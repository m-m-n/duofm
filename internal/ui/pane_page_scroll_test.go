package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sakura/duofm/internal/fs"
)

func TestMoveCursorPageDown_NormalCase(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 100 test files
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i%10))+string(rune('0'+i/10))+".txt")
		if err := os.WriteFile(filename, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	pane, err := NewPane(LeftPane, tmpDir, 40, 24, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	if len(pane.entries) < 20 {
		t.Fatalf("Not enough entries for test: got %d, want at least 20", len(pane.entries))
	}

	// Set cursor to 0
	pane.cursor = 0

	// Move page down (24 - 4 = 20 visible lines)
	pane.MoveCursorPageDown()

	expectedCursor := 20
	if pane.cursor != expectedCursor {
		t.Errorf("MoveCursorPageDown() cursor = %d, want %d", pane.cursor, expectedCursor)
	}
}

func TestMoveCursorPageDown_NearBottom(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 50 test files
	for i := 0; i < 50; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i%10))+string(rune('0'+i/10))+".txt")
		if err := os.WriteFile(filename, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	pane, err := NewPane(LeftPane, tmpDir, 40, 24, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	if len(pane.entries) < 50 {
		t.Fatalf("Not enough entries: got %d, want at least 50", len(pane.entries))
	}

	// Set cursor near bottom
	pane.cursor = 40

	// Move page down
	pane.MoveCursorPageDown()

	// Should move to last entry (49)
	expectedCursor := len(pane.entries) - 1
	if pane.cursor != expectedCursor {
		t.Errorf("MoveCursorPageDown() cursor = %d, want %d (last entry)", pane.cursor, expectedCursor)
	}
}

func TestMoveCursorPageDown_AtBottom(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 50 test files
	for i := 0; i < 50; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i%10))+string(rune('0'+i/10))+".txt")
		if err := os.WriteFile(filename, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	pane, err := NewPane(LeftPane, tmpDir, 40, 24, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Set cursor to last entry
	pane.cursor = len(pane.entries) - 1
	initialCursor := pane.cursor

	// Move page down
	pane.MoveCursorPageDown()

	// Should stay at bottom
	if pane.cursor != initialCursor {
		t.Errorf("MoveCursorPageDown() at bottom: cursor = %d, want %d (unchanged)", pane.cursor, initialCursor)
	}
}

func TestMoveCursorPageUp_NormalCase(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 100 test files
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i%10))+string(rune('0'+i/10))+".txt")
		if err := os.WriteFile(filename, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	pane, err := NewPane(LeftPane, tmpDir, 40, 24, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	if len(pane.entries) < 60 {
		t.Fatalf("Not enough entries for test: got %d, want at least 60", len(pane.entries))
	}

	// Set cursor to middle
	pane.cursor = 50

	// Move page up (24 - 4 = 20 visible lines)
	pane.MoveCursorPageUp()

	expectedCursor := 30
	if pane.cursor != expectedCursor {
		t.Errorf("MoveCursorPageUp() cursor = %d, want %d", pane.cursor, expectedCursor)
	}
}

func TestMoveCursorPageUp_NearTop(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 50 test files
	for i := 0; i < 50; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i%10))+string(rune('0'+i/10))+".txt")
		if err := os.WriteFile(filename, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	pane, err := NewPane(LeftPane, tmpDir, 40, 24, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Set cursor near top
	pane.cursor = 10

	// Move page up (should clamp to 0)
	pane.MoveCursorPageUp()

	if pane.cursor != 0 {
		t.Errorf("MoveCursorPageUp() near top: cursor = %d, want 0", pane.cursor)
	}
}

func TestMoveCursorPageUp_AtTop(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 50 test files
	for i := 0; i < 50; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i%10))+string(rune('0'+i/10))+".txt")
		if err := os.WriteFile(filename, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	pane, err := NewPane(LeftPane, tmpDir, 40, 24, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Set cursor to top
	pane.cursor = 0

	// Move page up
	pane.MoveCursorPageUp()

	// Should stay at top
	if pane.cursor != 0 {
		t.Errorf("MoveCursorPageUp() at top: cursor = %d, want 0", pane.cursor)
	}
}

func TestPageScroll_SmallPane(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 100 test files
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i%10))+string(rune('0'+i/10))+".txt")
		if err := os.WriteFile(filename, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Very small pane (height 5 -> 1 visible line)
	pane, err := NewPane(LeftPane, tmpDir, 40, 5, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	pane.cursor = 0

	// Move page down (should move by at least 1 line)
	pane.MoveCursorPageDown()

	if pane.cursor < 1 {
		t.Errorf("MoveCursorPageDown() in small pane: cursor = %d, want at least 1", pane.cursor)
	}
}

func TestPageScroll_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	pane, err := NewPane(LeftPane, tmpDir, 40, 24, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Empty directory handling
	pane.entries = []fs.FileEntry{}
	pane.cursor = 0

	// Should not crash
	pane.MoveCursorPageDown()

	if pane.cursor != 0 {
		t.Errorf("MoveCursorPageDown() in empty dir: cursor = %d, want 0", pane.cursor)
	}

	pane.MoveCursorPageUp()

	if pane.cursor != 0 {
		t.Errorf("MoveCursorPageUp() in empty dir: cursor = %d, want 0", pane.cursor)
	}
}

func TestGetVisibleLines(t *testing.T) {
	tests := []struct {
		name     string
		height   int
		expected int
	}{
		{"Normal height", 24, 20},
		{"Small height", 10, 6},
		{"Very small height", 5, 1},
		{"Minimum height", 4, 1},     // height - 4 = 0, clamped to 1
		{"Tiny height", 3, 1},        // height - 4 = -1, clamped to 1
		{"Extremely tiny", 1, 1},     // height - 4 = -3, clamped to 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			pane, err := NewPane(LeftPane, tmpDir, 40, tt.height, true, nil)
			if err != nil {
				t.Fatalf("NewPane() failed: %v", err)
			}

			visible := pane.getVisibleLines()

			// Check exact expected value for all cases
			if visible != tt.expected {
				t.Errorf("getVisibleLines() = %d, want %d", visible, tt.expected)
			}

			// getVisibleLines must always return at least 1
			if visible < 1 {
				t.Errorf("getVisibleLines() = %d, want at least 1", visible)
			}
		})
	}
}
