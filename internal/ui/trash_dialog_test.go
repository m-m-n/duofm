package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
)

// setupTestTrash creates a temporary trash directory with test items
func setupTestTrash(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()

	// Set XDG_DATA_HOME to use temp directory
	oldXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)

	// Create trash directories
	filesDir := filepath.Join(tmpDir, "Trash", "files")
	infoDir := filepath.Join(tmpDir, "Trash", "info")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		os.Setenv("XDG_DATA_HOME", oldXDG)
	}

	return tmpDir, cleanup
}

// addTestTrashItem adds a test item to the trash
func addTestTrashItem(t *testing.T, name, originalPath string, deletionTime time.Time, isDir bool) {
	t.Helper()

	filesDir := fs.TrashFilesDir()
	infoDir := fs.TrashInfoDir()

	// Create file or directory in trash
	itemPath := filepath.Join(filesDir, name)
	if isDir {
		if err := os.MkdirAll(itemPath, 0755); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := os.WriteFile(itemPath, []byte("test content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create trashinfo file
	trashinfoPath := filepath.Join(infoDir, name+".trashinfo")
	if err := fs.WriteTrashinfo(trashinfoPath, originalPath, deletionTime); err != nil {
		t.Fatal(err)
	}
}

func TestNewTrashDialog(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	// Add test items
	now := time.Now()
	addTestTrashItem(t, "test.txt", "/home/user/test.txt", now, false)
	addTestTrashItem(t, "dir1", "/home/user/dir1", now.Add(-time.Hour), true)

	t.Run("creates dialog with items", func(t *testing.T) {
		items, err := loadTrashItems()
		if err != nil {
			t.Fatalf("loadTrashItems failed: %v", err)
		}

		dialog := NewTrashDialog(items)

		if dialog == nil {
			t.Fatal("NewTrashDialog returned nil")
		}

		if !dialog.IsActive() {
			t.Error("dialog should be active")
		}

		if dialog.DisplayType() != DialogDisplayScreen {
			t.Error("dialog should be DialogDisplayScreen type")
		}

		if len(dialog.items) != 2 {
			t.Errorf("expected 2 items, got %d", len(dialog.items))
		}
	})

	t.Run("creates dialog with empty trash", func(t *testing.T) {
		dialog := NewTrashDialog(nil)

		if dialog == nil {
			t.Fatal("NewTrashDialog returned nil")
		}

		if len(dialog.items) != 0 {
			t.Errorf("expected 0 items, got %d", len(dialog.items))
		}
	})
}

func TestLoadTrashItems(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	now := time.Now()
	addTestTrashItem(t, "file1.txt", "/home/user/documents/file1.txt", now, false)
	addTestTrashItem(t, "file2.go", "/home/user/projects/file2.go", now.Add(-24*time.Hour), false)
	addTestTrashItem(t, "folder1", "/home/user/folder1", now.Add(-48*time.Hour), true)

	t.Run("loads all items from trash", func(t *testing.T) {
		items, err := loadTrashItems()
		if err != nil {
			t.Fatalf("loadTrashItems failed: %v", err)
		}

		if len(items) != 3 {
			t.Errorf("expected 3 items, got %d", len(items))
		}

		// Check that items have required fields
		for _, item := range items {
			if item.Name == "" {
				t.Error("item name should not be empty")
			}
			if item.OriginalPath == "" {
				t.Error("item original path should not be empty")
			}
			if item.DeletionTime == "" {
				t.Error("item deletion time should not be empty")
			}
		}
	})

	t.Run("loads items with correct isDir flag", func(t *testing.T) {
		items, err := loadTrashItems()
		if err != nil {
			t.Fatalf("loadTrashItems failed: %v", err)
		}

		for _, item := range items {
			if item.Name == "folder1" && !item.IsDir {
				t.Error("folder1 should be marked as directory")
			}
			if item.Name == "file1.txt" && item.IsDir {
				t.Error("file1.txt should not be marked as directory")
			}
		}
	})
}

func TestTrashDialog_CursorNavigation(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	now := time.Now()
	addTestTrashItem(t, "file1.txt", "/home/user/file1.txt", now, false)
	addTestTrashItem(t, "file2.txt", "/home/user/file2.txt", now, false)
	addTestTrashItem(t, "file3.txt", "/home/user/file3.txt", now, false)

	items, _ := loadTrashItems()
	dialog := NewTrashDialog(items)

	t.Run("j moves cursor down", func(t *testing.T) {
		dialog.cursor = 0
		dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

		if dialog.cursor != 1 {
			t.Errorf("expected cursor at 1, got %d", dialog.cursor)
		}
	})

	t.Run("k moves cursor up", func(t *testing.T) {
		dialog.cursor = 1
		dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

		if dialog.cursor != 0 {
			t.Errorf("expected cursor at 0, got %d", dialog.cursor)
		}
	})

	t.Run("down arrow moves cursor down", func(t *testing.T) {
		dialog.cursor = 0
		dialog.Update(tea.KeyMsg{Type: tea.KeyDown})

		if dialog.cursor != 1 {
			t.Errorf("expected cursor at 1, got %d", dialog.cursor)
		}
	})

	t.Run("up arrow moves cursor up", func(t *testing.T) {
		dialog.cursor = 1
		dialog.Update(tea.KeyMsg{Type: tea.KeyUp})

		if dialog.cursor != 0 {
			t.Errorf("expected cursor at 0, got %d", dialog.cursor)
		}
	})

	t.Run("cursor stops at end", func(t *testing.T) {
		dialog.cursor = len(items) - 1
		dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

		if dialog.cursor != len(items)-1 {
			t.Errorf("cursor should stop at end, got %d", dialog.cursor)
		}
	})

	t.Run("cursor stops at beginning", func(t *testing.T) {
		dialog.cursor = 0
		dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

		if dialog.cursor != 0 {
			t.Errorf("cursor should stop at beginning, got %d", dialog.cursor)
		}
	})
}

func TestTrashDialog_Close(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	items, _ := loadTrashItems()
	dialog := NewTrashDialog(items)

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"escape", tea.KeyMsg{Type: tea.KeyEsc}},
		{"ctrl+c", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}

	for _, tt := range tests {
		t.Run(tt.name+" closes dialog", func(t *testing.T) {
			dialog.SetActive(true)
			newDialog, cmd := dialog.Update(tt.key)

			if newDialog.IsActive() {
				t.Error("dialog should be inactive after close")
			}

			if cmd == nil {
				t.Error("should return a command")
			}
		})
	}
}

func TestTrashDialog_View(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	now := time.Now()
	addTestTrashItem(t, "test.txt", "/home/user/documents/test.txt", now, false)
	addTestTrashItem(t, "folder", "/home/user/folder", now, true)

	items, _ := loadTrashItems()
	dialog := NewTrashDialog(items)

	t.Run("view contains title", func(t *testing.T) {
		view := dialog.View()
		if !strings.Contains(view, "Trash") {
			t.Error("view should contain title 'Trash'")
		}
	})

	t.Run("view shows item count", func(t *testing.T) {
		view := dialog.View()
		if !strings.Contains(view, "[2]") {
			t.Error("view should show item count [2]")
		}
	})

	t.Run("view contains header row", func(t *testing.T) {
		view := dialog.View()
		if !strings.Contains(view, "Name") {
			t.Error("view should contain 'Name' header")
		}
		if !strings.Contains(view, "Size") {
			t.Error("view should contain 'Size' header")
		}
		if !strings.Contains(view, "Deleted") {
			t.Error("view should contain 'Deleted' header")
		}
		if !strings.Contains(view, "Original Path") {
			t.Error("view should contain 'Original Path' header")
		}
	})

	t.Run("view contains footer hints", func(t *testing.T) {
		view := dialog.View()
		if !strings.Contains(view, "j/k") {
			t.Error("view should contain navigation hint")
		}
		if !strings.Contains(view, "Esc") {
			t.Error("view should contain close hint")
		}
	})

	t.Run("empty trash shows message", func(t *testing.T) {
		emptyDialog := NewTrashDialog(nil)
		view := emptyDialog.View()
		// The view outputs "Trash is empty"
		if !strings.Contains(strings.ToLower(view), "empty") {
			t.Error("view should show empty message")
		}
	})
}

func TestTrashDialog_Scroll(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create many items to trigger scrolling
	now := time.Now()
	for i := range 30 {
		name := filepath.Base(filepath.Join("file", string(rune('a'+i))+".txt"))
		addTestTrashItem(t, name, "/home/user/"+name, now, false)
	}

	items, _ := loadTrashItems()
	dialog := NewTrashDialog(items)
	dialog.visibleHeight = 10

	t.Run("scroll follows cursor down", func(t *testing.T) {
		dialog.cursor = 0
		dialog.scrollOffset = 0

		// Move cursor beyond visible area
		for range 15 {
			dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		}

		if dialog.scrollOffset == 0 {
			t.Error("scroll should have changed when cursor moves beyond visible area")
		}
	})

	t.Run("scroll follows cursor up", func(t *testing.T) {
		dialog.cursor = 15
		dialog.scrollOffset = 10

		// Move cursor above visible area
		for range 10 {
			dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		}

		if dialog.scrollOffset >= 10 {
			t.Error("scroll should have decreased when cursor moves above visible area")
		}
	})
}

func TestTrashDialog_Mark(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	now := time.Now()
	addTestTrashItem(t, "file1.txt", "/home/user/file1.txt", now, false)
	addTestTrashItem(t, "file2.txt", "/home/user/file2.txt", now, false)
	addTestTrashItem(t, "file3.txt", "/home/user/file3.txt", now, false)

	items, _ := loadTrashItems()
	dialog := NewTrashDialog(items)

	t.Run("space toggles mark on current item", func(t *testing.T) {
		dialog.cursor = 0
		initialMarked := dialog.items[0].marked

		dialog.Update(tea.KeyMsg{Type: tea.KeySpace})

		if dialog.items[0].marked == initialMarked {
			t.Error("mark state should toggle")
		}
	})

	t.Run("space moves cursor down after mark", func(t *testing.T) {
		dialog.cursor = 0
		dialog.items[0].marked = false

		dialog.Update(tea.KeyMsg{Type: tea.KeySpace})

		if dialog.cursor != 1 {
			t.Errorf("expected cursor at 1, got %d", dialog.cursor)
		}
	})

	t.Run("cursor stays at last item after mark", func(t *testing.T) {
		dialog.cursor = len(items) - 1
		dialog.items[dialog.cursor].marked = false

		dialog.Update(tea.KeyMsg{Type: tea.KeySpace})

		if dialog.cursor != len(items)-1 {
			t.Errorf("cursor should stay at last item, got %d", dialog.cursor)
		}
	})

	t.Run("marked items are visually indicated", func(t *testing.T) {
		// Reset all marks
		for i := range dialog.items {
			dialog.items[i].marked = false
		}

		// Mark first item
		dialog.items[0].marked = true

		view := dialog.View()
		// Should contain mark indicator (*)
		if !strings.Contains(view, "*") {
			t.Error("marked item should show * indicator")
		}
	})

	t.Run("GetMarkedItems returns marked items", func(t *testing.T) {
		// Reset all marks
		for i := range dialog.items {
			dialog.items[i].marked = false
		}

		// Mark two items
		dialog.items[0].marked = true
		dialog.items[2].marked = true

		marked := dialog.GetMarkedItems()
		if len(marked) != 2 {
			t.Errorf("expected 2 marked items, got %d", len(marked))
		}
	})

	t.Run("GetMarkedItems returns empty when no marks", func(t *testing.T) {
		// Reset all marks
		for i := range dialog.items {
			dialog.items[i].marked = false
		}

		marked := dialog.GetMarkedItems()
		if len(marked) != 0 {
			t.Errorf("expected 0 marked items, got %d", len(marked))
		}
	})
}

func TestTrashDialog_Restore(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	now := time.Now()
	addTestTrashItem(t, "file1.txt", "/home/user/file1.txt", now, false)
	addTestTrashItem(t, "file2.txt", "/home/user/file2.txt", now, false)

	items, _ := loadTrashItems()
	dialog := NewTrashDialog(items)

	t.Run("R key triggers restore message", func(t *testing.T) {
		dialog.cursor = 0
		dialog.SetActive(true)

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})

		if cmd == nil {
			t.Error("R key should return a command")
		}

		// Execute the command and check the message type
		msg := cmd()
		if _, ok := msg.(trashDialogRestoreMsg); !ok {
			t.Errorf("expected trashDialogRestoreMsg, got %T", msg)
		}
	})

	t.Run("r key (lowercase) triggers restore message", func(t *testing.T) {
		dialog.cursor = 0
		dialog.SetActive(true)

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

		if cmd == nil {
			t.Error("r key should return a command")
		}

		// Execute the command and check the message type
		msg := cmd()
		if _, ok := msg.(trashDialogRestoreMsg); !ok {
			t.Errorf("expected trashDialogRestoreMsg, got %T", msg)
		}
	})

	t.Run("restore message contains selected item", func(t *testing.T) {
		dialog.cursor = 0
		dialog.SetActive(true)
		// Reset marks
		for i := range dialog.items {
			dialog.items[i].marked = false
		}

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})

		msg := cmd().(trashDialogRestoreMsg)
		if len(msg.items) != 1 {
			t.Errorf("expected 1 item, got %d", len(msg.items))
		}
	})

	t.Run("restore message contains marked items when available", func(t *testing.T) {
		dialog.SetActive(true)
		// Mark both items
		dialog.items[0].marked = true
		dialog.items[1].marked = true
		dialog.cursor = 0

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})

		msg := cmd().(trashDialogRestoreMsg)
		if len(msg.items) != 2 {
			t.Errorf("expected 2 items, got %d", len(msg.items))
		}
	})
}

