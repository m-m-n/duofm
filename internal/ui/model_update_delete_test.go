package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestExecuteDeleteOperation_SingleFile_CursorPosition tests cursor positioning after single file deletion
func TestExecuteDeleteOperation_SingleFile_CursorPosition(t *testing.T) {
	tests := []struct {
		name           string
		initialCursor  int
		expectedCursor int
		description    string
	}{
		{
			name:           "Delete middle file",
			initialCursor:  5, // file06.txt (index 5, since .. is index 0)
			expectedCursor: 5, // Should stay at index 5 (now points to file07.txt)
			description:    "Cursor should stay at same index after deleting middle file",
		},
		{
			name:           "Delete last file",
			initialCursor:  10, // file10.txt (last file)
			expectedCursor: 9,  // Should move to new last file (file09.txt)
			description:    "Cursor should move to previous file after deleting last file",
		},
		{
			name:           "Delete first real file",
			initialCursor:  1, // file01.txt (first file after ..)
			expectedCursor: 1, // Should stay at index 1 (now points to file02.txt)
			description:    "Cursor should stay at same index after deleting first real file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh directory and files for each test
			tmpDir := t.TempDir()
			fileCount := 10
			for i := 1; i <= fileCount; i++ {
				filename := fmt.Sprintf("file%02d.txt", i)
				if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			// Create fresh model for each test
			theme := DefaultTheme()
			pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, theme)
			if err != nil {
				t.Fatalf("Failed to create pane: %v", err)
			}

			model := Model{
				leftPane:   pane,
				rightPane:  pane, // Use same pane for simplicity
				activePane: LeftPane,
				theme:      theme,
			}

			// Set initial cursor position
			model.leftPane.cursor = tt.initialCursor

			// Get the filename at cursor position before deletion (check bounds)
			var entryNameBeforeDeletion string
			if tt.initialCursor >= 0 && tt.initialCursor < len(model.leftPane.entries) {
				entryNameBeforeDeletion = model.leftPane.entries[tt.initialCursor].Name
			}

			// Execute delete operation
			model = model.executeDeleteOperation()

			// Verify cursor position
			if model.leftPane.cursor != tt.expectedCursor {
				t.Errorf("%s: Expected cursor at %d, got %d",
					tt.description, tt.expectedCursor, model.leftPane.cursor)
			}

			// Verify cursor is within bounds
			if model.leftPane.cursor < 0 || model.leftPane.cursor >= len(model.leftPane.entries) {
				t.Errorf("Cursor out of bounds: %d (entries: %d)",
					model.leftPane.cursor, len(model.leftPane.entries))
			}

			// Verify the deleted file no longer exists (if we captured the name)
			if entryNameBeforeDeletion != "" && entryNameBeforeDeletion != ".." {
				deletedPath := filepath.Join(tmpDir, entryNameBeforeDeletion)
				if _, err := os.Stat(deletedPath); !os.IsNotExist(err) {
					t.Errorf("File should have been deleted: %s", entryNameBeforeDeletion)
				}
			}
		})
	}
}

// TestExecuteDeleteOperation_MultipleFiles_CursorPosition tests cursor positioning after multiple file deletion
func TestExecuteDeleteOperation_MultipleFiles_CursorPosition(t *testing.T) {
	tests := []struct {
		name           string
		markIndexes    []int
		expectedCursor int
		description    string
	}{
		{
			name:           "Delete 3 consecutive files from middle",
			markIndexes:    []int{4, 5, 6}, // file04, file05, file06
			expectedCursor: 4,              // Cursor should be at position where first marked file was
			description:    "Cursor should be at first marked file position",
		},
		{
			name:           "Delete files at end",
			markIndexes:    []int{8, 9, 10}, // file08, file09, file10 (last 3 files)
			expectedCursor: 7,               // Should be at new last file
			description:    "Cursor should move to new last file when deleting at end",
		},
		{
			name:           "Delete files at beginning",
			markIndexes:    []int{1, 2}, // file01, file02
			expectedCursor: 1,           // Should stay at index 1 (now file03)
			description:    "Cursor should stay at first marked position",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh directory and files for each test
			tmpDir := t.TempDir()
			fileCount := 10
			for i := 1; i <= fileCount; i++ {
				filename := fmt.Sprintf("file%02d.txt", i)
				if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			// Create fresh model for each test
			theme := DefaultTheme()
			pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, theme)
			if err != nil {
				t.Fatalf("Failed to create pane: %v", err)
			}

			model := Model{
				leftPane:   pane,
				rightPane:  pane,
				activePane: LeftPane,
				theme:      theme,
			}

			// Mark files for deletion
			for _, idx := range tt.markIndexes {
				if idx >= 0 && idx < len(model.leftPane.entries) {
					entry := &model.leftPane.entries[idx]
					model.leftPane.markedFiles[entry.Name] = true
				}
			}

			// Execute delete operation
			model = model.executeDeleteOperation()

			// Verify cursor position
			if model.leftPane.cursor != tt.expectedCursor {
				t.Errorf("%s: Expected cursor at %d, got %d",
					tt.description, tt.expectedCursor, model.leftPane.cursor)
			}

			// Verify cursor is within bounds
			if model.leftPane.cursor < 0 || model.leftPane.cursor >= len(model.leftPane.entries) {
				t.Errorf("Cursor out of bounds: %d (entries: %d)",
					model.leftPane.cursor, len(model.leftPane.entries))
			}

			// Verify marks are cleared
			if len(model.leftPane.markedFiles) != 0 {
				t.Errorf("Marked files should be cleared after deletion, got %d marks",
					len(model.leftPane.markedFiles))
			}
		})
	}
}

