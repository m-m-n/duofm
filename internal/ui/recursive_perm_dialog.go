package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	fsops "github.com/sakura/duofm/internal/fs"
)

// RecursivePermDialog is a two-step dialog for recursive permission changes
type RecursivePermDialog struct {
	targetName   string
	step         int    // 0: dir input, 1: file input
	dirMode      string // directory permission (from step 0)
	fileMode     string // file permission (from step 1)
	currentInput string
	cursorPos    int
	errorMsg     string
	active       bool
	width        int
	onConfirm    func(dirMode, fileMode string) tea.Cmd
}

// NewRecursivePermDialog creates a new recursive permission dialog
func NewRecursivePermDialog(targetName string) *RecursivePermDialog {
	return &RecursivePermDialog{
		targetName:   targetName,
		step:         0,
		dirMode:      "",
		fileMode:     "",
		currentInput: "",
		cursorPos:    0,
		errorMsg:     "",
		active:       true,
		width:        50,
		onConfirm:    nil,
	}
}

// SetOnConfirm sets the confirmation callback
func (d *RecursivePermDialog) SetOnConfirm(callback func(dirMode, fileMode string) tea.Cmd) {
	d.onConfirm = callback
}

// Update handles messages
func (d *RecursivePermDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear error message on any key press
		d.errorMsg = ""

		switch msg.Type {
		case tea.KeyEnter:
			// Validate current input
			if err := fsops.ValidatePermissionMode(d.currentInput); err != nil {
				d.errorMsg = err.Error()
				return d, nil
			}

			// If we're on step 0 (directory permission), save and move to step 1
			if d.step == 0 {
				d.dirMode = d.currentInput
				d.currentInput = ""
				d.cursorPos = 0
				d.step = 1
				return d, nil
			}

			// If we're on step 1 (file permission), execute operation
			d.fileMode = d.currentInput
			d.active = false
			if d.onConfirm != nil {
				return d, d.onConfirm(d.dirMode, d.fileMode)
			}
			return d, nil

		case tea.KeyEsc:
			d.active = false
			return d, nil

		case tea.KeyBackspace:
			if len(d.currentInput) > 0 {
				runes := []rune(d.currentInput)
				d.currentInput = string(runes[:len(runes)-1])
				d.cursorPos = len(d.currentInput)
			}
			return d, nil

		case tea.KeyRunes:
			// Handle digit input and presets
			if len(msg.Runes) > 0 {
				for _, r := range msg.Runes {
					// Check if it's a preset number (1-4) only when input is empty
					if r >= '1' && r <= '4' && len(d.currentInput) == 0 {
						presetIdx := int(r - '1')
						var presets []PermissionPreset
						if d.step == 0 {
							presets = dirPresets
						} else {
							presets = filePresets
						}

						if presetIdx < len(presets) {
							d.currentInput = presets[presetIdx].Mode
							d.cursorPos = len(d.currentInput)
							return d, nil
						}
					}

					// Check if it's a valid octal digit (0-7)
					if r >= '0' && r <= '7' {
						// Limit to 3 digits
						if len(d.currentInput) < 3 {
							d.currentInput += string(r)
							d.cursorPos++
						}
						// Silently ignore 4th+ digits
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
func (d *RecursivePermDialog) View() string {
	if !d.active {
		return ""
	}

	var b strings.Builder
	width := d.width

	// Title with step indicator
	titleStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 1).
		Bold(true).
		Foreground(lipgloss.Color("39"))

	stepText := fmt.Sprintf("Recursive Permissions (%d/2)", d.step+1)
	b.WriteString(titleStyle.Render(stepText))
	b.WriteString("\n\n")

	// Step-specific content
	if d.step == 0 {
		// Step 1: Directory permissions
		labelStyle := lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1).
			Foreground(lipgloss.Color("248"))
		b.WriteString(labelStyle.Render("Permissions for DIRECTORIES:"))
		b.WriteString("\n\n")

		// Input field
		b.WriteString(d.renderInputField(width-8, true))
		b.WriteString("\n\n")

		// Directory presets
		presetsStyle := lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1).
			Foreground(lipgloss.Color("240"))
		b.WriteString(presetsStyle.Render("── Quick Presets ──"))
		b.WriteString("\n")

		for _, preset := range dirPresets {
			presetStyle := lipgloss.NewStyle().
				Width(width-4).
				Padding(0, 1).
				Foreground(lipgloss.Color("248"))
			presetLine := fmt.Sprintf("[%d] %s  %s  (%s)", preset.Number, preset.Mode, preset.Symbolic, preset.Description)
			b.WriteString(presetStyle.Render(presetLine))
			b.WriteString("\n")
		}
	} else {
		// Step 2: File permissions
		infoStyle := lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1).
			Foreground(lipgloss.Color("248"))
		b.WriteString(infoStyle.Render("Permissions for FILES:"))
		b.WriteString("\n")
		b.WriteString(infoStyle.Render(fmt.Sprintf("(Directories will use: %s)", d.dirMode)))
		b.WriteString("\n\n")

		// Input field
		b.WriteString(d.renderInputField(width-8, false))
		b.WriteString("\n\n")

		// File presets
		presetsStyle := lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1).
			Foreground(lipgloss.Color("240"))
		b.WriteString(presetsStyle.Render("── Quick Presets ──"))
		b.WriteString("\n")

		for _, preset := range filePresets {
			presetStyle := lipgloss.NewStyle().
				Width(width-4).
				Padding(0, 1).
				Foreground(lipgloss.Color("248"))
			presetLine := fmt.Sprintf("[%d] %s  %s  (%s)", preset.Number, preset.Mode, preset.Symbolic, preset.Description)
			b.WriteString(presetStyle.Render(presetLine))
			b.WriteString("\n")
		}
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

	if d.step == 0 {
		b.WriteString(footerStyle.Render("[Enter] Next  [Esc] Cancel"))
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
func (d *RecursivePermDialog) renderInputField(width int, isDir bool) string {
	var b strings.Builder

	// Show input with cursor
	displayInput := d.currentInput
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
	if len(d.currentInput) < 3 {
		cursorPos := len(d.currentInput)
		if cursorPos < len(displayInput) {
			inputWithCursor = displayInput[:cursorPos] +
				cursorStyle.Render(string(displayInput[cursorPos])) +
				displayInput[cursorPos+1:]
		}
	}

	// Symbolic notation (real-time update)
	symbolic := "----------"
	if len(d.currentInput) == 3 {
		if mode, err := fsops.ParsePermissionMode(d.currentInput); err == nil {
			symbolic = fsops.FormatSymbolic(mode, isDir)
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
func (d *RecursivePermDialog) IsActive() bool {
	return d.active
}

// DisplayType returns the dialog display type
func (d *RecursivePermDialog) DisplayType() DialogDisplayType {
	return DialogDisplayPane
}
