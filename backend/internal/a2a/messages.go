package a2a

import (
	"encoding/json"
)

// ACPTerminalResponse is the WebSocket response format for agent sessions.
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
	Type       MessageType `json:"type"`
	Message    string      `json:"message"`
	CostUSD    float64     `json:"cost_usd"`
	DurationMs int         `json:"duration_ms"`
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

// NewUserMessage creates a user message.
func NewUserMessage(content string) *ACPMessage {
	data, _ := marshalACPContent("user_message", content)
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
