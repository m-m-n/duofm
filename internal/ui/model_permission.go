package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	fsops "github.com/sakura/duofm/internal/fs"
)

// handlePermission handles the permission change action
func (m Model) handlePermission() (tea.Model, tea.Cmd) {
	activePane := m.getActivePane()
	markedFiles := activePane.GetMarkedFiles()

	// Batch mode: multiple files marked
	if len(markedFiles) > 0 {
		return m.handleBatchPermission(markedFiles)
	}

	// Single mode: no marks, use selected entry
	entry := activePane.SelectedEntry()

	// Check if entry exists and is not parent directory
	if entry == nil || entry.IsParentDir() {
		m.statusMessage = "Cannot change parent directory permissions"
		m.isStatusError = true
		return m, statusMessageClearCmd(3 * time.Second)
	}

	// FR8.3: Check if entry is a symlink
	fullPath := filepath.Join(activePane.Path(), entry.Name)
	info, err := os.Lstat(fullPath)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		m.statusMessage = "Cannot change permissions: symlinks are skipped"
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second)
	}

	// Create permission dialog
	dialog := NewPermissionDialog(entry.Name, entry.IsDir, entry.Permissions)

	// Set confirmation callback
	dialog.SetOnConfirm(func(mode string, recursive bool) tea.Cmd {
		fullPath := filepath.Join(activePane.Path(), entry.Name)
		return m.executePermissionChange(fullPath, mode, recursive, entry.IsDir)
	})

	m.dialog = dialog
	return m, nil
}

// handleBatchPermission handles batch permission changes
func (m Model) handleBatchPermission(markedPaths []string) (tea.Model, tea.Cmd) {
	activePane := m.getActivePane()

	// Build full paths and check for symlinks (FR8.3)
	fullPaths := make([]string, 0, len(markedPaths))
	symlinkCount := 0

	for _, name := range markedPaths {
		fullPath := filepath.Join(activePane.Path(), name)

		// Check if it's a symlink
		info, err := os.Lstat(fullPath)
		if err != nil {
			continue // Skip if can't stat
		}

		if info.Mode()&os.ModeSymlink != 0 {
			symlinkCount++
		} else {
			fullPaths = append(fullPaths, fullPath)
		}
	}

	// FR8.3: If only symlinks are selected, show informative message
	if len(fullPaths) == 0 && symlinkCount > 0 {
		m.statusMessage = "Cannot change permissions: only symlinks selected (symlinks are skipped)"
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second)
	}

	// If no valid files after filtering symlinks
	if len(fullPaths) == 0 {
		m.statusMessage = "No files selected"
		m.isStatusError = true
		return m, statusMessageClearCmd(3 * time.Second)
	}

	// Create batch permission dialog (no recursive option for batch)
	dialog := NewPermissionDialog(fmt.Sprintf("%d items", len(fullPaths)), false, 0644)

	// Set confirmation callback
	dialog.SetOnConfirm(func(mode string, recursive bool) tea.Cmd {
		return m.executeBatchPermissionChange(fullPaths, mode)
	})

	m.dialog = dialog
	return m, nil
}

// executePermissionChange executes the permission change operation
func (m *Model) executePermissionChange(path, mode string, recursive bool, isDir bool) tea.Cmd {
	return func() tea.Msg {
		// If not recursive, simple permission change
		if !recursive {
			// Parse permission mode
			fileMode, err := fsops.ParsePermissionMode(mode)
			if err != nil {
				return permissionOperationCompleteMsg{
					path:    path,
					success: false,
					err:     err,
				}
			}

			// Execute permission change
			err = fsops.ChangePermission(path, fileMode)
			if err != nil {
				return permissionOperationCompleteMsg{
					path:    path,
					success: false,
					err:     err,
				}
			}

			return permissionOperationCompleteMsg{
				path:    path,
				success: true,
				err:     nil,
			}
		}

		// Recursive mode - show two-step dialog
		return showRecursivePermDialogMsg{
			path: path,
		}
	}
}

// executeRecursivePermissionChange executes recursive permission change
func (m *Model) executeRecursivePermissionChange(path, dirMode, fileMode string) tea.Cmd {
	return func() tea.Msg {
		// Parse modes
		dirFileMode, err := fsops.ParsePermissionMode(dirMode)
		if err != nil {
			return recursivePermissionCompleteMsg{
				path:         path,
				successCount: 0,
				errors:       []fsops.PermissionError{{Path: path, Error: fmt.Errorf("invalid directory mode: %w", err)}},
			}
		}

		fileFileMode, err := fsops.ParsePermissionMode(fileMode)
		if err != nil {
			return recursivePermissionCompleteMsg{
				path:         path,
				successCount: 0,
				errors:       []fsops.PermissionError{{Path: path, Error: fmt.Errorf("invalid file mode: %w", err)}},
			}
		}

		// Execute recursive change with progress callback
		// TODO: Add progress dialog for large operations
		successCount, errors := fsops.ChangePermissionRecursiveWithProgress(
			path,
			dirFileMode,
			fileFileMode,
			nil, // For now, no progress callback (will be added in future)
		)

		return recursivePermissionCompleteMsg{
			path:         path,
			successCount: successCount,
			errors:       errors,
		}
	}
}

