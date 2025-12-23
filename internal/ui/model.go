package ui

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/fs"
)

// ANSIエスケープシーケンスを除去するための正規表現
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Model はアプリケーション全体の状態を保持
type Model struct {
	leftPane           *Pane
	rightPane          *Pane
	leftPath           string
	rightPath          string
	activePane         PanePosition
	dialog             Dialog
	width              int
	height             int
	ready              bool
	lastDiskSpaceCheck time.Time    // 最後のディスク容量チェック時刻
	leftDiskSpace      uint64       // 左ペインのディスク空き容量
	rightDiskSpace     uint64       // 右ペインのディスク空き容量
	pendingAction      func() error // 確認待ちのアクション（コンテキストメニューの削除用）
	statusMessage      string       // ステータスバーに表示するメッセージ
	isStatusError      bool         // エラーメッセージかどうか
	searchState        SearchState  // 検索状態
	minibuffer         *Minibuffer  // ミニバッファ
	ctrlCPending       bool         // Ctrl+Cが1回押された状態かどうか
}

// PanePosition はペインの位置を表す
type PanePosition int

const (
	// LeftPane は左ペイン
	LeftPane PanePosition = iota
	// RightPane は右ペイン
	RightPane
)

// NewModel は初期モデルを作成
func NewModel() Model {
	// 初期ディレクトリの取得
	cwd, err := fs.CurrentDirectory()
	if err != nil {
		cwd = "/"
	}

	home, err := fs.HomeDirectory()
	if err != nil {
		home = "/"
	}

	return Model{
		leftPane:    nil, // Updateで初期化
		rightPane:   nil, // Updateで初期化
		leftPath:    cwd,
		rightPath:   home,
		activePane:  LeftPane,
		dialog:      nil,
		ready:       false,
		searchState: SearchState{Mode: SearchModeNone},
		minibuffer:  NewMinibuffer(),
	}
}

// Init はBubble Teaの初期化
func (m Model) Init() tea.Cmd {
	return nil
}

