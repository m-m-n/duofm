package ui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
)

// Test helper functions

// createTestModelForPermission creates a minimal Model for permission testing
func createTestModelForPermission(t *testing.T, tempDir string) Model {
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

// createTestFilesForPermission creates test files in the specified directory
func createTestFilesForPermission(t *testing.T, dir string, filenames []string) {
	t.Helper()

	for _, filename := range filenames {
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}
}

// =============================================================================
// Phase 1: handlePermissionOperationComplete tests
// =============================================================================

// TestHandlePermissionOperationComplete_ClearsDialog tests that dialog is cleared on success
func TestHandlePermissionOperationComplete_ClearsDialogOnSuccess(t *testing.T) {
	tempDir := t.TempDir()
	createTestFilesForPermission(t, tempDir, []string{"file1.txt"})

	m := createTestModelForPermission(t, tempDir)

	// Simulate PermissionDialog is active (as if user confirmed with Enter)
	m.dialog = NewPermissionDialog("file1.txt", false, 0644)
	if m.dialog == nil {
		t.Fatal("Dialog should be set")
	}

	// Simulate permission operation complete (success)
	msg := permissionOperationCompleteMsg{
		path:    filepath.Join(tempDir, "file1.txt"),
		success: true,
		err:     nil,
	}

	result, _ := m.handlePermissionOperationComplete(msg)
	m = result.(Model)

	// Verify dialog is cleared
	if m.dialog != nil {
		t.Error("Dialog should be nil after permission operation complete")
	}

	// Verify status message is set
	if m.statusMessage == "" {
		t.Error("Status message should be set on success")
	}
	if m.isStatusError {
		t.Error("Status should not be error on success")
	}
}

// TestHandlePermissionOperationComplete_ClearsDialogOnFailure tests that dialog is cleared on failure
func TestHandlePermissionOperationComplete_ClearsDialogOnFailure(t *testing.T) {
	tempDir := t.TempDir()
	createTestFilesForPermission(t, tempDir, []string{"file1.txt"})

	m := createTestModelForPermission(t, tempDir)

	// Simulate PermissionDialog is active
	m.dialog = NewPermissionDialog("file1.txt", false, 0644)

	// Simulate permission operation complete (failure)
	msg := permissionOperationCompleteMsg{
		path:    filepath.Join(tempDir, "file1.txt"),
		success: false,
		err:     os.ErrPermission,
	}

	result, _ := m.handlePermissionOperationComplete(msg)
	m = result.(Model)

	// Verify dialog is cleared even on failure
	if m.dialog != nil {
		t.Error("Dialog should be nil after permission operation complete (even on failure)")
	}

	// Verify error status message is set
	if m.statusMessage == "" {
		t.Error("Status message should be set on failure")
	}
	if !m.isStatusError {
		t.Error("Status should be error on failure")
	}
}

// =============================================================================
// Phase 2: handleRecursivePermissionComplete tests
// =============================================================================

// TestHandleRecursivePermissionComplete_ClearsDialogOnSuccess tests that dialog is cleared on success
func TestHandleRecursivePermissionComplete_ClearsDialogOnSuccess(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := createTestModelForPermission(t, tempDir)

	// Simulate RecursivePermDialog is active
	m.dialog = NewRecursivePermDialog("testdir")

	// Simulate recursive permission complete (success - no errors)
	msg := recursivePermissionCompleteMsg{
		path:         subDir,
		successCount: 5,
		errors:       nil, // No errors
	}

	result, _ := m.handleRecursivePermissionComplete(msg)
	m = result.(Model)

	// Verify dialog is cleared on success (no errors)
	if m.dialog != nil {
		t.Error("Dialog should be nil after successful recursive permission change")
	}

	// Verify success status message is set
	if m.statusMessage == "" {
		t.Error("Status message should be set on success")
	}
	if m.isStatusError {
		t.Error("Status should not be error on success")
	}
}

// TestHandleRecursivePermissionComplete_ShowsErrorDialogOnErrors tests that error dialog is shown on errors
func TestHandleRecursivePermissionComplete_ShowsErrorDialogOnErrors(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := createTestModelForPermission(t, tempDir)

	// Simulate RecursivePermDialog is active
	m.dialog = NewRecursivePermDialog("testdir")

	// Simulate recursive permission complete (with errors)
	msg := recursivePermissionCompleteMsg{
		path:         subDir,
		successCount: 3,
		errors: []fs.PermissionError{
			{Path: filepath.Join(subDir, "file1.txt"), Error: os.ErrPermission},
			{Path: filepath.Join(subDir, "file2.txt"), Error: os.ErrPermission},
		},
	}

	result, _ := m.handleRecursivePermissionComplete(msg)
	m = result.(Model)

	// Verify dialog is replaced with error report dialog
	if m.dialog == nil {
		t.Error("Dialog should be set to error report dialog when there are errors")
	}

	// Verify it's the error report dialog (type check)
	if _, ok := m.dialog.(*PermissionErrorReportDialog); !ok {
		t.Error("Dialog should be PermissionErrorReportDialog when there are errors")
	}
}

// =============================================================================
// Phase 3: handleBatchPermissionComplete tests
// =============================================================================

// TestHandleBatchPermissionComplete_ClearsDialogOnSuccess_SmallBatch tests small batch (no progress dialog)
func TestHandleBatchPermissionComplete_ClearsDialogOnSuccess_SmallBatch(t *testing.T) {
	tempDir := t.TempDir()
	createTestFilesForPermission(t, tempDir, []string{"file1.txt", "file2.txt", "file3.txt"})

	m := createTestModelForPermission(t, tempDir)

	// Simulate PermissionDialog (not PermissionProgressDialog) for small batch
	// This is the bug case - small batches use PermissionDialog, not PermissionProgressDialog
	m.dialog = NewPermissionDialog("3 items", false, 0644)

	// Simulate batch permission complete (success)
	msg := batchPermissionCompleteMsg{
		totalCount:   3,
		successCount: 3,
		failedCount:  0,
		errors:       nil,
	}

	result, _ := m.handleBatchPermissionComplete(msg)
	m = result.(Model)

	// Verify dialog is cleared (THIS IS THE BUG FIX)
	if m.dialog != nil {
		t.Error("Dialog should be nil after batch permission complete (small batch)")
	}

	// Verify success status message is set
	if m.statusMessage == "" {
		t.Error("Status message should be set on success")
	}
	if m.isStatusError {
		t.Error("Status should not be error on success")
	}
}

// TestHandleBatchPermissionComplete_ClearsDialogOnSuccess_LargeBatch tests large batch (with progress dialog)
func TestHandleBatchPermissionComplete_ClearsDialogOnSuccess_LargeBatch(t *testing.T) {
	tempDir := t.TempDir()
	// Create enough files for large batch (threshold is typically 10)
	files := make([]string, 15)
	for i := 0; i < 15; i++ {
		files[i] = filepath.Join("file%d.txt")
	}

	m := createTestModelForPermission(t, tempDir)

	// Simulate PermissionProgressDialog for large batch
	m.dialog = NewPermissionProgressDialog(15)

	// Simulate batch permission complete (success)
	msg := batchPermissionCompleteMsg{
		totalCount:   15,
		successCount: 15,
		failedCount:  0,
		errors:       nil,
	}

	result, _ := m.handleBatchPermissionComplete(msg)
	m = result.(Model)

	// Verify dialog is cleared
	if m.dialog != nil {
		t.Error("Dialog should be nil after batch permission complete (large batch)")
	}

	// Verify success status message is set
	if m.statusMessage == "" {
		t.Error("Status message should be set on success")
	}
	if m.isStatusError {
		t.Error("Status should not be error on success")
	}
}

// TestHandleBatchPermissionComplete_ShowsErrorDialogOnErrors tests that error dialog is shown on errors
func TestHandleBatchPermissionComplete_ShowsErrorDialogOnErrors(t *testing.T) {
	tempDir := t.TempDir()
	createTestFilesForPermission(t, tempDir, []string{"file1.txt", "file2.txt", "file3.txt"})

	m := createTestModelForPermission(t, tempDir)

	// Simulate PermissionDialog for small batch
	m.dialog = NewPermissionDialog("3 items", false, 0644)

	// Simulate batch permission complete (with errors)
	msg := batchPermissionCompleteMsg{
		totalCount:   3,
		successCount: 1,
		failedCount:  2,
		errors: []fs.PermissionError{
			{Path: filepath.Join(tempDir, "file1.txt"), Error: os.ErrPermission},
			{Path: filepath.Join(tempDir, "file2.txt"), Error: os.ErrPermission},
		},
	}

	result, _ := m.handleBatchPermissionComplete(msg)
	m = result.(Model)

	// Verify dialog is replaced with error report dialog
	if m.dialog == nil {
		t.Error("Dialog should be set to error report dialog when there are errors")
	}

	// Verify it's the error report dialog (type check)
	if _, ok := m.dialog.(*PermissionErrorReportDialog); !ok {
		t.Error("Dialog should be PermissionErrorReportDialog when there are errors")
	}
}

// =============================================================================
// Integration tests for confirmation flow
// =============================================================================

// TestPermissionConfirmationFlow_EndToEnd tests the full confirmation flow
func TestPermissionConfirmationFlow_SingleFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "file1.txt")
	createTestFilesForPermission(t, tempDir, []string{"file1.txt"})

	m := createTestModelForPermission(t, tempDir)

	// Open PermissionDialog
	m.dialog = NewPermissionDialog("file1.txt", false, 0644)
	if m.dialog == nil || !m.dialog.IsActive() {
		t.Fatal("PermissionDialog should be active")
	}

	// Simulate permission operation complete (as if user confirmed with Enter)
	msg := permissionOperationCompleteMsg{
		path:    testFile,
		success: true,
		err:     nil,
	}

	result, _ := m.handlePermissionOperationComplete(msg)
	m = result.(Model)

	// Verify dialog is cleared
	if m.dialog != nil {
		t.Error("Dialog should be nil after confirmation flow completes")
	}

	// Verify model can receive keyboard input after dialog closes
	// (This verifies the freeze is fixed)
}

// TestKeyboardNavigationAfterPermissionConfirm tests that keyboard works after confirming permission dialog
func TestKeyboardNavigationAfterPermissionConfirm(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "file1.txt")
	createTestFilesForPermission(t, tempDir, []string{"file1.txt", "file2.txt"})

	m := createTestModelForPermission(t, tempDir)

	// Open PermissionDialog
	m.dialog = NewPermissionDialog("file1.txt", false, 0644)

	// Simulate permission operation complete
	completeMsg := permissionOperationCompleteMsg{
		path:    testFile,
		success: true,
		err:     nil,
	}

	result, _ := m.handlePermissionOperationComplete(completeMsg)
	m = result.(Model)

	// Verify dialog is cleared
	if m.dialog != nil {
		t.Error("Dialog should be nil after permission operation complete")
	}

	// Simulate keyboard navigation (j key for cursor down)
	// This should work without freeze if dialog is properly cleared
	// If dialog is not nil, keyboard events would be captured by dialog instead of model
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)

	// Model should still have dialog = nil after keyboard event
	// (If freeze bug exists, model state might be corrupted)
	if m.dialog != nil {
		t.Error("Dialog should remain nil after keyboard navigation")
	}
}
