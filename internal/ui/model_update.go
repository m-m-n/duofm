package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/archive"
	"github.com/sakura/duofm/internal/fs"
)

// Update はメッセージを処理してモデルを更新
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// カスタムメッセージの処理を優先
	if newModel, cmd, handled := m.handleCustomMessages(msg); handled {
		return newModel, cmd
	}

	// システムメッセージの処理
	return m.handleSystemMessages(msg)
}

// handleCustomMessages はカスタムメッセージを処理する
// 処理された場合は handled=true を返す
func (m Model) handleCustomMessages(msg tea.Msg) (Model, tea.Cmd, bool) {
	// コンテキストメニュー結果
	if newModel, cmd, handled := m.handleContextMenuResult(msg); handled {
		return newModel, cmd, true
	}

	// ダイアログ関連メッセージ
	if newModel, cmd, handled := m.handleDialogMessages(msg); handled {
		return newModel, cmd, true
	}

	// ブックマーク関連メッセージ
	if newModel, cmd, handled := m.handleBookmarkMessages(msg); handled {
		return newModel, cmd, true
	}

	// Path Jump関連メッセージ
	if newModel, cmd, handled := m.handlePathJumpMessages(msg); handled {
		return newModel, cmd, true
	}

	// アーカイブ関連メッセージ
	if newModel, cmd, handled := m.handleArchiveMessages(msg); handled {
		return newModel, cmd, true
	}

	// パーミッション関連メッセージ
	if newModel, cmd, handled := m.handlePermissionMessages(msg); handled {
		return newModel, cmd, true
	}

	return m, nil, false
}

// handleContextMenuResult はコンテキストメニューの結果を処理する
func (m Model) handleContextMenuResult(msg tea.Msg) (Model, tea.Cmd, bool) {
	result, ok := msg.(contextMenuResultMsg)
	if !ok {
		return m, nil, false
	}

	prevDialog := m.dialog
	m.dialog = nil

	if _, ok := prevDialog.(*ContextMenuDialog); !ok {
		return m, nil, true
	}

	if result.cancelled {
		return m, nil, true
	}

	activePane := m.getActivePane()
	markedFiles := activePane.GetMarkedFiles()

	// Open with xdg-open
	if result.actionID == "open" {
		entry := activePane.SelectedEntry()
		if entry != nil {
			// Support parent directory (..) per US4
			var target string
			if entry.IsParentDir() {
				target = ".."
			} else {
				target = entry.Name
			}
			return m, openWithXDG(target, activePane.Path()), true
		}
		return m, nil, true
	}

	// Open with custom application
	if result.actionID == "open_with" {
		entry := activePane.SelectedEntry()
		var files []string

		if len(markedFiles) > 0 {
			// Use marked files
			files = markedFiles
		} else if entry != nil && !entry.IsParentDir() {
			// Use selected file
			files = []string{entry.Name}
		} else {
			return m, nil, true
		}

		m.dialog = NewOpenWithDialog(files, activePane.Path())
		return m, nil, true
	}

	// 削除の場合は確認ダイアログを表示
	if result.actionID == "delete" {
		if len(markedFiles) > 0 {
			m.dialog = NewConfirmDialog(
				fmt.Sprintf("Delete %d files?", len(markedFiles)),
				"This action cannot be undone.",
			)
		} else {
			entry := activePane.SelectedEntry()
			if entry != nil && !entry.IsParentDir() {
				m.pendingAction = result.action
				m.dialog = NewConfirmDialog(
					"Delete file?",
					entry.DisplayName(),
				)
			}
		}
		return m, nil, true
	}

	// 圧縮の場合
	if result.actionID == "compress" {
		m.dialog = NewCompressFormatDialog()
		return m, nil, true
	}

	// 展開の場合
	if result.actionID == "extract" {
		entry := activePane.SelectedEntry()
		if entry != nil && !entry.IsParentDir() {
			archivePath := filepath.Join(activePane.Path(), entry.Name)
			destDir := m.getInactivePane().Path()
			return m, m.checkExtractSecurity(archivePath, destDir), true
		}
		return m, nil, true
	}

	// コピー/移動の場合
	if result.actionID == "copy" || result.actionID == "move" {
		if len(markedFiles) > 0 {
			return m, m.startBatchOperation(markedFiles, result.actionID), true
		}
		entry := activePane.SelectedEntry()
		if entry != nil && !entry.IsParentDir() {
			srcPath := filepath.Join(activePane.Path(), entry.Name)
			destPath := m.getInactivePane().Path()
			return m, m.checkFileConflict(srcPath, destPath, result.actionID), true
		}
		return m, nil, true
	}

	// その他のアクションは直接実行
	if result.action != nil {
		if err := result.action(); err != nil {
			m.dialog = NewErrorDialog(fmt.Sprintf("Operation failed: %v", err))
			return m, nil, true
		}
		activePane.LoadDirectory()
		m.getInactivePane().LoadDirectory()
	}

	return m, nil, true
}

