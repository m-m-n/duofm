package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
)

// handleTrash shows confirmation dialog before moving files to trash
func (m Model) handleTrash() (tea.Model, tea.Cmd) {
	activePane := m.getActivePane()
	markedFiles := activePane.GetMarkedFiles()

	var paths []string

	if len(markedFiles) > 0 {
		// Multiple files are marked
		paths = markedFiles
	} else {
		// Single file operation
		entry := activePane.SelectedEntry()
		if entry == nil || entry.IsParentDir() {
			return m, nil
		}
		paths = []string{filepath.Join(activePane.Path(), entry.Name)}
	}

	// Show confirmation dialog
	m.dialog = NewMoveToTrashDialog(paths)
	return m, nil
}

// handleTrashConfirmed executes trash after user confirmation
func (m Model) handleTrashConfirmed(paths []string) (Model, tea.Cmd, bool) {
	if len(paths) == 0 {
		return m, nil, true
	}

	// Execute trash for each file
	cmds := make([]tea.Cmd, 0, len(paths))
	for _, path := range paths {
		cmds = append(cmds, m.executeTrash(path))
	}

	return m, tea.Batch(cmds...), true
}

// executeTrash executes the trash operation for a single file
func (m *Model) executeTrash(path string) tea.Cmd {
	return func() tea.Msg {
		err := fs.MoveToTrash(path)
		if err != nil {
			return trashErrorMsg{path: path, err: err}
		}
		return trashSuccessMsg{path: path}
	}
}

// trashSuccessMsg is sent when a file is successfully trashed
type trashSuccessMsg struct {
	path string
}

// trashErrorMsg is sent when a trash operation fails
type trashErrorMsg struct {
	path string
	err  error
}

// restoreSuccessMsg is sent when a file is successfully restored
type restoreSuccessMsg struct {
	trashName    string
	originalPath string
}

// restoreErrorMsg is sent when a restore operation fails
type restoreErrorMsg struct {
	trashName string
	err       error
}

// emptyTrashSuccessMsg is sent when trash is successfully emptied
type emptyTrashSuccessMsg struct{}

// emptyTrashErrorMsg is sent when emptying trash fails
type emptyTrashErrorMsg struct {
	err error
}

// handleOpenTrashDialog opens the trash dialog
func (m Model) handleOpenTrashDialog() (tea.Model, tea.Cmd) {
	// Ensure trash directories exist
	if err := fs.EnsureTrashDirs(); err != nil {
		m.statusMessage = fmt.Sprintf("Cannot access trash: %v", err)
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second)
	}

	// Load trash items
	items, err := loadTrashItems()
	if err != nil {
		m.statusMessage = fmt.Sprintf("Cannot load trash: %v", err)
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second)
	}

	// Open trash dialog
	m.dialog = NewTrashDialog(items)
	return m, nil
}

// validateTrashName validates that a trash name is safe (no path traversal)
func validateTrashName(trashName string) error {
	// Prevent path traversal attacks
	if strings.Contains(trashName, "/") || strings.Contains(trashName, "\\") {
		return fmt.Errorf("invalid trash name: contains path separator")
	}
	if trashName == ".." || trashName == "." {
		return fmt.Errorf("invalid trash name: special directory")
	}
	if trashName == "" {
		return fmt.Errorf("invalid trash name: empty")
	}
	return nil
}

// executeRestore performs the actual restore operation
func (m *Model) executeRestore(trashName string) tea.Cmd {
	return func() tea.Msg {
		// Validate trash name to prevent path traversal
		if err := validateTrashName(trashName); err != nil {
			return restoreErrorMsg{trashName: trashName, err: err}
		}

		info, err := fs.GetTrashItemInfo(trashName)
		if err != nil {
			return restoreErrorMsg{trashName: trashName, err: err}
		}

		err = fs.RestoreFromTrash(trashName)
		if err != nil {
			return restoreErrorMsg{trashName: trashName, err: err}
		}
		return restoreSuccessMsg{trashName: trashName, originalPath: info.OriginalPath}
	}
}

