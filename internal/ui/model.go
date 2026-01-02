package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/archive"
	"github.com/sakura/duofm/internal/config"
	"github.com/sakura/duofm/internal/fs"
)

// ANSIエスケープシーケンスを除去するための正規表現
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// BatchOperation holds state for batch file operations
type BatchOperation struct {
	Files      []string // List of source file paths
	CurrentIdx int      // Current file index
	DestPath   string   // Destination directory
	Operation  string   // "copy" or "move"
	Completed  []string // Successfully completed files
	Failed     []string // Failed files
}

// ArchiveOperationState holds state for in-progress archive operations
type ArchiveOperationState struct {
	Sources     []string              // Source files/directories to archive
	DestDir     string                // Destination directory
	Format      archive.ArchiveFormat // Selected archive format
	Level       int                   // Compression level (0-9)
	ArchiveName string                // Archive filename
	TaskID      string                // Task ID for background operation
}

// Model はアプリケーション全体の状態を保持
type Model struct {
	leftPane           *Pane
	rightPane          *Pane
	leftPath           string
	rightPath          string
	activePane         PanePosition
	dialog             Dialog
	width              int
	height             int
	ready              bool
	lastDiskSpaceCheck time.Time                  // 最後のディスク容量チェック時刻
	leftDiskSpace      uint64                     // 左ペインのディスク空き容量
	rightDiskSpace     uint64                     // 右ペインのディスク空き容量
	pendingAction      func() error               // 確認待ちのアクション（コンテキストメニューの削除用）
	statusMessage      string                     // ステータスバーに表示するメッセージ
	isStatusError      bool                       // エラーメッセージかどうか
	searchState        SearchState                // 検索状態
	minibuffer         *Minibuffer                // ミニバッファ
	ctrlCPending       bool                       // Ctrl+Cが1回押された状態かどうか
	batchOp            *BatchOperation            // Active batch operation (nil if none)
	sortDialog         *SortDialog                // ソートダイアログ（nil = 非表示）
	shellCommandMode   bool                       // シェルコマンドモードかどうか
	keybindingMap      *KeybindingMap             // キーバインドマップ
	configWarnings     []string                   // 設定ファイルの警告
	theme              *Theme                     // カラーテーマ
	bookmarks          []config.Bookmark          // ブックマークリスト
	bookmarkEditIndex  int                        // 編集中のブックマークインデックス
	archiveOp          *ArchiveOperationState     // アーカイブ操作の状態
	archiveController  *archive.ArchiveController // アーカイブコントローラー
}

// PanePosition はペインの位置を表す
type PanePosition int

const (
	// LeftPane は左ペイン
	LeftPane PanePosition = iota
	// RightPane は右ペイン
	RightPane
)

// NewModel は初期モデルを作成（デフォルトキーバインドを使用）
func NewModel() Model {
	return NewModelWithConfig(nil, nil, nil)
}

// NewModelWithConfig は設定付きの初期モデルを作成
func NewModelWithConfig(keybindingMap *KeybindingMap, theme *Theme, warnings []string) Model {
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
	var bookmarks []config.Bookmark
	configPath, configErr := config.GetConfigPath()
	if configErr != nil {
		warnings = append(warnings, fmt.Sprintf("Warning: failed to get config path: %v", configErr))
	} else {
		var bookmarkWarnings []string
		bookmarks, bookmarkWarnings = config.LoadBookmarks(configPath)
		warnings = append(warnings, bookmarkWarnings...)
	}

	return Model{
		leftPane:          nil, // Updateで初期化
		rightPane:         nil, // Updateで初期化
		leftPath:          cwd,
		rightPath:         home,
		activePane:        LeftPane,
		dialog:            nil,
		ready:             false,
		searchState:       SearchState{Mode: SearchModeNone},
		minibuffer:        NewMinibuffer(),
		keybindingMap:     keybindingMap,
		configWarnings:    warnings,
		theme:             theme,
		bookmarks:         bookmarks,
		bookmarkEditIndex: -1,
		archiveController: archive.NewArchiveController(),
	}
}

