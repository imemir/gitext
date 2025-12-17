package commands

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gitext/gitext/pkg/config"
	"github.com/gitext/gitext/pkg/git"
	"github.com/gitext/gitext/pkg/ui"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *Options) *cobra.Command {
	var with, mode string

	cmd := &cobra.Command{
		Use:   "update feature",
		Short: "Update feature branch with changes from stage or production",
		Long: `Update the current feature branch with changes from stage or production.
Uses rebase or merge based on the --mode flag.`,
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
			if with == "" {
				return fmt.Errorf("--with is required (stage or production)")
			}
			if mode == "" {
				mode = "rebase" // default
			}
			if mode != "rebase" && mode != "merge" {
				return fmt.Errorf("--mode must be 'rebase' or 'merge'")
			}

			var sourceBranch string
			switch with {
			case "stage":
				sourceBranch = cfg.Branch.Stage
			case "production":
				sourceBranch = cfg.Branch.Production
			default:
				return fmt.Errorf("--with must be 'stage' or 'production'")
			}

			// Get current branch
			currentBranch, err := g.GetCurrentBranch()
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}

			// Validate current branch is a feature branch
			pattern := strings.ReplaceAll(cfg.Naming.Feature, "*", ".*")
			matched, err := regexp.MatchString("^"+pattern+"$", currentBranch)
			if err != nil {
				return fmt.Errorf("failed to validate branch name: %w", err)
			}
			if !matched {
				return fmt.Errorf("current branch '%s' does not match feature pattern '%s'", currentBranch, cfg.Naming.Feature)
			}

			// Validate remote
			if err := g.ValidateRemote(cfg.Remote.Name); err != nil {
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

			// Fetch latest
			output.Doing("Fetching latest from %s", cfg.Remote.Name)
			if _, err := g.RunWithTimeout("fetch", cfg.Remote.Name); err != nil {
				return fmt.Errorf("failed to fetch: %w", err)
			}

			// Update source branch first (sync)
			output.Doing("Updating %s", sourceBranch)
			if _, err := g.RunWithTimeout("fetch", cfg.Remote.Name, fmt.Sprintf("%s:%s", sourceBranch, sourceBranch)); err != nil {
				output.Verbose("Note: %s may not exist locally, using remote reference", sourceBranch)
			}

			// Apply changes
			remoteRef := fmt.Sprintf("%s/%s", cfg.Remote.Name, sourceBranch)
			if mode == "rebase" {
				output.Doing("Rebasing onto %s", remoteRef)
				if _, err := g.RunWithTimeout("rebase", remoteRef); err != nil {
					output.Error("Rebase encountered conflicts")
					output.Next("resolve conflicts, then run: git rebase --continue")
					return fmt.Errorf("rebase failed: %w", err)
				}
				output.Did("Rebased onto %s", remoteRef)
			} else {
				output.Doing("Merging %s", remoteRef)
				if _, err := g.RunWithTimeout("merge", remoteRef); err != nil {
					output.Error("Merge encountered conflicts")
					output.Next("resolve conflicts, then run: git commit")
					return fmt.Errorf("merge failed: %w", err)
				}
				output.Did("Merged %s", remoteRef)
			}

			output.Next("push changes: git push %s %s", cfg.Remote.Name, currentBranch)

			return nil
		},
	}

	cmd.Flags().StringVar(&with, "with", "", "Source branch to update from (stage or production)")
	cmd.Flags().StringVar(&mode, "mode", "rebase", "Update mode: rebase or merge")

	return cmd
}

