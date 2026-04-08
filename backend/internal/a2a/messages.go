package a2a

import (
	"encoding/json"
)

// ACPTerminalResponse is the WebSocket response format for agent sessions.
// Defined here so ConvertACPToWS can return it without depending on ws package.
type ACPTerminalResponse struct {
	Type string `json:"type"`

	// Connected
	SessionID string `json:"sessionId,omitempty"`

	// Assistant message
	Content string `json:"content,omitempty"`

	// Tool use notification
	ToolName  string          `json:"tool_name,omitempty"`
	ToolInput json.RawMessage `json:"tool_input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`

	// Result
	Message    string  `json:"message,omitempty"`
	CostUSD    float64 `json:"cost_usd,omitempty"`
	DurationMs int     `json:"duration_ms,omitempty"`

	// Status
	Status string `json:"status,omitempty"`

	// Error
	Error string `json:"error,omitempty"`
	Code  int    `json:"code,omitempty"`
}

// ACP message parsing helpers for backward compatibility with ACPBridge and WebSocket handlers.

// UserMessage represents a user message sent to an agent.
type UserMessage struct {
	Type    MessageType `json:"type"`
	Content string      `json:"content"`
}

// AssistantMessage represents a message from an agent.
type AssistantMessage struct {
	Type    MessageType `json:"type"`
	Content string      `json:"content"`
}

// ToolUseMessage represents a tool invocation request.
type ToolUseMessage struct {
	Type      MessageType     `json:"type"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
}

// ToolResultMessage represents a tool execution result.
type ToolResultMessage struct {
	Type      MessageType `json:"type"`
	ToolUseID string      `json:"tool_use_id"`
	Content   string      `json:"content"`
	IsError   bool        `json:"is_error"`
}

// ResultMessage represents a session completion result.
type ResultMessage struct {
	Type      MessageType `json:"type"`
	Message   string      `json:"message"`
	CostUSD   float64     `json:"cost_usd"`
	DurationMs int        `json:"duration_ms"`
}

// ErrorMessage represents an error from an agent.
type ErrorMessage struct {
	Type  MessageType `json:"type"`
	Error string      `json:"error"`
}

// SystemMessage represents a system-level notification.
type SystemMessage struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
	Level   string      `json:"level"`
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

// ParseToolUseMessage parses an ACPMessage as a tool use message.
func (m *ACPMessage) ParseToolUseMessage() (*ToolUseMessage, error) {
	var msg ToolUseMessage
	if err := json.Unmarshal(m.Content, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// NewUserMessage creates a user message.
func NewUserMessage(content string) *ACPMessage {
	data, _ := json.Marshal(UserMessage{
		Type:    TypeUserMessage,
		Content: content,
	})
	return &ACPMessage{
		Type:    TypeUserMessage,
		Content: data,
	}
}

// NewToolResult creates a tool result message.
func NewToolResult(toolUseID, content string, isError bool) *ACPMessage {
	data, _ := json.Marshal(ToolResultMessage{
		Type:      TypeToolResult,
		ToolUseID: toolUseID,
		Content:   content,
		IsError:   isError,
	})
	return &ACPMessage{
		Type:    TypeToolResult,
		Content: data,
	}
}
