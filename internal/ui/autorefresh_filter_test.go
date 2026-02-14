package ui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// setupModelWithTmpDir creates a Model initialized with temporary directories for both panes.
// Returns the model and the temp dir path.
func setupModelWithTmpDir(t *testing.T, refreshRate int) (Model, string) {
	t.Helper()

	tmpDir := t.TempDir()
	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte(""), 0644)

	model := NewModelWithConfig(ModelOptions{
		RefreshRate: refreshRate,
	})

	// Initialize with WindowSizeMsg
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m := updatedModel.(Model)

	// Navigate both panes to tmpDir
	m.leftPane.path = tmpDir
	m.leftPane.LoadDirectory()
	m.rightPane.path = tmpDir
	m.rightPane.LoadDirectory()

	return m, tmpDir
}

// sendAutoRefresh simulates an autoRefreshMsg through Model.Update().
func sendAutoRefresh(t *testing.T, m Model) Model {
	t.Helper()
	updatedModel, _ := m.Update(autoRefreshMsg{})
	return updatedModel.(Model)
}

func TestAutoRefreshPreservesIncrementalFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	// Apply incremental filter on active pane
	activePane := m.getActivePane()
	err := activePane.ApplyFilter("txt", SearchModeIncremental)
	if err != nil {
		t.Fatalf("ApplyFilter failed: %v", err)
	}

	// Verify filter is applied
	if !activePane.IsFiltered() {
		t.Fatal("Filter should be applied before auto-refresh")
	}
	filteredCountBefore := len(activePane.entries)
	if filteredCountBefore == 0 {
		t.Fatal("Filtered entries should not be empty")
	}

	// Simulate auto-refresh
	m = sendAutoRefresh(t, m)

	// Verify filter is preserved
	activePane = m.getActivePane()
	if !activePane.IsFiltered() {
		t.Error("Incremental filter should be preserved after auto-refresh")
	}
	if activePane.FilterPattern() != "txt" {
		t.Errorf("FilterPattern() = %q, want %q", activePane.FilterPattern(), "txt")
	}
	if activePane.FilterMode() != SearchModeIncremental {
		t.Errorf("FilterMode() = %v, want %v", activePane.FilterMode(), SearchModeIncremental)
	}
	if len(activePane.entries) != filteredCountBefore {
		t.Errorf("Filtered entry count changed: got %d, want %d", len(activePane.entries), filteredCountBefore)
	}
}

func TestAutoRefreshPreservesRegexFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	activePane := m.getActivePane()
	err := activePane.ApplyFilter(`\.go$`, SearchModeRegex)
	if err != nil {
		t.Fatalf("ApplyFilter failed: %v", err)
	}

	if !activePane.IsFiltered() {
		t.Fatal("Regex filter should be applied before auto-refresh")
	}
	filteredCountBefore := len(activePane.entries)

	// Simulate auto-refresh
	m = sendAutoRefresh(t, m)

	activePane = m.getActivePane()
	if !activePane.IsFiltered() {
		t.Error("Regex filter should be preserved after auto-refresh")
	}
	if activePane.FilterPattern() != `\.go$` {
		t.Errorf("FilterPattern() = %q, want %q", activePane.FilterPattern(), `\.go$`)
	}
	if activePane.FilterMode() != SearchModeRegex {
		t.Errorf("FilterMode() = %v, want %v", activePane.FilterMode(), SearchModeRegex)
	}
	if len(activePane.entries) != filteredCountBefore {
		t.Errorf("Filtered entry count changed: got %d, want %d", len(activePane.entries), filteredCountBefore)
	}
}

func TestAutoRefreshPreservesSQLLikeFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	activePane := m.getActivePane()
	query := "name LIKE '%.txt'"
	err := activePane.ApplyFilter(query, SearchModeSQLLike)
	if err != nil {
		t.Fatalf("ApplyFilter failed: %v", err)
	}

	if !activePane.IsFiltered() {
		t.Fatal("SQL LIKE filter should be applied before auto-refresh")
	}
	filteredCountBefore := len(activePane.entries)

	// Simulate auto-refresh
	m = sendAutoRefresh(t, m)

	activePane = m.getActivePane()
	if !activePane.IsFiltered() {
		t.Error("SQL LIKE filter should be preserved after auto-refresh")
	}
	if activePane.FilterPattern() != query {
		t.Errorf("FilterPattern() = %q, want %q", activePane.FilterPattern(), query)
	}
	if activePane.FilterMode() != SearchModeSQLLike {
		t.Errorf("FilterMode() = %v, want %v", activePane.FilterMode(), SearchModeSQLLike)
	}
	if len(activePane.entries) != filteredCountBefore {
		t.Errorf("Filtered entry count changed: got %d, want %d", len(activePane.entries), filteredCountBefore)
	}
}

