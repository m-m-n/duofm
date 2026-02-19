package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ShellLogger manages a session log file at <dir>/duofm-shell-<PID>.log.
// It records shell command execution headers and footers for session review.
// The log file is lazily created on the first AppendHeader call and
// deleted on Close (normal program exit).
//
// ShellLogger is not safe for concurrent use. All methods must be called
// from the Bubble Tea Update loop (single goroutine).
type ShellLogger struct {
	logPath string
	logFile *os.File
}

// NewShellLogger creates a new ShellLogger with the given directory.
// The log file is not created until the first AppendHeader call.
// logDir defaults to "/tmp" if empty.
// If logDir does not exist, attempts os.MkdirAll; on failure falls back to "/tmp".
func NewShellLogger(logDir string) *ShellLogger {
	if logDir == "" {
		logDir = "/tmp"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logDir = "/tmp"
	}
	pid := os.Getpid()
	logPath := filepath.Join(logDir, fmt.Sprintf("duofm-shell-%d.log", pid))
	return &ShellLogger{logPath: logPath}
}

// AppendHeader writes the command header to the log file (before execution).
// Creates the file with 0600 permissions on first call.
//
// Format:
//
//	════════════════════════════════════════════════════════════════
//	[2024-01-15 14:30:05] $ command
//	Directory: /path/to/dir
//	════════════════════════════════════════════════════════════════
func (sl *ShellLogger) AppendHeader(command, workDir string) error {
	if sl.logFile == nil {
		f, err := os.OpenFile(sl.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return err
		}
		sl.logFile = f
	}
	separator := "════════════════════════════════════════════════════════════════"
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	header := fmt.Sprintf("%s\n[%s] $ %s\nDirectory: %s\n%s\n",
		separator, timestamp, command, workDir, separator)
	_, err := sl.logFile.WriteString(header)
	return err
}

// AppendFooter writes a blank line separator after command output.
// If the log file has not been created yet, this is a no-op.
func (sl *ShellLogger) AppendFooter() error {
	if sl.logFile == nil {
		return nil
	}
	_, err := sl.logFile.WriteString("\n")
	return err
}

// LogPath returns the log file path for command output capture.
func (sl *ShellLogger) LogPath() string {
	return sl.logPath
}

// HasLog returns true if log file has been created (first AppendHeader call).
func (sl *ShellLogger) HasLog() bool {
	return sl.logFile != nil
}

// Close closes the file handle and deletes the log file.
// Called on normal program exit.
func (sl *ShellLogger) Close() error {
	var firstErr error
	if sl.logFile != nil {
		firstErr = sl.logFile.Close()
		sl.logFile = nil
	}
	// Delete log file, ignore error if file doesn't exist
	if err := os.Remove(sl.logPath); err != nil && !os.IsNotExist(err) && firstErr == nil {
		firstErr = err
	}
	return firstErr
}
