package archive

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseZipCompressOutput(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "adding with deflated",
			line:     "  adding: file.txt (deflated 45%)",
			expected: "file.txt",
		},
		{
			name:     "adding with stored",
			line:     "  adding: data.bin (stored 0%)",
			expected: "data.bin",
		},
		{
			name:     "adding path with deflated",
			line:     "  adding: dir/subdir/file.txt (deflated 32%)",
			expected: "dir/subdir/file.txt",
		},
		{
			name:     "empty line",
			line:     "",
			expected: "",
		},
		{
			name:     "non-adding line",
			line:     "updating: file.txt",
			expected: "",
		},
		{
			name:     "directory entry",
			line:     "  adding: mydir/ (stored 0%)",
			expected: "mydir/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseZipCompressOutput(tt.line)
			if result != tt.expected {
				t.Errorf("ParseZipCompressOutput(%q) = %q, want %q", tt.line, result, tt.expected)
			}
		})
	}
}

func TestParseZipExtractOutput(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "inflating file",
			line:     "  inflating: file.txt",
			expected: "file.txt",
		},
		{
			name:     "creating directory",
			line:     "   creating: mydir/",
			expected: "mydir/",
		},
		{
			name:     "extracting file",
			line:     " extracting: data.bin",
			expected: "data.bin",
		},
		{
			name:     "inflating path",
			line:     "  inflating: dir/subdir/file.txt",
			expected: "dir/subdir/file.txt",
		},
		{
			name:     "empty line",
			line:     "",
			expected: "",
		},
		{
			name:     "non-extract line",
			line:     "Archive:  test.zip",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseZipExtractOutput(tt.line)
			if result != tt.expected {
				t.Errorf("ParseZipExtractOutput(%q) = %q, want %q", tt.line, result, tt.expected)
			}
		})
	}
}

func TestZipExecutor_Compress(t *testing.T) {
	if !IsFormatAvailable(FormatZip) {
		t.Skip("zip command not available")
	}

	executor := NewZipExecutor()
	ctx := context.Background()

	// Create temp directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "test.zip")

	err := executor.Compress(ctx, []string{testFile}, outputFile, 6, nil)
	if err != nil {
		t.Errorf("Compress() error = %v", err)
	}

	// Verify archive was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Compress() did not create archive file")
	}
}

func TestZipExecutor_Extract(t *testing.T) {
	if !IsFormatAvailable(FormatZip) {
		t.Skip("zip/unzip commands not available")
	}

	executor := NewZipExecutor()
	ctx := context.Background()

	// Create temp directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create archive
	archiveFile := filepath.Join(tempDir, "test.zip")
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

func TestZipExecutor_ListContents(t *testing.T) {
	if !IsFormatAvailable(FormatZip) {
		t.Skip("zip/unzip commands not available")
	}

	executor := NewZipExecutor()
	ctx := context.Background()

	// Create temp directory with test files
	tempDir := t.TempDir()
	testFile1 := filepath.Join(tempDir, "test1.txt")
	testFile2 := filepath.Join(tempDir, "test2.txt")
	os.WriteFile(testFile1, []byte("content1"), 0644)
	os.WriteFile(testFile2, []byte("content2"), 0644)

	// Create archive
	archiveFile := filepath.Join(tempDir, "test.zip")
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

func TestSanitizeZipPath(t *testing.T) {
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
			result := sanitizeZipPath(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeZipPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
