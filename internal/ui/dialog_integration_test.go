package ui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Integration tests for dialog cancellation flows.
// These tests verify that dialogs correctly integrate with Model
// by sending cancellation messages and allowing Model to clear dialog references.

// TestPermissionDialogCancellationIntegration tests PermissionDialog cancellation
func TestPermissionDialogCancellationIntegration(t *testing.T) {
	tempDir := t.TempDir()
	createTestFilesForIntegration(t, tempDir, []string{"file1.txt"})

	m := createTestModel(t, tempDir)

	// Open PermissionDialog
	m.dialog = NewPermissionDialog("file1.txt", false, 0644)
	if m.dialog == nil || !m.dialog.IsActive() {
		t.Fatal("PermissionDialog should be active")
	}

	// Send Esc key
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if cmd == nil {
		t.Fatal("Dialog should return cancel message command")
	}

	// Process cancel message
	msg := cmd()
	updated, _ = m.Update(msg)
	m = updated.(Model)

	// Verify dialog is cleared
	if m.dialog != nil {
		t.Error("Model should clear dialog after cancel message")
	}
}

// TestRecursivePermDialogCancellationIntegration tests RecursivePermDialog cancellation
func TestRecursivePermDialogCancellationIntegration(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := createTestModel(t, tempDir)

	// Open RecursivePermDialog
	m.dialog = NewRecursivePermDialog("testdir")
	if m.dialog == nil || !m.dialog.IsActive() {
		t.Fatal("RecursivePermDialog should be active")
	}

	// Send Esc key
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if cmd == nil {
		t.Fatal("Dialog should return cancel message command")
	}

	// Process cancel message
	msg := cmd()
	updated, _ = m.Update(msg)
	m = updated.(Model)

	// Verify dialog is cleared
	if m.dialog != nil {
		t.Error("Model should clear dialog after cancel message")
	}
}

// TestInputDialogCancellationIntegration tests InputDialog cancellation
func TestInputDialogCancellationIntegration(t *testing.T) {
	tempDir := t.TempDir()
	m := createTestModel(t, tempDir)

	callbackInvoked := false
	m.dialog = NewInputDialog("Test", func(input string) tea.Cmd {
		callbackInvoked = true
		return nil
	})

	if m.dialog == nil || !m.dialog.IsActive() {
		t.Fatal("InputDialog should be active")
	}

	// Send Esc key
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if cmd == nil {
		t.Fatal("Dialog should return cancel message command")
	}

	// Process cancel message
	msg := cmd()
	updated, _ = m.Update(msg)
	m = updated.(Model)

	// Verify dialog is cleared
	if m.dialog != nil {
		t.Error("Model should clear dialog after cancel message")
	}

	// Verify callback was NOT invoked
	if callbackInvoked {
		t.Error("Callback should not be invoked when dialog is cancelled")
	}
}

// TestRenameInputDialogCancellationIntegration tests RenameInputDialog cancellation
func TestRenameInputDialogCancellationIntegration(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	createTestFilesForIntegration(t, tempDir, []string{"source.txt"})

	m := createTestModel(t, tempDir)

	// Open RenameInputDialog
	m.dialog = NewRenameInputDialog(tempDir, srcFile, "copy")
	if m.dialog == nil || !m.dialog.IsActive() {
		t.Fatal("RenameInputDialog should be active")
	}

	// Send Esc key
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if cmd == nil {
		t.Fatal("Dialog should return cancel message command")
	}

	// Process cancel message
	msg := cmd()
	updated, _ = m.Update(msg)
	m = updated.(Model)

	// Verify dialog is cleared
	if m.dialog != nil {
		t.Error("Model should clear dialog after cancel message")
	}
}

// TestMultipleDialogCancellationSequence tests multiple dialog cancellations in sequence
func TestMultipleDialogCancellationSequence(t *testing.T) {
	tempDir := t.TempDir()
	createTestFilesForIntegration(t, tempDir, []string{"file1.txt"})

	m := createTestModel(t, tempDir)

	dialogs := []Dialog{
		NewPermissionDialog("file1.txt", false, 0644),
		NewErrorDialog("Test error"),
		NewHelpDialog(),
	}

	for i, dialog := range dialogs {
		// Open dialog
		m.dialog = dialog
		if m.dialog == nil || !m.dialog.IsActive() {
			t.Fatalf("Dialog %d should be active", i)
		}

		// Cancel dialog
		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = updated.(Model)

		if cmd != nil {
			msg := cmd()
			updated, _ = m.Update(msg)
			m = updated.(Model)
		}

		// Verify dialog cleared
		if m.dialog != nil {
			t.Errorf("Dialog %d should be cleared after cancel", i)
		}
	}
}

// TestDialogCancelAndReopenIntegration tests reopening a dialog after cancelling previous one
func TestDialogCancelAndReopenIntegration(t *testing.T) {
	tempDir := t.TempDir()
	createTestFilesForIntegration(t, tempDir, []string{"file1.txt", "file2.txt"})

	m := createTestModel(t, tempDir)

	// First dialog: PermissionDialog
	m.dialog = NewPermissionDialog("file1.txt", false, 0644)
	if m.dialog == nil || !m.dialog.IsActive() {
		t.Fatal("First PermissionDialog should be active")
	}

	// Cancel first dialog with Esc
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if cmd != nil {
		msg := cmd()
		updated, _ = m.Update(msg)
		m = updated.(Model)
	}

	// Verify first dialog is cleared
	if m.dialog != nil {
		t.Error("First dialog should be cleared after cancel")
	}

	// Simulate cursor movement (using 'j' key)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)

	// Second dialog: Another PermissionDialog for different file
	m.dialog = NewPermissionDialog("file2.txt", false, 0755)

	if m.dialog == nil || !m.dialog.IsActive() {
		t.Fatal("Second PermissionDialog should be active")
	}

	// Verify second dialog can be cancelled normally
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if cmd != nil {
		msg := cmd()
		updated, _ = m.Update(msg)
		m = updated.(Model)
	}

	// Verify second dialog is cleared
	if m.dialog != nil {
		t.Error("Second dialog should be cleared after cancel")
	}
}

