package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/archive"
	"github.com/sakura/duofm/internal/config"
	"github.com/sakura/duofm/internal/fs"
)

// ANSIエスケープシーケンスを除去するための正規表現
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Model はアプリケーション全体の状態を保持
type Model struct {
	// Pane management
	leftPane   *Pane
	rightPane  *Pane
	leftPath   string
	rightPath  string
	activePane PanePosition

	// UI state
	dialog Dialog
	width  int
	height int
	ready  bool

	// Disk space monitoring (delegated)
	diskSpaceMonitor *DiskSpaceMonitor

	// Status bar
	statusMessage  string   // ステータスバーに表示するメッセージ
	isStatusError  bool     // エラーメッセージかどうか
	configWarnings []string // 設定ファイルの警告

	// Search and minibuffer
	searchState  SearchState    // 検索状態
	minibuffer   *Minibuffer    // ミニバッファ
	regexHistory *SearchHistory // 正規表現検索履歴
	queryHistory *SearchHistory // クエリ検索履歴

	// Input state
	ctrlCPending         bool // Ctrl+Cが1回押された状態かどうか
	tabCompletionPending bool // TAB1回目が押された状態（次のTABで候補一覧表示）
	shellCommandMode     bool // シェルコマンドモードかどうか
	historySearching     bool // 履歴検索モードかどうか
	historySearchPattern string
	shellHistory         *ShellHistory
	historySearcher      *HistorySearcher
	historyIndex         int    // History navigation position: -1=at input, 0+=history positions
	historyEditBuf       string // Preserve original input before navigation

	// Dialogs
	sortDialog *SortDialog // ソートダイアログ（nil = 非表示）

	// Operations
	pendingAction  func() error           // 確認待ちのアクション（コンテキストメニューの削除用）
	batchOpManager *BatchOperationManager // Batch copy/move manager (delegated)

	// Configuration
	keybindingMap *KeybindingMap            // キーバインドマップ
	theme         *Theme                    // カラーテーマ
	refreshRate   int                       // Auto-refresh interval in seconds (0 = disabled)
	enterBehavior config.EnterBehavior      // Enterキーの動作設定
	mimeBehavior  config.MIMEBehaviorConfig // MIME type to command mapping

	// Bookmarks (delegated)
	bookmarkManager *BookmarkManager

	// Archive operations (delegated)
	archiveOpManager *ArchiveOperationManager

	// Configuration hot-reload
	configPath          string                   // Path to the config file
	configWatcher       *config.ConfigWatcher    // File watcher reference (for SuppressFor)
	pendingReloadResult *config.ConfigLoadResult // Reload result pending dialog decision
	pendingConfigError  *config.ConfigLoadResult // Queued error when another dialog is open

	// Per-directory sort settings
	dirSortStore *config.DirSortStore

	// Shell log
	shellLogger *ShellLogger

	// Tab completion
	tabCompleter *TabCompleter

	// Background shell command
	bgMode          bool              // true when in background input mode (pink prompt)
	bgRunner        *BackgroundRunner // manages the background process
	bgOutputBuffer  *OutputBuffer     // circular buffer for output lines
	bgOutputFocused bool              // true when output area has keyboard focus
	bgClosing       bool              // true during 2-sec auto-close delay
	bgOutputCh      chan string       // channel for output lines from background goroutine
	bgDoneCh        chan error        // channel for completion signal from background goroutine
	bgCommand       string            // the background command string (for done msg)
	bgWorkDir       string            // the working directory for bg command (for done msg)
}

// PanePosition はペインの位置を表す
type PanePosition int

const (
	// LeftPane は左ペイン
	LeftPane PanePosition = iota
	// RightPane は右ペイン
	RightPane
)

// ModelOptions holds all configuration for Model initialization.
type ModelOptions struct {
	KeybindingMap *KeybindingMap
	Theme         *Theme
	Warnings      []string
	HistoryLimit  int
	RefreshRate   int
	EnterBehavior config.EnterBehavior
	MIMEBehavior  config.MIMEBehaviorConfig
	ConfigPath    string
	DirSortStore  *config.DirSortStore
	ShellLogDir   string
}

// NewModel は初期モデルを作成（デフォルトキーバインドを使用）
func NewModel() Model {
	return NewModelWithConfig(ModelOptions{
		RefreshRate: config.DefaultRefreshRate,
	})
}

