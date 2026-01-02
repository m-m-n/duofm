package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ArchiveNameDialog はアーカイブ名入力ダイアログ
type ArchiveNameDialog struct {
	title     string // ダイアログタイトル
	input     string // 現在の入力テキスト
	cursorPos int    // カーソル位置
	active    bool   // ダイアログがアクティブ
	width     int    // ダイアログの幅
	errorMsg  string // バリデーションエラーメッセージ
}

// NewArchiveNameDialog は新しいアーカイブ名入力ダイアログを作成
func NewArchiveNameDialog(defaultName string) *ArchiveNameDialog {
	return &ArchiveNameDialog{
		title:     "Archive Name",
		input:     defaultName,
		cursorPos: len(defaultName),
		active:    true,
		width:     60,
		errorMsg:  "",
	}
}

// Update はメッセージを処理
func (d *ArchiveNameDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// エラーメッセージをクリア（何かキーを押したら）
		d.errorMsg = ""

		switch msg.Type {
		case tea.KeyEsc:
			// Escapeでキャンセル
			d.active = false
			return d, func() tea.Msg {
				return archiveNameResultMsg{cancelled: true}
			}

		case tea.KeyEnter:
			// Enterで確定
			name := strings.TrimSpace(d.input)

			// バリデーション
			if name == "" {
				d.errorMsg = "Archive name cannot be empty"
				return d, nil
			}

			// 不正な文字チェック（NUL、制御文字）
			for _, c := range name {
				if c == 0 || (c < 32 && c != '\t') {
					d.errorMsg = "Archive name contains invalid characters"
					return d, nil
				}
			}

			d.active = false
			return d, func() tea.Msg {
				return archiveNameResultMsg{name: name, cancelled: false}
			}

		case tea.KeyBackspace, tea.KeyDelete:
			if d.cursorPos > 0 {
				d.input = d.input[:d.cursorPos-1] + d.input[d.cursorPos:]
				d.cursorPos--
			}

		case tea.KeyLeft:
			if d.cursorPos > 0 {
				d.cursorPos--
			}

		case tea.KeyRight:
			if d.cursorPos < len(d.input) {
				d.cursorPos++
			}

		case tea.KeyHome, tea.KeyCtrlA:
			d.cursorPos = 0

		case tea.KeyEnd, tea.KeyCtrlE:
			d.cursorPos = len(d.input)

		case tea.KeyRunes:
			// 文字入力
			d.input = d.input[:d.cursorPos] + string(msg.Runes) + d.input[d.cursorPos:]
			d.cursorPos += len(msg.Runes)
		}
	}

	return d, nil
}

// View はダイアログを描画
func (d *ArchiveNameDialog) View() string {
	if !d.active {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Width(d.width - 6)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		MarginTop(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var content string
	content += titleStyle.Render(d.title) + "\n\n"

	// 入力フィールド（カーソル表示）
	inputText := d.input
	if d.cursorPos < len(inputText) {
		inputText = inputText[:d.cursorPos] + "█" + inputText[d.cursorPos+1:]
	} else {
		inputText += "█"
	}
	content += inputStyle.Render(inputText) + "\n"

	// エラーメッセージ
	if d.errorMsg != "" {
		content += errorStyle.Render("✗ "+d.errorMsg) + "\n"
	}

	content += helpStyle.Render("[Enter] Confirm  [Esc] Cancel")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(d.width)

	return boxStyle.Render(content)
}

// IsActive はダイアログがアクティブかを返す
func (d *ArchiveNameDialog) IsActive() bool {
	return d.active
}

// SetActive はダイアログのアクティブ状態を設定
func (d *ArchiveNameDialog) SetActive(active bool) {
	d.active = active
}

// DisplayType はダイアログの表示タイプを返す
func (d *ArchiveNameDialog) DisplayType() DialogDisplayType {
	return DialogDisplayScreen
}
