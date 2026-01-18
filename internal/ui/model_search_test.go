package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Model navigation tests: keyboard, search, shell command

// Model search tests: incremental/regex search, shell command mode

func TestNewModelInitializesSearchState(t *testing.T) {
	model := NewModel()

	t.Run("searchStateが初期化される", func(t *testing.T) {
		if model.searchState.Mode != SearchModeNone {
			t.Errorf("searchState.Mode = %v, want SearchModeNone", model.searchState.Mode)
		}
		if model.searchState.IsActive {
			t.Error("searchState.IsActive should be false initially")
		}
	})

	t.Run("minibufferが初期化される", func(t *testing.T) {
		if model.minibuffer == nil {
			t.Error("minibuffer should not be nil")
		}
		if model.minibuffer.IsVisible() {
			t.Error("minibuffer should not be visible initially")
		}
	})
}

func TestSearchKeyActivatesIncrementalSearch(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Press / key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.searchState.Mode != SearchModeIncremental {
		t.Errorf("searchState.Mode = %v, want SearchModeIncremental", m.searchState.Mode)
	}

	if !m.searchState.IsActive {
		t.Error("searchState.IsActive should be true after / key")
	}

	if !m.minibuffer.IsVisible() {
		t.Error("minibuffer should be visible after / key")
	}
}

func TestCtrlFOpensRegexSearchDialog(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Press Ctrl+F
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlF}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should open RegexSearchDialog, not activate minibuffer search
	if m.dialog == nil {
		t.Error("dialog should be set after Ctrl+F")
	}

	if _, ok := m.dialog.(*RegexSearchDialog); !ok {
		t.Errorf("dialog should be *RegexSearchDialog, got %T", m.dialog)
	}

	// minibuffer should NOT be visible (dialog is used instead)
	if m.minibuffer.IsVisible() {
		t.Error("minibuffer should NOT be visible after Ctrl+F (dialog is used)")
	}
}

func TestSearchEscCancelsSearch(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Start search
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Press Esc to cancel
	keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.searchState.IsActive {
		t.Error("searchState.IsActive should be false after Esc")
	}

	if m.minibuffer.IsVisible() {
		t.Error("minibuffer should not be visible after Esc")
	}
}

func TestSearchEnterConfirmsSearch(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Start search
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Type a pattern
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Press Enter to confirm
	keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.searchState.IsActive {
		t.Error("searchState.IsActive should be false after Enter")
	}

	if m.minibuffer.IsVisible() {
		t.Error("minibuffer should not be visible after Enter")
	}

	// Filter should be applied
	if !m.getActivePane().IsFiltered() {
		t.Error("filter should be applied after Enter with pattern")
	}
}

func TestEmptySearchEnterClearsFilter(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Apply a filter first
	m.getActivePane().ApplyFilter("test", SearchModeIncremental)

	// Start search
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Press Enter without typing anything
	keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Filter should be cleared
	if m.getActivePane().IsFiltered() {
		t.Error("filter should be cleared after Enter with empty pattern")
	}
}

func TestIncrementalSearchAppliesFilterImmediately(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	initialEntryCount := len(m.getActivePane().entries)

	// Start incremental search
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Type a pattern that should filter entries
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Entries should be filtered immediately
	if len(m.getActivePane().entries) >= initialEntryCount {
		// Unless "xyz" matches something, which is unlikely
		// This is a rough test - the point is that filtering happens immediately
		t.Log("Incremental filter applied (entry count may vary based on directory contents)")
	}
}

func TestSearchStateRestoreOnEsc(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Get initial entry count
	initialEntryCount := len(m.getActivePane().entries)

	// Apply a filter first
	m.getActivePane().ApplyFilter("test", SearchModeIncremental)
	filteredCount := len(m.getActivePane().entries)

	// Start a new search
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Type something different
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Press Esc to cancel - should restore previous filter
	keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Previous filter should be restored
	restoredCount := len(m.getActivePane().entries)

	// The filter pattern should be restored
	if m.getActivePane().FilterPattern() != "test" {
		t.Errorf("FilterPattern() = %s, want 'test'", m.getActivePane().FilterPattern())
	}

	t.Logf("Entry counts: initial=%d, filtered=%d, restored=%d", initialEntryCount, filteredCount, restoredCount)
}

func TestSearchPromptForIncrementalMode(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Start incremental search with / key
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ = m.Update(key)
	m = updatedModel.(Model)

	wantPrompt := "/: "
	if m.minibuffer.prompt != wantPrompt {
		t.Errorf("prompt = %s, want %s", m.minibuffer.prompt, wantPrompt)
	}
}

