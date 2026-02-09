package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
)

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
			// キャンセル時: ライブプレビューで変更された設定を元に戻す
			m.getActivePane().SetSortConfig(result.config)
			m.getActivePane().ApplySortAndPreserveCursor()
		} else {
			// 確定時: 現在の設定をストアに保存
			if m.dirSortStore != nil {
				pane := m.getActivePane()
				sc := pane.GetSortConfig()
				m.dirSortStore.Set(pane.Path(), sortFieldToString(sc.Field), sortOrderToString(sc.Order))
			}
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
			m.getActivePane().RefreshDirectoryPreserveCursor()
			m.getInactivePane().RefreshDirectoryPreserveCursor()
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
			// Note: ClearMarks is called inside ReloadDirectoryWithFilter
			if err := activePane.ReloadDirectoryWithFilter(); err != nil {
				m.dialog = NewErrorDialog(fmt.Sprintf("Failed to reload directory: %v", err))
				return m
			}

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
				if err := activePane.ReloadDirectoryWithFilter(); err != nil {
					m.dialog = NewErrorDialog(fmt.Sprintf("Failed to reload directory: %v", err))
					return m
				}

				// Calculate and set new cursor position
				newCursor := activePane.calculateCursorAfterDeletion(cursorBeforeDeletion)
				activePane.SetCursor(newCursor)
				activePane.EnsureCursorVisible()
			}
		}
	}

	return m
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
		m.getActivePane().RefreshDirectoryPreserveCursor()
		m.getInactivePane().RefreshDirectoryPreserveCursor()
	}
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

	if err := fs.Rename(oldPath, msg.newName); err != nil {
		m.dialog = NewErrorDialog(fmt.Sprintf("Failed to rename: %v", err))
		return m, nil
	}

	// Refresh panes
	m.getActivePane().RefreshDirectoryPreserveCursor()
	m.getInactivePane().RefreshDirectoryPreserveCursor()

	// Move cursor to renamed file
	m.moveCursorToFileAfterRename(msg.oldName, msg.newName)

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

	m.getActivePane().RefreshDirectoryPreserveCursor()
	m.getInactivePane().RefreshDirectoryPreserveCursor()

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
	m.getActivePane().RefreshDirectoryPreserveCursor()
	m.getInactivePane().RefreshDirectoryPreserveCursor()
	return m, nil
}
