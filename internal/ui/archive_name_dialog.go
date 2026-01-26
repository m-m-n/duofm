package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ArchiveNameDialog はアーカイブ名入力ダイアログ
type ArchiveNameDialog struct {
	BaseDialog
	title     string     // ダイアログタイトル
	textInput *TextInput // reusable text input component
	errorMsg  string     // バリデーションエラーメッセージ
	styles    DialogStyles
}

// NewArchiveNameDialog は新しいアーカイブ名入力ダイアログを作成
func NewArchiveNameDialog(defaultName string) *ArchiveNameDialog {
	base := NewBaseDialog(DialogDisplayScreen)
	base.SetWidth(60)
	return &ArchiveNameDialog{
		BaseDialog: base,
		title:      "Archive Name",
		textInput:  NewTextInput(defaultName),
		errorMsg:   "",
		styles:     NewDialogStyles(60, ColorBorder),
	}
}

// Update はメッセージを処理
func (d *ArchiveNameDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// エラーメッセージをクリア（何かキーを押したら）
		d.errorMsg = ""

		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			// Escapeでキャンセル
			d.Close()
			return d, func() tea.Msg {
				return archiveNameResultMsg{cancelled: true}
			}

		case tea.KeyEnter:
			// Enterで確定
			name := strings.TrimSpace(d.textInput.Value)

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

			d.Close()
			return d, func() tea.Msg {
				return archiveNameResultMsg{name: name, cancelled: false}
			}

		default:
			// Delegate text editing to TextInput
			if d.textInput.HandleKey(msg) {
				return d, nil
			}
		}
	}

	return d, nil
}

// View はダイアログを描画
func (d *ArchiveNameDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder
	width := d.Width()

	// Title
	b.WriteString(d.styles.Title.Render(d.title))
	b.WriteString("\n\n")

	// 入力フィールド（カーソル表示）
	inputWidth := width - 8
	inputText := d.textInput.RenderWithCursor(inputWidth - 2)
	b.WriteString(d.styles.Input.Width(inputWidth).Render(inputText))
	b.WriteString("\n")

	// エラーメッセージ
	if d.errorMsg != "" {
		b.WriteString("\n")
		b.WriteString(d.styles.Error.Render("✗ " + d.errorMsg))
	}

	b.WriteString("\n")
	b.WriteString(d.styles.Footer.Render("[Enter] Confirm  [Esc] Cancel"))

	return d.styles.Box.Render(b.String())
}

// Input returns the current input value.
func (d *ArchiveNameDialog) Input() string {
	return d.textInput.Value
}

// SetInput sets the input value and positions cursor at the end.
func (d *ArchiveNameDialog) SetInput(value string) {
	d.textInput.Value = value
	d.textInput.CursorPos = len([]rune(value))
}

// CursorPos returns the current cursor position.
func (d *ArchiveNameDialog) CursorPos() int {
	return d.textInput.CursorPos
}
