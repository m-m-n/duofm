package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/sakura/duofm/internal/config"
)

// handleKeyInput はキーボード入力を処理する
func (m Model) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// バックグラウンド出力フォーカス中
	if m.bgOutputFocused {
		return m.handleBgOutputFocusedInput(msg)
	}

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

	// TABキーでバックグラウンド出力エリアにフォーカス
	if msg.Type == tea.KeyTab {
		if m.bgRunner != nil && m.bgRunner.IsRunning() && m.bgRunner.Pane() == m.activePane {
			m.bgOutputFocused = true
			return m, nil
		}
	}

	// keybindingMapを使ってアクションを決定
	action := m.keybindingMap.GetAction(msg.String())
	return m.handleAction(action)
}

// handleBgOutputFocusedInput handles keyboard input when the background output area is focused.
// Only Ctrl+C (cancel), TAB, and Esc (return to file list) are accepted.
func (m Model) handleBgOutputFocusedInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		// Cancel background command
		if m.bgRunner != nil {
			m.bgRunner.Cancel()
		}
		// Log footer
		if m.shellLogger != nil {
			m.shellLogger.AppendFooter()
		}
		// Clean up state and refresh panes
		m.bgCleanupAndRefresh()
		return m, nil

	case tea.KeyTab, tea.KeyEsc:
		// Return focus to file list
		m.bgOutputFocused = false
		return m, nil
	}

	// All other keys are ignored
	return m, nil
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
	// Handle Ctrl+R for history search
	if msg.Type == tea.KeyCtrlR {
		m.tabCompletionPending = false
		return m.handleHistorySearch()
	}

	// In history search mode, handle input differently
	if m.historySearching {
		return m.handleHistorySearchInput(msg)
	}

	// Handle TAB for completion
	if msg.Type == tea.KeyTab {
		return m.handleShellCommandTab()
	}

	// Non-TAB key: reset tab completion pending state
	if m.tabCompletionPending {
		m.tabCompletionPending = false
		m.statusMessage = ""
	}

	// Handle Up/Down arrow keys for history navigation
	if msg.Type == tea.KeyUp {
		return m.handleHistoryUp()
	}
	if msg.Type == tea.KeyDown {
		return m.handleHistoryDown()
	}

	// Handle ! key for background mode toggle
	if msg.Type == tea.KeyRunes && string(msg.Runes) == "!" {
		if !m.bgMode {
			// Enter background mode
			m.bgMode = true
			m.minibuffer.SetPrompt(m.bgModePrompt())
			return m, nil
		}
		// Already in bgMode: append ! character normally (fall through to default)
	}

	// Handle Backspace in background mode
	if m.bgMode && msg.Type == tea.KeyBackspace {
		if m.minibuffer.Input() == "" {
			// Empty input + backspace = exit background mode
			m.bgMode = false
			m.minibuffer.SetPrompt("!: ")
			return m, nil
		}
		// Non-empty: normal backspace handling (fall through to default)
	}

	switch msg.Type {
	case tea.KeyEnter:
		command := m.minibuffer.Input()
		if command == "" {
			m.shellCommandMode = false
			m.bgMode = false
			m.minibuffer.Hide()
			return m, nil
		}

		if m.bgMode {
			// Background mode execution
			return m.handleBgEnter(command)
		}

		// Foreground execution
		workDir := m.getActivePane().Path()
		m.shellCommandMode = false
		m.minibuffer.Hide()

		// Add to history before executing
		if m.shellHistory != nil && m.shellHistory.IsEnabled() {
			m.shellHistory.Add(command)
		}

		// Log command header before execution and get log path for script
		logFile := m.shellLogger.LogPath()
		if err := m.shellLogger.AppendHeader(command, workDir); err != nil {
			m.statusMessage = fmt.Sprintf("Shell log error: %v", err)
			m.isStatusError = true
		}

		return m, executeShellCommand(command, workDir, logFile)

	case tea.KeyEsc, tea.KeyCtrlC:
		m.shellCommandMode = false
		m.bgMode = false
		m.minibuffer.Hide()
		return m, nil

	default:
		m.minibuffer.HandleKey(msg)
		return m, nil
	}
}

// bgModePrompt returns the pink-colored prompt for background mode
func (m Model) bgModePrompt() string {
	return lipgloss.NewStyle().Foreground(highlightColor).Render("!") + ": "
}

