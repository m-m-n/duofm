package config

import (
	"strings"
	"testing"
)

func TestParseEnterBehavior(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedType    EnterBehaviorType
		expectedPath    string
		expectWarning   bool
		warningContains string
	}{
		{
			name:          "less",
			input:         "less",
			expectedType:  EnterBehaviorLess,
			expectedPath:  "",
			expectWarning: false,
		},
		{
			name:          "xdg-open",
			input:         "xdg-open",
			expectedType:  EnterBehaviorXDGOpen,
			expectedPath:  "",
			expectWarning: false,
		},
		{
			name:          "custom path with path prefix",
			input:         "path:/usr/bin/vim",
			expectedType:  EnterBehaviorCustom,
			expectedPath:  "/usr/bin/vim",
			expectWarning: false,
		},
		{
			name:          "custom path with spaces",
			input:         "path:/path/to/my app",
			expectedType:  EnterBehaviorCustom,
			expectedPath:  "/path/to/my app",
			expectWarning: false,
		},
		{
			name:            "empty string",
			input:           "",
			expectedType:    EnterBehaviorLess,
			expectedPath:    "",
			expectWarning:   true,
			warningContains: "invalid",
		},
		{
			name:          "whitespace trimmed to less",
			input:         "  less  ",
			expectedType:  EnterBehaviorLess,
			expectedPath:  "",
			expectWarning: false,
		},
		{
			name:          "whitespace trimmed to xdg-open",
			input:         "  xdg-open  ",
			expectedType:  EnterBehaviorXDGOpen,
			expectedPath:  "",
			expectWarning: false,
		},
		{
			name:            "unknown value",
			input:           "unknown",
			expectedType:    EnterBehaviorLess,
			expectedPath:    "",
			expectWarning:   true,
			warningContains: "invalid",
		},
		{
			name:            "path with empty path",
			input:           "path:",
			expectedType:    EnterBehaviorLess,
			expectedPath:    "",
			expectWarning:   true,
			warningContains: "empty",
		},
		{
			name:            "path with only whitespace",
			input:           "path:   ",
			expectedType:    EnterBehaviorLess,
			expectedPath:    "",
			expectWarning:   true,
			warningContains: "empty",
		},
		{
			name:          "custom command name without absolute path",
			input:         "path:nvim",
			expectedType:  EnterBehaviorCustom,
			expectedPath:  "nvim",
			expectWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, warning := ParseEnterBehavior(tt.input)

			if result.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", result.Type, tt.expectedType)
			}

			if result.CustomPath != tt.expectedPath {
				t.Errorf("CustomPath = %q, want %q", result.CustomPath, tt.expectedPath)
			}

			if tt.expectWarning {
				if warning == "" {
					t.Error("expected warning but got none")
				} else if !strings.Contains(warning, tt.warningContains) {
					t.Errorf("warning %q should contain %q", warning, tt.warningContains)
				}
			} else {
				if warning != "" {
					t.Errorf("unexpected warning: %q", warning)
				}
			}
		})
	}
}

func TestDefaultEnterBehavior(t *testing.T) {
	result := DefaultEnterBehavior()

	if result.Type != EnterBehaviorLess {
		t.Errorf("DefaultEnterBehavior().Type = %v, want %v", result.Type, EnterBehaviorLess)
	}

	if result.CustomPath != "" {
		t.Errorf("DefaultEnterBehavior().CustomPath = %q, want empty string", result.CustomPath)
	}
}

func TestEnterBehaviorString(t *testing.T) {
	tests := []struct {
		name     string
		behavior EnterBehavior
		expected string
	}{
		{
			name:     "less",
			behavior: EnterBehavior{Type: EnterBehaviorLess},
			expected: "less",
		},
		{
			name:     "xdg-open",
			behavior: EnterBehavior{Type: EnterBehaviorXDGOpen},
			expected: "xdg-open",
		},
		{
			name:     "custom path",
			behavior: EnterBehavior{Type: EnterBehaviorCustom, CustomPath: "/usr/bin/vim"},
			expected: "path:/usr/bin/vim",
		},
		{
			name:     "custom command",
			behavior: EnterBehavior{Type: EnterBehaviorCustom, CustomPath: "nvim"},
			expected: "path:nvim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.behavior.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}
