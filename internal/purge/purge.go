// Package purge force-cleans gitt's registered worktrees off disk. Used by
// destructive flows (e.g. `gitt update`) that need to drop every worktree
// folder and its `git worktree` admin record before resetting the runtime
// directory.
package purge

import (
	"fmt"
	"os"

	"github.com/foreverfl/gitt/internal/gitx"
	"github.com/foreverfl/gitt/internal/store"
)

// LoadRegistered opens the daemon's SQLite store and returns every worktree
// row. Failures (missing db, open error, query error) are reported to stderr
// and treated as "nothing to clean" so the caller can still proceed.
func LoadRegistered(dbpath string) []store.Worktree {
	if _, err := os.Stat(dbpath); err != nil {
		return nil
	}
	storeHandle, err := store.Open(dbpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: open db at %s: %v (skipping worktree folder cleanup)\n", dbpath, err)
		return nil
	}
	defer storeHandle.Close()

	worktrees, err := storeHandle.ListWorktrees()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: list worktrees: %v\n", err)
		return nil
	}
	return worktrees
}

// RemoveRegistered force-removes the on-disk folder for each worktree row,
// then runs `git worktree prune` once per unique repo so orphaned admin
// records under .git/worktrees/<name> are dropped too. All failures are
// logged as warnings; the caller keeps going.
func RemoveRegistered(worktrees []store.Worktree) {
	if len(worktrees) == 0 {
		return
	}
	repos := make(map[string]struct{})
	removed := 0
	for _, worktree := range worktrees {
		repos[worktree.RepoRoot] = struct{}{}
		if _, err := os.Stat(worktree.WorktreePath); err != nil {
			continue
		}
		if err := os.RemoveAll(worktree.WorktreePath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: remove %s: %v\n", worktree.WorktreePath, err)
			continue
		}
		removed++
	}
	for repoRoot := range repos {
		if _, err := os.Stat(repoRoot); err != nil {
			continue
		}
		if err := gitx.WorktreePrune(repoRoot); err != nil {
			fmt.Fprintf(os.Stderr, "warning: prune %s: %v\n", repoRoot, err)
		}
	}
	if removed > 0 {
		fmt.Printf("removed %d worktree folder(s)\n", removed)
	}
}
