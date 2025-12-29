package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sakura/duofm/internal/fs"
)

// setupTestPane creates a pane with test entries for mark testing
func setupTestPane(t *testing.T) (*Pane, string) {
	t.Helper()

	// Create a temporary directory with test files
	tmpDir, err := os.MkdirTemp("", "duofm-mark-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test files with different sizes
	testFiles := []struct {
		name  string
		size  int
		isDir bool
	}{
		{"file1.txt", 100, false},
		{"file2.txt", 200, false},
		{"file3.txt", 300, false},
		{"subdir", 0, true},
		{".hidden", 50, false},
	}

	for _, tf := range testFiles {
		path := filepath.Join(tmpDir, tf.name)
		if tf.isDir {
			if err := os.Mkdir(path, 0755); err != nil {
				t.Fatalf("Failed to create test dir: %v", err)
			}
		} else {
			if err := os.WriteFile(path, make([]byte, tf.size), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}
	}

	pane, err := NewPane(tmpDir, 80, 24, true, nil)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create pane: %v", err)
	}

	return pane, tmpDir
}

func TestToggleMark(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	// Move cursor past parent directory (first entry is ..)
	pane.cursor = 1
	entry := pane.SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		t.Fatalf("Expected non-parent entry at cursor 1")
	}

	// Toggle mark on unmarked file
	result := pane.ToggleMark()
	if !result {
		t.Error("ToggleMark should return true for regular file")
	}

	if !pane.IsMarked(entry.Name) {
		t.Errorf("File %s should be marked after toggle", entry.Name)
	}

	// Toggle mark on already marked file (unmark)
	result = pane.ToggleMark()
	if !result {
		t.Error("ToggleMark should return true when unmarking")
	}

	if pane.IsMarked(entry.Name) {
		t.Errorf("File %s should be unmarked after second toggle", entry.Name)
	}
}

func TestToggleMarkOnParentDir(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	// Cursor should be on parent directory (..)
	pane.cursor = 0
	entry := pane.SelectedEntry()
	if entry == nil || !entry.IsParentDir() {
		t.Skip("First entry is not parent directory")
	}

	// ToggleMark should return false for parent directory
	result := pane.ToggleMark()
	if result {
		t.Error("ToggleMark should return false for parent directory")
	}

	// Parent directory should not be marked
	if pane.IsMarked(entry.Name) {
		t.Error("Parent directory should not be marked")
	}
}

func TestClearMarks(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	// Mark multiple files
	for i := 1; i < len(pane.entries) && i <= 3; i++ {
		pane.cursor = i
		pane.ToggleMark()
	}

	markedBefore := pane.GetMarkedFiles()
	if len(markedBefore) == 0 {
		t.Fatal("Expected some files to be marked")
	}

	// Clear all marks
	pane.ClearMarks()

	markedAfter := pane.GetMarkedFiles()
	if len(markedAfter) != 0 {
		t.Errorf("Expected 0 marked files after ClearMarks, got %d", len(markedAfter))
	}
}

func TestIsMarked(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	// Get a regular file entry
	pane.cursor = 1
	entry := pane.SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		t.Fatal("Expected non-parent entry")
	}

	// Initially not marked
	if pane.IsMarked(entry.Name) {
		t.Error("File should not be marked initially")
	}

	// Mark the file
	pane.ToggleMark()

	// Now should be marked
	if !pane.IsMarked(entry.Name) {
		t.Error("File should be marked after toggle")
	}

	// Check non-existent file
	if pane.IsMarked("nonexistent.txt") {
		t.Error("Non-existent file should not be marked")
	}
}

func TestGetMarkedFiles(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	// Mark specific files
	var markedNames []string
	for i := 1; i < len(pane.entries) && i <= 2; i++ {
		pane.cursor = i
		entry := pane.SelectedEntry()
		if entry != nil && !entry.IsParentDir() {
			pane.ToggleMark()
			markedNames = append(markedNames, entry.Name)
		}
	}

	markedFiles := pane.GetMarkedFiles()
	if len(markedFiles) != len(markedNames) {
		t.Errorf("Expected %d marked files, got %d", len(markedNames), len(markedFiles))
	}

	// Verify all marked names are in the result
	for _, name := range markedNames {
		found := false
		for _, marked := range markedFiles {
			if marked == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Marked file %s not found in GetMarkedFiles result", name)
		}
	}
}