// TestExecuteDeleteOperation_AllFiles_CursorPosition tests cursor positioning when all files are deleted
func TestExecuteDeleteOperation_AllFiles_CursorPosition(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only 2 test files
	for i := 1; i <= 2; i++ {
		filename := fmt.Sprintf("file%d.txt", i)
		if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	theme := DefaultTheme()
	pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, theme)
	if err != nil {
		t.Fatalf("Failed to create pane: %v", err)
	}

	model := Model{
		leftPane:   pane,
		rightPane:  pane,
		activePane: LeftPane,
		theme:      theme,
	}

	// Mark all files (not parent directory)
	for i := 1; i < len(model.leftPane.entries); i++ {
		entry := &model.leftPane.entries[i]
		if !entry.IsParentDir() {
			model.leftPane.markedFiles[entry.Name] = true
		}
	}

	// Execute delete operation
	model = model.executeDeleteOperation()

	// Verify cursor is at position 0 (parent directory)
	if model.leftPane.cursor != 0 {
		t.Errorf("Expected cursor at 0 (parent directory), got %d", model.leftPane.cursor)
	}

	// Verify only parent directory remains
	if len(model.leftPane.entries) != 1 {
		t.Errorf("Expected only parent directory to remain, got %d entries",
			len(model.leftPane.entries))
	}
}

// TestExecuteDeleteOperation_ParentDirectory_NoOp tests that parent directory cannot be deleted
func TestExecuteDeleteOperation_ParentDirectory_NoOp(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	theme := DefaultTheme()
	pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, theme)
	if err != nil {
		t.Fatalf("Failed to create pane: %v", err)
	}

	model := Model{
		leftPane:   pane,
		rightPane:  pane,
		activePane: LeftPane,
		theme:      theme,
	}

	// Set cursor to parent directory (index 0)
	model.leftPane.cursor = 0
	initialEntryCount := len(model.leftPane.entries)

	// Execute delete operation
	model = model.executeDeleteOperation()

	// Verify nothing was deleted
	if len(model.leftPane.entries) != initialEntryCount {
		t.Errorf("Parent directory deletion should be no-op, but entry count changed from %d to %d",
			initialEntryCount, len(model.leftPane.entries))
	}

	// Verify cursor remained at 0
	if model.leftPane.cursor != 0 {
		t.Errorf("Expected cursor to remain at 0, got %d", model.leftPane.cursor)
	}
}

