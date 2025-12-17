package commands

import (
	"github.com/spf13/cobra"
)

// NewAICmd creates the 'ai' command group
func NewAICmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "AI-related commands for commit message generation",
		Long: `Commands for configuring and managing AI providers for automatic
commit message generation. Supports OpenAI and OpenRouter.`,
	}

	// Add subcommands
	cmd.AddCommand(NewAISetupCmd(opts))
	cmd.AddCommand(NewAIConfigCmd(opts))

	return cmd
}
