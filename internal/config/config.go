package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration.
type Config struct {
	Keybindings  map[string][]string `toml:"keybindings"`
	Colors       *ColorConfig
	HistoryLimit int `toml:"history_limit"`
}

// DefaultHistoryLimit is the default number of shell command history entries.
const DefaultHistoryLimit = 20000

// rawConfig is used for TOML parsing to handle the [keybindings] and [colors] sections.
type rawConfig struct {
	Keybindings  map[string][]string    `toml:"keybindings"`
	Colors       map[string]interface{} `toml:"colors"`
	HistoryLimit *int                   `toml:"history_limit"`
}

// LoadConfig loads the configuration from the specified path.
// If the file does not exist, returns default configuration.
// If parsing fails, returns default configuration with a warning.
// Missing configuration items are automatically merged into the file.
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

	// Merge missing configuration items into the file
	if err := MergeConfig(path, &raw); err != nil {
		warnings = append(warnings, fmt.Sprintf("Warning: failed to merge config: %v", err))
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

	// Load history_limit (use explicit value if provided, otherwise keep default)
	if raw.HistoryLimit != nil {
		cfg.HistoryLimit = *raw.HistoryLimit
	}

	return cfg, warnings
}

// defaultConfig returns the default configuration.
func defaultConfig() *Config {
	return &Config{
		Keybindings:  DefaultKeybindings(),
		Colors:       DefaultColors(),
		HistoryLimit: DefaultHistoryLimit,
	}
}

// GetHistoryPath returns the path to the shell command history file.
// The path is fixed to ~/.config/duofm/history to prevent path traversal.
func GetHistoryPath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "history"), nil
}

// getConfigDir returns the configuration directory path.
func getConfigDir() (string, error) {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "duofm"), nil
	}

	// Fall back to ~/.config
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "duofm"), nil
}
