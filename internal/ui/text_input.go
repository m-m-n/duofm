package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TextInput provides reusable text input handling with cursor management.
// It encapsulates common text editing operations like cursor movement,
// character insertion/deletion, and Emacs-style keybindings.
type TextInput struct {
	Value     string // Current input text
	CursorPos int    // Cursor position (in runes)
}

// NewTextInput creates a new TextInput with initial value.
// Cursor is positioned at the end of the initial value.
func NewTextInput(initialValue string) *TextInput {
	return &TextInput{
		Value:     initialValue,
		CursorPos: len([]rune(initialValue)),
	}
}

// HandleKey processes keyboard input for text editing.
// Returns true if the key was handled, false otherwise.
// The caller should handle application-specific keys like Enter and Escape.
func (t *TextInput) HandleKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyRunes:
		t.InsertRunes(msg.Runes)
		return true

	case tea.KeySpace:
		// Space key is handled separately from KeyRunes in Bubble Tea
		t.InsertRunes([]rune{' '})
		return true

	case tea.KeyBackspace:
		t.DeleteBackward()
		return true

	case tea.KeyDelete:
		t.DeleteForward()
		return true

	case tea.KeyLeft:
		t.MoveCursorLeft()
		return true

	case tea.KeyRight:
		t.MoveCursorRight()
		return true

	case tea.KeyHome, tea.KeyCtrlA:
		t.MoveCursorToStart()
		return true

	case tea.KeyEnd, tea.KeyCtrlE:
		t.MoveCursorToEnd()
		return true

	case tea.KeyCtrlU:
		t.DeleteToStart()
		return true

	case tea.KeyCtrlK:
		t.DeleteToEnd()
		return true
	}

	return false
}

// InsertRunes inserts runes at the current cursor position.
func (t *TextInput) InsertRunes(runes []rune) {
	currentRunes := []rune(t.Value)
	newRunes := make([]rune, 0, len(currentRunes)+len(runes))
	newRunes = append(newRunes, currentRunes[:t.CursorPos]...)
	newRunes = append(newRunes, runes...)
	newRunes = append(newRunes, currentRunes[t.CursorPos:]...)
	t.Value = string(newRunes)
	t.CursorPos += len(runes)
}

// DeleteBackward deletes the character before the cursor (Backspace).
func (t *TextInput) DeleteBackward() {
	if t.CursorPos > 0 {
		runes := []rune(t.Value)
		newRunes := make([]rune, 0, len(runes)-1)
		newRunes = append(newRunes, runes[:t.CursorPos-1]...)
		newRunes = append(newRunes, runes[t.CursorPos:]...)
		t.Value = string(newRunes)
		t.CursorPos--
	}
}

// DeleteForward deletes the character at the cursor (Delete key).
func (t *TextInput) DeleteForward() {
	runes := []rune(t.Value)
	if t.CursorPos < len(runes) {
		newRunes := make([]rune, 0, len(runes)-1)
		newRunes = append(newRunes, runes[:t.CursorPos]...)
		newRunes = append(newRunes, runes[t.CursorPos+1:]...)
		t.Value = string(newRunes)
	}
}

// MoveCursorLeft moves the cursor one position to the left.
func (t *TextInput) MoveCursorLeft() {
	if t.CursorPos > 0 {
		t.CursorPos--
	}
}

// MoveCursorRight moves the cursor one position to the right.
func (t *TextInput) MoveCursorRight() {
	if t.CursorPos < len([]rune(t.Value)) {
		t.CursorPos++
	}
}

// MoveCursorToStart moves the cursor to the beginning of the line.
func (t *TextInput) MoveCursorToStart() {
	t.CursorPos = 0
}

// MoveCursorToEnd moves the cursor to the end of the line.
func (t *TextInput) MoveCursorToEnd() {
	t.CursorPos = len([]rune(t.Value))
}

// DeleteToStart deletes from cursor to the beginning of the line (Ctrl+U).
func (t *TextInput) DeleteToStart() {
	runes := []rune(t.Value)
	t.Value = string(runes[t.CursorPos:])
	t.CursorPos = 0
}

// DeleteToEnd deletes from cursor to the end of the line (Ctrl+K).
func (t *TextInput) DeleteToEnd() {
	runes := []rune(t.Value)
	t.Value = string(runes[:t.CursorPos])
}

// SetValue sets the input value and adjusts cursor position if needed.
func (t *TextInput) SetValue(value string) {
	t.Value = value
	runeLen := len([]rune(value))
	if t.CursorPos > runeLen {
		t.CursorPos = runeLen
	}
}

// Clear clears the input value and resets cursor position.
func (t *TextInput) Clear() {
	t.Value = ""
	t.CursorPos = 0
}

// IsEmpty returns true if the input value is empty.
func (t *TextInput) IsEmpty() bool {
	return t.Value == ""
}

// Len returns the length of the input value in runes.
func (t *TextInput) Len() int {
	return len([]rune(t.Value))
}

// RenderWithCursor renders the input text with a visible cursor.
// width specifies the display width for scrolling long text.
// If width is 0 or negative, no scrolling is applied.
func (t *TextInput) RenderWithCursor(width int) string {
	runes := []rune(t.Value)
	displayInput := t.Value
	cursorDisplayPos := t.CursorPos
	startPos := 0

	// Apply horizontal scrolling if text is longer than width
	if width > 0 && len(runes) > width-2 {
		if t.CursorPos > width-3 {
			startPos = t.CursorPos - width + 3
		}
		endPos := min(startPos+width-2, len(runes))
		displayInput = string(runes[startPos:endPos])
		cursorDisplayPos = t.CursorPos - startPos
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

	return result.String()
}
