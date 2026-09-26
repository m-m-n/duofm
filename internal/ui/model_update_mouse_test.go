package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
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

// --- AC-6: click, drag and double-click align with the renderer for a
// long, no-branch path that would wrap header line 1 before the fix ---

// newLongPathNoBranchModel builds a Model whose left pane shows a deeply
// nested, git-branch-less directory containing "subdir" (index1), and
// "alpha-file"/"bravo-file"/"charlie-file" (indices2-4), with ".." at
// index0 (SortEntries' parent-then-dirs-then-files ordering). The displayed
// path is wider than the pane at width 40 regardless of the temp root.
func newLongPathNoBranchModel(t *testing.T, clock *fakeClock) Model {
	t.Helper()
	const width, height = 80, 24
	paneWidth, paneHeight := width/2, height-2

	dir := t.TempDir()
	nested := dir
	for i := 0; i < 20; i++ {
		nested = filepath.Join(nested, fmt.Sprintf("nested-directory-segment-%02d", i))
	}
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.Mkdir(filepath.Join(nested, "subdir"), 0755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}
	for _, name := range []string{"alpha-file", "bravo-file", "charlie-file"} {
		if err := os.WriteFile(filepath.Join(nested, name), nil, 0644); err != nil {
			t.Fatalf("write file %s: %v", name, err)
		}
	}

	left, err := NewPane(LeftPane, nested, paneWidth, paneHeight, true, nil)
	if err != nil {
		t.Fatalf("NewPane left: %v", err)
	}
	left.gitBranch = "" // do not depend on whether the temp root is inside a git repo
	right := newFilesPane(t, RightPane, 5, paneWidth, paneHeight, false)

	return Model{
		leftPane:   left,
		rightPane:  right,
		activePane: LeftPane,
		width:      width,
		height:     height,
		detector:   newDoubleClickDetector(clock.now),
	}
}

// rowOfEntry finds the screen row of the entry containing name by reading
// the pane's own rendered output (offset by the title row), not the
// hit-test constant. This is what makes these tests detect a
// renderer / hit-test mismatch instead of hiding it behind a shared value.
func rowOfEntry(t *testing.T, p *Pane, name string) int {
	t.Helper()
	lines := strings.Split(stripANSI(p.View()), "\n")
	for i, line := range lines {
		if strings.Contains(line, name) {
			return mouseTitleRow + 1 + i
		}
	}
	t.Fatalf("entry %q not found in rendered pane view:\n%s", name, strings.Join(lines, "\n"))
	return -1
}

func TestHandleMouse_LongPathNoBranch_ClickMovesCursor(t *testing.T) {
	clock := newFakeClock()
	model := newLongPathNoBranchModel(t, clock)
	row := rowOfEntry(t, model.leftPane, "alpha-file")

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, row, tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, row, tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	if model.leftPane.cursor != 2 { // entries: ".." "subdir" "alpha-file" "bravo-file" "charlie-file"
		t.Errorf("cursor = %d, want 2", model.leftPane.cursor)
	}
}

func TestHandleMouse_LongPathNoBranch_DragMarksRangeExcludingParent(t *testing.T) {
	clock := newFakeClock()
	model := newLongPathNoBranchModel(t, clock)

	// The parent-dir row is not located by searching for "..": the
	// truncated header ends with "...", which also contains "..". It is
	// derived instead from "subdir"'s row, which is rendered directly below
	// it with no gap.
	subdirRow := rowOfEntry(t, model.leftPane, "subdir")
	parentRow := subdirRow - 1
	bravoRow := rowOfEntry(t, model.leftPane, "bravo-file")

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, parentRow, tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, bravoRow, tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, bravoRow, tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	assertMarks(t, model.leftPane,
		model.leftPane.entries[1].Name, model.leftPane.entries[2].Name, model.leftPane.entries[3].Name,
	)
	if model.leftPane.IsMarked("..") {
		t.Error("\"..\" must never be marked")
	}
}

