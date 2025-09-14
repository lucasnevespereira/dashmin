package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/lucasnevespereira/dashmin/config"
	"github.com/lucasnevespereira/dashmin/ui"
)

var testCmd = &cobra.Command{
	Use:   "test <app>",
	Short: "Test database connection for an app",
	Long:  "Test the database connection for a specific app to help debug connection issues.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

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

		fmt.Printf("🧪 Testing connection to '%s' (%s)...\n", appName, app.Type)
		fmt.Printf("Connection: %s\n\n", maskConnectionString(app.Connection))

		// Test connection
		conn, err := ui.ConnectDatabase(app)
		if err != nil {
			fmt.Printf("❌ Connection failed: %v\n\n", err)
			
			fmt.Printf("💡 Troubleshooting tips:\n")
			fmt.Printf("• Check if the database server is running\n")
			fmt.Printf("• Verify the connection string format:\n")
			switch app.Type {
			case "postgres":
				fmt.Printf("  postgres://user:password@host:port/database?sslmode=disable\n")
			case "mysql":
				fmt.Printf("  user:password@tcp(host:port)/database\n")
			case "mongodb":
				fmt.Printf("  mongodb://user:password@host:port/database\n")
			}
			fmt.Printf("• Check network connectivity (ping, telnet, etc.)\n")
			fmt.Printf("• Verify credentials are correct\n")
			fmt.Printf("• Check firewall settings\n")
			return
		}
		defer func() { _ = conn.Close() }()

		fmt.Printf("✅ Connection successful!\n\n")
		
		// Test a simple query
		fmt.Printf("🔍 Testing basic query...\n")
		var testQuery string
		switch app.Type {
		case "postgres", "mysql":
			testQuery = "SELECT 1 as test"
		case "mongodb":
			testQuery = "test.count({})"
		}

		result, err := conn.Query(testQuery)
		if err != nil {
			fmt.Printf("❌ Query failed: %v\n", err)
			return
		}

		if result.Error != nil {
			fmt.Printf("❌ Query error: %v\n", result.Error)
			return
		}

		fmt.Printf("✅ Query executed successfully!\n")
		if len(result.Rows) > 0 && len(result.Rows[0]) > 0 {
			fmt.Printf("Result: %v\n", result.Rows[0][0])
		}

		fmt.Printf("\n🎉 Connection and basic queries working correctly!\n")
		fmt.Printf("You can now add custom queries with:\n")
		fmt.Printf("  dashmin query %s <label> \"<your-sql-query>\"\n", appName)
	},
}

func maskConnectionString(conn string) string {
	if strings.Contains(conn, "://") {
		// URL format: postgres://user:pass@host:port/db
		parts := strings.SplitN(conn, "://", 2)
		if len(parts) == 2 {
			scheme := parts[0]
			rest := parts[1]
			
			if strings.Contains(rest, "@") {
				atParts := strings.SplitN(rest, "@", 2)
				userPass := atParts[0]
				hostDb := atParts[1]
				
				if strings.Contains(userPass, ":") {
					userPassParts := strings.SplitN(userPass, ":", 2)
					return fmt.Sprintf("%s://%s:***@%s", scheme, userPassParts[0], hostDb)
				}
			}
		}
	} else if strings.Contains(conn, ":") && strings.Contains(conn, "@") {
		// MySQL format: user:pass@tcp(host:port)/db
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