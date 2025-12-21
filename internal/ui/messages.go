package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
)

// diskSpaceUpdateMsg はディスク容量の定期更新を通知
type diskSpaceUpdateMsg struct{}

// diskSpaceTickCmd は5秒後にdiskSpaceUpdateMsgを送信するコマンド
func diskSpaceTickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return diskSpaceUpdateMsg{}
	})
}

// clearStatusMsg はステータスメッセージをクリアするメッセージ
type clearStatusMsg struct{}

// statusMessageClearCmd は指定時間後にclearStatusMsgを送信するコマンド
func statusMessageClearCmd(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// directoryLoadStartMsg はディレクトリ読み込み開始を通知
type directoryLoadStartMsg struct {
	panePath string
}

// directoryLoadCompleteMsg はディレクトリ読み込み完了を通知
type directoryLoadCompleteMsg struct {
	panePath      string
	entries       []fs.FileEntry
	err           error
	attemptedPath string // エラー時にメッセージに表示するパス
}

// directoryLoadProgressMsg は読み込み進捗を通知（オプション）
type directoryLoadProgressMsg struct {
	panePath  string
	fileCount int
}
