package ui

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ShellHistory manages shell command history with persistence.
// It provides debounced async saving and atomic file writes.
type ShellHistory struct {
	mu        sync.RWMutex
	commands  []string
	limit     int
	filePath  string
	saveQueue chan struct{}
	done      chan struct{}
	wg        sync.WaitGroup
	dirty     bool
}

// NewShellHistory creates a new ShellHistory instance.
// If limit is 0, history is disabled.
func NewShellHistory(filePath string, limit int) *ShellHistory {
	sh := &ShellHistory{
		commands:  make([]string, 0),
		limit:     limit,
		filePath:  filePath,
		saveQueue: make(chan struct{}, 1),
		done:      make(chan struct{}),
		dirty:     false,
	}

	if limit > 0 {
		sh.wg.Add(1)
		go sh.saveLoop()
	}

	return sh
}

// IsEnabled returns true if history is enabled (limit > 0).
func (sh *ShellHistory) IsEnabled() bool {
	return sh.limit > 0
}

// Add adds a command to the history.
// Empty commands (or whitespace-only) are ignored.
// Duplicate commands are moved to the front.
func (sh *ShellHistory) Add(command string) {
	if !sh.IsEnabled() {
		return
	}

	// Trim whitespace
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}

	sh.mu.Lock()
	defer sh.mu.Unlock()

	// Remove duplicate if exists
	for i, cmd := range sh.commands {
		if cmd == command {
			// Remove from current position
			sh.commands = append(sh.commands[:i], sh.commands[i+1:]...)
			break
		}
	}

	// Add to front
	sh.commands = append([]string{command}, sh.commands...)

	// Trim to limit
	if len(sh.commands) > sh.limit {
		sh.commands = sh.commands[:sh.limit]
	}

	sh.dirty = true

	// Trigger async save (non-blocking)
	select {
	case sh.saveQueue <- struct{}{}:
	default:
		// Save already queued
	}
}

// Commands returns a copy of the command history.
// Most recent commands are first.
func (sh *ShellHistory) Commands() []string {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	result := make([]string, len(sh.commands))
	copy(result, sh.commands)
	return result
}

// Load loads the command history from the file.
// If the file doesn't exist, an empty history is used.
// If limit is exceeded, older entries are trimmed.
func (sh *ShellHistory) Load() error {
	if !sh.IsEnabled() {
		return nil
	}

	sh.mu.Lock()
	defer sh.mu.Unlock()

	file, err := os.Open(sh.filePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			commands = append(commands, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Reverse so most recent is first (file stores oldest first)
	for i, j := 0, len(commands)-1; i < j; i, j = i+1, j-1 {
		commands[i], commands[j] = commands[j], commands[i]
	}

	// Trim to limit
	if len(commands) > sh.limit {
		commands = commands[:sh.limit]
	}

	sh.commands = commands
	return nil
}

// Close flushes any pending saves and stops the background goroutine.
func (sh *ShellHistory) Close() {
	if !sh.IsEnabled() {
		return
	}

	// Signal goroutine to stop
	close(sh.done)

	// Wait for goroutine to finish
	sh.wg.Wait()

	// Final flush if dirty
	sh.mu.Lock()
	dirty := sh.dirty
	sh.mu.Unlock()

	if dirty {
		sh.atomicWrite()
	}
}

// saveLoop is the background goroutine that handles debounced saves.
func (sh *ShellHistory) saveLoop() {
	defer sh.wg.Done()

	debounceTimer := time.NewTimer(0)
	if !debounceTimer.Stop() {
		<-debounceTimer.C
	}

	for {
		select {
		case <-sh.done:
			debounceTimer.Stop()
			return
		case <-sh.saveQueue:
			// Reset debounce timer
			debounceTimer.Reset(500 * time.Millisecond)
		case <-debounceTimer.C:
			sh.atomicWrite()
		}
	}
}

// atomicWrite writes the history to a temporary file and renames it atomically.
func (sh *ShellHistory) atomicWrite() {
	sh.mu.Lock()
	commands := make([]string, len(sh.commands))
	copy(commands, sh.commands)
	sh.dirty = false
	sh.mu.Unlock()

	if len(commands) == 0 {
		return
	}

	// Ensure parent directory exists with 0700 permissions
	dir := filepath.Dir(sh.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		// Log error but don't fail
		return
	}

	// Write to temporary file
	tmpFile := sh.filePath + ".tmp"
	file, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return
	}

	// Reverse commands so oldest is first in file (newest at end)
	var writeErr error
	for i := len(commands) - 1; i >= 0; i-- {
		if _, err := file.WriteString(commands[i] + "\n"); err != nil {
			writeErr = err
			break
		}
	}

	// Close the file and check for errors
	if err := file.Close(); err != nil {
		// Remove temp file on close error
		os.Remove(tmpFile)
		return
	}

	// If write had an error, remove temp file and return
	if writeErr != nil {
		os.Remove(tmpFile)
		return
	}

	// Atomic rename (best effort - errors are logged silently)
	if err := os.Rename(tmpFile, sh.filePath); err != nil {
		// Clean up temp file if rename fails
		os.Remove(tmpFile)
	}
}
