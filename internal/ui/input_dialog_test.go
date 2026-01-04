package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInputDialog_New(t *testing.T) {
	title := "Test Title:"
	dialog := NewInputDialog(title, nil)

	if dialog.title != title {
		t.Errorf("Expected title %q, got %q", title, dialog.title)
	}
	if dialog.Input() != "" {
		t.Error("Input should be empty initially")
	}
	if dialog.CursorPos() != 0 {
		t.Error("Cursor position should be 0 initially")
	}
	if !dialog.active {
		t.Error("Dialog should be active when created")
	}
	if !dialog.IsActive() {
		t.Error("IsActive() should return true")
	}
	if dialog.DisplayType() != DialogDisplayPane {
		t.Error("DisplayType should be DialogDisplayPane")
	}
}

func TestInputDialog_CharacterInput(t *testing.T) {
	dialog := NewInputDialog("Test:", nil)

	// Type "hello"
	keys := []string{"h", "e", "l", "l", "o"}
	for _, key := range keys {
		dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	}

	if dialog.Input() != "hello" {
		t.Errorf("Expected input %q, got %q", "hello", dialog.Input())
	}
	if dialog.CursorPos() != 5 {
		t.Errorf("Expected cursor at 5, got %d", dialog.CursorPos())
	}
}

func TestInputDialog_CursorMovement(t *testing.T) {
	dialog := NewInputDialog("Test:", nil)
	dialog.SetInput("hello")

	// Move left
	dialog.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if dialog.CursorPos() != 4 {
		t.Errorf("Expected cursor at 4 after left, got %d", dialog.CursorPos())
	}

	// Move left again
	dialog.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if dialog.CursorPos() != 3 {
		t.Errorf("Expected cursor at 3 after left, got %d", dialog.CursorPos())
	}

	// Move right
	dialog.Update(tea.KeyMsg{Type: tea.KeyRight})
	if dialog.CursorPos() != 4 {
		t.Errorf("Expected cursor at 4 after right, got %d", dialog.CursorPos())
	}

	// Move to beginning (Ctrl+A)
	dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	if dialog.CursorPos() != 0 {
		t.Errorf("Expected cursor at 0 after Ctrl+A, got %d", dialog.CursorPos())
	}

	// Try to move left at beginning
	dialog.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if dialog.CursorPos() != 0 {
		t.Errorf("Cursor should stay at 0, got %d", dialog.CursorPos())
	}

	// Move to end (Ctrl+E)
	dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	if dialog.CursorPos() != 5 {
		t.Errorf("Expected cursor at 5 after Ctrl+E, got %d", dialog.CursorPos())
	}

	// Try to move right at end
	dialog.Update(tea.KeyMsg{Type: tea.KeyRight})
	if dialog.CursorPos() != 5 {
		t.Errorf("Cursor should stay at 5, got %d", dialog.CursorPos())
	}
}

func TestInputDialog_Backspace(t *testing.T) {
	dialog := NewInputDialog("Test:", nil)
	dialog.SetInput("hello")

	// Delete last character
	dialog.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if dialog.Input() != "hell" {
		t.Errorf("Expected %q after backspace, got %q", "hell", dialog.Input())
	}
	if dialog.CursorPos() != 4 {
		t.Errorf("Expected cursor at 4, got %d", dialog.CursorPos())
	}

	// Move cursor to middle and backspace
	dialog.textInput.CursorPos = 2
	dialog.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if dialog.Input() != "hll" {
		t.Errorf("Expected %q after backspace at position 2, got %q", "hll", dialog.Input())
	}
	if dialog.CursorPos() != 1 {
		t.Errorf("Expected cursor at 1, got %d", dialog.CursorPos())
	}

	// Backspace at beginning should do nothing
	dialog.textInput.CursorPos = 0
	dialog.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if dialog.Input() != "hll" {
		t.Errorf("Input should not change when backspace at beginning, got %q", dialog.Input())
	}
}

func TestInputDialog_Delete(t *testing.T) {
	dialog := NewInputDialog("Test:", nil)
	dialog.textInput.Value = "hello"
	dialog.textInput.CursorPos = 0

	// Delete first character
	dialog.Update(tea.KeyMsg{Type: tea.KeyDelete})
	if dialog.Input() != "ello" {
		t.Errorf("Expected %q after delete, got %q", "ello", dialog.Input())
	}

	// Delete at middle
	dialog.textInput.CursorPos = 2
	dialog.Update(tea.KeyMsg{Type: tea.KeyDelete})
	if dialog.Input() != "elo" {
		t.Errorf("Expected %q after delete at position 2, got %q", "elo", dialog.Input())
	}

	// Delete at end should do nothing
	dialog.textInput.CursorPos = 3
	dialog.Update(tea.KeyMsg{Type: tea.KeyDelete})
	if dialog.Input() != "elo" {
		t.Errorf("Input should not change when delete at end, got %q", dialog.Input())
	}
}

func TestInputDialog_CtrlU_CtrlK(t *testing.T) {
	t.Run("Ctrl+U deletes to beginning", func(t *testing.T) {
		dialog := NewInputDialog("Test:", nil)
		dialog.textInput.Value = "hello world"
		dialog.textInput.CursorPos = 6

		dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
		if dialog.Input() != "world" {
			t.Errorf("Expected %q after Ctrl+U, got %q", "world", dialog.Input())
		}
		if dialog.CursorPos() != 0 {
			t.Errorf("Cursor should be at 0, got %d", dialog.CursorPos())
		}
	})

	t.Run("Ctrl+K deletes to end", func(t *testing.T) {
		dialog := NewInputDialog("Test:", nil)
		dialog.textInput.Value = "hello world"
		dialog.textInput.CursorPos = 6

		dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
		if dialog.Input() != "hello " {
			t.Errorf("Expected %q after Ctrl+K, got %q", "hello ", dialog.Input())
		}
		if dialog.CursorPos() != 6 {
			t.Errorf("Cursor should stay at 6, got %d", dialog.CursorPos())
		}
	})
}

