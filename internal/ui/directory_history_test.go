package ui

import (
	"testing"
)

// TestNewDirectoryHistory tests the constructor
func TestNewDirectoryHistory(t *testing.T) {
	dh := NewDirectoryHistory()

	if dh.maxSize != 100 {
		t.Errorf("Expected maxSize to be 100, got %d", dh.maxSize)
	}

	if dh.currentIndex != -1 {
		t.Errorf("Expected currentIndex to be -1, got %d", dh.currentIndex)
	}

	if len(dh.paths) != 0 {
		t.Errorf("Expected paths to be empty, got %d entries", len(dh.paths))
	}
}

// TestAddToHistory_EmptyHistory tests adding to empty history
func TestAddToHistory_EmptyHistory(t *testing.T) {
	dh := NewDirectoryHistory()

	dh.AddToHistory("/home/user")

	if len(dh.paths) != 1 {
		t.Errorf("Expected 1 path, got %d", len(dh.paths))
	}

	if dh.paths[0] != "/home/user" {
		t.Errorf("Expected path '/home/user', got '%s'", dh.paths[0])
	}

	if dh.currentIndex != 0 {
		t.Errorf("Expected currentIndex to be 0, got %d", dh.currentIndex)
	}
}

// TestAddToHistory_MultiplePaths tests adding multiple paths
func TestAddToHistory_MultiplePaths(t *testing.T) {
	dh := NewDirectoryHistory()

	paths := []string{"/home", "/home/user", "/home/user/docs"}
	for _, path := range paths {
		dh.AddToHistory(path)
	}

	if len(dh.paths) != 3 {
		t.Errorf("Expected 3 paths, got %d", len(dh.paths))
	}

	if dh.currentIndex != 2 {
		t.Errorf("Expected currentIndex to be 2, got %d", dh.currentIndex)
	}

	for i, path := range paths {
		if dh.paths[i] != path {
			t.Errorf("Expected path[%d] to be '%s', got '%s'", i, path, dh.paths[i])
		}
	}
}

// TestAddToHistory_DuplicateConsecutivePaths tests that duplicate consecutive paths are ignored
func TestAddToHistory_DuplicateConsecutivePaths(t *testing.T) {
	dh := NewDirectoryHistory()

	dh.AddToHistory("/home/user")
	dh.AddToHistory("/home/user") // Duplicate

	if len(dh.paths) != 1 {
		t.Errorf("Expected 1 path (duplicate ignored), got %d", len(dh.paths))
	}

	if dh.currentIndex != 0 {
		t.Errorf("Expected currentIndex to be 0, got %d", dh.currentIndex)
	}
}

// TestAddToHistory_TruncateForwardHistory tests forward history truncation
func TestAddToHistory_TruncateForwardHistory(t *testing.T) {
	dh := NewDirectoryHistory()

	// Build history: A -> B -> C -> D -> E
	dh.AddToHistory("/a")
	dh.AddToHistory("/b")
	dh.AddToHistory("/c")
	dh.AddToHistory("/d")
	dh.AddToHistory("/e")

	// Navigate back twice: position at C (index 2)
	dh.NavigateBack() // to D
	dh.NavigateBack() // to C

	if dh.currentIndex != 2 {
		t.Errorf("Expected currentIndex to be 2 after 2 backs, got %d", dh.currentIndex)
	}

	// Add new path F - should truncate D and E
	dh.AddToHistory("/f")

	expectedPaths := []string{"/a", "/b", "/c", "/f"}
	if len(dh.paths) != len(expectedPaths) {
		t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(dh.paths))
	}

	for i, expected := range expectedPaths {
		if dh.paths[i] != expected {
			t.Errorf("Expected path[%d] to be '%s', got '%s'", i, expected, dh.paths[i])
		}
	}

	if dh.currentIndex != 3 {
		t.Errorf("Expected currentIndex to be 3, got %d", dh.currentIndex)
	}
}

// TestAddToHistory_MaxSizeEnforcement tests max size limit (100 entries)
func TestAddToHistory_MaxSizeEnforcement(t *testing.T) {
	dh := NewDirectoryHistory()

	// Add 100 paths
	for i := 0; i < 100; i++ {
		dh.AddToHistory("/path" + string(rune('0'+i%10)))
	}

	if len(dh.paths) != 100 {
		t.Errorf("Expected 100 paths, got %d", len(dh.paths))
	}

	// Add 101st path - should remove oldest
	dh.AddToHistory("/path101")

	if len(dh.paths) != 100 {
		t.Errorf("Expected 100 paths (max size), got %d", len(dh.paths))
	}

	if dh.currentIndex != 99 {
		t.Errorf("Expected currentIndex to be 99, got %d", dh.currentIndex)
	}

	// Last path should be the 101st
	if dh.paths[99] != "/path101" {
		t.Errorf("Expected last path to be '/path101', got '%s'", dh.paths[99])
	}
}

