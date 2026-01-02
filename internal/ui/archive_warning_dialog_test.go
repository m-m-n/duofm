package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewCompressionBombWarningDialog(t *testing.T) {
	dialog := NewCompressionBombWarningDialog("/test/archive.tar.gz", 1024, 2048000, 2000.0)

	if dialog == nil {
		t.Fatal("NewCompressionBombWarningDialog returned nil")
	}
	if dialog.warningType != ArchiveWarningCompressionBomb {
		t.Errorf("Expected warning type %v, got %v", ArchiveWarningCompressionBomb, dialog.warningType)
	}
	if dialog.archivePath != "/test/archive.tar.gz" {
		t.Errorf("Expected archive path '/test/archive.tar.gz', got '%s'", dialog.archivePath)
	}
	if dialog.archiveSize != 1024 {
		t.Errorf("Expected archive size 1024, got %d", dialog.archiveSize)
	}
	if dialog.extractedSize != 2048000 {
		t.Errorf("Expected extracted size 2048000, got %d", dialog.extractedSize)
	}
	if dialog.ratio != 2000.0 {
		t.Errorf("Expected ratio 2000.0, got %f", dialog.ratio)
	}
	if !dialog.active {
		t.Error("Expected dialog to be active")
	}
	if dialog.selectedIndex != 1 { // Default to Cancel
		t.Errorf("Expected selectedIndex 1 (Cancel), got %d", dialog.selectedIndex)
	}
}

func TestNewDiskSpaceWarningDialog(t *testing.T) {
	dialog := NewDiskSpaceWarningDialog("/test/archive.zip", 1024*1024*1024, 500*1024*1024)

	if dialog == nil {
		t.Fatal("NewDiskSpaceWarningDialog returned nil")
	}
	if dialog.warningType != ArchiveWarningDiskSpace {
		t.Errorf("Expected warning type %v, got %v", ArchiveWarningDiskSpace, dialog.warningType)
	}
	if dialog.archivePath != "/test/archive.zip" {
		t.Errorf("Expected archive path '/test/archive.zip', got '%s'", dialog.archivePath)
	}
	if dialog.extractedSize != 1024*1024*1024 {
		t.Errorf("Expected extracted size 1073741824, got %d", dialog.extractedSize)
	}
	if dialog.availableSize != 500*1024*1024 {
		t.Errorf("Expected available size 524288000, got %d", dialog.availableSize)
	}
	if !dialog.active {
		t.Error("Expected dialog to be active")
	}
	if dialog.selectedIndex != 1 { // Default to Cancel
		t.Errorf("Expected selectedIndex 1 (Cancel), got %d", dialog.selectedIndex)
	}
}

func TestArchiveWarningDialog_Update_Cancel(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"Escape key", "esc"},
		{"n key", "n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewCompressionBombWarningDialog("/test/archive.tar.gz", 1024, 2048000, 2000.0)

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "esc" {
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			}

			_, cmd := dialog.Update(msg)

			if dialog.active {
				t.Error("Expected dialog to be inactive after cancel")
			}

			if cmd == nil {
				t.Fatal("Expected command to be returned")
			}

			result := cmd()
			resultMsg, ok := result.(archiveWarningResultMsg)
			if !ok {
				t.Fatal("Expected archiveWarningResultMsg")
			}

			if resultMsg.choice != ArchiveWarningCancel {
				t.Errorf("Expected choice %v, got %v", ArchiveWarningCancel, resultMsg.choice)
			}
			if resultMsg.warningType != ArchiveWarningCompressionBomb {
				t.Errorf("Expected warning type %v, got %v", ArchiveWarningCompressionBomb, resultMsg.warningType)
			}
		})
	}
}

func TestArchiveWarningDialog_Update_Continue(t *testing.T) {
	dialog := NewCompressionBombWarningDialog("/test/archive.tar.gz", 1024, 2048000, 2000.0)

	// Press 'y' to continue
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_, cmd := dialog.Update(msg)

	if dialog.active {
		t.Error("Expected dialog to be inactive after continue")
	}

	if cmd == nil {
		t.Fatal("Expected command to be returned")
	}

	result := cmd()
	resultMsg, ok := result.(archiveWarningResultMsg)
	if !ok {
		t.Fatal("Expected archiveWarningResultMsg")
	}

	if resultMsg.choice != ArchiveWarningContinue {
		t.Errorf("Expected choice %v, got %v", ArchiveWarningContinue, resultMsg.choice)
	}
}

func TestArchiveWarningDialog_Update_Navigation(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		expectedIndex int
	}{
		{"Left arrow selects Continue", "left", 0},
		{"h key selects Continue", "h", 0},
		{"Right arrow selects Cancel", "right", 1},
		{"l key selects Cancel", "l", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewCompressionBombWarningDialog("/test/archive.tar.gz", 1024, 2048000, 2000.0)

			var msg tea.KeyMsg
			switch tt.key {
			case "left":
				msg = tea.KeyMsg{Type: tea.KeyLeft}
			case "right":
				msg = tea.KeyMsg{Type: tea.KeyRight}
			default:
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			dialog.Update(msg)

			if dialog.selectedIndex != tt.expectedIndex {
				t.Errorf("Expected selectedIndex %d, got %d", tt.expectedIndex, dialog.selectedIndex)
			}
		})
	}
}

