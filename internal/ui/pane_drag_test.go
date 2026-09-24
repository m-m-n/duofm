package ui

import (
	"testing"

	"github.com/sakura/duofm/internal/fs"
)

// assertMarks fails the test unless pane.markedFiles contains exactly the
// given names.
func assertMarks(t *testing.T, p *Pane, want ...string) {
	t.Helper()
	wantSet := make(map[string]bool, len(want))
	for _, w := range want {
		wantSet[w] = true
	}
	if len(p.markedFiles) != len(wantSet) {
		t.Fatalf("marks = %v, want %v", p.markedFiles, wantSet)
	}
	for name := range wantSet {
		if !p.markedFiles[name] {
			t.Errorf("marks = %v, want %s marked", p.markedFiles, name)
		}
	}
}

// --- AC-4: applyDragMarks ---

func TestApplyDragMarks_GrowsOverBaseline(t *testing.T) {
	pane := newFilesPane(t, LeftPane, 8, 40, 22, true) // entries: "..",f00..f07 (indices 0..8)
	baseline := map[string]bool{pane.entries[7].Name: true}

	ok := pane.applyDragMarks(1, 4, baseline)
	if !ok {
		t.Fatal("applyDragMarks(1,4) = false, want true")
	}
	assertMarks(t, pane, pane.entries[1].Name, pane.entries[2].Name, pane.entries[3].Name, pane.entries[4].Name, pane.entries[7].Name)
	if pane.cursor != 4 {
		t.Errorf("cursor = %d, want 4", pane.cursor)
	}
}

func TestApplyDragMarks_RecomputesFromSameBaselineEachCall(t *testing.T) {
	pane := newFilesPane(t, LeftPane, 8, 40, 22, true)
	baseline := map[string]bool{pane.entries[2].Name: true}

	if ok := pane.applyDragMarks(1, 5, baseline); !ok {
		t.Fatal("applyDragMarks(1,5) = false, want true")
	}
	// Same baseline reused for a second call, simulating the drag shrinking
	// back toward the anchor: the result must not accumulate from the
	// first call.
	if ok := pane.applyDragMarks(1, 2, baseline); !ok {
		t.Fatal("applyDragMarks(1,2) = false, want true")
	}
	assertMarks(t, pane, pane.entries[1].Name, pane.entries[2].Name)
}

func TestApplyDragMarks_ReverseAnchorTarget(t *testing.T) {
	pane := newFilesPane(t, LeftPane, 8, 40, 22, true)

	ok := pane.applyDragMarks(5, 3, map[string]bool{})
	if !ok {
		t.Fatal("applyDragMarks(5,3) = false, want true")
	}
	assertMarks(t, pane, pane.entries[3].Name, pane.entries[4].Name, pane.entries[5].Name)
	if pane.cursor != 3 {
		t.Errorf("cursor = %d, want 3 (target, not the lower bound)", pane.cursor)
	}
}

func TestApplyDragMarks_ExcludesParentDir(t *testing.T) {
	pane := newFilesPane(t, LeftPane, 8, 40, 22, true)
	if !pane.entries[0].IsParentDir() {
		t.Fatalf("entries[0] = %q, want \"..\"", pane.entries[0].Name)
	}

	ok := pane.applyDragMarks(0, 3, map[string]bool{})
	if !ok {
		t.Fatal("applyDragMarks(0,3) = false, want true")
	}
	assertMarks(t, pane, pane.entries[1].Name, pane.entries[2].Name, pane.entries[3].Name)
	if pane.IsMarked("..") {
		t.Error("\"..\" must never be marked")
	}
}

