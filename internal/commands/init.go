package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/imemir/gitext/pkg/config"
	"github.com/imemir/gitext/pkg/ui"
	"github.com/spf13/cobra"
)

func NewInitCmd(opts *Options) *cobra.Command {
	var installHooks bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize gitext configuration",
		Long: `Initialize gitext by creating a .gitext configuration file in the repository root.
Optionally install git hooks to prevent direct pushes to protected branches.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			output := ui.NewOutput(opts.Verbose)
			cfg, err := config.Load()
			if err != nil {
				// Not in a git repo
				return fmt.Errorf("failed to initialize: %w", err)
			}

			gitRoot, err := config.GetGitRoot()
			if err != nil {
				return fmt.Errorf("failed to get git root: %w", err)
			}

			configPath := filepath.Join(gitRoot, ".gitext")
			_, err = os.Stat(configPath)
			if err == nil {
				output.Info(".gitext already exists at %s", configPath)
				output.Next("edit .gitext to customize configuration")
			} else {
				output.Doing("Creating .gitext configuration file")
				if err := cfg.Save(); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}
				output.Did("Created .gitext at %s", configPath)
			}

			if installHooks {
				if err := installPrePushHook(gitRoot, cfg, output); err != nil {
					return fmt.Errorf("failed to install hooks: %w", err)
				}
			} else {
				output.Next("run 'gitext init --install-hooks' to install git hooks")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&installHooks, "install-hooks", false, "Install git hooks to prevent direct pushes to protected branches")

	return cmd
}

func installPrePushHook(gitRoot string, cfg *config.Config, output *ui.Output) error {
	hooksDir := filepath.Join(gitRoot, ".git", "hooks")
	hookPath := filepath.Join(hooksDir, "pre-push")

	output.Doing("Installing pre-push hook")

	hookContent := generatePrePushHook(cfg)

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("failed to write hook: %w", err)
	}

	output.Did("Installed pre-push hook at %s", hookPath)
	return nil
}

func generatePrePushHook(cfg *config.Config) string {
	return fmt.Sprintf(`#!/bin/sh
# gitext pre-push hook
# Prevents direct pushes to protected branches (stage/production)

protected_branches="%s %s"
remote="$1"
url="$2"

while read local_ref local_sha remote_ref remote_sha
do
    branch=$(echo "$remote_ref" | sed 's|refs/heads/||')
    
    for protected in $protected_branches; do
        if [ "$branch" = "$protected" ]; then
            # Allow CI users (detected via environment variables)
            if [ -n "$CI" ] || [ -n "$GITHUB_ACTIONS" ] || [ -n "$GITLAB_CI" ]; then
                exit 0
            fi
            
            echo "Error: Direct push to protected branch '$branch' is not allowed."
            echo "Please create a pull request instead."
            exit 1
        fi
    done
done

exit 0
`, cfg.Branch.Production, cfg.Branch.Stage)
}
