package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	gitTimeout = 30 * time.Second
)

// Git wraps git command execution
type Git struct {
	dryRun  bool
	verbose bool
}

// NewGit creates a new Git instance
func NewGit(dryRun, verbose bool) *Git {
	return &Git{
		dryRun:  dryRun,
		verbose: verbose,
	}
}

// Run executes a git command and returns the output
func (g *Git) Run(ctx context.Context, args ...string) (string, error) {
	return g.RunWithDir(ctx, "", args...)
}

// RunWithDir executes a git command in a specific directory
func (g *Git) RunWithDir(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}

	if g.dryRun {
		fmt.Printf("[DRY RUN] git %s\n", strings.Join(args, " "))
		return "", nil
	}

	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if g.verbose {
		fmt.Printf("$ git %s\n%s\n", strings.Join(args, " "), outputStr)
	}

	if err != nil {
		return outputStr, fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, outputStr)
	}

	return outputStr, nil
}

// RunWithTimeout executes a git command with a timeout
func (g *Git) RunWithTimeout(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	return g.Run(ctx, args...)
}

// RunWithTimeoutAndDir executes a git command in a directory with a timeout
func (g *Git) RunWithTimeoutAndDir(dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	return g.RunWithDir(ctx, dir, args...)
}

