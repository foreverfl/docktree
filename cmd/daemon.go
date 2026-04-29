package cmd

import (
	"github.com/foreverfl/gitt/internal/daemon/server"
	"github.com/foreverfl/gitt/internal/paths"
	"github.com/spf13/cobra"
)

// daemonRunCmd is the in-process entrypoint that `gitt on` fork-execs into.
// Hidden from --help: end users don't run this directly.
var daemonRunCmd = &cobra.Command{
	Use:    "daemon-run",
	Short:  "Internal: run the gitt daemon in foreground",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		sockpath, err := paths.SockPath()
		if err != nil {
			return err
		}
		dbpath, err := paths.DBPath()
		if err != nil {
			return err
		}
		return server.Run(sockpath, dbpath)
	},
}

func init() {
	rootCmd.AddCommand(daemonRunCmd)
}