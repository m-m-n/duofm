package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPathJumpDialog_NewDialog(t *testing.T) {
	dialog := NewPathJumpDialog()

	if !dialog.IsActive() {
		t.Error("Dialog should be active when created")
	}
	if dialog.Input() != "" {
		t.Error("Input should be empty initially")
	}
	if dialog.DisplayType() != DialogDisplayPane {
		t.Error("DisplayType should be DialogDisplayPane")
	}
	if dialog.suggestion != "" {
		t.Error("Suggestion should be empty initially")
	}
	if dialog.errorMsg != "" {
		t.Error("Error message should be empty initially")
	}
}

func TestPathJumpDialog_TabCompletion(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	dialog := NewPathJumpDialog()
	dialog.SetInput(filepath.Join(tmpDir, "test"))

	// Recalculate suggestion
	dialog.updateSuggestion()

	// Verify suggestion exists
	if dialog.suggestion != "dir" {
		t.Errorf("Expected suggestion %q, got %q", "dir", dialog.suggestion)
	}

	// Press Tab to confirm suggestion
	dialog.Update(tea.KeyMsg{Type: tea.KeyTab})

	expected := filepath.Join(tmpDir, "testdir")
	if dialog.Input() != expected {
		t.Errorf("After Tab, expected input %q, got %q", expected, dialog.Input())
	}
}

func TestPathJumpDialog_TabNoSuggestion(t *testing.T) {
	dialog := NewPathJumpDialog()
	dialog.SetInput("/nonexistent/path")

	// Ensure no suggestion
	dialog.updateSuggestion()
	if dialog.suggestion != "" {
		t.Fatal("Expected no suggestion for nonexistent path")
	}

	initialInput := dialog.Input()

	// Press Tab - should do nothing
	dialog.Update(tea.KeyMsg{Type: tea.KeyTab})

	if dialog.Input() != initialInput {
		t.Errorf("Tab without suggestion should not change input, got %q", dialog.Input())
	}
}

func TestPathJumpDialog_EnterValidPath(t *testing.T) {
	tmpDir := t.TempDir()

	dialog := NewPathJumpDialog()
	dialog.SetInput(tmpDir)

	// Press Enter
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Dialog should be closed
	if dialog.IsActive() {
		t.Error("Dialog should be inactive after successful Enter")
	}

	// Should return a command
	if cmd == nil {
		t.Fatal("Enter with valid path should return a command")
	}

	// Execute command and check message type
	msg := cmd()
	resultMsg, ok := msg.(pathJumpResultMsg)
	if !ok {
		t.Fatalf("Expected pathJumpResultMsg, got %T", msg)
	}

	if resultMsg.path != tmpDir {
		t.Errorf("Expected path %q, got %q", tmpDir, resultMsg.path)
	}
}

func TestPathJumpDialog_EnterInvalidPath(t *testing.T) {
	dialog := NewPathJumpDialog()
	dialog.SetInput("/nonexistent/path/12345")

	// Press Enter
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Dialog should stay active
	if !dialog.IsActive() {
		t.Error("Dialog should stay active after invalid path")
	}

	// Error message should be set
	if dialog.errorMsg == "" {
		t.Error("Error message should be set for invalid path")
	}
}

func TestPathJumpDialog_EnterFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	dialog := NewPathJumpDialog()
	dialog.SetInput(filePath)

	// Press Enter
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Dialog should stay active
	if !dialog.IsActive() {
		t.Error("Dialog should stay active when path is a file")
	}

	// Error message should indicate not a directory
	if dialog.errorMsg == "" {
		t.Error("Error message should be set when path is a file")
	}
	if !strings.Contains(dialog.errorMsg, "directory") {
		t.Errorf("Error message should mention 'directory', got %q", dialog.errorMsg)
	}
}

func TestPathJumpDialog_EnterEmptyPath(t *testing.T) {
	dialog := NewPathJumpDialog()
	// Leave input empty

	// Press Enter
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Dialog should stay active
	if !dialog.IsActive() {
		t.Error("Dialog should stay active for empty path")
	}

	// Error message should be set
	if dialog.errorMsg == "" {
		t.Error("Error message should be set for empty path")
	}
}

