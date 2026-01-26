package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewMoveToTrashDialog_SingleFile(t *testing.T) {
	paths := []string{"/home/user/file.txt"}
	dialog := NewMoveToTrashDialog(paths)

	if dialog == nil {
		t.Fatal("expected dialog to be created")
	}

	if !dialog.IsActive() {
		t.Error("expected dialog to be active")
	}

	if dialog.DisplayType() != DialogDisplayPane {
		t.Errorf("expected DialogDisplayPane, got %v", dialog.DisplayType())
	}

	if len(dialog.paths) != 1 {
		t.Errorf("expected 1 path, got %d", len(dialog.paths))
	}

	if dialog.paths[0] != "/home/user/file.txt" {
		t.Errorf("expected path '/home/user/file.txt', got '%s'", dialog.paths[0])
	}
}

func TestNewMoveToTrashDialog_MultipleFiles(t *testing.T) {
	paths := []string{
		"/home/user/file1.txt",
		"/home/user/file2.txt",
		"/home/user/file3.txt",
	}
	dialog := NewMoveToTrashDialog(paths)

	if dialog == nil {
		t.Fatal("expected dialog to be created")
	}

	if len(dialog.paths) != 3 {
		t.Errorf("expected 3 paths, got %d", len(dialog.paths))
	}
}

func TestMoveToTrashDialog_View_SingleFile(t *testing.T) {
	paths := []string{"/home/user/document.txt"}
	dialog := NewMoveToTrashDialog(paths)

	view := dialog.View()

	// Check title
	if !strings.Contains(view, "Move to Trash") {
		t.Error("expected view to contain title 'Move to Trash'")
	}

	// Check filename is displayed
	if !strings.Contains(view, "document.txt") {
		t.Error("expected view to contain filename 'document.txt'")
	}

	// Check warning message (singular form for single file)
	if !strings.Contains(view, "File will not be permanently deleted") {
		t.Error("expected view to contain singular warning 'File will not be permanently deleted'")
	}

	if !strings.Contains(view, "Disk space will not be freed") {
		t.Error("expected view to contain warning about disk space")
	}

	// Check buttons
	if !strings.Contains(view, "[Y]es") {
		t.Error("expected view to contain '[Y]es' button")
	}

	if !strings.Contains(view, "[N]o") {
		t.Error("expected view to contain '[N]o' button")
	}
}

func TestMoveToTrashDialog_View_MultipleFiles(t *testing.T) {
	paths := []string{
		"/home/user/file1.txt",
		"/home/user/file2.txt",
		"/home/user/file3.txt",
	}
	dialog := NewMoveToTrashDialog(paths)

	view := dialog.View()

	// Check title
	if !strings.Contains(view, "Move to Trash") {
		t.Error("expected view to contain title 'Move to Trash'")
	}

	// Check item count is displayed
	if !strings.Contains(view, "3 items") {
		t.Error("expected view to contain '3 items'")
	}

	// Check warning message (plural form for multiple files)
	if !strings.Contains(view, "Files will not be permanently deleted") {
		t.Error("expected view to contain plural warning 'Files will not be permanently deleted'")
	}
}

func TestMoveToTrashDialog_Update_YKey(t *testing.T) {
	paths := []string{"/home/user/file.txt"}
	dialog := NewMoveToTrashDialog(paths)

	// Press 'y' key
	updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	if updatedDialog.IsActive() {
		t.Error("expected dialog to be closed after pressing 'y'")
	}

	if cmd == nil {
		t.Fatal("expected command to be returned")
	}

	// Execute the command and check the message
	msg := cmd()
	resultMsg, ok := msg.(trashConfirmResultMsg)
	if !ok {
		t.Fatalf("expected trashConfirmResultMsg, got %T", msg)
	}

	if !resultMsg.confirmed {
		t.Error("expected confirmed to be true")
	}

	if len(resultMsg.paths) != 1 {
		t.Errorf("expected 1 path, got %d", len(resultMsg.paths))
	}
}

func TestMoveToTrashDialog_Update_YKeyUppercase(t *testing.T) {
	paths := []string{"/home/user/file.txt"}
	dialog := NewMoveToTrashDialog(paths)

	// Press 'Y' key (uppercase)
	updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}})

	if updatedDialog.IsActive() {
		t.Error("expected dialog to be closed after pressing 'Y'")
	}

	if cmd == nil {
		t.Fatal("expected command to be returned")
	}

	msg := cmd()
	resultMsg, ok := msg.(trashConfirmResultMsg)
	if !ok {
		t.Fatalf("expected trashConfirmResultMsg, got %T", msg)
	}

	if !resultMsg.confirmed {
		t.Error("expected confirmed to be true")
	}
}

