package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewConfirmDialog(t *testing.T) {
	title := "Confirm Action"
	message := "Are you sure?"
	dialog := NewConfirmDialog(title, message)

	if dialog == nil {
		t.Fatal("NewConfirmDialog() returned nil")
	}

	if !dialog.IsActive() {
		t.Error("NewConfirmDialog() should be active by default")
	}

	if dialog.title != title {
		t.Errorf("NewConfirmDialog() title = %q, want %q", dialog.title, title)
	}

	if dialog.message != message {
		t.Errorf("NewConfirmDialog() message = %q, want %q", dialog.message, message)
	}
}

func TestConfirmDialog_IsActive(t *testing.T) {
	dialog := NewConfirmDialog("Title", "Message")

	if !dialog.IsActive() {
		t.Error("IsActive() should return true for active dialog")
	}

	dialog.active = false
	if dialog.IsActive() {
		t.Error("IsActive() should return false for inactive dialog")
	}
}

func TestConfirmDialog_DisplayType(t *testing.T) {
	dialog := NewConfirmDialog("Title", "Message")

	if dialog.DisplayType() != DialogDisplayPane {
		t.Errorf("DisplayType() = %v, want DialogDisplayPane", dialog.DisplayType())
	}
}

func TestConfirmDialog_Update(t *testing.T) {
	t.Run("y confirms dialog", func(t *testing.T) {
		dialog := NewConfirmDialog("Title", "Message")
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
		d := updated.(*ConfirmDialog)

		if d.IsActive() {
			t.Error("Dialog should be inactive after y")
		}
		if cmd == nil {
			t.Fatal("y should return a command")
		}

		// Check the command returns a confirmed result
		msg := cmd()
		result, ok := msg.(dialogResultMsg)
		if !ok {
			t.Errorf("Command should return dialogResultMsg, got %T", msg)
		}
		if !result.result.Confirmed {
			t.Error("Result should be confirmed")
		}
	})

	t.Run("n cancels dialog", func(t *testing.T) {
		dialog := NewConfirmDialog("Title", "Message")
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
		d := updated.(*ConfirmDialog)

		if d.IsActive() {
			t.Error("Dialog should be inactive after n")
		}
		if cmd == nil {
			t.Fatal("n should return a command")
		}

		// Check the command returns a cancelled result
		msg := cmd()
		result, ok := msg.(dialogResultMsg)
		if !ok {
			t.Errorf("Command should return dialogResultMsg, got %T", msg)
		}
		if !result.result.Cancelled {
			t.Error("Result should be cancelled")
		}
	})

	t.Run("Esc cancels dialog", func(t *testing.T) {
		dialog := NewConfirmDialog("Title", "Message")
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})
		d := updated.(*ConfirmDialog)

		if d.IsActive() {
			t.Error("Dialog should be inactive after Esc")
		}
		if cmd == nil {
			t.Error("Esc should return a command")
		}
	})

	t.Run("Ctrl+C cancels dialog", func(t *testing.T) {
		dialog := NewConfirmDialog("Title", "Message")
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		d := updated.(*ConfirmDialog)

		if d.IsActive() {
			t.Error("Dialog should be inactive after Ctrl+C")
		}
		if cmd == nil {
			t.Error("Ctrl+C should return a command")
		}
	})

	t.Run("inactive dialog ignores input", func(t *testing.T) {
		dialog := NewConfirmDialog("Title", "Message")
		dialog.active = false
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

		if updated == nil {
			t.Error("Updated dialog should not be nil")
		}
		if cmd != nil {
			t.Error("Command should be nil for inactive dialog")
		}
	})

	t.Run("other keys are ignored", func(t *testing.T) {
		dialog := NewConfirmDialog("Title", "Message")
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
		d := updated.(*ConfirmDialog)

		if !d.IsActive() {
			t.Error("Dialog should remain active after unrecognized key")
		}
		if cmd != nil {
			t.Error("Command should be nil for unrecognized key")
		}
	})
}

func TestConfirmDialog_View(t *testing.T) {
	t.Run("active dialog shows content", func(t *testing.T) {
		dialog := NewConfirmDialog("Confirm Delete", "Delete file.txt?")
		view := dialog.View()

		if view == "" {
			t.Error("View should not be empty for active dialog")
		}

		if !strings.Contains(view, "Confirm Delete") {
			t.Error("View should contain title")
		}

		if !strings.Contains(view, "Delete file.txt?") {
			t.Error("View should contain message")
		}

		if !strings.Contains(view, "[y] Yes") {
			t.Error("View should contain yes button")
		}

		if !strings.Contains(view, "[n] No") {
			t.Error("View should contain no button")
		}
	})

	t.Run("inactive dialog shows empty", func(t *testing.T) {
		dialog := NewConfirmDialog("Title", "Message")
		dialog.active = false
		view := dialog.View()

		if view != "" {
			t.Error("View should be empty for inactive dialog")
		}
	})
}
