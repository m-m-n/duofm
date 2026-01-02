package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewArchiveController(t *testing.T) {
	controller := NewArchiveController()
	if controller == nil {
		t.Fatal("NewArchiveController() returned nil")
	}
}

func TestArchiveController_CreateArchive(t *testing.T) {
	if !CheckCommand("tar") {
		t.Skip("tar command not available")
	}

	controller := NewArchiveController()

	// Create temp directory with test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "test.tar")

	taskID, err := controller.CreateArchive([]string{testFile}, outputFile, FormatTar, 6)
	if err != nil {
		t.Errorf("CreateArchive() error = %v", err)
	}

	if taskID == "" {
		t.Error("CreateArchive() returned empty task ID")
	}

	// Wait for task to complete
	controller.WaitForTask(taskID)

	// Verify archive was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("CreateArchive() did not create archive file")
	}
}

func TestArchiveController_ExtractArchive(t *testing.T) {
	if !CheckCommand("tar") {
		t.Skip("tar command not available")
	}

	controller := NewArchiveController()

	// Create temp directory with test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create archive
	archiveFile := filepath.Join(tempDir, "test.tar")
	taskID1, err := controller.CreateArchive([]string{testFile}, archiveFile, FormatTar, 6)
	if err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}
	controller.WaitForTask(taskID1)

	// Extract to new directory
	extractDir := filepath.Join(tempDir, "extract")
	taskID2, err := controller.ExtractArchive(archiveFile, extractDir)
	if err != nil {
		t.Errorf("ExtractArchive() error = %v", err)
	}

	if taskID2 == "" {
		t.Error("ExtractArchive() returned empty task ID")
	}

	controller.WaitForTask(taskID2)

	// Verify extracted file
	extractedFile := filepath.Join(extractDir, "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("ExtractArchive() did not extract file")
	}
}

func TestArchiveController_CancelTask(t *testing.T) {
	controller := NewArchiveController()

	// Test canceling a non-existent task
	err := controller.CancelTask("non-existent-task-id")
	if err == nil {
		t.Error("CancelTask() should return error for non-existent task")
	}
}

func TestArchiveController_GetTaskStatus(t *testing.T) {
	controller := NewArchiveController()

	// Test getting status of non-existent task
	status := controller.GetTaskStatus("non-existent-task-id")
	if status != nil {
		t.Error("GetTaskStatus() should return nil for non-existent task")
	}
}

func TestArchiveController_GetArchiveMetadata(t *testing.T) {
	if !CheckCommand("tar") {
		t.Skip("tar command not available")
	}

	controller := NewArchiveController()

	// Create temp directory with test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("test content for metadata")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create archive
	archiveFile := filepath.Join(tempDir, "test.tar")
	taskID, err := controller.CreateArchive([]string{testFile}, archiveFile, FormatTar, 6)
	if err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}
	controller.WaitForTask(taskID)

	// Get metadata
	metadata, err := controller.GetArchiveMetadata(archiveFile)
	if err != nil {
		t.Errorf("GetArchiveMetadata() error = %v", err)
	}

	if metadata == nil {
		t.Fatal("GetArchiveMetadata() returned nil metadata")
	}

	if metadata.FileCount != 1 {
		t.Errorf("GetArchiveMetadata() FileCount = %d, want 1", metadata.FileCount)
	}
}

func TestArchiveController_GetArchiveMetadata_NonExistent(t *testing.T) {
	controller := NewArchiveController()

	_, err := controller.GetArchiveMetadata("/nonexistent/file.tar")
	if err == nil {
		t.Error("GetArchiveMetadata() should return error for non-existent file")
	}
}

func TestArchiveController_CreateArchive_EmptySources(t *testing.T) {
	controller := NewArchiveController()

	_, err := controller.CreateArchive([]string{}, "/tmp/test.tar", FormatTar, 6)
	if err == nil {
		t.Error("CreateArchive() should return error for empty sources")
	}
}

func TestArchiveController_CreateArchive_NonExistentSource(t *testing.T) {
	controller := NewArchiveController()

	_, err := controller.CreateArchive([]string{"/nonexistent/file.txt"}, "/tmp/test.tar", FormatTar, 6)
	if err == nil {
		t.Error("CreateArchive() should return error for non-existent source")
	}
}

func TestArchiveController_ExtractArchive_NonExistent(t *testing.T) {
	controller := NewArchiveController()

	_, err := controller.ExtractArchive("/nonexistent/archive.tar", "/tmp/extract")
	if err == nil {
		t.Error("ExtractArchive() should return error for non-existent archive")
	}
}

func TestArchiveController_TaskStatusDuringExecution(t *testing.T) {
	if !CheckCommand("tar") {
		t.Skip("tar command not available")
	}

	controller := NewArchiveController()

	// Create temp directory with test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "test.tar")

	taskID, err := controller.CreateArchive([]string{testFile}, outputFile, FormatTar, 6)
	if err != nil {
		t.Errorf("CreateArchive() error = %v", err)
	}

	// Get status immediately after starting
	status := controller.GetTaskStatus(taskID)
	if status == nil {
		t.Error("GetTaskStatus() returned nil for running task")
	}

	// Wait for completion
	controller.WaitForTask(taskID)

	// Get status after completion
	status = controller.GetTaskStatus(taskID)
	if status == nil {
		t.Error("GetTaskStatus() returned nil after task completion")
	}
}

func TestCalculateTotalSize(t *testing.T) {
	// Create temp directory with test files
	tempDir := t.TempDir()
	testFile1 := filepath.Join(tempDir, "file1.txt")
	testFile2 := filepath.Join(tempDir, "file2.txt")

	content1 := []byte("content1")
	content2 := []byte("content2")

	if err := os.WriteFile(testFile1, content1, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(testFile2, content2, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	size := calculateTotalSize(tempDir)
	expectedSize := int64(len(content1) + len(content2))
	if size != expectedSize {
		t.Errorf("calculateTotalSize() = %d, want %d", size, expectedSize)
	}
}

func TestCalculateTotalSize_SingleFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("test content")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	size := calculateTotalSize(testFile)
	if size != int64(len(content)) {
		t.Errorf("calculateTotalSize() = %d, want %d", size, len(content))
	}
}

func TestCalculateTotalSize_NonExistent(t *testing.T) {
	size := calculateTotalSize("/nonexistent/path")
	if size != 0 {
		t.Errorf("calculateTotalSize() = %d, want 0 for non-existent path", size)
	}
}
