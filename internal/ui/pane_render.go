package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/sakura/duofm/internal/fs"
)

// View はペインをレンダリング（後方互換性のため）
func (p *Pane) View() string {
	return p.ViewWithDiskSpace(0)
}

// ViewWithDiskSpace はペインをレンダリング（ディスク容量情報付き）
func (p *Pane) ViewWithDiskSpace(diskSpace uint64) string {
	return p.viewInternal(diskSpace, nil)
}

// ViewWithMinibuffer はペインをレンダリング（ミニバッファ付き）
func (p *Pane) ViewWithMinibuffer(diskSpace uint64, minibuffer *Minibuffer) string {
	return p.viewInternal(diskSpace, minibuffer)
}

// viewInternal は内部レンダリング関数
func (p *Pane) viewInternal(diskSpace uint64, minibuffer *Minibuffer) string {
	var b strings.Builder

	// パス表示（ホームディレクトリは ~ に置換）
	displayPath := p.formatPath()
	// 隠しファイル表示中は [H] インジケーターを追加
	if p.showHidden {
		displayPath = "[H] " + displayPath
	}
	// フィルタ適用中はインジケーターを追加
	if p.IsFiltered() {
		filterIndicator := p.formatFilterIndicator()
		displayPath = filterIndicator + " " + displayPath
	}
	pathStyle := lipgloss.NewStyle().
		Width(p.width-2).
		Padding(0, 1).
		Bold(true)

	if p.isActive {
		pathStyle = pathStyle.Foreground(p.theme.PathFg)
	} else {
		pathStyle = pathStyle.Foreground(p.theme.PathFgInactive)
	}

	b.WriteString(pathStyle.Render(displayPath))
	b.WriteString("\n")

	// ヘッダー2行目（マーク情報と空き容量、またはローディング）
	headerLine2 := p.renderHeaderLine2(diskSpace)
	headerStyle := lipgloss.NewStyle().
		Width(p.width-2).
		Padding(0, 1)
	if p.isActive {
		headerStyle = headerStyle.Foreground(p.theme.HeaderFg)
	} else {
		headerStyle = headerStyle.Foreground(p.theme.HeaderFgInactive)
	}
	b.WriteString(headerStyle.Render(headerLine2))
	b.WriteString("\n")

	// 区切り線
	border := strings.Repeat("─", p.width-2)
	borderStyle := lipgloss.NewStyle().Padding(0, 1).Foreground(p.theme.BorderFg)
	b.WriteString(borderStyle.Render(border))
	b.WriteString("\n")

	// ファイルリスト（ミニバッファ表示時は1行少なく）
	visibleLines := p.height - 4 // ヘッダー2行 + ボーダー1行 = 3行
	if minibuffer != nil && minibuffer.IsVisible() {
		visibleLines-- // ミニバッファ分1行減らす
	}

	// フィルタ適用中で結果が空の場合
	if p.IsFiltered() && len(p.entries) == 0 {
		// "(No matches)" メッセージを表示
		noMatchStyle := lipgloss.NewStyle().
			Width(p.width-2).
			Padding(0, 1).
			Foreground(p.theme.DimmedFg).
			Italic(true)
		b.WriteString(noMatchStyle.Render("(No matches)"))
		b.WriteString("\n")

		// 残りを空行で埋める
		for i := 1; i < visibleLines; i++ {
			b.WriteString(strings.Repeat(" ", p.width))
			b.WriteString("\n")
		}
	} else {
		endIdx := p.scrollOffset + visibleLines
		if endIdx > len(p.entries) {
			endIdx = len(p.entries)
		}

		for i := p.scrollOffset; i < endIdx; i++ {
			entry := p.entries[i]
			line := p.formatEntry(entry, i == p.cursor)
			b.WriteString(line)
			b.WriteString("\n")
		}

		// 空行で埋める
		for i := endIdx - p.scrollOffset; i < visibleLines; i++ {
			b.WriteString(strings.Repeat(" ", p.width))
			b.WriteString("\n")
		}
	}

	// ミニバッファの表示
	if minibuffer != nil && minibuffer.IsVisible() {
		b.WriteString(minibuffer.View())
		b.WriteString("\n")
	}

	return b.String()
}

// formatPath はパスを表示用にフォーマット
func (p *Pane) formatPath() string {
	home, _ := fs.HomeDirectory()
	if strings.HasPrefix(p.path, home) {
		return "~" + strings.TrimPrefix(p.path, home)
	}
	return p.path
}

