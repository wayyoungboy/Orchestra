package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/orchestra/backend/internal/cli/provider"
	"github.com/spf13/cobra"
)

func newSkillsInstallCmd() *cobra.Command {
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "install [skill-name]",
		Short: "Install a skill to all detected AI providers",
		Long: `Install a skill to all detected AI providers (Claude Code, Gemini CLI, etc.).

The skill must exist in ~/.orchestra/skills/<skill-name>/ with a SKILL.md file.
Use --all to install all available skills.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestraSkillsDir := getOrchestraSkillsDir()

			if allFlag {
				return installAllSkills(orchestraSkillsDir)
			}

			if len(args) == 0 {
				return fmt.Errorf("please specify a skill name, or use --all to install all skills")
			}

			skillName := args[0]
			skillPath := filepath.Join(orchestraSkillsDir, skillName)

			if _, err := os.Stat(skillPath); err != nil {
				return fmt.Errorf("skill %q not found in %s", skillName, orchestraSkillsDir)
			}

			return installSingleSkill(skillPath, skillName)
		},
	}

	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Install all available skills")

	return cmd
}

func installAllSkills(skillsDir string) error {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no skills directory found at %s — create it and add SKILL.md files first", skillsDir)
		}
		return fmt.Errorf("failed to read skills directory: %w", err)
	}

	installed := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		skillPath := filepath.Join(skillsDir, skillName)

		if err := installSingleSkill(skillPath, skillName); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", skillName, err)
		} else {
			installed++
		}
	}

	if installed == 0 {
		fmt.Println("No skills were installed.")
	} else {
		fmt.Printf("\nInstalled %d skill(s).\n", installed)
	}

	return nil
}

func installSingleSkill(skillPath, skillName string) error {
	providers := provider.DetectProviders()
	if len(providers) == 0 {
		return fmt.Errorf("no AI providers detected — run 'orchestra providers' to check")
	}

	fmt.Printf("Installing skill: %s\n", skillName)

	for _, p := range providers {
		if !p.SupportsSkills {
			fmt.Printf("  ⚠ %s: does not support skills (skipped)\n", p.Name)
			continue
		}

		if err := provider.ClaudeInstallSkill(skillPath, skillName); err != nil {
			fmt.Printf("  ✗ %s: %v\n", p.Name, err)
		} else {
			fmt.Printf("  ✓ %s: installed\n", p.Name)
		}
	}

	return nil
}
