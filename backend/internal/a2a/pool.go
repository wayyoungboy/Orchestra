package a2a

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2aclient"
	"github.com/a2aproject/a2a-go/v2/a2aclient/agentcard"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/pkg/utils"
)

// Pool manages agent sessions (both A2A HTTP-based and Local CLI).
type Pool struct {
	mu            sync.RWMutex
	sessions      map[string]*Session // sessionID -> Session
	agentURLs     map[string]string   // "workspaceID:memberID" -> agentURL
	registry      *AgentRegistry
	toolHandler   *ToolHandler
	outputHook    func(*Session, *ACPMessage)
	idleTimeout   time.Duration
	workspacePath string // server-side workspace root path for local agents
}

// NewPool creates a new session pool.
func NewPool(idleTimeout time.Duration, registry *AgentRegistry, workspacePath string) *Pool {
	p := &Pool{
		sessions:      make(map[string]*Session),
		agentURLs:     make(map[string]string),
		registry:      registry,
		idleTimeout:   idleTimeout,
		workspacePath: workspacePath,
	}
	go p.cleanupIdleSessions()
	return p
}

// SessionConfig contains parameters for creating a session.
type SessionConfig struct {
	ID            string
	WorkspaceID   string
	WorkspaceDir  string // per-workspace directory path
	MemberID      string
	MemberName    string
	TerminalType  string
	Member        *models.Member
}

// Acquire creates or retrieves a session for a member.
// Supports both A2A (HTTP-based) and Local CLI agents.
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

	// Priority 1: Check for local CLI agent (ACP mode)
	if config.Member != nil && config.Member.ACPEnabled && config.Member.ACPCommand != "" {
		sess, err := p.createLocalSession(ctx, config)
		if err != nil {
			return nil, err
		}
		if sess != nil {
			return sess, nil
		}
	}

	// Priority 2: Check for A2A (HTTP-based) agent
	agentURL := p.resolveAgentURL(config)
	if agentURL != "" {
		sess, err := p.createA2ASession(ctx, config, agentURL)
		if err != nil {
			return nil, err
		}
		if sess != nil {
			return sess, nil
		}
	}

	// No agent configured for this member
	return nil, nil
}

// createLocalSession creates a session using a local CLI subprocess.
func (p *Pool) createLocalSession(ctx context.Context, config SessionConfig) (*Session, error) {
	command := config.Member.ACPCommand
	args := config.Member.ACPArgs
	if args == nil {
		args = []string{}
	}

	runner := NewLocalRunner(command, args, config.WorkspaceDir)

	// Generate session ID if not provided
	sessionID := config.ID
	if sessionID == "" {
		sessionID = "local_" + utils.GenerateID()
	}

	// Create session without A2A client
	sess := NewSession(
		sessionID,
		config.WorkspaceID,
		config.MemberID,
		config.MemberName,
		config.TerminalType,
		nil, // no A2A client for local runner
		"",  // no agent URL
	)
	sess.localRunner = runner

	// Wire up runner callbacks
	runner.onOutput = func(text string) {
		if text == "" {
			return
		}
		// Convert raw text to ACP message
		acpMsg := &ACPMessage{
			Type: TypeAssistantMessage,
		}
		// Try to parse as JSON first
		content, _ := marshalACPContent("assistant_message", text)
		acpMsg.Content = content
		sess.OutputChan <- acpMsg
	}
	runner.onEvent = func(typ string, data any) {
		switch typ {
		case "assistant":
			// Extract text from assistant message and emit as ACP
			if text := extractAssistantTextFromRaw(data); text != "" {
				acpMsg := &ACPMessage{
					Type: TypeAssistantMessage,
				}
				acpMsg.Content, _ = marshalACPContent("assistant_message", text)
				sess.OutputChan <- acpMsg
			}
		case "result":
			// Final result from Claude - skip, assistant message already emitted
			// Only log for debugging
			if raw, ok := data.(map[string]any); ok {
				if res, _ := raw["result"].(string); ok && res != "" {
					log.Printf("[local-runner] Claude session completed: result=%s", res)
				}
			}
		case "control_request":
			// Auto-approved internally, no output needed
		case "system":
			// System initialized, log it
			if raw, ok := data.(map[string]any); ok {
				if sid, _ := raw["session_id"].(string); sid != "" {
					log.Printf("[local-runner] Claude session ID: %s", sid)
				}
			}
		}
	}
	runner.onError = func(err error) {
		log.Printf("[local-runner] Error from %s: %v", config.MemberName, err)
		sess.ErrorChan <- err
	}

	// Start the runner
	if err := runner.Start(ctx); err != nil {
		return nil, err
	}

	// Start output processing goroutine
	go p.processOutput(sess)

	// Register in pool
	p.sessions[sessionID] = sess

	return sess, nil
}

// createA2ASession creates a session using A2A HTTP protocol.
func (p *Pool) createA2ASession(ctx context.Context, config SessionConfig, agentURL string) (*Session, error) {
	client, err := p.createClient(agentURL)
	if err != nil {
		return nil, err
	}

	sessionID := config.ID
	if sessionID == "" {
		sessionID = utils.GenerateID()
	}

	sess := NewSession(
		sessionID,
		config.WorkspaceID,
		config.MemberID,
		config.MemberName,
		config.TerminalType,
		client,
		agentURL,
	)

	go p.processOutput(sess)

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
		if sess.AgentURL != "" {
			for key, url := range p.agentURLs {
				if url == sess.AgentURL {
					delete(p.agentURLs, key)
				}
			}
		}
	}
	p.mu.Unlock()

	if sess != nil {
		sess.Release()
		// Stop local runner if present
		if sess.localRunner != nil {
			sess.localRunner.Stop()
		} else if sess.Client != nil {
			_ = sess.Client.Destroy()
		}
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
				PID:       0,
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
	if config.Member != nil && config.Member.A2AEnabled && config.Member.A2AAgentURL != nil && *config.Member.A2AAgentURL != "" {
		return *config.Member.A2AAgentURL
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
				go func() {
					// Capture tool use ID and name for result correlation
					toolUseParsed, _ := msg.ParseToolUseMessage()
					toolUseID := ""
					toolName := ""
					if toolUseParsed != nil {
						toolUseID = toolUseParsed.ToolUseID
						toolName = toolUseParsed.Name
					}

					// Execute the tool (broadcasts to frontend)
					p.toolHandler.ExecuteTool(msg, sess)

					// Send tool result back to agent's stdin so it can continue
					content := fmt.Sprintf("Tool '%s' executed successfully", toolName)
					if err := sess.SendToolResultToAgent(toolUseID, content, false); err != nil {
						log.Printf("[a2a-pool] Failed to send tool result to agent: %v", err)
					}
				}()
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
