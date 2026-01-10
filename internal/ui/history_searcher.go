package ui

import (
	"strings"
)

// HistorySearcher provides incremental search functionality for shell history.
// It maintains search state including pattern and match index.
type HistorySearcher struct {
	history    *ShellHistory
	pattern    string
	matches    []int // Indices into history.Commands()
	matchIndex int   // Current position in matches
}

// NewHistorySearcher creates a new HistorySearcher for the given history.
func NewHistorySearcher(history *ShellHistory) *HistorySearcher {
	return &HistorySearcher{
		history:    history,
		pattern:    "",
		matches:    nil,
		matchIndex: -1,
	}
}

// SetPattern sets the search pattern and updates matches.
// Search is case-insensitive and matches anywhere in the command.
func (hs *HistorySearcher) SetPattern(pattern string) {
	hs.pattern = pattern
	hs.updateMatches()
	hs.matchIndex = 0
}

// Current returns the currently matched command, or empty string if no match.
func (hs *HistorySearcher) Current() string {
	if len(hs.matches) == 0 || hs.matchIndex < 0 || hs.matchIndex >= len(hs.matches) {
		return ""
	}

	commands := hs.history.Commands()
	idx := hs.matches[hs.matchIndex]
	if idx >= len(commands) {
		return ""
	}
	return commands[idx]
}

// Next moves to the next match and returns it.
// Wraps around to the first match after the last.
func (hs *HistorySearcher) Next() string {
	if len(hs.matches) == 0 {
		return ""
	}

	hs.matchIndex++
	if hs.matchIndex >= len(hs.matches) {
		hs.matchIndex = 0
	}

	return hs.Current()
}

// Reset clears the search state.
func (hs *HistorySearcher) Reset() {
	hs.pattern = ""
	hs.matches = nil
	hs.matchIndex = -1
}

// updateMatches rebuilds the matches list based on the current pattern.
func (hs *HistorySearcher) updateMatches() {
	hs.matches = nil

	if hs.pattern == "" {
		return
	}

	lowerPattern := strings.ToLower(hs.pattern)
	commands := hs.history.Commands()

	for i, cmd := range commands {
		if strings.Contains(strings.ToLower(cmd), lowerPattern) {
			hs.matches = append(hs.matches, i)
		}
	}
}