// Update はメッセージを処理してモデルを更新
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// コンテキストメニューの結果処理
	if result, ok := msg.(contextMenuResultMsg); ok {
		prevDialog := m.dialog
		m.dialog = nil

		if _, ok := prevDialog.(*ContextMenuDialog); ok {
			if result.cancelled {
				// メニューがキャンセルされた
				return m, nil
			}

			if result.action != nil {
				// 削除の場合は確認ダイアログを表示
				if result.actionID == "delete" {
					entry := m.getActivePane().SelectedEntry()
					if entry != nil && !entry.IsParentDir() {
						m.pendingAction = result.action
						m.dialog = NewConfirmDialog(
							"Delete file?",
							entry.DisplayName(),
						)
					}
					return m, nil
				}

				// その他のアクションは直接実行
				if err := result.action(); err != nil {
					m.dialog = NewErrorDialog(fmt.Sprintf("Operation failed: %v", err))
					return m, nil
				}

				// 両ペインを再読み込みして変更を反映
				m.getActivePane().LoadDirectory()
				m.getInactivePane().LoadDirectory()
			}
		}

		return m, nil
	}

	// ダイアログの結果処理
	if result, ok := msg.(dialogResultMsg); ok {
		prevDialog := m.dialog
		m.dialog = nil

		// 削除確認の結果
		if result.result.Confirmed {
			if _, ok := prevDialog.(*ConfirmDialog); ok {
				// コンテキストメニューからの削除（pendingActionあり）
				if m.pendingAction != nil {
					if err := m.pendingAction(); err != nil {
						m.dialog = NewErrorDialog(fmt.Sprintf("Failed to delete: %v", err))
					} else {
						// 両ペインを再読み込み
						m.getActivePane().LoadDirectory()
						m.getInactivePane().LoadDirectory()
					}
					m.pendingAction = nil
					return m, nil
				}

				// 通常の削除（dキーから）
				entry := m.getActivePane().SelectedEntry()
				if entry != nil && !entry.IsParentDir() {
					fullPath := filepath.Join(m.getActivePane().Path(), entry.Name)

					if err := fs.Delete(fullPath); err != nil {
						// エラーダイアログを表示
						m.dialog = NewErrorDialog(fmt.Sprintf("Failed to delete: %v", err))
					} else {
						// ディレクトリを再読み込み
						m.getActivePane().LoadDirectory()
					}
				}
			}
		} else {
			// キャンセルされた場合、pendingActionをクリア
			m.pendingAction = nil
		}

		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			// 初回のみペインを作成
			paneWidth := msg.Width / 2
			paneHeight := msg.Height - 2 // ステータスバー分を引く

			var err error
			m.leftPane, err = NewPane(m.leftPath, paneWidth, paneHeight, true)
			if err != nil {
				// エラーハンドリングは Phase 3 で実装
				return m, tea.Quit
			}

			m.rightPane, err = NewPane(m.rightPath, paneWidth, paneHeight, false)
			if err != nil {
				return m, tea.Quit
			}

			// 初回のディスク容量を取得してタイマー開始
			m.updateDiskSpace()
			m.ready = true

			return m, diskSpaceTickCmd()
		} else {
			// リサイズ時のペインサイズ更新
			paneWidth := msg.Width / 2
			paneHeight := msg.Height - 2
			m.leftPane.SetSize(paneWidth, paneHeight)
			m.rightPane.SetSize(paneWidth, paneHeight)
		}

		return m, nil

	case diskSpaceUpdateMsg:
		// ディスク容量を更新して次の更新をスケジュール
		m.updateDiskSpace()
		return m, diskSpaceTickCmd()

	case clearStatusMsg:
		// ステータスメッセージをクリア
		m.statusMessage = ""
		m.isStatusError = false
		return m, nil

	case ctrlCTimeoutMsg:
		// Ctrl+Cタイムアウト - 状態をリセット
		if m.ctrlCPending {
			m.ctrlCPending = false
			m.statusMessage = ""
		}
		return m, nil

	case directoryLoadCompleteMsg:
		// ディレクトリ読み込み完了
		var targetPane *Pane
		// どのペインの読み込みかを判定（pendingPathも確認）
		if msg.panePath == m.leftPane.Path() || msg.panePath == m.leftPane.pendingPath {
			targetPane = m.leftPane
		} else if msg.panePath == m.rightPane.Path() || msg.panePath == m.rightPane.pendingPath {
			targetPane = m.rightPane
		}

		if targetPane != nil {
			targetPane.loading = false
			targetPane.loadingProgress = ""

			if msg.err != nil {
				// エラー時: パスを復元してステータスバーにメッセージ表示
				targetPane.restorePreviousPath()
				m.statusMessage = formatDirectoryError(msg.err, msg.attemptedPath)
				m.isStatusError = true
				return m, statusMessageClearCmd(5 * time.Second)
			}

			// 成功時: エントリを更新
			entries := msg.entries
			if !targetPane.showHidden {
				entries = filterHiddenFiles(entries)
			}
			targetPane.entries = entries
			targetPane.cursor = 0
			targetPane.scrollOffset = 0
			targetPane.pendingPath = ""

			// ディスク容量を更新
			m.updateDiskSpace()
		}
		return m, nil

	case execFinishedMsg:
		// 外部コマンド完了
		// 両ペインを再読み込みして変更を反映
		m.getActivePane().LoadDirectory()
		m.getInactivePane().LoadDirectory()

		if msg.err != nil {
			m.statusMessage = fmt.Sprintf("Command failed: %v", msg.err)
			m.isStatusError = true
			return m, statusMessageClearCmd(5 * time.Second)
		}
		return m, nil

	case inputDialogResultMsg:
		// 入力ダイアログの結果処理
		m.dialog = nil

		if msg.err != nil {
			m.statusMessage = msg.err.Error()
			m.isStatusError = true
			return m, statusMessageClearCmd(5 * time.Second)
		}

		// 両ペインを再読み込み
		m.getActivePane().LoadDirectory()
		m.getInactivePane().LoadDirectory()

		// カーソル位置を調整
		switch msg.operation {
		case "create_file", "create_dir":
			m.moveCursorToFile(msg.input)
		case "rename":
			m.moveCursorToFileAfterRename(msg.oldName, msg.input)
		}
		return m, nil

	case tea.KeyMsg:
		// ダイアログが開いている場合はダイアログに処理を委譲
		if m.dialog != nil {
			var cmd tea.Cmd
			m.dialog, cmd = m.dialog.Update(msg)
			return m, cmd
		}

		// ミニバッファがアクティブな場合（検索中）
		if m.searchState.IsActive {
			switch msg.Type {
			case tea.KeyEnter:
				// 検索を確定
				m.confirmSearch()
				return m, nil

			case tea.KeyEsc, tea.KeyCtrlC:
				// 検索をキャンセル
				m.cancelSearch()
				return m, nil

			default:
				// ミニバッファにキー入力を渡す
				if m.minibuffer.HandleKey(msg) {
					// インクリメンタル検索の場合は即座にフィルタを適用
					m.applyIncrementalFilter()
					return m, nil
				}
			}
			return m, nil
		}

		// Ctrl+Cのダブルプレス処理（他のキー処理より先に実行）
		if msg.String() == "ctrl+c" {
			if m.ctrlCPending {
				// 2回目のCtrl+C - 終了
				return m, tea.Quit
			}
			// 1回目のCtrl+C - メッセージ表示とタイマー開始
			m.ctrlCPending = true
			m.statusMessage = "Press Ctrl+C again to quit"
			m.isStatusError = false
			return m, ctrlCTimeoutCmd(2 * time.Second)
		}

		// ステータスメッセージがあればクリア、Ctrl+C待機状態もリセット
		if m.statusMessage != "" || m.ctrlCPending {
			m.statusMessage = ""
			m.isStatusError = false
			m.ctrlCPending = false
		}

		switch msg.String() {

		case KeyRefresh, KeyRefreshAlt:
			return m, m.RefreshBothPanes()

		case KeySyncPane:
			m.SyncOppositePane()
			return m, nil

		case KeyQuit:
			return m, tea.Quit

		case KeyHelp:
			// ヘルプダイアログを表示
			m.dialog = NewHelpDialog()
			return m, nil

		case KeySearch:
			// インクリメンタル検索を開始
			m.startSearch(SearchModeIncremental)
			return m, nil

		case KeyRegexSearch:
			// 正規表現検索を開始
			m.startSearch(SearchModeRegex)
			return m, nil

		case KeyMoveDown, KeyArrowDown:
			m.getActivePane().MoveCursorDown()

		case KeyMoveUp, KeyArrowUp:
			m.getActivePane().MoveCursorUp()

		case KeyMoveLeft, KeyArrowLeft:
			if m.activePane == LeftPane {
				// 左ペインで h/← -> 親ディレクトリへ（非同期版）
				cmd := m.leftPane.MoveToParentAsync()
				return m, cmd
			} else {
				// 右ペインで h/← -> 左ペインへ切り替え
				m.switchToPane(LeftPane)
			}

		case KeyMoveRight, KeyArrowRight:
			if m.activePane == RightPane {
				// 右ペインで l/→ -> 親ディレクトリへ（非同期版）
				cmd := m.rightPane.MoveToParentAsync()
				return m, cmd
			} else {
				// 左ペインで l/→ -> 右ペインへ切り替え
				m.switchToPane(RightPane)
			}

		case KeyEnter:
			entry := m.getActivePane().SelectedEntry()
			if entry != nil && !entry.IsParentDir() && !entry.IsDir {
				// ファイル選択時: ビューアー(less)で開く（vキーと同じ）
				fullPath := filepath.Join(m.getActivePane().Path(), entry.Name)
				if err := checkReadPermission(fullPath); err != nil {
					m.statusMessage = fmt.Sprintf("Cannot read file: %v", err)
					m.isStatusError = true
					return m, statusMessageClearCmd(5 * time.Second)
				}
				return m, openWithViewer(fullPath)
			}
			// ディレクトリまたは親ディレクトリ: 既存の動作（非同期版）
			cmd := m.getActivePane().EnterDirectoryAsync()
			return m, cmd

		case KeyToggleInfo:
			// 表示モードを切り替え（端末が十分な幅の場合のみ）
			activePane := m.getActivePane()
			if activePane.CanToggleMode() {
				activePane.ToggleDisplayMode()
			}
			return m, nil

		case KeyCopy:
			// コピー操作
			entry := m.getActivePane().SelectedEntry()
			if entry != nil && !entry.IsParentDir() {
				srcPath := filepath.Join(m.getActivePane().Path(), entry.Name)
				dstPath := m.getInactivePane().Path()

				if err := fs.Copy(srcPath, dstPath); err != nil {
					m.dialog = NewErrorDialog(fmt.Sprintf("Failed to copy: %v", err))
				} else {
					// 対象ペインを再読み込み
					m.getInactivePane().LoadDirectory()
				}
			}
			return m, nil

		case KeyMove:
			// 移動操作
			entry := m.getActivePane().SelectedEntry()
			if entry != nil && !entry.IsParentDir() {
				srcPath := filepath.Join(m.getActivePane().Path(), entry.Name)
				dstPath := m.getInactivePane().Path()

				if err := fs.MoveFile(srcPath, dstPath); err != nil {
					m.dialog = NewErrorDialog(fmt.Sprintf("Failed to move: %v", err))
				} else {
					// 両ペインを再読み込み
					m.getActivePane().LoadDirectory()
					m.getInactivePane().LoadDirectory()
				}
			}
			return m, nil

		case KeyDelete:
			// 削除確認ダイアログを表示
			entry := m.getActivePane().SelectedEntry()
			if entry != nil && !entry.IsParentDir() {
				m.dialog = NewConfirmDialog(
					"Delete file?",
					entry.DisplayName(),
				)
			}
			return m, nil

		case KeyContextMenu:
			// コンテキストメニューを表示
			activePane := m.getActivePane()
			entry := activePane.SelectedEntry()

			if entry != nil && !entry.IsParentDir() {
				m.dialog = NewContextMenuDialogWithPane(
					entry,
					activePane.Path(),
					m.getInactivePane().Path(),
					activePane,
				)
			}
			return m, nil

		case KeyToggleHidden:
			// 隠しファイル表示をトグル
			m.getActivePane().ToggleHidden()
			return m, nil

		case KeyHome:
			// ホームディレクトリへ移動（非同期版）
			cmd := m.getActivePane().NavigateToHomeAsync()
			return m, cmd

		case KeyPrevDir:
			// 直前のディレクトリへ移動（非同期版）
			cmd := m.getActivePane().NavigateToPreviousAsync()
			return m, cmd

		case KeyView:
			// ファイルをビューアー(less)で開く
			entry := m.getActivePane().SelectedEntry()
			if entry != nil && !entry.IsParentDir() && !entry.IsDir {
				fullPath := filepath.Join(m.getActivePane().Path(), entry.Name)
				if err := checkReadPermission(fullPath); err != nil {
					m.statusMessage = fmt.Sprintf("Cannot read file: %v", err)
					m.isStatusError = true
					return m, statusMessageClearCmd(5 * time.Second)
				}
				return m, openWithViewer(fullPath)
			}
			return m, nil

		case KeyEdit:
			// ファイルをエディタ(vim)で開く
			entry := m.getActivePane().SelectedEntry()
			if entry != nil && !entry.IsParentDir() && !entry.IsDir {
				fullPath := filepath.Join(m.getActivePane().Path(), entry.Name)
				if err := checkReadPermission(fullPath); err != nil {
					m.statusMessage = fmt.Sprintf("Cannot read file: %v", err)
					m.isStatusError = true
					return m, statusMessageClearCmd(5 * time.Second)
				}
				return m, openWithEditor(fullPath)
			}
			return m, nil

		case KeyNewFile:
			// 新規ファイル作成ダイアログを表示
			pane := m.getActivePane()
			m.dialog = NewInputDialog("New file:", func(filename string) tea.Cmd {
				return m.handleCreateFile(pane.Path(), filename)
			})
			return m, nil

		case KeyNewDirectory:
			// 新規ディレクトリ作成ダイアログを表示
			pane := m.getActivePane()
			m.dialog = NewInputDialog("New directory:", func(dirname string) tea.Cmd {
				return m.handleCreateDirectory(pane.Path(), dirname)
			})
			return m, nil

		case KeyRename:
			// リネームダイアログを表示
			entry := m.getActivePane().SelectedEntry()
			if entry == nil || entry.IsParentDir() {
				// 親ディレクトリは無視
				return m, nil
			}
			pane := m.getActivePane()
			oldName := entry.Name
			m.dialog = NewInputDialog("Rename to:", func(newName string) tea.Cmd {
				return m.handleRename(pane.Path(), oldName, newName)
			})
			return m, nil
		}
	}

	return m, nil
}