// TestInputDialogCancelWithPartialInputIntegration tests cancelling InputDialog with partial input
func TestInputDialogCancelWithPartialInputIntegration(t *testing.T) {
	tempDir := t.TempDir()
	m := createTestModel(t, tempDir)

	callbackInvoked := false
	var capturedInput string

	m.dialog = NewInputDialog("Test Input", func(input string) tea.Cmd {
		callbackInvoked = true
		capturedInput = input
		return nil
	})

	if m.dialog == nil || !m.dialog.IsActive() {
		t.Fatal("InputDialog should be active")
	}

	// Type partial input
	testInput := "testfile"
	for _, r := range testInput {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	// Cancel with Esc
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if cmd != nil {
		msg := cmd()
		updated, _ = m.Update(msg)
		m = updated.(Model)
	}

	// Verify dialog is cleared
	if m.dialog != nil {
		t.Error("Dialog should be cleared after cancel")
	}

	// Verify callback was NOT invoked
	if callbackInvoked {
		t.Error("Callback should not be invoked when dialog is cancelled")
	}

	// Verify input was not captured
	if capturedInput != "" {
		t.Errorf("Input should not be captured on cancel, got: %s", capturedInput)
	}
}

// TestMultipleEscKeyPressesIntegration tests multiple consecutive Esc key presses
func TestMultipleEscKeyPressesIntegration(t *testing.T) {
	tempDir := t.TempDir()
	createTestFilesForIntegration(t, tempDir, []string{"file1.txt"})

	m := createTestModel(t, tempDir)

	// Open PermissionDialog
	m.dialog = NewPermissionDialog("file1.txt", false, 0644)
	if m.dialog == nil || !m.dialog.IsActive() {
		t.Fatal("PermissionDialog should be active")
	}

	// Press Esc multiple times consecutively
	for i := 0; i < 5; i++ {
		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = updated.(Model)

		if cmd != nil {
			msg := cmd()
			updated, _ = m.Update(msg)
			m = updated.(Model)
		}

		// After first Esc, dialog should be closed
		if i == 0 && m.dialog != nil {
			t.Error("Dialog should be cleared after first Esc")
		}

		// Subsequent Esc presses should not cause errors
		if m.dialog != nil {
			t.Errorf("Dialog should remain nil after Esc press %d", i+1)
		}
	}

	// Verify model is still functional by performing navigation
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)

	// No error should occur - this verifies model is in normal state
}

// TestBookmarkDialogCancellationIntegration tests BookmarkDialog cancellation
func TestBookmarkDialogCancellationIntegration(t *testing.T) {
	tempDir := t.TempDir()
	m := createTestModel(t, tempDir)

	// Note: BookmarkDialog expects []config.Bookmark, but for testing we'll create empty slice
	// since we only need to test cancellation behavior, not bookmark functionality
	m.dialog = NewBookmarkDialog(nil)

	if m.dialog == nil || !m.dialog.IsActive() {
		t.Fatal("BookmarkDialog should be active")
	}

	// Send Esc key
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if cmd != nil {
		msg := cmd()
		updated, _ = m.Update(msg)
		m = updated.(Model)
	}

	// Verify dialog is cleared
	if m.dialog != nil {
		t.Error("BookmarkDialog should be cleared after cancel")
	}

	// Verify model returns to normal operation
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
}

// TestContextMenuCancellationIntegration tests ContextMenu cancellation
func TestContextMenuCancellationIntegration(t *testing.T) {
	tempDir := t.TempDir()
	createTestFilesForIntegration(t, tempDir, []string{"file1.txt"})

	m := createTestModel(t, tempDir)

	// Get first file entry for context menu
	if len(m.leftPane.entries) == 0 {
		t.Fatal("Expected at least one file in pane")
	}

	entry := &m.leftPane.entries[0]

	// Open ContextMenuDialog
	m.dialog = NewContextMenuDialog(entry, tempDir, tempDir)

	if m.dialog == nil || !m.dialog.IsActive() {
		t.Fatal("ContextMenuDialog should be active")
	}

	// Send Esc key
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if cmd != nil {
		msg := cmd()
		updated, _ = m.Update(msg)
		m = updated.(Model)
	}

	// Verify dialog is cleared
	if m.dialog != nil {
		t.Error("ContextMenuDialog should be cleared after cancel")
	}

	// Verify model returns to normal operation
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
}

// Helper functions

// createTestModel creates a minimal Model for testing
func createTestModel(t *testing.T, tempDir string) Model {
	t.Helper()

	leftPane, err := NewPane(LeftPane, tempDir, 40, 20, true, DefaultTheme())
	if err != nil {
		t.Fatalf("Failed to create left pane: %v", err)
	}

	rightPane, err := NewPane(RightPane, tempDir, 40, 20, false, DefaultTheme())
	if err != nil {
		t.Fatalf("Failed to create right pane: %v", err)
	}

	return Model{
		leftPane:   leftPane,
		rightPane:  rightPane,
		leftPath:   tempDir,
		rightPath:  tempDir,
		activePane: LeftPane,
	}
}

// createTestFilesForIntegration creates test files in the specified directory
func createTestFilesForIntegration(t *testing.T, dir string, filenames []string) {
	t.Helper()

	for _, filename := range filenames {
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}
}
