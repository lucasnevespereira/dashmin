package cmd

import (
	"fmt"
	"strings"

	"github.com/lucasnevespereira/dashmin/internal/ai"
	"github.com/lucasnevespereira/dashmin/internal/config"
	"github.com/spf13/cobra"
)

var (
	aiProvider string
	aiAPIKey   string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `Manage Dashmin configuration.

Examples:
  dashmin config show                              # Show current config
  dashmin config path                              # Show config file path
  dashmin config ai --provider openai --key sk-xx  # Configure AI`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Display the current configuration file contents.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		fmt.Printf("Config location: %s\n\n", config.GetConfigPath())

		if len(cfg.Apps) == 0 && cfg.AI == nil {
			fmt.Println("No configuration yet.")
			return
		}

		if cfg.AI != nil && cfg.AI.APIKey != "" {
			fmt.Println("ai:")
			fmt.Printf("  provider: %s\n", cfg.AI.Provider)
			maskedKey := cfg.AI.APIKey[:8] + "..." + cfg.AI.APIKey[len(cfg.AI.APIKey)-4:]
			fmt.Printf("  api_key: %s\n", maskedKey)
			fmt.Println()
		}

		if len(cfg.Apps) > 0 {
			fmt.Println("apps:")
			for name, app := range cfg.Apps {
				fmt.Printf("  %s:\n", name)
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
		}
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.GetConfigPath())
	},
}

var configAiCmd = &cobra.Command{
	Use:   "ai",
	Short: "Configure AI for natural language queries",
	Long: `Configure AI provider and API key for query generation.

Examples:
  dashmin config ai --provider openai --key sk-your-key
  dashmin config ai --provider anthropic --key your-key
  dashmin config ai status
  dashmin config ai reset`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if aiProvider != "" || aiAPIKey != "" {
			setAIConfiguration(cfg, aiProvider, aiAPIKey)
			return
		}

		showAIConfiguration(cfg)
	},
}

var configAiStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show AI configuration status",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		showAIConfiguration(cfg)
	},
}

var configAiResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset AI configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		resetAIConfiguration(cfg)
	},
}

func showAIConfiguration(cfg *config.Config) {
	fmt.Printf("AI Configuration\n")
	fmt.Printf("================\n\n")

	if cfg.AI == nil || cfg.AI.APIKey == "" {
		fmt.Printf("Status: Not configured\n")
		fmt.Printf("\nSetup:\n")
		fmt.Printf("  dashmin config ai --provider openai --key sk-your-key\n")
		return
	}

	fmt.Printf("Status: Configured\n")
	fmt.Printf("Provider: %s\n", cfg.AI.Provider)

	maskedKey := cfg.AI.APIKey[:8] + "..." + cfg.AI.APIKey[len(cfg.AI.APIKey)-4:]
	fmt.Printf("API Key: %s\n", maskedKey)

	fmt.Printf("\nUsage:\n")
	fmt.Printf("  dashmin query generate <app> \"<question>\"\n")
}

func setAIConfiguration(cfg *config.Config, provider, apiKey string) {
	if provider != "" {
		providers := ai.GetAvailableProviders()
		validProvider := false
		for _, p := range providers {
			if p == provider {
				validProvider = true
				break
			}
		}
		if !validProvider {
			fmt.Printf("Invalid provider. Available: %s\n", strings.Join(providers, ", "))
			return
		}
	}

	if cfg.AI == nil {
		cfg.AI = &config.AIConfig{}
	}

	if provider != "" {
		cfg.AI.Provider = provider
		fmt.Printf("Provider: %s\n", provider)
	}

	if apiKey != "" {
		cfg.AI.APIKey = apiKey
		fmt.Printf("API key: updated\n")
	}

	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving: %v\n", err)
		return
	}

	fmt.Printf("\nUsage:\n")
	fmt.Printf("  dashmin query generate <app> \"<question>\"\n")
}

func resetAIConfiguration(cfg *config.Config) {
	if cfg.AI == nil || cfg.AI.APIKey == "" {
		fmt.Printf("AI not configured.\n")
		return
	}

	cfg.AI = nil
	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving: %v\n", err)
		return
	}

	fmt.Printf("AI configuration removed.\n")
}

func maskConnection(conn string) string {
	if strings.Contains(conn, "://") {
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

func init() {
	configAiCmd.Flags().StringVar(&aiProvider, "provider", "", "AI provider (openai, anthropic)")
	configAiCmd.Flags().StringVar(&aiAPIKey, "key", "", "API key for the provider")

	configAiCmd.AddCommand(configAiStatusCmd)
	configAiCmd.AddCommand(configAiResetCmd)

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configAiCmd)
}
