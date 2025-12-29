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
