// Package worktree owns gitt's per-branch worktree layout convention:
// <repo>/.worktrees/<safe-branch>.
package worktree

import (
	"path/filepath"
	"strings"
)

// SafeBranch turns a git branch name into a single path segment by replacing
// directory separators with dashes. e.g. "feature/foo" -> "feature-foo".
func SafeBranch(branch string) string {
	return strings.NewReplacer("/", "-", "\\", "-").Replace(branch)
}

// Path returns the directory where the worktree for branch should live,
// following gitt's layout: <repo>/.worktrees/<safe-branch>. mainRoot must be
// the main repository's top-level directory (see gitx.MainRepoRoot).
func Path(mainRoot, branch string) string {
	return filepath.Join(mainRoot, ".worktrees", SafeBranch(branch))
}
