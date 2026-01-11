package ui

import (
	"testing"

	"github.com/sakura/duofm/internal/archive"
)

func TestNewArchiveOperationManager(t *testing.T) {
	manager := NewArchiveOperationManager()

	if manager == nil {
		t.Fatal("NewArchiveOperationManager() returned nil")
	}

	if manager.IsActive() {
		t.Error("NewArchiveOperationManager() should not be active initially")
	}

	if manager.State() != nil {
		t.Error("NewArchiveOperationManager() state should be nil initially")
	}

	if manager.TaskID() != "" {
		t.Error("NewArchiveOperationManager() TaskID should be empty initially")
	}
}

func TestArchiveOperationManager_IsActive(t *testing.T) {
	manager := NewArchiveOperationManager()

	// Initially not active
	if manager.IsActive() {
		t.Error("IsActive() should return false initially")
	}

	// Manually set state to test IsActive
	manager.state = &ArchiveOperationState{
		Sources: []string{"/test/file.txt"},
	}

	if !manager.IsActive() {
		t.Error("IsActive() should return true when state is set")
	}
}

func TestArchiveOperationManager_State(t *testing.T) {
	manager := NewArchiveOperationManager()

	if manager.State() != nil {
		t.Error("State() should return nil when not active")
	}

	expectedState := &ArchiveOperationState{
		Sources:     []string{"/test/file.txt"},
		DestDir:     "/output",
		Format:      archive.FormatZip,
		Level:       5,
		ArchiveName: "test.zip",
	}
	manager.state = expectedState

	state := manager.State()
	if state != expectedState {
		t.Error("State() should return the current state")
	}

	if len(state.Sources) != 1 || state.Sources[0] != "/test/file.txt" {
		t.Errorf("State() sources = %v, want [/test/file.txt]", state.Sources)
	}

	if state.DestDir != "/output" {
		t.Errorf("State() DestDir = %s, want /output", state.DestDir)
	}

	if state.Format != archive.FormatZip {
		t.Errorf("State() Format = %v, want FormatZip", state.Format)
	}

	if state.Level != 5 {
		t.Errorf("State() Level = %d, want 5", state.Level)
	}

	if state.ArchiveName != "test.zip" {
		t.Errorf("State() ArchiveName = %s, want test.zip", state.ArchiveName)
	}
}

func TestArchiveOperationManager_TaskID(t *testing.T) {
	manager := NewArchiveOperationManager()

	// TaskID when state is nil
	if manager.TaskID() != "" {
		t.Error("TaskID() should return empty string when state is nil")
	}

	// Set state without TaskID
	manager.state = &ArchiveOperationState{
		Sources: []string{"/test/file.txt"},
	}

	if manager.TaskID() != "" {
		t.Error("TaskID() should return empty string when TaskID not set")
	}

	// Set TaskID
	manager.state.TaskID = "test-task-123"

	if manager.TaskID() != "test-task-123" {
		t.Errorf("TaskID() = %s, want test-task-123", manager.TaskID())
	}
}

func TestArchiveOperationManager_SetTaskID(t *testing.T) {
	manager := NewArchiveOperationManager()

	// SetTaskID when state is nil (should not panic)
	manager.SetTaskID("test-task-123")

	if manager.TaskID() != "" {
		t.Error("SetTaskID() should not set TaskID when state is nil")
	}

	// Set state and then TaskID
	manager.state = &ArchiveOperationState{
		Sources: []string{"/test/file.txt"},
	}
	manager.SetTaskID("test-task-456")

	if manager.TaskID() != "test-task-456" {
		t.Errorf("TaskID() = %s, want test-task-456", manager.TaskID())
	}

	// Update TaskID
	manager.SetTaskID("test-task-789")

	if manager.TaskID() != "test-task-789" {
		t.Errorf("TaskID() = %s, want test-task-789", manager.TaskID())
	}
}

func TestArchiveOperationManager_PrepareCompression(t *testing.T) {
	manager := NewArchiveOperationManager()

	sources := []string{"/test/file1.txt", "/test/file2.txt"}
	destDir := "/output"
	format := archive.FormatTarGz
	level := 6
	archiveName := "archive.tar.gz"

	cmd := manager.PrepareCompression(sources, destDir, format, level, archiveName)

	if cmd == nil {
		t.Fatal("PrepareCompression() returned nil command")
	}

	// Verify state is set
	if manager.state == nil {
		t.Fatal("PrepareCompression() should set state")
	}

	if len(manager.state.Sources) != 2 {
		t.Errorf("state.Sources length = %d, want 2", len(manager.state.Sources))
	}

	if manager.state.DestDir != destDir {
		t.Errorf("state.DestDir = %s, want %s", manager.state.DestDir, destDir)
	}

	if manager.state.Format != format {
		t.Errorf("state.Format = %v, want %v", manager.state.Format, format)
	}

	if manager.state.Level != level {
		t.Errorf("state.Level = %d, want %d", manager.state.Level, level)
	}

	if manager.state.ArchiveName != archiveName {
		t.Errorf("state.ArchiveName = %s, want %s", manager.state.ArchiveName, archiveName)
	}

	// Execute command and verify message
	msg := cmd()
	progressMsg, ok := msg.(showArchiveProgressMsg)
	if !ok {
		t.Fatalf("PrepareCompression() command returned %T, want showArchiveProgressMsg", msg)
	}

	if progressMsg.operation != "compress" {
		t.Errorf("message operation = %s, want compress", progressMsg.operation)
	}

	if progressMsg.archivePath != archiveName {
		t.Errorf("message archivePath = %s, want %s", progressMsg.archivePath, archiveName)
	}
}