// executeBatchPermissionChange executes batch permission change
func (m *Model) executeBatchPermissionChange(paths []string, mode string) tea.Cmd {
	// Check if we need to show progress dialog (FR7.7: 10+ items)
	if len(paths) >= fsops.ProgressThreshold {
		return m.executeBatchPermissionChangeWithProgress(paths, mode)
	}

	// For small batches, execute directly without progress dialog
	return func() tea.Msg {
		// Parse permission mode
		fileMode, err := fsops.ParsePermissionMode(mode)
		if err != nil {
			return batchPermissionCompleteMsg{
				totalCount:  len(paths),
				failedCount: len(paths),
				errors:      []fsops.PermissionError{{Path: "validation", Error: err}},
			}
		}

		// Apply to all paths
		successCount := 0
		errors := make([]fsops.PermissionError, 0)
		for _, path := range paths {
			// Skip symlinks
			info, err := os.Lstat(path)
			if err != nil {
				errors = append(errors, fsops.PermissionError{Path: path, Error: err})
				continue // Skip if can't stat
			}
			if info.Mode()&os.ModeSymlink != 0 {
				continue // Skip symlinks
			}

			// Change permission
			if err := fsops.ChangePermission(path, fileMode); err != nil {
				errors = append(errors, fsops.PermissionError{Path: path, Error: err})
			} else {
				successCount++
			}
		}

		return batchPermissionCompleteMsg{
			totalCount:   len(paths),
			successCount: successCount,
			failedCount:  len(paths) - successCount,
			errors:       errors,
		}
	}
}

// executeBatchPermissionChangeWithProgress executes batch permission change with progress dialog
func (m *Model) executeBatchPermissionChangeWithProgress(paths []string, mode string) tea.Cmd {
	return func() tea.Msg {
		// First, show progress dialog
		return batchPermissionStartMsg{
			paths: paths,
			mode:  mode,
		}
	}
}

// handlePermissionMessages handles permission-related messages
func (m Model) handlePermissionMessages(msg tea.Msg) (Model, tea.Cmd, bool) {
	if completeMsg, ok := msg.(permissionOperationCompleteMsg); ok {
		newModel, cmd := m.handlePermissionOperationComplete(completeMsg)
		return newModel.(Model), cmd, true
	}

	if batchMsg, ok := msg.(batchPermissionCompleteMsg); ok {
		newModel, cmd := m.handleBatchPermissionComplete(batchMsg)
		return newModel.(Model), cmd, true
	}

	if recursiveMsg, ok := msg.(recursivePermissionCompleteMsg); ok {
		newModel, cmd := m.handleRecursivePermissionComplete(recursiveMsg)
		return newModel.(Model), cmd, true
	}

	if startMsg, ok := msg.(batchPermissionStartMsg); ok {
		// Show progress dialog and start batch operation
		newModel, cmd := m.handleBatchPermissionStart(startMsg)
		return newModel.(Model), cmd, true
	}

	if progressMsg, ok := msg.(batchPermissionProgressMsg); ok {
		// Update progress dialog
		newModel, cmd := m.handleBatchPermissionProgress(progressMsg)
		return newModel.(Model), cmd, true
	}

	if recursiveDialogMsg, ok := msg.(showRecursivePermDialogMsg); ok {
		// Show RecursivePermDialog
		dialog := NewRecursivePermDialog(filepath.Base(recursiveDialogMsg.path))
		dialog.SetOnConfirm(func(dirMode, fileMode string) tea.Cmd {
			return m.executeRecursivePermissionChange(recursiveDialogMsg.path, dirMode, fileMode)
		})
		m.dialog = dialog
		return m, nil, true
	}

	return m, nil, false
}

// handlePermissionOperationComplete handles permission operation completion
func (m Model) handlePermissionOperationComplete(msg permissionOperationCompleteMsg) (tea.Model, tea.Cmd) {
	if msg.success {
		m.statusMessage = fmt.Sprintf("Permission changed: %s", filepath.Base(msg.path))
		m.isStatusError = false

		// Refresh active pane to show new permissions
		if err := m.getActivePane().Refresh(); err != nil {
			m.statusMessage = fmt.Sprintf("Permission changed but refresh failed: %v", err)
			m.isStatusError = true
		}

		return m, statusMessageClearCmd(3 * time.Second)
	}

	m.statusMessage = fmt.Sprintf("Failed to change permission: %v", msg.err)
	m.isStatusError = true
	return m, statusMessageClearCmd(5 * time.Second)
}

