package cmd

import (
	"fmt"

	"github.com/foreverfl/gitt/internal/daemon/client"
	"github.com/spf13/cobra"
)

var sqliteCmd = &cobra.Command{
	Use:   "sqlite",
	Short: "Run a SQLite self-test against the daemon's database",
	Long: "Asks the running gitt daemon to create a scratch table, insert a\n" +
		"row, read it back, and drop the table. Useful for confirming the\n" +
		"daemon's database connection is healthy.\n\n" +
		"Requires `gitt on` to be running.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireDaemon(); err != nil {
			return err
		}
		message, err := client.SqliteTest()
		if err != nil {
			return fmt.Errorf("sqlite test failed: %w", err)
		}
		fmt.Println(message)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sqliteCmd)
}
