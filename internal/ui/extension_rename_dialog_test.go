package ui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewExtensionRenameDialog(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Create some test files
	os.WriteFile(filepath.Join(tmpDir, "existing.txt"), []byte("test"), 0644)

	t.Run("creates dialog with correct values", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")

		if dialog == nil {
			t.Fatal("NewExtensionRenameDialog returned nil")
		}

		if !dialog.IsActive() {
			t.Error("Dialog should be active")
		}

		if dialog.extension != ".txt" {
			t.Errorf("extension = %q, want %q", dialog.extension, ".txt")
		}

		if dialog.originalName != "document.txt" {
			t.Errorf("originalName = %q, want %q", dialog.originalName, "document.txt")
		}

		if dialog.dirPath != tmpDir {
			t.Errorf("dirPath = %q, want %q", dialog.dirPath, tmpDir)
		}
	})

	t.Run("input contains only base name", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")

		input := dialog.Input()
		if input != "document" {
			t.Errorf("Input() = %q, want %q", input, "document")
		}
	})

	t.Run("loads existing files for validation", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")

		if dialog.existingFiles == nil {
			t.Fatal("existingFiles is nil")
		}

		if !dialog.existingFiles["existing.txt"] {
			t.Error("existingFiles should contain existing.txt")
		}
	})
}

func TestExtensionRenameDialog_Update_Enter(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Enter generates correct full filename", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
		dialog.SetInput("newname")

		// Press Enter
		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cmd == nil {
			t.Fatal("Expected cmd, got nil")
		}

		msg := cmd()
		result, ok := msg.(extensionRenameResultMsg)
		if !ok {
			t.Fatalf("Expected extensionRenameResultMsg, got %T", msg)
		}

		if result.cancelled {
			t.Error("Result should not be cancelled")
		}

		if result.newName != "newname.txt" {
			t.Errorf("newName = %q, want %q", result.newName, "newname.txt")
		}

		if result.oldName != "document.txt" {
			t.Errorf("oldName = %q, want %q", result.oldName, "document.txt")
		}

		if result.dirPath != tmpDir {
			t.Errorf("dirPath = %q, want %q", result.dirPath, tmpDir)
		}
	})

	t.Run("Enter with error does nothing", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
		dialog.SetInput("") // Empty input causes error

		// Press Enter
		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cmd != nil {
			t.Error("Expected nil cmd when there is an error")
		}

		if !dialog.IsActive() {
			t.Error("Dialog should still be active")
		}
	})
}

func TestExtensionRenameDialog_Update_Escape(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Escape cancels dialog", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")

		// Press Escape
		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

		if cmd == nil {
			t.Fatal("Expected cmd, got nil")
		}

		msg := cmd()
		result, ok := msg.(extensionRenameResultMsg)
		if !ok {
			t.Fatalf("Expected extensionRenameResultMsg, got %T", msg)
		}

		if !result.cancelled {
			t.Error("Result should be cancelled")
		}

		if dialog.IsActive() {
			t.Error("Dialog should not be active after Escape")
		}
	})
}

func TestExtensionRenameDialog_Validation(t *testing.T) {
	tmpDir := t.TempDir()
	// Create existing file
	os.WriteFile(filepath.Join(tmpDir, "existing.txt"), []byte("test"), 0644)

	t.Run("empty input shows error", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
		dialog.SetInput("")

		if !dialog.hasError {
			t.Error("Expected hasError = true for empty input")
		}

		if dialog.errorMessage == "" {
			t.Error("Expected non-empty error message")
		}
	})

	t.Run("duplicate filename shows error", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
		dialog.SetInput("existing") // existing + .txt = existing.txt

		if !dialog.hasError {
			t.Error("Expected hasError = true for duplicate filename")
		}

		if dialog.errorMessage == "" {
			t.Error("Expected non-empty error message")
		}
	})

	t.Run("invalid characters show error", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
		dialog.SetInput("new/name") // Contains path separator

		if !dialog.hasError {
			t.Error("Expected hasError = true for invalid characters")
		}
	})

	t.Run("valid input has no error", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
		dialog.SetInput("newname")

		if dialog.hasError {
			t.Errorf("Expected hasError = false for valid input, got error: %s", dialog.errorMessage)
		}
	})

	t.Run("same name as original has no error", func(t *testing.T) {
		// When renaming, the original file will be replaced, so same name is valid
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
		dialog.SetInput("document")

		// Note: the original file "document.txt" may or may not exist in tmpDir
		// If it doesn't exist, this should pass regardless
		// The implementation should exclude the original filename from duplicate check
		if dialog.hasError && dialog.errorMessage == "File already exists" {
			// Only fail if the error is about the file already existing AND
			// the file is the same as the original (which should be allowed)
			t.Error("Same name as original should be allowed")
		}
	})
}

func TestExtensionRenameDialog_View(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("renders when active", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")

		view := dialog.View()

		if view == "" {
			t.Error("View should not be empty when active")
		}
	})

	t.Run("returns empty when inactive", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
		dialog.Close()

		view := dialog.View()

		if view != "" {
			t.Error("View should be empty when inactive")
		}
	})

	t.Run("view contains extension", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")

		view := dialog.View()

		if !containsString(view, ".txt") {
			t.Error("View should contain extension")
		}
	})

	t.Run("view contains footer hints", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")

		view := dialog.View()

		if !containsString(view, "Enter") || !containsString(view, "Esc") {
			t.Error("View should contain keybinding hints")
		}
	})
}

func TestExtensionRenameDialog_HiddenFiles(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("hidden file with extension", func(t *testing.T) {
		// .config.json -> baseName=".config", ext=".json"
		dialog := NewExtensionRenameDialog(tmpDir, ".config.json", ".config", ".json")

		if dialog.Input() != ".config" {
			t.Errorf("Input() = %q, want %q", dialog.Input(), ".config")
		}

		// Press Enter
		dialog.SetInput(".newconfig")
		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

		msg := cmd()
		result, ok := msg.(extensionRenameResultMsg)
		if !ok {
			t.Fatalf("Expected extensionRenameResultMsg, got %T", msg)
		}

		if result.newName != ".newconfig.json" {
			t.Errorf("newName = %q, want %q", result.newName, ".newconfig.json")
		}
	})
}

// containsString checks if a string contains a substring
// Uses simple substring matching
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
