package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/config"
)

func TestNewBookmarkDialog(t *testing.T) {
	bookmarks := []config.Bookmark{
		{Name: "Projects", Path: "/path/to/projects"},
		{Name: "Downloads", Path: "/path/to/downloads"},
	}

	dialog := NewBookmarkDialog(bookmarks)

	if !dialog.IsActive() {
		t.Error("expected dialog to be active")
	}
	if dialog.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", dialog.cursor)
	}
	if len(dialog.bookmarks) != 2 {
		t.Errorf("expected 2 bookmarks, got %d", len(dialog.bookmarks))
	}
}

func TestBookmarkDialogNavigation(t *testing.T) {
	bookmarks := []config.Bookmark{
		{Name: "First", Path: "/first"},
		{Name: "Second", Path: "/second"},
		{Name: "Third", Path: "/third"},
	}

	t.Run("j moves cursor down", func(t *testing.T) {
		dialog := NewBookmarkDialog(bookmarks)
		dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

		if dialog.cursor != 1 {
			t.Errorf("expected cursor at 1, got %d", dialog.cursor)
		}
	})

	t.Run("k moves cursor up", func(t *testing.T) {
		dialog := NewBookmarkDialog(bookmarks)
		dialog.cursor = 2
		dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

		if dialog.cursor != 1 {
			t.Errorf("expected cursor at 1, got %d", dialog.cursor)
		}
	})

	t.Run("j at last item wraps to 0", func(t *testing.T) {
		dialog := NewBookmarkDialog(bookmarks)
		dialog.cursor = 2
		dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

		if dialog.cursor != 0 {
			t.Errorf("expected cursor to wrap to 0, got %d", dialog.cursor)
		}
	})

	t.Run("k at first item wraps to last", func(t *testing.T) {
		dialog := NewBookmarkDialog(bookmarks)
		dialog.cursor = 0
		dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

		if dialog.cursor != 2 {
			t.Errorf("expected cursor to wrap to 2, got %d", dialog.cursor)
		}
	})

	t.Run("down arrow moves cursor down", func(t *testing.T) {
		dialog := NewBookmarkDialog(bookmarks)
		dialog.Update(tea.KeyMsg{Type: tea.KeyDown})

		if dialog.cursor != 1 {
			t.Errorf("expected cursor at 1, got %d", dialog.cursor)
		}
	})

	t.Run("up arrow moves cursor up", func(t *testing.T) {
		dialog := NewBookmarkDialog(bookmarks)
		dialog.cursor = 1
		dialog.Update(tea.KeyMsg{Type: tea.KeyUp})

		if dialog.cursor != 0 {
			t.Errorf("expected cursor at 0, got %d", dialog.cursor)
		}
	})
}

func TestBookmarkDialogActions(t *testing.T) {
	bookmarks := []config.Bookmark{
		{Name: "Test", Path: "/tmp"}, // Use /tmp which exists
	}

	t.Run("Enter on existing path returns jump message", func(t *testing.T) {
		dialog := NewBookmarkDialog(bookmarks)
		// Mark path as existing
		dialog.pathExists[0] = true

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cmd == nil {
			t.Error("expected command for jump")
		}
	})

	t.Run("Enter on non-existent path does nothing", func(t *testing.T) {
		dialog := NewBookmarkDialog(bookmarks)
		dialog.pathExists[0] = false

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cmd != nil {
			t.Error("expected no command for non-existent path")
		}
	})

	t.Run("D returns delete message", func(t *testing.T) {
		dialog := NewBookmarkDialog(bookmarks)

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

		if cmd == nil {
			t.Error("expected command for delete")
		}
	})

	t.Run("E returns edit message", func(t *testing.T) {
		dialog := NewBookmarkDialog(bookmarks)

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

		if cmd == nil {
			t.Error("expected command for edit")
		}
	})

	t.Run("Esc closes dialog", func(t *testing.T) {
		dialog := NewBookmarkDialog(bookmarks)

		dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

		if dialog.IsActive() {
			t.Error("expected dialog to be inactive after Esc")
		}
	})
}

