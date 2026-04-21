// Package tmux provides tmux-based agent session management.
package tmux

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// MessageType matches a2a.MessageType but defined here to avoid import cycles.
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

// ACPMessage is the internal message format compatible with a2a.ACPMessage.
type ACPMessage struct {
	Type    MessageType
	Content json.RawMessage
}

// MarshalACPContent creates JSON content for the given type and text.
func MarshalACPContent(typ, text string) (json.RawMessage, error) {
	return json.Marshal(map[string]string{
		"type":    typ,
		"content": text,
	})
}

// SessionState represents the lifecycle state of a tmux session.
type SessionState string

const (
	StateConnecting SessionState = "connecting" // starting, post-ready steps
	StateOnline     SessionState = "online"     // ready for input
	StateWorking    SessionState = "working"    // actively processing
	StateOffline    SessionState = "offline"    // terminated
)

// TmuxSession represents a tmux-backed agent session.
type TmuxSession struct {
	Name         string // tmux session name (e.g., "orchestra-abc123-def456")
	SessionID    string // Orchestra session ID
	WorkspaceID  string
	MemberID     string
	MemberName   string
	TerminalType string
	CWD          string
	Command      string
	Args         []string

	// Output channel for pool/bridge integration
	OutputChan chan *ACPMessage
	ErrorChan  chan error
	DoneChan   chan struct{}

	// Chat bridge fields
	streamSpanID string
	streamSeq    uint64
	streamSpanMu sync.Mutex

	// State machine
	state      SessionState
	stateMu    sync.Mutex
	lastActive time.Time
	createdAt  time.Time

	// Output capture: pipe-pane log file
	LogFile       string
	logReaderDone chan struct{}
	logReaderCtx  context.CancelFunc

	// Control
	mu   sync.Mutex
	done bool

	// Post-ready
	shellReady bool

	// Pending tool use tracking
	pendingToolUses map[string]string
	toolUseMu       sync.Mutex

	// Read offset tracking for recovery
	readOffset int64
}

// NewTmuxSession creates a new TmuxSession without starting it.
func NewTmuxSession(sessionID, workspaceID, memberID, memberName, terminalType, name, cwd, command string, args []string) *TmuxSession {
	return &TmuxSession{
		Name:            name,
		SessionID:       sessionID,
		WorkspaceID:     workspaceID,
		MemberID:        memberID,
		MemberName:      memberName,
		TerminalType:    terminalType,
		CWD:             cwd,
		Command:         command,
		Args:            args,
		state:           StateConnecting,
		createdAt:       time.Now(),
		lastActive:      time.Now(),
		OutputChan:      make(chan *ACPMessage, 256),
		ErrorChan:       make(chan error, 16),
		DoneChan:        make(chan struct{}),
		logReaderDone:   make(chan struct{}),
		pendingToolUses: make(map[string]string),
	}
}

// SendInput sends text to the tmux session using literal mode, followed by Enter.
func (s *TmuxSession) SendInput(text string) error {
	s.mu.Lock()
	if s.done {
		s.mu.Unlock()
		return fmt.Errorf("session is closed")
	}
	s.lastActive = time.Now()
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mgr := NewManager("")
	if err := mgr.SendKeys(ctx, s.Name, text); err != nil {
		return fmt.Errorf("tmux send-keys: %w", err)
	}
	if err := mgr.SendEnter(ctx, s.Name); err != nil {
		return fmt.Errorf("tmux send-keys Enter: %w", err)
	}
	return nil
}

// CaptureScrollback captures the last N lines of pane output.
func (s *TmuxSession) CaptureScrollback(ctx context.Context, lines int) (string, error) {
	mgr := NewManager("")
	return mgr.CapturePane(ctx, s.Name, lines)
}

