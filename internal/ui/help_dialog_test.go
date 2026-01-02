package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHelpDialogContainsShellCommand(t *testing.T) {
	dialog := NewHelpDialog()
	view := dialog.View()

	// Check that the help dialog contains the shell command key binding
	if !strings.Contains(view, "!") {
		t.Error("Help dialog should contain '!' key binding for shell command")
	}
}

func TestHelpDialogScrolling(t *testing.T) {
	dialog := NewHelpDialog()

	// Initial state
	if dialog.scrollOffset != 0 {
		t.Errorf("Initial scrollOffset = %d, want 0", dialog.scrollOffset)
	}

	// Scroll down
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if dialog.scrollOffset != 1 {
		t.Errorf("After j, scrollOffset = %d, want 1", dialog.scrollOffset)
	}

	// Scroll up
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if dialog.scrollOffset != 0 {
		t.Errorf("After k, scrollOffset = %d, want 0", dialog.scrollOffset)
	}

	// Don't scroll above 0
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if dialog.scrollOffset != 0 {
		t.Errorf("After extra k, scrollOffset = %d, want 0", dialog.scrollOffset)
	}

	// Page down with space
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if dialog.scrollOffset != dialog.visibleHeight {
		t.Errorf("After space, scrollOffset = %d, want %d", dialog.scrollOffset, dialog.visibleHeight)
	}

	// Go to top
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if dialog.scrollOffset != 0 {
		t.Errorf("After g, scrollOffset = %d, want 0", dialog.scrollOffset)
	}

	// Go to end
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	maxOffset := len(dialog.contentLines) - dialog.visibleHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if dialog.scrollOffset != maxOffset {
		t.Errorf("After G, scrollOffset = %d, want %d", dialog.scrollOffset, maxOffset)
	}
}

func TestHelpDialogContentHasColorPalette(t *testing.T) {
	dialog := NewHelpDialog()

	// Check content contains color palette section
	contentStr := strings.Join(dialog.contentLines, "\n")

	if !strings.Contains(contentStr, "Color Palette Reference") {
		t.Error("Content should contain 'Color Palette Reference'")
	}
	if !strings.Contains(contentStr, "Standard Colors (0-15)") {
		t.Error("Content should contain 'Standard Colors (0-15)'")
	}
	if !strings.Contains(contentStr, "6x6x6 Color Cube (16-231)") {
		t.Error("Content should contain '6x6x6 Color Cube (16-231)'")
	}
	if !strings.Contains(contentStr, "Grayscale (232-255)") {
		t.Error("Content should contain 'Grayscale (232-255)'")
	}
}

func TestColorCubeToHex(t *testing.T) {
	tests := []struct {
		index    int
		expected string
	}{
		{0, "#000000"},   // 16: black (r=0, g=0, b=0)
		{1, "#00005f"},   // 17: (r=0, g=0, b=1)
		{6, "#005f00"},   // 22: green (r=0, g=1, b=0)
		{36, "#5f0000"},  // 52: red (r=1, g=0, b=0)
		{215, "#ffffff"}, // 231: white (r=5, g=5, b=5)
	}

	for _, tt := range tests {
		result := colorCubeToHex(tt.index)
		if result != tt.expected {
			t.Errorf("colorCubeToHex(%d) = %s, want %s", tt.index, result, tt.expected)
		}
	}
}

func TestGrayscaleToHex(t *testing.T) {
	tests := []struct {
		index    int
		expected string
	}{
		{0, "#080808"},  // 232
		{1, "#121212"},  // 233
		{23, "#eeeeee"}, // 255
	}

	for _, tt := range tests {
		result := grayscaleToHex(tt.index)
		if result != tt.expected {
			t.Errorf("grayscaleToHex(%d) = %s, want %s", tt.index, result, tt.expected)
		}
	}
}

