package ui

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/config"
)

// Theme provides lipgloss colors for UI components.
// It converts ColorConfig integer values to lipgloss.Color.
type Theme struct {
	// Cursor (current row)
	CursorFg         lipgloss.Color
	CursorBg         lipgloss.Color
	CursorBgInactive lipgloss.Color

	// Marked rows
	MarkFg         lipgloss.Color
	MarkFgInactive lipgloss.Color
	MarkBg         lipgloss.Color
	MarkBgInactive lipgloss.Color

	// Cursor + Marked rows
	CursorMarkFg         lipgloss.Color
	CursorMarkBg         lipgloss.Color
	CursorMarkBgInactive lipgloss.Color

	// Path display
	PathFg         lipgloss.Color
	PathFgInactive lipgloss.Color

	// Header (column names)
	HeaderFg         lipgloss.Color
	HeaderFgInactive lipgloss.Color

	// Pane structure
	BorderFg lipgloss.Color
	DimmedBg lipgloss.Color
	DimmedFg lipgloss.Color

	// File types
	DirectoryFg  lipgloss.Color
	SymlinkFg    lipgloss.Color
	ExecutableFg lipgloss.Color

	// Dialog
	DialogTitleFg    lipgloss.Color
	DialogBorderFg   lipgloss.Color
	DialogSelectedFg lipgloss.Color
	DialogSelectedBg lipgloss.Color
	DialogFooterFg   lipgloss.Color

	// Input fields
	InputFg       lipgloss.Color
	InputBg       lipgloss.Color
	InputBorderFg lipgloss.Color

	// Minibuffer
	MinibufferFg lipgloss.Color
	MinibufferBg lipgloss.Color

	// Error and warning
	ErrorFg       lipgloss.Color
	ErrorBorderFg lipgloss.Color
	WarningFg     lipgloss.Color

	// Status bar
	StatusFg lipgloss.Color
	StatusBg lipgloss.Color
}

// NewTheme creates a Theme from ColorConfig.
// If cfg is nil, returns DefaultTheme.
func NewTheme(cfg *config.ColorConfig) *Theme {
	if cfg == nil {
		return NewTheme(config.DefaultColors())
	}
	return &Theme{
		// Cursor
		CursorFg:         lipgloss.Color(colorCodeToString(cfg.CursorFg)),
		CursorBg:         lipgloss.Color(colorCodeToString(cfg.CursorBg)),
		CursorBgInactive: lipgloss.Color(colorCodeToString(cfg.CursorBgInactive)),

		// Mark
		MarkFg:         lipgloss.Color(colorCodeToString(cfg.MarkFg)),
		MarkFgInactive: lipgloss.Color(colorCodeToString(cfg.MarkFgInactive)),
		MarkBg:         lipgloss.Color(colorCodeToString(cfg.MarkBg)),
		MarkBgInactive: lipgloss.Color(colorCodeToString(cfg.MarkBgInactive)),

		// Cursor + Mark
		CursorMarkFg:         lipgloss.Color(colorCodeToString(cfg.CursorMarkFg)),
		CursorMarkBg:         lipgloss.Color(colorCodeToString(cfg.CursorMarkBg)),
		CursorMarkBgInactive: lipgloss.Color(colorCodeToString(cfg.CursorMarkBgInactive)),

		// Path
		PathFg:         lipgloss.Color(colorCodeToString(cfg.PathFg)),
		PathFgInactive: lipgloss.Color(colorCodeToString(cfg.PathFgInactive)),

		// Header
		HeaderFg:         lipgloss.Color(colorCodeToString(cfg.HeaderFg)),
		HeaderFgInactive: lipgloss.Color(colorCodeToString(cfg.HeaderFgInactive)),

		// Pane structure
		BorderFg: lipgloss.Color(colorCodeToString(cfg.BorderFg)),
		DimmedBg: lipgloss.Color(colorCodeToString(cfg.DimmedBg)),
		DimmedFg: lipgloss.Color(colorCodeToString(cfg.DimmedFg)),

		// File types
		DirectoryFg:  lipgloss.Color(colorCodeToString(cfg.DirectoryFg)),
		SymlinkFg:    lipgloss.Color(colorCodeToString(cfg.SymlinkFg)),
		ExecutableFg: lipgloss.Color(colorCodeToString(cfg.ExecutableFg)),

		// Dialog
		DialogTitleFg:    lipgloss.Color(colorCodeToString(cfg.DialogTitleFg)),
		DialogBorderFg:   lipgloss.Color(colorCodeToString(cfg.DialogBorderFg)),
		DialogSelectedFg: lipgloss.Color(colorCodeToString(cfg.DialogSelectedFg)),
		DialogSelectedBg: lipgloss.Color(colorCodeToString(cfg.DialogSelectedBg)),
		DialogFooterFg:   lipgloss.Color(colorCodeToString(cfg.DialogFooterFg)),

		// Input
		InputFg:       lipgloss.Color(colorCodeToString(cfg.InputFg)),
		InputBg:       lipgloss.Color(colorCodeToString(cfg.InputBg)),
		InputBorderFg: lipgloss.Color(colorCodeToString(cfg.InputBorderFg)),

		// Minibuffer
		MinibufferFg: lipgloss.Color(colorCodeToString(cfg.MinibufferFg)),
		MinibufferBg: lipgloss.Color(colorCodeToString(cfg.MinibufferBg)),

		// Error and warning
		ErrorFg:       lipgloss.Color(colorCodeToString(cfg.ErrorFg)),
		ErrorBorderFg: lipgloss.Color(colorCodeToString(cfg.ErrorBorderFg)),
		WarningFg:     lipgloss.Color(colorCodeToString(cfg.WarningFg)),

		// Status bar
		StatusFg: lipgloss.Color(colorCodeToString(cfg.StatusFg)),
		StatusBg: lipgloss.Color(colorCodeToString(cfg.StatusBg)),
	}
}

// DefaultTheme creates a Theme with default colors.
func DefaultTheme() *Theme {
	return NewTheme(config.DefaultColors())
}

// colorCodeToString converts an integer color code to string.
func colorCodeToString(i int) string {
	return strconv.Itoa(i)
}
