package ui

import (
	"os"
	"path/filepath"
	"reflect"
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

	pane, err := NewPane(LeftPane, tmpDir, 80, 24, true, nil)
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

// displayedNames returns the names of the pane's current display list,
// excluding the parent directory entry.
func displayedNames(p *Pane) []string {
	names := make([]string, 0, len(p.entries))
	for _, e := range p.entries {
		if e.IsParentDir() {
			continue
		}
		names = append(names, e.Name)
	}
	return names
}

// markByName marks the entry with the given name via the pane's public
// ToggleMark API (moves the cursor to that entry first, then toggles it).
func markByName(t *testing.T, p *Pane, name string) {
	t.Helper()
	for i, e := range p.entries {
		if e.Name == name {
			p.cursor = i
			if !p.ToggleMark() {
				t.Fatalf("ToggleMark failed for %q", name)
			}
			return
		}
	}
	t.Fatalf("entry %q not found in display list", name)
}

// setupFilterHiddenMarksFixture creates a pane with 6 marked names: some
// remain visible after an incremental filter for "vv" is applied (they
// contain "vv"), the rest are hidden by the filter (they do not).
func setupFilterHiddenMarksFixture(t *testing.T) (*Pane, string) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "duofm-mark-filter-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	entries := []struct {
		name  string
		isDir bool
	}{
		{"avvdir", true},        // directory, visible after filter
		{"zdir", true},          // directory, hidden after filter
		{"bvvfile.txt", false},  // file, visible after filter
		{"yvvfile.txt", false},  // file, visible after filter
		{"hidden_a.txt", false}, // file, hidden after filter
		{"hidden_b.txt", false}, // file, hidden after filter
	}

	for _, e := range entries {
		path := filepath.Join(tmpDir, e.name)
		if e.isDir {
			if err := os.Mkdir(path, 0755); err != nil {
				t.Fatalf("Failed to create dir %s: %v", e.name, err)
			}
		} else {
			if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
				t.Fatalf("Failed to create file %s: %v", e.name, err)
			}
		}
	}

	pane, err := NewPane(LeftPane, tmpDir, 80, 24, true, nil)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create pane: %v", err)
	}

	// Mark every name while the display list is still unfiltered.
	for _, e := range entries {
		markByName(t, pane, e.name)
	}

	if err := pane.ApplyFilter("vv", SearchModeIncremental); err != nil {
		t.Fatalf("ApplyFilter failed: %v", err)
	}

	return pane, tmpDir
}

// AC-1: with at least three names marked in the reverse of their display
// order, GetMarkedFiles returns them in display order, and this holds
// across many repeated calls within the same run (against an unordered
// map-based implementation, the wrong order shows up probabilistically).
func TestGetMarkedFilesOrderMatchesDisplayOrderAcrossRepeatedCalls(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	displayOrder := displayedNames(pane)
	if len(displayOrder) < 3 {
		t.Fatalf("need at least 3 displayable entries, got %d", len(displayOrder))
	}

	// Mark every displayed name in the reverse of display order.
	for i := len(displayOrder) - 1; i >= 0; i-- {
		markByName(t, pane, displayOrder[i])
	}

	for i := 0; i < 50; i++ {
		got := pane.GetMarkedFiles()
		if !reflect.DeepEqual(got, displayOrder) {
			t.Fatalf("iteration %d: GetMarkedFiles() = %v, want %v", i, got, displayOrder)
		}
	}
}

// AC-2: with directories and files marked in an order different from
// display order, GetMarkedFiles equals the display list's order restricted
// to marked names, and every marked directory precedes every marked file.
func TestGetMarkedFilesDirsPrecedeFilesWhenMarkedOutOfOrder(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "duofm-mark-dirs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	entries := []struct {
		name  string
		isDir bool
	}{
		{"zdir", true},
		{"adir", true},
		{"file1.txt", false},
		{"file2.txt", false},
		{"file3.txt", false},
	}
	for _, e := range entries {
		path := filepath.Join(tmpDir, e.name)
		if e.isDir {
			if err := os.Mkdir(path, 0755); err != nil {
				t.Fatalf("Failed to create dir %s: %v", e.name, err)
			}
		} else {
			if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
				t.Fatalf("Failed to create file %s: %v", e.name, err)
			}
		}
	}

	pane, err := NewPane(LeftPane, tmpDir, 80, 24, true, nil)
	if err != nil {
		t.Fatalf("Failed to create pane: %v", err)
	}

	displayOrder := displayedNames(pane)

	// Mark a subset in an order different from display order.
	marked := map[string]bool{"file2.txt": true, "zdir": true, "file1.txt": true, "adir": true}
	markByName(t, pane, "file2.txt")
	markByName(t, pane, "zdir")
	markByName(t, pane, "file1.txt")
	markByName(t, pane, "adir")

	var expected []string
	for _, name := range displayOrder {
		if marked[name] {
			expected = append(expected, name)
		}
	}

	got := pane.GetMarkedFiles()
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("GetMarkedFiles() = %v, want %v", got, expected)
	}

	seenFile := false
	for _, name := range got {
		isDir := name == "zdir" || name == "adir"
		if isDir && seenFile {
			t.Fatalf("marked directory %q appeared after a marked file in %v", name, got)
		}
		if !isDir {
			seenFile = true
		}
	}
}

