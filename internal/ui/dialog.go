package ui

import tea "github.com/charmbracelet/bubbletea"

// Dialog はモーダルダイアログのインターフェース
type Dialog interface {
	Update(msg tea.Msg) (Dialog, tea.Cmd)
	View() string
	IsActive() bool
}

// DialogResult はダイアログの結果
type DialogResult struct {
	Confirmed bool
	Cancelled bool
}

// dialogResultMsg はダイアログ結果のメッセージ
type dialogResultMsg struct {
	result DialogResult
}
