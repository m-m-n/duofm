package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// BaseDialog provides common dialog state and behavior.
// Embed this struct in dialog implementations to reduce boilerplate.
type BaseDialog struct {
	active      bool
	width       int
	displayType DialogDisplayType
}

// NewBaseDialog creates a new BaseDialog with default settings.
func NewBaseDialog(displayType DialogDisplayType) BaseDialog {
	return BaseDialog{
		active:      true,
		width:       50,
		displayType: displayType,
	}
}

// IsActive returns whether the dialog is active.
func (b *BaseDialog) IsActive() bool {
	return b.active
}

// SetActive sets the dialog's active state.
func (b *BaseDialog) SetActive(active bool) {
	b.active = active
}

// DisplayType returns the dialog's display type.
func (b *BaseDialog) DisplayType() DialogDisplayType {
	return b.displayType
}

// Width returns the dialog width.
func (b *BaseDialog) Width() int {
	return b.width
}

// SetWidth sets the dialog width.
func (b *BaseDialog) SetWidth(width int) {
	b.width = width
}

// Close deactivates the dialog.
func (b *BaseDialog) Close() {
	b.active = false
}

// DialogStyles provides common styles for dialog rendering.
type DialogStyles struct {
	// Title style for dialog headers
	Title lipgloss.Style
	// Body style for main content
	Body lipgloss.Style
	// Footer style for hints/buttons
	Footer lipgloss.Style
	// Box style for the outer border
	Box lipgloss.Style
	// Error style for error messages
	Error lipgloss.Style
	// Input style for text input fields
	Input lipgloss.Style
}

// DialogColor defines standard dialog colors
type DialogColor string

const (
	ColorPrimary   DialogColor = "39"  // Cyan - primary actions
	ColorDanger    DialogColor = "196" // Red - errors, destructive actions
	ColorMuted     DialogColor = "240" // Gray - hints, secondary text
	ColorHighlight DialogColor = "15"  // White - highlighted text
	ColorInputBg   DialogColor = "236" // Dark gray - input background
	ColorBorder    DialogColor = "62"  // Purple - borders
)

// NewDialogStyles creates a new DialogStyles with the given width and border color.
func NewDialogStyles(width int, borderColor DialogColor) DialogStyles {
	return DialogStyles{
		Title: lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1).
			Bold(true).
			Foreground(lipgloss.Color(string(ColorPrimary))),

		Body: lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1),

		Footer: lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1).
			Foreground(lipgloss.Color(string(ColorMuted))),

		Box: lipgloss.NewStyle().
			Width(width).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(string(borderColor))).
			Padding(1, 2),

		Error: lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1).
			Foreground(lipgloss.Color(string(ColorDanger))),

		Input: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color(string(ColorHighlight))).
			Background(lipgloss.Color(string(ColorInputBg))).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(string(ColorMuted))),
	}
}

// DefaultDialogStyles creates styles with default settings.
func DefaultDialogStyles(width int) DialogStyles {
	return NewDialogStyles(width, ColorPrimary)
}

// ErrorDialogStyles creates styles for error dialogs.
func ErrorDialogStyles(width int) DialogStyles {
	styles := NewDialogStyles(width, ColorDanger)
	styles.Title = styles.Title.Foreground(lipgloss.Color(string(ColorDanger)))
	return styles
}
