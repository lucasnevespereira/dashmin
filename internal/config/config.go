package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type App struct {
	Name       string            `yaml:"name"`
	Type       string            `yaml:"type"` // postgres, mysql, mongodb
	Connection string            `yaml:"connection"`
	Queries    map[string]string `yaml:"queries"`
}

type AIConfig struct {
	Provider string `yaml:"provider,omitempty"` // openai, anthropic
	APIKey   string `yaml:"api_key,omitempty"`
}

type Config struct {
	AI   *AIConfig       `yaml:"ai,omitempty"`
	Apps map[string]App `yaml:"apps"`
}

func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "dashmin", "config.yaml")
}

func EnsureConfigDir() error {
	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)
	return os.MkdirAll(configDir, 0755)
}

func Load() (*Config, error) {
	configPath := GetConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{Apps: make(map[string]App)}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if config.Apps == nil {
		config.Apps = make(map[string]App)
	}

	return &config, nil
}

func (c *Config) Save() error {
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := GetConfigPath()
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