// handleDialogMessages はダイアログ関連のメッセージを処理する
func (m Model) handleDialogMessages(msg tea.Msg) (Model, tea.Cmd, bool) {
	// Regex search dialog result
	if result, ok := msg.(regexSearchResultMsg); ok {
		m.dialog = nil
		if result.cancelled {
			return m, nil, true
		}
		pane := m.getActivePane()
		if result.pattern == "" {
			pane.ClearFilter()
		} else {
			if err := pane.ApplyFilter(result.pattern, SearchModeRegex); err != nil {
				m.statusMessage = fmt.Sprintf("Regex error: %v", err)
				m.isStatusError = true
				return m, statusMessageClearCmd(5 * time.Second), true
			}
		}
		return m, nil, true
	}

	// Query search dialog result
	if result, ok := msg.(querySearchResultMsg); ok {
		m.dialog = nil
		if result.cancelled {
			return m, nil, true
		}
		pane := m.getActivePane()
		if result.query == "" {
			pane.ClearFilter()
		} else {
			if err := pane.ApplyFilter(result.query, SearchModeSQLLike); err != nil {
				m.statusMessage = fmt.Sprintf("Query error: %v", err)
				m.isStatusError = true
				return m, statusMessageClearCmd(5 * time.Second), true
			}
		}
		return m, nil, true
	}

	// ソートダイアログの結果処理
	if result, ok := msg.(sortDialogResultMsg); ok {
		m.sortDialog = nil
		if result.cancelled {
			m.getActivePane().SetSortConfig(result.config)
			m.getActivePane().ApplySortAndPreserveCursor()
		}
		return m, nil, true
	}

	// ソートダイアログの設定変更（ライブプレビュー）
	if result, ok := msg.(sortDialogConfigChangedMsg); ok {
		if m.sortDialog != nil {
			m.getActivePane().SetSortConfig(result.config)
			m.getActivePane().ApplySortAndPreserveCursor()
		}
		return m, nil, true
	}

	// 圧縮フォーマット選択の結果処理
	if newModel, cmd, handled := m.handleCompressFormatResult(msg); handled {
		return newModel, cmd, true
	}

	// 圧縮レベル選択の結果処理
	if newModel, cmd, handled := m.handleCompressionLevelResult(msg); handled {
		return newModel, cmd, true
	}

	// アーカイブ名入力の結果処理
	if newModel, cmd, handled := m.handleArchiveNameResult(msg); handled {
		return newModel, cmd, true
	}

	// アーカイブ衝突解決の結果処理
	if newModel, cmd, handled := m.handleArchiveConflictResult(msg); handled {
		return newModel, cmd, true
	}

	// 上書き確認ダイアログの結果処理
	if newModel, cmd, handled := m.handleOverwriteDialogResult(msg); handled {
		return newModel, cmd, true
	}

	// 確認ダイアログの結果処理
	if newModel, cmd, handled := m.handleConfirmDialogResult(msg); handled {
		return newModel, cmd, true
	}

	// ステータスメッセージ処理
	if result, ok := msg.(showStatusMsg); ok {
		m.dialog = nil
		m.statusMessage = result.message
		m.isStatusError = result.isError
		duration := 3 * time.Second
		if result.isError {
			duration = 5 * time.Second
		}
		return m, statusMessageClearCmd(duration), true
	}

	// Open with xdg-open finished
	if result, ok := msg.(openWithFinishedMsg); ok {
		if result.err != nil {
			// Check if it's "command not found" error
			errStr := result.err.Error()
			if strings.Contains(errStr, "executable file not found") || strings.Contains(errStr, "not found") {
				m.statusMessage = "Cannot open file: xdg-open not found. Install xdg-utils package."
			} else {
				m.statusMessage = fmt.Sprintf("Failed to open file: %v", result.err)
			}
			m.isStatusError = true
			return m, statusMessageClearCmd(5 * time.Second), true
		}
		m.statusMessage = "Opened with xdg-open"
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true
	}

	// Open with custom application result
	if result, ok := msg.(openWithDialogResultMsg); ok {
		m.dialog = nil

		if result.cancelled {
			return m, nil, true
		}

		// Launch custom application
		return m, openWithCustom(result.application, result.files, result.workDir), true
	}

	return m, nil, false
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

// handleOverwriteDialogResult は上書き確認ダイアログの結果を処理
func (m Model) handleOverwriteDialogResult(msg tea.Msg) (Model, tea.Cmd, bool) {
	result, ok := msg.(overwriteDialogResultMsg)
	if !ok {
		return m, nil, false
	}

	m.dialog = nil

	switch result.choice {
	case OverwriteChoiceOverwrite:
		destFile := filepath.Join(result.destPath, result.filename)
		if err := removeAllFiles(destFile); err != nil {
			if isPermissionError(err) {
				m.dialog = NewErrorDialog("Permission denied: cannot remove existing file")
			} else {
				m.dialog = NewErrorDialog(fmt.Sprintf("Failed to remove: %v", err))
			}
			if m.batchOpManager.IsActive() {
				return m, m.cancelBatchOperation(), true
			}
			return m, nil, true
		}
		return m, m.executeFileOperation(result.srcPath, result.destPath, result.operation), true

	case OverwriteChoiceCancel:
		if m.batchOpManager.IsActive() {
			return m, m.cancelBatchOperation(), true
		}
		return m, nil, true

	case OverwriteChoiceRename:
		m.dialog = NewRenameInputDialog(result.destPath, result.srcPath, result.operation)
		return m, nil, true
	}

	return m, nil, true
}

// handleConfirmDialogResult は確認ダイアログの結果を処理
func (m Model) handleConfirmDialogResult(msg tea.Msg) (Model, tea.Cmd, bool) {
	result, ok := msg.(dialogResultMsg)
	if !ok {
		return m, nil, false
	}

	prevDialog := m.dialog
	m.dialog = nil

	if !result.result.Confirmed {
		m.pendingAction = nil
		return m, nil, true
	}

	if _, ok := prevDialog.(*ConfirmDialog); !ok {
		return m, nil, true
	}

	// コンテキストメニューからの削除
	if m.pendingAction != nil {
		if err := m.pendingAction(); err != nil {
			m.dialog = NewErrorDialog(fmt.Sprintf("Failed to delete: %v", err))
		} else {
			m.getActivePane().LoadDirectory()
			m.getInactivePane().LoadDirectory()
		}
		m.pendingAction = nil
		return m, nil, true
	}

	// 通常の削除
	return m.executeDeleteOperation(), nil, true
}

// executeDeleteOperation は削除操作を実行
func (m Model) executeDeleteOperation() Model {
	activePane := m.getActivePane()
	markedFiles := activePane.GetMarkedFiles()

	if len(markedFiles) > 0 {
		// Find the smallest index among marked files (first marked file position)
		minMarkedIndex := -1
		for i, entry := range activePane.entries {
			if activePane.markedFiles[entry.Name] {
				if minMarkedIndex == -1 || i < minMarkedIndex {
					minMarkedIndex = i
				}
			}
		}

		// Delete multiple files
		var deleteErr error
		successCount := 0
		for _, name := range markedFiles {
			fullPath := filepath.Join(activePane.Path(), name)
			if err := deleteFile(fullPath); err != nil {
				deleteErr = err
				break
			}
			successCount++
		}

		// Only reload and clear marks if at least one file was deleted successfully
		if successCount > 0 {
			activePane.ClearMarks()
			activePane.LoadDirectory()

			// Calculate and set new cursor position based on first marked file position
			if minMarkedIndex >= 0 {
				newCursor := activePane.calculateCursorAfterDeletion(minMarkedIndex)
				activePane.SetCursor(newCursor)
				activePane.EnsureCursorVisible()
			}
		}

		// Show error dialog if any deletion failed
		if deleteErr != nil {
			if successCount > 0 {
				m.dialog = NewErrorDialog(fmt.Sprintf("Partially failed to delete: %v (deleted %d of %d files)", deleteErr, successCount, len(markedFiles)))
			} else {
				m.dialog = NewErrorDialog(fmt.Sprintf("Failed to delete: %v", deleteErr))
			}
		}
	} else {
		entry := activePane.SelectedEntry()
		if entry != nil && !entry.IsParentDir() {
			// Remember cursor position before deletion
			cursorBeforeDeletion := activePane.cursor

			fullPath := filepath.Join(activePane.Path(), entry.Name)
			if err := deleteFile(fullPath); err != nil {
				m.dialog = NewErrorDialog(fmt.Sprintf("Failed to delete: %v", err))
				// Don't reload or adjust cursor on error for single file
			} else {
				activePane.LoadDirectory()

				// Calculate and set new cursor position
				newCursor := activePane.calculateCursorAfterDeletion(cursorBeforeDeletion)
				activePane.SetCursor(newCursor)
				activePane.EnsureCursorVisible()
			}
		}
	}

	return m
}

// handleBookmarkMessages はブックマーク関連のメッセージを処理する
func (m Model) handleBookmarkMessages(msg tea.Msg) (Model, tea.Cmd, bool) {
	// ブックマークジャンプ
	if result, ok := msg.(bookmarkJumpMsg); ok {
		m.dialog = nil
		cmd := m.getActivePane().ChangeDirectoryAsync(result.path)
		return m, cmd, true
	}

	// ブックマーク削除
	if result, ok := msg.(bookmarkDeleteMsg); ok {
		m.dialog = nil
		cmd := m.bookmarkManager.Delete(result.index)
		return m, cmd, true
	}

	// ブックマーク削除完了
	if result, ok := msg.(bookmarkDeletedMsg); ok {
		m.bookmarkManager.SetBookmarks(result.bookmarks)
		m.statusMessage = "Bookmark removed"
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true
	}

	// ブックマーク編集
	if result, ok := msg.(bookmarkEditMsg); ok {
		m.dialog = nil
		m.bookmarkManager.SetEditIndex(result.index)
		editIndex := result.index
		dialog := NewInputDialog("Edit bookmark name:", func(newAlias string) tea.Cmd {
			return m.bookmarkManager.Edit(editIndex, newAlias)
		})
		dialog.SetEmptyErrorMsg("Bookmark name cannot be empty")
		dialog.SetInput(result.bookmark.Name)
		m.dialog = dialog
		return m, nil, true
	}

	// ブックマークダイアログ閉じる
	if _, ok := msg.(bookmarkCloseMsg); ok {
		m.dialog = nil
		return m, nil, true
	}

	// ブックマーク追加完了
	if result, ok := msg.(bookmarkAddedMsg); ok {
		m.dialog = nil
		m.bookmarkManager.SetBookmarks(result.bookmarks)
		m.statusMessage = fmt.Sprintf("Bookmarked: %s", result.alias)
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true
	}

	// ブックマーク編集完了
	if result, ok := msg.(bookmarkEditedMsg); ok {
		m.dialog = nil
		m.bookmarkManager.SetBookmarks(result.bookmarks)
		m.bookmarkManager.ClearEditIndex()
		m.statusMessage = fmt.Sprintf("Bookmark updated: %s", result.alias)
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true
	}

	return m, nil, false
}

// handlePathJumpMessages はPath Jump関連のメッセージを処理する
func (m Model) handlePathJumpMessages(msg tea.Msg) (Model, tea.Cmd, bool) {
	// Path Jumpジャンプ
	if result, ok := msg.(pathJumpResultMsg); ok {
		m.dialog = nil
		cmd := m.getActivePane().ChangeDirectoryAsync(result.path)
		return m, cmd, true
	}

	// Path Jumpキャンセル
	if _, ok := msg.(pathJumpCancelMsg); ok {
		m.dialog = nil
		return m, nil, true
	}

	return m, nil, false
}

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

// handleSystemMessages はシステムメッセージを処理する
func (m Model) handleSystemMessages(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case diskSpaceUpdateMsg:
		m.updateDiskSpace()
		return m, diskSpaceTickCmd()

	case clearStatusMsg:
		m.statusMessage = ""
		m.isStatusError = false
		return m, nil

	case ctrlCTimeoutMsg:
		if m.ctrlCPending {
			m.ctrlCPending = false
			m.statusMessage = ""
		}
		return m, nil

	case directoryLoadCompleteMsg:
		return m.handleDirectoryLoadComplete(msg)

	case execFinishedMsg:
		return m.handleExecFinished(msg)

	case shellCommandFinishedMsg:
		return m.handleShellCommandFinished(msg)

	case inputDialogResultMsg:
		return m.handleInputDialogResult(msg)

	case showErrorDialogMsg:
		m.dialog = NewErrorDialog(msg.message)
		return m, nil

	case showOverwriteDialogMsg:
		m.dialog = NewOverwriteDialog(
			msg.filename,
			msg.destPath,
			msg.srcInfo,
			msg.destInfo,
			msg.operation,
			msg.srcPath,
		)
		return m, nil

	case fileOperationCompleteMsg:
		return m.handleFileOperationComplete(msg)

	case batchCompleteMsg:
		// Clear marks and reload panes after batch operation completes
		m.getActivePane().ClearMarks()
		m.getActivePane().LoadDirectory()
		m.getInactivePane().LoadDirectory()

		m.statusMessage = fmt.Sprintf("%s %d files completed", strings.Title(msg.operation), msg.completed)
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second)

	case batchCancelledMsg:
		// Clear marks and reload panes after cancellation
		m.getActivePane().ClearMarks()
		m.getActivePane().LoadDirectory()
		m.getInactivePane().LoadDirectory()

		m.statusMessage = fmt.Sprintf("%s cancelled (%d completed, %d remaining)",
			strings.Title(msg.operation), msg.completed, msg.remaining)
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second)

	case batchNextFileMsg:
		// Process the next file in the batch
		return m, m.checkFileConflict(msg.srcPath, msg.destPath, m.batchOpManager.Operation())

	case renameInputResultMsg:
		return m.handleRenameInputResult(msg)

	case extensionRenameResultMsg:
		return m.handleExtensionRenameResult(msg)

	case tea.KeyMsg:
		return m.handleKeyInput(msg)
	}

	return m, nil
}

