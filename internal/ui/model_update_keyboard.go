package ui

import (
	"fmt"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyInput はキーボード入力を処理する
func (m Model) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// ソートダイアログが開いている場合
	if m.sortDialog != nil && m.sortDialog.IsActive() {
		var cmd tea.Cmd
		_, cmd = m.sortDialog.Update(msg)
		return m, cmd
	}

	// ダイアログが開いている場合はダイアログに処理を委譲
	if m.dialog != nil {
		var cmd tea.Cmd
		m.dialog, cmd = m.dialog.Update(msg)
		return m, cmd
	}

	// ミニバッファがアクティブな場合（検索中）
	if m.searchState.IsActive {
		return m.handleSearchInput(msg)
	}

	// シェルコマンドモードの入力処理
	if m.shellCommandMode {
		return m.handleShellCommandInput(msg)
	}

	// Ctrl+Cのダブルプレス処理
	if msg.String() == "ctrl+c" {
		return m.handleCtrlC()
	}

	// ステータスメッセージがあればクリア
	if m.statusMessage != "" || m.ctrlCPending {
		m.statusMessage = ""
		m.isStatusError = false
		m.ctrlCPending = false
	}

	// keybindingMapを使ってアクションを決定
	action := m.keybindingMap.GetAction(msg.String())
	return m.handleAction(action)
}

// handleSearchInput は検索中の入力を処理
func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.confirmSearch()
		return m, nil

	case tea.KeyEsc, tea.KeyCtrlC:
		m.cancelSearch()
		return m, nil

	default:
		if m.minibuffer.HandleKey(msg) {
			m.applyIncrementalFilter()
			return m, nil
		}
	}
	return m, nil
}

// handleShellCommandInput はシェルコマンドモードの入力を処理
func (m Model) handleShellCommandInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		command := m.minibuffer.Input()
		if command == "" {
			m.shellCommandMode = false
			m.minibuffer.Hide()
			return m, nil
		}
		workDir := m.getActivePane().Path()
		m.shellCommandMode = false
		m.minibuffer.Hide()
		return m, executeShellCommand(command, workDir)

	case tea.KeyEsc, tea.KeyCtrlC:
		m.shellCommandMode = false
		m.minibuffer.Hide()
		return m, nil

	default:
		m.minibuffer.HandleKey(msg)
		return m, nil
	}
}

// handleCtrlC はCtrl+Cのダブルプレスを処理
func (m Model) handleCtrlC() (tea.Model, tea.Cmd) {
	if m.ctrlCPending {
		return m, tea.Quit
	}
	m.ctrlCPending = true
	m.statusMessage = "Press Ctrl+C again to quit"
	m.isStatusError = false
	return m, ctrlCTimeoutCmd(2 * time.Second)
}

// handleAction はキーバインドアクションを処理
func (m Model) handleAction(action Action) (tea.Model, tea.Cmd) {
	switch action {
	case ActionRefresh:
		return m, m.RefreshBothPanes()

	case ActionSyncPane:
		m.SyncOppositePane()
		return m, nil

	case ActionQuit:
		return m, tea.Quit

	case ActionHelp:
		m.dialog = NewHelpDialog()
		return m, nil

	case ActionSearch:
		m.startSearch(SearchModeIncremental)
		return m, nil

	case ActionRegexSearch:
		m.startSearch(SearchModeRegex)
		return m, nil

	case ActionShellCommand:
		m.startShellCommandMode()
		return m, nil

	case ActionMoveDown:
		m.getActivePane().MoveCursorDown()
		return m, nil

	case ActionMoveUp:
		m.getActivePane().MoveCursorUp()
		return m, nil

	case ActionMoveLeft:
		return m.handleMoveLeft()

	case ActionMoveRight:
		return m.handleMoveRight()

	case ActionEnter:
		return m.handleEnter()

	case ActionMark:
		return m.handleMark()

	case ActionToggleInfo:
		return m.handleToggleInfo()

	case ActionCopy:
		return m.handleCopy()

	case ActionMove:
		return m.handleMove()

	case ActionDelete:
		return m.handleDelete()

	case ActionContextMenu:
		return m.handleContextMenu()

	case ActionToggleHidden:
		m.getActivePane().ToggleHidden()
		return m, nil

	case ActionHome:
		cmd := m.getActivePane().NavigateToHomeAsync()
		return m, cmd

	case ActionPrevDir:
		cmd := m.getActivePane().NavigateToPreviousAsync()
		return m, cmd

	case ActionHistoryBack:
		cmd := m.getActivePane().NavigateHistoryBackAsync()
		return m, cmd

	case ActionHistoryForward:
		cmd := m.getActivePane().NavigateHistoryForwardAsync()
		return m, cmd

	case ActionView:
		return m.handleView()

	case ActionEdit:
		return m.handleEdit()

	case ActionNewFile:
		return m.handleNewFile()

	case ActionNewDirectory:
		return m.handleNewDirectory()

	case ActionRename:
		return m.handleRenameUI()

	case ActionSort:
		m.sortDialog = NewSortDialog(m.getActivePane().GetSortConfig())
		return m, nil

	case ActionBookmark:
		m.dialog = NewBookmarkDialog(m.bookmarks)
		return m, nil

	case ActionAddBookmark:
		return m.handleAddBookmarkUI()
	}

	return m, nil
}

