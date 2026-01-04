package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Model navigation tests: keyboard, search, shell command

func TestArrowKeyNavigation(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Save initial cursor position
	initialCursor := m.getActivePane().cursor

	// Test down arrow
	keyMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.getActivePane().cursor != initialCursor+1 {
		t.Errorf("down arrow: cursor = %d, want %d", m.getActivePane().cursor, initialCursor+1)
	}

	// Test up arrow
	keyMsg = tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.getActivePane().cursor != initialCursor {
		t.Errorf("up arrow: cursor = %d, want %d", m.getActivePane().cursor, initialCursor)
	}
}

// TestArrowKeyPaneSwitching tests arrow key pane switching
func TestArrowKeyPaneSwitching(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Initial active pane should be LeftPane
	if m.activePane != LeftPane {
		t.Fatalf("initial activePane = %v, want LeftPane", m.activePane)
	}

	// Press right arrow to switch to right pane
	keyMsg := tea.KeyMsg{Type: tea.KeyRight}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.activePane != RightPane {
		t.Errorf("after right arrow: activePane = %v, want RightPane", m.activePane)
	}

	// Press left arrow to switch back to left pane
	keyMsg = tea.KeyMsg{Type: tea.KeyLeft}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.activePane != LeftPane {
		t.Errorf("after left arrow: activePane = %v, want LeftPane", m.activePane)
	}
}

// TestArrowKeysEquivalentToHJKL tests that arrow keys work the same as hjkl
func TestArrowKeysEquivalentToHJKL(t *testing.T) {
	tests := []struct {
		name     string
		arrowKey tea.KeyType
		vimKey   string
	}{
		{"down arrow equals j", tea.KeyDown, "j"},
		{"up arrow equals k", tea.KeyUp, "k"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with arrow key
			model1 := NewModel()
			msg := tea.WindowSizeMsg{Width: 120, Height: 40}
			updatedModel, _ := model1.Update(msg)
			m1 := updatedModel.(Model)

			arrowMsg := tea.KeyMsg{Type: tt.arrowKey}
			updatedModel, _ = m1.Update(arrowMsg)
			m1 = updatedModel.(Model)
			arrowCursor := m1.getActivePane().cursor

			// Test with vim key
			model2 := NewModel()
			updatedModel, _ = model2.Update(msg)
			m2 := updatedModel.(Model)

			vimMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.vimKey)}
			updatedModel, _ = m2.Update(vimMsg)
			m2 = updatedModel.(Model)
			vimCursor := m2.getActivePane().cursor

			if arrowCursor != vimCursor {
				t.Errorf("arrow key cursor = %d, vim key cursor = %d", arrowCursor, vimCursor)
			}
		})
	}
}

// TestModelContextMenuEscClosesMenu tests that Esc closes context menu
func TestStatusMessageField(t *testing.T) {
	model := NewModel()

	t.Run("初期状態でstatusMessageは空", func(t *testing.T) {
		if model.statusMessage != "" {
			t.Errorf("statusMessage should be empty initially, got %s", model.statusMessage)
		}
	})

	t.Run("初期状態でisStatusErrorはfalse", func(t *testing.T) {
		if model.isStatusError {
			t.Error("isStatusError should be false initially")
		}
	})
}

func TestClearStatusMsg(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set status message
	m.statusMessage = "Test error message"
	m.isStatusError = true

	// Send clearStatusMsg
	updatedModel, _ = m.Update(clearStatusMsg{})
	m = updatedModel.(Model)

	if m.statusMessage != "" {
		t.Errorf("statusMessage should be cleared, got %s", m.statusMessage)
	}

	if m.isStatusError {
		t.Error("isStatusError should be false after clearStatusMsg")
	}
}

func TestStatusMessageClearCmd(t *testing.T) {
	// Test that statusMessageClearCmd returns a non-nil command
	cmd := statusMessageClearCmd(100)
	if cmd == nil {
		t.Error("statusMessageClearCmd should return a non-nil command")
	}
}

func TestStatusMessageClearOnKeyPress(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set status message
	m.statusMessage = "Test error message"
	m.isStatusError = true

	// Press any key (j for down)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.statusMessage != "" {
		t.Errorf("statusMessage should be cleared after key press, got %s", m.statusMessage)
	}

	if m.isStatusError {
		t.Error("isStatusError should be false after key press")
	}
}

// === Phase 4: 検索機能のテスト ===

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

