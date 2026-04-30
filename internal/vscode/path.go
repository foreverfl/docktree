package vscode

import "path/filepath"

// WorkspacePath returns the conventional .code-workspace path for mainRoot:
// <mainRoot>/<repoName>.code-workspace, where repoName is the basename of
// the main repository directory. Mirrors gitx.WorktreePath in spirit —
// vscode's per-repo file layout, kept separate from the workspace document
// shape and daemon I/O in workspace.go.
func WorkspacePath(mainRoot string) string {
	return filepath.Join(mainRoot, filepath.Base(mainRoot)+".code-workspace")
}
