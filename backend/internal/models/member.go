package models

import "time"

type MemberRole string

const (
	RoleOwner      MemberRole = "owner"
	RoleAdmin      MemberRole = "admin"
	RoleSecretary  MemberRole = "secretary" // 秘书（监工/协调式监督角色）
	RoleAssistant  MemberRole = "assistant"
	RoleMember     MemberRole = "member"
)

type Member struct {
	ID                string     `json:"id"`
	WorkspaceID       string     `json:"workspaceId"`
	Name              string     `json:"name"`
	RoleType          MemberRole `json:"roleType"`
	RoleKey           string     `json:"roleKey,omitempty"`
	Avatar            string     `json:"avatar,omitempty"`
	TerminalType      string     `json:"terminalType,omitempty"`
	TerminalCommand   string     `json:"terminalCommand,omitempty"`
	TerminalPath      string     `json:"terminalPath,omitempty"`
	AutoStartTerminal bool       `json:"autoStartTerminal"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"createdAt"`

	// ACP configuration - for structured JSON communication
	ACPEnabled bool     `json:"acpEnabled"`           // Use ACP protocol instead of PTY
	ACPCommand string   `json:"acpCommand,omitempty"` // e.g., "claude", "gemini"
	ACPArgs    []string `json:"acpArgs,omitempty"`    // e.g., ["--input-format", "stream-json", "--output-format", "stream-json"]

	// A2A configuration - Agent-to-Agent Protocol
	A2AEnabled   bool    `json:"a2aEnabled"`
	A2AAgentURL  *string `json:"a2aAgentUrl,omitempty"`  // A2A agent service endpoint
	A2AAuthType  *string `json:"a2aAuthType,omitempty"`  // "none", "api_key", "oauth2"
	A2AAuthToken *string `json:"a2aAuthToken,omitempty"` // Auth token for A2A requests
}

type MemberCreate struct {
	Name            string     `json:"name" binding:"required"`
	RoleType        MemberRole `json:"roleType" binding:"required"`
	TerminalType    string     `json:"terminalType,omitempty"`
	TerminalCommand string     `json:"terminalCommand,omitempty"`
	// ACP configuration
	ACPEnabled bool     `json:"acpEnabled"`
	ACPCommand string   `json:"acpCommand,omitempty"`
	ACPArgs    []string `json:"acpArgs,omitempty"`

	// A2A configuration
	A2AEnabled   bool   `json:"a2aEnabled"`
	A2AAgentURL  string `json:"a2aAgentUrl,omitempty"`
	A2AAuthType  string `json:"a2aAuthType,omitempty"`
	A2AAuthToken string `json:"a2aAuthToken,omitempty"`
}

// Presence represents a member's real-time activity status
type Presence struct {
	MemberID      string `json:"memberId"`
	WorkspaceID   string `json:"workspaceId"`
	Activity      string `json:"activity"`       // "typing", "viewing", "idle"
	TargetID      string `json:"targetId"`       // conversationId or terminalId
	TargetType    string `json:"targetType"`     // "conversation" or "terminal"
	Timestamp     int64  `json:"timestamp"`
}

// PresenceUpdate is the request body for updating presence
type PresenceUpdate struct {
	Activity   string `json:"activity"`
	TargetID   string `json:"targetId,omitempty"`
	TargetType string `json:"targetType,omitempty"`
}