package models

import "time"

// AgentStatus represents an AI agent's current activity status
type AgentStatus struct {
	MemberID      string    `json:"memberId"`
	WorkspaceID   string    `json:"workspaceId"`
	ConversationID string   `json:"conversationId,omitempty"`
	Status        string    `json:"status"`        // "thinking", "reading_file", "writing_code", "idle", "error"
	Message       string    `json:"message,omitempty"` // Human-readable status message
	Progress      int       `json:"progress,omitempty"` // 0-100 progress percentage
	Timestamp     time.Time `json:"timestamp"`
}

// AgentStatusUpdate is the request body for updating agent status
type AgentStatusUpdate struct {
	MemberID      string `json:"memberId" binding:"required"`
	WorkspaceID   string `json:"workspaceId" binding:"required"`
	ConversationID string `json:"conversationId,omitempty"`
	Status        string `json:"status" binding:"required"`
	Message       string `json:"message,omitempty"`
	Progress      int    `json:"progress,omitempty"`
}

// Agent status constants
const (
	AgentStatusThinking     = "thinking"
	AgentStatusReadingFile  = "reading_file"
	AgentStatusWritingCode  = "writing_code"
	AgentStatusRunningTests = "running_tests"
	AgentStatusIdle         = "idle"
	AgentStatusError        = "error"
)