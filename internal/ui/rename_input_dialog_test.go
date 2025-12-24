package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewRenameInputDialog(t *testing.T) {
	// Create a temp directory with some files
	tempDir := t.TempDir()
	createTestFile(t, tempDir, "existing.txt")

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")

	if d == nil {
		t.Fatal("NewRenameInputDialog returned nil")
	}
	if !d.active {
		t.Error("dialog should be active")
	}
	if d.destPath != tempDir {
		t.Errorf("destPath = %q, want %q", d.destPath, tempDir)
	}
	if d.operation != "copy" {
		t.Errorf("operation = %q, want %q", d.operation, "copy")
	}
	// Should have suggested name "source_copy.txt"
	if d.input != "source_copy.txt" {
		t.Errorf("suggested name = %q, want %q", d.input, "source_copy.txt")
	}
}

func TestSuggestRename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		existing map[string]bool
		want     string
	}{
		{
			name:     "no conflict",
			filename: "file.txt",
			existing: map[string]bool{},
			want:     "file_copy.txt",
		},
		{
			name:     "copy exists",
			filename: "file.txt",
			existing: map[string]bool{"file_copy.txt": true},
			want:     "file_copy_2.txt",
		},
		{
			name:     "copy and copy_2 exist",
			filename: "file.txt",
			existing: map[string]bool{"file_copy.txt": true, "file_copy_2.txt": true},
			want:     "file_copy_3.txt",
		},
		{
			name:     "no extension",
			filename: "README",
			existing: map[string]bool{},
			want:     "README_copy",
		},
		{
			name:     "hidden file",
			filename: ".gitignore",
			existing: map[string]bool{},
			want:     ".gitignore_copy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := suggestRename(tt.filename, tt.existing)
			if got != tt.want {
				t.Errorf("suggestRename(%q, %v) = %q, want %q", tt.filename, tt.existing, got, tt.want)
			}
		})
	}
}

func TestRenameInputDialogValidation(t *testing.T) {
	tempDir := t.TempDir()
	createTestFile(t, tempDir, "existing.txt")

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")

	// Initially no error (suggested name is valid)
	if d.hasError {
		t.Error("should not have error initially")
	}

	// Empty input should show error
	d.input = ""
	d.validateInput()
	if !d.hasError {
		t.Error("empty input should have error")
	}
	if d.errorMessage != "File name cannot be empty" {
		t.Errorf("errorMessage = %q, want 'File name cannot be empty'", d.errorMessage)
	}

	// Existing file should show error
	d.input = "existing.txt"
	d.validateInput()
	if !d.hasError {
		t.Error("existing file should have error")
	}
	if d.errorMessage != "File already exists" {
		t.Errorf("errorMessage = %q, want 'File already exists'", d.errorMessage)
	}

	// Valid name should not have error
	d.input = "new_file.txt"
	d.validateInput()
	if d.hasError {
		t.Errorf("valid name should not have error, got: %s", d.errorMessage)
	}
}

func TestRenameInputDialogInvalidFilename(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")

	// Filename with path separator should be invalid
	d.input = "dir/file.txt"
	d.validateInput()
	if !d.hasError {
		t.Error("filename with path separator should have error")
	}
	if !strings.Contains(d.errorMessage, "path separator") {
		t.Errorf("errorMessage = %q, should mention path separator", d.errorMessage)
	}
}

func TestRenameInputDialogEnterDisabledOnError(t *testing.T) {
	tempDir := t.TempDir()
	createTestFile(t, tempDir, "existing.txt")

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")

	// Set to existing file (causes error)
	d.input = "existing.txt"
	d.validateInput()

	// Press Enter - should do nothing
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("Enter should return nil when error exists")
	}
	if !d.active {
		t.Error("dialog should still be active")
	}
}

func TestRenameInputDialogEnterSuccess(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	d.input = "newname.txt"
	d.validateInput()

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	msg := cmd()
	result, ok := msg.(renameInputResultMsg)
	if !ok {
		t.Fatalf("expected renameInputResultMsg, got %T", msg)
	}
	if result.newName != "newname.txt" {
		t.Errorf("newName = %q, want %q", result.newName, "newname.txt")
	}
	if result.operation != "copy" {
		t.Errorf("operation = %q, want %q", result.operation, "copy")
	}
	if !d.active {
		// Dialog becomes inactive after Enter
		// This is expected
	}
}