// executeRestoreWithOverwrite performs restore by first deleting the existing file
func (m *Model) executeRestoreWithOverwrite(trashName, originalPath string) tea.Cmd {
	return func() tea.Msg {
		// Validate trash name to prevent path traversal
		if err := validateTrashName(trashName); err != nil {
			return restoreErrorMsg{trashName: trashName, err: err}
		}

		// Delete existing file
		if err := fs.Delete(originalPath); err != nil {
			return restoreErrorMsg{trashName: trashName, err: fmt.Errorf("failed to delete existing file: %w", err)}
		}

		// Restore from trash
		if err := fs.RestoreFromTrash(trashName); err != nil {
			return restoreErrorMsg{trashName: trashName, err: err}
		}
		return restoreSuccessMsg{trashName: trashName, originalPath: originalPath}
	}
}

// executeRestoreWithRename performs restore by renaming the restored file
func (m *Model) executeRestoreWithRename(trashName, originalPath string) tea.Cmd {
	return func() tea.Msg {
		// Validate trash name to prevent path traversal
		if err := validateTrashName(trashName); err != nil {
			return restoreErrorMsg{trashName: trashName, err: err}
		}

		info, err := fs.GetTrashItemInfo(trashName)
		if err != nil {
			return restoreErrorMsg{trashName: trashName, err: err}
		}

		// Find a unique name at the destination
		dir := filepath.Dir(info.OriginalPath)
		baseName := filepath.Base(info.OriginalPath)
		newName := fs.ResolveNameCollision(dir, baseName)
		newPath := filepath.Join(dir, newName)

		// Get source path in trash
		trashFilePath := filepath.Join(fs.TrashFilesDir(), trashName)

		// Move to new path
		if err := os.Rename(trashFilePath, newPath); err != nil {
			// Try copy+delete for cross-filesystem
			if copyErr := fs.Copy(trashFilePath, newPath); copyErr != nil {
				return restoreErrorMsg{trashName: trashName, err: copyErr}
			}
			if deleteErr := fs.Delete(trashFilePath); deleteErr != nil {
				return restoreErrorMsg{trashName: trashName, err: deleteErr}
			}
		}

		// Remove trashinfo - ignore error as restore was successful
		trashinfoPath := filepath.Join(fs.TrashInfoDir(), trashName+".trashinfo")
		_ = os.Remove(trashinfoPath)

		return restoreSuccessMsg{trashName: trashName, originalPath: newPath}
	}
}

// executeEmptyTrash performs the actual empty trash operation
func (m *Model) executeEmptyTrash() tea.Cmd {
	return func() tea.Msg {
		if err := fs.EmptyTrash(); err != nil {
			return emptyTrashErrorMsg{err: err}
		}
		return emptyTrashSuccessMsg{}
	}
}

// handleTrashDialogRestore handles restore request from trash dialog
func (m Model) handleTrashDialogRestore(items []TrashItem) (Model, tea.Cmd, bool) {
	if len(items) == 0 {
		return m, nil, true
	}

	// For single item, check for conflict and restore
	if len(items) == 1 {
		item := items[0]
		trashName := item.Name

		// Get original path from trashinfo
		info, err := fs.GetTrashItemInfo(trashName)
		if err != nil {
			m.statusMessage = fmt.Sprintf("Cannot restore: %v", err)
			m.isStatusError = true
			return m, statusMessageClearCmd(5 * time.Second), true
		}

		// Check if destination exists
		if _, err := os.Stat(info.OriginalPath); err == nil {
			// File exists at destination - show conflict dialog
			// Keep trash dialog open in background
			m.dialog = NewRestoreConflictDialog(trashName, info.OriginalPath)
			return m, nil, true
		}

		// No conflict - restore directly
		m.dialog = nil
		return m, m.executeRestore(trashName), true
	}

	// For multiple items, restore all (skip conflicts)
	m.dialog = nil
	cmds := make([]tea.Cmd, 0, len(items))
	for _, item := range items {
		trashName := item.Name

		// Get original path from trashinfo
		info, err := fs.GetTrashItemInfo(trashName)
		if err != nil {
			continue
		}

		// Check if destination exists - skip if conflict
		if _, err := os.Stat(info.OriginalPath); err == nil {
			continue
		}

		cmds = append(cmds, m.executeRestore(trashName))
	}

	if len(cmds) == 0 {
		m.statusMessage = "No items restored (all have conflicts)"
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second), true
	}

	return m, tea.Batch(cmds...), true
}

