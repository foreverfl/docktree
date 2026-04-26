// Package gitx wraps the git CLI calls doctree needs. It shells out rather
// than linking a git library to keep the binary small and stay close to
// observable git behavior.
package gitx

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RepoRoot returns the absolute path to the enclosing git repository's
// top-level directory. Errors if the current working directory is not inside
// a git repo.
func RepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// WorktreeAdd runs `git worktree add <target> <branch>`, streaming git's
// progress output to the caller's stdout/stderr.
func WorktreeAdd(target, branch string) error {
	cmd := exec.Command("git", "worktree", "add", target, branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git worktree add: %w", err)
	}
	return nil
}
