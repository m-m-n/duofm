package ui

import (
	"io/fs"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewPermissionDialog tests dialog creation
func TestNewPermissionDialog(t *testing.T) {
	dialog := NewPermissionDialog("testfile.txt", false, 0644)

	if !dialog.IsActive() {
		t.Error("Dialog should be active after creation")
	}

	if dialog.DisplayType() != DialogDisplayPane {
		t.Errorf("Dialog should be pane-local, got %v", dialog.DisplayType())
	}
}

// TestPermissionDialogPresetSelection tests preset key handling
func TestPermissionDialogPresetSelection(t *testing.T) {
	tests := []struct {
		name         string
		isDir        bool
		presetKey    string
		expectedMode string
	}{
		{
			name:         "file preset 1 (644)",
			isDir:        false,
			presetKey:    "1",
			expectedMode: "644",
		},
		{
			name:         "file preset 2 (755)",
			isDir:        false,
			presetKey:    "2",
			expectedMode: "755",
		},
		{
			name:         "dir preset 1 (755)",
			isDir:        true,
			presetKey:    "1",
			expectedMode: "755",
		},
		{
			name:         "dir preset 2 (700)",
			isDir:        true,
			presetKey:    "2",
			expectedMode: "700",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewPermissionDialog("test", tt.isDir, 0644)

			// Simulate preset key press
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.presetKey)}
			updatedDialog, _ := dialog.Update(msg)
			pd := updatedDialog.(*PermissionDialog)

			if pd.inputValue != tt.expectedMode {
				t.Errorf("Expected input value %s, got %s", tt.expectedMode, pd.inputValue)
			}
		})
	}
}

// TestPermissionDialogDigitInput tests numeric digit input
func TestPermissionDialogDigitInput(t *testing.T) {
	var dialog Dialog = NewPermissionDialog("test.txt", false, 0644)

	// Enter "7"
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("7")})
	pd := dialog.(*PermissionDialog)
	if pd.inputValue != "7" {
		t.Errorf("Expected input '7', got '%s'", pd.inputValue)
	}

	// Enter "5"
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("5")})
	pd = dialog.(*PermissionDialog)
	if pd.inputValue != "75" {
		t.Errorf("Expected input '75', got '%s'", pd.inputValue)
	}

	// Enter "5"
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("5")})
	pd = dialog.(*PermissionDialog)
	if pd.inputValue != "755" {
		t.Errorf("Expected input '755', got '%s'", pd.inputValue)
	}

	// Try to enter 4th digit (should be ignored)
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	pd = dialog.(*PermissionDialog)
	if pd.inputValue != "755" {
		t.Errorf("Expected input '755' (4th digit ignored), got '%s'", pd.inputValue)
	}
}

// TestPermissionDialogInvalidDigit tests rejection of invalid digits
func TestPermissionDialogInvalidDigit(t *testing.T) {
	var dialog Dialog = NewPermissionDialog("test.txt", false, 0644)

	// Try to enter "8" (invalid)
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("8")})
	pd := dialog.(*PermissionDialog)

	if pd.inputValue != "" {
		t.Errorf("Expected empty input after invalid digit, got '%s'", pd.inputValue)
	}

	if pd.errorMsg == "" {
		t.Error("Expected error message for invalid digit")
	}
}

// TestPermissionDialogBackspace tests backspace handling
func TestPermissionDialogBackspace(t *testing.T) {
	var dialog Dialog = NewPermissionDialog("test.txt", false, 0644)

	// Enter "755"
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("755")})

	// Backspace
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	pd := dialog.(*PermissionDialog)

	if pd.inputValue != "75" {
		t.Errorf("Expected input '75' after backspace, got '%s'", pd.inputValue)
	}
}

// TestPermissionDialogEscape tests dialog cancellation
func TestPermissionDialogEscape(t *testing.T) {
	var dialog Dialog = NewPermissionDialog("test.txt", false, 0644)

	// Press Escape
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if dialog.IsActive() {
		t.Error("Dialog should be inactive after Escape")
	}
}

// TestPermissionDialogRecursiveOption tests recursive option for directories
func TestPermissionDialogRecursiveOption(t *testing.T) {
	var dialog Dialog = NewPermissionDialog("testdir", true, 0755)
	pd := dialog.(*PermissionDialog)

	// Initially "this only" should be selected
	if pd.recursiveOption != 0 {
		t.Errorf("Expected recursiveOption 0 (this only), got %d", pd.recursiveOption)
	}

	// Press Tab to toggle
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyTab})
	pd = dialog.(*PermissionDialog)

	if pd.recursiveOption != 1 {
		t.Errorf("Expected recursiveOption 1 (recursive), got %d", pd.recursiveOption)
	}

	// Press Tab again to toggle back
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyTab})
	pd = dialog.(*PermissionDialog)

	if pd.recursiveOption != 0 {
		t.Errorf("Expected recursiveOption 0 (this only), got %d", pd.recursiveOption)
	}
}

