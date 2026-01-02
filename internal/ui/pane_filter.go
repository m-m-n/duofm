package ui

import (
	"fmt"
	"strings"

	"github.com/sakura/duofm/internal/fs"
)

// filterHiddenFiles は隠しファイル（.で始まるファイル）を除外する
// ただし親ディレクトリ（..）は常に表示する
func filterHiddenFiles(entries []fs.FileEntry) []fs.FileEntry {
	result := make([]fs.FileEntry, 0, len(entries))
	for _, e := range entries {
		// 親ディレクトリは常に表示
		if e.IsParentDir() || !strings.HasPrefix(e.Name, ".") {
			result = append(result, e)
		}
	}
	return result
}

// ToggleHidden は隠しファイルの表示/非表示を切り替える
// カーソル位置は可能な限り維持する
func (p *Pane) ToggleHidden() {
	// 現在選択中のファイル名を記憶
	var selectedName string
	if p.cursor >= 0 && p.cursor < len(p.entries) {
		selectedName = p.entries[p.cursor].Name
	}

	// 非表示に切り替える場合、隠しファイルのマークをクリア
	if p.showHidden {
		for filename := range p.markedFiles {
			if strings.HasPrefix(filename, ".") {
				delete(p.markedFiles, filename)
			}
		}
	}

	p.showHidden = !p.showHidden
	p.LoadDirectory()

	// Note: LoadDirectory() clears all marks, so we need to preserve them
	// This is handled by saving marks before and restoring after if needed
	// For now, marks are cleared on directory reload which is acceptable

	// カーソル位置の復元を試みる
	if selectedName != "" {
		for i, e := range p.entries {
			if e.Name == selectedName {
				p.cursor = i
				p.adjustScroll()
				return
			}
		}
	}
	// 見つからない場合（隠しファイルだった場合）は先頭にリセット
	p.cursor = 0
	p.scrollOffset = 0
}

// IsShowingHidden は隠しファイルが表示中かどうかを返す
func (p *Pane) IsShowingHidden() bool {
	return p.showHidden
}

// ApplyFilter はフィルタパターンを適用してエントリをフィルタリングする
func (p *Pane) ApplyFilter(pattern string, mode SearchMode) error {
	p.filterPattern = pattern
	p.filterMode = mode

	if pattern == "" {
		// パターンが空の場合はフィルタをクリア
		p.entries = p.allEntries
		p.cursor = 0
		p.scrollOffset = 0
		return nil
	}

	var filtered []fs.FileEntry
	var err error

	switch mode {
	case SearchModeIncremental:
		filtered = filterIncremental(p.allEntries, pattern)
	case SearchModeRegex:
		filtered, err = filterRegex(p.allEntries, pattern)
		if err != nil {
			return err
		}
	default:
		filtered = p.allEntries
	}

	p.entries = filtered

	// カーソル位置を調整
	if p.cursor >= len(p.entries) {
		if len(p.entries) > 0 {
			p.cursor = len(p.entries) - 1
		} else {
			p.cursor = 0
		}
	}
	p.scrollOffset = 0
	p.adjustScroll()

	return nil
}

// ClearFilter はフィルタをクリアしてすべてのエントリを表示する
func (p *Pane) ClearFilter() {
	p.filterPattern = ""
	p.filterMode = SearchModeNone
	p.entries = p.allEntries

	// カーソル位置を調整
	if p.cursor >= len(p.entries) {
		if len(p.entries) > 0 {
			p.cursor = len(p.entries) - 1
		} else {
			p.cursor = 0
		}
	}
	p.adjustScroll()
}

// ResetToFullList はディレクトリを再読み込みしてフィルタをクリアする
func (p *Pane) ResetToFullList() error {
	return p.LoadDirectory()
}

// IsFiltered はフィルタが適用中かどうかを返す
func (p *Pane) IsFiltered() bool {
	return p.filterPattern != ""
}

// FilterPattern は現在のフィルタパターンを返す
func (p *Pane) FilterPattern() string {
	return p.filterPattern
}

// FilterMode は現在のフィルタモードを返す
func (p *Pane) FilterMode() SearchMode {
	return p.filterMode
}

// TotalEntryCount はフィルタ前のエントリ数を返す（親ディレクトリを除く）
func (p *Pane) TotalEntryCount() int {
	count := len(p.allEntries)
	if count > 0 && p.allEntries[0].IsParentDir() {
		count--
	}
	return count
}

// FilteredEntryCount はフィルタ後のエントリ数を返す（親ディレクトリを除く）
func (p *Pane) FilteredEntryCount() int {
	count := len(p.entries)
	if count > 0 && p.entries[0].IsParentDir() {
		count--
	}
	return count
}

// formatFilterIndicator はフィルタインジケーターをフォーマットする
// 例: [/pattern] または [re/pattern]
func (p *Pane) formatFilterIndicator() string {
	if !p.IsFiltered() {
		return ""
	}

	pattern := p.filterPattern
	// パターンが長い場合は切り詰める
	maxLen := 15
	if len(pattern) > maxLen {
		pattern = pattern[:maxLen-2] + ".."
	}

	switch p.filterMode {
	case SearchModeIncremental:
		return fmt.Sprintf("[/%s]", pattern)
	case SearchModeRegex:
		return fmt.Sprintf("[re/%s]", pattern)
	default:
		return ""
	}
}

// RefreshDirectoryPreserveCursor reloads directory contents while preserving cursor position.
// If the previously selected file no longer exists, cursor resets to the beginning.
func (p *Pane) RefreshDirectoryPreserveCursor() error {
	// Store current selected file name
	var selectedName string
	if entry := p.SelectedEntry(); entry != nil {
		selectedName = entry.Name
	}

	// Reload directory entries
	entries, err := fs.ReadDirectory(p.path)
	if err != nil {
		return err
	}

	entries = SortEntries(entries, p.sortConfig)

	// Filter hidden files
	if !p.showHidden {
		entries = filterHiddenFiles(entries)
	}

	p.allEntries = entries
	p.entries = entries
	p.filterPattern = ""
	p.filterMode = SearchModeNone

	// Find the previously selected file in new entries
	newCursor := 0 // Default to beginning if file not found
	if selectedName != "" {
		for i, e := range entries {
			if e.Name == selectedName {
				newCursor = i
				break
			}
		}
	}

	p.cursor = newCursor
	p.adjustScroll()

	// Clear marks on refresh (same as LoadDirectory)
	p.markedFiles = make(map[string]bool)

	return nil
}
