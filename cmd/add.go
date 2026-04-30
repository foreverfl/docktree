package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/foreverfl/gitt/internal/daemon/client"
	"github.com/foreverfl/gitt/internal/gitx"
	"github.com/foreverfl/gitt/internal/vscode"
	"github.com/spf13/cobra"
)

var addPrintPath bool

var addCmd = &cobra.Command{
	Use:   "add <branch>",
	Short: "Create a new git worktree for <branch>",
	Long: "Creates a worktree for <branch> under <repo>/.worktrees/<safe-name>\n" +
		"and registers it with the daemon.\n\n" +
		"With --print-path, all human-readable output is sent to stderr and\n" +
		"the worktree path is written as a single line to stdout. Intended\n" +
		"for shell wrappers that capture the path to `cd` into the new worktree.\n\n" +
		"Requires `gitt on` to be running.",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireDaemon(); err != nil {
			return err
		}
		branch := args[0]

		var info io.Writer = cmd.OutOrStdout()
		if addPrintPath {
			info = cmd.ErrOrStderr()
		}

		mainRoot, err := gitx.MainRepoRoot()
		if err != nil {
			return err
		}
		target := gitx.WorktreePath(mainRoot, branch)

		existingPath, err := gitx.WorktreeForBranch(branch)
		if err != nil {
			return err
		}
		if existingPath != "" {
			fmt.Fprintf(info, "branch '%s' is already checked out\n  path:   %s\n", branch, existingPath)
			if err := client.RegisterWorktree(mainRoot, branch, existingPath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: daemon registration failed: %v\n", err)
			}
			
			vscode.Sync(mainRoot, info)
			if addPrintPath {
				fmt.Fprintln(cmd.OutOrStdout(), existingPath)
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

		if err := gitx.WorktreeAdd("", target, branch, !exists); err != nil {
			return err
		}

		if err := client.RegisterWorktree(mainRoot, branch, target); err != nil {
			fmt.Fprintf(os.Stderr, "warning: worktree created but daemon registration failed: %v\n", err)
		}

		if exists {
			fmt.Fprintf(info, "created worktree\n  path:   %s\n  branch: %s\n", target, branch)
		} else {
			fmt.Fprintf(info, "created worktree (new branch)\n  path:   %s\n  branch: %s\n", target, branch)
		}
		vscode.Sync(mainRoot, info)
		fmt.Fprintf(info, "\nOpen a new terminal, then run:\n  cd %s\n  # start your AI CLI here\n", target)

		if addPrintPath {
			fmt.Fprintln(cmd.OutOrStdout(), target)
		}
		return nil
	},
}

func init() {
	addCmd.Flags().BoolVar(&addPrintPath, "print-path", false, "print only the worktree path to stdout (for shell wrappers)")
	rootCmd.AddCommand(addCmd)
}