// NewModelWithConfig は設定付きの初期モデルを作成
func NewModelWithConfig(opts ModelOptions) Model {
	keybindingMap := opts.KeybindingMap
	theme := opts.Theme
	warnings := opts.Warnings
	historyLimit := opts.HistoryLimit
	refreshRate := opts.RefreshRate
	enterBehavior := opts.EnterBehavior
	mimeBehavior := opts.MIMEBehavior
	configPath := opts.ConfigPath
	// 初期ディレクトリの取得
	cwd, err := fs.CurrentDirectory()
	if err != nil {
		cwd = "/"
	}

	home, err := fs.HomeDirectory()
	if err != nil {
		home = "/"
	}

	// keybindingMapがnilの場合はデフォルトを使用
	if keybindingMap == nil {
		keybindingMap = DefaultKeybindingMap()
	}

	// themeがnilの場合はデフォルトを使用
	if theme == nil {
		theme = DefaultTheme()
	}

	// ブックマークを読み込み
	bookmarkManager, bookmarkWarnings := NewBookmarkManager()
	warnings = append(warnings, bookmarkWarnings...)

	// シェルコマンド履歴を初期化
	var shellHistory *ShellHistory
	if historyLimit > 0 {
		historyPath, err := getHistoryPath()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Warning: could not get history path: %v", err))
		} else {
			shellHistory = NewShellHistory(historyPath, historyLimit)
			if err := shellHistory.Load(); err != nil {
				warnings = append(warnings, fmt.Sprintf("Warning: could not load history: %v", err))
			}
		}
	}

	return Model{
		leftPane:         nil, // Updateで初期化
		rightPane:        nil, // Updateで初期化
		leftPath:         cwd,
		rightPath:        home,
		activePane:       LeftPane,
		dialog:           nil,
		ready:            false,
		diskSpaceMonitor: NewDiskSpaceMonitor(),
		searchState:      SearchState{Mode: SearchModeNone},
		minibuffer:       NewMinibuffer(),
		regexHistory:     NewSearchHistory(DefaultSearchHistorySize),
		queryHistory:     NewSearchHistory(DefaultSearchHistorySize),
		keybindingMap:    keybindingMap,
		configWarnings:   warnings,
		theme:            theme,
		refreshRate:      refreshRate,
		enterBehavior:    enterBehavior,
		mimeBehavior:     mimeBehavior,
		bookmarkManager:  bookmarkManager,
		batchOpManager:   NewBatchOperationManager(),
		archiveOpManager: NewArchiveOperationManager(),
		shellHistory:     shellHistory,
		configPath:       configPath,
		dirSortStore:     opts.DirSortStore,
		shellLogger:      NewShellLogger(opts.ShellLogDir),
		tabCompleter:     NewTabCompleter(),
		bgRunner:         NewBackgroundRunner(),
		bgOutputBuffer:   NewOutputBuffer(10000),
	}
}

// getHistoryPath returns the shell command history file path.
func getHistoryPath() (string, error) {
	// Use XDG_CONFIG_HOME or fall back to ~/.config
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "duofm", "history"), nil
}

// SetPendingReloadResult sets the pending reload result for startup error handling.
func (m *Model) SetPendingReloadResult(result *config.ConfigLoadResult) {
	m.pendingReloadResult = result
}

// SetConfigWatcher sets the config file watcher reference.
func (m *Model) SetConfigWatcher(watcher *config.ConfigWatcher) {
	m.configWatcher = watcher
}

// applyConfig applies a new configuration to the model.
func (m *Model) applyConfig(cfg *config.Config) {
	m.keybindingMap = NewKeybindingMap(cfg)
	m.theme = NewTheme(cfg.Colors)
	m.enterBehavior = cfg.EnterBehavior
	m.mimeBehavior = cfg.MIMEBehavior

	// Update pane themes
	if m.leftPane != nil {
		m.leftPane.SetTheme(m.theme)
	}
	if m.rightPane != nil {
		m.rightPane.SetTheme(m.theme)
	}

	// Update history limit
	if m.shellHistory != nil {
		m.shellHistory.SetLimit(cfg.HistoryLimit)
	}

	// Update refresh rate
	m.refreshRate = cfg.RefreshRate
}

