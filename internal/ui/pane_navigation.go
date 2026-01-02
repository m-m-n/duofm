package ui

import (
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
)

// StartLoadingDirectory はローディング状態を開始
func (p *Pane) StartLoadingDirectory() {
	p.loading = true
	p.loadingProgress = "Loading directory..."
}

// LoadDirectoryAsync は非同期でディレクトリを読み込む
func LoadDirectoryAsync(paneID PanePosition, panePath string, sortConfig SortConfig) tea.Cmd {
	return loadDirectoryAsyncInternal(paneID, panePath, sortConfig, false, false)
}

// LoadDirectoryAsyncWithHistory は非同期でディレクトリを読み込む（履歴ナビゲーションフラグ付き）
// isForward: true=前進操作、false=後退操作（エラー時の復元用）
func LoadDirectoryAsyncWithHistory(paneID PanePosition, panePath string, sortConfig SortConfig, isForward bool) tea.Cmd {
	return loadDirectoryAsyncInternal(paneID, panePath, sortConfig, true, isForward)
}

// loadDirectoryAsyncInternal は非同期でディレクトリを読み込む内部関数
func loadDirectoryAsyncInternal(paneID PanePosition, panePath string, sortConfig SortConfig, isHistoryNavigation bool, isForward bool) tea.Cmd {
	return func() tea.Msg {
		entries, err := fs.ReadDirectory(panePath)
		if err != nil {
			return directoryLoadCompleteMsg{
				paneID:                   paneID,
				panePath:                 panePath,
				entries:                  nil,
				err:                      err,
				attemptedPath:            panePath,
				isHistoryNavigation:      isHistoryNavigation,
				historyNavigationForward: isForward,
			}
		}

		entries = SortEntries(entries, sortConfig)
		return directoryLoadCompleteMsg{
			paneID:                   paneID,
			panePath:                 panePath,
			entries:                  entries,
			err:                      nil,
			attemptedPath:            panePath,
			isHistoryNavigation:      isHistoryNavigation,
			historyNavigationForward: isForward,
		}
	}
}

// recordPreviousPath はナビゲーション前に現在のパスを記録する
func (p *Pane) recordPreviousPath() {
	p.previousPath = p.path
}

// addToHistory は現在のパスを履歴に追加する
// 通常のディレクトリ遷移で呼び出され、履歴ナビゲーション自体では呼び出されない
func (p *Pane) addToHistory() {
	p.history.AddToHistory(p.path)
}

// restorePreviousPath は読み込み失敗時に前のパスに復元する
func (p *Pane) restorePreviousPath() {
	if p.previousPath != "" {
		p.path = p.previousPath
		p.pendingPath = ""
	}
}

// EnterDirectoryAsync はディレクトリへの移動を開始し、Cmdを返す
// 親ディレクトリ(..)に入る場合は、読み込み完了後に直前のサブディレクトリに
// カーソルを合わせるため、pendingCursorTarget にサブディレクトリ名を保存する
func (p *Pane) EnterDirectoryAsync() tea.Cmd {
	entry := p.SelectedEntry()
	if entry == nil {
		return nil
	}

	// シンボリックリンクの処理
	if entry.IsSymlink {
		if entry.LinkBroken {
			// リンク切れの場合は何もしない
			return nil
		}

		// リンク先がディレクトリかチェック
		isDir, err := fs.IsDirectory(entry.LinkTarget)
		if err != nil || !isDir {
			// リンク先がファイルまたはエラーの場合は何もしない
			return nil
		}
	}

	// 通常のディレクトリ処理
	if !entry.IsDir && !entry.IsSymlink {
		return nil // ファイルの場合は何もしない
	}

	var newPath string
	if entry.IsParentDir() {
		// 親ディレクトリに移動 - サブディレクトリ名を記憶
		p.pendingCursorTarget = p.extractSubdirName()
		newPath = filepath.Dir(p.path)
	} else {
		// サブディレクトリに移動（シンボリックリンク含む）- カーソル記憶をクリア
		p.pendingCursorTarget = ""
		newPath = filepath.Join(p.path, entry.Name)
	}

	// 現在のパスを記録（復元用）
	// 履歴への追加は成功時に directoryLoadCompleteMsg ハンドラで行う
	p.recordPreviousPath()
	p.pendingPath = newPath
	p.path = newPath

	// ローディング状態を開始
	p.StartLoadingDirectory()

	return LoadDirectoryAsync(p.paneID, newPath, p.sortConfig)
}

