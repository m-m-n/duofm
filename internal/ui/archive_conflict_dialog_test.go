package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewArchiveConflictDialog(t *testing.T) {
	t.Run("with non-existent file", func(t *testing.T) {
		dialog := NewArchiveConflictDialog("/tmp/non-existent-test-12345.zip")

		if dialog == nil {
			t.Fatal("NewArchiveConflictDialog() returned nil")
		}

		if !dialog.active {
			t.Error("dialog should be active by default")
		}

		if dialog.cursor != 0 {
			t.Errorf("dialog.cursor = %d, want 0", dialog.cursor)
		}

		if dialog.width != 55 {
			t.Errorf("dialog.width = %d, want 55", dialog.width)
		}

		if dialog.filename != "non-existent-test-12345.zip" {
			t.Errorf("dialog.filename = %s, want non-existent-test-12345.zip", dialog.filename)
		}

		if dialog.destDir != "/tmp" {
			t.Errorf("dialog.destDir = %s, want /tmp", dialog.destDir)
		}
	})

	t.Run("with existing file", func(t *testing.T) {
		// Create temp file
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.zip")
		err := os.WriteFile(tmpFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		dialog := NewArchiveConflictDialog(tmpFile)

		if dialog == nil {
			t.Fatal("NewArchiveConflictDialog() returned nil")
		}

		if dialog.filename != "test.zip" {
			t.Errorf("dialog.filename = %s, want test.zip", dialog.filename)
		}

		if dialog.destDir != tmpDir {
			t.Errorf("dialog.destDir = %s, want %s", dialog.destDir, tmpDir)
		}

		if dialog.existingSize <= 0 {
			t.Errorf("dialog.existingSize = %d, should be > 0", dialog.existingSize)
		}

		if dialog.existingMod.IsZero() {
			t.Error("dialog.existingMod should not be zero")
		}
	})
}

func TestArchiveConflictDialogUpdate_Navigation(t *testing.T) {
	tests := []struct {
		name           string
		keyType        tea.KeyType
		keyRunes       []rune
		initialCursor  int
		expectedCursor int
	}{
		{"j moves down from 0", tea.KeyRunes, []rune{'j'}, 0, 1},
		{"j moves down from 1", tea.KeyRunes, []rune{'j'}, 1, 2},
		{"j wraps from 2 to 0", tea.KeyRunes, []rune{'j'}, 2, 0},
		{"k moves up from 1", tea.KeyRunes, []rune{'k'}, 1, 0},
		{"k moves up from 2", tea.KeyRunes, []rune{'k'}, 2, 1},
		{"k wraps from 0 to 2", tea.KeyRunes, []rune{'k'}, 0, 2},
		{"down arrow moves down", tea.KeyDown, nil, 0, 1},
		{"up arrow moves up", tea.KeyUp, nil, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewArchiveConflictDialog("/tmp/test.zip")
			dialog.cursor = tt.initialCursor

			var msg tea.KeyMsg
			if tt.keyRunes != nil {
				msg = tea.KeyMsg{Type: tt.keyType, Runes: tt.keyRunes}
			} else {
				msg = tea.KeyMsg{Type: tt.keyType}
			}

			updatedDialog, cmd := dialog.Update(msg)
			if cmd != nil {
				t.Error("navigation should not return a command")
			}

			d := updatedDialog.(*ArchiveConflictDialog)
			if d.cursor != tt.expectedCursor {
				t.Errorf("cursor = %d, want %d", d.cursor, tt.expectedCursor)
			}
		})
	}
}

func TestArchiveConflictDialogUpdate_NumberKeys(t *testing.T) {
	tests := []struct {
		name           string
		key            rune
		expectedChoice ArchiveConflictChoice
	}{
		{"1 selects overwrite", '1', ArchiveConflictOverwrite},
		{"2 selects rename", '2', ArchiveConflictRename},
		{"3 selects cancel", '3', ArchiveConflictCancel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewArchiveConflictDialog("/tmp/test.zip")

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}}
			updatedDialog, cmd := dialog.Update(msg)

			d := updatedDialog.(*ArchiveConflictDialog)
			if d.active {
				t.Error("dialog should be inactive after selection")
			}

			if cmd == nil {
				t.Fatal("command should be returned")
			}

			result := cmd()
			resultMsg, ok := result.(archiveConflictResultMsg)
			if !ok {
				t.Fatal("command should return archiveConflictResultMsg")
			}

			if resultMsg.choice != tt.expectedChoice {
				t.Errorf("resultMsg.choice = %v, want %v", resultMsg.choice, tt.expectedChoice)
			}

			if resultMsg.archivePath != "/tmp/test.zip" {
				t.Errorf("resultMsg.archivePath = %s, want /tmp/test.zip", resultMsg.archivePath)
			}
		})
	}
}

