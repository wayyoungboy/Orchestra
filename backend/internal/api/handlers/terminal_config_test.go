package handlers

import (
	"testing"

	"github.com/orchestra/backend/internal/models"
)

func TestMergeSessionMemberConfigUsesStoredAgentCommand(t *testing.T) {
	stored := &models.Member{
		ID:              "assistant-1",
		WorkspaceID:     "ws-1",
		Name:            "Assistant",
		RoleType:        models.RoleAssistant,
		TerminalType:    "claude",
		TerminalCommand: "claude",
		ACPEnabled:      true,
		ACPCommand:      "claude",
		ACPArgs:         []string{"--output-format", "stream-json"},
	}

	merged := mergeSessionMemberConfig(stored, CreateSessionRequest{})

	if merged.ID != stored.ID || merged.WorkspaceID != stored.WorkspaceID {
		t.Fatalf("identity changed: %+v", merged)
	}
	if !merged.ACPEnabled {
		t.Fatal("expected stored ACP configuration to remain enabled")
	}
	if merged.ACPCommand != "claude" {
		t.Fatalf("ACPCommand = %q, want claude", merged.ACPCommand)
	}
	if len(merged.ACPArgs) != 2 || merged.ACPArgs[1] != "stream-json" {
		t.Fatalf("ACPArgs = %#v", merged.ACPArgs)
	}
	if merged.TerminalType != "claude" {
		t.Fatalf("TerminalType = %q, want claude", merged.TerminalType)
	}
}

func TestMergeSessionMemberConfigAllowsRequestOverride(t *testing.T) {
	stored := &models.Member{
		ID:              "assistant-1",
		WorkspaceID:     "ws-1",
		Name:            "Assistant",
		RoleType:        models.RoleAssistant,
		TerminalType:    "claude",
		TerminalCommand: "claude",
		ACPEnabled:      true,
		ACPCommand:      "claude",
	}

	merged := mergeSessionMemberConfig(stored, CreateSessionRequest{
		Command:      "gemini",
		Args:         []string{"--yolo"},
		TerminalType: "gemini",
		MemberName:   "Gemini",
	})

	if merged.Name != "Gemini" {
		t.Fatalf("Name = %q, want Gemini", merged.Name)
	}
	if merged.ACPCommand != "gemini" {
		t.Fatalf("ACPCommand = %q, want gemini", merged.ACPCommand)
	}
	if len(merged.ACPArgs) != 1 || merged.ACPArgs[0] != "--yolo" {
		t.Fatalf("ACPArgs = %#v", merged.ACPArgs)
	}
	if merged.TerminalType != "gemini" {
		t.Fatalf("TerminalType = %q, want gemini", merged.TerminalType)
	}
}

func TestValidateAgentCommandRejectsCommandsOutsideAllowlist(t *testing.T) {
	if err := validateAgentCommand("/bin/cat", []string{"/bin/cat", "claude"}); err != nil {
		t.Fatalf("allowed command rejected: %v", err)
	}
	if err := validateAgentCommand("/bin/missing-agent", []string{"/bin/cat", "claude"}); err == nil {
		t.Fatal("expected command outside allowlist to be rejected")
	}
}

func TestTerminalSnapshotLineCount(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want int
	}{
		{name: "default", raw: "", want: 200},
		{name: "invalid", raw: "many", want: 200},
		{name: "negative", raw: "-5", want: 200},
		{name: "valid", raw: "80", want: 80},
		{name: "capped", raw: "5000", want: 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := terminalSnapshotLineCount(tt.raw); got != tt.want {
				t.Fatalf("terminalSnapshotLineCount(%q) = %d, want %d", tt.raw, got, tt.want)
			}
		})
	}
}
