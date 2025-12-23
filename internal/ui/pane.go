package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/fs"
)

// DisplayMode は表示モードを表す
type DisplayMode int

const (
	// DisplayMinimal は名前のみを表示（端末幅が狭い場合に自動）
	DisplayMinimal DisplayMode = iota
	// DisplayBasic は名前 + サイズ + タイムスタンプを表示（デフォルト）
	DisplayBasic
	// DisplayDetail は名前 + パーミッション + 所有者 + グループを表示
	DisplayDetail
)

// ダイアログオーバーレイ用のdimmedスタイル色
var (
	dimmedBgColor = lipgloss.Color("236") // 濃いグレー背景
	dimmedFgColor = lipgloss.Color("243") // 暗いテキスト
)

// Pane は1つのファイルリストペインを表現
type Pane struct {
	path            string
	entries         []fs.FileEntry // フィルタ適用後の表示用エントリ
	allEntries      []fs.FileEntry // フィルタ適用前のすべてのエントリ
	cursor          int
	scrollOffset    int
	width           int
	height          int
	isActive        bool
	displayMode     DisplayMode // ユーザーが選択した表示モード
	loading         bool        // ローディング中フラグ
	loadingProgress string      // ローディングメッセージ
	showHidden      bool        // 隠しファイル表示フラグ（デフォルト: false）
	previousPath    string      // 直前のディレクトリパス（履歴なしの場合は空文字列）
	pendingPath     string      // 読み込み中の暫定パス（エラー時に元に戻す）
	filterPattern   string      // 現在のフィルタパターン（空の場合はフィルタなし）
	filterMode      SearchMode  // 現在のフィルタモード
}

// NewPane は新しいペインを作成
func NewPane(path string, width, height int, isActive bool) (*Pane, error) {
	pane := &Pane{
		path:            path,
		width:           width,
		height:          height,
		isActive:        isActive,
		cursor:          0,
		scrollOffset:    0,
		displayMode:     DisplayBasic, // デフォルトは基本情報モード
		loading:         false,
		loadingProgress: "",
	}

	if err := pane.LoadDirectory(); err != nil {
		return nil, err
	}

	return pane, nil
}

// LoadDirectory はディレクトリを読み込む（同期版）
func (p *Pane) LoadDirectory() error {
	entries, err := fs.ReadDirectory(p.path)
	if err != nil {
		return err
	}

	fs.SortEntries(entries)

	// 隠しファイルをフィルタリング
	if !p.showHidden {
		entries = filterHiddenFiles(entries)
	}

	// allEntriesにすべてのエントリを保存
	p.allEntries = entries
	// フィルタをクリアして全エントリを表示
	p.entries = entries
	p.filterPattern = ""
	p.filterMode = SearchModeNone
	p.cursor = 0
	p.scrollOffset = 0

	return nil
}

// filterHiddenFiles は隠しファイル（.で始まるファイル）を除外する
// ただし親ディレクトリ（..）は常に表示する
func filterHiddenFiles(entries []fs.FileEntry) []fs.FileEntry {
	result := make([]fs.FileEntry, 0, len(entries))
	for _, e := range entries {
		// 親ディレクトリは常に表示
		if e.IsParentDir() || !strings.HasPrefix(e.Name, ".") {
			result = append(result, e)
		}
	}
	return result
}

// StartLoadingDirectory はローディング状態を開始
func (p *Pane) StartLoadingDirectory() {
	p.loading = true
	p.loadingProgress = "Loading directory..."
}

// LoadDirectoryAsync は非同期でディレクトリを読み込む
func LoadDirectoryAsync(panePath string) tea.Cmd {
	return func() tea.Msg {
		entries, err := fs.ReadDirectory(panePath)
		if err != nil {
			return directoryLoadCompleteMsg{
				panePath:      panePath,
				entries:       nil,
				err:           err,
				attemptedPath: panePath,
			}
		}

		fs.SortEntries(entries)
		return directoryLoadCompleteMsg{
			panePath:      panePath,
			entries:       entries,
			err:           nil,
			attemptedPath: panePath,
		}
	}
}

// MoveCursorDown はカーソルを下に移動
func (p *Pane) MoveCursorDown() {
	if p.cursor < len(p.entries)-1 {
		p.cursor++
		p.adjustScroll()
	}
}

// MoveCursorUp はカーソルを上に移動
func (p *Pane) MoveCursorUp() {
	if p.cursor > 0 {
		p.cursor--
		p.adjustScroll()
	}
}

