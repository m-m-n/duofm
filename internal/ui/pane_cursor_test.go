package ui

import (
	"testing"

	"github.com/sakura/duofm/internal/fs"
)

// TestCalculateCursorTargetAfterBatchMove tests cursor target calculation
// for batch move operations. The method should find the nearest unmarked file
// by searching upward first (skipping ".."), then downward.
func TestCalculateCursorTargetAfterBatchMove(t *testing.T) {
	tests := []struct {
		name        string
		entries     []fs.FileEntry
		cursor      int
		markedFiles map[string]bool
		expected    string
	}{
		{
			name: "B-1: cursor on unmarked file",
			entries: []fs.FileEntry{
				{Name: "..", IsDir: true},
				{Name: "a"},
				{Name: "b"},
				{Name: "c"},
				{Name: "d"},
			},
			cursor:      4, // d
			markedFiles: map[string]bool{"b": true, "c": true},
			expected:    "d",
		},
		{
			name: "B-2: search up finds unmarked",
			entries: []fs.FileEntry{
				{Name: "..", IsDir: true},
				{Name: "a"},
				{Name: "b"},
				{Name: "c"},
			},
			cursor:      3, // c (marked)
			markedFiles: map[string]bool{"b": true, "c": true},
			expected:    "a",
		},
		{
			name: "B-3: up all marked, search down finds unmarked",
			entries: []fs.FileEntry{
				{Name: "..", IsDir: true},
				{Name: "a"},
				{Name: "b"},
				{Name: "c"},
				{Name: "d"},
			},
			cursor:      1, // a (marked)
			markedFiles: map[string]bool{"a": true, "b": true},
			expected:    "c",
		},
		{
			name: "B-4: all files marked",
			entries: []fs.FileEntry{
				{Name: "..", IsDir: true},
				{Name: "a"},
				{Name: "b"},
			},
			cursor:      1,
			markedFiles: map[string]bool{"a": true, "b": true},
			expected:    "",
		},
		{
			name: "B-5: single file marked",
			entries: []fs.FileEntry{
				{Name: "..", IsDir: true},
				{Name: "a"},
			},
			cursor:      1,
			markedFiles: map[string]bool{"a": true},
			expected:    "",
		},
		{
			name: "B-6: unmarked files between marked",
			entries: []fs.FileEntry{
				{Name: "..", IsDir: true},
				{Name: "a"},
				{Name: "b"},
				{Name: "c"},
				{Name: "d"},
				{Name: "e"},
			},
			cursor:      4, // d
			markedFiles: map[string]bool{"a": true, "c": true, "e": true},
			expected:    "d",
		},
		{
			name: "B-7: cursor at 0 (on ..)",
			entries: []fs.FileEntry{
				{Name: "..", IsDir: true},
				{Name: "a"},
				{Name: "b"},
			},
			cursor:      0,
			markedFiles: map[string]bool{"a": true, "b": true},
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane := &Pane{
				entries: tt.entries,
				cursor:  tt.cursor,
			}
			result := pane.calculateCursorTargetAfterBatchMove(tt.markedFiles)
			if result != tt.expected {
				t.Errorf("calculateCursorTargetAfterBatchMove() = %q, want %q", result, tt.expected)
			}
		})
	}
}