// EnterDirectory はディレクトリに入る
// 親ディレクトリ(..)に入る場合は、直前にいたサブディレクトリにカーソルを合わせる
func (p *Pane) EnterDirectory() error {
	entry := p.SelectedEntry()
	if entry == nil {
		return nil
	}

	// シンボリックリンクの処理
	if entry.IsSymlink {
		if entry.LinkBroken {
			// リンク切れの場合は何もしない
			return nil
		}

		// リンク先がディレクトリかチェック
		isDir, err := fs.IsDirectory(entry.LinkTarget)
		if err != nil || !isDir {
			// リンク先がファイルまたはエラーの場合は何もしない
			return nil
		}

		// 直前のパスを記録してから論理パス（シンボリックリンク自体のパス）に移動
		// これにより、..で論理的な親ディレクトリに戻れる
		p.recordPreviousPath()
		p.path = filepath.Join(p.path, entry.Name)
		if err := p.LoadDirectory(); err != nil {
			return err
		}
		// 成功時に履歴に追加
		p.addToHistory()
		return nil
	}

	// 通常のディレクトリ処理
	if !entry.IsDir {
		return nil // ファイルの場合は何もしない
	}

	var newPath string
	var subdirName string // 親ディレクトリ遷移時のカーソル位置決定用

	if entry.IsParentDir() {
		// 親ディレクトリに移動 - サブディレクトリ名を記憶
		subdirName = p.extractSubdirName()
		newPath = filepath.Dir(p.path)
	} else {
		// サブディレクトリに移動
		newPath = filepath.Join(p.path, entry.Name)
	}

	// 直前のパスを記録（復元用）
	p.recordPreviousPath()
	p.path = newPath

	if err := p.LoadDirectory(); err != nil {
		return err
	}

	// 成功時に履歴に追加
	p.addToHistory()

	// 親ディレクトリ遷移の場合、直前のサブディレクトリにカーソルを合わせる
	if subdirName != "" {
		if index := p.findEntryIndex(subdirName); index >= 0 {
			p.cursor = index
			p.adjustScroll()
		}
	}

	return nil
}

// MoveToParent は親ディレクトリに移動
// 移動後、直前にいたサブディレクトリにカーソルを合わせる
func (p *Pane) MoveToParent() error {
	if p.path == "/" {
		return nil // ルートより上には行けない
	}

	// 親ディレクトリ遷移前にサブディレクトリ名を記憶
	subdirName := p.extractSubdirName()

	p.recordPreviousPath()
	p.path = filepath.Dir(p.path)

	if err := p.LoadDirectory(); err != nil {
		return err
	}

	// 成功時に履歴に追加
	p.addToHistory()

	// 直前のサブディレクトリにカーソルを合わせる
	if index := p.findEntryIndex(subdirName); index >= 0 {
		p.cursor = index
		p.adjustScroll()
	}
	// 見つからない場合は LoadDirectory() で設定された cursor = 0 のまま

	return nil
}

// MoveToParentAsync は親ディレクトリへの移動を開始
// 読み込み完了後に直前のサブディレクトリにカーソルを合わせるため、
// pendingCursorTarget にサブディレクトリ名を保存する
func (p *Pane) MoveToParentAsync() tea.Cmd {
	if p.path == "/" {
		return nil
	}

	// 親ディレクトリ遷移後のカーソル位置決定用にサブディレクトリ名を記憶
	p.pendingCursorTarget = p.extractSubdirName()

	newPath := filepath.Dir(p.path)
	p.recordPreviousPath()
	// 履歴への追加は成功時に directoryLoadCompleteMsg ハンドラで行う
	p.pendingPath = newPath
	p.path = newPath
	p.StartLoadingDirectory()
	return LoadDirectoryAsync(p.paneID, newPath, p.sortConfig)
}

// ChangeDirectory は指定されたパスに移動
func (p *Pane) ChangeDirectory(path string) error {
	p.recordPreviousPath()
	p.path = path
	if err := p.LoadDirectory(); err != nil {
		return err
	}
	// 成功時に履歴に追加
	p.addToHistory()
	return nil
}

