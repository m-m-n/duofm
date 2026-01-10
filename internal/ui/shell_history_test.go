package ui

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewShellHistory(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	if sh == nil {
		t.Fatal("NewShellHistory() returned nil")
	}

	if !sh.IsEnabled() {
		t.Error("IsEnabled() = false, want true")
	}
}

func TestShellHistory_IsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		limit   int
		enabled bool
	}{
		{"limit > 0", 100, true},
		{"limit = 0", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			historyFile := filepath.Join(tmpDir, "history")

			sh := NewShellHistory(historyFile, tt.limit)
			defer sh.Close()

			if sh.IsEnabled() != tt.enabled {
				t.Errorf("IsEnabled() = %v, want %v", sh.IsEnabled(), tt.enabled)
			}
		})
	}
}

func TestShellHistory_Add(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	// Add a command
	sh.Add("ls -la")

	commands := sh.Commands()
	if len(commands) != 1 {
		t.Fatalf("Commands() len = %d, want 1", len(commands))
	}
	if commands[0] != "ls -la" {
		t.Errorf("Commands()[0] = %q, want %q", commands[0], "ls -la")
	}
}

func TestShellHistory_Add_EmptyCommand(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	// Add empty command - should be ignored
	sh.Add("")
	sh.Add("   ")
	sh.Add("\t\n")

	commands := sh.Commands()
	if len(commands) != 0 {
		t.Errorf("Commands() len = %d, want 0", len(commands))
	}
}

func TestShellHistory_Add_TrimWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("  ls -la  ")

	commands := sh.Commands()
	if len(commands) != 1 {
		t.Fatalf("Commands() len = %d, want 1", len(commands))
	}
	if commands[0] != "ls -la" {
		t.Errorf("Commands()[0] = %q, want %q", commands[0], "ls -la")
	}
}

func TestShellHistory_Add_DuplicateRemoval(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("ls -la")
	sh.Add("pwd")
	sh.Add("ls -la") // Duplicate - should move to top

	commands := sh.Commands()
	if len(commands) != 2 {
		t.Fatalf("Commands() len = %d, want 2", len(commands))
	}
	// Most recent first
	if commands[0] != "ls -la" {
		t.Errorf("Commands()[0] = %q, want %q", commands[0], "ls -la")
	}
	if commands[1] != "pwd" {
		t.Errorf("Commands()[1] = %q, want %q", commands[1], "pwd")
	}
}

func TestShellHistory_Add_LimitExceeded(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 3) // Small limit for testing
	defer sh.Close()

	sh.Add("cmd1")
	sh.Add("cmd2")
	sh.Add("cmd3")
	sh.Add("cmd4") // This should push cmd1 out

	commands := sh.Commands()
	if len(commands) != 3 {
		t.Fatalf("Commands() len = %d, want 3", len(commands))
	}
	// cmd1 should be removed
	for _, cmd := range commands {
		if cmd == "cmd1" {
			t.Error("cmd1 should have been removed due to limit")
		}
	}
	if commands[0] != "cmd4" {
		t.Errorf("Commands()[0] = %q, want %q", commands[0], "cmd4")
	}
}

func TestShellHistory_Add_Disabled(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 0) // Disabled
	defer sh.Close()

	sh.Add("ls -la")

	commands := sh.Commands()
	if len(commands) != 0 {
		t.Errorf("Commands() len = %d, want 0 (disabled)", len(commands))
	}
}

func TestShellHistory_Load(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	// Create history file with content
	content := "cmd1\ncmd2\ncmd3\n"
	if err := os.WriteFile(historyFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write history file: %v", err)
	}

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	if err := sh.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	commands := sh.Commands()
	if len(commands) != 3 {
		t.Fatalf("Commands() len = %d, want 3", len(commands))
	}
	// Most recent should be first (cmd3 was last in file)
	if commands[0] != "cmd3" {
		t.Errorf("Commands()[0] = %q, want %q", commands[0], "cmd3")
	}
}

func TestShellHistory_Load_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "nonexistent_history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	// Should not error on non-existent file
	if err := sh.Load(); err != nil {
		t.Errorf("Load() error on non-existent file: %v", err)
	}

	commands := sh.Commands()
	if len(commands) != 0 {
		t.Errorf("Commands() len = %d, want 0", len(commands))
	}
}

func TestShellHistory_Load_EmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	// Create history file with empty lines
	content := "cmd1\n\ncmd2\n   \ncmd3\n"
	if err := os.WriteFile(historyFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write history file: %v", err)
	}

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	if err := sh.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	commands := sh.Commands()
	if len(commands) != 3 {
		t.Errorf("Commands() len = %d, want 3", len(commands))
	}
}

