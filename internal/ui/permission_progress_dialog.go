package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PermissionProgressDialog はパーミッション変更操作の進捗表示ダイアログ
type PermissionProgressDialog struct {
	BaseDialog
	totalFiles     int
	processedFiles int
	currentFile    string
	startTime      time.Time
	onCancel       func()
	styles         DialogStyles
}

// NewPermissionProgressDialog は新しい進捗表示ダイアログを作成
func NewPermissionProgressDialog(totalFiles int) *PermissionProgressDialog {
	base := NewBaseDialog(DialogDisplayScreen)
	base.SetWidth(70)
	return &PermissionProgressDialog{
		BaseDialog:     base,
		totalFiles:     totalFiles,
		processedFiles: 0,
		currentFile:    "",
		startTime:      time.Now(),
		onCancel:       nil,
		styles:         NewDialogStyles(70, ColorBorder),
	}
}

// SetOnCancel はキャンセルコールバックを設定
func (d *PermissionProgressDialog) SetOnCancel(callback func()) {
	d.onCancel = callback
}

// UpdateProgress は進捗情報を更新
func (d *PermissionProgressDialog) UpdateProgress(processed int, currentFile string) {
	d.processedFiles = processed
	d.currentFile = currentFile
}

// Complete は操作完了を通知
func (d *PermissionProgressDialog) Complete() {
	d.Close()
}

// Update はメッセージを処理
func (d *PermissionProgressDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			if d.onCancel != nil {
				d.onCancel()
			}
			return d, nil
		}
	}

	return d, nil
}

// View はダイアログを描画
func (d *PermissionProgressDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(d.styles.Title.Render("Changing Permissions"))
	b.WriteString("\n\n")

	// プログレスバー
	percentage := 0
	if d.totalFiles > 0 {
		percentage = (d.processedFiles * 100) / d.totalFiles
	}

	barWidth := 50
	filledWidth := (percentage * barWidth) / 100
	emptyWidth := barWidth - filledWidth

	progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(string(ColorBorder)))
	bar := strings.Repeat("▓", filledWidth) + strings.Repeat("░", emptyWidth)
	b.WriteString(progressStyle.Render(fmt.Sprintf("[%s] %d%%", bar, percentage)))
	b.WriteString("\n\n")

	// ファイル数
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	b.WriteString(infoStyle.Render(fmt.Sprintf("Progress: %d / %d files", d.processedFiles, d.totalFiles)))
	b.WriteString("\n")

	// 現在処理中のファイル
	if d.currentFile != "" {
		currentFile := d.currentFile
		maxPathLen := 60
		if len(currentFile) > maxPathLen {
			currentFile = "..." + currentFile[len(currentFile)-(maxPathLen-3):]
		}
		b.WriteString(infoStyle.Render(fmt.Sprintf("Current: %s", currentFile)))
		b.WriteString("\n")
	}

	// 経過時間
	elapsed := time.Since(d.startTime)
	b.WriteString("\n")
	b.WriteString(infoStyle.Render(fmt.Sprintf("Elapsed: %s", formatDuration(elapsed))))
	b.WriteString("\n\n")

	b.WriteString(d.styles.Footer.Render("[Ctrl+C] Cancel"))

	return d.styles.Box.Render(b.String())
}
