package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewBatchOperationManager(t *testing.T) {
	manager := NewBatchOperationManager()

	if manager == nil {
		t.Fatal("NewBatchOperationManager() returned nil")
	}

	if manager.IsActive() {
		t.Error("NewBatchOperationManager() should not be active initially")
	}

	if manager.Current() != nil {
		t.Error("NewBatchOperationManager() current should be nil initially")
	}
}

func TestBatchOperationManager_IsActive(t *testing.T) {
	manager := NewBatchOperationManager()

	// Initially not active
	if manager.IsActive() {
		t.Error("IsActive() should return false initially")
	}

	// Manually set current to test IsActive
	manager.current = &BatchOperation{
		Files:     []string{"/test/file.txt"},
		Operation: "copy",
	}

	if !manager.IsActive() {
		t.Error("IsActive() should return true when current is set")
	}
}

func TestBatchOperationManager_Current(t *testing.T) {
	manager := NewBatchOperationManager()

	if manager.Current() != nil {
		t.Error("Current() should return nil when not active")
	}

	expectedOp := &BatchOperation{
		Files:      []string{"/test/file1.txt", "/test/file2.txt"},
		CurrentIdx: 0,
		DestPath:   "/output",
		Operation:  "copy",
	}
	manager.current = expectedOp

	op := manager.Current()
	if op != expectedOp {
		t.Error("Current() should return the current operation")
	}

	if len(op.Files) != 2 {
		t.Errorf("Current() Files length = %d, want 2", len(op.Files))
	}

	if op.DestPath != "/output" {
		t.Errorf("Current() DestPath = %s, want /output", op.DestPath)
	}

	if op.Operation != "copy" {
		t.Errorf("Current() Operation = %s, want copy", op.Operation)
	}
}

func TestBatchOperationManager_Start(t *testing.T) {
	manager := NewBatchOperationManager()

	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	srcDir := "/source"
	destDir := "/dest"
	operation := "copy"

	cmd := manager.Start(files, srcDir, destDir, operation)

	if cmd == nil {
		t.Fatal("Start() returned nil command")
	}

	// Verify current is set
	if manager.current == nil {
		t.Fatal("Start() should set current")
	}

	// Verify files have full paths
	if len(manager.current.Files) != 3 {
		t.Errorf("current.Files length = %d, want 3", len(manager.current.Files))
	}

	expectedPath := filepath.Join(srcDir, "file1.txt")
	if manager.current.Files[0] != expectedPath {
		t.Errorf("current.Files[0] = %s, want %s", manager.current.Files[0], expectedPath)
	}

	if manager.current.CurrentIdx != 0 {
		t.Errorf("current.CurrentIdx = %d, want 0", manager.current.CurrentIdx)
	}

	if manager.current.DestPath != destDir {
		t.Errorf("current.DestPath = %s, want %s", manager.current.DestPath, destDir)
	}

	if manager.current.Operation != operation {
		t.Errorf("current.Operation = %s, want %s", manager.current.Operation, operation)
	}

	if len(manager.current.Completed) != 0 {
		t.Errorf("current.Completed should be empty initially")
	}

	if len(manager.current.Failed) != 0 {
		t.Errorf("current.Failed should be empty initially")
	}

	// Execute command and verify message
	msg := cmd()
	if _, ok := msg.(batchStartedMsg); !ok {
		t.Fatalf("Start() command returned %T, want batchStartedMsg", msg)
	}
}

func TestBatchOperationManager_CurrentFile(t *testing.T) {
	manager := NewBatchOperationManager()

	// CurrentFile when current is nil
	if manager.CurrentFile() != "" {
		t.Error("CurrentFile() should return empty string when current is nil")
	}

	// Set up batch operation
	manager.current = &BatchOperation{
		Files:      []string{"/test/file1.txt", "/test/file2.txt", "/test/file3.txt"},
		CurrentIdx: 0,
	}

	if manager.CurrentFile() != "/test/file1.txt" {
		t.Errorf("CurrentFile() = %s, want /test/file1.txt", manager.CurrentFile())
	}

	// Move to next file
	manager.current.CurrentIdx = 1
	if manager.CurrentFile() != "/test/file2.txt" {
		t.Errorf("CurrentFile() = %s, want /test/file2.txt", manager.CurrentFile())
	}

	// Move to last file
	manager.current.CurrentIdx = 2
	if manager.CurrentFile() != "/test/file3.txt" {
		t.Errorf("CurrentFile() = %s, want /test/file3.txt", manager.CurrentFile())
	}

	// Move past last file
	manager.current.CurrentIdx = 3
	if manager.CurrentFile() != "" {
		t.Errorf("CurrentFile() should return empty string when past last file")
	}
}

