package ui

import (
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
)

// BatchOperation holds state for batch file operations
type BatchOperation struct {
	Files      []string // List of source file paths
	CurrentIdx int      // Current file index
	DestPath   string   // Destination directory
	Operation  string   // "copy" or "move"
	Completed  []string // Successfully completed files
	Failed     []string // Failed files
}

// BatchOperationManager manages batch copy/move operations.
// It uses a message-based approach to stay decoupled from Model/Pane.
type BatchOperationManager struct {
	current *BatchOperation
}

// NewBatchOperationManager creates a new batch operation manager.
func NewBatchOperationManager() *BatchOperationManager {
	return &BatchOperationManager{}
}

// IsActive returns true if a batch operation is in progress.
func (m *BatchOperationManager) IsActive() bool {
	return m.current != nil
}

// Current returns the current batch operation state.
func (m *BatchOperationManager) Current() *BatchOperation {
	return m.current
}

// Start initializes a new batch operation.
// Returns a command that produces batchStartedMsg.
func (m *BatchOperationManager) Start(files []string, srcDir, destDir, operation string) tea.Cmd {
	// Build full paths
	fullPaths := make([]string, len(files))
	for i, f := range files {
		fullPaths[i] = filepath.Join(srcDir, f)
	}

	m.current = &BatchOperation{
		Files:      fullPaths,
		CurrentIdx: 0,
		DestPath:   destDir,
		Operation:  operation,
		Completed:  make([]string, 0),
		Failed:     make([]string, 0),
	}

	return func() tea.Msg {
		return batchStartedMsg{}
	}
}

// CurrentFile returns the current file to process.
// Returns empty string if no more files.
func (m *BatchOperationManager) CurrentFile() string {
	if m.current == nil || m.current.CurrentIdx >= len(m.current.Files) {
		return ""
	}
	return m.current.Files[m.current.CurrentIdx]
}

// DestPath returns the destination path.
func (m *BatchOperationManager) DestPath() string {
	if m.current == nil {
		return ""
	}
	return m.current.DestPath
}

// Operation returns the current operation type ("copy" or "move").
func (m *BatchOperationManager) Operation() string {
	if m.current == nil {
		return ""
	}
	return m.current.Operation
}

// Advance moves to the next file in the batch.
// Returns a command that produces either batchNextFileMsg or batchCompleteMsg.
func (m *BatchOperationManager) Advance(success bool, srcPath string) tea.Cmd {
	if m.current == nil {
		return nil
	}

	if success {
		m.current.Completed = append(m.current.Completed, srcPath)
	} else {
		m.current.Failed = append(m.current.Failed, srcPath)
	}

	m.current.CurrentIdx++

	// Check if there are more files
	if m.current.CurrentIdx >= len(m.current.Files) {
		return m.complete()
	}

	return func() tea.Msg {
		return batchNextFileMsg{
			srcPath:  m.current.Files[m.current.CurrentIdx],
			destPath: m.current.DestPath,
		}
	}
}

// complete finishes the batch operation and returns completion message.
func (m *BatchOperationManager) complete() tea.Cmd {
	if m.current == nil {
		return nil
	}

	operation := m.current.Operation
	completed := len(m.current.Completed)
	failed := len(m.current.Failed)
	m.current = nil

	return func() tea.Msg {
		return batchCompleteMsg{
			operation: operation,
			completed: completed,
			failed:    failed,
		}
	}
}

// Cancel cancels the batch operation.
// Returns a command that produces batchCancelledMsg.
func (m *BatchOperationManager) Cancel() tea.Cmd {
	if m.current == nil {
		return nil
	}

	operation := m.current.Operation
	completed := len(m.current.Completed)
	remaining := len(m.current.Files) - m.current.CurrentIdx
	m.current = nil

	return func() tea.Msg {
		return batchCancelledMsg{
			operation: operation,
			completed: completed,
			remaining: remaining,
		}
	}
}

// ExecuteCurrentFile executes the file operation for the current file.
// Returns a command that produces the operation result.
func (m *BatchOperationManager) ExecuteCurrentFile() tea.Cmd {
	if m.current == nil {
		return nil
	}

	srcPath := m.CurrentFile()
	destPath := m.current.DestPath
	operation := m.current.Operation

	return func() tea.Msg {
		var err error
		if operation == "copy" {
			err = fs.Copy(srcPath, destPath)
		} else {
			err = fs.MoveFile(srcPath, destPath)
		}

		return batchFileResultMsg{
			srcPath: srcPath,
			success: err == nil,
			err:     err,
		}
	}
}

// --- Messages ---

// batchStartedMsg is sent when a batch operation starts.
type batchStartedMsg struct{}

// batchNextFileMsg is sent when ready to process the next file.
type batchNextFileMsg struct {
	srcPath  string
	destPath string
}

// batchCompleteMsg is sent when a batch operation completes.
type batchCompleteMsg struct {
	operation string
	completed int
	failed    int
}

// batchCancelledMsg is sent when a batch operation is cancelled.
type batchCancelledMsg struct {
	operation string
	completed int
	remaining int
}

// batchFileResultMsg is sent when a single file operation completes.
type batchFileResultMsg struct {
	srcPath string
	success bool
	err     error
}
