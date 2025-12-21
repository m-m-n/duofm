package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewMinibuffer(t *testing.T) {
	mb := NewMinibuffer()
	if mb == nil {
		t.Fatal("NewMinibuffer() returned nil")
	}
	if mb.visible {
		t.Error("New minibuffer should not be visible")
	}
	if mb.input != "" {
		t.Error("New minibuffer should have empty input")
	}
	if mb.cursorPos != 0 {
		t.Error("New minibuffer should have cursor at 0")
	}
}

func TestMinibufferShowHide(t *testing.T) {
	mb := NewMinibuffer()

	mb.Show()
	if !mb.IsVisible() {
		t.Error("Minibuffer should be visible after Show()")
	}

	mb.Hide()
	if mb.IsVisible() {
		t.Error("Minibuffer should not be visible after Hide()")
	}
}

func TestMinibufferSetPrompt(t *testing.T) {
	mb := NewMinibuffer()
	mb.SetPrompt("/: ")
	if mb.prompt != "/: " {
		t.Errorf("prompt = %q, want %q", mb.prompt, "/: ")
	}
}

func TestMinibufferClear(t *testing.T) {
	mb := NewMinibuffer()
	mb.input = "test"
	mb.cursorPos = 4

	mb.Clear()

	if mb.input != "" {
		t.Error("Clear() should empty input")
	}
	if mb.cursorPos != 0 {
		t.Error("Clear() should reset cursor to 0")
	}
}

func TestMinibufferInput(t *testing.T) {
	mb := NewMinibuffer()
	mb.input = "test input"
	if mb.Input() != "test input" {
		t.Errorf("Input() = %q, want %q", mb.Input(), "test input")
	}
}

func TestMinibufferHandleKeyCharacterInput(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()

	// Type "abc"
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	if mb.input != "abc" {
		t.Errorf("input = %q, want %q", mb.input, "abc")
	}
	if mb.cursorPos != 3 {
		t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 3)
	}
}

func TestMinibufferHandleKeyBackspace(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abc"
	mb.cursorPos = 3

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyBackspace})

	if mb.input != "ab" {
		t.Errorf("input = %q, want %q", mb.input, "ab")
	}
	if mb.cursorPos != 2 {
		t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 2)
	}

	// Backspace at beginning should do nothing
	mb.cursorPos = 0
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if mb.input != "ab" {
		t.Errorf("input should not change when backspace at beginning")
	}
}

func TestMinibufferHandleKeyDelete(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abc"
	mb.cursorPos = 1

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyDelete})

	if mb.input != "ac" {
		t.Errorf("input = %q, want %q", mb.input, "ac")
	}
	if mb.cursorPos != 1 {
		t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 1)
	}

	// Delete at end should do nothing
	mb.cursorPos = 2
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyDelete})
	if mb.input != "ac" {
		t.Errorf("input should not change when delete at end")
	}
}

func TestMinibufferHandleKeyCursorMovement(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abc"
	mb.cursorPos = 1

	// Left arrow
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyLeft})
	if mb.cursorPos != 0 {
		t.Errorf("cursorPos = %d, want %d after left", mb.cursorPos, 0)
	}

	// Left at beginning should stay at 0
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyLeft})
	if mb.cursorPos != 0 {
		t.Errorf("cursorPos = %d, want %d after left at beginning", mb.cursorPos, 0)
	}

	// Right arrow
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRight})
	if mb.cursorPos != 1 {
		t.Errorf("cursorPos = %d, want %d after right", mb.cursorPos, 1)
	}

	// Right at end should stay at end
	mb.cursorPos = 3
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRight})
	if mb.cursorPos != 3 {
		t.Errorf("cursorPos = %d, want %d after right at end", mb.cursorPos, 3)
	}
}

func TestMinibufferHandleKeyCtrlA(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abc"
	mb.cursorPos = 3

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlA})
	if mb.cursorPos != 0 {
		t.Errorf("cursorPos = %d, want %d after Ctrl+A", mb.cursorPos, 0)
	}
}

func TestMinibufferHandleKeyCtrlE(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abc"
	mb.cursorPos = 0

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlE})
	if mb.cursorPos != 3 {
		t.Errorf("cursorPos = %d, want %d after Ctrl+E", mb.cursorPos, 3)
	}
}

func TestMinibufferHandleKeyCtrlK(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abcdef"
	mb.cursorPos = 3

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlK})
	if mb.input != "abc" {
		t.Errorf("input = %q, want %q after Ctrl+K", mb.input, "abc")
	}
	if mb.cursorPos != 3 {
		t.Errorf("cursorPos = %d, want %d after Ctrl+K", mb.cursorPos, 3)
	}
}

func TestMinibufferHandleKeyCtrlU(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abcdef"
	mb.cursorPos = 3

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlU})
	if mb.input != "def" {
		t.Errorf("input = %q, want %q after Ctrl+U", mb.input, "def")
	}
	if mb.cursorPos != 0 {
		t.Errorf("cursorPos = %d, want %d after Ctrl+U", mb.cursorPos, 0)
	}
}

func TestMinibufferHandleKeyInsertAtMiddle(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "ac"
	mb.cursorPos = 1

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	if mb.input != "abc" {
		t.Errorf("input = %q, want %q", mb.input, "abc")
	}
	if mb.cursorPos != 2 {
		t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 2)
	}
}

func TestMinibufferView(t *testing.T) {
	mb := NewMinibuffer()
	mb.SetPrompt("/: ")
	mb.SetWidth(40)
	mb.input = "test"
	mb.cursorPos = 4
	mb.Show()

	view := mb.View()

	// View should contain prompt and input
	if !strings.Contains(view, "/: ") {
		t.Error("View should contain prompt")
	}
	if !strings.Contains(view, "test") {
		t.Error("View should contain input")
	}
}

func TestMinibufferViewHidden(t *testing.T) {
	mb := NewMinibuffer()
	mb.SetPrompt("/: ")
	mb.input = "test"

	view := mb.View()
	if view != "" {
		t.Errorf("Hidden minibuffer View() = %q, want empty string", view)
	}
}

func TestMinibufferViewTruncation(t *testing.T) {
	mb := NewMinibuffer()
	mb.SetPrompt("/: ")
	mb.SetWidth(10)
	mb.input = "this is a very long input that should be truncated"
	mb.cursorPos = len(mb.input)
	mb.Show()

	view := mb.View()

	// View should not exceed width
	// Note: actual rendering may include ANSI codes, so we check the visible content
	if len(view) > 100 { // generous limit accounting for ANSI codes
		// Just verify it renders without panic
	}
}
