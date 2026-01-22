package ui

import (
	"testing"
)

// setDesktopEnvironmentForTest sets the cached desktop environment value for testing.
// This function allows tests to control the cached value without modifying environment variables.
// It is used by multiple test files in this package (env_test.go, context_menu_dialog_test.go,
// model_operations_test.go) to ensure consistent test behavior across desktop environment scenarios.
func setDesktopEnvironmentForTest(value bool) {
	hasDesktop = value
}

// TestDetectDesktopEnvironment tests the detection of desktop environment via environment variables.
func TestDetectDesktopEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		display  string
		wayland  string
		expected bool
	}{
		{
			name:     "DISPLAY set",
			display:  ":0",
			wayland:  "",
			expected: true,
		},
		{
			name:     "WAYLAND_DISPLAY set",
			display:  "",
			wayland:  "wayland-0",
			expected: true,
		},
		{
			name:     "both unset",
			display:  "",
			wayland:  "",
			expected: false,
		},
		{
			name:     "DISPLAY empty string",
			display:  "",
			wayland:  "",
			expected: false,
		},
		{
			name:     "both set",
			display:  ":0",
			wayland:  "wayland-0",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectDesktopEnvironmentWithValues(tt.display, tt.wayland)
			if result != tt.expected {
				t.Errorf("detectDesktopEnvironmentWithValues(%q, %q) = %v, want %v",
					tt.display, tt.wayland, result, tt.expected)
			}
		})
	}
}

// TestHasDesktopEnvironment tests the HasDesktopEnvironment function with cached values.
func TestHasDesktopEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		setValue bool
	}{
		{
			name:     "cached true",
			setValue: true,
		},
		{
			name:     "cached false",
			setValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the cached value
			setDesktopEnvironmentForTest(tt.setValue)

			result := HasDesktopEnvironment()
			if result != tt.setValue {
				t.Errorf("HasDesktopEnvironment() = %v, want %v", result, tt.setValue)
			}
		})
	}
}
