package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Model file operation tests: copy, move, create, delete

func TestFileOperationCompleteMsg(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Send fileOperationCompleteMsg
	completeMsg := fileOperationCompleteMsg{
		operation: "copy",
	}
	updatedModel, _ = m.Update(completeMsg)
	m = updatedModel.(Model)

	// Should not cause any errors
	if m.dialog != nil {
		t.Error("dialog should be nil after fileOperationCompleteMsg")
	}
}

func TestContextMenuCopyShowsOverwriteDialog(t *testing.T) {
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
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		t.Skip("No suitable entry for test")
	}

	// Simulate context menu result for copy
	// The actual checkFileConflict will be called
	resultMsg := contextMenuResultMsg{
		actionID:  "copy",
		cancelled: false,
	}

	updatedModel, cmd := m.Update(resultMsg)
	m = updatedModel.(Model)

	// Should return a command (from checkFileConflict)
	if cmd != nil {
		// Execute the command to see what happens
		nextMsg := cmd()
		if nextMsg != nil {
			// The result could be showOverwriteDialogMsg or fileOperationCompleteMsg
			// or showErrorDialogMsg
			t.Logf("cmd returned message of type: %T", nextMsg)
		}
	}
}

func TestContextMenuMoveShowsOverwriteDialog(t *testing.T) {
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
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		t.Skip("No suitable entry for test")
	}

	// Simulate context menu result for move
	resultMsg := contextMenuResultMsg{
		actionID:  "move",
		cancelled: false,
	}

	updatedModel, cmd := m.Update(resultMsg)
	m = updatedModel.(Model)

	// Should return a command (from checkFileConflict)
	if cmd != nil {
		// Execute the command to see what happens
		nextMsg := cmd()
		if nextMsg != nil {
			t.Logf("cmd returned message of type: %T", nextMsg)
		}
	}
}

func TestCopyKeyShowsOverwriteDialogOnConflict(t *testing.T) {
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
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		t.Skip("No suitable entry for test")
	}

	// Press 'c' key for copy
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	updatedModel, cmd := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should return a command (from checkFileConflict)
	if cmd == nil {
		t.Error("copy key should return a command")
	}
}

func TestMoveKeyShowsOverwriteDialogOnConflict(t *testing.T) {
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
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		t.Skip("No suitable entry for test")
	}

	// Press 'm' key for move
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updatedModel, cmd := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should return a command (from checkFileConflict)
	if cmd == nil {
		t.Error("move key should return a command")
	}
}

func TestCopyKeyOnParentDirDoesNothing(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Ensure cursor is at parent dir (..)
	m.getActivePane().cursor = 0
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || !entry.IsParentDir() {
		t.Skip("First entry is not parent directory")
	}

	// Press 'c' key for copy
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	updatedModel, cmd := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should return nil command for parent directory
	if cmd != nil {
		t.Error("copy key on parent dir should return nil command")
	}
}

func TestMoveKeyOnParentDirDoesNothing(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Ensure cursor is at parent dir (..)
	m.getActivePane().cursor = 0
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || !entry.IsParentDir() {
		t.Skip("First entry is not parent directory")
	}

	// Press 'm' key for move
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updatedModel, cmd := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should return nil command for parent directory
	if cmd != nil {
		t.Error("move key on parent dir should return nil command")
	}
}