// Init はBubble Teaの初期化
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	// 起動時設定エラーがあればダイアログ表示
	if m.pendingReloadResult != nil && m.pendingReloadResult.HasErrors() {
		cmds = append(cmds, func() tea.Msg {
			return configStartupErrorMsg{result: m.pendingReloadResult}
		})
	}

	// 設定ファイルの警告があれば最初の警告をステータスバーに表示
	if len(m.configWarnings) > 0 {
		m.statusMessage = m.configWarnings[0]
		m.isStatusError = false
	}

	// Note: Auto-refresh ticker is started in handleWindowSize (first WindowSizeMsg),
	// not here, because panes must be initialized first.

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

// Update is now defined in model_update.go
// View is now defined in model_view.go

// getActivePane は現在アクティブなペインを返す
func (m *Model) getActivePane() *Pane {
	if m.activePane == LeftPane {
		return m.leftPane
	}
	return m.rightPane
}

// getInactivePane は非アクティブなペインを返す
func (m *Model) getInactivePane() *Pane {
	if m.activePane == LeftPane {
		return m.rightPane
	}
	return m.leftPane
}

// switchToPane はアクティブペインを切り替え
func (m *Model) switchToPane(pos PanePosition) {
	// 検索中の場合はキャンセル
	if m.searchState.IsActive {
		m.cancelSearch()
	}

	m.activePane = pos
	m.leftPane.SetActive(pos == LeftPane)
	m.rightPane.SetActive(pos == RightPane)
}

// updateDiskSpace はディスク容量を更新
func (m *Model) updateDiskSpace() {
	leftPath := ""
	rightPath := ""
	if m.leftPane != nil {
		leftPath = m.leftPane.Path()
	}
	if m.rightPane != nil {
		rightPath = m.rightPane.Path()
	}
	m.diskSpaceMonitor.Update(leftPath, rightPath)
}

// bgCleanup resets all background state and refreshes both panes.
func (m *Model) bgCleanup() {
	m.bgClosing = false
	m.bgOutputFocused = false
	m.bgOutputBuffer.Clear()
	m.bgOutputCh = nil
	m.bgDoneCh = nil
	m.bgCommand = ""
	m.bgWorkDir = ""
}

// isBgActive returns true if a background command is running or closing.
func (m *Model) isBgActive() bool {
	return (m.bgRunner != nil && m.bgRunner.IsRunning()) || m.bgClosing
}

// startShellCommandMode はシェルコマンドモードを開始する
func (m *Model) startShellCommandMode() {
	m.shellCommandMode = true
	m.historyIndex = -1   // Reset navigation position
	m.historyEditBuf = "" // Clear edit buffer
	m.minibuffer.SetPrompt("!: ")
	m.minibuffer.Clear()
	m.minibuffer.SetWidth(m.getActivePane().width)
	m.minibuffer.Show()
}

// startSearch は検索モードを開始する
// Note: Only handles incremental search now. Regex and query searches use dialogs.
func (m *Model) startSearch(mode SearchMode) {
	// Only handle incremental search - other modes use dialogs
	if mode != SearchModeIncremental {
		return
	}

	// 現在のフィルタ状態を保存（Esc時に復元するため）
	pane := m.getActivePane()
	if pane.IsFiltered() {
		m.searchState.PreviousResult = &SearchResult{
			Mode:    pane.FilterMode(),
			Pattern: pane.FilterPattern(),
		}
	} else {
		m.searchState.PreviousResult = nil
	}

	m.searchState.Mode = mode
	m.searchState.Pattern = ""
	m.searchState.IsActive = true

	m.minibuffer.SetPrompt("/: ")
	m.minibuffer.Clear()
	m.minibuffer.SetWidth(m.getActivePane().width)
	m.minibuffer.Show()
}

