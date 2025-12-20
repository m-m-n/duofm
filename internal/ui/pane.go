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

// Pane は1つのファイルリストペインを表現
type Pane struct {
	path            string
	entries         []fs.FileEntry
	cursor          int
	scrollOffset    int
	width           int
	height          int
	isActive        bool
	displayMode     DisplayMode // ユーザーが選択した表示モード
	loading         bool        // ローディング中フラグ
	loadingProgress string      // ローディングメッセージ
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
	p.entries = entries
	p.cursor = 0
	p.scrollOffset = 0

	return nil
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
				panePath: panePath,
				entries:  nil,
				err:      err,
			}
		}

		fs.SortEntries(entries)
		return directoryLoadCompleteMsg{
			panePath: panePath,
			entries:  entries,
			err:      nil,
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

// SelectedEntry は選択中のエントリを返す
func (p *Pane) SelectedEntry() *fs.FileEntry {
	if p.cursor < 0 || p.cursor >= len(p.entries) {
		return nil
	}
	return &p.entries[p.cursor]
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

		// リンク先のディレクトリに移動
		p.path = entry.LinkTarget
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

	p.path = newPath
	return p.LoadDirectory()
}

// MoveToParent は親ディレクトリに移動
func (p *Pane) MoveToParent() error {
	if p.path == "/" {
		return nil // ルートより上には行けない
	}

	p.path = filepath.Dir(p.path)
	return p.LoadDirectory()
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
	var b strings.Builder

	// パス表示（ホームディレクトリは ~ に置換）
	displayPath := p.formatPath()
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

	// ファイルリスト
	visibleLines := p.height - 4 // ヘッダー2行 + ボーダー1行 = 3行
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
	totalCount := len(p.entries)
	// 親ディレクトリエントリを除外
	if totalCount > 0 && p.entries[0].IsParentDir() {
		totalCount--
	}

	markedSize := int64(0)

	// マーク情報を作成
	markedInfo := fmt.Sprintf("Marked %d/%d %s", markedCount, totalCount, FormatSize(markedSize))

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