func TestArchiveConflictDialogUpdate_Enter(t *testing.T) {
	tests := []struct {
		name           string
		cursor         int
		expectedChoice ArchiveConflictChoice
	}{
		{"enter at cursor 0 selects overwrite", 0, ArchiveConflictOverwrite},
		{"enter at cursor 1 selects rename", 1, ArchiveConflictRename},
		{"enter at cursor 2 selects cancel", 2, ArchiveConflictCancel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewArchiveConflictDialog("/tmp/test.zip")
			dialog.cursor = tt.cursor

			msg := tea.KeyMsg{Type: tea.KeyEnter}
			updatedDialog, cmd := dialog.Update(msg)

			d := updatedDialog.(*ArchiveConflictDialog)
			if d.active {
				t.Error("dialog should be inactive after Enter")
			}

			if cmd == nil {
				t.Fatal("command should be returned")
			}

			result := cmd()
			resultMsg, ok := result.(archiveConflictResultMsg)
			if !ok {
				t.Fatal("command should return archiveConflictResultMsg")
			}

			if resultMsg.choice != tt.expectedChoice {
				t.Errorf("resultMsg.choice = %v, want %v", resultMsg.choice, tt.expectedChoice)
			}
		})
	}
}

func TestArchiveConflictDialogUpdate_Cancel(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"esc cancels", tea.KeyMsg{Type: tea.KeyEsc}},
		{"ctrl+c cancels", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewArchiveConflictDialog("/tmp/test.zip")

			updatedDialog, cmd := dialog.Update(tt.key)

			d := updatedDialog.(*ArchiveConflictDialog)
			if d.active {
				t.Error("dialog should be inactive after cancel")
			}

			if cmd == nil {
				t.Fatal("command should be returned")
			}

			result := cmd()
			resultMsg, ok := result.(archiveConflictResultMsg)
			if !ok {
				t.Fatal("command should return archiveConflictResultMsg")
			}

			if resultMsg.choice != ArchiveConflictCancel {
				t.Errorf("resultMsg.choice = %v, want ArchiveConflictCancel", resultMsg.choice)
			}

			if !resultMsg.cancelled {
				t.Error("resultMsg.cancelled should be true")
			}
		})
	}
}

func TestArchiveConflictDialogUpdate_Inactive(t *testing.T) {
	dialog := NewArchiveConflictDialog("/tmp/test.zip")
	dialog.active = false

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedDialog, cmd := dialog.Update(msg)

	if updatedDialog == nil {
		t.Fatal("updated dialog should not be nil")
	}

	if cmd != nil {
		t.Error("command should be nil when inactive")
	}
}

func TestArchiveConflictDialogView(t *testing.T) {
	t.Run("active dialog shows content", func(t *testing.T) {
		dialog := NewArchiveConflictDialog("/tmp/test.zip")
		view := dialog.View()

		if view == "" {
			t.Error("view should not be empty for active dialog")
		}

		if !strings.Contains(view, "Archive already exists") {
			t.Error("view should contain title")
		}

		if !strings.Contains(view, "test.zip") {
			t.Error("view should contain filename")
		}

		if !strings.Contains(view, "Overwrite") {
			t.Error("view should contain Overwrite option")
		}

		if !strings.Contains(view, "Rename") {
			t.Error("view should contain Rename option")
		}

		if !strings.Contains(view, "Cancel") {
			t.Error("view should contain Cancel option")
		}
	})

	t.Run("inactive dialog shows empty", func(t *testing.T) {
		dialog := NewArchiveConflictDialog("/tmp/test.zip")
		dialog.active = false
		view := dialog.View()

		if view != "" {
			t.Error("view should be empty for inactive dialog")
		}
	})
}

