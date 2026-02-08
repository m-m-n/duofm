package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/config"
)

// Model basic tests: initialization, view, settings

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
	t.Run("Init returns nil (ticker starts from handleWindowSize)", func(t *testing.T) {
		model := NewModel() // default refreshRate=3
		cmd := model.Init()
		// Auto-refresh ticker is started from handleWindowSize, not Init
		if cmd != nil {
			t.Error("Init() should return nil (ticker starts from handleWindowSize)")
		}
	})

	t.Run("Init returns nil when refreshRate is 0", func(t *testing.T) {
		model := NewModelWithConfig(nil, nil, nil, 0, 0, config.DefaultEnterBehavior(), config.MIMEBehaviorConfig{}, "")
		cmd := model.Init()
		if cmd != nil {
			t.Error("Init() should return nil when refreshRate=0")
		}
	})
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

func TestModelInitWithWarnings(t *testing.T) {
	t.Run("calls Init without panic", func(t *testing.T) {
		model := NewModelWithConfig(nil, nil, []string{"Warning: test"}, 0, config.DefaultRefreshRate, config.DefaultEnterBehavior(), config.MIMEBehaviorConfig{}, "")
		cmd := model.Init()
		// Init returns nil (ticker starts from handleWindowSize, not Init)
		if cmd != nil {
			t.Error("Init should return nil (ticker starts from handleWindowSize)")
		}
	})

	t.Run("configWarnings are stored", func(t *testing.T) {
		model := NewModelWithConfig(nil, nil, []string{"Warning: test"}, 0, config.DefaultRefreshRate, config.DefaultEnterBehavior(), config.MIMEBehaviorConfig{}, "")
		if len(model.configWarnings) != 1 {
			t.Errorf("Expected 1 warning, got %d", len(model.configWarnings))
		}
	})
}
