package ui

import (
	"regexp"
	"strings"

	"github.com/sakura/duofm/internal/fs"
)

// SearchMode represents the type of search
type SearchMode int

const (
	// SearchModeNone indicates no active search
	SearchModeNone SearchMode = iota
	// SearchModeIncremental is real-time filtering as user types
	SearchModeIncremental
	// SearchModeRegex is regex pattern matching applied on confirm
	SearchModeRegex
)

// String returns the string representation of SearchMode
func (m SearchMode) String() string {
	switch m {
	case SearchModeIncremental:
		return "incremental"
	case SearchModeRegex:
		return "regex"
	default:
		return "none"
	}
}

// SearchState holds the current search context
type SearchState struct {
	Mode           SearchMode
	Pattern        string
	PreviousResult *SearchResult // For restoring on Esc
	IsActive       bool          // Minibuffer is open
}

// SearchResult stores a confirmed search
type SearchResult struct {
	Mode    SearchMode
	Pattern string
}

// isSmartCaseSensitive returns true if pattern contains uppercase letters
// (meaning case-sensitive search should be used)
func isSmartCaseSensitive(pattern string) bool {
	return pattern != strings.ToLower(pattern)
}

// filterIncremental filters entries by substring match with smart case
func filterIncremental(entries []fs.FileEntry, pattern string) []fs.FileEntry {
	if pattern == "" {
		return entries
	}

	caseSensitive := isSmartCaseSensitive(pattern)
	searchPattern := pattern
	if !caseSensitive {
		searchPattern = strings.ToLower(pattern)
	}

	result := make([]fs.FileEntry, 0)
	for _, e := range entries {
		name := e.Name
		if !caseSensitive {
			name = strings.ToLower(name)
		}
		if strings.Contains(name, searchPattern) {
			result = append(result, e)
		}
	}
	return result
}

// filterRegex filters entries by regex pattern with smart case
func filterRegex(entries []fs.FileEntry, pattern string) ([]fs.FileEntry, error) {
	if pattern == "" {
		return entries, nil
	}

	// Apply smart case: add case-insensitive flag if pattern is all lowercase
	regexPattern := pattern
	if !isSmartCaseSensitive(pattern) {
		regexPattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}

	result := make([]fs.FileEntry, 0)
	for _, e := range entries {
		if re.MatchString(e.Name) {
			result = append(result, e)
		}
	}
	return result, nil
}
