package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewQuerySearchDialog(t *testing.T) {
	history := NewSearchHistory(50)
	dialog := NewQuerySearchDialog(history)

	if dialog == nil {
		t.Fatal("NewQuerySearchDialog returned nil")
	}

	if !dialog.IsActive() {
		t.Error("dialog should be active when created")
	}

	if dialog.DisplayType() != DialogDisplayPane {
		t.Errorf("DisplayType = %v, want DialogDisplayPane", dialog.DisplayType())
	}
}

func TestQuerySearchDialog_View(t *testing.T) {
	history := NewSearchHistory(50)
	dialog := NewQuerySearchDialog(history)

	view := dialog.View()

	// Check title
	if !strings.Contains(view, "Query Filter") {
		t.Error("view should contain title 'Query Filter'")
	}

	// Check hints (2 lines)
	if !strings.Contains(view, "size") || !strings.Contains(view, "ext") {
		t.Error("view should contain syntax hints for size and ext")
	}
	if !strings.Contains(view, "LIKE") {
		t.Error("view should contain syntax hint for LIKE")
	}

	// Check footer
	if !strings.Contains(view, "Enter") || !strings.Contains(view, "Esc") {
		t.Error("view should contain footer with key hints")
	}
}

func TestQuerySearchDialog_EnterValidQuery(t *testing.T) {
	history := NewSearchHistory(50)
	dialog := NewQuerySearchDialog(history)

	// Type a valid query
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("size > 1MB")})

	// Press Enter
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Enter with valid query should return a command")
	}

	// Execute the command and check the message
	msg := cmd()
	resultMsg, ok := msg.(querySearchResultMsg)
	if !ok {
		t.Fatalf("command returned %T, want querySearchResultMsg", msg)
	}

	if resultMsg.cancelled {
		t.Error("result should not be cancelled")
	}

	if resultMsg.query != "size > 1MB" {
		t.Errorf("query = %q, want %q", resultMsg.query, "size > 1MB")
	}

	// History should be updated
	if len(history.patterns) != 1 || history.patterns[0] != "size > 1MB" {
		t.Error("query should be added to history")
	}

	if dialog.IsActive() {
		t.Error("dialog should be closed after Enter")
	}
}

func TestQuerySearchDialog_EnterInvalidQuery(t *testing.T) {
	history := NewSearchHistory(50)
	dialog := NewQuerySearchDialog(history)

	// Type an invalid query
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("invalid query syntax !!!")})

	// Press Enter
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("Enter with invalid query should not return a command")
	}

	if !dialog.IsActive() {
		t.Error("dialog should stay open on invalid query")
	}
}

func TestQuerySearchDialog_EnterEmptyInput(t *testing.T) {
	history := NewSearchHistory(50)
	dialog := NewQuerySearchDialog(history)

	// Press Enter with empty input (clears filter)
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Enter with empty input should return a command")
	}

	msg := cmd()
	resultMsg, ok := msg.(querySearchResultMsg)
	if !ok {
		t.Fatalf("command returned %T, want querySearchResultMsg", msg)
	}

	if resultMsg.cancelled {
		t.Error("result should not be cancelled")
	}

	if resultMsg.query != "" {
		t.Errorf("query = %q, want empty string", resultMsg.query)
	}

	// History should NOT be updated for empty input
	if len(history.patterns) != 0 {
		t.Error("empty query should not be added to history")
	}
}

func TestQuerySearchDialog_Escape(t *testing.T) {
	history := NewSearchHistory(50)
	dialog := NewQuerySearchDialog(history)

	// Type some text
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("some query")})

	// Press Escape
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("Escape should return a command")
	}

	msg := cmd()
	resultMsg, ok := msg.(querySearchResultMsg)
	if !ok {
		t.Fatalf("command returned %T, want querySearchResultMsg", msg)
	}

	if !resultMsg.cancelled {
		t.Error("result should be cancelled")
	}

	// History should NOT be updated on cancel
	if len(history.patterns) != 0 {
		t.Error("query should not be added to history on cancel")
	}

	if dialog.IsActive() {
		t.Error("dialog should be closed after Escape")
	}
}