// adjustScroll はスクロール位置を調整
func (p *Pane) adjustScroll() {
	visibleLines := p.height - 4 // ヘッダー2行 + ボーダー1行 = 3行を除く

	// カーソルが表示範囲外に出たらスクロール
	if p.cursor < p.scrollOffset {
		p.scrollOffset = p.cursor
	} else if p.cursor >= p.scrollOffset+visibleLines {
		p.scrollOffset = p.cursor - visibleLines + 1
	}
}

// EnsureCursorVisible はカーソルが表示範囲内に収まるようスクロールを調整
func (p *Pane) EnsureCursorVisible() {
	p.adjustScroll()
}

// SelectedEntry は選択中のエントリを返す
func (p *Pane) SelectedEntry() *fs.FileEntry {
	if p.cursor < 0 || p.cursor >= len(p.entries) {
		return nil
	}
	return &p.entries[p.cursor]
}

// recordPreviousPath はナビゲーション前に現在のパスを記録する
func (p *Pane) recordPreviousPath() {
	p.previousPath = p.path
}

// restorePreviousPath は読み込み失敗時に前のパスに復元する
func (p *Pane) restorePreviousPath() {
	if p.previousPath != "" {
		p.path = p.previousPath
		p.pendingPath = ""
	}
}

// EnterDirectoryAsync はディレクトリへの移動を開始し、Cmdを返す
func (p *Pane) EnterDirectoryAsync() tea.Cmd {
	entry := p.SelectedEntry()
	if entry == nil {
		return nil
	}

	// シンボリックリンクの処理
	if entry.IsSymlink {
		if entry.LinkBroken {
			// リンク切れの場合は何もしない
			return nil
		}

		// リンク先がディレクトリかチェック
		isDir, err := fs.IsDirectory(entry.LinkTarget)
		if err != nil || !isDir {
			// リンク先がファイルまたはエラーの場合は何もしない
			return nil
		}
	}

	// 通常のディレクトリ処理
	if !entry.IsDir && !entry.IsSymlink {
		return nil // ファイルの場合は何もしない
	}

	var newPath string
	if entry.IsParentDir() {
		// 親ディレクトリに移動
		newPath = filepath.Dir(p.path)
	} else {
		// サブディレクトリに移動（シンボリックリンク含む）
		newPath = filepath.Join(p.path, entry.Name)
	}

	// 現在のパスを記録（復元用）
	p.recordPreviousPath()
	p.pendingPath = newPath
	p.path = newPath

	// ローディング状態を開始
	p.StartLoadingDirectory()

	return LoadDirectoryAsync(newPath)
}

// EnterDirectory はディレクトリに入る
func (p *Pane) EnterDirectory() error {
	entry := p.SelectedEntry()
	if entry == nil {
		return nil
	}

	// シンボリックリンクの処理
	if entry.IsSymlink {
		if entry.LinkBroken {
			// リンク切れの場合は何もしない
			return nil
		}

		// リンク先がディレクトリかチェック
		isDir, err := fs.IsDirectory(entry.LinkTarget)
		if err != nil || !isDir {
			// リンク先がファイルまたはエラーの場合は何もしない
			return nil
		}

		// 直前のパスを記録してから論理パス（シンボリックリンク自体のパス）に移動
		// これにより、..で論理的な親ディレクトリに戻れる
		p.recordPreviousPath()
		p.path = filepath.Join(p.path, entry.Name)
		return p.LoadDirectory()
	}

	// 通常のディレクトリ処理
	if !entry.IsDir {
		return nil // ファイルの場合は何もしない
	}

	var newPath string
	if entry.IsParentDir() {
		// 親ディレクトリに移動
		newPath = filepath.Dir(p.path)
	} else {
		// サブディレクトリに移動
		newPath = filepath.Join(p.path, entry.Name)
	}

	// 直前のパスを記録
	p.recordPreviousPath()
	p.path = newPath
	return p.LoadDirectory()
}

// MoveToParent は親ディレクトリに移動
func (p *Pane) MoveToParent() error {
	if p.path == "/" {
		return nil // ルートより上には行けない
	}

	p.recordPreviousPath()
	p.path = filepath.Dir(p.path)
	return p.LoadDirectory()
}

// MoveToParentAsync は親ディレクトリへの移動を開始
func (p *Pane) MoveToParentAsync() tea.Cmd {
	if p.path == "/" {
		return nil
	}
	newPath := filepath.Dir(p.path)
	p.recordPreviousPath()
	p.pendingPath = newPath
	p.path = newPath
	p.StartLoadingDirectory()
	return LoadDirectoryAsync(newPath)
}

