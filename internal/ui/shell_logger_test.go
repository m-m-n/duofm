package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewShellLogger_DefaultDir(t *testing.T) {
	sl := NewShellLogger("/tmp")
	defer sl.Close()

	expected := filepath.Join("/tmp", fmt.Sprintf("duofm-shell-%d.log", os.Getpid()))
	if sl.LogPath() != expected {
		t.Errorf("LogPath() = %q, want %q", sl.LogPath(), expected)
	}
}

func TestNewShellLogger_EmptyDir(t *testing.T) {
	sl := NewShellLogger("")
	defer sl.Close()

	expected := filepath.Join("/tmp", fmt.Sprintf("duofm-shell-%d.log", os.Getpid()))
	if sl.LogPath() != expected {
		t.Errorf("LogPath() = %q, want %q", sl.LogPath(), expected)
	}
}

func TestNewShellLogger_CustomDir(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)
	defer sl.Close()

	expected := filepath.Join(tmpDir, fmt.Sprintf("duofm-shell-%d.log", os.Getpid()))
	if sl.LogPath() != expected {
		t.Errorf("LogPath() = %q, want %q", sl.LogPath(), expected)
	}
}

func TestNewShellLogger_NonExistentDir(t *testing.T) {
	// Use a path under /proc that cannot be created
	sl := NewShellLogger("/proc/nonexistent/deeply/nested/dir")
	defer sl.Close()

	expected := filepath.Join("/tmp", fmt.Sprintf("duofm-shell-%d.log", os.Getpid()))
	if sl.LogPath() != expected {
		t.Errorf("LogPath() = %q, want %q (should fall back to /tmp)", sl.LogPath(), expected)
	}
}

func TestNewShellLogger_MkdirAllCreatesDir(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "a", "b", "c")

	sl := NewShellLogger(nestedDir)
	defer sl.Close()

	expected := filepath.Join(nestedDir, fmt.Sprintf("duofm-shell-%d.log", os.Getpid()))
	if sl.LogPath() != expected {
		t.Errorf("LogPath() = %q, want %q", sl.LogPath(), expected)
	}

	// Verify directory was created
	info, err := os.Stat(nestedDir)
	if err != nil {
		t.Fatalf("MkdirAll did not create directory: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected %q to be a directory", nestedDir)
	}
}

func TestShellLogger_AppendHeader_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)
	defer sl.Close()

	logPath := sl.LogPath()

	// File should not exist before AppendHeader
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("log file should not exist before AppendHeader")
	}

	err := sl.AppendHeader("ls -la", "/home/user")
	if err != nil {
		t.Fatalf("AppendHeader() error = %v", err)
	}

	// File should exist after AppendHeader
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("log file should exist after AppendHeader")
	}
}

func TestShellLogger_AppendHeader_Format(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)
	defer sl.Close()

	err := sl.AppendHeader("echo hello", "/home/user/docs")
	if err != nil {
		t.Fatalf("AppendHeader() error = %v", err)
	}

	data, err := os.ReadFile(sl.LogPath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)
	separator := "════════════════════════════════════════════════════════════════"

	// Verify header contains separator lines
	if strings.Count(content, separator) != 2 {
		t.Errorf("header should contain exactly 2 separator lines, got %d", strings.Count(content, separator))
	}

	// Verify header contains the command
	if !strings.Contains(content, "$ echo hello") {
		t.Errorf("header should contain command, got:\n%s", content)
	}

	// Verify header contains the directory
	if !strings.Contains(content, "Directory: /home/user/docs") {
		t.Errorf("header should contain directory, got:\n%s", content)
	}

	// Verify timestamp format [YYYY-MM-DD HH:MM:SS]
	if !strings.Contains(content, "[") || !strings.Contains(content, "]") {
		t.Errorf("header should contain timestamp in brackets, got:\n%s", content)
	}
}

func TestShellLogger_AppendFooter(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)
	defer sl.Close()

	// Write header first to create the file
	err := sl.AppendHeader("ls", "/tmp")
	if err != nil {
		t.Fatalf("AppendHeader() error = %v", err)
	}

	beforeData, err := os.ReadFile(sl.LogPath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	err = sl.AppendFooter()
	if err != nil {
		t.Fatalf("AppendFooter() error = %v", err)
	}

	afterData, err := os.ReadFile(sl.LogPath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Footer should add a blank line
	added := string(afterData[len(beforeData):])
	if added != "\n" {
		t.Errorf("AppendFooter() should write a blank line, got %q", added)
	}
}

func TestShellLogger_AppendFooter_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)
	defer sl.Close()

	// AppendFooter without any prior write should not error
	err := sl.AppendFooter()
	if err != nil {
		t.Errorf("AppendFooter() on uninitialized logger should not error, got %v", err)
	}
}

