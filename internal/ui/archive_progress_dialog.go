package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/archive"
)

// ArchiveProgressDialog はアーカイブ操作の進捗表示ダイアログ
type ArchiveProgressDialog struct {
	BaseDialog
	operation   string                  // "compress" or "extract"
	archivePath string                  // アーカイブファイルのパス
	progress    *archive.ProgressUpdate // 現在の進捗情報
	onCancel    func()                  // キャンセル時のコールバック
	styles      DialogStyles
}

// NewArchiveProgressDialog は新しい進捗表示ダイアログを作成
func NewArchiveProgressDialog(operation string, archivePath string) *ArchiveProgressDialog {
	base := NewBaseDialog(DialogDisplayScreen)
	base.SetWidth(70)
	return &ArchiveProgressDialog{
		BaseDialog:  base,
		operation:   operation,
		archivePath: archivePath,
		progress:    nil,
		onCancel:    nil,
		styles:      NewDialogStyles(70, ColorBorder),
	}
}

// SetOnCancel はキャンセルコールバックを設定
func (d *ArchiveProgressDialog) SetOnCancel(callback func()) {
	d.onCancel = callback
}

// UpdateProgress は進捗情報を更新
func (d *ArchiveProgressDialog) UpdateProgress(progress *archive.ProgressUpdate) {
	d.progress = progress
}

// Complete は操作完了を通知
func (d *ArchiveProgressDialog) Complete() {
	d.Close()
}

// Update はメッセージを処理
func (d *ArchiveProgressDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			// Escapeでキャンセル
			if d.onCancel != nil {
				d.onCancel()
			}
			return d, nil
		}
	}

	return d, nil
}

// View はダイアログを描画
func (d *ArchiveProgressDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder

	progressBarStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(string(ColorBorder)))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))

	var title string
	if d.operation == "compress" {
		title = "Compressing Archive"
	} else {
		title = "Extracting Archive"
	}

	b.WriteString(d.styles.Title.Render(title))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render(d.archivePath))
	b.WriteString("\n\n")

	if d.progress != nil {
		// プログレスバー
		percentage := d.progress.Percentage()
		barWidth := 50
		filledWidth := (percentage * barWidth) / 100
		emptyWidth := barWidth - filledWidth

		bar := "[" + strings.Repeat("█", filledWidth) + strings.Repeat("░", emptyWidth) + "]"
		b.WriteString(progressBarStyle.Render(fmt.Sprintf("%s %d%%", bar, percentage)))
		b.WriteString("\n\n")

		// ファイル数
		b.WriteString(infoStyle.Render(fmt.Sprintf("Files: %d/%d", d.progress.ProcessedFiles, d.progress.TotalFiles)))
		b.WriteString("\n")

		// 現在処理中のファイル
		if d.progress.CurrentFile != "" {
			currentFile := d.progress.CurrentFile
			if len(currentFile) > 50 {
				currentFile = "..." + currentFile[len(currentFile)-47:]
			}
			b.WriteString(infoStyle.Render(fmt.Sprintf("Current: %s", currentFile)))
			b.WriteString("\n")
		}

		// 経過時間
		elapsed := d.progress.ElapsedTime()
		b.WriteString(infoStyle.Render(fmt.Sprintf("Elapsed: %s", formatDuration(elapsed))))
		b.WriteString("\n")

		// 推定残り時間
		if d.progress.ProcessedFiles > 0 {
			remaining := d.progress.EstimatedRemaining()
			b.WriteString(infoStyle.Render(fmt.Sprintf("Remaining: %s", formatDuration(remaining))))
			b.WriteString("\n")
		}
	} else {
		// 進捗情報がまだない場合
		b.WriteString(infoStyle.Render("Starting..."))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(d.styles.Footer.Render("[Esc] Cancel"))

	return d.styles.Box.Render(b.String())
}

// formatDuration は時間を MM:SS 形式にフォーマット
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
