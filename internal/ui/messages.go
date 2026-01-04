package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
)

// diskSpaceUpdateMsg notifies periodic disk space updates
type diskSpaceUpdateMsg struct{}

// diskSpaceTickCmd returns a command that sends diskSpaceUpdateMsg after 5 seconds
func diskSpaceTickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return diskSpaceUpdateMsg{}
	})
}

// clearStatusMsg is a message to clear status messages
type clearStatusMsg struct{}

// statusMessageClearCmd returns a command that sends clearStatusMsg after the specified duration
func statusMessageClearCmd(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// directoryLoadStartMsg notifies the start of directory loading
type directoryLoadStartMsg struct {
	panePath string
}

// directoryLoadCompleteMsg notifies the completion of directory loading
type directoryLoadCompleteMsg struct {
	paneID                   PanePosition // which pane is being loaded
	panePath                 string
	entries                  []fs.FileEntry
	err                      error
	attemptedPath            string // path to display in error message
	isHistoryNavigation      bool   // whether via history navigation (history navigation itself is not recorded)
	historyNavigationForward bool   // true=forward, false=backward (for restoration on history navigation error)
}

// directoryLoadProgressMsg notifies loading progress (optional)
type directoryLoadProgressMsg struct {
	panePath  string
	fileCount int
}

// ctrlCTimeoutMsg notifies timeout for Ctrl+C quit confirmation
type ctrlCTimeoutMsg struct{}

// ctrlCTimeoutCmd returns a command that sends ctrlCTimeoutMsg after the specified duration
func ctrlCTimeoutCmd(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return ctrlCTimeoutMsg{}
	})
}

// inputDialogResultMsg notifies the result of an input dialog
type inputDialogResultMsg struct {
	operation string // "create_file", "create_dir", "rename"
	input     string // the entered name
	oldName   string // the original name for rename operations
	cancelled bool   // true if cancelled
	err       error  // error if any
}

// archiveOperationStartMsg notifies the start of an archive operation
type archiveOperationStartMsg struct {
	taskID string // task ID
}

// archiveProgressUpdateMsg notifies progress updates for an archive operation
type archiveProgressUpdateMsg struct {
	taskID          string        // task ID
	progress        float64       // progress rate (0.0-1.0)
	processedFiles  int           // number of processed files
	totalFiles      int           // total number of files
	currentFile     string        // currently processing file
	elapsedTime     time.Duration // elapsed time
	estimatedRemain time.Duration // estimated remaining time
}

// archiveOperationCompleteMsg notifies the completion of an archive operation
type archiveOperationCompleteMsg struct {
	taskID      string // task ID
	success     bool   // whether successful
	cancelled   bool   // whether cancelled
	archivePath string // path of created archive (for compress/extract)
	err         error  // error (on failure)
}

// archiveOperationErrorMsg notifies an error in an archive operation
type archiveOperationErrorMsg struct {
	taskID  string // task ID
	err     error  // error
	message string // user-facing error message
}

// compressionLevelResultMsg notifies the result of compression level selection
type compressionLevelResultMsg struct {
	level     int  // selected compression level (0-9)
	cancelled bool // whether cancelled
}

// archiveNameResultMsg notifies the result of archive name input
type archiveNameResultMsg struct {
	name      string // entered archive name
	cancelled bool   // whether cancelled
}

// extractSecurityCheckMsg notifies security check result before archive extraction
type extractSecurityCheckMsg struct {
	archivePath   string  // archive path
	destDir       string  // destination directory
	archiveSize   int64   // archive size
	extractedSize int64   // extracted size
	availableSize int64   // available space at destination
	compressionOK bool    // whether compression ratio is normal
	diskSpaceOK   bool    // whether disk space is sufficient
	ratio         float64 // compression ratio (extracted size / archive size)
	err           error   // error (on metadata retrieval failure)
}

// permissionOperationStartMsg notifies the start of a permission change operation
type permissionOperationStartMsg struct {
	path      string // target path
	mode      string // new permission
	recursive bool   // whether recursive change
}

// permissionOperationCompleteMsg notifies the completion of a permission change operation
type permissionOperationCompleteMsg struct {
	path    string // target path
	success bool   // whether successful
	err     error  // error (on failure)
}

// showRecursivePermDialogMsg notifies to show RecursivePermDialog
type showRecursivePermDialogMsg struct {
	path string // target path
}

// batchPermissionStartMsg notifies the start of batch permission change
type batchPermissionStartMsg struct {
	paths []string
	mode  string
}

// batchPermissionCompleteMsg notifies the completion of batch permission change
type batchPermissionCompleteMsg struct {
	totalCount   int
	successCount int
	failedCount  int
	errors       []fs.PermissionError
}

// batchPermissionProgressMsg notifies the progress of batch permission change
type batchPermissionProgressMsg struct {
	processed   int
	total       int
	currentPath string
}

// recursivePermissionCompleteMsg notifies the completion of recursive permission change
type recursivePermissionCompleteMsg struct {
	path         string
	successCount int
	errors       []fs.PermissionError
}
