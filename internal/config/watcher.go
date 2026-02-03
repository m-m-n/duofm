package config

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// MsgSender is the callback type for sending messages.
// Uses interface{} to avoid config package depending on tea package.
type MsgSender func(msg interface{})

// ConfigFileChangedMsg notifies that the config file has been changed.
type ConfigFileChangedMsg struct{}

// ConfigWatchLostMsg notifies that the file watch has been lost.
type ConfigWatchLostMsg struct {
	Error error
}

// ConfigWatcher watches for config file changes using fsnotify.
type ConfigWatcher struct {
	configPath    string
	configDir     string
	watcher       *fsnotify.Watcher
	sendMsg       MsgSender
	done          chan struct{}
	suppressUntil time.Time
	mu            sync.Mutex
}

// NewConfigWatcher creates a new ConfigWatcher.
// It watches the parent directory (for file creation) and the file itself (if it exists).
func NewConfigWatcher(configPath string, sendMsg MsgSender) (*ConfigWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Dir(configPath)

	// Watch the parent directory for file creation
	if err := w.Add(configDir); err != nil {
		w.Close()
		return nil, err
	}

	// Also watch the file itself if it exists
	// Ignore error if file doesn't exist yet
	w.Add(configPath)

	return &ConfigWatcher{
		configPath: configPath,
		configDir:  configDir,
		watcher:    w,
		sendMsg:    sendMsg,
		done:       make(chan struct{}),
	}, nil
}

// Start begins watching for config file changes in a goroutine.
func (cw *ConfigWatcher) Start() {
	go cw.eventLoop()
}

// Stop stops the watcher and cleans up resources.
func (cw *ConfigWatcher) Stop() {
	close(cw.done)
	cw.watcher.Close()
}

// SuppressFor suppresses event processing for the given duration.
// Call this before writing to the config file to prevent self-triggered reloads.
func (cw *ConfigWatcher) SuppressFor(d time.Duration) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.suppressUntil = time.Now().Add(d)
}

// isSuppressed checks if events should currently be suppressed.
func (cw *ConfigWatcher) isSuppressed() bool {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	return time.Now().Before(cw.suppressUntil)
}

// eventLoop processes fsnotify events with debouncing.
func (cw *ConfigWatcher) eventLoop() {
	var debounceTimer *time.Timer

	for {
		select {
		case <-cw.done:
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return

		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}

			// Only handle events for the config file
			if !cw.isConfigFileEvent(event) {
				continue
			}

			// Check suppression
			if cw.isSuppressed() {
				continue
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				// Reset debounce timer
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(200*time.Millisecond, func() {
					cw.sendMsg(ConfigFileChangedMsg{})
				})

				// Re-add the file to the watcher if it was recreated (e.g., rename+create)
				if event.Has(fsnotify.Create) {
					cw.watcher.Add(cw.configPath)
				}
			}

			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				// File was removed or renamed, try to re-register watch after a delay
				go cw.retryWatch()
			}

		case _, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}
			// Errors are typically transient, ignore
		}
	}
}

// isConfigFileEvent checks if the event is for the config file.
func (cw *ConfigWatcher) isConfigFileEvent(event fsnotify.Event) bool {
	absEvent, err := filepath.Abs(event.Name)
	if err != nil {
		return event.Name == cw.configPath
	}
	absConfig, err := filepath.Abs(cw.configPath)
	if err != nil {
		return event.Name == cw.configPath
	}
	return absEvent == absConfig
}

// retryWatch attempts to re-register the file watch after a delay.
func (cw *ConfigWatcher) retryWatch() {
	time.Sleep(1 * time.Second)

	select {
	case <-cw.done:
		return
	default:
	}

	// Try to add the file back
	if err := cw.watcher.Add(cw.configPath); err != nil {
		// Retry failed - notify
		cw.sendMsg(ConfigWatchLostMsg{Error: err})
	}
}
