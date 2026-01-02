package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePath_PathTraversal(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid relative path",
			path:    "dir/file.txt",
			wantErr: false,
		},
		{
			name:    "path traversal with ..",
			path:    "../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path traversal in middle",
			path:    "dir/../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "absolute path",
			path:    "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "current directory",
			path:    "./file.txt",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckCompressionRatio(t *testing.T) {
	tests := []struct {
		name          string
		archiveSize   int64
		extractedSize int64
		wantWarning   bool
	}{
		{
			name:          "normal ratio",
			archiveSize:   1024 * 1024,      // 1 MB
			extractedSize: 10 * 1024 * 1024, // 10 MB
			wantWarning:   false,
		},
		{
			name:          "high ratio but acceptable",
			archiveSize:   1024 * 1024,       // 1 MB
			extractedSize: 100 * 1024 * 1024, // 100 MB
			wantWarning:   false,
		},
		{
			name:          "compression bomb",
			archiveSize:   1024 * 1024,        // 1 MB
			extractedSize: 2000 * 1024 * 1024, // 2 GB (ratio > 1:1000)
			wantWarning:   true,
		},
		{
			name:          "zero archive size",
			archiveSize:   0,
			extractedSize: 1024,
			wantWarning:   false, // Cannot calculate ratio
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warning := CheckCompressionRatio(tt.archiveSize, tt.extractedSize)
			if warning != tt.wantWarning {
				t.Errorf("CheckCompressionRatio() warning = %v, want %v", warning, tt.wantWarning)
			}
		})
	}
}

func TestCheckDiskSpace(t *testing.T) {
	// This test requires actual disk space checking
	// We just verify the function doesn't panic
	available := GetAvailableDiskSpace("/tmp")
	if available < 0 {
		t.Error("GetAvailableDiskSpace() returned negative value")
	}

	// Check if 1GB would fit
	insufficient, err := CheckDiskSpace("/tmp", 1024*1024*1024)
	if err != nil {
		t.Errorf("CheckDiskSpace() returned unexpected error: %v", err)
	}
	// We can't reliably test the result as it depends on actual disk space
	_ = insufficient
}