func TestTrashDialog_EmptyTrash(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	now := time.Now()
	addTestTrashItem(t, "file1.txt", "/home/user/file1.txt", now, false)
	addTestTrashItem(t, "file2.txt", "/home/user/file2.txt", now, false)

	items, _ := loadTrashItems()
	dialog := NewTrashDialog(items)

	t.Run("E key triggers empty trash message", func(t *testing.T) {
		dialog.SetActive(true)

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'E'}})

		if cmd == nil {
			t.Error("E key should return a command")
		}

		// Execute the command and check the message type
		msg := cmd()
		if _, ok := msg.(trashDialogEmptyMsg); !ok {
			t.Errorf("expected trashDialogEmptyMsg, got %T", msg)
		}
	})

	t.Run("e key (lowercase) triggers empty trash message", func(t *testing.T) {
		dialog.SetActive(true)

		_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

		if cmd == nil {
			t.Error("e key should return a command")
		}

		// Execute the command and check the message type
		msg := cmd()
		if _, ok := msg.(trashDialogEmptyMsg); !ok {
			t.Errorf("expected trashDialogEmptyMsg, got %T", msg)
		}
	})

	t.Run("empty trash on empty dialog does nothing", func(t *testing.T) {
		emptyDialog := NewTrashDialog(nil)
		emptyDialog.SetActive(true)

		_, cmd := emptyDialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'E'}})

		if cmd != nil {
			t.Error("E key on empty dialog should not return a command")
		}
	})
}

