package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/foreverfl/gitt/internal/daemon"
	"github.com/foreverfl/gitt/internal/paths"
	"github.com/foreverfl/gitt/internal/process"
	"github.com/foreverfl/gitt/internal/prompt"
	"github.com/foreverfl/gitt/internal/purge"
	"github.com/foreverfl/gitt/internal/release"
	"github.com/foreverfl/gitt/internal/version"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update gitt to the latest release (force-removes all registered worktrees)",
	Long: `Fetch and install the latest gitt release.

Before replacing the binary, update:
  1. Reads every registered worktree from ~/.gitt/gitt.db.
  2. Force-removes each worktree folder on disk (os.RemoveAll).
     Any uncommitted or untracked changes inside those folders will be lost.
     The safety checks that 'gitt remove' normally performs are bypassed.
  3. Runs 'git worktree prune' on each affected repository to drop orphaned
     admin records under .git/worktrees/<name>.
  4. Stops the running daemon (if any).
  5. Deletes ~/.gitt/ entirely (db, sock, pid, log).
  6. Replaces the binary and restarts the daemon.

Partial failures (folder removal, prune) are printed as warnings and do not
abort the update. Use -y/--yes to skip the confirmation prompt.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		current := version.Installed()

		fmt.Println("checking latest release...")
		latest, err := release.LatestTag()
		if err != nil {
			return err
		}

		if current != "" && current == latest {
			fmt.Printf("already at latest (%s)\n", current)
			return nil
		}
		if current == "" {
			fmt.Printf("updating to %s\n", latest)
		} else {
			fmt.Printf("updating %s -> %s\n", current, latest)
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
		logpath, err := paths.LogPath()
		if err != nil {
			return err
		}
		dbpath, err := paths.DBPath()
		if err != nil {
			return err
		}

		daemonRunning := false
		if pid, ok := process.ReadPid(pidpath); ok && process.Alive(pid) {
			if err := daemon.Ping(sockpath); err == nil {
				daemonRunning = true
			}
		}

		registered := purge.LoadRegistered(dbpath)

		fmt.Println("update will reset gitt runtime data:")
		fmt.Printf("  - %s (db, sock, pid, log)\n", runtime)
		if len(registered) > 0 {
			fmt.Printf("  - %d registered worktree folder(s) will be force-removed\n", len(registered))
			fmt.Println("    (any uncommitted or untracked changes inside them will be lost)")
		}
		if daemonRunning {
			fmt.Println("  - the running daemon will be stopped and restarted")
		}
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

		selfPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("locate self: %w", err)
		}
		newPath := selfPath + ".new"

		fmt.Println("downloading...")
		if err := release.Download(latest, newPath); err != nil {
			_ = os.Remove(newPath)
			return err
		}

		if daemonRunning {
			if err := daemon.Shutdown(sockpath, pidpath, os.Stdout, os.Stderr); err != nil {
				_ = os.Remove(newPath)
				return fmt.Errorf("stop daemon: %w", err)
			}
		}

		// Re-list after the daemon is down so no row gets created mid-update.
		final := purge.LoadRegistered(dbpath)
		purge.RemoveRegistered(final)

		if err := os.RemoveAll(runtime); err != nil {
			_ = os.Remove(newPath)
			return fmt.Errorf("reset %s: %w", runtime, err)
		}

		if err := os.Rename(newPath, selfPath); err != nil {
			_ = os.Remove(newPath)
			return fmt.Errorf("replace binary: %w", err)
		}

		if vpath, verr := paths.VersionPath(); verr == nil {
			_ = os.WriteFile(vpath, []byte(latest+"\n"), 0o644)
		}

		fmt.Printf("updated to %s\n", latest)

		if daemonRunning {
			fmt.Println("restarting daemon...")
			pid, err := daemon.Spawn(selfPath, sockpath, pidpath, logpath)
			if err != nil {
				return fmt.Errorf("restart daemon: %w", err)
			}
			fmt.Printf("gitt daemon started (pid=%d)\n", pid)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