// AC-3: after marking, the display order is changed (sort switched to name
// descending and re-applied via ApplySortAndPreserveCursor, which keeps
// marks); GetMarkedFiles follows the new display order at call time.
func TestGetMarkedFilesFollowsSortOrderChangeAfterMarking(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	markByName(t, pane, "file1.txt")
	markByName(t, pane, "file2.txt")
	markByName(t, pane, "subdir")

	preOrder := displayedNames(pane)
	marked := map[string]bool{"file1.txt": true, "file2.txt": true, "subdir": true}
	var preExpected []string
	for _, name := range preOrder {
		if marked[name] {
			preExpected = append(preExpected, name)
		}
	}

	pane.SetSortConfig(SortConfig{Field: SortByName, Order: SortDesc})
	pane.ApplySortAndPreserveCursor()

	postOrder := displayedNames(pane)
	var postExpected []string
	for _, name := range postOrder {
		if marked[name] {
			postExpected = append(postExpected, name)
		}
	}

	if reflect.DeepEqual(preExpected, postExpected) {
		t.Fatalf("test setup did not change display order: %v", postExpected)
	}

	got := pane.GetMarkedFiles()
	if !reflect.DeepEqual(got, postExpected) {
		t.Fatalf("GetMarkedFiles() after sort change = %v, want %v", got, postExpected)
	}
}

// AC-4: with an active filter hiding some marked names (at least two
// hidden, one of them a directory), GetMarkedFiles returns the visible
// marked names in display order followed by the hidden marked names in
// ascending name order; length equals MarkCount() and elements equal the
// mark set.
func TestGetMarkedFilesHiddenNamesAppendedAscending(t *testing.T) {
	pane, tmpDir := setupFilterHiddenMarksFixture(t)
	defer os.RemoveAll(tmpDir)

	if got := pane.MarkCount(); got != 6 {
		t.Fatalf("expected MarkCount()=6, got %d", got)
	}

	visibleOrder := displayedNames(pane)
	expected := append([]string{}, visibleOrder...)
	expected = append(expected, "hidden_a.txt", "hidden_b.txt", "zdir")

	got := pane.GetMarkedFiles()
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("GetMarkedFiles() = %v, want %v", got, expected)
	}

	if len(got) != pane.MarkCount() {
		t.Fatalf("GetMarkedFiles() length = %d, want MarkCount()=%d", len(got), pane.MarkCount())
	}

	gotSet := make(map[string]bool, len(got))
	for _, name := range got {
		gotSet[name] = true
	}
	for _, name := range []string{"avvdir", "zdir", "bvvfile.txt", "yvvfile.txt", "hidden_a.txt", "hidden_b.txt"} {
		if !gotSet[name] {
			t.Errorf("expected marked name %q in GetMarkedFiles() result, missing", name)
		}
	}
}

// AC-5: with zero marks, both queries return empty lists. With exactly one
// mark, GetMarkedFiles returns exactly that name and GetMarkedFilePaths
// returns exactly that name's full path.
func TestGetMarkedFilesAndPathsEmptyAndSingleMark(t *testing.T) {
	pane, tmpDir := setupTestPane(t)
	defer os.RemoveAll(tmpDir)

	if got := pane.GetMarkedFiles(); len(got) != 0 {
		t.Fatalf("expected empty GetMarkedFiles() with zero marks, got %v", got)
	}
	if got := pane.GetMarkedFilePaths(); len(got) != 0 {
		t.Fatalf("expected empty GetMarkedFilePaths() with zero marks, got %v", got)
	}

	pane.cursor = 1
	entry := pane.SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		t.Fatal("expected non-parent entry at cursor 1")
	}
	if !pane.ToggleMark() {
		t.Fatal("ToggleMark failed for single mark")
	}

	names := pane.GetMarkedFiles()
	if len(names) != 1 || names[0] != entry.Name {
		t.Fatalf("GetMarkedFiles() = %v, want [%s]", names, entry.Name)
	}

	paths := pane.GetMarkedFilePaths()
	expectedPath := filepath.Join(tmpDir, entry.Name)
	if len(paths) != 1 || paths[0] != expectedPath {
		t.Fatalf("GetMarkedFilePaths() = %v, want [%s]", paths, expectedPath)
	}
}

// AC-6: for a state with several marks including at least one filter-hidden
// mark, GetMarkedFilePaths has the same length as GetMarkedFiles, and for
// every i its i-th element equals the pane path joined with the i-th name.
func TestGetMarkedFilePathsMatchesNamesWithHiddenMarks(t *testing.T) {
	pane, tmpDir := setupFilterHiddenMarksFixture(t)
	defer os.RemoveAll(tmpDir)

	names := pane.GetMarkedFiles()
	paths := pane.GetMarkedFilePaths()

	if len(paths) != len(names) {
		t.Fatalf("GetMarkedFilePaths() length = %d, want %d (GetMarkedFiles() length)", len(paths), len(names))
	}

	for i, name := range names {
		expected := filepath.Join(tmpDir, name)
		if paths[i] != expected {
			t.Errorf("paths[%d] = %s, want %s", i, paths[i], expected)
		}
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