func TestAutoRefreshPreservesFilterOnBothPanes(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	// Apply different filters on both panes
	m.leftPane.ApplyFilter("txt", SearchModeIncremental)
	m.rightPane.ApplyFilter(`\.go$`, SearchModeRegex)

	if !m.leftPane.IsFiltered() || !m.rightPane.IsFiltered() {
		t.Fatal("Both panes should have filters applied")
	}

	// Simulate auto-refresh
	m = sendAutoRefresh(t, m)

	// Verify left pane filter preserved
	if !m.leftPane.IsFiltered() {
		t.Error("Left pane filter should be preserved after auto-refresh")
	}
	if m.leftPane.FilterPattern() != "txt" {
		t.Errorf("Left pane FilterPattern() = %q, want %q", m.leftPane.FilterPattern(), "txt")
	}
	if m.leftPane.FilterMode() != SearchModeIncremental {
		t.Errorf("Left pane FilterMode() = %v, want %v", m.leftPane.FilterMode(), SearchModeIncremental)
	}

	// Verify right pane filter preserved
	if !m.rightPane.IsFiltered() {
		t.Error("Right pane filter should be preserved after auto-refresh")
	}
	if m.rightPane.FilterPattern() != `\.go$` {
		t.Errorf("Right pane FilterPattern() = %q, want %q", m.rightPane.FilterPattern(), `\.go$`)
	}
	if m.rightPane.FilterMode() != SearchModeRegex {
		t.Errorf("Right pane FilterMode() = %v, want %v", m.rightPane.FilterMode(), SearchModeRegex)
	}
}

func TestAutoRefreshWithNoFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	// No filter applied - verify refresh works normally
	totalBefore := len(m.getActivePane().entries)

	m = sendAutoRefresh(t, m)

	activePane := m.getActivePane()
	if activePane.IsFiltered() {
		t.Error("No filter should be active after auto-refresh when none was set")
	}
	if len(activePane.entries) != totalBefore {
		t.Errorf("Entry count changed: got %d, want %d", len(activePane.entries), totalBefore)
	}
}

func TestAutoRefreshPreservesCursorWithFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	activePane := m.getActivePane()
	activePane.ApplyFilter("file", SearchModeIncremental)

	// Move cursor to a specific position
	if len(activePane.entries) > 1 {
		activePane.cursor = 1
		selectedName := activePane.entries[1].Name

		m = sendAutoRefresh(t, m)

		activePane = m.getActivePane()
		if activePane.cursor < len(activePane.entries) {
			if activePane.entries[activePane.cursor].Name != selectedName {
				t.Errorf("Cursor should point to %q, got %q",
					selectedName, activePane.entries[activePane.cursor].Name)
			}
		}
	}
}

func TestAutoRefreshPreservesMarksWithFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	activePane := m.getActivePane()
	// Mark a file before applying filter
	for i, e := range activePane.entries {
		if e.Name == "file1.txt" {
			activePane.cursor = i
			activePane.ToggleMark()
			break
		}
	}

	// Apply filter
	activePane.ApplyFilter("file", SearchModeIncremental)

	if !activePane.IsMarked("file1.txt") {
		t.Fatal("file1.txt should be marked before auto-refresh")
	}

	m = sendAutoRefresh(t, m)

	activePane = m.getActivePane()
	if !activePane.IsMarked("file1.txt") {
		t.Error("Marks should be preserved after auto-refresh with filter")
	}
}

func TestMultipleAutoRefreshesPreserveFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	activePane := m.getActivePane()
	activePane.ApplyFilter("txt", SearchModeIncremental)

	// Simulate multiple consecutive auto-refreshes
	for i := 0; i < 5; i++ {
		m = sendAutoRefresh(t, m)

		activePane = m.getActivePane()
		if !activePane.IsFiltered() {
			t.Errorf("Filter should be preserved after auto-refresh #%d", i+1)
		}
		if activePane.FilterPattern() != "txt" {
			t.Errorf("FilterPattern() = %q after refresh #%d, want %q",
				activePane.FilterPattern(), i+1, "txt")
		}
	}
}

