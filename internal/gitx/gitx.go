// Package gitx wraps the git CLI calls gitt needs. It shells out rather
// than linking a git library to keep the binary small and stay close to
// observable git behavior.
package gitx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// MainRepoRoot returns the main repository's top-level directory. When called
// from inside a linked worktree, this differs from RepoRoot, which returns the
// worktree's own toplevel.
func MainRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-common-dir").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository: %w", err)
	}
	return filepath.Dir(strings.TrimSpace(string(out))), nil
}

// CurrentBranch returns the short branch name of HEAD, or empty string if
// HEAD is detached.
func CurrentBranch() (string, error) {
	out, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return "", fmt.Errorf("git branch --show-current: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// IsClean reports whether the working tree has no staged, unstaged, or
// untracked changes.
func IsClean() (bool, error) {
	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	return len(strings.TrimSpace(string(out))) == 0, nil
}

// HasConflicts reports whether the working tree has any unmerged paths.
// Conflicts can outlive an ongoing operation (e.g. `git stash pop` may leave
// unmerged files behind without setting MERGE_HEAD), so this is checked
// independently from OngoingOp.
func HasConflicts() (bool, error) {
	out, err := exec.Command("git", "diff", "--name-only", "--diff-filter=U").Output()
	if err != nil {
		return false, fmt.Errorf("git diff --diff-filter=U: %w", err)
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

// OngoingOp returns the name of the in-progress git operation (merging,
// rebasing, cherry-picking, reverting, bisecting), or an empty string if no
// operation is in progress.
func OngoingOp() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-dir").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --git-dir: %w", err)
	}
	gitDir := strings.TrimSpace(string(out))
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(gitDir, name))
		return err == nil
	}
	switch {
	case exists("rebase-merge"), exists("rebase-apply"):
		return "rebasing", nil
	case exists("MERGE_HEAD"):
		return "merging", nil
	case exists("CHERRY_PICK_HEAD"):
		return "cherry-picking", nil
	case exists("REVERT_HEAD"):
		return "reverting", nil
	case exists("BISECT_LOG"):
		return "bisecting", nil
	}
	return "", nil
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

// WorktreeRemove runs `git worktree remove <target>`, streaming git's output
// to the caller's stdout/stderr. Fails if the worktree has uncommitted or
// untracked changes; git's own message explains the cause.
func WorktreeRemove(target string) error {
	cmd := exec.Command("git", "worktree", "remove", target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git worktree remove: %w", err)
	}
	return nil
}