// getActivePane は現在アクティブなペインを返す
func (m *Model) getActivePane() *Pane {
	if m.activePane == LeftPane {
		return m.leftPane
	}
	return m.rightPane
}

// getInactivePane は非アクティブなペインを返す
func (m *Model) getInactivePane() *Pane {
	if m.activePane == LeftPane {
		return m.rightPane
	}
	return m.leftPane
}

// switchToPane はアクティブペインを切り替え
func (m *Model) switchToPane(pos PanePosition) {
	// 検索中の場合はキャンセル
	if m.searchState.IsActive {
		m.cancelSearch()
	}

	m.activePane = pos
	m.leftPane.SetActive(pos == LeftPane)
	m.rightPane.SetActive(pos == RightPane)
}

// View はUIをレンダリング
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// タイトルバー
	title := titleStyle.Render("duofm v0.1.0")

	// 2つのペインを横に並べる（ディスク容量情報付き）
	// 検索モードの場合はアクティブペインにミニバッファを渡す
	var leftView, rightView string
	if m.searchState.IsActive {
		if m.activePane == LeftPane {
			leftView = m.leftPane.ViewWithMinibuffer(m.leftDiskSpace, m.minibuffer)
			rightView = m.rightPane.ViewWithDiskSpace(m.rightDiskSpace)
		} else {
			leftView = m.leftPane.ViewWithDiskSpace(m.leftDiskSpace)
			rightView = m.rightPane.ViewWithMinibuffer(m.rightDiskSpace, m.minibuffer)
		}
	} else {
		leftView = m.leftPane.ViewWithDiskSpace(m.leftDiskSpace)
		rightView = m.rightPane.ViewWithDiskSpace(m.rightDiskSpace)
	}
	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)

	// ステータスバー
	statusBar := m.renderStatusBar()

	// 全体を縦に結合
	mainView := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		panes,
		statusBar,
	)

	// ダイアログがある場合は表示タイプに応じて描画
	if m.dialog != nil && m.dialog.IsActive() {
		switch m.dialog.DisplayType() {
		case DialogDisplayScreen:
			return m.renderDialogScreen()
		case DialogDisplayPane:
			return m.renderDialogPane()
		}
	}

	return mainView
}

