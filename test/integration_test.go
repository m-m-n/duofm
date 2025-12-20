package test

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/ui"
)

// testModel はテスト用のモデル実行
func testModel(t *testing.T, inputs []tea.Msg) ui.Model {
	m := ui.NewModel()

	// ウィンドウサイズを設定
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(ui.Model)

	// 各入力を処理
	for _, input := range inputs {
		result, cmd := m.Update(input)
		m = result.(ui.Model)

		// コマンドが返された場合は実行
		if cmd != nil {
			msg := cmd()
			if msg != nil {
				result, _ := m.Update(msg)
				m = result.(ui.Model)
			}
		}
	}

	return m
}

func TestNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// テストファイルを作成
	os.Mkdir("testdir", 0755)
	os.WriteFile("file1.txt", []byte("test"), 0644)
	os.WriteFile("file2.txt", []byte("test"), 0644)

	inputs := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, // 下に移動
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}, // 上に移動
	}

	m := testModel(t, inputs)

	// ビューがレンダリングされることを確認
	view := m.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// タイトルが表示されることを確認
	if !contains(view, "duofm") {
		t.Error("View should contain title")
	}
}

func TestPaneSwitch(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// 左ペインから右ペインに切り替え
	inputs := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}, // 右ペインへ
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}, // 左ペインへ
	}

	m := testModel(t, inputs)

	view := m.View()
	if view == "" {
		t.Error("View should not be empty after pane switch")
	}
}

func TestHelpDialog(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	inputs := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, // ヘルプを開く
	}

	m := testModel(t, inputs)
	view := m.View()

	// ヘルプダイアログが表示されていることを確認
	if !contains(view, "Keybindings") {
		t.Error("Help dialog should be visible")
	}

	// Escでヘルプを閉じる
	inputs2 := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyEsc},
	}

	for _, input := range inputs2 {
		result, cmd := m.Update(input)
		m = result.(ui.Model)
		if cmd != nil {
			msg := cmd()
			if msg != nil {
				result, _ := m.Update(msg)
				m = result.(ui.Model)
			}
		}
	}

	view = m.View()
	// ヘルプが閉じていることを確認（Keybindingsが表示されていない）
	if contains(view, "Keybindings") {
		t.Error("Help dialog should be closed after Esc")
	}
}

func TestDirectoryNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// サブディレクトリを作成
	os.Mkdir("subdir", 0755)
	os.WriteFile(filepath.Join("subdir", "file.txt"), []byte("test"), 0644)

	m := ui.NewModel()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(ui.Model)

	// 初期ビュー確認
	view := m.View()
	if !contains(view, "subdir") {
		t.Error("Should show subdir in initial view")
	}

	// カーソルをsubdirに移動してEnter
	inputs := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, // subdirに移動（..の次）
		tea.KeyMsg{Type: tea.KeyEnter},                     // ディレクトリに入る
	}

	for _, input := range inputs {
		result, _ := m.Update(input)
		m = result.(ui.Model)
	}

	view = m.View()
	if !contains(view, "file.txt") {
		t.Error("Should show file.txt after entering subdir")
	}
}

func TestDeleteConfirmation(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// テストファイルを作成
	testFile := "delete_me.txt"
	os.WriteFile(testFile, []byte("test"), 0644)

	m := ui.NewModel()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(ui.Model)

	// カーソルをファイルに移動してdキーで削除ダイアログを開く
	inputs := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, // ファイルに移動
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, // 削除ダイアログ
	}

	for _, input := range inputs {
		result, _ := m.Update(input)
		m = result.(ui.Model)
	}

	view := m.View()
	// 確認ダイアログが表示されることを確認
	if !contains(view, "Delete file?") {
		t.Error("Delete confirmation dialog should be shown")
	}

	// nでキャンセル
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = result.(ui.Model)
	if cmd != nil {
		msg := cmd()
		if msg != nil {
			result, _ := m.Update(msg)
			m = result.(ui.Model)
		}
	}

	// ファイルがまだ存在することを確認
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("File should still exist after cancel")
	}
}

func TestErrorDialog(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// テストファイルを作成
	os.WriteFile("file1.txt", []byte("test"), 0644)

	m := ui.NewModel()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(ui.Model)

	// 存在しないファイルのコピーを試みる（エラーを発生させる）
	// 実際にはこのテストは手動テストに近い形になる
	view := m.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestQuit(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	m := ui.NewModel()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(ui.Model)

	// qキーで終了
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Error("Quit command should be returned")
	}

	// tea.Quit コマンドが返されることを確認
	msg := cmd()
	if msg != tea.Quit() {
		t.Error("Should return Quit message")
	}
}

// contains は文字列が部分文字列を含むかチェック
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		len(s) >= len(substr) &&
		stringContains(s, substr)
}

// stringContains は文字列検索のヘルパー
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
