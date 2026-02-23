package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sakura/duofm/internal/fs"
)

// --- FR4: getVisibleLines() returns reduced height during bg output ---

func TestGetVisibleLines_NormalWhenBgNotActive(t *testing.T) {
	tmpDir := t.TempDir()
	pane, err := NewPane(LeftPane, tmpDir, 40, 40, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// bg output not active: should return normal height
	visible := pane.getVisibleLines()
	expected := 40 - 4 // header(2) + border(1) + status(1) = 4
	if visible != expected {
		t.Errorf("getVisibleLines() without bg = %d, want %d", visible, expected)
	}
}

func TestGetVisibleLines_ReducedWhenBgActive(t *testing.T) {
	tmpDir := t.TempDir()
	pane, err := NewPane(LeftPane, tmpDir, 40, 40, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	pane.SetBgOutputActive(true)

	visible := pane.getVisibleLines()

	// Calculate expected: totalContent = 40-3 = 37, outputHeight = 37/3 = 12,
	// fileListHeight = 37 - 12 - 1 = 24
	if visible != 24 {
		t.Errorf("getVisibleLines() with bg = %d, want 24", visible)
	}
}

func TestGetVisibleLines_ReturnsNormalAfterBgDeactivated(t *testing.T) {
	tmpDir := t.TempDir()
	pane, err := NewPane(LeftPane, tmpDir, 40, 40, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Activate then deactivate
	pane.SetBgOutputActive(true)
	pane.SetBgOutputActive(false)

	visible := pane.getVisibleLines()
	expected := 40 - 4
	if visible != expected {
		t.Errorf("getVisibleLines() after bg deactivated = %d, want %d", visible, expected)
	}
}

func TestGetVisibleLines_BgActive_SmallHeight(t *testing.T) {
	tmpDir := t.TempDir()
	pane, err := NewPane(LeftPane, tmpDir, 40, 10, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	pane.SetBgOutputActive(true)

	visible := pane.getVisibleLines()

	// totalContent = 10-3 = 7, outputHeight = max(2, 7/3) = max(2, 2) = 2,
	// fileListHeight = 7 - 2 - 1 = 4
	if visible != 4 {
		t.Errorf("getVisibleLines() bg active small height = %d, want 4", visible)
	}

	// Must always be at least 1
	if visible < 1 {
		t.Errorf("getVisibleLines() must be >= 1, got %d", visible)
	}
}

func TestGetVisibleLines_BgActive_VerySmallHeight(t *testing.T) {
	tmpDir := t.TempDir()
	pane, err := NewPane(LeftPane, tmpDir, 40, 5, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	pane.SetBgOutputActive(true)

	visible := pane.getVisibleLines()

	// Must always be at least 1
	if visible < 1 {
		t.Errorf("getVisibleLines() bg active very small height must be >= 1, got %d", visible)
	}
}

// --- FR5: Cursor constrained to reduced visible area ---

func TestCursorConstraint_PageDownRespectsBgHeight(t *testing.T) {
	tmpDir := t.TempDir()
	for i := 0; i < 100; i++ {
		fname := filepath.Join(tmpDir, "file"+string(rune('a'+i%26))+string(rune('a'+i/26))+".txt")
		os.WriteFile(fname, []byte(""), 0644)
	}

	pane, err := NewPane(LeftPane, tmpDir, 40, 40, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	pane.SetBgOutputActive(true)
	pane.cursor = 0

	pane.MoveCursorPageDown()

	visibleLines := pane.getVisibleLines()
	if pane.cursor != visibleLines {
		t.Errorf("PageDown with bg: cursor = %d, want %d (visibleLines)", pane.cursor, visibleLines)
	}
}

func TestCursorConstraint_PageUpRespectsBgHeight(t *testing.T) {
	tmpDir := t.TempDir()
	for i := 0; i < 100; i++ {
		fname := filepath.Join(tmpDir, "file"+string(rune('a'+i%26))+string(rune('a'+i/26))+".txt")
		os.WriteFile(fname, []byte(""), 0644)
	}

	pane, err := NewPane(LeftPane, tmpDir, 40, 40, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	pane.SetBgOutputActive(true)
	pane.cursor = 50

	pane.MoveCursorPageUp()

	visibleLines := pane.getVisibleLines()
	expected := 50 - visibleLines
	if pane.cursor != expected {
		t.Errorf("PageUp with bg: cursor = %d, want %d", pane.cursor, expected)
	}
}

func TestCursorConstraint_ScrollAdjustsForBgHeight(t *testing.T) {
	tmpDir := t.TempDir()
	for i := 0; i < 100; i++ {
		fname := filepath.Join(tmpDir, "file"+string(rune('a'+i%26))+string(rune('a'+i/26))+".txt")
		os.WriteFile(fname, []byte(""), 0644)
	}

	pane, err := NewPane(LeftPane, tmpDir, 40, 40, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	pane.SetBgOutputActive(true)
	visibleLines := pane.getVisibleLines()

	// Move cursor past visible area
	pane.cursor = visibleLines + 5
	pane.adjustScroll()

	// scrollOffset should adjust so cursor is within visible range
	if pane.cursor < pane.scrollOffset || pane.cursor >= pane.scrollOffset+visibleLines {
		t.Errorf("cursor %d not in visible range [%d, %d)", pane.cursor, pane.scrollOffset, pane.scrollOffset+visibleLines)
	}
}

// --- NFR1: Normal behavior unaffected when no bg command ---

func TestNormalBehavior_UnaffectedWithoutBg(t *testing.T) {
	tmpDir := t.TempDir()
	for i := 0; i < 100; i++ {
		fname := filepath.Join(tmpDir, "file"+string(rune('a'+i%26))+string(rune('a'+i/26))+".txt")
		os.WriteFile(fname, []byte(""), 0644)
	}

	pane, err := NewPane(LeftPane, tmpDir, 40, 24, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Normal getVisibleLines
	visible := pane.getVisibleLines()
	expected := 24 - 4
	if visible != expected {
		t.Errorf("getVisibleLines() = %d, want %d", visible, expected)
	}

	// Normal page down
	pane.cursor = 0
	pane.MoveCursorPageDown()
	if pane.cursor != expected {
		t.Errorf("PageDown cursor = %d, want %d", pane.cursor, expected)
	}
}

// --- FR1/FR2: Separator line color tests ---

func TestViewWithBgOutput_SeparatorContainsCommand(t *testing.T) {
	pane := &Pane{
		path:        "/tmp",
		width:       80,
		height:      30,
		theme:       DefaultTheme(),
		entries:     []fs.FileEntry{},
		markedFiles: make(map[string]bool),
		sortConfig:  DefaultSortConfig(),
	}
	buf := NewOutputBuffer(100)
	buf.Append("test output")

	// Test unfocused - should not panic, should contain command text
	result := pane.ViewWithBgOutput(0, buf, "ls", false)
	if !strings.Contains(result, "ls") {
		t.Errorf("unfocused: expected separator to contain command text 'ls', got:\n%s", result)
	}

	// Test focused - should not panic, should contain command text
	result = pane.ViewWithBgOutput(0, buf, "ls", true)
	if !strings.Contains(result, "ls") {
		t.Errorf("focused: expected separator to contain command text 'ls', got:\n%s", result)
	}
}

func TestViewWithBgOutput_RenderDoesNotPanic(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		height  int
		focused bool
	}{
		{"normal unfocused", 80, 30, false},
		{"normal focused", 80, 30, true},
		{"small pane unfocused", 40, 10, false},
		{"small pane focused", 40, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane := &Pane{
				path:        "/tmp",
				width:       tt.width,
				height:      tt.height,
				theme:       DefaultTheme(),
				entries:     []fs.FileEntry{},
				markedFiles: make(map[string]bool),
				sortConfig:  DefaultSortConfig(),
			}
			buf := NewOutputBuffer(100)
			buf.Append("output line")

			// Should not panic
			result := pane.ViewWithBgOutput(0, buf, "echo test", tt.focused)
			if result == "" {
				t.Error("expected non-empty render output")
			}
		})
	}
}

// --- SetBgOutputActive/IsBgOutputActive tests ---

func TestSetBgOutputActive(t *testing.T) {
	tmpDir := t.TempDir()
	pane, err := NewPane(LeftPane, tmpDir, 40, 40, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	if pane.IsBgOutputActive() {
		t.Error("expected bgOutputActive=false initially")
	}

	pane.SetBgOutputActive(true)
	if !pane.IsBgOutputActive() {
		t.Error("expected bgOutputActive=true after Set(true)")
	}

	pane.SetBgOutputActive(false)
	if pane.IsBgOutputActive() {
		t.Error("expected bgOutputActive=false after Set(false)")
	}
}