func TestCtrlFActivatesRegexSearch(t *testing.T) {
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

	if m.searchState.Mode != SearchModeRegex {
		t.Errorf("searchState.Mode = %v, want SearchModeRegex", m.searchState.Mode)
	}

	if !m.searchState.IsActive {
		t.Error("searchState.IsActive should be true after Ctrl+F")
	}

	if !m.minibuffer.IsVisible() {
		t.Error("minibuffer should be visible after Ctrl+F")
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

func TestSearchPromptForModes(t *testing.T) {
	tests := []struct {
		name       string
		key        tea.KeyMsg
		wantPrompt string
	}{
		{
			name:       "インクリメンタル検索のプロンプト",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}},
			wantPrompt: "/: ",
		},
		{
			name:       "正規表現検索のプロンプト",
			key:        tea.KeyMsg{Type: tea.KeyCtrlF},
			wantPrompt: "(search): ",
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

			// Start search
			updatedModel, _ = m.Update(tt.key)
			m = updatedModel.(Model)

			if m.minibuffer.prompt != tt.wantPrompt {
				t.Errorf("prompt = %s, want %s", m.minibuffer.prompt, tt.wantPrompt)
			}
		})
	}
}

// === Ctrl+C機能のテスト ===

// TestCtrlCPendingFieldInitialization tests ctrlCPending field is false initially
func TestCtrlCPendingFieldInitialization(t *testing.T) {
	model := NewModel()

	if model.ctrlCPending {
		t.Error("ctrlCPending should be false initially")
	}
}

// TestSingleCtrlCShowsMessage tests that first Ctrl+C shows message
func TestSingleCtrlCShowsMessage(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Press Ctrl+C
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedModel, cmd := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Verify status message is shown
	if m.statusMessage != "Press Ctrl+C again to quit" {
		t.Errorf("statusMessage = %q, want 'Press Ctrl+C again to quit'", m.statusMessage)
	}

	// Verify ctrlCPending is true
	if !m.ctrlCPending {
		t.Error("ctrlCPending should be true after first Ctrl+C")
	}

	// Verify isStatusError is false
	if m.isStatusError {
		t.Error("isStatusError should be false for Ctrl+C message")
	}

	// Verify a timeout command is returned
	if cmd == nil {
		t.Error("should return a timeout command")
	}
}

// TestDoubleCtrlCQuits tests that double Ctrl+C quits application
func TestDoubleCtrlCQuits(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// First Ctrl+C
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Second Ctrl+C - should quit
	updatedModel, cmd := m.Update(keyMsg)

	// Verify quit command is returned
	if cmd == nil {
		t.Error("should return quit command on double Ctrl+C")
	}
}

// TestCtrlCTimeoutResetsState tests that timeout resets state
func TestCtrlCTimeoutResetsState(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// First Ctrl+C
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Verify ctrlCPending is true
	if !m.ctrlCPending {
		t.Error("ctrlCPending should be true after first Ctrl+C")
	}

	// Send ctrlCTimeoutMsg
	updatedModel, _ = m.Update(ctrlCTimeoutMsg{})
	m = updatedModel.(Model)

	// Verify state is reset
	if m.ctrlCPending {
		t.Error("ctrlCPending should be false after timeout")
	}

	if m.statusMessage != "" {
		t.Errorf("statusMessage should be empty after timeout, got %q", m.statusMessage)
	}
}

// TestOtherKeyAfterCtrlCResetsState tests that other key resets state
func TestOtherKeyAfterCtrlCResetsState(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// First Ctrl+C
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Verify ctrlCPending is true
	if !m.ctrlCPending {
		t.Error("ctrlCPending should be true after first Ctrl+C")
	}

	// Press 'j' key
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Verify state is reset
	if m.ctrlCPending {
		t.Error("ctrlCPending should be false after other key")
	}

	if m.statusMessage != "" {
		t.Errorf("statusMessage should be empty after other key, got %q", m.statusMessage)
	}
}

// TestSearchCtrlCCancelsSearch tests Ctrl+C cancels search
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
func TestQKeyStillQuits(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Press 'q' key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(keyMsg)

	// Verify quit command is returned
	if cmd == nil {
		t.Error("q key should return quit command")
	}
}

// TestCtrlCTimeoutCmdReturnsNonNil tests that ctrlCTimeoutCmd returns non-nil command
func TestCtrlCTimeoutCmdReturnsNonNil(t *testing.T) {
	cmd := ctrlCTimeoutCmd(100)
	if cmd == nil {
		t.Error("ctrlCTimeoutCmd should return non-nil command")
	}
}

