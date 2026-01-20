package ui

import (
	"fmt"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/archive"
)

// handleArchiveMessages はアーカイブ関連のメッセージを処理する
func (m Model) handleArchiveMessages(msg tea.Msg) (Model, tea.Cmd, bool) {
	// アーカイブ操作開始
	if result, ok := msg.(archiveOperationStartMsg); ok {
		if !m.archiveOpManager.IsActive() {
			return m, nil, true
		}
		m.archiveOpManager.SetTaskID(result.taskID)
		return m, m.pollArchiveProgress(result.taskID), true
	}

	// アーカイブ進捗更新
	if result, ok := msg.(archiveProgressUpdateMsg); ok {
		if progressDialog, ok := m.dialog.(*ArchiveProgressDialog); ok {
			progressDialog.UpdateProgress(&archive.ProgressUpdate{
				ProcessedFiles: result.processedFiles,
				TotalFiles:     result.totalFiles,
				CurrentFile:    result.currentFile,
				StartTime:      time.Now().Add(-result.elapsedTime),
			})
		}
		return m, m.pollArchiveProgress(result.taskID), true
	}

	// アーカイブ操作完了
	if result, ok := msg.(archiveOperationCompleteMsg); ok {
		m.dialog = nil
		m.archiveOpManager.Clear()

		if result.cancelled {
			m.statusMessage = "Archive operation cancelled"
			m.isStatusError = false
		} else if result.success {
			m.getActivePane().ClearMarks()
			m.getActivePane().LoadDirectory()
			m.getInactivePane().LoadDirectory()
			m.statusMessage = fmt.Sprintf("Archive created: %s", filepath.Base(result.archivePath))
			m.isStatusError = false
		} else {
			errMsg := "Archive operation failed"
			if result.err != nil {
				errMsg = fmt.Sprintf("Archive operation failed: %v", result.err)
			}
			m.statusMessage = errMsg
			m.isStatusError = true
		}
		return m, statusMessageClearCmd(5 * time.Second), true
	}

	// アーカイブ操作エラー
	if result, ok := msg.(archiveOperationErrorMsg); ok {
		m.dialog = nil
		m.archiveOpManager.Clear()
		m.statusMessage = result.message
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second), true
	}

	// 展開セキュリティチェック結果
	if newModel, cmd, handled := m.handleExtractSecurityCheck(msg); handled {
		return newModel, cmd, true
	}

	// アーカイブ警告ダイアログ結果
	if result, ok := msg.(archiveWarningResultMsg); ok {
		m.dialog = nil
		if result.choice == ArchiveWarningCancel {
			m.statusMessage = "Extraction cancelled"
			m.isStatusError = false
			return m, statusMessageClearCmd(3 * time.Second), true
		}
		destDir := m.getInactivePane().Path()
		return m, m.startArchiveExtraction(result.archivePath, destDir), true
	}

	return m, nil, false
}

// handleExtractSecurityCheck は展開セキュリティチェック結果を処理
func (m Model) handleExtractSecurityCheck(msg tea.Msg) (Model, tea.Cmd, bool) {
	result, ok := msg.(extractSecurityCheckMsg)
	if !ok {
		return m, nil, false
	}

	if result.err != nil {
		m.statusMessage = fmt.Sprintf("Failed to check archive: %v", result.err)
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second), true
	}

	if !result.compressionOK {
		m.dialog = NewCompressionBombWarningDialog(
			result.archivePath,
			result.archiveSize,
			result.extractedSize,
			result.ratio,
		)
		return m, nil, true
	}

	if !result.diskSpaceOK {
		m.dialog = NewDiskSpaceWarningDialog(
			result.archivePath,
			result.extractedSize,
			result.availableSize,
		)
		return m, nil, true
	}

	return m, m.startArchiveExtraction(result.archivePath, result.destDir), true
}

