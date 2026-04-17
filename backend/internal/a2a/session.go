// Package a2a provides A2A (Agent-to-Agent) protocol integration for Orchestra.
// It replaces the ACP (stdin/stdout) approach with HTTP-based agent communication.
package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2aclient"
)

// agentSender is the interface implemented by LocalRunner (and test doubles)
// for sending messages to the agent process.
type agentSender interface {
	SendUserMessage(text string) error
	SendToolResult(toolUseID, content string, isError bool) error
	Stop()
	IsRunning() bool
}

// Session represents an agent session (A2A HTTP or Local CLI).
// It implements the same interface as the old ACP session for ACPBridge compatibility.
type Session struct {
	ID           string
	WorkspaceID  string
	MemberID     string
	MemberName   string
	TerminalType string

	// A2A (HTTP-based) fields
	Client   *a2aclient.Client
	AgentURL string

	// Local CLI runner (PTY alternative)
	localRunner agentSender

	mu               sync.Mutex
	lastActive       time.Time
	createdAt        time.Time
	lastChatConvID   string
	chatStreamMu     sync.Mutex
	chatStream       chan<- []byte
	streamSpanMu     sync.Mutex
	streamSpanID     string
	streamSeq        uint64

	// Output channels for ACPBridge / WebSocket relay compatibility
	OutputChan chan *ACPMessage
	ErrorChan  chan error
	DoneChan   chan struct{}
	done       bool
	released   bool

	// Active subscriptions
	subscriptions map[string]context.CancelFunc // taskID -> cancel
	subMu         sync.Mutex

	// Pending tool use tracking (for correlation when tool results arrive)
	pendingToolUses map[string]string // toolUseID -> taskID that initiated it
	toolUseMu       sync.Mutex
}

// ACPMessage mirrors the old ACP message type for backward compatibility.
type ACPMessage struct {
	Type    MessageType
	Content json.RawMessage
}

// marshalACPContent creates ACP content JSON for the given type and text.
func marshalACPContent(typ, text string) (json.RawMessage, error) {
	return json.Marshal(map[string]string{
		"type":    typ,
		"content": text,
	})
}

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

// NewSession creates a new session for a workspace member.
func NewSession(id, workspaceID, memberID, memberName, terminalType string, client *a2aclient.Client, agentURL string) *Session {
	return &Session{
		ID:            id,
		WorkspaceID:   workspaceID,
		MemberID:      memberID,
		MemberName:    memberName,
		TerminalType:  terminalType,
		Client:        client,
		AgentURL:      agentURL,
		createdAt:     time.Now(),
		lastActive:    time.Now(),
		OutputChan:    make(chan *ACPMessage, 256),
		ErrorChan:     make(chan error, 16),
		DoneChan:      make(chan struct{}),
		subscriptions: make(map[string]context.CancelFunc),
		pendingToolUses: make(map[string]string),
	}
}

// SendUserMessage sends a user message to the agent.
// Supports both A2A HTTP-based and Local CLI agents.
func (s *Session) SendUserMessage(content string) error {
	s.mu.Lock()
	if s.released {
		s.mu.Unlock()
		return fmt.Errorf("session already released")
	}
	s.lastActive = time.Now()
	s.mu.Unlock()

	// Local CLI runner path
	if s.localRunner != nil {
		return s.localRunner.SendUserMessage(content)
	}

	// A2A HTTP path
	if s.Client == nil {
		return fmt.Errorf("no agent configured for session %s", s.ID)
	}

	msg := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart(content))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.Client.SendMessage(ctx, &a2a.SendMessageRequest{
		Message: msg,
	})
	if err != nil {
		return err
	}

	// Process the response: if it's a Task, subscribe to SSE for updates
	if task, ok := resp.(*a2a.Task); ok {
		s.subscribeToTask(string(task.ID))
	}

	// If it's a direct message response, convert to ACP message
	if directMsg, ok := resp.(a2a.SendMessageResult); ok {
		if msg, ok := directMsg.(*a2a.Message); ok {
			acpMsg := convertA2AMessageToACP(msg)
			if acpMsg != nil {
				s.OutputChan <- acpMsg
			}
		}
	}

	return nil
}

