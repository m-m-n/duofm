package config

import (
	"math"
	"testing"
)

func TestDefaultColors(t *testing.T) {
	cfg := DefaultColors()

	// Verify all 35 color fields have expected default values
	tests := []struct {
		name     string
		got      int
		expected int
	}{
		// Cursor
		{"CursorFg", cfg.CursorFg, 15},
		{"CursorBg", cfg.CursorBg, 39},
		{"CursorBgInactive", cfg.CursorBgInactive, 240},

		// Mark
		{"MarkFg", cfg.MarkFg, 0},
		{"MarkFgInactive", cfg.MarkFgInactive, 15},
		{"MarkBg", cfg.MarkBg, 136},
		{"MarkBgInactive", cfg.MarkBgInactive, 94},

		// Cursor + Mark
		{"CursorMarkFg", cfg.CursorMarkFg, 15},
		{"CursorMarkBg", cfg.CursorMarkBg, 30},
		{"CursorMarkBgInactive", cfg.CursorMarkBgInactive, 23},

		// Path
		{"PathFg", cfg.PathFg, 39},
		{"PathFgInactive", cfg.PathFgInactive, 240},

		// Header
		{"HeaderFg", cfg.HeaderFg, 245},
		{"HeaderFgInactive", cfg.HeaderFgInactive, 240},

		// Pane structure
		{"BorderFg", cfg.BorderFg, 240},
		{"DimmedBg", cfg.DimmedBg, 236},
		{"DimmedFg", cfg.DimmedFg, 243},

		// File types
		{"DirectoryFg", cfg.DirectoryFg, 39},
		{"SymlinkFg", cfg.SymlinkFg, 14},
		{"ExecutableFg", cfg.ExecutableFg, 9},

		// Dialog
		{"DialogTitleFg", cfg.DialogTitleFg, 39},
		{"DialogBorderFg", cfg.DialogBorderFg, 39},
		{"DialogSelectedFg", cfg.DialogSelectedFg, 0},
		{"DialogSelectedBg", cfg.DialogSelectedBg, 39},
		{"DialogFooterFg", cfg.DialogFooterFg, 240},

		// Input
		{"InputFg", cfg.InputFg, 15},
		{"InputBg", cfg.InputBg, 236},
		{"InputBorderFg", cfg.InputBorderFg, 240},

		// Minibuffer
		{"MinibufferFg", cfg.MinibufferFg, 15},
		{"MinibufferBg", cfg.MinibufferBg, 236},

		// Error and warning
		{"ErrorFg", cfg.ErrorFg, 196},
		{"ErrorBorderFg", cfg.ErrorBorderFg, 196},
		{"WarningFg", cfg.WarningFg, 240},

		// Status bar
		{"StatusFg", cfg.StatusFg, 15},
		{"StatusBg", cfg.StatusBg, 240},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestValidateColor(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"valid_zero", 0, false},
		{"valid_255", 255, false},
		{"valid_middle", 128, false},
		{"invalid_negative", -1, true},
		{"invalid_256", 256, true},
		{"invalid_large", 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateColor(tt.value, "test_key")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateColor(%d) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestLoadColors_NilColors(t *testing.T) {
	cfg, warnings := LoadColors(nil)

	if len(warnings) != 0 {
		t.Errorf("Expected no warnings, got %v", warnings)
	}

	// Should return defaults
	defaults := DefaultColors()
	if cfg.CursorFg != defaults.CursorFg {
		t.Errorf("CursorFg = %d, want default %d", cfg.CursorFg, defaults.CursorFg)
	}
}

func TestLoadColors_PartialOverride(t *testing.T) {
	rawColors := map[string]interface{}{
		"cursor_fg": int64(200),
		"cursor_bg": int64(100),
	}

	cfg, warnings := LoadColors(rawColors)

	if len(warnings) != 0 {
		t.Errorf("Expected no warnings, got %v", warnings)
	}

	if cfg.CursorFg != 200 {
		t.Errorf("CursorFg = %d, want 200", cfg.CursorFg)
	}
	if cfg.CursorBg != 100 {
		t.Errorf("CursorBg = %d, want 100", cfg.CursorBg)
	}

	// Other values should be defaults
	defaults := DefaultColors()
	if cfg.CursorBgInactive != defaults.CursorBgInactive {
		t.Errorf("CursorBgInactive = %d, want default %d", cfg.CursorBgInactive, defaults.CursorBgInactive)
	}
}

func TestLoadColors_InvalidValue(t *testing.T) {
	rawColors := map[string]interface{}{
		"cursor_fg": int64(300), // out of range
	}

	cfg, warnings := LoadColors(rawColors)

	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d: %v", len(warnings), warnings)
	}

	// Should use default
	defaults := DefaultColors()
	if cfg.CursorFg != defaults.CursorFg {
		t.Errorf("CursorFg = %d, want default %d", cfg.CursorFg, defaults.CursorFg)
	}
}

func TestLoadColors_InvalidType(t *testing.T) {
	rawColors := map[string]interface{}{
		"cursor_fg": "blue", // string instead of int
	}

	cfg, warnings := LoadColors(rawColors)

	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d: %v", len(warnings), warnings)
	}

	// Should use default
	defaults := DefaultColors()
	if cfg.CursorFg != defaults.CursorFg {
		t.Errorf("CursorFg = %d, want default %d", cfg.CursorFg, defaults.CursorFg)
	}
}

func TestLoadColors_UnknownKey(t *testing.T) {
	rawColors := map[string]interface{}{
		"unknown_key": int64(100),
	}

	cfg, warnings := LoadColors(rawColors)

	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d: %v", len(warnings), warnings)
	}

	// All values should be defaults
	defaults := DefaultColors()
	if cfg.CursorFg != defaults.CursorFg {
		t.Errorf("CursorFg = %d, want default %d", cfg.CursorFg, defaults.CursorFg)
	}
}

