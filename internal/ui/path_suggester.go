package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PathSuggester provides filesystem-based path completion suggestions.
// It looks up directories in the filesystem and returns completion suffixes.
type PathSuggester struct{}

// NewPathSuggester creates a new PathSuggester instance.
func NewPathSuggester() *PathSuggester {
	return &PathSuggester{}
}

// Suggest returns the completion suffix for the given partial path.
// For example, if input is "/home/us" and "/home/user" exists,
// it returns "er" (the suffix to complete the path).
//
// Returns empty string if:
// - Input is empty
// - Input is not an absolute path
// - Parent directory doesn't exist
// - No matching directories found
// - The input exactly matches an existing directory
func (s *PathSuggester) Suggest(input string) string {
	if input == "" {
		return ""
	}

	// Only handle absolute paths
	if !strings.HasPrefix(input, "/") {
		return ""
	}

	// Handle trailing slash: suggest children of that directory
	if strings.HasSuffix(input, "/") {
		return s.suggestChildren(input)
	}

	// Split into parent directory and prefix
	parent := filepath.Dir(input)
	prefix := filepath.Base(input)

	// Read parent directory
	entries, err := os.ReadDir(parent)
	if err != nil {
		return ""
	}

	// Filter and collect matching directories
	var matches []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && name != prefix {
			matches = append(matches, name)
		}
	}

	if len(matches) == 0 {
		return ""
	}

	// Sort alphabetically and return suffix of first match
	sort.Strings(matches)
	firstMatch := matches[0]

	// Return suffix (part after the prefix)
	return firstMatch[len(prefix):]
}

// suggestChildren suggests the first child directory of the given directory path.
// The input path must end with "/".
func (s *PathSuggester) suggestChildren(dirPath string) string {
	// Remove trailing slash to get the directory path
	dir := strings.TrimSuffix(dirPath, "/")
	if dir == "" {
		dir = "/"
	}

	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	// Collect directories
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	if len(dirs) == 0 {
		return ""
	}

	// Sort and return first
	sort.Strings(dirs)
	return dirs[0]
}
