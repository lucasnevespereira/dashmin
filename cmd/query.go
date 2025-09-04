package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/lucasnevespereira/dashmin/config"
)

var queryCmd = &cobra.Command{
	Use:   "query <app> <label> <query>",
	Short: "Add a custom query to an app",
	Long: `Add a custom query to monitor specific metrics for an app.

Examples:
  dashmin query blogbuddy users "SELECT COUNT(*) FROM users"
  dashmin query blogbuddy posts "SELECT COUNT(*) FROM posts WHERE created_at > NOW() - INTERVAL '30 days'"
  dashmin query webapp revenue "SELECT SUM(amount) FROM payments WHERE DATE(created_at) = CURDATE()"
  dashmin query analytics active_users "users.count({\"status\": \"active\"})"`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		label := args[1]
		query := args[2]

		// Load config
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		// Check if app exists
		app, exists := cfg.Apps[appName]
		if !exists {
			fmt.Printf("Error: App '%s' not found.\n", appName)
			fmt.Printf("Available apps: ")
			for name := range cfg.Apps {
				fmt.Printf("%s ", name)
			}
			fmt.Printf("\n")
			return
		}

		// Add query
		if app.Queries == nil {
			app.Queries = make(map[string]string)
		}
		app.Queries[label] = query
		cfg.Apps[appName] = app

		// Save config
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("✅ Added query '%s' to app '%s'\n", label, appName)
		fmt.Printf("Query: %s\n", query)
		fmt.Printf("\nView results: dashmin status\n")
	},
}