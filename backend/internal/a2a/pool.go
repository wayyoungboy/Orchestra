package a2a

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2aclient"
	"github.com/a2aproject/a2a-go/v2/a2aclient/agentcard"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/pkg/utils"
)

// Pool manages A2A agent sessions.
type Pool struct {
	mu          sync.RWMutex
	sessions    map[string]*Session // sessionID -> Session
	agentURLs   map[string]string   // "workspaceID:memberID" -> agentURL
	registry    *AgentRegistry
	toolHandler *ToolHandler
	outputHook  func(*Session, *ACPMessage)
	idleTimeout time.Duration
}

// NewPool creates a new A2A session pool.
func NewPool(idleTimeout time.Duration, registry *AgentRegistry) *Pool {
	p := &Pool{
		sessions:    make(map[string]*Session),
		agentURLs:   make(map[string]string),
		registry:    registry,
		idleTimeout: idleTimeout,
	}
	go p.cleanupIdleSessions()
	return p
}

// SessionConfig contains parameters for creating an A2A session.
type SessionConfig struct {
	ID           string
	WorkspaceID  string
	MemberID     string
	MemberName   string
	TerminalType string
	Member       *models.Member
}

// Acquire creates or retrieves an A2A session for a member.
func (p *Pool) Acquire(ctx context.Context, config SessionConfig) (*Session, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we already have an active session for this member
	for _, sess := range p.sessions {
		if sess.WorkspaceID == config.WorkspaceID && sess.MemberID == config.MemberID {
			sess.mu.Lock()
			sess.lastActive = time.Now()
			sess.mu.Unlock()
			return sess, nil
		}
	}

	// Resolve agent URL from member config or registry
	agentURL := p.resolveAgentURL(config)
	if agentURL == "" {
		return nil, nil
	}

	// Create A2A client
	client, err := p.createClient(agentURL)
	if err != nil {
		return nil, err
	}

	// Generate session ID if not provided
	sessionID := config.ID
	if sessionID == "" {
		sessionID = utils.GenerateID()
	}

	// Create session
	sess := NewSession(
		sessionID,
		config.WorkspaceID,
		config.MemberID,
		config.MemberName,
		config.TerminalType,
		client,
		agentURL,
	)

	// Start output processing goroutine
	go p.processOutput(sess)

	// Register in pool
	p.sessions[sessionID] = sess
	p.agentURLs[config.WorkspaceID+":"+config.MemberID] = agentURL

	return sess, nil
}

// Get retrieves a session by ID.
func (p *Pool) Get(id string) *Session {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.sessions[id]
}

// Release closes and removes a session.
func (p *Pool) Release(id string) {
	p.mu.Lock()
	sess := p.sessions[id]
	if sess != nil {
		delete(p.sessions, id)
		// Clean up agent URL mapping
		for key, url := range p.agentURLs {
			if url == sess.AgentURL {
				delete(p.agentURLs, key)
			}
		}
	}
	p.mu.Unlock()

	if sess != nil {
		sess.Release()
		_ = sess.Client.Destroy()
	}
}

// SessionForWorkspaceMember returns the session for a specific workspace member.
func (p *Pool) SessionForWorkspaceMember(workspaceID, memberID string) *Session {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, sess := range p.sessions {
		if sess.WorkspaceID == workspaceID && sess.MemberID == memberID {
			return sess
		}
	}
	return nil
}

// ListSessionsForWorkspace returns session info for a workspace.
func (p *Pool) ListSessionsForWorkspace(workspaceID string) []WorkspaceSessionInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var infos []WorkspaceSessionInfo
	for _, sess := range p.sessions {
		if sess.WorkspaceID == workspaceID {
			infos = append(infos, WorkspaceSessionInfo{
				MemberID:  sess.MemberID,
				SessionID: sess.ID,
				PID:       0, // No PID for HTTP-based sessions
			})
		}
	}
	return infos
}

// SetToolHandler sets the tool execution handler.
func (p *Pool) SetToolHandler(h *ToolHandler) {
	p.toolHandler = h
}

// SetOutputHook sets the output message callback (used by ACPBridge/AgentBridge).
func (p *Pool) SetOutputHook(fn func(*Session, *ACPMessage)) {
	p.outputHook = fn
}

// WorkspaceSessionInfo provides session metadata for the frontend.
type WorkspaceSessionInfo struct {
	MemberID  string `json:"memberId"`
	SessionID string `json:"sessionId"`
	PID       int    `json:"pid"`
}

// resolveAgentURL determines the A2A agent URL for a session.
func (p *Pool) resolveAgentURL(config SessionConfig) string {
	// Check member's A2A config first
	if config.Member != nil && config.Member.A2AEnabled && config.Member.A2AAgentURL != "" {
		return config.Member.A2AAgentURL
	}

	// Check registry
	key := config.WorkspaceID + ":" + config.MemberID
	if entry := p.registry.Get(key); entry != nil {
		return entry.AgentURL
	}

	return ""
}

// createClient creates an A2A client for the given agent URL.
func (p *Pool) createClient(agentURL string) (*a2aclient.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to resolve agent card first
	resolver := agentcard.NewResolver(nil)
	if card, err := resolver.Resolve(ctx, agentURL); err == nil && card != nil {
		return a2aclient.NewFromCard(ctx, card)
	}

	// Fallback: create client directly with endpoint
	return a2aclient.NewFromEndpoints(ctx, []*a2a.AgentInterface{
		a2a.NewAgentInterface(agentURL, a2a.TransportProtocolJSONRPC),
	})
}

// processOutput reads from a session's OutputChan and invokes the output hook.
// It exits when the session is done OR when the session has been released from the pool.
func (p *Pool) processOutput(sess *Session) {
	for {
		select {
		case <-sess.DoneChan:
			return
		case msg, ok := <-sess.OutputChan:
			if !ok {
				return
			}
			// Check if session was released from pool before processing
			p.mu.RLock()
			_, exists := p.sessions[sess.ID]
			p.mu.RUnlock()
			if !exists {
				return // Session was released, stop processing
			}
			if msg.Type == TypeToolUse && p.toolHandler != nil {
				go p.toolHandler.ExecuteTool(msg, sess)
			}
			if p.outputHook != nil {
				p.outputHook(sess, msg)
			}
		case err := <-sess.ErrorChan:
			log.Printf("[a2a-pool] Error from session %s: %v", sess.MemberName, err)
		}
	}
}

// cleanupIdleSessions periodically removes idle sessions.
func (p *Pool) cleanupIdleSessions() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		var toRelease []string
		for id, sess := range p.sessions {
			sess.mu.Lock()
			if time.Since(sess.lastActive) > p.idleTimeout {
				toRelease = append(toRelease, id)
			}
			sess.mu.Unlock()
		}
		p.mu.Unlock()

		for _, id := range toRelease {
			p.Release(id)
		}
	}
}
