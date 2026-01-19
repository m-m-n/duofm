package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FocusTarget represents which major item has focus in the dialog.
type FocusTarget int

const (
	// FocusTargetSortBy is the Sort by dropdown (major item 0).
	FocusTargetSortBy FocusTarget = iota
	// FocusTargetOrder is the Order dropdown (major item 1).
	FocusTargetOrder
	// FocusTargetOK is the OK button (major item 2).
	FocusTargetOK
)

// SortDialog is a dialog for changing sort settings using dropdown menus.
type SortDialog struct {
	BaseDialog
	config         SortConfig  // Current selection
	originalConfig SortConfig  // For restoration on cancel
	focusedItem    FocusTarget // 0: Sort by, 1: Order, 2: OK button
	fieldDropdown  *Dropdown   // Sort by dropdown
	orderDropdown  *Dropdown   // Order dropdown
	styles         DialogStyles
}

// NewSortDialog creates a new sort dialog with dropdown menus.
func NewSortDialog(current SortConfig) *SortDialog {
	base := NewBaseDialog(DialogDisplayPane)
	base.SetWidth(36)

	// Create Sort by dropdown
	fieldOptions := []DropdownOption{
		{Label: "Name", Value: "name"},
		{Label: "Size", Value: "size"},
		{Label: "Date", Value: "date"},
	}
	fieldDropdown := NewDropdown(fieldOptions, int(current.Field))

	// Create Order dropdown
	orderOptions := []DropdownOption{
		{Label: "\u2191Asc", Value: "asc"},   // Up arrow
		{Label: "\u2193Desc", Value: "desc"}, // Down arrow
	}
	orderDropdown := NewDropdown(orderOptions, int(current.Order))

	return &SortDialog{
		BaseDialog:     base,
		config:         current,
		originalConfig: current,
		focusedItem:    FocusTargetSortBy,
		fieldDropdown:  fieldDropdown,
		orderDropdown:  orderDropdown,
		styles:         NewDialogStyles(36, ColorPrimary),
	}
}

// HandleKey processes key input and returns confirmed/cancelled status.
func (d *SortDialog) HandleKey(key string) (confirmed bool, cancelled bool) {
	// q always cancels the dialog, even when dropdown is expanded
	if key == "q" {
		d.config = d.originalConfig
		d.Close()
		return false, true
	}

	// Check if any dropdown is expanded
	if d.fieldDropdown.IsExpanded() {
		return d.handleExpandedDropdownKey(d.fieldDropdown, key, true)
	}
	if d.orderDropdown.IsExpanded() {
		return d.handleExpandedDropdownKey(d.orderDropdown, key, false)
	}

	// No dropdown expanded - handle dialog-level keys
	return d.handleDialogKey(key)
}

// handleExpandedDropdownKey handles keys when a dropdown is expanded.
func (d *SortDialog) handleExpandedDropdownKey(dropdown *Dropdown, key string, isFieldDropdown bool) (confirmed bool, cancelled bool) {
	action := dropdown.HandleKey(key)

	switch action {
	case DropdownActionSelected:
		// Update config based on which dropdown was changed
		if isFieldDropdown {
			d.config.Field = SortField(dropdown.SelectedIndex())
		} else {
			d.config.Order = SortOrder(dropdown.SelectedIndex())
		}
		return false, false

	case DropdownActionCancelled:
		// Escape closes dropdown only, does not cancel dialog
		return false, false

	default:
		// DropdownActionNone - key was handled internally (cursor moved)
		return false, false
	}
}