// Init はBubble Teaの初期化
func (m Model) Init() tea.Cmd {
	// 設定ファイルの警告があれば最初の警告をステータスバーに表示
	if len(m.configWarnings) > 0 {
		m.statusMessage = m.configWarnings[0]
		m.isStatusError = false
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
	if m.leftPane != nil {
		if freeBytes, _, err := fs.GetDiskSpace(m.leftPane.Path()); err == nil {
			m.leftDiskSpace = freeBytes
		}
	}

	if m.rightPane != nil {
		if freeBytes, _, err := fs.GetDiskSpace(m.rightPane.Path()); err == nil {
			m.rightDiskSpace = freeBytes
		}
	}

	m.lastDiskSpaceCheck = time.Now()
}

// startShellCommandMode はシェルコマンドモードを開始する
func (m *Model) startShellCommandMode() {
	m.shellCommandMode = true
	m.minibuffer.SetPrompt("!: ")
	m.minibuffer.Clear()
	m.minibuffer.SetWidth(m.getActivePane().width)
	m.minibuffer.Show()
}

// startSearch は検索モードを開始する
func (m *Model) startSearch(mode SearchMode) {
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

	// ミニバッファの設定
	if mode == SearchModeIncremental {
		m.minibuffer.SetPrompt("/: ")
	} else {
		m.minibuffer.SetPrompt("(search): ")
	}
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

// batchFileCompleteMsg is sent when one file in a batch operation completes
type batchFileCompleteMsg struct {
	success bool
	srcPath string
}

// startBatchOperation initializes a batch copy/move operation
func (m *Model) startBatchOperation(files []string, operation string) tea.Cmd {
	srcDir := m.getActivePane().Path()
	destDir := m.getInactivePane().Path()

	// Build full paths
	fullPaths := make([]string, len(files))
	for i, f := range files {
		fullPaths[i] = filepath.Join(srcDir, f)
	}

	m.batchOp = &BatchOperation{
		Files:      fullPaths,
		CurrentIdx: 0,
		DestPath:   destDir,
		Operation:  operation,
		Completed:  make([]string, 0),
		Failed:     make([]string, 0),
	}

	// Process first file
	return m.processBatchFile()
}

// processBatchFile processes the current file in the batch operation
func (m *Model) processBatchFile() tea.Cmd {
	if m.batchOp == nil || m.batchOp.CurrentIdx >= len(m.batchOp.Files) {
		return m.completeBatchOperation()
	}

	srcPath := m.batchOp.Files[m.batchOp.CurrentIdx]
	return m.checkFileConflict(srcPath, m.batchOp.DestPath, m.batchOp.Operation)
}

// advanceBatchOperation moves to the next file in the batch
func (m *Model) advanceBatchOperation(success bool, srcPath string) tea.Cmd {
	if m.batchOp == nil {
		return nil
	}

	if success {
		m.batchOp.Completed = append(m.batchOp.Completed, srcPath)
	} else {
		m.batchOp.Failed = append(m.batchOp.Failed, srcPath)
	}

	m.batchOp.CurrentIdx++
	return m.processBatchFile()
}

// completeBatchOperation finishes the batch operation
func (m *Model) completeBatchOperation() tea.Cmd {
	if m.batchOp == nil {
		return nil
	}

	operation := m.batchOp.Operation
	completed := len(m.batchOp.Completed)
	m.batchOp = nil

	// Clear marks
	m.getActivePane().ClearMarks()

	// Reload both panes
	m.getActivePane().LoadDirectory()
	m.getInactivePane().LoadDirectory()

	return func() tea.Msg {
		return batchOperationCompleteMsg{operation: operation, count: completed}
	}
}

// cancelBatchOperation cancels the remaining batch operation
func (m *Model) cancelBatchOperation() {
	if m.batchOp == nil {
		return
	}

	// Clear marks and batch state
	m.getActivePane().ClearMarks()
	m.batchOp = nil

	// Reload both panes
	m.getActivePane().LoadDirectory()
	m.getInactivePane().LoadDirectory()
}

// batchOperationCompleteMsg is sent when a batch operation finishes
type batchOperationCompleteMsg struct {
	operation string
	count     int
}

// handleAddBookmark はブックマーク追加処理
func (m *Model) handleAddBookmark(currentBookmarks []config.Bookmark, path, alias string) tea.Cmd {
	return func() tea.Msg {
		newBookmarks, err := config.AddBookmark(currentBookmarks, alias, path)
		if err != nil {
			if err == config.ErrEmptyAlias {
				return showStatusMsg{message: "Bookmark name cannot be empty", isError: true}
			}
			if err == config.ErrDuplicatePath {
				return showStatusMsg{message: "Already bookmarked", isError: false}
			}
			return showStatusMsg{message: fmt.Sprintf("Failed to add bookmark: %v", err), isError: true}
		}

		// 設定ファイルに保存
		if saveErr := saveBookmarksToConfig(newBookmarks); saveErr != nil {
			return showStatusMsg{message: saveErr.Error(), isError: true}
		}

		return bookmarkAddedMsg{bookmarks: newBookmarks, alias: alias}
	}
}

// handleBookmarkEdit はブックマーク編集処理
func (m *Model) handleBookmarkEdit(currentBookmarks []config.Bookmark, index int, newAlias string) tea.Cmd {
	return func() tea.Msg {
		if index < 0 || index >= len(currentBookmarks) {
			return showStatusMsg{message: "Invalid bookmark index", isError: true}
		}

		newBookmarks, err := config.UpdateBookmarkAlias(currentBookmarks, index, newAlias)
		if err != nil {
			if err == config.ErrEmptyAlias {
				return showStatusMsg{message: "Bookmark name cannot be empty", isError: true}
			}
			return showStatusMsg{message: fmt.Sprintf("Failed to edit bookmark: %v", err), isError: true}
		}

		// 設定ファイルに保存
		if saveErr := saveBookmarksToConfig(newBookmarks); saveErr != nil {
			return showStatusMsg{message: saveErr.Error(), isError: true}
		}

		return bookmarkEditedMsg{bookmarks: newBookmarks, alias: newAlias}
	}
}

// saveBookmarksToConfig saves bookmarks to the configuration file.
// Returns an error with a user-friendly message if saving fails.
func saveBookmarksToConfig(bookmarks []config.Bookmark) error {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}
	if err := config.SaveBookmarks(configPath, bookmarks); err != nil {
		return fmt.Errorf("failed to save bookmarks: %w", err)
	}
	return nil
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

// startArchiveCompression starts the background archive compression task
func (m *Model) startArchiveCompression(archivePath string) tea.Cmd {
	if m.archiveOp == nil || m.archiveController == nil {
		return nil
	}

	sources := m.archiveOp.Sources
	format := m.archiveOp.Format
	level := m.archiveOp.Level
	controller := m.archiveController

	// Show progress dialog
	progressDialog := NewArchiveProgressDialog("compress", archivePath)
	progressDialog.SetOnCancel(func() {
		if m.archiveOp != nil && m.archiveOp.TaskID != "" {
			controller.CancelTask(m.archiveOp.TaskID)
		}
	})
	m.dialog = progressDialog

	return func() tea.Msg {
		taskID, err := controller.CreateArchive(sources, archivePath, format, level)
		if err != nil {
			return archiveOperationErrorMsg{
				err:     err,
				message: fmt.Sprintf("Failed to start compression: %v", err),
			}
		}
		return archiveOperationStartMsg{taskID: taskID}
	}
}

// pollArchiveProgress polls for archive operation progress
func (m *Model) pollArchiveProgress(taskID string) tea.Cmd {
	if m.archiveController == nil {
		return nil
	}

	controller := m.archiveController

	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		status := controller.GetTaskStatus(taskID)
		if status == nil {
			return archiveOperationCompleteMsg{
				taskID:  taskID,
				success: false,
				err:     fmt.Errorf("task not found"),
			}
		}

		switch status.State {
		case archive.TaskStateRunning:
			// Return progress update
			if status.Progress != nil {
				return archiveProgressUpdateMsg{
					taskID:          taskID,
					progress:        float64(status.Progress.Percentage()) / 100.0,
					processedFiles:  status.Progress.ProcessedFiles,
					totalFiles:      status.Progress.TotalFiles,
					currentFile:     status.Progress.CurrentFile,
					elapsedTime:     status.Progress.ElapsedTime(),
					estimatedRemain: status.Progress.EstimatedRemaining(),
				}
			}
			return archiveProgressUpdateMsg{taskID: taskID}

		case archive.TaskStateCompleted:
			archivePath := ""
			if status.Progress != nil {
				archivePath = status.Progress.ArchivePath
			}
			return archiveOperationCompleteMsg{
				taskID:      taskID,
				success:     true,
				archivePath: archivePath,
			}

		case archive.TaskStateCancelled:
			return archiveOperationCompleteMsg{
				taskID:    taskID,
				cancelled: true,
			}

		case archive.TaskStateFailed:
			return archiveOperationCompleteMsg{
				taskID:  taskID,
				success: false,
				err:     status.Error,
			}

		default:
			return archiveProgressUpdateMsg{taskID: taskID}
		}
	})
}

