package ui

import (
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// RegexSearchDialog is a dialog for regex search with syntax hints and history navigation.
type RegexSearchDialog struct {
	BaseDialog
	textInput *TextInput
	history   *SearchHistory
	errorMsg  string
	styles    DialogStyles
}

// NewRegexSearchDialog creates a new regex search dialog.
// The history is shared with Model and persists across dialog sessions.
func NewRegexSearchDialog(history *SearchHistory) *RegexSearchDialog {
	base := NewBaseDialog(DialogDisplayPane)
	history.Reset()
	return &RegexSearchDialog{
		BaseDialog: base,
		textInput:  NewTextInput(""),
		history:    history,
		errorMsg:   "",
		styles:     DefaultDialogStyles(base.Width()),
	}
}

// Update handles keyboard input for the dialog.
func (d *RegexSearchDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear error message on any key press
		d.errorMsg = ""

		switch msg.Type {
		case tea.KeyEnter:
			pattern := d.textInput.Value
			// Empty input clears filter
			if pattern == "" {
				d.Close()
				return d, func() tea.Msg {
					return regexSearchResultMsg{pattern: ""}
				}
			}
			// Validate regex
			if _, err := regexp.Compile(pattern); err != nil {
				d.errorMsg = "Invalid regex: " + err.Error()
				return d, nil
			}
			// Add to history and return result
			d.history.Add(pattern)
			d.Close()
			return d, func() tea.Msg {
				return regexSearchResultMsg{pattern: pattern}
			}

		case tea.KeyEsc:
			d.Close()
			return d, func() tea.Msg {
				return regexSearchResultMsg{cancelled: true}
			}

		case tea.KeyUp:
			newValue := d.history.NavigateUp(d.textInput.Value)
			d.textInput.SetValue(newValue)
			d.textInput.MoveCursorToEnd()
			return d, nil

		case tea.KeyDown:
			newValue := d.history.NavigateDown()
			d.textInput.SetValue(newValue)
			d.textInput.MoveCursorToEnd()
			return d, nil

		default:
			// Delegate text editing to TextInput
			d.textInput.HandleKey(msg)
		}
	}

	return d, nil
}

// View renders the dialog.
func (d *RegexSearchDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder
	width := d.Width()

	// Title
	b.WriteString(d.styles.Title.Render("Regex Search"))
	b.WriteString("\n\n")

	// Input field
	inputWidth := width - 8 // subtract padding and border
	inputStyle := d.styles.Input.Width(inputWidth)
	b.WriteString(inputStyle.Render(d.textInput.RenderWithCursor(inputWidth - 2)))
	b.WriteString("\n")

	// Error message (if any)
	if d.errorMsg != "" {
		b.WriteString("\n")
		b.WriteString(d.styles.Error.Render(d.errorMsg))
	}

	b.WriteString("\n")

	// Syntax hints
	b.WriteString(d.styles.Footer.Render("Examples: ^prefix  suffix$  \\.txt$"))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(d.styles.Footer.Render("Enter: Search  Esc: Cancel  Up/Down: History"))

	return d.styles.Box.Render(b.String())
}
