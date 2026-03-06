package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load loads configuration from file and environment
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Load from file if exists
	configPath := getConfigPath()
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("invalid config: %w", err)
		}
	}

	return cfg, nil
}

func getConfigPath() string {
	// Check XDG_CONFIG_HOME first
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "lazylinear", "config.yaml")
	}
	// Fall back to ~/.config
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lazylinear", "config.yaml")
}
