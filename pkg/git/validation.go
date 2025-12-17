package git

import (
	"fmt"
)

// ValidateGitRepo checks if we're in a git repository
func (g *Git) ValidateGitRepo() error {
	_, err := g.RunWithTimeout("rev-parse", "--git-dir")
	return err
}

// ValidateRemote checks if a remote exists
func (g *Git) ValidateRemote(remote string) error {
	output, err := g.RunWithTimeout("remote", "get-url", remote)
	if err != nil {
		return fmt.Errorf("remote '%s' does not exist → run 'git remote add %s <url>'", remote, remote)
	}
	if output == "" {
		return fmt.Errorf("remote '%s' has no URL configured", remote)
	}
	return nil
}

// ValidateBranchExists checks if a branch exists (local or remote)
func (g *Git) ValidateBranchExists(branch, remote string) error {
	// Check local first
	exists, err := g.BranchExists(branch)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// Check remote
	exists, err = g.RemoteBranchExists(remote, branch)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("branch '%s' does not exist locally or on '%s'", branch, remote)
	}

	return nil
}