// TestExecuteDeleteOperation_PreservesIncrementalFilter tests filter preservation after single file deletion
func TestExecuteDeleteOperation_PreservesIncrementalFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "apple.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "apricot.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "banana.txt"), []byte("test"), 0644)

	theme := DefaultTheme()
	pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, theme)
	if err != nil {
		t.Fatalf("Failed to create pane: %v", err)
	}

	model := Model{
		leftPane:   pane,
		rightPane:  pane,
		activePane: LeftPane,
		theme:      theme,
	}

	// Apply incremental filter
	if err := model.leftPane.ApplyFilter("ap", SearchModeIncremental); err != nil {
		t.Fatalf("ApplyFilter() failed: %v", err)
	}

	// Verify filter is applied (should have 2 entries: apple.txt, apricot.txt)
	if model.leftPane.filterPattern != "ap" {
		t.Fatalf("Filter pattern not set: got %q", model.leftPane.filterPattern)
	}
	if model.leftPane.filterMode != SearchModeIncremental {
		t.Fatalf("Filter mode not set: got %v", model.leftPane.filterMode)
	}

	// Find and select apple.txt
	for i, entry := range model.leftPane.entries {
		if entry.Name == "apple.txt" {
			model.leftPane.cursor = i
			break
		}
	}

	// Execute delete operation
	model = model.executeDeleteOperation()

	// Verify filter is preserved
	if model.leftPane.filterPattern != "ap" {
		t.Errorf("Filter pattern should be preserved, got %q", model.leftPane.filterPattern)
	}
	if model.leftPane.filterMode != SearchModeIncremental {
		t.Errorf("Filter mode should be preserved, got %v", model.leftPane.filterMode)
	}

	// Verify filtered entries are updated (only apricot.txt should remain in filtered view)
	foundApricot := false
	foundApple := false
	for _, entry := range model.leftPane.entries {
		if entry.Name == "apricot.txt" {
			foundApricot = true
		}
		if entry.Name == "apple.txt" {
			foundApple = true
		}
	}
	if !foundApricot {
		t.Error("apricot.txt should still be in filtered entries")
	}
	if foundApple {
		t.Error("apple.txt should have been removed from filtered entries")
	}

	// Verify cursor is within bounds
	if model.leftPane.cursor < 0 || model.leftPane.cursor >= len(model.leftPane.entries) {
		t.Errorf("Cursor out of bounds: %d (entries: %d)",
			model.leftPane.cursor, len(model.leftPane.entries))
	}
}

// TestExecuteDeleteOperation_PreservesRegexFilter tests filter preservation with regex filter
func TestExecuteDeleteOperation_PreservesRegexFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "data.csv"), []byte("test"), 0644)

	theme := DefaultTheme()
	pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, theme)
	if err != nil {
		t.Fatalf("Failed to create pane: %v", err)
	}

	model := Model{
		leftPane:   pane,
		rightPane:  pane,
		activePane: LeftPane,
		theme:      theme,
	}

	// Apply regex filter
	if err := model.leftPane.ApplyFilter(`file\d`, SearchModeRegex); err != nil {
		t.Fatalf("ApplyFilter() failed: %v", err)
	}

	// Find and select file1.txt
	for i, entry := range model.leftPane.entries {
		if entry.Name == "file1.txt" {
			model.leftPane.cursor = i
			break
		}
	}

	// Execute delete operation
	model = model.executeDeleteOperation()

	// Verify filter is preserved
	if model.leftPane.filterPattern != `file\d` {
		t.Errorf("Filter pattern should be preserved, got %q", model.leftPane.filterPattern)
	}
	if model.leftPane.filterMode != SearchModeRegex {
		t.Errorf("Filter mode should be preserved, got %v", model.leftPane.filterMode)
	}
}

// TestExecuteDeleteOperation_PreservesSQLLikeFilter tests filter preservation with SQL-like filter
func TestExecuteDeleteOperation_PreservesSQLLikeFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "document.pdf"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "document.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "image.png"), []byte("test"), 0644)

	theme := DefaultTheme()
	pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, theme)
	if err != nil {
		t.Fatalf("Failed to create pane: %v", err)
	}

	model := Model{
		leftPane:   pane,
		rightPane:  pane,
		activePane: LeftPane,
		theme:      theme,
	}

	// Apply SQL-like filter
	if err := model.leftPane.ApplyFilter("name LIKE 'document%'", SearchModeSQLLike); err != nil {
		t.Fatalf("ApplyFilter() failed: %v", err)
	}

	// Find and select document.pdf
	for i, entry := range model.leftPane.entries {
		if entry.Name == "document.pdf" {
			model.leftPane.cursor = i
			break
		}
	}

	// Execute delete operation
	model = model.executeDeleteOperation()

	// Verify filter is preserved
	if model.leftPane.filterPattern != "name LIKE 'document%'" {
		t.Errorf("Filter pattern should be preserved, got %q", model.leftPane.filterPattern)
	}
	if model.leftPane.filterMode != SearchModeSQLLike {
		t.Errorf("Filter mode should be preserved, got %v", model.leftPane.filterMode)
	}
}