func TestMoveToTrashDialog_Update_NKey(t *testing.T) {
	paths := []string{"/home/user/file.txt"}
	dialog := NewMoveToTrashDialog(paths)

	// Press 'n' key
	updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if updatedDialog.IsActive() {
		t.Error("expected dialog to be closed after pressing 'n'")
	}

	if cmd == nil {
		t.Fatal("expected command to be returned")
	}

	msg := cmd()
	resultMsg, ok := msg.(trashConfirmResultMsg)
	if !ok {
		t.Fatalf("expected trashConfirmResultMsg, got %T", msg)
	}

	if resultMsg.confirmed {
		t.Error("expected confirmed to be false")
	}
}

func TestMoveToTrashDialog_Update_NKeyUppercase(t *testing.T) {
	paths := []string{"/home/user/file.txt"}
	dialog := NewMoveToTrashDialog(paths)

	// Press 'N' key (uppercase)
	updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})

	if updatedDialog.IsActive() {
		t.Error("expected dialog to be closed after pressing 'N'")
	}

	if cmd == nil {
		t.Fatal("expected command to be returned")
	}

	msg := cmd()
	resultMsg, ok := msg.(trashConfirmResultMsg)
	if !ok {
		t.Fatalf("expected trashConfirmResultMsg, got %T", msg)
	}

	if resultMsg.confirmed {
		t.Error("expected confirmed to be false")
	}
}

func TestMoveToTrashDialog_Update_EscKey(t *testing.T) {
	paths := []string{"/home/user/file.txt"}
	dialog := NewMoveToTrashDialog(paths)

	// Press Esc key
	updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if updatedDialog.IsActive() {
		t.Error("expected dialog to be closed after pressing Esc")
	}

	if cmd == nil {
		t.Fatal("expected command to be returned")
	}

	msg := cmd()
	resultMsg, ok := msg.(trashConfirmResultMsg)
	if !ok {
		t.Fatalf("expected trashConfirmResultMsg, got %T", msg)
	}

	if resultMsg.confirmed {
		t.Error("expected confirmed to be false")
	}
}

func TestMoveToTrashDialog_Update_OtherKey(t *testing.T) {
	paths := []string{"/home/user/file.txt"}
	dialog := NewMoveToTrashDialog(paths)

	// Press some other key (e.g., 'x')
	updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	if !updatedDialog.IsActive() {
		t.Error("expected dialog to remain active after pressing unhandled key")
	}

	if cmd != nil {
		t.Error("expected no command for unhandled key")
	}
}

func TestMoveToTrashDialog_View_Inactive(t *testing.T) {
	paths := []string{"/home/user/file.txt"}
	dialog := NewMoveToTrashDialog(paths)
	dialog.Close()

	view := dialog.View()

	if view != "" {
		t.Error("expected empty view when dialog is inactive")
	}
}

func TestMoveToTrashDialog_Update_WhenInactive(t *testing.T) {
	paths := []string{"/home/user/file.txt"}
	dialog := NewMoveToTrashDialog(paths)
	dialog.Close()

	// Try to update when inactive
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	if cmd != nil {
		t.Error("expected no command when dialog is inactive")
	}
}

func TestMoveToTrashDialog_PathsReturnedInMessage(t *testing.T) {
	paths := []string{
		"/home/user/file1.txt",
		"/home/user/file2.txt",
	}
	dialog := NewMoveToTrashDialog(paths)

	// Press 'y' key
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	msg := cmd()
	resultMsg, ok := msg.(trashConfirmResultMsg)
	if !ok {
		t.Fatalf("expected trashConfirmResultMsg, got %T", msg)
	}

	if len(resultMsg.paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(resultMsg.paths))
	}

	if resultMsg.paths[0] != "/home/user/file1.txt" {
		t.Errorf("expected first path '/home/user/file1.txt', got '%s'", resultMsg.paths[0])
	}

	if resultMsg.paths[1] != "/home/user/file2.txt" {
		t.Errorf("expected second path '/home/user/file2.txt', got '%s'", resultMsg.paths[1])
	}
}
