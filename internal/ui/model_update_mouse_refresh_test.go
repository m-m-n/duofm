package ui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// task0001: cancel a drag session when its pane's displayed list changes
// (SPEC TS-1..TS-8). These tests are driven through the Model's update entry
// point with temp-directory panes and real refresh messages, and assert only
// through the pre-existing dragSession fields (armed, active, pane, anchor,
// baseline) and the pane's mark set -- never through the new Pane helpers or
// the new session field -- so this file alone compiles against the pre-fix
// code (SPEC AC8: fail before the fix, pass after for AC-1, AC-2 and AC-6;
// AC-3, AC-4 and AC-5-style checks pass both before and after).

// newDragRefreshTestModel builds the same fixture as newDragLoadTestModel
// (disk-space monitor initialized, avoiding a nil dereference in the
// refresh paths) plus a positive refresh rate, since autoRefreshMsg is a
// no-op otherwise.
func newDragRefreshTestModel(t *testing.T, leftN, rightN int) Model {
	t.Helper()
	m := newDragLoadTestModel(t, leftN, rightN)
	m.refreshRate = 60
	return m
}

// AC-1 (TS-1): a session armed by a press (no motion) is cancelled at the
// next motion once the left pane's displayed list has grown via an
// auto-refresh; the release leaves marks exactly as captured right after the
// refresh.
func TestHandleMouse_DragRefresh_CancelledWhenFileAddedBeforeMotion(t *testing.T) {
	model := newDragRefreshTestModel(t, 20, 20)
	const anchor = 2

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(anchor, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	if !model.mouseDrag.armed || model.mouseDrag.active {
		t.Fatalf("session after press = %+v, want armed and not active", model.mouseDrag)
	}

	if err := os.WriteFile(filepath.Join(model.leftPane.Path(), "added-file"), nil, 0644); err != nil {
		t.Fatalf("write added-file: %v", err)
	}

	updated, _ = model.Update(autoRefreshMsg{})
	model = updated.(Model)

	wantMarks := model.leftPane.snapshotMarks()

	const motion = 5
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motion, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	if model.mouseDrag.armed || model.mouseDrag.active {
		t.Fatalf("session after motion = %+v, want neither armed nor active", model.mouseDrag)
	}

	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motion, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	assertMarksEqual(t, model.leftPane.markedFiles, wantMarks)
}

// AC-2 (TS-2): a session already active (anchor + motion marked a range) is
// cancelled once an auto-refresh removes an unmarked file from the left
// pane's directory; the mark set right after the refresh equals the set
// right before it, and a later motion/release leave it unchanged while the
// session ends up neither armed nor active.
func TestHandleMouse_DragRefresh_CancelledWhenUnmarkedFileDeletedDuringActiveDrag(t *testing.T) {
	model := newDragRefreshTestModel(t, 20, 20)
	const anchor, motionA = 2, 6

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(anchor, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motionA, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	if !model.mouseDrag.active {
		t.Fatal("session should be active after motion")
	}

	// The last entry (alphabetically last, outside anchor..motionA) is not
	// marked by the anchor..motionA range.
	unmarkedName := model.leftPane.entries[len(model.leftPane.entries)-1].Name
	if err := os.Remove(filepath.Join(model.leftPane.Path(), unmarkedName)); err != nil {
		t.Fatalf("remove: %v", err)
	}

	beforeRefresh := model.leftPane.snapshotMarks()
	if len(beforeRefresh) == 0 {
		t.Fatal("expected marks from anchor..motionA before refresh")
	}

	updated, _ = model.Update(autoRefreshMsg{})
	model = updated.(Model)
	assertMarksEqual(t, model.leftPane.markedFiles, beforeRefresh)

	const motionB = 4
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motionB, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	if model.mouseDrag.armed || model.mouseDrag.active {
		t.Fatalf("session after motion = %+v, want neither armed nor active", model.mouseDrag)
	}
	assertMarksEqual(t, model.leftPane.markedFiles, beforeRefresh)

	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motionB, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)
	assertMarksEqual(t, model.leftPane.markedFiles, beforeRefresh)
}

// AC-3 (TS-3, TS-6, TS-7): when the left pane's displayed list is unchanged
// by an auto-refresh -- no change at all, an existing file's content (size)
// changes with names/order unchanged, or a filter hides an added file -- the
// session survives the refresh (still armed, same pane and anchor), and a
// later motion marks exactly the press-time marks plus the displayed range
// from the anchor to the motion row (parent excluded).
func TestHandleMouse_DragRefresh_SurvivesWhenDisplayedListUnchanged(t *testing.T) {
	t.Run("no file change", func(t *testing.T) {
		model := newDragRefreshTestModel(t, 20, 20)
		runDragSurvivesRefreshCase(t, model, 2, 6, 15, func(m *Model) {})
	})

	t.Run("existing file content changes, names and order unchanged", func(t *testing.T) {
		model := newDragRefreshTestModel(t, 20, 20)
		runDragSurvivesRefreshCase(t, model, 2, 6, 15, func(m *Model) {
			target := filepath.Join(m.leftPane.Path(), m.leftPane.entries[10].Name)
			if err := os.WriteFile(target, []byte("more content now"), 0644); err != nil {
				t.Fatalf("rewrite file: %v", err)
			}
		})
	})

	t.Run("filter applied, added file does not match filter", func(t *testing.T) {
		model := newDragRefreshTestModel(t, 20, 20)
		if err := model.leftPane.ApplyFilter("f0", SearchModeIncremental); err != nil {
			t.Fatalf("ApplyFilter: %v", err)
		}
		// Filtered entries: "f00".."f09" (10 entries, indices 0..9); "f10".."f19"
		// do not contain "f0" and are excluded, along with "..".
		runDragSurvivesRefreshCase(t, model, 1, 4, 8, func(m *Model) {
			target := filepath.Join(m.leftPane.Path(), "zzz-no-match")
			if err := os.WriteFile(target, nil, 0644); err != nil {
				t.Fatalf("write file: %v", err)
			}
		})
	})
}

// runDragSurvivesRefreshCase marks entries[preMarkIndex] before the press,
// presses at anchor, applies mutate to the left pane's directory, processes
// an auto-refresh message, confirms the session is still armed with the
// same pane and anchor, then moves to motion and confirms the resulting
// mark set is exactly the press-time marks plus the displayed range from
// anchor to motion (parent excluded).
func runDragSurvivesRefreshCase(t *testing.T, model Model, anchor, motion, preMarkIndex int, mutate func(m *Model)) {
	t.Helper()
	preMarkName := model.leftPane.entries[preMarkIndex].Name
	model.leftPane.markedFiles[preMarkName] = true

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(anchor, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	wantBaseline := model.leftPane.snapshotMarks()

	mutate(&model)

	updated, _ = model.Update(autoRefreshMsg{})
	model = updated.(Model)

	if !model.mouseDrag.armed || model.mouseDrag.active {
		t.Fatalf("session after refresh = %+v, want armed and not active", model.mouseDrag)
	}
	if model.mouseDrag.pane != LeftPane || model.mouseDrag.anchor != anchor {
		t.Fatalf("session pane/anchor after refresh = %+v, want pane=LeftPane anchor=%d", model.mouseDrag, anchor)
	}

	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motion, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motion, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	wantMarks := make(map[string]bool, len(wantBaseline)+1)
	for name, marked := range wantBaseline {
		wantMarks[name] = marked
	}
	lo, hi := anchor, motion
	if lo > hi {
		lo, hi = hi, lo
	}
	for i := lo; i <= hi; i++ {
		entry := model.leftPane.entries[i]
		if entry.IsParentDir() {
			continue
		}
		wantMarks[entry.Name] = true
	}
	assertMarksEqual(t, model.leftPane.markedFiles, wantMarks)
}

// AC-4 (TS-4): a change to only the RIGHT pane's directory leaves an armed
// left-pane session's armed, active, pane, anchor and baseline unchanged
// across an auto-refresh; a later left-pane motion marks exactly the range
// from the original anchor to the motion row.
func TestHandleMouse_DragRefresh_UnaffectedWhenOnlyOtherPaneChanges(t *testing.T) {
	model := newDragRefreshTestModel(t, 20, 20)
	const anchor = 3

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(anchor, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	wantSession := model.mouseDrag
	wantBaseline := model.mouseDrag.baseline

	if err := os.WriteFile(filepath.Join(model.rightPane.Path(), "right-added"), nil, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	updated, _ = model.Update(autoRefreshMsg{})
	model = updated.(Model)

	if !model.mouseDrag.armed || model.mouseDrag.active {
		t.Fatalf("session after refresh = %+v, want armed and not active", model.mouseDrag)
	}
	if model.mouseDrag.pane != wantSession.pane || model.mouseDrag.anchor != wantSession.anchor {
		t.Fatalf("session pane/anchor changed: got %+v, want %+v", model.mouseDrag, wantSession)
	}
	assertMarksEqual(t, model.mouseDrag.baseline, wantBaseline)

	const motion = 7
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motion, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motion, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	var wantNames []string
	for i := anchor; i <= motion; i++ {
		wantNames = append(wantNames, model.leftPane.entries[i].Name)
	}
	assertMarks(t, model.leftPane, wantNames...)
}

// AC-5 (TS-5): the same setup as AC-1, refreshed once through an exec
// completion message and once through a shell command completion message
// instead of the auto-refresh message: after the motion and release, the
// left pane's mark set equals the set captured right after the refresh.
func TestHandleMouse_DragRefresh_CancelledByExecOrShellCompletion(t *testing.T) {
	cases := []struct {
		name string
		msg  tea.Msg
	}{
		{"exec completion", execFinishedMsg{}},
		{"shell command completion", shellCommandFinishedMsg{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			model := newDragRefreshTestModel(t, 20, 20)
			const anchor = 2

			updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(anchor, 0), tea.MouseButtonLeft, tea.MouseActionPress))
			model = updated.(Model)

			if err := os.WriteFile(filepath.Join(model.leftPane.Path(), "added-file"), nil, 0644); err != nil {
				t.Fatalf("write added-file: %v", err)
			}

			updated, _ = model.Update(c.msg)
			model = updated.(Model)

			wantMarks := model.leftPane.snapshotMarks()

			const motion = 5
			updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motion, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
			model = updated.(Model)
			if model.mouseDrag.armed || model.mouseDrag.active {
				t.Fatalf("session after motion = %+v, want neither armed nor active", model.mouseDrag)
			}

			updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motion, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
			model = updated.(Model)

			assertMarksEqual(t, model.leftPane.markedFiles, wantMarks)
		})
	}
}

// AC-6 (TS-8): two files marked and then hidden by an applied filter; the
// press happens inside the filtered list, one of the two hidden files is
// deleted on disk (which does not change the displayed list, since it was
// already hidden), and an auto-refresh is processed. After a motion and
// release, the deleted file's name is not in the mark set, the other hidden
// file is still marked, and the anchor..motion range is marked.
func TestHandleMouse_DragRefresh_RestrictsBaselineToExistingHiddenMarks(t *testing.T) {
	model := newDragRefreshTestModel(t, 20, 20)

	hiddenKeepName := model.leftPane.entries[15].Name   // "f14": hidden by filter below, survives
	hiddenDeleteName := model.leftPane.entries[16].Name // "f15": hidden by filter below, deleted
	model.leftPane.markedFiles[hiddenKeepName] = true
	model.leftPane.markedFiles[hiddenDeleteName] = true

	if err := model.leftPane.ApplyFilter("f0", SearchModeIncremental); err != nil {
		t.Fatalf("ApplyFilter: %v", err)
	}

	const anchor = 1
	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(anchor, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)

	if err := os.Remove(filepath.Join(model.leftPane.Path(), hiddenDeleteName)); err != nil {
		t.Fatalf("remove: %v", err)
	}

	updated, _ = model.Update(autoRefreshMsg{})
	model = updated.(Model)

	if !model.mouseDrag.armed || model.mouseDrag.active {
		t.Fatalf("session after refresh = %+v, want armed and not active", model.mouseDrag)
	}

	const motion = 4
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motion, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motion, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	if model.leftPane.markedFiles[hiddenDeleteName] {
		t.Error("deleted file's name must not be in the mark set")
	}
	if !model.leftPane.markedFiles[hiddenKeepName] {
		t.Error("the other hidden file must still be marked")
	}
	for i := anchor; i <= motion; i++ {
		name := model.leftPane.entries[i].Name
		if !model.leftPane.markedFiles[name] {
			t.Errorf("entry %q in anchor..motion range must be marked", name)
		}
	}
}
