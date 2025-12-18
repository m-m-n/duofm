package ui

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/fs"
)

// Pane は1つのファイルリストペインを表現
type Pane struct {
	path         string
	entries      []fs.FileEntry
	cursor       int
	scrollOffset int
	width        int
	height       int
	isActive     bool
}

// NewPane は新しいペインを作成
func NewPane(path string, width, height int, isActive bool) (*Pane, error) {
	pane := &Pane{
		path:         path,
		width:        width,
		height:       height,
		isActive:     isActive,
		cursor:       0,
		scrollOffset: 0,
	}

	if err := pane.LoadDirectory(); err != nil {
		return nil, err
	}

	return pane, nil
}

// LoadDirectory はディレクトリを読み込む
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
	visibleLines := p.height - 3 // ヘッダーとボーダーを除く

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
	if entry == nil || !entry.IsDir {
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

// View はペインをレンダリング
func (p *Pane) View() string {
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

	// 区切り線
	border := strings.Repeat("─", p.width-2)
	b.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(border))
	b.WriteString("\n")

	// ファイルリスト
	visibleLines := p.height - 3
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

// formatEntry はエントリを1行にフォーマット
func (p *Pane) formatEntry(entry fs.FileEntry, isCursor bool) string {
	displayName := entry.DisplayName()

	// 表示幅を調整（パディングを考慮）
	maxWidth := p.width - 4
	if len(displayName) > maxWidth {
		displayName = displayName[:maxWidth-3] + "..."
	}

	style := lipgloss.NewStyle().
		Width(p.width-2).
		Padding(0, 1)

	if isCursor {
		if p.isActive {
			style = style.Background(lipgloss.Color("39")).
				Foreground(lipgloss.Color("15"))
		} else {
			style = style.Background(lipgloss.Color("240")).
				Foreground(lipgloss.Color("15"))
		}
	} else if entry.IsDir {
		style = style.Foreground(lipgloss.Color("39"))
	}

	return style.Render(displayName)
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
