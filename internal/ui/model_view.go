package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/sakura/duofm/internal/version"
)

// View はUIをレンダリング
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// タイトルバー
	title := titleStyle.Render("duofm " + version.Version)

	// 2つのペインを横に並べる（ディスク容量情報付き）
	// 検索モードまたはシェルコマンドモードの場合はアクティブペインにミニバッファを渡す
	var leftView, rightView string
	if m.searchState.IsActive || m.shellCommandMode {
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

	// ソートダイアログがある場合はペインローカル表示
	if m.sortDialog != nil && m.sortDialog.IsActive() {
		return m.renderSortDialogPane()
	}

	return mainView
}

// renderDialogScreen は画面全体表示ダイアログをレンダリング（両ペインdimmed）
func (m Model) renderDialogScreen() string {
	// タイトルバー
	title := titleStyle.Render("duofm " + version.Version)

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
	title := titleStyle.Render("duofm " + version.Version)

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

// renderSortDialogPane はソートダイアログをペインローカル表示
func (m Model) renderSortDialogPane() string {
	paneWidth := m.width / 2
	paneHeight := m.height - 2

	title := titleStyle.Render("duofm " + version.Version)
	statusBar := m.renderStatusBar()

	var leftView, rightView string

	if m.activePane == LeftPane {
		dimmedLeft := m.leftPane.ViewDimmedWithDiskSpace(m.leftDiskSpace)
		leftView = m.overlaySortDialogOnPane(dimmedLeft, paneWidth, paneHeight)
		rightView = m.rightPane.ViewWithDiskSpace(m.rightDiskSpace)
	} else {
		leftView = m.leftPane.ViewWithDiskSpace(m.leftDiskSpace)
		dimmedRight := m.rightPane.ViewDimmedWithDiskSpace(m.rightDiskSpace)
		rightView = m.overlaySortDialogOnPane(dimmedRight, paneWidth, paneHeight)
	}

	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
	return lipgloss.JoinVertical(lipgloss.Left, title, panes, statusBar)
}

// overlaySortDialogOnPane はdimmedペインの上にソートダイアログをオーバーレイ
func (m Model) overlaySortDialogOnPane(dimmedPane string, paneWidth, paneHeight int) string {
	dialogView := lipgloss.Place(
		paneWidth,
		paneHeight,
		lipgloss.Center,
		lipgloss.Center,
		m.sortDialog.View(),
		lipgloss.WithWhitespaceBackground(dimmedBgColor),
	)

	paneLines := strings.Split(dimmedPane, "\n")
	dialogLines := strings.Split(dialogView, "\n")

	var result strings.Builder
	for i := 0; i < len(paneLines) && i < len(dialogLines); i++ {
		dialogLine := dialogLines[i]
		paneLine := paneLines[i]

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
		if runewidth.StringWidth(msg) > maxLen {
			msg = runewidth.Truncate(msg, maxLen-3, "") + "..."
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
	padding := m.width - runewidth.StringWidth(posInfo) - runewidth.StringWidth(hints) - 4
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
