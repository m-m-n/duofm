package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTabComplete_EmptyInput(t *testing.T) {
	tc := NewTabCompleter()
	cwd := t.TempDir()

	result := tc.Complete("", 0, cwd)
	if result.NewInput != "" {
		t.Errorf("input = %q, want %q", result.NewInput, "")
	}
	if result.NewCursorPos != 0 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 0)
	}
	if result.HasProgress {
		t.Error("HasProgress should be false for empty input")
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

	result := tc.Complete("myuniquetestcmd1", 16, cwd)
	// Single match should auto-complete with trailing space
	if result.NewInput != "myuniquetestcmd123 " {
		t.Errorf("input = %q, want %q", result.NewInput, "myuniquetestcmd123 ")
	}
	if result.NewCursorPos != 19 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 19)
	}
	if !result.HasProgress {
		t.Error("HasProgress should be true for single match")
	}
	if len(result.Candidates) != 1 {
		t.Errorf("Candidates = %d, want 1", len(result.Candidates))
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

	result := tc.Complete("testcmd_", 8, cwd)
	// Multiple matches should return common prefix only
	if result.NewInput != "testcmd_" {
		t.Errorf("input = %q, want %q (common prefix, no progress)", result.NewInput, "testcmd_")
	}
	// No progress made, should return original cursor
	if result.NewCursorPos != 8 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 8)
	}
	if result.HasProgress {
		t.Error("HasProgress should be false when no progress made")
	}
	if len(result.Candidates) != 2 {
		t.Errorf("Candidates = %d, want 2", len(result.Candidates))
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

	result := tc.Complete("testx", 5, cwd)
	// Common prefix is "testxyz_"
	if result.NewInput != "testxyz_" {
		t.Errorf("input = %q, want %q", result.NewInput, "testxyz_")
	}
	if result.NewCursorPos != 8 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 8)
	}
	if !result.HasProgress {
		t.Error("HasProgress should be true when common prefix extends input")
	}
}

func TestTabComplete_SingleFile(t *testing.T) {
	cwd := t.TempDir()
	os.WriteFile(filepath.Join(cwd, "uniquefile.txt"), []byte(""), 0644)

	tc := NewTabCompleter()

	result := tc.Complete("ls uniquef", 10, cwd)
	if result.NewInput != "ls uniquefile.txt" {
		t.Errorf("input = %q, want %q", result.NewInput, "ls uniquefile.txt")
	}
	if result.NewCursorPos != 17 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 17)
	}
}

func TestTabComplete_MultipleFiles(t *testing.T) {
	cwd := t.TempDir()
	os.WriteFile(filepath.Join(cwd, "report_2024.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(cwd, "report_2025.txt"), []byte(""), 0644)

	tc := NewTabCompleter()

	result := tc.Complete("cat rep", 7, cwd)
	// Common prefix of "report_2024.txt" and "report_2025.txt" is "report_202"
	if result.NewInput != "cat report_202" {
		t.Errorf("input = %q, want %q", result.NewInput, "cat report_202")
	}
	if result.NewCursorPos != 14 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 14)
	}
}

func TestTabComplete_DirectoryAppendSlash(t *testing.T) {
	cwd := t.TempDir()
	os.Mkdir(filepath.Join(cwd, "mysubdir"), 0755)

	tc := NewTabCompleter()

	result := tc.Complete("cd mysub", 8, cwd)
	if result.NewInput != "cd mysubdir/" {
		t.Errorf("input = %q, want %q", result.NewInput, "cd mysubdir/")
	}
	if result.NewCursorPos != 12 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 12)
	}
}

func TestTabComplete_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "absolutetarget.txt"), []byte(""), 0644)

	tc := NewTabCompleter()
	cwd := t.TempDir() // different cwd

	result := tc.Complete("cat "+tmpDir+"/absolut", len("cat "+tmpDir+"/absolut"), cwd)
	expected := "cat " + tmpDir + "/absolutetarget.txt"
	if result.NewInput != expected {
		t.Errorf("input = %q, want %q", result.NewInput, expected)
	}
	if result.NewCursorPos != len(expected) {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, len(expected))
	}
}