func TestShellLogger_Multiple_HeaderFooter(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)
	defer sl.Close()

	// Write multiple header/footer pairs
	commands := []struct {
		cmd string
		dir string
	}{
		{"ls -la", "/home/user"},
		{"cat file.txt", "/home/user/docs"},
		{"git status", "/home/user/project"},
	}

	for _, c := range commands {
		if err := sl.AppendHeader(c.cmd, c.dir); err != nil {
			t.Fatalf("AppendHeader(%q) error = %v", c.cmd, err)
		}
		if err := sl.AppendFooter(); err != nil {
			t.Fatalf("AppendFooter() error = %v", err)
		}
	}

	data, err := os.ReadFile(sl.LogPath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)

	// Each header has 2 separators, so 3 commands = 6 separators
	separator := "════════════════════════════════════════════════════════════════"
	if strings.Count(content, separator) != 6 {
		t.Errorf("expected 6 separator lines for 3 commands, got %d", strings.Count(content, separator))
	}

	// Verify all commands are present
	for _, c := range commands {
		if !strings.Contains(content, "$ "+c.cmd) {
			t.Errorf("log should contain command %q", c.cmd)
		}
		if !strings.Contains(content, "Directory: "+c.dir) {
			t.Errorf("log should contain directory %q", c.dir)
		}
	}
}

func TestShellLogger_HasLog_BeforeWrite(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)
	defer sl.Close()

	if sl.HasLog() {
		t.Errorf("HasLog() = true before any writes, want false")
	}
}

func TestShellLogger_HasLog_AfterWrite(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)
	defer sl.Close()

	err := sl.AppendHeader("ls", "/tmp")
	if err != nil {
		t.Fatalf("AppendHeader() error = %v", err)
	}

	if !sl.HasLog() {
		t.Errorf("HasLog() = false after AppendHeader, want true")
	}
}

func TestShellLogger_Close_DeletesFile(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)

	// Write something to create the file
	err := sl.AppendHeader("ls", "/tmp")
	if err != nil {
		t.Fatalf("AppendHeader() error = %v", err)
	}

	logPath := sl.LogPath()

	// Verify file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Fatalf("log file should exist before Close")
	}

	err = sl.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// File should be deleted after Close
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("log file should not exist after Close")
	}
}

func TestShellLogger_Close_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)

	// Close without any writes should not error
	err := sl.Close()
	if err != nil {
		t.Errorf("Close() on logger with no writes should not error, got %v", err)
	}
}

func TestShellLogger_Close_SetsHasLogFalse(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)

	err := sl.AppendHeader("ls", "/tmp")
	if err != nil {
		t.Fatalf("AppendHeader() error = %v", err)
	}

	if !sl.HasLog() {
		t.Fatalf("HasLog() should be true after write")
	}

	sl.Close()

	if sl.HasLog() {
		t.Errorf("HasLog() = true after Close, want false")
	}
}

func TestShellLogger_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)
	defer sl.Close()

	err := sl.AppendHeader("ls", "/tmp")
	if err != nil {
		t.Fatalf("AppendHeader() error = %v", err)
	}

	info, err := os.Stat(sl.LogPath())
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	// File should have 0600 permissions
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

func TestShellLogger_Unicode(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)
	defer sl.Close()

	unicodeCmd := "echo 'こんにちは世界 🌍'"
	unicodeDir := "/home/ユーザー/ドキュメント"

	err := sl.AppendHeader(unicodeCmd, unicodeDir)
	if err != nil {
		t.Fatalf("AppendHeader() error = %v", err)
	}

	data, err := os.ReadFile(sl.LogPath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)

	if !strings.Contains(content, unicodeCmd) {
		t.Errorf("log should preserve unicode command, got:\n%s", content)
	}

	if !strings.Contains(content, unicodeDir) {
		t.Errorf("log should preserve unicode directory, got:\n%s", content)
	}
}

func TestShellLogger_LogPath_Immutable(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)
	defer sl.Close()

	path1 := sl.LogPath()
	path2 := sl.LogPath()

	if path1 != path2 {
		t.Errorf("LogPath() should return consistent value: %q != %q", path1, path2)
	}
}

func TestShellLogger_LogPath_ContainsPID(t *testing.T) {
	tmpDir := t.TempDir()
	sl := NewShellLogger(tmpDir)
	defer sl.Close()

	expectedSuffix := fmt.Sprintf("duofm-shell-%d.log", os.Getpid())
	if !strings.HasSuffix(sl.LogPath(), expectedSuffix) {
		t.Errorf("LogPath() = %q, should end with %q", sl.LogPath(), expectedSuffix)
	}
}
