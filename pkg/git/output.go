package git

import (
	"fmt"
	"strings"
)

// GetCurrentBranch returns the current branch name
func (g *Git) GetCurrentBranch() (string, error) {
	output, err := g.RunWithTimeout("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// IsWorkingTreeClean checks if the working tree is clean
func (g *Git) IsWorkingTreeClean() (bool, error) {
	output, err := g.RunWithTimeout("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) == "", nil
}

// GetRemoteBranches returns a list of remote branch names
func (g *Git) GetRemoteBranches(remote string) ([]string, error) {
	output, err := g.RunWithTimeout("branch", "-r", "--format", "%(refname:short)")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var branches []string
	prefix := remote + "/"

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, prefix) {
			branches = append(branches, strings.TrimPrefix(line, prefix))
		}
	}

	return branches, nil
}

// BranchExists checks if a branch exists locally
func (g *Git) BranchExists(branch string) (bool, error) {
	output, err := g.RunWithTimeout("branch", "--list", branch)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) != "", nil
}

// RemoteBranchExists checks if a remote branch exists
func (g *Git) RemoteBranchExists(remote, branch string) (bool, error) {
	output, err := g.RunWithTimeout("ls-remote", "--heads", remote, branch)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) != "", nil
}

// GetAheadBehind returns the ahead/behind counts for the current branch vs a remote branch
func (g *Git) GetAheadBehind(remote, branch string) (ahead, behind int, err error) {
	currentBranch, err := g.GetCurrentBranch()
	if err != nil {
		return 0, 0, err
	}

	output, err := g.RunWithTimeout("rev-list", "--left-right", "--count",
		fmt.Sprintf("%s/%s...%s", remote, branch, currentBranch))
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Fields(output)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected output from rev-list: %s", output)
	}

	behind, err = parseInt(parts[0])
	if err != nil {
		return 0, 0, err
	}

	ahead, err = parseInt(parts[1])
	if err != nil {
		return 0, 0, err
	}

	return ahead, behind, nil
}

// GetCommitAuthors returns unique authors from recent commits (last N commits)
func (g *Git) GetCommitAuthors(count int) ([]string, error) {
	output, err := g.RunWithTimeout("log", "-n", fmt.Sprintf("%d", count), "--format=%an")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	authorMap := make(map[string]bool)
	var authors []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !authorMap[line] {
			authorMap[line] = true
			authors = append(authors, line)
		}
	}

	return authors, nil
}

// GetMergedBranches returns local branches that are merged into the given branch
func (g *Git) GetMergedBranches(intoBranch string) ([]string, error) {
	output, err := g.RunWithTimeout("branch", "--merged", intoBranch, "--format", "%(refname:short)")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var branches []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == intoBranch || strings.HasPrefix(line, "*") {
			continue
		}
		branches = append(branches, line)
	}

	return branches, nil
}

// IsDetachedHEAD checks if HEAD is detached
func (g *Git) IsDetachedHEAD() (bool, error) {
	output, err := g.RunWithTimeout("symbolic-ref", "-q", "HEAD")
	if err != nil {
		// If command fails, we're likely in detached HEAD
		return true, nil
	}
	return strings.TrimSpace(output) == "", nil
}

// GetStagedDiff returns the diff of staged changes
func (g *Git) GetStagedDiff() (string, error) {
	output, err := g.RunWithTimeout("diff", "--cached")
	if err != nil {
		return "", err
	}
	return output, nil
}

// HasStagedChanges checks if there are any staged changes
func (g *Git) HasStagedChanges() (bool, error) {
	_, err := g.RunWithTimeout("diff", "--cached", "--quiet")
	if err != nil {
		// If command fails, there are staged changes
		return true, nil
	}
	// If command succeeds, there are no staged changes
	return false, nil
}

// parseInt parses an integer from a string
func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