func TestValidateFileName(t *testing.T) {
	tests := []struct {
		name    string
		fname   string
		wantErr bool
	}{
		{
			name:    "valid filename",
			fname:   "archive.tar.gz",
			wantErr: false,
		},
		{
			name:    "empty filename",
			fname:   "",
			wantErr: true,
		},
		{
			name:    "filename with null character",
			fname:   "test\x00.tar",
			wantErr: true,
		},
		{
			name:    "filename with control character",
			fname:   "test\x01.tar",
			wantErr: true,
		},
		{
			name:    "filename with newline",
			fname:   "test\n.tar",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileName(tt.fname)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePath_MoreEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "simple filename",
			path:    "file.txt",
			wantErr: false,
		},
		{
			name:    "nested directory",
			path:    "a/b/c/d/file.txt",
			wantErr: false,
		},
		{
			name:    "path starting with dot",
			path:    ".hidden",
			wantErr: false,
		},
		{
			name:    "windows-style absolute path",
			path:    "C:/Windows/System32",
			wantErr: false, // Not absolute on Linux
		},
		{
			name:    "double dots in name",
			path:    "file..name.txt",
			wantErr: false, // Contains ".." but not as a path component, so it's valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestCheckDiskSpace_Error(t *testing.T) {
	// Test with non-existent path to trigger error case
	_, err := CheckDiskSpace("/nonexistent/path/that/does/not/exist", 1024)
	// Should return error because we can't check disk space
	if err == nil {
		t.Error("CheckDiskSpace() for non-existent path expected error, got nil")
	}
}

func TestGetAvailableDiskSpace_Error(t *testing.T) {
	// Test with non-existent path
	space := GetAvailableDiskSpace("/nonexistent/path/that/does/not/exist")
	if space != -1 {
		t.Errorf("GetAvailableDiskSpace() for non-existent path = %d, want -1", space)
	}
}

func TestCalculateFileHash(t *testing.T) {
	// Create a temp file with known content
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate hash
	hash, err := CalculateFileHash(testFile)
	if err != nil {
		t.Fatalf("CalculateFileHash() error = %v", err)
	}

	// SHA256 of "hello world" is known
	expectedHash := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expectedHash {
		t.Errorf("CalculateFileHash() = %s, want %s", hash, expectedHash)
	}
}

func TestCalculateFileHash_NonExistent(t *testing.T) {
	_, err := CalculateFileHash("/nonexistent/path")
	if err == nil {
		t.Error("CalculateFileHash() expected error for non-existent file, got nil")
	}
}

func TestVerifyFileHash(t *testing.T) {
	// Create a temp file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate hash
	hash, err := CalculateFileHash(testFile)
	if err != nil {
		t.Fatalf("CalculateFileHash() error = %v", err)
	}

	// Verify should pass with correct hash
	if err := VerifyFileHash(testFile, hash); err != nil {
		t.Errorf("VerifyFileHash() with correct hash error = %v", err)
	}

	// Verify should fail with incorrect hash
	if err := VerifyFileHash(testFile, "incorrect_hash"); err == nil {
		t.Error("VerifyFileHash() with incorrect hash expected error, got nil")
	}
}

func TestVerifyFileHash_NonExistent(t *testing.T) {
	err := VerifyFileHash("/nonexistent/path", "somehash")
	if err == nil {
		t.Error("VerifyFileHash() expected error for non-existent file, got nil")
	}
}

func TestValidateExtractedSymlinks(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	// Test 1: Directory with no symlinks should pass
	t.Run("no symlinks", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "no_symlinks")
		os.MkdirAll(testDir, 0755)
		os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("test"), 0644)

		err := ValidateExtractedSymlinks(testDir)
		if err != nil {
			t.Errorf("ValidateExtractedSymlinks() unexpected error: %v", err)
		}
	})

	// Test 2: Valid symlink within directory should pass
	t.Run("valid symlink", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "valid_symlink")
		os.MkdirAll(testDir, 0755)
		targetFile := filepath.Join(testDir, "target.txt")
		os.WriteFile(targetFile, []byte("test"), 0644)
		linkFile := filepath.Join(testDir, "link.txt")
		os.Symlink("target.txt", linkFile)

		err := ValidateExtractedSymlinks(testDir)
		if err != nil {
			t.Errorf("ValidateExtractedSymlinks() unexpected error for valid symlink: %v", err)
		}
	})

	// Test 3: Symlink with path traversal should fail
	t.Run("path traversal symlink", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "traversal_symlink")
		os.MkdirAll(testDir, 0755)
		linkFile := filepath.Join(testDir, "evil_link.txt")
		os.Symlink("../../../etc/passwd", linkFile)

		err := ValidateExtractedSymlinks(testDir)
		if err == nil {
			t.Error("ValidateExtractedSymlinks() expected error for path traversal symlink")
		}
		// Verify symlink was removed
		if _, statErr := os.Lstat(linkFile); !os.IsNotExist(statErr) {
			t.Error("Dangerous symlink was not removed")
		}
	})

	// Test 4: Absolute path symlink should fail
	t.Run("absolute symlink", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "absolute_symlink")
		os.MkdirAll(testDir, 0755)
		linkFile := filepath.Join(testDir, "abs_link.txt")
		os.Symlink("/etc/passwd", linkFile)

		err := ValidateExtractedSymlinks(testDir)
		if err == nil {
			t.Error("ValidateExtractedSymlinks() expected error for absolute symlink")
		}
		// Verify symlink was removed
		if _, statErr := os.Lstat(linkFile); !os.IsNotExist(statErr) {
			t.Error("Dangerous symlink was not removed")
		}
	})
}
