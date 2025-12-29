package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration.
type Config struct {
	Keybindings map[string][]string `toml:"keybindings"`
	Colors      *ColorConfig
}

// rawConfig is used for TOML parsing to handle the [keybindings] and [colors] sections.
type rawConfig struct {
	Keybindings map[string][]string    `toml:"keybindings"`
	Colors      map[string]interface{} `toml:"colors"`
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

	// Start with defaults
	cfg := defaultConfig()

	// Merge keybindings with defaults
	for action, keys := range raw.Keybindings {
		cfg.Keybindings[action] = keys
	}

	// Load colors (merges with defaults, generates warnings for invalid values)
	colors, colorWarnings := LoadColors(raw.Colors)
	cfg.Colors = colors
	warnings = append(warnings, colorWarnings...)

	return cfg, warnings
}

// defaultConfig returns the default configuration.
func defaultConfig() *Config {
	return &Config{
		Keybindings: DefaultKeybindings(),
		Colors:      DefaultColors(),
	}
}
