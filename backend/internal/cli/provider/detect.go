package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Provider represents a detected AI CLI provider.
type Provider struct {
	Name             string
	Installed        bool
	Version          string
	ConfigDir        string
	SkillsDir        string
	SupportsSkills   bool
	SkillsInstallCmd string
}

// DetectProviders checks the system for installed AI CLI tools.
func DetectProviders() []*Provider {
	var providers []*Provider

	// Detect Claude Code
	if claude := detectClaude(); claude != nil {
		providers = append(providers, claude)
	}

	// Detect Gemini CLI
	if gemini := detectGemini(); gemini != nil {
		providers = append(providers, gemini)
	}

	return providers
}

// GetProvider returns a provider by name.
func GetProvider(name string) *Provider {
	for _, p := range DetectProviders() {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

func runVersion(cmd string, args ...string) string {
	c := exec.Command(cmd, args...)
	out, err := c.CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
