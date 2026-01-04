package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/config"
)

// BookmarkManager manages bookmark state and operations.
type BookmarkManager struct {
	bookmarks []config.Bookmark
	editIndex int // Index of bookmark being edited (-1 if none)
}

// NewBookmarkManager creates a new bookmark manager.
// It loads bookmarks from the configuration file.
func NewBookmarkManager() (*BookmarkManager, []string) {
	var bookmarks []config.Bookmark
	var warnings []string

	configPath, err := config.GetConfigPath()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Warning: failed to get config path: %v", err))
	} else {
		var bookmarkWarnings []string
		bookmarks, bookmarkWarnings = config.LoadBookmarks(configPath)
		warnings = append(warnings, bookmarkWarnings...)
	}

	return &BookmarkManager{
		bookmarks: bookmarks,
		editIndex: -1,
	}, warnings
}

// Bookmarks returns the current list of bookmarks.
func (m *BookmarkManager) Bookmarks() []config.Bookmark {
	return m.bookmarks
}

// SetBookmarks updates the bookmark list.
func (m *BookmarkManager) SetBookmarks(bookmarks []config.Bookmark) {
	m.bookmarks = bookmarks
}

// EditIndex returns the current edit index.
func (m *BookmarkManager) EditIndex() int {
	return m.editIndex
}

// SetEditIndex sets the edit index.
func (m *BookmarkManager) SetEditIndex(index int) {
	m.editIndex = index
}

// ClearEditIndex resets the edit index to -1.
func (m *BookmarkManager) ClearEditIndex() {
	m.editIndex = -1
}

// Add adds a new bookmark at the given path with the given alias.
// Returns a tea.Cmd that produces the appropriate message.
func (m *BookmarkManager) Add(path, alias string) tea.Cmd {
	return func() tea.Msg {
		newBookmarks, err := config.AddBookmark(m.bookmarks, alias, path)
		if err != nil {
			if err == config.ErrEmptyAlias {
				return showStatusMsg{message: "Bookmark name cannot be empty", isError: true}
			}
			if err == config.ErrDuplicatePath {
				return showStatusMsg{message: "Already bookmarked", isError: false}
			}
			return showStatusMsg{message: fmt.Sprintf("Failed to add bookmark: %v", err), isError: true}
		}

		// Save to config file
		if saveErr := m.save(newBookmarks); saveErr != nil {
			return showStatusMsg{message: saveErr.Error(), isError: true}
		}

		return bookmarkAddedMsg{bookmarks: newBookmarks, alias: alias}
	}
}

// Edit updates an existing bookmark's alias.
// Returns a tea.Cmd that produces the appropriate message.
func (m *BookmarkManager) Edit(index int, newAlias string) tea.Cmd {
	return func() tea.Msg {
		if index < 0 || index >= len(m.bookmarks) {
			return showStatusMsg{message: "Invalid bookmark index", isError: true}
		}

		newBookmarks, err := config.UpdateBookmarkAlias(m.bookmarks, index, newAlias)
		if err != nil {
			if err == config.ErrEmptyAlias {
				return showStatusMsg{message: "Bookmark name cannot be empty", isError: true}
			}
			return showStatusMsg{message: fmt.Sprintf("Failed to edit bookmark: %v", err), isError: true}
		}

		// Save to config file
		if saveErr := m.save(newBookmarks); saveErr != nil {
			return showStatusMsg{message: saveErr.Error(), isError: true}
		}

		return bookmarkEditedMsg{bookmarks: newBookmarks, alias: newAlias}
	}
}

// Delete removes a bookmark at the given index.
// Returns a tea.Cmd that produces the appropriate message.
func (m *BookmarkManager) Delete(index int) tea.Cmd {
	return func() tea.Msg {
		if index < 0 || index >= len(m.bookmarks) {
			return showStatusMsg{message: "Invalid bookmark index", isError: true}
		}

		newBookmarks := append(m.bookmarks[:index], m.bookmarks[index+1:]...)

		// Save to config file
		if saveErr := m.save(newBookmarks); saveErr != nil {
			return showStatusMsg{message: saveErr.Error(), isError: true}
		}

		return bookmarkDeletedMsg{bookmarks: newBookmarks}
	}
}

// save saves the bookmarks to the configuration file.
func (m *BookmarkManager) save(bookmarks []config.Bookmark) error {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}
	if err := config.SaveBookmarks(configPath, bookmarks); err != nil {
		return fmt.Errorf("failed to save bookmarks: %w", err)
	}
	return nil
}

// bookmarkDeletedMsg is sent when a bookmark is successfully deleted
type bookmarkDeletedMsg struct {
	bookmarks []config.Bookmark
}
