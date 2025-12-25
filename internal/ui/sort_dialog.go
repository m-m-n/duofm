package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SortDialog はソート設定を変更するためのダイアログ
type SortDialog struct {
	config         SortConfig // 現在の選択
	originalConfig SortConfig // キャンセル時の復元用
	focusedRow     int        // 0: Sort by, 1: Order
	active         bool
	width          int
}

// NewSortDialog は新しいソートダイアログを作成
func NewSortDialog(current SortConfig) *SortDialog {
	return &SortDialog{
		config:         current,
		originalConfig: current,
		focusedRow:     0,
		active:         true,
		width:          36,
	}
}

// HandleKey はキー入力を処理し、確定またはキャンセル状態を返す
func (d *SortDialog) HandleKey(key string) (confirmed bool, cancelled bool) {
	switch key {
	case "h", "left":
		d.moveLeft()
	case "l", "right":
		d.moveRight()
	case "j", "down":
		if d.focusedRow < 1 {
			d.focusedRow = 1
		}
	case "k", "up":
		if d.focusedRow > 0 {
			d.focusedRow = 0
		}
	case "enter":
		d.active = false
		return true, false
	case "esc", "q":
		d.config = d.originalConfig
		d.active = false
		return false, true
	}
	return false, false
}

// moveLeft は現在の行で左に移動
func (d *SortDialog) moveLeft() {
	if d.focusedRow == 0 {
		// Sort by: Name <- Size <- Date
		if d.config.Field > SortByName {
			d.config.Field--
		}
	} else {
		// Order: Asc <- Desc
		d.config.Order = SortAsc
	}
}

// moveRight は現在の行で右に移動
func (d *SortDialog) moveRight() {
	if d.focusedRow == 0 {
		// Sort by: Name -> Size -> Date
		if d.config.Field < SortByDate {
			d.config.Field++
		}
	} else {
		// Order: Asc -> Desc
		d.config.Order = SortDesc
	}
}

// Config は現在の設定を返す
func (d *SortDialog) Config() SortConfig {
	return d.config
}

// OriginalConfig は元の設定を返す
func (d *SortDialog) OriginalConfig() SortConfig {
	return d.originalConfig
}

// IsActive はダイアログがアクティブかどうかを返す
func (d *SortDialog) IsActive() bool {
	return d.active
}

// DisplayType はダイアログの表示タイプを返す
func (d *SortDialog) DisplayType() DialogDisplayType {
	return DialogDisplayPane
}

// Update はbubbletea互換のUpdate実装
func (d *SortDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		confirmed, cancelled := d.HandleKey(msg.String())

		if confirmed {
			return d, func() tea.Msg {
				return sortDialogResultMsg{config: d.config, confirmed: true}
			}
		}

		if cancelled {
			return d, func() tea.Msg {
				return sortDialogResultMsg{config: d.originalConfig, cancelled: true}
			}
		}

		// ライブプレビュー用: 設定変更メッセージ
		return d, func() tea.Msg {
			return sortDialogConfigChangedMsg{config: d.config}
		}
	}

	return d, nil
}

// View はダイアログをレンダリング
func (d *SortDialog) View() string {
	if !d.active {
		return ""
	}

	var b strings.Builder

	// タイトル
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39"))
	b.WriteString(titleStyle.Render("Sort"))
	b.WriteString("\n\n")

	// Sort by 行
	b.WriteString(d.renderSortByRow())
	b.WriteString("\n")

	// Order 行
	b.WriteString(d.renderOrderRow())
	b.WriteString("\n\n")

	// ヘルプテキスト
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	b.WriteString(helpStyle.Render("h/l:change  j/k:row"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Enter:OK  Esc:cancel"))

	// ボックススタイル
	boxStyle := lipgloss.NewStyle().
		Width(d.width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

// renderSortByRow はSort by行をレンダリング
func (d *SortDialog) renderSortByRow() string {
	labelStyle := lipgloss.NewStyle().Width(10)

	fields := []struct {
		field SortField
		label string
	}{
		{SortByName, "Name"},
		{SortBySize, "Size"},
		{SortByDate, "Date"},
	}

	var options []string
	for _, f := range fields {
		if d.config.Field == f.field {
			options = append(options, d.renderSelected(f.label, d.focusedRow == 0))
		} else {
			options = append(options, d.renderUnselected(f.label, d.focusedRow == 0))
		}
	}

	return labelStyle.Render("Sort by") + "  " + strings.Join(options, "  ")
}

// renderOrderRow はOrder行をレンダリング
func (d *SortDialog) renderOrderRow() string {
	labelStyle := lipgloss.NewStyle().Width(10)

	orders := []struct {
		order SortOrder
		label string
	}{
		{SortAsc, "↑Asc"},
		{SortDesc, "↓Desc"},
	}

	var options []string
	for _, o := range orders {
		if d.config.Order == o.order {
			options = append(options, d.renderSelected(o.label, d.focusedRow == 1))
		} else {
			options = append(options, d.renderUnselected(o.label, d.focusedRow == 1))
		}
	}

	return labelStyle.Render("Order") + "  " + strings.Join(options, "  ")
}

// renderSelected は選択中の項目をレンダリング
func (d *SortDialog) renderSelected(label string, isFocusedRow bool) string {
	style := lipgloss.NewStyle()
	if isFocusedRow {
		style = style.
			Background(lipgloss.Color("39")).
			Foreground(lipgloss.Color("0"))
	} else {
		style = style.
			Foreground(lipgloss.Color("39")).
			Bold(true)
	}
	return "[" + style.Render(label) + "]"
}

// renderUnselected は非選択の項目をレンダリング
func (d *SortDialog) renderUnselected(label string, isFocusedRow bool) string {
	style := lipgloss.NewStyle()
	if isFocusedRow {
		style = style.Foreground(lipgloss.Color("240"))
	} else {
		style = style.Foreground(lipgloss.Color("243"))
	}
	return " " + style.Render(label) + " "
}

// sortDialogResultMsg はソートダイアログの結果メッセージ
type sortDialogResultMsg struct {
	config    SortConfig
	confirmed bool
	cancelled bool
}

// sortDialogConfigChangedMsg はソート設定変更時のメッセージ（ライブプレビュー用）
type sortDialogConfigChangedMsg struct {
	config SortConfig
}