// confirmSearch は検索を確定する
func (m *Model) confirmSearch() {
	pattern := m.minibuffer.Input()
	pane := m.getActivePane()

	if pattern == "" {
		// 空のパターンでEnter→フィルタをクリア
		pane.ClearFilter()
		m.searchState.PreviousResult = nil
	} else {
		// パターンがある場合はフィルタを適用
		if err := pane.ApplyFilter(pattern, m.searchState.Mode); err != nil {
			// 正規表現エラーの場合はステータスバーに表示してミニバッファを維持
			m.statusMessage = fmt.Sprintf("Invalid regex: %v", err)
			m.isStatusError = true
			return
		}
		// 成功した場合、現在のフィルタを「前の結果」として保存
		m.searchState.PreviousResult = &SearchResult{
			Mode:    m.searchState.Mode,
			Pattern: pattern,
		}
	}

	// ミニバッファを閉じる
	m.minibuffer.Hide()
	m.searchState.IsActive = false
	m.searchState.Mode = SearchModeNone
}

// cancelSearch は検索をキャンセルする
func (m *Model) cancelSearch() {
	pane := m.getActivePane()

	// 前の検索結果があれば復元
	if m.searchState.PreviousResult != nil {
		pane.ApplyFilter(m.searchState.PreviousResult.Pattern, m.searchState.PreviousResult.Mode)
	} else {
		// 前の結果がなければフィルタをクリア
		pane.ClearFilter()
	}

	// ミニバッファを閉じる
	m.minibuffer.Hide()
	m.searchState.IsActive = false
	m.searchState.Mode = SearchModeNone
}

// applyIncrementalFilter はインクリメンタル検索のフィルタを適用する
func (m *Model) applyIncrementalFilter() {
	pattern := m.minibuffer.Input()
	pane := m.getActivePane()

	// インクリメンタル検索の場合は即座にフィルタを適用
	if m.searchState.Mode == SearchModeIncremental {
		pane.ApplyFilter(pattern, SearchModeIncremental)
	}
}

// RefreshBothPanes refreshes both panes
func (m *Model) RefreshBothPanes() tea.Cmd {
	var cmds []tea.Cmd

	// Refresh left pane
	if err := m.leftPane.Refresh(); err != nil {
		m.dialog = NewErrorDialog(fmt.Sprintf("Failed to refresh left pane: %v", err))
	}

	// Refresh right pane
	if err := m.rightPane.Refresh(); err != nil {
		m.dialog = NewErrorDialog(fmt.Sprintf("Failed to refresh right pane: %v", err))
	}

	// Update disk space
	m.updateDiskSpace()

	return tea.Batch(cmds...)
}

// SyncOppositePane synchronizes the opposite pane to the active pane's directory
func (m *Model) SyncOppositePane() {
	activePane := m.getActivePane()
	oppositePane := m.getInactivePane()

	if err := oppositePane.SyncTo(activePane.path); err != nil {
		m.dialog = NewErrorDialog(fmt.Sprintf("Failed to sync pane: %v", err))
	}
}

// handleCreateFile は新規ファイル作成を処理
func (m *Model) handleCreateFile(dirPath, filename string) tea.Cmd {
	return func() tea.Msg {
		// バリデーション
		if err := fs.ValidateFilename(filename); err != nil {
			return inputDialogResultMsg{
				operation: "create_file",
				err:       err,
			}
		}

		fullPath := filepath.Join(dirPath, filename)
		if err := fs.CreateFile(fullPath); err != nil {
			return inputDialogResultMsg{
				operation: "create_file",
				err:       err,
			}
		}

		return inputDialogResultMsg{
			operation: "create_file",
			input:     filename,
		}
	}
}

// handleCreateDirectory は新規ディレクトリ作成を処理
func (m *Model) handleCreateDirectory(dirPath, dirname string) tea.Cmd {
	return func() tea.Msg {
		// バリデーション
		if err := fs.ValidateFilename(dirname); err != nil {
			return inputDialogResultMsg{
				operation: "create_dir",
				err:       err,
			}
		}

		fullPath := filepath.Join(dirPath, dirname)
		if err := fs.CreateDirectory(fullPath); err != nil {
			return inputDialogResultMsg{
				operation: "create_dir",
				err:       err,
			}
		}

		return inputDialogResultMsg{
			operation: "create_dir",
			input:     dirname,
		}
	}
}

// handleRename はリネームを処理
func (m *Model) handleRename(dirPath, oldName, newName string) tea.Cmd {
	return func() tea.Msg {
		// バリデーション
		if err := fs.ValidateFilename(newName); err != nil {
			return inputDialogResultMsg{
				operation: "rename",
				err:       err,
			}
		}

		oldPath := filepath.Join(dirPath, oldName)
		if err := fs.Rename(oldPath, newName); err != nil {
			return inputDialogResultMsg{
				operation: "rename",
				err:       err,
			}
		}

		return inputDialogResultMsg{
			operation: "rename",
			input:     newName,
			oldName:   oldName,
		}
	}
}

