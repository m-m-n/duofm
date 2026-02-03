package ui

import (
	"fmt"
	"maps"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

	// トラッシュ関連メッセージ
	if newModel, cmd, handled := m.handleTrashMessages(msg); handled {
		return newModel, cmd, true
	}

	// 設定リロード関連メッセージ
	if newModel, cmd, handled := m.handleConfigMessages(msg); handled {
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
		activePane.RefreshDirectoryPreserveCursor()
		m.getInactivePane().RefreshDirectoryPreserveCursor()
	}

	return m, nil, true
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
		m.refreshPanesAfterBatchOperation(msg.operation)

		m.statusMessage = fmt.Sprintf("%s %d files completed", cases.Title(language.English).String(msg.operation), msg.completed)
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second)

	case batchCancelledMsg:
		m.refreshPanesAfterBatchOperation(msg.operation)

		m.statusMessage = fmt.Sprintf("%s cancelled (%d completed, %d remaining)",
			cases.Title(language.English).String(msg.operation), msg.completed, msg.remaining)
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

// refreshPanesAfterBatchOperation refreshes both panes after a batch operation.
// For move operations, it calculates the cursor target using mark information
// before clearing marks and refreshing. For copy operations, the source files
// remain so filename match handles cursor preservation naturally.
func (m Model) refreshPanesAfterBatchOperation(operation string) {
	activePane := m.getActivePane()
	if operation == "move" {
		// For move: calculate cursor target before clearing marks
		markedCopy := make(map[string]bool, len(activePane.markedFiles))
		maps.Copy(markedCopy, activePane.markedFiles)
		cursorTarget := activePane.calculateCursorTargetAfterBatchMove(markedCopy)
		activePane.ClearMarks()
		activePane.RefreshDirectoryPreserveCursor()
		if idx := activePane.findEntryIndex(cursorTarget); idx >= 0 {
			activePane.SetCursor(idx)
			activePane.EnsureCursorVisible()
		}
	} else {
		// For copy: files remain in source, filename match handles cursor
		activePane.ClearMarks()
		activePane.RefreshDirectoryPreserveCursor()
	}
	m.getInactivePane().RefreshDirectoryPreserveCursor()
}