// renderDialogScreen は画面全体表示ダイアログをレンダリング（両ペインdimmed）
func (m Model) renderDialogScreen() string {
	// タイトルバー
	title := titleStyle.Render("duofm v0.1.0")

	// 両方のペインをdimmedスタイルで描画
	leftView := m.leftPane.ViewDimmedWithDiskSpace(m.leftDiskSpace)
	rightView := m.rightPane.ViewDimmedWithDiskSpace(m.rightDiskSpace)
	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)

	// ステータスバー
	statusBar := m.renderStatusBar()

	// ペイン全体のサイズ
	panesHeight := m.height - 2 // タイトルバーとステータスバー分を引く

	// ダイアログを画面中央に配置（背景をdimmed色で埋める）
	dialogView := lipgloss.Place(
		m.width,
		panesHeight,
		lipgloss.Center,
		lipgloss.Center,
		m.dialog.View(),
		lipgloss.WithWhitespaceBackground(dimmedBgColor),
	)

	// dimmedペインの上にダイアログをオーバーレイ
	// 各行を結合してオーバーレイ効果を出す
	panesLines := strings.Split(panes, "\n")
	dialogLines := strings.Split(dialogView, "\n")

	var result strings.Builder
	for i := 0; i < len(panesLines) && i < len(dialogLines); i++ {
		// ダイアログ行が空白のみなら背景を使用、そうでなければダイアログを使用
		dialogLine := dialogLines[i]
		paneLine := panesLines[i]

		// ダイアログ行の内容をチェック（ANSIエスケープシーケンスを除去してから空白判定）
		stripped := ansiRegex.ReplaceAllString(dialogLine, "")
		trimmed := strings.TrimSpace(stripped)
		if trimmed == "" {
			result.WriteString(paneLine)
		} else {
			result.WriteString(dialogLine)
		}
		if i < len(panesLines)-1 {
			result.WriteString("\n")
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, result.String(), statusBar)
}