func TestBatchOperationManager_DestPath(t *testing.T) {
	manager := NewBatchOperationManager()

	// DestPath when current is nil
	if manager.DestPath() != "" {
		t.Error("DestPath() should return empty string when current is nil")
	}

	manager.current = &BatchOperation{
		DestPath: "/destination",
	}

	if manager.DestPath() != "/destination" {
		t.Errorf("DestPath() = %s, want /destination", manager.DestPath())
	}
}

func TestBatchOperationManager_Operation(t *testing.T) {
	manager := NewBatchOperationManager()

	// Operation when current is nil
	if manager.Operation() != "" {
		t.Error("Operation() should return empty string when current is nil")
	}

	manager.current = &BatchOperation{
		Operation: "move",
	}

	if manager.Operation() != "move" {
		t.Errorf("Operation() = %s, want move", manager.Operation())
	}

	// Test copy operation
	manager.current.Operation = "copy"
	if manager.Operation() != "copy" {
		t.Errorf("Operation() = %s, want copy", manager.Operation())
	}
}

func TestBatchOperationManager_Advance_Success(t *testing.T) {
	manager := NewBatchOperationManager()

	manager.current = &BatchOperation{
		Files:      []string{"/test/file1.txt", "/test/file2.txt", "/test/file3.txt"},
		CurrentIdx: 0,
		DestPath:   "/dest",
		Operation:  "copy",
		Completed:  []string{},
		Failed:     []string{},
	}

	// Advance with success
	cmd := manager.Advance(true, "/test/file1.txt")

	if cmd == nil {
		t.Fatal("Advance() returned nil command")
	}

	// Verify file was added to Completed
	if len(manager.current.Completed) != 1 {
		t.Errorf("Completed length = %d, want 1", len(manager.current.Completed))
	}

	if manager.current.Completed[0] != "/test/file1.txt" {
		t.Errorf("Completed[0] = %s, want /test/file1.txt", manager.current.Completed[0])
	}

	// Verify CurrentIdx was incremented
	if manager.current.CurrentIdx != 1 {
		t.Errorf("CurrentIdx = %d, want 1", manager.current.CurrentIdx)
	}

	// Execute command and verify message
	msg := cmd()
	nextMsg, ok := msg.(batchNextFileMsg)
	if !ok {
		t.Fatalf("Advance() command returned %T, want batchNextFileMsg", msg)
	}

	if nextMsg.srcPath != "/test/file2.txt" {
		t.Errorf("nextMsg.srcPath = %s, want /test/file2.txt", nextMsg.srcPath)
	}

	if nextMsg.destPath != "/dest" {
		t.Errorf("nextMsg.destPath = %s, want /dest", nextMsg.destPath)
	}
}

func TestBatchOperationManager_Advance_Failure(t *testing.T) {
	manager := NewBatchOperationManager()

	manager.current = &BatchOperation{
		Files:      []string{"/test/file1.txt", "/test/file2.txt"},
		CurrentIdx: 0,
		DestPath:   "/dest",
		Operation:  "copy",
		Completed:  []string{},
		Failed:     []string{},
	}

	// Advance with failure
	cmd := manager.Advance(false, "/test/file1.txt")

	if cmd == nil {
		t.Fatal("Advance() returned nil command")
	}

	// Verify file was added to Failed
	if len(manager.current.Failed) != 1 {
		t.Errorf("Failed length = %d, want 1", len(manager.current.Failed))
	}

	if manager.current.Failed[0] != "/test/file1.txt" {
		t.Errorf("Failed[0] = %s, want /test/file1.txt", manager.current.Failed[0])
	}

	// Verify Completed is still empty
	if len(manager.current.Completed) != 0 {
		t.Errorf("Completed length = %d, want 0", len(manager.current.Completed))
	}
}

func TestBatchOperationManager_Advance_LastFile(t *testing.T) {
	manager := NewBatchOperationManager()

	manager.current = &BatchOperation{
		Files:      []string{"/test/file1.txt"},
		CurrentIdx: 0,
		DestPath:   "/dest",
		Operation:  "copy",
		Completed:  []string{},
		Failed:     []string{},
	}

	// Advance with last file
	cmd := manager.Advance(true, "/test/file1.txt")

	if cmd == nil {
		t.Fatal("Advance() returned nil command for last file")
	}

	// Execute command and verify batchCompleteMsg
	msg := cmd()
	completeMsg, ok := msg.(batchCompleteMsg)
	if !ok {
		t.Fatalf("Advance() command returned %T, want batchCompleteMsg", msg)
	}

	if completeMsg.operation != "copy" {
		t.Errorf("completeMsg.operation = %s, want copy", completeMsg.operation)
	}

	if completeMsg.completed != 1 {
		t.Errorf("completeMsg.completed = %d, want 1", completeMsg.completed)
	}

	if completeMsg.failed != 0 {
		t.Errorf("completeMsg.failed = %d, want 0", completeMsg.failed)
	}

	// Verify current is nil after completion
	if manager.current != nil {
		t.Error("current should be nil after batch completion")
	}
}

