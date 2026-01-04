package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/archive"
)

// ArchiveOperationState holds state for in-progress archive operations
type ArchiveOperationState struct {
	Sources     []string              // Source files/directories to archive
	DestDir     string                // Destination directory
	Format      archive.ArchiveFormat // Selected archive format
	Level       int                   // Compression level (0-9)
	ArchiveName string                // Archive filename
	TaskID      string                // Task ID for background operation
}

// ArchiveOperationManager manages archive compress/extract operations.
// It uses a message-based approach to stay decoupled from Model/Dialog.
type ArchiveOperationManager struct {
	state      *ArchiveOperationState
	controller *archive.ArchiveController
}

// NewArchiveOperationManager creates a new archive operation manager.
func NewArchiveOperationManager() *ArchiveOperationManager {
	return &ArchiveOperationManager{
		controller: archive.NewArchiveController(),
	}
}

// IsActive returns true if an archive operation is in progress.
func (m *ArchiveOperationManager) IsActive() bool {
	return m.state != nil
}

// State returns the current operation state.
func (m *ArchiveOperationManager) State() *ArchiveOperationState {
	return m.state
}

// TaskID returns the current task ID.
func (m *ArchiveOperationManager) TaskID() string {
	if m.state == nil {
		return ""
	}
	return m.state.TaskID
}

// SetTaskID sets the task ID for the current operation.
func (m *ArchiveOperationManager) SetTaskID(taskID string) {
	if m.state != nil {
		m.state.TaskID = taskID
	}
}

// PrepareCompression prepares state for a compression operation.
// Returns a message to show the progress dialog.
func (m *ArchiveOperationManager) PrepareCompression(sources []string, destDir string, format archive.ArchiveFormat, level int, archiveName string) tea.Cmd {
	m.state = &ArchiveOperationState{
		Sources:     sources,
		DestDir:     destDir,
		Format:      format,
		Level:       level,
		ArchiveName: archiveName,
	}

	return func() tea.Msg {
		return showArchiveProgressMsg{
			operation:   "compress",
			archivePath: archiveName,
		}
	}
}

// StartCompression starts the compression task.
// Should be called after the progress dialog is shown.
func (m *ArchiveOperationManager) StartCompression(archivePath string) tea.Cmd {
	if m.state == nil || m.controller == nil {
		return nil
	}

	sources := m.state.Sources
	format := m.state.Format
	level := m.state.Level
	controller := m.controller

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

// PrepareExtraction prepares state for an extraction operation.
// Returns a message to show the progress dialog.
func (m *ArchiveOperationManager) PrepareExtraction(archivePath, destDir string) tea.Cmd {
	m.state = &ArchiveOperationState{
		Sources: []string{archivePath},
		DestDir: destDir,
	}

	return func() tea.Msg {
		return showArchiveProgressMsg{
			operation:   "extract",
			archivePath: archivePath,
		}
	}
}

// StartExtraction starts the extraction task.
// Should be called after the progress dialog is shown.
func (m *ArchiveOperationManager) StartExtraction(archivePath, destDir string) tea.Cmd {
	if m.controller == nil {
		return nil
	}

	controller := m.controller

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

// CheckSecurity performs security checks before extraction.
func (m *ArchiveOperationManager) CheckSecurity(archivePath, destDir string) tea.Cmd {
	if m.controller == nil {
		return nil
	}

	controller := m.controller

	return func() tea.Msg {
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
		availableSize := int64(archive.GetAvailableDiskSpace(destDir))
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

// PollProgress polls for operation progress.
func (m *ArchiveOperationManager) PollProgress(taskID string) tea.Cmd {
	if m.controller == nil {
		return nil
	}

	controller := m.controller

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

// CancelTask cancels the current task.
func (m *ArchiveOperationManager) CancelTask() {
	if m.state != nil && m.state.TaskID != "" && m.controller != nil {
		m.controller.CancelTask(m.state.TaskID)
	}
}

// Clear clears the current operation state.
func (m *ArchiveOperationManager) Clear() {
	m.state = nil
}

// --- Messages ---

// showArchiveProgressMsg requests Model to show a progress dialog.
type showArchiveProgressMsg struct {
	operation   string // "compress" or "extract"
	archivePath string
}

// archiveOperationStartMsg is sent when archive operation starts.
type archiveOperationStartMsg struct {
	taskID string
}

// archiveOperationErrorMsg is sent when an error occurs.
type archiveOperationErrorMsg struct {
	err     error
	message string
}

// archiveProgressUpdateMsg is sent for progress updates.
type archiveProgressUpdateMsg struct {
	taskID          string
	progress        float64
	processedFiles  int
	totalFiles      int
	currentFile     string
	elapsedTime     time.Duration
	estimatedRemain time.Duration
}

// archiveOperationCompleteMsg is sent when operation completes.
type archiveOperationCompleteMsg struct {
	taskID      string
	success     bool
	cancelled   bool
	archivePath string
	err         error
}

// extractSecurityCheckMsg is sent with security check results.
type extractSecurityCheckMsg struct {
	archivePath   string
	destDir       string
	archiveSize   int64
	extractedSize int64
	availableSize int64
	compressionOK bool
	diskSpaceOK   bool
	ratio         float64
	err           error
}
