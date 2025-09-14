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
  dashmin all`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help by default
		cmd.Help()
	},
}

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Show dashboard for all apps",
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
			fmt.Println("  dashmin all")
			return
		}

		// Launch TUI dashboard
		if err := ui.RunDashboard(cfg, ""); err != nil {
			fmt.Printf("Error running dashboard: %v\n", err)
			os.Exit(1)
		}
	},
}

var seeCmd = &cobra.Command{
	Use:   "see [app]",
	Short: "Show dashboard for a specific app",
	Long:  "Display the real-time dashboard for a single app and its metrics.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		appName := args[0]
		if _, exists := cfg.Apps[appName]; !exists {
			fmt.Printf("App '%s' not found.\n", appName)
			fmt.Println("\nAvailable apps:")
			for name := range cfg.Apps {
				fmt.Printf("  %s\n", name)
			}
			return
		}

		// Launch TUI dashboard for specific app
		if err := ui.RunDashboard(cfg, appName); err != nil {
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
	rootCmd.AddCommand(allCmd)
	rootCmd.AddCommand(seeCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(testCmd)
}