// Package vscode owns the VSCode multi-root workspace file that gitt
// generates from registered worktrees: discovery via the daemon, and the
// shape of the .code-workspace JSON document on disk.
package vscode

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/foreverfl/gitt/internal/daemon/client"
)

// Folder is one entry under the "folders" key of a .code-workspace file.
// The Name field controls how VSCode labels the folder in its sidebar and
// title bar; setting it per-branch is the whole point of this package, since
// every gitt worktree otherwise shows up as ".worktrees" or its safe-branch
// dir name and is hard to tell apart across multiple windows.
type Folder struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Folders fetches every worktree row from the daemon, keeps the ones that
// belong to mainRoot, and returns folder entries with paths relative to
// mainRoot so the workspace file is portable across machines.
func Folders(mainRoot string) ([]Folder, error) {
	worktrees, err := client.ListWorktrees()
	if err != nil {
		if errors.Is(err, client.ErrNotRunning) {
			return nil, fmt.Errorf("gitt daemon is not running. start it first: gitt on")
		}
		return nil, err
	}

	var folders []Folder
	for _, w := range worktrees {
		if w.RepoRoot != mainRoot {
			continue
		}
		path, err := filepath.Rel(mainRoot, w.WorktreePath)
		if err != nil {
			path = w.WorktreePath
		}
		folders = append(folders, Folder{Name: w.BranchName, Path: path})
	}
	sort.Slice(folders, func(i, j int) bool {
		return folders[i].Name < folders[j].Name
	})
	return folders, nil
}

// WriteWorkspace replaces only the "folders" key of an existing workspace
// file, preserving any user-edited "settings", "extensions", etc. When the
// file does not yet exist, it writes a minimal skeleton.
func WriteWorkspace(workspacePath string, folders []Folder) error {
	doc := map[string]any{}
	if existing, err := os.ReadFile(workspacePath); err == nil {
		if err := json.Unmarshal(existing, &doc); err != nil {
			return fmt.Errorf("parse existing %s: %w", filepath.Base(workspacePath), err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read %s: %w", filepath.Base(workspacePath), err)
	}

	doc["folders"] = folders
	if _, ok := doc["settings"]; !ok {
		doc["settings"] = map[string]any{}
	}

	buf, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("encode workspace: %w", err)
	}
	buf = append(buf, '\n')

	if err := os.WriteFile(workspacePath, buf, 0o644); err != nil {
		return fmt.Errorf("write workspace file: %w", err)
	}
	return nil
}
