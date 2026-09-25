package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newFilesPane creates a pane over a fresh temp directory containing n
// regular files named "f00", "f01", ... Entries are sorted name-ascending
// with the parent directory first (SortEntries' behavior), so for a
// non-root temp directory (always the case for t.TempDir()):
//
//	entries[0]      == ".."
//	entries[1+i]    == "f0i" (i in [0, n))
//
// This is shared by mouse_hittest_test.go, pane_drag_test.go and
// model_update_mouse_test.go.
func newFilesPane(t *testing.T, id PanePosition, n, width, height int, active bool) *Pane {
	t.Helper()
	dir := t.TempDir()
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("f%02d", i)
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0644); err != nil {
			t.Fatalf("write file %s: %v", name, err)
		}
	}
	pane, err := NewPane(id, dir, width, height, active, nil)
	if err != nil {
		t.Fatalf("NewPane: %v", err)
	}
	return pane
}

// --- AC-1: hitTest at width 80/height 24, more entries than visible lines ---

func TestHitTest_BasicGridAndPaneSelection(t *testing.T) {
	// visibleLines=18 matches a real 80x24 window (paneHeight=22, getVisibleLines=18).
	layout := paneHitLayout{scrollOffset: 0, visibleLines: 18, entryCount: 30}

	tests := []struct {
		name      string
		x, y      int
		wantKind  hitKind
		wantPane  PanePosition
		wantIndex int
	}{
		{"left pane top entry", 5, 4, hitEntry, LeftPane, 0},
		{"left pane third entry", 39, 6, hitEntry, LeftPane, 2},
		{"right pane top entry", 40, 4, hitEntry, RightPane, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hit := hitTest(tt.x, tt.y, 80, 24, layout, layout)
			if hit.kind != tt.wantKind || hit.pane != tt.wantPane || (tt.wantKind == hitEntry && hit.index != tt.wantIndex) {
				t.Errorf("hitTest(%d,%d) = %+v, want kind=%v pane=%v index=%d", tt.x, tt.y, hit, tt.wantKind, tt.wantPane, tt.wantIndex)
			}
		})
	}
}

func TestHitTest_ScrollOffsetShiftsEntryIndex(t *testing.T) {
	const visibleLines = 18
	layout := paneHitLayout{scrollOffset: 10, visibleLines: visibleLines, entryCount: 100}

	// y=4 -> index10 (first visible row)
	hit := hitTest(5, 4, 80, 24, layout, layout)
	if hit.kind != hitEntry || hit.index != 10 {
		t.Errorf("y=4: got %+v, want hitEntry index=10", hit)
	}

	// y=4+V-1 -> index 10+V-1 (last visible row)
	hit = hitTest(5, 4+visibleLines-1, 80, 24, layout, layout)
	if hit.kind != hitEntry || hit.index != 10+visibleLines-1 {
		t.Errorf("y=4+V-1: got %+v, want hitEntry index=%d", hit, 10+visibleLines-1)
	}

	// y=4+V -> non-entry (one past the visible window)
	hit = hitTest(5, 4+visibleLines, 80, 24, layout, layout)
	if hit.kind != hitNonEntry {
		t.Errorf("y=4+V: got %+v, want hitNonEntry", hit)
	}
}

// --- AC-2: bars, out-of-screen, headers, filtered-empty, bg split, live adapters ---

func TestHitTest_OutOfScreenAndBars(t *testing.T) {
	layout := paneHitLayout{scrollOffset: 0, visibleLines: 18, entryCount: 3}

	cases := []struct {
		name string
		x, y int
	}{
		{"title bar", 5, 0},
		{"status bar", 5, 23}, // height-1 = 23
		{"leftover column of odd width", 80, 5},
		{"negative x", -1, 5},
		{"negative y", 5, -1},
		{"y beyond height", 5, 24},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hit := hitTest(c.x, c.y, 81, 24, layout, layout)
			if hit.kind != hitNone {
				t.Errorf("hitTest(%d,%d) = %+v, want hitNone", c.x, c.y, hit)
			}
		})
	}
}

func TestHitTest_HeaderRowsAndBlankRowAreNonEntry(t *testing.T) {
	layout := paneHitLayout{scrollOffset: 0, visibleLines: 18, entryCount: 3}

	for _, y := range []int{1, 2, 3, 8} {
		hit := hitTest(5, y, 81, 24, layout, layout)
		if hit.kind != hitNonEntry {
			t.Errorf("hitTest(_, %d) = %+v, want hitNonEntry", y, hit)
		}
	}
}

func TestHitTest_FilterWithNoMatchesIsNonEntry(t *testing.T) {
	empty := paneHitLayout{scrollOffset: 0, visibleLines: 18, entryCount: 0}
	hit := hitTest(5, mouseFirstEntryRow, 81, 24, empty, empty)
	if hit.kind != hitNonEntry {
		t.Errorf("hitTest at row4 with entryCount=0 = %+v, want hitNonEntry", hit)
	}
}