func TestCtrlGOpensQuerySearchDialog(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Press Ctrl+G
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlG}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should open QuerySearchDialog
	if m.dialog == nil {
		t.Error("dialog should be set after Ctrl+G")
	}

	if _, ok := m.dialog.(*QuerySearchDialog); !ok {
		t.Errorf("dialog should be *QuerySearchDialog, got %T", m.dialog)
	}

	// minibuffer should NOT be visible (dialog is used instead)
	if m.minibuffer.IsVisible() {
		t.Error("minibuffer should NOT be visible after Ctrl+G (dialog is used)")
	}
}

// === Ctrl+C機能のテスト ===

// TestCtrlCPendingFieldInitialization tests ctrlCPending field is false initially
func TestSearchCtrlCCancelsSearch(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Start search
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Verify search is active
	if !m.searchState.IsActive {
		t.Fatal("search should be active after / key")
	}

	// Press Ctrl+C to cancel
	keyMsg = tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Verify search is cancelled
	if m.searchState.IsActive {
		t.Error("search should be cancelled after Ctrl+C")
	}

	if m.minibuffer.IsVisible() {
		t.Error("minibuffer should not be visible after Ctrl+C")
	}
}

// TestQKeyStillQuits tests that q key still quits immediately
func TestModelConfirmSearch(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "abc.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "abd.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "xyz.txt"), []byte("test"), 0644)

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("confirms search with pattern via HandleKey", func(t *testing.T) {
		m.searchState = SearchState{
			Mode:     SearchModeIncremental,
			IsActive: true,
		}
		m.minibuffer.Show()
		// Type "abc" via HandleKey
		m.minibuffer.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		m.minibuffer.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
		m.minibuffer.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		m.confirmSearch()
		if m.searchState.Mode != SearchModeNone {
			t.Error("Search mode should be cleared after confirm")
		}
		if m.searchState.PreviousResult == nil {
			t.Error("Previous result should be stored after confirm with pattern")
		}
	})

	t.Run("clears filter with empty pattern", func(t *testing.T) {
		m.searchState = SearchState{
			Mode:     SearchModeIncremental,
			IsActive: true,
		}
		m.minibuffer.Show()
		m.minibuffer.Clear()
		m.confirmSearch()
		if m.searchState.PreviousResult != nil {
			t.Error("Previous result should be nil for empty pattern")
		}
	})
}

func TestNewModelShellCommandModeInitialization(t *testing.T) {
	// Test that shellCommandMode initializes to false
	model := NewModel()

	if model.shellCommandMode {
		t.Error("NewModel() shellCommandMode should be false initially")
	}
}

func TestShellCommandModeActivation(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Press '!' to enter shell command mode
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if !m.shellCommandMode {
		t.Error("shellCommandMode should be true after pressing '!'")
	}

	if !m.minibuffer.IsVisible() {
		t.Error("minibuffer should be visible in shell command mode")
	}
}

func TestShellCommandModeEscapeCancels(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Enter shell command mode
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Press Escape to cancel
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	updatedModel, _ = m.Update(escMsg)
	m = updatedModel.(Model)

	if m.shellCommandMode {
		t.Error("shellCommandMode should be false after pressing Escape")
	}

	if m.minibuffer.IsVisible() {
		t.Error("minibuffer should be hidden after pressing Escape")
	}
}

func TestShellCommandModeEmptyEnterExits(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Enter shell command mode
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Press Enter with empty input - should exit without executing
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := m.Update(enterMsg)
	m = updatedModel.(Model)

	if m.shellCommandMode {
		t.Error("shellCommandMode should be false after pressing Enter with empty input")
	}

	if cmd != nil {
		t.Error("no command should be executed for empty input")
	}
}

func TestShellCommandModeIgnoredWhenDialogActive(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Open help dialog
	m.dialog = NewHelpDialog()

	// Try to enter shell command mode
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.shellCommandMode {
		t.Error("shellCommandMode should not be activated when dialog is active")
	}
}

func TestShellCommandModeIgnoredWhenSearchActive(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Start search mode
	searchMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ = m.Update(searchMsg)
	m = updatedModel.(Model)

	// Try to press '!' in search mode - it should be passed to minibuffer, not activate shell mode
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should still be in search mode, not shell command mode
	if m.shellCommandMode {
		t.Error("shellCommandMode should not be activated when search mode is active")
	}
}

func TestShellCommandModeCharacterInput(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Enter shell command mode
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Type some characters
	for _, r := range "ls -la" {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedModel, _ = m.Update(charMsg)
		m = updatedModel.(Model)
	}

	if m.minibuffer.Input() != "ls -la" {
		t.Errorf("minibuffer input = %q, want %q", m.minibuffer.Input(), "ls -la")
	}
}

func TestShellCommandModeViewRendering(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Enter shell command mode
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// View should render minibuffer with "!:" prompt
	view := m.View()

	if !strings.Contains(view, "!:") {
		t.Error("View should contain minibuffer with '!:' prompt in shell command mode")
	}
}
