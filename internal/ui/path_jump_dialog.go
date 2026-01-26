package ui

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// pathJumpResultMsg is sent when user confirms a valid path.
type pathJumpResultMsg struct {
	path string
}

// pathJumpCancelMsg is sent when user cancels the dialog.
type pathJumpCancelMsg struct{}

// PathJumpDialog is a dialog for jumping to a directory by path.
// It provides inline autocomplete suggestions from the filesystem.
type PathJumpDialog struct {
	BaseDialog
	textInput  *TextInput
	suggester  *PathSuggester
	suggestion string // Current suggestion suffix
	errorMsg   string
	styles     DialogStyles
}

// NewPathJumpDialog creates a new path jump dialog.
func NewPathJumpDialog() *PathJumpDialog {
	base := NewBaseDialog(DialogDisplayPane)
	base.SetWidth(60) // Wider dialog for path display
	return &PathJumpDialog{
		BaseDialog: base,
		textInput:  NewTextInput(""),
		suggester:  NewPathSuggester(),
		suggestion: "",
		errorMsg:   "",
		styles:     DefaultDialogStyles(base.Width()),
	}
}

// Update handles input messages.
func (d *PathJumpDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear error on any key press
		d.errorMsg = ""

		switch msg.Type {
		case tea.KeyTab:
			// Confirm suggestion if present
			if d.suggestion != "" {
				d.textInput.Value += d.suggestion
				d.textInput.CursorPos = len([]rune(d.textInput.Value))
				d.updateSuggestion()
			}
			return d, nil

		case tea.KeyEnter:
			path := d.textInput.Value

			// Validate path
			errMsg := d.validatePath(path)
			if errMsg != "" {
				d.errorMsg = errMsg
				return d, nil
			}

			// Close dialog and send result
			d.Close()
			return d, func() tea.Msg {
				return pathJumpResultMsg{path: path}
			}

		case tea.KeyEsc, tea.KeyCtrlC:
			d.Close()
			return d, func() tea.Msg {
				return pathJumpCancelMsg{}
			}

		default:
			// Delegate text editing to TextInput
			if d.textInput.HandleKey(msg) {
				d.updateSuggestion()
				return d, nil
			}
		}
	}

	return d, nil
}

// validatePath checks if the path is valid for navigation.
// Returns empty string if valid, or error message if invalid.
func (d *PathJumpDialog) validatePath(path string) string {
	if path == "" {
		return "Path cannot be empty"
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "Directory does not exist: " + path
		}
		if os.IsPermission(err) {
			return "Permission denied: " + path
		}
		return "Error accessing path: " + path
	}

	if !info.IsDir() {
		return "Not a directory: " + path
	}

	return ""
}

// updateSuggestion recalculates the suggestion based on current input.
func (d *PathJumpDialog) updateSuggestion() {
	d.suggestion = d.suggester.Suggest(d.textInput.Value)
}

// View renders the dialog.
func (d *PathJumpDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder
	width := d.Width()

	// Title
	b.WriteString(d.styles.Title.Render("Jump to Directory"))
	b.WriteString("\n\n")

	// Input field with suggestion
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
	b.WriteString(d.styles.Footer.Render("Tab: Complete  Enter: Jump  Esc: Cancel"))

	return d.styles.Box.Render(b.String())
}

// renderInputField renders the input field with inline suggestion.
func (d *PathJumpDialog) renderInputField(width int) string {
	inputStyle := d.styles.Input.Width(width)

	// Build input display with cursor and suggestion
	content := d.renderInputWithSuggestion(width - 2)

	return inputStyle.Render(content)
}

// renderInputWithSuggestion renders the input text with cursor and grayed-out suggestion.
func (d *PathJumpDialog) renderInputWithSuggestion(width int) string {
	runes := []rune(d.textInput.Value)
	displayInput := d.textInput.Value
	cursorDisplayPos := d.textInput.CursorPos
	startPos := 0

	// Calculate total length including suggestion for scrolling
	totalLen := len(runes) + len([]rune(d.suggestion))

	// Apply horizontal scrolling if content is longer than width
	if width > 0 && totalLen > width-2 {
		if d.textInput.CursorPos > width-3 {
			startPos = d.textInput.CursorPos - width + 3
		}
		endPos := startPos + width - 2 - len([]rune(d.suggestion))
		if endPos > len(runes) {
			endPos = len(runes)
		}
		if startPos < len(runes) && endPos > startPos {
			displayInput = string(runes[startPos:endPos])
		} else if startPos >= len(runes) {
			displayInput = ""
		}
		cursorDisplayPos = d.textInput.CursorPos - startPos
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

	// If cursor is at end of input, show block cursor
	if cursorDisplayPos >= len(displayRunes) {
		if d.suggestion != "" {
			// Show first char of suggestion with cursor styling
			suggRunes := []rune(d.suggestion)
			if len(suggRunes) > 0 {
				result.WriteString(lipgloss.NewStyle().Reverse(true).Foreground(lipgloss.Color(string(ColorMuted))).Render(string(suggRunes[0])))
				// Show rest of suggestion grayed out
				if len(suggRunes) > 1 {
					result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(string(ColorMuted))).Render(string(suggRunes[1:])))
				}
			}
		} else {
			result.WriteString(lipgloss.NewStyle().Reverse(true).Render(" "))
		}
	} else if d.suggestion != "" {
		// Cursor is in middle, show suggestion after input
		result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(string(ColorMuted))).Render(d.suggestion))
	}

	return result.String()
}

// SetInput sets the input value and positions cursor at the end.
func (d *PathJumpDialog) SetInput(value string) {
	d.textInput.Value = value
	d.textInput.CursorPos = len([]rune(value))
	d.updateSuggestion()
}

// Input returns the current input value.
func (d *PathJumpDialog) Input() string {
	return d.textInput.Value
}

// CursorPos returns the current cursor position.
func (d *PathJumpDialog) CursorPos() int {
	return d.textInput.CursorPos
}