func TestHitTest_BgSplit_FileListRowsAreEntriesSeparatorIsNot(t *testing.T) {
	pane := newFilesPane(t, LeftPane, 20, 40, 22, true)
	pane.SetBgOutputActive(true)

	fileListHeight, _ := pane.bgSplitHeights()
	layout := pane.hitLayout()
	if layout.visibleLines != fileListHeight {
		t.Fatalf("hitLayout().visibleLines = %d, want bg split file list height %d", layout.visibleLines, fileListHeight)
	}

	other := paneHitLayout{scrollOffset: 0, visibleLines: 0, entryCount: 0}

	// A file-list row within the reduced height still maps to an entry.
	hit := hitTest(5, mouseFirstEntryRow, 40, 26, layout, other)
	if hit.kind != hitEntry || hit.index != 0 {
		t.Errorf("bg split file-list row: got %+v, want hitEntry index=0", hit)
	}

	// The row right after the file-list height is the separator: non-entry.
	hit = hitTest(5, mouseFirstEntryRow+fileListHeight, 40, 26, layout, other)
	if hit.kind != hitNonEntry {
		t.Errorf("bg split separator row: got %+v, want hitNonEntry", hit)
	}

	// Verify the same classification holds through the Model adapter.
	m := &Model{leftPane: pane, rightPane: newFilesPane(t, RightPane, 0, 40, 22, false), width: 80, height: 26}
	mhit := m.mouseHitAt(5, mouseFirstEntryRow)
	if mhit.kind != hitEntry {
		t.Errorf("mouseHitAt bg split file-list row: got %+v, want hitEntry", mhit)
	}
	mhit = m.mouseHitAt(5, mouseFirstEntryRow+fileListHeight)
	if mhit.kind != hitNonEntry {
		t.Errorf("mouseHitAt bg split separator row: got %+v, want hitNonEntry", mhit)
	}
}

func TestPaneHitLayout_FollowsLiveState(t *testing.T) {
	pane := newFilesPane(t, LeftPane, 20, 40, 22, true)

	before := pane.hitLayout()

	pane.scrollOffset = 3
	afterScroll := pane.hitLayout()
	if afterScroll.scrollOffset != 3 || afterScroll.scrollOffset == before.scrollOffset {
		t.Errorf("hitLayout() did not follow scrollOffset change: %+v", afterScroll)
	}

	pane.SetBgOutputActive(true)
	afterBg := pane.hitLayout()
	if afterBg.visibleLines == before.visibleLines {
		t.Errorf("hitLayout() did not follow bg state change: %+v", afterBg)
	}
	pane.SetBgOutputActive(false)

	if err := pane.ApplyFilter("f00", SearchModeIncremental); err != nil {
		t.Fatalf("ApplyFilter: %v", err)
	}
	afterFilter := pane.hitLayout()
	if afterFilter.entryCount == before.entryCount {
		t.Errorf("hitLayout() did not follow filter change: %+v", afterFilter)
	}
}

// --- AC-3: clampedEntryIndex and the Model adapters' absent-pane defense ---

func TestClampedEntryIndex(t *testing.T) {
	layout := paneHitLayout{scrollOffset: 3, visibleLines: 5, entryCount: 10} // maxIndex=7

	tests := []struct {
		name      string
		y         int
		wantIndex int
		wantOK    bool
	}{
		{"above first entry row: zero", 0, 3, true},
		{"above first entry row: negative", -5, 3, true},
		{"below last displayed row: far beyond", 100, 7, true},
		{"row inside the list", 6, 5, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx, ok := clampedEntryIndex(tt.y, layout)
			if ok != tt.wantOK || idx != tt.wantIndex {
				t.Errorf("clampedEntryIndex(%d) = (%d,%v), want (%d,%v)", tt.y, idx, ok, tt.wantIndex, tt.wantOK)
			}
		})
	}
}

func TestClampedEntryIndex_BlankRowInShortListClampsToLastEntry(t *testing.T) {
	layout := paneHitLayout{scrollOffset: 0, visibleLines: 18, entryCount: 3}
	idx, ok := clampedEntryIndex(8, layout) // blank row well past the 3 entries
	if !ok || idx != 2 {
		t.Errorf("clampedEntryIndex(8) = (%d,%v), want (2,true)", idx, ok)
	}
}

func TestClampedEntryIndex_NothingDisplayed(t *testing.T) {
	cases := []paneHitLayout{
		{scrollOffset: 5, visibleLines: 5, entryCount: 5},  // entryCount <= scrollOffset
		{scrollOffset: 0, visibleLines: 0, entryCount: 10}, // visibleLines <= 0
	}
	for _, layout := range cases {
		if _, ok := clampedEntryIndex(4, layout); ok {
			t.Errorf("clampedEntryIndex with %+v: got ok=true, want false", layout)
		}
	}
}

func TestModelAdapters_AbsentPane_NoPanicNoIndex(t *testing.T) {
	m := &Model{}

	hit := m.mouseHitAt(1, 1)
	if hit.kind != hitNone {
		t.Errorf("mouseHitAt with absent panes = %+v, want hitNone", hit)
	}

	if _, ok := m.mouseDragIndexAt(LeftPane, 5); ok {
		t.Error("mouseDragIndexAt with absent pane: got ok=true, want false")
	}
}

