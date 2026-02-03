package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestConfigWatcher_WriteTriggersMessage(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	os.WriteFile(configPath, []byte("initial"), 0644)

	var mu sync.Mutex
	var msgs []interface{}
	sender := func(msg interface{}) {
		mu.Lock()
		msgs = append(msgs, msg)
		mu.Unlock()
	}

	watcher, err := NewConfigWatcher(configPath, sender)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}
	watcher.Start()
	defer watcher.Stop()

	// Wait for watcher to be ready
	time.Sleep(100 * time.Millisecond)

	// Modify the file
	os.WriteFile(configPath, []byte("modified"), 0644)

	// Wait for debounce + processing
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	found := false
	for _, msg := range msgs {
		if _, ok := msg.(ConfigFileChangedMsg); ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected ConfigFileChangedMsg to be sent after file write")
	}
}

func TestConfigWatcher_CreateTriggersMessage(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	// File does not exist initially

	var mu sync.Mutex
	var msgs []interface{}
	sender := func(msg interface{}) {
		mu.Lock()
		msgs = append(msgs, msg)
		mu.Unlock()
	}

	watcher, err := NewConfigWatcher(configPath, sender)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}
	watcher.Start()
	defer watcher.Stop()

	time.Sleep(100 * time.Millisecond)

	// Create the file
	os.WriteFile(configPath, []byte("new content"), 0644)

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	found := false
	for _, msg := range msgs {
		if _, ok := msg.(ConfigFileChangedMsg); ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected ConfigFileChangedMsg to be sent after file creation")
	}
}

func TestConfigWatcher_SuppressForIgnoresEvents(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	os.WriteFile(configPath, []byte("initial"), 0644)

	var mu sync.Mutex
	var msgs []interface{}
	sender := func(msg interface{}) {
		mu.Lock()
		msgs = append(msgs, msg)
		mu.Unlock()
	}

	watcher, err := NewConfigWatcher(configPath, sender)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}
	watcher.Start()
	defer watcher.Stop()

	time.Sleep(100 * time.Millisecond)

	// Suppress for 1 second
	watcher.SuppressFor(1 * time.Second)

	// Modify the file during suppression
	os.WriteFile(configPath, []byte("suppressed write"), 0644)

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	msgCount := len(msgs)
	mu.Unlock()

	if msgCount > 0 {
		t.Error("Expected no messages during suppression period")
	}
}

func TestConfigWatcher_SuppressForExpiresAndResumesEvents(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	os.WriteFile(configPath, []byte("initial"), 0644)

	var mu sync.Mutex
	var msgs []interface{}
	sender := func(msg interface{}) {
		mu.Lock()
		msgs = append(msgs, msg)
		mu.Unlock()
	}

	watcher, err := NewConfigWatcher(configPath, sender)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}
	watcher.Start()
	defer watcher.Stop()

	time.Sleep(100 * time.Millisecond)

	// Suppress for a short time
	watcher.SuppressFor(200 * time.Millisecond)

	// Wait for suppression to expire
	time.Sleep(300 * time.Millisecond)

	// Modify the file after suppression
	os.WriteFile(configPath, []byte("after suppression"), 0644)

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	found := false
	for _, msg := range msgs {
		if _, ok := msg.(ConfigFileChangedMsg); ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected ConfigFileChangedMsg after suppression period expired")
	}
}

func TestConfigWatcher_StopTerminatesGoroutine(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	os.WriteFile(configPath, []byte("initial"), 0644)

	sender := func(msg interface{}) {}

	watcher, err := NewConfigWatcher(configPath, sender)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}
	watcher.Start()

	// Stop should not hang
	done := make(chan struct{})
	go func() {
		watcher.Stop()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(3 * time.Second):
		t.Fatal("Stop() did not complete within timeout")
	}
}

func TestConfigWatcher_DebounceCoalescesEvents(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	os.WriteFile(configPath, []byte("initial"), 0644)

	var mu sync.Mutex
	var msgs []interface{}
	sender := func(msg interface{}) {
		mu.Lock()
		msgs = append(msgs, msg)
		mu.Unlock()
	}

	watcher, err := NewConfigWatcher(configPath, sender)
	if err != nil {
		t.Fatalf("NewConfigWatcher failed: %v", err)
	}
	watcher.Start()
	defer watcher.Stop()

	time.Sleep(100 * time.Millisecond)

	// Write multiple times rapidly
	for i := 0; i < 5; i++ {
		os.WriteFile(configPath, []byte("rapid write"), 0644)
		time.Sleep(30 * time.Millisecond)
	}

	// Wait for debounce to settle
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	changeCount := 0
	for _, msg := range msgs {
		if _, ok := msg.(ConfigFileChangedMsg); ok {
			changeCount++
		}
	}

	// Should get at most 1-2 messages (debounce coalesces)
	if changeCount > 2 {
		t.Errorf("Expected at most 2 ConfigFileChangedMsg (debounced), got %d", changeCount)
	}
}
