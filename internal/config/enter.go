package config

import (
	"fmt"
	"strings"
)

// EnterBehaviorType represents the type of enter key behavior.
type EnterBehaviorType int

const (
	// EnterBehaviorLess opens files with pager (foreground, default)
	EnterBehaviorLess EnterBehaviorType = iota
	// EnterBehaviorXDGOpen opens files with system default app (background)
	EnterBehaviorXDGOpen
	// EnterBehaviorCustom opens files with a custom application (foreground)
	EnterBehaviorCustom
	// EnterBehaviorMIME opens files based on MIME type configuration
	EnterBehaviorMIME
)

// EnterBehavior represents the configured enter key behavior.
type EnterBehavior struct {
	Type       EnterBehaviorType
	CustomPath string // Only used when Type == EnterBehaviorCustom
}

// ParseEnterBehavior parses the enter_behavior config value.
// Returns the parsed EnterBehavior and any warning message.
// Valid values: "less", "xdg-open", "path:/path/to/app"
// Invalid values return default (less) with a warning.
//
// Processing:
//   - Input is trimmed (strings.TrimSpace) before parsing
//   - For "path:" format, PATH existence is NOT validated here
//   - Validation occurs at runtime in openWithCustomForeground() using exec.LookPath()
func ParseEnterBehavior(value string) (EnterBehavior, string) {
	value = strings.TrimSpace(value)

	switch value {
	case "less":
		return EnterBehavior{Type: EnterBehaviorLess}, ""
	case "xdg-open":
		return EnterBehavior{Type: EnterBehaviorXDGOpen}, ""
	case "mime:":
		return EnterBehavior{Type: EnterBehaviorMIME}, ""
	default:
		if strings.HasPrefix(value, "path:") {
			path := strings.TrimSpace(strings.TrimPrefix(value, "path:"))
			if path == "" {
				return DefaultEnterBehavior(), "empty path in enter_behavior, using default"
			}
			return EnterBehavior{Type: EnterBehaviorCustom, CustomPath: path}, ""
		}
		// Unknown value or empty string
		return DefaultEnterBehavior(), fmt.Sprintf("invalid enter_behavior value '%s', using default", value)
	}
}

// DefaultEnterBehavior returns the default enter behavior (less).
func DefaultEnterBehavior() EnterBehavior {
	return EnterBehavior{Type: EnterBehaviorLess}
}

// String returns the string representation of EnterBehavior.
func (e EnterBehavior) String() string {
	switch e.Type {
	case EnterBehaviorLess:
		return "less"
	case EnterBehaviorXDGOpen:
		return "xdg-open"
	case EnterBehaviorCustom:
		return "path:" + e.CustomPath
	case EnterBehaviorMIME:
		return "mime:"
	default:
		return "less"
	}
}