// renderHeaderLine2 はヘッダー2行目（マーク情報と空き容量）をレンダリング
func (p *Pane) renderHeaderLine2(diskSpace uint64) string {
	if p.loading {
		// ローディング中はローディングメッセージを表示
		return p.loadingProgress
	}

	// マーク情報を計算
	markInfo := p.CalculateMarkInfo()
	markedCount := markInfo.Count
	markedSize := markInfo.TotalSize

	// フィルタ適用中は "Marked 0/5 (15) 0 B" 形式（5=フィルタ後、15=フィルタ前）
	// 通常は "Marked 0/15 0 B" 形式
	var markedInfo string
	if p.IsFiltered() {
		filteredCount := p.FilteredEntryCount()
		totalCount := p.TotalEntryCount()
		markedInfo = fmt.Sprintf("Marked %d/%d (%d) %s", markedCount, filteredCount, totalCount, FormatSize(markedSize))
	} else {
		totalCount := p.TotalEntryCount()
		markedInfo = fmt.Sprintf("Marked %d/%d %s", markedCount, totalCount, FormatSize(markedSize))
	}

	// 空き容量情報
	freeInfo := ""
	if diskSpace > 0 {
		freeInfo = fmt.Sprintf("%s Free", FormatSize(int64(diskSpace)))
	}

	// レイアウト: 左にマーク情報、右に空き容量
	availableWidth := p.width - 4 // パディングを考慮
	markedLen := runewidth.StringWidth(markedInfo)
	freeLen := runewidth.StringWidth(freeInfo)
	padding := availableWidth - markedLen - freeLen
	if padding < 1 {
		padding = 1
	}

	return markedInfo + strings.Repeat(" ", padding) + freeInfo
}

// formatEntry はエントリを1行にフォーマット
func (p *Pane) formatEntry(entry fs.FileEntry, isCursor bool) string {
	mode := p.GetEffectiveDisplayMode()
	nameWidth, _ := CalculateColumnWidths(p.width)

	var line string

	switch mode {
	case DisplayMinimal:
		// 名前のみ
		line = p.formatMinimalEntry(entry, nameWidth)

	case DisplayBasic:
		// 名前 + サイズ + タイムスタンプ
		line = p.formatBasicEntry(entry, nameWidth)

	case DisplayDetail:
		// 名前 + パーミッション + 所有者 + グループ
		line = p.formatDetailEntry(entry, nameWidth)
	}

	// スタイルを適用
	style := lipgloss.NewStyle().
		Width(p.width-2).
		Padding(0, 1)

	isMarked := p.IsMarked(entry.Name)

	// 4つの状態を処理: 通常、カーソルのみ、マークのみ、カーソル+マーク
	if isCursor && isMarked {
		// Cursor + Mark combined
		if p.isActive {
			style = style.Background(p.theme.CursorMarkBg).
				Foreground(p.theme.CursorMarkFg)
		} else {
			style = style.Background(p.theme.CursorMarkBgInactive).
				Foreground(p.theme.CursorMarkFg)
		}
	} else if isCursor {
		// Cursor only
		if p.isActive {
			style = style.Background(p.theme.CursorBg).
				Foreground(p.theme.CursorFg)
		} else {
			style = style.Background(p.theme.CursorBgInactive).
				Foreground(p.theme.CursorFg)
		}
	} else if isMarked {
		// Marked only
		if p.isActive {
			style = style.Background(p.theme.MarkBg).
				Foreground(p.theme.MarkFg)
		} else {
			style = style.Background(p.theme.MarkBgInactive).
				Foreground(p.theme.MarkFgInactive)
		}
	} else {
		// Normal - ファイルタイプによる色付け
		if entry.IsSymlink {
			if entry.LinkBroken {
				style = style.Foreground(p.theme.ExecutableFg) // 赤色（壊れたリンク）
			} else {
				style = style.Foreground(p.theme.SymlinkFg) // シアン色
			}
		} else if entry.IsDir {
			style = style.Foreground(p.theme.DirectoryFg) // 青色
		}
	}

	return style.Render(line)
}

// formatMinimalEntry は名前のみのエントリをフォーマット
func (p *Pane) formatMinimalEntry(entry fs.FileEntry, nameWidth int) string {
	return entry.DisplayNameWithLimit(nameWidth)
}

// formatBasicEntry は基本情報（名前 + サイズ + タイムスタンプ）をフォーマット
func (p *Pane) formatBasicEntry(entry fs.FileEntry, nameWidth int) string {
	// ファイル名
	name := entry.DisplayNameWithLimit(nameWidth)

	// サイズ
	var sizeStr string
	if entry.IsSymlink && entry.LinkBroken {
		sizeStr = "?"
	} else if entry.IsDir {
		sizeStr = "-"
	} else {
		sizeStr = FormatSize(entry.Size)
	}

	// タイムスタンプ
	timestamp := FormatTimestamp(entry.ModTime)

	// カラムを組み立て
	// 名前幅を確保（nameWidthまで）
	namePadding := nameWidth - runewidth.StringWidth(name)
	if namePadding < 0 {
		namePadding = 0
	}

	// サイズは右揃えで10文字
	sizePadded := fmt.Sprintf("%10s", sizeStr)

	return fmt.Sprintf("%s%s  %s  %s", name, strings.Repeat(" ", namePadding), sizePadded, timestamp)
}

