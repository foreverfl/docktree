package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/foreverfl/gitt/internal/gitx"
	"github.com/foreverfl/gitt/internal/vscode"
	"github.com/spf13/cobra"
)

var vscodeCmd = &cobra.Command{
	Use:   "vscode",
	Short: "Generate a VSCode multi-root workspace file from registered worktrees",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireDaemon(); err != nil {
			return err
		}
		mainRoot, err := gitx.MainRepoRoot()
		if err != nil {
			return err
		}

		folders, err := vscode.Folders(mainRoot)
		if err != nil {
			return err
		}
		if len(folders) == 0 {
			return fmt.Errorf("no worktrees registered for %s", mainRoot)
		}

		repoName := filepath.Base(mainRoot)
		workspacePath := filepath.Join(mainRoot, repoName+".code-workspace")

		if err := vscode.WriteWorkspace(workspacePath, folders); err != nil {
			return err
		}

		fmt.Printf("wrote VSCode workspace\n  path:    %s\n  folders: %d\n", workspacePath, len(folders))
		fmt.Printf("\nOpen it:\n  code %s\n", workspacePath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(vscodeCmd)
}