func TestBatchOperationManager_Advance_NilCurrent(t *testing.T) {
	manager := NewBatchOperationManager()

	// Advance when current is nil
	cmd := manager.Advance(true, "/test/file.txt")

	if cmd != nil {
		t.Error("Advance() should return nil when current is nil")
	}
}

func TestBatchOperationManager_Cancel(t *testing.T) {
	manager := NewBatchOperationManager()

	manager.current = &BatchOperation{
		Files:      []string{"/test/file1.txt", "/test/file2.txt", "/test/file3.txt"},
		CurrentIdx: 1,
		DestPath:   "/dest",
		Operation:  "move",
		Completed:  []string{"/test/file1.txt"},
		Failed:     []string{},
	}

	cmd := manager.Cancel()

	if cmd == nil {
		t.Fatal("Cancel() returned nil command")
	}

	// Verify current is nil after cancel
	if manager.current != nil {
		t.Error("current should be nil after cancel")
	}

	// Execute command and verify message
	msg := cmd()
	cancelMsg, ok := msg.(batchCancelledMsg)
	if !ok {
		t.Fatalf("Cancel() command returned %T, want batchCancelledMsg", msg)
	}

	if cancelMsg.operation != "move" {
		t.Errorf("cancelMsg.operation = %s, want move", cancelMsg.operation)
	}

	if cancelMsg.completed != 1 {
		t.Errorf("cancelMsg.completed = %d, want 1", cancelMsg.completed)
	}

	if cancelMsg.remaining != 2 {
		t.Errorf("cancelMsg.remaining = %d, want 2", cancelMsg.remaining)
	}
}

func TestBatchOperationManager_Cancel_NilCurrent(t *testing.T) {
	manager := NewBatchOperationManager()

	// Cancel when current is nil
	cmd := manager.Cancel()

	if cmd != nil {
		t.Error("Cancel() should return nil when current is nil")
	}
}

func TestBatchOperationManager_ExecuteCurrentFile_NilCurrent(t *testing.T) {
	manager := NewBatchOperationManager()

	// ExecuteCurrentFile when current is nil
	cmd := manager.ExecuteCurrentFile()

	if cmd != nil {
		t.Error("ExecuteCurrentFile() should return nil when current is nil")
	}
}

func TestBatchOperationManager_ExecuteCurrentFile_Copy(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(destDir, 0755)

	srcFile := filepath.Join(srcDir, "test.txt")
	os.WriteFile(srcFile, []byte("test content"), 0644)

	manager := NewBatchOperationManager()
	manager.current = &BatchOperation{
		Files:      []string{srcFile},
		CurrentIdx: 0,
		DestPath:   destDir,
		Operation:  "copy",
		Completed:  []string{},
		Failed:     []string{},
	}

	cmd := manager.ExecuteCurrentFile()

	if cmd == nil {
		t.Fatal("ExecuteCurrentFile() returned nil command")
	}

	// Execute command
	msg := cmd()
	resultMsg, ok := msg.(batchFileResultMsg)
	if !ok {
		t.Fatalf("ExecuteCurrentFile() command returned %T, want batchFileResultMsg", msg)
	}

	if resultMsg.srcPath != srcFile {
		t.Errorf("resultMsg.srcPath = %s, want %s", resultMsg.srcPath, srcFile)
	}

	if !resultMsg.success {
		t.Errorf("resultMsg.success = false, want true: %v", resultMsg.err)
	}

	// Verify file was copied
	destFile := filepath.Join(destDir, "test.txt")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("File was not copied to destination")
	}

	// Verify source file still exists (copy, not move)
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		t.Error("Source file should still exist after copy")
	}
}

func TestBatchOperationManager_ExecuteCurrentFile_Move(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(destDir, 0755)

	srcFile := filepath.Join(srcDir, "test.txt")
	os.WriteFile(srcFile, []byte("test content"), 0644)

	manager := NewBatchOperationManager()
	manager.current = &BatchOperation{
		Files:      []string{srcFile},
		CurrentIdx: 0,
		DestPath:   destDir,
		Operation:  "move",
		Completed:  []string{},
		Failed:     []string{},
	}

	cmd := manager.ExecuteCurrentFile()

	if cmd == nil {
		t.Fatal("ExecuteCurrentFile() returned nil command")
	}

	// Execute command
	msg := cmd()
	resultMsg, ok := msg.(batchFileResultMsg)
	if !ok {
		t.Fatalf("ExecuteCurrentFile() command returned %T, want batchFileResultMsg", msg)
	}

	if !resultMsg.success {
		t.Errorf("resultMsg.success = false, want true: %v", resultMsg.err)
	}

	// Verify file was moved
	destFile := filepath.Join(destDir, "test.txt")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("File was not moved to destination")
	}

	// Verify source file no longer exists (move)
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("Source file should not exist after move")
	}
}

