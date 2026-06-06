package provider

import (
	"fmt"
	"os"
	"path/filepath"
)

func detectCodex() *Provider {
	home, _ := os.UserHomeDir()
	if home == "" {
		return nil
	}

	configDir := filepath.Join(home, ".codex")
	skillsDir := filepath.Join(configDir, "skills")

	if !dirExists(configDir) {
		return nil
	}

	version := runVersion("codex", "--version")

	return &Provider{
		Name:             "codex",
		Installed:        true,
		Version:          version,
		ConfigDir:        configDir,
		SkillsDir:        skillsDir,
		SupportsSkills:   true,
		SkillsInstallCmd: "symlink",
	}
}

func CodexInstallSkill(orchestraSkillPath, skillName string) error {
	p := GetProvider("codex")
	if p == nil || !p.Installed {
		return fmt.Errorf("Codex not installed")
	}
	return installSkillSymlink(p.SkillsDir, orchestraSkillPath, skillName)
}

func CodexUninstallSkill(skillName string) error {
	p := GetProvider("codex")
	if p == nil || !p.Installed {
		return fmt.Errorf("Codex not installed")
	}
	return uninstallSkillSymlink(p.SkillsDir, skillName)
}

func CodexIsSkillInstalled(skillName string) bool {
	p := GetProvider("codex")
	if p == nil || !p.Installed {
		return false
	}
	return isSkillSymlinkInstalled(p.SkillsDir, skillName)
}