// subscribeToTask subscribes to SSE events for a task and converts them to ACP messages.
// It implements exponential backoff reconnection on SSE stream failure with a maximum retry limit.
func (s *Session) subscribeToTask(taskID string) {
	ctx, cancel := context.WithCancel(context.Background())

	s.subMu.Lock()
	s.subscriptions[taskID] = cancel
	s.subMu.Unlock()

	go func() {
		defer func() {
			s.subMu.Lock()
			delete(s.subscriptions, taskID)
			s.subMu.Unlock()
		}()

		backoff := 1 * time.Second
		const maxBackoff = 30 * time.Second
		retryCount := 0
		const maxRetries = 12

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.DoneChan:
				return
			default:
			}

			reconnected := false
			for event, err := range s.Client.SubscribeToTask(ctx, &a2a.SubscribeToTaskRequest{
				ID: a2a.TaskID(taskID),
			}) {
				if err != nil {
					s.ErrorChan <- err
					reconnected = true
					continue
				}

				// Reset backoff and retry count on successful reception
				backoff = 1 * time.Second
				retryCount = 0

				acpMsg := s.convertA2AEventToACP(event)
				if acpMsg != nil {
					s.OutputChan <- acpMsg
				}
			}

			// SSE stream ended or errored — check if we need to recover
			if !reconnected {
				// Stream completed normally, check task status
				return
			}

			// Check retry limit
			if retryCount >= maxRetries {
				log.Printf("[a2a] Task %s subscription exceeded max retries (%d)", taskID, maxRetries)
				return
			}
			retryCount++

			// Exponential backoff before reconnect
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return
			case <-s.DoneChan:
				return
			}

			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			// Verify task still exists before reconnecting
			getCtx, getCancel := context.WithTimeout(context.Background(), 5*time.Second)
			task, err := s.Client.GetTask(getCtx, &a2a.GetTaskRequest{
				ID: a2a.TaskID(taskID),
			})
			getCancel()
			if err != nil {
				log.Printf("[a2a] Task %s not recoverable: %v", taskID, err)
				return
			}
			if task != nil {
				switch task.Status.State {
				case a2a.TaskStateCompleted, a2a.TaskStateFailed, a2a.TaskStateCanceled:
					return // Terminal state, no need to reconnect
				}
			}
		}
	}()
}

// SendToolResult sends a tool result back to the A2A agent.
func (s *Session) SendToolResult(toolUseID, content string, isError bool) error {
	s.mu.Lock()
	if s.released {
		s.mu.Unlock()
		return fmt.Errorf("session already released")
	}
	s.lastActive = time.Now()
	s.mu.Unlock()

	// Build a structured tool result message for the A2A agent
	resultContent := map[string]any{
		"type":      "tool_result",
		"tool_use_id": toolUseID,
		"content":   content,
		"is_error":  isError,
	}
	rawContent, _ := json.Marshal(resultContent)

	msg := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart(string(rawContent)))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.Client.SendMessage(ctx, &a2a.SendMessageRequest{
		Message: msg,
	})
	if err != nil {
		return err
	}

	// Clean up pending tool use tracking
	s.toolUseMu.Lock()
	delete(s.pendingToolUses, toolUseID)
	s.toolUseMu.Unlock()

	if task, ok := resp.(*a2a.Task); ok {
		s.subscribeToTask(string(task.ID))
	}

	return nil
}

// SendToolResultToAgent sends a tool result back to the agent via the correct transport.
// For local CLI agents, it sends in Claude's stream-json format.
// For A2A agents, it uses the A2A protocol.
func (s *Session) SendToolResultToAgent(toolUseID, content string, isError bool) error {
	// Local CLI runner path - use Claude's native format
	if s.localRunner != nil {
		return s.localRunner.SendToolResult(toolUseID, content, isError)
	}

	// A2A HTTP path - use existing method
	return s.SendToolResult(toolUseID, content, isError)
}

// IsAlive returns true if the session is still active.
func (s *Session) IsAlive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.localRunner != nil {
		return s.localRunner.IsRunning()
	}
	return s.Client != nil
}

// Release closes the session and all active subscriptions.
func (s *Session) Release() {
	s.subMu.Lock()
	for _, cancel := range s.subscriptions {
		cancel()
	}
	s.subscriptions = make(map[string]context.CancelFunc)
	s.subMu.Unlock()

	// Stop local runner if present
	if s.localRunner != nil {
		s.localRunner.Stop()
	}

	s.mu.Lock()
	s.released = true
	if !s.done {
		s.done = true
		close(s.DoneChan)
	}
	s.mu.Unlock()
}

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

