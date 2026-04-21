package cli

import (
	"fmt"

	"github.com/orchestra/backend/internal/cli/provider"
	"github.com/spf13/cobra"
)

func newSkillsUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall [skill-name]",
		Short: "Uninstall a skill from all detected AI providers",
		Long: `Remove a skill symlink from all detected AI providers.

The skill must have been previously installed via 'orchestra skills install'.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName := args[0]

			providers := provider.DetectProviders()
			if len(providers) == 0 {
				return fmt.Errorf("no AI providers detected")
			}

			uninstalled := 0
			for _, p := range providers {
				if !p.SupportsSkills {
					continue
				}

				if err := provider.ClaudeUninstallSkill(skillName); err != nil {
					fmt.Printf("  ✗ %s: %v\n", p.Name, err)
				} else {
					fmt.Printf("  ✓ %s: uninstalled\n", p.Name)
					uninstalled++
				}
			}

			if uninstalled == 0 {
				fmt.Printf("Skill %q was not installed on any provider.\n", skillName)
			}

			return nil
		},
	}

	return cmd
}