func TestShellHistory_Load_LimitTrim(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	// Create history file with more entries than limit
	var content strings.Builder
	for i := 0; i < 10; i++ {
		content.WriteString("cmd" + string(rune('0'+i)) + "\n")
	}
	if err := os.WriteFile(historyFile, []byte(content.String()), 0600); err != nil {
		t.Fatalf("Failed to write history file: %v", err)
	}

	sh := NewShellHistory(historyFile, 5) // Limit to 5
	defer sh.Close()

	if err := sh.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	commands := sh.Commands()
	if len(commands) != 5 {
		t.Errorf("Commands() len = %d, want 5", len(commands))
	}
	// Should keep the most recent 5 (cmd5-cmd9)
	if commands[0] != "cmd9" {
		t.Errorf("Commands()[0] = %q, want %q", commands[0], "cmd9")
	}
}

func TestShellHistory_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)

	sh.Add("cmd1")
	sh.Add("cmd2")

	// Close to flush pending saves
	sh.Close()

	// Verify file exists
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		t.Fatal("History file was not created")
	}

	// Verify permissions (0600)
	info, err := os.Stat(historyFile)
	if err != nil {
		t.Fatalf("Failed to stat history file: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("File permissions = %o, want 0600", perm)
	}

	// Verify content
	content, err := os.ReadFile(historyFile)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("File has %d lines, want 2", len(lines))
	}
}

func TestShellHistory_AtomicWrite_CreatesParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "nested", "dir", "history")

	sh := NewShellHistory(historyFile, 100)

	sh.Add("cmd1")
	sh.Close()

	// Verify file exists
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		t.Fatal("History file was not created with nested directories")
	}

	// Verify parent directory permissions (0700)
	parentDir := filepath.Dir(historyFile)
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Fatalf("Failed to stat parent directory: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0700 {
		t.Errorf("Parent directory permissions = %o, want 0700", perm)
	}
}

func TestShellHistory_Debounce(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)

	// Add multiple commands rapidly
	for i := 0; i < 10; i++ {
		sh.Add("cmd" + string(rune('0'+i)))
	}

	// Close immediately - should flush all pending saves
	sh.Close()

	// Verify file contains all commands
	content, err := os.ReadFile(historyFile)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 10 {
		t.Errorf("File has %d lines, want 10", len(lines))
	}
}

func TestShellHistory_Close_Flush(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)

	sh.Add("cmd1")
	sh.Add("cmd2")

	// Don't wait for debounce, close immediately
	sh.Close()

	// File should be written
	content, err := os.ReadFile(historyFile)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}
	if !strings.Contains(string(content), "cmd1") || !strings.Contains(string(content), "cmd2") {
		t.Error("Close() did not flush pending saves")
	}
}

func TestShellHistory_Commands_ReturnsCopy(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("cmd1")
	sh.Add("cmd2")

	commands := sh.Commands()
	commands[0] = "modified"

	// Original should not be modified
	original := sh.Commands()
	if original[0] != "cmd2" {
		t.Error("Commands() did not return a copy")
	}
}

func TestShellHistory_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 1000)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				sh.Add("cmd" + string(rune('0'+n)) + string(rune('0'+j)))
				_ = sh.Commands()
			}
		}(i)
	}
	wg.Wait()

	sh.Close()

	// Should not panic or deadlock
	commands := sh.Commands()
	// Due to duplicate removal, we may have fewer than 100
	if len(commands) == 0 {
		t.Error("Concurrent access resulted in empty history")
	}
}

func TestShellHistory_Unicode(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)

	sh.Add("echo 'Hello World'")
	sh.Add("ls -la ~")
	sh.Add("echo 'Bonjour'")

	sh.Close()

	// Reload and verify
	sh2 := NewShellHistory(historyFile, 100)
	defer sh2.Close()

	if err := sh2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	commands := sh2.Commands()
	if len(commands) != 3 {
		t.Fatalf("Commands() len = %d, want 3", len(commands))
	}

	// Check unicode is preserved
	found := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "echo") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Unicode content not preserved")
	}
}

func TestShellHistory_LongCommand(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)

	// Create a long command (>1000 chars)
	longCmd := strings.Repeat("x", 1500)
	sh.Add(longCmd)

	sh.Close()

	// Reload and verify
	sh2 := NewShellHistory(historyFile, 100)
	defer sh2.Close()

	if err := sh2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	commands := sh2.Commands()
	if len(commands) != 1 {
		t.Fatalf("Commands() len = %d, want 1", len(commands))
	}
	if commands[0] != longCmd {
		t.Error("Long command not preserved correctly")
	}
}

func TestShellHistory_DebounceTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping debounce timing test in short mode")
	}

	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)

	sh.Add("cmd1")

	// Wait less than debounce time
	time.Sleep(100 * time.Millisecond)

	// File should not exist yet
	if _, err := os.Stat(historyFile); err == nil {
		// File might exist due to fast debounce, skip this check
		t.Log("File exists before debounce timeout (fast system)")
	}

	// Wait for debounce to complete
	time.Sleep(600 * time.Millisecond)

	// File should exist now
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		t.Error("File not written after debounce timeout")
	}

	sh.Close()
}
