package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Geometry shared by these tests: width=80, height=24 -> paneWidth=40,
// paneHeight=22, mouseFirstEntryRow=4, visibleLines=18 (no bg split).
const (
	mouseTestLeftX  = 5
	mouseTestRightX = 45
)

func mouseMsg(x, y int, button tea.MouseButton, action tea.MouseAction) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Button: button, Action: action}
}

func rowFor(index, scrollOffset int) int {
	return mouseFirstEntryRow + index - scrollOffset
}

// newMouseTestModel builds a ready Model with two independent panes over
// generated files (see newFilesPane), at the shared 80x24 geometry.
func newMouseTestModel(t *testing.T, leftN, rightN int) Model {
	t.Helper()
	const width, height = 80, 24
	paneWidth, paneHeight := width/2, height-2
	left := newFilesPane(t, LeftPane, leftN, paneWidth, paneHeight, true)
	right := newFilesPane(t, RightPane, rightN, paneWidth, paneHeight, false)
	return Model{
		leftPane:   left,
		rightPane:  right,
		activePane: LeftPane,
		width:      width,
		height:     height,
		detector:   newDoubleClickDetector(nil),
	}
}

// newDoubleClickTestModel builds a Model whose left/right panes both list
// the same directory, containing exactly ".." (index0), "subdir" (index1,
// a directory) and "file.txt" (index2, a regular file) -- per SortEntries'
// parent-then-dirs-then-files ordering. The detector uses clock.
func newDoubleClickTestModel(t *testing.T, clock *fakeClock) Model {
	t.Helper()
	const width, height = 80, 24
	paneWidth, paneHeight := width/2, height-2

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	left, err := NewPane(LeftPane, dir, paneWidth, paneHeight, true, nil)
	if err != nil {
		t.Fatalf("NewPane left: %v", err)
	}
	right, err := NewPane(RightPane, dir, paneWidth, paneHeight, false, nil)
	if err != nil {
		t.Fatalf("NewPane right: %v", err)
	}

	return Model{
		leftPane:   left,
		rightPane:  right,
		activePane: LeftPane,
		width:      width,
		height:     height,
		detector:   newDoubleClickDetector(clock.now),
	}
}

func assertMarksEqual(t *testing.T, got, want map[string]bool) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("marks = %v, want %v", got, want)
	}
	for name, marked := range want {
		if got[name] != marked {
			t.Errorf("marks[%s] = %v, want %v", name, got[name], marked)
		}
	}
}

// --- AC-6: wiring ---

