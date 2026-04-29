package cmd

import (
	"errors"
	"fmt"

	"github.com/foreverfl/gitt/internal/daemon/client"
	"github.com/foreverfl/gitt/internal/paths"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "gitt",
	Short:        "gitt — git worktree + docker compose orchestrator",
	Long:         "Coordinates per-branch git worktrees and their docker compose stacks via a small SQLite-backed daemon.",
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "skip confirmation prompts")
}

func Execute() error {
	return rootCmd.Execute()
}

// requireDaemon errors out with an init hint when the daemon isn't reachable.
func requireDaemon() error {
	sockpath, err := paths.SockPath()
	if err != nil {
		return err
	}
	if err := client.Ping(sockpath); err != nil {
		if errors.Is(err, client.ErrNotRunning) {
			return fmt.Errorf("gitt daemon is not running. start it first: gitt on")
		}
		return err
	}
	return nil
}