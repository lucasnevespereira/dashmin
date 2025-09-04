package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/lucasnevespereira/dashmin/config"
	"github.com/lucasnevespereira/dashmin/ui"
)

var rootCmd = &cobra.Command{
	Use:   "dashmin",
	Short: "Minimal dashboard for your apps",
	Long: `A lightweight CLI tool to monitor multiple databases and applications from one place.
Built for developers who want quick insights without the overhead.

Examples:
  dashmin add myapp postgres "postgres://readonly:password@localhost:5432/myapp?sslmode=disable"
  dashmin query myapp users "SELECT COUNT(*) FROM users"
  dashmin status`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help by default
		cmd.Help()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show interactive dashboard",
	Long:  "Display the real-time dashboard with all your apps and metrics.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		if len(cfg.Apps) == 0 {
			fmt.Println("No apps configured yet.")
			fmt.Println("\nQuick start:")
			fmt.Println("  dashmin add myapp postgres \"postgres://readonly:password@localhost:5432/myapp?sslmode=disable\"")
			fmt.Println("  dashmin query myapp users \"SELECT COUNT(*) FROM users\"")
			fmt.Println("  dashmin status")
			return
		}

		// Launch TUI dashboard
		if err := ui.RunDashboard(cfg); err != nil {
			fmt.Printf("Error running dashboard: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(testCmd)
}