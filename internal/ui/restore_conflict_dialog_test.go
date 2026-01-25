package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewRestoreConflictDialog(t *testing.T) {
	d := NewRestoreConflictDialog("file.txt", "/home/user/Documents/file.txt")

	if d == nil {
		t.Fatal("NewRestoreConflictDialog() returned nil")
	}

	if d.trashName != "file.txt" {
		t.Errorf("trashName = %q, want %q", d.trashName, "file.txt")
	}

	if d.originalPath != "/home/user/Documents/file.txt" {
		t.Errorf("originalPath = %q, want %q", d.originalPath, "/home/user/Documents/file.txt")
	}

	if !d.IsActive() {
		t.Error("dialog should be active after creation")
	}
}

func TestRestoreConflictDialogUpdate(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		expectedChoice RestoreConflictChoice
	}{
		{
			name:           "overwrite with o",
			key:            "o",
			expectedChoice: RestoreChoiceOverwrite,
		},
		{
			name:           "overwrite with O",
			key:            "O",
			expectedChoice: RestoreChoiceOverwrite,
		},
		{
			name:           "rename with r",
			key:            "r",
			expectedChoice: RestoreChoiceRename,
		},
		{
			name:           "rename with R",
			key:            "R",
			expectedChoice: RestoreChoiceRename,
		},
		{
			name:           "skip with s",
			key:            "s",
			expectedChoice: RestoreChoiceSkip,
		},
		{
			name:           "skip with S",
			key:            "S",
			expectedChoice: RestoreChoiceSkip,
		},
		{
			name:           "cancel with esc",
			key:            "esc",
			expectedChoice: RestoreChoiceCancelled,
		},
		{
			name:           "cancel with ctrl+c",
			key:            "ctrl+c",
			expectedChoice: RestoreChoiceCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewRestoreConflictDialog("file.txt", "/home/user/Documents/file.txt")

			keyMsg := tea.KeyMsg{Type: tea.KeyRunes}
			switch tt.key {
			case "esc":
				keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
			case "ctrl+c":
				keyMsg = tea.KeyMsg{Type: tea.KeyCtrlC}
			default:
				keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			_, cmd := d.Update(keyMsg)

			if cmd == nil {
				t.Fatal("Update() returned nil cmd")
			}

			// Execute the command
			msg := cmd()
			result, ok := msg.(restoreConflictResultMsg)
			if !ok {
				t.Fatalf("expected restoreConflictResultMsg, got %T", msg)
			}

			if result.choice != tt.expectedChoice {
				t.Errorf("choice = %v, want %v", result.choice, tt.expectedChoice)
			}

			if result.trashName != "file.txt" {
				t.Errorf("trashName = %q, want %q", result.trashName, "file.txt")
			}

			if result.originalPath != "/home/user/Documents/file.txt" {
				t.Errorf("originalPath = %q, want %q", result.originalPath, "/home/user/Documents/file.txt")
			}

			if d.IsActive() {
				t.Error("dialog should be closed after choice")
			}
		})
	}
}

func TestRestoreConflictDialogUpdateIgnoresUnrelatedKeys(t *testing.T) {
	d := NewRestoreConflictDialog("file.txt", "/home/user/Documents/file.txt")

	// Send an unrelated key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	_, cmd := d.Update(keyMsg)

	if cmd != nil {
		t.Error("Update() should return nil cmd for unrelated keys")
	}

	if !d.IsActive() {
		t.Error("dialog should remain active after unrelated key")
	}
}

func TestRestoreConflictDialogUpdateInactive(t *testing.T) {
	d := NewRestoreConflictDialog("file.txt", "/home/user/Documents/file.txt")
	d.Close()

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")}
	_, cmd := d.Update(keyMsg)

	if cmd != nil {
		t.Error("Update() should return nil cmd when dialog is inactive")
	}
}

func TestRestoreConflictDialogView(t *testing.T) {
	d := NewRestoreConflictDialog("file.txt", "/home/user/Documents/file.txt")

	view := d.View()

	// Check that the view contains expected elements
	if !strings.Contains(view, "File already exists") {
		t.Error("View should contain 'File already exists'")
	}

	if !strings.Contains(view, "/home/user/Documents/file.txt") {
		t.Error("View should contain the original path")
	}

	if !strings.Contains(view, "[O]verwrite") {
		t.Error("View should contain '[O]verwrite' option")
	}

	if !strings.Contains(view, "[R]ename") {
		t.Error("View should contain '[R]ename' option")
	}

	if !strings.Contains(view, "[S]kip") {
		t.Error("View should contain '[S]kip' option")
	}
}

func TestRestoreConflictDialogViewInactive(t *testing.T) {
	d := NewRestoreConflictDialog("file.txt", "/home/user/Documents/file.txt")
	d.Close()

	view := d.View()

	if view != "" {
		t.Error("View() should return empty string when dialog is inactive")
	}
}