// ChangeDirectoryAsync は指定パスへの移動を開始
func (p *Pane) ChangeDirectoryAsync(path string) tea.Cmd {
	p.pendingCursorTarget = "" // 非親ディレクトリ遷移ではカーソル記憶をクリア
	p.recordPreviousPath()
	// 履歴への追加は成功時に directoryLoadCompleteMsg ハンドラで行う
	p.pendingPath = path
	p.path = path
	p.StartLoadingDirectory()
	return LoadDirectoryAsync(p.paneID, path, p.sortConfig)
}

// NavigateToHome はホームディレクトリに移動する
func (p *Pane) NavigateToHome() error {
	home, err := fs.HomeDirectory()
	if err != nil {
		return err
	}

	// すでにホームにいる場合は何もしない
	if p.path == home {
		return nil
	}

	p.recordPreviousPath()
	p.path = home
	if err := p.LoadDirectory(); err != nil {
		return err
	}
	// 成功時に履歴に追加
	p.addToHistory()
	return nil
}

// NavigateToHomeAsync はホームディレクトリへの移動を開始
func (p *Pane) NavigateToHomeAsync() tea.Cmd {
	home, err := fs.HomeDirectory()
	if err != nil {
		return nil
	}
	if p.path == home {
		return nil
	}
	p.pendingCursorTarget = "" // 非親ディレクトリ遷移ではカーソル記憶をクリア
	p.recordPreviousPath()
	// 履歴への追加は成功時に directoryLoadCompleteMsg ハンドラで行う
	p.pendingPath = home
	p.path = home
	p.StartLoadingDirectory()
	return LoadDirectoryAsync(p.paneID, home, p.sortConfig)
}

// NavigateToPrevious は直前のディレクトリに移動する（トグル動作）
func (p *Pane) NavigateToPrevious() error {
	if p.previousPath == "" {
		return nil // 履歴がない場合は何もしない
	}

	// 現在のパスと直前のパスをスワップ（トグル動作）
	current := p.path
	p.path = p.previousPath
	p.previousPath = current

	if err := p.LoadDirectory(); err != nil {
		return err
	}
	// 成功時に履歴に追加（トグル動作でも記録）
	p.addToHistory()
	return nil
}

// NavigateToPreviousAsync は直前のディレクトリへの移動を開始（トグル動作）
func (p *Pane) NavigateToPreviousAsync() tea.Cmd {
	if p.previousPath == "" {
		return nil
	}
	p.pendingCursorTarget = "" // 非親ディレクトリ遷移ではカーソル記憶をクリア

	// 履歴への追加は成功時に directoryLoadCompleteMsg ハンドラで行う（トグル動作でも記録）

	current := p.path
	p.pendingPath = p.previousPath
	p.path = p.previousPath
	p.previousPath = current
	p.StartLoadingDirectory()
	return LoadDirectoryAsync(p.paneID, p.path, p.sortConfig)
}

// Refresh reloads the current directory, preserving cursor position and marks
func (p *Pane) Refresh() error {
	// Save currently selected filename
	var selectedName string
	if p.cursor >= 0 && p.cursor < len(p.entries) {
		selectedName = p.entries[p.cursor].Name
	}
	savedCursor := p.cursor

	// Save marks before reload
	savedMarks := make(map[string]bool)
	for k, v := range p.markedFiles {
		savedMarks[k] = v
	}

	// Reload directory with existence check
	currentPath := p.path
	for {
		if fs.DirectoryExists(currentPath) {
			break
		}
		// Navigate up to parent directory
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			// Reached root but it doesn't exist
			home, err := fs.HomeDirectory()
			if err == nil && fs.DirectoryExists(home) {
				currentPath = home
				break
			}
			currentPath = "/"
			break
		}
		currentPath = parent
	}

	if currentPath != p.path {
		// Directory was changed, update previousPath for navigation history
		p.previousPath = p.path
		p.path = currentPath
		// Don't restore marks when directory changes
		savedMarks = nil
	}

	err := p.LoadDirectory()
	if err != nil {
		return err
	}

	// Restore marks for files that still exist
	if savedMarks != nil {
		existingFiles := make(map[string]bool)
		for _, entry := range p.allEntries {
			existingFiles[entry.Name] = true
		}
		for filename := range savedMarks {
			if existingFiles[filename] {
				p.markedFiles[filename] = true
			}
		}
	}

	// Restore cursor position
	if selectedName != "" {
		// Search for the same filename
		for i, e := range p.entries {
			if e.Name == selectedName {
				p.cursor = i
				p.adjustScroll()
				return nil
			}
		}
	}

	// If file not found, use previous index
	if savedCursor < len(p.entries) {
		p.cursor = savedCursor
	} else if len(p.entries) > 0 {
		p.cursor = len(p.entries) - 1
	} else {
		p.cursor = 0
	}
	p.adjustScroll()

	return nil
}

