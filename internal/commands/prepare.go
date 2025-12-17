package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/imemir/gitext/pkg/config"
	"github.com/imemir/gitext/pkg/git"
	"github.com/imemir/gitext/pkg/ui"
	"github.com/spf13/cobra"
)

func NewPrepareCmd(opts *Options) *cobra.Command {
	var to string

	cmd := &cobra.Command{
		Use:   "prepare pr",
		Short: "Prepare a pull request",
		Long: `Run CI checks and generate PR text for the current branch.
CI commands are run based on the target branch (stage or production).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "pr" {
				return fmt.Errorf("only 'pr' is supported")
			}

			output := ui.NewOutput(opts.Verbose)
			g := git.NewGit(opts.DryRun, opts.Verbose)

			if err := g.ValidateGitRepo(); err != nil {
				return ui.NewError("not in a git repository", "run this command from within a git repository")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Validate flags
			if to == "" {
				return fmt.Errorf("--to is required (stage or production)")
			}

			var targetBranch string
			switch to {
			case "stage":
				targetBranch = cfg.Branch.Stage
			case "production":
				targetBranch = cfg.Branch.Production
			default:
				return fmt.Errorf("--to must be 'stage' or 'production'")
			}

			// Get current branch
			currentBranch, err := g.GetCurrentBranch()
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}

			// Get CI commands for target
			var ciCommands []string
			if to == "stage" {
				ciCommands = cfg.CI.Stage
			} else {
				ciCommands = cfg.CI.Production
			}

			// Run CI commands
			if len(ciCommands) > 0 {
				output.Doing("Running CI checks for %s", to)
				for _, cmdStr := range ciCommands {
					output.Verbose("Running: %s", cmdStr)
					if !opts.DryRun {
						if err := runCICommand(cmdStr); err != nil {
							output.Error("CI check failed: %s", cmdStr)
							return fmt.Errorf("CI check failed: %w", err)
						}
					}
				}
				output.Did("All CI checks passed")
			} else {
				output.Info("No CI commands configured for %s", to)
			}

			// Generate PR text
			output.Doing("Generating PR text")

			prText := generatePRText(cfg, currentBranch, targetBranch, g, output)

			// Print PR text to stdout
			output.Print("\n" + prText + "\n")

			output.Did("PR text generated")
			output.Next("create PR on GitHub/GitLab or copy the text above")

			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "Target branch for PR (stage or production)")

	return cmd
}

func runCICommand(cmdStr string) error {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func generatePRText(cfg *config.Config, currentBranch, targetBranch string, g *git.Git, output *ui.Output) string {
	var prText strings.Builder

	// Load template if configured
	if cfg.PR.TemplatePath != "" {
		gitRoot, err := config.GetGitRoot()
		if err == nil {
			templatePath := filepath.Join(gitRoot, cfg.PR.TemplatePath)
			if data, err := os.ReadFile(templatePath); err == nil {
				prText.WriteString(string(data))
				prText.WriteString("\n\n---\n\n")
			}
		}
	}

	// Extract ticket from branch name if possible
	ticket := extractTicketFromBranch(currentBranch)

	// Add branch info
	prText.WriteString(fmt.Sprintf("## Branch: %s\n\n", currentBranch))
	if ticket != "" {
		prText.WriteString(fmt.Sprintf("**Ticket:** %s\n\n", ticket))
	}
	prText.WriteString(fmt.Sprintf("**Target:** %s\n\n", targetBranch))

	// Get commit summary
	commits, err := getCommitSummary(cfg.Remote.Name, targetBranch, g)
	if err == nil && commits != "" {
		prText.WriteString("## Commits\n\n")
		prText.WriteString(commits)
		prText.WriteString("\n")
	}

	// Add description placeholder
	prText.WriteString("\n## Description\n\n")
	prText.WriteString("<!-- Add description here -->\n")

	return prText.String()
}

func extractTicketFromBranch(branch string) string {
	// Try to extract ticket ID from branch name (e.g., feature/KWS-123-slug -> KWS-123)
	parts := strings.Split(branch, "/")
	if len(parts) > 1 {
		nameParts := strings.Split(parts[1], "-")
		if len(nameParts) >= 2 {
			// Check if it looks like a ticket ID (e.g., KWS-123)
			if len(nameParts[0]) >= 2 && len(nameParts[1]) >= 1 {
				return fmt.Sprintf("%s-%s", nameParts[0], nameParts[1])
			}
		}
	}
	return ""
}

func getCommitSummary(remote, targetBranch string, g *git.Git) (string, error) {
	currentBranch, err := g.GetCurrentBranch()
	if err != nil {
		return "", err
	}

	targetRef := fmt.Sprintf("%s/%s", remote, targetBranch)
	output, err := g.RunWithTimeout("log", "--oneline", fmt.Sprintf("%s..%s", targetRef, currentBranch))
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(output) == "" {
		return "No commits (branch is up to date or behind)", nil
	}

	lines := strings.Split(output, "\n")
	var summary strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			summary.WriteString("- ")
			summary.WriteString(line)
			summary.WriteString("\n")
		}
	}

	return summary.String(), nil
}
