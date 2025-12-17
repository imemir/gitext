package commands

import (
	"github.com/spf13/cobra"
)

// AddCommands adds all subcommands to the root command
func AddCommands(rootCmd *cobra.Command, opts *Options) {
	rootCmd.AddCommand(NewInitCmd(opts))
	rootCmd.AddCommand(NewStatusCmd(opts))
	rootCmd.AddCommand(NewSyncCmd(opts))
	rootCmd.AddCommand(NewStartCmd(opts))
	rootCmd.AddCommand(NewUpdateCmd(opts))
	rootCmd.AddCommand(NewRetargetCmd(opts))
	rootCmd.AddCommand(NewPrepareCmd(opts))
	rootCmd.AddCommand(NewCleanupCmd(opts))
	rootCmd.AddCommand(NewCommitCmd(opts))
	rootCmd.AddCommand(NewAICmd(opts))
	rootCmd.AddCommand(NewSelfUpdateCmd(opts))
	rootCmd.AddCommand(NewCompletionCmd())
}

