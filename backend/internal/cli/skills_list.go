package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/orchestra/backend/internal/cli/provider"
	"github.com/spf13/cobra"
)

func newSkillsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available skills and their installation status",
		Long:  `List all skills in ~/.orchestra/skills/ and show whether they are installed on detected providers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			skillsDir := getOrchestraSkillsDir()
			providers := provider.DetectProviders()

			fmt.Println("Available skills:")
			fmt.Println()

			entries, err := os.ReadDir(skillsDir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Printf("  No skills directory found at %s\n", skillsDir)
					fmt.Println("  Create it and add SKILL.md files, then run 'orchestra skills install <name>'")
					return nil
				}
				return fmt.Errorf("failed to read skills directory: %w", err)
			}

			if len(entries) == 0 {
				fmt.Println("  No skills found in", skillsDir)
				return nil
			}

			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}

				skillName := entry.Name()
				// Check if SKILL.md exists
				skillMd := filepath.Join(skillsDir, skillName, "SKILL.md")
				valid := ""
				if _, err := os.Stat(skillMd); err != nil {
					valid = " (no SKILL.md)"
				}

				fmt.Printf("  %s%s\n", skillName, valid)

				for _, p := range providers {
					installed := false
					if p.SupportsSkills {
						installed = provider.ClaudeIsSkillInstalled(skillName)
					}

					status := "not installed"
					if installed {
						status = "installed"
					}
					fmt.Printf("    %s: %s\n", p.Name, status)
				}
				fmt.Println()
			}

			return nil
		},
	}

	return cmd
}
