package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sakura/duofm/internal/fs"
)

func TestGotoTop_FromMiddle(t *testing.T) {
	tmpDir := t.TempDir()

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

	// Set cursor to middle
	pane.cursor = 25
	pane.scrollOffset = 10

	pane.GotoTop()

	if pane.cursor != 0 {
		t.Errorf("GotoTop() cursor = %d, want 0", pane.cursor)
	}
	if pane.scrollOffset != 0 {
		t.Errorf("GotoTop() scrollOffset = %d, want 0", pane.scrollOffset)
	}
}

func TestGotoTop_AlreadyAtTop(t *testing.T) {
	tmpDir := t.TempDir()

	for i := 0; i < 10; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".txt")
		if err := os.WriteFile(filename, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	pane, err := NewPane(LeftPane, tmpDir, 40, 24, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	pane.cursor = 0
	pane.scrollOffset = 0

	pane.GotoTop()

	if pane.cursor != 0 {
		t.Errorf("GotoTop() already at top: cursor = %d, want 0", pane.cursor)
	}
}

func TestGotoTop_EmptyList(t *testing.T) {
	tmpDir := t.TempDir()

	pane, err := NewPane(LeftPane, tmpDir, 40, 24, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	pane.entries = []fs.FileEntry{}
	pane.cursor = 0

	// Should not crash
	pane.GotoTop()

	if pane.cursor != 0 {
		t.Errorf("GotoTop() empty list: cursor = %d, want 0", pane.cursor)
	}
}

func TestGotoBottom_FromMiddle(t *testing.T) {
	tmpDir := t.TempDir()

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

	pane.cursor = 10

	pane.GotoBottom()

	expectedCursor := len(pane.entries) - 1
	if pane.cursor != expectedCursor {
		t.Errorf("GotoBottom() cursor = %d, want %d", pane.cursor, expectedCursor)
	}
}

func TestGotoBottom_AlreadyAtBottom(t *testing.T) {
	tmpDir := t.TempDir()

	for i := 0; i < 10; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".txt")
		if err := os.WriteFile(filename, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	pane, err := NewPane(LeftPane, tmpDir, 40, 24, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	lastIdx := len(pane.entries) - 1
	pane.cursor = lastIdx

	pane.GotoBottom()

	if pane.cursor != lastIdx {
		t.Errorf("GotoBottom() already at bottom: cursor = %d, want %d", pane.cursor, lastIdx)
	}
}

func TestGotoBottom_EmptyList(t *testing.T) {
	tmpDir := t.TempDir()

	pane, err := NewPane(LeftPane, tmpDir, 40, 24, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	pane.entries = []fs.FileEntry{}
	pane.cursor = 0

	// Should not crash
	pane.GotoBottom()

	if pane.cursor != 0 {
		t.Errorf("GotoBottom() empty list: cursor = %d, want 0", pane.cursor)
	}
}

func TestGotoBottom_AdjustsScroll(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 100 files to ensure scrolling is needed
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

	pane.cursor = 0
	pane.scrollOffset = 0

	pane.GotoBottom()

	// scrollOffset should be adjusted so the last entry is visible
	if pane.scrollOffset == 0 {
		t.Errorf("GotoBottom() scrollOffset = 0, expected non-zero for long list")
	}
}

func TestGotoTop_AdjustsScroll(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 100 files
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

	// Set cursor far down with scroll offset
	pane.cursor = 80
	pane.scrollOffset = 60

	pane.GotoTop()

	if pane.cursor != 0 {
		t.Errorf("GotoTop() cursor = %d, want 0", pane.cursor)
	}
	if pane.scrollOffset != 0 {
		t.Errorf("GotoTop() scrollOffset = %d, want 0", pane.scrollOffset)
	}
}

func TestGoto_SingleEntry(t *testing.T) {
	tmpDir := t.TempDir()

	pane, err := NewPane(LeftPane, tmpDir, 40, 24, true, nil)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	// Set entries to exactly one item
	pane.entries = []fs.FileEntry{{Name: "only_file.txt"}}
	pane.cursor = 0

	pane.GotoTop()
	if pane.cursor != 0 {
		t.Errorf("GotoTop() single entry: cursor = %d, want 0", pane.cursor)
	}

	pane.GotoBottom()
	if pane.cursor != 0 {
		t.Errorf("GotoBottom() single entry: cursor = %d, want 0", pane.cursor)
	}
}
