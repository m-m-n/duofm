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

// =============================================================================
// Phase 4: handlePermissionMessages tests
// =============================================================================

func TestHandlePermissionMessages_PermissionDialogCancelMsg(t *testing.T) {
	tempDir := t.TempDir()
	createTestFilesForPermission(t, tempDir, []string{"file1.txt"})

	m := createTestModelForPermission(t, tempDir)
	m.dialog = NewPermissionDialog("file1.txt", false, 0644)

	if m.dialog == nil {
		t.Fatal("Dialog should be set before cancel")
	}

	newModel, _, handled := m.handlePermissionMessages(permissionDialogCancelMsg{})

	if !handled {
		t.Error("handlePermissionMessages should handle permissionDialogCancelMsg")
	}

	if newModel.dialog != nil {
		t.Error("Dialog should be nil after cancel")
	}
}

func TestHandlePermissionMessages_RecursivePermDialogCancelMsg(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := createTestModelForPermission(t, tempDir)
	m.dialog = NewRecursivePermDialog("testdir")

	if m.dialog == nil {
		t.Fatal("Dialog should be set before cancel")
	}

	newModel, _, handled := m.handlePermissionMessages(recursivePermDialogCancelMsg{})

	if !handled {
		t.Error("handlePermissionMessages should handle recursivePermDialogCancelMsg")
	}

	if newModel.dialog != nil {
		t.Error("Dialog should be nil after cancel")
	}
}

func TestHandlePermissionMessages_PermissionOperationCompleteMsg(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "file1.txt")
	createTestFilesForPermission(t, tempDir, []string{"file1.txt"})

	m := createTestModelForPermission(t, tempDir)
	m.dialog = NewPermissionDialog("file1.txt", false, 0644)

	msg := permissionOperationCompleteMsg{
		path:    testFile,
		success: true,
		err:     nil,
	}

	newModel, _, handled := m.handlePermissionMessages(msg)

	if !handled {
		t.Error("handlePermissionMessages should handle permissionOperationCompleteMsg")
	}

	if newModel.dialog != nil {
		t.Error("Dialog should be nil after operation complete")
	}
}

func TestHandlePermissionMessages_BatchPermissionCompleteMsg(t *testing.T) {
	tempDir := t.TempDir()
	createTestFilesForPermission(t, tempDir, []string{"file1.txt", "file2.txt", "file3.txt"})

	m := createTestModelForPermission(t, tempDir)
	m.dialog = NewPermissionDialog("3 items", false, 0644)

	msg := batchPermissionCompleteMsg{
		totalCount:   3,
		successCount: 3,
		failedCount:  0,
		errors:       nil,
	}

	newModel, _, handled := m.handlePermissionMessages(msg)

	if !handled {
		t.Error("handlePermissionMessages should handle batchPermissionCompleteMsg")
	}

	if newModel.dialog != nil {
		t.Error("Dialog should be nil after batch operation complete")
	}
}

func TestHandlePermissionMessages_RecursivePermissionCompleteMsg(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := createTestModelForPermission(t, tempDir)
	m.dialog = NewRecursivePermDialog("testdir")

	msg := recursivePermissionCompleteMsg{
		path:         subDir,
		successCount: 5,
		errors:       nil,
	}

	newModel, _, handled := m.handlePermissionMessages(msg)

	if !handled {
		t.Error("handlePermissionMessages should handle recursivePermissionCompleteMsg")
	}

	if newModel.dialog != nil {
		t.Error("Dialog should be nil after recursive operation complete")
	}
}

func TestHandlePermissionMessages_ShowRecursivePermDialogMsg(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := createTestModelForPermission(t, tempDir)

	msg := showRecursivePermDialogMsg{
		path: subDir,
	}

	newModel, _, handled := m.handlePermissionMessages(msg)

	if !handled {
		t.Error("handlePermissionMessages should handle showRecursivePermDialogMsg")
	}

	if newModel.dialog == nil {
		t.Error("Dialog should be set to RecursivePermDialog")
	}

	if _, ok := newModel.dialog.(*RecursivePermDialog); !ok {
		t.Errorf("Dialog should be RecursivePermDialog, got %T", newModel.dialog)
	}
}

func TestHandlePermissionMessages_UnhandledMessage(t *testing.T) {
	tempDir := t.TempDir()
	m := createTestModelForPermission(t, tempDir)

	// Create an unrelated message type
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}

	_, _, handled := m.handlePermissionMessages(msg)

	if handled {
		t.Error("handlePermissionMessages should not handle unrelated messages")
	}
}

// =============================================================================
// Phase 5: executePermissionChange tests
// =============================================================================

func TestExecutePermissionChange_NonRecursive_Success(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "file1.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	m := createTestModelForPermission(t, tempDir)

	cmd := m.executePermissionChange(testFile, "755", false, false)
	if cmd == nil {
		t.Fatal("executePermissionChange should return a command")
	}

	msg := cmd()
	completeMsg, ok := msg.(permissionOperationCompleteMsg)
	if !ok {
		t.Fatalf("Command should return permissionOperationCompleteMsg, got %T", msg)
	}

	if !completeMsg.success {
		t.Errorf("Operation should succeed: %v", completeMsg.err)
	}

	// Verify permission was changed
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatal(err)
	}

	// Permission should be 0755
	expectedPerm := os.FileMode(0755)
	actualPerm := info.Mode().Perm()
	if actualPerm != expectedPerm {
		t.Errorf("File permission = %o, want %o", actualPerm, expectedPerm)
	}
}

