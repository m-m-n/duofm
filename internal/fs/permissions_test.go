package fs

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestValidatePermissionMode tests permission mode validation
func TestValidatePermissionMode(t *testing.T) {
	tests := []struct {
		name      string
		mode      string
		wantError bool
	}{
		{
			name:      "valid 644",
			mode:      "644",
			wantError: false,
		},
		{
			name:      "valid 755",
			mode:      "755",
			wantError: false,
		},
		{
			name:      "valid boundary 000",
			mode:      "000",
			wantError: false,
		},
		{
			name:      "valid boundary 777",
			mode:      "777",
			wantError: false,
		},
		{
			name:      "invalid digit 888",
			mode:      "888",
			wantError: true,
		},
		{
			name:      "too short 64",
			mode:      "64",
			wantError: true,
		},
		{
			name:      "too long 6440",
			mode:      "6440",
			wantError: true,
		},
		{
			name:      "non-numeric abc",
			mode:      "abc",
			wantError: true,
		},
		{
			name:      "empty string",
			mode:      "",
			wantError: true,
		},
		{
			name:      "mixed valid and invalid 759",
			mode:      "759",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePermissionMode(tt.mode)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePermissionMode(%q) error = %v, wantError %v", tt.mode, err, tt.wantError)
			}
		})
	}
}

// TestParsePermissionMode tests octal string to FileMode conversion
func TestParsePermissionMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		expected fs.FileMode
		wantErr  bool
	}{
		{
			name:     "644",
			mode:     "644",
			expected: 0644,
			wantErr:  false,
		},
		{
			name:     "755",
			mode:     "755",
			expected: 0755,
			wantErr:  false,
		},
		{
			name:     "000",
			mode:     "000",
			expected: 0000,
			wantErr:  false,
		},
		{
			name:     "777",
			mode:     "777",
			expected: 0777,
			wantErr:  false,
		},
		{
			name:     "invalid 888",
			mode:     "888",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePermissionMode(tt.mode)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePermissionMode(%q) error = %v, wantErr %v", tt.mode, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("ParsePermissionMode(%q) = %o, want %o", tt.mode, got, tt.expected)
			}
		})
	}
}

// TestFormatSymbolic tests FileMode to symbolic string formatting
func TestFormatSymbolic(t *testing.T) {
	tests := []struct {
		name     string
		mode     fs.FileMode
		isDir    bool
		expected string
	}{
		{
			name:     "file 644",
			mode:     0644,
			isDir:    false,
			expected: "-rw-r--r--",
		},
		{
			name:     "dir 755",
			mode:     0755,
			isDir:    true,
			expected: "drwxr-xr-x",
		},
		{
			name:     "file 777",
			mode:     0777,
			isDir:    false,
			expected: "-rwxrwxrwx",
		},
		{
			name:     "file 000",
			mode:     0000,
			isDir:    false,
			expected: "----------",
		},
		{
			name:     "dir 700",
			mode:     0700,
			isDir:    true,
			expected: "drwx------",
		},
		{
			name:     "file 600",
			mode:     0600,
			isDir:    false,
			expected: "-rw-------",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSymbolic(tt.mode, tt.isDir)
			if got != tt.expected {
				t.Errorf("FormatSymbolic(%o, %v) = %q, want %q", tt.mode, tt.isDir, got, tt.expected)
			}
		})
	}
}

