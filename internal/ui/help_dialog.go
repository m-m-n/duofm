package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpDialog はヘルプダイアログ
type HelpDialog struct {
	active        bool
	scrollOffset  int
	contentLines  []string
	visibleHeight int
}

// NewHelpDialog は新しいヘルプダイアログを作成
func NewHelpDialog() *HelpDialog {
	d := &HelpDialog{
		active:        true,
		scrollOffset:  0,
		visibleHeight: 20, // デフォルトの表示行数
	}
	d.contentLines = d.buildContent()
	return d
}

// Update はメッセージを処理
func (d *HelpDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "?", "ctrl+c":
			d.active = false
			return d, func() tea.Msg {
				return dialogResultMsg{
					result: DialogResult{Cancelled: true},
				}
			}
		case "j", "down":
			d.scrollDown(1)
		case "k", "up":
			d.scrollUp(1)
		case " ":
			d.scrollDown(d.visibleHeight)
		case "shift+space":
			d.scrollUp(d.visibleHeight)
		case "g":
			d.scrollOffset = 0
		case "G":
			d.scrollToEnd()
		}
	}

	return d, nil
}

// scrollDown スクロールを下に移動
func (d *HelpDialog) scrollDown(lines int) {
	maxOffset := len(d.contentLines) - d.visibleHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	d.scrollOffset += lines
	if d.scrollOffset > maxOffset {
		d.scrollOffset = maxOffset
	}
}

// scrollUp スクロールを上に移動
func (d *HelpDialog) scrollUp(lines int) {
	d.scrollOffset -= lines
	if d.scrollOffset < 0 {
		d.scrollOffset = 0
	}
}

// scrollToEnd スクロールを最後に移動
func (d *HelpDialog) scrollToEnd() {
	maxOffset := len(d.contentLines) - d.visibleHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	d.scrollOffset = maxOffset
}

// View はダイアログをレンダリング
func (d *HelpDialog) View() string {
	if !d.active {
		return ""
	}

	var b strings.Builder
	width := 70

	// 総ページ数とページ計算
	totalPages := (len(d.contentLines) + d.visibleHeight - 1) / d.visibleHeight
	if totalPages < 1 {
		totalPages = 1
	}
	currentPage := d.scrollOffset/d.visibleHeight + 1
	if currentPage > totalPages {
		currentPage = totalPages
	}

	// ページインジケータ
	pageIndicator := fmt.Sprintf("[%d/%d]", currentPage, totalPages)
	titleWidth := width - 8 - len(pageIndicator)
	titleStyle := lipgloss.NewStyle().
		Width(titleWidth).
		Bold(true).
		Foreground(lipgloss.Color("39"))
	pageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	b.WriteString(titleStyle.Render("Help"))
	b.WriteString(strings.Repeat(" ", width-8-titleWidth-len(pageIndicator)))
	b.WriteString(pageStyle.Render(pageIndicator))
	b.WriteString("\n\n")

	// 表示範囲の計算
	endOffset := d.scrollOffset + d.visibleHeight
	if endOffset > len(d.contentLines) {
		endOffset = len(d.contentLines)
	}

	// コンテンツをレンダリング
	visibleContent := d.contentLines[d.scrollOffset:endOffset]
	for _, line := range visibleContent {
		b.WriteString(line)
		b.WriteString("\n")
	}

	// フッター
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("[j/k: scroll] [Space: page down] [?/Esc: close]"))

	// ボーダーで囲む
	boxStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

