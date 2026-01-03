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
	totalFiles     int
	processedFiles int
	currentFile    string
	startTime      time.Time
	active         bool
	width          int
	onCancel       func()
}

// NewPermissionProgressDialog は新しい進捗表示ダイアログを作成
func NewPermissionProgressDialog(totalFiles int) *PermissionProgressDialog {
	return &PermissionProgressDialog{
		totalFiles:     totalFiles,
		processedFiles: 0,
		currentFile:    "",
		startTime:      time.Now(),
		active:         true,
		width:          70,
		onCancel:       nil,
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
	d.active = false
}

// Update はメッセージを処理
func (d *PermissionProgressDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			// Esc or Ctrl+Cでキャンセル
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

	var content string
	content += titleStyle.Render("Changing Permissions") + "\n\n"

	// プログレスバー
	percentage := 0
	if d.totalFiles > 0 {
		percentage = (d.processedFiles * 100) / d.totalFiles
	}

	barWidth := 50
	filledWidth := (percentage * barWidth) / 100
	emptyWidth := barWidth - filledWidth

	// プログレスバーを構築
	bar := strings.Repeat("▓", filledWidth) + strings.Repeat("░", emptyWidth)
	content += progressBarStyle.Render(fmt.Sprintf("[%s] %d%%", bar, percentage)) + "\n\n"

	// ファイル数
	content += infoStyle.Render(
		fmt.Sprintf("Progress: %d / %d files", d.processedFiles, d.totalFiles),
	) + "\n"

	// 現在処理中のファイル
	if d.currentFile != "" {
		currentFile := d.currentFile
		// 長いパスは省略
		maxPathLen := 60
		if len(currentFile) > maxPathLen {
			currentFile = "..." + currentFile[len(currentFile)-(maxPathLen-3):]
		}
		content += infoStyle.Render(fmt.Sprintf("Current: %s", currentFile)) + "\n"
	}

	// 経過時間
	elapsed := time.Since(d.startTime)
	content += "\n"
	content += infoStyle.Render(
		fmt.Sprintf("Elapsed: %s", formatDuration(elapsed)),
	) + "\n"

	content += "\n"
	content += helpStyle.Render("[Ctrl+C] Cancel")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(d.width)

	return boxStyle.Render(content)
}

// IsActive はダイアログがアクティブかを返す
func (d *PermissionProgressDialog) IsActive() bool {
	return d.active
}

// SetActive はダイアログのアクティブ状態を設定
func (d *PermissionProgressDialog) SetActive(active bool) {
	d.active = active
}

// DisplayType はダイアログの表示タイプを返す
func (d *PermissionProgressDialog) DisplayType() DialogDisplayType {
	return DialogDisplayScreen
}
