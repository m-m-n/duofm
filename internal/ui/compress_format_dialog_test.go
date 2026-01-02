package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/archive"
)

func TestNewCompressFormatDialog(t *testing.T) {
	dialog := NewCompressFormatDialog()

	if dialog == nil {
		t.Fatal("NewCompressFormatDialog() returned nil")
	}

	if !dialog.active {
		t.Error("dialog should be active by default")
	}

	if dialog.cursor != 0 {
		t.Errorf("dialog.cursor = %d, want 0", dialog.cursor)
	}

	if dialog.width != 50 {
		t.Errorf("dialog.width = %d, want 50", dialog.width)
	}

	if len(dialog.formats) == 0 {
		t.Error("dialog.formats should not be empty")
	}
}

func TestCompressFormatDialogUpdate_Navigation(t *testing.T) {
	tests := []struct {
		name           string
		keyType        tea.KeyType
		keyRunes       []rune
		initialCursor  int
		expectedCursor int
	}{
		{"j moves down", tea.KeyRunes, []rune{'j'}, 0, 1},
		{"down arrow moves down", tea.KeyDown, nil, 0, 1},
		{"k at 0 wraps to end", tea.KeyRunes, []rune{'k'}, 0, -1}, // -1 = len(formats)-1
		{"up arrow at 0 wraps", tea.KeyUp, nil, 0, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewCompressFormatDialog()
			dialog.cursor = tt.initialCursor
			formatCount := len(dialog.formats)

			var msg tea.KeyMsg
			if tt.keyRunes != nil {
				msg = tea.KeyMsg{Type: tt.keyType, Runes: tt.keyRunes}
			} else {
				msg = tea.KeyMsg{Type: tt.keyType}
			}

			updated, cmd := dialog.Update(msg)
			if cmd != nil {
				t.Error("navigation should not return a command")
			}

			d := updated.(*CompressFormatDialog)
			expected := tt.expectedCursor
			if expected == -1 {
				expected = formatCount - 1
			}

			if d.cursor != expected {
				t.Errorf("cursor = %d, want %d", d.cursor, expected)
			}
		})
	}
}

func TestCompressFormatDialogUpdate_CursorWrap(t *testing.T) {
	t.Run("cursor wraps from end to start", func(t *testing.T) {
		dialog := NewCompressFormatDialog()
		dialog.cursor = len(dialog.formats) - 1 // Last item

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
		updated, _ := dialog.Update(msg)

		d := updated.(*CompressFormatDialog)
		if d.cursor != 0 {
			t.Errorf("cursor = %d, want 0 (should wrap to start)", d.cursor)
		}
	})

	t.Run("cursor wraps from start to end", func(t *testing.T) {
		dialog := NewCompressFormatDialog()
		dialog.cursor = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
		updated, _ := dialog.Update(msg)

		d := updated.(*CompressFormatDialog)
		expected := len(d.formats) - 1
		if d.cursor != expected {
			t.Errorf("cursor = %d, want %d (should wrap to end)", d.cursor, expected)
		}
	})
}

func TestCompressFormatDialogUpdate_Cancel(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"esc cancels", tea.KeyMsg{Type: tea.KeyEsc}},
		{"ctrl+c cancels", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewCompressFormatDialog()

			updated, cmd := dialog.Update(tt.key)

			d := updated.(*CompressFormatDialog)
			if d.active {
				t.Error("dialog should be inactive after cancel")
			}

			if cmd == nil {
				t.Fatal("command should be returned after cancel")
			}

			result := cmd()
			resultMsg, ok := result.(compressFormatResultMsg)
			if !ok {
				t.Fatal("command should return compressFormatResultMsg")
			}

			if !resultMsg.cancelled {
				t.Error("resultMsg.cancelled should be true")
			}
		})
	}
}

func TestCompressFormatDialogUpdate_Enter(t *testing.T) {
	dialog := NewCompressFormatDialog()
	dialog.cursor = 0
	expectedFormat := dialog.formats[0]

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := dialog.Update(msg)

	d := updated.(*CompressFormatDialog)
	if d.active {
		t.Error("dialog should be inactive after Enter")
	}

	if cmd == nil {
		t.Fatal("command should be returned after Enter")
	}

	result := cmd()
	resultMsg, ok := result.(compressFormatResultMsg)
	if !ok {
		t.Fatal("command should return compressFormatResultMsg")
	}

	if resultMsg.cancelled {
		t.Error("resultMsg.cancelled should be false")
	}

	if resultMsg.format != expectedFormat {
		t.Errorf("resultMsg.format = %v, want %v", resultMsg.format, expectedFormat)
	}
}

