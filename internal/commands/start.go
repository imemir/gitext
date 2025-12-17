package commands

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/imemir/gitext/pkg/config"
	"github.com/imemir/gitext/pkg/git"
	"github.com/imemir/gitext/pkg/ui"
	"github.com/spf13/cobra"
)

func NewStartCmd(opts *Options) *cobra.Command {
	var ticket, slug, from string

	cmd := &cobra.Command{
		Use:   "start feature",
		Short: "Start a new feature branch",
		Long: `Create a new feature branch from stage or production.
The branch name will be: feature/<ticket>-<slug>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "feature" {
				return fmt.Errorf("only 'feature' is supported currently")
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
			if ticket == "" {
				return fmt.Errorf("--ticket is required")
			}
			if slug == "" {
				return fmt.Errorf("--slug is required")
			}
			if from == "" {
				return fmt.Errorf("--from is required (stage or production)")
			}

			var sourceBranch string
			switch from {
			case "stage":
				sourceBranch = cfg.Branch.Stage
			case "production":
				sourceBranch = cfg.Branch.Production
			default:
				return fmt.Errorf("--from must be 'stage' or 'production'")
			}

			// Validate remote
			if err := g.ValidateRemote(cfg.Remote.Name); err != nil {
				return err
			}

			// Validate source branch exists
			if err := g.ValidateBranchExists(sourceBranch, cfg.Remote.Name); err != nil {
				return err
			}

			// Check working tree
			isClean, err := g.IsWorkingTreeClean()
			if err != nil {
				return fmt.Errorf("failed to check working tree: %w", err)
			}
			if !isClean {
				return ui.NewError("working tree has uncommitted changes", "commit or stash changes first")
			}

			// Generate branch name
			branchName := fmt.Sprintf("feature/%s-%s", ticket, slug)

			// Validate branch name matches pattern
			pattern := strings.ReplaceAll(cfg.Naming.Feature, "*", ".*")
			matched, err := regexp.MatchString("^"+pattern+"$", branchName)
			if err != nil {
				return fmt.Errorf("failed to validate branch name: %w", err)
			}
			if !matched {
				return fmt.Errorf("branch name '%s' does not match pattern '%s'", branchName, cfg.Naming.Feature)
			}

			// Check if branch already exists
			exists, err := g.BranchExists(branchName)
			if err != nil {
				return fmt.Errorf("failed to check if branch exists: %w", err)
			}
			if exists {
				return fmt.Errorf("branch '%s' already exists", branchName)
			}

			// Fetch latest
			output.Doing("Fetching latest from %s", cfg.Remote.Name)
			if _, err := g.RunWithTimeout("fetch", cfg.Remote.Name); err != nil {
				return fmt.Errorf("failed to fetch: %w", err)
			}

			// Checkout source branch
			output.Doing("Checking out %s", sourceBranch)
			if _, err := g.RunWithTimeout("checkout", sourceBranch); err != nil {
				return fmt.Errorf("failed to checkout %s: %w", sourceBranch, err)
			}

			// Pull latest
			output.Doing("Pulling latest changes")
			if _, err := g.RunWithTimeout("pull", "--ff-only", cfg.Remote.Name, sourceBranch); err != nil {
				output.Warning("Fast-forward pull failed, continuing anyway")
			}

			// Create and checkout new branch
			output.Doing("Creating branch %s", branchName)
			if _, err := g.RunWithTimeout("checkout", "-b", branchName); err != nil {
				return fmt.Errorf("failed to create branch: %w", err)
			}
			output.Did("Created and checked out %s", branchName)

			output.Next("start making changes, then run: gitext prepare pr --to stage")

			return nil
		},
	}

	cmd.Flags().StringVar(&ticket, "ticket", "", "Ticket ID (e.g., KWS-123)")
	cmd.Flags().StringVar(&slug, "slug", "", "Feature slug (e.g., retry-policy)")
	cmd.Flags().StringVar(&from, "from", "", "Source branch (stage or production)")

	return cmd
}
