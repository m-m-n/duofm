package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TabCompleter provides TAB completion for shell command mode.
// It completes PATH executables for command position and
// file paths for argument positions.
type TabCompleter struct {
	pathCache []string // cached PATH executables
	pathEnv   string   // PATH value when cache was built
}

// NewTabCompleter creates a new TabCompleter
func NewTabCompleter() *TabCompleter {
	return &TabCompleter{}
}

// Complete returns updated input and cursor position after completion.
// It determines whether to complete a command or file path based on cursor position.
func (tc *TabCompleter) Complete(input string, cursorPos int, cwd string) (string, int) {
	if input == "" {
		return input, cursorPos
	}

	word, wordStart := extractWordAtCursor(input, cursorPos)
	if word == "" {
		return input, cursorPos
	}

	var candidates []string
	if isCommandPosition(input, cursorPos) {
		candidates = tc.completeCommand(word)
	} else {
		candidates = tc.completePath(word, cwd)
	}

	if len(candidates) == 0 {
		return input, cursorPos
	}

	var replacement string
	if len(candidates) == 1 {
		replacement = candidates[0]
		// Add trailing space for commands, / is already included for dirs
		if isCommandPosition(input, cursorPos) && !strings.HasSuffix(replacement, "/") {
			replacement += " "
		}
	} else {
		replacement = commonPrefix(candidates)
		if replacement == word {
			// No progress made
			return input, cursorPos
		}
	}

	// Reconstruct input
	runes := []rune(input)
	wordEnd := cursorPos
	newRunes := make([]rune, 0, len(runes)+len(replacement))
	newRunes = append(newRunes, runes[:wordStart]...)
	newRunes = append(newRunes, []rune(replacement)...)
	newRunes = append(newRunes, runes[wordEnd:]...)

	newInput := string(newRunes)
	newCursorPos := wordStart + len([]rune(replacement))

	return newInput, newCursorPos
}

// isCommandPosition returns true if the cursor is in the first word (command position)
func isCommandPosition(input string, cursorPos int) bool {
	runes := []rune(input)
	// Check if there's a space before the current word start
	for i := 0; i < cursorPos && i < len(runes); i++ {
		if runes[i] == ' ' {
			return false
		}
	}
	return true
}

// extractWordAtCursor extracts the word being completed and its start position
func extractWordAtCursor(input string, cursorPos int) (string, int) {
	runes := []rune(input)
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	// Find word start (scan backwards from cursor)
	wordStart := cursorPos
	for wordStart > 0 && runes[wordStart-1] != ' ' {
		wordStart--
	}

	return string(runes[wordStart:cursorPos]), wordStart
}

// completeCommand finds matching executables from PATH
func (tc *TabCompleter) completeCommand(prefix string) []string {
	tc.ensurePathCache()

	var matches []string
	for _, cmd := range tc.pathCache {
		if strings.HasPrefix(cmd, prefix) {
			matches = append(matches, cmd)
		}
	}
	sort.Strings(matches)
	return matches
}

// completePath finds matching files/directories in the given directory
func (tc *TabCompleter) completePath(prefix string, cwd string) []string {
	var dir, filePrefix string

	if strings.Contains(prefix, "/") {
		dir = filepath.Dir(prefix)
		filePrefix = filepath.Base(prefix)
		// Handle trailing slash (e.g., "internal/")
		if strings.HasSuffix(prefix, "/") {
			dir = prefix
			filePrefix = ""
		}
	} else {
		dir = "."
		filePrefix = prefix
	}

	// Resolve relative to cwd
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(cwd, dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var matches []string
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, filePrefix) {
			continue
		}

		// Build the completion result preserving the original prefix path
		var completion string
		if strings.Contains(prefix, "/") {
			if strings.HasSuffix(prefix, "/") {
				completion = prefix + name
			} else {
				completion = filepath.Dir(prefix) + "/" + name
			}
		} else {
			completion = name
		}

		if entry.IsDir() {
			completion += "/"
		}

		matches = append(matches, completion)
	}

	sort.Strings(matches)
	return matches
}

// ensurePathCache rebuilds the PATH cache if needed
func (tc *TabCompleter) ensurePathCache() {
	currentPath := os.Getenv("PATH")
	if currentPath == tc.pathEnv && tc.pathCache != nil {
		return
	}

	tc.buildPathCache()
}

// buildPathCache scans PATH directories for executables
func (tc *TabCompleter) buildPathCache() {
	tc.pathEnv = os.Getenv("PATH")
	tc.pathCache = nil

	if tc.pathEnv == "" {
		return
	}

	seen := make(map[string]bool)
	dirs := filepath.SplitList(tc.pathEnv)

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if seen[name] {
				continue
			}
			// Check if executable
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.Mode()&0111 != 0 {
				seen[name] = true
				tc.pathCache = append(tc.pathCache, name)
			}
		}
	}

	sort.Strings(tc.pathCache)
}

// commonPrefix returns the longest common prefix of candidates.
// It operates on runes to correctly handle multi-byte UTF-8 characters.
func commonPrefix(candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}
	if len(candidates) == 1 {
		return candidates[0]
	}

	prefix := []rune(candidates[0])
	for _, s := range candidates[1:] {
		runes := []rune(s)
		i := 0
		for i < len(prefix) && i < len(runes) && prefix[i] == runes[i] {
			i++
		}
		prefix = prefix[:i]
		if len(prefix) == 0 {
			return ""
		}
	}
	return string(prefix)
}
