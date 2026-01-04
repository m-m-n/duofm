package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewOpenWithDialog tests dialog creation
func TestNewOpenWithDialog(t *testing.T) {
	files := []string{"test.txt"}
	workDir := "/test/dir"

	dialog := NewOpenWithDialog(files, workDir)

	if dialog == nil {
		t.Fatal("NewOpenWithDialog returned nil")
	}

	if !dialog.IsActive() {
		t.Error("dialog should be active after creation")
	}

	if len(dialog.fileList) != 1 {
		t.Errorf("fileList length = %d, want 1", len(dialog.fileList))
	}

	if dialog.workDir != workDir {
		t.Errorf("workDir = %s, want %s", dialog.workDir, workDir)
	}
}

// TestOpenWithDialog_FileListFormatting tests file list display formatting
func TestOpenWithDialog_FileListFormatting(t *testing.T) {
	tests := []struct {
		name           string
		files          []string
		expectContains string
	}{
		{
			name:           "single file",
			files:          []string{"test.txt"},
			expectContains: `"test.txt"`,
		},
		{
			name:           "multiple files",
			files:          []string{"file1.txt", "file2.txt"},
			expectContains: `"file1.txt" "file2.txt"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewOpenWithDialog(tt.files, "/test")

			if !strings.Contains(dialog.filesDisplay, tt.expectContains) {
				t.Errorf("filesDisplay = %s, want to contain %s", dialog.filesDisplay, tt.expectContains)
			}
		})
	}
}

// TestOpenWithDialog_Update_Enter tests Enter key sends result
func TestOpenWithDialog_Update_Enter(t *testing.T) {
	dialog := NewOpenWithDialog([]string{"test.txt"}, "/test")

	// Set application name
	dialog.applicationInput.SetValue("mpv")

	// Press Enter
	updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if updatedDialog.IsActive() {
		t.Error("dialog should be closed after Enter")
	}

	if cmd == nil {
		t.Fatal("cmd should not be nil after Enter")
	}

	msg := cmd()
	result, ok := msg.(openWithDialogResultMsg)
	if !ok {
		t.Fatal("cmd() did not return openWithDialogResultMsg")
	}

	if result.cancelled {
		t.Error("result should not be cancelled after Enter")
	}

	if result.application != "mpv" {
		t.Errorf("result.application = %s, want mpv", result.application)
	}

	if len(result.files) != 1 {
		t.Errorf("result.files length = %d, want 1", len(result.files))
	}
}

// TestOpenWithDialog_Update_Esc tests Esc key cancels
func TestOpenWithDialog_Update_Esc(t *testing.T) {
	dialog := NewOpenWithDialog([]string{"test.txt"}, "/test")

	// Press Esc
	updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if updatedDialog.IsActive() {
		t.Error("dialog should be closed after Esc")
	}

	if cmd == nil {
		t.Fatal("cmd should not be nil after Esc")
	}

	msg := cmd()
	result, ok := msg.(openWithDialogResultMsg)
	if !ok {
		t.Fatal("cmd() did not return openWithDialogResultMsg")
	}

	if !result.cancelled {
		t.Error("result should be cancelled after Esc")
	}
}

// TestOpenWithDialog_Update_EmptyApplication tests Enter with empty application does nothing
func TestOpenWithDialog_Update_EmptyApplication(t *testing.T) {
	dialog := NewOpenWithDialog([]string{"test.txt"}, "/test")

	// Application input is empty by default
	// Press Enter
	updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Dialog should remain active
	if !updatedDialog.IsActive() {
		t.Error("dialog should remain active when application is empty")
	}

	// No command should be returned
	if cmd != nil {
		t.Error("cmd should be nil when application is empty")
	}
}