// moveCursorToFile は作成されたファイルにカーソルを移動
func (m *Model) moveCursorToFile(filename string) {
	pane := m.getActivePane()

	// 隠しファイルで表示OFFの場合はカーソル移動しない
	if strings.HasPrefix(filename, ".") && !pane.showHidden {
		return
	}

	// ファイルを探してカーソルを移動
	for i, entry := range pane.entries {
		if entry.Name == filename {
			pane.cursor = i
			pane.EnsureCursorVisible()
			return
		}
	}
}

// moveCursorToFileAfterRename はリネーム後にカーソルを移動
func (m *Model) moveCursorToFileAfterRename(oldName, newName string) {
	pane := m.getActivePane()

	// 隠しファイルにリネームされ、表示OFFの場合
	if strings.HasPrefix(newName, ".") && !pane.showHidden {
		// 現在のカーソル位置が有効範囲を超えていたら調整
		if pane.cursor >= len(pane.entries) {
			if len(pane.entries) > 0 {
				pane.cursor = len(pane.entries) - 1
			} else {
				pane.cursor = 0
			}
		}
		pane.EnsureCursorVisible()
		return
	}

	// リネームされたファイルを探してカーソルを移動
	m.moveCursorToFile(newName)
}

// checkFileConflict checks if destination file exists and returns appropriate action
func (m *Model) checkFileConflict(srcPath, destDir, operation string) tea.Cmd {
	filename := filepath.Base(srcPath)
	destPath := filepath.Join(destDir, filename)

	// Check destination using Lstat to handle symlinks properly
	destInfo, err := os.Lstat(destPath)
	if os.IsNotExist(err) {
		// No conflict - execute immediately
		return m.executeFileOperation(srcPath, destDir, operation)
	}

	if err != nil {
		// Other error
		return func() tea.Msg {
			return showErrorDialogMsg{message: fmt.Sprintf("Failed to check destination: %v", err)}
		}
	}

	srcInfo, err := os.Lstat(srcPath)
	if err != nil {
		return func() tea.Msg {
			return showErrorDialogMsg{message: fmt.Sprintf("Failed to check source: %v", err)}
		}
	}

	// Directory conflict - show error dialog
	if srcInfo.IsDir() && destInfo.IsDir() {
		return func() tea.Msg {
			return showErrorDialogMsg{
				message: fmt.Sprintf("Directory \"%s\" already exists in\n%s", filename, destDir),
			}
		}
	}

	// File conflict - show overwrite dialog
	return func() tea.Msg {
		return showOverwriteDialogMsg{
			filename:  filename,
			srcPath:   srcPath,
			destPath:  destDir,
			srcInfo:   OverwriteFileInfo{Size: srcInfo.Size(), ModTime: srcInfo.ModTime()},
			destInfo:  OverwriteFileInfo{Size: destInfo.Size(), ModTime: destInfo.ModTime()},
			operation: operation,
		}
	}
}

// executeFileOperation executes a copy or move operation
func (m *Model) executeFileOperation(srcPath, destPath, operation string) tea.Cmd {
	return func() tea.Msg {
		var err error
		if operation == "copy" {
			err = fs.Copy(srcPath, destPath)
		} else {
			err = fs.MoveFile(srcPath, destPath)
		}

		if err != nil {
			return showErrorDialogMsg{message: fmt.Sprintf("Failed to %s: %v", operation, err)}
		}
		return fileOperationCompleteMsg{operation: operation}
	}
}

// showErrorDialogMsg is a message to show an error dialog
type showErrorDialogMsg struct {
	message string
}

// showOverwriteDialogMsg is a message to show the overwrite confirmation dialog
type showOverwriteDialogMsg struct {
	filename  string
	srcPath   string
	destPath  string
	srcInfo   OverwriteFileInfo
	destInfo  OverwriteFileInfo
	operation string
}

// fileOperationCompleteMsg is sent when a file operation completes successfully
type fileOperationCompleteMsg struct {
	operation string
}

