package cmd

import (
	"os"

	"github.com/foreverfl/gitt/internal/daemon/client"
	"github.com/foreverfl/gitt/internal/paths"
	"github.com/spf13/cobra"
)

var offCmd = &cobra.Command{
	Use:   "off",
	Short: "Stop the gitt daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		sockpath, err := paths.SockPath()
		if err != nil {
			return err
		}
		pidpath, err := paths.PidPath()
		if err != nil {
			return err
		}
		return client.Shutdown(sockpath, pidpath, os.Stdout, os.Stderr)
	},
}

func init() {
	rootCmd.AddCommand(offCmd)
}