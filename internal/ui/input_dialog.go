package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// InputDialog はテキスト入力ダイアログ
type InputDialog struct {
	BaseDialog
	title         string               // dialog title/prompt
	textInput     *TextInput           // reusable text input component
	onConfirm     func(string) tea.Cmd // callback on Enter key
	errorMsg      string               // validation error message
	emptyErrorMsg string               // error message for empty input (customizable)
	styles        DialogStyles
}

// NewInputDialog creates a new input dialog
func NewInputDialog(title string, onConfirm func(string) tea.Cmd) *InputDialog {
	base := NewBaseDialog(DialogDisplayPane)
	return &InputDialog{
		BaseDialog:    base,
		title:         title,
		textInput:     NewTextInput(""),
		onConfirm:     onConfirm,
		errorMsg:      "",
		emptyErrorMsg: "Input cannot be empty",
		styles:        DefaultDialogStyles(base.Width()),
	}
}

// SetEmptyErrorMsg sets a custom error message for empty input
func (d *InputDialog) SetEmptyErrorMsg(msg string) {
	d.emptyErrorMsg = msg
}

// Update handles messages
func (d *InputDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
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
			d.Close()
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
			d.Close()
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
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder
	width := d.Width()

	// Title
	b.WriteString(d.styles.Title.Render(d.title))
	b.WriteString("\n\n")

	// Input field
	inputWidth := width - 8 // subtract padding and border
	b.WriteString(d.renderInputField(inputWidth))
	b.WriteString("\n")

	// Error message (if any)
	if d.errorMsg != "" {
		b.WriteString("\n")
		b.WriteString(d.styles.Error.Render(d.errorMsg))
	}

	b.WriteString("\n")

	// Footer
	b.WriteString(d.styles.Footer.Render("Enter: Confirm  Esc: Cancel"))

	return d.styles.Box.Render(b.String())
}

// renderInputField renders the input field
func (d *InputDialog) renderInputField(width int) string {
	inputStyle := d.styles.Input.Width(width)
	return inputStyle.Render(d.textInput.RenderWithCursor(width - 2))
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
