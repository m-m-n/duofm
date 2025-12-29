package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration.
type Config struct {
	Keybindings map[string][]string `toml:"keybindings"`
}

// rawConfig is used for TOML parsing to handle the [keybindings] section.
type rawConfig struct {
	Keybindings map[string][]string `toml:"keybindings"`
}

// LoadConfig loads the configuration from the specified path.
// If the file does not exist, returns default configuration.
// If parsing fails, returns default configuration with a warning.
func LoadConfig(path string) (*Config, []string) {
	var warnings []string

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return defaultConfig(), warnings
	}

	// Parse TOML file
	var raw rawConfig
	if _, err := toml.DecodeFile(path, &raw); err != nil {
		warnings = append(warnings, fmt.Sprintf("Warning: config parse error, using defaults: %v", err))
		return defaultConfig(), warnings
	}

	// If no keybindings section, use defaults
	if raw.Keybindings == nil {
		return defaultConfig(), warnings
	}

	// Merge with defaults
	cfg := defaultConfig()
	for action, keys := range raw.Keybindings {
		cfg.Keybindings[action] = keys
	}

	return cfg, warnings
}

// defaultConfig returns the default configuration.
func defaultConfig() *Config {
	return &Config{
		Keybindings: DefaultKeybindings(),
	}
}
