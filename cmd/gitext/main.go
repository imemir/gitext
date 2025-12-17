package main

import (
	"fmt"
	"os"

	"github.com/gitext/gitext/internal/commands"
	"github.com/spf13/cobra"
)

var (
	dryRun  bool
	verbose bool
)

// Version and BuildTime are set during build via ldflags
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gitext",
		Short: "Safe git workflow automation for engineering teams",
		Long: fmt.Sprintf(`gitext is a CLI tool that replaces manual git workflow steps with safe,
repeatable commands. It enforces branch protection rules and prevents
accidental production contamination.

Version: %s
Build Time: %s`, Version, BuildTime),
	}

	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without executing")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Show detailed git command output")

	// Add subcommands
	commands.AddCommands(rootCmd, &commands.Options{
		DryRun:  dryRun,
		Verbose: verbose,
		Version: Version,
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