func TestBookmarkDialogView(t *testing.T) {
	t.Run("displays bookmarks in two-line format", func(t *testing.T) {
		bookmarks := []config.Bookmark{
			{Name: "Projects", Path: "/path/to/projects"},
		}
		dialog := NewBookmarkDialog(bookmarks)
		dialog.pathExists[0] = true

		view := dialog.View()

		if !strings.Contains(view, "Projects") {
			t.Error("expected view to contain bookmark name")
		}
		if !strings.Contains(view, "/path/to/projects") {
			t.Error("expected view to contain bookmark path")
		}
	})

	t.Run("shows warning indicator for non-existent paths", func(t *testing.T) {
		bookmarks := []config.Bookmark{
			{Name: "Missing", Path: "/nonexistent/path"},
		}
		dialog := NewBookmarkDialog(bookmarks)
		dialog.pathExists[0] = false

		view := dialog.View()

		// Should contain warning emoji
		if !strings.Contains(view, "\u26a0") {
			t.Error("expected view to contain warning indicator")
		}
	})

	t.Run("shows footer with key hints", func(t *testing.T) {
		bookmarks := []config.Bookmark{
			{Name: "Test", Path: "/test"},
		}
		dialog := NewBookmarkDialog(bookmarks)

		view := dialog.View()

		if !strings.Contains(view, "Enter") {
			t.Error("expected view to contain Enter hint")
		}
		if !strings.Contains(view, "D") {
			t.Error("expected view to contain D hint")
		}
		if !strings.Contains(view, "E") {
			t.Error("expected view to contain E hint")
		}
		if !strings.Contains(view, "Esc") {
			t.Error("expected view to contain Esc hint")
		}
	})

	t.Run("shows empty message when no bookmarks", func(t *testing.T) {
		dialog := NewBookmarkDialog([]config.Bookmark{})

		view := dialog.View()

		if !strings.Contains(view, "No bookmarks") {
			t.Error("expected view to show 'No bookmarks' message")
		}
	})
}

func TestBookmarkDialogDisplayType(t *testing.T) {
	dialog := NewBookmarkDialog([]config.Bookmark{})

	if dialog.DisplayType() != DialogDisplayPane {
		t.Error("expected DialogDisplayPane")
	}
}

func TestBookmarkResultMsg(t *testing.T) {
	t.Run("jump result contains path", func(t *testing.T) {
		msg := bookmarkJumpMsg{path: "/test/path"}
		if msg.path != "/test/path" {
			t.Errorf("expected path '/test/path', got '%s'", msg.path)
		}
	})

	t.Run("delete result contains index", func(t *testing.T) {
		msg := bookmarkDeleteMsg{index: 5}
		if msg.index != 5 {
			t.Errorf("expected index 5, got %d", msg.index)
		}
	})

	t.Run("edit result contains index and bookmark", func(t *testing.T) {
		b := config.Bookmark{Name: "Test", Path: "/test"}
		msg := bookmarkEditMsg{index: 3, bookmark: b}
		if msg.index != 3 {
			t.Errorf("expected index 3, got %d", msg.index)
		}
		if msg.bookmark.Name != "Test" {
			t.Errorf("expected name 'Test', got '%s'", msg.bookmark.Name)
		}
	})
}

func TestBookmarkDialog_SetWidth(t *testing.T) {
	bookmarks := []config.Bookmark{
		{Name: "Test", Path: "/test"},
	}
	dialog := NewBookmarkDialog(bookmarks)

	initialWidth := dialog.width
	newWidth := 100

	dialog.SetWidth(newWidth)

	if dialog.width != newWidth {
		t.Errorf("SetWidth() did not set width, got %d, want %d", dialog.width, newWidth)
	}

	if dialog.width == initialWidth && newWidth != initialWidth {
		t.Error("SetWidth() should change the width")
	}
}

func TestBookmarkDialog_WrapPath(t *testing.T) {
	bookmarks := []config.Bookmark{
		{Name: "Test", Path: "/test"},
	}
	dialog := NewBookmarkDialog(bookmarks)

	t.Run("short path returns unchanged", func(t *testing.T) {
		path := "/short/path"
		result := dialog.wrapPath(path, 50)
		if result != path {
			t.Errorf("wrapPath() = %q, want %q", result, path)
		}
	})

	t.Run("long path gets wrapped", func(t *testing.T) {
		path := "/very/long/path/that/exceeds/the/maximum/width/allowed"
		result := dialog.wrapPath(path, 20)
		if !strings.Contains(result, "\n") {
			t.Error("wrapPath() should wrap long paths with newlines")
		}
	})

	t.Run("path exactly at max width", func(t *testing.T) {
		path := "/exact"
		result := dialog.wrapPath(path, 6)
		if result != path {
			t.Errorf("wrapPath() = %q, want %q", result, path)
		}
	})

	t.Run("empty path", func(t *testing.T) {
		path := ""
		result := dialog.wrapPath(path, 50)
		if result != path {
			t.Errorf("wrapPath() = %q, want %q", result, path)
		}
	})
}

func TestBookmarkDialogWithVeryLongPath(t *testing.T) {
	longPath := "/very/long/path/that/exceeds/the/normal/display/width/and/should/be/wrapped"
	bookmarks := []config.Bookmark{
		{Name: "LongPath", Path: longPath},
	}
	dialog := NewBookmarkDialog(bookmarks)
	dialog.width = 30

	view := dialog.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	if !strings.Contains(view, "LongPath") {
		t.Error("View should contain bookmark name")
	}
}