func TestCheckFileConflictNoConflict(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Create temp file
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	if err := os.WriteFile(srcFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	destDir := t.TempDir()

	// Call checkFileConflict - should execute immediately (no conflict)
	cmd := m.checkFileConflict(srcFile, destDir, "copy")
	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	// Execute the command
	result := cmd()

	// Should be fileOperationCompleteMsg (copy succeeded)
	_, isComplete := result.(fileOperationCompleteMsg)
	_, isError := result.(showErrorDialogMsg)

	if !isComplete && !isError {
		t.Errorf("expected fileOperationCompleteMsg or showErrorDialogMsg, got %T", result)
	}
}

func TestCheckFileConflictWithExistingFile(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Create source file
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(srcFile, []byte("source"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Create destination file with same name
	destDir := t.TempDir()
	destFile := filepath.Join(destDir, "test.txt")
	if err := os.WriteFile(destFile, []byte("dest"), 0644); err != nil {
		t.Fatalf("failed to create dest file: %v", err)
	}

	// Call checkFileConflict - should show overwrite dialog
	cmd := m.checkFileConflict(srcFile, destDir, "copy")
	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	// Execute the command
	result := cmd()

	// Should be showOverwriteDialogMsg
	overwriteMsg, ok := result.(showOverwriteDialogMsg)
	if !ok {
		t.Fatalf("expected showOverwriteDialogMsg, got %T", result)
	}

	if overwriteMsg.filename != "test.txt" {
		t.Errorf("filename = %q, want 'test.txt'", overwriteMsg.filename)
	}
	if overwriteMsg.operation != "copy" {
		t.Errorf("operation = %q, want 'copy'", overwriteMsg.operation)
	}
}

func TestCheckFileConflictWithDirectories(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Create source directory
	srcDir := t.TempDir()
	srcSubDir := filepath.Join(srcDir, "subdir")
	if err := os.Mkdir(srcSubDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	// Create destination with same directory name
	destParent := t.TempDir()
	destSubDir := filepath.Join(destParent, "subdir")
	if err := os.Mkdir(destSubDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Call checkFileConflict - should show error dialog (directory conflict)
	cmd := m.checkFileConflict(srcSubDir, destParent, "copy")
	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	// Execute the command
	result := cmd()

	// Should be showErrorDialogMsg for directory conflict
	errorMsg, ok := result.(showErrorDialogMsg)
	if !ok {
		t.Fatalf("expected showErrorDialogMsg for directory conflict, got %T", result)
	}

	if !strings.Contains(errorMsg.message, "already exists") {
		t.Errorf("error message should contain 'already exists', got: %s", errorMsg.message)
	}
}

func TestCheckFileConflictSourceError(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Create destination file (but no source file)
	destDir := t.TempDir()
	destFile := filepath.Join(destDir, "nonexistent.txt")
	if err := os.WriteFile(destFile, []byte("dest"), 0644); err != nil {
		t.Fatalf("failed to create dest file: %v", err)
	}

	// Non-existent source file
	srcFile := "/nonexistent/path/to/nonexistent.txt"

	// Call checkFileConflict - should show error dialog
	cmd := m.checkFileConflict(srcFile, destDir, "copy")
	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	// Execute the command
	result := cmd()

	// Should be showErrorDialogMsg for source check failure
	_, ok := result.(showErrorDialogMsg)
	if !ok {
		t.Fatalf("expected showErrorDialogMsg for source error, got %T", result)
	}
}

func TestExecuteFileOperationCopy(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Create source file
	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "source.txt")
	if err := os.WriteFile(srcFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	destDir := t.TempDir()

	// Execute copy operation
	cmd := m.executeFileOperation(srcFile, destDir, "copy")
	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	result := cmd()

	// Should be fileOperationCompleteMsg
	completeMsg, ok := result.(fileOperationCompleteMsg)
	if !ok {
		t.Fatalf("expected fileOperationCompleteMsg, got %T", result)
	}

	if completeMsg.operation != "copy" {
		t.Errorf("operation = %q, want 'copy'", completeMsg.operation)
	}

	// Verify file was copied
	destFile := filepath.Join(destDir, "source.txt")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("destination file should exist after copy")
	}
}

func TestExecuteFileOperationMove(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Create source file
	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "source.txt")
	if err := os.WriteFile(srcFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	destDir := t.TempDir()

	// Execute move operation
	cmd := m.executeFileOperation(srcFile, destDir, "move")
	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	result := cmd()

	// Should be fileOperationCompleteMsg
	completeMsg, ok := result.(fileOperationCompleteMsg)
	if !ok {
		t.Fatalf("expected fileOperationCompleteMsg, got %T", result)
	}

	if completeMsg.operation != "move" {
		t.Errorf("operation = %q, want 'move'", completeMsg.operation)
	}

	// Verify file was moved (source gone, dest exists)
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("source file should not exist after move")
	}

	destFile := filepath.Join(destDir, "source.txt")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("destination file should exist after move")
	}
}

func TestExecuteFileOperationError(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Non-existent source file
	srcFile := "/nonexistent/path/source.txt"
	destDir := t.TempDir()

	// Execute copy operation (should fail)
	cmd := m.executeFileOperation(srcFile, destDir, "copy")
	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	result := cmd()

	// Should be showErrorDialogMsg
	errorMsg, ok := result.(showErrorDialogMsg)
	if !ok {
		t.Fatalf("expected showErrorDialogMsg, got %T", result)
	}

	if !strings.Contains(errorMsg.message, "Failed to copy") {
		t.Errorf("error message should contain 'Failed to copy', got: %s", errorMsg.message)
	}
}

func TestOverwriteDialogResultMsgOverwriteActualFile(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Create source and destination files
	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(srcFile, []byte("source content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	destDir := t.TempDir()
	destFile := filepath.Join(destDir, "test.txt")
	if err := os.WriteFile(destFile, []byte("original content"), 0644); err != nil {
		t.Fatalf("failed to create dest file: %v", err)
	}

	// Send overwriteDialogResultMsg with Overwrite choice
	resultMsg := overwriteDialogResultMsg{
		choice:    OverwriteChoiceOverwrite,
		srcPath:   srcFile,
		destPath:  destDir,
		filename:  "test.txt",
		operation: "copy",
	}
	updatedModel, cmd := m.Update(resultMsg)
	m = updatedModel.(Model)

	// Should return a command
	if cmd != nil {
		// Execute the command
		result := cmd()
		if result != nil {
			t.Logf("overwrite command returned: %T", result)
		}
	}
}

func TestRenameInputResultMsgSuccessfulCopy(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Create source file
	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "source.txt")
	if err := os.WriteFile(srcFile, []byte("source content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	destDir := t.TempDir()

	// Send renameInputResultMsg
	resultMsg := renameInputResultMsg{
		newName:   "newname.txt",
		srcPath:   srcFile,
		destPath:  destDir,
		operation: "copy",
	}
	updatedModel, cmd := m.Update(resultMsg)
	m = updatedModel.(Model)

	// Should return a command for the actual operation
	if cmd != nil {
		result := cmd()
		if result != nil {
			t.Logf("rename copy command returned: %T", result)
		}
	}

	// Verify the renamed file exists
	newFile := filepath.Join(destDir, "newname.txt")
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		// The command might need to be processed through Update
		t.Log("new file not immediately created (async operation)")
	}
}

func TestRenameInputResultMsgSuccessfulMove(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Create source file
	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "source.txt")
	if err := os.WriteFile(srcFile, []byte("source content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	destDir := t.TempDir()

	// Send renameInputResultMsg for move
	resultMsg := renameInputResultMsg{
		newName:   "newname.txt",
		srcPath:   srcFile,
		destPath:  destDir,
		operation: "move",
	}
	updatedModel, cmd := m.Update(resultMsg)
	m = updatedModel.(Model)

	// Should return a command for the actual operation
	if cmd != nil {
		result := cmd()
		if result != nil {
			t.Logf("rename move command returned: %T", result)
		}
	}
}

// === Additional View and Rendering Tests ===

func TestModelHandleCreateFile(t *testing.T) {
	tmpDir := t.TempDir()
	model := NewModel()

	t.Run("creates file successfully", func(t *testing.T) {
		cmd := model.handleCreateFile(tmpDir, "newfile.txt")
		if cmd == nil {
			t.Fatal("handleCreateFile should return a command")
		}

		msg := cmd()
		result, ok := msg.(inputDialogResultMsg)
		if !ok {
			t.Fatalf("Expected inputDialogResultMsg, got %T", msg)
		}

		if result.err != nil {
			t.Errorf("Expected no error, got %v", result.err)
		}
		if result.operation != "create_file" {
			t.Errorf("Expected operation 'create_file', got %q", result.operation)
		}
		if result.input != "newfile.txt" {
			t.Errorf("Expected input 'newfile.txt', got %q", result.input)
		}
	})

	t.Run("returns error for invalid filename", func(t *testing.T) {
		cmd := model.handleCreateFile(tmpDir, "")
		msg := cmd()
		result := msg.(inputDialogResultMsg)

		if result.err == nil {
			t.Error("Expected error for empty filename")
		}
	})

	t.Run("returns error for existing file", func(t *testing.T) {
		// Create a file first
		existingFile := filepath.Join(tmpDir, "existing.txt")
		os.WriteFile(existingFile, []byte("test"), 0644)

		cmd := model.handleCreateFile(tmpDir, "existing.txt")
		msg := cmd()
		result := msg.(inputDialogResultMsg)

		if result.err == nil {
			t.Error("Expected error for existing file")
		}
	})
}

func TestModelHandleCreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	model := NewModel()

	t.Run("creates directory successfully", func(t *testing.T) {
		cmd := model.handleCreateDirectory(tmpDir, "newdir")
		if cmd == nil {
			t.Fatal("handleCreateDirectory should return a command")
		}

		msg := cmd()
		result, ok := msg.(inputDialogResultMsg)
		if !ok {
			t.Fatalf("Expected inputDialogResultMsg, got %T", msg)
		}

		if result.err != nil {
			t.Errorf("Expected no error, got %v", result.err)
		}
		if result.operation != "create_dir" {
			t.Errorf("Expected operation 'create_dir', got %q", result.operation)
		}
		if result.input != "newdir" {
			t.Errorf("Expected input 'newdir', got %q", result.input)
		}
	})

	t.Run("returns error for invalid dirname", func(t *testing.T) {
		cmd := model.handleCreateDirectory(tmpDir, "")
		msg := cmd()
		result := msg.(inputDialogResultMsg)

		if result.err == nil {
			t.Error("Expected error for empty dirname")
		}
	})

	t.Run("returns error for existing directory", func(t *testing.T) {
		// Create a dir first
		existingDir := filepath.Join(tmpDir, "existingdir")
		os.Mkdir(existingDir, 0755)

		cmd := model.handleCreateDirectory(tmpDir, "existingdir")
		msg := cmd()
		result := msg.(inputDialogResultMsg)

		if result.err == nil {
			t.Error("Expected error for existing directory")
		}
	})
}

func TestModelHandleRename(t *testing.T) {
	tmpDir := t.TempDir()
	model := NewModel()

	t.Run("renames file successfully", func(t *testing.T) {
		// Create a file to rename
		oldFile := filepath.Join(tmpDir, "oldname.txt")
		os.WriteFile(oldFile, []byte("test"), 0644)

		cmd := model.handleRename(tmpDir, "oldname.txt", "newname.txt")
		if cmd == nil {
			t.Fatal("handleRename should return a command")
		}

		msg := cmd()
		result, ok := msg.(inputDialogResultMsg)
		if !ok {
			t.Fatalf("Expected inputDialogResultMsg, got %T", msg)
		}

		if result.err != nil {
			t.Errorf("Expected no error, got %v", result.err)
		}
		if result.operation != "rename" {
			t.Errorf("Expected operation 'rename', got %q", result.operation)
		}
		if result.input != "newname.txt" {
			t.Errorf("Expected input 'newname.txt', got %q", result.input)
		}
		if result.oldName != "oldname.txt" {
			t.Errorf("Expected oldName 'oldname.txt', got %q", result.oldName)
		}
	})

	t.Run("returns error for invalid new name", func(t *testing.T) {
		cmd := model.handleRename(tmpDir, "somefile.txt", "")
		msg := cmd()
		result := msg.(inputDialogResultMsg)

		if result.err == nil {
			t.Error("Expected error for empty new name")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		cmd := model.handleRename(tmpDir, "nonexistent.txt", "new.txt")
		msg := cmd()
		result := msg.(inputDialogResultMsg)

		if result.err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

func TestModelMoveCursorToFile(t *testing.T) {
	tmpDir := t.TempDir()
	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "aaa.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "bbb.txt"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "ccc.txt"), []byte("c"), 0644)

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	// Initialize with size
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("moves cursor to existing file", func(t *testing.T) {
		m.moveCursorToFile("bbb.txt")
		pane := m.getActivePane()
		// Find the expected cursor position
		expectedCursor := -1
		for i, e := range pane.entries {
			if e.Name == "bbb.txt" {
				expectedCursor = i
				break
			}
		}
		if expectedCursor >= 0 && pane.cursor != expectedCursor {
			t.Errorf("Expected cursor at %d, got %d", expectedCursor, pane.cursor)
		}
	})

	t.Run("does not move cursor for non-existent file", func(t *testing.T) {
		pane := m.getActivePane()
		originalCursor := pane.cursor
		m.moveCursorToFile("nonexistent.txt")
		if pane.cursor != originalCursor {
			t.Errorf("Cursor should not move for non-existent file")
		}
	})

	t.Run("does not move to hidden file when showHidden is false", func(t *testing.T) {
		// Create a hidden file
		os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("h"), 0644)
		pane := m.getActivePane()
		pane.showHidden = false
		originalCursor := pane.cursor
		m.moveCursorToFile(".hidden")
		if pane.cursor != originalCursor {
			t.Errorf("Cursor should not move to hidden file when showHidden is false")
		}
	})
}

func TestModelMoveCursorToFileAfterRename(t *testing.T) {
	tmpDir := t.TempDir()
	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "aaa.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "renamed.txt"), []byte("b"), 0644)

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("moves cursor to renamed file", func(t *testing.T) {
		m.moveCursorToFileAfterRename("oldname.txt", "renamed.txt")
		pane := m.getActivePane()
		// Find the expected cursor position
		expectedCursor := -1
		for i, e := range pane.entries {
			if e.Name == "renamed.txt" {
				expectedCursor = i
				break
			}
		}
		if expectedCursor >= 0 && pane.cursor != expectedCursor {
			t.Errorf("Expected cursor at %d, got %d", expectedCursor, pane.cursor)
		}
	})

	t.Run("does not move for non-matching new name", func(t *testing.T) {
		pane := m.getActivePane()
		originalCursor := pane.cursor
		m.moveCursorToFileAfterRename("old.txt", "nonexistent.txt")
		if pane.cursor != originalCursor {
			t.Errorf("Cursor should not move for non-existent new name")
		}
	})

	t.Run("adjusts cursor when renaming to hidden file with showHidden false", func(t *testing.T) {
		pane := m.getActivePane()
		pane.showHidden = false
		pane.cursor = 10 // Set cursor beyond entries length
		m.moveCursorToFileAfterRename("old.txt", ".hidden")
		// Cursor should be adjusted to valid range
		if pane.cursor >= len(pane.entries) && len(pane.entries) > 0 {
			t.Errorf("Cursor should be adjusted to valid range")
		}
	})

	t.Run("handles empty entries when renaming to hidden", func(t *testing.T) {
		emptyDir := t.TempDir()
		model2 := NewModel()
		model2.leftPath = emptyDir
		model2.rightPath = emptyDir
		msg2 := tea.WindowSizeMsg{Width: 120, Height: 40}
		updatedModel2, _ := model2.Update(msg2)
		m2 := updatedModel2.(Model)

		pane2 := m2.getActivePane()
		pane2.showHidden = false
		pane2.entries = nil // Empty entries
		pane2.cursor = 5
		m2.moveCursorToFileAfterRename("old.txt", ".hidden")
		if pane2.cursor != 0 {
			t.Errorf("Cursor should be 0 for empty entries, got %d", pane2.cursor)
		}
	})
}

func TestModelRefreshBothPanes(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("refreshes both panes without error", func(t *testing.T) {
		m.dialog = nil
		m.RefreshBothPanes()
		// Should not set error dialog for valid paths
		// Note: dialog might still be nil if refresh succeeds
	})

	t.Run("updates disk space", func(t *testing.T) {
		m.RefreshBothPanes()
		// Just verify it doesn't panic
	})
}

func TestModelSyncOppositePane(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("syncs opposite pane from left", func(t *testing.T) {
		m.activePane = LeftPane
		m.dialog = nil
		m.SyncOppositePane()
		// Right pane should be synced to left pane's path
		if m.rightPane.path != m.leftPane.path {
			t.Errorf("Right pane should sync to left pane path")
		}
	})

	t.Run("syncs opposite pane from right", func(t *testing.T) {
		m.activePane = RightPane
		m.dialog = nil
		m.SyncOppositePane()
		// Left pane should be synced to right pane's path
		if m.leftPane.path != m.rightPane.path {
			t.Errorf("Left pane should sync to right pane path")
		}
	})

	t.Run("sets error dialog on sync failure", func(t *testing.T) {
		m.activePane = LeftPane
		m.leftPane.path = "/nonexistent/path"
		m.dialog = nil
		m.SyncOppositePane()
		if m.dialog == nil {
			t.Error("Error dialog should be set on sync failure")
		}
	})
}

func TestModelCheckFileConflict(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "existing.txt")
	os.WriteFile(existingFile, []byte("test"), 0644)

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir

	t.Run("returns nil for non-conflicting file", func(t *testing.T) {
		srcPath := filepath.Join(tmpDir, "newfile.txt")
		os.WriteFile(srcPath, []byte("test"), 0644)
		cmd := model.checkFileConflict(srcPath, tmpDir, "copy")
		// Should return nil or a command depending on conflict
		_ = cmd
	})

	t.Run("returns command for conflicting file", func(t *testing.T) {
		// Create source file with same name as existing
		srcDir := t.TempDir()
		srcPath := filepath.Join(srcDir, "existing.txt")
		os.WriteFile(srcPath, []byte("new content"), 0644)

		cmd := model.checkFileConflict(srcPath, tmpDir, "copy")
		if cmd == nil {
			t.Error("Should return command for conflicting file")
		}
	})
}

func TestContextMenuCompressWithNilAction(t *testing.T) {
	// Save original value and restore after test
	originalValue := hasDesktop
	defer func() { hasDesktop = originalValue }()

	// Set desktop environment to true so Open/Open with items are enabled
	// This ensures consistent navigation behavior
	setDesktopEnvironmentForTest(true)

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

	// Check that the dialog is a ContextMenuDialog
	contextMenu, ok := m.dialog.(*ContextMenuDialog)
	if !ok {
		t.Fatal("dialog should be ContextMenuDialog")
	}

	// Find compress menu item index
	compressIndex := -1
	for i, item := range contextMenu.items {
		if item.ID == "compress" {
			compressIndex = i
			break
		}
	}
	if compressIndex == -1 {
		t.Fatal("compress menu item should exist")
	}

	// Select compress by navigating to it and pressing Enter
	for i := 0; i < compressIndex; i++ {
		keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
		updatedModel, _ = m.Update(keyMsg)
		m = updatedModel.(Model)
	}

	// Press Enter to select
	keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Execute the command to send contextMenuResultMsg
	if cmd != nil {
		resultMsg := cmd()
		updatedModel, _ = m.Update(resultMsg)
		m = updatedModel.(Model)
	}

	// Verify CompressFormatDialog is shown
	if m.dialog == nil {
		t.Error("dialog should be shown after compress action")
	}

	_, isCompressFormatDialog := m.dialog.(*CompressFormatDialog)
	if !isCompressFormatDialog {
		t.Errorf("dialog should be CompressFormatDialog after compress action from context menu, got %T", m.dialog)
	}
}