func TestCompressFormatDialogUpdate_NumberKeys(t *testing.T) {
	dialog := NewCompressFormatDialog()

	// Test valid number key (format count dependent)
	if len(dialog.formats) >= 1 {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
		updated, cmd := dialog.Update(msg)

		d := updated.(*CompressFormatDialog)
		if d.active {
			t.Error("dialog should be inactive after number key selection")
		}

		if cmd == nil {
			t.Fatal("command should be returned")
		}

		result := cmd()
		resultMsg, ok := result.(compressFormatResultMsg)
		if !ok {
			t.Fatal("command should return compressFormatResultMsg")
		}

		if resultMsg.cancelled {
			t.Error("resultMsg.cancelled should be false")
		}

		if resultMsg.format != dialog.formats[0] {
			t.Errorf("resultMsg.format = %v, want %v", resultMsg.format, dialog.formats[0])
		}
	}
}

func TestCompressFormatDialogUpdate_InvalidNumberKey(t *testing.T) {
	dialog := NewCompressFormatDialog()

	// Try number key larger than format count
	if len(dialog.formats) < 9 {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'9'}}
		updated, cmd := dialog.Update(msg)

		d := updated.(*CompressFormatDialog)
		if !d.active {
			t.Error("dialog should remain active with invalid number key")
		}

		if cmd != nil {
			t.Error("command should be nil with invalid number key")
		}
	}
}

func TestCompressFormatDialogUpdate_Inactive(t *testing.T) {
	dialog := NewCompressFormatDialog()
	dialog.active = false
	initialCursor := dialog.cursor

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updated, cmd := dialog.Update(msg)

	if updated == nil {
		t.Fatal("updated dialog should not be nil")
	}

	if cmd != nil {
		t.Error("command should be nil when inactive")
	}

	d := updated.(*CompressFormatDialog)
	if d.cursor != initialCursor {
		t.Errorf("cursor = %d, want %d (should not change when inactive)", d.cursor, initialCursor)
	}
}

func TestCompressFormatDialogView(t *testing.T) {
	t.Run("active dialog shows content", func(t *testing.T) {
		dialog := NewCompressFormatDialog()
		view := dialog.View()

		if view == "" {
			t.Error("view should not be empty for active dialog")
		}

		if !contains(view, "Select Archive Format") {
			t.Error("view should contain title")
		}

		if !contains(view, "Navigate") {
			t.Error("view should contain navigation hint")
		}

		if !contains(view, "Cancel") {
			t.Error("view should contain cancel hint")
		}
	})

	t.Run("inactive dialog shows empty", func(t *testing.T) {
		dialog := NewCompressFormatDialog()
		dialog.active = false
		view := dialog.View()

		if view != "" {
			t.Error("view should be empty for inactive dialog")
		}
	})
}

func TestCompressFormatDialogView_ShowsAllFormats(t *testing.T) {
	dialog := NewCompressFormatDialog()
	view := dialog.View()

	// Check that format descriptions are present
	for _, format := range dialog.formats {
		var expected string
		switch format {
		case archive.FormatTar:
			expected = "tar"
		case archive.FormatTarGz:
			expected = "tar.gz"
		case archive.FormatTarBz2:
			expected = "tar.bz2"
		case archive.FormatTarXz:
			expected = "tar.xz"
		case archive.FormatZip:
			expected = "zip"
		case archive.Format7z:
			expected = "7z"
		}

		if expected != "" && !contains(view, expected) {
			t.Errorf("view should contain format %s", expected)
		}
	}
}

func TestCompressFormatDialogIsActive(t *testing.T) {
	t.Run("new dialog is active", func(t *testing.T) {
		dialog := NewCompressFormatDialog()
		if !dialog.IsActive() {
			t.Error("new dialog should be active")
		}
	})

	t.Run("cancelled dialog is inactive", func(t *testing.T) {
		dialog := NewCompressFormatDialog()
		msg := tea.KeyMsg{Type: tea.KeyEsc}
		updated, _ := dialog.Update(msg)

		d := updated.(*CompressFormatDialog)
		if d.IsActive() {
			t.Error("cancelled dialog should be inactive")
		}
	})
}

func TestCompressFormatDialogDisplayType(t *testing.T) {
	dialog := NewCompressFormatDialog()
	if dialog.DisplayType() != DialogDisplayScreen {
		t.Errorf("DisplayType() = %v, want DialogDisplayScreen", dialog.DisplayType())
	}
}

// contains is a helper function to check if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
