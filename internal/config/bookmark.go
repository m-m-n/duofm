package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

// Bookmark represents a single directory bookmark.
type Bookmark struct {
	Name string `toml:"name"`
	Path string `toml:"path"`
}

// ErrDuplicatePath is returned when attempting to bookmark an already bookmarked path.
var ErrDuplicatePath = errors.New("path is already bookmarked")

// ErrInvalidIndex is returned when bookmark index is out of range.
var ErrInvalidIndex = errors.New("bookmark index out of range")

// ErrEmptyAlias is returned when alias is empty.
var ErrEmptyAlias = errors.New("alias cannot be empty")

// bookmarksConfig is used for TOML parsing of bookmarks section.
type bookmarksConfig struct {
	Bookmarks []Bookmark `toml:"bookmarks"`
}

// LoadBookmarks loads bookmarks from a TOML configuration file.
// Returns empty slice if file doesn't exist or has no bookmarks section.
func LoadBookmarks(path string) ([]Bookmark, []string) {
	var warnings []string

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []Bookmark{}, warnings
	}

	// Parse TOML file
	var cfg bookmarksConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		warnings = append(warnings, fmt.Sprintf("Warning: failed to parse bookmarks: %v", err))
		return []Bookmark{}, warnings
	}

	// Validate entries
	var valid []Bookmark
	for _, b := range cfg.Bookmarks {
		if b.Name == "" {
			warnings = append(warnings, "Warning: skipping bookmark with empty name")
			continue
		}
		if b.Path == "" {
			warnings = append(warnings, fmt.Sprintf("Warning: skipping bookmark '%s' with empty path", b.Name))
			continue
		}
		valid = append(valid, b)
	}

	return valid, warnings
}

// SaveBookmarks saves bookmarks to a TOML configuration file.
// Preserves existing sections (keybindings, colors).
func SaveBookmarks(path string, bookmarks []Bookmark) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Read existing content if file exists
	var existingContent string
	if data, err := os.ReadFile(path); err == nil {
		existingContent = string(data)
	}

	// Remove existing bookmarks sections
	newContent := removeBookmarksSection(existingContent)

	// Build bookmarks TOML
	var bookmarksTOML strings.Builder
	for _, b := range bookmarks {
		bookmarksTOML.WriteString("[[bookmarks]]\n")
		bookmarksTOML.WriteString(fmt.Sprintf("name = %q\n", b.Name))
		bookmarksTOML.WriteString(fmt.Sprintf("path = %q\n", b.Path))
		bookmarksTOML.WriteString("\n")
	}

	// Append bookmarks to content
	if newContent != "" && !strings.HasSuffix(newContent, "\n\n") {
		if !strings.HasSuffix(newContent, "\n") {
			newContent += "\n"
		}
		newContent += "\n"
	}
	newContent += bookmarksTOML.String()

	// Write to file
	return os.WriteFile(path, []byte(newContent), 0644)
}

// removeBookmarksSection removes all [[bookmarks]] sections from TOML content.
func removeBookmarksSection(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inBookmarksSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if we're entering a bookmarks section
		if trimmed == "[[bookmarks]]" {
			inBookmarksSection = true
			continue
		}

		// Check if we're entering a different section
		if strings.HasPrefix(trimmed, "[") {
			inBookmarksSection = false
		}

		// Skip lines that are part of bookmarks section
		if inBookmarksSection {
			continue
		}

		result = append(result, line)
	}

	// Join and clean up extra blank lines
	output := strings.Join(result, "\n")
	output = regexp.MustCompile(`\n{3,}`).ReplaceAllString(output, "\n\n")
	output = strings.TrimRight(output, "\n")
	if output != "" {
		output += "\n"
	}

	return output
}

// AddBookmark adds a new bookmark to the beginning of the list.
// Returns ErrEmptyAlias if name is empty.
// Returns ErrDuplicatePath if path is already bookmarked.
// The path is normalized using filepath.Clean before storing.
func AddBookmark(bookmarks []Bookmark, name, path string) ([]Bookmark, error) {
	if name == "" {
		return bookmarks, ErrEmptyAlias
	}
	normalizedPath := filepath.Clean(path)
	if IsPathBookmarked(bookmarks, normalizedPath) {
		return bookmarks, ErrDuplicatePath
	}

	newBookmark := Bookmark{Name: name, Path: normalizedPath}
	return append([]Bookmark{newBookmark}, bookmarks...), nil
}

// RemoveBookmark removes a bookmark at the given index.
// Returns a new slice without modifying the original.
func RemoveBookmark(bookmarks []Bookmark, index int) ([]Bookmark, error) {
	if index < 0 || index >= len(bookmarks) {
		return bookmarks, ErrInvalidIndex
	}

	result := make([]Bookmark, 0, len(bookmarks)-1)
	result = append(result, bookmarks[:index]...)
	result = append(result, bookmarks[index+1:]...)
	return result, nil
}

// UpdateBookmarkAlias updates the alias of a bookmark at the given index.
func UpdateBookmarkAlias(bookmarks []Bookmark, index int, newAlias string) ([]Bookmark, error) {
	if index < 0 || index >= len(bookmarks) {
		return bookmarks, ErrInvalidIndex
	}
	if newAlias == "" {
		return bookmarks, ErrEmptyAlias
	}

	result := make([]Bookmark, len(bookmarks))
	copy(result, bookmarks)
	result[index].Name = newAlias
	return result, nil
}

// IsPathBookmarked checks if a path is already bookmarked.
// Paths are normalized using filepath.Clean before comparison.
func IsPathBookmarked(bookmarks []Bookmark, path string) bool {
	normalizedPath := filepath.Clean(path)
	for _, b := range bookmarks {
		if filepath.Clean(b.Path) == normalizedPath {
			return true
		}
	}
	return false
}

// DefaultAliasFromPath returns the default alias for a path (the basename).
func DefaultAliasFromPath(path string) string {
	base := filepath.Base(path)
	if base == "" || base == "." {
		return path
	}
	return base
}
