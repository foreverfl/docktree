package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/foreverfl/gitt/internal/daemon"
	"github.com/foreverfl/gitt/internal/gitx"
	"github.com/foreverfl/gitt/internal/paths"
	"github.com/foreverfl/gitt/internal/worktree"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <branch>",
	Short: "Create a new git worktree for <branch>",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireDaemon(); err != nil {
			return err
		}
		branch := args[0]

		mainRoot, err := gitx.MainRepoRoot()
		if err != nil {
			return err
		}
		target := worktree.Path(mainRoot, branch)

		existingPath, err := gitx.WorktreeForBranch(branch)
		if err != nil {
			return err
		}
		if existingPath != "" {
			fmt.Printf("branch '%s' is already checked out\n  path:   %s\n", branch, existingPath)
			if err := registerWorktree(mainRoot, branch, existingPath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: daemon registration failed: %v\n", err)
			}
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create worktree parent: %w", err)
		}

		exists, err := gitx.BranchExists(branch)
		if err != nil {
			return err
		}

		if err := gitx.WorktreeAdd(target, branch, !exists); err != nil {
			return err
		}

		if err := registerWorktree(mainRoot, branch, target); err != nil {
			fmt.Fprintf(os.Stderr, "warning: worktree created but daemon registration failed: %v\n", err)
		}

		if exists {
			fmt.Printf("created worktree\n  path:   %s\n  branch: %s\n", target, branch)
		} else {
			fmt.Printf("created worktree (new branch)\n  path:   %s\n  branch: %s\n", target, branch)
		}
		fmt.Printf("\nOpen a new terminal, then run:\n  cd %s\n  # start your AI CLI here\n", target)
		return nil
	},
}

func registerWorktree(mainRoot, branch, target string) error {
	sockpath, err := paths.SockPath()
	if err != nil {
		return err
	}
	response, err := daemon.Call(sockpath, daemon.Request{
		Op: daemon.OpRegisterWorktree,
		Args: map[string]any{
			"repo_root":        mainRoot,
			"repo_name":        filepath.Base(mainRoot),
			"branch_name":      branch,
			"safe_branch_name": worktree.SafeBranch(branch),
			"worktree_path":    target,
		},
	})
	if err != nil {
		return err
	}
	if !response.OK {
		return fmt.Errorf("%s", response.Error)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(addCmd)
}
