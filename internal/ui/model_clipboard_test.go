package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestHandleContextMenuResult_CopyName tests that copy_name sets the status message.
func TestHandleContextMenuResult_CopyName(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Move to a file (not parent directory)
	m.getActivePane().MoveCursorDown()
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		t.Skip("No suitable entry for test")
	}

	// Set context menu dialog (needed for handleContextMenuResult guard)
	m.dialog = NewContextMenuDialog(entry, m.getActivePane().Path(), m.getInactivePane().Path())

	// Simulate copy_name context menu result
	resultMsg := contextMenuResultMsg{
		actionID: "copy_name",
	}

	updatedModel, cmd := m.Update(resultMsg)
	m = updatedModel.(Model)

	// Verify status message is set
	expectedPrefix := "Copied: " + entry.Name
	if m.statusMessage != expectedPrefix {
		t.Errorf("statusMessage = %q, want %q", m.statusMessage, expectedPrefix)
	}

	// Verify isStatusError is false (success message)
	if m.isStatusError {
		t.Error("isStatusError should be false for copy_name success")
	}

	// Verify a command was returned (clipboard write + status clear)
	if cmd == nil {
		t.Error("cmd should not be nil (should include clipboard write and status clear)")
	}

	// Verify dialog is closed
	if m.dialog != nil {
		t.Error("dialog should be nil after context menu result")
	}
}

// TestHandleContextMenuResult_CopyPath tests that copy_path sets the status message with full path.
func TestHandleContextMenuResult_CopyPath(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Move to a file (not parent directory)
	m.getActivePane().MoveCursorDown()
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		t.Skip("No suitable entry for test")
	}

	panePath := m.getActivePane().Path()

	// Set context menu dialog
	m.dialog = NewContextMenuDialog(entry, panePath, m.getInactivePane().Path())

	// Simulate copy_path context menu result
	resultMsg := contextMenuResultMsg{
		actionID: "copy_path",
	}

	updatedModel, cmd := m.Update(resultMsg)
	m = updatedModel.(Model)

	// Verify status message contains the full path
	expectedPrefix := "Copied: "
	if len(m.statusMessage) <= len(expectedPrefix) {
		t.Errorf("statusMessage = %q, should start with %q and contain a path", m.statusMessage, expectedPrefix)
	}

	if m.statusMessage[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("statusMessage should start with %q, got %q", expectedPrefix, m.statusMessage)
	}

	// Verify isStatusError is false
	if m.isStatusError {
		t.Error("isStatusError should be false for copy_path success")
	}

	// Verify a command was returned
	if cmd == nil {
		t.Error("cmd should not be nil")
	}

	// Verify dialog is closed
	if m.dialog != nil {
		t.Error("dialog should be nil after context menu result")
	}
}

// TestHandleClipboardResultMsg_Success tests that successful clipboard result does nothing.
func TestHandleClipboardResultMsg_Success(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set status message as if copy_name just succeeded
	m.statusMessage = "Copied: test.txt"
	m.isStatusError = false

	// Send successful clipboard result
	resultMsg := clipboardResultMsg{err: nil}
	updatedModel, _ = m.Update(resultMsg)
	m = updatedModel.(Model)

	// Status message should remain unchanged (optimistic UI)
	if m.statusMessage != "Copied: test.txt" {
		t.Errorf("statusMessage = %q, should remain unchanged on success", m.statusMessage)
	}
	if m.isStatusError {
		t.Error("isStatusError should remain false on success")
	}
}

// TestHandleClipboardResultMsg_Error tests that clipboard error overwrites status message.
func TestHandleClipboardResultMsg_Error(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set status message as if copy_name just succeeded (optimistic UI)
	m.statusMessage = "Copied: test.txt"
	m.isStatusError = false

	// Send error clipboard result
	resultMsg := clipboardResultMsg{err: errForTest("clipboard command failed")}
	updatedModel, cmd := m.Update(resultMsg)
	m = updatedModel.(Model)

	// Status message should be overwritten with error
	expected := "Copy failed: clipboard command failed"
	if m.statusMessage != expected {
		t.Errorf("statusMessage = %q, want %q", m.statusMessage, expected)
	}
	if !m.isStatusError {
		t.Error("isStatusError should be true on error")
	}

	// A clear command should be returned
	if cmd == nil {
		t.Error("cmd should not be nil (should include status clear)")
	}
}

// errForTest is a simple error type for testing.
type errForTest string

func (e errForTest) Error() string { return string(e) }
