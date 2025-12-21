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
)