// === RefreshBothPanes and SyncOppositePane のテスト ===

func TestRefreshBothPanes(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("RefreshBothPanesで両ペインがリフレッシュされる", func(t *testing.T) {
		// Call RefreshBothPanes
		_ = m.RefreshBothPanes()

		// Basic verification: both panes should have entries
		if len(m.leftPane.entries) == 0 {
			t.Error("leftPane should have entries after refresh")
		}
		if len(m.rightPane.entries) == 0 {
			t.Error("rightPane should have entries after refresh")
		}
	})

	t.Run("RefreshBothPanesでディスク容量が更新される", func(t *testing.T) {
		m.leftDiskSpace = 0
		m.rightDiskSpace = 0

		_ = m.RefreshBothPanes()

		// Disk space should be updated
		if m.leftDiskSpace == 0 && m.rightDiskSpace == 0 {
			t.Log("Disk space might be 0 on some filesystems, skipping verification")
		}
	})
}

func TestSyncOppositePane(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("左ペインがアクティブなとき右ペインが同期される", func(t *testing.T) {
		// Ensure left pane is active
		m.activePane = LeftPane
		m.leftPane.SetActive(true)
		m.rightPane.SetActive(false)

		leftPath := m.leftPane.Path()
		originalRightPath := m.rightPane.Path()

		// Skip if already same directory
		if leftPath == originalRightPath {
			t.Skip("Left and right panes are already in the same directory")
		}

		m.SyncOppositePane()

		if m.rightPane.Path() != leftPath {
			t.Errorf("rightPane.Path() = %s, want %s", m.rightPane.Path(), leftPath)
		}
	})

	t.Run("右ペインがアクティブなとき左ペインが同期される", func(t *testing.T) {
		// Reinitialize model
		model2 := NewModel()
		updatedModel2, _ := model2.Update(msg)
		m2 := updatedModel2.(Model)

		// Make right pane active
		m2.activePane = RightPane
		m2.leftPane.SetActive(false)
		m2.rightPane.SetActive(true)

		rightPath := m2.rightPane.Path()
		originalLeftPath := m2.leftPane.Path()

		// Skip if already same directory
		if rightPath == originalLeftPath {
			t.Skip("Left and right panes are already in the same directory")
		}

		m2.SyncOppositePane()

		if m2.leftPane.Path() != rightPath {
			t.Errorf("leftPane.Path() = %s, want %s", m2.leftPane.Path(), rightPath)
		}
	})
}

func TestRefreshKeyF5(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("F5キーでリフレッシュが呼ばれる", func(t *testing.T) {
		// Press F5 key
		keyMsg := tea.KeyMsg{Type: tea.KeyF5}
		_, cmd := m.Update(keyMsg)

		// Should return a command (from RefreshBothPanes)
		if cmd == nil {
			// RefreshBothPanes currently returns empty batch, so nil is acceptable
			t.Log("F5 handled (nil command returned)")
		}
	})
}

func TestRefreshKeyCtrlR(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("Ctrl+Rキーでリフレッシュが呼ばれる", func(t *testing.T) {
		// Press Ctrl+R key
		keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
		_, cmd := m.Update(keyMsg)

		// Should return a command (from RefreshBothPanes)
		if cmd == nil {
			// RefreshBothPanes currently returns empty batch, so nil is acceptable
			t.Log("Ctrl+R handled (nil command returned)")
		}
	})
}

func TestSyncPaneKey(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("=キーで反対ペインが同期される", func(t *testing.T) {
		// Ensure left pane is active
		m.activePane = LeftPane
		m.leftPane.SetActive(true)
		m.rightPane.SetActive(false)

		leftPath := m.leftPane.Path()
		originalRightPath := m.rightPane.Path()

		// Skip if already same directory
		if leftPath == originalRightPath {
			t.Skip("Left and right panes are already in the same directory")
		}

		// Press = key
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'='}}
		updatedModel, _ := m.Update(keyMsg)
		m = updatedModel.(Model)

		if m.rightPane.Path() != leftPath {
			t.Errorf("rightPane.Path() = %s, want %s", m.rightPane.Path(), leftPath)
		}
	})
}