// handleBgEnter handles Enter in background mode
func (m Model) handleBgEnter(command string) (tea.Model, tea.Cmd) {
	workDir := m.getActivePane().Path()
	launchPane := m.activePane

	m.shellCommandMode = false
	m.bgMode = false
	m.minibuffer.Hide()

	// Add to history
	if m.shellHistory != nil && m.shellHistory.IsEnabled() {
		m.shellHistory.Add(command)
	}

	// Log command header
	if err := m.shellLogger.AppendHeader(command, workDir); err != nil {
		m.statusMessage = fmt.Sprintf("Shell log error: %v", err)
		m.isStatusError = true
	}

	// Clear output buffer for new command
	m.bgOutputBuffer.Clear()

	// Increment session ID to invalidate any stale timers
	m.bgSessionID++

	// Create channels and store on model
	m.bgOutputCh = make(chan string, 100)
	m.bgDoneCh = make(chan error, 1)
	m.bgCommand = command
	m.bgWorkDir = workDir

	err := m.bgRunner.Start(command, workDir, launchPane,
		func(line string) {
			m.bgOutputCh <- line
		},
		func(err error) {
			m.bgDoneCh <- err
		},
	)
	if err != nil {
		m.statusMessage = fmt.Sprintf("Background command failed: %v", err)
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second)
	}

	// Return a tea.Cmd that waits for the first output or completion
	return m, m.waitForBgEvent()
}

// waitForBgEvent returns a tea.Cmd that waits for the next background event
func (m Model) waitForBgEvent() tea.Cmd {
	outputCh := m.bgOutputCh
	doneCh := m.bgDoneCh
	command := m.bgCommand
	workDir := m.bgWorkDir
	sessionID := m.bgSessionID
	if outputCh == nil || doneCh == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case line, ok := <-outputCh:
			if !ok {
				return nil
			}
			return bgOutputMsg{line: line}
		case err := <-doneCh:
			return bgCommandDoneMsg{err: err, command: command, workDir: workDir, sessionID: sessionID}
		}
	}
}

// handleHistoryUp handles Up arrow key for history navigation in shell command mode
func (m Model) handleHistoryUp() (tea.Model, tea.Cmd) {
	// Check if history is enabled and non-empty
	if m.shellHistory == nil || !m.shellHistory.IsEnabled() {
		return m, nil
	}

	commands := m.shellHistory.Commands()
	if len(commands) == 0 {
		return m, nil
	}

	// If this is the first navigation (historyIndex == -1), save current input
	if m.historyIndex == -1 {
		m.historyEditBuf = m.minibuffer.Input()
	}

	// Navigate to older command if not at the oldest
	if m.historyIndex < len(commands)-1 {
		m.historyIndex++
		m.minibuffer.SetInput(commands[m.historyIndex])
	}

	return m, nil
}

// handleHistoryDown handles Down arrow key for history navigation in shell command mode
func (m Model) handleHistoryDown() (tea.Model, tea.Cmd) {
	// If not navigating history, do nothing
	if m.historyIndex < 0 {
		return m, nil
	}

	// Navigate to newer command
	m.historyIndex--

	if m.historyIndex == -1 {
		// Restore original input
		m.minibuffer.SetInput(m.historyEditBuf)
	} else {
		// Show command at new index (defensive check for nil shellHistory)
		if m.shellHistory == nil {
			return m, nil
		}
		commands := m.shellHistory.Commands()
		if m.historyIndex < len(commands) {
			m.minibuffer.SetInput(commands[m.historyIndex])
		}
	}

	return m, nil
}

