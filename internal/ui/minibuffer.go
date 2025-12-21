package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Minibuffer is a single-line text input area displayed at the bottom of the pane
type Minibuffer struct {
	prompt    string
	input     string
	cursorPos int
	visible   bool
	width     int
}

// NewMinibuffer creates a new minibuffer instance
func NewMinibuffer() *Minibuffer {
	return &Minibuffer{
		prompt:    "",
		input:     "",
		cursorPos: 0,
		visible:   false,
		width:     80,
	}
}

// SetPrompt sets the prompt text (e.g., "/: " or "(search): ")
func (m *Minibuffer) SetPrompt(prompt string) {
	m.prompt = prompt
}

// SetWidth sets the display width
func (m *Minibuffer) SetWidth(width int) {
	m.width = width
}

// Clear resets the input and cursor position
func (m *Minibuffer) Clear() {
	m.input = ""
	m.cursorPos = 0
}

// Show makes the minibuffer visible
func (m *Minibuffer) Show() {
	m.visible = true
}

// Hide makes the minibuffer invisible
func (m *Minibuffer) Hide() {
	m.visible = false
}

// Input returns the current input text
func (m *Minibuffer) Input() string {
	return m.input
}

// IsVisible returns whether the minibuffer is visible
func (m *Minibuffer) IsVisible() bool {
	return m.visible
}

// HandleKey processes a key message and returns true if handled
func (m *Minibuffer) HandleKey(msg tea.KeyMsg) bool {
	if !m.visible {
		return false
	}

	switch msg.Type {
	case tea.KeyRunes:
		// Insert character at cursor position
		runes := []rune(m.input)
		newRunes := make([]rune, 0, len(runes)+len(msg.Runes))
		newRunes = append(newRunes, runes[:m.cursorPos]...)
		newRunes = append(newRunes, msg.Runes...)
		newRunes = append(newRunes, runes[m.cursorPos:]...)
		m.input = string(newRunes)
		m.cursorPos += len(msg.Runes)
		return true

	case tea.KeyBackspace:
		if m.cursorPos > 0 {
			runes := []rune(m.input)
			newRunes := make([]rune, 0, len(runes)-1)
			newRunes = append(newRunes, runes[:m.cursorPos-1]...)
			newRunes = append(newRunes, runes[m.cursorPos:]...)
			m.input = string(newRunes)
			m.cursorPos--
		}
		return true

	case tea.KeyDelete:
		runes := []rune(m.input)
		if m.cursorPos < len(runes) {
			newRunes := make([]rune, 0, len(runes)-1)
			newRunes = append(newRunes, runes[:m.cursorPos]...)
			newRunes = append(newRunes, runes[m.cursorPos+1:]...)
			m.input = string(newRunes)
		}
		return true

	case tea.KeyLeft:
		if m.cursorPos > 0 {
			m.cursorPos--
		}
		return true

	case tea.KeyRight:
		if m.cursorPos < len([]rune(m.input)) {
			m.cursorPos++
		}
		return true

	case tea.KeyCtrlA:
		// Move to beginning
		m.cursorPos = 0
		return true

	case tea.KeyCtrlE:
		// Move to end
		m.cursorPos = len([]rune(m.input))
		return true

	case tea.KeyCtrlK:
		// Kill to end of line
		runes := []rune(m.input)
		m.input = string(runes[:m.cursorPos])
		return true

	case tea.KeyCtrlU:
		// Kill to beginning of line
		runes := []rune(m.input)
		m.input = string(runes[m.cursorPos:])
		m.cursorPos = 0
		return true

	case tea.KeyCtrlB:
		// Move backward (same as left)
		if m.cursorPos > 0 {
			m.cursorPos--
		}
		return true

	case tea.KeyCtrlF:
		// Move forward (same as right)
		// Note: This conflicts with regex search key, so it's only active when minibuffer is visible
		if m.cursorPos < len([]rune(m.input)) {
			m.cursorPos++
		}
		return true
	}

	return false
}

// View renders the minibuffer
func (m *Minibuffer) View() string {
	if !m.visible {
		return ""
	}

	// Calculate available width for input
	promptLen := len(m.prompt)
	availableWidth := m.width - promptLen - 2 // -2 for padding

	// Build the display string
	runes := []rune(m.input)
	displayInput := m.input

	// Calculate visible portion if input is too long
	cursorDisplayPos := m.cursorPos
	startPos := 0

	if len(runes) > availableWidth {
		// Ensure cursor is visible
		if m.cursorPos > availableWidth-1 {
			startPos = m.cursorPos - availableWidth + 1
		}
		endPos := startPos + availableWidth
		if endPos > len(runes) {
			endPos = len(runes)
		}
		displayInput = string(runes[startPos:endPos])
		cursorDisplayPos = m.cursorPos - startPos
	}

	// Build cursor display
	displayRunes := []rune(displayInput)
	var result strings.Builder
	for i, r := range displayRunes {
		if i == cursorDisplayPos {
			// Highlight cursor position
			result.WriteString(lipgloss.NewStyle().Reverse(true).Render(string(r)))
		} else {
			result.WriteRune(r)
		}
	}
	// If cursor is at end, show a block cursor
	if cursorDisplayPos >= len(displayRunes) {
		result.WriteString(lipgloss.NewStyle().Reverse(true).Render(" "))
	}

	// Style the whole minibuffer line
	style := lipgloss.NewStyle().
		Width(m.width-2).
		Padding(0, 1).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("236"))

	return style.Render(m.prompt + result.String())
}
