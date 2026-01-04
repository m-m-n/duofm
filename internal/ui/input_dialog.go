package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputDialog はテキスト入力ダイアログ
type InputDialog struct {
	title         string               // dialog title/prompt
	textInput     *TextInput           // reusable text input component
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
		textInput:     NewTextInput(""),
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
			if d.textInput.IsEmpty() {
				d.errorMsg = d.emptyErrorMsg
				return d, nil
			}
			d.active = false
			if d.onConfirm != nil {
				return d, d.onConfirm(d.textInput.Value)
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

		default:
			// Delegate text editing to TextInput
			if d.textInput.HandleKey(msg) {
				return d, nil
			}
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
	// Input field style
	fieldStyle := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("236")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	return fieldStyle.Render(d.textInput.RenderWithCursor(width - 2))
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

// SetInput sets the input value and positions cursor at the end.
// This is useful for pre-populating the input field.
func (d *InputDialog) SetInput(value string) {
	d.textInput.Value = value
	d.textInput.CursorPos = len([]rune(value))
}

// Input returns the current input value (for testing).
func (d *InputDialog) Input() string {
	return d.textInput.Value
}

// CursorPos returns the current cursor position (for testing).
func (d *InputDialog) CursorPos() int {
	return d.textInput.CursorPos
}