func TestHandleMouse_PressViaUpdateMovesCursor(t *testing.T) {
	model := newMouseTestModel(t, 20, 20)

	updated, cmd := model.Update(mouseMsg(mouseTestLeftX, rowFor(2, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	m := updated.(Model)

	if m.leftPane.cursor != 2 {
		t.Errorf("cursor = %d, want 2", m.leftPane.cursor)
	}
	if cmd != nil {
		t.Errorf("cmd = %v, want nil", cmd)
	}
}

func TestHandleMouse_BeforePanesExist_NoPanicNilCmd(t *testing.T) {
	model := Model{}

	updated, cmd := model.Update(mouseMsg(5, 5, tea.MouseButtonLeft, tea.MouseActionPress))
	m := updated.(Model)

	if cmd != nil {
		t.Errorf("cmd = %v, want nil", cmd)
	}
	if m.leftPane != nil || m.rightPane != nil {
		t.Error("panes should remain nil")
	}
}

func TestNewModel_InitializesDetector(t *testing.T) {
	m := NewModel()
	if m.detector == nil {
		t.Error("NewModel() did not initialize the double-click detector")
	}
}

// --- AC-7: modal gating and other buttons ---

func TestHandleMouse_ModalStates_IgnoreGesture(t *testing.T) {
	setups := map[string]func(m *Model){
		"dialog present":            func(m *Model) { m.dialog = NewErrorDialog("boom") },
		"sort dialog active":        func(m *Model) { m.sortDialog = NewSortDialog(DefaultSortConfig()) },
		"incremental search active": func(m *Model) { m.searchState.IsActive = true },
		"shell command mode":        func(m *Model) { m.shellCommandMode = true },
		"bg output focused":         func(m *Model) { m.bgOutputFocused = true },
	}

	for name, setup := range setups {
		t.Run(name, func(t *testing.T) {
			model := newMouseTestModel(t, 20, 20)
			setup(&model)

			wantActivePane := model.activePane
			wantLeftCursor := model.leftPane.cursor
			wantRightCursor := model.rightPane.cursor
			wantLeftMarks := model.leftPane.snapshotMarks()
			wantRightMarks := model.rightPane.snapshotMarks()

			msgs := []tea.MouseMsg{
				mouseMsg(mouseTestLeftX, rowFor(1, 0), tea.MouseButtonLeft, tea.MouseActionPress),
				mouseMsg(mouseTestLeftX, rowFor(3, 0), tea.MouseButtonLeft, tea.MouseActionMotion),
				mouseMsg(mouseTestLeftX, rowFor(3, 0), tea.MouseButtonLeft, tea.MouseActionRelease),
			}
			for _, msg := range msgs {
				updated, cmd := model.Update(msg)
				model = updated.(Model)
				if cmd != nil {
					t.Errorf("cmd = %v, want nil", cmd)
				}
			}

			if model.activePane != wantActivePane {
				t.Errorf("activePane changed: got %v, want %v", model.activePane, wantActivePane)
			}
			if model.leftPane.cursor != wantLeftCursor || model.rightPane.cursor != wantRightCursor {
				t.Errorf("cursor changed: left=%d right=%d, want left=%d right=%d",
					model.leftPane.cursor, model.rightPane.cursor, wantLeftCursor, wantRightCursor)
			}
			assertMarksEqual(t, model.leftPane.markedFiles, wantLeftMarks)
			assertMarksEqual(t, model.rightPane.markedFiles, wantRightMarks)
		})
	}
}

func TestHandleMouse_ArmedSessionDisarmedByModal(t *testing.T) {
	model := newMouseTestModel(t, 20, 20)

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(1, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	if !model.mouseDrag.armed {
		t.Fatal("session not armed after press")
	}

	model.dialog = NewErrorDialog("boom")
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(4, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	if model.mouseDrag.armed {
		t.Fatal("session should be disarmed while a modal state is active")
	}

	model.dialog = nil
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(6, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	if model.leftPane.HasMarkedFiles() {
		t.Error("motion after the modal state ended should mark nothing")
	}
}

func TestHandleMouse_OtherButtons_NoEffect(t *testing.T) {
	buttons := []tea.MouseButton{tea.MouseButtonWheelUp, tea.MouseButtonWheelDown, tea.MouseButtonRight, tea.MouseButtonMiddle}
	for _, btn := range buttons {
		t.Run(fmt.Sprintf("button_%d", btn), func(t *testing.T) {
			model := newMouseTestModel(t, 20, 20)
			wantCursor := model.leftPane.cursor
			wantActive := model.activePane
			wantScroll := model.leftPane.scrollOffset

			updated, cmd := model.Update(mouseMsg(mouseTestLeftX, rowFor(2, 0), btn, tea.MouseActionPress))
			model = updated.(Model)

			if cmd != nil {
				t.Errorf("cmd = %v, want nil", cmd)
			}
			if model.leftPane.cursor != wantCursor {
				t.Errorf("cursor changed: %d, want %d", model.leftPane.cursor, wantCursor)
			}
			if model.activePane != wantActive {
				t.Errorf("activePane changed: %v, want %v", model.activePane, wantActive)
			}
			if model.leftPane.HasMarkedFiles() {
				t.Error("marks changed, want none")
			}
			if model.leftPane.scrollOffset != wantScroll {
				t.Errorf("scrollOffset changed: %d, want %d", model.leftPane.scrollOffset, wantScroll)
			}
		})
	}
}

// --- AC-8: click semantics ---

func TestHandleMouse_ClickMovesCursorKeepsMarks(t *testing.T) {
	model := newMouseTestModel(t, 20, 20)
	model.leftPane.markedFiles[model.leftPane.entries[1].Name] = true
	model.leftPane.markedFiles[model.leftPane.entries[5].Name] = true
	wantMarks := model.leftPane.snapshotMarks()

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(3, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(3, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	if model.leftPane.cursor != 3 {
		t.Errorf("cursor = %d, want 3", model.leftPane.cursor)
	}
	if model.activePane != LeftPane {
		t.Errorf("activePane = %v, want LeftPane", model.activePane)
	}
	assertMarksEqual(t, model.leftPane.markedFiles, wantMarks)
}

func TestHandleMouse_ClickActivatesInactivePane(t *testing.T) {
	model := newMouseTestModel(t, 20, 20)
	wantLeftCursor := model.leftPane.cursor

	updated, _ := model.Update(mouseMsg(mouseTestRightX, rowFor(2, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(2, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	if model.activePane != RightPane {
		t.Errorf("activePane = %v, want RightPane", model.activePane)
	}
	if model.rightPane.cursor != 2 {
		t.Errorf("right cursor = %d, want 2", model.rightPane.cursor)
	}
	if model.leftPane.cursor != wantLeftCursor {
		t.Errorf("left cursor changed: %d, want %d", model.leftPane.cursor, wantLeftCursor)
	}
}

func TestHandleMouse_PressReleaseNoMove_NoMarks(t *testing.T) {
	model := newMouseTestModel(t, 20, 20)

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(2, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(2, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	if model.leftPane.cursor != 2 {
		t.Errorf("cursor = %d, want 2", model.leftPane.cursor)
	}
	if model.leftPane.HasMarkedFiles() {
		t.Error("marks should be empty")
	}
}

func TestHandleMouse_ClickOnParentDir_CursorOnParent(t *testing.T) {
	model := newMouseTestModel(t, 5, 5)

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(0, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)

	if model.leftPane.cursor != 0 || !model.leftPane.entries[model.leftPane.cursor].IsParentDir() {
		t.Errorf("cursor = %d, want on parent dir (index 0)", model.leftPane.cursor)
	}
}

func TestHandleMouse_NonEntryClickOnRightPane_ActivatesOnly(t *testing.T) {
	for _, row := range []int{1, 2, 3, 8} {
		t.Run(fmt.Sprintf("row_%d", row), func(t *testing.T) {
			model := newMouseTestModel(t, 3, 3) // few entries: row 8 is a blank row
			wantCursor := model.rightPane.cursor
			wantMarks := model.rightPane.snapshotMarks()

			updated, cmd := model.Update(mouseMsg(mouseTestRightX, row, tea.MouseButtonLeft, tea.MouseActionPress))
			model = updated.(Model)

			if model.activePane != RightPane {
				t.Errorf("activePane = %v, want RightPane", model.activePane)
			}
			if model.rightPane.cursor != wantCursor {
				t.Errorf("cursor changed: %d, want %d", model.rightPane.cursor, wantCursor)
			}
			assertMarksEqual(t, model.rightPane.markedFiles, wantMarks)
			if cmd != nil {
				t.Errorf("cmd = %v, want nil", cmd)
			}
		})
	}
}

func TestHandleMouse_NoneHits_ChangeNothing(t *testing.T) {
	cases := []struct {
		name string
		x, y int
	}{
		{"title bar", mouseTestLeftX, 0},
		{"status bar", mouseTestLeftX, 23},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			model := newMouseTestModel(t, 20, 20)
			wantActive := model.activePane
			wantCursor := model.leftPane.cursor

			updated, cmd := model.Update(mouseMsg(c.x, c.y, tea.MouseButtonLeft, tea.MouseActionPress))
			model = updated.(Model)

			if model.activePane != wantActive || model.leftPane.cursor != wantCursor {
				t.Errorf("state changed: activePane=%v cursor=%d, want %v/%d",
					model.activePane, model.leftPane.cursor, wantActive, wantCursor)
			}
			if cmd != nil {
				t.Errorf("cmd = %v, want nil", cmd)
			}
		})
	}
}

func TestHandleMouse_OddWidthLeftoverColumn_NoEffect(t *testing.T) {
	const width, height = 81, 24
	paneWidth, paneHeight := width/2, height-2 // 40, 22
	left := newFilesPane(t, LeftPane, 20, paneWidth, paneHeight, true)
	right := newFilesPane(t, RightPane, 20, paneWidth, paneHeight, false)
	model := Model{
		leftPane: left, rightPane: right, activePane: LeftPane,
		width: width, height: height, detector: newDoubleClickDetector(nil),
	}

	wantCursor := model.leftPane.cursor
	updated, cmd := model.Update(mouseMsg(80, rowFor(2, 0), tea.MouseButtonLeft, tea.MouseActionPress)) // x=80 is the leftover column
	model = updated.(Model)

	if model.leftPane.cursor != wantCursor || model.activePane != LeftPane {
		t.Error("leftover column click changed state")
	}
	if cmd != nil {
		t.Errorf("cmd = %v, want nil", cmd)
	}
}

func TestHandleMouse_BgSplit_FileListClickMovesCursorSeparatorDoesNot(t *testing.T) {
	model := newMouseTestModel(t, 20, 20)
	model.leftPane.SetBgOutputActive(true)
	fileListHeight, _ := model.leftPane.bgSplitHeights()

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(1, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	if model.leftPane.cursor != 1 {
		t.Errorf("cursor = %d, want 1", model.leftPane.cursor)
	}

	wantCursor := model.leftPane.cursor
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, mouseFirstEntryRow+fileListHeight, tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	if model.leftPane.cursor != wantCursor {
		t.Errorf("separator click changed cursor: %d, want %d", model.leftPane.cursor, wantCursor)
	}
}

// --- AC-9: drag marking via mouse events ---

func TestHandleMouse_Drag_ExtendsRangeOverExistingMark(t *testing.T) {
	model := newMouseTestModel(t, 20, 20)
	model.leftPane.markedFiles[model.leftPane.entries[7].Name] = true

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(1, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(4, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(4, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	assertMarks(t, model.leftPane,
		model.leftPane.entries[1].Name, model.leftPane.entries[2].Name,
		model.leftPane.entries[3].Name, model.leftPane.entries[4].Name,
		model.leftPane.entries[7].Name,
	)
	if model.leftPane.cursor != 4 {
		t.Errorf("cursor = %d, want 4", model.leftPane.cursor)
	}
}

func TestHandleMouse_Drag_ShrinkBackRemovesAddedMarks(t *testing.T) {
	model := newMouseTestModel(t, 20, 20)
	model.leftPane.markedFiles[model.leftPane.entries[2].Name] = true

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(1, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(5, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(2, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(2, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	assertMarks(t, model.leftPane, model.leftPane.entries[1].Name, model.leftPane.entries[2].Name)
}

func TestHandleMouse_Drag_ReverseDirection(t *testing.T) {
	model := newMouseTestModel(t, 20, 20)

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(5, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(3, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(3, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	assertMarks(t, model.leftPane, model.leftPane.entries[3].Name, model.leftPane.entries[4].Name, model.leftPane.entries[5].Name)
}

func TestHandleMouse_Drag_FromParentDir(t *testing.T) {
	model := newMouseTestModel(t, 20, 20)

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(0, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(3, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(3, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	assertMarks(t, model.leftPane, model.leftPane.entries[1].Name, model.leftPane.entries[2].Name, model.leftPane.entries[3].Name)
	if model.leftPane.IsMarked("..") {
		t.Error("\"..\" must never be marked")
	}
}

func TestHandleMouse_Drag_ClampsAtScrolledWindowEdgesAndIgnoresColumn(t *testing.T) {
	model := newMouseTestModel(t, 30, 5)
	model.leftPane.scrollOffset = 5
	visible := model.leftPane.getVisibleLines()
	minIdx, maxIdx := 5, 5+visible-1
	anchor := minIdx + 2 // away from both edges, so each clamp below differs from it

	rightCursorBefore := model.rightPane.cursor
	rightMarksBefore := model.rightPane.snapshotMarks()

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(anchor, 5), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)

	// Above the first entry row, onto the title bar: clamps to the first visible entry.
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, 0, tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	assertMarks(t, model.leftPane, model.leftPane.entries[minIdx].Name, model.leftPane.entries[minIdx+1].Name, model.leftPane.entries[anchor].Name)

	// Below the last visible row, onto the status bar, using the OTHER
	// pane's x column: the pointer column is never read during a drag.
	updated, _ = model.Update(mouseMsg(mouseTestRightX, 23, tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestRightX, 23, tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	if model.leftPane.scrollOffset != 5 {
		t.Errorf("scrollOffset changed: %d, want 5", model.leftPane.scrollOffset)
	}
	if model.rightPane.cursor != rightCursorBefore {
		t.Errorf("right pane cursor changed: %d, want %d", model.rightPane.cursor, rightCursorBefore)
	}
	assertMarksEqual(t, model.rightPane.markedFiles, rightMarksBefore)

	// Final range spans from the anchor to the last visible entry.
	var names []string
	for i := anchor; i <= maxIdx; i++ {
		names = append(names, model.leftPane.entries[i].Name)
	}
	assertMarks(t, model.leftPane, names...)
}

func TestHandleMouse_Drag_ActivatesInactivePaneFirst(t *testing.T) {
	model := newMouseTestModel(t, 20, 20)

	updated, _ := model.Update(mouseMsg(mouseTestRightX, rowFor(1, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(4, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(4, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	if model.activePane != RightPane {
		t.Errorf("activePane = %v, want RightPane", model.activePane)
	}
	assertMarks(t, model.rightPane,
		model.rightPane.entries[1].Name, model.rightPane.entries[2].Name,
		model.rightPane.entries[3].Name, model.rightPane.entries[4].Name,
	)
}

// --- AC-10: double-click via mouse events, on an injected fake clock ---

func TestHandleMouse_DoubleClick_OnDirectory_StartsLoading(t *testing.T) {
	clock := newFakeClock()
	model := newDoubleClickTestModel(t, clock)
	const dirEntryIndex = 1
	wantPath := filepath.Join(model.leftPane.Path(), "subdir")

	updated, cmd1 := model.Update(mouseMsg(mouseTestLeftX, rowFor(dirEntryIndex, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	if cmd1 != nil {
		t.Errorf("first press cmd = %v, want nil", cmd1)
	}

	clock.advance(200 * time.Millisecond)
	updated, cmd2 := model.Update(mouseMsg(mouseTestLeftX, rowFor(dirEntryIndex, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)

	if cmd2 == nil {
		t.Fatal("second press cmd = nil, want non-nil (Enter on directory)")
	}
	if model.leftPane.Path() != wantPath {
		t.Errorf("pane path = %q, want %q", model.leftPane.Path(), wantPath)
	}
	if !model.leftPane.IsLoading() {
		t.Error("pane should be loading the new directory")
	}
}

func TestHandleMouse_DoubleClick_OnParentDir_RemembersSubdirAndLoads(t *testing.T) {
	clock := newFakeClock()
	model := newDoubleClickTestModel(t, clock)
	originalDir := model.leftPane.Path()
	wantSubdirName := filepath.Base(originalDir)
	wantParentPath := filepath.Dir(originalDir)

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(0, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	clock.advance(200 * time.Millisecond)
	updated, cmd := model.Update(mouseMsg(mouseTestLeftX, rowFor(0, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("second press cmd = nil, want non-nil (Enter on \"..\")")
	}
	if model.leftPane.Path() != wantParentPath {
		t.Errorf("pane path = %q, want %q", model.leftPane.Path(), wantParentPath)
	}
	if model.leftPane.GetPendingCursorTarget() != wantSubdirName {
		t.Errorf("pendingCursorTarget = %q, want %q", model.leftPane.GetPendingCursorTarget(), wantSubdirName)
	}
}

func TestHandleMouse_DoubleClick_OnFile_ReturnsCommandPathUnchanged(t *testing.T) {
	clock := newFakeClock()
	model := newDoubleClickTestModel(t, clock)
	const fileEntryIndex = 2
	wantPath := model.leftPane.Path()

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(fileEntryIndex, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	clock.advance(200 * time.Millisecond)
	updated, cmd := model.Update(mouseMsg(mouseTestLeftX, rowFor(fileEntryIndex, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("second press cmd = nil, want non-nil (Enter on file)")
	}
	if model.leftPane.Path() != wantPath {
		t.Errorf("pane path changed to %q, want unchanged %q", model.leftPane.Path(), wantPath)
	}
}

func TestHandleMouse_NoDoubleClick_NonTriggeringSequences(t *testing.T) {
	t.Run("600ms apart", func(t *testing.T) {
		clock := newFakeClock()
		model := newDoubleClickTestModel(t, clock)
		updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(2, 0), tea.MouseButtonLeft, tea.MouseActionPress))
		model = updated.(Model)
		clock.advance(600 * time.Millisecond)
		updated, cmd := model.Update(mouseMsg(mouseTestLeftX, rowFor(2, 0), tea.MouseButtonLeft, tea.MouseActionPress))
		model = updated.(Model)
		if cmd != nil {
			t.Errorf("cmd = %v, want nil", cmd)
		}
	})

	t.Run("two different entries 100ms apart", func(t *testing.T) {
		clock := newFakeClock()
		model := newDoubleClickTestModel(t, clock)
		updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(1, 0), tea.MouseButtonLeft, tea.MouseActionPress))
		model = updated.(Model)
		clock.advance(100 * time.Millisecond)
		updated, cmd := model.Update(mouseMsg(mouseTestLeftX, rowFor(2, 0), tea.MouseButtonLeft, tea.MouseActionPress))
		model = updated.(Model)
		if cmd != nil {
			t.Errorf("cmd = %v, want nil", cmd)
		}
	})

	t.Run("A-B-A within 500ms", func(t *testing.T) {
		clock := newFakeClock()
		model := newDoubleClickTestModel(t, clock)
		rows := []int{1, 2, 1}
		var lastCmd tea.Cmd
		for i, idx := range rows {
			if i > 0 {
				clock.advance(100 * time.Millisecond)
			}
			updated, cmd := model.Update(mouseMsg(mouseTestLeftX, rowFor(idx, 0), tea.MouseButtonLeft, tea.MouseActionPress))
			model = updated.(Model)
			lastCmd = cmd
		}
		if lastCmd != nil {
			t.Errorf("cmd after A-B-A = %v, want nil", lastCmd)
		}
	})
}

func TestHandleMouse_ThreePressesWithin500ms_ExactlyOneEnter(t *testing.T) {
	clock := newFakeClock()
	model := newDoubleClickTestModel(t, clock)
	const fileEntryIndex = 2

	var cmds []tea.Cmd
	for i := 0; i < 3; i++ {
		if i > 0 {
			clock.advance(200 * time.Millisecond)
		}
		updated, cmd := model.Update(mouseMsg(mouseTestLeftX, rowFor(fileEntryIndex, 0), tea.MouseButtonLeft, tea.MouseActionPress))
		model = updated.(Model)
		cmds = append(cmds, cmd)
	}

	nonNil := 0
	for _, c := range cmds {
		if c != nil {
			nonNil++
		}
	}
	if nonNil != 1 {
		t.Errorf("non-nil cmds = %d, want exactly 1 (only the second press)", nonNil)
	}
	if cmds[1] == nil {
		t.Error("the second press should be the one that triggers Enter")
	}
}

func TestHandleMouse_DoubleClick_OnInactiveRightPane_Activates(t *testing.T) {
	clock := newFakeClock()
	model := newDoubleClickTestModel(t, clock)
	const dirEntryIndex = 1

	updated, _ := model.Update(mouseMsg(mouseTestRightX, rowFor(dirEntryIndex, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	clock.advance(200 * time.Millisecond)
	updated, cmd := model.Update(mouseMsg(mouseTestRightX, rowFor(dirEntryIndex, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)

	if model.activePane != RightPane {
		t.Errorf("activePane = %v, want RightPane", model.activePane)
	}
	if model.rightPane.cursor != dirEntryIndex {
		t.Errorf("cursor = %d, want %d", model.rightPane.cursor, dirEntryIndex)
	}
	if cmd == nil {
		t.Fatal("cmd = nil, want non-nil (Enter on directory)")
	}
}

func TestHandleMouse_PressAfterDragActivated_NoEnterEvenWithinWindow(t *testing.T) {
	clock := newFakeClock()
	model := newDoubleClickTestModel(t, clock)
	const dirEntryIndex = 1

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(dirEntryIndex, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(2, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(2, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	clock.advance(300 * time.Millisecond)
	updated, cmd := model.Update(mouseMsg(mouseTestLeftX, rowFor(dirEntryIndex, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)

	if cmd != nil {
		t.Errorf("cmd = %v, want nil (drag activation must have reset the detector)", cmd)
	}
}