// handleWindowSize はウィンドウサイズ変更を処理
func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	if !m.ready {
		paneWidth := msg.Width / 2
		paneHeight := msg.Height - 2

		var err error
		m.leftPane, err = NewPane(LeftPane, m.leftPath, paneWidth, paneHeight, true, m.theme)
		if err != nil {
			return m, tea.Quit
		}

		m.rightPane, err = NewPane(RightPane, m.rightPath, paneWidth, paneHeight, false, m.theme)
		if err != nil {
			return m, tea.Quit
		}

		m.updateDiskSpace()
		m.ready = true
		return m, diskSpaceTickCmd()
	}

	paneWidth := msg.Width / 2
	paneHeight := msg.Height - 2
	m.leftPane.SetSize(paneWidth, paneHeight)
	m.rightPane.SetSize(paneWidth, paneHeight)

	return m, nil
}

// handleDirectoryLoadComplete はディレクトリ読み込み完了を処理
func (m Model) handleDirectoryLoadComplete(msg directoryLoadCompleteMsg) (tea.Model, tea.Cmd) {
	var targetPane *Pane
	if msg.paneID == LeftPane {
		targetPane = m.leftPane
	} else if msg.paneID == RightPane {
		targetPane = m.rightPane
	}

	if targetPane == nil {
		return m, nil
	}

	if targetPane.pendingPath != "" && targetPane.pendingPath != msg.panePath {
		return m, nil
	}

	targetPane.loading = false
	targetPane.loadingProgress = ""

	if msg.err != nil {
		targetPane.restorePreviousPath()
		targetPane.pendingCursorTarget = ""

		if msg.isHistoryNavigation {
			if msg.historyNavigationForward {
				targetPane.history.NavigateBack()
			} else {
				targetPane.history.NavigateForward()
			}
		}

		m.statusMessage = formatDirectoryError(msg.err, msg.attemptedPath)
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second)
	}

	entries := msg.entries
	if !targetPane.showHidden {
		entries = filterHiddenFiles(entries)
	}
	targetPane.allEntries = entries
	targetPane.entries = entries
	targetPane.filterPattern = ""
	targetPane.filterMode = SearchModeNone

	if targetPane.pendingCursorTarget != "" {
		if index := targetPane.findEntryIndex(targetPane.pendingCursorTarget); index >= 0 {
			targetPane.cursor = index
		} else {
			targetPane.cursor = 0
		}
		targetPane.pendingCursorTarget = ""
	} else {
		targetPane.cursor = 0
	}

	targetPane.scrollOffset = 0
	targetPane.adjustScroll()
	targetPane.pendingPath = ""

	// Gitブランチを更新
	targetPane.gitBranch = fs.GetGitBranch(targetPane.path)

	if !msg.isHistoryNavigation {
		targetPane.addToHistory()
	}

	m.updateDiskSpace()
	return m, nil
}

