package cli

import (
	"fmt"

	"github.com/orchestra/backend/internal/cli/provider"
	"github.com/spf13/cobra"
)

func newProvidersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "List detected AI providers",
		Long:  `Detect and list AI CLI providers installed on the system (Claude Code, Gemini CLI, etc.).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			providers := provider.DetectProviders()

			if len(providers) == 0 {
				fmt.Println("No AI providers detected.")
				fmt.Println()
				fmt.Println("Supported providers:")
				fmt.Println("  Claude Code  — install from https://claude.ai/code")
				fmt.Println("  Gemini CLI   — install with 'npm install -g @anthropic-ai/gemini-cli' (or equivalent)")
				return nil
			}

			fmt.Printf("Detected %d provider(s):\n\n", len(providers))

			for _, p := range providers {
				skillsSupport := "no"
				if p.SupportsSkills {
					skillsSupport = "yes"
				}

				fmt.Printf("  %s\n", p.Name)
				fmt.Printf("    Version:       %s\n", orEmpty(p.Version))
				fmt.Printf("    Config dir:    %s\n", p.ConfigDir)
				fmt.Printf("    Skills support: %s\n", skillsSupport)
				fmt.Println()
			}

			return nil
		},
	}

	return cmd
}

func orEmpty(s string) string {
	if s == "" {
		return "(unknown)"
	}
	return s
}