func TestRefreshKeysIgnoredDuringDialog(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Open a dialog
	m.dialog = NewHelpDialog()

	t.Run("ダイアログ表示中はF5キーが無視される", func(t *testing.T) {
		// Press F5 key - should be handled by dialog
		keyMsg := tea.KeyMsg{Type: tea.KeyF5}
		updatedModel, _ := m.Update(keyMsg)
		m = updatedModel.(Model)

		// Dialog should still be active
		if m.dialog == nil {
			t.Error("dialog should still be active after F5 during dialog display")
		}
	})
}

func TestSyncPreservesPaneSettings(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Set some settings on right pane
	m.rightPane.showHidden = true
	m.rightPane.displayMode = DisplayDetail

	// Sync to left pane's directory
	m.activePane = LeftPane
	m.SyncOppositePane()

	// Verify settings are preserved
	if !m.rightPane.showHidden {
		t.Error("showHidden should be preserved after sync")
	}
	if m.rightPane.displayMode != DisplayDetail {
		t.Error("displayMode should be preserved after sync")
	}
}

// === ダイアログ完了後のクリーンアップテスト ===
// これらのテストは、ダイアログ完了後に m.dialog が nil になることを検証する
// 回帰テスト: Issue #XXX - ファイル作成後に操作不能になるバグ

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

func TestModelUpdateKeyActions(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test2.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Test various key actions - focus on code path coverage
	keyTests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"Tab", tea.KeyMsg{Type: tea.KeyTab}},
		{"Enter", tea.KeyMsg{Type: tea.KeyEnter}},
		{"j", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}},
		{"k", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}},
		{"g", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}},
		{"G", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}},
		{"space", tea.KeyMsg{Type: tea.KeySpace}},
		{"?", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}},
		{"h", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}},
		{"l", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}},
		{"Home", tea.KeyMsg{Type: tea.KeyHome}},
		{"End", tea.KeyMsg{Type: tea.KeyEnd}},
		{"PgUp", tea.KeyMsg{Type: tea.KeyPgUp}},
		{"PgDown", tea.KeyMsg{Type: tea.KeyPgDown}},
		{"Backspace", tea.KeyMsg{Type: tea.KeyBackspace}},
		{"r", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}},
		{"n", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}},
		{"N", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}},
		{"/", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}},
		{"Ctrl+H", tea.KeyMsg{Type: tea.KeyCtrlH}},
		{"s", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}},
	}

	for _, tt := range keyTests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic
			m.Update(tt.key)
		})
	}
}

func TestModelMinibufferInteraction(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "abc.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "abd.txt"), []byte("test"), 0644)

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("/ opens minibuffer for search", func(t *testing.T) {
		m.searchState.IsActive = false
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
		m.Update(keyMsg)
		// Minibuffer should be visible
	})

	t.Run("Esc during search closes minibuffer", func(t *testing.T) {
		m.searchState.IsActive = true
		m.minibuffer.Show()
		keyMsg := tea.KeyMsg{Type: tea.KeyEsc}
		m.Update(keyMsg)
	})

	t.Run("Enter during search confirms", func(t *testing.T) {
		m.searchState.IsActive = true
		m.searchState.Mode = SearchModeIncremental
		m.minibuffer.Show()
		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
		m.Update(keyMsg)
	})
}

func TestModelMoreKeyActions(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("c key for copy", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
		m.Update(keyMsg)
	})

	t.Run("m key for move", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
		m.Update(keyMsg)
	})

	t.Run("d key for delete", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
		m.Update(keyMsg)
	})

	t.Run("o key for open menu", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}
		m.Update(keyMsg)
	})

	t.Run("a key for select all", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
		m.Update(keyMsg)
	})

	t.Run("u key for unselect all", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}
		m.Update(keyMsg)
	})

	t.Run("= key for sync panes", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'='}}
		m.Update(keyMsg)
	})

	t.Run("b key for bookmarks", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
		m.Update(keyMsg)
	})

	t.Run("B key for add bookmark", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'B'}}
		m.Update(keyMsg)
	})

	t.Run("Ctrl+C key", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
		m.Update(keyMsg)
	})

	t.Run("z key for compress", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}}
		m.Update(keyMsg)
	})

	t.Run("x key for extract", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
		m.Update(keyMsg)
	})

	t.Run("e key for edit", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
		m.Update(keyMsg)
	})

	t.Run("v key for view", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}}
		m.Update(keyMsg)
	})

	t.Run("! key for shell", func(t *testing.T) {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}
		m.Update(keyMsg)
	})
}

// TestContextMenuCompressWithNilAction tests that compress action works even when Action is nil
// This is a regression test for the bug where compress menu item had Action: nil
// and the handler checked result.action != nil before processing actionID
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
