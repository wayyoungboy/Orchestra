package terminal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type CommandValidator interface {
	ValidateCommand(cmd string) error
}

var (
	ErrProcessPoolFull   = errors.New("process pool is full")
	ErrSessionNotFound   = errors.New("session not found")
	ErrCommandNotAllowed = errors.New("command not allowed by whitelist")
)

type ProcessConfig struct {
	ID           string
	Command      string
	Args         []string
	Workspace    string
	WorkspaceID  string
	MemberID     string
	MemberName   string
	TerminalType string
	Env          []string
}

type ProcessPool struct {
	mu          sync.RWMutex
	sessions    map[string]*ProcessSession
	maxSessions int
	idleTimeout time.Duration
	validator   CommandValidator
	outputHook  func(*ProcessSession, []byte)
}

func NewProcessPool(maxSessions int, idleTimeout time.Duration) *ProcessPool {
	p := &ProcessPool{
		sessions:    make(map[string]*ProcessSession),
		maxSessions: maxSessions,
		idleTimeout: idleTimeout,
	}
	go p.cleanupIdleSessions()
	return p
}

func (p *ProcessPool) SetValidator(v CommandValidator) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.validator = v
}

// SetOutputHook receives raw PTY output for optional side effects (e.g. chat bridge).
func (p *ProcessPool) SetOutputHook(fn func(*ProcessSession, []byte)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.outputHook = fn
}

// WorkspaceTerminalSession is a serializable summary for HTTP APIs (REQ-303 polling).
type WorkspaceTerminalSession struct {
	MemberID  string `json:"memberId"`
	SessionID string `json:"sessionId"`
	PID       int    `json:"pid"`
}

// ListSessionsForWorkspace returns all in-memory PTY sessions bound to the workspace.
func (p *ProcessPool) ListSessionsForWorkspace(workspaceID string) []WorkspaceTerminalSession {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]WorkspaceTerminalSession, 0)
	for _, s := range p.sessions {
		if s.WorkspaceID == workspaceID {
			out = append(out, WorkspaceTerminalSession{
				MemberID:  s.MemberID,
				SessionID: s.ID,
				PID:       s.PID,
			})
		}
	}
	return out
}

// SessionForWorkspaceMember returns a running session for the given workspace and member, if any.
func (p *ProcessPool) SessionForWorkspaceMember(workspaceID, memberID string) *ProcessSession {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, s := range p.sessions {
		if s.WorkspaceID == workspaceID && s.MemberID == memberID {
			return s
		}
	}
	return nil
}

func (p *ProcessPool) Acquire(ctx context.Context, config ProcessConfig) (*ProcessSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Validate command against whitelist
	if p.validator != nil {
		if err := p.validator.ValidateCommand(config.Command); err != nil {
			return nil, ErrCommandNotAllowed
		}
	}

	if len(p.sessions) >= p.maxSessions {
		return nil, ErrProcessPoolFull
	}

	// Generate session ID if not provided
	if config.ID == "" {
		config.ID = fmt.Sprintf("session-%d", time.Now().UnixNano())
	}

	session, err := createSession(config)
	if err != nil {
		return nil, err
	}
	if p.outputHook != nil {
		session.SetOutputHook(p.outputHook)
	}

	p.sessions[session.ID] = session
	return session, nil
}

func (p *ProcessPool) Get(id string) (*ProcessSession, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	session, ok := p.sessions[id]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (p *ProcessPool) Release(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if session, ok := p.sessions[id]; ok {
		session.Kill()
		delete(p.sessions, id)
	}
}

func (p *ProcessPool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.sessions)
}

func (p *ProcessPool) cleanupIdleSessions() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		for id, session := range p.sessions {
			if time.Since(session.LastActive) > p.idleTimeout {
				session.Kill()
				delete(p.sessions, id)
			}
		}
		p.mu.Unlock()
	}
}
