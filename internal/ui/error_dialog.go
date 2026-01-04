package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ErrorDialog はエラーダイアログ
type ErrorDialog struct {
	BaseDialog
	message string
	styles  DialogStyles
}

// NewErrorDialog は新しいエラーダイアログを作成
func NewErrorDialog(message string) *ErrorDialog {
	base := NewBaseDialog(DialogDisplayScreen)
	return &ErrorDialog{
		BaseDialog: base,
		message:    message,
		styles:     ErrorDialogStyles(base.Width()),
	}
}

// Update はメッセージを処理
func (d *ErrorDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter", "ctrl+c":
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
func (d *ErrorDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder

	// タイトル
	b.WriteString(d.styles.Title.Render("Error"))
	b.WriteString("\n\n")

	// メッセージ
	b.WriteString(d.styles.Body.Render(d.message))
	b.WriteString("\n\n")

	// ヒント
	b.WriteString(d.styles.Footer.Render("Press Esc to close"))

	return d.styles.Box.Render(b.String())
}