// handleTrashDialogEmpty handles empty trash request from trash dialog
func (m Model) handleTrashDialogEmpty() (Model, tea.Cmd, bool) {
	// Count items in trash
	filesDir := fs.TrashFilesDir()
	entries, err := os.ReadDir(filesDir)
	if err != nil {
		m.statusMessage = fmt.Sprintf("Cannot read trash: %v", err)
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second), true
	}

	if len(entries) == 0 {
		m.statusMessage = "Trash is already empty"
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true
	}

	// Show confirmation dialog (trash dialog stays in background)
	m.dialog = NewEmptyTrashDialog(len(entries))
	return m, nil, true
}

// handleTrashMessages handles trash-related messages
func (m Model) handleTrashMessages(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case trashConfirmResultMsg:
		m.dialog = nil
		if msg.confirmed && len(msg.paths) > 0 {
			return m.handleTrashConfirmed(msg.paths)
		}
		return m, nil, true

	case trashDialogCloseMsg:
		m.dialog = nil
		return m, nil, true

	case trashDialogRestoreMsg:
		// Restore items from trash dialog
		return m.handleTrashDialogRestore(msg.items)

	case trashDialogEmptyMsg:
		// Show confirmation dialog for emptying trash
		return m.handleTrashDialogEmpty()

	case trashSuccessMsg:
		// Refresh both panes
		m.getActivePane().LoadDirectory()
		m.getInactivePane().LoadDirectory()
		// Clear marks
		m.getActivePane().ClearMarks()

		m.statusMessage = fmt.Sprintf("Moved to trash: %s", filepath.Base(msg.path))
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true

	case trashErrorMsg:
		m.statusMessage = fmt.Sprintf("Failed to trash %s: %v", filepath.Base(msg.path), msg.err)
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second), true

	case restoreSuccessMsg:
		// Refresh both panes
		m.getActivePane().LoadDirectory()
		m.getInactivePane().LoadDirectory()

		m.statusMessage = fmt.Sprintf("Restored: %s", filepath.Base(msg.originalPath))
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true

	case restoreErrorMsg:
		m.statusMessage = fmt.Sprintf("Failed to restore %s: %v", msg.trashName, msg.err)
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second), true

	case restoreConflictResultMsg:
		m.dialog = nil
		switch msg.choice {
		case RestoreChoiceOverwrite:
			return m, m.executeRestoreWithOverwrite(msg.trashName, msg.originalPath), true
		case RestoreChoiceRename:
			return m, m.executeRestoreWithRename(msg.trashName, msg.originalPath), true
		case RestoreChoiceSkip, RestoreChoiceCancelled:
			return m, nil, true
		}
		return m, nil, true

	case emptyTrashResultMsg:
		m.dialog = nil
		if msg.confirmed {
			return m, m.executeEmptyTrash(), true
		}
		return m, nil, true

	case emptyTrashSuccessMsg:
		// Refresh both panes
		m.getActivePane().LoadDirectory()
		m.getInactivePane().LoadDirectory()

		m.statusMessage = "Trash emptied"
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true

	case emptyTrashErrorMsg:
		m.statusMessage = fmt.Sprintf("Failed to empty trash: %v", msg.err)
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second), true
	}

	return m, nil, false
}