func TestArchiveWarningDialog_Update_Tab(t *testing.T) {
	dialog := NewCompressionBombWarningDialog("/test/archive.tar.gz", 1024, 2048000, 2000.0)
	initialIndex := dialog.selectedIndex

	// Tab should toggle between Continue and Cancel
	msg := tea.KeyMsg{Type: tea.KeyTab}
	dialog.Update(msg)

	if dialog.selectedIndex == initialIndex {
		t.Error("Expected selectedIndex to change after Tab")
	}

	// Tab again should go back
	dialog.Update(msg)
	if dialog.selectedIndex != initialIndex {
		t.Errorf("Expected selectedIndex to return to %d, got %d", initialIndex, dialog.selectedIndex)
	}
}

func TestArchiveWarningDialog_Update_EnterWithContinue(t *testing.T) {
	dialog := NewCompressionBombWarningDialog("/test/archive.tar.gz", 1024, 2048000, 2000.0)

	// Navigate to Continue
	dialog.Update(tea.KeyMsg{Type: tea.KeyLeft})

	// Press Enter
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Expected command to be returned")
	}

	result := cmd()
	resultMsg, ok := result.(archiveWarningResultMsg)
	if !ok {
		t.Fatal("Expected archiveWarningResultMsg")
	}

	if resultMsg.choice != ArchiveWarningContinue {
		t.Errorf("Expected choice %v, got %v", ArchiveWarningContinue, resultMsg.choice)
	}
}

func TestArchiveWarningDialog_Update_EnterWithCancel(t *testing.T) {
	dialog := NewCompressionBombWarningDialog("/test/archive.tar.gz", 1024, 2048000, 2000.0)
	// Default is Cancel (selectedIndex = 1)

	// Press Enter
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Expected command to be returned")
	}

	result := cmd()
	resultMsg, ok := result.(archiveWarningResultMsg)
	if !ok {
		t.Fatal("Expected archiveWarningResultMsg")
	}

	if resultMsg.choice != ArchiveWarningCancel {
		t.Errorf("Expected choice %v, got %v", ArchiveWarningCancel, resultMsg.choice)
	}
}

func TestArchiveWarningDialog_View_CompressionBomb(t *testing.T) {
	dialog := NewCompressionBombWarningDialog("/test/archive.tar.gz", 1024*1024, 2*1024*1024*1024, 2000.0)
	view := dialog.View()

	// Check for expected content
	expectedStrings := []string{
		"Warning",
		"extraction ratio",
		"Archive size",
		"Extracted size",
		"ratio",
		"zip bomb",
		"Continue",
		"Cancel",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(view, expected) {
			t.Errorf("Expected view to contain '%s'", expected)
		}
	}
}

func TestArchiveWarningDialog_View_DiskSpace(t *testing.T) {
	dialog := NewDiskSpaceWarningDialog("/test/archive.zip", 2*1024*1024*1024, 500*1024*1024)
	view := dialog.View()

	// Check for expected content
	expectedStrings := []string{
		"Warning",
		"disk space",
		"Required",
		"Available",
		"Continue",
		"Cancel",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(view, expected) {
			t.Errorf("Expected view to contain '%s'", expected)
		}
	}
}

func TestArchiveWarningDialog_IsActive(t *testing.T) {
	dialog := NewCompressionBombWarningDialog("/test/archive.tar.gz", 1024, 2048000, 2000.0)

	if !dialog.IsActive() {
		t.Error("Expected dialog to be active")
	}

	// Cancel the dialog
	dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if dialog.IsActive() {
		t.Error("Expected dialog to be inactive after cancel")
	}
}

func TestArchiveWarningDialog_DisplayType(t *testing.T) {
	dialog := NewCompressionBombWarningDialog("/test/archive.tar.gz", 1024, 2048000, 2000.0)

	if dialog.DisplayType() != DialogDisplayScreen {
		t.Errorf("Expected DisplayType to be DialogDisplayScreen, got %v", dialog.DisplayType())
	}
}

func TestArchiveWarningDialog_InactiveReturnsEmpty(t *testing.T) {
	dialog := NewCompressionBombWarningDialog("/test/archive.tar.gz", 1024, 2048000, 2000.0)
	dialog.active = false

	view := dialog.View()
	if view != "" {
		t.Error("Expected empty view when dialog is inactive")
	}
}

func TestArchiveWarningDialog_InactiveIgnoresInput(t *testing.T) {
	dialog := NewCompressionBombWarningDialog("/test/archive.tar.gz", 1024, 2048000, 2000.0)
	dialog.active = false

	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("Expected nil command when dialog is inactive")
	}
}