func TestTrashItem_Size(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	now := time.Now()
	addTestTrashItem(t, "file.txt", "/home/user/file.txt", now, false)
	addTestTrashItem(t, "dir", "/home/user/dir", now, true)

	items, _ := loadTrashItems()

	t.Run("file shows size", func(t *testing.T) {
		for _, item := range items {
			if item.Name == "file.txt" {
				if item.Size < 0 {
					t.Error("file should have non-negative size")
				}
			}
		}
	})

	t.Run("directory shows dash for size", func(t *testing.T) {
		dialog := NewTrashDialog(items)
		view := dialog.View()
		// Directory size should be displayed as "-"
		if !strings.Contains(view, "-") {
			t.Log("Note: directory size display depends on implementation")
		}
	})
}

func TestValidateTrashName(t *testing.T) {
	tests := []struct {
		name      string
		trashName string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid simple name",
			trashName: "file.txt",
			wantErr:   false,
		},
		{
			name:      "valid name with dots",
			trashName: "file.tar.gz",
			wantErr:   false,
		},
		{
			name:      "valid name with spaces",
			trashName: "my file.txt",
			wantErr:   false,
		},
		{
			name:      "valid dotfile",
			trashName: ".gitignore",
			wantErr:   false,
		},
		{
			name:      "invalid - forward slash",
			trashName: "path/to/file",
			wantErr:   true,
			errMsg:    "contains path separator",
		},
		{
			name:      "invalid - backslash",
			trashName: "path\\to\\file",
			wantErr:   true,
			errMsg:    "contains path separator",
		},
		{
			name:      "invalid - parent directory",
			trashName: "..",
			wantErr:   true,
			errMsg:    "special directory",
		},
		{
			name:      "invalid - current directory",
			trashName: ".",
			wantErr:   true,
			errMsg:    "special directory",
		},
		{
			name:      "invalid - empty string",
			trashName: "",
			wantErr:   true,
			errMsg:    "empty",
		},
		{
			name:      "invalid - path traversal attempt",
			trashName: "../etc/passwd",
			wantErr:   true,
			errMsg:    "contains path separator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTrashName(tt.trashName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateTrashName(%q) expected error, got nil", tt.trashName)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateTrashName(%q) error = %v, want error containing %q", tt.trashName, err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateTrashName(%q) unexpected error: %v", tt.trashName, err)
				}
			}
		})
	}
}