// TestExecuteDeleteOperation_MultipleFiles_PreservesFilter tests filter preservation after multiple file deletion
func TestExecuteDeleteOperation_MultipleFiles_PreservesFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "alpha.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "another.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "apple.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "banana.txt"), []byte("test"), 0644)

	theme := DefaultTheme()
	pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, theme)
	if err != nil {
		t.Fatalf("Failed to create pane: %v", err)
	}

	model := Model{
		leftPane:   pane,
		rightPane:  pane,
		activePane: LeftPane,
		theme:      theme,
	}

	// Apply filter
	if err := model.leftPane.ApplyFilter("a", SearchModeIncremental); err != nil {
		t.Fatalf("ApplyFilter() failed: %v", err)
	}

	// Mark multiple files for deletion
	for _, entry := range model.leftPane.entries {
		if entry.Name == "alpha.txt" || entry.Name == "another.txt" {
			model.leftPane.markedFiles[entry.Name] = true
		}
	}

	// Execute delete operation
	model = model.executeDeleteOperation()

	// Verify filter is preserved
	if model.leftPane.filterPattern != "a" {
		t.Errorf("Filter pattern should be preserved, got %q", model.leftPane.filterPattern)
	}
	if model.leftPane.filterMode != SearchModeIncremental {
		t.Errorf("Filter mode should be preserved, got %v", model.leftPane.filterMode)
	}

	// Verify only apple.txt remains in filtered view (matches "a")
	foundApple := false
	for _, entry := range model.leftPane.entries {
		if entry.Name == "apple.txt" {
			foundApple = true
		}
		if entry.Name == "alpha.txt" || entry.Name == "another.txt" {
			t.Errorf("Deleted file %s should not be in filtered entries", entry.Name)
		}
	}
	if !foundApple {
		t.Error("apple.txt should still be in filtered entries")
	}

	// Verify marks are cleared
	if len(model.leftPane.markedFiles) != 0 {
		t.Errorf("Marks should be cleared after deletion, got %d marks", len(model.leftPane.markedFiles))
	}
}

// TestExecuteDeleteOperation_NoFilter_NoRegression tests that delete without filter still works
func TestExecuteDeleteOperation_NoFilter_NoRegression(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte("test"), 0644)

	theme := DefaultTheme()
	pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, theme)
	if err != nil {
		t.Fatalf("Failed to create pane: %v", err)
	}

	model := Model{
		leftPane:   pane,
		rightPane:  pane,
		activePane: LeftPane,
		theme:      theme,
	}

	// Ensure no filter is applied
	if model.leftPane.filterPattern != "" {
		t.Fatalf("Filter should not be applied initially")
	}

	// Find file2.txt
	for i, entry := range model.leftPane.entries {
		if entry.Name == "file2.txt" {
			model.leftPane.cursor = i
			break
		}
	}

	initialEntryCount := len(model.leftPane.entries)

	// Execute delete operation
	model = model.executeDeleteOperation()

	// Verify filter remains empty
	if model.leftPane.filterPattern != "" {
		t.Errorf("Filter pattern should remain empty, got %q", model.leftPane.filterPattern)
	}
	if model.leftPane.filterMode != SearchModeNone {
		t.Errorf("Filter mode should remain SearchModeNone, got %v", model.leftPane.filterMode)
	}

	// Verify entry count decreased by 1
	if len(model.leftPane.entries) != initialEntryCount-1 {
		t.Errorf("Entry count should decrease by 1, got %d (was %d)",
			len(model.leftPane.entries), initialEntryCount)
	}

	// Verify file2.txt is deleted
	for _, entry := range model.leftPane.entries {
		if entry.Name == "file2.txt" {
			t.Error("file2.txt should have been deleted")
		}
	}
}

// TestExecuteDeleteOperation_ErrorHandling_CursorStillSet tests cursor positioning when deletion fails
func TestExecuteDeleteOperation_ErrorHandling_CursorStillSet(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a read-only directory to cause deletion failure
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create a file inside
	testFile := filepath.Join(readOnlyDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Make directory read-only (deletion will fail)
	if err := os.Chmod(readOnlyDir, 0555); err != nil {
		t.Skipf("Cannot change permissions (may need root): %v", err)
	}
	defer os.Chmod(readOnlyDir, 0755) // Restore for cleanup

	theme := DefaultTheme()
	pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, theme)
	if err != nil {
		t.Fatalf("Failed to create pane: %v", err)
	}

	model := Model{
		leftPane:   pane,
		rightPane:  pane,
		activePane: LeftPane,
		theme:      theme,
	}

	// Find and try to delete the read-only directory
	targetIndex := -1
	for i, entry := range model.leftPane.entries {
		if entry.Name == "readonly" {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		t.Fatal("Could not find readonly directory in entries")
	}

	model.leftPane.cursor = targetIndex

	// Execute delete operation (should fail but not crash)
	model = model.executeDeleteOperation()

	// Verify cursor is still set to a valid position
	if model.leftPane.cursor < 0 || model.leftPane.cursor >= len(model.leftPane.entries) {
		t.Errorf("Cursor should be within valid bounds even after error, got %d (entries: %d)",
			model.leftPane.cursor, len(model.leftPane.entries))
	}

	// Verify error dialog was set
	if model.dialog == nil {
		t.Error("Expected error dialog to be shown after deletion failure")
	}
}
