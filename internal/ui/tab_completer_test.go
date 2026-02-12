package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTabComplete_EmptyInput(t *testing.T) {
	tc := NewTabCompleter()
	cwd := t.TempDir()

	input, cursor := tc.Complete("", 0, cwd)
	if input != "" {
		t.Errorf("input = %q, want %q", input, "")
	}
	if cursor != 0 {
		t.Errorf("cursor = %d, want %d", cursor, 0)
	}
}

func TestTabComplete_SingleCommand(t *testing.T) {
	// Create a temp dir with a fake executable to use as PATH
	binDir := t.TempDir()
	createExecutable(t, binDir, "myuniquetestcmd123")

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir)
	defer os.Setenv("PATH", origPath)

	tc := NewTabCompleter()
	cwd := t.TempDir()

	input, cursor := tc.Complete("myuniquetestcmd1", 16, cwd)
	// Single match should auto-complete with trailing space
	if input != "myuniquetestcmd123 " {
		t.Errorf("input = %q, want %q", input, "myuniquetestcmd123 ")
	}
	if cursor != 19 {
		t.Errorf("cursor = %d, want %d", cursor, 19)
	}
}

func TestTabComplete_MultipleCommands(t *testing.T) {
	binDir := t.TempDir()
	createExecutable(t, binDir, "testcmd_alpha")
	createExecutable(t, binDir, "testcmd_beta")

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir)
	defer os.Setenv("PATH", origPath)

	tc := NewTabCompleter()
	cwd := t.TempDir()

	input, cursor := tc.Complete("testcmd_", 8, cwd)
	// Multiple matches should return common prefix only
	if input != "testcmd_" {
		t.Errorf("input = %q, want %q (common prefix, no progress)", input, "testcmd_")
	}
	// No progress made, should return original cursor
	if cursor != 8 {
		t.Errorf("cursor = %d, want %d", cursor, 8)
	}
}

func TestTabComplete_MultipleCommandsWithProgress(t *testing.T) {
	binDir := t.TempDir()
	createExecutable(t, binDir, "testxyz_alpha")
	createExecutable(t, binDir, "testxyz_beta")

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir)
	defer os.Setenv("PATH", origPath)

	tc := NewTabCompleter()
	cwd := t.TempDir()

	input, cursor := tc.Complete("testx", 5, cwd)
	// Common prefix is "testxyz_"
	if input != "testxyz_" {
		t.Errorf("input = %q, want %q", input, "testxyz_")
	}
	if cursor != 8 {
		t.Errorf("cursor = %d, want %d", cursor, 8)
	}
}

func TestTabComplete_SingleFile(t *testing.T) {
	cwd := t.TempDir()
	os.WriteFile(filepath.Join(cwd, "uniquefile.txt"), []byte(""), 0644)

	tc := NewTabCompleter()

	input, cursor := tc.Complete("ls uniquef", 10, cwd)
	if input != "ls uniquefile.txt" {
		t.Errorf("input = %q, want %q", input, "ls uniquefile.txt")
	}
	if cursor != 17 {
		t.Errorf("cursor = %d, want %d", cursor, 17)
	}
}

func TestTabComplete_MultipleFiles(t *testing.T) {
	cwd := t.TempDir()
	os.WriteFile(filepath.Join(cwd, "report_2024.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(cwd, "report_2025.txt"), []byte(""), 0644)

	tc := NewTabCompleter()

	input, cursor := tc.Complete("cat rep", 7, cwd)
	// Common prefix of "report_2024.txt" and "report_2025.txt" is "report_202"
	if input != "cat report_202" {
		t.Errorf("input = %q, want %q", input, "cat report_202")
	}
	if cursor != 14 {
		t.Errorf("cursor = %d, want %d", cursor, 14)
	}
}

func TestTabComplete_DirectoryAppendSlash(t *testing.T) {
	cwd := t.TempDir()
	os.Mkdir(filepath.Join(cwd, "mysubdir"), 0755)

	tc := NewTabCompleter()

	input, cursor := tc.Complete("cd mysub", 8, cwd)
	if input != "cd mysubdir/" {
		t.Errorf("input = %q, want %q", input, "cd mysubdir/")
	}
	if cursor != 12 {
		t.Errorf("cursor = %d, want %d", cursor, 12)
	}
}

func TestTabComplete_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "absolutetarget.txt"), []byte(""), 0644)

	tc := NewTabCompleter()
	cwd := t.TempDir() // different cwd

	input, cursor := tc.Complete("cat "+tmpDir+"/absolut", len("cat "+tmpDir+"/absolut"), cwd)
	expected := "cat " + tmpDir + "/absolutetarget.txt"
	if input != expected {
		t.Errorf("input = %q, want %q", input, expected)
	}
	if cursor != len(expected) {
		t.Errorf("cursor = %d, want %d", cursor, len(expected))
	}
}

