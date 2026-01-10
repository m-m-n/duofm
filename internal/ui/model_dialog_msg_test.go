package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Model dialog tests: context menu, dialogs, messages

// Model dialog message tests: dialog results, view rendering

func TestInputDialogResultMsgClearsDialog(t *testing.T) {
	tests := []struct {
		name      string
		msg       inputDialogResultMsg
		wantError bool
	}{
		{
			name: "ファイル作成成功後にdialogがクリアされる",
			msg: inputDialogResultMsg{
				operation: "create_file",
				input:     "test.txt",
			},
			wantError: false,
		},
		{
			name: "ディレクトリ作成成功後にdialogがクリアされる",
			msg: inputDialogResultMsg{
				operation: "create_dir",
				input:     "testdir",
			},
			wantError: false,
		},
		{
			name: "リネーム成功後にdialogがクリアされる",
			msg: inputDialogResultMsg{
				operation: "rename",
				input:     "newname.txt",
				oldName:   "oldname.txt",
			},
			wantError: false,
		},
		{
			name: "エラー時もdialogがクリアされる",
			msg: inputDialogResultMsg{
				operation: "create_file",
				err:       fmt.Errorf("file already exists"),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()

			// Initialize with WindowSizeMsg
			msg := tea.WindowSizeMsg{
				Width:  120,
				Height: 40,
			}
			updatedModel, _ := model.Update(msg)
			m := updatedModel.(Model)

			// Simulate an open InputDialog
			m.dialog = NewInputDialog("Test:", func(s string) tea.Cmd { return nil })

			// Verify dialog is not nil
			if m.dialog == nil {
				t.Fatal("dialog should not be nil before test")
			}

			// Send inputDialogResultMsg
			updatedModel, _ = m.Update(tt.msg)
			m = updatedModel.(Model)

			// CRITICAL: dialog must be nil after inputDialogResultMsg
			if m.dialog != nil {
				t.Error("dialog should be nil after inputDialogResultMsg - this causes the app to become unresponsive")
			}

			// Verify error handling
			if tt.wantError {
				if m.statusMessage == "" {
					t.Error("statusMessage should be set on error")
				}
				if !m.isStatusError {
					t.Error("isStatusError should be true on error")
				}
			}
		})
	}
}

func TestDialogResultMsgClearsDialog(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Simulate an open ConfirmDialog
	m.dialog = NewConfirmDialog("Delete?", "test.txt")

	// Verify dialog is not nil
	if m.dialog == nil {
		t.Fatal("dialog should not be nil before test")
	}

	// Send dialogResultMsg (cancelled)
	resultMsg := dialogResultMsg{
		result: DialogResult{Confirmed: false},
	}
	updatedModel, _ = m.Update(resultMsg)
	m = updatedModel.(Model)

	// dialog must be nil after dialogResultMsg
	if m.dialog != nil {
		t.Error("dialog should be nil after dialogResultMsg")
	}
}

func TestContextMenuResultMsgClearsDialog(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Move to a file and open context menu
	m.getActivePane().MoveCursorDown()
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		t.Skip("No suitable entry for context menu test")
	}

	m.dialog = NewContextMenuDialogWithPane(
		entry,
		m.getActivePane().Path(),
		m.getInactivePane().Path(),
		m.getActivePane(),
	)

	// Verify dialog is not nil
	if m.dialog == nil {
		t.Fatal("dialog should not be nil before test")
	}

	// Send contextMenuResultMsg (cancelled)
	resultMsg := contextMenuResultMsg{
		cancelled: true,
	}
	updatedModel, _ = m.Update(resultMsg)
	m = updatedModel.(Model)

	// dialog must be nil after contextMenuResultMsg
	if m.dialog != nil {
		t.Error("dialog should be nil after contextMenuResultMsg")
	}
}

func TestNavigationWorksAfterDialogClose(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Simulate file creation dialog completion
	m.dialog = NewInputDialog("New file:", func(s string) tea.Cmd { return nil })
	resultMsg := inputDialogResultMsg{
		operation: "create_file",
		input:     "test.txt",
	}
	updatedModel, _ = m.Update(resultMsg)
	m = updatedModel.(Model)

	// Get initial cursor position
	initialCursor := m.getActivePane().cursor

	// Try to navigate with j key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Cursor should have moved (navigation works)
	if m.getActivePane().cursor == initialCursor && len(m.getActivePane().entries) > 1 {
		t.Error("navigation should work after dialog close - cursor didn't move")
	}

	// Try q key to quit
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(keyMsg)

	// q should return quit command
	if cmd == nil {
		t.Error("q key should work after dialog close - no quit command returned")
	}
}

