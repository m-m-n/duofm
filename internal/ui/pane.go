package ui

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/fs"
)

// DisplayMode は表示モードを表す
type DisplayMode int

const (
	// DisplayMinimal は名前のみを表示（端末幅が狭い場合に自動）
	DisplayMinimal DisplayMode = iota
	// DisplayBasic は名前 + サイズ + タイムスタンプを表示（デフォルト）
	DisplayBasic
	// DisplayDetail は名前 + パーミッション + 所有者 + グループを表示
	DisplayDetail
)

// dimmedBgColor is used for dialog overlay background.
// Note: Most dimmed colors are now in the theme, but this is kept
// for backward compatibility with model_view.go dialog rendering.
var dimmedBgColor = lipgloss.Color("236") // 濃いグレー背景

// MarkInfo holds mark statistics
type MarkInfo struct {
	Count     int   // Number of marked files
	TotalSize int64 // Total size in bytes
}

// Pane は1つのファイルリストペインを表現
type Pane struct {
	paneID              PanePosition // このペインの識別子（LeftPane or RightPane）
	path                string
	entries             []fs.FileEntry // フィルタ適用後の表示用エントリ
	allEntries          []fs.FileEntry // フィルタ適用前のすべてのエントリ
	cursor              int
	scrollOffset        int
	width               int
	height              int
	isActive            bool
	displayMode         DisplayMode      // ユーザーが選択した表示モード
	loading             bool             // ローディング中フラグ
	loadingProgress     string           // ローディングメッセージ
	showHidden          bool             // 隠しファイル表示フラグ（デフォルト: false）
	previousPath        string           // 直前のディレクトリパス（履歴なしの場合は空文字列）
	pendingPath         string           // 読み込み中の暫定パス（エラー時に元に戻す）
	filterPattern       string           // 現在のフィルタパターン（空の場合はフィルタなし）
	filterMode          SearchMode       // 現在のフィルタモード
	markedFiles         map[string]bool  // key: filename, value: marked state
	sortConfig          SortConfig       // ソート設定
	theme               *Theme           // カラーテーマ
	pendingCursorTarget string           // 親ディレクトリ遷移後のカーソル位置決定用（サブディレクトリ名）
	history             DirectoryHistory // ディレクトリ履歴（ブラウザ風のback/forward）
}

// NewPane は新しいペインを作成
func NewPane(paneID PanePosition, path string, width, height int, isActive bool, theme *Theme) (*Pane, error) {
	if theme == nil {
		theme = DefaultTheme()
	}
	pane := &Pane{
		paneID:          paneID,
		path:            path,
		width:           width,
		height:          height,
		isActive:        isActive,
		cursor:          0,
		scrollOffset:    0,
		displayMode:     DisplayBasic, // デフォルトは基本情報モード
		loading:         false,
		loadingProgress: "",
		markedFiles:     make(map[string]bool),
		sortConfig:      DefaultSortConfig(), // デフォルトは名前昇順
		theme:           theme,
		history:         NewDirectoryHistory(),
	}

	if err := pane.LoadDirectory(); err != nil {
		return nil, err
	}

	// 初期ディレクトリを履歴に追加
	pane.history.AddToHistory(pane.path)

	return pane, nil
}

// LoadDirectory はディレクトリを読み込む（同期版）
func (p *Pane) LoadDirectory() error {
	entries, err := fs.ReadDirectory(p.path)
	if err != nil {
		return err
	}

	entries = SortEntries(entries, p.sortConfig)

	// 隠しファイルをフィルタリング
	if !p.showHidden {
		entries = filterHiddenFiles(entries)
	}

	// allEntriesにすべてのエントリを保存
	p.allEntries = entries
	// フィルタをクリアして全エントリを表示
	p.entries = entries
	p.filterPattern = ""
	p.filterMode = SearchModeNone
	p.cursor = 0
	p.scrollOffset = 0
	// Clear marks on directory change
	p.markedFiles = make(map[string]bool)

	return nil
}

// MoveCursorDown はカーソルを下に移動
func (p *Pane) MoveCursorDown() {
	if p.cursor < len(p.entries)-1 {
		p.cursor++
		p.adjustScroll()
	}
}

// MoveCursorUp はカーソルを上に移動
func (p *Pane) MoveCursorUp() {
	if p.cursor > 0 {
		p.cursor--
		p.adjustScroll()
	}
}

// adjustScroll はスクロール位置を調整
func (p *Pane) adjustScroll() {
	visibleLines := p.height - 4 // ヘッダー2行 + ボーダー1行 = 3行を除く

	// カーソルが表示範囲外に出たらスクロール
	if p.cursor < p.scrollOffset {
		p.scrollOffset = p.cursor
	} else if p.cursor >= p.scrollOffset+visibleLines {
		p.scrollOffset = p.cursor - visibleLines + 1
	}
}

// EnsureCursorVisible はカーソルが表示範囲内に収まるようスクロールを調整
func (p *Pane) EnsureCursorVisible() {
	p.adjustScroll()
}

// SelectedEntry は選択中のエントリを返す
func (p *Pane) SelectedEntry() *fs.FileEntry {
	if p.cursor < 0 || p.cursor >= len(p.entries) {
		return nil
	}
	return &p.entries[p.cursor]
}

// Path は現在のパスを返す
func (p *Pane) Path() string {
	return p.path
}