func TestTabComplete_RelativePath(t *testing.T) {
	cwd := t.TempDir()
	os.Mkdir(filepath.Join(cwd, "subdir"), 0755)
	os.WriteFile(filepath.Join(cwd, "subdir", "relfile.txt"), []byte(""), 0644)

	tc := NewTabCompleter()

	input, cursor := tc.Complete("cat subdir/relf", 15, cwd)
	if input != "cat subdir/relfile.txt" {
		t.Errorf("input = %q, want %q", input, "cat subdir/relfile.txt")
	}
	if cursor != 22 {
		t.Errorf("cursor = %d, want %d", cursor, 22)
	}
}

func TestTabComplete_NoMatches(t *testing.T) {
	cwd := t.TempDir()

	tc := NewTabCompleter()

	input, cursor := tc.Complete("cat nonexistent", 15, cwd)
	if input != "cat nonexistent" {
		t.Errorf("input = %q, want %q", input, "cat nonexistent")
	}
	if cursor != 15 {
		t.Errorf("cursor = %d, want %d", cursor, 15)
	}
}

func TestTabComplete_MiddleOfWord(t *testing.T) {
	// Completion should work at cursor position, not end of input
	binDir := t.TempDir()
	createExecutable(t, binDir, "midtestcmd")

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir)
	defer os.Setenv("PATH", origPath)

	tc := NewTabCompleter()
	cwd := t.TempDir()

	// Cursor is at position 7 ("midtest") but there's more text after
	input, cursor := tc.Complete("midtest --flag", 7, cwd)
	if input != "midtestcmd  --flag" {
		t.Errorf("input = %q, want %q", input, "midtestcmd  --flag")
	}
	if cursor != 11 {
		t.Errorf("cursor = %d, want %d", cursor, 11)
	}
}

func TestTabComplete_HiddenFiles(t *testing.T) {
	cwd := t.TempDir()
	os.WriteFile(filepath.Join(cwd, ".hiddenconfig"), []byte(""), 0644)

	tc := NewTabCompleter()

	input, cursor := tc.Complete("cat .hidden", 11, cwd)
	if input != "cat .hiddenconfig" {
		t.Errorf("input = %q, want %q", input, "cat .hiddenconfig")
	}
	if cursor != 17 {
		t.Errorf("cursor = %d, want %d", cursor, 17)
	}
}

func TestTabComplete_PathCacheInvalidation(t *testing.T) {
	binDir1 := t.TempDir()
	binDir2 := t.TempDir()
	createExecutable(t, binDir1, "cachetest1cmd")
	createExecutable(t, binDir2, "cachetest2cmd")

	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	tc := NewTabCompleter()
	cwd := t.TempDir()

	// First with binDir1
	os.Setenv("PATH", binDir1)
	input1, _ := tc.Complete("cachetest", 9, cwd)

	// Change PATH to binDir2
	os.Setenv("PATH", binDir2)
	input2, _ := tc.Complete("cachetest", 9, cwd)

	// Should find cachetest2cmd now
	if input1 == input2 {
		t.Error("cache should have been invalidated when PATH changed")
	}
	if input1 != "cachetest1cmd " {
		t.Errorf("first completion = %q, want %q", input1, "cachetest1cmd ")
	}
	if input2 != "cachetest2cmd " {
		t.Errorf("second completion = %q, want %q", input2, "cachetest2cmd ")
	}
}

func TestIsCommandPosition(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		cursor   int
		expected bool
	}{
		{"first word", "ls", 2, true},
		{"typing first word", "cat", 3, true},
		{"empty input", "", 0, true},
		{"after space - argument", "ls f", 4, false},
		{"in middle of command", "ls -la foo", 5, false},
		{"just typed space", "cmd ", 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCommandPosition(tt.input, tt.cursor)
			if got != tt.expected {
				t.Errorf("isCommandPosition(%q, %d) = %v, want %v",
					tt.input, tt.cursor, got, tt.expected)
			}
		})
	}
}

