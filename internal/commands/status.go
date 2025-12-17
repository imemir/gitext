package commands

import (
	"fmt"

	"github.com/gitext/gitext/pkg/config"
	"github.com/gitext/gitext/pkg/git"
	"github.com/gitext/gitext/pkg/ui"
	"github.com/spf13/cobra"
)

func NewStatusCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current git status and suggest next steps",
		Long: `Show the current branch, ahead/behind status vs stage and production,
working tree state, and suggest the next recommended command.`,
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

			// Check if detached HEAD
			isDetached, err := g.IsDetachedHEAD()
			if err == nil && isDetached {
				output.Warning("HEAD is detached")
				output.Next("checkout a branch: git checkout -b <branch-name>")
				return nil
			}

			output.Info("Current branch: %s", currentBranch)

			// Check working tree
			isClean, err := g.IsWorkingTreeClean()
			if err != nil {
				return fmt.Errorf("failed to check working tree: %w", err)
			}

			if !isClean {
				output.Warning("Working tree has uncommitted changes")
				output.Next("commit or stash changes: git commit -am 'message' or git stash")
			} else {
				output.Success("Working tree is clean")
			}

			// Validate remote
			if err := g.ValidateRemote(cfg.Remote.Name); err != nil {
				output.Warning("Remote '%s' not configured", cfg.Remote.Name)
				return nil
			}

			// Fetch to get latest remote state
			output.Verbose("Fetching from remote...")
			if _, err := g.RunWithTimeout("fetch", cfg.Remote.Name); err != nil {
				output.Warning("Failed to fetch from remote: %v", err)
			}

			// Check status vs remote branch
			remoteBranchExists, err := g.RemoteBranchExists(cfg.Remote.Name, currentBranch)
			if err == nil && remoteBranchExists {
				ahead, behind, err := g.GetAheadBehind(cfg.Remote.Name, currentBranch)
				if err == nil {
					if ahead > 0 {
						output.Info("Ahead of %s/%s by %d commit(s)", cfg.Remote.Name, currentBranch, ahead)
					}
					if behind > 0 {
						output.Warning("Behind %s/%s by %d commit(s)", cfg.Remote.Name, currentBranch, behind)
						output.Next("sync with remote: gitext sync %s", currentBranch)
					}
					if ahead > 0 && behind == 0 {
						output.Next("push changes: git push %s %s", cfg.Remote.Name, currentBranch)
					}
				}
			}

			// Check status vs stage
			if currentBranch != cfg.Branch.Stage {
				stageExists, err := g.RemoteBranchExists(cfg.Remote.Name, cfg.Branch.Stage)
				if err == nil && stageExists {
					_, behind, err := g.GetAheadBehind(cfg.Remote.Name, cfg.Branch.Stage)
					if err == nil {
						if behind > 0 {
							output.Info("Behind %s by %d commit(s)", cfg.Branch.Stage, behind)
							if isClean {
								output.Next("update with stage: gitext update feature --with stage")
							}
						}
					}
				}
			}

			// Check status vs production
			if currentBranch != cfg.Branch.Production {
				prodExists, err := g.RemoteBranchExists(cfg.Remote.Name, cfg.Branch.Production)
				if err == nil && prodExists {
					_, behind, err := g.GetAheadBehind(cfg.Remote.Name, cfg.Branch.Production)
					if err == nil {
						if behind > 0 {
							output.Info("Behind %s by %d commit(s)", cfg.Branch.Production, behind)
						}
					}
				}
			}

			// Suggest next steps based on branch type
			if isClean {
				if currentBranch == cfg.Branch.Stage || currentBranch == cfg.Branch.Production {
					output.Next("sync latest changes: gitext sync %s", currentBranch)
				} else {
					output.Next("prepare PR: gitext prepare pr --to stage")
				}
			}

			return nil
		},
	}

	return cmd
}

