package provider

import (
	"os"
	"path/filepath"
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
	reg.Register(NewCodexProvider(""))

	all := reg.List()
	if len(all) != 3 {
		t.Errorf("expected 3 providers, got %d", len(all))
	}
}

func TestRegistry_Installed(t *testing.T) {
	binDir := t.TempDir()
	claudePath := filepath.Join(binDir, "claude")
	if err := os.WriteFile(claudePath, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatalf("write fake claude executable: %v", err)
	}
	t.Setenv("PATH", binDir)

	reg := NewRegistry()
	reg.Register(NewClaudeProvider("claude"))
	reg.Register(NewGeminiProvider())

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

func TestCodexProvider_Name(t *testing.T) {
	p := NewCodexProvider("")
	if p.Name() != ProviderCodex {
		t.Errorf("expected %q, got %q", ProviderCodex, p.Name())
	}
}

func TestCodexProvider_DisplayName(t *testing.T) {
	p := NewCodexProvider("")
	if p.DisplayName() != "OpenAI Codex" {
		t.Errorf("expected %q, got %q", "OpenAI Codex", p.DisplayName())
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

func TestCodexProvider_CustomCommand(t *testing.T) {
	p := NewCodexProvider("/custom/path/codex")
	if p.Name() != ProviderCodex {
		t.Errorf("expected %q, got %q", ProviderCodex, p.Name())
	}
	if p.IsInstalled() {
		t.Error("custom command should not be installed")
	}
}
