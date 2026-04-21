package provider

import (
	"fmt"
	"os"
	"path/filepath"
)

func detectGemini() *Provider {
	home, _ := os.UserHomeDir()
	if home == "" {
		return nil
	}

	configDir := filepath.Join(home, ".gemini")

	if !dirExists(configDir) {
		return nil
	}

	version := runVersion("gemini", "--version")

	return &Provider{
		Name:           "gemini",
		Installed:      true,
		Version:        version,
		ConfigDir:      configDir,
		SkillsDir:      "",
		SupportsSkills: false,
	}
}

// GeminiInstallSkill is a no-op since Gemini CLI doesn't support skills.
func GeminiInstallSkill(orchestraSkillPath, skillName string) error {
	p := GetProvider("gemini")
	if p == nil || !p.Installed {
		return fmt.Errorf("Gemini CLI not installed")
	}

	return fmt.Errorf("Gemini CLI does not support skills — skipping %q", skillName)
}

// GeminiUninstallSkill is a no-op since Gemini CLI doesn't support skills.
func GeminiUninstallSkill(skillName string) error {
	p := GetProvider("gemini")
	if p == nil || !p.Installed {
		return fmt.Errorf("Gemini CLI not installed")
	}

	return fmt.Errorf("Gemini CLI does not support skills — skipping %q", skillName)
}

// GeminiIsSkillInstalled always returns false.
func GeminiIsSkillInstalled(skillName string) bool {
	return false
}
