package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/foreverfl/gitt/internal/daemon/client"
	"github.com/foreverfl/gitt/internal/paths"
	"github.com/foreverfl/gitt/internal/prompt"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the gitt binary and all runtime data",
	Long:  "Stops the daemon, deletes ~/.gitt (sock, pid, log, db), then removes the binary itself.",
	RunE: func(cmd *cobra.Command, args []string) error {
		binpath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolve binary path: %w", err)
		}
		runtime, err := paths.RuntimeDir()
		if err != nil {
			return err
		}
		sockpath, err := paths.SockPath()
		if err != nil {
			return err
		}
		pidpath, err := paths.PidPath()
		if err != nil {
			return err
		}

		fmt.Println("gitt uninstall will remove:")
		fmt.Printf("  - runtime: %s (sock, pid, log, db)\n", runtime)
		fmt.Printf("  - binary:  %s\n", binpath)
		fmt.Println()

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			ok, err := prompt.Confirm("proceed?", false)
			if err != nil {
				if errors.Is(err, prompt.ErrNoTTY) {
					return fmt.Errorf("non-interactive shell — pass --yes to confirm")
				}
				return err
			}
			if !ok {
				fmt.Println("aborted.")
				return nil
			}
		}

		if err := client.Shutdown(sockpath, pidpath, os.Stdout, os.Stderr); err != nil {
			return fmt.Errorf("stop daemon: %w", err)
		}

		if err := os.RemoveAll(runtime); err != nil {
			return fmt.Errorf("remove %s: %w", runtime, err)
		}
		fmt.Printf("removed %s\n", runtime)

		if err := os.Remove(binpath); err != nil {
			return fmt.Errorf("remove %s: %w", binpath, err)
		}
		fmt.Printf("removed %s\n", binpath)
		fmt.Println("gitt uninstalled.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}