// ChangeDirectory は指定されたパスに移動
func (p *Pane) ChangeDirectory(path string) error {
	p.recordPreviousPath()
	p.path = path
	return p.LoadDirectory()
}

// ChangeDirectoryAsync は指定パスへの移動を開始
func (p *Pane) ChangeDirectoryAsync(path string) tea.Cmd {
	p.recordPreviousPath()
	p.pendingPath = path
	p.path = path
	p.StartLoadingDirectory()
	return LoadDirectoryAsync(path)
}

// Path は現在のパスを返す
func (p *Pane) Path() string {
	return p.path
}

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
		pathStyle = pathStyle.Foreground(lipgloss.Color("39"))
	} else {
		pathStyle = pathStyle.Foreground(lipgloss.Color("240"))
	}

	b.WriteString(pathStyle.Render(displayPath))
	b.WriteString("\n")

	// ヘッダー2行目（マーク情報と空き容量、またはローディング）
	headerLine2 := p.renderHeaderLine2(diskSpace)
	headerStyle := lipgloss.NewStyle().
		Width(p.width-2).
		Padding(0, 1)
	if p.isActive {
		headerStyle = headerStyle.Foreground(lipgloss.Color("245"))
	} else {
		headerStyle = headerStyle.Foreground(lipgloss.Color("240"))
	}
	b.WriteString(headerStyle.Render(headerLine2))
	b.WriteString("\n")

	// 区切り線
	border := strings.Repeat("─", p.width-2)
	b.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(border))
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
			Foreground(lipgloss.Color("243")).
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

	// マーク情報（現在は未実装なので0）
	markedCount := 0
	markedSize := int64(0)

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
	markedLen := len(markedInfo)
	freeLen := len(freeInfo)
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

	// カーソル位置のハイライト
	if isCursor {
		if p.isActive {
			style = style.Background(lipgloss.Color("39")).
				Foreground(lipgloss.Color("15"))
		} else {
			style = style.Background(lipgloss.Color("240")).
				Foreground(lipgloss.Color("15"))
		}
	} else {
		// ファイルタイプによる色付け
		if entry.IsSymlink {
			if entry.LinkBroken {
				style = style.Foreground(lipgloss.Color("9")) // 赤色
			} else {
				style = style.Foreground(lipgloss.Color("14")) // シアン色
			}
		} else if entry.IsDir {
			style = style.Foreground(lipgloss.Color("39")) // 青色
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
	namePadding := nameWidth - len(name)
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
	namePadding := nameWidth - len(name)
	if namePadding < 0 {
		namePadding = 0
	}

	// 所有者とグループを左揃えで各10文字
	ownerPadded := fmt.Sprintf("%-10s", owner)
	groupPadded := fmt.Sprintf("%-10s", group)

	return fmt.Sprintf("%s%s  %s  %s  %s", name, strings.Repeat(" ", namePadding), perms, ownerPadded, groupPadded)
}

// SetSize はペインサイズを設定
func (p *Pane) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetActive はアクティブ状態を設定
func (p *Pane) SetActive(active bool) {
	p.isActive = active
}

// ToggleDisplayMode は表示モードをBasicとDetailの間で切り替える
// 端末幅が狭い場合（ShouldUseMinimalMode() == true）は何もしない
func (p *Pane) ToggleDisplayMode() {
	if p.ShouldUseMinimalMode() {
		// 端末幅が狭い場合は切り替えない
		return
	}

	// displayMode を切り替え（これがユーザーの選択を保存）
	if p.displayMode == DisplayBasic {
		p.displayMode = DisplayDetail
	} else {
		p.displayMode = DisplayBasic
	}
}

// ShouldUseMinimalMode は端末幅に基づいてMinimalモードを使うべきか判定
func (p *Pane) ShouldUseMinimalMode() bool {
	_, hasSpace := CalculateColumnWidths(p.width)
	return !hasSpace
}

// GetEffectiveDisplayMode は実際に使用される表示モードを返す
// 端末幅が狭い場合は自動的にMinimalモードになる
func (p *Pane) GetEffectiveDisplayMode() DisplayMode {
	if p.ShouldUseMinimalMode() {
		return DisplayMinimal
	}
	return p.displayMode
}

// CanToggleMode は現在iキーが有効かどうかを返す
func (p *Pane) CanToggleMode() bool {
	return !p.ShouldUseMinimalMode()
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
		Background(dimmedBgColor).
		Foreground(dimmedFgColor)

	b.WriteString(pathStyle.Render(displayPath))
	b.WriteString("\n")

	// ヘッダー2行目（マーク情報と空き容量）
	headerLine2 := p.renderHeaderLine2(diskSpace)
	headerStyle := lipgloss.NewStyle().
		Width(p.width-2).
		Padding(0, 1).
		Background(dimmedBgColor).
		Foreground(dimmedFgColor)
	b.WriteString(headerStyle.Render(headerLine2))
	b.WriteString("\n")

	// 区切り線
	border := strings.Repeat("─", p.width-2)
	borderStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Background(dimmedBgColor).
		Foreground(dimmedFgColor)
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
		Background(dimmedBgColor)
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

	// dimmedスタイルを適用（カーソルハイライトなし、ファイルタイプ色なし）
	style := lipgloss.NewStyle().
		Width(p.width-2).
		Padding(0, 1).
		Background(dimmedBgColor).
		Foreground(dimmedFgColor)

	return style.Render(line)
}

// ToggleHidden は隠しファイルの表示/非表示を切り替える
// カーソル位置は可能な限り維持する
func (p *Pane) ToggleHidden() {
	// 現在選択中のファイル名を記憶
	var selectedName string
	if p.cursor >= 0 && p.cursor < len(p.entries) {
		selectedName = p.entries[p.cursor].Name
	}

	p.showHidden = !p.showHidden
	p.LoadDirectory()

	// カーソル位置の復元を試みる
	if selectedName != "" {
		for i, e := range p.entries {
			if e.Name == selectedName {
				p.cursor = i
				p.adjustScroll()
				return
			}
		}
	}
	// 見つからない場合（隠しファイルだった場合）は先頭にリセット
	p.cursor = 0
	p.scrollOffset = 0
}

// IsShowingHidden は隠しファイルが表示中かどうかを返す
func (p *Pane) IsShowingHidden() bool {
	return p.showHidden
}

// NavigateToHome はホームディレクトリに移動する
func (p *Pane) NavigateToHome() error {
	home, err := fs.HomeDirectory()
	if err != nil {
		return err
	}

	// すでにホームにいる場合は何もしない
	if p.path == home {
		return nil
	}

	p.recordPreviousPath()
	p.path = home
	return p.LoadDirectory()
}

// NavigateToHomeAsync はホームディレクトリへの移動を開始
func (p *Pane) NavigateToHomeAsync() tea.Cmd {
	home, err := fs.HomeDirectory()
	if err != nil {
		return nil
	}
	if p.path == home {
		return nil
	}
	p.recordPreviousPath()
	p.pendingPath = home
	p.path = home
	p.StartLoadingDirectory()
	return LoadDirectoryAsync(home)
}

// NavigateToPrevious は直前のディレクトリに移動する（トグル動作）
func (p *Pane) NavigateToPrevious() error {
	if p.previousPath == "" {
		return nil // 履歴がない場合は何もしない
	}

	// 現在のパスと直前のパスをスワップ（トグル動作）
	current := p.path
	p.path = p.previousPath
	p.previousPath = current

	return p.LoadDirectory()
}

// NavigateToPreviousAsync は直前のディレクトリへの移動を開始（トグル動作）
func (p *Pane) NavigateToPreviousAsync() tea.Cmd {
	if p.previousPath == "" {
		return nil
	}
	current := p.path
	p.pendingPath = p.previousPath
	p.path = p.previousPath
	p.previousPath = current
	p.StartLoadingDirectory()
	return LoadDirectoryAsync(p.path)
}

// ApplyFilter はフィルタパターンを適用してエントリをフィルタリングする
func (p *Pane) ApplyFilter(pattern string, mode SearchMode) error {
	p.filterPattern = pattern
	p.filterMode = mode

	if pattern == "" {
		// パターンが空の場合はフィルタをクリア
		p.entries = p.allEntries
		p.cursor = 0
		p.scrollOffset = 0
		return nil
	}

	var filtered []fs.FileEntry
	var err error

	switch mode {
	case SearchModeIncremental:
		filtered = filterIncremental(p.allEntries, pattern)
	case SearchModeRegex:
		filtered, err = filterRegex(p.allEntries, pattern)
		if err != nil {
			return err
		}
	default:
		filtered = p.allEntries
	}

	p.entries = filtered

	// カーソル位置を調整
	if p.cursor >= len(p.entries) {
		if len(p.entries) > 0 {
			p.cursor = len(p.entries) - 1
		} else {
			p.cursor = 0
		}
	}
	p.scrollOffset = 0
	p.adjustScroll()

	return nil
}

// ClearFilter はフィルタをクリアしてすべてのエントリを表示する
func (p *Pane) ClearFilter() {
	p.filterPattern = ""
	p.filterMode = SearchModeNone
	p.entries = p.allEntries

	// カーソル位置を調整
	if p.cursor >= len(p.entries) {
		if len(p.entries) > 0 {
			p.cursor = len(p.entries) - 1
		} else {
			p.cursor = 0
		}
	}
	p.adjustScroll()
}

// ResetToFullList はディレクトリを再読み込みしてフィルタをクリアする
func (p *Pane) ResetToFullList() error {
	return p.LoadDirectory()
}

// IsFiltered はフィルタが適用中かどうかを返す
func (p *Pane) IsFiltered() bool {
	return p.filterPattern != ""
}

// FilterPattern は現在のフィルタパターンを返す
func (p *Pane) FilterPattern() string {
	return p.filterPattern
}

// FilterMode は現在のフィルタモードを返す
func (p *Pane) FilterMode() SearchMode {
	return p.filterMode
}

// TotalEntryCount はフィルタ前のエントリ数を返す（親ディレクトリを除く）
func (p *Pane) TotalEntryCount() int {
	count := len(p.allEntries)
	if count > 0 && p.allEntries[0].IsParentDir() {
		count--
	}
	return count
}

// FilteredEntryCount はフィルタ後のエントリ数を返す（親ディレクトリを除く）
func (p *Pane) FilteredEntryCount() int {
	count := len(p.entries)
	if count > 0 && p.entries[0].IsParentDir() {
		count--
	}
	return count
}

// formatFilterIndicator はフィルタインジケーターをフォーマットする
// 例: [/pattern] または [re/pattern]
func (p *Pane) formatFilterIndicator() string {
	if !p.IsFiltered() {
		return ""
	}

	pattern := p.filterPattern
	// パターンが長い場合は切り詰める
	maxLen := 15
	if len(pattern) > maxLen {
		pattern = pattern[:maxLen-2] + ".."
	}

	switch p.filterMode {
	case SearchModeIncremental:
		return fmt.Sprintf("[/%s]", pattern)
	case SearchModeRegex:
		return fmt.Sprintf("[re/%s]", pattern)
	default:
		return ""
	}
}

// Refresh reloads the current directory, preserving cursor position
func (p *Pane) Refresh() error {
	// Save currently selected filename
	var selectedName string
	if p.cursor >= 0 && p.cursor < len(p.entries) {
		selectedName = p.entries[p.cursor].Name
	}
	savedCursor := p.cursor

	// Reload directory with existence check
	currentPath := p.path
	for {
		if fs.DirectoryExists(currentPath) {
			break
		}
		// Navigate up to parent directory
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			// Reached root but it doesn't exist
			home, err := fs.HomeDirectory()
			if err == nil && fs.DirectoryExists(home) {
				currentPath = home
				break
			}
			currentPath = "/"
			break
		}
		currentPath = parent
	}

	if currentPath != p.path {
		// Directory was changed, update previousPath for navigation history
		p.previousPath = p.path
		p.path = currentPath
	}

	err := p.LoadDirectory()
	if err != nil {
		return err
	}

	// Restore cursor position
	if selectedName != "" {
		// Search for the same filename
		for i, e := range p.entries {
			if e.Name == selectedName {
				p.cursor = i
				p.adjustScroll()
				return nil
			}
		}
	}

	// If file not found, use previous index
	if savedCursor < len(p.entries) {
		p.cursor = savedCursor
	} else if len(p.entries) > 0 {
		p.cursor = len(p.entries) - 1
	} else {
		p.cursor = 0
	}
	p.adjustScroll()

	return nil
}

// SyncTo synchronizes this pane to the specified directory
// Preserves display settings but resets cursor to top
func (p *Pane) SyncTo(path string) error {
	// Do nothing if already in the same directory
	if p.path == path {
		return nil
	}

	// Update previousPath for navigation history
	p.previousPath = p.path

	// Change directory
	p.path = path
	err := p.LoadDirectory()
	if err != nil {
		return err
	}

	// Reset cursor and scroll to top
	p.cursor = 0
	p.scrollOffset = 0

	return nil
}
