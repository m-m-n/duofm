package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Model dialog tests: context menu, dialogs, messages

// Model menu tests: context menu, sort dialog, file/directory dialogs

func TestModelContextMenuOpen(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Verify no dialog initially
	if m.dialog != nil {
		t.Error("dialog should be nil initially")
	}

	// Move cursor to a file (not parent directory ..)
	// Assuming first entry is "..", move to second entry
	m.getActivePane().MoveCursorDown()

	// Press @ key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'@'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Verify context menu is opened
	if m.dialog == nil {
		t.Error("dialog should be opened after @ key")
	}

	_, isContextMenu := m.dialog.(*ContextMenuDialog)
	if !isContextMenu {
		t.Error("dialog should be ContextMenuDialog")
	}
}

// TestModelContextMenuParentDirProtection tests that @ key does nothing for parent directory
func TestModelContextMenuParentDirProtection(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Cursor is at position 0 which should be ".." (parent directory)
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || !entry.IsParentDir() {
		// If first entry is not "..", skip this test
		t.Skip("First entry is not parent directory, skipping test")
	}

	// Press @ key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'@'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Verify no dialog is opened for parent directory
	if m.dialog != nil {
		t.Error("dialog should not be opened for parent directory")
	}
}

// TestModelContextMenuDeleteShowsConfirmDialog tests that delete action shows confirmation dialog
func TestModelContextMenuDeleteShowsConfirmDialog(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Move to a file (not parent directory)
	m.getActivePane().MoveCursorDown()

	// Press @ key to open context menu
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'@'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.dialog == nil {
		t.Fatal("context menu should be opened")
	}

	// Simulate selecting delete (press '5' for delete - now at index 4)
	// Menu items: 1=open, 2=open_with, 3=copy, 4=move, 5=delete
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	updatedModel, cmd := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Execute the command to send contextMenuResultMsg
	if cmd != nil {
		resultMsg := cmd()
		updatedModel, _ = m.Update(resultMsg)
		m = updatedModel.(Model)
	}

	// Verify ConfirmDialog is shown (not direct deletion)
	if m.dialog == nil {
		t.Error("dialog should be shown after delete action")
	}

	_, isConfirmDialog := m.dialog.(*ConfirmDialog)
	if !isConfirmDialog {
		t.Error("dialog should be ConfirmDialog after delete action from context menu")
	}

	// Verify pendingAction is set
	if m.pendingAction == nil {
		t.Error("pendingAction should be set for delete confirmation")
	}
}

// TestModelContextMenuCancelledClearsPendingAction tests that cancelling clears pendingAction
func TestModelContextMenuCancelledClearsPendingAction(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set a pending action manually
	m.pendingAction = func() error { return nil }

	// Create a ConfirmDialog
	m.dialog = NewConfirmDialog("Test", "test")

	// Simulate pressing 'n' to cancel
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, cmd := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Execute the command to send dialogResultMsg
	if cmd != nil {
		resultMsg := cmd()
		updatedModel, _ = m.Update(resultMsg)
		m = updatedModel.(Model)
	}

	// Verify pendingAction is cleared
	if m.pendingAction != nil {
		t.Error("pendingAction should be cleared after cancellation")
	}
}

// TestArrowKeyNavigation tests arrow key navigation in main view
func TestModelContextMenuEscClosesMenu(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Move to a file
	m.getActivePane().MoveCursorDown()

	// Press @ key to open context menu
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'@'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.dialog == nil {
		t.Fatal("context menu should be opened")
	}

	// Press Esc to close
	keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, cmd := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Execute the command
	if cmd != nil {
		resultMsg := cmd()
		updatedModel, _ = m.Update(resultMsg)
		m = updatedModel.(Model)
	}

	// Verify dialog is closed
	if m.dialog != nil {
		t.Error("dialog should be closed after Esc")
	}
}

// === Phase 2: ステータスバーメッセージ機能のテスト ===

func TestHelpDialogToggle(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Press '?' to open help
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.dialog == nil {
		t.Error("dialog should be set after ? key")
	}

	_, isHelpDialog := m.dialog.(*HelpDialog)
	if !isHelpDialog {
		t.Error("dialog should be HelpDialog")
	}
}

func TestDeleteKeyShowsConfirmDialog(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Move to a file (not parent dir)
	m.getActivePane().MoveCursorDown()
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		t.Skip("No suitable entry for test")
	}

	// Press 'd' for delete
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.dialog == nil {
		t.Error("dialog should be set after d key")
	}

	_, isConfirmDialog := m.dialog.(*ConfirmDialog)
	if !isConfirmDialog {
		t.Error("dialog should be ConfirmDialog")
	}
}

