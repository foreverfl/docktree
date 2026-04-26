package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/foreverfl/gitt/internal/gitx"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current worktree's repo, branch, path, and clean/dirty state",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := gitx.RepoRoot()
		if err != nil {
			return err
		}
		mainRoot, err := gitx.MainRepoRoot()
		if err != nil {
			return err
		}
		branch, err := gitx.CurrentBranch()
		if err != nil {
			return err
		}
		clean, err := gitx.IsClean()
		if err != nil {
			return err
		}
		op, err := gitx.OngoingOp()
		if err != nil {
			return err
		}
		conflicts, err := gitx.HasConflicts()
		if err != nil {
			return err
		}

		if branch == "" {
			branch = "(detached)"
		}

		var parts []string
		if op != "" {
			parts = append(parts, op)
		}
		if conflicts {
			parts = append(parts, "conflict")
		}
		if len(parts) == 0 {
			if clean {
				parts = append(parts, "clean")
			} else {
				parts = append(parts, "dirty")
			}
		}

		fmt.Printf("Repo: %s\n", filepath.Base(mainRoot))
		fmt.Printf("Worktree: %s\n", filepath.Base(repoRoot))
		fmt.Printf("Branch: %s\n", branch)
		fmt.Printf("Path: %s\n", repoRoot)
		fmt.Printf("State: %s\n", strings.Join(parts, ", "))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
