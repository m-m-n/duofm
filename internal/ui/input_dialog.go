package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputDialog はテキスト入力ダイアログ
type InputDialog struct {
	title     string               // ダイアログタイトル/プロンプト
	input     string               // 現在の入力テキスト
	cursorPos int                  // カーソル位置
	active    bool                 // ダイアログがアクティブ
	width     int                  // ダイアログの幅
	onConfirm func(string) tea.Cmd // Enter時のコールバック
	errorMsg  string               // バリデーションエラーメッセージ
}

// NewInputDialog は新しい入力ダイアログを作成
func NewInputDialog(title string, onConfirm func(string) tea.Cmd) *InputDialog {
	return &InputDialog{
		title:     title,
		input:     "",
		cursorPos: 0,
		active:    true,
		width:     50,
		onConfirm: onConfirm,
		errorMsg:  "",
	}
}

// Update はメッセージを処理
func (d *InputDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// エラーメッセージをクリア（何かキーを押したら）
		d.errorMsg = ""

		switch msg.Type {
		case tea.KeyEnter:
			// 空文字列チェック
			if d.input == "" {
				d.errorMsg = "File name cannot be empty"
				return d, nil
			}
			d.active = false
			if d.onConfirm != nil {
				return d, d.onConfirm(d.input)
			}
			return d, nil

		case tea.KeyEsc:
			d.active = false
			return d, nil

		case tea.KeyRunes:
			// 文字入力
			runes := []rune(d.input)
			newRunes := make([]rune, 0, len(runes)+len(msg.Runes))
			newRunes = append(newRunes, runes[:d.cursorPos]...)
			newRunes = append(newRunes, msg.Runes...)
			newRunes = append(newRunes, runes[d.cursorPos:]...)
			d.input = string(newRunes)
			d.cursorPos += len(msg.Runes)
			return d, nil

		case tea.KeyBackspace:
			if d.cursorPos > 0 {
				runes := []rune(d.input)
				newRunes := make([]rune, 0, len(runes)-1)
				newRunes = append(newRunes, runes[:d.cursorPos-1]...)
				newRunes = append(newRunes, runes[d.cursorPos:]...)
				d.input = string(newRunes)
				d.cursorPos--
			}
			return d, nil

		case tea.KeyDelete:
			runes := []rune(d.input)
			if d.cursorPos < len(runes) {
				newRunes := make([]rune, 0, len(runes)-1)
				newRunes = append(newRunes, runes[:d.cursorPos]...)
				newRunes = append(newRunes, runes[d.cursorPos+1:]...)
				d.input = string(newRunes)
			}
			return d, nil

		case tea.KeyLeft:
			if d.cursorPos > 0 {
				d.cursorPos--
			}
			return d, nil

		case tea.KeyRight:
			if d.cursorPos < len([]rune(d.input)) {
				d.cursorPos++
			}
			return d, nil

		case tea.KeyCtrlA:
			// 行頭へ移動
			d.cursorPos = 0
			return d, nil

		case tea.KeyCtrlE:
			// 行末へ移動
			d.cursorPos = len([]rune(d.input))
			return d, nil

		case tea.KeyCtrlU:
			// カーソルから行頭まで削除
			runes := []rune(d.input)
			d.input = string(runes[d.cursorPos:])
			d.cursorPos = 0
			return d, nil

		case tea.KeyCtrlK:
			// カーソルから行末まで削除
			runes := []rune(d.input)
			d.input = string(runes[:d.cursorPos])
			return d, nil
		}
	}

	return d, nil
}

// View はダイアログをレンダリング
func (d *InputDialog) View() string {
	if !d.active {
		return ""
	}

	var b strings.Builder
	width := d.width

	// タイトル
	titleStyle := lipgloss.NewStyle().
		Width(width - 4).
		Padding(0, 1).
		Bold(true).
		Foreground(lipgloss.Color("39"))
	b.WriteString(titleStyle.Render(d.title))
	b.WriteString("\n\n")

	// 入力フィールド
	inputWidth := width - 8 // パディングとボーダー分を引く
	b.WriteString(d.renderInputField(inputWidth))
	b.WriteString("\n")

	// エラーメッセージ（あれば）
	if d.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Width(width - 4).
			Padding(0, 1).
			Foreground(lipgloss.Color("196")) // 赤色
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(d.errorMsg))
	}

	b.WriteString("\n")

	// フッター
	footerStyle := lipgloss.NewStyle().
		Width(width - 4).
		Padding(0, 1).
		Foreground(lipgloss.Color("240"))
	b.WriteString(footerStyle.Render("Enter: Confirm  Esc: Cancel"))

	// ボーダーで囲む
	boxStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

// renderInputField は入力フィールドをレンダリング
func (d *InputDialog) renderInputField(width int) string {
	runes := []rune(d.input)
	displayInput := d.input

	// 表示可能な範囲を計算
	cursorDisplayPos := d.cursorPos
	startPos := 0

	if len(runes) > width-2 {
		// カーソルが表示範囲内になるように調整
		if d.cursorPos > width-3 {
			startPos = d.cursorPos - width + 3
		}
		endPos := startPos + width - 2
		if endPos > len(runes) {
			endPos = len(runes)
		}
		displayInput = string(runes[startPos:endPos])
		cursorDisplayPos = d.cursorPos - startPos
	}

	// カーソル付きで表示文字列を構築
	displayRunes := []rune(displayInput)
	var result strings.Builder
	for i, r := range displayRunes {
		if i == cursorDisplayPos {
			// カーソル位置を反転表示
			result.WriteString(lipgloss.NewStyle().Reverse(true).Render(string(r)))
		} else {
			result.WriteRune(r)
		}
	}
	// カーソルが末尾の場合はブロックカーソルを表示
	if cursorDisplayPos >= len(displayRunes) {
		result.WriteString(lipgloss.NewStyle().Reverse(true).Render(" "))
	}

	// 入力フィールドのスタイル
	fieldStyle := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("236")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	return fieldStyle.Render(result.String())
}

// IsActive はダイアログがアクティブかどうかを返す
func (d *InputDialog) IsActive() bool {
	return d.active
}

// DisplayType はダイアログの表示タイプを返す
func (d *InputDialog) DisplayType() DialogDisplayType {
	return DialogDisplayPane
}

// SetWidth はダイアログの幅を設定
func (d *InputDialog) SetWidth(width int) {
	d.width = width
}
