package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CompletionResult holds the result of a tab completion attempt.
type CompletionResult struct {
	NewInput     string   // updated input string
	NewCursorPos int      // updated cursor position
	Candidates   []string // matched candidates
	HasProgress  bool     // whether input changed (completion made progress)
}

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

// Complete returns a CompletionResult with updated input, cursor position, and candidate info.
// It determines whether to complete a command or file path based on cursor position.
func (tc *TabCompleter) Complete(input string, cursorPos int, cwd string) CompletionResult {
	if input == "" {
		return CompletionResult{NewInput: input, NewCursorPos: cursorPos}
	}

	word, wordStart := extractWordAtCursor(input, cursorPos)
	if word == "" {
		return CompletionResult{NewInput: input, NewCursorPos: cursorPos}
	}

	var candidates []string
	if isCommandPosition(input, cursorPos) {
		if strings.ContainsRune(word, '/') {
			// Path-like command (e.g., ./script.sh, ../bin/tool, /usr/local/bin/cmd)
			candidates = tc.completeExecutablePath(word, cwd)
		} else {
			candidates = tc.completeCommand(word)
		}
	} else {
		candidates = tc.completePath(word, cwd)
	}

	if len(candidates) == 0 {
		return CompletionResult{NewInput: input, NewCursorPos: cursorPos, Candidates: candidates}
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
			return CompletionResult{
				NewInput:     input,
				NewCursorPos: cursorPos,
				Candidates:   candidates,
			}
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

	return CompletionResult{
		NewInput:     newInput,
		NewCursorPos: newCursorPos,
		Candidates:   candidates,
		HasProgress:  true,
	}
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

	// Precompute base prefix for completion string construction.
	// Avoids repeated filepath.Dir() calls and handles root path correctly
	// (filepath.Dir("/ho") returns "/" which would produce "//home" with naive concatenation).
	var basePrefix string
	hasSlash := strings.Contains(prefix, "/")
	if hasSlash {
		if strings.HasSuffix(prefix, "/") {
			basePrefix = prefix
		} else {
			dirPart := filepath.Dir(prefix)
			if dirPart == "/" {
				basePrefix = "/"
			} else {
				basePrefix = dirPart + "/"
			}
		}
	}

	var matches []string
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, filePrefix) {
			continue
		}

		// Build the completion result preserving the original prefix path
		var completion string
		if hasSlash {
			completion = basePrefix + name
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

// completeExecutablePath finds matching executable files and directories for command position.
// Used when the input contains a path separator (e.g., ./script.sh, ../bin/tool).
func (tc *TabCompleter) completeExecutablePath(prefix string, cwd string) []string {
	var dir, filePrefix string

	if strings.HasSuffix(prefix, "/") {
		dir = prefix
		filePrefix = ""
	} else {
		dir = filepath.Dir(prefix)
		filePrefix = filepath.Base(prefix)
	}

	// Resolve relative to cwd
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(cwd, dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	// Precompute base prefix (same root-path fix as completePath)
	var basePrefix string
	if strings.HasSuffix(prefix, "/") {
		basePrefix = prefix
	} else {
		dirPart := filepath.Dir(prefix)
		if dirPart == "/" {
			basePrefix = "/"
		} else {
			basePrefix = dirPart + "/"
		}
	}

	var matches []string
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, filePrefix) {
			continue
		}

		completion := basePrefix + name

		if entry.IsDir() {
			completion += "/"
			matches = append(matches, completion)
			continue
		}

		// Only include executable files
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.Mode()&0111 != 0 {
			matches = append(matches, completion)
		}
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
