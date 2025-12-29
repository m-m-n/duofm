package config

import (
	"fmt"
	"regexp"
	"strings"
)

// specialKeyMap maps PascalCase special key names to their Bubble Tea equivalents.
var specialKeyMap = map[string]string{
	"enter":     "enter",
	"esc":       "esc",
	"space":     " ",
	"tab":       "tab",
	"backspace": "backspace",
	"up":        "up",
	"down":      "down",
	"left":      "left",
	"right":     "right",
}

// functionKeyRegex matches function keys like F1, F5, F12.
var functionKeyRegex = regexp.MustCompile(`^[Ff]([1-9]|1[0-2])$`)

// NormalizeKey converts a configuration key string to Bubble Tea's internal format.
// Examples:
//   - "J" -> "j"
//   - "Ctrl+H" -> "ctrl+h"
//   - "Enter" -> "enter"
//   - "Space" -> " "
//   - "Shift+N" -> "N"
//   - "F5" -> "f5"
func NormalizeKey(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("empty key")
	}

	// Handle modifier keys
	if strings.Contains(key, "+") {
		return normalizeModifierKey(key)
	}

	// Handle single character keys (alphabets)
	if len(key) == 1 {
		ch := key[0]
		// Uppercase letters -> lowercase for Bubble Tea
		if ch >= 'A' && ch <= 'Z' {
			return strings.ToLower(key), nil
		}
		// Lowercase letters
		if ch >= 'a' && ch <= 'z' {
			return key, nil
		}
		// Symbols - keep as is
		return key, nil
	}

	// Handle special keys (Enter, Esc, Space, etc.)
	lowerKey := strings.ToLower(key)
	if mapped, ok := specialKeyMap[lowerKey]; ok {
		return mapped, nil
	}

	// Handle function keys
	if functionKeyRegex.MatchString(key) {
		return strings.ToLower(key), nil
	}

	return "", fmt.Errorf("invalid key format: %q", key)
}

// normalizeModifierKey handles keys with modifiers like Ctrl+H, Shift+N.
func normalizeModifierKey(key string) (string, error) {
	parts := strings.Split(key, "+")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid modifier key format: %q", key)
	}

	// Get the last part (the actual key)
	actualKey := parts[len(parts)-1]
	if actualKey == "" {
		return "", fmt.Errorf("invalid modifier key format: %q", key)
	}

	// Get modifiers
	modifiers := parts[:len(parts)-1]

	// Check for Shift modifier
	hasShift := false
	hasCtrl := false
	hasAlt := false

	for _, mod := range modifiers {
		switch strings.ToLower(mod) {
		case "shift":
			hasShift = true
		case "ctrl":
			hasCtrl = true
		case "alt":
			hasAlt = true
		default:
			return "", fmt.Errorf("unknown modifier: %q", mod)
		}
	}

	// Handle Shift+Letter -> uppercase letter
	if hasShift && !hasCtrl && !hasAlt {
		if len(actualKey) == 1 {
			ch := actualKey[0]
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				return strings.ToUpper(actualKey), nil
			}
		}
	}

	// Build Bubble Tea format
	var result strings.Builder

	if hasCtrl {
		result.WriteString("ctrl+")
	}
	if hasAlt {
		result.WriteString("alt+")
	}
	if hasShift && (hasCtrl || hasAlt) {
		result.WriteString("shift+")
	}

	// Normalize the actual key
	if len(actualKey) == 1 {
		ch := actualKey[0]
		if ch >= 'A' && ch <= 'Z' {
			result.WriteString(strings.ToLower(actualKey))
		} else if ch >= 'a' && ch <= 'z' {
			result.WriteString(actualKey)
		} else {
			// Symbol
			result.WriteString(actualKey)
		}
	} else {
		// Special key (Enter, etc.)
		lowerKey := strings.ToLower(actualKey)
		if mapped, ok := specialKeyMap[lowerKey]; ok {
			result.WriteString(mapped)
		} else if functionKeyRegex.MatchString(actualKey) {
			result.WriteString(strings.ToLower(actualKey))
		} else {
			return "", fmt.Errorf("invalid key in modifier combination: %q", actualKey)
		}
	}

	return result.String(), nil
}

// ValidateAction checks if the given action name is valid.
func ValidateAction(action string) bool {
	for _, a := range AllActions() {
		if a == action {
			return true
		}
	}
	return false
}

// ValidateKeybindings checks for duplicate key assignments and returns warnings.
func ValidateKeybindings(cfg *Config) []string {
	var warnings []string

	// Map of normalized key -> action name
	keyToAction := make(map[string]string)

	for action, keys := range cfg.Keybindings {
		for _, key := range keys {
			normalized, err := NormalizeKey(key)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Warning: invalid key %q in config for %s", key, action))
				continue
			}

			if existingAction, exists := keyToAction[normalized]; exists {
				warnings = append(warnings, fmt.Sprintf("Warning: key %q assigned to both %s and %s", key, existingAction, action))
			}
			keyToAction[normalized] = action
		}
	}

	return warnings
}
