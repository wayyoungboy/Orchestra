package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root CLI command for Orchestra.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "orchestra",
		Short: "Orchestra multi-agent collaboration platform",
		Long:  "Orchestra is a web-based multi-agent collaboration system. This CLI manages local skills and provider configuration.",
	}

	rootCmd.AddCommand(
		newSkillsCmd(),
		newProvidersCmd(),
	)

	return rootCmd
}
