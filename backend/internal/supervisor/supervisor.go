// Package supervisor manages agent process lifecycles with worker pooling and queuing.
package supervisor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/orchestra/backend/internal/eventbus"
	"github.com/orchestra/backend/internal/messagequeue"
	"github.com/orchestra/backend/internal/provider"
)

// Config holds supervisor configuration.
type Config struct {
	MaxWorkers       int
	StaleThreshold   time.Duration
	BucketSwapPeriod time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxWorkers:       5,
		StaleThreshold:   5 * time.Minute,
		BucketSwapPeriod: 15 * time.Second,
	}
}

// SessionState tracks the current state of a session.
type SessionState string

const (
	StateIdle        SessionState = "idle"
	StateInTurn      SessionState = "in-turn"
	StateWaitingInput SessionState = "waiting-input"
	StateTerminated  SessionState = "terminated"
)

// SessionInfo holds metadata about an active session.
type SessionInfo struct {
	SessionID     string
	WorkspaceID   string
	MemberID      string
	State         SessionState
	Provider      string
	LastMessageAt time.Time
	CreatedAt     time.Time
}

// Supervisor manages agent sessions with worker pooling and staleness detection.
type Supervisor struct {
	mu              sync.RWMutex
	sessions        map[string]*runningSession
	everOwnedSessions map[string]bool
	config          Config
	eventBus        *eventbus.EventBus
	registry        *provider.Registry
	cancelFunc      context.CancelFunc
}

type runningSession struct {
	info          SessionInfo
	session       provider.AgentSession
	provider      provider.AgentProvider
	msgQueue      *messagequeue.Queue[provider.AgentMessage]
	cancel        context.CancelFunc
	lastMsgTime   time.Time
	messageBucket messageBucket
}

// messageBucket implements a two-bucket swap pattern for bounded history.
type messageBucket struct {
	current  []provider.AgentMessage
	previous []provider.AgentMessage
	timer    *time.Timer
}

// New creates a new Supervisor.
func New(cfg Config, eb *eventbus.EventBus, registry *provider.Registry) *Supervisor {
	if eb == nil {
		eb = eventbus.New()
	}
	return &Supervisor{
		sessions:        make(map[string]*runningSession),
		everOwnedSessions: make(map[string]bool),
		config:          cfg,
		eventBus:        eb,
		registry:        registry,
	}
}

// StartSession creates or resumes a session for a member.
func (s *Supervisor) StartSession(ctx context.Context, workspaceID, memberID string, opts provider.SessionOptions) (provider.AgentSession, error) {
	sessionID := opts.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s-%s", workspaceID, memberID)
	}

	s.mu.Lock()

	// Check if session already exists
	if existing, ok := s.sessions[sessionID]; ok {
		if existing.session.IsAlive() {
			s.mu.Unlock()
			s.eventBus.EmitPayload(eventbus.EventProcessStateChanged, map[string]any{
				"sessionId": sessionID,
				"state":     string(StateInTurn),
			})
			return existing.session, nil
		}
		// Session died — clean up
		s.removeSessionLocked(sessionID)
	}

	s.everOwnedSessions[sessionID] = true
	s.mu.Unlock()

	// Find the right provider
	prov := s.registry.Get(provider.ProviderName(opts.WorkspacePath))
	if prov == nil {
		// Default to Claude
		prov = s.registry.Get(provider.ProviderClaude)
	}
	if prov == nil {
		return nil, fmt.Errorf("no provider available")
	}

	// Create new session context
	sessCtx, cancel := context.WithCancel(ctx)
	sess, err := prov.StartSession(sessCtx, opts)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	msgQ := messagequeue.New[provider.AgentMessage](64)

	info := SessionInfo{
		SessionID:   sessionID,
		WorkspaceID: workspaceID,
		MemberID:    memberID,
		State:       StateInTurn,
		Provider:    string(prov.Name()),
		LastMessageAt: time.Now(),
		CreatedAt:   time.Now(),
	}

	rs := &runningSession{
		info:     info,
		session:  sess,
		provider: prov,
		msgQueue: msgQ,
		cancel:   cancel,
		lastMsgTime: time.Now(),
		messageBucket: messageBucket{
			current: make([]provider.AgentMessage, 0, 128),
			timer:   time.NewTimer(s.config.BucketSwapPeriod),
		},
	}

	s.mu.Lock()
	s.sessions[sessionID] = rs
	s.mu.Unlock()

	s.eventBus.EmitPayload(eventbus.EventSessionCreated, map[string]any{
		"sessionId": sessionID,
		"provider":  prov.Name(),
		"workspaceId": workspaceID,
		"memberId":  memberID,
	})
	s.eventBus.EmitPayload(eventbus.EventProcessStateChanged, map[string]any{
		"sessionId": sessionID,
		"state":     string(StateInTurn),
	})

	return sess, nil
}

// GetSession returns a session by ID.
func (s *Supervisor) GetSession(sessionID string) (SessionInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rs, ok := s.sessions[sessionID]
	if !ok {
		return SessionInfo{}, false
	}
	return rs.info, true
}

// TerminateSession stops a session and cleans up.
func (s *Supervisor) TerminateSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.removeSessionLocked(sessionID)
}

// ActiveSessions returns all active session IDs.
func (s *Supervisor) ActiveSessions() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.sessions))
	for id := range s.sessions {
		ids = append(ids, id)
	}
	return ids
}

// ActiveSessionCount returns the number of active sessions.
func (s *Supervisor) ActiveSessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

// IsAtCapacity reports whether the supervisor has reached max workers.
func (s *Supervisor) IsAtCapacity() bool {
	if s.config.MaxWorkers <= 0 {
		return false
	}
	return s.ActiveSessionCount() >= s.config.MaxWorkers
}

// CheckStaleSessions terminates sessions that have been silent too long.
func (s *Supervisor) CheckStaleSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, rs := range s.sessions {
		if rs.info.State != StateInTurn {
			continue
		}
		silent := now.Sub(rs.lastMsgTime)
		if silent < s.config.StaleThreshold {
			continue
		}
		if rs.session != nil && rs.session.IsAlive() {
			continue // Still alive, probably doing heavy work
		}
		_ = s.removeSessionLocked(id)
		s.eventBus.EmitPayload(eventbus.EventProcessStateChanged, map[string]any{
			"sessionId": id,
			"state":     string(StateTerminated),
			"reason":    "stale",
		})
	}
}

func (s *Supervisor) removeSessionLocked(sessionID string) error {
	rs, ok := s.sessions[sessionID]
	if !ok {
		return nil
	}
	rs.cancel()
	if rs.messageBucket.timer != nil {
		rs.messageBucket.timer.Stop()
	}
	delete(s.sessions, sessionID)
	return nil
}