// === Overwrite Confirmation Dialog Tests ===

func TestShowOverwriteDialogMsg(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Send showOverwriteDialogMsg
	overwriteMsg := showOverwriteDialogMsg{
		filename:  "test.txt",
		srcPath:   "/src/test.txt",
		destPath:  "/dest",
		srcInfo:   OverwriteFileInfo{Size: 1234},
		destInfo:  OverwriteFileInfo{Size: 5678},
		operation: "copy",
	}
	updatedModel, _ = m.Update(overwriteMsg)
	m = updatedModel.(Model)

	// Verify OverwriteDialog is shown
	if m.dialog == nil {
		t.Fatal("dialog should not be nil after showOverwriteDialogMsg")
	}

	_, isOverwriteDialog := m.dialog.(*OverwriteDialog)
	if !isOverwriteDialog {
		t.Error("dialog should be OverwriteDialog")
	}
}

func TestShowErrorDialogMsg(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Send showErrorDialogMsg
	errorMsg := showErrorDialogMsg{
		message: "Test error message",
	}
	updatedModel, _ = m.Update(errorMsg)
	m = updatedModel.(Model)

	// Verify ErrorDialog is shown
	if m.dialog == nil {
		t.Fatal("dialog should not be nil after showErrorDialogMsg")
	}

	_, isErrorDialog := m.dialog.(*ErrorDialog)
	if !isErrorDialog {
		t.Error("dialog should be ErrorDialog")
	}
}