// renderDialogPane はペインローカルダイアログをレンダリング（アクティブペインのみdimmed）
func (m Model) renderDialogPane() string {
	paneWidth := m.width / 2
	paneHeight := m.height - 2 // タイトルバーとステータスバー分を引く

	// タイトルバー
	title := titleStyle.Render("duofm v0.1.0")

	// ステータスバー
	statusBar := m.renderStatusBar()

	var leftView, rightView string

	if m.activePane == LeftPane {
		// 左ペインをdimmedで描画してダイアログをオーバーレイ
		dimmedLeft := m.leftPane.ViewDimmedWithDiskSpace(m.leftDiskSpace)
		leftView = m.overlayDialogOnPane(dimmedLeft, paneWidth, paneHeight)
		rightView = m.rightPane.ViewWithDiskSpace(m.rightDiskSpace)
	} else {
		// 右ペインをdimmedで描画してダイアログをオーバーレイ
		leftView = m.leftPane.ViewWithDiskSpace(m.leftDiskSpace)
		dimmedRight := m.rightPane.ViewDimmedWithDiskSpace(m.rightDiskSpace)
		rightView = m.overlayDialogOnPane(dimmedRight, paneWidth, paneHeight)
	}

	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
	return lipgloss.JoinVertical(lipgloss.Left, title, panes, statusBar)
}

