package ui

// SearchHistory manages search pattern history with navigation.
// It provides Up/Down navigation through previously used patterns
// while preserving the user's current input.
type SearchHistory struct {
	patterns []string // History entries (newest at index 0)
	index    int      // Current navigation position (-1 = at input, 0+ = in history)
	editBuf  string   // Original input before navigation started
	maxSize  int      // Maximum number of entries to keep
}

// DefaultSearchHistorySize is the default maximum number of history entries.
const DefaultSearchHistorySize = 50

// NewSearchHistory creates a new SearchHistory with the given max size.
func NewSearchHistory(maxSize int) *SearchHistory {
	return &SearchHistory{
		patterns: make([]string, 0),
		index:    -1,
		editBuf:  "",
		maxSize:  maxSize,
	}
}

// Add adds a pattern to history.
// - Empty patterns are ignored
// - If pattern already exists, it's moved to the front (deduplication)
// - History is truncated to maxSize if necessary
func (h *SearchHistory) Add(pattern string) {
	if pattern == "" {
		return
	}

	// Remove existing occurrence (deduplication)
	for i, p := range h.patterns {
		if p == pattern {
			h.patterns = append(h.patterns[:i], h.patterns[i+1:]...)
			break
		}
	}

	// Add to front
	h.patterns = append([]string{pattern}, h.patterns...)

	// Trim to max size
	if len(h.patterns) > h.maxSize {
		h.patterns = h.patterns[:h.maxSize]
	}
}

// NavigateUp moves to an older entry in history.
// On first call, saves currentInput to editBuf.
// Returns the pattern at the new position, or currentInput if history is empty.
func (h *SearchHistory) NavigateUp(currentInput string) string {
	if len(h.patterns) == 0 {
		return currentInput
	}

	// First navigation - save current input
	if h.index == -1 {
		h.editBuf = currentInput
	}

	// Move to older entry (bounded)
	if h.index < len(h.patterns)-1 {
		h.index++
	}

	return h.patterns[h.index]
}

// NavigateDown moves to a newer entry in history.
// Returns the pattern at the new position, or editBuf if at input position.
func (h *SearchHistory) NavigateDown() string {
	if h.index < 0 {
		return h.editBuf
	}

	h.index--

	if h.index == -1 {
		return h.editBuf
	}

	return h.patterns[h.index]
}

// Reset clears navigation state for a new dialog session.
// The patterns themselves are preserved.
func (h *SearchHistory) Reset() {
	h.index = -1
	h.editBuf = ""
}