// SetChatStreamSink sets the channel for streaming chat data.
func (s *Session) SetChatStreamSink(ch chan<- []byte) {
	s.chatStreamMu.Lock()
	defer s.chatStreamMu.Unlock()
	s.chatStream = ch
}

// TrySendChatStream sends data to the chat stream sink if configured.
// Attempts to send with a 5-second timeout to avoid blocking indefinitely.
func (s *Session) TrySendChatStream(data []byte) {
	s.chatStreamMu.Lock()
	ch := s.chatStream
	s.chatStreamMu.Unlock()

	if ch != nil {
		select {
		case ch <- data:
		case <-time.After(5 * time.Second):
			log.Printf("[a2a] WARN: output channel full for session %s, message dropped (type=%s)", s.ID, "chat_stream")
		}
	}
}

// convertA2AMessageToACP converts an A2A Message to ACP format.
func convertA2AMessageToACP(msg *a2a.Message) *ACPMessage {
	if msg == nil {
		return nil
	}
	content := extractTextFromParts(msg.Parts)
	if content == "" {
		return nil
	}
	data, _ := json.Marshal(map[string]string{
		"type":    "assistant_message",
		"content": content,
	})
	return &ACPMessage{
		Type:    TypeAssistantMessage,
		Content: data,
	}
}

