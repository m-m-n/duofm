package archive

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParse7zCompressOutput(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "plus prefix",
			line:     "+ file.txt",
			expected: "file.txt",
		},
		{
			name:     "plus prefix with path",
			line:     "+ dir/subdir/file.txt",
			expected: "dir/subdir/file.txt",
		},
		{
			name:     "compressing prefix",
			line:     "Compressing  file.txt",
			expected: "file.txt",
		},
		{
			name:     "compressing with path",
			line:     "Compressing  dir/subdir/file.txt",
			expected: "dir/subdir/file.txt",
		},
		{
			name:     "empty line",
			line:     "",
			expected: "",
		},
		{
			name:     "header line",
			line:     "7-Zip 16.02 : Copyright",
			expected: "",
		},
		{
			name:     "directory with plus",
			line:     "+ mydir/",
			expected: "mydir/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse7zCompressOutput(tt.line)
			if result != tt.expected {
				t.Errorf("Parse7zCompressOutput(%q) = %q, want %q", tt.line, result, tt.expected)
			}
		})
	}
}

func TestParse7zExtractOutput(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "minus prefix",
			line:     "- file.txt",
			expected: "file.txt",
		},
		{
			name:     "minus prefix with path",
			line:     "- dir/subdir/file.txt",
			expected: "dir/subdir/file.txt",
		},
		{
			name:     "extracting prefix",
			line:     "Extracting  file.txt",
			expected: "file.txt",
		},
		{
			name:     "extracting with path",
			line:     "Extracting  dir/subdir/file.txt",
			expected: "dir/subdir/file.txt",
		},
		{
			name:     "empty line",
			line:     "",
			expected: "",
		},
		{
			name:     "header line",
			line:     "7-Zip 16.02 : Copyright",
			expected: "",
		},
		{
			name:     "directory with minus",
			line:     "- mydir/",
			expected: "mydir/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse7zExtractOutput(tt.line)
			if result != tt.expected {
				t.Errorf("Parse7zExtractOutput(%q) = %q, want %q", tt.line, result, tt.expected)
			}
		})
	}
}

func TestSevenZipExecutor_Compress(t *testing.T) {
	if !IsFormatAvailable(Format7z) {
		t.Skip("7z command not available")
	}

	executor := NewSevenZipExecutor()
	ctx := context.Background()

	// Create temp directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "test.7z")

	err := executor.Compress(ctx, []string{testFile}, outputFile, 6, nil)
	if err != nil {
		t.Errorf("Compress() error = %v", err)
	}

	// Verify archive was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Compress() did not create archive file")
	}
}

func TestSevenZipExecutor_Extract(t *testing.T) {
	if !IsFormatAvailable(Format7z) {
		t.Skip("7z command not available")
	}

	executor := NewSevenZipExecutor()
	ctx := context.Background()

	// Create temp directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create archive
	archiveFile := filepath.Join(tempDir, "test.7z")
	if err := executor.Compress(ctx, []string{testFile}, archiveFile, 6, nil); err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}

	// Extract to new directory
	extractDir := filepath.Join(tempDir, "extract")
	if err := os.Mkdir(extractDir, 0755); err != nil {
		t.Fatalf("Failed to create extract directory: %v", err)
	}

	err := executor.Extract(ctx, archiveFile, extractDir, nil)
	if err != nil {
		t.Errorf("Extract() error = %v", err)
	}

	// Verify extracted file exists
	extractedFile := filepath.Join(extractDir, "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("Extract() did not extract file")
	}
}

func TestSevenZipExecutor_ListContents(t *testing.T) {
	if !IsFormatAvailable(Format7z) {
		t.Skip("7z command not available")
	}

	executor := NewSevenZipExecutor()
	ctx := context.Background()

	// Create temp directory with test files
	tempDir := t.TempDir()
	testFile1 := filepath.Join(tempDir, "test1.txt")
	testFile2 := filepath.Join(tempDir, "test2.txt")
	os.WriteFile(testFile1, []byte("content1"), 0644)
	os.WriteFile(testFile2, []byte("content2"), 0644)

	// Create archive
	archiveFile := filepath.Join(tempDir, "test.7z")
	executor.Compress(ctx, []string{testFile1, testFile2}, archiveFile, 6, nil)

	// List contents
	contents, err := executor.ListContents(ctx, archiveFile)
	if err != nil {
		t.Errorf("ListContents() error = %v", err)
	}

	if len(contents) < 2 {
		t.Errorf("ListContents() returned %d items, want at least 2", len(contents))
	}
}

func TestSanitize7zPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal path",
			input:    "file.txt",
			expected: "file.txt",
		},
		{
			name:     "path starting with dash",
			input:    "-filename.txt",
			expected: "./-filename.txt",
		},
		{
			name:     "path with multiple dashes",
			input:    "--dangerous-file",
			expected: "./--dangerous-file",
		},
		{
			name:     "path with dash in middle",
			input:    "my-file.txt",
			expected: "my-file.txt",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitize7zPath(tt.input)
			if result != tt.expected {
				t.Errorf("sanitize7zPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
