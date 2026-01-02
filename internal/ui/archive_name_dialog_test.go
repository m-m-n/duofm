package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewArchiveNameDialog(t *testing.T) {
	dialog := NewArchiveNameDialog("test.tar.gz")

	if dialog == nil {
		t.Fatal("NewArchiveNameDialog() returned nil")
	}

	if !dialog.IsActive() {
		t.Error("NewArchiveNameDialog() should be active by default")
	}

	if dialog.input != "test.tar.gz" {
		t.Errorf("NewArchiveNameDialog() input = %q, want %q", dialog.input, "test.tar.gz")
	}
}

func TestArchiveNameDialog_EmptyInput(t *testing.T) {
	dialog := NewArchiveNameDialog("test.tar.gz")
	dialog.input = ""

	// Try to confirm with empty input
	updated, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})
	d := updated.(*ArchiveNameDialog)

	if d.errorMsg == "" {
		t.Error("Error message should be set for empty input")
	}

	if !d.IsActive() {
		t.Error("Dialog should remain active on validation error")
	}
}

func TestArchiveNameDialog_InvalidCharacters(t *testing.T) {
	dialog := NewArchiveNameDialog("test.tar.gz")

	tests := []struct {
		name  string
		input string
	}{
		{"null character", "test\x00.tar"},
		{"control character", "test\x01.tar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.input = tt.input
			updated, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})
			d := updated.(*ArchiveNameDialog)

			if d.errorMsg == "" {
				t.Error("Error message should be set for invalid characters")
			}
		})
	}
}

func TestArchiveNameDialog_ValidInput(t *testing.T) {
	dialog := NewArchiveNameDialog("default.tar.gz")
	dialog.input = "myarchive.tar.xz"

	// Confirm with valid input
	updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})
	d := updated.(*ArchiveNameDialog)

	if d.IsActive() {
		t.Error("Dialog should be inactive after confirmation")
	}

	// Check that a command is returned
	if cmd == nil {
		t.Error("Command should be returned after confirmation")
	}

	// Execute the command and check the message
	if cmd != nil {
		msg := cmd()
		result, ok := msg.(archiveNameResultMsg)
		if !ok {
			t.Error("Command should return archiveNameResultMsg")
		}
		if result.name != "myarchive.tar.xz" {
			t.Errorf("archiveNameResultMsg.name = %q, want %q", result.name, "myarchive.tar.xz")
		}
		if result.cancelled {
			t.Error("archiveNameResultMsg.cancelled should be false")
		}
	}
}

func TestArchiveNameDialog_Cancel(t *testing.T) {
	dialog := NewArchiveNameDialog("test.tar.gz")

	// Press Escape
	updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})
	d := updated.(*ArchiveNameDialog)

	if d.IsActive() {
		t.Error("Dialog should be inactive after cancellation")
	}

	// Check that a command is returned
	if cmd == nil {
		t.Error("Command should be returned after cancellation")
	}

	// Execute the command and check the message
	if cmd != nil {
		msg := cmd()
		result, ok := msg.(archiveNameResultMsg)
		if !ok {
			t.Error("Command should return archiveNameResultMsg")
		}
		if !result.cancelled {
			t.Error("archiveNameResultMsg.cancelled should be true")
		}
	}
}

func TestArchiveNameDialog_SetActive(t *testing.T) {
	dialog := NewArchiveNameDialog("test.tar.gz")

	dialog.SetActive(false)
	if dialog.IsActive() {
		t.Error("SetActive(false) should deactivate dialog")
	}

	dialog.SetActive(true)
	if !dialog.IsActive() {
		t.Error("SetActive(true) should activate dialog")
	}
}

func TestArchiveNameDialog_DisplayType(t *testing.T) {
	dialog := NewArchiveNameDialog("test.tar.gz")
	if dialog.DisplayType() != DialogDisplayScreen {
		t.Errorf("DisplayType() = %v, want DialogDisplayScreen", dialog.DisplayType())
	}
}

func TestArchiveNameDialog_View(t *testing.T) {
	t.Run("active dialog shows content", func(t *testing.T) {
		dialog := NewArchiveNameDialog("test.tar.gz")
		view := dialog.View()

		if view == "" {
			t.Error("View should not be empty for active dialog")
		}

		if !containsDialogStr(view, "Archive Name") {
			t.Error("View should contain title")
		}

		if !containsDialogStr(view, "[Enter] Confirm") {
			t.Error("View should contain confirm hint")
		}

		if !containsDialogStr(view, "[Esc] Cancel") {
			t.Error("View should contain cancel hint")
		}
	})

	t.Run("dialog with error shows error message", func(t *testing.T) {
		dialog := NewArchiveNameDialog("test.tar.gz")
		dialog.input = ""
		// Trigger validation error
		dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

		view := dialog.View()

		if !containsDialogStr(view, "Error") {
			// After setting error, check if it's shown
			dialog.errorMsg = "Name cannot be empty"
			view = dialog.View()
			if !containsDialogStr(view, "Error") && !containsDialogStr(view, "empty") {
				t.Error("View should show error message when present")
			}
		}
	})

	t.Run("inactive dialog shows empty", func(t *testing.T) {
		dialog := NewArchiveNameDialog("test.tar.gz")
		dialog.SetActive(false)
		view := dialog.View()

		if view != "" {
			t.Error("View should be empty for inactive dialog")
		}
	})
}

func TestArchiveNameDialog_Backspace(t *testing.T) {
	dialog := NewArchiveNameDialog("test.tar.gz")
	initialLen := len(dialog.input)

	// Press backspace
	updated, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	d := updated.(*ArchiveNameDialog)

	if len(d.input) != initialLen-1 {
		t.Errorf("Backspace should remove one character, got len=%d, want %d", len(d.input), initialLen-1)
	}
}

func TestArchiveNameDialog_TypeCharacter(t *testing.T) {
	dialog := NewArchiveNameDialog("test")

	// Type a character
	updated, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	d := updated.(*ArchiveNameDialog)

	if d.input != "testa" {
		t.Errorf("Typing 'a' should append, got %q, want %q", d.input, "testa")
	}
}

func TestArchiveNameDialog_Update_Inactive(t *testing.T) {
	dialog := NewArchiveNameDialog("test.tar.gz")
	dialog.SetActive(false)

	updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	if updated == nil {
		t.Error("Updated dialog should not be nil")
	}

	if cmd != nil {
		t.Error("Command should be nil for inactive dialog")
	}
}

// containsDialogStr checks if s contains substr
func containsDialogStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
