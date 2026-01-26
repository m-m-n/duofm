package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CompressionLevelDialog は圧縮レベル選択ダイアログ
type CompressionLevelDialog struct {
	BaseDialog
	selectedLevel int // 選択された圧縮レベル (0-9)
	styles        DialogStyles
}

// NewCompressionLevelDialog は新しい圧縮レベル選択ダイアログを作成
func NewCompressionLevelDialog() *CompressionLevelDialog {
	base := NewBaseDialog(DialogDisplayScreen)
	base.SetWidth(60)
	return &CompressionLevelDialog{
		BaseDialog:    base,
		selectedLevel: 6, // デフォルト: Normal (推奨)
		styles:        NewDialogStyles(60, ColorBorder),
	}
}

// Update はメッセージを処理
func (d *CompressionLevelDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			d.Close()
			return d, func() tea.Msg {
				return compressionLevelResultMsg{level: 6, cancelled: true}
			}

		case tea.KeyEnter:
			d.Close()
			return d, func() tea.Msg {
				return compressionLevelResultMsg{level: d.selectedLevel, cancelled: false}
			}

		case tea.KeyRunes:
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
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(d.styles.Title.Render("Select Compression Level"))
	b.WriteString("\n\n")

	// Level list
	selectedStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Background(lipgloss.Color(string(ColorBorder))).
		Foreground(lipgloss.Color("230")).
		Bold(true)

	levelStyle := lipgloss.NewStyle().Padding(0, 2)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(string(ColorMuted)))

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
			b.WriteString(selectedStyle.Render("→ "+line) + " " + descStyle.Render(l.desc) + "\n")
		} else {
			b.WriteString(levelStyle.Render("  "+line) + " " + descStyle.Render(l.desc) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(d.styles.Footer.Render("[j/k] Navigate  [0-9] Direct select  [Enter] Confirm  [Esc] Use default (6)"))

	return d.styles.Box.Render(b.String())
}
