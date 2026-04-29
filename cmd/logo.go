package cmd

import (
	"os"

	"github.com/foreverfl/gitt/internal/ui"
	"github.com/spf13/cobra"
)

var logoCmd = &cobra.Command{
	Use:   "logo",
	Short: "Print the gitt logo art in a sky-blue box",
	Run: func(cmd *cobra.Command, args []string) {
		ui.Logo(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(logoCmd)
}