func TestLoadColors_FloatValue(t *testing.T) {
	rawColors := map[string]interface{}{
		"cursor_fg": float64(100),
	}

	cfg, warnings := LoadColors(rawColors)

	if len(warnings) != 0 {
		t.Errorf("Expected no warnings, got %v", warnings)
	}

	if cfg.CursorFg != 100 {
		t.Errorf("CursorFg = %d, want 100", cfg.CursorFg)
	}
}

func TestAllColorKeys(t *testing.T) {
	keys := AllColorKeys()

	// Should have 35 keys
	if len(keys) != 35 {
		t.Errorf("Expected 35 color keys, got %d", len(keys))
	}

	// Verify some expected keys exist
	expectedKeys := []string{
		"cursor_fg", "cursor_bg", "mark_fg", "mark_bg",
		"directory_fg", "dialog_title_fg", "error_fg",
	}

	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	for _, expected := range expectedKeys {
		if !keyMap[expected] {
			t.Errorf("Expected key %q not found in AllColorKeys()", expected)
		}
	}
}

func TestLoadColors_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		wantWarn bool
	}{
		{
			name:     "fractional float",
			input:    map[string]interface{}{"cursor_fg": float64(100.5)},
			wantWarn: true,
		},
		{
			name:     "NaN",
			input:    map[string]interface{}{"cursor_fg": math.NaN()},
			wantWarn: true,
		},
		{
			name:     "positive Inf",
			input:    map[string]interface{}{"cursor_fg": math.Inf(1)},
			wantWarn: true,
		},
		{
			name:     "negative Inf",
			input:    map[string]interface{}{"cursor_fg": math.Inf(-1)},
			wantWarn: true,
		},
		{
			name:     "negative float",
			input:    map[string]interface{}{"cursor_fg": float64(-10)},
			wantWarn: true,
		},
		{
			name:     "float out of range",
			input:    map[string]interface{}{"cursor_fg": float64(300)},
			wantWarn: true,
		},
		{
			name:     "int64 out of range high",
			input:    map[string]interface{}{"cursor_fg": int64(99999)},
			wantWarn: true,
		},
		{
			name:     "int64 out of range negative",
			input:    map[string]interface{}{"cursor_fg": int64(-100)},
			wantWarn: true,
		},
		{
			name:     "valid integer float",
			input:    map[string]interface{}{"cursor_fg": float64(100.0)},
			wantWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, warnings := LoadColors(tt.input)
			if tt.wantWarn && len(warnings) == 0 {
				t.Error("Expected warning, got none")
			}
			if !tt.wantWarn && len(warnings) > 0 {
				t.Errorf("Expected no warning, got: %v", warnings)
			}
		})
	}
}
