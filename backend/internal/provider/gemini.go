package provider

import (
	"context"
	"fmt"
	"os/exec"
)

// GeminiProvider implements AgentProvider for Gemini CLI.
type GeminiProvider struct {
	command string
}

// NewGeminiProvider creates a Gemini CLI provider.
func NewGeminiProvider() *GeminiProvider {
	return &GeminiProvider{command: "gemini"}
}

func (p *GeminiProvider) Name() ProviderName {
	return ProviderGemini
}

func (p *GeminiProvider) DisplayName() string {
	return "Gemini CLI"
}

func (p *GeminiProvider) IsInstalled() bool {
	_, err := exec.LookPath(p.command)
	return err == nil
}

func (p *GeminiProvider) SupportsPermissionMode() bool {
	return false
}

func (p *GeminiProvider) StartSession(ctx context.Context, opts SessionOptions) (AgentSession, error) {
	return nil, fmt.Errorf("gemini CLI does not yet support agent sessions")
}
