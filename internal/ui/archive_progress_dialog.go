package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/archive"
)

// ArchiveProgressDialog はアーカイブ操作の進捗表示ダイアログ
type ArchiveProgressDialog struct {
	operation   string                  // "compress" or "extract"
	archivePath string                  // アーカイブファイルのパス
	progress    *archive.ProgressUpdate // 現在の進捗情報
	active      bool                    // ダイアログがアクティブ
	width       int                     // ダイアログの幅
	onCancel    func()                  // キャンセル時のコールバック
}

// NewArchiveProgressDialog は新しい進捗表示ダイアログを作成
func NewArchiveProgressDialog(operation string, archivePath string) *ArchiveProgressDialog {
	return &ArchiveProgressDialog{
		operation:   operation,
		archivePath: archivePath,
		progress:    nil,
		active:      true,
		width:       70,
		onCancel:    nil,
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
	d.active = false
}

// Update はメッセージを処理
func (d *ArchiveProgressDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
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
	if !d.active {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	progressBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62"))

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("246"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var title string
	if d.operation == "compress" {
		title = "Compressing Archive"
	} else {
		title = "Extracting Archive"
	}

	var content string
	content += titleStyle.Render(title) + "\n"
	content += infoStyle.Render(d.archivePath) + "\n\n"

	if d.progress != nil {
		// プログレスバー
		percentage := d.progress.Percentage()
		barWidth := 50
		filledWidth := (percentage * barWidth) / 100
		emptyWidth := barWidth - filledWidth

		bar := progressBarStyle.Render(
			"[" + string(make([]rune, filledWidth)) + string(make([]rune, emptyWidth)) + "]",
		)
		for i := 0; i < filledWidth; i++ {
			bar = bar[:i+1] + "█" + bar[i+2:]
		}
		for i := 0; i < emptyWidth; i++ {
			bar = bar[:filledWidth+i+1] + "░" + bar[filledWidth+i+2:]
		}

		content += bar + fmt.Sprintf(" %d%%\n\n", percentage)

		// ファイル数
		content += infoStyle.Render(
			fmt.Sprintf("Files: %d/%d", d.progress.ProcessedFiles, d.progress.TotalFiles),
		) + "\n"

		// 現在処理中のファイル
		if d.progress.CurrentFile != "" {
			currentFile := d.progress.CurrentFile
			if len(currentFile) > 50 {
				currentFile = "..." + currentFile[len(currentFile)-47:]
			}
			content += infoStyle.Render(fmt.Sprintf("Current: %s", currentFile)) + "\n"
		}

		// 経過時間
		elapsed := d.progress.ElapsedTime()
		content += infoStyle.Render(
			fmt.Sprintf("Elapsed: %s", formatDuration(elapsed)),
		) + "\n"

		// 推定残り時間
		if d.progress.ProcessedFiles > 0 {
			remaining := d.progress.EstimatedRemaining()
			content += infoStyle.Render(
				fmt.Sprintf("Remaining: %s", formatDuration(remaining)),
			) + "\n"
		}
	} else {
		// 進捗情報がまだない場合
		content += infoStyle.Render("Starting...") + "\n"
	}

	content += "\n"
	content += helpStyle.Render("[Esc] Cancel")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(d.width)

	return boxStyle.Render(content)
}

// IsActive はダイアログがアクティブかを返す
func (d *ArchiveProgressDialog) IsActive() bool {
	return d.active
}

// SetActive はダイアログのアクティブ状態を設定
func (d *ArchiveProgressDialog) SetActive(active bool) {
	d.active = active
}

// DisplayType はダイアログの表示タイプを返す
func (d *ArchiveProgressDialog) DisplayType() DialogDisplayType {
	return DialogDisplayScreen
}

// formatDuration は時間を MM:SS 形式にフォーマット
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
