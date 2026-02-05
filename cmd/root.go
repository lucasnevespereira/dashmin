package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dashmin",
	Short: "Minimal terminal dashboard for your apps",
	Long: `Monitor your apps from the terminal.

Examples:
  dashmin app add myapp postgres "postgres://user:pass@host/db"
  dashmin query add myapp users "SELECT COUNT(*) FROM users"
  dashmin show`,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(appCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}