// handleExecFinished は外部コマンド完了を処理
func (m Model) handleExecFinished(msg execFinishedMsg) (tea.Model, tea.Cmd) {
	m.getActivePane().RefreshDirectoryPreserveCursor()
	m.getInactivePane().RefreshDirectoryPreserveCursor()

	if msg.err != nil {
		m.statusMessage = fmt.Sprintf("Command failed: %v", msg.err)
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second)
	}
	return m, nil
}

// handleShellCommandFinished はシェルコマンド完了を処理
func (m Model) handleShellCommandFinished(msg shellCommandFinishedMsg) (tea.Model, tea.Cmd) {
	m.getActivePane().RefreshDirectoryPreserveCursor()
	m.getInactivePane().RefreshDirectoryPreserveCursor()

	if msg.err != nil {
		m.statusMessage = fmt.Sprintf("Shell command failed: %v", msg.err)
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second)
	}
	return m, nil
}

// handleInputDialogResult は入力ダイアログの結果を処理
func (m Model) handleInputDialogResult(msg inputDialogResultMsg) (tea.Model, tea.Cmd) {
	m.dialog = nil

	// If cancelled, do nothing (just clear dialog)
	if msg.cancelled {
		return m, nil
	}

	if msg.err != nil {
		m.statusMessage = msg.err.Error()
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second)
	}

	m.getActivePane().LoadDirectory()
	m.getInactivePane().LoadDirectory()

	switch msg.operation {
	case "create_file", "create_dir":
		m.moveCursorToFile(msg.input)
	case "rename":
		m.moveCursorToFileAfterRename(msg.oldName, msg.input)
	}
	return m, nil
}