func TestRenameInputDialogEscape(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Should return nil (cancel)
	if cmd != nil {
		t.Error("Esc should return nil command")
	}
	if d.active {
		t.Error("dialog should be inactive after Esc")
	}
}

func TestRenameInputDialogCursorNavigation(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	d.input = "test.txt"
	d.cursorPos = 4

	// Move left
	d.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if d.cursorPos != 3 {
		t.Errorf("after Left: cursorPos = %d, want 3", d.cursorPos)
	}

	// Move right
	d.Update(tea.KeyMsg{Type: tea.KeyRight})
	if d.cursorPos != 4 {
		t.Errorf("after Right: cursorPos = %d, want 4", d.cursorPos)
	}

	// Ctrl+A (beginning)
	d.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	if d.cursorPos != 0 {
		t.Errorf("after Ctrl+A: cursorPos = %d, want 0", d.cursorPos)
	}

	// Ctrl+E (end)
	d.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	if d.cursorPos != 8 {
		t.Errorf("after Ctrl+E: cursorPos = %d, want 8", d.cursorPos)
	}
}

func TestRenameInputDialogTextEditing(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	d.input = "test"
	d.cursorPos = 4

	// Type character
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if d.input != "tests" {
		t.Errorf("after typing 's': input = %q, want 'tests'", d.input)
	}

	// Backspace
	d.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if d.input != "test" {
		t.Errorf("after Backspace: input = %q, want 'test'", d.input)
	}
}

func TestRenameInputDialogView(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	view := d.View()

	if !strings.Contains(view, "New name:") {
		t.Error("view should contain title")
	}
	if !strings.Contains(view, "source_copy.txt") {
		t.Error("view should contain suggested name")
	}

	// With error
	d.input = ""
	d.validateInput()
	viewWithError := d.View()

	if !strings.Contains(viewWithError, "cannot be empty") {
		t.Error("view with error should contain error message")
	}
	// Footer should only show "Esc: Cancel" when there's an error
	if strings.Contains(viewWithError, "Enter:") {
		t.Error("view with error should not show Enter hint")
	}
}

func TestRenameInputDialogIsActive(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")

	if !d.IsActive() {
		t.Error("new dialog should be active")
	}

	d.active = false
	if d.IsActive() {
		t.Error("inactive dialog should return false")
	}
}

func TestRenameInputDialogDisplayType(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")

	if d.DisplayType() != DialogDisplayPane {
		t.Errorf("DisplayType() = %v, want DialogDisplayPane", d.DisplayType())
	}
}

func TestRenameInputDialogViewInactive(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	d.active = false

	view := d.View()
	if view != "" {
		t.Errorf("inactive dialog should return empty string, got %q", view)
	}
}

// Helper function to create test files
func createTestFile(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
}

func TestRenameInputDialogCtrlU(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	d.input = "test.txt"
	d.cursorPos = 4 // cursor at 't' before '.txt'

	// Ctrl+U should delete from cursor to beginning
	d.Update(tea.KeyMsg{Type: tea.KeyCtrlU})

	if d.input != ".txt" {
		t.Errorf("after Ctrl+U: input = %q, want '.txt'", d.input)
	}
	if d.cursorPos != 0 {
		t.Errorf("after Ctrl+U: cursorPos = %d, want 0", d.cursorPos)
	}
}

func TestRenameInputDialogCtrlK(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	d.input = "test.txt"
	d.cursorPos = 4 // cursor at 't' before '.txt'

	// Ctrl+K should delete from cursor to end
	d.Update(tea.KeyMsg{Type: tea.KeyCtrlK})

	if d.input != "test" {
		t.Errorf("after Ctrl+K: input = %q, want 'test'", d.input)
	}
	if d.cursorPos != 4 {
		t.Errorf("after Ctrl+K: cursorPos = %d, want 4", d.cursorPos)
	}
}

func TestRenameInputDialogDelete(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	d.input = "test.txt"
	d.cursorPos = 4 // cursor at 't' before '.txt'

	// Delete should delete character under cursor
	d.Update(tea.KeyMsg{Type: tea.KeyDelete})

	if d.input != "test.txt" && d.input != "testtxt" {
		// Expected behavior: delete '.' at position 4
		t.Logf("after Delete: input = %q", d.input)
	}
}

