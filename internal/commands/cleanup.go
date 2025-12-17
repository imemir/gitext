package commands

import (
	"fmt"

	"github.com/imemir/gitext/pkg/config"
	"github.com/imemir/gitext/pkg/git"
	"github.com/imemir/gitext/pkg/ui"
	"github.com/spf13/cobra"
)

func NewCleanupCmd(opts *Options) *cobra.Command {
	var hard bool

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up merged local branches",
		Long: `List and optionally delete local branches that have been merged.
By default, shows what would be deleted. Use --hard to actually delete branches.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			output := ui.NewOutput(opts.Verbose)
			g := git.NewGit(opts.DryRun, opts.Verbose)

			if err := g.ValidateGitRepo(); err != nil {
				return ui.NewError("not in a git repository", "run this command from within a git repository")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get current branch
			currentBranch, err := g.GetCurrentBranch()
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}

			// Get merged branches from both stage and production
			var allMergedBranches []string
			mergedMap := make(map[string]bool)

			// Check merged into stage
			stageMerged, err := g.GetMergedBranches(cfg.Branch.Stage)
			if err == nil {
				for _, branch := range stageMerged {
					if !mergedMap[branch] && branch != currentBranch {
						mergedMap[branch] = true
						allMergedBranches = append(allMergedBranches, branch)
					}
				}
			}

			// Check merged into production
			prodMerged, err := g.GetMergedBranches(cfg.Branch.Production)
			if err == nil {
				for _, branch := range prodMerged {
					if !mergedMap[branch] && branch != currentBranch {
						mergedMap[branch] = true
						allMergedBranches = append(allMergedBranches, branch)
					}
				}
			}

			if len(allMergedBranches) == 0 {
				output.Info("No merged branches to clean up")
				return nil
			}

			output.Info("Found %d merged branch(es):", len(allMergedBranches))
			for _, branch := range allMergedBranches {
				output.Print("  - %s", branch)
			}

			if !hard {
				output.Warning("Dry run mode - no branches deleted")
				output.Next("run with --hard to delete these branches: gitext cleanup --hard")
				return nil
			}

			// Delete branches
			output.Doing("Deleting merged branches")
			deleted := 0
			for _, branch := range allMergedBranches {
				// Skip protected branches
				if branch == cfg.Branch.Stage || branch == cfg.Branch.Production {
					output.Warning("Skipping protected branch: %s", branch)
					continue
				}

				if _, err := g.RunWithTimeout("branch", "-d", branch); err != nil {
					output.Warning("Failed to delete %s: %v", branch, err)
					// Try force delete if regular delete fails (for unmerged branches)
					if _, err := g.RunWithTimeout("branch", "-D", branch); err != nil {
						output.Error("Failed to force delete %s: %v", branch, err)
					} else {
						output.Verbose("Force deleted %s", branch)
						deleted++
					}
				} else {
					output.Verbose("Deleted %s", branch)
					deleted++
				}
			}

			output.Did("Deleted %d branch(es)", deleted)
			output.Next("run: gitext status")

			return nil
		},
	}

	cmd.Flags().BoolVar(&hard, "hard", false, "Actually delete branches (default is dry-run)")

	return cmd
}