func TestApplyDragMarks_ScrollOffsetNeverTouched(t *testing.T) {
	pane := newFilesPane(t, LeftPane, 30, 40, 22, true)
	pane.scrollOffset = 5

	if ok := pane.applyDragMarks(6, 10, map[string]bool{}); !ok {
		t.Fatal("applyDragMarks(6,10) = false, want true")
	}
	if pane.scrollOffset != 5 {
		t.Errorf("scrollOffset = %d, want unchanged at 5", pane.scrollOffset)
	}

	// Target on the window's edge, still with a non-zero offset.
	visible := pane.getVisibleLines()
	edge := pane.scrollOffset + visible - 1
	if ok := pane.applyDragMarks(6, edge, map[string]bool{}); !ok {
		t.Fatal("applyDragMarks at window edge = false, want true")
	}
	if pane.scrollOffset != 5 {
		t.Errorf("scrollOffset = %d, want unchanged at 5", pane.scrollOffset)
	}
}

func TestApplyDragMarks_OutOfRangeReturnsFalseUnchanged(t *testing.T) {
	pane := newFilesPane(t, LeftPane, 3, 40, 22, true) // entries 0..3
	pane.markedFiles[pane.entries[1].Name] = true
	pane.cursor = 1

	before := pane.snapshotMarks()

	if ok := pane.applyDragMarks(-1, 2, map[string]bool{}); ok {
		t.Error("applyDragMarks(-1,2) = true, want false")
	}
	if ok := pane.applyDragMarks(2, 99, map[string]bool{}); ok {
		t.Error("applyDragMarks(2,99) = true, want false")
	}
	if pane.cursor != 1 {
		t.Errorf("cursor changed to %d after a rejected call, want unchanged 1", pane.cursor)
	}
	assertMarks(t, pane, markedNames(before)...)
}

func TestApplyDragMarks_EmptyEntriesReturnsFalse(t *testing.T) {
	pane := &Pane{entries: []fs.FileEntry{}, markedFiles: make(map[string]bool)}
	if ok := pane.applyDragMarks(0, 0, map[string]bool{}); ok {
		t.Error("applyDragMarks on empty entries = true, want false")
	}
	if len(pane.markedFiles) != 0 {
		t.Errorf("markedFiles = %v, want empty", pane.markedFiles)
	}
}

// --- AC-4: snapshotMarks independence ---

func TestSnapshotMarks_IndependentBothDirections(t *testing.T) {
	pane := newFilesPane(t, LeftPane, 5, 40, 22, true)
	pane.markedFiles[pane.entries[1].Name] = true

	snapshot := pane.snapshotMarks()

	// Mutating the pane afterward must not affect the snapshot.
	pane.markedFiles[pane.entries[2].Name] = true
	if snapshot[pane.entries[2].Name] {
		t.Error("snapshot changed after mutating pane.markedFiles")
	}

	// Mutating the snapshot afterward must not affect the pane.
	snapshot[pane.entries[3].Name] = true
	if pane.markedFiles[pane.entries[3].Name] {
		t.Error("pane.markedFiles changed after mutating the snapshot")
	}
}

func TestApplyDragMarks_NeverMutatesBaseline(t *testing.T) {
	pane := newFilesPane(t, LeftPane, 8, 40, 22, true)
	baseline := map[string]bool{pane.entries[7].Name: true}
	baselineCopy := map[string]bool{pane.entries[7].Name: true}

	if ok := pane.applyDragMarks(1, 4, baseline); !ok {
		t.Fatal("applyDragMarks = false, want true")
	}

	if len(baseline) != len(baselineCopy) {
		t.Fatalf("baseline was mutated: %v, want %v", baseline, baselineCopy)
	}
	for k, v := range baselineCopy {
		if baseline[k] != v {
			t.Errorf("baseline[%s] = %v, want %v (baseline must never be mutated)", k, baseline[k], v)
		}
	}
}

// markedNames converts a mark-set snapshot into a slice of marked names,
// for use with assertMarks.
func markedNames(marks map[string]bool) []string {
	names := make([]string, 0, len(marks))
	for name, marked := range marks {
		if marked {
			names = append(names, name)
		}
	}
	return names
}