func TestNewFileDialogOpens(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Press 'n' for new file
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.dialog == nil {
		t.Error("dialog should be set after n key")
	}

	_, isInputDialog := m.dialog.(*InputDialog)
	if !isInputDialog {
		t.Error("dialog should be InputDialog")
	}
}

func TestNewDirectoryDialogOpens(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Press 'N' (shift+n) for new directory
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.dialog == nil {
		t.Error("dialog should be set after N key")
	}

	_, isInputDialog := m.dialog.(*InputDialog)
	if !isInputDialog {
		t.Error("dialog should be InputDialog")
	}
}

func TestRenameDialogOpens(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Move to a file (not parent dir)
	m.getActivePane().MoveCursorDown()
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		t.Skip("No suitable entry for test")
	}

	// Press 'r' for rename
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.dialog == nil {
		t.Error("dialog should be set after r key")
	}

	// Dialog type depends on whether the file has an extension
	// Files with extension -> ExtensionRenameDialog
	// Files without extension or directories -> InputDialog
	_, baseName, hasExt := hasEditableExtension(entry.Name, entry.IsDir)
	_ = baseName // suppress unused variable warning

	if hasExt {
		_, isExtRenameDialog := m.dialog.(*ExtensionRenameDialog)
		if !isExtRenameDialog {
			t.Errorf("dialog should be ExtensionRenameDialog for file with extension, got %T", m.dialog)
		}
	} else {
		_, isInputDialog := m.dialog.(*InputDialog)
		if !isInputDialog {
			t.Errorf("dialog should be InputDialog for file without extension, got %T", m.dialog)
		}
	}
}

func TestModelSortDialogResultConfirmed(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set sort dialog active
	m.sortDialog = NewSortDialog(SortConfig{Field: SortBySize, Order: SortDesc})

	// Send confirmed result
	resultMsg := sortDialogResultMsg{
		config:    SortConfig{Field: SortBySize, Order: SortDesc},
		confirmed: true,
		cancelled: false,
	}
	updatedModel, _ = m.Update(resultMsg)
	m = updatedModel.(Model)

	// Sort dialog should be closed
	if m.sortDialog != nil {
		t.Error("Sort dialog should be nil after confirmed")
	}
}

func TestModelSortDialogResultCancelled(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set sort dialog active
	originalConfig := SortConfig{Field: SortByName, Order: SortAsc}
	m.sortDialog = NewSortDialog(originalConfig)

	// Send cancelled result with original config
	resultMsg := sortDialogResultMsg{
		config:    originalConfig,
		confirmed: false,
		cancelled: true,
	}
	updatedModel, _ = m.Update(resultMsg)
	m = updatedModel.(Model)

	// Sort dialog should be closed
	if m.sortDialog != nil {
		t.Error("Sort dialog should be nil after cancelled")
	}

	// Active pane should have original config
	if m.getActivePane().GetSortConfig().Field != SortByName {
		t.Error("Sort config should be restored to original")
	}
}

func TestModelSortDialogConfigChanged(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set sort dialog active
	m.sortDialog = NewSortDialog(SortConfig{Field: SortByName, Order: SortAsc})

	// Send config changed message
	configMsg := sortDialogConfigChangedMsg{
		config: SortConfig{Field: SortByDate, Order: SortDesc},
	}
	updatedModel, _ = m.Update(configMsg)
	m = updatedModel.(Model)

	// Active pane should have new config (live preview)
	if m.getActivePane().GetSortConfig().Field != SortByDate {
		t.Error("Sort config should be updated for live preview")
	}
}

func TestModelSortDialogConfigChangedWithoutDialog(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// sortDialog is nil

	// Send config changed message (should be ignored)
	configMsg := sortDialogConfigChangedMsg{
		config: SortConfig{Field: SortByDate, Order: SortDesc},
	}
	updatedModel, _ = m.Update(configMsg)
	m = updatedModel.(Model)

	// Should not crash, config should remain default
	if m.getActivePane().GetSortConfig().Field != SortByName {
		t.Error("Sort config should remain unchanged when dialog is nil")
	}
}

func TestModelViewWithSortDialog(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Activate sort dialog
	m.sortDialog = NewSortDialog(SortConfig{Field: SortByName, Order: SortAsc})
	m.sortDialog.width = 30

	// Render view
	view := m.View()

	// View should contain sort dialog content
	if view == "" {
		t.Error("View should not be empty when sort dialog is active")
	}
	if !strings.Contains(view, "Sort") {
		t.Error("View should contain 'Sort' when sort dialog is active")
	}
}

func TestModelSortKeyOpensDialog(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Press 's' to open sort dialog
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Sort dialog should be active
	if m.sortDialog == nil {
		t.Error("Sort dialog should be active after pressing 's'")
	}
}
