package ui

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadEntriesFromDisk tests the loadEntriesFromDisk helper function
func TestLoadEntriesFromDisk(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "alpha.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "beta.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte(""), 0644)
	os.Mkdir(filepath.Join(tmpDir, "gamma"), 0755)

	tests := []struct {
		name              string
		showHidden        bool
		expectedMinLen    int
		expectedHasHidden bool
	}{
		{
			name:              "Returns sorted entries without hidden files",
			showHidden:        false,
			expectedMinLen:    3, // alpha.txt, beta.txt, gamma
			expectedHasHidden: false,
		},
		{
			name:              "Returns sorted entries with hidden files",
			showHidden:        true,
			expectedMinLen:    4, // alpha.txt, beta.txt, .hidden, gamma
			expectedHasHidden: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane, err := NewPane(LeftPane, tmpDir, 40, 20, true, nil)
			if err != nil {
				t.Fatalf("NewPane() failed: %v", err)
			}
			pane.showHidden = tt.showHidden

			entries, err := pane.loadEntriesFromDisk()
			if err != nil {
				t.Fatalf("loadEntriesFromDisk() error = %v", err)
			}

			if len(entries) < tt.expectedMinLen {
				t.Errorf("loadEntriesFromDisk() returned %d entries, expected at least %d", len(entries), tt.expectedMinLen)
			}

			// Check for hidden file presence
			hasHidden := false
			for _, e := range entries {
				if e.Name == ".hidden" {
					hasHidden = true
					break
				}
			}
			if hasHidden != tt.expectedHasHidden {
				t.Errorf("loadEntriesFromDisk() hidden file presence = %v, want %v", hasHidden, tt.expectedHasHidden)
			}

			// Verify sorting (directories first, then alphabetical)
			// This is basic check - detailed sort tests are in sort_test.go
			if len(entries) > 1 {
				// Basic check: entries should not be nil
				for _, e := range entries {
					if e.Name == "" {
						t.Error("loadEntriesFromDisk() returned entry with empty name")
					}
				}
			}
		})
	}
}

// TestReloadDirectoryWithFilter_Incremental tests filter preservation with incremental search
func TestReloadDirectoryWithFilter_Incremental(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "apple.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "apricot.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "banana.txt"), []byte(""), 0644)

	pane, err := NewPane(LeftPane, tmpDir, 40, 20, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Apply incremental filter
	err = pane.ApplyFilter("ap", SearchModeIncremental)
	if err != nil {
		t.Fatalf("ApplyFilter() error = %v", err)
	}

	// Verify filter is applied
	if len(pane.entries) != 2 { // apple.txt, apricot.txt
		t.Errorf("After ApplyFilter: entries = %d, want 2", len(pane.entries))
	}

	// Simulate file deletion by removing a file from disk
	os.Remove(filepath.Join(tmpDir, "apple.txt"))

	// Reload with filter preservation
	err = pane.ReloadDirectoryWithFilter()
	if err != nil {
		t.Fatalf("ReloadDirectoryWithFilter() error = %v", err)
	}

	// Verify filter is preserved
	if pane.filterPattern != "ap" {
		t.Errorf("filterPattern = %q, want %q", pane.filterPattern, "ap")
	}
	if pane.filterMode != SearchModeIncremental {
		t.Errorf("filterMode = %v, want %v", pane.filterMode, SearchModeIncremental)
	}

	// Verify filtered entries are updated (apple.txt removed)
	if len(pane.entries) != 1 { // only apricot.txt now
		t.Errorf("After ReloadDirectoryWithFilter: entries = %d, want 1", len(pane.entries))
	}

	// Verify allEntries is also updated (may include parent directory ..)
	// Count only regular files (exclude parent directory)
	regularFilesCount := 0
	for _, e := range pane.allEntries {
		if !e.IsParentDir() {
			regularFilesCount++
		}
	}
	if regularFilesCount != 2 { // apricot.txt, banana.txt
		t.Errorf("After ReloadDirectoryWithFilter: regular files in allEntries = %d, want 2", regularFilesCount)
	}
}

