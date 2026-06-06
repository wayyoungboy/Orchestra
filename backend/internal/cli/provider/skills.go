package provider

import (
	"fmt"
	"os"
	"path/filepath"
)

func installSkillSymlink(skillsDir, orchestraSkillPath, skillName string) error {
	skillMd := filepath.Join(orchestraSkillPath, "SKILL.md")
	if _, err := os.Stat(skillMd); err != nil {
		return fmt.Errorf("invalid skill: SKILL.md not found in %s", orchestraSkillPath)
	}

	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	linkPath := filepath.Join(skillsDir, skillName)
	if info, err := os.Lstat(linkPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(linkPath); err != nil {
				return fmt.Errorf("failed to remove existing symlink: %w", err)
			}
		} else {
			return fmt.Errorf("a non-symlink file/directory already exists at %s", linkPath)
		}
	}

	if err := os.Symlink(orchestraSkillPath, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

func uninstallSkillSymlink(skillsDir, skillName string) error {
	linkPath := filepath.Join(skillsDir, skillName)

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

func isSkillSymlinkInstalled(skillsDir, skillName string) bool {
	linkPath := filepath.Join(skillsDir, skillName)
	info, err := os.Lstat(linkPath)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}