func TestArchiveOperationManager_StartCompression_NilState(t *testing.T) {
	manager := NewArchiveOperationManager()

	// Should return nil when state is nil
	cmd := manager.StartCompression("/output/archive.zip")

	if cmd != nil {
		t.Error("StartCompression() should return nil when state is nil")
	}
}

func TestArchiveOperationManager_StartCompression_NilController(t *testing.T) {
	manager := NewArchiveOperationManager()
	manager.state = &ArchiveOperationState{
		Sources: []string{"/test/file.txt"},
	}
	manager.controller = nil

	// Should return nil when controller is nil
	cmd := manager.StartCompression("/output/archive.zip")

	if cmd != nil {
		t.Error("StartCompression() should return nil when controller is nil")
	}
}

func TestArchiveOperationManager_PrepareExtraction(t *testing.T) {
	manager := NewArchiveOperationManager()

	archivePath := "/test/archive.zip"
	destDir := "/output"

	cmd := manager.PrepareExtraction(archivePath, destDir)

	if cmd == nil {
		t.Fatal("PrepareExtraction() returned nil command")
	}

	// Verify state is set
	if manager.state == nil {
		t.Fatal("PrepareExtraction() should set state")
	}

	if len(manager.state.Sources) != 1 || manager.state.Sources[0] != archivePath {
		t.Errorf("state.Sources = %v, want [%s]", manager.state.Sources, archivePath)
	}

	if manager.state.DestDir != destDir {
		t.Errorf("state.DestDir = %s, want %s", manager.state.DestDir, destDir)
	}

	// Execute command and verify message
	msg := cmd()
	progressMsg, ok := msg.(showArchiveProgressMsg)
	if !ok {
		t.Fatalf("PrepareExtraction() command returned %T, want showArchiveProgressMsg", msg)
	}

	if progressMsg.operation != "extract" {
		t.Errorf("message operation = %s, want extract", progressMsg.operation)
	}

	if progressMsg.archivePath != archivePath {
		t.Errorf("message archivePath = %s, want %s", progressMsg.archivePath, archivePath)
	}
}

func TestArchiveOperationManager_StartExtraction_NilController(t *testing.T) {
	manager := NewArchiveOperationManager()
	manager.controller = nil

	// Should return nil when controller is nil
	cmd := manager.StartExtraction("/test/archive.zip", "/output")

	if cmd != nil {
		t.Error("StartExtraction() should return nil when controller is nil")
	}
}

func TestArchiveOperationManager_CheckSecurity_NilController(t *testing.T) {
	manager := NewArchiveOperationManager()
	manager.controller = nil

	// Should return nil when controller is nil
	cmd := manager.CheckSecurity("/test/archive.zip", "/output")

	if cmd != nil {
		t.Error("CheckSecurity() should return nil when controller is nil")
	}
}

func TestArchiveOperationManager_PollProgress_NilController(t *testing.T) {
	manager := NewArchiveOperationManager()
	manager.controller = nil

	// Should return nil when controller is nil
	cmd := manager.PollProgress("test-task-123")

	if cmd != nil {
		t.Error("PollProgress() should return nil when controller is nil")
	}
}

func TestArchiveOperationManager_CancelTask(t *testing.T) {
	manager := NewArchiveOperationManager()

	// CancelTask when state is nil (should not panic)
	manager.CancelTask()

	// CancelTask when TaskID is empty
	manager.state = &ArchiveOperationState{}
	manager.CancelTask()

	// CancelTask when controller is nil
	manager.state.TaskID = "test-task-123"
	manager.controller = nil
	manager.CancelTask()

	// All calls should complete without panic
}

func TestArchiveOperationManager_Clear(t *testing.T) {
	manager := NewArchiveOperationManager()

	// Set state
	manager.state = &ArchiveOperationState{
		Sources:     []string{"/test/file.txt"},
		DestDir:     "/output",
		Format:      archive.FormatZip,
		Level:       5,
		ArchiveName: "test.zip",
		TaskID:      "test-task-123",
	}

	if !manager.IsActive() {
		t.Fatal("Manager should be active before Clear()")
	}

	// Clear state
	manager.Clear()

	if manager.IsActive() {
		t.Error("Clear() should deactivate manager")
	}

	if manager.State() != nil {
		t.Error("Clear() should set state to nil")
	}

	if manager.TaskID() != "" {
		t.Error("Clear() should clear TaskID")
	}
}

