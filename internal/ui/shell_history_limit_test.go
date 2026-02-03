package ui

import (
	"path/filepath"
	"testing"
)

func TestShellHistory_SetLimit(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.SetLimit(50)

	if sh.limit != 50 {
		t.Errorf("Expected limit=50, got %d", sh.limit)
	}
}

func TestShellHistory_SetLimit_TruncatesEntries(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	// Add more entries than new limit
	for i := 0; i < 10; i++ {
		sh.Add("cmd" + string(rune('a'+i)))
	}

	commands := sh.Commands()
	if len(commands) != 10 {
		t.Fatalf("Expected 10 commands, got %d", len(commands))
	}

	// Set limit to 5
	sh.SetLimit(5)

	commands = sh.Commands()
	if len(commands) != 5 {
		t.Errorf("Expected 5 commands after SetLimit(5), got %d", len(commands))
	}
}

func TestShellHistory_SetLimit_IncreaseKeepsEntries(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 5)
	defer sh.Close()

	for i := 0; i < 5; i++ {
		sh.Add("cmd" + string(rune('a'+i)))
	}

	// Increase limit
	sh.SetLimit(100)

	commands := sh.Commands()
	if len(commands) != 5 {
		t.Errorf("Expected 5 commands after increasing limit, got %d", len(commands))
	}
}