func TestOverwriteDialogResultMsgOverwrite(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Create a temporary test scenario
	// Set up an OverwriteDialog
	m.dialog = NewOverwriteDialog("test.txt", "/dest", OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")

	// Send overwriteDialogResultMsg with Cancel choice (safer for testing)
	resultMsg := overwriteDialogResultMsg{
		choice:    OverwriteChoiceCancel,
		srcPath:   "/src/test.txt",
		destPath:  "/dest",
		filename:  "test.txt",
		operation: "copy",
	}
	updatedModel, _ = m.Update(resultMsg)
	m = updatedModel.(Model)

	// Dialog should be nil
	if m.dialog != nil {
		t.Error("dialog should be nil after overwriteDialogResultMsg with Cancel")
	}
}

func TestOverwriteDialogResultMsgRename(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set up an OverwriteDialog
	m.dialog = NewOverwriteDialog("test.txt", m.getInactivePane().Path(), OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")

	// Send overwriteDialogResultMsg with Rename choice
	resultMsg := overwriteDialogResultMsg{
		choice:    OverwriteChoiceRename,
		srcPath:   "/src/test.txt",
		destPath:  m.getInactivePane().Path(),
		filename:  "test.txt",
		operation: "copy",
	}
	updatedModel, _ = m.Update(resultMsg)
	m = updatedModel.(Model)

	// Should show RenameInputDialog
	if m.dialog == nil {
		t.Fatal("dialog should not be nil after Rename choice")
	}

	_, isRenameDialog := m.dialog.(*RenameInputDialog)
	if !isRenameDialog {
		t.Error("dialog should be RenameInputDialog after Rename choice")
	}
}

func TestRenameInputResultMsg(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set up a RenameInputDialog
	m.dialog = NewRenameInputDialog(m.getInactivePane().Path(), "/src/test.txt", "copy")

	// Send renameInputResultMsg with a new name
	// Note: This will fail for actual copy since /src/test.txt doesn't exist,
	// but we're testing the message handling flow
	resultMsg := renameInputResultMsg{
		newName:   "newname.txt",
		srcPath:   "/nonexistent/test.txt", // Use nonexistent to trigger error
		destPath:  m.getInactivePane().Path(),
		operation: "copy",
	}
	updatedModel, _ = m.Update(resultMsg)
	m = updatedModel.(Model)

	// Dialog should be replaced with ErrorDialog due to copy failure
	if m.dialog == nil {
		// OK - the original dialog was cleared
	} else {
		_, isErrorDialog := m.dialog.(*ErrorDialog)
		if !isErrorDialog {
			t.Error("dialog should be either nil or ErrorDialog after failed rename operation")
		}
	}
}

func TestOverwriteDialogNavigationInModel(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Create an OverwriteDialog
	m.dialog = NewOverwriteDialog("test.txt", "/dest", OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")

	// Verify dialog exists
	if m.dialog == nil {
		t.Fatal("dialog should not be nil")
	}

	// Press 'j' to navigate in dialog
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Dialog should still be active
	if m.dialog == nil || !m.dialog.IsActive() {
		t.Error("dialog should still be active after navigation")
	}

	// Press Esc to close dialog
	keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, cmd := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should return a command (overwriteDialogResultMsg)
	if cmd != nil {
		resultMsg := cmd()
		updatedModel, _ = m.Update(resultMsg)
		m = updatedModel.(Model)

		// Dialog should be closed
		if m.dialog != nil {
			t.Error("dialog should be nil after Esc and processing result")
		}
	}
}

func TestRenameInputDialogNavigationInModel(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Create a RenameInputDialog
	m.dialog = NewRenameInputDialog(m.getInactivePane().Path(), "/src/test.txt", "copy")

	// Verify dialog exists
	if m.dialog == nil {
		t.Fatal("dialog should not be nil")
	}

	// Type a character
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Dialog should still be active
	if m.dialog == nil || !m.dialog.IsActive() {
		t.Error("dialog should still be active after typing")
	}

	// Press Esc to close dialog
	keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Dialog should be inactive
	if m.dialog != nil && m.dialog.IsActive() {
		t.Error("dialog should be inactive after Esc")
	}
}

// === checkFileConflict and executeFileOperation Tests ===

func TestModelViewWithDialog(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set a dialog
	m.dialog = NewOverwriteDialog("test.txt", "/dest", OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")

	// View should render without error
	view := m.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestModelViewWithErrorDialog(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set an error dialog
	m.dialog = NewErrorDialog("Test error message")

	// View should render without error
	view := m.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestModelViewWithRenameInputDialog(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set a rename input dialog
	m.dialog = NewRenameInputDialog(m.getInactivePane().Path(), "/src/test.txt", "copy")

	// View should render without error
	view := m.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestModelViewWithStatusMessage(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set status message
	m.statusMessage = "Test status message"
	m.isStatusError = false

	// View should render without error
	view := m.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestModelViewWithErrorStatusMessage(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set error status message
	m.statusMessage = "Test error message"
	m.isStatusError = true

	// View should render without error
	view := m.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestModelUpdateWithDialog(t *testing.T) {
	tmpDir := t.TempDir()

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("dialog receives key events", func(t *testing.T) {
		m.dialog = NewHelpDialog()
		keyMsg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, _ := m.Update(keyMsg)
		updated := updatedModel.(Model)
		// Dialog should be closed after Esc
		if updated.dialog != nil && updated.dialog.IsActive() {
			t.Error("Esc should close dialog")
		}
	})
}

func TestModelUpdateMessageTypes(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("handles directoryLoadCompleteMsg", func(t *testing.T) {
		loadMsg := directoryLoadCompleteMsg{
			paneID:   LeftPane,
			panePath: tmpDir,
			entries:  nil,
		}
		m.Update(loadMsg)
	})

	t.Run("handles diskSpaceUpdateMsg", func(t *testing.T) {
		diskMsg := diskSpaceUpdateMsg{}
		m.Update(diskMsg)
	})

	t.Run("handles clearStatusMsg", func(t *testing.T) {
		m.statusMessage = "test message"
		clearMsg := clearStatusMsg{}
		updatedModel, _ := m.Update(clearMsg)
		updated := updatedModel.(Model)
		if updated.statusMessage != "" {
			t.Error("clearStatusMsg should clear status")
		}
	})

	t.Run("handles dialogResultMsg confirmed", func(t *testing.T) {
		resultMsg := dialogResultMsg{
			result: DialogResult{Confirmed: true},
		}
		m.Update(resultMsg)
	})

	t.Run("handles dialogResultMsg cancelled", func(t *testing.T) {
		resultMsg := dialogResultMsg{
			result: DialogResult{Cancelled: true},
		}
		m.Update(resultMsg)
	})

	t.Run("handles inputDialogResultMsg success", func(t *testing.T) {
		resultMsg := inputDialogResultMsg{
			operation: "create_file",
			input:     "newfile.txt",
		}
		m.Update(resultMsg)
	})

	t.Run("handles inputDialogResultMsg error", func(t *testing.T) {
		resultMsg := inputDialogResultMsg{
			operation: "create_file",
			err:       fmt.Errorf("test error"),
		}
		m.Update(resultMsg)
	})
}

// === Path Jump Dialog Tests ===

func TestActionPathJump_OpensDialog(t *testing.T) {
	tmpDir := t.TempDir()

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Press Ctrl+J
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlJ}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Dialog should be PathJumpDialog
	if m.dialog == nil {
		t.Fatal("dialog should not be nil after Ctrl+J")
	}

	_, isPathJumpDialog := m.dialog.(*PathJumpDialog)
	if !isPathJumpDialog {
		t.Error("dialog should be PathJumpDialog")
	}
}

func TestPathJumpResultMsg_ChangesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set PathJumpDialog
	m.dialog = NewPathJumpDialog()

	// Send pathJumpResultMsg
	resultMsg := pathJumpResultMsg{path: targetDir}
	updatedModel, cmd := m.Update(resultMsg)
	m = updatedModel.(Model)

	// Dialog should be cleared
	if m.dialog != nil {
		t.Error("dialog should be nil after pathJumpResultMsg")
	}

	// Should return a command to change directory
	if cmd == nil {
		t.Error("pathJumpResultMsg should return a command to change directory")
	}
}

func TestPathJumpCancelMsg_ClearsDialog(t *testing.T) {
	tmpDir := t.TempDir()

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set PathJumpDialog
	m.dialog = NewPathJumpDialog()

	// Send pathJumpCancelMsg
	cancelMsg := pathJumpCancelMsg{}
	updatedModel, _ = m.Update(cancelMsg)
	m = updatedModel.(Model)

	// Dialog should be cleared
	if m.dialog != nil {
		t.Error("dialog should be nil after pathJumpCancelMsg")
	}
}

func TestPathJumpDialog_NavigationWorksAfterClose(t *testing.T) {
	tmpDir := t.TempDir()

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Open and close dialog
	m.dialog = NewPathJumpDialog()
	cancelMsg := pathJumpCancelMsg{}
	updatedModel, _ = m.Update(cancelMsg)
	m = updatedModel.(Model)

	// Get initial cursor position
	initialCursor := m.getActivePane().cursor

	// Try to navigate with j key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Cursor should have moved (navigation works)
	if m.getActivePane().cursor == initialCursor && len(m.getActivePane().entries) > 1 {
		t.Error("navigation should work after dialog close - cursor didn't move")
	}
}

func TestModelUpdateMoreMessages(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("handles ctrlCTimeoutMsg", func(t *testing.T) {
		timeoutMsg := ctrlCTimeoutMsg{}
		m.Update(timeoutMsg)
	})

	t.Run("handles showErrorDialogMsg", func(t *testing.T) {
		errorMsg := showErrorDialogMsg{
			message: "test error",
		}
		updatedModel, _ := m.Update(errorMsg)
		updated := updatedModel.(Model)
		if updated.dialog == nil {
			t.Error("showErrorDialogMsg should set dialog")
		}
	})

	t.Run("handles showStatusMsg", func(t *testing.T) {
		statusMsg := showStatusMsg{
			message: "test status",
			isError: false,
		}
		updatedModel, _ := m.Update(statusMsg)
		updated := updatedModel.(Model)
		if updated.statusMessage != "test status" {
			t.Errorf("Expected status 'test status', got %q", updated.statusMessage)
		}
	})

	t.Run("handles fileOperationCompleteMsg", func(t *testing.T) {
		opMsg := fileOperationCompleteMsg{
			operation: "copy",
		}
		m.Update(opMsg)
	})

	t.Run("handles batchCompleteMsg", func(t *testing.T) {
		batchMsg := batchCompleteMsg{
			operation: "copy",
			completed: 5,
			failed:    0,
		}
		m.Update(batchMsg)
	})
}