func TestPathJumpDialog_EscCancel(t *testing.T) {
	dialog := NewPathJumpDialog()
	dialog.SetInput("/some/path")

	// Press Esc
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Dialog should be closed
	if dialog.IsActive() {
		t.Error("Dialog should be inactive after Esc")
	}

	// Should return a command
	if cmd == nil {
		t.Fatal("Esc should return a command")
	}

	// Execute command and check message type
	msg := cmd()
	_, ok := msg.(pathJumpCancelMsg)
	if !ok {
		t.Fatalf("Expected pathJumpCancelMsg, got %T", msg)
	}
}

func TestPathJumpDialog_ErrorClearsOnInput(t *testing.T) {
	dialog := NewPathJumpDialog()
	dialog.errorMsg = "Some error"

	// Type a character
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})

	// Error should be cleared
	if dialog.errorMsg != "" {
		t.Errorf("Error message should be cleared on input, got %q", dialog.errorMsg)
	}
}

func TestPathJumpDialog_InactiveIgnoresInput(t *testing.T) {
	dialog := NewPathJumpDialog()
	dialog.SetActive(false)
	dialog.SetInput("/initial")

	// Try to type
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	// Input should not change
	if dialog.Input() != "/initial" {
		t.Error("Inactive dialog should not process input")
	}
}

func TestPathJumpDialog_View(t *testing.T) {
	dialog := NewPathJumpDialog()
	dialog.SetInput("/home/user")

	view := dialog.View()

	// Check that view contains title
	if !strings.Contains(view, "Jump to Directory") {
		t.Error("View should contain title")
	}

	// Check that view contains the input
	if !strings.Contains(view, "/home/user") {
		t.Error("View should contain the input path")
	}

	// Check that view contains footer with key hints
	if !strings.Contains(view, "Tab") || !strings.Contains(view, "Enter") || !strings.Contains(view, "Esc") {
		t.Error("View should contain key hints (Tab, Enter, Esc)")
	}
}