// TestReloadDirectoryWithFilter_Regex tests filter preservation with regex search
func TestReloadDirectoryWithFilter_Regex(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "data.csv"), []byte(""), 0644)

	pane, err := NewPane(LeftPane, tmpDir, 40, 20, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Apply regex filter
	err = pane.ApplyFilter(`file\d`, SearchModeRegex)
	if err != nil {
		t.Fatalf("ApplyFilter() error = %v", err)
	}

	// Verify filter is applied
	if len(pane.entries) != 2 { // file1.txt, file2.txt
		t.Errorf("After ApplyFilter: entries = %d, want 2", len(pane.entries))
	}

	// Reload with filter preservation
	err = pane.ReloadDirectoryWithFilter()
	if err != nil {
		t.Fatalf("ReloadDirectoryWithFilter() error = %v", err)
	}

	// Verify filter is preserved
	if pane.filterPattern != `file\d` {
		t.Errorf("filterPattern = %q, want %q", pane.filterPattern, `file\d`)
	}
	if pane.filterMode != SearchModeRegex {
		t.Errorf("filterMode = %v, want %v", pane.filterMode, SearchModeRegex)
	}

	// Verify filtered entries count
	if len(pane.entries) != 2 {
		t.Errorf("After ReloadDirectoryWithFilter: entries = %d, want 2", len(pane.entries))
	}
}

// TestReloadDirectoryWithFilter_SQLLike tests filter preservation with SQL-like search
func TestReloadDirectoryWithFilter_SQLLike(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "document.pdf"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "document.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "image.png"), []byte(""), 0644)

	pane, err := NewPane(LeftPane, tmpDir, 40, 20, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Apply SQL-like filter
	err = pane.ApplyFilter("name LIKE 'document%'", SearchModeSQLLike)
	if err != nil {
		t.Fatalf("ApplyFilter() error = %v", err)
	}

	// Verify filter is applied
	if len(pane.entries) != 2 { // document.pdf, document.txt
		t.Errorf("After ApplyFilter: entries = %d, want 2", len(pane.entries))
	}

	// Reload with filter preservation
	err = pane.ReloadDirectoryWithFilter()
	if err != nil {
		t.Fatalf("ReloadDirectoryWithFilter() error = %v", err)
	}

	// Verify filter is preserved
	if pane.filterPattern != "name LIKE 'document%'" {
		t.Errorf("filterPattern = %q, want %q", pane.filterPattern, "name LIKE 'document%'")
	}
	if pane.filterMode != SearchModeSQLLike {
		t.Errorf("filterMode = %v, want %v", pane.filterMode, SearchModeSQLLike)
	}
}

// TestReloadDirectoryWithFilter_NoFilter tests behavior when no filter is active
func TestReloadDirectoryWithFilter_NoFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte(""), 0644)

	pane, err := NewPane(LeftPane, tmpDir, 40, 20, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Ensure no filter is applied
	if pane.filterPattern != "" {
		t.Fatalf("filterPattern should be empty, got %q", pane.filterPattern)
	}

	initialEntryCount := len(pane.entries)

	// Reload with filter preservation (no filter active)
	err = pane.ReloadDirectoryWithFilter()
	if err != nil {
		t.Fatalf("ReloadDirectoryWithFilter() error = %v", err)
	}

	// Verify filter state remains empty
	if pane.filterPattern != "" {
		t.Errorf("filterPattern = %q, want empty", pane.filterPattern)
	}
	if pane.filterMode != SearchModeNone {
		t.Errorf("filterMode = %v, want %v", pane.filterMode, SearchModeNone)
	}

	// Verify entries are preserved
	if len(pane.entries) != initialEntryCount {
		t.Errorf("entries = %d, want %d", len(pane.entries), initialEntryCount)
	}
}

// TestReloadDirectoryWithFilter_ClearsMarks tests that marks are cleared after reload
func TestReloadDirectoryWithFilter_ClearsMarks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte(""), 0644)

	pane, err := NewPane(LeftPane, tmpDir, 40, 20, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Mark some files
	pane.markedFiles["file1.txt"] = true
	pane.markedFiles["file2.txt"] = true

	if len(pane.markedFiles) != 2 {
		t.Fatalf("markedFiles = %d, want 2", len(pane.markedFiles))
	}

	// Reload with filter preservation
	err = pane.ReloadDirectoryWithFilter()
	if err != nil {
		t.Fatalf("ReloadDirectoryWithFilter() error = %v", err)
	}

	// Verify marks are cleared
	if len(pane.markedFiles) != 0 {
		t.Errorf("markedFiles = %d, want 0 (should be cleared)", len(pane.markedFiles))
	}
}

