package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
)

func TestConfirmDialog(t *testing.T) {
	t.Run("確認ダイアログの作成", func(t *testing.T) {
		dialog := NewConfirmDialog("Test Title", "Test message")

		if dialog == nil {
			t.Fatal("NewConfirmDialog() should return non-nil dialog")
		}

		if !dialog.IsActive() {
			t.Error("NewConfirmDialog() should create active dialog")
		}
	})

	t.Run("yキーで確認", func(t *testing.T) {
		dialog := NewConfirmDialog("Test", "Message")

		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
		updatedDialog, cmd := dialog.Update(keyMsg)

		if updatedDialog.IsActive() {
			t.Error("Dialog should be inactive after 'y' key")
		}

		if cmd == nil {
			t.Error("Dialog should return command after 'y' key")
		}

		// コマンドを実行して結果を確認
		if cmd != nil {
			msg := cmd()
			if result, ok := msg.(dialogResultMsg); ok {
				if !result.result.Confirmed {
					t.Error("Result should be confirmed")
				}
				if result.result.Cancelled {
					t.Error("Result should not be cancelled")
				}
			} else {
				t.Error("Command should return dialogResultMsg")
			}
		}
	})

	t.Run("nキーでキャンセル", func(t *testing.T) {
		dialog := NewConfirmDialog("Test", "Message")

		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		updatedDialog, cmd := dialog.Update(keyMsg)

		if updatedDialog.IsActive() {
			t.Error("Dialog should be inactive after 'n' key")
		}

		if cmd == nil {
			t.Error("Dialog should return command after 'n' key")
		}

		// コマンドを実行して結果を確認
		if cmd != nil {
			msg := cmd()
			if result, ok := msg.(dialogResultMsg); ok {
				if result.result.Confirmed {
					t.Error("Result should not be confirmed")
				}
				if !result.result.Cancelled {
					t.Error("Result should be cancelled")
				}
			}
		}
	})

	t.Run("ビューのレンダリング", func(t *testing.T) {
		dialog := NewConfirmDialog("Test Title", "Test Message")

		view := dialog.View()

		if view == "" {
			t.Error("View() should return non-empty string")
		}

		if !strings.Contains(view, "Test Title") {
			t.Error("View() should contain title")
		}

		if !strings.Contains(view, "Test Message") {
			t.Error("View() should contain message")
		}
	})
}

func TestErrorDialog(t *testing.T) {
	t.Run("エラーダイアログの作成", func(t *testing.T) {
		dialog := NewErrorDialog("Error message")

		if dialog == nil {
			t.Fatal("NewErrorDialog() should return non-nil dialog")
		}

		if !dialog.IsActive() {
			t.Error("NewErrorDialog() should create active dialog")
		}
	})

	t.Run("Escキーで閉じる", func(t *testing.T) {
		dialog := NewErrorDialog("Error message")

		keyMsg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedDialog, cmd := dialog.Update(keyMsg)

		if updatedDialog.IsActive() {
			t.Error("Dialog should be inactive after Esc key")
		}

		if cmd == nil {
			t.Error("Dialog should return command after Esc key")
		}
	})

	t.Run("ビューのレンダリング", func(t *testing.T) {
		dialog := NewErrorDialog("Test error message")

		view := dialog.View()

		if view == "" {
			t.Error("View() should return non-empty string")
		}

		if !strings.Contains(view, "Error") {
			t.Error("View() should contain 'Error'")
		}

		if !strings.Contains(view, "Test error message") {
			t.Error("View() should contain error message")
		}
	})
}