// TestChangePermission tests single file/directory permission change
func TestChangePermission(t *testing.T) {
	// Create temp directory for tests
	tempDir, err := os.MkdirTemp("", "permission-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("change file permission", func(t *testing.T) {
		// Create test file with 644
		testFile := filepath.Join(tempDir, "testfile.txt")
		err := os.WriteFile(testFile, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Change to 755
		err = ChangePermission(testFile, 0755)
		if err != nil {
			t.Errorf("ChangePermission() error = %v", err)
		}

		// Verify permission changed
		info, err := os.Stat(testFile)
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}
		if info.Mode().Perm() != 0755 {
			t.Errorf("Permission not changed: got %o, want %o", info.Mode().Perm(), 0755)
		}
	})

	t.Run("change directory permission", func(t *testing.T) {
		// Create test directory with 755
		testDir := filepath.Join(tempDir, "testdir")
		err := os.Mkdir(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test dir: %v", err)
		}

		// Change to 700
		err = ChangePermission(testDir, 0700)
		if err != nil {
			t.Errorf("ChangePermission() error = %v", err)
		}

		// Verify permission changed
		info, err := os.Stat(testDir)
		if err != nil {
			t.Fatalf("Failed to stat dir: %v", err)
		}
		if info.Mode().Perm() != 0700 {
			t.Errorf("Permission not changed: got %o, want %o", info.Mode().Perm(), 0700)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		nonExistent := filepath.Join(tempDir, "does-not-exist.txt")
		err := ChangePermission(nonExistent, 0644)
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
}

// TestChangePermissionRecursive tests recursive permission changes
func TestChangePermissionRecursive(t *testing.T) {
	// Create temp directory for tests
	tempDir, err := os.MkdirTemp("", "permission-recursive-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create directory tree
	subdir1 := filepath.Join(tempDir, "subdir1")
	subdir2 := filepath.Join(tempDir, "subdir1", "subdir2")
	os.MkdirAll(subdir2, 0755)

	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(subdir1, "file2.txt")
	file3 := filepath.Join(subdir2, "file3.txt")
	os.WriteFile(file1, []byte("test"), 0644)
	os.WriteFile(file2, []byte("test"), 0644)
	os.WriteFile(file3, []byte("test"), 0644)

	// Create symlink
	symlink := filepath.Join(tempDir, "symlink")
	os.Symlink(file1, symlink)

	t.Run("recursive change all directories and files", func(t *testing.T) {
		successCount, errors := ChangePermissionRecursive(tempDir, 0700, 0600)

		if len(errors) > 0 {
			t.Errorf("Expected no errors, got %d: %v", len(errors), errors)
		}

		if successCount != 6 { // tempDir, subdir1, subdir2, file1, file2, file3
			t.Errorf("Expected 6 successful changes, got %d", successCount)
		}

		// Verify directory permissions
		verifyPerm(t, tempDir, 0700)
		verifyPerm(t, subdir1, 0700)
		verifyPerm(t, subdir2, 0700)

		// Verify file permissions
		verifyPerm(t, file1, 0600)
		verifyPerm(t, file2, 0600)
		verifyPerm(t, file3, 0600)

		// Verify symlink was skipped (should still point to original)
		linkInfo, err := os.Lstat(symlink)
		if err != nil {
			t.Fatalf("Failed to stat symlink: %v", err)
		}
		if linkInfo.Mode()&os.ModeSymlink == 0 {
			t.Error("Symlink was not preserved")
		}
	})
}

// verifyPerm verifies that a file/directory has the expected permission
func verifyPerm(t *testing.T, path string, expected fs.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat %s: %v", path, err)
	}
	if info.Mode().Perm() != expected {
		t.Errorf("%s: expected permission %o, got %o", path, expected, info.Mode().Perm())
	}
}

// TestChangePermissionRecursiveWithProgress tests recursive permission changes with progress callback
func TestChangePermissionRecursiveWithProgress(t *testing.T) {
	// Create temp directory for tests
	tempDir, err := os.MkdirTemp("", "permission-progress-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create directory tree with multiple files
	subdir1 := filepath.Join(tempDir, "subdir1")
	subdir2 := filepath.Join(tempDir, "subdir2")
	os.Mkdir(subdir1, 0755)
	os.Mkdir(subdir2, 0755)

	for i := 0; i < 10; i++ {
		filename := filepath.Join(tempDir, "file"+string(rune('0'+i))+".txt")
		os.WriteFile(filename, []byte("test"), 0644)
	}

	t.Run("progress callback called", func(t *testing.T) {
		var mu sync.Mutex
		callbackCalled := false
		lastProcessed := 0
		var lastPath string

		callback := func(processed, total int, path string) {
			mu.Lock()
			defer mu.Unlock()
			callbackCalled = true
			lastProcessed = processed
			lastPath = path
		}

		successCount, errors := ChangePermissionRecursiveWithProgress(tempDir, 0755, 0644, callback)

		if !callbackCalled {
			t.Error("Expected progress callback to be called")
		}

		if len(errors) > 0 {
			t.Errorf("Expected no errors, got %d", len(errors))
		}

		// Should have processed tempDir + 2 subdirs + 10 files = 13 items
		if successCount != 13 {
			t.Errorf("Expected 13 successful changes, got %d", successCount)
		}

		mu.Lock()
		if lastProcessed == 0 {
			t.Error("Expected lastProcessed to be updated")
		}
		if lastPath == "" {
			t.Error("Expected lastPath to be set")
		}
		mu.Unlock()
	})

	t.Run("progress callback nil (should not crash)", func(t *testing.T) {
		successCount, errors := ChangePermissionRecursiveWithProgress(tempDir, 0755, 0644, nil)

		if len(errors) > 0 {
			t.Errorf("Expected no errors, got %d", len(errors))
		}

		if successCount != 13 {
			t.Errorf("Expected 13 successful changes, got %d", successCount)
		}
	})
}
