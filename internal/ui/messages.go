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
	paneID                   PanePosition // どちらのペインの読み込みか
	panePath                 string
	entries                  []fs.FileEntry
	err                      error
	attemptedPath            string // エラー時にメッセージに表示するパス
	isHistoryNavigation      bool   // 履歴ナビゲーション経由かどうか（履歴ナビゲーション自体は記録しない）
	historyNavigationForward bool   // true=前進、false=後退（履歴ナビゲーションエラー時の復元用）
}

// directoryLoadProgressMsg は読み込み進捗を通知（オプション）
type directoryLoadProgressMsg struct {
	panePath  string
	fileCount int
}

// ctrlCTimeoutMsg はCtrl+C終了確認のタイムアウトを通知
type ctrlCTimeoutMsg struct{}

// ctrlCTimeoutCmd は指定時間後にctrlCTimeoutMsgを送信するコマンド
func ctrlCTimeoutCmd(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return ctrlCTimeoutMsg{}
	})
}

// inputDialogResultMsg は入力ダイアログの結果を通知
type inputDialogResultMsg struct {
	operation string // "create_file", "create_dir", "rename"
	input     string // 入力された名前
	oldName   string // リネームの場合の元の名前
	err       error  // エラー
}

// archiveOperationStartMsg はアーカイブ操作の開始を通知
type archiveOperationStartMsg struct {
	taskID string // タスクID
}

// archiveProgressUpdateMsg はアーカイブ操作の進捗更新を通知
type archiveProgressUpdateMsg struct {
	taskID          string        // タスクID
	progress        float64       // 進捗率 (0.0-1.0)
	processedFiles  int           // 処理済みファイル数
	totalFiles      int           // 総ファイル数
	currentFile     string        // 現在処理中のファイル
	elapsedTime     time.Duration // 経過時間
	estimatedRemain time.Duration // 推定残り時間
}

// archiveOperationCompleteMsg はアーカイブ操作の完了を通知
type archiveOperationCompleteMsg struct {
	taskID      string // タスクID
	success     bool   // 成功したかどうか
	cancelled   bool   // キャンセルされたかどうか
	archivePath string // 作成されたアーカイブのパス（圧縮/展開の場合）
	err         error  // エラー（失敗時）
}

// archiveOperationErrorMsg はアーカイブ操作のエラーを通知
type archiveOperationErrorMsg struct {
	taskID  string // タスクID
	err     error  // エラー
	message string // ユーザー向けエラーメッセージ
}

// compressionLevelResultMsg は圧縮レベル選択の結果を通知
type compressionLevelResultMsg struct {
	level     int  // 選択された圧縮レベル (0-9)
	cancelled bool // キャンセルされた場合
}

// archiveNameResultMsg はアーカイブ名入力の結果を通知
type archiveNameResultMsg struct {
	name      string // 入力されたアーカイブ名
	cancelled bool   // キャンセルされた場合
}

// extractSecurityCheckMsg はアーカイブ展開前のセキュリティチェック結果を通知
type extractSecurityCheckMsg struct {
	archivePath   string  // アーカイブパス
	destDir       string  // 展開先ディレクトリ
	archiveSize   int64   // アーカイブサイズ
	extractedSize int64   // 展開後サイズ
	availableSize int64   // 展開先の空き容量
	compressionOK bool    // 圧縮率が正常か
	diskSpaceOK   bool    // ディスク容量が十分か
	ratio         float64 // 圧縮率（展開サイズ/アーカイブサイズ）
	err           error   // エラー（メタデータ取得失敗時）
}

// permissionOperationStartMsg はパーミッション変更操作の開始を通知
type permissionOperationStartMsg struct {
	path      string // 対象パス
	mode      string // 新しいパーミッション
	recursive bool   // 再帰的変更かどうか
}

// permissionOperationCompleteMsg はパーミッション変更操作の完了を通知
type permissionOperationCompleteMsg struct {
	path    string // 対象パス
	success bool   // 成功したかどうか
	err     error  // エラー（失敗時）
}

// showRecursivePermDialogMsg はRecursivePermDialogの表示を通知
type showRecursivePermDialogMsg struct {
	path string // 対象パス
}

// batchPermissionStartMsg はバッチパーミッション変更開始を通知
type batchPermissionStartMsg struct {
	paths []string
	mode  string
}

// batchPermissionCompleteMsg はバッチパーミッション変更完了を通知
type batchPermissionCompleteMsg struct {
	totalCount   int
	successCount int
	failedCount  int
	errors       []fs.PermissionError
}

// batchPermissionProgressMsg はバッチパーミッション変更進捗を通知
type batchPermissionProgressMsg struct {
	processed   int
	total       int
	currentPath string
}

// recursivePermissionCompleteMsg は再帰的パーミッション変更完了を通知
type recursivePermissionCompleteMsg struct {
	path         string
	successCount int
	errors       []fs.PermissionError
}