// handleShellCommandTab handles TAB key for command/path completion
func (m Model) handleShellCommandTab() (tea.Model, tea.Cmd) {
	input := m.minibuffer.Input()
	cursorPos := m.minibuffer.CursorPos()
	cwd := m.getActivePane().Path()

	result := m.tabCompleter.Complete(input, cursorPos, cwd)

	if result.HasProgress {
		// Completion made progress - update input
		m.minibuffer.SetInput(result.NewInput)
		m.minibuffer.SetCursorPos(result.NewCursorPos)
		m.tabCompletionPending = false
		return m, nil
	}

	// No progress
	if len(result.Candidates) == 0 {
		return m, nil
	}

	// Extract the word being completed for the message
	word, _ := extractWordAtCursor(input, cursorPos)

	if !m.tabCompletionPending {
		// TAB 1st press: show candidate count
		m.tabCompletionPending = true
		m.statusMessage = fmt.Sprintf("%d matches for '%s'", len(result.Candidates), word)
		m.isStatusError = false
		return m, nil
	}

	// TAB 2nd press: show candidate list
	m.tabCompletionPending = false
	m.statusMessage = formatCandidateList(result.Candidates, m.width)
	m.isStatusError = false
	return m, statusMessageClearCmd(5 * time.Second)
}

// formatCandidateList formats candidates to fit within the given width.
func formatCandidateList(candidates []string, maxWidth int) string {
	if len(candidates) == 0 {
		return ""
	}

	if maxWidth <= 0 {
		maxWidth = 80
	}

	var parts []string
	remaining := maxWidth
	shown := 0

	for i, c := range candidates {
		cw := runewidth.StringWidth(c)
		needed := cw
		if shown > 0 {
			needed++ // space separator
		}

		isLast := i == len(candidates)-1

		// If this is not the last candidate, reserve space for "(+N more)" suffix
		if shown > 0 && !isLast {
			moreText := fmt.Sprintf(" (+%d more)", len(candidates)-shown)
			if remaining-needed < len(moreText) {
				// Can't fit this candidate plus the "(+N more)" text
				parts = append(parts, fmt.Sprintf("(+%d more)", len(candidates)-shown))
				break
			}
		}

		if shown > 0 {
			remaining-- // space
		}
		if remaining < cw && shown > 0 {
			parts = append(parts, fmt.Sprintf("(+%d more)", len(candidates)-shown))
			break
		}

		parts = append(parts, c)
		remaining -= cw
		shown++
	}

	return strings.Join(parts, " ")
}

// handleHistorySearch handles Ctrl+R press in shell command mode
func (m Model) handleHistorySearch() (tea.Model, tea.Cmd) {
	// If history is disabled or not initialized, do nothing
	if m.shellHistory == nil || !m.shellHistory.IsEnabled() {
		return m, nil
	}

	if !m.historySearching {
		// Start history search mode
		m.historySearching = true
		m.historySearchPattern = ""
		m.historySearcher = NewHistorySearcher(m.shellHistory)
		m.minibuffer.SetPrompt("(reverse-i-search)'': ")
		m.minibuffer.Clear()
	} else {
		// Already in search mode - move to next match
		if m.historySearcher != nil {
			next := m.historySearcher.Next()
			if next != "" {
				m.minibuffer.SetInput(next)
			}
		}
	}

	return m, nil
}

// handleHistorySearchInput handles input during history search mode
func (m Model) handleHistorySearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Execute the selected command
		command := m.minibuffer.Input()
		m.historySearching = false
		m.historySearchPattern = ""
		m.historySearcher = nil
		m.shellCommandMode = false
		m.minibuffer.Hide()

		if command == "" {
			return m, nil
		}

		workDir := m.getActivePane().Path()

		// Add to history before executing
		if m.shellHistory != nil && m.shellHistory.IsEnabled() {
			m.shellHistory.Add(command)
		}

		// Log command header before execution and get log path for script
		logFile := m.shellLogger.LogPath()
		if err := m.shellLogger.AppendHeader(command, workDir); err != nil {
			m.statusMessage = fmt.Sprintf("Shell log error: %v", err)
			m.isStatusError = true
		}

		return m, executeShellCommand(command, workDir, logFile)

	case tea.KeyEsc, tea.KeyCtrlC:
		// Cancel history search, return to shell command mode
		m.historySearching = false
		m.historySearchPattern = ""
		m.historySearcher = nil
		m.minibuffer.SetPrompt("!: ")
		m.minibuffer.Clear()
		return m, nil

	case tea.KeyBackspace:
		// Handle backspace - update search pattern
		if len(m.historySearchPattern) > 0 {
			runes := []rune(m.historySearchPattern)
			m.historySearchPattern = string(runes[:len(runes)-1])
		}
		m.updateHistorySearch()
		return m, nil

	case tea.KeyRunes:
		// Regular input - update search pattern
		m.historySearchPattern += string(msg.Runes)
		m.updateHistorySearch()
		return m, nil

	case tea.KeySpace:
		// Space input - update search pattern
		m.historySearchPattern += " "
		m.updateHistorySearch()
		return m, nil

	default:
		return m, nil
	}
}

