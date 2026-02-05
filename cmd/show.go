package cmd

import (
	"fmt"
	"os"

	"github.com/lucasnevespereira/dashmin/internal/config"
	"github.com/lucasnevespereira/dashmin/ui"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show [app]",
	Short: "Show the dashboard",
	Long: `Display the real-time dashboard with metrics.

Without arguments, shows all configured apps.
With an app name, shows only that specific app.

Examples:
  dashmin show           # Show all apps
  dashmin show myapp     # Show specific app`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		if len(cfg.Apps) == 0 {
			fmt.Println("No apps configured yet.")
			fmt.Println("\nQuick start:")
			fmt.Println("  dashmin app add myapp postgres \"postgres://readonly:password@localhost:5432/myapp?sslmode=disable\"")
			fmt.Println("  dashmin query add myapp users \"SELECT COUNT(*) FROM users\"")
			fmt.Println("  dashmin show")
			return
		}

		appFilter := ""
		if len(args) == 1 {
			appFilter = args[0]
			if _, exists := cfg.Apps[appFilter]; !exists {
				fmt.Printf("App '%s' not found.\n", appFilter)
				fmt.Println("\nAvailable apps:")
				for name := range cfg.Apps {
					fmt.Printf("  %s\n", name)
				}
				return
			}
		}

		if err := ui.RunDashboard(cfg, appFilter); err != nil {
			fmt.Printf("Error running dashboard: %v\n", err)
			os.Exit(1)
		}
	},
}
