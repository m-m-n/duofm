package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpDialog はヘルプダイアログ
type HelpDialog struct {
	active bool
}

// NewHelpDialog は新しいヘルプダイアログを作成
func NewHelpDialog() *HelpDialog {
	return &HelpDialog{
		active: true,
	}
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
		}
	}

	return d, nil
}

// View はダイアログをレンダリング
func (d *HelpDialog) View() string {
	if !d.active {
		return ""
	}

	var b strings.Builder

	width := 70

	// タイトル
	titleStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 2).
		Bold(true).
		Foreground(lipgloss.Color("39"))
	b.WriteString(titleStyle.Render("Keybindings"))
	b.WriteString("\n\n")

	// カテゴリとキーバインディング
	contentStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 2)

	content := []string{
		"Navigation",
		"  J/K/Up/Down    : move cursor down/up",
		"  H/L/Left/Right : move to left/right pane or parent",
		"  Enter          : enter directory / view file",
		"  ~              : go to home directory",
		"  -              : go to previous directory",
		"  Q              : quit",
		"",
		"File Operations",
		"  C              : copy to opposite pane",
		"  M              : move to opposite pane",
		"  D              : delete (with confirmation)",
		"  R              : rename file/directory",
		"  N              : create new file",
		"  Shift+N        : create new directory",
		"  Space          : mark/unmark file",
		"  @              : show context menu",
		"  !              : execute shell command",
		"",
		"Display & Search",
		"  I              : toggle info mode",
		"  Ctrl+H         : toggle hidden files",
		"  S              : sort settings",
		"  /              : incremental search",
		"  Ctrl+F         : regex search",
		"",
		"External Apps",
		"  V              : view file with pager",
		"  E              : edit file with editor",
		"",
		"Other",
		"  F5 / Ctrl+R    : refresh view",
		"  =              : sync opposite pane",
		"  ?              : show this help",
		"",
		"Press Esc or ? to close",
	}

	b.WriteString(contentStyle.Render(strings.Join(content, "\n")))

	// ボーダーで囲む
	boxStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

// IsActive はダイアログがアクティブかどうかを返す
func (d *HelpDialog) IsActive() bool {
	return d.active
}

// DisplayType はダイアログの表示タイプを返す
func (d *HelpDialog) DisplayType() DialogDisplayType {
	return DialogDisplayScreen
}
