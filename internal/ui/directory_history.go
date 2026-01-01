package ui

// DirectoryHistory manages the directory navigation history for a pane.
// It maintains a stack of visited directories with a current position,
// similar to browser history (back/forward navigation).
type DirectoryHistory struct {
	paths        []string // History path list (max maxSize entries)
	currentIndex int      // Current position index (-1 = no history)
	maxSize      int      // Maximum size (100)
}

// NewDirectoryHistory creates a new DirectoryHistory with a maximum size of 100.
func NewDirectoryHistory() DirectoryHistory {
	return DirectoryHistory{
		paths:        []string{},
		currentIndex: -1,
		maxSize:      100,
	}
}

// AddToHistory adds a new directory path to the history.
// - Duplicate consecutive paths are ignored
// - All entries after the current position are deleted
// - New path is appended to history
// - currentIndex is set to the last position
// - If history exceeds maxSize entries, the oldest entry is removed
func (dh *DirectoryHistory) AddToHistory(path string) {
	// Ignore duplicate consecutive paths
	if dh.currentIndex >= 0 && dh.currentIndex < len(dh.paths) {
		if dh.paths[dh.currentIndex] == path {
			return
		}
	}

	// Truncate forward history (all entries after current position)
	if dh.currentIndex >= 0 && dh.currentIndex < len(dh.paths)-1 {
		dh.paths = dh.paths[:dh.currentIndex+1]
	}

	// Append new path
	dh.paths = append(dh.paths, path)

	// Enforce max size limit
	if len(dh.paths) > dh.maxSize {
		// Remove oldest entry (first element)
		dh.paths = dh.paths[1:]
	}

	// Set currentIndex to last position
	dh.currentIndex = len(dh.paths) - 1
}

// NavigateBack moves the current position backward in history.
// Returns the path at the new position and true if successful.
// Returns empty string and false if already at the beginning or history is empty.
func (dh *DirectoryHistory) NavigateBack() (string, bool) {
	if !dh.CanGoBack() {
		return "", false
	}

	dh.currentIndex--
	return dh.paths[dh.currentIndex], true
}

// NavigateForward moves the current position forward in history.
// Returns the path at the new position and true if successful.
// Returns empty string and false if already at the end or history is empty.
func (dh *DirectoryHistory) NavigateForward() (string, bool) {
	if !dh.CanGoForward() {
		return "", false
	}

	dh.currentIndex++
	return dh.paths[dh.currentIndex], true
}

// CanGoBack returns whether backward navigation is possible.
func (dh *DirectoryHistory) CanGoBack() bool {
	return dh.currentIndex > 0
}

// CanGoForward returns whether forward navigation is possible.
func (dh *DirectoryHistory) CanGoForward() bool {
	return dh.currentIndex >= 0 && dh.currentIndex < len(dh.paths)-1
}