// handleMoveLeft は左移動を処理
func (m Model) handleMoveLeft() (tea.Model, tea.Cmd) {
	if m.activePane == LeftPane {
		cmd := m.leftPane.MoveToParentAsync()
		return m, cmd
	}
	m.switchToPane(LeftPane)
	return m, nil
}

// handleMoveRight は右移動を処理
func (m Model) handleMoveRight() (tea.Model, tea.Cmd) {
	if m.activePane == RightPane {
		cmd := m.rightPane.MoveToParentAsync()
		return m, cmd
	}
	m.switchToPane(RightPane)
	return m, nil
}

// handleEnter はEnterキーを処理
func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	entry := m.getActivePane().SelectedEntry()
	if entry != nil && !entry.IsParentDir() && !entry.IsDir {
		fullPath := filepath.Join(m.getActivePane().Path(), entry.Name)
		if err := checkReadPermission(fullPath); err != nil {
			m.statusMessage = fmt.Sprintf("Cannot read file: %v", err)
			m.isStatusError = true
			return m, statusMessageClearCmd(5 * time.Second)
		}
		return m, openWithViewer(fullPath, m.getActivePane().Path())
	}
	cmd := m.getActivePane().EnterDirectoryAsync()
	return m, cmd
}

// handleMark はマークを処理
func (m Model) handleMark() (tea.Model, tea.Cmd) {
	activePane := m.getActivePane()
	if activePane.ToggleMark() {
		activePane.MoveCursorDown()
	}
	return m, nil
}

// handleToggleInfo は情報表示切り替えを処理
func (m Model) handleToggleInfo() (tea.Model, tea.Cmd) {
	activePane := m.getActivePane()
	if activePane.CanToggleMode() {
		activePane.ToggleDisplayMode()
	}
	return m, nil
}

// handleCopy はコピーを処理
func (m Model) handleCopy() (tea.Model, tea.Cmd) {
	activePane := m.getActivePane()
	markedFiles := activePane.GetMarkedFiles()

	if len(markedFiles) > 0 {
		return m, m.startBatchOperation(markedFiles, "copy")
	}

	entry := activePane.SelectedEntry()
	if entry != nil && !entry.IsParentDir() {
		srcPath := filepath.Join(activePane.Path(), entry.Name)
		destPath := m.getInactivePane().Path()
		return m, m.checkFileConflict(srcPath, destPath, "copy")
	}
	return m, nil
}

// handleMove は移動を処理
func (m Model) handleMove() (tea.Model, tea.Cmd) {
	activePane := m.getActivePane()
	markedFiles := activePane.GetMarkedFiles()

	if len(markedFiles) > 0 {
		return m, m.startBatchOperation(markedFiles, "move")
	}

	entry := activePane.SelectedEntry()
	if entry != nil && !entry.IsParentDir() {
		srcPath := filepath.Join(activePane.Path(), entry.Name)
		destPath := m.getInactivePane().Path()
		return m, m.checkFileConflict(srcPath, destPath, "move")
	}
	return m, nil
}

