package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewCompressionLevelDialog(t *testing.T) {
	dialog := NewCompressionLevelDialog()

	if dialog == nil {
		t.Fatal("NewCompressionLevelDialog() returned nil")
	}

	if !dialog.IsActive() {
		t.Error("NewCompressionLevelDialog() should be active by default")
	}

	if dialog.selectedLevel != 6 {
		t.Errorf("NewCompressionLevelDialog() default level = %d, want 6", dialog.selectedLevel)
	}
}

func TestCompressionLevelDialog_Navigation(t *testing.T) {
	dialog := NewCompressionLevelDialog()

	// Test j key (move down)
	updated, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	dialog = updated.(*CompressionLevelDialog)
	if dialog.selectedLevel != 7 {
		t.Errorf("After 'j' key, selectedLevel = %d, want 7", dialog.selectedLevel)
	}

	// Test k key (move up)
	updated, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	dialog = updated.(*CompressionLevelDialog)
	if dialog.selectedLevel != 6 {
		t.Errorf("After 'k' key, selectedLevel = %d, want 6", dialog.selectedLevel)
	}

	// Test boundary - cannot go below 0
	dialog.selectedLevel = 0
	updated, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	dialog = updated.(*CompressionLevelDialog)
	if dialog.selectedLevel != 0 {
		t.Errorf("After 'k' at level 0, selectedLevel = %d, want 0", dialog.selectedLevel)
	}

	// Test boundary - cannot go above 9
	dialog.selectedLevel = 9
	updated, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	dialog = updated.(*CompressionLevelDialog)
	if dialog.selectedLevel != 9 {
		t.Errorf("After 'j' at level 9, selectedLevel = %d, want 9", dialog.selectedLevel)
	}
}

func TestCompressionLevelDialog_DirectSelection(t *testing.T) {
	dialog := NewCompressionLevelDialog()

	tests := []struct {
		key  rune
		want int
	}{
		{'0', 0},
		{'1', 1},
		{'5', 5},
		{'9', 9},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			updated, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})
			d := updated.(*CompressionLevelDialog)
			if d.selectedLevel != tt.want {
				t.Errorf("After '%c' key, selectedLevel = %d, want %d", tt.key, d.selectedLevel, tt.want)
			}
		})
	}
}

func TestCompressionLevelDialog_Confirm(t *testing.T) {
	dialog := NewCompressionLevelDialog()
	dialog.selectedLevel = 8

	// Press Enter
	updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})
	d := updated.(*CompressionLevelDialog)

	if d.IsActive() {
		t.Error("Dialog should be inactive after confirmation")
	}

	// Check that a command is returned
	if cmd == nil {
		t.Error("Command should be returned after confirmation")
	}

	// Execute the command and check the message
	if cmd != nil {
		msg := cmd()
		result, ok := msg.(compressionLevelResultMsg)
		if !ok {
			t.Error("Command should return compressionLevelResultMsg")
		}
		if result.level != 8 {
			t.Errorf("compressionLevelResultMsg.level = %d, want 8", result.level)
		}
		if result.cancelled {
			t.Error("compressionLevelResultMsg.cancelled should be false")
		}
	}
}

func TestCompressionLevelDialog_Cancel(t *testing.T) {
	dialog := NewCompressionLevelDialog()

	// Press Escape
	updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})
	d := updated.(*CompressionLevelDialog)

	if d.IsActive() {
		t.Error("Dialog should be inactive after cancellation")
	}

	// Check that a command is returned
	if cmd == nil {
		t.Error("Command should be returned after cancellation")
	}

	// Execute the command and check the message
	if cmd != nil {
		msg := cmd()
		result, ok := msg.(compressionLevelResultMsg)
		if !ok {
			t.Error("Command should return compressionLevelResultMsg")
		}
		if !result.cancelled {
			t.Error("compressionLevelResultMsg.cancelled should be true")
		}
	}
}

func TestCompressionLevelDialog_SetActive(t *testing.T) {
	dialog := NewCompressionLevelDialog()

	dialog.SetActive(false)
	if dialog.IsActive() {
		t.Error("SetActive(false) should deactivate dialog")
	}

	dialog.SetActive(true)
	if !dialog.IsActive() {
		t.Error("SetActive(true) should activate dialog")
	}
}

func TestCompressionLevelDialog_DisplayType(t *testing.T) {
	dialog := NewCompressionLevelDialog()
	if dialog.DisplayType() != DialogDisplayScreen {
		t.Errorf("DisplayType() = %v, want DialogDisplayScreen", dialog.DisplayType())
	}
}

func TestCompressionLevelDialog_View(t *testing.T) {
	t.Run("active dialog shows content", func(t *testing.T) {
		dialog := NewCompressionLevelDialog()
		view := dialog.View()

		if view == "" {
			t.Error("View should not be empty for active dialog")
		}

		if !containsLevelStr(view, "Select Compression Level") {
			t.Error("View should contain title")
		}

		// Check for level options
		if !containsLevelStr(view, "Level 0") {
			t.Error("View should contain level 0")
		}

		if !containsLevelStr(view, "Level 9") {
			t.Error("View should contain level 9")
		}

		if !containsLevelStr(view, "[Enter] Confirm") {
			t.Error("View should contain confirm hint")
		}

		if !containsLevelStr(view, "[Esc]") {
			t.Error("View should contain Esc hint")
		}
	})

	t.Run("inactive dialog shows empty", func(t *testing.T) {
		dialog := NewCompressionLevelDialog()
		dialog.SetActive(false)
		view := dialog.View()

		if view != "" {
			t.Error("View should be empty for inactive dialog")
		}
	})
}

func TestCompressionLevelDialog_Update_Inactive(t *testing.T) {
	dialog := NewCompressionLevelDialog()
	dialog.SetActive(false)

	updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	if updated == nil {
		t.Error("Updated dialog should not be nil")
	}

	if cmd != nil {
		t.Error("Command should be nil for inactive dialog")
	}
}

// containsLevelStr checks if s contains substr
func containsLevelStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