// handleCompressFormatResult は圧縮フォーマット選択の結果を処理
func (m Model) handleCompressFormatResult(msg tea.Msg) (Model, tea.Cmd, bool) {
	result, ok := msg.(compressFormatResultMsg)
	if !ok {
		return m, nil, false
	}

	m.dialog = nil

	if result.cancelled {
		m.archiveOpManager.Clear()
		return m, nil, true
	}

	activePane := m.getActivePane()
	markedFiles := activePane.GetMarkedFiles()
	var sources []string

	if len(markedFiles) > 0 {
		for _, name := range markedFiles {
			sources = append(sources, filepath.Join(activePane.Path(), name))
		}
	} else {
		entry := activePane.SelectedEntry()
		if entry != nil && !entry.IsParentDir() {
			sources = append(sources, filepath.Join(activePane.Path(), entry.Name))
		}
	}

	if len(sources) == 0 {
		m.statusMessage = "No files selected for compression"
		m.isStatusError = true
		return m, statusMessageClearCmd(3 * time.Second), true
	}

	// Prepare compression state via manager
	destDir := m.getInactivePane().Path()
	m.archiveOpManager.PrepareCompression(sources, destDir, result.format, 6, "")

	if result.format == archive.FormatTar {
		defaultName := m.generateDefaultArchiveName(sources, result.format)
		m.dialog = NewArchiveNameDialog(defaultName)
	} else {
		m.dialog = NewCompressionLevelDialog()
	}
	return m, nil, true
}

// handleCompressionLevelResult は圧縮レベル選択の結果を処理
func (m Model) handleCompressionLevelResult(msg tea.Msg) (Model, tea.Cmd, bool) {
	result, ok := msg.(compressionLevelResultMsg)
	if !ok {
		return m, nil, false
	}

	m.dialog = nil

	state := m.archiveOpManager.State()
	if result.cancelled || state == nil {
		m.archiveOpManager.Clear()
		return m, nil, true
	}

	state.Level = result.level
	defaultName := m.generateDefaultArchiveName(state.Sources, state.Format)
	m.dialog = NewArchiveNameDialog(defaultName)
	return m, nil, true
}

// handleArchiveNameResult はアーカイブ名入力の結果を処理
func (m Model) handleArchiveNameResult(msg tea.Msg) (Model, tea.Cmd, bool) {
	result, ok := msg.(archiveNameResultMsg)
	if !ok {
		return m, nil, false
	}

	m.dialog = nil

	state := m.archiveOpManager.State()
	if result.cancelled || state == nil {
		m.archiveOpManager.Clear()
		return m, nil, true
	}

	state.ArchiveName = result.name
	archivePath := filepath.Join(state.DestDir, result.name)

	exists, err := fileExists(archivePath)
	if err != nil {
		m.dialog = NewErrorDialog(fmt.Sprintf("Cannot check file: %v", err))
		m.archiveOpManager.Clear()
		return m, nil, true
	}
	if exists {
		m.dialog = NewArchiveConflictDialog(archivePath)
		return m, nil, true
	}

	return m, m.startArchiveCompression(archivePath), true
}

// handleArchiveConflictResult はアーカイブ衝突解決の結果を処理
func (m Model) handleArchiveConflictResult(msg tea.Msg) (Model, tea.Cmd, bool) {
	result, ok := msg.(archiveConflictResultMsg)
	if !ok {
		return m, nil, false
	}

	m.dialog = nil

	state := m.archiveOpManager.State()
	if result.cancelled || state == nil {
		m.archiveOpManager.Clear()
		return m, nil, true
	}

	archivePath := result.archivePath

	switch result.choice {
	case ArchiveConflictOverwrite:
		if err := removeFile(archivePath); err != nil {
			m.statusMessage = fmt.Sprintf("Failed to remove existing file: %v", err)
			m.isStatusError = true
			m.archiveOpManager.Clear()
			return m, statusMessageClearCmd(5 * time.Second), true
		}
		return m, m.startArchiveCompression(archivePath), true

	case ArchiveConflictRename:
		newPath := GenerateUniqueArchiveName(archivePath)
		newName := filepath.Base(newPath)
		m.dialog = NewArchiveNameDialog(newName)
		return m, nil, true

	case ArchiveConflictCancel:
		m.archiveOpManager.Clear()
		return m, nil, true
	}

	return m, nil, true
}