func TestCalculateMarkInfo(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	// Initially no marks
	info := pane.CalculateMarkInfo()
	if info.Count != 0 {
		t.Errorf("Expected Count=0, got %d", info.Count)
	}
	if info.TotalSize != 0 {
		t.Errorf("Expected TotalSize=0, got %d", info.TotalSize)
	}

	// Mark files with known sizes (100 + 200 = 300)
	var totalExpectedSize int64 = 0
	markedCount := 0
	for i := 1; i < len(pane.entries); i++ {
		pane.cursor = i
		entry := pane.SelectedEntry()
		if entry != nil && !entry.IsParentDir() && !entry.IsDir {
			pane.ToggleMark()
			totalExpectedSize += entry.Size
			markedCount++
			if markedCount >= 2 {
				break
			}
		}
	}

	info = pane.CalculateMarkInfo()
	if info.Count != markedCount {
		t.Errorf("Expected Count=%d, got %d", markedCount, info.Count)
	}
	if info.TotalSize != totalExpectedSize {
		t.Errorf("Expected TotalSize=%d, got %d", totalExpectedSize, info.TotalSize)
	}
}

func TestCalculateMarkInfoWithDirectory(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	// Find and mark the directory
	var dirEntry *fs.FileEntry
	for i := 1; i < len(pane.entries); i++ {
		entry := &pane.entries[i]
		if entry.IsDir && !entry.IsParentDir() {
			dirEntry = entry
			pane.cursor = i
			pane.ToggleMark()
			break
		}
	}

	if dirEntry == nil {
		t.Skip("No directory found in test entries")
	}

	// Directory should be counted as 0 bytes
	info := pane.CalculateMarkInfo()
	if info.Count != 1 {
		t.Errorf("Expected Count=1, got %d", info.Count)
	}
	if info.TotalSize != 0 {
		t.Errorf("Expected TotalSize=0 for directory, got %d", info.TotalSize)
	}
}

func TestMarksClearedOnDirectoryChange(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	// Mark a file
	pane.cursor = 1
	pane.ToggleMark()

	markedBefore := pane.GetMarkedFiles()
	if len(markedBefore) == 0 {
		t.Fatal("Expected some files to be marked")
	}

	// Change directory (reload)
	err := pane.LoadDirectory()
	if err != nil {
		t.Fatalf("LoadDirectory failed: %v", err)
	}

	// Marks should be cleared
	markedAfter := pane.GetMarkedFiles()
	if len(markedAfter) != 0 {
		t.Errorf("Expected 0 marked files after directory change, got %d", len(markedAfter))
	}
}

func TestGetMarkedFilePaths(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	// Mark a file
	pane.cursor = 1
	entry := pane.SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		t.Fatal("Expected non-parent entry")
	}
	pane.ToggleMark()

	paths := pane.GetMarkedFilePaths()
	if len(paths) != 1 {
		t.Fatalf("Expected 1 marked file path, got %d", len(paths))
	}

	expectedPath := filepath.Join(tmpDir, entry.Name)
	if paths[0] != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, paths[0])
	}
}

func TestMarkCount(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	// Initially 0
	if count := pane.MarkCount(); count != 0 {
		t.Errorf("Expected MarkCount=0, got %d", count)
	}

	// Mark 2 files
	for i := 1; i < len(pane.entries) && i <= 2; i++ {
		pane.cursor = i
		if entry := pane.SelectedEntry(); entry != nil && !entry.IsParentDir() {
			pane.ToggleMark()
		}
	}

	if count := pane.MarkCount(); count != 2 {
		t.Errorf("Expected MarkCount=2, got %d", count)
	}
}

func TestHasMarkedFiles(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	// Initially false
	if pane.HasMarkedFiles() {
		t.Error("Expected HasMarkedFiles=false initially")
	}

	// Mark a file
	pane.cursor = 1
	pane.ToggleMark()

	// Now true
	if !pane.HasMarkedFiles() {
		t.Error("Expected HasMarkedFiles=true after marking")
	}
}
