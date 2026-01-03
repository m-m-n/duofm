package ui

import (
	"fmt"
	"io/fs"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	fsops "github.com/sakura/duofm/internal/fs"
)

// PermissionPreset represents a quick preset for permission values
type PermissionPreset struct {
	Number      int
	Mode        string
	Symbolic    string
	Description string
}

// File presets
var filePresets = []PermissionPreset{
	{Number: 1, Mode: "644", Symbolic: "-rw-r--r--", Description: "Default file"},
	{Number: 2, Mode: "755", Symbolic: "-rwxr-xr-x", Description: "Executable"},
	{Number: 3, Mode: "600", Symbolic: "-rw-------", Description: "Private"},
	{Number: 4, Mode: "777", Symbolic: "-rwxrwxrwx", Description: "Full access"},
}

// Directory presets
var dirPresets = []PermissionPreset{
	{Number: 1, Mode: "755", Symbolic: "drwxr-xr-x", Description: "Default directory"},
	{Number: 2, Mode: "700", Symbolic: "drwx------", Description: "Private directory"},
	{Number: 3, Mode: "775", Symbolic: "drwxrwxr-x", Description: "Group writable"},
	{Number: 4, Mode: "777", Symbolic: "drwxrwxrwx", Description: "Full access"},
}

// PermissionDialog is the permission change dialog
type PermissionDialog struct {
	targetName      string
	isDir           bool
	currentMode     fs.FileMode
	inputValue      string
	cursorPos       int
	recursiveOption int  // 0: this only, 1: recursive
	showRecursive   bool // true for directories
	presets         []PermissionPreset
	errorMsg        string
	active          bool
	width           int
	onConfirm       func(mode string, recursive bool) tea.Cmd
}

// NewPermissionDialog creates a new permission dialog
func NewPermissionDialog(targetName string, isDir bool, currentMode fs.FileMode) *PermissionDialog {
	presets := filePresets
	if isDir {
		presets = dirPresets
	}

	return &PermissionDialog{
		targetName:      targetName,
		isDir:           isDir,
		currentMode:     currentMode,
		inputValue:      "",
		cursorPos:       0,
		recursiveOption: 0,
		showRecursive:   isDir,
		presets:         presets,
		errorMsg:        "",
		active:          true,
		width:           50,
		onConfirm:       nil,
	}
}

// SetOnConfirm sets the confirmation callback
func (d *PermissionDialog) SetOnConfirm(callback func(mode string, recursive bool) tea.Cmd) {
	d.onConfirm = callback
}

// Update handles messages
func (d *PermissionDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear error message on any key press
		d.errorMsg = ""

		switch msg.Type {
		case tea.KeyEnter:
			// Validate and confirm
			if err := fsops.ValidatePermissionMode(d.inputValue); err != nil {
				d.errorMsg = err.Error()
				return d, nil
			}
			d.active = false
			if d.onConfirm != nil {
				recursive := d.recursiveOption == 1
				return d, d.onConfirm(d.inputValue, recursive)
			}
			return d, nil

		case tea.KeyEsc:
			d.active = false
			return d, nil

		case tea.KeyTab:
			// Toggle recursive option (directories only)
			if d.showRecursive {
				d.recursiveOption = (d.recursiveOption + 1) % 2
			}
			return d, nil

		case tea.KeySpace:
			// Toggle recursive option with Space key (FR3.4, FR10.6)
			if d.showRecursive {
				d.recursiveOption = (d.recursiveOption + 1) % 2
			}
			return d, nil

		case tea.KeyBackspace:
			if len(d.inputValue) > 0 {
				runes := []rune(d.inputValue)
				d.inputValue = string(runes[:len(runes)-1])
				d.cursorPos = len(d.inputValue)
			}
			return d, nil

		case tea.KeyRunes:
			// Handle preset selection and digit input
			if len(msg.Runes) > 0 {
				// Process each rune
				for _, r := range msg.Runes {
					// Handle j/k navigation for recursive option (FR3.3, FR10.5)
					if d.showRecursive && (r == 'j' || r == 'k') {
						if r == 'j' {
							// Move down (increment with wrap)
							d.recursiveOption = (d.recursiveOption + 1) % 2
						} else if r == 'k' {
							// Move up (decrement with wrap)
							d.recursiveOption = (d.recursiveOption - 1 + 2) % 2
						}
						return d, nil
					}

					// Check if it's a preset number (1-4) only when input is empty
					if r >= '1' && r <= '4' && len(d.inputValue) == 0 {
						presetIdx := int(r - '1')
						if presetIdx < len(d.presets) {
							d.inputValue = d.presets[presetIdx].Mode
							d.cursorPos = len(d.inputValue)
							return d, nil
						}
					}

					// Check if it's a valid octal digit (0-7)
					if r >= '0' && r <= '7' {
						// Limit to 3 digits
						if len(d.inputValue) < 3 {
							d.inputValue += string(r)
							d.cursorPos++
						}
						// Silently ignore 4th+ digits
						continue
					}

					// Invalid digit (skip j/k for files to avoid error message)
					if !d.showRecursive && (r == 'j' || r == 'k') {
						// Silently ignore j/k for file dialogs
						continue
					}

					// Invalid digit
					d.errorMsg = "Invalid digit: must be 0-7"
					return d, nil
				}
				return d, nil
			}
		}
	}

	return d, nil
}

