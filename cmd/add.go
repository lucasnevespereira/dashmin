package cmd

import (
	"fmt"

	"github.com/lucasnevespereira/dashmin/internal/config"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <name> <type> <connection-string>",
	Short: "Add a new app to monitor",
	Long: `Add a new application to monitor with its database connection.

Examples:
  dashmin add blogbuddy postgres "postgres://readonly:password@localhost:5432/blogbuddy_prod?sslmode=disable"
  dashmin add webapp mysql "user:pass@tcp(localhost:3306)/webapp"
  dashmin add analytics mongodb "mongodb://user:pass@localhost:27017/analytics"`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		dbType := args[1]
		connection := args[2]

		// Validate database type
		validTypes := map[string]bool{
			"postgres": true,
			"mysql":    true,
			"mongodb":  true,
		}

		if !validTypes[dbType] {
			fmt.Printf("Error: Invalid database type '%s'. Supported: postgres, mysql, mongodb\n", dbType)
			return
		}

		// Load config
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		// Add app
		app := config.App{
			Name:       name,
			Type:       dbType,
			Connection: connection,
			Queries:    make(map[string]string),
		}

		// Add default queries based on database type
		switch dbType {
		case "postgres", "mysql":
			app.Queries["users"] = "SELECT COUNT(*) FROM users"
		case "mongodb":
			app.Queries["users"] = "users.count({})"
		}

		cfg.Apps[name] = app

		// Save config
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("✅ Added app '%s' (%s)\n", name, dbType)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  • Add custom queries: dashmin query %s <label> \"<sql>\"\n", name)
		fmt.Printf("  • View dashboard: dashmin all\n")
	},
}