func TestInputDialog_EnterConfirm(t *testing.T) {
	confirmCalled := false
	inputReceived := ""

	dialog := NewInputDialog("Test:", func(input string) tea.Cmd {
		confirmCalled = true
		inputReceived = input
		return nil
	})

	dialog.SetInput("testfile.txt")

	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !confirmCalled {
		t.Error("Confirm callback should be called")
	}
	if inputReceived != "testfile.txt" {
		t.Errorf("Expected input %q, got %q", "testfile.txt", inputReceived)
	}
	if dialog.active {
		t.Error("Dialog should be inactive after confirm")
	}
	// cmd should be nil since our callback returns nil
	if cmd != nil {
		t.Error("Command should be nil")
	}
}

func TestInputDialog_EscCancel(t *testing.T) {
	confirmCalled := false

	dialog := NewInputDialog("Test:", func(input string) tea.Cmd {
		confirmCalled = true
		return nil
	})

	dialog.SetInput("testfile.txt")

	newDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if confirmCalled {
		t.Error("Confirm callback should not be called on cancel")
	}
	if newDialog.IsActive() {
		t.Error("Dialog should be inactive after cancel")
	}

	// CRITICAL: Verify cancel message is returned
	if cmd == nil {
		t.Fatal("Esc key should return a command, got nil - this will cause the freeze bug")
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

func TestInputDialog_EmptyInputError(t *testing.T) {
	confirmCalled := false

	dialog := NewInputDialog("Test:", func(input string) tea.Cmd {
		confirmCalled = true
		return nil
	})

	// Try to confirm with empty input
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if confirmCalled {
		t.Error("Confirm callback should not be called for empty input")
	}
	if !dialog.active {
		t.Error("Dialog should stay active for empty input")
	}
	if dialog.errorMsg == "" {
		t.Error("Error message should be set for empty input")
	}
}

func TestInputDialog_View(t *testing.T) {
	dialog := NewInputDialog("New file:", nil)
	dialog.SetInput("test.txt")

	view := dialog.View()

	// Check that view contains the title
	if !strings.Contains(view, "New file:") {
		t.Error("View should contain the title")
	}

	// Check that view contains the input
	if !strings.Contains(view, "test") {
		t.Error("View should contain the input text")
	}

	// Check that view contains the footer
	if !strings.Contains(view, "Enter") && !strings.Contains(view, "Esc") {
		t.Error("View should contain help text")
	}
}

func TestInputDialog_ViewWithError(t *testing.T) {
	dialog := NewInputDialog("New file:", nil)
	dialog.errorMsg = "File name cannot be empty"

	view := dialog.View()

	// Check that view contains the error message
	if !strings.Contains(view, "File name cannot be empty") {
		t.Error("View should contain the error message")
	}
}

func TestInputDialog_UnicodeInput(t *testing.T) {
	dialog := NewInputDialog("Test:", nil)

	// Type Japanese characters
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("テ")})
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ス")})
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ト")})

	if dialog.Input() != "テスト" {
		t.Errorf("Expected input %q, got %q", "テスト", dialog.Input())
	}
	if dialog.CursorPos() != 3 {
		t.Errorf("Expected cursor at 3, got %d", dialog.CursorPos())
	}
}

func TestInputDialog_InactiveDoesNotProcess(t *testing.T) {
	dialog := NewInputDialog("Test:", nil)
	dialog.active = false
	dialog.textInput.Value = "initial"

	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if dialog.Input() != "initial" {
		t.Error("Inactive dialog should not process input")
	}
}

func TestInputDialog_SetWidth(t *testing.T) {
	dialog := NewInputDialog("Test:", nil)

	dialog.SetWidth(80)

	if dialog.width != 80 {
		t.Errorf("SetWidth(80) width = %d, want 80", dialog.width)
	}
}

func TestInputDialog_SetEmptyErrorMsg(t *testing.T) {
	dialog := NewInputDialog("Test:", nil)

	customMsg := "Custom error message"
	dialog.SetEmptyErrorMsg(customMsg)

	if dialog.emptyErrorMsg != customMsg {
		t.Errorf("SetEmptyErrorMsg() did not set message, got %q, want %q", dialog.emptyErrorMsg, customMsg)
	}
}

func TestInputDialog_LongInputView(t *testing.T) {
	dialog := NewInputDialog("Test:", nil)

	// Create a long input string that exceeds typical display width
	longInput := "this_is_a_very_long_filename_that_exceeds_the_normal_display_width.txt"
	dialog.SetInput(longInput)
	dialog.width = 50

	view := dialog.View()

	// View should still contain content
	if view == "" {
		t.Error("View should not be empty for long input")
	}

	// View should still contain the title
	if !strings.Contains(view, "Test:") {
		t.Error("View should contain the title")
	}
}

func TestInputDialog_CursorAtMiddleOfLongInput(t *testing.T) {
	dialog := NewInputDialog("Test:", nil)

	// Create a long input string
	longInput := "this_is_a_very_long_filename_that_exceeds_the_normal_display_width.txt"
	dialog.textInput.Value = longInput
	dialog.textInput.CursorPos = len(longInput) / 2
	dialog.width = 50

	view := dialog.View()

	if view == "" {
		t.Error("View should not be empty for long input with cursor in middle")
	}
}
