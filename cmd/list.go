package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/foreverfl/gitt/internal/daemon/client"
	"github.com/foreverfl/gitt/internal/gitx"
	"github.com/foreverfl/gitt/internal/store/repo"
	"github.com/spf13/cobra"
)

var listGlobal bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List gitt-managed worktrees for the current repo",
	Long: "By default lists worktrees belonging to the repository that\n" +
		"contains the current working directory. Pass --global / -g to\n" +
		"list every worktree the daemon knows about, across all repos.",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireDaemon(); err != nil {
			return err
		}

		var (
			worktrees []repo.Worktree
			err       error
		)
		if listGlobal {
			worktrees, err = client.ListWorktrees()
		} else {
			mainRoot, mainErr := gitx.MainRepoRoot()
			if mainErr != nil {
				return fmt.Errorf("%w (or pass --global to list every repo)", mainErr)
			}
			worktrees, err = client.ListWorktreesForRepo(mainRoot)
		}
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		if len(worktrees) == 0 {
			if listGlobal {
				fmt.Fprintln(out, "(no worktrees registered)")
			} else {
				fmt.Fprintln(out, "(no worktrees registered for this repo)")
			}
			return nil
		}

		writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		if listGlobal {
			fmt.Fprintln(writer, "REPO\tBRANCH\tPROTECTED\tSTATUS\tPATH")
			for _, worktree := range worktrees {
				fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
					worktree.RepoName, worktree.BranchName, protectedMark(worktree.IsProtected), worktree.Status, worktree.WorktreePath)
			}
		} else {
			fmt.Fprintln(writer, "BRANCH\tPROTECTED\tSTATUS\tPATH")
			for _, worktree := range worktrees {
				fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
					worktree.BranchName, protectedMark(worktree.IsProtected), worktree.Status, worktree.WorktreePath)
			}
		}
		return writer.Flush()
	},
}

func init() {
	listCmd.Flags().BoolVarP(&listGlobal, "global", "g", false, "list worktrees across all repos")
	rootCmd.AddCommand(listCmd)
}

// protectedMark renders the PROTECTED column cell. A single "*" keeps the
// column slim so the existing BRANCH/STATUS/PATH layout barely shifts; the
// blank case stays empty rather than "no" / "-" for the same reason.
func protectedMark(isProtected bool) string {
	if isProtected {
		return "*"
	}
	return ""
}