func TestTabComplete_RelativePath(t *testing.T) {
	cwd := t.TempDir()
	os.Mkdir(filepath.Join(cwd, "subdir"), 0755)
	os.WriteFile(filepath.Join(cwd, "subdir", "relfile.txt"), []byte(""), 0644)

	tc := NewTabCompleter()

	result := tc.Complete("cat subdir/relf", 15, cwd)
	if result.NewInput != "cat subdir/relfile.txt" {
		t.Errorf("input = %q, want %q", result.NewInput, "cat subdir/relfile.txt")
	}
	if result.NewCursorPos != 22 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 22)
	}
}

func TestTabComplete_NoMatches(t *testing.T) {
	cwd := t.TempDir()

	tc := NewTabCompleter()

	result := tc.Complete("cat nonexistent", 15, cwd)
	if result.NewInput != "cat nonexistent" {
		t.Errorf("input = %q, want %q", result.NewInput, "cat nonexistent")
	}
	if result.NewCursorPos != 15 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 15)
	}
	if result.HasProgress {
		t.Error("HasProgress should be false for no matches")
	}
	if len(result.Candidates) != 0 {
		t.Errorf("Candidates = %d, want 0", len(result.Candidates))
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
	result := tc.Complete("midtest --flag", 7, cwd)
	if result.NewInput != "midtestcmd  --flag" {
		t.Errorf("input = %q, want %q", result.NewInput, "midtestcmd  --flag")
	}
	if result.NewCursorPos != 11 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 11)
	}
}

func TestTabComplete_HiddenFiles(t *testing.T) {
	cwd := t.TempDir()
	os.WriteFile(filepath.Join(cwd, ".hiddenconfig"), []byte(""), 0644)

	tc := NewTabCompleter()

	result := tc.Complete("cat .hidden", 11, cwd)
	if result.NewInput != "cat .hiddenconfig" {
		t.Errorf("input = %q, want %q", result.NewInput, "cat .hiddenconfig")
	}
	if result.NewCursorPos != 17 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 17)
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
	result1 := tc.Complete("cachetest", 9, cwd)

	// Change PATH to binDir2
	os.Setenv("PATH", binDir2)
	result2 := tc.Complete("cachetest", 9, cwd)

	// Should find cachetest2cmd now
	if result1.NewInput == result2.NewInput {
		t.Error("cache should have been invalidated when PATH changed")
	}
	if result1.NewInput != "cachetest1cmd " {
		t.Errorf("first completion = %q, want %q", result1.NewInput, "cachetest1cmd ")
	}
	if result2.NewInput != "cachetest2cmd " {
		t.Errorf("second completion = %q, want %q", result2.NewInput, "cachetest2cmd ")
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
	result := tc.Complete("cat linkf", 9, cwd)
	if result.NewInput != "cat linkfile.txt" {
		t.Errorf("input = %q, want %q", result.NewInput, "cat linkfile.txt")
	}
	if result.NewCursorPos != 16 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 16)
	}
}

func TestTabComplete_DotSlashPrefix(t *testing.T) {
	cwd := t.TempDir()
	os.WriteFile(filepath.Join(cwd, "dotslashfile.txt"), []byte(""), 0644)

	tc := NewTabCompleter()

	result := tc.Complete("cat ./dotslash", 14, cwd)
	if result.NewInput != "cat ./dotslashfile.txt" {
		t.Errorf("input = %q, want %q", result.NewInput, "cat ./dotslashfile.txt")
	}
	if result.NewCursorPos != 22 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 22)
	}
}

