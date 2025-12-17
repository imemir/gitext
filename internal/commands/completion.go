package commands

import (
	"github.com/spf13/cobra"
)

func NewCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for gitext.

To load completions:

Bash:
  $ source <(gitext completion bash)

  # To load completions for each session, add the above to your ~/.bashrc or ~/.bash_profile

Zsh:
  $ source <(gitext completion zsh)

  # To load completions for each session, add the above to your ~/.zshrc

Fish:
  $ gitext completion fish | source

  # To load completions for each session, add the above to your ~/.config/fish/config.fish

PowerShell:
  PS> gitext completion powershell | Out-String | Invoke-Expression

  # To load completions for each session, add the above to your PowerShell profile`,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]
			rootCmd := cmd.Root()

			switch shell {
			case "bash":
				return rootCmd.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return rootCmd.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return rootCmd.GenPowerShellCompletion(cmd.OutOrStdout())
			default:
				return cmd.Help()
			}
		},
	}

	return cmd
}
