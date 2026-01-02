package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CompressionLevelDialog は圧縮レベル選択ダイアログ
type CompressionLevelDialog struct {
	selectedLevel int  // 選択された圧縮レベル (0-9)
	active        bool // ダイアログがアクティブ
	width         int  // ダイアログの幅
}

// NewCompressionLevelDialog は新しい圧縮レベル選択ダイアログを作成
func NewCompressionLevelDialog() *CompressionLevelDialog {
	return &CompressionLevelDialog{
		selectedLevel: 6, // デフォルト: Normal (推奨)
		active:        true,
		width:         60,
	}
}

// Update はメッセージを処理
func (d *CompressionLevelDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			// Escapeでキャンセル
			d.active = false
			return d, func() tea.Msg {
				return compressionLevelResultMsg{level: 6, cancelled: true}
			}

		case tea.KeyEnter:
			// Enterで確定
			d.active = false
			return d, func() tea.Msg {
				return compressionLevelResultMsg{level: d.selectedLevel, cancelled: false}
			}

		case tea.KeyRunes:
			// j/k で上下移動、0-9で直接選択
			switch msg.String() {
			case "j":
				if d.selectedLevel < 9 {
					d.selectedLevel++
				}
			case "k":
				if d.selectedLevel > 0 {
					d.selectedLevel--
				}
			case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
				d.selectedLevel = int(msg.Runes[0] - '0')
			}
		}
	}

	return d, nil
}

// View はダイアログを描画
func (d *CompressionLevelDialog) View() string {
	if !d.active {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	levelStyle := lipgloss.NewStyle().
		Padding(0, 2)

	selectedStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var content string
	content += titleStyle.Render("Select Compression Level") + "\n\n"

	levels := []struct {
		level int
		desc  string
	}{
		{0, "No compression (fastest)"},
		{1, "Fast compression"},
		{2, "Fast compression"},
		{3, "Fast compression"},
		{4, "Normal compression"},
		{5, "Normal compression"},
		{6, "Normal compression (recommended)"},
		{7, "Best compression"},
		{8, "Best compression"},
		{9, "Best compression (slowest)"},
	}

	for _, l := range levels {
		line := fmt.Sprintf("Level %d", l.level)
		if d.selectedLevel == l.level {
			content += selectedStyle.Render("→ "+line) + " " + descStyle.Render(l.desc) + "\n"
		} else {
			content += levelStyle.Render("  "+line) + " " + descStyle.Render(l.desc) + "\n"
		}
	}

	content += "\n"
	content += helpStyle.Render("[j/k] Navigate  [0-9] Direct select  [Enter] Confirm  [Esc] Use default (6)")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(d.width)

	return boxStyle.Render(content)
}

// IsActive はダイアログがアクティブかを返す
func (d *CompressionLevelDialog) IsActive() bool {
	return d.active
}

// SetActive はダイアログのアクティブ状態を設定
func (d *CompressionLevelDialog) SetActive(active bool) {
	d.active = active
}

// DisplayType はダイアログの表示タイプを返す
func (d *CompressionLevelDialog) DisplayType() DialogDisplayType {
	return DialogDisplayScreen
}