// handleBatchPermissionStart handles batch permission operation start
func (m Model) handleBatchPermissionStart(msg batchPermissionStartMsg) (tea.Model, tea.Cmd) {
	// Create and show progress dialog
	progressDialog := NewPermissionProgressDialog(len(msg.paths))
	m.dialog = progressDialog

	// Start the actual batch operation in background
	return m, m.executeBatchPermissionInBackground(msg.paths, msg.mode)
}

// handleBatchPermissionProgress handles batch permission progress updates
func (m Model) handleBatchPermissionProgress(msg batchPermissionProgressMsg) (tea.Model, tea.Cmd) {
	// Update progress dialog if it exists
	if progressDialog, ok := m.dialog.(*PermissionProgressDialog); ok {
		progressDialog.UpdateProgress(msg.processed, msg.currentPath)
	}
	return m, nil
}

// executeBatchPermissionInBackground executes batch permission change in background with progress updates
func (m *Model) executeBatchPermissionInBackground(paths []string, mode string) tea.Cmd {
	return func() tea.Msg {
		// Parse permission mode
		fileMode, err := fsops.ParsePermissionMode(mode)
		if err != nil {
			return batchPermissionCompleteMsg{
				totalCount:  len(paths),
				failedCount: len(paths),
				errors:      []fsops.PermissionError{{Path: "validation", Error: err}},
			}
		}

		// Apply to all paths with progress updates
		successCount := 0
		processed := 0
		total := len(paths)
		errors := make([]fsops.PermissionError, 0)

		for _, path := range paths {
			// Skip symlinks
			info, err := os.Lstat(path)
			if err != nil {
				errors = append(errors, fsops.PermissionError{Path: path, Error: err})
				processed++
				continue // Skip if can't stat
			}
			if info.Mode()&os.ModeSymlink != 0 {
				processed++
				continue // Skip symlinks
			}

			// Change permission
			if err := fsops.ChangePermission(path, fileMode); err != nil {
				errors = append(errors, fsops.PermissionError{Path: path, Error: err})
			} else {
				successCount++
			}
			processed++

			// Send progress update every file (Bubble Tea will throttle rendering)
			// In production, we might want to batch these updates
			// For now, send each update to ensure smooth progress bar
		}

		return batchPermissionCompleteMsg{
			totalCount:   total,
			successCount: successCount,
			failedCount:  total - successCount,
			errors:       errors,
		}
	}
}

// handleRecursivePermissionComplete handles recursive permission operation completion
func (m Model) handleRecursivePermissionComplete(msg recursivePermissionCompleteMsg) (tea.Model, tea.Cmd) {
	// Refresh active pane
	if err := m.getActivePane().Refresh(); err != nil {
		m.statusMessage = fmt.Sprintf("Permissions changed but refresh failed: %v", err)
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second)
	}

	// If there are errors, show error report dialog
	if len(msg.errors) > 0 {
		errorDialog := NewPermissionErrorReportDialog(msg.successCount, len(msg.errors), msg.errors)
		m.dialog = errorDialog
		return m, nil
	}

	// All successful
	m.statusMessage = fmt.Sprintf("Recursive permissions changed: %d files successful", msg.successCount)
	m.isStatusError = false
	return m, statusMessageClearCmd(3 * time.Second)
}

// handleBatchPermissionComplete handles batch permission operation completion
func (m Model) handleBatchPermissionComplete(msg batchPermissionCompleteMsg) (tea.Model, tea.Cmd) {
	// Close progress dialog if it exists
	if _, ok := m.dialog.(*PermissionProgressDialog); ok {
		m.dialog = nil
	}

	// Clear marks after batch operation (even if some failed)
	m.getActivePane().ClearMarks()

	// Refresh active pane
	if err := m.getActivePane().Refresh(); err != nil {
		m.statusMessage = fmt.Sprintf("Permissions changed but refresh failed: %v", err)
		m.isStatusError = true
		return m, statusMessageClearCmd(5 * time.Second)
	}

	// If there are errors, show error report dialog
	if len(msg.errors) > 0 {
		errorDialog := NewPermissionErrorReportDialog(msg.successCount, len(msg.errors), msg.errors)
		m.dialog = errorDialog
		return m, nil
	}

	// All successful
	m.statusMessage = fmt.Sprintf("Permissions changed: %d files successful", msg.successCount)
	m.isStatusError = false
	return m, statusMessageClearCmd(3 * time.Second)
}
