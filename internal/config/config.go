package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration.
type Config struct {
	Keybindings   map[string][]string `toml:"keybindings"`
	Colors        *ColorConfig
	HistoryLimit  int `toml:"history_limit"`
	RefreshRate   int `toml:"refresh_rate"`
	EnterBehavior EnterBehavior
	MIMEBehavior  MIMEBehaviorConfig
}

// DefaultHistoryLimit is the default number of shell command history entries.
const DefaultHistoryLimit = 20000

// DefaultRefreshRate is the default auto-refresh interval in seconds.
const DefaultRefreshRate = 3

// rawConfig is used for TOML parsing to handle the [keybindings] and [colors] sections.
type rawConfig struct {
	Keybindings       map[string][]string    `toml:"keybindings"`
	Colors            map[string]interface{} `toml:"colors"`
	HistoryLimit      *int                   `toml:"history_limit"`
	RefreshRate       *int                   `toml:"refresh_rate"`
	EnterBehavior     *string                `toml:"enter_behavior"`
	EnterBehaviorMIME map[string][]string    `toml:"enter_behavior_mime"`
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

	// Load refresh_rate (validate range 0-60)
	if raw.RefreshRate != nil {
		rate := *raw.RefreshRate
		if rate < 0 || rate > 60 {
			warnings = append(warnings, fmt.Sprintf("Warning: refresh_rate %d out of range (0-60), using default %d", rate, DefaultRefreshRate))
		} else {
			cfg.RefreshRate = rate
		}
	}

	// Load enter_behavior (parse if provided, otherwise keep default)
	if raw.EnterBehavior != nil {
		enterBehavior, warning := ParseEnterBehavior(*raw.EnterBehavior)
		cfg.EnterBehavior = enterBehavior
		if warning != "" {
			warnings = append(warnings, fmt.Sprintf("Warning: %s", warning))
		}
	}

	// Load MIME behavior if enter_behavior is "mime:"
	if cfg.EnterBehavior.Type == EnterBehaviorMIME {
		mimeBehavior, mimeWarnings := ParseMIMEBehavior(raw.EnterBehaviorMIME)
		cfg.MIMEBehavior = mimeBehavior
		for _, w := range mimeWarnings {
			warnings = append(warnings, fmt.Sprintf("Warning: %s", w))
		}
	}

	return cfg, warnings
}

// defaultConfig returns the default configuration.
func defaultConfig() *Config {
	return &Config{
		Keybindings:   DefaultKeybindings(),
		Colors:        DefaultColors(),
		HistoryLimit:  DefaultHistoryLimit,
		RefreshRate:   DefaultRefreshRate,
		EnterBehavior: DefaultEnterBehavior(),
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