// SetupPipePane configures tmux pipe-pane to append output to a log file.
// Should be called after the tmux session is created.
func (s *TmuxSession) SetupPipePane(ctx context.Context) error {
	if s.LogFile == "" {
		s.LogFile = fmt.Sprintf("/tmp/orch-%s.log", s.Name)
	}

	// Ensure log file exists
	f, err := os.OpenFile(s.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("create log file: %w", err)
	}
	f.Close()

	mgr := NewManager("")
	return mgr.SetupPipePane(ctx, s.Name, s.LogFile)
}

// StartOutputReader starts reading from the log file and parsing JSON lines.
// For recovered sessions, reads from beginning; for new sessions, from current end.
func (s *TmuxSession) StartOutputReader(ctx context.Context, fromBeginning bool) error {
	if s.LogFile == "" {
		s.LogFile = fmt.Sprintf("/tmp/orch-%s.log", s.Name)
	}

	// Ensure log file exists
	f, err := os.OpenFile(s.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("create log file: %w", err)
	}

	if fromBeginning {
		// For recovery: start from beginning
		s.readOffset = 0
	} else {
		// For new sessions: skip existing content, only read new output
		if info, err := f.Stat(); err == nil {
			s.readOffset = info.Size()
		}
	}
	f.Close()

	// Start reader goroutine
	readerCtx, cancel := context.WithCancel(ctx)
	s.logReaderCtx = cancel
	go s.readLogFile(readerCtx)

	return nil
}

// readLogFile tails the log file, starting from s.readOffset.
func (s *TmuxSession) readLogFile(ctx context.Context) {
	defer close(s.logReaderDone)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		file, err := os.Open(s.LogFile)
		if err != nil {
			s.ErrorChan <- fmt.Errorf("open log file: %w", err)
			return
		}

		// Seek to last read position
		if _, err := file.Seek(s.readOffset, 0); err != nil {
			file.Close()
			s.ErrorChan <- fmt.Errorf("seek log file: %w", err)
			return
		}

		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024) // 10MB max line

		readAny := false
		for scanner.Scan() {
			readAny = true
			line := scanner.Text()
			if line != "" {
				s.ParseAndEmit(line)
			}
			// Update offset
			s.readOffset = fileOffset(scanner, file)
		}

		if err := scanner.Err(); err != nil {
			file.Close()
			s.ErrorChan <- fmt.Errorf("log scanner: %w", err)
			return
		}

		file.Close()

		// If no new data, wait before retrying
		if !readAny {
			select {
			case <-time.After(200 * time.Millisecond):
				continue
			case <-ctx.Done():
				return
			}
		}
		// Small delay before retry to avoid busy-looping
		select {
		case <-time.After(50 * time.Millisecond):
		case <-ctx.Done():
			return
		}
	}
}

// fileOffset returns the current file position after the last successful scan.
func fileOffset(scanner *bufio.Scanner, file *os.File) int64 {
	// Approximate: use current file position
	pos, _ := file.Seek(0, 1)
	return pos
}

// ParseAndEmit parses a line of agent output and emits the appropriate ACP message.
// This replaces local_runner.go's handleJSONLine.
func (s *TmuxSession) ParseAndEmit(line string) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		// Not JSON, emit as assistant message if non-empty
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			s.emitAssistantMessage(trimmed)
		}
		return
	}

	typ, _ := raw["type"].(string)
	if typ == "" {
		// Check for JSON-RPC format
		if _, hasJSONRPC := raw["jsonrpc"]; hasJSONRPC {
			s.handleJSONRPC(raw)
		}
		return
	}

	switch typ {
	case "system":
		subtype, _ := raw["subtype"].(string)
		if subtype == "init" {
			sessionID, _ := raw["session_id"].(string)
			if sessionID != "" {
				log.Printf("[tmux-session] Agent %s session ID: %s", s.MemberName, sessionID)
			}
			// System initialized → mark shell ready
			s.mu.Lock()
			s.shellReady = true
			s.mu.Unlock()
			s.SetState(StateOnline)
		}
		s.emitSystemMessage(raw)

	case "assistant":
		// Extract text and tool_use from assistant message
		text := extractAssistantText(raw)
		toolUses := extractToolUses(raw)

		if len(toolUses) > 0 {
			for _, tu := range toolUses {
				s.toolUseMu.Lock()
				s.pendingToolUses[tu.ToolUseID] = ""
				s.toolUseMu.Unlock()

				data, _ := json.Marshal(tu)
				s.OutputChan <- &ACPMessage{
					Type:    TypeToolUse,
					Content: data,
				}
			}
		}

		if text != "" {
			s.SetState(StateWorking)
			s.emitAssistantMessage(text)
		}

	case "user":
		// Tool result echoed back — ignore
	case "result":
		resultText, _ := raw["result"].(string)
		costUSD, _ := raw["total_cost_usd"].(float64)
		durationMs, _ := raw["duration_ms"].(float64)
		data, _ := json.Marshal(map[string]any{
			"type":        "result",
			"message":     resultText,
			"cost_usd":    costUSD,
			"duration_ms": durationMs,
		})
		s.OutputChan <- &ACPMessage{
			Type:    TypeResult,
			Content: data,
		}

	case "control_request":
		// Auto-approve
		s.autoApproveTool(raw)

	case "control_cancel_request":
		log.Printf("[tmux-session] Control request cancelled for %s", s.MemberName)

	default:
		// Unknown type, pass as assistant message
		s.emitAssistantMessage(line)
	}
}