func TestExecutePermissionChange_NonRecursive_InvalidMode(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "file1.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	m := createTestModelForPermission(t, tempDir)

	cmd := m.executePermissionChange(testFile, "invalid", false, false)
	if cmd == nil {
		t.Fatal("executePermissionChange should return a command")
	}

	msg := cmd()
	completeMsg, ok := msg.(permissionOperationCompleteMsg)
	if !ok {
		t.Fatalf("Command should return permissionOperationCompleteMsg, got %T", msg)
	}

	if completeMsg.success {
		t.Error("Operation should fail with invalid mode")
	}

	if completeMsg.err == nil {
		t.Error("Error should be set for invalid mode")
	}
}

func TestExecutePermissionChange_Recursive_ShowsDialog(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := createTestModelForPermission(t, tempDir)

	cmd := m.executePermissionChange(testDir, "755", true, true)
	if cmd == nil {
		t.Fatal("executePermissionChange should return a command")
	}

	msg := cmd()
	dialogMsg, ok := msg.(showRecursivePermDialogMsg)
	if !ok {
		t.Fatalf("Recursive mode should return showRecursivePermDialogMsg, got %T", msg)
	}

	if dialogMsg.path != testDir {
		t.Errorf("Dialog path = %s, want %s", dialogMsg.path, testDir)
	}
}

// =============================================================================
// Phase 6: executeRecursivePermissionChange tests
// =============================================================================

