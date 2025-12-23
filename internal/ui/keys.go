package ui

// キーバインディング定義
const (
	KeyMoveDown    = "j"
	KeyMoveUp      = "k"
	KeyMoveLeft    = "h"
	KeyMoveRight   = "l"
	KeyEnter       = "enter"
	KeyCopy        = "c"
	KeyMove        = "m"
	KeyDelete      = "d"
	KeyHelp        = "?"
	KeyEscape      = "esc"
	KeyQuit        = "q"
	KeyToggleInfo  = "i" // 表示モード切り替え
	KeyContextMenu = "@" // コンテキストメニュー

	// ナビゲーション強化
	KeyToggleHidden = "ctrl+h" // 隠しファイル表示トグル
	KeyHome         = "~"      // ホームディレクトリへ移動
	KeyPrevDir      = "-"      // 直前のディレクトリへ移動

	// カーソルキー（hjklの代替）
	KeyArrowDown  = "down"
	KeyArrowUp    = "up"
	KeyArrowLeft  = "left"
	KeyArrowRight = "right"

	// 検索機能
	KeySearch      = "/"      // インクリメンタル検索
	KeyRegexSearch = "ctrl+f" // 正規表現検索

	// ファイル操作
	KeyView = "v" // ファイルをビューアー(less)で開く
	KeyEdit = "e" // ファイルをエディタ(vim)で開く

	// Refresh and sync
	KeyRefresh    = "f5"     // Refresh view
	KeyRefreshAlt = "ctrl+r" // Refresh view (alternative)
	KeySyncPane   = "="      // Pane synchronization

	// File/directory creation and renaming
	KeyNewFile      = "n" // Create new file
	KeyNewDirectory = "N" // Create new directory (Shift+n)
	KeyRename       = "r" // Rename file/directory
)
