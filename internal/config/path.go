package config

import (
	"os"
	"path/filepath"
)

// GetConfigPath returns the path to the configuration file.
// It respects XDG_CONFIG_HOME if set, otherwise uses ~/.config/duofm/config.toml
func GetConfigPath() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "duofm", "config.toml"), nil
}

// GetConfigDir returns the directory containing the configuration file.
func GetConfigDir() (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(configPath), nil
}