// overlayDialogOnPane はdimmedペインの上にダイアログをオーバーレイする
func (m Model) overlayDialogOnPane(dimmedPane string, paneWidth, paneHeight int) string {
	// ダイアログをペイン中央に配置（背景をdimmed色で埋める）
	dialogView := lipgloss.Place(
		paneWidth,
		paneHeight,
		lipgloss.Center,
		lipgloss.Center,
		m.dialog.View(),
		lipgloss.WithWhitespaceBackground(dimmedBgColor),
	)

	// dimmedペインの上にダイアログをオーバーレイ
	paneLines := strings.Split(dimmedPane, "\n")
	dialogLines := strings.Split(dialogView, "\n")

	var result strings.Builder
	for i := 0; i < len(paneLines) && i < len(dialogLines); i++ {
		dialogLine := dialogLines[i]
		paneLine := paneLines[i]

		// ダイアログ行が空白のみなら背景を使用（ANSIエスケープシーケンスを除去してから判定）
		stripped := ansiRegex.ReplaceAllString(dialogLine, "")
		trimmed := strings.TrimSpace(stripped)
		if trimmed == "" {
			result.WriteString(paneLine)
		} else {
			result.WriteString(dialogLine)
		}
		if i < len(paneLines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// renderStatusBar はステータスバーをレンダリング
func (m Model) renderStatusBar() string {
	// ステータスメッセージがある場合はそれを優先表示
	if m.statusMessage != "" {
		style := lipgloss.NewStyle().
			Width(m.width).
			Padding(0, 1)

		if m.isStatusError {
			// エラーメッセージは赤背景で表示
			style = style.
				Background(lipgloss.Color("124")). // 暗めの赤
				Foreground(lipgloss.Color("15"))   // 白
		} else {
			style = style.
				Background(lipgloss.Color("240")).
				Foreground(lipgloss.Color("15"))
		}

		// メッセージを幅に合わせて切り詰め
		msg := m.statusMessage
		maxLen := m.width - 2 // パディング分を引く
		if len(msg) > maxLen {
			msg = msg[:maxLen-3] + "..."
		}

		return style.Render(msg)
	}

	activePane := m.getActivePane()

	// 選択位置情報
	posInfo := fmt.Sprintf("%d/%d", activePane.cursor+1, len(activePane.entries))

	// キーヒント（動的に変更）
	hints := "?:help q:quit"
	if activePane != nil && activePane.CanToggleMode() {
		hints = "i:info " + hints
	}

	// スペースで埋める
	padding := m.width - len(posInfo) - len(hints) - 4
	if padding < 0 {
		padding = 0
	}

	statusBar := fmt.Sprintf(" %s%s%s ",
		posInfo,
		strings.Repeat(" ", padding),
		hints,
	)

	style := lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("15"))

	return style.Render(statusBar)
}

// updateDiskSpace はディスク容量を更新
func (m *Model) updateDiskSpace() {
	if m.leftPane != nil {
		if freeBytes, _, err := fs.GetDiskSpace(m.leftPane.Path()); err == nil {
			m.leftDiskSpace = freeBytes
		}
	}

	if m.rightPane != nil {
		if freeBytes, _, err := fs.GetDiskSpace(m.rightPane.Path()); err == nil {
			m.rightDiskSpace = freeBytes
		}
	}

	m.lastDiskSpaceCheck = time.Now()
}

// startSearch は検索モードを開始する
func (m *Model) startSearch(mode SearchMode) {
	// 現在のフィルタ状態を保存（Esc時に復元するため）
	pane := m.getActivePane()
	if pane.IsFiltered() {
		m.searchState.PreviousResult = &SearchResult{
			Mode:    pane.FilterMode(),
			Pattern: pane.FilterPattern(),
		}
	} else {
		m.searchState.PreviousResult = nil
	}

	m.searchState.Mode = mode
	m.searchState.Pattern = ""
	m.searchState.IsActive = true

	// ミニバッファの設定
	if mode == SearchModeIncremental {
		m.minibuffer.SetPrompt("/: ")
	} else {
		m.minibuffer.SetPrompt("(search): ")
	}
	m.minibuffer.Clear()
	m.minibuffer.SetWidth(m.getActivePane().width)
	m.minibuffer.Show()
}

// confirmSearch は検索を確定する
func (m *Model) confirmSearch() {
	pattern := m.minibuffer.Input()
	pane := m.getActivePane()

	if pattern == "" {
		// 空のパターンでEnter→フィルタをクリア
		pane.ClearFilter()
		m.searchState.PreviousResult = nil
	} else {
		// パターンがある場合はフィルタを適用
		if err := pane.ApplyFilter(pattern, m.searchState.Mode); err != nil {
			// 正規表現エラーの場合はステータスバーに表示してミニバッファを維持
			m.statusMessage = fmt.Sprintf("Invalid regex: %v", err)
			m.isStatusError = true
			return
		}
		// 成功した場合、現在のフィルタを「前の結果」として保存
		m.searchState.PreviousResult = &SearchResult{
			Mode:    m.searchState.Mode,
			Pattern: pattern,
		}
	}

	// ミニバッファを閉じる
	m.minibuffer.Hide()
	m.searchState.IsActive = false
	m.searchState.Mode = SearchModeNone
}

// cancelSearch は検索をキャンセルする
func (m *Model) cancelSearch() {
	pane := m.getActivePane()

	// 前の検索結果があれば復元
	if m.searchState.PreviousResult != nil {
		pane.ApplyFilter(m.searchState.PreviousResult.Pattern, m.searchState.PreviousResult.Mode)
	} else {
		// 前の結果がなければフィルタをクリア
		pane.ClearFilter()
	}

	// ミニバッファを閉じる
	m.minibuffer.Hide()
	m.searchState.IsActive = false
	m.searchState.Mode = SearchModeNone
}

// applyIncrementalFilter はインクリメンタル検索のフィルタを適用する
func (m *Model) applyIncrementalFilter() {
	pattern := m.minibuffer.Input()
	pane := m.getActivePane()

	// インクリメンタル検索の場合は即座にフィルタを適用
	if m.searchState.Mode == SearchModeIncremental {
		pane.ApplyFilter(pattern, SearchModeIncremental)
	}
}

// RefreshBothPanes refreshes both panes
func (m *Model) RefreshBothPanes() tea.Cmd {
	var cmds []tea.Cmd

	// Refresh left pane
	if err := m.leftPane.Refresh(); err != nil {
		m.dialog = NewErrorDialog(fmt.Sprintf("Failed to refresh left pane: %v", err))
	}

	// Refresh right pane
	if err := m.rightPane.Refresh(); err != nil {
		m.dialog = NewErrorDialog(fmt.Sprintf("Failed to refresh right pane: %v", err))
	}

	// Update disk space
	m.updateDiskSpace()

	return tea.Batch(cmds...)
}

// SyncOppositePane synchronizes the opposite pane to the active pane's directory
func (m *Model) SyncOppositePane() {
	activePane := m.getActivePane()
	oppositePane := m.getInactivePane()

	if err := oppositePane.SyncTo(activePane.path); err != nil {
		m.dialog = NewErrorDialog(fmt.Sprintf("Failed to sync pane: %v", err))
	}
}

// handleCreateFile は新規ファイル作成を処理
func (m *Model) handleCreateFile(dirPath, filename string) tea.Cmd {
	return func() tea.Msg {
		// バリデーション
		if err := fs.ValidateFilename(filename); err != nil {
			return inputDialogResultMsg{
				operation: "create_file",
				err:       err,
			}
		}

		fullPath := filepath.Join(dirPath, filename)
		if err := fs.CreateFile(fullPath); err != nil {
			return inputDialogResultMsg{
				operation: "create_file",
				err:       err,
			}
		}

		return inputDialogResultMsg{
			operation: "create_file",
			input:     filename,
		}
	}
}

// handleCreateDirectory は新規ディレクトリ作成を処理
func (m *Model) handleCreateDirectory(dirPath, dirname string) tea.Cmd {
	return func() tea.Msg {
		// バリデーション
		if err := fs.ValidateFilename(dirname); err != nil {
			return inputDialogResultMsg{
				operation: "create_dir",
				err:       err,
			}
		}

		fullPath := filepath.Join(dirPath, dirname)
		if err := fs.CreateDirectory(fullPath); err != nil {
			return inputDialogResultMsg{
				operation: "create_dir",
				err:       err,
			}
		}

		return inputDialogResultMsg{
			operation: "create_dir",
			input:     dirname,
		}
	}
}

// handleRename はリネームを処理
func (m *Model) handleRename(dirPath, oldName, newName string) tea.Cmd {
	return func() tea.Msg {
		// バリデーション
		if err := fs.ValidateFilename(newName); err != nil {
			return inputDialogResultMsg{
				operation: "rename",
				err:       err,
			}
		}

		oldPath := filepath.Join(dirPath, oldName)
		if err := fs.Rename(oldPath, newName); err != nil {
			return inputDialogResultMsg{
				operation: "rename",
				err:       err,
			}
		}

		return inputDialogResultMsg{
			operation: "rename",
			input:     newName,
			oldName:   oldName,
		}
	}
}

// moveCursorToFile は作成されたファイルにカーソルを移動
func (m *Model) moveCursorToFile(filename string) {
	pane := m.getActivePane()

	// 隠しファイルで表示OFFの場合はカーソル移動しない
	if strings.HasPrefix(filename, ".") && !pane.showHidden {
		return
	}

	// ファイルを探してカーソルを移動
	for i, entry := range pane.entries {
		if entry.Name == filename {
			pane.cursor = i
			pane.EnsureCursorVisible()
			return
		}
	}
}

// moveCursorToFileAfterRename はリネーム後にカーソルを移動
func (m *Model) moveCursorToFileAfterRename(oldName, newName string) {
	pane := m.getActivePane()

	// 隠しファイルにリネームされ、表示OFFの場合
	if strings.HasPrefix(newName, ".") && !pane.showHidden {
		// 現在のカーソル位置が有効範囲を超えていたら調整
		if pane.cursor >= len(pane.entries) {
			if len(pane.entries) > 0 {
				pane.cursor = len(pane.entries) - 1
			} else {
				pane.cursor = 0
			}
		}
		pane.EnsureCursorVisible()
		return
	}

	// リネームされたファイルを探してカーソルを移動
	m.moveCursorToFile(newName)
}
