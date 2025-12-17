package commands

import (
	"fmt"

	"github.com/gitext/gitext/pkg/ai"
	"github.com/gitext/gitext/pkg/aiconfig"
	"github.com/gitext/gitext/pkg/git"
	"github.com/gitext/gitext/pkg/ui"
	"github.com/spf13/cobra"
)

func NewCommitCmd(opts *Options) *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Generate and create a commit with AI",
		Long: `Generate a commit message using AI based on staged changes and create the commit.
The message follows Conventional Commits specification.

If --message is provided, it will be used instead of generating one.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			output := ui.NewOutput(opts.Verbose)
			aiOutput := ui.NewAIOutput(opts.Verbose)
			g := git.NewGit(opts.DryRun, opts.Verbose)

			// Validate git repo
			if err := g.ValidateGitRepo(); err != nil {
				return ui.NewError("not in a git repository", "run this command from within a git repository")
			}

			// Check for staged changes
			hasStaged, err := g.HasStagedChanges()
			if err != nil {
				return fmt.Errorf("failed to check staged changes: %w", err)
			}
			if !hasStaged {
				return ui.NewError("no staged changes", "stage your changes first with 'git add'")
			}

			// Get commit message
			var commitMessage string
			if message != "" {
				commitMessage = message
				output.Info("Using provided commit message")
			} else {
				// Load AI configuration
				manager, err := aiconfig.NewManager()
				if err != nil {
					return fmt.Errorf("failed to create config manager: %w", err)
				}

				if !manager.Exists() {
					return ui.NewError(
						"AI configuration not found",
						"run 'gitext ai setup' to configure AI provider",
					)
				}

				cfg, err := manager.Load()
				if err != nil {
					return fmt.Errorf("failed to load AI configuration: %w", err)
				}

				// Create AI service
				service, err := ai.NewService(cfg)
				if err != nil {
					return fmt.Errorf("failed to create AI service: %w", err)
				}

				// Get staged diff
				output.Doing("Getting staged changes")
				diff, err := g.GetStagedDiff()
				if err != nil {
					return fmt.Errorf("failed to get staged diff: %w", err)
				}

				if diff == "" {
					return ui.NewError("no changes in diff", "ensure you have staged changes")
				}

				// Generate commit message
				aiOutput.GeneratingCommitMessage()
				commitMessage, err = service.GenerateCommitMessage(diff)
				if err != nil {
					return fmt.Errorf("failed to generate commit message: %w", err)
				}

				aiOutput.CommitMessageGenerated(commitMessage)
			}

			// Show commit message
			if opts.Verbose {
				output.Verbose("Commit message: %s", commitMessage)
			}

			// Confirm commit
			if !opts.DryRun {
				confirmed, err := ui.PromptConfirm("Create commit with this message?", true)
				if err != nil {
					return err
				}
				if !confirmed {
					output.Info("Commit cancelled")
					return nil
				}
			}

			// Create commit
			output.Doing("Creating commit")
			if _, err := g.RunWithTimeout("commit", "-m", commitMessage); err != nil {
				return fmt.Errorf("failed to create commit: %w", err)
			}

			output.Success("Commit created successfully")
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Use this commit message instead of generating one")

	return cmd
}