func TestHelpDialog(t *testing.T) {
	t.Run("ヘルプダイアログの作成", func(t *testing.T) {
		dialog := NewHelpDialog()

		if dialog == nil {
			t.Fatal("NewHelpDialog() should return non-nil dialog")
		}

		if !dialog.IsActive() {
			t.Error("NewHelpDialog() should create active dialog")
		}
	})

	t.Run("?キーで閉じる", func(t *testing.T) {
		dialog := NewHelpDialog()

		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
		updatedDialog, cmd := dialog.Update(keyMsg)

		if updatedDialog.IsActive() {
			t.Error("Dialog should be inactive after '?' key")
		}

		if cmd != nil {
			msg := cmd()
			if result, ok := msg.(dialogResultMsg); ok {
				if !result.result.Cancelled {
					t.Error("Result should be cancelled")
				}
			}
		}
	})

	t.Run("Escキーで閉じる", func(t *testing.T) {
		dialog := NewHelpDialog()

		keyMsg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedDialog, cmd := dialog.Update(keyMsg)

		if updatedDialog.IsActive() {
			t.Error("Dialog should be inactive after Esc key")
		}

		if cmd == nil {
			t.Error("Dialog should return command after Esc key")
		}
	})

	t.Run("ビューのレンダリング", func(t *testing.T) {
		dialog := NewHelpDialog()

		view := dialog.View()

		if view == "" {
			t.Error("View() should return non-empty string")
		}

		if !strings.Contains(view, "Keybindings") {
			t.Error("View() should contain 'Keybindings'")
		}

		// キーバインディングの説明が含まれているか確認（PascalCase形式）
		if !strings.Contains(view, "J/K") {
			t.Error("View() should contain navigation keys")
		}
	})
}

func TestDialogResult(t *testing.T) {
	t.Run("DialogResultの作成", func(t *testing.T) {
		result := DialogResult{
			Confirmed: true,
			Cancelled: false,
		}

		if !result.Confirmed {
			t.Error("Confirmed should be true")
		}

		if result.Cancelled {
			t.Error("Cancelled should be false")
		}
	})
}

func TestDialogInterface(t *testing.T) {
	t.Run("ConfirmDialogがDialogインターフェースを実装", func(t *testing.T) {
		var _ Dialog = &ConfirmDialog{}
	})

	t.Run("ErrorDialogがDialogインターフェースを実装", func(t *testing.T) {
		var _ Dialog = &ErrorDialog{}
	})

	t.Run("HelpDialogがDialogインターフェースを実装", func(t *testing.T) {
		var _ Dialog = &HelpDialog{}
	})
}

// === Ctrl+Cダイアログキャンセルのテスト ===

func TestConfirmDialogCtrlCCancels(t *testing.T) {
	dialog := NewConfirmDialog("Test", "Message")

	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedDialog, cmd := dialog.Update(keyMsg)

	if updatedDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C")
	}

	if cmd == nil {
		t.Error("Dialog should return command after Ctrl+C")
	}

	// Execute and verify result
	if cmd != nil {
		msg := cmd()
		if result, ok := msg.(dialogResultMsg); ok {
			if result.result.Confirmed {
				t.Error("Result should not be confirmed")
			}
			if !result.result.Cancelled {
				t.Error("Result should be cancelled")
			}
		}
	}
}

func TestErrorDialogCtrlCCloses(t *testing.T) {
	dialog := NewErrorDialog("Error message")

	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedDialog, cmd := dialog.Update(keyMsg)

	if updatedDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C")
	}

	if cmd == nil {
		t.Error("Dialog should return command after Ctrl+C")
	}
}

func TestHelpDialogCtrlCCloses(t *testing.T) {
	dialog := NewHelpDialog()

	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedDialog, cmd := dialog.Update(keyMsg)

	if updatedDialog.IsActive() {
		t.Error("Dialog should be inactive after Ctrl+C")
	}

	if cmd == nil {
		t.Error("Dialog should return command after Ctrl+C")
	}

	// Execute and verify result
	if cmd != nil {
		msg := cmd()
		if result, ok := msg.(dialogResultMsg); ok {
			if !result.result.Cancelled {
				t.Error("Result should be cancelled")
			}
		}
	}
}

// === DisplayType のテスト ===

func TestConfirmDialogDisplayType(t *testing.T) {
	dialog := NewConfirmDialog("Test", "Message")

	displayType := dialog.DisplayType()
	if displayType != DialogDisplayPane {
		t.Errorf("ConfirmDialog.DisplayType() = %v, want DialogDisplayPane", displayType)
	}
}

func TestHelpDialogDisplayType(t *testing.T) {
	dialog := NewHelpDialog()

	displayType := dialog.DisplayType()
	if displayType != DialogDisplayScreen {
		t.Errorf("HelpDialog.DisplayType() = %v, want DialogDisplayScreen", displayType)
	}
}

func TestContextMenuDialogDisplayType(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}
	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	displayType := dialog.DisplayType()
	if displayType != DialogDisplayPane {
		t.Errorf("ContextMenuDialog.DisplayType() = %v, want DialogDisplayPane", displayType)
	}
}