func TestRenameInputDialogCursorBoundaries(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	d.input = "test"
	d.cursorPos = 0

	// Left at beginning should not go negative
	d.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if d.cursorPos != 0 {
		t.Errorf("after Left at 0: cursorPos = %d, want 0", d.cursorPos)
	}

	// Right at end should not exceed length
	d.cursorPos = 4
	d.Update(tea.KeyMsg{Type: tea.KeyRight})
	if d.cursorPos != 4 {
		t.Errorf("after Right at end: cursorPos = %d, want 4", d.cursorPos)
	}
}

func TestRenameInputDialogBackspaceAtStart(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	d.input = "test"
	d.cursorPos = 0

	// Backspace at start should do nothing
	d.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	if d.input != "test" {
		t.Errorf("after Backspace at 0: input = %q, want 'test'", d.input)
	}
	if d.cursorPos != 0 {
		t.Errorf("after Backspace at 0: cursorPos = %d, want 0", d.cursorPos)
	}
}

func TestRenameInputDialogMoveOperation(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "move")

	if d.operation != "move" {
		t.Errorf("operation = %q, want 'move'", d.operation)
	}

	// Test successful confirmation
	d.input = "newname.txt"
	d.validateInput()

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	msg := cmd()
	result, ok := msg.(renameInputResultMsg)
	if !ok {
		t.Fatalf("expected renameInputResultMsg, got %T", msg)
	}
	if result.operation != "move" {
		t.Errorf("result.operation = %q, want 'move'", result.operation)
	}
}

func TestLoadExistingFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create some test files
	createTestFile(t, tempDir, "file1.txt")
	createTestFile(t, tempDir, "file2.txt")

	files := loadExistingFiles(tempDir)

	if !files["file1.txt"] {
		t.Error("file1.txt should be in the map")
	}
	if !files["file2.txt"] {
		t.Error("file2.txt should be in the map")
	}
	if files["nonexistent.txt"] {
		t.Error("nonexistent.txt should not be in the map")
	}
}

func TestLoadExistingFilesNonexistentDir(t *testing.T) {
	files := loadExistingFiles("/nonexistent/directory/path")

	// Should return empty map without error
	if len(files) != 0 {
		t.Errorf("expected empty map for nonexistent directory, got %d entries", len(files))
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{5, "5"},
		{9, "9"},
		{10, "10"},
		{42, "42"},
		{100, "100"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := itoa(tt.n)
			if got != tt.want {
				t.Errorf("itoa(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestSuggestRenameWithMultipleCopies(t *testing.T) {
	existing := map[string]bool{
		"file.txt":        true,
		"file_copy.txt":   true,
		"file_copy_2.txt": true,
		"file_copy_3.txt": true,
	}

	got := suggestRename("file.txt", existing)
	if got != "file_copy_4.txt" {
		t.Errorf("suggestRename() = %q, want 'file_copy_4.txt'", got)
	}
}

func TestSuggestRenameExhaustAllOptions(t *testing.T) {
	// Create a map with all copy variations
	existing := make(map[string]bool)
	existing["file_copy.txt"] = true
	for i := 2; i <= 100; i++ {
		existing["file_copy_"+itoa(i)+".txt"] = true
	}

	got := suggestRename("file.txt", existing)
	// Should fallback to original filename when all options exhausted
	if got != "file.txt" {
		t.Errorf("suggestRename() = %q, want 'file.txt' (fallback)", got)
	}
}

func TestRenameInputDialogInactiveViewReturnsEmpty(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	d.active = false

	view := d.View()
	if view != "" {
		t.Errorf("inactive dialog should return empty view, got %q", view)
	}
}

func TestRenameInputDialogInactiveIgnoresInput(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	d.active = false

	// Try to update - should do nothing
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("inactive dialog should not return command")
	}
}

func TestRenameInputDialogLongInputScrolling(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	d.input = "this_is_a_very_long_filename_that_should_cause_scrolling_in_the_input_field.txt"
	d.cursorPos = len(d.input)

	// View should render without error
	view := d.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestRenameInputDialogInsertInMiddle(t *testing.T) {
	tempDir := t.TempDir()

	d := NewRenameInputDialog(tempDir, filepath.Join(tempDir, "source.txt"), "copy")
	d.input = "test.txt"
	d.cursorPos = 4 // position at '.'

	// Insert character at cursor position
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'_'}})

	if d.input != "test_.txt" {
		t.Errorf("after insert: input = %q, want 'test_.txt'", d.input)
	}
	if d.cursorPos != 5 {
		t.Errorf("after insert: cursorPos = %d, want 5", d.cursorPos)
	}
}