func TestAutoRefreshDisabledDoesNotAffectFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 0) // refreshRate = 0 (disabled)

	activePane := m.getActivePane()
	activePane.ApplyFilter("go", SearchModeIncremental)

	// autoRefreshMsg with refreshRate=0 should be a no-op
	m = sendAutoRefresh(t, m)

	activePane = m.getActivePane()
	if !activePane.IsFiltered() {
		t.Error("Filter should still be present when auto-refresh is disabled")
	}
}

func TestAutoRefreshSuppressedDuringDialog(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	activePane := m.getActivePane()
	activePane.ApplyFilter("txt", SearchModeIncremental)

	// Open a dialog
	m.dialog = NewErrorDialog("test")

	m = sendAutoRefresh(t, m)

	// Filter should be preserved because refresh is suppressed during dialog
	activePane = m.getActivePane()
	if !activePane.IsFiltered() {
		t.Error("Filter should be preserved when auto-refresh is suppressed during dialog")
	}
	if activePane.FilterPattern() != "txt" {
		t.Errorf("FilterPattern() = %q, want %q", activePane.FilterPattern(), "txt")
	}
}

func TestAutoRefreshPicksUpNewFilesWithFilter(t *testing.T) {
	m, tmpDir := setupModelWithTmpDir(t, 3)

	activePane := m.getActivePane()
	activePane.ApplyFilter("txt", SearchModeIncremental)
	countBefore := len(activePane.entries)

	// Add a new .txt file
	os.WriteFile(filepath.Join(tmpDir, "new_file.txt"), []byte(""), 0644)

	m = sendAutoRefresh(t, m)

	activePane = m.getActivePane()
	if !activePane.IsFiltered() {
		t.Error("Filter should be preserved after auto-refresh")
	}
	// New file should appear in filtered results (matches "txt")
	if len(activePane.entries) != countBefore+1 {
		t.Errorf("Filtered entry count: got %d, want %d (new .txt file should appear)",
			len(activePane.entries), countBefore+1)
	}
}

// === Pane.Refresh() tests (F5/Ctrl+R path) ===

func TestPaneRefreshPreservesIncrementalFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	activePane := m.getActivePane()
	activePane.ApplyFilter("txt", SearchModeIncremental)

	if !activePane.IsFiltered() {
		t.Fatal("Filter should be applied before Refresh()")
	}
	filteredCountBefore := len(activePane.entries)

	err := activePane.Refresh()
	if err != nil {
		t.Fatalf("Refresh() failed: %v", err)
	}

	if !activePane.IsFiltered() {
		t.Error("Incremental filter should be preserved after Refresh()")
	}
	if activePane.FilterPattern() != "txt" {
		t.Errorf("FilterPattern() = %q, want %q", activePane.FilterPattern(), "txt")
	}
	if activePane.FilterMode() != SearchModeIncremental {
		t.Errorf("FilterMode() = %v, want %v", activePane.FilterMode(), SearchModeIncremental)
	}
	if len(activePane.entries) != filteredCountBefore {
		t.Errorf("Filtered entry count changed: got %d, want %d", len(activePane.entries), filteredCountBefore)
	}
}

func TestPaneRefreshPreservesRegexFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	activePane := m.getActivePane()
	activePane.ApplyFilter(`\.go$`, SearchModeRegex)
	filteredCountBefore := len(activePane.entries)

	err := activePane.Refresh()
	if err != nil {
		t.Fatalf("Refresh() failed: %v", err)
	}

	if !activePane.IsFiltered() {
		t.Error("Regex filter should be preserved after Refresh()")
	}
	if activePane.FilterPattern() != `\.go$` {
		t.Errorf("FilterPattern() = %q, want %q", activePane.FilterPattern(), `\.go$`)
	}
	if len(activePane.entries) != filteredCountBefore {
		t.Errorf("Filtered entry count changed: got %d, want %d", len(activePane.entries), filteredCountBefore)
	}
}

func TestPaneRefreshPreservesSQLLikeFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	activePane := m.getActivePane()
	query := "name LIKE '%.txt'"
	activePane.ApplyFilter(query, SearchModeSQLLike)
	filteredCountBefore := len(activePane.entries)

	err := activePane.Refresh()
	if err != nil {
		t.Fatalf("Refresh() failed: %v", err)
	}

	if !activePane.IsFiltered() {
		t.Error("SQL LIKE filter should be preserved after Refresh()")
	}
	if activePane.FilterPattern() != query {
		t.Errorf("FilterPattern() = %q, want %q", activePane.FilterPattern(), query)
	}
	if len(activePane.entries) != filteredCountBefore {
		t.Errorf("Filtered entry count changed: got %d, want %d", len(activePane.entries), filteredCountBefore)
	}
}

func TestRefreshBothPanesPreservesFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	// Apply different filters on both panes
	m.leftPane.ApplyFilter("txt", SearchModeIncremental)
	m.rightPane.ApplyFilter(`\.go$`, SearchModeRegex)

	// Call RefreshBothPanes (F5/Ctrl+R path)
	m.RefreshBothPanes()

	// Verify left pane filter preserved
	if !m.leftPane.IsFiltered() {
		t.Error("Left pane filter should be preserved after RefreshBothPanes()")
	}
	if m.leftPane.FilterPattern() != "txt" {
		t.Errorf("Left pane FilterPattern() = %q, want %q", m.leftPane.FilterPattern(), "txt")
	}

	// Verify right pane filter preserved
	if !m.rightPane.IsFiltered() {
		t.Error("Right pane filter should be preserved after RefreshBothPanes()")
	}
	if m.rightPane.FilterPattern() != `\.go$` {
		t.Errorf("Right pane FilterPattern() = %q, want %q", m.rightPane.FilterPattern(), `\.go$`)
	}
}

func TestF5KeyPreservesFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	activePane := m.getActivePane()
	activePane.ApplyFilter("txt", SearchModeIncremental)

	// Simulate F5 key press through Model.Update()
	keyMsg := tea.KeyMsg{Type: tea.KeyF5}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	activePane = m.getActivePane()
	if !activePane.IsFiltered() {
		t.Error("Filter should be preserved after F5 key press")
	}
	if activePane.FilterPattern() != "txt" {
		t.Errorf("FilterPattern() = %q, want %q", activePane.FilterPattern(), "txt")
	}
}

func TestPaneRefreshPreservesMarksWithFilter(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	activePane := m.getActivePane()
	// Mark a file
	for i, e := range activePane.entries {
		if e.Name == "file1.txt" {
			activePane.cursor = i
			activePane.ToggleMark()
			break
		}
	}

	// Apply filter
	activePane.ApplyFilter("file", SearchModeIncremental)

	if !activePane.IsMarked("file1.txt") {
		t.Fatal("file1.txt should be marked before Refresh()")
	}

	err := activePane.Refresh()
	if err != nil {
		t.Fatalf("Refresh() failed: %v", err)
	}

	if !activePane.IsMarked("file1.txt") {
		t.Error("Marks should be preserved after Refresh() with filter")
	}
	if !activePane.IsFiltered() {
		t.Error("Filter should be preserved after Refresh()")
	}
}

func TestPaneRefreshPicksUpNewFilesWithFilter(t *testing.T) {
	m, tmpDir := setupModelWithTmpDir(t, 3)

	activePane := m.getActivePane()
	activePane.ApplyFilter("txt", SearchModeIncremental)
	countBefore := len(activePane.entries)

	// Add a new .txt file
	os.WriteFile(filepath.Join(tmpDir, "new_file.txt"), []byte(""), 0644)

	err := activePane.Refresh()
	if err != nil {
		t.Fatalf("Refresh() failed: %v", err)
	}

	if !activePane.IsFiltered() {
		t.Error("Filter should be preserved after Refresh()")
	}
	if len(activePane.entries) != countBefore+1 {
		t.Errorf("Filtered entry count: got %d, want %d (new .txt file should appear)",
			len(activePane.entries), countBefore+1)
	}
}

func TestAutoRefreshTickCmdReturnsCommand(t *testing.T) {
	duration := 3 * time.Second
	cmd := autoRefreshTickCmd(duration)
	if cmd == nil {
		t.Error("autoRefreshTickCmd should return a non-nil command")
	}
}

func TestAutoRefreshReturnsNextTickCommand(t *testing.T) {
	m, _ := setupModelWithTmpDir(t, 3)

	_, cmd := m.Update(autoRefreshMsg{})
	if cmd == nil {
		t.Error("autoRefreshMsg handler should return a tick command for the next refresh")
	}
}