// convertA2AEventToACP converts an A2A Event to ACP format.
func (s *Session) convertA2AEventToACP(event a2a.Event) *ACPMessage {
	if event == nil {
		return nil
	}

	switch e := event.(type) {
	case *a2a.TaskArtifactUpdateEvent:
		// Check for tool_use parts first
		if toolUses := extractToolUsesFromParts(e.Artifact.Parts); len(toolUses) > 0 {
			for _, tu := range toolUses {
				// Track this tool use for correlation
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
		content := extractTextFromParts(e.Artifact.Parts)
		if content == "" {
			return nil
		}
		data, _ := json.Marshal(map[string]string{
			"type":    "assistant_message",
			"content": content,
		})
		return &ACPMessage{
			Type:    TypeAssistantMessage,
			Content: data,
		}

	case *a2a.TaskStatusUpdateEvent:
		switch e.Status.State {
		case a2a.TaskStateCompleted:
			// Check for tool_use in the final message parts
			if e.Status.Message != nil {
				if toolUses := extractToolUsesFromParts(e.Status.Message.Parts); len(toolUses) > 0 {
					for _, tu := range toolUses {
						data, _ := json.Marshal(tu)
						s.OutputChan <- &ACPMessage{
							Type:    TypeToolUse,
							Content: data,
						}
					}
				}
			}
			msg := ""
			if e.Status.Message != nil {
				msg = extractTextFromParts(e.Status.Message.Parts)
			}
			data, _ := json.Marshal(map[string]any{
				"type":        "result",
				"message":     msg,
				"cost_usd":    0.0,
				"duration_ms": 0,
			})
			return &ACPMessage{
				Type:    TypeResult,
				Content: data,
			}

		case a2a.TaskStateFailed, a2a.TaskStateCanceled:
			msg := ""
			if e.Status.Message != nil {
				msg = extractTextFromParts(e.Status.Message.Parts)
			}
			data, _ := json.Marshal(map[string]string{
				"type":  "error",
				"error": msg,
			})
			return &ACPMessage{
				Type:    TypeError,
				Content: data,
			}

		case a2a.TaskStateInputRequired:
			// Signal that the agent needs more input
			msg := ""
			if e.Status.Message != nil {
				msg = extractTextFromParts(e.Status.Message.Parts)
			}
			data, _ := json.Marshal(map[string]string{
				"type":    "system",
				"message": msg,
				"level":   "info",
			})
			return &ACPMessage{
				Type:    TypeSystem,
				Content: data,
			}

		default:
			// working, submitted, etc. - internal states, skip
			return nil
		}

	default:
		// Unknown event type, skip
		return nil
	}
}

// convertRawEventToACP converts a raw map[string]any event (e.g., from LocalRunner's
// Claude stream-json output) into an ACPMessage. Returns nil for events that should
// be silently ignored (system hooks, control responses, etc.).
func convertRawEventToACP(raw map[string]any) *ACPMessage {
	typ, _ := raw["type"].(string)
	switch typ {
	case "assistant":
		// Extract text content from assistant message
		text := extractAssistantTextFromRaw(raw)
		if text == "" {
			return nil
		}
		data, _ := json.Marshal(map[string]string{
			"type":    "assistant_message",
			"content": text,
		})
		return &ACPMessage{
			Type:    TypeAssistantMessage,
			Content: data,
		}

	case "result":
		// Final result
		resultText, _ := raw["result"].(string)
		costUSD, _ := raw["total_cost_usd"].(float64)
		durationMs, _ := raw["duration_ms"].(float64)
		data, _ := json.Marshal(map[string]any{
			"type":        "result",
			"message":     resultText,
			"cost_usd":    costUSD,
			"duration_ms": durationMs,
		})
		return &ACPMessage{
			Type:    TypeResult,
			Content: data,
		}

	case "system":
		// Only log init, skip hook events
		subtype, _ := raw["subtype"].(string)
		if subtype == "init" {
			sessionID, _ := raw["session_id"].(string)
			data, _ := json.Marshal(map[string]string{
				"type":       "system",
				"session_id": sessionID,
				"level":      "info",
			})
			return &ACPMessage{
				Type:    TypeSystem,
				Content: data,
			}
		}
		return nil

	case "control_request":
		// Auto-approve already handled in LocalRunner, skip
		return nil

	default:
		return nil
	}
}

// extractTextFromParts extracts text content from A2A Part slice.
func extractTextFromParts(parts []*a2a.Part) string {
	var result string
	for _, p := range parts {
		if p == nil {
			continue
		}
		if t := p.Text(); t != "" {
			result += t
		}
	}
	return result
}

// extractToolUsesFromParts extracts tool_use calls from Part slices.
// A2A agents return tool_use as DataContent parts with structured JSON.
func extractToolUsesFromParts(parts []*a2a.Part) []*ToolUseMessage {
	var results []*ToolUseMessage
	for _, p := range parts {
		if p == nil {
			continue
		}
		// Check for DataContent parts (tool_use is typically structured data)
		if d := p.Data(); d != nil {
			if toolUse, ok := tryParseToolUse(d); ok {
				results = append(results, toolUse)
			}
		}
		// Also check Raw parts for JSON-encoded tool_use
		if raw := p.Raw(); len(raw) > 0 {
			var data any
			if json.Unmarshal(raw, &data) == nil {
				if toolUse, ok := tryParseToolUse(data); ok {
					results = append(results, toolUse)
				}
			}
		}
	}
	return results
}

// tryParseToolUse attempts to parse a Data value as a tool_use message.
func tryParseToolUse(data any) (*ToolUseMessage, bool) {
	m, ok := data.(map[string]any)
	if !ok {
		return nil, false
	}
	// Check for tool_use indicator
	if typ, _ := m["type"].(string); typ == "tool_use" {
		inputRaw, _ := json.Marshal(m["input"])
		return &ToolUseMessage{
			Type:      TypeToolUse,
			Name:      m["name"].(string),
			Input:     json.RawMessage(inputRaw),
			ToolUseID: m["id"].(string),
		}, true
	}
	// Some agents use "tool" or "function_call" as the type
	if _, hasName := m["name"]; hasName {
		if _, hasInput := m["input"]; hasInput {
			id, _ := m["id"].(string)
			if id == "" {
				id, _ = m["tool_use_id"].(string)
			}
			inputRaw, _ := json.Marshal(m["input"])
			return &ToolUseMessage{
				Type:      TypeToolUse,
				Name:      m["name"].(string),
				Input:     json.RawMessage(inputRaw),
				ToolUseID: id,
			}, true
		}
	}
	return nil, false
}

// convertACPToWS converts an ACP message to a WebSocket response.
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
		parsed, err := msg.ParseErrorMessage() // reuse error parser for system messages
		if err != nil {
			return nil
		}
		return &ACPTerminalResponse{
			Type:   "status",
			Status: parsed.Error,
		}

	default:
		return nil
	}
}