// handleDialogKey handles keys when no dropdown is expanded.
func (d *SortDialog) handleDialogKey(key string) (confirmed bool, cancelled bool) {
	switch key {
	case "j", "down":
		// Move focus to next major item (no cycling)
		if d.focusedItem < FocusTargetOK {
			d.focusedItem++
		}
		return false, false

	case "k", "up":
		// Move focus to previous major item (no cycling)
		if d.focusedItem > FocusTargetSortBy {
			d.focusedItem--
		}
		return false, false

	case "tab":
		// Tab moves to next major item (no cycling)
		if d.focusedItem < FocusTargetOK {
			d.focusedItem++
		}
		return false, false

	case "shift+tab":
		// Shift+Tab moves to previous major item (no cycling)
		if d.focusedItem > FocusTargetSortBy {
			d.focusedItem--
		}
		return false, false

	case "enter":
		// If on OK button, confirm dialog
		if d.focusedItem == FocusTargetOK {
			d.Close()
			return true, false
		}
		// Otherwise, expand the focused dropdown
		d.getFocusedDropdown().Expand()
		return false, false

	case " ":
		// Space expands dropdown, but does nothing on OK button
		if d.focusedItem < FocusTargetOK {
			d.getFocusedDropdown().Expand()
		}
		return false, false

	case "esc":
		// Cancel dialog
		d.config = d.originalConfig
		d.Close()
		return false, true
	}

	return false, false
}

// getFocusedDropdown returns the currently focused dropdown.
// Returns nil if OK button is focused.
func (d *SortDialog) getFocusedDropdown() *Dropdown {
	switch d.focusedItem {
	case FocusTargetSortBy:
		return d.fieldDropdown
	case FocusTargetOrder:
		return d.orderDropdown
	default:
		return nil
	}
}

// Config returns the current sort configuration.
func (d *SortDialog) Config() SortConfig {
	return d.config
}

// OriginalConfig returns the original sort configuration.
func (d *SortDialog) OriginalConfig() SortConfig {
	return d.originalConfig
}

// Update implements the Bubble Tea Update interface.
func (d *SortDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		confirmed, cancelled := d.HandleKey(msg.String())

		if confirmed {
			return d, func() tea.Msg {
				return sortDialogResultMsg{config: d.config, confirmed: true}
			}
		}

		if cancelled {
			return d, func() tea.Msg {
				return sortDialogResultMsg{config: d.originalConfig, cancelled: true}
			}
		}

		// Live preview: emit config change message
		return d, func() tea.Msg {
			return sortDialogConfigChangedMsg{config: d.config}
		}
	}

	return d, nil
}

// View renders the dialog.
func (d *SortDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(d.styles.Title.Render("Sort"))
	b.WriteString("\n\n")

	// Sort by row
	b.WriteString(d.renderSortByRow())
	b.WriteString("\n")

	// Order row
	b.WriteString(d.renderOrderRow())
	b.WriteString("\n\n")

	// OK button
	b.WriteString(d.renderOKButton())
	b.WriteString("\n\n")

	// Help text
	b.WriteString(d.styles.Footer.Render("j/k:move  Enter:select  q:quit"))

	return d.styles.Box.Render(b.String())
}

// renderSortByRow renders the Sort by row with dropdown.
func (d *SortDialog) renderSortByRow() string {
	labelStyle := lipgloss.NewStyle().Width(10)
	isFocused := d.focusedItem == FocusTargetSortBy

	return labelStyle.Render("Sort by") + "  " + d.fieldDropdown.View(isFocused)
}

// renderOrderRow renders the Order row with dropdown.
func (d *SortDialog) renderOrderRow() string {
	labelStyle := lipgloss.NewStyle().Width(10)
	isFocused := d.focusedItem == FocusTargetOrder

	return labelStyle.Render("Order") + "  " + d.orderDropdown.View(isFocused)
}

// renderOKButton renders the OK button.
func (d *SortDialog) renderOKButton() string {
	labelStyle := lipgloss.NewStyle().Width(10)
	isFocused := d.focusedItem == FocusTargetOK

	var buttonStyle lipgloss.Style
	if isFocused {
		buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("39"))
	} else {
		buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
	}

	return labelStyle.Render("") + "  [" + buttonStyle.Render("OK") + "]"
}

// sortDialogResultMsg is the result message for the sort dialog.
type sortDialogResultMsg struct {
	config    SortConfig
	confirmed bool
	cancelled bool
}

// sortDialogConfigChangedMsg is emitted when sort config changes (for live preview).
type sortDialogConfigChangedMsg struct {
	config SortConfig
}
