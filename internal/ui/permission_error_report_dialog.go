package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/fs"
)

// PermissionErrorReportDialog はパーミッション変更のエラーレポートダイアログ
type PermissionErrorReportDialog struct {
	BaseDialog
	successCount int
	failureCount int
	errors       []fs.PermissionError
	scrollOffset int
	visibleLines int
	styles       DialogStyles
}

// NewPermissionErrorReportDialog は新しいエラーレポートダイアログを作成
func NewPermissionErrorReportDialog(successCount, failureCount int, errors []fs.PermissionError) *PermissionErrorReportDialog {
	base := NewBaseDialog(DialogDisplayScreen)
	base.SetWidth(70)
	return &PermissionErrorReportDialog{
		BaseDialog:   base,
		successCount: successCount,
		failureCount: failureCount,
		errors:       errors,
		scrollOffset: 0,
		visibleLines: 10, // デフォルトの可視行数
		styles:       NewDialogStyles(70, ColorDanger),
	}
}

// Update はメッセージを処理
func (d *PermissionErrorReportDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyEsc, tea.KeyCtrlC:
			// Enter or Esc or Ctrl+C で閉じる
			d.Close()
			return d, nil

		case tea.KeyRunes:
			switch msg.String() {
			case "j":
				// Scroll down
				maxOffset := len(d.errors) - d.visibleLines
				if maxOffset < 0 {
					maxOffset = 0
				}
				if d.scrollOffset < maxOffset {
					d.scrollOffset++
				}
			case "k":
				// Scroll up
				if d.scrollOffset > 0 {
					d.scrollOffset--
				}
			}

		case tea.KeyCtrlD, tea.KeyPgDown:
			// Page Down
			maxOffset := len(d.errors) - d.visibleLines
			if maxOffset < 0 {
				maxOffset = 0
			}
			d.scrollOffset += d.visibleLines
			if d.scrollOffset > maxOffset {
				d.scrollOffset = maxOffset
			}

		case tea.KeyCtrlU, tea.KeyPgUp:
			// Page Up
			d.scrollOffset -= d.visibleLines
			if d.scrollOffset < 0 {
				d.scrollOffset = 0
			}
		}
	}

	return d, nil
}

// View はダイアログを描画
func (d *PermissionErrorReportDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9"))

	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("246"))

	reasonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("203")).
		Italic(true)

	var content string
	content += d.styles.Title.Render("Permission Change Report") + "\n\n"

	// Success/Failure counts
	content += successStyle.Render(fmt.Sprintf("Success: %d files", d.successCount)) + "\n"
	content += errorStyle.Render(fmt.Sprintf("Failed: %d files", d.failureCount)) + "\n\n"

	// Error list
	if len(d.errors) > 0 {
		content += lipgloss.NewStyle().Bold(true).Render("Failed files:") + "\n\n"

		// Calculate visible range
		start := d.scrollOffset
		end := start + d.visibleLines
		if end > len(d.errors) {
			end = len(d.errors)
		}

		// Show scroll indicator if needed
		if d.scrollOffset > 0 {
			content += pathStyle.Render("  ↑ (more above)") + "\n"
		}

		// Display visible errors
		for i := start; i < end; i++ {
			permErr := d.errors[i]

			// Truncate long paths
			displayPath := permErr.Path
			maxPathLen := 60
			if len(displayPath) > maxPathLen {
				displayPath = "..." + displayPath[len(displayPath)-(maxPathLen-3):]
			}

			content += pathStyle.Render(fmt.Sprintf("  - %s", displayPath)) + "\n"
			content += reasonStyle.Render(fmt.Sprintf("    Error: %s", permErr.Error.Error())) + "\n"
		}

		// Show scroll indicator if more content below
		if end < len(d.errors) {
			content += pathStyle.Render("  ↓ (more below)") + "\n"
		}

		// Show position indicator
		if len(d.errors) > d.visibleLines {
			content += "\n"
			content += pathStyle.Render(
				fmt.Sprintf("Showing %d-%d of %d errors", start+1, end, len(d.errors)),
			) + "\n"
		}
	}

	content += "\n"

	// Help text
	helpText := "[Enter] Close"
	if len(d.errors) > d.visibleLines {
		helpText = "[j/k] Scroll  [PgUp/PgDn] Page  " + helpText
	}
	content += d.styles.Footer.Render(helpText)

	return d.styles.Box.Render(content)
}
