package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
)

// autoRefreshMsg triggers periodic auto-refresh (directory listings + disk space).
type autoRefreshMsg struct{}

// autoRefreshTickCmd returns a command that sends autoRefreshMsg after the specified interval.
func autoRefreshTickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return autoRefreshMsg{}
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

// openWithXDGMsg notifies to open file with xdg-open
type openWithXDGMsg struct {
	file    string // file path
	workDir string // working directory
}

// openWithFinishedMsg notifies the completion of open operation
type openWithFinishedMsg struct {
	err error // error if any
}

// openWithDialogResultMsg notifies the result of open with dialog
type openWithDialogResultMsg struct {
	application string   // application name with options
	files       []string // files to open
	workDir     string   // working directory
	cancelled   bool     // true if cancelled
}

// clipboardResultMsg notifies the result of a clipboard write operation.
// err is nil on success; non-nil on failure (overrides optimistic status message).
type clipboardResultMsg struct {
	err error
}

// regexSearchResultMsg notifies the result of regex search dialog
type regexSearchResultMsg struct {
	pattern   string // search pattern (empty string means clear filter)
	cancelled bool   // true if cancelled
}

// querySearchResultMsg notifies the result of query search dialog
type querySearchResultMsg struct {
	query     string // search query (empty string means clear filter)
	cancelled bool   // true if cancelled
}

// bgOutputMsg delivers a line of output from the background command
type bgOutputMsg struct {
	line string
}

// bgCommandDoneMsg signals that the background command has finished
type bgCommandDoneMsg struct {
	err     error
	command string
	workDir string
}

// bgAutoCloseMsg signals that the 2-second post-completion timer has fired
type bgAutoCloseMsg struct{}