// --- AC-4: the shared pane header row count ---

func TestMouseFirstEntryRow_EqualsTitleRowPlusSharedHeaderRows(t *testing.T) {
	if paneHeaderRows != 3 {
		t.Fatalf("paneHeaderRows = %d, want 3", paneHeaderRows)
	}
	if mouseFirstEntryRow != mouseTitleRow+1+paneHeaderRows {
		t.Errorf("mouseFirstEntryRow = %d, want %d", mouseFirstEntryRow, mouseTitleRow+1+paneHeaderRows)
	}
}

func TestVisibleLinesAndBgSplitHeight_UnchangedAtGivenHeights(t *testing.T) {
	// Pinned values from before this change: getVisibleLines() and the bg
	// split file-list height must stay identical now that both derive from
	// the shared paneHeaderRows definition instead of a repeated literal.
	cases := []struct {
		height               int
		wantVisible          int
		wantBgFileListHeight int
	}{
		{22, 18, 11},
		{26, 22, 14},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("height=%d", c.height), func(t *testing.T) {
			pane := newFilesPane(t, LeftPane, 30, 40, c.height, true)

			if got := pane.getVisibleLines(); got != c.wantVisible {
				t.Errorf("getVisibleLines() = %d, want %d", got, c.wantVisible)
			}

			pane.SetBgOutputActive(true)
			fileListHeight, _ := pane.bgSplitHeights()
			if fileListHeight != c.wantBgFileListHeight {
				t.Errorf("bgSplitHeights() fileListHeight = %d, want %d", fileListHeight, c.wantBgFileListHeight)
			}
		})
	}
}

func TestViewDimmed_LineCount_MatchesSharedHeaderRowsPlusEntryRows(t *testing.T) {
	const height = 22
	pane := newFilesPane(t, LeftPane, 30, 40, height, true)

	rendered := pane.ViewDimmedWithDiskSpace(0)
	lines := strings.Split(rendered, "\n")
	// The rendered output ends with a trailing "\n", which Split turns into
	// an empty trailing element; drop it before counting.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	want := paneHeaderRows + (height - paneHeaderRows - 1)
	if len(lines) != want {
		t.Errorf("dimmed view line count = %d, want %d", len(lines), want)
	}
}

// --- AC-5: render / hit-test row alignment across all three views, for a
// long, no-branch path that would wrap header line 1 before the fix ---

func TestRenderAndHitTest_LongPathNoBranch_AllViewsAlign(t *testing.T) {
	const paneWidth, paneHeight = 40, 22
	const modelWidth, modelHeight = paneWidth * 2, paneHeight + 2

	dir := t.TempDir()
	nested := dir
	for i := 0; i < 20; i++ {
		nested = filepath.Join(nested, fmt.Sprintf("nested-directory-segment-%02d", i))
	}
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	names := []string{"alpha-file", "bravo-file", "charlie-file"}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(nested, name), nil, 0644); err != nil {
			t.Fatalf("write file %s: %v", name, err)
		}
	}

	pane, err := NewPane(LeftPane, nested, paneWidth, paneHeight, true, nil)
	if err != nil {
		t.Fatalf("NewPane: %v", err)
	}
	pane.gitBranch = "" // do not depend on whether the temp root is inside a git repo
	other := paneHitLayout{}

	findLine := func(t *testing.T, rendered, name string) int {
		t.Helper()
		lines := strings.Split(stripANSI(rendered), "\n")
		for i, line := range lines {
			if strings.Contains(line, name) {
				return i
			}
		}
		t.Fatalf("entry %q not found in rendered output:\n%s", name, rendered)
		return -1
	}

	checkView := func(t *testing.T, rendered string, layout paneHitLayout) {
		t.Helper()
		for _, name := range names {
			idx := pane.findEntryIndex(name) // ".." occupies index0, so this is not the names slice position
			if idx < 0 {
				t.Fatalf("entry %q not found in pane.entries", name)
			}
			line := findLine(t, rendered, name)
			if line != paneHeaderRows+idx {
				t.Errorf("entry %q: line = %d, want %d (paneHeaderRows + %d)", name, line, paneHeaderRows+idx, idx)
			}
			screenRow := mouseTitleRow + 1 + line
			hit := hitTest(5, screenRow, modelWidth, modelHeight, layout, other)
			if hit.kind != hitEntry || hit.index != idx {
				t.Errorf("entry %q: hitTest at row %d = %+v, want hitEntry index=%d", name, screenRow, hit, idx)
			}
		}
	}

	t.Run("normal view", func(t *testing.T) {
		checkView(t, pane.View(), pane.hitLayout())
	})

	t.Run("dimmed view", func(t *testing.T) {
		checkView(t, pane.ViewDimmedWithDiskSpace(0), pane.hitLayout())
	})

	t.Run("bg output split view", func(t *testing.T) {
		pane.SetBgOutputActive(true)
		defer pane.SetBgOutputActive(false)
		buf := NewOutputBuffer(100)
		checkView(t, pane.ViewWithBgOutput(0, buf, "echo hi", false), pane.hitLayout())
	})
}
