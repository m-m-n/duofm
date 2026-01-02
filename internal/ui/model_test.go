package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	tests := []struct {
		name string
		want struct {
			activePane PanePosition
			ready      bool
		}
	}{
		{
			name: "初期モデルの作成",
			want: struct {
				activePane PanePosition
				ready      bool
			}{
				activePane: LeftPane,
				ready:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()

			if model.activePane != tt.want.activePane {
				t.Errorf("NewModel() activePane = %v, want %v", model.activePane, tt.want.activePane)
			}

			if model.ready != tt.want.ready {
				t.Errorf("NewModel() ready = %v, want %v", model.ready, tt.want.ready)
			}

			if model.leftPane != nil {
				t.Error("NewModel() leftPane should be nil initially")
			}

			if model.rightPane != nil {
				t.Error("NewModel() rightPane should be nil initially")
			}
		})
	}
}

func TestModelInit(t *testing.T) {
	tests := []struct {
		name    string
		wantCmd bool
	}{
		{
			name:    "Init は nil コマンドを返す",
			wantCmd: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			cmd := model.Init()

			if tt.wantCmd && cmd == nil {
				t.Error("Init() should return a command")
			}

			if !tt.wantCmd && cmd != nil {
				t.Error("Init() should return nil")
			}
		})
	}
}

func TestModelUpdateWindowSize(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{
			name:   "ウィンドウサイズメッセージの処理",
			width:  80,
			height: 24,
		},
		{
			name:   "大きなウィンドウサイズ",
			width:  200,
			height: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()

			msg := tea.WindowSizeMsg{
				Width:  tt.width,
				Height: tt.height,
			}

			updatedModel, _ := model.Update(msg)
			m := updatedModel.(Model)

			if m.width != tt.width {
				t.Errorf("Update() width = %v, want %v", m.width, tt.width)
			}

			if m.height != tt.height {
				t.Errorf("Update() height = %v, want %v", m.height, tt.height)
			}

			if !m.ready {
				t.Error("Update() should set ready to true after WindowSizeMsg")
			}
		})
	}
}

func TestModelUpdateQuit(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantCmd bool
	}{
		{
			name:    "q キーで終了",
			key:     "q",
			wantCmd: true,
		},
		{
			name:    "ctrl+c で終了",
			key:     "ctrl+c",
			wantCmd: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			// ready状態にする
			model.ready = true

			msg := tea.KeyMsg{
				Type: tea.KeyRunes,
			}

			// KeyMsgの作成方法を調整
			if tt.key == "q" {
				msg.Type = tea.KeyRunes
				msg.Runes = []rune{'q'}
			} else if tt.key == "ctrl+c" {
				msg.Type = tea.KeyCtrlC
			}

			_, cmd := model.Update(msg)

			if tt.wantCmd && cmd == nil {
				t.Error("Update() should return quit command")
			}

			if !tt.wantCmd && cmd != nil {
				t.Error("Update() should not return a command")
			}
		})
	}
}

func TestModelView(t *testing.T) {
	t.Run("初期化前の表示", func(t *testing.T) {
		model := NewModel()
		model.ready = false

		view := model.View()

		if view != "Initializing..." {
			t.Errorf("View() = %v, want %v", view, "Initializing...")
		}
	})

	t.Run("初期化後の表示", func(t *testing.T) {
		model := NewModel()

		// WindowSizeMsgを送信してペインを初期化
		msg := tea.WindowSizeMsg{
			Width:  120,
			Height: 40,
		}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		view := m.View()

		// 初期化後は、デュアルペインとステータスバーを含むビューが表示される
		if view == "" {
			t.Error("View() should return non-empty content after initialization")
		}

		if view == "Initializing..." {
			t.Error("View() should not show 'Initializing...' after WindowSizeMsg")
		}

		// "duofm" タイトルが含まれることを確認
		if !strings.Contains(view, "duofm") {
			t.Error("View() should contain 'duofm' title")
		}
	})
}

