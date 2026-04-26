package cmd

import (
	"os"

	"github.com/foreverfl/doctree/internal/banner"
	"github.com/spf13/cobra"
)

var logoCmd = &cobra.Command{
	Use:   "logo",
	Short: "Print the doctree logo art in a sky-blue box",
	Run: func(cmd *cobra.Command, args []string) {
		banner.PrintLogo(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(logoCmd)
}