// SyncTo synchronizes this pane to the specified directory
// Preserves display settings but resets cursor to top
func (p *Pane) SyncTo(path string) error {
	// Do nothing if already in the same directory
	if p.path == path {
		return nil
	}

	// Update previousPath for navigation history
	p.previousPath = p.path

	// Change directory
	p.path = path
	err := p.LoadDirectory()
	if err != nil {
		return err
	}

	// 成功時に履歴に追加
	p.addToHistory()

	// Reset cursor and scroll to top
	p.cursor = 0
	p.scrollOffset = 0

	return nil
}

// NavigateHistoryBack はディレクトリ履歴を遡って移動する
// 成功時はディレクトリを読み込んでnilを返す
// 履歴がない場合や、ディレクトリが存在しない場合はエラーを返す
func (p *Pane) NavigateHistoryBack() error {
	path, ok := p.history.NavigateBack()
	if !ok {
		return nil // 履歴がない場合は何もしない
	}

	// 履歴ナビゲーションではpreviousPathを更新しない（独立動作）
	// また、addToHistoryも呼ばない（履歴ナビゲーション自体は記録しない）
	oldPath := p.path // エラー時の復元用
	p.path = path

	// ディレクトリが存在するか確認
	if err := p.LoadDirectory(); err != nil {
		// エラーの場合はパスと履歴位置を元に戻す（ナビゲーションをキャンセル）
		p.path = oldPath
		p.history.NavigateForward()
		return err
	}

	return nil
}

// NavigateHistoryForward はディレクトリ履歴を進んで移動する
// 成功時はディレクトリを読み込んでnilを返す
// 履歴がない場合や、ディレクトリが存在しない場合はエラーを返す
func (p *Pane) NavigateHistoryForward() error {
	path, ok := p.history.NavigateForward()
	if !ok {
		return nil // 履歴がない場合は何もしない
	}

	// 履歴ナビゲーションではpreviousPathを更新しない（独立動作）
	// また、addToHistoryも呼ばない（履歴ナビゲーション自体は記録しない）
	oldPath := p.path // エラー時の復元用
	p.path = path

	// ディレクトリが存在するか確認
	if err := p.LoadDirectory(); err != nil {
		// エラーの場合はパスと履歴位置を元に戻す（ナビゲーションをキャンセル）
		p.path = oldPath
		p.history.NavigateBack()
		return err
	}

	return nil
}

// NavigateHistoryBackAsync はディレクトリ履歴を遡って移動を開始
func (p *Pane) NavigateHistoryBackAsync() tea.Cmd {
	path, ok := p.history.NavigateBack()
	if !ok {
		return nil // 履歴がない場合は何もしない
	}

	// 履歴ナビゲーションではカーソル記憶をクリア
	p.pendingCursorTarget = ""

	// 履歴ナビゲーションでは addToHistory を呼ばない（isHistoryNavigation=trueで識別）
	// 履歴ナビゲーションでは previousPath を更新しない（- キーとは独立動作）
	p.pendingPath = path
	p.path = path
	p.StartLoadingDirectory()
	// isForward=false（後退操作）
	return LoadDirectoryAsyncWithHistory(p.paneID, path, p.sortConfig, false)
}

// NavigateHistoryForwardAsync はディレクトリ履歴を進んで移動を開始
func (p *Pane) NavigateHistoryForwardAsync() tea.Cmd {
	path, ok := p.history.NavigateForward()
	if !ok {
		return nil // 履歴がない場合は何もしない
	}

	// 履歴ナビゲーションではカーソル記憶をクリア
	p.pendingCursorTarget = ""

	// 履歴ナビゲーションでは addToHistory を呼ばない（isHistoryNavigation=trueで識別）
	// 履歴ナビゲーションでは previousPath を更新しない（- キーとは独立動作）
	p.pendingPath = path
	p.path = path
	p.StartLoadingDirectory()
	// isForward=true（前進操作）
	return LoadDirectoryAsyncWithHistory(p.paneID, path, p.sortConfig, true)
}
