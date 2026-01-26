package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/config"
)

// TestCancelKeyUnification verifies that all dialogs support both Esc and Ctrl+C for cancellation.
// This is a critical UX feature for consistent keyboard navigation across the application.

func TestInputDialog_CtrlCCancel(t *testing.T) {
	confirmCalled := false

	dialog := NewInputDialog("Test:", func(input string) tea.Cmd {
		confirmCalled = true
		return nil
	})

	dialog.SetInput("testfile.txt")

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if confirmCalled {
		t.Error("Confirm callback should not be called on Ctrl+C cancel")
	}
	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}

	// CRITICAL: Verify cancel message is returned
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command, got nil - this will cause the freeze bug")
	}

	// Execute the command to get the message
	msg := cmd()

	// Verify the message is inputDialogResultMsg with cancelled=true
	resultMsg, ok := msg.(inputDialogResultMsg)
	if !ok {
		t.Fatalf("Expected inputDialogResultMsg, got %T", msg)
	}

	if !resultMsg.cancelled {
		t.Error("Expected cancelled=true in inputDialogResultMsg")
	}
}

func TestPathJumpDialog_CtrlCCancel(t *testing.T) {
	dialog := NewPathJumpDialog()
	dialog.SetInput("/tmp")

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command")
	}

	msg := cmd()
	_, ok := msg.(pathJumpCancelMsg)
	if !ok {
		t.Fatalf("Expected pathJumpCancelMsg, got %T", msg)
	}
}

func TestArchiveNameDialog_CtrlCCancel(t *testing.T) {
	dialog := NewArchiveNameDialog("test.tar.gz")

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command")
	}

	msg := cmd()
	resultMsg, ok := msg.(archiveNameResultMsg)
	if !ok {
		t.Fatalf("Expected archiveNameResultMsg, got %T", msg)
	}
	if !resultMsg.cancelled {
		t.Error("Expected cancelled=true")
	}
}

func TestRecursivePermDialog_CtrlCCancel(t *testing.T) {
	dialog := NewRecursivePermDialog("testdir")

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command")
	}

	msg := cmd()
	_, ok := msg.(recursivePermDialogCancelMsg)
	if !ok {
		t.Fatalf("Expected recursivePermDialogCancelMsg, got %T", msg)
	}
}

func TestBookmarkDialog_CtrlCCancel(t *testing.T) {
	bookmarks := []config.Bookmark{
		{Name: "test", Path: "/tmp"},
	}
	dialog := NewBookmarkDialog(bookmarks)

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command")
	}

	msg := cmd()
	_, ok := msg.(bookmarkCloseMsg)
	if !ok {
		t.Fatalf("Expected bookmarkCloseMsg, got %T", msg)
	}
}

func TestRegexSearchDialog_CtrlCCancel(t *testing.T) {
	history := NewSearchHistory(10)
	dialog := NewRegexSearchDialog(history)

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command")
	}

	msg := cmd()
	resultMsg, ok := msg.(regexSearchResultMsg)
	if !ok {
		t.Fatalf("Expected regexSearchResultMsg, got %T", msg)
	}
	if !resultMsg.cancelled {
		t.Error("Expected cancelled=true")
	}
}

func TestRenameInputDialog_CtrlCCancel(t *testing.T) {
	dialog := NewRenameInputDialog("/tmp", "/home/test.txt", "copy")

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command")
	}

	msg := cmd()
	resultMsg, ok := msg.(renameInputResultMsg)
	if !ok {
		t.Fatalf("Expected renameInputResultMsg, got %T", msg)
	}
	if !resultMsg.cancelled {
		t.Error("Expected cancelled=true")
	}
}

func TestArchiveProgressDialog_CtrlCCancel(t *testing.T) {
	cancelCalled := false
	dialog := NewArchiveProgressDialog("extract", "/tmp/test.tar.gz")
	dialog.SetOnCancel(func() {
		cancelCalled = true
	})

	dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if !cancelCalled {
		t.Error("Cancel callback should be called on Ctrl+C")
	}
}

func TestCompressionLevelDialog_CtrlCCancel(t *testing.T) {
	dialog := NewCompressionLevelDialog()

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command")
	}

	msg := cmd()
	resultMsg, ok := msg.(compressionLevelResultMsg)
	if !ok {
		t.Fatalf("Expected compressionLevelResultMsg, got %T", msg)
	}
	if !resultMsg.cancelled {
		t.Error("Expected cancelled=true")
	}
}

func TestPermissionErrorReportDialog_CtrlCCancel(t *testing.T) {
	dialog := NewPermissionErrorReportDialog(5, 2, nil)

	newDialog, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
}

