package archive

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseTarOutput(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "simple filename",
			line:     "file.txt",
			expected: "file.txt",
		},
		{
			name:     "filename with path",
			line:     "dir/subdir/file.txt",
			expected: "dir/subdir/file.txt",
		},
		{
			name:     "directory",
			line:     "mydir/",
			expected: "mydir/",
		},
		{
			name:     "empty line",
			line:     "",
			expected: "",
		},
		{
			name:     "whitespace only",
			line:     "   ",
			expected: "",
		},
		{
			name:     "line with leading/trailing spaces",
			line:     "  file.txt  ",
			expected: "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTarOutput(tt.line)
			if result != tt.expected {
				t.Errorf("ParseTarOutput(%q) = %q, want %q", tt.line, result, tt.expected)
			}
		})
	}
}

func TestTarExecutor_BuildCompressArgs(t *testing.T) {
	executor := NewTarExecutor()

	tests := []struct {
		name    string
		format  ArchiveFormat
		sources []string
		output  string
		level   int
		wantLen int // Expected number of arguments
	}{
		{
			name:    "tar format (no compression)",
			format:  FormatTar,
			sources: []string{"file1.txt", "file2.txt"},
			output:  "archive.tar",
			level:   6,
			wantLen: 4, // -cvf output sources...
		},
		{
			name:    "tar.gz format",
			format:  FormatTarGz,
			sources: []string{"file1.txt"},
			output:  "archive.tar.gz",
			level:   6,
			wantLen: 3, // -czvf output sources...
		},
		{
			name:    "tar.bz2 format",
			format:  FormatTarBz2,
			sources: []string{"file1.txt"},
			output:  "archive.tar.bz2",
			level:   9,
			wantLen: 3, // -cjvf output sources...
		},
		{
			name:    "tar.xz format",
			format:  FormatTarXz,
			sources: []string{"dir/"},
			output:  "archive.tar.xz",
			level:   3,
			wantLen: 3, // -cJvf output sources...
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := executor.BuildCompressArgs(tt.format, tt.sources, tt.output, tt.level)
			if len(args) < tt.wantLen {
				t.Errorf("BuildCompressArgs() returned %d args, want at least %d", len(args), tt.wantLen)
			}
			// Verify output file is in arguments
			found := false
			for _, arg := range args {
				if arg == tt.output {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("BuildCompressArgs() args = %v, output file %s not found", args, tt.output)
			}
		})
	}
}

func TestTarExecutor_BuildExtractArgs(t *testing.T) {
	executor := NewTarExecutor()

	tests := []struct {
		name        string
		format      ArchiveFormat
		archivePath string
		destDir     string
		wantLen     int
	}{
		{
			name:        "tar extract",
			format:      FormatTar,
			archivePath: "archive.tar",
			destDir:     "/tmp/extract",
			wantLen:     6, // -xvf archive --no-same-permissions --no-same-owner -C destDir
		},
		{
			name:        "tar.gz extract",
			format:      FormatTarGz,
			archivePath: "archive.tar.gz",
			destDir:     "/tmp/extract",
			wantLen:     6, // -xzvf archive --no-same-permissions --no-same-owner -C destDir
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := executor.BuildExtractArgs(tt.format, tt.archivePath, tt.destDir)
			if len(args) != tt.wantLen {
				t.Errorf("BuildExtractArgs() returned %d args, want %d", len(args), tt.wantLen)
			}
		})
	}
}

func TestTarExecutor_Compress(t *testing.T) {
	if !CheckCommand("tar") {
		t.Skip("tar command not available")
	}

	executor := NewTarExecutor()
	ctx := context.Background()

	// Create temp directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "test.tar")

	err := executor.Compress(ctx, FormatTar, []string{testFile}, outputFile, 6, nil)
	if err != nil {
		t.Errorf("Compress() error = %v", err)
	}

	// Verify archive was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Compress() did not create archive file")
	}
}

func TestTarExecutor_Extract(t *testing.T) {
	if !CheckCommand("tar") {
		t.Skip("tar command not available")
	}

	executor := NewTarExecutor()
	ctx := context.Background()

	// Create temp directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create archive
	archiveFile := filepath.Join(tempDir, "test.tar")
	if err := executor.Compress(ctx, FormatTar, []string{testFile}, archiveFile, 6, nil); err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}

	// Extract to new directory
	extractDir := filepath.Join(tempDir, "extract")
	if err := os.Mkdir(extractDir, 0755); err != nil {
		t.Fatalf("Failed to create extract directory: %v", err)
	}

	err := executor.Extract(ctx, FormatTar, archiveFile, extractDir, nil)
	if err != nil {
		t.Errorf("Extract() error = %v", err)
	}

	// Verify extracted file exists and has correct content
	extractedFile := filepath.Join(extractDir, "test.txt")
	content, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Errorf("Failed to read extracted file: %v", err)
	}
	if string(content) != string(testContent) {
		t.Errorf("Extracted file content = %q, want %q", content, testContent)
	}
}