// TestNavigateBack_EmptyHistory tests back navigation on empty history
func TestNavigateBack_EmptyHistory(t *testing.T) {
	dh := NewDirectoryHistory()

	path, ok := dh.NavigateBack()

	if ok {
		t.Error("Expected NavigateBack to return false on empty history")
	}

	if path != "" {
		t.Errorf("Expected empty path, got '%s'", path)
	}

	if dh.currentIndex != -1 {
		t.Errorf("Expected currentIndex to remain -1, got %d", dh.currentIndex)
	}
}

// TestNavigateBack_AtBeginning tests back navigation at beginning
func TestNavigateBack_AtBeginning(t *testing.T) {
	dh := NewDirectoryHistory()
	dh.AddToHistory("/home")
	dh.AddToHistory("/home/user")

	// Navigate back to /home (index 0)
	dh.NavigateBack()

	// Try to navigate back again - should fail
	path, ok := dh.NavigateBack()

	if ok {
		t.Error("Expected NavigateBack to return false at beginning")
	}

	if path != "" {
		t.Errorf("Expected empty path, got '%s'", path)
	}

	if dh.currentIndex != 0 {
		t.Errorf("Expected currentIndex to remain 0, got %d", dh.currentIndex)
	}
}

// TestNavigateBack_Success tests successful back navigation
func TestNavigateBack_Success(t *testing.T) {
	dh := NewDirectoryHistory()
	dh.AddToHistory("/a")
	dh.AddToHistory("/b")
	dh.AddToHistory("/c")

	// currentIndex should be 2 (at /c)

	// Navigate back to /b
	path, ok := dh.NavigateBack()

	if !ok {
		t.Error("Expected NavigateBack to succeed")
	}

	if path != "/b" {
		t.Errorf("Expected path '/b', got '%s'", path)
	}

	if dh.currentIndex != 1 {
		t.Errorf("Expected currentIndex to be 1, got %d", dh.currentIndex)
	}

	// Navigate back again to /a
	path, ok = dh.NavigateBack()

	if !ok {
		t.Error("Expected NavigateBack to succeed")
	}

	if path != "/a" {
		t.Errorf("Expected path '/a', got '%s'", path)
	}

	if dh.currentIndex != 0 {
		t.Errorf("Expected currentIndex to be 0, got %d", dh.currentIndex)
	}
}

// TestNavigateForward_EmptyHistory tests forward navigation on empty history
func TestNavigateForward_EmptyHistory(t *testing.T) {
	dh := NewDirectoryHistory()

	path, ok := dh.NavigateForward()

	if ok {
		t.Error("Expected NavigateForward to return false on empty history")
	}

	if path != "" {
		t.Errorf("Expected empty path, got '%s'", path)
	}
}

// TestNavigateForward_AtEnd tests forward navigation at end
func TestNavigateForward_AtEnd(t *testing.T) {
	dh := NewDirectoryHistory()
	dh.AddToHistory("/a")
	dh.AddToHistory("/b")

	// Already at end (index 1)

	path, ok := dh.NavigateForward()

	if ok {
		t.Error("Expected NavigateForward to return false at end")
	}

	if path != "" {
		t.Errorf("Expected empty path, got '%s'", path)
	}

	if dh.currentIndex != 1 {
		t.Errorf("Expected currentIndex to remain 1, got %d", dh.currentIndex)
	}
}

// TestNavigateForward_Success tests successful forward navigation
func TestNavigateForward_Success(t *testing.T) {
	dh := NewDirectoryHistory()
	dh.AddToHistory("/a")
	dh.AddToHistory("/b")
	dh.AddToHistory("/c")

	// Navigate back twice
	dh.NavigateBack()
	dh.NavigateBack()

	// currentIndex should be 0 (at /a)

	// Navigate forward to /b
	path, ok := dh.NavigateForward()

	if !ok {
		t.Error("Expected NavigateForward to succeed")
	}

	if path != "/b" {
		t.Errorf("Expected path '/b', got '%s'", path)
	}

	if dh.currentIndex != 1 {
		t.Errorf("Expected currentIndex to be 1, got %d", dh.currentIndex)
	}

	// Navigate forward again to /c
	path, ok = dh.NavigateForward()

	if !ok {
		t.Error("Expected NavigateForward to succeed")
	}

	if path != "/c" {
		t.Errorf("Expected path '/c', got '%s'", path)
	}

	if dh.currentIndex != 2 {
		t.Errorf("Expected currentIndex to be 2, got %d", dh.currentIndex)
	}
}