func TestArchiveConflictDialogIsActive(t *testing.T) {
	t.Run("new dialog is active", func(t *testing.T) {
		dialog := NewArchiveConflictDialog("/tmp/test.zip")
		if !dialog.IsActive() {
			t.Error("new dialog should be active")
		}
	})

	t.Run("cancelled dialog is inactive", func(t *testing.T) {
		dialog := NewArchiveConflictDialog("/tmp/test.zip")
		msg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedDialog, _ := dialog.Update(msg)

		d := updatedDialog.(*ArchiveConflictDialog)
		if d.IsActive() {
			t.Error("cancelled dialog should be inactive")
		}
	})
}

func TestArchiveConflictDialogDisplayType(t *testing.T) {
	dialog := NewArchiveConflictDialog("/tmp/test.zip")
	if dialog.DisplayType() != DialogDisplayScreen {
		t.Errorf("DisplayType() = %v, want DialogDisplayScreen", dialog.DisplayType())
	}
}

func TestFormatFileSizeForDialog(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatFileSizeForDialog(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatFileSizeForDialog(%d) = %s, want %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatModTimeForDialog(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		result := formatModTimeForDialog(time.Time{})
		if result != "unknown" {
			t.Errorf("formatModTimeForDialog(zero) = %s, want unknown", result)
		}
	})

	t.Run("valid time", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := formatModTimeForDialog(tm)
		if result != "2024-01-15 10:30" {
			t.Errorf("formatModTimeForDialog() = %s, want 2024-01-15 10:30", result)
		}
	})
}

func TestTruncatePathForDialog(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		maxLen   int
		expected string
	}{
		{"short path", "/tmp", 10, "/tmp"},
		{"exact length", "/tmp/test", 9, "/tmp/test"},
		{"long path", "/home/user/documents/folder", 15, "...ments/folder"},
		{"very short max", "/long/path", 3, "..."},
		{"edge case max 4", "/long/path", 4, "...h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncatePathForDialog(tt.path, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncatePathForDialog(%s, %d) = %s, want %s", tt.path, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestGenerateUniqueArchiveName(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("non-existent file returns with _1", func(t *testing.T) {
		originalPath := filepath.Join(tmpDir, "test.zip")
		result := GenerateUniqueArchiveName(originalPath)
		expected := filepath.Join(tmpDir, "test_1.zip")
		if result != expected {
			t.Errorf("GenerateUniqueArchiveName() = %s, want %s", result, expected)
		}
	})

	t.Run("existing _1 file returns _2", func(t *testing.T) {
		originalPath := filepath.Join(tmpDir, "exist.zip")
		existPath := filepath.Join(tmpDir, "exist_1.zip")

		// Create exist_1.zip
		err := os.WriteFile(existPath, []byte("test"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		result := GenerateUniqueArchiveName(originalPath)
		expected := filepath.Join(tmpDir, "exist_2.zip")
		if result != expected {
			t.Errorf("GenerateUniqueArchiveName() = %s, want %s", result, expected)
		}
	})

	t.Run("handles tar.gz extension", func(t *testing.T) {
		originalPath := filepath.Join(tmpDir, "archive.tar.gz")
		result := GenerateUniqueArchiveName(originalPath)
		expected := filepath.Join(tmpDir, "archive_1.tar.gz")
		if result != expected {
			t.Errorf("GenerateUniqueArchiveName() = %s, want %s", result, expected)
		}
	})

	t.Run("handles tar.bz2 extension", func(t *testing.T) {
		originalPath := filepath.Join(tmpDir, "archive.tar.bz2")
		result := GenerateUniqueArchiveName(originalPath)
		expected := filepath.Join(tmpDir, "archive_1.tar.bz2")
		if result != expected {
			t.Errorf("GenerateUniqueArchiveName() = %s, want %s", result, expected)
		}
	})

	t.Run("handles tar.xz extension", func(t *testing.T) {
		originalPath := filepath.Join(tmpDir, "archive.tar.xz")
		result := GenerateUniqueArchiveName(originalPath)
		expected := filepath.Join(tmpDir, "archive_1.tar.xz")
		if result != expected {
			t.Errorf("GenerateUniqueArchiveName() = %s, want %s", result, expected)
		}
	})

	t.Run("handles 7z extension", func(t *testing.T) {
		originalPath := filepath.Join(tmpDir, "archive.7z")
		result := GenerateUniqueArchiveName(originalPath)
		expected := filepath.Join(tmpDir, "archive_1.7z")
		if result != expected {
			t.Errorf("GenerateUniqueArchiveName() = %s, want %s", result, expected)
		}
	})
}
