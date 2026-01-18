package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/filter"
)

// QuerySearchDialog is a dialog for SQL-like query search with syntax hints and history navigation.
type QuerySearchDialog struct {
	BaseDialog
	textInput *TextInput
	history   *SearchHistory
	errorMsg  string
	styles    DialogStyles
}

// NewQuerySearchDialog creates a new query search dialog.
// The history is shared with Model and persists across dialog sessions.
func NewQuerySearchDialog(history *SearchHistory) *QuerySearchDialog {
	base := NewBaseDialog(DialogDisplayPane)
	history.Reset()
	return &QuerySearchDialog{
		BaseDialog: base,
		textInput:  NewTextInput(""),
		history:    history,
		errorMsg:   "",
		styles:     DefaultDialogStyles(base.Width()),
	}
}

// Update handles keyboard input for the dialog.
func (d *QuerySearchDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear error message on any key press
		d.errorMsg = ""

		switch msg.Type {
		case tea.KeyEnter:
			query := d.textInput.Value
			// Empty input clears filter
			if query == "" {
				d.Close()
				return d, func() tea.Msg {
					return querySearchResultMsg{query: ""}
				}
			}
			// Validate query
			if err := filter.ValidateQuery(query); err != nil {
				d.errorMsg = "Invalid query: " + err.Error()
				return d, nil
			}
			// Add to history and return result
			d.history.Add(query)
			d.Close()
			return d, func() tea.Msg {
				return querySearchResultMsg{query: query}
			}

		case tea.KeyEsc:
			d.Close()
			return d, func() tea.Msg {
				return querySearchResultMsg{cancelled: true}
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
func (d *QuerySearchDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder
	width := d.Width()

	// Title
	b.WriteString(d.styles.Title.Render("Query Filter"))
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

	// Syntax hints (2 lines)
	b.WriteString(d.styles.Footer.Render("Examples: size > 1MB  ext = \".go\""))
	b.WriteString("\n")
	b.WriteString(d.styles.Footer.Render("          name LIKE \"test%\""))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(d.styles.Footer.Render("Enter: Filter  Esc: Cancel  Up/Down: History"))

	return d.styles.Box.Render(b.String())
}
