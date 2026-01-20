package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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
		cmd := m.bookmarkManager.Delete(result.index)
		return m, cmd, true
	}

	// ブックマーク削除完了
	if result, ok := msg.(bookmarkDeletedMsg); ok {
		m.bookmarkManager.SetBookmarks(result.bookmarks)
		m.statusMessage = "Bookmark removed"
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true
	}

	// ブックマーク編集
	if result, ok := msg.(bookmarkEditMsg); ok {
		m.dialog = nil
		m.bookmarkManager.SetEditIndex(result.index)
		editIndex := result.index
		dialog := NewInputDialog("Edit bookmark name:", func(newAlias string) tea.Cmd {
			return m.bookmarkManager.Edit(editIndex, newAlias)
		})
		dialog.SetEmptyErrorMsg("Bookmark name cannot be empty")
		dialog.SetInput(result.bookmark.Name)
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
		m.bookmarkManager.SetBookmarks(result.bookmarks)
		m.statusMessage = fmt.Sprintf("Bookmarked: %s", result.alias)
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true
	}

	// ブックマーク編集完了
	if result, ok := msg.(bookmarkEditedMsg); ok {
		m.dialog = nil
		m.bookmarkManager.SetBookmarks(result.bookmarks)
		m.bookmarkManager.ClearEditIndex()
		m.statusMessage = fmt.Sprintf("Bookmark updated: %s", result.alias)
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true
	}

	return m, nil, false
}

// handlePathJumpMessages はPath Jump関連のメッセージを処理する
func (m Model) handlePathJumpMessages(msg tea.Msg) (Model, tea.Cmd, bool) {
	// Path Jumpジャンプ
	if result, ok := msg.(pathJumpResultMsg); ok {
		m.dialog = nil
		cmd := m.getActivePane().ChangeDirectoryAsync(result.path)
		return m, cmd, true
	}

	// Path Jumpキャンセル
	if _, ok := msg.(pathJumpCancelMsg); ok {
		m.dialog = nil
		return m, nil, true
	}

	return m, nil, false
}
