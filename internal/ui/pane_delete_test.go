package ui

import (
	"testing"

	"github.com/sakura/duofm/internal/fs"
)

// TestCalculateCursorAfterDeletion_MiddleFile tests cursor calculation when deleting a file in the middle of the list
func TestCalculateCursorAfterDeletion_MiddleFile(t *testing.T) {
	pane := &Pane{
		entries: make([]fs.FileEntry, 5), // Simulate 5 entries after deletion
	}

	// Cursor was at index 2, one file deleted
	deletedIndex := 2
	result := pane.calculateCursorAfterDeletion(deletedIndex)

	if result != 2 {
		t.Errorf("Expected cursor at 2, got %d", result)
	}
}

// TestCalculateCursorAfterDeletion_LastFile tests cursor calculation when deleting the last file
func TestCalculateCursorAfterDeletion_LastFile(t *testing.T) {
	pane := &Pane{
		entries: make([]fs.FileEntry, 5), // Simulate 5 entries after deletion
	}

	// Cursor was at index 5 (last file before deletion), one file deleted
	deletedIndex := 5
	result := pane.calculateCursorAfterDeletion(deletedIndex)

	if result != 4 {
		t.Errorf("Expected cursor at 4 (new last file), got %d", result)
	}
}

// TestCalculateCursorAfterDeletion_MultipleFiles tests cursor calculation when deleting multiple marked files
func TestCalculateCursorAfterDeletion_MultipleFiles(t *testing.T) {
	tests := []struct {
		name           string
		deletedIndex   int
		remainingCount int
		expectedCursor int
	}{
		{
			name:           "Delete 3 files from middle",
			deletedIndex:   3,
			remainingCount: 7,
			expectedCursor: 3,
		},
		{
			name:           "Delete all files after cursor",
			deletedIndex:   5,
			remainingCount: 1, // Only .. remains
			expectedCursor: 0,
		},
		{
			name:           "Delete files at beginning",
			deletedIndex:   1,
			remainingCount: 8,
			expectedCursor: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane := &Pane{
				entries: make([]fs.FileEntry, tt.remainingCount),
			}
			result := pane.calculateCursorAfterDeletion(tt.deletedIndex)
			if result != tt.expectedCursor {
				t.Errorf("Expected %d, got %d", tt.expectedCursor, result)
			}
		})
	}
}

// TestCalculateCursorAfterDeletion_AllFiles tests cursor calculation when all files are deleted
func TestCalculateCursorAfterDeletion_AllFiles(t *testing.T) {
	pane := &Pane{
		entries: make([]fs.FileEntry, 1), // Only parent directory remains
	}

	deletedIndex := 2
	result := pane.calculateCursorAfterDeletion(deletedIndex)

	if result != 0 {
		t.Errorf("Expected cursor at 0 (parent dir), got %d", result)
	}
}

// TestCalculateCursorAfterDeletion_TwoEntriesOnly tests edge case with only 2 entries
func TestCalculateCursorAfterDeletion_TwoEntriesOnly(t *testing.T) {
	pane := &Pane{
		entries: make([]fs.FileEntry, 1), // After deletion: only .. remains
	}

	// Before: [.., file.txt]  cursor=1
	// After:  [..]            cursor=0
	deletedIndex := 1
	result := pane.calculateCursorAfterDeletion(deletedIndex)

	if result != 0 {
		t.Errorf("Expected cursor at 0, got %d", result)
	}
}

// TestCalculateCursorAfterDeletion_FirstRealFile tests deleting the first file after parent directory
func TestCalculateCursorAfterDeletion_FirstRealFile(t *testing.T) {
	pane := &Pane{
		entries: make([]fs.FileEntry, 3), // After deletion: [.., file2, file3]
	}

	// Before: [.., file1, file2, file3]  cursor=1
	// After:  [.., file2, file3]         cursor=1
	deletedIndex := 1
	result := pane.calculateCursorAfterDeletion(deletedIndex)

	if result != 1 {
		t.Errorf("Expected cursor at 1, got %d", result)
	}
}

// TestCalculateCursorAfterDeletion_EmptyDirectory tests edge case with empty directory
func TestCalculateCursorAfterDeletion_EmptyDirectory(t *testing.T) {
	pane := &Pane{
		entries: make([]fs.FileEntry, 0), // No entries at all (should not happen, but handle gracefully)
	}

	deletedIndex := 0
	result := pane.calculateCursorAfterDeletion(deletedIndex)

	if result != 0 {
		t.Errorf("Expected cursor at 0, got %d", result)
	}
}

// TestCalculateCursorAfterDeletion_BoundaryConditions tests various boundary conditions
func TestCalculateCursorAfterDeletion_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name           string
		deletedIndex   int
		remainingCount int
		expectedCursor int
	}{
		{
			name:           "Deleted index equals remaining count",
			deletedIndex:   5,
			remainingCount: 5,
			expectedCursor: 4,
		},
		{
			name:           "Deleted index exceeds remaining count",
			deletedIndex:   10,
			remainingCount: 3,
			expectedCursor: 2,
		},
		{
			name:           "Single entry remains",
			deletedIndex:   0,
			remainingCount: 1,
			expectedCursor: 0,
		},
		{
			name:           "Negative deleted index (defensive)",
			deletedIndex:   -1,
			remainingCount: 5,
			expectedCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane := &Pane{
				entries: make([]fs.FileEntry, tt.remainingCount),
			}
			result := pane.calculateCursorAfterDeletion(tt.deletedIndex)
			if result != tt.expectedCursor {
				t.Errorf("Expected %d, got %d", tt.expectedCursor, result)
			}
		})
	}
}
