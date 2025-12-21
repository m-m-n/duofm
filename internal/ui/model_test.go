package ui

import (
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