// buildContent はヘルプダイアログのコンテンツを生成
func (d *HelpDialog) buildContent() []string {
	var lines []string

	// Keybindings section
	lines = append(lines, "Keybindings")
	lines = append(lines, "")
	lines = append(lines, "Navigation")
	lines = append(lines, "  J/K/Up/Down    : move cursor down/up")
	lines = append(lines, "  H/L/Left/Right : move to left/right pane or parent")
	lines = append(lines, "  Enter          : enter directory / view file")
	lines = append(lines, "  ~              : go to home directory")
	lines = append(lines, "  -              : go to previous directory")
	lines = append(lines, "  Q              : quit")
	lines = append(lines, "")
	lines = append(lines, "File Operations")
	lines = append(lines, "  C              : copy to opposite pane")
	lines = append(lines, "  M              : move to opposite pane")
	lines = append(lines, "  D              : delete (with confirmation)")
	lines = append(lines, "  R              : rename file/directory")
	lines = append(lines, "  N              : create new file")
	lines = append(lines, "  Shift+N        : create new directory")
	lines = append(lines, "  Space          : mark/unmark file")
	lines = append(lines, "  @              : show context menu")
	lines = append(lines, "  !              : execute shell command")
	lines = append(lines, "")
	lines = append(lines, "Display & Search")
	lines = append(lines, "  I              : toggle info mode")
	lines = append(lines, "  Ctrl+H         : toggle hidden files")
	lines = append(lines, "  S              : sort settings")
	lines = append(lines, "  /              : incremental search")
	lines = append(lines, "  Ctrl+F         : regex search")
	lines = append(lines, "")
	lines = append(lines, "External Apps")
	lines = append(lines, "  V              : view file with pager")
	lines = append(lines, "  E              : edit file with editor")
	lines = append(lines, "")
	lines = append(lines, "Other")
	lines = append(lines, "  F5 / Ctrl+R    : refresh view")
	lines = append(lines, "  =              : sync opposite pane")
	lines = append(lines, "  ?              : show this help")
	lines = append(lines, "")
	lines = append(lines, "")

	// Color Palette section
	lines = append(lines, "Color Palette Reference")
	lines = append(lines, "")
	lines = append(lines, "Standard Colors (0-15): Terminal-dependent")
	lines = append(lines, d.renderStandardColors()...)
	lines = append(lines, "")
	lines = append(lines, "6x6x6 Color Cube (16-231):")
	lines = append(lines, d.renderColorCube()...)
	lines = append(lines, "")
	lines = append(lines, "Grayscale (232-255):")
	lines = append(lines, d.renderGrayscale()...)

	return lines
}

// renderStandardColors は標準色（0-15）をレンダリング
func (d *HelpDialog) renderStandardColors() []string {
	var lines []string
	var line1, line2 strings.Builder

	for i := 0; i < 8; i++ {
		colorStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(fmt.Sprintf("%d", i)))
		line1.WriteString(colorStyle.Render("  "))
		line1.WriteString(fmt.Sprintf("%3d ", i))
	}
	lines = append(lines, line1.String())

	for i := 8; i < 16; i++ {
		colorStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(fmt.Sprintf("%d", i)))
		line2.WriteString(colorStyle.Render("  "))
		line2.WriteString(fmt.Sprintf("%3d ", i))
	}
	lines = append(lines, line2.String())

	return lines
}

// renderColorCube は6x6x6カラーキューブ（16-231）をレンダリング
func (d *HelpDialog) renderColorCube() []string {
	var lines []string

	// 6x6x6 = 216 colors, displayed 4 per row
	for row := 0; row < 54; row++ {
		var line strings.Builder
		for col := 0; col < 4; col++ {
			colorNum := 16 + row*4 + col
			if colorNum > 231 {
				break
			}
			colorStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(fmt.Sprintf("%d", colorNum)))
			hexValue := colorCubeToHex(colorNum - 16)
			line.WriteString(colorStyle.Render("  "))
			line.WriteString(fmt.Sprintf("%3d=%s ", colorNum, hexValue))
		}
		lines = append(lines, line.String())
	}

	return lines
}

// renderGrayscale はグレースケール（232-255）をレンダリング
func (d *HelpDialog) renderGrayscale() []string {
	var lines []string

	for row := 0; row < 6; row++ {
		var line strings.Builder
		for col := 0; col < 4; col++ {
			colorNum := 232 + row*4 + col
			if colorNum > 255 {
				break
			}
			colorStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(fmt.Sprintf("%d", colorNum)))
			hexValue := grayscaleToHex(colorNum - 232)
			line.WriteString(colorStyle.Render("  "))
			line.WriteString(fmt.Sprintf("%3d=%s ", colorNum, hexValue))
		}
		lines = append(lines, line.String())
	}

	return lines
}

// colorCubeToHex は6x6x6カラーキューブのインデックスをHex値に変換
func colorCubeToHex(index int) string {
	// 6x6x6 cube: r, g, b each from 0-5
	// Values: 0, 95, 135, 175, 215, 255
	levels := []int{0, 95, 135, 175, 215, 255}

	r := index / 36
	g := (index % 36) / 6
	b := index % 6

	return fmt.Sprintf("#%02x%02x%02x", levels[r], levels[g], levels[b])
}

// grayscaleToHex はグレースケールのインデックスをHex値に変換
func grayscaleToHex(index int) string {
	// Grayscale: 24 shades from 8 to 238, step of 10
	// 232: #080808, 233: #121212, ..., 255: #eeeeee
	gray := 8 + index*10
	return fmt.Sprintf("#%02x%02x%02x", gray, gray, gray)
}

// IsActive はダイアログがアクティブかどうかを返す
func (d *HelpDialog) IsActive() bool {
	return d.active
}

// DisplayType はダイアログの表示タイプを返す
func (d *HelpDialog) DisplayType() DialogDisplayType {
	return DialogDisplayScreen
}
