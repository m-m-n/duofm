package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/archive"
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
		m.archiveOp = nil
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

	m.archiveOp = &ArchiveOperationState{
		Sources: sources,
		DestDir: m.getInactivePane().Path(),
		Format:  result.format,
		Level:   6,
	}

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

	if result.cancelled || m.archiveOp == nil {
		m.archiveOp = nil
		return m, nil, true
	}

	m.archiveOp.Level = result.level
	defaultName := m.generateDefaultArchiveName(m.archiveOp.Sources, m.archiveOp.Format)
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

	if result.cancelled || m.archiveOp == nil {
		m.archiveOp = nil
		return m, nil, true
	}

	m.archiveOp.ArchiveName = result.name
	archivePath := filepath.Join(m.archiveOp.DestDir, result.name)

	exists, err := fileExists(archivePath)
	if err != nil {
		m.dialog = NewErrorDialog(fmt.Sprintf("Cannot check file: %v", err))
		m.archiveOp = nil
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

	if result.cancelled || m.archiveOp == nil {
		m.archiveOp = nil
		return m, nil, true
	}

	archivePath := result.archivePath

	switch result.choice {
	case ArchiveConflictOverwrite:
		if err := removeFile(archivePath); err != nil {
			m.statusMessage = fmt.Sprintf("Failed to remove existing file: %v", err)
			m.isStatusError = true
			m.archiveOp = nil
			return m, statusMessageClearCmd(5 * time.Second), true
		}
		return m, m.startArchiveCompression(archivePath), true

	case ArchiveConflictRename:
		newPath := GenerateUniqueArchiveName(archivePath)
		newName := filepath.Base(newPath)
		m.dialog = NewArchiveNameDialog(newName)
		return m, nil, true

	case ArchiveConflictCancel:
		m.archiveOp = nil
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
			if m.batchOp != nil {
				m.cancelBatchOperation()
			}
			return m, nil, true
		}
		return m, m.executeFileOperation(result.srcPath, result.destPath, result.operation), true

	case OverwriteChoiceCancel:
		if m.batchOp != nil {
			m.cancelBatchOperation()
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
		var deleteErr error
		for _, name := range markedFiles {
			fullPath := filepath.Join(activePane.Path(), name)
			if err := deleteFile(fullPath); err != nil {
				deleteErr = err
				break
			}
		}
		if deleteErr != nil {
			m.dialog = NewErrorDialog(fmt.Sprintf("Failed to delete: %v", deleteErr))
		}
		activePane.ClearMarks()
		activePane.LoadDirectory()
	} else {
		entry := activePane.SelectedEntry()
		if entry != nil && !entry.IsParentDir() {
			fullPath := filepath.Join(activePane.Path(), entry.Name)
			if err := deleteFile(fullPath); err != nil {
				m.dialog = NewErrorDialog(fmt.Sprintf("Failed to delete: %v", err))
			} else {
				activePane.LoadDirectory()
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
		newBookmarks, err := removeBookmark(m.bookmarks, result.index)
		if err != nil {
			m.statusMessage = fmt.Sprintf("Failed to remove bookmark: %v", err)
			m.isStatusError = true
			return m, statusMessageClearCmd(5 * time.Second), true
		}
		if saveErr := saveBookmarksToConfig(newBookmarks); saveErr != nil {
			m.statusMessage = saveErr.Error()
			m.isStatusError = true
			return m, statusMessageClearCmd(5 * time.Second), true
		}
		m.bookmarks = newBookmarks
		m.statusMessage = "Bookmark removed"
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true
	}

	// ブックマーク編集
	if result, ok := msg.(bookmarkEditMsg); ok {
		m.dialog = nil
		m.bookmarkEditIndex = result.index
		currentBookmarks := m.bookmarks
		editIndex := result.index
		dialog := NewInputDialog("Edit bookmark name:", func(newAlias string) tea.Cmd {
			return m.handleBookmarkEdit(currentBookmarks, editIndex, newAlias)
		})
		dialog.SetEmptyErrorMsg("Bookmark name cannot be empty")
		dialog.input = result.bookmark.Name
		dialog.cursorPos = len(result.bookmark.Name)
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
		m.bookmarks = result.bookmarks
		m.statusMessage = fmt.Sprintf("Bookmarked: %s", result.alias)
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true
	}

	// ブックマーク編集完了
	if result, ok := msg.(bookmarkEditedMsg); ok {
		m.dialog = nil
		m.bookmarks = result.bookmarks
		m.bookmarkEditIndex = -1
		m.statusMessage = fmt.Sprintf("Bookmark updated: %s", result.alias)
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true
	}

	return m, nil, false
}

// handleArchiveMessages はアーカイブ関連のメッセージを処理する
func (m Model) handleArchiveMessages(msg tea.Msg) (Model, tea.Cmd, bool) {
	// アーカイブ操作開始
	if result, ok := msg.(archiveOperationStartMsg); ok {
		if m.archiveOp == nil {
			return m, nil, true
		}
		m.archiveOp.TaskID = result.taskID
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
		m.archiveOp = nil

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
		m.archiveOp = nil
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

	case batchOperationCompleteMsg:
		m.statusMessage = fmt.Sprintf("%s %d files completed", strings.Title(msg.operation), msg.count)
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second)

	case renameInputResultMsg:
		return m.handleRenameInputResult(msg)

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
	if m.batchOp != nil {
		srcPath := m.batchOp.Files[m.batchOp.CurrentIdx]
		return m, m.advanceBatchOperation(true, srcPath)
	}
	m.getActivePane().LoadDirectory()
	m.getInactivePane().LoadDirectory()
	return m, nil
}

// handleRenameInputResult はリネーム入力ダイアログの結果を処理
func (m Model) handleRenameInputResult(msg renameInputResultMsg) (tea.Model, tea.Cmd) {
	m.dialog = nil

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
