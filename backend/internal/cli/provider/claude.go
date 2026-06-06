package provider

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func detectClaude() *Provider {
	home, _ := os.UserHomeDir()
	if home == "" {
		return nil
	}

	configDir := filepath.Join(home, ".claude")
	skillsDir := filepath.Join(configDir, "skills")

	if !dirExists(configDir) {
		return nil
	}

	version := runVersion("claude", "--version")

	// Check if binary is in PATH (informational — config dir presence is the primary signal)
	if _, err := exec.LookPath("claude"); err == nil {
		_ = version // version already captured above
	}

	return &Provider{
		Name:             "claude",
		Installed:        true,
		Version:          version,
		ConfigDir:        configDir,
		SkillsDir:        skillsDir,
		SupportsSkills:   true,
		SkillsInstallCmd: "symlink",
	}
}

// ClaudeInstallSkill creates a symlink from Orchestra's skill directory to Claude's skills directory.
func ClaudeInstallSkill(orchestraSkillPath, skillName string) error {
	p := GetProvider("claude")
	if p == nil || !p.Installed {
		return fmt.Errorf("Claude Code not installed")
	}
	return installSkillSymlink(p.SkillsDir, orchestraSkillPath, skillName)
}

// ClaudeUninstallSkill removes a skill symlink from Claude's skills directory.
func ClaudeUninstallSkill(skillName string) error {
	p := GetProvider("claude")
	if p == nil || !p.Installed {
		return fmt.Errorf("Claude Code not installed")
	}
	return uninstallSkillSymlink(p.SkillsDir, skillName)
}

// ClaudeIsSkillInstalled checks if a skill is symlinked in Claude's skills directory.
func ClaudeIsSkillInstalled(skillName string) bool {
	p := GetProvider("claude")
	if p == nil || !p.Installed {
		return false
	}
	return isSkillSymlinkInstalled(p.SkillsDir, skillName)
}
