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

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "Configure AI for natural language queries",
	Long: `Configure AI provider and API key for query generation.

Examples:
  dashmin ai --provider openai --key sk-your-key-here
  dashmin ai --provider anthropic --key your-key
  dashmin ai status
  dashmin ai reset`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		// Set provider and key
		if aiProvider != "" || aiAPIKey != "" {
			setAIConfig(cfg, aiProvider, aiAPIKey)
			return
		}

		// Default: show current status
		showAIStatus(cfg)
	},
}

var aiStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show AI configuration status",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		showAIStatus(cfg)
	},
}

var aiResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset AI configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		resetAIConfig(cfg)
	},
}

func showAIStatus(cfg *config.Config) {
	fmt.Printf("🤖 AI Configuration\n")
	fmt.Printf("═══════════════════\n")

	if cfg.AI == nil || cfg.AI.APIKey == "" {
		fmt.Printf("Status: ❌ Not configured\n")
		fmt.Printf("\nSetup: dashmin ai --provider openai --key sk-your-key\n")
		return
	}

	fmt.Printf("Status: ✅ Configured\n")
	fmt.Printf("Provider: %s\n", cfg.AI.Provider)

	maskedKey := cfg.AI.APIKey[:8] + "..." + cfg.AI.APIKey[len(cfg.AI.APIKey)-4:]
	fmt.Printf("API Key: %s\n", maskedKey)

	fmt.Printf("\n🚀 Ready: dashmin prompt <app> \"<question>\"\n")
}

func setAIConfig(cfg *config.Config, provider, apiKey string) {
	// Validate provider if provided
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
			fmt.Printf("❌ Invalid provider. Available: %s\n", strings.Join(providers, ", "))
			return
		}
	}

	// Initialize AI config if needed
	if cfg.AI == nil {
		cfg.AI = &config.AIConfig{}
	}

	// Update provider
	if provider != "" {
		cfg.AI.Provider = provider
		fmt.Printf("✅ Provider: %s\n", provider)
	}

	// Update API key
	if apiKey != "" {
		cfg.AI.APIKey = apiKey
		fmt.Printf("✅ API key: updated\n")
	}

	// Save config
	if err := cfg.Save(); err != nil {
		fmt.Printf("❌ Error saving: %v\n", err)
		return
	}

	fmt.Printf("\n🚀 Try: dashmin prompt <app> \"<question>\"\n")
}

func resetAIConfig(cfg *config.Config) {
	if cfg.AI == nil || cfg.AI.APIKey == "" {
		fmt.Printf("AI not configured.\n")
		return
	}

	cfg.AI = nil
	if err := cfg.Save(); err != nil {
		fmt.Printf("❌ Error saving: %v\n", err)
		return
	}

	fmt.Printf("✅ AI configuration removed.\n")
}

func init() {
	aiCmd.Flags().StringVar(&aiProvider, "provider", "", "AI provider (openai, anthropic)")
	aiCmd.Flags().StringVar(&aiAPIKey, "key", "", "API key for the provider")

	aiCmd.AddCommand(aiStatusCmd)
	aiCmd.AddCommand(aiResetCmd)
}