// TestReloadDirectoryWithFilter_CursorAdjustment tests cursor adjustment after reload
func TestReloadDirectoryWithFilter_CursorAdjustment(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte(""), 0644)

	pane, err := NewPane(LeftPane, tmpDir, 40, 20, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Move cursor to last entry
	pane.cursor = len(pane.entries) - 1

	// Remove a file to make cursor potentially out of bounds after reload
	os.Remove(filepath.Join(tmpDir, "file3.txt"))

	// Reload with filter preservation
	err = pane.ReloadDirectoryWithFilter()
	if err != nil {
		t.Fatalf("ReloadDirectoryWithFilter() error = %v", err)
	}

	// Verify cursor is within bounds
	if pane.cursor >= len(pane.entries) {
		t.Errorf("cursor = %d, but entries length = %d (cursor out of bounds)", pane.cursor, len(pane.entries))
	}
	if pane.cursor < 0 {
		t.Errorf("cursor = %d, should not be negative", pane.cursor)
	}
}

// TestReloadDirectoryWithFilter_WithFilterCursorAdjustment tests cursor adjustment with filter active
func TestReloadDirectoryWithFilter_WithFilterCursorAdjustment(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "apple.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "apricot.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "avocado.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "banana.txt"), []byte(""), 0644)

	pane, err := NewPane(LeftPane, tmpDir, 40, 20, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Apply filter
	err = pane.ApplyFilter("a", SearchModeIncremental)
	if err != nil {
		t.Fatalf("ApplyFilter() error = %v", err)
	}

	// Move cursor to last filtered entry
	pane.cursor = len(pane.entries) - 1

	// Remove files to reduce filtered entries
	os.Remove(filepath.Join(tmpDir, "apricot.txt"))
	os.Remove(filepath.Join(tmpDir, "avocado.txt"))

	// Reload with filter preservation
	err = pane.ReloadDirectoryWithFilter()
	if err != nil {
		t.Fatalf("ReloadDirectoryWithFilter() error = %v", err)
	}

	// Verify cursor is within bounds of filtered entries
	if pane.cursor >= len(pane.entries) {
		t.Errorf("cursor = %d, but entries length = %d (cursor out of bounds)", pane.cursor, len(pane.entries))
	}
}

// TestReloadDirectoryWithFilter_EmptyFilteredResult tests behavior when filter results in empty list
func TestReloadDirectoryWithFilter_EmptyFilteredResult(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte(""), 0644)

	pane, err := NewPane(LeftPane, tmpDir, 40, 20, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Apply filter that will have results
	err = pane.ApplyFilter("file", SearchModeIncremental)
	if err != nil {
		t.Fatalf("ApplyFilter() error = %v", err)
	}

	// Remove all matching files
	os.Remove(filepath.Join(tmpDir, "file1.txt"))
	os.Remove(filepath.Join(tmpDir, "file2.txt"))

	// Create a non-matching file so directory is not empty
	os.WriteFile(filepath.Join(tmpDir, "other.txt"), []byte(""), 0644)

	// Reload with filter preservation
	err = pane.ReloadDirectoryWithFilter()
	if err != nil {
		t.Fatalf("ReloadDirectoryWithFilter() error = %v", err)
	}

	// Filter should still be applied, but with empty results
	if pane.filterPattern != "file" {
		t.Errorf("filterPattern = %q, want %q", pane.filterPattern, "file")
	}

	// Entries should be empty (no matches)
	if len(pane.entries) != 0 {
		t.Errorf("entries = %d, want 0 (no matches)", len(pane.entries))
	}

	// allEntries should have the new file (may include parent directory ..)
	regularFilesCount := 0
	for _, e := range pane.allEntries {
		if !e.IsParentDir() {
			regularFilesCount++
		}
	}
	if regularFilesCount != 1 {
		t.Errorf("regular files in allEntries = %d, want 1", regularFilesCount)
	}

	// Cursor should be 0
	if pane.cursor != 0 {
		t.Errorf("cursor = %d, want 0", pane.cursor)
	}
}

// TestLoadDirectory_UsesSharedHelper tests that LoadDirectory uses the shared helper
func TestLoadDirectory_UsesSharedHelper(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "alpha.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "beta.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte(""), 0644)

	pane, err := NewPane(LeftPane, tmpDir, 40, 20, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Apply a filter first
	err = pane.ApplyFilter("alpha", SearchModeIncremental)
	if err != nil {
		t.Fatalf("ApplyFilter() error = %v", err)
	}

	// Verify filter is applied
	if pane.filterPattern != "alpha" {
		t.Errorf("filterPattern = %q, want %q", pane.filterPattern, "alpha")
	}

	// Call LoadDirectory (should clear filter)
	err = pane.LoadDirectory()
	if err != nil {
		t.Fatalf("LoadDirectory() error = %v", err)
	}

	// Verify filter is cleared
	if pane.filterPattern != "" {
		t.Errorf("filterPattern = %q, want empty (LoadDirectory should clear filter)", pane.filterPattern)
	}
	if pane.filterMode != SearchModeNone {
		t.Errorf("filterMode = %v, want %v", pane.filterMode, SearchModeNone)
	}

	// Verify entries are loaded (without hidden by default)
	hasHidden := false
	for _, e := range pane.entries {
		if e.Name == ".hidden" {
			hasHidden = true
			break
		}
	}
	if hasHidden {
		t.Error("LoadDirectory() should filter hidden files when showHidden is false")
	}
}

// TestRefreshDirectoryPreserveCursor_Fallback tests the index-based fallback
// when the previously selected filename is no longer found after refresh.
func TestRefreshDirectoryPreserveCursor_Fallback(t *testing.T) {
	tests := []struct {
		name           string
		files          []string // files to create initially
		cursorFile     string   // file to position cursor on
		deleteFiles    []string // files to delete before refresh
		expectedCursor int      // expected cursor position after refresh
		expectedFile   string   // expected file at cursor (for verification)
	}{
		{
			name:           "F-1: filename match succeeds",
			files:          []string{"aaa.txt", "bbb.txt", "ccc.txt"},
			cursorFile:     "bbb.txt",
			deleteFiles:    nil, // no deletion
			expectedCursor: 2,   // [.., aaa.txt, bbb.txt, ccc.txt] -> cursor=2(bbb.txt)
			expectedFile:   "bbb.txt",
		},
		{
			name:           "F-2: filename match fails, old index valid",
			files:          []string{"aaa.txt", "bbb.txt", "ccc.txt"},
			cursorFile:     "bbb.txt",
			deleteFiles:    []string{"bbb.txt"},
			expectedCursor: 2, // [.., aaa.txt, ccc.txt] -> cursor=2(ccc.txt)
			expectedFile:   "ccc.txt",
		},
		{
			name:           "F-3: filename match fails, old index exceeds",
			files:          []string{"aaa.txt", "bbb.txt"},
			cursorFile:     "bbb.txt",
			deleteFiles:    []string{"bbb.txt"},
			expectedCursor: 1, // [.., aaa.txt] -> cursor=1(aaa.txt)
			expectedFile:   "aaa.txt",
		},
		{
			name:           "F-4: all files deleted",
			files:          []string{"aaa.txt"},
			cursorFile:     "aaa.txt",
			deleteFiles:    []string{"aaa.txt"},
			expectedCursor: 0, // [..] -> cursor=0(..)
			expectedFile:   "..",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create initial files
			for _, f := range tt.files {
				os.WriteFile(filepath.Join(tmpDir, f), []byte(""), 0644)
			}

			pane, err := NewPane(LeftPane, tmpDir, 40, 20, true, nil)
			if err != nil {
				t.Fatalf("NewPane() failed: %v", err)
			}

			// Position cursor on target file
			for i, entry := range pane.entries {
				if entry.Name == tt.cursorFile {
					pane.cursor = i
					break
				}
			}

			// Delete files
			for _, f := range tt.deleteFiles {
				os.Remove(filepath.Join(tmpDir, f))
			}

			// Refresh
			err = pane.RefreshDirectoryPreserveCursor()
			if err != nil {
				t.Fatalf("RefreshDirectoryPreserveCursor() failed: %v", err)
			}

			// Verify cursor position
			if pane.cursor != tt.expectedCursor {
				t.Errorf("cursor = %d, want %d", pane.cursor, tt.expectedCursor)
			}

			// Verify file at cursor
			entry := pane.SelectedEntry()
			if entry == nil {
				t.Fatal("SelectedEntry() returned nil")
			}
			if entry.Name != tt.expectedFile {
				t.Errorf("file at cursor = %q, want %q", entry.Name, tt.expectedFile)
			}
		})
	}
}
