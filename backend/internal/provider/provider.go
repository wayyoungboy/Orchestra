// Package provider defines the AgentProvider interface for abstracting AI CLI runtimes.
package provider

import (
	"context"
)

// ProviderName identifies a supported AI provider.
type ProviderName string

const (
	ProviderClaude  ProviderName = "claude"
	ProviderGemini  ProviderName = "gemini"
	ProviderCodex   ProviderName = "codex"
	ProviderOpenCode ProviderName = "opencode"
	ProviderAider   ProviderName = "aider"
	ProviderCustom  ProviderName = "custom"
)

// SessionOptions contains parameters for starting a new agent session.
type SessionOptions struct {
	WorkspacePath string
	MemberID      string
	SessionID     string
	PermissionMode string // e.g., "default", "acceptEdits", "bypassPermissions"
	ModelSettings  *ModelSettings
}

// ModelSettings holds optional model configuration.
type ModelSettings struct {
	Model         string
	MaxTokens     int
	Temperature   float64
	MaxThinking   int
}

// AgentProvider is the abstraction interface for AI CLI runtimes.
// Each provider (Claude Code, Gemini CLI, etc.) implements this interface.
type AgentProvider interface {
	// Name returns the provider identifier (e.g., "claude").
	Name() ProviderName

	// DisplayName returns a human-readable name.
	DisplayName() string

	// IsInstalled checks whether the provider CLI is available on the system.
	IsInstalled() bool

	// SupportsPermissionMode reports whether the provider supports interactive permission modes.
	SupportsPermissionMode() bool

	// StartSession launches a new agent session.
	StartSession(ctx context.Context, opts SessionOptions) (AgentSession, error)
}

// AgentSession represents a running agent session.
type AgentSession interface {
	// Messages returns a channel that receives agent output messages.
	Messages() <-chan AgentMessage

	// Send sends a user message to the agent.
	Send(msg UserMessage) error

	// Abort terminates the session.
	Abort() error

	// IsAlive reports whether the underlying process is still running.
	IsAlive() bool

	// PID returns the process ID, if applicable.
	PID() int
}

// AgentMessage is a message from the agent to the client.
type AgentMessage struct {
	Type    string // "text", "tool_use", "tool_result", "status", "error"
	Content string
	Meta    map[string]any
}

// UserMessage is a message from the user to the agent.
type UserMessage struct {
	Text        string
	Attachments []string
}
