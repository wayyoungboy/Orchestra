// Package a2a provides agent session management for Orchestra.
// Sessions are backed by tmux for true process persistence.
package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/orchestra/backend/internal/tmux"
)

// MessageType represents ACP message types.
type MessageType string

const (
	TypeUserMessage      MessageType = "user_message"
	TypeToolResult       MessageType = "tool_result"
	TypeAssistantMessage MessageType = "assistant_message"
	TypeToolUse          MessageType = "tool_use"
	TypeResult           MessageType = "result"
	TypeError            MessageType = "error"
	TypeSystem           MessageType = "system"
)

// ACPMessage is the internal message format for backward compatibility.
type ACPMessage struct {
	Type    MessageType
	Content json.RawMessage
}

// ParseAssistantMessage parses an ACPMessage as an assistant message.
func (m *ACPMessage) ParseAssistantMessage() (*AssistantMessage, error) {
	var msg AssistantMessage
	if err := json.Unmarshal(m.Content, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// ParseResultMessage parses an ACPMessage as a result message.
func (m *ACPMessage) ParseResultMessage() (*ResultMessage, error) {
	var msg ResultMessage
	if err := json.Unmarshal(m.Content, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// ParseErrorMessage parses an ACPMessage as an error message.
func (m *ACPMessage) ParseErrorMessage() (*ErrorMessage, error) {
	var msg ErrorMessage
	if err := json.Unmarshal(m.Content, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// ParseSystemMessage parses an ACPMessage as a system message.
func (m *ACPMessage) ParseSystemMessage() (*SystemMessage, error) {
	var msg SystemMessage
	if err := json.Unmarshal(m.Content, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// ParseToolUseMessage parses an ACPMessage as a tool use message.
func (m *ACPMessage) ParseToolUseMessage() (*ToolUseMessage, error) {
	var msg ToolUseMessage
	if err := json.Unmarshal(m.Content, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// Session represents an agent session backed by a tmux session.
// It implements the same interface as the old ACP session for ACPBridge compatibility.
type Session struct {
	ID           string
	WorkspaceID  string
	MemberID     string
	MemberName   string
	TerminalType string

	// Tmux session (replaces localRunner and A2A client)
	tmuxSession *tmux.TmuxSession

	mu             sync.Mutex
	lastActive     time.Time
	createdAt      time.Time
	lastChatConvID string
	terminalSubsMu sync.RWMutex
	outputSubs     map[chan *ACPMessage]struct{}
	chatStreamSubs map[chan []byte]struct{}
	errorSubs      map[chan error]struct{}
	streamSpanMu   sync.Mutex
	streamSpanID   string
	streamSeq      uint64

	// Output channels for ACPBridge / AgentBridge compatibility
	OutputChan chan *ACPMessage
	ErrorChan  chan error
	DoneChan   chan struct{}
	done       bool
	released   bool

	// Active subscriptions (kept for potential future use)
	subscriptions map[string]func() // taskID -> cancel
	subMu         sync.Mutex

	// Pending tool use tracking (for correlation when tool results arrive)
	pendingToolUses map[string]string // toolUseID -> taskID that initiated it
	toolUseMu       sync.Mutex

	// Test-only capture hook (set by tests to intercept SendUserMessage calls)
	testCaptureHook func(content string)
}

// NewSession creates a new session backed by a tmux session.
func NewSession(id, workspaceID, memberID, memberName, terminalType string, tmuxSession *tmux.TmuxSession) *Session {
	s := &Session{
		ID:              id,
		WorkspaceID:     workspaceID,
		MemberID:        memberID,
		MemberName:      memberName,
		TerminalType:    terminalType,
		tmuxSession:     tmuxSession,
		createdAt:       time.Now(),
		lastActive:      time.Now(),
		OutputChan:      make(chan *ACPMessage, 256),
		ErrorChan:       make(chan error, 16),
		DoneChan:        make(chan struct{}),
		outputSubs:      make(map[chan *ACPMessage]struct{}),
		chatStreamSubs:  make(map[chan []byte]struct{}),
		errorSubs:       make(map[chan error]struct{}),
		subscriptions:   make(map[string]func()),
		pendingToolUses: make(map[string]string),
	}
	return s
}

// SendUserMessage sends a user message to the agent via tmux.
func (s *Session) SendUserMessage(content string) error {
	s.mu.Lock()
	if s.released {
		s.mu.Unlock()
		return fmt.Errorf("session already released")
	}
	s.lastActive = time.Now()
	s.mu.Unlock()

	// Test capture hook
	if s.testCaptureHook != nil {
		s.testCaptureHook(content)
	}

	if s.tmuxSession == nil {
		return fmt.Errorf("no tmux session configured for %s", s.ID)
	}
	return s.tmuxSession.SendInput(content)
}

// SendToolResultToAgent sends a tool result back to the agent via tmux.
// For Claude's stream-json format, this sends a structured JSON message.
func (s *Session) SendToolResultToAgent(toolUseID, content string, isError bool) error {
	if s.tmuxSession == nil {
		return fmt.Errorf("no tmux session configured for %s", s.ID)
	}

	msg := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role": "user",
			"content": []map[string]any{
				{
					"type":        "tool_result",
					"tool_use_id": toolUseID,
					"content":     content,
					"is_error":    isError,
				},
			},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return s.tmuxSession.SendInput(string(data))
}

// SendToolResult sends a tool result (alias for SendToolResultToAgent for WS compatibility).
func (s *Session) SendToolResult(toolUseID, content string, isError bool) error {
	return s.SendToolResultToAgent(toolUseID, content, isError)
}

// IsAlive returns true if the tmux session still exists.
func (s *Session) IsAlive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tmuxSession == nil {
		return false
	}
	return s.tmuxSession.IsAlive()
}

// Release closes the session but keeps the tmux session alive.
func (s *Session) Release() {
	s.subMu.Lock()
	for _, cancel := range s.subscriptions {
		cancel()
	}
	s.subscriptions = make(map[string]func())
	s.subMu.Unlock()

	// Stop tmux session (does NOT kill the tmux process)
	if s.tmuxSession != nil {
		s.tmuxSession.Stop()
	}
	s.closeTerminalSubscribers()

	s.mu.Lock()
	s.released = true
	if !s.done {
		s.done = true
		close(s.DoneChan)
	}
	s.mu.Unlock()
}

// Kill closes the session and kills the tmux session.
func (s *Session) Kill() {
	s.subMu.Lock()
	for _, cancel := range s.subscriptions {
		cancel()
	}
	s.subscriptions = make(map[string]func())
	s.subMu.Unlock()

	if s.tmuxSession != nil {
		s.tmuxSession.Kill()
	}
	s.closeTerminalSubscribers()

	s.mu.Lock()
	s.released = true
	if !s.done {
		s.done = true
		select {
		case <-s.DoneChan:
		default:
			close(s.DoneChan)
		}
	}
	s.mu.Unlock()
}

// Chat bridge methods

// LastChatTargetConversation returns the conversation ID bound for chat bridging.
func (s *Session) LastChatTargetConversation() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastChatConvID
}

// SetLastChatTargetConversation binds a conversation ID for chat bridging.
func (s *Session) SetLastChatTargetConversation(convID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastChatConvID = convID
}

// GetWorkspaceID returns the workspace ID.
func (s *Session) GetWorkspaceID() string {
	return s.WorkspaceID
}

// GetMemberID returns the member ID.
func (s *Session) GetMemberID() string {
	return s.MemberID
}

// GetMemberName returns the member name.
func (s *Session) GetMemberName() string {
	return s.MemberName
}

// SetStreamSpanID sets the span ID for streaming chat.
func (s *Session) SetStreamSpanID(id string) {
	s.streamSpanMu.Lock()
	defer s.streamSpanMu.Unlock()
	s.streamSpanID = id
}

// StreamSpanID returns the current stream span ID.
func (s *Session) StreamSpanID() string {
	s.streamSpanMu.Lock()
	defer s.streamSpanMu.Unlock()
	return s.streamSpanID
}

// NextStreamSeq returns and increments the stream sequence number.
func (s *Session) NextStreamSeq() uint64 {
	s.streamSpanMu.Lock()
	defer s.streamSpanMu.Unlock()
	s.streamSeq++
	return s.streamSeq
}

// SubscribeTerminal creates a dedicated stream for one terminal WebSocket.
// Pool owns OutputChan and fans messages out to these subscriptions, so chat,
// tool execution, and multiple viewers cannot consume each other's events.
func (s *Session) SubscribeTerminal() (<-chan *ACPMessage, <-chan []byte, <-chan error, func()) {
	output := make(chan *ACPMessage, 256)
	chatStream := make(chan []byte, 256)
	errors := make(chan error, 16)

	s.terminalSubsMu.Lock()
	s.outputSubs[output] = struct{}{}
	s.chatStreamSubs[chatStream] = struct{}{}
	s.errorSubs[errors] = struct{}{}
	s.terminalSubsMu.Unlock()

	var once sync.Once
	return output, chatStream, errors, func() {
		once.Do(func() {
			s.terminalSubsMu.Lock()
			delete(s.outputSubs, output)
			delete(s.chatStreamSubs, chatStream)
			delete(s.errorSubs, errors)
			close(output)
			close(chatStream)
			close(errors)
			s.terminalSubsMu.Unlock()
		})
	}
}

// PublishOutput sends an agent event to each terminal subscriber without
// blocking the agent pipeline on a slow browser.
func (s *Session) PublishOutput(msg *ACPMessage) {
	s.terminalSubsMu.RLock()
	defer s.terminalSubsMu.RUnlock()
	for ch := range s.outputSubs {
		select {
		case ch <- msg:
		default:
			log.Printf("[a2a] WARN: terminal output dropped for session %s", s.ID)
		}
	}
}

// PublishError sends an error to each terminal subscriber.
func (s *Session) PublishError(err error) {
	s.terminalSubsMu.RLock()
	defer s.terminalSubsMu.RUnlock()
	for ch := range s.errorSubs {
		select {
		case ch <- err:
		default:
			log.Printf("[a2a] WARN: terminal error dropped for session %s", s.ID)
		}
	}
}

// TrySendChatStream sends data to every terminal subscriber.
func (s *Session) TrySendChatStream(data []byte) {
	s.terminalSubsMu.RLock()
	defer s.terminalSubsMu.RUnlock()
	for ch := range s.chatStreamSubs {
		select {
		case ch <- data:
		default:
			log.Printf("[a2a] WARN: terminal chat stream dropped for session %s", s.ID)
		}
	}
}

func (s *Session) closeTerminalSubscribers() {
	s.terminalSubsMu.Lock()
	defer s.terminalSubsMu.Unlock()
	// WebSocket writers also select on DoneChan and then invoke their own
	// unsubscribe callback. Do not close the per-client channels here: that
	// callback may already be pending and would otherwise double-close them.
	s.outputSubs = make(map[chan *ACPMessage]struct{})
	s.chatStreamSubs = make(map[chan []byte]struct{})
	s.errorSubs = make(map[chan error]struct{})
}

// CaptureScrollback captures the last N lines of pane output (for recovery).
func (s *Session) CaptureScrollback(ctx context.Context, lines int) (string, error) {
	if s.tmuxSession == nil {
		return "", fmt.Errorf("no tmux session")
	}
	return s.tmuxSession.CaptureScrollback(ctx, lines)
}

// GetState returns the current session state.
func (s *Session) GetState() string {
	if s.tmuxSession == nil {
		return "offline"
	}
	return string(s.tmuxSession.GetState())
}

// TmuxSession returns the underlying tmux session (for direct access).
func (s *Session) TmuxSession() *tmux.TmuxSession {
	return s.tmuxSession
}

// marshalACPContent creates ACP content JSON.
func marshalACPContent(typ, text string) (json.RawMessage, error) {
	return json.Marshal(map[string]string{
		"type":    typ,
		"content": text,
	})
}

// ConvertACPToWS converts an ACP message to a WebSocket response.
// This function is used by the A2A terminal handler.
func ConvertACPToWS(msg *ACPMessage) *ACPTerminalResponse {
	if msg == nil {
		return nil
	}
	switch msg.Type {
	case TypeAssistantMessage:
		parsed, err := msg.ParseAssistantMessage()
		if err != nil {
			return nil
		}
		return &ACPTerminalResponse{
			Type:    "assistant_message",
			Content: parsed.Content,
		}

	case TypeToolUse:
		parsed, err := msg.ParseToolUseMessage()
		if err != nil {
			return nil
		}
		return &ACPTerminalResponse{
			Type:      "tool_use",
			ToolName:  parsed.Name,
			ToolInput: parsed.Input,
			ToolUseID: parsed.ToolUseID,
		}

	case TypeResult:
		parsed, err := msg.ParseResultMessage()
		if err != nil {
			return nil
		}
		return &ACPTerminalResponse{
			Type:       "result",
			Message:    parsed.Message,
			CostUSD:    parsed.CostUSD,
			DurationMs: parsed.DurationMs,
		}

	case TypeError:
		parsed, err := msg.ParseErrorMessage()
		if err != nil {
			return nil
		}
		return &ACPTerminalResponse{
			Type:  "error",
			Error: parsed.Error,
		}

	case TypeSystem:
		parsed, err := msg.ParseSystemMessage()
		if err != nil {
			return nil
		}
		return &ACPTerminalResponse{
			Type:   "status",
			Status: parsed.Message,
		}

	default:
		return nil
	}
}
