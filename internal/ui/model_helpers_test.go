package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists_Exists(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "testfile")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	exists, err := fileExists(tmpFile)
	if err != nil {
		t.Errorf("fileExists() returned error: %v", err)
	}
	if !exists {
		t.Error("fileExists() = false, want true for existing file")
	}
}

func TestFileExists_NotExists(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "nonexistent")

	exists, err := fileExists(tmpFile)
	if err != nil {
		t.Errorf("fileExists() returned error: %v", err)
	}
	if exists {
		t.Error("fileExists() = true, want false for non-existing file")
	}
}

func TestFileExists_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	exists, err := fileExists(tmpDir)
	if err != nil {
		t.Errorf("fileExists() returned error: %v", err)
	}
	if !exists {
		t.Error("fileExists() = false, want true for existing directory")
	}
}

func TestRemoveFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "testfile")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := removeFile(tmpFile)
	if err != nil {
		t.Errorf("removeFile() returned error: %v", err)
	}

	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("removeFile() did not remove the file")
	}
}

func TestRemoveFile_NotExists(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "nonexistent")

	err := removeFile(tmpFile)
	if err == nil {
		t.Error("removeFile() should return error for non-existing file")
	}
}

func TestRemoveAllFiles(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	testFile := filepath.Join(subDir, "testfile")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := removeAllFiles(subDir)
	if err != nil {
		t.Errorf("removeAllFiles() returned error: %v", err)
	}

	if _, err := os.Stat(subDir); !os.IsNotExist(err) {
		t.Error("removeAllFiles() did not remove the directory")
	}
}

func TestIsPermissionError_True(t *testing.T) {
	// Create permission error
	err := os.ErrPermission

	if !isPermissionError(err) {
		t.Error("isPermissionError() = false, want true for permission error")
	}
}

func TestIsPermissionError_False(t *testing.T) {
	err := os.ErrNotExist

	if isPermissionError(err) {
		t.Error("isPermissionError() = true, want false for non-permission error")
	}
}

func TestIsPermissionError_Nil(t *testing.T) {
	if isPermissionError(nil) {
		t.Error("isPermissionError(nil) = true, want false")
	}
}
