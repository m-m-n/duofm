package config

import (
	"fmt"
	"math"
)

// ColorConfig holds all color values for UI theming.
// Colors are specified as ANSI 256-color codes (0-255).
type ColorConfig struct {
	// Cursor (current row)
	CursorFg         int `toml:"cursor_fg"`
	CursorBg         int `toml:"cursor_bg"`
	CursorBgInactive int `toml:"cursor_bg_inactive"`

	// Marked rows
	MarkFg         int `toml:"mark_fg"`
	MarkFgInactive int `toml:"mark_fg_inactive"`
	MarkBg         int `toml:"mark_bg"`
	MarkBgInactive int `toml:"mark_bg_inactive"`

	// Cursor + Marked rows
	CursorMarkFg         int `toml:"cursor_mark_fg"`
	CursorMarkBg         int `toml:"cursor_mark_bg"`
	CursorMarkBgInactive int `toml:"cursor_mark_bg_inactive"`

	// Path display
	PathFg         int `toml:"path_fg"`
	PathFgInactive int `toml:"path_fg_inactive"`

	// Header (column names)
	HeaderFg         int `toml:"header_fg"`
	HeaderFgInactive int `toml:"header_fg_inactive"`

	// Pane structure
	BorderFg int `toml:"border_fg"`
	DimmedBg int `toml:"dimmed_bg"`
	DimmedFg int `toml:"dimmed_fg"`

	// File types
	DirectoryFg  int `toml:"directory_fg"`
	SymlinkFg    int `toml:"symlink_fg"`
	ExecutableFg int `toml:"executable_fg"`

	// Dialog
	DialogTitleFg    int `toml:"dialog_title_fg"`
	DialogBorderFg   int `toml:"dialog_border_fg"`
	DialogSelectedFg int `toml:"dialog_selected_fg"`
	DialogSelectedBg int `toml:"dialog_selected_bg"`
	DialogFooterFg   int `toml:"dialog_footer_fg"`

	// Input fields
	InputFg       int `toml:"input_fg"`
	InputBg       int `toml:"input_bg"`
	InputBorderFg int `toml:"input_border_fg"`

	// Minibuffer
	MinibufferFg int `toml:"minibuffer_fg"`
	MinibufferBg int `toml:"minibuffer_bg"`

	// Error and warning
	ErrorFg       int `toml:"error_fg"`
	ErrorBorderFg int `toml:"error_border_fg"`
	WarningFg     int `toml:"warning_fg"`

	// Status bar
	StatusFg int `toml:"status_fg"`
	StatusBg int `toml:"status_bg"`
}

// DefaultColors returns the default ColorConfig matching current hardcoded values.
func DefaultColors() *ColorConfig {
	return &ColorConfig{
		// Cursor
		CursorFg:         15,  // white
		CursorBg:         39,  // blue (active pane)
		CursorBgInactive: 240, // gray (inactive pane)

		// Mark
		MarkFg:         0,   // black (active pane)
		MarkFgInactive: 15,  // white (inactive pane)
		MarkBg:         136, // dark yellow (active pane)
		MarkBgInactive: 94,  // darker yellow (inactive pane)

		// Cursor + Mark
		CursorMarkFg:         15, // white
		CursorMarkBg:         30, // cyan (active pane)
		CursorMarkBgInactive: 23, // dark cyan (inactive pane)

		// Path
		PathFg:         39,  // blue (active pane)
		PathFgInactive: 240, // gray (inactive pane)

		// Header
		HeaderFg:         245, // light gray (active pane)
		HeaderFgInactive: 240, // gray (inactive pane)

		// Pane structure
		BorderFg: 240, // gray
		DimmedBg: 236, // dark gray
		DimmedFg: 243, // medium gray

		// File types
		DirectoryFg:  39, // blue
		SymlinkFg:    14, // cyan
		ExecutableFg: 9,  // red

		// Dialog
		DialogTitleFg:    39,  // blue
		DialogBorderFg:   39,  // blue
		DialogSelectedFg: 0,   // black
		DialogSelectedBg: 39,  // blue
		DialogFooterFg:   240, // gray

		// Input
		InputFg:       15,  // white
		InputBg:       236, // dark gray
		InputBorderFg: 240, // gray

		// Minibuffer
		MinibufferFg: 15,  // white
		MinibufferBg: 236, // dark gray

		// Error and warning
		ErrorFg:       196, // red
		ErrorBorderFg: 196, // red
		WarningFg:     240, // gray

		// Status bar
		StatusFg: 15,  // white
		StatusBg: 240, // gray
	}
}

// ValidateColor checks if a color value is in the valid range (0-255).
// Returns an error if the value is out of range.
func ValidateColor(value int, key string) error {
	if value < 0 || value > 255 {
		return fmt.Errorf("color value %d out of range for %s, must be 0-255", value, key)
	}
	return nil
}