// handleDelete は削除を処理
func (m Model) handleDelete() (tea.Model, tea.Cmd) {
	activePane := m.getActivePane()
	markedFiles := activePane.GetMarkedFiles()

	if len(markedFiles) > 0 {
		m.dialog = NewConfirmDialog(
			fmt.Sprintf("Delete %d files?", len(markedFiles)),
			"This action cannot be undone.",
		)
	} else {
		entry := activePane.SelectedEntry()
		if entry != nil && !entry.IsParentDir() {
			m.dialog = NewConfirmDialog(
				"Delete file?",
				entry.DisplayName(),
			)
		}
	}
	return m, nil
}

// handleContextMenu はコンテキストメニューを処理
func (m Model) handleContextMenu() (tea.Model, tea.Cmd) {
	activePane := m.getActivePane()
	entry := activePane.SelectedEntry()

	if entry != nil && !entry.IsParentDir() {
		m.dialog = NewContextMenuDialogWithPane(
			entry,
			activePane.Path(),
			m.getInactivePane().Path(),
			activePane,
		)
	}
	return m, nil
}

// handleView はビューアー表示を処理
func (m Model) handleView() (tea.Model, tea.Cmd) {
	entry := m.getActivePane().SelectedEntry()
	if entry != nil && !entry.IsParentDir() && !entry.IsDir {
		fullPath := filepath.Join(m.getActivePane().Path(), entry.Name)
		if err := checkReadPermission(fullPath); err != nil {
			m.statusMessage = fmt.Sprintf("Cannot read file: %v", err)
			m.isStatusError = true
			return m, statusMessageClearCmd(5 * time.Second)
		}
		return m, openWithViewer(fullPath, m.getActivePane().Path())
	}
	return m, nil
}

// handleEdit はエディター表示を処理
func (m Model) handleEdit() (tea.Model, tea.Cmd) {
	entry := m.getActivePane().SelectedEntry()
	if entry != nil && !entry.IsParentDir() && !entry.IsDir {
		fullPath := filepath.Join(m.getActivePane().Path(), entry.Name)
		if err := checkReadPermission(fullPath); err != nil {
			m.statusMessage = fmt.Sprintf("Cannot read file: %v", err)
			m.isStatusError = true
			return m, statusMessageClearCmd(5 * time.Second)
		}
		return m, openWithEditor(fullPath, m.getActivePane().Path())
	}
	return m, nil
}

// handleNewFile は新規ファイル作成を処理
func (m Model) handleNewFile() (tea.Model, tea.Cmd) {
	pane := m.getActivePane()
	m.dialog = NewInputDialog("New file:", func(filename string) tea.Cmd {
		return m.handleCreateFile(pane.Path(), filename)
	})
	return m, nil
}

// handleNewDirectory は新規ディレクトリ作成を処理
func (m Model) handleNewDirectory() (tea.Model, tea.Cmd) {
	pane := m.getActivePane()
	m.dialog = NewInputDialog("New directory:", func(dirname string) tea.Cmd {
		return m.handleCreateDirectory(pane.Path(), dirname)
	})
	return m, nil
}

// handleRenameUI はリネームダイアログを表示
func (m Model) handleRenameUI() (tea.Model, tea.Cmd) {
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		return m, nil
	}
	pane := m.getActivePane()
	oldName := entry.Name
	m.dialog = NewInputDialog("Rename to:", func(newName string) tea.Cmd {
		return m.handleRename(pane.Path(), oldName, newName)
	})
	return m, nil
}

// handleAddBookmarkUI はブックマーク追加ダイアログを表示
func (m Model) handleAddBookmarkUI() (tea.Model, tea.Cmd) {
	currentPath := m.getActivePane().Path()

	if isPathBookmarked(m.bookmarks, currentPath) {
		m.statusMessage = "Already bookmarked"
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second)
	}

	defaultAlias := defaultAliasFromPath(currentPath)
	currentBookmarks := m.bookmarks
	dialog := NewInputDialog("Bookmark name:", func(alias string) tea.Cmd {
		return m.handleAddBookmark(currentBookmarks, currentPath, alias)
	})
	dialog.SetEmptyErrorMsg("Bookmark name cannot be empty")
	dialog.input = defaultAlias
	dialog.cursorPos = len(defaultAlias)
	m.dialog = dialog
	return m, nil
}