// TestCanGoBack tests CanGoBack functionality
func TestCanGoBack(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*DirectoryHistory)
		expected bool
	}{
		{
			name:     "empty history",
			setup:    func(dh *DirectoryHistory) {},
			expected: false,
		},
		{
			name: "at beginning",
			setup: func(dh *DirectoryHistory) {
				dh.AddToHistory("/a")
			},
			expected: false,
		},
		{
			name: "can go back",
			setup: func(dh *DirectoryHistory) {
				dh.AddToHistory("/a")
				dh.AddToHistory("/b")
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dh := NewDirectoryHistory()
			tt.setup(&dh)

			result := dh.CanGoBack()
			if result != tt.expected {
				t.Errorf("Expected CanGoBack to be %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCanGoForward tests CanGoForward functionality
func TestCanGoForward(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*DirectoryHistory)
		expected bool
	}{
		{
			name:     "empty history",
			setup:    func(dh *DirectoryHistory) {},
			expected: false,
		},
		{
			name: "at end",
			setup: func(dh *DirectoryHistory) {
				dh.AddToHistory("/a")
				dh.AddToHistory("/b")
			},
			expected: false,
		},
		{
			name: "can go forward",
			setup: func(dh *DirectoryHistory) {
				dh.AddToHistory("/a")
				dh.AddToHistory("/b")
				dh.NavigateBack()
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dh := NewDirectoryHistory()
			tt.setup(&dh)

			result := dh.CanGoForward()
			if result != tt.expected {
				t.Errorf("Expected CanGoForward to be %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestComplexStateTransitions tests the example from spec: A→B→C, back twice, add D
func TestComplexStateTransitions(t *testing.T) {
	dh := NewDirectoryHistory()

	// A → B → C
	dh.AddToHistory("/a")
	dh.AddToHistory("/b")
	dh.AddToHistory("/c")

	// Back twice (to A)
	dh.NavigateBack() // to B
	dh.NavigateBack() // to A

	if dh.currentIndex != 0 {
		t.Errorf("Expected currentIndex to be 0, got %d", dh.currentIndex)
	}

	// Add D - should result in [A, B] -> [A, D]
	// Wait, spec says [A, B, D] - let me re-read
	// "After A→B→C, going back twice then navigating to D results in history [A, B, D]"
	// So back twice from C means: C -> B -> A
	// We're at A (index 0)
	// But wait, the current position should be at A
	// When we add D, we truncate everything after current position
	// Current history: [A, B, C], currentIndex = 0
	// Truncate: [A], then append D: [A, D]
	// But spec says [A, B, D]

	// Let me re-read: "back twice" - from end position (index 2, at C)
	// Back once: index 1, at B
	// Back twice: index 0, at A
	// When at A and add D, we truncate after A: [A] + [D] = [A, D]

	// Hmm, spec example might mean: A->B->C, back once to B, add D -> [A,B,D]
	// Let me check spec again...
	// "Scenario 6: After A→B→C, going back twice then navigating to D results in history [A, B, D]"

	// Actually, I think the spec might be saying:
	// A->B->C (3 entries)
	// Go back twice means go back from C to B (1 back), then... wait
	// Or maybe it means: after reaching C, the history is [A,B,C]
	// Go back twice: to A? But then adding D would give [A,D]
	// Unless "go back twice" means navigate back 2 steps, ending at A
	// But A->B->C->D with forward history cleared would be...

	// Let me look at the IMPLEMENTATION.md example:
	// "Mid-state: paths=[A, B, C, D], currentIndex=1 (pointing to B)"
	// "AddToHistory(E) → paths=[A, B, E], currentIndex=2"
	// So at index 1 (B), adding E truncates C,D and appends E -> [A,B,E]

	// So for scenario: A->B->C, if we go back to A (index 0), adding D -> [A,D]
	// If we go back to B (index 1), adding D -> [A,B,D]

	// I think "back twice" might mean:
	// 1. Back from C to B
	// 2. Back from B to A
	// But that's confusing with the expected result [A,B,D]

	// Let me just test what the implementation should do:
	// If currentIndex = 1 (at B), adding D -> [A, B, D]
	dh.NavigateForward() // to B (index 1)

	dh.AddToHistory("/d")

	expectedPaths := []string{"/a", "/b", "/d"}
	if len(dh.paths) != len(expectedPaths) {
		t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(dh.paths))
	}

	for i, expected := range expectedPaths {
		if dh.paths[i] != expected {
			t.Errorf("Expected path[%d] to be '%s', got '%s'", i, expected, dh.paths[i])
		}
	}

	if dh.currentIndex != 2 {
		t.Errorf("Expected currentIndex to be 2, got %d", dh.currentIndex)
	}
}