func TestExtractWordAtCursor(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		cursorPos     int
		expectedWord  string
		expectedStart int
	}{
		{"full word", "hello", 5, "hello", 0},
		{"first word before space", "hello world", 5, "hello", 0},
		{"second word", "hello world", 11, "world", 6},
		{"partial second word", "hello wor", 9, "wor", 6},
		{"at beginning", "hello", 0, "", 0},
		{"cursor beyond end", "hello", 10, "hello", 0},
		{"middle of word", "hello", 3, "hel", 0},
		{"empty input", "", 0, "", 0},
		{"spaces only at cursor", "ls  file", 3, "", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word, start := extractWordAtCursor(tt.input, tt.cursorPos)
			if word != tt.expectedWord {
				t.Errorf("word = %q, want %q", word, tt.expectedWord)
			}
			if start != tt.expectedStart {
				t.Errorf("start = %d, want %d", start, tt.expectedStart)
			}
		})
	}
}

func TestCommonPrefix(t *testing.T) {
	tests := []struct {
		name       string
		candidates []string
		expected   string
	}{
		{"empty", nil, ""},
		{"single", []string{"hello"}, "hello"},
		{"common prefix", []string{"hello", "help"}, "hel"},
		{"no common", []string{"abc", "xyz"}, ""},
		{"identical", []string{"same", "same"}, "same"},
		{"one empty", []string{"", "hello"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := commonPrefix(tt.candidates)
			if got != tt.expected {
				t.Errorf("commonPrefix(%v) = %q, want %q",
					tt.candidates, got, tt.expected)
			}
		})
	}
}

func TestTabComplete_SymlinksIncluded(t *testing.T) {
	cwd := t.TempDir()
	// Create a real file and a symlink
	realFile := filepath.Join(cwd, "realfile.txt")
	os.WriteFile(realFile, []byte(""), 0644)

	linkPath := filepath.Join(cwd, "linkfile.txt")
	err := os.Symlink(realFile, linkPath)
	if err != nil {
		t.Skip("could not create symlink:", err)
	}

	tc := NewTabCompleter()

	// Complete "linkf" should find the symlink
	input, cursor := tc.Complete("cat linkf", 9, cwd)
	if input != "cat linkfile.txt" {
		t.Errorf("input = %q, want %q", input, "cat linkfile.txt")
	}
	if cursor != 16 {
		t.Errorf("cursor = %d, want %d", cursor, 16)
	}
}

func TestTabComplete_DotSlashPrefix(t *testing.T) {
	cwd := t.TempDir()
	os.WriteFile(filepath.Join(cwd, "dotslashfile.txt"), []byte(""), 0644)

	tc := NewTabCompleter()

	input, cursor := tc.Complete("cat ./dotslash", 14, cwd)
	if input != "cat ./dotslashfile.txt" {
		t.Errorf("input = %q, want %q", input, "cat ./dotslashfile.txt")
	}
	if cursor != 22 {
		t.Errorf("cursor = %d, want %d", cursor, 22)
	}
}

func TestTabComplete_TrailingSlashDir(t *testing.T) {
	cwd := t.TempDir()
	subDir := filepath.Join(cwd, "mydir")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "innerfile.txt"), []byte(""), 0644)

	tc := NewTabCompleter()

	// Complete inside "mydir/" directory
	input, cursor := tc.Complete("ls mydir/inner", 14, cwd)
	if input != "ls mydir/innerfile.txt" {
		t.Errorf("input = %q, want %q", input, "ls mydir/innerfile.txt")
	}
	if cursor != 22 {
		t.Errorf("cursor = %d, want %d", cursor, 22)
	}
}

// TestMinibufferCursorPosGetSet tests the new CursorPos and SetCursorPos methods
func TestMinibufferCursorPosGetSet(t *testing.T) {
	mb := NewMinibuffer()
	mb.input = "hello world"

	t.Run("CursorPos returns current position", func(t *testing.T) {
		mb.cursorPos = 5
		if mb.CursorPos() != 5 {
			t.Errorf("CursorPos() = %d, want %d", mb.CursorPos(), 5)
		}
	})

	t.Run("SetCursorPos sets valid position", func(t *testing.T) {
		mb.SetCursorPos(3)
		if mb.cursorPos != 3 {
			t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 3)
		}
	})

	t.Run("SetCursorPos clamps negative to 0", func(t *testing.T) {
		mb.SetCursorPos(-5)
		if mb.cursorPos != 0 {
			t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 0)
		}
	})

	t.Run("SetCursorPos clamps beyond length", func(t *testing.T) {
		mb.SetCursorPos(100)
		if mb.cursorPos != 11 { // len("hello world") = 11
			t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 11)
		}
	})

	t.Run("SetCursorPos at exact end", func(t *testing.T) {
		mb.SetCursorPos(11)
		if mb.cursorPos != 11 {
			t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 11)
		}
	})
}

// createExecutable creates a fake executable file in the given directory
func createExecutable(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0755)
	if err != nil {
		t.Fatal(err)
	}
}