// checkExtractSecurity performs security checks before archive extraction
func (m *Model) checkExtractSecurity(archivePath, destDir string) tea.Cmd {
	if m.archiveController == nil {
		return nil
	}

	controller := m.archiveController

	return func() tea.Msg {
		// Get archive metadata
		metadata, err := controller.GetArchiveMetadata(archivePath)
		if err != nil {
			return extractSecurityCheckMsg{
				archivePath: archivePath,
				destDir:     destDir,
				err:         err,
			}
		}

		// Check compression ratio (warn if > 1:1000)
		var ratio float64
		compressionOK := true
		if metadata.ArchiveSize > 0 {
			ratio = float64(metadata.ExtractedSize) / float64(metadata.ArchiveSize)
			if ratio > 1000.0 {
				compressionOK = false
			}
		}

		// Check available disk space
		availableSize := archive.GetAvailableDiskSpace(destDir)
		diskSpaceOK := true
		if availableSize > 0 && metadata.ExtractedSize > availableSize {
			diskSpaceOK = false
		}

		return extractSecurityCheckMsg{
			archivePath:   archivePath,
			destDir:       destDir,
			archiveSize:   metadata.ArchiveSize,
			extractedSize: metadata.ExtractedSize,
			availableSize: availableSize,
			compressionOK: compressionOK,
			diskSpaceOK:   diskSpaceOK,
			ratio:         ratio,
		}
	}
}

// startArchiveExtraction starts the background archive extraction task
func (m *Model) startArchiveExtraction(archivePath, destDir string) tea.Cmd {
	if m.archiveController == nil {
		return nil
	}

	controller := m.archiveController

	// Show progress dialog
	progressDialog := NewArchiveProgressDialog("extract", archivePath)
	progressDialog.SetOnCancel(func() {
		if m.archiveOp != nil && m.archiveOp.TaskID != "" {
			controller.CancelTask(m.archiveOp.TaskID)
		}
	})
	m.dialog = progressDialog

	// Initialize archiveOp for extraction tracking
	m.archiveOp = &ArchiveOperationState{
		Sources: []string{archivePath},
		DestDir: destDir,
	}

	return func() tea.Msg {
		taskID, err := controller.ExtractArchive(archivePath, destDir)
		if err != nil {
			return archiveOperationErrorMsg{
				err:     err,
				message: fmt.Sprintf("Failed to start extraction: %v", err),
			}
		}
		return archiveOperationStartMsg{taskID: taskID}
	}
}
