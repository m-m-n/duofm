package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewErrorDialog(t *testing.T) {
	msg := "Test error message"
	dialog := NewErrorDialog(msg)

	if dialog == nil {
		t.Fatal("NewErrorDialog() returned nil")
	}

	if !dialog.IsActive() {
		t.Error("NewErrorDialog() should be active by default")
	}

	if dialog.message != msg {
		t.Errorf("NewErrorDialog() message = %q, want %q", dialog.message, msg)
	}
}

func TestErrorDialog_IsActive(t *testing.T) {
	dialog := NewErrorDialog("Error")

	if !dialog.IsActive() {
		t.Error("IsActive() should return true for active dialog")
	}

	dialog.active = false
	if dialog.IsActive() {
		t.Error("IsActive() should return false for inactive dialog")
	}
}

func TestErrorDialog_DisplayType(t *testing.T) {
	dialog := NewErrorDialog("Error")

	if dialog.DisplayType() != DialogDisplayScreen {
		t.Errorf("DisplayType() = %v, want DialogDisplayScreen", dialog.DisplayType())
	}
}

func TestErrorDialog_Update(t *testing.T) {
	t.Run("Esc closes dialog", func(t *testing.T) {
		dialog := NewErrorDialog("Error")
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})
		d := updated.(*ErrorDialog)

		if d.IsActive() {
			t.Error("Dialog should be inactive after Esc")
		}
		if cmd == nil {
			t.Error("Esc should return a command")
		}
	})

	t.Run("Enter closes dialog", func(t *testing.T) {
		dialog := NewErrorDialog("Error")
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})
		d := updated.(*ErrorDialog)

		if d.IsActive() {
			t.Error("Dialog should be inactive after Enter")
		}
		if cmd == nil {
			t.Error("Enter should return a command")
		}
	})

	t.Run("Ctrl+C closes dialog", func(t *testing.T) {
		dialog := NewErrorDialog("Error")
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		d := updated.(*ErrorDialog)

		if d.IsActive() {
			t.Error("Dialog should be inactive after Ctrl+C")
		}
		if cmd == nil {
			t.Error("Ctrl+C should return a command")
		}
	})

	t.Run("inactive dialog ignores input", func(t *testing.T) {
		dialog := NewErrorDialog("Error")
		dialog.active = false
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

		if updated == nil {
			t.Error("Updated dialog should not be nil")
		}
		if cmd != nil {
			t.Error("Command should be nil for inactive dialog")
		}
	})

	t.Run("other keys are ignored", func(t *testing.T) {
		dialog := NewErrorDialog("Error")
		updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		d := updated.(*ErrorDialog)

		if !d.IsActive() {
			t.Error("Dialog should remain active after unrecognized key")
		}
		if cmd != nil {
			t.Error("Command should be nil for unrecognized key")
		}
	})
}

func TestErrorDialog_View(t *testing.T) {
	t.Run("active dialog shows content", func(t *testing.T) {
		dialog := NewErrorDialog("Test error message")
		view := dialog.View()

		if view == "" {
			t.Error("View should not be empty for active dialog")
		}

		if !strings.Contains(view, "Error") {
			t.Error("View should contain 'Error' title")
		}

		if !strings.Contains(view, "Test error message") {
			t.Error("View should contain error message")
		}

		if !strings.Contains(view, "Esc") {
			t.Error("View should contain escape hint")
		}
	})

	t.Run("inactive dialog shows empty", func(t *testing.T) {
		dialog := NewErrorDialog("Error")
		dialog.active = false
		view := dialog.View()

		if view != "" {
			t.Error("View should be empty for inactive dialog")
		}
	})
}

func TestErrorDialog_UpdateCommand(t *testing.T) {
	dialog := NewErrorDialog("Error")
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("Command should not be nil")
	}

	// Execute the command and check the message type
	msg := cmd()
	_, ok := msg.(dialogResultMsg)
	if !ok {
		t.Errorf("Command should return dialogResultMsg, got %T", msg)
	}
}
