package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// ConfigLoadResult holds the detailed result of a config load operation.
type ConfigLoadResult struct {
	Config        *Config
	Warnings      []string
	Errors        []ConfigError
	HasSyntaxErr  bool
	SyntaxErrLine int    // 1-based line number of syntax error
	SyntaxErrMsg  string // Formatted error message with position info
}

// ConfigError represents an individual config value error.
type ConfigError struct {
	Field   string // Field name (e.g., "history_limit", "colors.cursor_fg")
	Message string // Error description
	Line    int    // Line number (if available)
}

// HasErrors returns whether any errors exist.
func (r *ConfigLoadResult) HasErrors() bool {
	return len(r.Errors) > 0 || r.HasSyntaxErr
}

// LoadConfigDetailed loads configuration with detailed error reporting.
// Unlike LoadConfig, it reports specific errors and attempts partial recovery.
func LoadConfigDetailed(path string) *ConfigLoadResult {
	result := &ConfigLoadResult{}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		result.Config = defaultConfig()
		return result
	}

	// Read the file content for partial parsing
	fileContent, err := os.ReadFile(path)
	if err != nil {
		result.Config = defaultConfig()
		result.Errors = append(result.Errors, ConfigError{
			Field:   "file",
			Message: fmt.Sprintf("Failed to read config file: %v", err),
		})
		return result
	}

	// Try parsing the TOML file
	var raw rawConfig
	_, parseErr := toml.Decode(string(fileContent), &raw)

	if parseErr != nil {
		// Check if it's a syntax error
		var tomlErr toml.ParseError
		if errors.As(parseErr, &tomlErr) {
			result.HasSyntaxErr = true
			result.SyntaxErrMsg = tomlErr.ErrorWithPosition()

			// Extract line number from ParseError
			result.SyntaxErrLine = tomlErr.Position.Line

			// Try partial parse of content before the error line
			partialRaw, partialErr := partialParse(string(fileContent), result.SyntaxErrLine)
			if partialErr == nil && partialRaw != nil {
				// Build config from partial parse + defaults
				result.Config = buildConfigFromRaw(partialRaw, result)
			} else {
				// Partial parse failed, use full defaults
				result.Config = defaultConfig()
			}
		} else {
			// Non-TOML parse error
			result.HasSyntaxErr = true
			result.SyntaxErrMsg = parseErr.Error()
			result.Config = defaultConfig()
		}
		return result
	}

	// Parse succeeded - validate values
	result.Config = buildConfigFromRaw(&raw, result)
	return result
}

// buildConfigFromRaw builds a Config from rawConfig, validating values and
// recording errors for invalid ones.
func buildConfigFromRaw(raw *rawConfig, result *ConfigLoadResult) *Config {
	cfg := defaultConfig()

	// Merge keybindings
	if raw.Keybindings != nil {
		for action, keys := range raw.Keybindings {
			cfg.Keybindings[action] = keys
		}
	}

	// Load colors with validation
	if raw.Colors != nil {
		colors, colorWarnings := LoadColors(raw.Colors)
		cfg.Colors = colors
		// Convert color warnings to errors
		for _, w := range colorWarnings {
			// Extract field name from warning message
			field := extractColorFieldFromWarning(w, raw.Colors)
			result.Errors = append(result.Errors, ConfigError{
				Field:   "colors." + field,
				Message: w,
			})
		}
	}

	// Load history_limit
	if raw.HistoryLimit != nil {
		cfg.HistoryLimit = *raw.HistoryLimit
	}

	// Load shell_log_dir
	if raw.ShellLogDir != nil {
		cfg.ShellLogDir = *raw.ShellLogDir
	}

	// Load refresh_rate with validation
	if raw.RefreshRate != nil {
		rate := *raw.RefreshRate
		if rate < 0 || rate > 60 {
			result.Errors = append(result.Errors, ConfigError{
				Field:   "refresh_rate",
				Message: fmt.Sprintf("refresh_rate %d out of range (0-60), using default %d", rate, DefaultRefreshRate),
			})
		} else {
			cfg.RefreshRate = rate
		}
	}

	// Load enter_behavior with validation
	if raw.EnterBehavior != nil {
		enterBehavior, warning := ParseEnterBehavior(*raw.EnterBehavior)
		if warning != "" {
			result.Errors = append(result.Errors, ConfigError{
				Field:   "enter_behavior",
				Message: warning,
			})
			// Keep default
		} else {
			cfg.EnterBehavior = enterBehavior
		}
	}

	// Load MIME behavior
	if cfg.EnterBehavior.Type == EnterBehaviorMIME {
		mimeBehavior, mimeWarnings := ParseMIMEBehavior(raw.EnterBehaviorMIME)
		cfg.MIMEBehavior = mimeBehavior
		for _, w := range mimeWarnings {
			result.Warnings = append(result.Warnings, w)
		}
	}

	return cfg
}

// partialParse attempts to parse TOML content up to (but not including) the given line number.
// Returns nil and error if the truncated content is not valid TOML.
func partialParse(content string, upToLine int) (*rawConfig, error) {
	lines := strings.Split(content, "\n")

	// Keep lines before the error line (upToLine is 1-based)
	cutIndex := upToLine - 1
	if cutIndex <= 0 {
		return nil, fmt.Errorf("no content before line %d", upToLine)
	}
	if cutIndex > len(lines) {
		cutIndex = len(lines)
	}

	truncated := strings.Join(lines[:cutIndex], "\n")

	var raw rawConfig
	if _, err := toml.Decode(truncated, &raw); err != nil {
		return nil, fmt.Errorf("partial parse failed: %w", err)
	}

	return &raw, nil
}

// extractColorFieldFromWarning tries to extract the color field name from a warning message.
func extractColorFieldFromWarning(warning string, rawColors map[string]interface{}) string {
	// Try to find a color key name in the warning message
	for key := range rawColors {
		if strings.Contains(warning, key) {
			return key
		}
	}
	return "unknown"
}