// handleJSONRPC handles JSON-RPC 2.0 responses.
func (s *TmuxSession) handleJSONRPC(raw map[string]any) {
	// Check for method (notification)
	if method, ok := raw["method"].(string); ok {
		if method != "" {
			data, _ := json.Marshal(raw)
			s.OutputChan <- &ACPMessage{
				Type:    TypeSystem,
				Content: data,
			}
		}
	}
}

// autoApproveTool responds to Claude's tool permission requests.
func (s *TmuxSession) autoApproveTool(raw map[string]any) {
	requestID, _ := raw["request_id"].(string)
	if requestID == "" {
		return
	}

	resp := map[string]any{
		"type": "control_response",
		"response": map[string]any{
			"subtype":    "success",
			"request_id": requestID,
			"response": map[string]any{
				"behavior": "allow",
			},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[tmux-session] Failed to marshal auto-approve: %v", err)
		return
	}

	// Send the approval back via tmux send-keys
	if err := s.SendInput(string(data)); err != nil {
		log.Printf("[tmux-session] Failed to send auto-approve: %v", err)
	}
}

// emitAssistantMessage creates and sends an assistant ACP message.
func (s *TmuxSession) emitAssistantMessage(text string) {
	data, _ := MarshalACPContent("assistant_message", text)
	s.OutputChan <- &ACPMessage{
		Type:    TypeAssistantMessage,
		Content: data,
	}
}

// emitSystemMessage creates and sends a system ACP message.
func (s *TmuxSession) emitSystemMessage(raw map[string]any) {
	msg := ""
	if m, ok := raw["message"].(string); ok {
		msg = m
	} else if sid, ok := raw["session_id"].(string); ok {
		msg = fmt.Sprintf("session initialized: %s", sid)
	}
	data, _ := json.Marshal(map[string]any{
		"type":    "system",
		"message": msg,
		"level":   "info",
	})
	s.OutputChan <- &ACPMessage{
		Type:    TypeSystem,
		Content: data,
	}
}

// SetState updates the session state.
func (s *TmuxSession) SetState(state SessionState) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.state != state {
		s.state = state
	}
}

// GetState returns the current session state.
func (s *TmuxSession) GetState() SessionState {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	return s.state
}

// Stop closes the session without killing the tmux session.
func (s *TmuxSession) Stop() error {
	s.mu.Lock()
	if s.done {
		s.mu.Unlock()
		return nil
	}
	s.done = true
	s.mu.Unlock()

	// Stop log reader
	if s.logReaderCtx != nil {
		s.logReaderCtx()
	}
	// Wait for reader to finish (with timeout)
	select {
	case <-s.logReaderDone:
	case <-time.After(2 * time.Second):
	}

	// Close done channel
	s.SetState(StateOffline)
	select {
	case <-s.DoneChan:
		// already closed
	default:
		close(s.DoneChan)
	}
	return nil
}

