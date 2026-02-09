package ui

import (
	"fmt"
	"path/filepath"
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

	// ヘッダー1行目: パス + Gitブランチ
	headerLine1 := p.renderHeaderLine1()
	pathStyle := lipgloss.NewStyle().
		Width(p.width-2).
		Padding(0, 1).
		Bold(true)

	if p.isActive {
		pathStyle = pathStyle.Foreground(p.theme.PathFg)
	} else {
		pathStyle = pathStyle.Foreground(p.theme.PathFgInactive)
	}

	b.WriteString(pathStyle.Render(headerLine1))
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
	visibleLines := p.getVisibleLines()
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

// renderHeaderLine1 はヘッダー1行目（パス + Gitブランチ）をレンダリング
func (p *Pane) renderHeaderLine1() string {
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

	// ブランチ表示がない場合はパスのみ返す
	if p.gitBranch == "" {
		return displayPath
	}

	// ブランチ表示を構築 [branch]
	branchDisplay := "[" + p.gitBranch + "]"

	// 利用可能な幅を計算（パディング分を除く）
	availableWidth := p.width - 4 // 左右パディング(1+1) + 余白

	pathWidth := runewidth.StringWidth(displayPath)
	branchWidth := runewidth.StringWidth(branchDisplay)

	// パスとブランチの間に最低1スペース必要
	minTotalWidth := pathWidth + 1 + branchWidth

	if minTotalWidth <= availableWidth {
		// 十分なスペースがある場合: パス + パディング + ブランチ
		padding := availableWidth - pathWidth - branchWidth
		return displayPath + strings.Repeat(" ", padding) + branchDisplay
	}

	// スペースが不足する場合: パスを切り詰めてブランチを優先
	// 最小限必要: ブランチ幅 + 1スペース + 切り詰め記号(...)分
	maxPathWidth := availableWidth - branchWidth - 1
	if maxPathWidth < 4 {
		// パス表示スペースが極端に狭い場合はブランチのみ
		return branchDisplay
	}

	// パスを切り詰め
	truncatedPath := truncateStringWithEllipsis(displayPath, maxPathWidth)
	truncatedPathWidth := runewidth.StringWidth(truncatedPath)
	padding := availableWidth - truncatedPathWidth - branchWidth
	if padding < 1 {
		padding = 1
	}

	return truncatedPath + strings.Repeat(" ", padding) + branchDisplay
}

// truncateStringWithEllipsis は文字列を指定幅に切り詰め、省略記号(...)を追加する
// 文字列が指定幅に収まる場合は切り詰めずにそのまま返す
func truncateStringWithEllipsis(s string, maxWidth int) string {
	// 文字列の実際の表示幅を計算
	stringWidth := runewidth.StringWidth(s)

	// 文字列が指定幅に収まる場合は切り詰め不要
	if stringWidth <= maxWidth {
		return s
	}

	// 極端に狭い場合は省略記号のみ
	if maxWidth <= 3 {
		return "..."[:maxWidth]
	}

	currentWidth := 0
	var result strings.Builder
	ellipsis := "..."
	ellipsisWidth := 3

	targetWidth := maxWidth - ellipsisWidth

	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if currentWidth+rw > targetWidth {
			break
		}
		result.WriteRune(r)
		currentWidth += rw
	}

	result.WriteString(ellipsis)
	return result.String()
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

	// ソート情報
	sortInfo := p.sortConfig.String()

	// レイアウト: 左にマーク情報、中央にソート情報、右に空き容量
	availableWidth := p.width - 4 // パディングを考慮
	markedLen := runewidth.StringWidth(markedInfo)
	sortLen := runewidth.StringWidth(sortInfo)
	freeLen := runewidth.StringWidth(freeInfo)

	totalContentWidth := markedLen + sortLen + freeLen
	remainingSpace := availableWidth - totalContentWidth
	if remainingSpace < 2 {
		// スペースが足りない場合は最小限のパディング
		return markedInfo + " " + sortInfo + " " + freeInfo
	}

	// ソート情報を中央に配置するため、左右のパディングを均等に分配
	leftPad := remainingSpace / 2
	rightPad := remainingSpace - leftPad

	return markedInfo + strings.Repeat(" ", leftPad) + sortInfo + strings.Repeat(" ", rightPad) + freeInfo
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
// ゴミ箱内では削除日時と元パスのディレクトリ部分を表示
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

	// カラムを組み立て
	// 名前幅を確保（nameWidthまで）
	namePadding := nameWidth - runewidth.StringWidth(name)
	if namePadding < 0 {
		namePadding = 0
	}

	// サイズは右揃えで10文字
	sizePadded := fmt.Sprintf("%10s", sizeStr)

	// ゴミ箱内では削除日時と元パスのディレクトリ部分を表示
	if p.IsInTrash() && !entry.IsParentDir() {
		return p.formatTrashEntry(name, namePadding, sizePadded, entry.Name)
	}

	// タイムスタンプ
	timestamp := FormatTimestamp(entry.ModTime)

	return fmt.Sprintf("%s%s  %s  %s", name, strings.Repeat(" ", namePadding), sizePadded, timestamp)
}

// formatTrashEntry はゴミ箱内でのエントリをフォーマット（削除日時 + 元パス）
func (p *Pane) formatTrashEntry(name string, namePadding int, sizePadded, trashName string) string {
	// trashinfoから情報を取得
	info, err := fs.GetTrashItemInfo(trashName)
	if err != nil {
		// trashinfoがない場合は通常表示
		return fmt.Sprintf("%s%s  %s  ?", name, strings.Repeat(" ", namePadding), sizePadded)
	}

	// 削除日時
	deletedTime := FormatTimestamp(info.DeletionDate)

	// 元パスのディレクトリ部分（ファイル名は除く）
	originalDir := filepath.Dir(info.OriginalPath)
	// ホームディレクトリを ~ に置換
	home, _ := fs.HomeDirectory()
	if strings.HasPrefix(originalDir, home) {
		originalDir = "~" + strings.TrimPrefix(originalDir, home)
	}

	// 幅が足りない場合は切り詰め
	maxDirWidth := 30
	if runewidth.StringWidth(originalDir) > maxDirWidth {
		originalDir = truncateStringWithEllipsis(originalDir, maxDirWidth)
	}

	return fmt.Sprintf("%s%s  %s  %s  %s", name, strings.Repeat(" ", namePadding), sizePadded, deletedTime, originalDir)
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

	// ヘッダー1行目（パス + Gitブランチ、暗いスタイル）
	headerLine1 := p.renderHeaderLine1()
	pathStyle := lipgloss.NewStyle().
		Width(p.width-2).
		Padding(0, 1).
		Bold(true).
		Background(p.theme.DimmedBg).
		Foreground(p.theme.DimmedFg)

	b.WriteString(pathStyle.Render(headerLine1))
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
