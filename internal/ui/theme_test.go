package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/config"
)

func TestNewTheme(t *testing.T) {
	cfg := &config.ColorConfig{
		CursorFg:         15,
		CursorBg:         39,
		CursorBgInactive: 240,
	}

	theme := NewTheme(cfg)

	// Verify colors are converted correctly
	if theme.CursorFg != lipgloss.Color("15") {
		t.Errorf("CursorFg = %v, want 15", theme.CursorFg)
	}
	if theme.CursorBg != lipgloss.Color("39") {
		t.Errorf("CursorBg = %v, want 39", theme.CursorBg)
	}
	if theme.CursorBgInactive != lipgloss.Color("240") {
		t.Errorf("CursorBgInactive = %v, want 240", theme.CursorBgInactive)
	}
}

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()
	defaults := config.DefaultColors()

	// Verify theme matches default colors
	if theme.CursorFg != lipgloss.Color(colorCodeToString(defaults.CursorFg)) {
		t.Errorf("CursorFg mismatch")
	}
	if theme.CursorBg != lipgloss.Color(colorCodeToString(defaults.CursorBg)) {
		t.Errorf("CursorBg mismatch")
	}
	if theme.DirectoryFg != lipgloss.Color(colorCodeToString(defaults.DirectoryFg)) {
		t.Errorf("DirectoryFg mismatch")
	}
	if theme.ErrorFg != lipgloss.Color(colorCodeToString(defaults.ErrorFg)) {
		t.Errorf("ErrorFg mismatch")
	}
}

func TestColorCodeToString(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{15, "15"},
		{39, "39"},
		{100, "100"},
		{196, "196"},
		{240, "240"},
		{255, "255"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := colorCodeToString(tt.input)
			if result != tt.expected {
				t.Errorf("colorCodeToString(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestThemeHasAllColors(t *testing.T) {
	theme := DefaultTheme()

	// Verify all theme fields are populated (not empty strings)
	colors := []struct {
		name  string
		color lipgloss.Color
	}{
		{"CursorFg", theme.CursorFg},
		{"CursorBg", theme.CursorBg},
		{"CursorBgInactive", theme.CursorBgInactive},
		{"MarkFg", theme.MarkFg},
		{"MarkFgInactive", theme.MarkFgInactive},
		{"MarkBg", theme.MarkBg},
		{"MarkBgInactive", theme.MarkBgInactive},
		{"CursorMarkFg", theme.CursorMarkFg},
		{"CursorMarkBg", theme.CursorMarkBg},
		{"CursorMarkBgInactive", theme.CursorMarkBgInactive},
		{"PathFg", theme.PathFg},
		{"PathFgInactive", theme.PathFgInactive},
		{"HeaderFg", theme.HeaderFg},
		{"HeaderFgInactive", theme.HeaderFgInactive},
		{"BorderFg", theme.BorderFg},
		{"DimmedBg", theme.DimmedBg},
		{"DimmedFg", theme.DimmedFg},
		{"DirectoryFg", theme.DirectoryFg},
		{"SymlinkFg", theme.SymlinkFg},
		{"ExecutableFg", theme.ExecutableFg},
		{"DialogTitleFg", theme.DialogTitleFg},
		{"DialogBorderFg", theme.DialogBorderFg},
		{"DialogSelectedFg", theme.DialogSelectedFg},
		{"DialogSelectedBg", theme.DialogSelectedBg},
		{"DialogFooterFg", theme.DialogFooterFg},
		{"InputFg", theme.InputFg},
		{"InputBg", theme.InputBg},
		{"InputBorderFg", theme.InputBorderFg},
		{"MinibufferFg", theme.MinibufferFg},
		{"MinibufferBg", theme.MinibufferBg},
		{"ErrorFg", theme.ErrorFg},
		{"ErrorBorderFg", theme.ErrorBorderFg},
		{"WarningFg", theme.WarningFg},
		{"StatusFg", theme.StatusFg},
		{"StatusBg", theme.StatusBg},
	}

	for _, c := range colors {
		if c.color == "" {
			t.Errorf("%s is empty", c.name)
		}
	}

	// Verify we have 35 colors
	if len(colors) != 35 {
		t.Errorf("Expected 35 colors, got %d", len(colors))
	}
}

func TestNewTheme_NilConfig(t *testing.T) {
	// Should not panic, should return default theme
	theme := NewTheme(nil)
	if theme == nil {
		t.Fatal("NewTheme(nil) returned nil")
	}

	// Verify it has default colors
	defaults := config.DefaultColors()
	if theme.CursorFg != lipgloss.Color(colorCodeToString(defaults.CursorFg)) {
		t.Errorf("CursorFg mismatch with default")
	}
	if theme.CursorBg != lipgloss.Color(colorCodeToString(defaults.CursorBg)) {
		t.Errorf("CursorBg mismatch with default")
	}
}

func TestColorCodeToString_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"negative", -1, "-1"},
		{"negative large", -255, "-255"},
		{"max int safe", 999999, "999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorCodeToString(tt.input)
			if result != tt.expected {
				t.Errorf("colorCodeToString(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