// handleFileOperationComplete はファイル操作完了を処理
func (m Model) handleFileOperationComplete(msg fileOperationCompleteMsg) (tea.Model, tea.Cmd) {
	if m.batchOpManager.IsActive() {
		srcPath := m.batchOpManager.CurrentFile()
		return m, m.advanceBatchOperation(true, srcPath)
	}
	m.getActivePane().LoadDirectory()
	m.getInactivePane().LoadDirectory()
	return m, nil
}

// handleExtensionRenameResult は拡張子保持リネームダイアログの結果を処理
func (m Model) handleExtensionRenameResult(msg extensionRenameResultMsg) (tea.Model, tea.Cmd) {
	m.dialog = nil

	// Handle cancellation
	if msg.cancelled {
		return m, nil
	}

	// Execute rename operation
	oldPath := filepath.Join(msg.dirPath, msg.oldName)
	newPath := filepath.Join(msg.dirPath, msg.newName)

	if err := fs.Rename(oldPath, newPath); err != nil {
		m.dialog = NewErrorDialog(fmt.Sprintf("Failed to rename: %v", err))
		return m, nil
	}

	// Refresh panes
	m.getActivePane().LoadDirectory()
	m.getInactivePane().LoadDirectory()

	// Move cursor to renamed file
	m.moveCursorToFileAfterRename(msg.oldName, msg.newName)

	return m, nil
}

// handleRenameInputResult はリネーム入力ダイアログの結果を処理
func (m Model) handleRenameInputResult(msg renameInputResultMsg) (tea.Model, tea.Cmd) {
	m.dialog = nil

	// Handle cancellation
	if msg.cancelled {
		return m, nil
	}

	newDestPath := filepath.Join(msg.destPath, msg.newName)

	var err error
	if msg.operation == "copy" {
		err = copyFile(msg.srcPath, newDestPath)
	} else {
		err = moveFile(msg.srcPath, newDestPath)
	}

	if err != nil {
		m.dialog = NewErrorDialog(fmt.Sprintf("Failed to %s: %v", msg.operation, err))
	} else {
		m.getActivePane().LoadDirectory()
		m.getInactivePane().LoadDirectory()
	}
	return m, nil
}
