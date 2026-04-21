package cli

import (
	"github.com/spf13/cobra"
)

func newSkillsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Manage AI assistant skills",
		Long:  "Install, uninstall, and list skills for detected AI providers (Claude Code, Gemini CLI, etc.).",
	}

	cmd.AddCommand(
		newSkillsInstallCmd(),
		newSkillsUninstallCmd(),
		newSkillsListCmd(),
	)

	return cmd
}