func TestHandleMouse_LongPathNoBranch_DoubleClickOnSubdir_StartsLoading(t *testing.T) {
	clock := newFakeClock()
	model := newLongPathNoBranchModel(t, clock)
	wantPath := filepath.Join(model.leftPane.Path(), "subdir")
	row := rowOfEntry(t, model.leftPane, "subdir")

	updated, cmd1 := model.Update(mouseMsg(mouseTestLeftX, row, tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	if cmd1 != nil {
		t.Errorf("first press cmd = %v, want nil", cmd1)
	}

	clock.advance(200 * time.Millisecond)
	updated, cmd2 := model.Update(mouseMsg(mouseTestLeftX, row, tea.MouseButtonLeft, tea.MouseActionPress))
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

// --- task0001: keep the pre-press scroll offset when an entry press
// activates a bg-split pane ---
//
// Shared fixture: 80x24 geometry, left pane active over 20 generated files,
// right pane inactive over 20 generated files (21 entries including ".." at
// index0), right cursor 15, right scroll offset 0, a background command
// attached to the right pane. The right pane shows 18 entry rows without the
// bg split and 11 with it (see TestVisibleLinesAndBgSplitHeight_UnchangedAtGivenHeights).

// newBgSplitPressTestModel builds the fixture shared by AC-1 through AC-6.
// running selects the "running" variant (TS-1): the runner reports the
// right pane and IsRunning() true. When false, the runner still reports the
// right pane but IsRunning() false, and the model's closing flag is set
// instead (the "closing" variant, TS-2) -- both make the right pane's
// bg-split appear once it becomes active, the same way isBgActive() does.
func newBgSplitPressTestModel(t *testing.T, running bool) Model {
	t.Helper()
	model := newMouseTestModel(t, 20, 20)
	model.rightPane.cursor = 15
	model.rightPane.scrollOffset = 0
	model.bgRunner = &BackgroundRunner{running: running, pane: RightPane}
	if !running {
		model.bgClosing = true
	}
	return model
}

// newBgSplitPressDirFixtureModel builds the same fixture as
// newBgSplitPressTestModel (left pane active, right pane inactive with
// cursor 15 and scroll offset 0, a running background command attached to
// the right pane), except the right pane's directory holds 7 subdirectories
// sorted before 13 regular files. SortEntries orders entries as parent, then
// directories, then files, both name-ascending, so index 7 (the generated
// files fixture has no directory at all) lands on a directory here -- needed
// by AC-7's double-click-into-directory check.
func newBgSplitPressDirFixtureModel(t *testing.T, clock *fakeClock) Model {
	t.Helper()
	const width, height = 80, 24
	paneWidth, paneHeight := width/2, height-2

	left := newFilesPane(t, LeftPane, 20, paneWidth, paneHeight, true)

	dir := t.TempDir()
	for i := 0; i < 7; i++ {
		name := fmt.Sprintf("d%02d", i)
		if err := os.Mkdir(filepath.Join(dir, name), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
	}
	for i := 0; i < 13; i++ {
		name := fmt.Sprintf("f%02d", i)
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0644); err != nil {
			t.Fatalf("write file %s: %v", name, err)
		}
	}
	right, err := NewPane(RightPane, dir, paneWidth, paneHeight, false, nil)
	if err != nil {
		t.Fatalf("NewPane right: %v", err)
	}
	right.cursor = 15
	right.scrollOffset = 0

	return Model{
		leftPane:   left,
		rightPane:  right,
		activePane: LeftPane,
		width:      width,
		height:     height,
		detector:   newDoubleClickDetector(clock.now),
		bgRunner:   &BackgroundRunner{running: true, pane: RightPane},
	}
}

// AC-1 (TS-1 running, TS-2 closing): a same-row press/release on right index
// 7 leaves the mark set, cursor and scroll offset exactly as before the
// press.
func TestHandleMouse_BgSplitActivation_SameRowClick_KeepsPrePressOffsetAndMarks(t *testing.T) {
	cases := []struct {
		name    string
		running bool
	}{
		{"running", true},
		{"closing", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			model := newBgSplitPressTestModel(t, c.running)
			wantMarks := model.rightPane.snapshotMarks()

			updated, _ := model.Update(mouseMsg(mouseTestRightX, rowFor(7, 0), tea.MouseButtonLeft, tea.MouseActionPress))
			model = updated.(Model)
			updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(7, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
			model = updated.(Model)

			if model.rightPane.cursor != 7 {
				t.Errorf("cursor = %d, want 7", model.rightPane.cursor)
			}
			if model.rightPane.scrollOffset != 0 {
				t.Errorf("scrollOffset = %d, want 0", model.rightPane.scrollOffset)
			}
			assertMarksEqual(t, model.rightPane.markedFiles, wantMarks)
		})
	}
}

// AC-2 (TS-3): pressed index above the shifted window (a), inside the kept
// range (b), and below the shifted window but on a pre-press visible row
// (c) -- a same-row press/release adds no marks and the cursor ends on the
// pressed index in every case.
func TestHandleMouse_BgSplitActivation_PressReleaseAcrossOffsetPositions_NoMarks(t *testing.T) {
	cases := []struct {
		name        string
		index       int
		checkOffset bool
	}{
		{"above shifted window", 3, true},
		{"inside kept range", 7, true},
		{"below shifted window on pre-press visible row", 16, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			model := newBgSplitPressTestModel(t, true)

			updated, _ := model.Update(mouseMsg(mouseTestRightX, rowFor(c.index, 0), tea.MouseButtonLeft, tea.MouseActionPress))
			model = updated.(Model)
			updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(c.index, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
			model = updated.(Model)

			if model.rightPane.cursor != c.index {
				t.Errorf("cursor = %d, want %d", model.rightPane.cursor, c.index)
			}
			if model.rightPane.HasMarkedFiles() {
				t.Error("marks should be empty")
			}
			if c.checkOffset && model.rightPane.scrollOffset != 0 {
				t.Errorf("scrollOffset = %d, want 0", model.rightPane.scrollOffset)
			}
		})
	}
}

// AC-3 (TS-4): pressing an index at or beyond the pre-press offset plus the
// post-activation visible lines lands the cursor on the bottom visible
// file-list row, and a same-row release adds no marks.
func TestHandleMouse_BgSplitActivation_PressBelowRestoredWindow_ScrollsToBottomRow(t *testing.T) {
	model := newBgSplitPressTestModel(t, true)

	updated, _ := model.Update(mouseMsg(mouseTestRightX, rowFor(13, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(13, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	if model.rightPane.cursor != 13 {
		t.Errorf("cursor = %d, want 13", model.rightPane.cursor)
	}
	if model.rightPane.scrollOffset != 3 {
		t.Errorf("scrollOffset = %d, want 3", model.rightPane.scrollOffset)
	}
	if model.rightPane.HasMarkedFiles() {
		t.Error("marks should be empty")
	}
}

// AC-4 (TS-5): with entries marked before the press, both inside and
// outside the range the unmodified handler would have marked, a same-row
// press/release leaves the mark set exactly as it was before the press.
func TestHandleMouse_BgSplitActivation_SameRowClick_PreservesExistingMarks(t *testing.T) {
	model := newBgSplitPressTestModel(t, true)
	model.rightPane.markedFiles[model.rightPane.entries[9].Name] = true // inside 7..12
	model.rightPane.markedFiles[model.rightPane.entries[2].Name] = true // outside 7..12
	wantMarks := model.rightPane.snapshotMarks()

	updated, _ := model.Update(mouseMsg(mouseTestRightX, rowFor(7, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(7, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	assertMarksEqual(t, model.rightPane.markedFiles, wantMarks)
}

// AC-5 (TS-6): a same-row press/release on the ".." row adds no marks and
// leaves the cursor on it.
func TestHandleMouse_BgSplitActivation_ParentDirRow_NoMarksCursorOnParent(t *testing.T) {
	model := newBgSplitPressTestModel(t, true)

	updated, _ := model.Update(mouseMsg(mouseTestRightX, rowFor(0, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(0, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	if model.rightPane.cursor != 0 || !model.rightPane.entries[0].IsParentDir() {
		t.Errorf("cursor = %d, want on parent dir (index 0)", model.rightPane.cursor)
	}
	if model.rightPane.HasMarkedFiles() {
		t.Error("marks should be empty")
	}
}

// AC-6 (TS-7): a drag that starts from a bg-split-activating press maps
// motion and release rows against the restored pre-press offset, both when
// it extends past the anchor and when it returns to the anchor before
// release.
func TestHandleMouse_BgSplitActivation_Drag_UsesPrePressOffsetThroughout(t *testing.T) {
	t.Run("extends to index 9", func(t *testing.T) {
		model := newBgSplitPressTestModel(t, true)
		wantMarks := model.rightPane.snapshotMarks()

		updated, _ := model.Update(mouseMsg(mouseTestRightX, rowFor(7, 0), tea.MouseButtonLeft, tea.MouseActionPress))
		model = updated.(Model)
		updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(9, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
		model = updated.(Model)
		updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(9, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
		model = updated.(Model)

		wantMarks[model.rightPane.entries[7].Name] = true
		wantMarks[model.rightPane.entries[8].Name] = true
		wantMarks[model.rightPane.entries[9].Name] = true
		assertMarksEqual(t, model.rightPane.markedFiles, wantMarks)
	})

	t.Run("returns to anchor before release", func(t *testing.T) {
		model := newBgSplitPressTestModel(t, true)
		wantMarks := model.rightPane.snapshotMarks()

		updated, _ := model.Update(mouseMsg(mouseTestRightX, rowFor(7, 0), tea.MouseButtonLeft, tea.MouseActionPress))
		model = updated.(Model)
		updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(9, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
		model = updated.(Model)
		updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(7, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
		model = updated.(Model)
		updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(7, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
		model = updated.(Model)

		wantMarks[model.rightPane.entries[7].Name] = true
		assertMarksEqual(t, model.rightPane.markedFiles, wantMarks)
		if model.rightPane.cursor != 7 {
			t.Errorf("cursor = %d, want 7", model.rightPane.cursor)
		}
	})
}

// AC-7 (TS-8): a second press on the same row, within the double-click
// window, after a bg-split-activating first press/release, still hits the
// same directory entry and runs the existing Enter action on it.
func TestHandleMouse_BgSplitActivation_DoubleClickOnDirectory_StartsLoading(t *testing.T) {
	clock := newFakeClock()
	model := newBgSplitPressDirFixtureModel(t, clock)
	if !model.rightPane.entries[7].IsDir {
		t.Fatalf("fixture entries[7] = %+v, want a directory", model.rightPane.entries[7])
	}
	wantPath := filepath.Join(model.rightPane.Path(), model.rightPane.entries[7].Name)

	updated, cmd1 := model.Update(mouseMsg(mouseTestRightX, rowFor(7, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	if cmd1 != nil {
		t.Errorf("first press cmd = %v, want nil", cmd1)
	}
	updated, _ = model.Update(mouseMsg(mouseTestRightX, rowFor(7, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	clock.advance(200 * time.Millisecond)
	updated, cmd2 := model.Update(mouseMsg(mouseTestRightX, rowFor(7, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)

	if cmd2 == nil {
		t.Fatal("second press cmd = nil, want non-nil (Enter on directory)")
	}
	if model.rightPane.Path() != wantPath {
		t.Errorf("pane path = %q, want %q", model.rightPane.Path(), wantPath)
	}
	if !model.rightPane.IsLoading() {
		t.Error("pane should be loading the new directory")
	}
}

// AC-8 (TS-9): keyboard pane switching keeps its current bg-split scroll
// behavior. This pins pre-existing, unmodified behavior -- it passes both
// before and after this task's change -- guarding NFR1/NFR2 (the shared
// scroll logic and keyboard pane switching are not touched by this task).
func TestHandleMoveRight_BgSplitActivation_KeepsCurrentScrollBehavior(t *testing.T) {
	model := newBgSplitPressTestModel(t, true)

	updated, _ := model.handleMoveRight()
	model = updated.(Model)

	if model.activePane != RightPane {
		t.Errorf("activePane = %v, want RightPane", model.activePane)
	}
	if model.rightPane.cursor != 15 {
		t.Errorf("cursor = %d, want 15", model.rightPane.cursor)
	}
	if model.rightPane.scrollOffset != 5 {
		t.Errorf("scrollOffset = %d, want 5", model.rightPane.scrollOffset)
	}
}

// --- task0001: cancel the drag session on a successful directory load of
// the drag pane ---
//
// newDragLoadTestModel builds the same fixture as newMouseTestModel, plus an
// initialized disk space monitor: the success path of
// handleDirectoryLoadComplete calls updateDiskSpace, which would otherwise
// dereference a nil *DiskSpaceMonitor.

func newDragLoadTestModel(t *testing.T, leftN, rightN int) Model {
	t.Helper()
	m := newMouseTestModel(t, leftN, rightN)
	m.diskSpaceMonitor = NewDiskSpaceMonitor()
	return m
}

// newLoadEntries builds n plain fs.FileEntry values named "<prefix><NN>",
// none starting with a dot, so the success path's hidden-entry filter never
// removes any of them and the resulting list's length is exactly n.
func newLoadEntries(prefix string, n int) []fs.FileEntry {
	entries := make([]fs.FileEntry, n)
	for i := range entries {
		entries[i] = fs.FileEntry{Name: fmt.Sprintf("%s%02d", prefix, i)}
	}
	return entries
}

// TS-1 / AC-1: a session armed by a press (no motion) is cancelled by a
// successful completion for its own pane; a subsequent motion and release
// leave the mark set exactly as it was right after the completion.
func TestHandleMouse_DragSession_CancelledOnLoadCompleteForArmedPane(t *testing.T) {
	model := newDragLoadTestModel(t, 20, 20)
	const anchor = 2

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(anchor, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	if !model.mouseDrag.armed || model.mouseDrag.active {
		t.Fatalf("session after press = %+v, want armed and not active", model.mouseDrag)
	}

	loadMsg := directoryLoadCompleteMsg{
		paneID:   LeftPane,
		panePath: model.leftPane.Path(),
		entries:  newLoadEntries("new", 8),
	}
	updated, _ = model.Update(loadMsg)
	model = updated.(Model)

	if model.mouseDrag.armed || model.mouseDrag.active {
		t.Fatalf("session after load complete = %+v, want neither armed nor active", model.mouseDrag)
	}

	wantMarks := model.leftPane.snapshotMarks()

	const motionIndex = 4
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motionIndex, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motionIndex, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	assertMarksEqual(t, model.leftPane.markedFiles, wantMarks)
}

// TS-2 / AC-2: a session made active by a press and a motion (an existing
// marked range) is cancelled by a successful completion for its own pane;
// the marks present immediately before the completion survive unchanged,
// not reverted to the press-time baseline, and a subsequent motion and
// release do not change them.
func TestHandleMouse_DragSession_CancelledOnLoadCompleteKeepsCurrentMarks(t *testing.T) {
	model := newDragLoadTestModel(t, 20, 20)
	const anchor, motionA = 2, 5

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(anchor, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motionA, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	if !model.mouseDrag.active {
		t.Fatal("session should be active after motion")
	}

	wantMarks := model.leftPane.snapshotMarks()
	if len(wantMarks) == 0 {
		t.Fatal("expected marks from anchor..motionA before completion")
	}

	loadMsg := directoryLoadCompleteMsg{
		paneID:   LeftPane,
		panePath: model.leftPane.Path(),
		entries:  newLoadEntries("new", 8),
	}
	updated, _ = model.Update(loadMsg)
	model = updated.(Model)

	if model.mouseDrag.armed || model.mouseDrag.active {
		t.Fatalf("session after load complete = %+v, want neither armed nor active", model.mouseDrag)
	}
	assertMarksEqual(t, model.leftPane.markedFiles, wantMarks)

	const motionB = 3
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motionB, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motionB, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	assertMarksEqual(t, model.leftPane.markedFiles, wantMarks)
}

// TS-3 / AC-3: a successful completion for the OTHER pane leaves an armed
// left-pane session completely unchanged; a subsequent left-pane motion
// marks exactly the range from the original anchor to the motion row.
func TestHandleMouse_DragSession_UnaffectedByLoadCompleteForOtherPane(t *testing.T) {
	model := newDragLoadTestModel(t, 20, 20)
	const anchor = 2

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(anchor, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	wantSession := model.mouseDrag
	wantBaseline := model.mouseDrag.baseline

	loadMsg := directoryLoadCompleteMsg{
		paneID:   RightPane,
		panePath: model.rightPane.Path(),
		entries:  newLoadEntries("new", 8),
	}
	updated, _ = model.Update(loadMsg)
	model = updated.(Model)

	if !model.mouseDrag.armed || model.mouseDrag.active {
		t.Fatalf("left session changed by right-pane completion: %+v", model.mouseDrag)
	}
	if model.mouseDrag.pane != wantSession.pane || model.mouseDrag.anchor != wantSession.anchor {
		t.Fatalf("left session pane/anchor changed: got %+v, want %+v", model.mouseDrag, wantSession)
	}
	assertMarksEqual(t, model.mouseDrag.baseline, wantBaseline)

	const motionIndex = 6
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motionIndex, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motionIndex, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	var wantNames []string
	for i := anchor; i <= motionIndex; i++ {
		wantNames = append(wantNames, model.leftPane.entries[i].Name)
	}
	assertMarks(t, model.leftPane, wantNames...)
}

// TS-4 / AC-4: a completion for the left pane that carries an error leaves
// the session unchanged; a subsequent motion marks exactly the range from
// the original anchor to the motion row.
func TestHandleMouse_DragSession_UnaffectedByLoadCompleteError(t *testing.T) {
	model := newDragLoadTestModel(t, 20, 20)
	const anchor = 3

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(anchor, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	wantSession := model.mouseDrag
	wantBaseline := model.mouseDrag.baseline

	loadMsg := directoryLoadCompleteMsg{
		paneID:        LeftPane,
		panePath:      model.leftPane.Path(),
		err:           errors.New("boom"),
		attemptedPath: model.leftPane.Path(),
	}
	updated, _ = model.Update(loadMsg)
	model = updated.(Model)

	if !model.mouseDrag.armed || model.mouseDrag.active {
		t.Fatalf("session changed by error completion: %+v", model.mouseDrag)
	}
	if model.mouseDrag.pane != wantSession.pane || model.mouseDrag.anchor != wantSession.anchor {
		t.Fatalf("session pane/anchor changed: got %+v, want %+v", model.mouseDrag, wantSession)
	}
	assertMarksEqual(t, model.mouseDrag.baseline, wantBaseline)

	const motionIndex = 7
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motionIndex, 0), tea.MouseButtonLeft, tea.MouseActionMotion))
	model = updated.(Model)
	updated, _ = model.Update(mouseMsg(mouseTestLeftX, rowFor(motionIndex, 0), tea.MouseButtonLeft, tea.MouseActionRelease))
	model = updated.(Model)

	var wantNames []string
	for i := anchor; i <= motionIndex; i++ {
		wantNames = append(wantNames, model.leftPane.entries[i].Name)
	}
	assertMarks(t, model.leftPane, wantNames...)
}

// TS-5 / AC-5: a stale completion (pane path differs from a non-empty
// pending path) for the left pane leaves both the session and the left
// pane's entry list unchanged.
func TestHandleMouse_DragSession_UnaffectedByStaleLoadComplete(t *testing.T) {
	model := newDragLoadTestModel(t, 20, 20)
	const anchor = 1

	updated, _ := model.Update(mouseMsg(mouseTestLeftX, rowFor(anchor, 0), tea.MouseButtonLeft, tea.MouseActionPress))
	model = updated.(Model)
	model.leftPane.pendingPath = "/some/other/pending/path"
	wantSession := model.mouseDrag
	wantBaseline := model.mouseDrag.baseline
	wantEntries := append([]fs.FileEntry(nil), model.leftPane.entries...)

	loadMsg := directoryLoadCompleteMsg{
		paneID:   LeftPane,
		panePath: "/does/not/match/pending",
		entries:  newLoadEntries("new", 8),
	}
	updated, _ = model.Update(loadMsg)
	model = updated.(Model)

	if !model.mouseDrag.armed || model.mouseDrag.active {
		t.Fatalf("session changed by stale completion: %+v", model.mouseDrag)
	}
	if model.mouseDrag.pane != wantSession.pane || model.mouseDrag.anchor != wantSession.anchor {
		t.Fatalf("session pane/anchor changed: got %+v, want %+v", model.mouseDrag, wantSession)
	}
	assertMarksEqual(t, model.mouseDrag.baseline, wantBaseline)

	if len(model.leftPane.entries) != len(wantEntries) {
		t.Fatalf("entries changed: got %d, want %d", len(model.leftPane.entries), len(wantEntries))
	}
	for i, e := range wantEntries {
		if model.leftPane.entries[i].Name != e.Name {
			t.Fatalf("entries[%d].Name = %q, want %q", i, model.leftPane.entries[i].Name, e.Name)
		}
	}
}