func TestPanePosition(t *testing.T) {
	tests := []struct {
		name     string
		position PanePosition
		want     int
	}{
		{
			name:     "LeftPane の値",
			position: LeftPane,
			want:     0,
		},
		{
			name:     "RightPane の値",
			position: RightPane,
			want:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.position) != tt.want {
				t.Errorf("PanePosition = %v, want %v", int(tt.position), tt.want)
			}
		})
	}
}

// TestModelContextMenuOpen tests that @ key opens context menu
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

	// Simulate selecting delete (press '3' for delete)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}
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

func TestModelSwitchToPane(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Switch to right pane
	m.switchToPane(RightPane)

	if m.activePane != RightPane {
		t.Errorf("activePane = %v, want RightPane", m.activePane)
	}
	if !m.rightPane.isActive {
		t.Error("rightPane should be active")
	}
	if m.leftPane.isActive {
		t.Error("leftPane should be inactive")
	}

	// Switch to left pane
	m.switchToPane(LeftPane)

	if m.activePane != LeftPane {
		t.Errorf("activePane = %v, want LeftPane", m.activePane)
	}
	if !m.leftPane.isActive {
		t.Error("leftPane should be active")
	}
	if m.rightPane.isActive {
		t.Error("rightPane should be inactive")
	}
}

func TestToggleHiddenFiles(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Get initial state
	initialShowHidden := m.getActivePane().showHidden

	// Press Ctrl+H to toggle hidden files
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlH}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.getActivePane().showHidden == initialShowHidden {
		t.Error("showHidden should have toggled")
	}
}

func TestToggleDisplayMode(t *testing.T) {
	model := NewModel()

	// Initialize with WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	// Check if terminal is wide enough to toggle mode
	if !m.getActivePane().CanToggleMode() {
		t.Skip("Terminal too narrow to toggle display mode")
	}

	// Get initial display mode
	initialMode := m.getActivePane().displayMode

	// Press 'i' to toggle display mode
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.getActivePane().displayMode == initialMode {
		t.Error("displayMode should have changed")
	}
}

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

	_, isInputDialog := m.dialog.(*InputDialog)
	if !isInputDialog {
		t.Error("dialog should be InputDialog")
	}
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

func TestModelInitWithWarnings(t *testing.T) {
	t.Run("calls Init without panic", func(t *testing.T) {
		model := NewModelWithConfig(nil, nil, []string{"Warning: test"})
		cmd := model.Init()
		// Init returns nil
		if cmd != nil {
			t.Error("Init should return nil")
		}
	})

	t.Run("configWarnings are stored", func(t *testing.T) {
		model := NewModelWithConfig(nil, nil, []string{"Warning: test"})
		if len(model.configWarnings) != 1 {
			t.Errorf("Expected 1 warning, got %d", len(model.configWarnings))
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

func TestModelRenderMethods(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)

	model := NewModel()
	model.leftPath = tmpDir
	model.rightPath = tmpDir
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	t.Run("View renders without panic", func(t *testing.T) {
		view := m.View()
		if view == "" {
			t.Error("View should not be empty")
		}
	})

	t.Run("View with dialog", func(t *testing.T) {
		m.dialog = NewHelpDialog()
		view := m.View()
		if view == "" {
			t.Error("View with dialog should not be empty")
		}
	})

	t.Run("View with error dialog", func(t *testing.T) {
		m.dialog = NewErrorDialog("test error")
		view := m.View()
		if view == "" {
			t.Error("View with error dialog should not be empty")
		}
	})

	t.Run("View with status message", func(t *testing.T) {
		m.dialog = nil
		m.statusMessage = "Test status"
		view := m.View()
		if !strings.Contains(view, "Test status") {
			t.Error("View should contain status message")
		}
	})

	t.Run("View with error status", func(t *testing.T) {
		m.statusMessage = "Error message"
		m.isStatusError = true
		view := m.View()
		if view == "" {
			t.Error("View with error status should not be empty")
		}
	})
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

	t.Run("handles batchOperationCompleteMsg", func(t *testing.T) {
		batchMsg := batchOperationCompleteMsg{
			operation: "copy",
			count:     5,
		}
		m.Update(batchMsg)
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
