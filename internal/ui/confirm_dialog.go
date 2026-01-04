package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ConfirmDialog は確認ダイアログ
type ConfirmDialog struct {
	BaseDialog
	title   string
	message string
	styles  DialogStyles
}

// NewConfirmDialog は新しい確認ダイアログを作成
func NewConfirmDialog(title, message string) *ConfirmDialog {
	base := NewBaseDialog(DialogDisplayPane)
	return &ConfirmDialog{
		BaseDialog: base,
		title:      title,
		message:    message,
		styles:     DefaultDialogStyles(base.Width()),
	}
}

// Update はメッセージを処理
func (d *ConfirmDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			d.Close()
			return d, func() tea.Msg {
				return dialogResultMsg{
					result: DialogResult{Confirmed: true},
				}
			}

		case "n", "esc", "ctrl+c":
			d.Close()
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
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder

	// タイトル
	b.WriteString(d.styles.Title.Render(d.title))
	b.WriteString("\n\n")

	// メッセージ
	b.WriteString(d.styles.Body.Render(d.message))
	b.WriteString("\n\n")

	// ボタン
	b.WriteString(d.styles.Footer.Render("[y] Yes  [n] No"))

	return d.styles.Box.Render(b.String())
}