// colorKeyMap maps config key names to ColorConfig field setters.
// This allows dynamic loading from TOML.
var colorKeyMap = map[string]func(*ColorConfig, int){
	"cursor_fg":               func(c *ColorConfig, v int) { c.CursorFg = v },
	"cursor_bg":               func(c *ColorConfig, v int) { c.CursorBg = v },
	"cursor_bg_inactive":      func(c *ColorConfig, v int) { c.CursorBgInactive = v },
	"mark_fg":                 func(c *ColorConfig, v int) { c.MarkFg = v },
	"mark_fg_inactive":        func(c *ColorConfig, v int) { c.MarkFgInactive = v },
	"mark_bg":                 func(c *ColorConfig, v int) { c.MarkBg = v },
	"mark_bg_inactive":        func(c *ColorConfig, v int) { c.MarkBgInactive = v },
	"cursor_mark_fg":          func(c *ColorConfig, v int) { c.CursorMarkFg = v },
	"cursor_mark_bg":          func(c *ColorConfig, v int) { c.CursorMarkBg = v },
	"cursor_mark_bg_inactive": func(c *ColorConfig, v int) { c.CursorMarkBgInactive = v },
	"path_fg":                 func(c *ColorConfig, v int) { c.PathFg = v },
	"path_fg_inactive":        func(c *ColorConfig, v int) { c.PathFgInactive = v },
	"header_fg":               func(c *ColorConfig, v int) { c.HeaderFg = v },
	"header_fg_inactive":      func(c *ColorConfig, v int) { c.HeaderFgInactive = v },
	"border_fg":               func(c *ColorConfig, v int) { c.BorderFg = v },
	"dimmed_bg":               func(c *ColorConfig, v int) { c.DimmedBg = v },
	"dimmed_fg":               func(c *ColorConfig, v int) { c.DimmedFg = v },
	"directory_fg":            func(c *ColorConfig, v int) { c.DirectoryFg = v },
	"symlink_fg":              func(c *ColorConfig, v int) { c.SymlinkFg = v },
	"executable_fg":           func(c *ColorConfig, v int) { c.ExecutableFg = v },
	"dialog_title_fg":         func(c *ColorConfig, v int) { c.DialogTitleFg = v },
	"dialog_border_fg":        func(c *ColorConfig, v int) { c.DialogBorderFg = v },
	"dialog_selected_fg":      func(c *ColorConfig, v int) { c.DialogSelectedFg = v },
	"dialog_selected_bg":      func(c *ColorConfig, v int) { c.DialogSelectedBg = v },
	"dialog_footer_fg":        func(c *ColorConfig, v int) { c.DialogFooterFg = v },
	"input_fg":                func(c *ColorConfig, v int) { c.InputFg = v },
	"input_bg":                func(c *ColorConfig, v int) { c.InputBg = v },
	"input_border_fg":         func(c *ColorConfig, v int) { c.InputBorderFg = v },
	"minibuffer_fg":           func(c *ColorConfig, v int) { c.MinibufferFg = v },
	"minibuffer_bg":           func(c *ColorConfig, v int) { c.MinibufferBg = v },
	"error_fg":                func(c *ColorConfig, v int) { c.ErrorFg = v },
	"error_border_fg":         func(c *ColorConfig, v int) { c.ErrorBorderFg = v },
	"warning_fg":              func(c *ColorConfig, v int) { c.WarningFg = v },
	"status_fg":               func(c *ColorConfig, v int) { c.StatusFg = v },
	"status_bg":               func(c *ColorConfig, v int) { c.StatusBg = v },
}

// AllColorKeys returns all valid color key names.
func AllColorKeys() []string {
	keys := make([]string, 0, len(colorKeyMap))
	for k := range colorKeyMap {
		keys = append(keys, k)
	}
	return keys
}

// LoadColors parses color values from the raw map and merges with defaults.
// Returns the merged ColorConfig and any warnings for invalid values.
func LoadColors(rawColors map[string]interface{}) (*ColorConfig, []string) {
	var warnings []string
	cfg := DefaultColors()

	if rawColors == nil {
		return cfg, warnings
	}

	for key, value := range rawColors {
		setter, ok := colorKeyMap[key]
		if !ok {
			warnings = append(warnings, fmt.Sprintf("Warning: unknown color key %q in config, ignored", key))
			continue
		}

		// Try to convert to int with validation
		var intVal int
		switch v := value.(type) {
		case int64:
			if v < 0 || v > 255 {
				warnings = append(warnings, fmt.Sprintf("Warning: color value %d out of range for %s, using default", v, key))
				continue
			}
			intVal = int(v)
		case int:
			intVal = v
		case float64:
			// Check for NaN, Inf, or non-integer values
			if math.IsNaN(v) || math.IsInf(v, 0) {
				warnings = append(warnings, fmt.Sprintf("Warning: invalid color value (NaN/Inf) for %s, using default", key))
				continue
			}
			if math.Trunc(v) != v {
				warnings = append(warnings, fmt.Sprintf("Warning: fractional color value %.2f for %s, using default", v, key))
				continue
			}
			if v < 0 || v > 255 {
				warnings = append(warnings, fmt.Sprintf("Warning: color value %.0f out of range for %s, using default", v, key))
				continue
			}
			intVal = int(v)
		default:
			warnings = append(warnings, fmt.Sprintf("Warning: invalid color value %v for %s, using default", value, key))
			continue
		}

		// Validate range (for int type)
		if err := ValidateColor(intVal, key); err != nil {
			warnings = append(warnings, fmt.Sprintf("Warning: %v, using default", err))
			continue
		}

		setter(cfg, intVal)
	}

	return cfg, warnings
}
