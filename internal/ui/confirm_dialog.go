package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmDialog は確認ダイアログ
type ConfirmDialog struct {
	title   string
	message string
	active  bool
}

// NewConfirmDialog は新しい確認ダイアログを作成
func NewConfirmDialog(title, message string) *ConfirmDialog {
	return &ConfirmDialog{
		title:   title,
		message: message,
		active:  true,
	}
}

// Update はメッセージを処理
func (d *ConfirmDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "enter":
			d.active = false
			return d, func() tea.Msg {
				return dialogResultMsg{
					result: DialogResult{Confirmed: true},
				}
			}

		case "n", "esc":
			d.active = false
			return d, func() tea.Msg {
				return dialogResultMsg{
					result: DialogResult{Cancelled: true},
				}
			}
		}
	}

	return d, nil
}

// View はダイアログをレンダリング
func (d *ConfirmDialog) View() string {
	if !d.active {
		return ""
	}

	var b strings.Builder

	width := 50

	// タイトル
	titleStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 2).
		Bold(true).
		Foreground(lipgloss.Color("39"))
	b.WriteString(titleStyle.Render(d.title))
	b.WriteString("\n\n")

	// メッセージ
	messageStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 2)
	b.WriteString(messageStyle.Render(d.message))
	b.WriteString("\n\n")

	// ボタン
	buttonStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 2).
		Foreground(lipgloss.Color("240"))
	b.WriteString(buttonStyle.Render("[y] Yes  [n] No"))

	// ボーダーで囲む
	boxStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

// IsActive はダイアログがアクティブかどうかを返す
func (d *ConfirmDialog) IsActive() bool {
	return d.active
}