// View renders the dialog
func (d *PermissionDialog) View() string {
	if !d.active {
		return ""
	}

	var b strings.Builder
	width := d.width

	// Title
	titleStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 1).
		Bold(true).
		Foreground(lipgloss.Color("39"))
	b.WriteString(titleStyle.Render(fmt.Sprintf("Permissions: %s", d.targetName)))
	b.WriteString("\n\n")

	// Current permission
	currentPerm := formatPermission(d.currentMode)
	currentSymbolic := fsops.FormatSymbolic(d.currentMode, d.isDir)
	currentStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 1).
		Foreground(lipgloss.Color("240"))
	b.WriteString(currentStyle.Render(fmt.Sprintf("Current: %s  →  %s", currentPerm, currentSymbolic)))
	b.WriteString("\n\n")

	// Input field with symbolic notation
	b.WriteString(d.renderInputField(width - 8))
	b.WriteString("\n\n")

	// Presets
	presetsStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 1).
		Foreground(lipgloss.Color("240"))
	b.WriteString(presetsStyle.Render("── Quick Presets ──"))
	b.WriteString("\n")

	for _, preset := range d.presets {
		presetStyle := lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1).
			Foreground(lipgloss.Color("248"))
		presetLine := fmt.Sprintf("[%d] %s  %s  (%s)", preset.Number, preset.Mode, preset.Symbolic, preset.Description)
		b.WriteString(presetStyle.Render(presetLine))
		b.WriteString("\n")
	}

	// Recursive option (directories only)
	if d.showRecursive {
		b.WriteString("\n")
		applyToStyle := lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1).
			Foreground(lipgloss.Color("240"))
		b.WriteString(applyToStyle.Render("Apply to:"))
		b.WriteString("\n")

		thisOnlyIcon := "( )"
		recursiveIcon := "( )"
		if d.recursiveOption == 0 {
			thisOnlyIcon = "(●)"
		} else {
			recursiveIcon = "(●)"
		}

		optionStyle := lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1).
			Foreground(lipgloss.Color("248"))
		b.WriteString(optionStyle.Render(fmt.Sprintf("  %s This directory only", thisOnlyIcon)))
		b.WriteString("\n")
		b.WriteString(optionStyle.Render(fmt.Sprintf("  %s Recursively (all contents)", recursiveIcon)))
		b.WriteString("\n")
	}

	// Error message (if any)
	if d.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1).
			Foreground(lipgloss.Color("196"))
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(d.errorMsg))
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	footerStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 1).
		Foreground(lipgloss.Color("240"))

	if d.showRecursive {
		b.WriteString(footerStyle.Render("[j/k/Space] Toggle  [Enter] Apply  [Esc] Cancel"))
	} else {
		b.WriteString(footerStyle.Render("[Enter] Apply  [Esc] Cancel"))
	}

	// Border
	boxStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

// renderInputField renders the input field with symbolic notation
func (d *PermissionDialog) renderInputField(width int) string {
	var b strings.Builder

	// Show input with cursor
	displayInput := d.inputValue
	if len(displayInput) == 0 {
		displayInput = "___"
	} else if len(displayInput) == 1 {
		displayInput = displayInput + "__"
	} else if len(displayInput) == 2 {
		displayInput = displayInput + "_"
	}

	// Add cursor at end
	cursorStyle := lipgloss.NewStyle().Reverse(true)
	inputWithCursor := displayInput
	if len(d.inputValue) < 3 {
		cursorPos := len(d.inputValue)
		if cursorPos < len(displayInput) {
			inputWithCursor = displayInput[:cursorPos] +
				cursorStyle.Render(string(displayInput[cursorPos])) +
				displayInput[cursorPos+1:]
		}
	}

	// Symbolic notation (real-time update)
	symbolic := "----------"
	if len(d.inputValue) == 3 {
		if mode, err := fsops.ParsePermissionMode(d.inputValue); err == nil {
			symbolic = fsops.FormatSymbolic(mode, d.isDir)
		}
	}

	inputLine := fmt.Sprintf("Mode: [%s]  →  %s", inputWithCursor, symbolic)

	inputStyle := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Foreground(lipgloss.Color("15"))

	b.WriteString(inputStyle.Render(inputLine))

	return b.String()
}

// IsActive returns whether the dialog is active
func (d *PermissionDialog) IsActive() bool {
	return d.active
}

// DisplayType returns the dialog display type
func (d *PermissionDialog) DisplayType() DialogDisplayType {
	return DialogDisplayPane
}

// formatPermission converts FileMode to octal string (e.g., 0644 -> "644")
func formatPermission(mode fs.FileMode) string {
	perm := mode.Perm()
	return fmt.Sprintf("%03o", perm)
}