func TestArchiveOperationState_Fields(t *testing.T) {
	state := ArchiveOperationState{
		Sources:     []string{"/test/file1.txt", "/test/file2.txt"},
		DestDir:     "/output",
		Format:      archive.FormatTarGz,
		Level:       9,
		ArchiveName: "backup.tar.gz",
		TaskID:      "task-001",
	}

	if len(state.Sources) != 2 {
		t.Errorf("Sources length = %d, want 2", len(state.Sources))
	}

	if state.Sources[0] != "/test/file1.txt" {
		t.Errorf("Sources[0] = %s, want /test/file1.txt", state.Sources[0])
	}

	if state.DestDir != "/output" {
		t.Errorf("DestDir = %s, want /output", state.DestDir)
	}

	if state.Format != archive.FormatTarGz {
		t.Errorf("Format = %v, want FormatTarGz", state.Format)
	}

	if state.Level != 9 {
		t.Errorf("Level = %d, want 9", state.Level)
	}

	if state.ArchiveName != "backup.tar.gz" {
		t.Errorf("ArchiveName = %s, want backup.tar.gz", state.ArchiveName)
	}

	if state.TaskID != "task-001" {
		t.Errorf("TaskID = %s, want task-001", state.TaskID)
	}
}

// Test message types
func TestArchiveMessages(t *testing.T) {
	t.Run("showArchiveProgressMsg", func(t *testing.T) {
		msg := showArchiveProgressMsg{
			operation:   "compress",
			archivePath: "/test/archive.zip",
		}

		if msg.operation != "compress" {
			t.Errorf("operation = %s, want compress", msg.operation)
		}

		if msg.archivePath != "/test/archive.zip" {
			t.Errorf("archivePath = %s, want /test/archive.zip", msg.archivePath)
		}
	})

	t.Run("archiveOperationStartMsg", func(t *testing.T) {
		msg := archiveOperationStartMsg{
			taskID: "task-123",
		}

		if msg.taskID != "task-123" {
			t.Errorf("taskID = %s, want task-123", msg.taskID)
		}
	})

	t.Run("archiveOperationErrorMsg", func(t *testing.T) {
		testErr := &testError{"test error"}
		msg := archiveOperationErrorMsg{
			err:     testErr,
			message: "Failed to compress",
		}

		if msg.err != testErr {
			t.Error("err mismatch")
		}

		if msg.message != "Failed to compress" {
			t.Errorf("message = %s, want Failed to compress", msg.message)
		}
	})

	t.Run("archiveProgressUpdateMsg", func(t *testing.T) {
		msg := archiveProgressUpdateMsg{
			taskID:         "task-123",
			progress:       0.5,
			processedFiles: 5,
			totalFiles:     10,
			currentFile:    "file5.txt",
		}

		if msg.taskID != "task-123" {
			t.Errorf("taskID = %s, want task-123", msg.taskID)
		}

		if msg.progress != 0.5 {
			t.Errorf("progress = %f, want 0.5", msg.progress)
		}

		if msg.processedFiles != 5 {
			t.Errorf("processedFiles = %d, want 5", msg.processedFiles)
		}

		if msg.totalFiles != 10 {
			t.Errorf("totalFiles = %d, want 10", msg.totalFiles)
		}

		if msg.currentFile != "file5.txt" {
			t.Errorf("currentFile = %s, want file5.txt", msg.currentFile)
		}
	})

	t.Run("archiveOperationCompleteMsg", func(t *testing.T) {
		msg := archiveOperationCompleteMsg{
			taskID:      "task-123",
			success:     true,
			cancelled:   false,
			archivePath: "/output/archive.zip",
		}

		if msg.taskID != "task-123" {
			t.Errorf("taskID = %s, want task-123", msg.taskID)
		}

		if !msg.success {
			t.Error("success should be true")
		}

		if msg.cancelled {
			t.Error("cancelled should be false")
		}

		if msg.archivePath != "/output/archive.zip" {
			t.Errorf("archivePath = %s, want /output/archive.zip", msg.archivePath)
		}
	})

	t.Run("extractSecurityCheckMsg", func(t *testing.T) {
		msg := extractSecurityCheckMsg{
			archivePath:   "/test/archive.zip",
			destDir:       "/output",
			archiveSize:   1000,
			extractedSize: 10000,
			availableSize: 50000,
			compressionOK: true,
			diskSpaceOK:   true,
			ratio:         10.0,
		}

		if msg.archivePath != "/test/archive.zip" {
			t.Errorf("archivePath = %s, want /test/archive.zip", msg.archivePath)
		}

		if msg.destDir != "/output" {
			t.Errorf("destDir = %s, want /output", msg.destDir)
		}

		if msg.archiveSize != 1000 {
			t.Errorf("archiveSize = %d, want 1000", msg.archiveSize)
		}

		if msg.extractedSize != 10000 {
			t.Errorf("extractedSize = %d, want 10000", msg.extractedSize)
		}

		if msg.availableSize != 50000 {
			t.Errorf("availableSize = %d, want 50000", msg.availableSize)
		}

		if !msg.compressionOK {
			t.Error("compressionOK should be true")
		}

		if !msg.diskSpaceOK {
			t.Error("diskSpaceOK should be true")
		}

		if msg.ratio != 10.0 {
			t.Errorf("ratio = %f, want 10.0", msg.ratio)
		}
	})
}

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