func TestPathJumpDialog_ViewWithSuggestion(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	dialog := NewPathJumpDialog()
	dialog.SetInput(filepath.Join(tmpDir, "test"))
	dialog.updateSuggestion()

	view := dialog.View()

	// Suggestion should be calculated
	if dialog.suggestion != "dir" {
		t.Errorf("Suggestion should be 'dir', got %q", dialog.suggestion)
	}

	// View should not be empty
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestPathJumpDialog_ViewWithError(t *testing.T) {
	dialog := NewPathJumpDialog()
	dialog.errorMsg = "Directory does not exist"

	view := dialog.View()

	// Check that view contains the error message
	if !strings.Contains(view, "Directory does not exist") {
		t.Error("View should contain the error message")
	}
}

func TestPathJumpDialog_SuggestionUpdatesOnInput(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	dialog := NewPathJumpDialog()

	// Type the path character by character
	path := tmpDir + "/test"
	for _, r := range path {
		dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Suggestion should be calculated
	if dialog.suggestion != "dir" {
		t.Errorf("Expected suggestion %q, got %q", "dir", dialog.suggestion)
	}
}

func TestPathJumpDialog_CharacterInput(t *testing.T) {
	dialog := NewPathJumpDialog()

	// Type "/home"
	keys := []string{"/", "h", "o", "m", "e"}
	for _, key := range keys {
		dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	}

	if dialog.Input() != "/home" {
		t.Errorf("Expected input %q, got %q", "/home", dialog.Input())
	}
}

func TestPathJumpDialog_CursorMovement(t *testing.T) {
	dialog := NewPathJumpDialog()
	dialog.SetInput("/home/user")

	// Move left
	dialog.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if dialog.CursorPos() != 9 {
		t.Errorf("Expected cursor at 9 after left, got %d", dialog.CursorPos())
	}

	// Move to beginning (Ctrl+A)
	dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	if dialog.CursorPos() != 0 {
		t.Errorf("Expected cursor at 0 after Ctrl+A, got %d", dialog.CursorPos())
	}

	// Move to end (Ctrl+E)
	dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	if dialog.CursorPos() != 10 {
		t.Errorf("Expected cursor at 10 after Ctrl+E, got %d", dialog.CursorPos())
	}
}

func TestPathJumpDialog_Backspace(t *testing.T) {
	dialog := NewPathJumpDialog()
	dialog.SetInput("/home/user")

	// Delete last character
	dialog.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if dialog.Input() != "/home/use" {
		t.Errorf("Expected %q after backspace, got %q", "/home/use", dialog.Input())
	}
}

func TestPathJumpDialog_Delete(t *testing.T) {
	dialog := NewPathJumpDialog()
	dialog.textInput.Value = "/home/user"
	dialog.textInput.CursorPos = 0

	// Delete first character
	dialog.Update(tea.KeyMsg{Type: tea.KeyDelete})
	if dialog.Input() != "home/user" {
		t.Errorf("Expected %q after delete, got %q", "home/user", dialog.Input())
	}
}

func TestPathJumpDialog_CtrlU_CtrlK(t *testing.T) {
	t.Run("Ctrl+U deletes to beginning", func(t *testing.T) {
		dialog := NewPathJumpDialog()
		dialog.textInput.Value = "/home/user"
		dialog.textInput.CursorPos = 5

		dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
		if dialog.Input() != "/user" {
			t.Errorf("Expected %q after Ctrl+U, got %q", "/user", dialog.Input())
		}
	})

	t.Run("Ctrl+K deletes to end", func(t *testing.T) {
		dialog := NewPathJumpDialog()
		dialog.textInput.Value = "/home/user"
		dialog.textInput.CursorPos = 5

		dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
		if dialog.Input() != "/home" {
			t.Errorf("Expected %q after Ctrl+K, got %q", "/home", dialog.Input())
		}
	})
}

func TestPathJumpDialog_InactiveView(t *testing.T) {
	dialog := NewPathJumpDialog()
	dialog.SetActive(false)

	view := dialog.View()

	if view != "" {
		t.Error("Inactive dialog should return empty view")
	}
}

func TestPathJumpDialog_EnterPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("Test cannot run as root")
	}

	tmpDir := t.TempDir()
	restrictedDir := filepath.Join(tmpDir, "restricted")
	if err := os.Mkdir(restrictedDir, 0755); err != nil {
		t.Fatalf("Failed to create restricted directory: %v", err)
	}
	// Create a subdirectory, then remove all permissions from parent
	subDir := filepath.Join(restrictedDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	// Remove permissions from parent so we cannot stat the subdir
	if err := os.Chmod(restrictedDir, 0000); err != nil {
		t.Fatalf("Failed to change permissions: %v", err)
	}
	defer os.Chmod(restrictedDir, 0755) // Cleanup

	dialog := NewPathJumpDialog()
	dialog.SetInput(subDir)

	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !dialog.IsActive() {
		t.Error("Dialog should stay active for permission denied")
	}
	// Either "Permission denied" or "Error accessing" is acceptable
	if !strings.Contains(dialog.errorMsg, "Permission denied") && !strings.Contains(dialog.errorMsg, "Error accessing") && !strings.Contains(dialog.errorMsg, "does not exist") {
		t.Errorf("Expected permission/access error, got %q", dialog.errorMsg)
	}
}

func TestPathJumpDialog_LongInputScrolling(t *testing.T) {
	dialog := NewPathJumpDialog()

	// Set a very long path to trigger scrolling
	longPath := "/home/user/very/long/path/that/exceeds/the/display/width/significantly/more/and/more"
	dialog.SetInput(longPath)

	// Verify view is rendered without panic
	view := dialog.View()
	if view == "" {
		t.Error("View should not be empty for long input")
	}

	// Move cursor to end and beyond display width
	dialog.textInput.CursorPos = len([]rune(longPath))
	view = dialog.View()
	if view == "" {
		t.Error("View should not be empty after cursor movement")
	}
}

func TestPathJumpDialog_CursorInMiddleWithSuggestion(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	dialog := NewPathJumpDialog()
	dialog.SetInput(filepath.Join(tmpDir, "test"))
	dialog.updateSuggestion()

	// Move cursor to middle of input
	dialog.textInput.CursorPos = 5

	view := dialog.View()
	if view == "" {
		t.Error("View should not be empty with cursor in middle")
	}
}

func TestPathJumpDialog_VeryLongInputWithStartPosBeyondRunes(t *testing.T) {
	dialog := NewPathJumpDialog()

	// Create a path that will cause startPos >= len(runes)
	path := "/a/b/c"
	dialog.textInput.Value = path
	dialog.textInput.CursorPos = 100 // Position beyond input length

	// This should trigger the else branch in renderInputWithSuggestion
	view := dialog.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}
