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
		leftPane:   nil, // Updateで初期化
		rightPane:  nil, // Updateで初期化
		leftPath:   cwd,
		rightPath:  home,
		activePane: LeftPane,
		dialog:     nil,
		ready:      false,
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

	case directoryLoadCompleteMsg:
		// ディレクトリ読み込み完了
		var targetPane *Pane
		if msg.panePath == m.leftPane.Path() {
			targetPane = m.leftPane
		} else if msg.panePath == m.rightPane.Path() {
			targetPane = m.rightPane
		}

		if targetPane != nil {
			if msg.err != nil {
				// エラーダイアログを表示
				m.dialog = NewErrorDialog(fmt.Sprintf("Failed to read directory: %v", msg.err))
				targetPane.loading = false
				targetPane.loadingProgress = ""
			} else {
				// エントリを更新
				targetPane.entries = msg.entries
				targetPane.cursor = 0
				targetPane.scrollOffset = 0
				targetPane.loading = false
				targetPane.loadingProgress = ""
			}
		}
		return m, nil

	case tea.KeyMsg:
		// ダイアログが開いている場合はダイアログに処理を委譲
		if m.dialog != nil {
			var cmd tea.Cmd
			m.dialog, cmd = m.dialog.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", KeyQuit:
			return m, tea.Quit

		case KeyHelp:
			// ヘルプダイアログを表示
			m.dialog = NewHelpDialog()
			return m, nil

		case KeyMoveDown, KeyArrowDown:
			m.getActivePane().MoveCursorDown()

		case KeyMoveUp, KeyArrowUp:
			m.getActivePane().MoveCursorUp()

		case KeyMoveLeft, KeyArrowLeft:
			if m.activePane == LeftPane {
				// 左ペインで h/← -> 親ディレクトリへ
				m.leftPane.MoveToParent()
			} else {
				// 右ペインで h/← -> 左ペインへ切り替え
				m.switchToPane(LeftPane)
			}

		case KeyMoveRight, KeyArrowRight:
			if m.activePane == RightPane {
				// 右ペインで l/→ -> 親ディレクトリへ
				m.rightPane.MoveToParent()
			} else {
				// 左ペインで l/→ -> 右ペインへ切り替え
				m.switchToPane(RightPane)
			}

		case KeyEnter:
			m.getActivePane().EnterDirectory()
			// ディレクトリ移動時にディスク容量を更新
			m.updateDiskSpace()

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
	panes := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.leftPane.ViewWithDiskSpace(m.leftDiskSpace),
		m.rightPane.ViewWithDiskSpace(m.rightDiskSpace),
	)

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