func TestTabComplete_TrailingSlashDir(t *testing.T) {
	cwd := t.TempDir()
	subDir := filepath.Join(cwd, "mydir")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "innerfile.txt"), []byte(""), 0644)

	tc := NewTabCompleter()

	// Complete inside "mydir/" directory
	result := tc.Complete("ls mydir/inner", 14, cwd)
	if result.NewInput != "ls mydir/innerfile.txt" {
		t.Errorf("input = %q, want %q", result.NewInput, "ls mydir/innerfile.txt")
	}
	if result.NewCursorPos != 22 {
		t.Errorf("cursor = %d, want %d", result.NewCursorPos, 22)
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

// TestCompletionResult_CandidatesReturned verifies candidates are included in result
func TestCompletionResult_CandidatesReturned(t *testing.T) {
	binDir := t.TempDir()
	createExecutable(t, binDir, "testcand_aaa")
	createExecutable(t, binDir, "testcand_bbb")
	createExecutable(t, binDir, "testcand_ccc")

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir)
	defer os.Setenv("PATH", origPath)

	tc := NewTabCompleter()
	cwd := t.TempDir()

	result := tc.Complete("testcand_", 9, cwd)
	if len(result.Candidates) != 3 {
		t.Fatalf("Candidates = %d, want 3", len(result.Candidates))
	}
	expected := []string{"testcand_aaa", "testcand_bbb", "testcand_ccc"}
	for i, c := range result.Candidates {
		if c != expected[i] {
			t.Errorf("Candidates[%d] = %q, want %q", i, c, expected[i])
		}
	}
}

// TestFormatCandidateList tests the candidate list formatting
func TestFormatCandidateList(t *testing.T) {
	tests := []struct {
		name       string
		candidates []string
		maxWidth   int
		check      func(t *testing.T, result string)
	}{
		{
			name:       "empty candidates",
			candidates: nil,
			maxWidth:   80,
			check: func(t *testing.T, result string) {
				if result != "" {
					t.Errorf("result = %q, want empty", result)
				}
			},
		},
		{
			name:       "all fit",
			candidates: []string{"ls", "cat", "rm"},
			maxWidth:   80,
			check: func(t *testing.T, result string) {
				if result != "ls cat rm" {
					t.Errorf("result = %q, want %q", result, "ls cat rm")
				}
			},
		},
		{
			name:       "truncated with more",
			candidates: []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot"},
			maxWidth:   25,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "more)") {
					t.Errorf("result should contain (+N more), got %q", result)
				}
				if !strings.HasPrefix(result, "alpha") {
					t.Errorf("result should start with first candidate, got %q", result)
				}
			},
		},
		{
			name:       "zero width defaults to 80",
			candidates: []string{"aaa", "bbb"},
			maxWidth:   0,
			check: func(t *testing.T, result string) {
				if result != "aaa bbb" {
					t.Errorf("result = %q, want %q", result, "aaa bbb")
				}
			},
		},
		{
			name:       "last candidate fits without more suffix",
			candidates: []string{"a", "b"},
			maxWidth:   5,
			check: func(t *testing.T, result string) {
				if result != "a b" {
					t.Errorf("result = %q, want %q", result, "a b")
				}
			},
		},
		{
			name:       "single candidate always shown",
			candidates: []string{"verylongname"},
			maxWidth:   5,
			check: func(t *testing.T, result string) {
				if result != "verylongname" {
					t.Errorf("result = %q, want %q", result, "verylongname")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCandidateList(tt.candidates, tt.maxWidth)
			tt.check(t, result)
		})
	}
}

// TestFormatCandidateList_MoreCount verifies the (+N more) count is accurate
func TestFormatCandidateList_MoreCount(t *testing.T) {
	// Create 10 candidates of 5 chars each
	candidates := make([]string, 10)
	for i := range candidates {
		candidates[i] = fmt.Sprintf("cmd%02d", i)
	}

	// Width 20: should show some and truncate rest
	result := formatCandidateList(candidates, 20)
	if !strings.Contains(result, "more)") {
		t.Errorf("expected truncation, got %q", result)
	}

	// Count shown candidates (everything before (+N more))
	parts := strings.Fields(result)
	moreIdx := -1
	for i, p := range parts {
		if strings.HasPrefix(p, "(+") {
			moreIdx = i
			break
		}
	}
	if moreIdx == -1 {
		t.Fatal("could not find (+N more) in result")
	}

	shownCount := moreIdx
	// Extract N from "(+N"
	moreStr := strings.TrimPrefix(parts[moreIdx], "(+")
	var moreCount int
	fmt.Sscanf(moreStr, "%d", &moreCount)

	if shownCount+moreCount != 10 {
		t.Errorf("shown(%d) + more(%d) = %d, want 10", shownCount, moreCount, shownCount+moreCount)
	}
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
