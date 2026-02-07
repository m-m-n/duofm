package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/sakura/duofm/internal/config"
	"github.com/sakura/duofm/internal/fs"
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
	leftSpace := m.diskSpaceMonitor.LeftSpace()
	rightSpace := m.diskSpaceMonitor.RightSpace()

	var leftView, rightView string
	if m.searchState.IsActive || m.shellCommandMode {
		if m.activePane == LeftPane {
			leftView = m.leftPane.ViewWithMinibuffer(leftSpace, m.minibuffer)
			rightView = m.rightPane.ViewWithDiskSpace(rightSpace)
		} else {
			leftView = m.leftPane.ViewWithDiskSpace(leftSpace)
			rightView = m.rightPane.ViewWithMinibuffer(rightSpace, m.minibuffer)
		}
	} else {
		leftView = m.leftPane.ViewWithDiskSpace(leftSpace)
		rightView = m.rightPane.ViewWithDiskSpace(rightSpace)
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
	leftSpace := m.diskSpaceMonitor.LeftSpace()
	rightSpace := m.diskSpaceMonitor.RightSpace()
	leftView := m.leftPane.ViewDimmedWithDiskSpace(leftSpace)
	rightView := m.rightPane.ViewDimmedWithDiskSpace(rightSpace)
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

	leftSpace := m.diskSpaceMonitor.LeftSpace()
	rightSpace := m.diskSpaceMonitor.RightSpace()

	var leftView, rightView string

	if m.activePane == LeftPane {
		// 左ペインをdimmedで描画してダイアログをオーバーレイ
		dimmedLeft := m.leftPane.ViewDimmedWithDiskSpace(leftSpace)
		leftView = m.overlayDialogOnPane(dimmedLeft, paneWidth, paneHeight)
		rightView = m.rightPane.ViewWithDiskSpace(rightSpace)
	} else {
		// 右ペインをdimmedで描画してダイアログをオーバーレイ
		leftView = m.leftPane.ViewWithDiskSpace(leftSpace)
		dimmedRight := m.rightPane.ViewDimmedWithDiskSpace(rightSpace)
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

	leftSpace := m.diskSpaceMonitor.LeftSpace()
	rightSpace := m.diskSpaceMonitor.RightSpace()

	var leftView, rightView string

	if m.activePane == LeftPane {
		dimmedLeft := m.leftPane.ViewDimmedWithDiskSpace(leftSpace)
		leftView = m.overlaySortDialogOnPane(dimmedLeft, paneWidth, paneHeight)
		rightView = m.rightPane.ViewWithDiskSpace(rightSpace)
	} else {
		leftView = m.leftPane.ViewWithDiskSpace(leftSpace)
		dimmedRight := m.rightPane.ViewDimmedWithDiskSpace(rightSpace)
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
			// ダイアログ行をペイン幅に調整（切り詰めまたはパディング）
			adjusted := truncateOrPadLineToWidth(dialogLine, paneWidth)
			result.WriteString(adjusted)
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
			// ダイアログ行をペイン幅に調整（切り詰めまたはパディング）
			adjusted := truncateOrPadLineToWidth(dialogLine, paneWidth)
			result.WriteString(adjusted)
		}
		if i < len(paneLines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// truncateOrPadLineToWidth は文字列を指定した表示幅に調整する
// ANSIエスケープシーケンスを考慮し、表示幅のみをカウントする
// 幅を超える場合は切り詰め、足りない場合はスペースでパディングする
func truncateOrPadLineToWidth(s string, targetWidth int) string {
	var result strings.Builder
	currentWidth := 0
	i := 0
	runes := []rune(s)

	for i < len(runes) {
		// ANSIエスケープシーケンスの検出（\x1b[...m）
		if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			// エスケープシーケンスの終わりを探す
			escStart := i
			i += 2 // \x1b[ をスキップ
			for i < len(runes) && runes[i] != 'm' {
				i++
			}
			if i < len(runes) {
				i++ // 'm' をスキップ
			}
			// エスケープシーケンス全体を結果に追加（幅はカウントしない）
			result.WriteString(string(runes[escStart:i]))
			continue
		}

		// 通常の文字
		r := runes[i]
		rw := runewidth.RuneWidth(r)

		// 幅を超える場合は終了
		if currentWidth+rw > targetWidth {
			break
		}

		result.WriteRune(r)
		currentWidth += rw
		i++
	}

	// 幅が足りない場合はスペースでパディング
	if currentWidth < targetWidth {
		result.WriteString(strings.Repeat(" ", targetWidth-currentWidth))
	}

	return result.String()
}

// entryTypeLabel はエントリの種別ラベルと表示色を返す。
// 判定優先順位: シンボリックリンク > ディレクトリ > 通常ファイル（MIMEタイプ）
func (m Model) entryTypeLabel(entry *fs.FileEntry) (label string, fg lipgloss.Color) {
	if entry == nil {
		return "", m.theme.StatusFg
	}
	switch {
	case entry.IsSymlink:
		return "SymbolicLink", m.theme.SymlinkFg
	case entry.IsDir:
		return "Directory", m.theme.DirectoryFg
	default:
		return config.GetMIMEType(entry.Name), m.theme.StatusFg
	}
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

	// MIMEタイプ / 種別情報
	typeInfo := ""
	typeInfoWidth := 0
	if entry := activePane.SelectedEntry(); entry != nil {
		label, fg := m.entryTypeLabel(entry)
		labelText := fmt.Sprintf(" [%s]", label)
		typeInfo = lipgloss.NewStyle().
			Foreground(fg).
			Background(m.theme.StatusBg).
			Render(labelText)
		typeInfoWidth = runewidth.StringWidth(labelText)
	}

	// キーヒント（動的に変更）
	hints := "?:help q:quit"
	if activePane != nil && activePane.CanToggleMode() {
		hints = "i:info " + hints
	}

	// スペースで埋める（typeInfoWidth を加算）
	padding := m.width - runewidth.StringWidth(posInfo) - typeInfoWidth - runewidth.StringWidth(hints) - 4
	if padding < 0 {
		// 幅が足りない場合はMIMEタイプ表示を省略
		typeInfo = ""
		padding = m.width - runewidth.StringWidth(posInfo) - runewidth.StringWidth(hints) - 4
		if padding < 0 {
			padding = 0
		}
	}

	// ステータスバーの組み立て
	// typeInfo は既に lipgloss でレンダリング済み（ANSI色付き）なので、
	// 全体を style.Render() に渡すと色が上書きされる。
	// そのため、左部分・typeInfo・右部分を個別にレンダリングして結合する。
	statusBg := m.theme.StatusBg
	statusFg := m.theme.StatusFg

	leftPart := lipgloss.NewStyle().
		Background(statusBg).
		Foreground(statusFg).
		Render(fmt.Sprintf(" %s", posInfo))

	rightPart := lipgloss.NewStyle().
		Background(statusBg).
		Foreground(statusFg).
		Render(fmt.Sprintf("%s%s ", strings.Repeat(" ", padding), hints))

	statusBar := leftPart + typeInfo + rightPart

	return lipgloss.NewStyle().
		Width(m.width).
		Render(statusBar)
}