// TestPermissionDialogRecursiveNotShownForFiles tests that recursive option is not shown for files
func TestPermissionDialogRecursiveNotShownForFiles(t *testing.T) {
	var dialog Dialog = NewPermissionDialog("test.txt", false, 0644)
	pd := dialog.(*PermissionDialog)

	if pd.showRecursive {
		t.Error("Recursive option should not be shown for files")
	}

	// Tab should not affect file dialogs
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyTab})
	pd = dialog.(*PermissionDialog)

	if pd.recursiveOption != 0 {
		t.Error("RecursiveOption should remain 0 for files")
	}
}

// TestPermissionDialogJKNavigation tests j/k key navigation for recursive options (FR3.3)
func TestPermissionDialogJKNavigation(t *testing.T) {
	var dialog Dialog = NewPermissionDialog("testdir", true, 0755)
	pd := dialog.(*PermissionDialog)

	// Initially "this only" should be selected
	if pd.recursiveOption != 0 {
		t.Errorf("Expected recursiveOption 0 (this only), got %d", pd.recursiveOption)
	}

	// Press j to move down to "recursive"
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	pd = dialog.(*PermissionDialog)

	if pd.recursiveOption != 1 {
		t.Errorf("Expected recursiveOption 1 (recursive) after 'j', got %d", pd.recursiveOption)
	}

	// Press k to move up to "this only"
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	pd = dialog.(*PermissionDialog)

	if pd.recursiveOption != 0 {
		t.Errorf("Expected recursiveOption 0 (this only) after 'k', got %d", pd.recursiveOption)
	}

	// Press k again (should wrap to bottom - recursive)
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	pd = dialog.(*PermissionDialog)

	if pd.recursiveOption != 1 {
		t.Errorf("Expected recursiveOption 1 (recursive) after 'k' wrap, got %d", pd.recursiveOption)
	}

	// Press j again (should wrap to top - this only)
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	pd = dialog.(*PermissionDialog)

	if pd.recursiveOption != 0 {
		t.Errorf("Expected recursiveOption 0 (this only) after 'j' wrap, got %d", pd.recursiveOption)
	}
}

// TestPermissionDialogSpaceSelection tests Space key selection for recursive options (FR3.4, FR10.6)
func TestPermissionDialogSpaceSelection(t *testing.T) {
	var dialog Dialog = NewPermissionDialog("testdir", true, 0755)
	pd := dialog.(*PermissionDialog)

	// Initially "this only" should be selected
	if pd.recursiveOption != 0 {
		t.Errorf("Expected recursiveOption 0 (this only), got %d", pd.recursiveOption)
	}

	// Press Space to toggle
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeySpace})
	pd = dialog.(*PermissionDialog)

	if pd.recursiveOption != 1 {
		t.Errorf("Expected recursiveOption 1 (recursive) after Space, got %d", pd.recursiveOption)
	}

	// Press Space again to toggle back
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeySpace})
	pd = dialog.(*PermissionDialog)

	if pd.recursiveOption != 0 {
		t.Errorf("Expected recursiveOption 0 (this only) after Space, got %d", pd.recursiveOption)
	}
}

// TestPermissionDialogJKForFilesNoEffect tests that j/k keys don't affect file dialogs
func TestPermissionDialogJKForFilesNoEffect(t *testing.T) {
	var dialog Dialog = NewPermissionDialog("test.txt", false, 0644)
	pd := dialog.(*PermissionDialog)

	// j/k should not navigate for files (no recursive option)
	dialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	pd = dialog.(*PermissionDialog)

	// Should remain 0 and not show error
	if pd.recursiveOption != 0 {
		t.Error("j key should not affect file dialogs")
	}
	if pd.errorMsg != "" {
		t.Errorf("j key should not produce error for files, got: %s", pd.errorMsg)
	}
}

// TestFormatPermission tests permission formatting
func TestFormatPermission(t *testing.T) {
	tests := []struct {
		name     string
		mode     fs.FileMode
		isDir    bool
		expected string
	}{
		{
			name:     "file 644",
			mode:     0644,
			isDir:    false,
			expected: "644",
		},
		{
			name:     "dir 755",
			mode:     0755,
			isDir:    true,
			expected: "755",
		},
		{
			name:     "file 777",
			mode:     0777,
			isDir:    false,
			expected: "777",
		},
		{
			name:     "file 000",
			mode:     0000,
			isDir:    false,
			expected: "000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPermission(tt.mode)
			if got != tt.expected {
				t.Errorf("formatPermission(%o) = %s, want %s", tt.mode, got, tt.expected)
			}
		})
	}
}
