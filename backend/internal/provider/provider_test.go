package provider

import (
	"testing"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	reg := NewRegistry()

	// Register a provider
	claude := NewClaudeProvider("claude")
	reg.Register(claude)

	// Get by name
	got := reg.Get(ProviderClaude)
	if got == nil {
		t.Fatal("expected Claude provider")
	}
	if got.Name() != ProviderClaude {
		t.Errorf("expected name %q, got %q", ProviderClaude, got.Name())
	}
	if got.DisplayName() != "Claude Code" {
		t.Errorf("expected display name %q, got %q", "Claude Code", got.DisplayName())
	}
}

func TestRegistry_List(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewClaudeProvider("claude"))
	reg.Register(NewGeminiProvider())

	all := reg.List()
	if len(all) != 2 {
		t.Errorf("expected 2 providers, got %d", len(all))
	}
}

func TestRegistry_Installed(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewClaudeProvider("claude"))
	reg.Register(NewGeminiProvider())

	// Claude is installed on this machine
	installed := reg.Installed()
	names := make(map[ProviderName]bool)
	for _, p := range installed {
		names[p.Name()] = true
	}
	if !names[ProviderClaude] {
		t.Error("expected Claude to be installed")
	}
	// Gemini may or may not be installed depending on environment
}

func TestClaudeProvider_Name(t *testing.T) {
	p := NewClaudeProvider("")
	if p.Name() != ProviderClaude {
		t.Errorf("expected %q, got %q", ProviderClaude, p.Name())
	}
}

func TestClaudeProvider_DisplayName(t *testing.T) {
	p := NewClaudeProvider("")
	if p.DisplayName() != "Claude Code" {
		t.Errorf("expected %q, got %q", "Claude Code", p.DisplayName())
	}
}

func TestClaudeProvider_IsInstalled(t *testing.T) {
	p := NewClaudeProvider("claude")
	if !p.IsInstalled() {
		t.Skip("claude not installed — skipping")
	}
}

func TestGeminiProvider_Name(t *testing.T) {
	p := NewGeminiProvider()
	if p.Name() != ProviderGemini {
		t.Errorf("expected %q, got %q", ProviderGemini, p.Name())
	}
}

func TestGeminiProvider_SupportsPermissionMode(t *testing.T) {
	p := NewGeminiProvider()
	if p.SupportsPermissionMode() {
		t.Error("gemini should not support permission mode")
	}
}

func TestClaudeProvider_SupportsPermissionMode(t *testing.T) {
	p := NewClaudeProvider("")
	if !p.SupportsPermissionMode() {
		t.Error("claude should support permission mode")
	}
}

func TestClaudeProvider_CustomCommand(t *testing.T) {
	p := NewClaudeProvider("/custom/path/claude")
	if p.Name() != ProviderClaude {
		t.Errorf("expected %q, got %q", ProviderClaude, p.Name())
	}
	// Custom command should not be installed
	if p.IsInstalled() {
		t.Error("custom command should not be installed")
	}
}