// Kill stops the session and kills the tmux session.
func (s *TmuxSession) Kill() error {
	if err := s.Stop(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mgr := NewManager("")
	return mgr.KillSession(ctx, s.Name)
}

// IsAlive returns true if the tmux session still exists and session not closed.
func (s *TmuxSession) IsAlive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done {
		return false
	}
	mgr := NewManager("")
	return mgr.SessionExists(s.Name)
}

// Chat bridge methods

// SetStreamSpanID sets the span ID for streaming chat.
func (s *TmuxSession) SetStreamSpanID(id string) {
	s.streamSpanMu.Lock()
	defer s.streamSpanMu.Unlock()
	s.streamSpanID = id
}

// StreamSpanID returns the current stream span ID.
func (s *TmuxSession) StreamSpanID() string {
	s.streamSpanMu.Lock()
	defer s.streamSpanMu.Unlock()
	return s.streamSpanID
}

// NextStreamSeq returns and increments the stream sequence number.
func (s *TmuxSession) NextStreamSeq() uint64 {
	s.streamSpanMu.Lock()
	defer s.streamSpanMu.Unlock()
	s.streamSeq++
	return s.streamSeq
}

// TrySendChatStream sends data to the chat stream sink (no-op placeholder).
func (s *TmuxSession) TrySendChatStream(data []byte) {
	// Actual chat streaming is handled by AgentBridge using OutputChan
	_ = data
}

// SessionInterface implementation for tool handler compatibility

// LastChatTargetConversation returns the conversation ID bound for chat bridging.
func (s *TmuxSession) LastChatTargetConversation() string {
	return ""
}

// SetLastChatTargetConversation binds a conversation ID for chat bridging.
func (s *TmuxSession) SetLastChatTargetConversation(convID string) {
	// Not needed in tmux mode — AgentBridge uses OutputChan directly
	_ = convID
}

// GetWorkspaceID returns the workspace ID.
func (s *TmuxSession) GetWorkspaceID() string {
	return s.WorkspaceID
}

// GetMemberID returns the member ID.
func (s *TmuxSession) GetMemberID() string {
	return s.MemberID
}

// GetMemberName returns the member name.
func (s *TmuxSession) GetMemberName() string {
	return s.MemberName
}

// Helper functions (migrated from local_runner.go)

// extractAssistantText extracts text content from an assistant message.
func extractAssistantText(raw map[string]any) string {
	message, ok := raw["message"].(map[string]any)
	if !ok {
		return ""
	}
	content, ok := message["content"].([]any)
	if !ok {
		if text, ok := message["content"].(string); ok {
			return text
		}
		return ""
	}

	var parts []string
	for _, block := range content {
		if blockMap, ok := block.(map[string]any); ok {
			switch blockMap["type"] {
			case "text":
				if text, ok := blockMap["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
	}
	return strings.Join(parts, "\n")
}

// extractToolUses extracts tool_use blocks from an assistant message.
func extractToolUses(raw map[string]any) []*ToolUseInfo {
	message, ok := raw["message"].(map[string]any)
	if !ok {
		return nil
	}
	content, ok := message["content"].([]any)
	if !ok {
		return nil
	}

	var results []*ToolUseInfo
	for _, block := range content {
		blockMap, ok := block.(map[string]any)
		if !ok {
			continue
		}
		if blockMap["type"] != "tool_use" {
			continue
		}
		name, _ := blockMap["name"].(string)
		id, _ := blockMap["id"].(string)
		input, _ := json.Marshal(blockMap["input"])
		if name != "" && id != "" {
			results = append(results, &ToolUseInfo{
				Type:      TypeToolUse,
				Name:      name,
				ToolUseID: id,
				Input:     json.RawMessage(input),
			})
		}
	}
	return results
}

// ToolUseInfo holds parsed tool use information.
type ToolUseInfo struct {
	Type      MessageType     `json:"type"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
}