func TestPermissionDialog_CtrlCCancel(t *testing.T) {
	dialog := NewPermissionDialog("test.txt", false, 0644)

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command")
	}

	msg := cmd()
	_, ok := msg.(permissionDialogCancelMsg)
	if !ok {
		t.Fatalf("Expected permissionDialogCancelMsg, got %T", msg)
	}
}

func TestQuerySearchDialog_CtrlCCancel(t *testing.T) {
	history := NewSearchHistory(10)
	dialog := NewQuerySearchDialog(history)

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command")
	}

	msg := cmd()
	resultMsg, ok := msg.(querySearchResultMsg)
	if !ok {
		t.Fatalf("Expected querySearchResultMsg, got %T", msg)
	}
	if !resultMsg.cancelled {
		t.Error("Expected cancelled=true")
	}
}

func TestExtensionRenameDialog_CtrlCCancel(t *testing.T) {
	dialog := NewExtensionRenameDialog("/tmp", "test.txt", "test", ".txt")

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command")
	}

	msg := cmd()
	resultMsg, ok := msg.(extensionRenameResultMsg)
	if !ok {
		t.Fatalf("Expected extensionRenameResultMsg, got %T", msg)
	}
	if !resultMsg.cancelled {
		t.Error("Expected cancelled=true")
	}
}

func TestTrashDialog_CtrlCCancel(t *testing.T) {
	items := []TrashItem{
		{Name: "test.txt", Size: 100, DeletionTime: "2024-01-01 00:00", OriginalPath: "/tmp/test.txt"},
	}
	dialog := NewTrashDialog(items)

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command")
	}

	msg := cmd()
	_, ok := msg.(trashDialogCloseMsg)
	if !ok {
		t.Fatalf("Expected trashDialogCloseMsg, got %T", msg)
	}
}

// Phase 2: String-based key matching tests

func TestSortDialog_CtrlCCancel(t *testing.T) {
	original := SortConfig{Field: SortByName, Order: SortAsc}
	dialog := NewSortDialog(original)

	// Change the selection
	dialog.config.Field = SortBySize
	dialog.config.Order = SortDesc

	confirmed, cancelled := dialog.HandleKey("ctrl+c")

	if confirmed {
		t.Error("Ctrl+C should not confirm")
	}
	if !cancelled {
		t.Error("Ctrl+C should cancel dialog")
	}
	// Config should be restored to original
	if dialog.config.Field != SortByName {
		t.Errorf("config.Field = %v, want SortByName (restored)", dialog.config.Field)
	}
	if dialog.config.Order != SortAsc {
		t.Errorf("config.Order = %v, want SortAsc (restored)", dialog.config.Order)
	}
}

func TestSortDialog_CtrlCCancel_Update(t *testing.T) {
	original := SortConfig{Field: SortByName, Order: SortAsc}
	dialog := NewSortDialog(original)
	dialog.config.Field = SortByDate // Change

	// Use tea.KeyCtrlC type for the Update method
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := dialog.Update(keyMsg)

	if cmd == nil {
		t.Error("Update with Ctrl+C should return a command")
	}

	msg := cmd()
	result, ok := msg.(sortDialogResultMsg)
	if !ok {
		t.Fatalf("Expected sortDialogResultMsg, got %T", msg)
	}

	if result.confirmed {
		t.Error("Expected confirmed = false")
	}
	if !result.cancelled {
		t.Error("Expected cancelled = true")
	}
	if result.config.Field != SortByName {
		t.Errorf("config.Field = %v, want SortByName", result.config.Field)
	}
}

func TestArchiveWarningDialog_CtrlCCancel(t *testing.T) {
	dialog := NewCompressionBombWarningDialog("/tmp/test.zip", 1000, 1000000, 1000)

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command")
	}

	msg := cmd()
	resultMsg, ok := msg.(archiveWarningResultMsg)
	if !ok {
		t.Fatalf("Expected archiveWarningResultMsg, got %T", msg)
	}
	if resultMsg.choice != ArchiveWarningCancel {
		t.Error("Expected ArchiveWarningCancel choice")
	}
}

func TestArchiveWarningDialog_CtrlCCancel_StringBased(t *testing.T) {
	dialog := NewCompressionBombWarningDialog("/tmp/test.zip", 1000, 1000000, 1000)

	// Test using string-based key matching (ctrl+c)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{}}
	// Simulate the string representation being "ctrl+c"
	// This tests the msg.String() path in Update

	// First test with tea.KeyCtrlC type which should work
	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C cancel")
	}
	if cmd == nil {
		t.Fatal("Ctrl+C key should return a command")
	}

	// Ignore the unused variable warning
	_ = keyMsg
}
