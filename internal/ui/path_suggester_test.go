package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathSuggester_BasicCompletion(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	s := NewPathSuggester()

	// Test: "/tmp/xxx/test" should suggest "dir" to complete "testdir"
	input := filepath.Join(tmpDir, "test")
	got := s.Suggest(input)
	expected := "dir"

	if got != expected {
		t.Errorf("Suggest(%q) = %q, want %q", input, got, expected)
	}
}

func TestPathSuggester_NoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "aaa")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	s := NewPathSuggester()

	// Test: prefix "xyz" should not match "aaa"
	input := filepath.Join(tmpDir, "xyz")
	got := s.Suggest(input)

	if got != "" {
		t.Errorf("Suggest(%q) = %q, want empty string", input, got)
	}
}

func TestPathSuggester_DirectoriesOnly(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	file := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a directory with same prefix
	dir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	s := NewPathSuggester()

	// Test: "test" prefix should only suggest directory "testdir", not file "testfile.txt"
	input := filepath.Join(tmpDir, "test")
	got := s.Suggest(input)
	expected := "dir" // Not "file.txt"

	if got != expected {
		t.Errorf("Suggest(%q) = %q, want %q", input, got, expected)
	}
}

func TestPathSuggester_HiddenDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create hidden directory
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	if err := os.Mkdir(hiddenDir, 0755); err != nil {
		t.Fatalf("Failed to create hidden directory: %v", err)
	}

	s := NewPathSuggester()

	// Test: ".hid" should suggest "den" to complete ".hidden"
	input := filepath.Join(tmpDir, ".hid")
	got := s.Suggest(input)
	expected := "den"

	if got != expected {
		t.Errorf("Suggest(%q) = %q, want %q", input, got, expected)
	}
}

func TestPathSuggester_RootPath(t *testing.T) {
	s := NewPathSuggester()

	// Test: "/" should suggest the first directory in root
	got := s.Suggest("/")

	// Should return something (there's always at least one directory in root)
	// We just check it doesn't crash and returns non-empty
	// The actual value depends on the system
	if got == "" {
		t.Log("Root path suggestion is empty (may be expected on some systems)")
	}
}

func TestPathSuggester_NonExistentParent(t *testing.T) {
	s := NewPathSuggester()

	// Test: non-existent parent should return empty
	input := "/nonexistent/path/prefix"
	got := s.Suggest(input)

	if got != "" {
		t.Errorf("Suggest(%q) = %q, want empty string", input, got)
	}
}

func TestPathSuggester_CaseSensitive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory with lowercase name
	dir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	s := NewPathSuggester()

	// Test: "Test" (uppercase T) should not match "testdir"
	input := filepath.Join(tmpDir, "Test")
	got := s.Suggest(input)

	if got != "" {
		t.Errorf("Suggest(%q) = %q, want empty string (case-sensitive)", input, got)
	}
}

func TestPathSuggester_TrailingSlash(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a child directory inside subdir
	childDir := filepath.Join(subDir, "child")
	if err := os.Mkdir(childDir, 0755); err != nil {
		t.Fatalf("Failed to create child directory: %v", err)
	}

	s := NewPathSuggester()

	// Test: path ending with "/" should suggest children of that directory
	input := subDir + "/"
	got := s.Suggest(input)
	expected := "child"

	if got != expected {
		t.Errorf("Suggest(%q) = %q, want %q", input, got, expected)
	}
}

func TestPathSuggester_EmptyInput(t *testing.T) {
	s := NewPathSuggester()

	// Test: empty input should return empty
	got := s.Suggest("")

	if got != "" {
		t.Errorf("Suggest(\"\") = %q, want empty string", got)
	}
}

func TestPathSuggester_RelativePath(t *testing.T) {
	s := NewPathSuggester()

	// Test: relative path (no leading /) should return empty
	got := s.Suggest("relative/path")

	if got != "" {
		t.Errorf("Suggest(\"relative/path\") = %q, want empty string", got)
	}
}

func TestPathSuggester_AlphabeticalOrder(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directories in non-alphabetical order
	dirs := []string{"zebra", "apple", "mango"}
	for _, d := range dirs {
		if err := os.Mkdir(filepath.Join(tmpDir, d), 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", d, err)
		}
	}

	s := NewPathSuggester()

	// Test: with no prefix, should suggest first alphabetically ("apple")
	input := tmpDir + "/"
	got := s.Suggest(input)
	expected := "apple"

	if got != expected {
		t.Errorf("Suggest(%q) = %q, want %q", input, got, expected)
	}
}

func TestPathSuggester_ExactMatch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory
	dir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	s := NewPathSuggester()

	// Test: exact match should return empty suffix (already complete)
	input := filepath.Join(tmpDir, "testdir")
	got := s.Suggest(input)

	if got != "" {
		t.Errorf("Suggest(%q) = %q, want empty string (exact match)", input, got)
	}
}

func TestPathSuggester_MultipleMatchesReturnsFirst(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple directories with same prefix
	dirs := []string{"test_c", "test_a", "test_b"}
	for _, d := range dirs {
		if err := os.Mkdir(filepath.Join(tmpDir, d), 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", d, err)
		}
	}

	s := NewPathSuggester()

	// Test: should return suffix for first alphabetical match (test_a)
	input := filepath.Join(tmpDir, "test_")
	got := s.Suggest(input)
	expected := "a"

	if got != expected {
		t.Errorf("Suggest(%q) = %q, want %q", input, got, expected)
	}
}

func TestPathSuggester_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	s := NewPathSuggester()

	// Test: empty directory with trailing slash
	input := emptyDir + "/"
	got := s.Suggest(input)

	if got != "" {
		t.Errorf("Suggest(%q) = %q, want empty string for empty directory", input, got)
	}
}

func TestPathSuggester_RootSlashOnly(t *testing.T) {
	s := NewPathSuggester()

	// Test "/" which goes through suggestChildren with dir = "" converted to "/"
	got := s.Suggest("/")

	// The result depends on the system, but this exercises the code path
	// where dir == "" is converted to "/"
	t.Logf("Suggest(\"/\") = %q", got)
}
