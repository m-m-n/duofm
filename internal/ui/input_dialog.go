package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputDialog はテキスト入力ダイアログ
type InputDialog struct {
	title         string               // dialog title/prompt
	input         string               // current input text
	cursorPos     int                  // cursor position
	active        bool                 // whether dialog is active
	width         int                  // dialog width
	onConfirm     func(string) tea.Cmd // callback on Enter key
	errorMsg      string               // validation error message
	emptyErrorMsg string               // error message for empty input (customizable)
}

// NewInputDialog creates a new input dialog
func NewInputDialog(title string, onConfirm func(string) tea.Cmd) *InputDialog {
	return &InputDialog{
		title:         title,
		input:         "",
		cursorPos:     0,
		active:        true,
		width:         50,
		onConfirm:     onConfirm,
		errorMsg:      "",
		emptyErrorMsg: "Input cannot be empty",
	}
}

// SetEmptyErrorMsg sets a custom error message for empty input
func (d *InputDialog) SetEmptyErrorMsg(msg string) {
	d.emptyErrorMsg = msg
}

// Update handles messages
func (d *InputDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear error message on any key press
		d.errorMsg = ""

		switch msg.Type {
		case tea.KeyEnter:
			// Check for empty input
			if d.input == "" {
				d.errorMsg = d.emptyErrorMsg
				return d, nil
			}
			d.active = false
			if d.onConfirm != nil {
				return d, d.onConfirm(d.input)
			}
			// Defensive: if onConfirm is nil, send cancel message to avoid freeze
			return d, func() tea.Msg {
				return inputDialogResultMsg{
					cancelled: true,
				}
			}

		case tea.KeyEsc:
			d.active = false
			return d, func() tea.Msg {
				return inputDialogResultMsg{
					cancelled: true,
				}
			}

		case tea.KeyRunes:
			// Character input
			runes := []rune(d.input)
			newRunes := make([]rune, 0, len(runes)+len(msg.Runes))
			newRunes = append(newRunes, runes[:d.cursorPos]...)
			newRunes = append(newRunes, msg.Runes...)
			newRunes = append(newRunes, runes[d.cursorPos:]...)
			d.input = string(newRunes)
			d.cursorPos += len(msg.Runes)
			return d, nil

		case tea.KeyBackspace:
			if d.cursorPos > 0 {
				runes := []rune(d.input)
				newRunes := make([]rune, 0, len(runes)-1)
				newRunes = append(newRunes, runes[:d.cursorPos-1]...)
				newRunes = append(newRunes, runes[d.cursorPos:]...)
				d.input = string(newRunes)
				d.cursorPos--
			}
			return d, nil

		case tea.KeyDelete:
			runes := []rune(d.input)
			if d.cursorPos < len(runes) {
				newRunes := make([]rune, 0, len(runes)-1)
				newRunes = append(newRunes, runes[:d.cursorPos]...)
				newRunes = append(newRunes, runes[d.cursorPos+1:]...)
				d.input = string(newRunes)
			}
			return d, nil

		case tea.KeyLeft:
			if d.cursorPos > 0 {
				d.cursorPos--
			}
			return d, nil

		case tea.KeyRight:
			if d.cursorPos < len([]rune(d.input)) {
				d.cursorPos++
			}
			return d, nil

		case tea.KeyCtrlA:
			// Move to beginning of line
			d.cursorPos = 0
			return d, nil

		case tea.KeyCtrlE:
			// Move to end of line
			d.cursorPos = len([]rune(d.input))
			return d, nil

		case tea.KeyCtrlU:
			// Delete from cursor to beginning of line
			runes := []rune(d.input)
			d.input = string(runes[d.cursorPos:])
			d.cursorPos = 0
			return d, nil

		case tea.KeyCtrlK:
			// Delete from cursor to end of line
			runes := []rune(d.input)
			d.input = string(runes[:d.cursorPos])
			return d, nil
		}
	}

	return d, nil
}

// View renders the dialog
func (d *InputDialog) View() string {
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
	b.WriteString(titleStyle.Render(d.title))
	b.WriteString("\n\n")

	// Input field
	inputWidth := width - 8 // subtract padding and border
	b.WriteString(d.renderInputField(inputWidth))
	b.WriteString("\n")

	// Error message (if any)
	if d.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1).
			Foreground(lipgloss.Color("196")) // red color
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(d.errorMsg))
	}

	b.WriteString("\n")

	// Footer
	footerStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 1).
		Foreground(lipgloss.Color("240"))
	b.WriteString(footerStyle.Render("Enter: Confirm  Esc: Cancel"))

	// Border
	boxStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

// renderInputField renders the input field
func (d *InputDialog) renderInputField(width int) string {
	runes := []rune(d.input)
	displayInput := d.input

	// Calculate displayable range
	cursorDisplayPos := d.cursorPos
	startPos := 0

	if len(runes) > width-2 {
		// Adjust so cursor is within display range
		if d.cursorPos > width-3 {
			startPos = d.cursorPos - width + 3
		}
		endPos := startPos + width - 2
		if endPos > len(runes) {
			endPos = len(runes)
		}
		displayInput = string(runes[startPos:endPos])
		cursorDisplayPos = d.cursorPos - startPos
	}

	// Build display string with cursor
	displayRunes := []rune(displayInput)
	var result strings.Builder
	for i, r := range displayRunes {
		if i == cursorDisplayPos {
			// Reverse display for cursor position
			result.WriteString(lipgloss.NewStyle().Reverse(true).Render(string(r)))
		} else {
			result.WriteRune(r)
		}
	}
	// If cursor is at end, show block cursor
	if cursorDisplayPos >= len(displayRunes) {
		result.WriteString(lipgloss.NewStyle().Reverse(true).Render(" "))
	}

	// Input field style
	fieldStyle := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("236")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	return fieldStyle.Render(result.String())
}

// IsActive returns whether the dialog is active
func (d *InputDialog) IsActive() bool {
	return d.active
}

// DisplayType returns the dialog display type
func (d *InputDialog) DisplayType() DialogDisplayType {
	return DialogDisplayPane
}

// SetWidth sets the dialog width
func (d *InputDialog) SetWidth(width int) {
	d.width = width
}