// SetSize はペインサイズを設定
func (p *Pane) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetActive はアクティブ状態を設定
func (p *Pane) SetActive(active bool) {
	p.isActive = active
}

// ToggleDisplayMode は表示モードをBasicとDetailの間で切り替える
// 端末幅が狭い場合（ShouldUseMinimalMode() == true）は何もしない
func (p *Pane) ToggleDisplayMode() {
	if p.ShouldUseMinimalMode() {
		// 端末幅が狭い場合は切り替えない
		return
	}

	// displayMode を切り替え（これがユーザーの選択を保存）
	if p.displayMode == DisplayBasic {
		p.displayMode = DisplayDetail
	} else {
		p.displayMode = DisplayBasic
	}
}

// ShouldUseMinimalMode は端末幅に基づいてMinimalモードを使うべきか判定
func (p *Pane) ShouldUseMinimalMode() bool {
	_, hasSpace := CalculateColumnWidths(p.width)
	return !hasSpace
}

// GetEffectiveDisplayMode は実際に使用される表示モードを返す
// 端末幅が狭い場合は自動的にMinimalモードになる
func (p *Pane) GetEffectiveDisplayMode() DisplayMode {
	if p.ShouldUseMinimalMode() {
		return DisplayMinimal
	}
	return p.displayMode
}

// CanToggleMode は現在iキーが有効かどうかを返す
func (p *Pane) CanToggleMode() bool {
	return !p.ShouldUseMinimalMode()
}

// GetSortConfig はソート設定を返す
func (p *Pane) GetSortConfig() SortConfig {
	return p.sortConfig
}

// SetSortConfig はソート設定を設定する
func (p *Pane) SetSortConfig(config SortConfig) {
	p.sortConfig = config
}

// extractSubdirName extracts the subdirectory name from the current path.
// This is used to remember which directory we came from when navigating to parent.
//
// Example:
//
//	Current path: /home/user/documents
//	Returns: "documents"
func (p *Pane) extractSubdirName() string {
	return filepath.Base(p.path)
}

// findEntryIndex finds the index of an entry by name in the current entries list.
// Returns -1 if not found.
func (p *Pane) findEntryIndex(name string) int {
	for i, entry := range p.entries {
		if entry.Name == name {
			return i
		}
	}
	return -1
}

// ApplySortAndPreserveCursor はソートを適用しながらカーソル位置を維持する
func (p *Pane) ApplySortAndPreserveCursor() {
	// 現在のカーソル位置のファイル名を記憶
	currentName := ""
	if p.cursor >= 0 && p.cursor < len(p.entries) {
		currentName = p.entries[p.cursor].Name
	}

	// allEntriesをソート
	p.allEntries = SortEntries(p.allEntries, p.sortConfig)

	// フィルタが適用されている場合は再適用
	if p.IsFiltered() {
		p.ApplyFilter(p.filterPattern, p.filterMode)
	} else {
		p.entries = p.allEntries
	}

	// カーソル位置を復元
	if currentName != "" {
		for i, e := range p.entries {
			if e.Name == currentName {
				p.cursor = i
				p.adjustScroll()
				return
			}
		}
	}

	// 見つからない場合は現在のインデックスを維持（範囲内に調整）
	if p.cursor >= len(p.entries) {
		p.cursor = max(0, len(p.entries)-1)
	}
	p.adjustScroll()
}

// SetLoading sets the loading state and progress message
func (p *Pane) SetLoading(loading bool, progress string) {
	p.loading = loading
	p.loadingProgress = progress
}

// IsLoading returns whether the pane is in loading state
func (p *Pane) IsLoading() bool {
	return p.loading
}

// SetEntries sets the entries directly (used by async loading)
func (p *Pane) SetEntries(entries []fs.FileEntry) {
	// Apply hidden file filter
	if !p.showHidden {
		entries = filterHiddenFiles(entries)
	}
	p.allEntries = entries
	p.entries = entries
	p.filterPattern = ""
	p.filterMode = SearchModeNone
}

// GetPendingCursorTarget returns the pending cursor target
func (p *Pane) GetPendingCursorTarget() string {
	return p.pendingCursorTarget
}

// ClearPendingCursorTarget clears the pending cursor target
func (p *Pane) ClearPendingCursorTarget() {
	p.pendingCursorTarget = ""
}

// SetCursor sets the cursor position
func (p *Pane) SetCursor(cursor int) {
	p.cursor = cursor
}

// GetPaneID returns the pane ID
func (p *Pane) GetPaneID() PanePosition {
	return p.paneID
}

// SetPath sets the current path (used by async navigation)
func (p *Pane) SetPath(path string) {
	p.path = path
}

// GetHistory returns the directory history
func (p *Pane) GetHistory() *DirectoryHistory {
	return &p.history
}

// restoreHistoryOnError restores history position on navigation error
func (p *Pane) restoreHistoryOnError(isForward bool) {
	if isForward {
		p.history.NavigateBack()
	} else {
		p.history.NavigateForward()
	}
}

// GetPendingPath returns the pending path during async navigation
func (p *Pane) GetPendingPath() string {
	return p.pendingPath
}

// Helper function to check if a string has a hidden file prefix
func hasHiddenPrefix(name string) bool {
	return strings.HasPrefix(name, ".")
}
