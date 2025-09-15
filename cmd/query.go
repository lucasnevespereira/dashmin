package cmd

import (
	"fmt"

	"github.com/lucasnevespereira/dashmin/internal/config"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query <app> <label> <query>",
	Short: "Manage queries for an app",
	Long: `Add or remove custom queries to monitor specific metrics for an app.

Examples:
  dashmin query blogbuddy users "SELECT COUNT(*) FROM users"
  dashmin query blogbuddy posts "SELECT COUNT(*) FROM posts WHERE created_at > NOW() - INTERVAL '30 days'"
  dashmin query webapp revenue "SELECT SUM(amount) FROM payments WHERE DATE(created_at) = CURDATE()"
  dashmin query analytics active_users "users.count({\"status\": \"active\"})"
  dashmin query remove blogbuddy users`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 3 {
			fmt.Printf("❌ Invalid usage. Use: dashmin query <app> <label> <query>\n")
			fmt.Printf("Or use subcommands: dashmin query remove <app> <label>\n")
			return
		}

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
		fmt.Printf("\nView results: dashmin all\n")
	},
}

var queryRemoveCmd = &cobra.Command{
	Use:   "remove <app> <label>",
	Short: "Remove a query from an app",
	Long: `Remove a custom query from an app.

Examples:
  dashmin query remove blogbuddy users
  dashmin query remove webapp revenue`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		label := args[1]

		// Load config
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		// Check if app exists
		app, exists := cfg.Apps[appName]
		if !exists {
			fmt.Printf("❌ App '%s' not found.\n", appName)
			fmt.Printf("Available apps: ")
			for name := range cfg.Apps {
				fmt.Printf("%s ", name)
			}
			fmt.Printf("\n")
			return
		}

		// Check if query exists
		if app.Queries == nil {
			fmt.Printf("❌ App '%s' has no queries.\n", appName)
			return
		}

		querySQL, queryExists := app.Queries[label]
		if !queryExists {
			fmt.Printf("❌ Query '%s' not found in app '%s'.\n", label, appName)
			fmt.Printf("Available queries: ")
			for queryLabel := range app.Queries {
				fmt.Printf("%s ", queryLabel)
			}
			fmt.Printf("\n")
			return
		}

		// Remove query
		delete(app.Queries, label)
		cfg.Apps[appName] = app

		// Save config
		if err := cfg.Save(); err != nil {
			fmt.Printf("❌ Error saving config: %v\n", err)
			return
		}

		fmt.Printf("✅ Removed query '%s' from app '%s'\n", label, appName)
		fmt.Printf("Query was: %s\n", querySQL)
	},
}

func init() {
	queryCmd.AddCommand(queryRemoveCmd)
}
