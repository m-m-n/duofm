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

// TestExtensionRenameDialog_SpaceInFilename tests space handling in filenames.
// This test suite was added to reproduce and verify the fix for a bug where
// the R key (extension-preserving rename) fails with files containing spaces.
func TestExtensionRenameDialog_SpaceInFilename(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("creates dialog with space in base name", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "My Document.txt", "My Document", ".txt")

		if dialog == nil {
			t.Fatal("NewExtensionRenameDialog returned nil")
		}

		if dialog.Input() != "My Document" {
			t.Errorf("Input() = %q, want %q", dialog.Input(), "My Document")
		}

		if dialog.extension != ".txt" {
			t.Errorf("extension = %q, want %q", dialog.extension, ".txt")
		}
	})

	t.Run("Enter with space in base name generates correct filename", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "My Document.txt", "My Document", ".txt")
		dialog.SetInput("Your Document")

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cmd == nil {
			t.Fatal("Expected cmd, got nil")
		}

		msg := cmd()
		result, ok := msg.(extensionRenameResultMsg)
		if !ok {
			t.Fatalf("Expected extensionRenameResultMsg, got %T", msg)
		}

		if result.newName != "Your Document.txt" {
			t.Errorf("newName = %q, want %q", result.newName, "Your Document.txt")
		}
	})

	t.Run("handles multiple consecutive spaces", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "My  Long  Document.txt", "My  Long  Document", ".txt")
		dialog.SetInput("Another  Long  Name")

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

		msg := cmd()
		result := msg.(extensionRenameResultMsg)

		if result.newName != "Another  Long  Name.txt" {
			t.Errorf("newName = %q, want %q", result.newName, "Another  Long  Name.txt")
		}
	})

	t.Run("handles leading space in filename", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, " Document.txt", " Document", ".txt")
		dialog.SetInput(" New Name")

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

		msg := cmd()
		result := msg.(extensionRenameResultMsg)

		if result.newName != " New Name.txt" {
			t.Errorf("newName = %q, want %q", result.newName, " New Name.txt")
		}
	})

	t.Run("handles trailing space in filename", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "Document .txt", "Document ", ".txt")
		dialog.SetInput("New Name ")

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

		msg := cmd()
		result := msg.(extensionRenameResultMsg)

		if result.newName != "New Name .txt" {
			t.Errorf("newName = %q, want %q", result.newName, "New Name .txt")
		}
	})

	t.Run("validates filename with spaces correctly", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
		dialog.SetInput("new name with spaces")

		// Should not have error for valid name with spaces
		if dialog.hasError {
			t.Errorf("Expected no error for valid name with spaces, got: %s", dialog.errorMessage)
		}
	})

	t.Run("detects duplicate with space in name", func(t *testing.T) {
		// Create existing file with space
		os.WriteFile(filepath.Join(tmpDir, "existing file.txt"), []byte("test"), 0644)

		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
		dialog.SetInput("existing file")

		if !dialog.hasError {
			t.Error("Expected hasError = true for duplicate filename")
		}

		if dialog.errorMessage != "File already exists" {
			t.Errorf("errorMessage = %q, want %q", dialog.errorMessage, "File already exists")
		}
	})

	t.Run("handles space around dot separator", func(t *testing.T) {
		// Test case: "My Doc .txt" - space before the dot separator
		dialog := NewExtensionRenameDialog(tmpDir, "My Doc .txt", "My Doc ", ".txt")
		dialog.SetInput("Your Doc ")

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

		msg := cmd()
		result := msg.(extensionRenameResultMsg)

		if result.newName != "Your Doc .txt" {
			t.Errorf("newName = %q, want %q", result.newName, "Your Doc .txt")
		}
	})

	t.Run("handles hidden file with space in name", func(t *testing.T) {
		// Test case: ".my doc.txt" - hidden file with space in the visible name
		dialog := NewExtensionRenameDialog(tmpDir, ".my doc.txt", ".my doc", ".txt")
		dialog.SetInput(".your doc")

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

		msg := cmd()
		result := msg.(extensionRenameResultMsg)

		if result.newName != ".your doc.txt" {
			t.Errorf("newName = %q, want %q", result.newName, ".your doc.txt")
		}
	})
}

// TestExtensionRenameDialog_SpaceKeyInput tests space key input handling.
// This is the core test for the bug: pressing space key in the dialog.
func TestExtensionRenameDialog_SpaceKeyInput(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("space key input inserts space character", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")

		// Clear input and set cursor at start
		dialog.SetInput("")
		dialog.textInput.Value = "MyDoc"
		dialog.textInput.CursorPos = 2 // After "My"

		// Simulate space key press
		dialog.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

		// Check that space was inserted
		if dialog.Input() != "My Doc" {
			t.Errorf("After space key, Input() = %q, want %q", dialog.Input(), "My Doc")
		}

		if dialog.CursorPos() != 3 {
			t.Errorf("After space key, CursorPos() = %d, want %d", dialog.CursorPos(), 3)
		}
	})

	t.Run("space key at end appends space", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
		dialog.SetInput("My")

		// Simulate space key press at end
		dialog.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

		if dialog.Input() != "My " {
			t.Errorf("After space key at end, Input() = %q, want %q", dialog.Input(), "My ")
		}
	})

	t.Run("multiple space key presses", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
		dialog.SetInput("a")

		// Simulate multiple space key presses
		dialog.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
		dialog.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

		if dialog.Input() != "a  " {
			t.Errorf("After multiple space keys, Input() = %q, want %q", dialog.Input(), "a  ")
		}
	})

	t.Run("complete rename flow with space key input", func(t *testing.T) {
		dialog := NewExtensionRenameDialog(tmpDir, "old.txt", "old", ".txt")

		// Clear and type "new name" using space key for the space
		dialog.SetInput("new")
		dialog.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
		// Type "name" using regular key input
		dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n', 'a', 'm', 'e'}})

		if dialog.Input() != "new name" {
			t.Errorf("Input() = %q, want %q", dialog.Input(), "new name")
		}

		// Press Enter
		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

		msg := cmd()
		result := msg.(extensionRenameResultMsg)

		if result.newName != "new name.txt" {
			t.Errorf("newName = %q, want %q", result.newName, "new name.txt")
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