func TestHelpDialogPageIndicator(t *testing.T) {
	dialog := NewHelpDialog()
	view := dialog.View()

	// View should contain page indicator
	if !strings.Contains(view, "[1/") {
		t.Error("Help dialog should contain page indicator [1/N]")
	}
}

func TestHelpDialog_CloseKeys(t *testing.T) {
	t.Run("Esc closes dialog", func(t *testing.T) {
		dialog := NewHelpDialog()
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})
		d := updated.(*HelpDialog)

		if d.IsActive() {
			t.Error("Dialog should be inactive after Esc")
		}
		if cmd == nil {
			t.Error("Esc should return a command")
		}
	})

	t.Run("? closes dialog", func(t *testing.T) {
		dialog := NewHelpDialog()
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
		d := updated.(*HelpDialog)

		if d.IsActive() {
			t.Error("Dialog should be inactive after ?")
		}
		if cmd == nil {
			t.Error("? should return a command")
		}
	})

	t.Run("Ctrl+C closes dialog", func(t *testing.T) {
		dialog := NewHelpDialog()
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		d := updated.(*HelpDialog)

		if d.IsActive() {
			t.Error("Dialog should be inactive after Ctrl+C")
		}
		if cmd == nil {
			t.Error("Ctrl+C should return a command")
		}
	})
}

func TestHelpDialog_InactiveIgnoresInput(t *testing.T) {
	dialog := NewHelpDialog()
	dialog.active = false

	updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if updated == nil {
		t.Error("Updated dialog should not be nil")
	}
	if cmd != nil {
		t.Error("Command should be nil for inactive dialog")
	}
}

func TestHelpDialog_InactiveReturnsEmpty(t *testing.T) {
	dialog := NewHelpDialog()
	dialog.active = false
	view := dialog.View()

	if view != "" {
		t.Error("View should be empty for inactive dialog")
	}
}

func TestHelpDialog_IsActive(t *testing.T) {
	dialog := NewHelpDialog()

	if !dialog.IsActive() {
		t.Error("IsActive() should return true for new dialog")
	}

	dialog.active = false
	if dialog.IsActive() {
		t.Error("IsActive() should return false for inactive dialog")
	}
}

func TestHelpDialog_DisplayType(t *testing.T) {
	dialog := NewHelpDialog()
	if dialog.DisplayType() != DialogDisplayScreen {
		t.Errorf("DisplayType() = %v, want DialogDisplayScreen", dialog.DisplayType())
	}
}

func TestHelpDialog_DownArrowKey(t *testing.T) {
	dialog := NewHelpDialog()
	dialog.Update(tea.KeyMsg{Type: tea.KeyDown})

	if dialog.scrollOffset != 1 {
		t.Errorf("After down arrow, scrollOffset = %d, want 1", dialog.scrollOffset)
	}
}

func TestHelpDialog_UpArrowKey(t *testing.T) {
	dialog := NewHelpDialog()
	dialog.scrollOffset = 5
	dialog.Update(tea.KeyMsg{Type: tea.KeyUp})

	if dialog.scrollOffset != 4 {
		t.Errorf("After up arrow, scrollOffset = %d, want 4", dialog.scrollOffset)
	}
}

func TestHelpDialog_ScrollDown_MaxOffset(t *testing.T) {
	dialog := NewHelpDialog()
	// Set visible height greater than content to test maxOffset < 0 case
	dialog.visibleHeight = len(dialog.contentLines) + 10

	dialog.scrollDown(5)

	if dialog.scrollOffset != 0 {
		t.Errorf("scrollOffset = %d, want 0 when content fits in view", dialog.scrollOffset)
	}
}

func TestHelpDialog_ScrollToEnd_ShortContent(t *testing.T) {
	dialog := NewHelpDialog()
	// Set visible height greater than content
	dialog.visibleHeight = len(dialog.contentLines) + 10

	dialog.scrollToEnd()

	if dialog.scrollOffset != 0 {
		t.Errorf("scrollOffset = %d, want 0 when content fits in view", dialog.scrollOffset)
	}
}
