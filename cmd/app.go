package cmd

import (
	"fmt"

	"github.com/lucasnevespereira/dashmin/internal/config"
	"github.com/lucasnevespereira/dashmin/internal/db"
	"github.com/spf13/cobra"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage apps",
	Long: `Manage applications to monitor.

Examples:
  dashmin app add myapp postgres "postgres://user:pass@host/db"
  dashmin app list
  dashmin app test myapp
  dashmin app remove myapp`,
}

var appAddCmd = &cobra.Command{
	Use:   "add <name> <type> <connection-string>",
	Short: "Add a new app to monitor",
	Long: `Add a new application to monitor with its database connection.

Supported database types: postgres, mysql, mongodb

Examples:
  dashmin app add myapp postgres "postgres://readonly:password@localhost:5432/myapp?sslmode=disable"
  dashmin app add webapp mysql "user:pass@tcp(localhost:3306)/webapp"
  dashmin app add analytics mongodb "mongodb://user:pass@localhost:27017/analytics"`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		dbType := args[1]
		connection := args[2]

		validTypes := map[string]bool{
			"postgres": true,
			"mysql":    true,
			"mongodb":  true,
		}

		if !validTypes[dbType] {
			return fmt.Errorf("invalid database type '%s'. Supported: postgres, mysql, mongodb", dbType)
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		app := config.App{
			Name:       name,
			Type:       dbType,
			Connection: connection,
			Queries:    make(map[string]string),
		}

		switch dbType {
		case "postgres", "mysql":
			app.Queries["users"] = "SELECT COUNT(*) FROM users"
		case "mongodb":
			app.Queries["users"] = "users.count({})"
		}

		// Validate connection
		fmt.Printf("Testing connection to '%s'...\n", name)
		conn, err := db.ConnectByType(dbType, connection)
		if err != nil {
			fmt.Printf("Warning: Connection failed: %v\n", err)
			fmt.Printf("App saved anyway. Fix the connection string and test with:\n")
			fmt.Printf("  dashmin app test %s\n\n", name)
		} else {
			_ = conn.Close()
			fmt.Printf("Connection successful!\n\n")
		}

		cfg.Apps[name] = app

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("Added app '%s' (%s)\n", name, dbType)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  dashmin query add %s <label> \"<query>\"  # Add custom query\n", name)
		fmt.Printf("  dashmin show                             # View dashboard\n")
		return nil
	},
}

var appRemoveCmd = &cobra.Command{
	Use:   "remove <app>",
	Short: "Remove an app",
	Long: `Remove an application from monitoring.

Examples:
  dashmin app remove myapp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if _, exists := cfg.Apps[appName]; !exists {
			fmt.Printf("Error: App '%s' not found.\n", appName)
			if len(cfg.Apps) > 0 {
				fmt.Printf("Available apps: ")
				for name := range cfg.Apps {
					fmt.Printf("%s ", name)
				}
				fmt.Printf("\n")
			}
			return fmt.Errorf("app '%s' not found", appName)
		}

		delete(cfg.Apps, appName)

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("Removed app '%s'\n", appName)
		return nil
	},
}

var appListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured apps",
	Long:  "Show all configured applications and their queries.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if len(cfg.Apps) == 0 {
			fmt.Println("No apps configured yet.")
			fmt.Println("\nAdd your first app:")
			fmt.Println("  dashmin app add myapp postgres \"postgres://readonly:password@localhost:5432/myapp?sslmode=disable\"")
			return nil
		}

		fmt.Printf("Configured Apps (%d):\n\n", len(cfg.Apps))

		for name, app := range cfg.Apps {
			fmt.Printf("  %s (%s)\n", name, app.Type)

			if len(app.Queries) > 0 {
				fmt.Printf("    Queries:\n")
				for label, query := range app.Queries {
					displayQuery := query
					if len(query) > 60 {
						displayQuery = query[:57] + "..."
					}
					fmt.Printf("      %s: %s\n", label, displayQuery)
				}
			} else {
				fmt.Printf("    No queries defined\n")
			}
			fmt.Printf("\n")
		}
		return nil
	},
}

var appTestCmd = &cobra.Command{
	Use:   "test <app>",
	Short: "Test database connection for an app",
	Long: `Test the database connection for a specific app to help debug connection issues.

Examples:
  dashmin app test myapp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		app, exists := cfg.Apps[appName]
		if !exists {
			fmt.Printf("Error: App '%s' not found.\n", appName)
			if len(cfg.Apps) > 0 {
				fmt.Printf("Available apps: ")
				for name := range cfg.Apps {
					fmt.Printf("%s ", name)
				}
				fmt.Printf("\n")
			}
			return fmt.Errorf("app '%s' not found", appName)
		}

		fmt.Printf("Testing connection to '%s' (%s)...\n", appName, app.Type)
		fmt.Printf("Connection: %s\n\n", maskConnection(app.Connection))

		conn, err := db.ConnectByType(app.Type, app.Connection)
		if err != nil {
			fmt.Printf("Troubleshooting tips:\n")
			fmt.Printf("  - Check if the database server is running\n")
			fmt.Printf("  - Verify the connection string format:\n")
			switch app.Type {
			case "postgres":
				fmt.Printf("    postgres://user:password@host:port/database?sslmode=disable\n")
			case "mysql":
				fmt.Printf("    user:password@tcp(host:port)/database\n")
			case "mongodb":
				fmt.Printf("    mongodb://user:password@host:port/database\n")
			}
			fmt.Printf("  - Check network connectivity\n")
			fmt.Printf("  - Verify credentials are correct\n")
			return fmt.Errorf("connection failed: %w", err)
		}
		defer func() { _ = conn.Close() }()

		fmt.Printf("Connection successful!\n\n")

		fmt.Printf("Testing basic query...\n")
		var testQuery string
		switch app.Type {
		case "postgres", "mysql":
			testQuery = "SELECT 1 as test"
		case "mongodb":
			testQuery = "test.count({})"
		}

		result, err := conn.Query(testQuery)
		if err != nil {
			return fmt.Errorf("query failed: %w", err)
		}

		if result.Error != nil {
			return fmt.Errorf("query error: %w", result.Error)
		}

		fmt.Printf("Query executed successfully!\n")
		if len(result.Rows) > 0 && len(result.Rows[0]) > 0 {
			fmt.Printf("Result: %v\n", result.Rows[0][0])
		}

		fmt.Printf("\nConnection and basic queries working correctly!\n")
		fmt.Printf("You can now add custom queries with:\n")
		fmt.Printf("  dashmin query add %s <label> \"<your-query>\"\n", appName)
		return nil
	},
}

func init() {
	appCmd.AddCommand(appAddCmd)
	appCmd.AddCommand(appRemoveCmd)
	appCmd.AddCommand(appListCmd)
	appCmd.AddCommand(appTestCmd)
}