func TestExecuteRecursivePermissionChange_Success(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create some test files
	testFile := filepath.Join(testDir, "file1.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(testDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := createTestModelForPermission(t, tempDir)

	cmd := m.executeRecursivePermissionChange(testDir, "755", "644")
	if cmd == nil {
		t.Fatal("executeRecursivePermissionChange should return a command")
	}

	msg := cmd()
	completeMsg, ok := msg.(recursivePermissionCompleteMsg)
	if !ok {
		t.Fatalf("Command should return recursivePermissionCompleteMsg, got %T", msg)
	}

	if completeMsg.successCount == 0 {
		t.Error("At least some files should be processed successfully")
	}
}

func TestExecuteRecursivePermissionChange_InvalidDirMode(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := createTestModelForPermission(t, tempDir)

	cmd := m.executeRecursivePermissionChange(testDir, "invalid", "644")
	if cmd == nil {
		t.Fatal("executeRecursivePermissionChange should return a command")
	}

	msg := cmd()
	completeMsg, ok := msg.(recursivePermissionCompleteMsg)
	if !ok {
		t.Fatalf("Command should return recursivePermissionCompleteMsg, got %T", msg)
	}

	if completeMsg.successCount != 0 {
		t.Error("No files should succeed with invalid mode")
	}

	if len(completeMsg.errors) == 0 {
		t.Error("Should have errors for invalid mode")
	}
}

func TestExecuteRecursivePermissionChange_InvalidFileMode(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := createTestModelForPermission(t, tempDir)

	cmd := m.executeRecursivePermissionChange(testDir, "755", "invalid")
	if cmd == nil {
		t.Fatal("executeRecursivePermissionChange should return a command")
	}

	msg := cmd()
	completeMsg, ok := msg.(recursivePermissionCompleteMsg)
	if !ok {
		t.Fatalf("Command should return recursivePermissionCompleteMsg, got %T", msg)
	}

	if completeMsg.successCount != 0 {
		t.Error("No files should succeed with invalid mode")
	}

	if len(completeMsg.errors) == 0 {
		t.Error("Should have errors for invalid mode")
	}
}

// =============================================================================
// Phase 7: executeBatchPermissionChange tests
// =============================================================================

func TestExecuteBatchPermissionChange_SmallBatch_Success(t *testing.T) {
	tempDir := t.TempDir()
	testFiles := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "file2.txt"),
		filepath.Join(tempDir, "file3.txt"),
	}

	for _, f := range testFiles {
		if err := os.WriteFile(f, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	m := createTestModelForPermission(t, tempDir)

	cmd := m.executeBatchPermissionChange(testFiles, "755")
	if cmd == nil {
		t.Fatal("executeBatchPermissionChange should return a command")
	}

	msg := cmd()
	completeMsg, ok := msg.(batchPermissionCompleteMsg)
	if !ok {
		t.Fatalf("Command should return batchPermissionCompleteMsg, got %T", msg)
	}

	if completeMsg.successCount != 3 {
		t.Errorf("successCount = %d, want 3", completeMsg.successCount)
	}

	if completeMsg.failedCount != 0 {
		t.Errorf("failedCount = %d, want 0", completeMsg.failedCount)
	}
}

func TestExecuteBatchPermissionChange_SmallBatch_InvalidMode(t *testing.T) {
	tempDir := t.TempDir()
	testFiles := []string{
		filepath.Join(tempDir, "file1.txt"),
	}

	if err := os.WriteFile(testFiles[0], []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	m := createTestModelForPermission(t, tempDir)

	cmd := m.executeBatchPermissionChange(testFiles, "invalid")
	if cmd == nil {
		t.Fatal("executeBatchPermissionChange should return a command")
	}

	msg := cmd()
	completeMsg, ok := msg.(batchPermissionCompleteMsg)
	if !ok {
		t.Fatalf("Command should return batchPermissionCompleteMsg, got %T", msg)
	}

	if completeMsg.successCount != 0 {
		t.Error("No files should succeed with invalid mode")
	}

	if len(completeMsg.errors) == 0 {
		t.Error("Should have errors for invalid mode")
	}
}

func TestExecuteBatchPermissionChange_SkipsSymlinks(t *testing.T) {
	tempDir := t.TempDir()

	// Create a real file
	realFile := filepath.Join(tempDir, "real.txt")
	if err := os.WriteFile(realFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a symlink
	symlink := filepath.Join(tempDir, "link.txt")
	if err := os.Symlink(realFile, symlink); err != nil {
		t.Skip("Cannot create symlink, skipping test")
	}

	m := createTestModelForPermission(t, tempDir)

	paths := []string{realFile, symlink}
	cmd := m.executeBatchPermissionChange(paths, "755")
	if cmd == nil {
		t.Fatal("executeBatchPermissionChange should return a command")
	}

	msg := cmd()
	completeMsg, ok := msg.(batchPermissionCompleteMsg)
	if !ok {
		t.Fatalf("Command should return batchPermissionCompleteMsg, got %T", msg)
	}

	// Only the real file should be processed, symlink should be skipped
	if completeMsg.successCount != 1 {
		t.Errorf("successCount = %d, want 1 (symlink should be skipped)", completeMsg.successCount)
	}
}

// =============================================================================
// Phase 8: Message type tests
// =============================================================================

func TestPermissionMessageTypes(t *testing.T) {
	t.Run("permissionOperationCompleteMsg", func(t *testing.T) {
		msg := permissionOperationCompleteMsg{
			path:    "/test/path",
			success: true,
			err:     nil,
		}

		if msg.path != "/test/path" {
			t.Errorf("path = %s, want /test/path", msg.path)
		}
		if !msg.success {
			t.Error("success should be true")
		}
	})

	t.Run("recursivePermissionCompleteMsg", func(t *testing.T) {
		msg := recursivePermissionCompleteMsg{
			path:         "/test/dir",
			successCount: 10,
			errors:       nil,
		}

		if msg.path != "/test/dir" {
			t.Errorf("path = %s, want /test/dir", msg.path)
		}
		if msg.successCount != 10 {
			t.Errorf("successCount = %d, want 10", msg.successCount)
		}
	})

	t.Run("batchPermissionCompleteMsg", func(t *testing.T) {
		msg := batchPermissionCompleteMsg{
			totalCount:   5,
			successCount: 3,
			failedCount:  2,
			errors:       nil,
		}

		if msg.totalCount != 5 {
			t.Errorf("totalCount = %d, want 5", msg.totalCount)
		}
		if msg.successCount != 3 {
			t.Errorf("successCount = %d, want 3", msg.successCount)
		}
		if msg.failedCount != 2 {
			t.Errorf("failedCount = %d, want 2", msg.failedCount)
		}
	})

	t.Run("batchPermissionStartMsg", func(t *testing.T) {
		msg := batchPermissionStartMsg{
			paths: []string{"/path1", "/path2"},
			mode:  "755",
		}

		if len(msg.paths) != 2 {
			t.Errorf("paths length = %d, want 2", len(msg.paths))
		}
		if msg.mode != "755" {
			t.Errorf("mode = %s, want 755", msg.mode)
		}
	})

	t.Run("batchPermissionProgressMsg", func(t *testing.T) {
		msg := batchPermissionProgressMsg{
			processed:   5,
			total:       10,
			currentPath: "/test/file.txt",
		}

		if msg.processed != 5 {
			t.Errorf("processed = %d, want 5", msg.processed)
		}
		if msg.total != 10 {
			t.Errorf("total = %d, want 10", msg.total)
		}
		if msg.currentPath != "/test/file.txt" {
			t.Errorf("currentPath = %s, want /test/file.txt", msg.currentPath)
		}
	})

	t.Run("showRecursivePermDialogMsg", func(t *testing.T) {
		msg := showRecursivePermDialogMsg{
			path: "/test/dir",
		}

		if msg.path != "/test/dir" {
			t.Errorf("path = %s, want /test/dir", msg.path)
		}
	})

	t.Run("permissionDialogCancelMsg", func(t *testing.T) {
		msg := permissionDialogCancelMsg{}
		_ = msg // Just verify type exists
	})

	t.Run("recursivePermDialogCancelMsg", func(t *testing.T) {
		msg := recursivePermDialogCancelMsg{}
		_ = msg // Just verify type exists
	})
}
