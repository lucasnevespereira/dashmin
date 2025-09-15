package cmd

import (
	"fmt"
	"strings"

	"github.com/lucasnevespereira/dashmin/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured apps",
	Long:  "Show all configured applications and their queries.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if len(cfg.Apps) == 0 {
			fmt.Println("No apps configured yet.")
			fmt.Println("\nAdd your first app:")
			fmt.Println("  dashmin add myapp postgres \"postgres://readonly:password@localhost:5432/myapp?sslmode=disable\"")
			return
		}

		fmt.Printf("📱 Configured Apps (%d):\n\n", len(cfg.Apps))

		for name, app := range cfg.Apps {
			fmt.Printf("▸ %s (%s)\n", name, app.Type)

			if len(app.Queries) > 0 {
				fmt.Printf("  Queries:\n")
				for label, query := range app.Queries {
					// Truncate long queries
					displayQuery := query
					if len(query) > 60 {
						displayQuery = query[:57] + "..."
					}
					fmt.Printf("    %s: %s\n", label, displayQuery)
				}
			} else {
				fmt.Printf("  No queries defined\n")
			}
			fmt.Printf("\n")
		}

		fmt.Printf("Commands:\n")
		fmt.Printf("  dashmin query <app> <label> \"<query>\"  # Add custom query\n")
		fmt.Printf("  dashmin all                            # View dashboard\n")
		fmt.Printf("  dashmin remove <app>                   # Remove app\n")
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove <app>",
	Short: "Remove an app",
	Long:  "Remove an application from monitoring.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if _, exists := cfg.Apps[appName]; !exists {
			fmt.Printf("Error: App '%s' not found.\n", appName)
			return
		}

		delete(cfg.Apps, appName)

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("✅ Removed app '%s'\n", appName)
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	Long:  "Display the current configuration file contents.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		fmt.Printf("📁 Config location: %s\n\n", config.GetConfigPath())

		if len(cfg.Apps) == 0 {
			fmt.Println("No apps configured.")
			return
		}

		// Show config in a readable format
		fmt.Println("apps:")
		for name, app := range cfg.Apps {
			fmt.Printf("  %s:\n", name)
			fmt.Printf("    name: %s\n", app.Name)
			fmt.Printf("    type: %s\n", app.Type)
			fmt.Printf("    connection: %s\n", maskConnection(app.Connection))
			if len(app.Queries) > 0 {
				fmt.Printf("    queries:\n")
				for label, query := range app.Queries {
					fmt.Printf("      %s: %s\n", label, query)
				}
			}
			fmt.Println()
		}
	},
}

func maskConnection(conn string) string {
	// Mask passwords in connection strings
	if strings.Contains(conn, "password=") {
		parts := strings.Split(conn, " ")
		for i, part := range parts {
			if strings.HasPrefix(part, "password=") {
				parts[i] = "password=***"
			}
		}
		return strings.Join(parts, " ")
	}

	if strings.Contains(conn, ":") && strings.Contains(conn, "@") {
		// Handle user:pass@host format
		parts := strings.Split(conn, "@")
		if len(parts) > 1 {
			userPass := strings.Split(parts[0], ":")
			if len(userPass) > 1 {
				return userPass[0] + ":***@" + strings.Join(parts[1:], "@")
			}
		}
	}

	return conn
}