func TestQuerySearchDialog_HistoryNavigation(t *testing.T) {
	history := NewSearchHistory(50)
	history.Add("ext = '.txt'")
	history.Add("size > 100KB")
	history.Add("name LIKE 'test%'")

	dialog := NewQuerySearchDialog(history)

	// Type some text
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("current")})

	// Navigate up
	dialog.Update(tea.KeyMsg{Type: tea.KeyUp})

	// The textInput should now show newest entry
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter should return a command")
	}

	msg := cmd()
	resultMsg, ok := msg.(querySearchResultMsg)
	if !ok {
		t.Fatalf("command returned %T, want querySearchResultMsg", msg)
	}

	if resultMsg.query != "name LIKE 'test%'" {
		t.Errorf("query = %q, want %q (from history)", resultMsg.query, "name LIKE 'test%'")
	}
}

func TestQuerySearchDialog_HistoryNavigateDown(t *testing.T) {
	history := NewSearchHistory(50)
	history.Add("size > 10KB")
	history.Add("size < 1MB")

	dialog := NewQuerySearchDialog(history)

	// Type some text
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("typed")})

	// Navigate up twice
	dialog.Update(tea.KeyMsg{Type: tea.KeyUp})
	dialog.Update(tea.KeyMsg{Type: tea.KeyUp})

	// Navigate down to return to newer entry
	dialog.Update(tea.KeyMsg{Type: tea.KeyDown})

	// Press Enter
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter should return a command")
	}

	msg := cmd()
	resultMsg := msg.(querySearchResultMsg)

	if resultMsg.query != "size < 1MB" {
		t.Errorf("query = %q, want %q", resultMsg.query, "size < 1MB")
	}
}

func TestQuerySearchDialog_TextInput(t *testing.T) {
	history := NewSearchHistory(50)
	dialog := NewQuerySearchDialog(history)

	// Type valid query with ext field
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ext = '.go'")})

	// Verify by pressing Enter
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msg := cmd()
	resultMsg := msg.(querySearchResultMsg)

	if resultMsg.query != "ext = '.go'" {
		t.Errorf("query = %q, want %q", resultMsg.query, "ext = '.go'")
	}
}

func TestQuerySearchDialog_InactiveView(t *testing.T) {
	history := NewSearchHistory(50)
	dialog := NewQuerySearchDialog(history)

	// Close the dialog
	dialog.Close()

	view := dialog.View()
	if view != "" {
		t.Errorf("inactive dialog View() = %q, want empty string", view)
	}
}

func TestQuerySearchDialog_ClearErrorOnKeyPress(t *testing.T) {
	history := NewSearchHistory(50)
	dialog := NewQuerySearchDialog(history)

	// Type invalid query
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("invalid!@#syntax")})

	// Press Enter to trigger error
	dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Type another character - error should clear
	dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
}

func TestQuerySearchDialog_ResetsHistoryOnCreate(t *testing.T) {
	history := NewSearchHistory(50)
	history.Add("a")
	history.Add("b")

	// Simulate previous navigation
	history.NavigateUp("old input")
	history.NavigateUp("old input")

	// Create new dialog - should reset history navigation
	dialog := NewQuerySearchDialog(history)
	_ = dialog

	// History patterns should remain but navigation state should be reset
	if history.index != -1 {
		t.Errorf("history.index = %d, want -1 (reset)", history.index)
	}
}

func TestQuerySearchDialog_ValidQueries(t *testing.T) {
	validQueries := []string{
		"size > 1MB",
		"size < 100KB",
		"ext = '.go'",
		"name LIKE 'test%'",
		"size >= 0",
		"isdir",
		"NOT isdir",
		"isfile",
	}

	for _, query := range validQueries {
		t.Run(query, func(t *testing.T) {
			history := NewSearchHistory(50)
			dialog := NewQuerySearchDialog(history)

			// Type the query
			dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(query)})

			// Press Enter
			_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

			if cmd == nil {
				t.Errorf("query %q should be valid", query)
			}
		})
	}
}
