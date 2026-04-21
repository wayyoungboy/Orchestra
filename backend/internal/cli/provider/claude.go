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

	// Validate the skill directory contains SKILL.md
	skillMd := filepath.Join(orchestraSkillPath, "SKILL.md")
	if _, err := os.Stat(skillMd); err != nil {
		return fmt.Errorf("invalid skill: SKILL.md not found in %s", orchestraSkillPath)
	}

	// Ensure Claude skills directory exists
	if err := os.MkdirAll(p.SkillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	linkPath := filepath.Join(p.SkillsDir, skillName)

	// Remove existing symlink if present
	if info, err := os.Lstat(linkPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(linkPath); err != nil {
				return fmt.Errorf("failed to remove existing symlink: %w", err)
			}
		} else {
			return fmt.Errorf("a non-symlink file/directory already exists at %s", linkPath)
		}
	}

	// Create symlink
	if err := os.Symlink(orchestraSkillPath, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// ClaudeUninstallSkill removes a skill symlink from Claude's skills directory.
func ClaudeUninstallSkill(skillName string) error {
	p := GetProvider("claude")
	if p == nil || !p.Installed {
		return fmt.Errorf("Claude Code not installed")
	}

	linkPath := filepath.Join(p.SkillsDir, skillName)

	info, err := os.Lstat(linkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill %q is not installed", skillName)
		}
		return fmt.Errorf("failed to check skill: %w", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("%q is not a symlink managed by Orchestra", skillName)
	}

	if err := os.Remove(linkPath); err != nil {
		return fmt.Errorf("failed to remove skill: %w", err)
	}

	return nil
}

// ClaudeIsSkillInstalled checks if a skill is symlinked in Claude's skills directory.
func ClaudeIsSkillInstalled(skillName string) bool {
	p := GetProvider("claude")
	if p == nil || !p.Installed {
		return false
	}

	linkPath := filepath.Join(p.SkillsDir, skillName)
	info, err := os.Lstat(linkPath)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}