// formatDetailEntry は詳細情報（名前 + パーミッション + 所有者 + グループ）をフォーマット
func (p *Pane) formatDetailEntry(entry fs.FileEntry, nameWidth int) string {
	// ファイル名
	name := entry.DisplayNameWithLimit(nameWidth)

	// パーミッション
	perms := FormatPermissions(entry.Permissions)

	// 所有者とグループ
	owner := entry.Owner
	if len(owner) > 10 {
		owner = owner[:10]
	}

	group := entry.Group
	if len(group) > 10 {
		group = group[:10]
	}

	// カラムを組み立て
	// 名前幅を確保（nameWidthまで）
	namePadding := nameWidth - runewidth.StringWidth(name)
	if namePadding < 0 {
		namePadding = 0
	}

	// 所有者とグループを左揃えで各10文字
	ownerPadded := fmt.Sprintf("%-10s", owner)
	groupPadded := fmt.Sprintf("%-10s", group)

	return fmt.Sprintf("%s%s  %s  %s  %s", name, strings.Repeat(" ", namePadding), perms, ownerPadded, groupPadded)
}

// ViewDimmedWithDiskSpace はダイアログオーバーレイ用にdimmedスタイルでペインをレンダリング
func (p *Pane) ViewDimmedWithDiskSpace(diskSpace uint64) string {
	var b strings.Builder

	// パス表示（暗いスタイル）
	displayPath := p.formatPath()
	// 隠しファイル表示中は [H] インジケーターを追加
	if p.showHidden {
		displayPath = "[H] " + displayPath
	}
	pathStyle := lipgloss.NewStyle().
		Width(p.width-2).
		Padding(0, 1).
		Bold(true).
		Background(p.theme.DimmedBg).
		Foreground(p.theme.DimmedFg)

	b.WriteString(pathStyle.Render(displayPath))
	b.WriteString("\n")

	// ヘッダー2行目（マーク情報と空き容量）
	headerLine2 := p.renderHeaderLine2(diskSpace)
	headerStyle := lipgloss.NewStyle().
		Width(p.width-2).
		Padding(0, 1).
		Background(p.theme.DimmedBg).
		Foreground(p.theme.DimmedFg)
	b.WriteString(headerStyle.Render(headerLine2))
	b.WriteString("\n")

	// 区切り線
	border := strings.Repeat("─", p.width-2)
	borderStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Background(p.theme.DimmedBg).
		Foreground(p.theme.DimmedFg)
	b.WriteString(borderStyle.Render(border))
	b.WriteString("\n")

	// ファイルリスト
	visibleLines := p.height - 4 // ヘッダー2行 + ボーダー1行 = 3行
	endIdx := p.scrollOffset + visibleLines
	if endIdx > len(p.entries) {
		endIdx = len(p.entries)
	}

	for i := p.scrollOffset; i < endIdx; i++ {
		entry := p.entries[i]
		line := p.formatEntryDimmed(entry)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// 空行で埋める（dimmedスタイル）
	emptyStyle := lipgloss.NewStyle().
		Width(p.width).
		Background(p.theme.DimmedBg)
	for i := endIdx - p.scrollOffset; i < visibleLines; i++ {
		b.WriteString(emptyStyle.Render(""))
		b.WriteString("\n")
	}

	return b.String()
}

// formatEntryDimmed はエントリをdimmedスタイルで1行にフォーマット
func (p *Pane) formatEntryDimmed(entry fs.FileEntry) string {
	mode := p.GetEffectiveDisplayMode()
	nameWidth, _ := CalculateColumnWidths(p.width)

	var line string

	switch mode {
	case DisplayMinimal:
		line = p.formatMinimalEntry(entry, nameWidth)
	case DisplayBasic:
		line = p.formatBasicEntry(entry, nameWidth)
	case DisplayDetail:
		line = p.formatDetailEntry(entry, nameWidth)
	}

	// dimmedスタイルを適用
	style := lipgloss.NewStyle().
		Width(p.width-2).
		Padding(0, 1).
		Background(p.theme.DimmedBg).
		Foreground(p.theme.DimmedFg)

	// マークされたファイルは薄いハイライトで表示
	if p.IsMarked(entry.Name) {
		style = style.Background(lipgloss.Color("58")) // Dim yellow-ish background
	}

	return style.Render(line)
}