// startBatchOperation initializes a batch copy/move operation.
// Delegates to BatchOperationManager and processes the first file.
func (m *Model) startBatchOperation(files []string, operation string) tea.Cmd {
	srcDir := m.getActivePane().Path()
	destDir := m.getInactivePane().Path()

	// Start the batch operation via manager
	m.batchOpManager.Start(files, srcDir, destDir, operation)

	// Process first file
	return m.processBatchFile()
}

// processBatchFile processes the current file in the batch operation.
func (m *Model) processBatchFile() tea.Cmd {
	srcPath := m.batchOpManager.CurrentFile()
	if srcPath == "" {
		// No more files - complete the operation
		return m.batchOpManager.Advance(true, "") // triggers completion
	}

	return m.checkFileConflict(srcPath, m.batchOpManager.DestPath(), m.batchOpManager.Operation())
}

// advanceBatchOperation moves to the next file in the batch.
func (m *Model) advanceBatchOperation(success bool, srcPath string) tea.Cmd {
	return m.batchOpManager.Advance(success, srcPath)
}

// cancelBatchOperation cancels the remaining batch operation.
func (m *Model) cancelBatchOperation() tea.Cmd {
	return m.batchOpManager.Cancel()
}

// showStatusMsg is a message to show a status message
type showStatusMsg struct {
	message string
	isError bool
}

// bookmarkAddedMsg is sent when a bookmark is successfully added
type bookmarkAddedMsg struct {
	bookmarks []config.Bookmark
	alias     string
}

// bookmarkEditedMsg is sent when a bookmark is successfully edited
type bookmarkEditedMsg struct {
	bookmarks []config.Bookmark
	alias     string
}

// generateDefaultArchiveName creates a default archive name based on source files
func (m *Model) generateDefaultArchiveName(sources []string, format archive.ArchiveFormat) string {
	if len(sources) == 0 {
		return "archive" + format.Extension()
	}

	if len(sources) == 1 {
		// Use source filename/dirname as base
		base := filepath.Base(sources[0])
		// Remove any existing extension for files
		if !isDirectory(sources[0]) {
			ext := filepath.Ext(base)
			if ext != "" {
				base = strings.TrimSuffix(base, ext)
			}
		}
		return base + format.Extension()
	}

	// Multiple sources - use parent directory name or "archive"
	parentDir := filepath.Dir(sources[0])
	dirName := filepath.Base(parentDir)
	if dirName == "" || dirName == "." || dirName == "/" {
		return "archive" + format.Extension()
	}
	return dirName + format.Extension()
}

// isDirectory checks if path is a directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// startArchiveCompression starts the background archive compression task.
// Shows progress dialog and delegates to ArchiveOperationManager.
func (m *Model) startArchiveCompression(archivePath string) tea.Cmd {
	if !m.archiveOpManager.IsActive() {
		return nil
	}

	// Show progress dialog with cancel callback
	progressDialog := NewArchiveProgressDialog("compress", archivePath)
	progressDialog.SetOnCancel(func() {
		m.archiveOpManager.CancelTask()
	})
	m.dialog = progressDialog

	return m.archiveOpManager.StartCompression(archivePath)
}

// pollArchiveProgress polls for archive operation progress.
// Delegates to ArchiveOperationManager.
func (m *Model) pollArchiveProgress(taskID string) tea.Cmd {
	return m.archiveOpManager.PollProgress(taskID)
}

// checkExtractSecurity performs security checks before archive extraction.
// Delegates to ArchiveOperationManager.
func (m *Model) checkExtractSecurity(archivePath, destDir string) tea.Cmd {
	return m.archiveOpManager.CheckSecurity(archivePath, destDir)
}

// startArchiveExtraction starts the background archive extraction task.
// Shows progress dialog and delegates to ArchiveOperationManager.
func (m *Model) startArchiveExtraction(archivePath, destDir string) tea.Cmd {
	// Prepare state via manager
	m.archiveOpManager.PrepareExtraction(archivePath, destDir)

	// Show progress dialog with cancel callback
	progressDialog := NewArchiveProgressDialog("extract", archivePath)
	progressDialog.SetOnCancel(func() {
		m.archiveOpManager.CancelTask()
	})
	m.dialog = progressDialog

	return m.archiveOpManager.StartExtraction(archivePath, destDir)
}