func TestBatchOperationManager_ExecuteCurrentFile_Error(t *testing.T) {
	manager := NewBatchOperationManager()
	manager.current = &BatchOperation{
		Files:      []string{"/nonexistent/path/file.txt"},
		CurrentIdx: 0,
		DestPath:   "/also/nonexistent",
		Operation:  "copy",
		Completed:  []string{},
		Failed:     []string{},
	}

	cmd := manager.ExecuteCurrentFile()

	if cmd == nil {
		t.Fatal("ExecuteCurrentFile() returned nil command")
	}

	// Execute command
	msg := cmd()
	resultMsg, ok := msg.(batchFileResultMsg)
	if !ok {
		t.Fatalf("ExecuteCurrentFile() command returned %T, want batchFileResultMsg", msg)
	}

	if resultMsg.success {
		t.Error("resultMsg.success = true, want false for nonexistent file")
	}

	if resultMsg.err == nil {
		t.Error("resultMsg.err should not be nil for failed operation")
	}
}

func TestBatchOperation_Fields(t *testing.T) {
	op := BatchOperation{
		Files:      []string{"/file1.txt", "/file2.txt"},
		CurrentIdx: 1,
		DestPath:   "/dest",
		Operation:  "copy",
		Completed:  []string{"/file1.txt"},
		Failed:     []string{"/file3.txt"},
	}

	if len(op.Files) != 2 {
		t.Errorf("Files length = %d, want 2", len(op.Files))
	}

	if op.CurrentIdx != 1 {
		t.Errorf("CurrentIdx = %d, want 1", op.CurrentIdx)
	}

	if op.DestPath != "/dest" {
		t.Errorf("DestPath = %s, want /dest", op.DestPath)
	}

	if op.Operation != "copy" {
		t.Errorf("Operation = %s, want copy", op.Operation)
	}

	if len(op.Completed) != 1 || op.Completed[0] != "/file1.txt" {
		t.Errorf("Completed = %v, want [/file1.txt]", op.Completed)
	}

	if len(op.Failed) != 1 || op.Failed[0] != "/file3.txt" {
		t.Errorf("Failed = %v, want [/file3.txt]", op.Failed)
	}
}

// Test message types
func TestBatchMessages(t *testing.T) {
	t.Run("batchStartedMsg", func(t *testing.T) {
		msg := batchStartedMsg{}
		_ = msg // Just verify type exists
	})

	t.Run("batchNextFileMsg", func(t *testing.T) {
		msg := batchNextFileMsg{
			srcPath:  "/source/file.txt",
			destPath: "/dest",
		}

		if msg.srcPath != "/source/file.txt" {
			t.Errorf("srcPath = %s, want /source/file.txt", msg.srcPath)
		}

		if msg.destPath != "/dest" {
			t.Errorf("destPath = %s, want /dest", msg.destPath)
		}
	})

	t.Run("batchCompleteMsg", func(t *testing.T) {
		msg := batchCompleteMsg{
			operation: "move",
			completed: 5,
			failed:    2,
		}

		if msg.operation != "move" {
			t.Errorf("operation = %s, want move", msg.operation)
		}

		if msg.completed != 5 {
			t.Errorf("completed = %d, want 5", msg.completed)
		}

		if msg.failed != 2 {
			t.Errorf("failed = %d, want 2", msg.failed)
		}
	})

	t.Run("batchCancelledMsg", func(t *testing.T) {
		msg := batchCancelledMsg{
			operation: "copy",
			completed: 3,
			remaining: 7,
		}

		if msg.operation != "copy" {
			t.Errorf("operation = %s, want copy", msg.operation)
		}

		if msg.completed != 3 {
			t.Errorf("completed = %d, want 3", msg.completed)
		}

		if msg.remaining != 7 {
			t.Errorf("remaining = %d, want 7", msg.remaining)
		}
	})

	t.Run("batchFileResultMsg", func(t *testing.T) {
		testErr := &testError{"test error"}
		msg := batchFileResultMsg{
			srcPath: "/test/file.txt",
			success: false,
			err:     testErr,
		}

		if msg.srcPath != "/test/file.txt" {
			t.Errorf("srcPath = %s, want /test/file.txt", msg.srcPath)
		}

		if msg.success {
			t.Error("success should be false")
		}

		if msg.err != testErr {
			t.Error("err mismatch")
		}
	})
}
