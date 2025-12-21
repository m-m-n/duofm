package ui

import tea "github.com/charmbracelet/bubbletea"

// DialogDisplayType はダイアログの表示タイプを表す
type DialogDisplayType int

const (
	// DialogDisplayPane はペインローカルダイアログ（アクティブペインのみdimmed）
	DialogDisplayPane DialogDisplayType = iota
	// DialogDisplayScreen は画面全体ダイアログ（両ペインdimmed）
	DialogDisplayScreen
)

// Dialog はモーダルダイアログのインターフェース
type Dialog interface {
	Update(msg tea.Msg) (Dialog, tea.Cmd)
	View() string
	IsActive() bool
	DisplayType() DialogDisplayType
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
