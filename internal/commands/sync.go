package commands

import (
	"fmt"

	"github.com/imemir/gitext/pkg/config"
	"github.com/imemir/gitext/pkg/git"
	"github.com/imemir/gitext/pkg/ui"
	"github.com/spf13/cobra"
)

func NewSyncCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync [stage|production]",
		Short: "Safely sync a branch with its remote",
		Long: `Fetch from remote and pull with --ff-only to safely update stage or production.
Fails if fast-forward is not possible, suggesting an update command instead.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			output := ui.NewOutput(opts.Verbose)
			g := git.NewGit(opts.DryRun, opts.Verbose)

			if err := g.ValidateGitRepo(); err != nil {
				return ui.NewError("not in a git repository", "run this command from within a git repository")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			var branch string
			switch target {
			case "stage":
				branch = cfg.Branch.Stage
			case "production":
				branch = cfg.Branch.Production
			default:
				return fmt.Errorf("invalid target '%s', must be 'stage' or 'production'", target)
			}

			// Validate remote
			if err := g.ValidateRemote(cfg.Remote.Name); err != nil {
				return err
			}

			// Check if branch exists
			if err := g.ValidateBranchExists(branch, cfg.Remote.Name); err != nil {
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

			currentBranch, err := g.GetCurrentBranch()
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}

			// Checkout branch if not already on it
			if currentBranch != branch {
				output.Doing("Checking out %s", branch)
				if _, err := g.RunWithTimeout("checkout", branch); err != nil {
					return fmt.Errorf("failed to checkout %s: %w", branch, err)
				}
				output.Did("Checked out %s", branch)
			}

			// Fetch from remote
			output.Doing("Fetching from %s", cfg.Remote.Name)
			if _, err := g.RunWithTimeout("fetch", cfg.Remote.Name); err != nil {
				return fmt.Errorf("failed to fetch: %w", err)
			}
			output.Did("Fetched from %s", cfg.Remote.Name)

			// Pull with --ff-only
			output.Doing("Pulling with --ff-only")
			remoteRef := fmt.Sprintf("%s/%s", cfg.Remote.Name, branch)
			if _, err := g.RunWithTimeout("pull", "--ff-only", cfg.Remote.Name, branch); err != nil {
				output.Error("Fast-forward pull failed")
				output.Next("branch has diverged, run: git pull --rebase %s %s", cfg.Remote.Name, branch)
				return fmt.Errorf("fast-forward not possible: %w", err)
			}
			output.Did("Pulled %s", remoteRef)

			// Show status
			ahead, behind, err := g.GetAheadBehind(cfg.Remote.Name, branch)
			if err == nil {
				if ahead == 0 && behind == 0 {
					output.Success("%s is up to date with %s", branch, remoteRef)
				} else {
					output.Info("Ahead: %d, Behind: %d", ahead, behind)
				}
			}

			output.Next("continue working or run: gitext status")

			return nil
		},
	}

	return cmd
}
