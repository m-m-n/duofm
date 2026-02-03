package config

import (
	"fmt"
	"os"
	"strings"
)

// RepairConfig repairs a broken configuration file based on the load result.
// For syntax errors: removes content from the error line onwards.
// For value errors: replaces invalid values with defaults.
func RepairConfig(path string, result *ConfigLoadResult) error {
	// Get file permissions before modification
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}
	fileMode := fileInfo.Mode()

	// Read the current file content
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var repairedContent string

	if result.HasSyntaxErr {
		repairedContent = repairSyntaxError(string(content), result.SyntaxErrLine)
	} else if len(result.Errors) > 0 {
		repairedContent = repairValueErrors(string(content), result.Errors)
	} else {
		// No errors to repair
		return nil
	}

	// Write the repaired content back, preserving permissions
	if err := os.WriteFile(path, []byte(repairedContent), fileMode); err != nil {
		return fmt.Errorf("failed to write repaired config: %w", err)
	}

	return nil
}

// repairSyntaxError removes content from the error line onwards and ensures valid TOML.
func repairSyntaxError(content string, errLine int) string {
	lines := strings.Split(content, "\n")

	// Keep lines before the error line (errLine is 1-based)
	cutIndex := errLine - 1
	if cutIndex <= 0 {
		// Error on first line, return empty (will get defaults)
		return ""
	}
	if cutIndex > len(lines) {
		cutIndex = len(lines)
	}

	// Remove trailing incomplete lines (empty lines, partial sections)
	kept := lines[:cutIndex]
	for len(kept) > 0 {
		trimmed := strings.TrimSpace(kept[len(kept)-1])
		if trimmed == "" || trimmed == "[" {
			kept = kept[:len(kept)-1]
		} else {
			break
		}
	}

	result := strings.Join(kept, "\n")
	if result != "" && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	return result
}

// repairValueErrors replaces invalid values in the content with defaults.
func repairValueErrors(content string, errors []ConfigError) string {
	lines := strings.Split(content, "\n")

	for _, configErr := range errors {
		if configErr.Line > 0 && configErr.Line <= len(lines) {
			// Replace by line number
			lineIdx := configErr.Line - 1
			lines[lineIdx] = getDefaultValueLine(configErr.Field)
		} else {
			// Search by key name
			keyName := extractKeyName(configErr.Field)
			for i, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, keyName+" ") || strings.HasPrefix(trimmed, keyName+"=") {
					lines[i] = getDefaultValueLine(configErr.Field)
					break
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// extractKeyName extracts the key name from a field path.
// e.g., "colors.cursor_fg" -> "cursor_fg", "enter_behavior" -> "enter_behavior"
func extractKeyName(field string) string {
	parts := strings.Split(field, ".")
	return parts[len(parts)-1]
}

// getDefaultValueLine returns the TOML line for a field's default value.
func getDefaultValueLine(field string) string {
	switch field {
	case "enter_behavior":
		return fmt.Sprintf("enter_behavior = %q", DefaultEnterBehavior().String())
	case "history_limit":
		return fmt.Sprintf("history_limit = %d", DefaultHistoryLimit)
	default:
		// Handle color fields
		keyName := extractKeyName(field)
		defaultValue := GetDefaultColorValue(keyName)
		if defaultValue >= 0 {
			return fmt.Sprintf("%s = %d", keyName, defaultValue)
		}
		// Unknown field - comment it out
		return fmt.Sprintf("# %s = <removed: unknown field>", keyName)
	}
}