func TestTarExecutor_ListContents(t *testing.T) {
	if !CheckCommand("tar") {
		t.Skip("tar command not available")
	}

	executor := NewTarExecutor()
	ctx := context.Background()

	// Create temp directory with test files
	tempDir := t.TempDir()
	testFile1 := filepath.Join(tempDir, "test1.txt")
	testFile2 := filepath.Join(tempDir, "test2.txt")
	os.WriteFile(testFile1, []byte("content1"), 0644)
	os.WriteFile(testFile2, []byte("content2"), 0644)

	// Create archive
	archiveFile := filepath.Join(tempDir, "test.tar")
	executor.Compress(ctx, FormatTar, []string{testFile1, testFile2}, archiveFile, 6, nil)

	// List contents
	contents, err := executor.ListContents(ctx, FormatTar, archiveFile)
	if err != nil {
		t.Errorf("ListContents() error = %v", err)
	}

	if len(contents) != 2 {
		t.Errorf("ListContents() returned %d items, want 2", len(contents))
	}
}

func TestTarExecutor_BuildExtractArgs_AllFormats(t *testing.T) {
	executor := NewTarExecutor()

	tests := []struct {
		name   string
		format ArchiveFormat
	}{
		{name: "tar", format: FormatTar},
		{name: "tar.gz", format: FormatTarGz},
		{name: "tar.bz2", format: FormatTarBz2},
		{name: "tar.xz", format: FormatTarXz},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := executor.BuildExtractArgs(tt.format, "archive.tar", "/tmp/extract")
			if len(args) < 4 {
				t.Errorf("BuildExtractArgs() for %s returned %d args, want at least 4", tt.name, len(args))
			}
		})
	}
}

func TestTarExecutor_Compress_WithProgress(t *testing.T) {
	if !CheckCommand("tar") {
		t.Skip("tar command not available")
	}

	executor := NewTarExecutor()
	ctx := context.Background()

	// Create temp directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "test.tar")
	progressChan := make(chan *ProgressUpdate, 10)

	go func() {
		for range progressChan {
			// Consume progress updates
		}
	}()

	err := executor.Compress(ctx, FormatTar, []string{testFile}, outputFile, 6, progressChan)
	close(progressChan)

	if err != nil {
		t.Errorf("Compress() with progress error = %v", err)
	}
}

func TestTarExecutor_Extract_TarGz(t *testing.T) {
	if !CheckCommand("tar") || !CheckCommand("gzip") {
		t.Skip("tar or gzip command not available")
	}

	executor := NewTarExecutor()
	ctx := context.Background()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("test content for gzip")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create tar.gz archive
	archiveFile := filepath.Join(tempDir, "test.tar.gz")
	if err := executor.Compress(ctx, FormatTarGz, []string{testFile}, archiveFile, 6, nil); err != nil {
		t.Fatalf("Failed to create tar.gz archive: %v", err)
	}

	// Extract
	extractDir := filepath.Join(tempDir, "extract")
	if err := os.Mkdir(extractDir, 0755); err != nil {
		t.Fatalf("Failed to create extract directory: %v", err)
	}

	err := executor.Extract(ctx, FormatTarGz, archiveFile, extractDir, nil)
	if err != nil {
		t.Errorf("Extract() tar.gz error = %v", err)
	}
}

func TestTarExecutor_ListContents_TarGz(t *testing.T) {
	if !CheckCommand("tar") || !CheckCommand("gzip") {
		t.Skip("tar or gzip command not available")
	}

	executor := NewTarExecutor()
	ctx := context.Background()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	archiveFile := filepath.Join(tempDir, "test.tar.gz")
	executor.Compress(ctx, FormatTarGz, []string{testFile}, archiveFile, 6, nil)

	contents, err := executor.ListContents(ctx, FormatTarGz, archiveFile)
	if err != nil {
		t.Errorf("ListContents() tar.gz error = %v", err)
	}

	if len(contents) != 1 {
		t.Errorf("ListContents() returned %d items, want 1", len(contents))
	}
}

func TestTarExecutor_CountFilesAndSize(t *testing.T) {
	executor := NewTarExecutor()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	count, size := executor.countFilesAndSize(tempDir)

	if count < 1 {
		t.Errorf("countFilesAndSize() count = %d, want at least 1", count)
	}
	if size < 7 {
		t.Errorf("countFilesAndSize() size = %d, want at least 7", size)
	}
}

func TestTarExecutor_CountFilesAndSize_NonExistent(t *testing.T) {
	executor := NewTarExecutor()

	count, size := executor.countFilesAndSize("/nonexistent/path")

	if count != 0 {
		t.Errorf("countFilesAndSize() count = %d, want 0 for non-existent", count)
	}
	if size != 0 {
		t.Errorf("countFilesAndSize() size = %d, want 0 for non-existent", size)
	}
}

func TestTarExecutor_CalculateSize(t *testing.T) {
	executor := NewTarExecutor()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	size := executor.calculateSize(tempDir)
	if size < 12 { // "test content" is 12 bytes
		t.Errorf("calculateSize() = %d, want at least 12", size)
	}
}

func TestTarExecutor_CalculateSize_SingleFile(t *testing.T) {
	executor := NewTarExecutor()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("hello world")
	os.WriteFile(testFile, content, 0644)

	size := executor.calculateSize(testFile)
	if size != int64(len(content)) {
		t.Errorf("calculateSize() = %d, want %d", size, len(content))
	}
}

func TestTarExecutor_CalculateSize_NonExistent(t *testing.T) {
	executor := NewTarExecutor()

	size := executor.calculateSize("/nonexistent/path")
	if size != 0 {
		t.Errorf("calculateSize() = %d, want 0 for non-existent", size)
	}
}

func TestSanitizePathForCommand(t *testing.T) {
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
			name:     "directory path starting with dash",
			input:    "-dir/file.txt",
			expected: "./-dir/file.txt",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizePathForCommand(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizePathForCommand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