// updateHistorySearch updates the history search based on current pattern
func (m *Model) updateHistorySearch() {
	if m.historySearcher == nil {
		return
	}

	m.historySearcher.SetPattern(m.historySearchPattern)

	// Update prompt to show current pattern (bash-style format)
	m.minibuffer.SetPrompt("(reverse-i-search)'" + m.historySearchPattern + "': ")

	// Get current match and display it
	current := m.historySearcher.Current()
	if current != "" {
		m.minibuffer.SetInput(current)
	} else {
		// No match - clear input but keep showing pattern being typed
		m.minibuffer.SetInput("")
	}
}

// handleShellLog はシェルログ表示を処理
func (m Model) handleShellLog() (tea.Model, tea.Cmd) {
	if m.shellLogger == nil || !m.shellLogger.HasLog() {
		m.statusMessage = "No shell log"
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second)
	}
	return m, openWithViewer(m.shellLogger.LogPath(), m.getActivePane().Path())
}

// handleCtrlC はCtrl+Cのダブルプレスを処理
func (m Model) handleCtrlC() (tea.Model, tea.Cmd) {
	if m.ctrlCPending {
		// Cancel background process to avoid orphans
		if m.bgRunner != nil && m.bgRunner.IsRunning() {
			m.bgRunner.Cancel()
		}
		// Close shell history and shell logger before quitting
		if m.shellHistory != nil {
			m.shellHistory.Close()
		}
		if m.shellLogger != nil {
			m.shellLogger.Close()
		}
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
		// Cancel background process to avoid orphans
		if m.bgRunner != nil && m.bgRunner.IsRunning() {
			m.bgRunner.Cancel()
		}
		// Close shell history and shell logger before quitting
		if m.shellHistory != nil {
			m.shellHistory.Close()
		}
		if m.shellLogger != nil {
			m.shellLogger.Close()
		}
		return m, tea.Quit

	case ActionHelp:
		m.dialog = NewHelpDialog()
		return m, nil

	case ActionSearch:
		m.startSearch(SearchModeIncremental)
		return m, nil

	case ActionRegexSearch:
		m.dialog = NewRegexSearchDialog(m.regexHistory)
		return m, nil

	case ActionSQLFilter:
		m.dialog = NewQuerySearchDialog(m.queryHistory)
		return m, nil

	case ActionShellCommand:
		// Block new shell commands during background execution
		if m.bgRunner != nil && m.bgRunner.IsRunning() {
			m.statusMessage = "Background command running"
			m.isStatusError = false
			return m, statusMessageClearCmd(3 * time.Second)
		}
		m.startShellCommandMode()
		return m, nil

	case ActionMoveDown:
		m.getActivePane().MoveCursorDown()
		return m, nil

	case ActionMoveUp:
		m.getActivePane().MoveCursorUp()
		return m, nil

	case ActionPageDown:
		m.getActivePane().MoveCursorPageDown()
		return m, nil

	case ActionPageUp:
		m.getActivePane().MoveCursorPageUp()
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

	case ActionRenameFullName:
		return m.handleRenameFullNameUI()

	case ActionSort:
		m.sortDialog = NewSortDialog(m.getActivePane().GetSortConfig())
		return m, nil

	case ActionBookmark:
		m.dialog = NewBookmarkDialog(m.bookmarkManager.Bookmarks())
		return m, nil

	case ActionAddBookmark:
		return m.handleAddBookmarkUI()

	case ActionPermission:
		return m.handlePermission()

	case ActionPathJump:
		m.dialog = NewPathJumpDialog()
		return m, nil

	case ActionTrash:
		return m.handleTrash()

	case ActionOpenTrash:
		return m.handleOpenTrashDialog()

	case ActionRestore:
		// Restore is now handled within TrashDialog only
		// R key outside dialog now always does rename
		return m.handleRenameUI()

	case ActionEmptyTrash:
		// Empty trash is now handled within TrashDialog only
		// This action does nothing outside the dialog
		return m, nil

	case ActionShellLog:
		return m.handleShellLog()

	case ActionGotoTop:
		m.getActivePane().GotoTop()
		return m, nil

	case ActionGotoBottom:
		m.getActivePane().GotoBottom()
		return m, nil
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
// EnterBehavior設定に基づいてファイルを開く動作を決定:
//   - less: ページャーで開く（フォアグラウンド、デフォルト）
//   - xdg-open: システムのデフォルトアプリで開く（バックグラウンド）
//   - path:XXX: 指定されたアプリケーションで開く（フォアグラウンド）
//   - mime: MIME type に基づいてアプリケーションを選択（フォアグラウンド）
func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	entry := m.getActivePane().SelectedEntry()
	if entry != nil && !entry.IsParentDir() && !entry.IsDir {
		fullPath := filepath.Join(m.getActivePane().Path(), entry.Name)
		workDir := m.getActivePane().Path()

		// xdg-open以外はパーミッションチェックが必要
		if m.enterBehavior.Type != config.EnterBehaviorXDGOpen {
			if err := checkReadPermission(fullPath); err != nil {
				m.statusMessage = fmt.Sprintf("Cannot read file: %v", err)
				m.isStatusError = true
				return m, statusMessageClearCmd(5 * time.Second)
			}
		}

		// EnterBehaviorに基づいてファイルを開く
		switch m.enterBehavior.Type {
		case config.EnterBehaviorXDGOpen:
			return m, openWithXDG(fullPath, workDir)
		case config.EnterBehaviorCustom:
			return m, openWithCustomForeground(m.enterBehavior.CustomPath, fullPath, workDir)
		case config.EnterBehaviorMIME:
			cmd, statusMsg := openWithMIME(fullPath, workDir, m.mimeBehavior)
			if statusMsg != "" {
				m.statusMessage = statusMsg
				m.isStatusError = true
				return m, tea.Batch(cmd, statusMessageClearCmd(5*time.Second))
			}
			return m, cmd
		default: // EnterBehaviorLess
			return m, openWithViewer(fullPath, workDir)
		}
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
// ファイルの種類に応じて拡張子保持モードまたはフルネーム編集モードを選択する
func (m Model) handleRenameUI() (tea.Model, tea.Cmd) {
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		return m, nil
	}
	pane := m.getActivePane()
	oldName := entry.Name

	// 拡張子保持モードが使えるか判定
	baseName, ext, hasExt := hasEditableExtension(entry.Name, entry.IsDir)

	if hasExt {
		// 拡張子保持モード（ExtensionRenameDialog）
		m.dialog = NewExtensionRenameDialog(pane.Path(), oldName, baseName, ext)
	} else {
		// フルネーム編集モード（InputDialog）
		dialog := NewInputDialog("Rename to:", func(newName string) tea.Cmd {
			return m.handleRename(pane.Path(), oldName, newName)
		})
		dialog.SetInput(oldName)
		m.dialog = dialog
	}
	return m, nil
}

// handleRenameFullNameUI はフルネームリネームダイアログを表示
// ファイル種類に関わらず常にフルネーム編集モードを使用する（Shift+R用）
func (m Model) handleRenameFullNameUI() (tea.Model, tea.Cmd) {
	entry := m.getActivePane().SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		return m, nil
	}
	pane := m.getActivePane()
	oldName := entry.Name

	dialog := NewInputDialog("Rename to:", func(newName string) tea.Cmd {
		return m.handleRename(pane.Path(), oldName, newName)
	})
	dialog.SetInput(oldName)
	m.dialog = dialog
	return m, nil
}

// handleAddBookmarkUI はブックマーク追加ダイアログを表示
func (m Model) handleAddBookmarkUI() (tea.Model, tea.Cmd) {
	currentPath := m.getActivePane().Path()

	if isPathBookmarked(m.bookmarkManager.Bookmarks(), currentPath) {
		m.statusMessage = "Already bookmarked"
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second)
	}

	defaultAlias := defaultAliasFromPath(currentPath)
	dialog := NewInputDialog("Bookmark name:", func(alias string) tea.Cmd {
		return m.bookmarkManager.Add(currentPath, alias)
	})
	dialog.SetEmptyErrorMsg("Bookmark name cannot be empty")
	dialog.SetInput(defaultAlias)
	m.dialog = dialog
	return m, nil
}
