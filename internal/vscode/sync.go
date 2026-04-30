package vscode

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// SyncWorkspaceIfPresent rewrites the .code-workspace file at mainRoot if it
// already exists, so commands that change the worktree set (add, remove,
// rename) keep the workspace file in step with the daemon. When the file is
// not present, the user has not opted into VSCode workspace tracking yet and
// this is a no-op. Returns the workspace path and whether a sync occurred.
func SyncWorkspaceIfPresent(mainRoot string) (string, bool, error) {
	workspacePath := WorkspacePath(mainRoot)
	if _, err := os.Stat(workspacePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("stat workspace: %w", err)
	}
	folders, err := Folders(mainRoot)
	if err != nil {
		return "", false, err
	}
	if err := WriteWorkspace(workspacePath, folders); err != nil {
		return "", false, err
	}
	return workspacePath, true, nil
}

// Sync wraps SyncWorkspaceIfPresent with CLI-friendly reporting: it emits a
// one-line confirmation to info on success and routes failures to stderr as
// a warning so they never block the surrounding command. A missing workspace
// file is silent because that just means the user has not opted in.
func Sync(mainRoot string, info io.Writer) {
	path, synced, err := SyncWorkspaceIfPresent(mainRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: vscode workspace sync failed: %v\n", err)
		return
	}
	if synced {
		fmt.Fprintf(info, "  vscode: updated %s\n", path)
	}
}