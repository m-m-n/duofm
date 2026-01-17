package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DropdownAction represents the result of handling a key in a dropdown.
type DropdownAction int

const (
	// DropdownActionNone indicates no action (cursor moved or invalid key)
	DropdownActionNone DropdownAction = iota
	// DropdownActionSelected indicates user pressed Enter and option was selected
	DropdownActionSelected
	// DropdownActionCancelled indicates user pressed Escape
	DropdownActionCancelled
)

// String returns a string representation of DropdownAction.
func (a DropdownAction) String() string {
	switch a {
	case DropdownActionNone:
		return "None"
	case DropdownActionSelected:
		return "Selected"
	case DropdownActionCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// DropdownOption represents a single option in the dropdown.
type DropdownOption struct {
	Label string // Display label
	Value string // Internal value
}

// Dropdown is a reusable dropdown component for selection from a list.
type Dropdown struct {
	options          []DropdownOption
	selectedIndex    int  // Currently confirmed selection
	highlightedIndex int  // Temporary cursor while navigating (reset on open)
	expanded         bool // Whether dropdown is currently open
}

// NewDropdown creates a new dropdown with the given options and initial selection.
func NewDropdown(options []DropdownOption, selectedIndex int) *Dropdown {
	if selectedIndex < 0 {
		selectedIndex = 0
	}
	if selectedIndex >= len(options) {
		selectedIndex = len(options) - 1
	}
	if selectedIndex < 0 {
		selectedIndex = 0
	}

	return &Dropdown{
		options:          options,
		selectedIndex:    selectedIndex,
		highlightedIndex: selectedIndex,
		expanded:         false,
	}
}

// Expand opens the dropdown and initializes highlightedIndex from selectedIndex.
func (d *Dropdown) Expand() {
	d.expanded = true
	d.highlightedIndex = d.selectedIndex
}

// Collapse closes the dropdown.
func (d *Dropdown) Collapse() {
	d.expanded = false
}

// IsExpanded returns whether the dropdown is currently open.
func (d *Dropdown) IsExpanded() bool {
	return d.expanded
}

// SelectedValue returns the value of the currently selected option.
func (d *Dropdown) SelectedValue() string {
	if d.selectedIndex >= 0 && d.selectedIndex < len(d.options) {
		return d.options[d.selectedIndex].Value
	}
	return ""
}

// SelectedLabel returns the label of the currently selected option.
func (d *Dropdown) SelectedLabel() string {
	if d.selectedIndex >= 0 && d.selectedIndex < len(d.options) {
		return d.options[d.selectedIndex].Label
	}
	return ""
}

// SelectedIndex returns the index of the currently selected option.
func (d *Dropdown) SelectedIndex() int {
	return d.selectedIndex
}

// SetSelectedIndex sets the selected index, clamping to valid range.
func (d *Dropdown) SetSelectedIndex(index int) {
	if index < 0 {
		index = 0
	}
	if index >= len(d.options) {
		index = len(d.options) - 1
	}
	if index < 0 {
		index = 0
	}
	d.selectedIndex = index
}

// OptionCount returns the number of options.
func (d *Dropdown) OptionCount() int {
	return len(d.options)
}

// HandleKey processes a key input when the dropdown is expanded.
// Returns DropdownAction indicating the result.
func (d *Dropdown) HandleKey(key string) DropdownAction {
	if !d.expanded {
		return DropdownActionNone
	}

	switch key {
	case "j", "down":
		if d.highlightedIndex < len(d.options)-1 {
			d.highlightedIndex++
		}
		return DropdownActionNone

	case "k", "up":
		if d.highlightedIndex > 0 {
			d.highlightedIndex--
		}
		return DropdownActionNone

	case "enter":
		d.selectedIndex = d.highlightedIndex
		d.Collapse()
		return DropdownActionSelected

	case "esc":
		d.Collapse()
		return DropdownActionCancelled
	}

	return DropdownActionNone
}

// View renders the dropdown.
// focused indicates whether this dropdown has focus in the parent dialog.
func (d *Dropdown) View(focused bool) string {
	if d.expanded {
		return d.viewExpanded(focused)
	}
	return d.viewClosed(focused)
}

// viewClosed renders the dropdown in closed state: [Value ▼]
func (d *Dropdown) viewClosed(focused bool) string {
	label := d.SelectedLabel() + " \u25BC" // ▼ down arrow indicator

	style := lipgloss.NewStyle()
	if focused {
		style = style.
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("39"))
	} else {
		style = style.
			Foreground(lipgloss.Color("252"))
	}

	return "[" + style.Render(label) + "]"
}

// viewExpanded renders the dropdown in expanded state with option list.
func (d *Dropdown) viewExpanded(focused bool) string {
	var b strings.Builder

	// Closed trigger showing current value
	triggerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("39"))

	b.WriteString("[" + triggerStyle.Render(d.SelectedLabel()) + "]")
	b.WriteString("\n")

	// Options list with border
	optionLines := make([]string, 0, len(d.options))

	for i, opt := range d.options {
		var line string
		if i == d.highlightedIndex {
			// Highlighted option
			highlightStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("39"))
			line = highlightStyle.Render(" " + opt.Label + " ")
		} else {
			// Normal option
			normalStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))
			line = normalStyle.Render(" " + opt.Label + " ")
		}
		optionLines = append(optionLines, line)
	}

	// Draw box around options
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	b.WriteString(boxStyle.Render(strings.Join(optionLines, "\n")))

	return b.String()
}
