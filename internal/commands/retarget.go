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

func NewRetargetCmd(opts *Options) *cobra.Command {
	var onto, from string
	var override, iKnowWhatImDoing bool

	cmd := &cobra.Command{
		Use:   "retarget feature",
		Short: "Retarget a feature branch from stage onto production",
		Long: `Rebase a feature branch that was based on stage onto production.
This is useful when a feature is ready for production but was developed from stage.
Uses 'git rebase --onto' to rewrite history safely.`,
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
			if onto == "" {
				onto = "production"
			}
			if from == "" {
				from = "stage"
			}

			if onto != "production" {
				return fmt.Errorf("--onto must be 'production'")
			}
			if from != "stage" {
				return fmt.Errorf("--from must be 'stage'")
			}

			ontoBranch := cfg.Branch.Production
			fromBranch := cfg.Branch.Stage

			// Get current branch
			currentBranch, err := g.GetCurrentBranch()
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}

			// Validate current branch is a feature branch (unless override)
			if !override {
				pattern := strings.ReplaceAll(cfg.Naming.Feature, "*", ".*")
				matched, err := regexp.MatchString("^"+pattern+"$", currentBranch)
				if err != nil {
					return fmt.Errorf("failed to validate branch name: %w", err)
				}
				if !matched {
					return fmt.Errorf("current branch '%s' does not match feature pattern '%s' (use --override to bypass)", currentBranch, cfg.Naming.Feature)
				}
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

			// Check if branch appears shared
			remoteBranchExists, err := g.RemoteBranchExists(cfg.Remote.Name, currentBranch)
			if err == nil && remoteBranchExists {
				// Check for multiple authors in recent commits
				authors, err := g.GetCommitAuthors(10)
				if err == nil && len(authors) > 1 {
					if !iKnowWhatImDoing {
						output.Error("Branch '%s' appears to be shared (multiple authors in recent commits)", currentBranch)
						output.Warning("Retargeting will rewrite history and may affect other developers")
						return fmt.Errorf("branch appears shared → use --i-know-what-im-doing to proceed")
					}
					output.Warning("Branch appears shared, proceeding with --i-know-what-im-doing flag")
				}
			}

			// Fetch latest
			output.Doing("Fetching latest from %s", cfg.Remote.Name)
			if _, err := g.RunWithTimeout("fetch", cfg.Remote.Name); err != nil {
				return fmt.Errorf("failed to fetch: %w", err)
			}

			// Validate branches exist
			if err := g.ValidateBranchExists(ontoBranch, cfg.Remote.Name); err != nil {
				return fmt.Errorf("target branch '%s' does not exist: %w", ontoBranch, err)
			}
			if err := g.ValidateBranchExists(fromBranch, cfg.Remote.Name); err != nil {
				return fmt.Errorf("source branch '%s' does not exist: %w", fromBranch, err)
			}

			// Execute rebase --onto
			ontoRef := fmt.Sprintf("%s/%s", cfg.Remote.Name, ontoBranch)
			fromRef := fmt.Sprintf("%s/%s", cfg.Remote.Name, fromBranch)

			output.Doing("Retargeting %s onto %s (from %s)", currentBranch, ontoRef, fromRef)
			output.Warning("This will rewrite history. If the branch is pushed, you'll need to force push.")

			if _, err := g.RunWithTimeout("rebase", "--onto", ontoRef, fromRef); err != nil {
				output.Error("Rebase encountered conflicts")
				output.Next("resolve conflicts, then run: git rebase --continue")
				return fmt.Errorf("rebase failed: %w", err)
			}

			output.Did("Retargeted %s onto %s", currentBranch, ontoRef)

			// Check if remote branch exists and warn about force push
			if remoteBranchExists {
				output.Warning("Remote branch exists. You'll need to force push with --force-with-lease")
				output.Next("push with force: git push --force-with-lease %s %s", cfg.Remote.Name, currentBranch)
			} else {
				output.Next("push branch: git push %s %s", cfg.Remote.Name, currentBranch)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&onto, "onto", "production", "Target branch (must be production)")
	cmd.Flags().StringVar(&from, "from", "stage", "Source branch (must be stage)")
	cmd.Flags().BoolVar(&override, "override", false, "Allow retargeting non-feature branches")
	cmd.Flags().BoolVar(&iKnowWhatImDoing, "i-know-what-im-doing", false, "Bypass shared branch safety check")

	return cmd
}

