package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/foreverfl/doctree/internal/gitx"
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

		repoRoot, err := gitx.RepoRoot()
		if err != nil {
			return err
		}
		repoName := filepath.Base(repoRoot)
		target := filepath.Join(filepath.Dir(repoRoot), ".worktrees", repoName, safeBranch(branch))

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create worktree parent: %w", err)
		}

		if err := gitx.WorktreeAdd(target, branch); err != nil {
			return err
		}

		fmt.Printf("created worktree\n  path:   %s\n  branch: %s\n", target, branch)
		return nil
	},
}

// safeBranch turns a git branch name into a single path segment by replacing
// directory separators with dashes. e.g. "feature/foo" -> "feature-foo".
func safeBranch(branch string) string {
	return strings.NewReplacer("/", "-", "\\", "-").Replace(branch)
}

func init() {
	rootCmd.AddCommand(addCmd)
}
