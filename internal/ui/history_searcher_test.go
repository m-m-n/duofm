package ui

import (
	"path/filepath"
	"testing"
)

func TestNewHistorySearcher(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("ls -la")
	sh.Add("pwd")
	sh.Add("echo hello")

	hs := NewHistorySearcher(sh)

	if hs == nil {
		t.Fatal("NewHistorySearcher() returned nil")
	}
}

func TestHistorySearcher_SetPattern(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("ls -la")
	sh.Add("pwd")
	sh.Add("echo hello")

	hs := NewHistorySearcher(sh)

	hs.SetPattern("ls")

	// Should have at least one match
	current := hs.Current()
	if current == "" {
		t.Error("Current() = empty, want match for 'ls'")
	}
	if current != "ls -la" {
		t.Errorf("Current() = %q, want %q", current, "ls -la")
	}
}

func TestHistorySearcher_SetPattern_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("LS -LA")
	sh.Add("pwd")

	hs := NewHistorySearcher(sh)

	hs.SetPattern("ls")

	current := hs.Current()
	if current != "LS -LA" {
		t.Errorf("Current() = %q, want %q (case insensitive)", current, "LS -LA")
	}
}

func TestHistorySearcher_SetPattern_Partial(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("echo hello world")
	sh.Add("pwd")

	hs := NewHistorySearcher(sh)

	hs.SetPattern("hello")

	current := hs.Current()
	if current != "echo hello world" {
		t.Errorf("Current() = %q, want %q (partial match)", current, "echo hello world")
	}
}

func TestHistorySearcher_SetPattern_NoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("ls -la")
	sh.Add("pwd")

	hs := NewHistorySearcher(sh)

	hs.SetPattern("xyz")

	current := hs.Current()
	if current != "" {
		t.Errorf("Current() = %q, want empty for no match", current)
	}
}

func TestHistorySearcher_Current(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("cmd1")
	sh.Add("cmd2")
	sh.Add("cmd3")

	hs := NewHistorySearcher(sh)

	// Without pattern, Current should return empty
	current := hs.Current()
	if current != "" {
		t.Errorf("Current() without pattern = %q, want empty", current)
	}

	hs.SetPattern("cmd")

	// With pattern, Current should return the first match
	current = hs.Current()
	if current != "cmd3" {
		t.Errorf("Current() = %q, want %q (most recent)", current, "cmd3")
	}
}

func TestHistorySearcher_Next(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("match1")
	sh.Add("other")
	sh.Add("match2")
	sh.Add("match3")

	hs := NewHistorySearcher(sh)
	hs.SetPattern("match")

	// First should be match3 (most recent)
	if hs.Current() != "match3" {
		t.Errorf("Current() = %q, want %q", hs.Current(), "match3")
	}

	// Next should return match2
	next := hs.Next()
	if next != "match2" {
		t.Errorf("Next() = %q, want %q", next, "match2")
	}

	// Next should return match1
	next = hs.Next()
	if next != "match1" {
		t.Errorf("Next() = %q, want %q", next, "match1")
	}

	// Next should wrap around to match3
	next = hs.Next()
	if next != "match3" {
		t.Errorf("Next() wrap = %q, want %q", next, "match3")
	}
}

func TestHistorySearcher_Next_NoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("cmd1")
	sh.Add("cmd2")

	hs := NewHistorySearcher(sh)
	hs.SetPattern("xyz")

	next := hs.Next()
	if next != "" {
		t.Errorf("Next() with no match = %q, want empty", next)
	}
}

func TestHistorySearcher_Reset(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("cmd1")
	sh.Add("cmd2")

	hs := NewHistorySearcher(sh)
	hs.SetPattern("cmd")

	// Move to next
	hs.Next()

	// Reset
	hs.Reset()

	// Pattern should be cleared
	if hs.Current() != "" {
		t.Errorf("Current() after Reset = %q, want empty", hs.Current())
	}
}

func TestHistorySearcher_EmptyHistory(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	hs := NewHistorySearcher(sh)
	hs.SetPattern("cmd")

	if hs.Current() != "" {
		t.Errorf("Current() on empty history = %q, want empty", hs.Current())
	}

	if hs.Next() != "" {
		t.Errorf("Next() on empty history = %q, want empty", hs.Next())
	}
}

func TestHistorySearcher_PatternUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("apple")
	sh.Add("banana")
	sh.Add("apricot")

	hs := NewHistorySearcher(sh)

	// Search for "a"
	hs.SetPattern("a")
	if hs.Current() != "apricot" {
		t.Errorf("Current() for 'a' = %q, want %q", hs.Current(), "apricot")
	}

	// Refine to "ap"
	hs.SetPattern("ap")
	if hs.Current() != "apricot" {
		t.Errorf("Current() for 'ap' = %q, want %q", hs.Current(), "apricot")
	}

	// Change to "ban"
	hs.SetPattern("ban")
	if hs.Current() != "banana" {
		t.Errorf("Current() for 'ban' = %q, want %q", hs.Current(), "banana")
	}
}
