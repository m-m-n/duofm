package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTrashDir(t *testing.T) {
	// Save and restore XDG_DATA_HOME
	origXDG := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	tests := []struct {
		name        string
		xdgDataHome string
		wantSuffix  string
	}{
		{
			name:        "with XDG_DATA_HOME set",
			xdgDataHome: "/custom/data",
			wantSuffix:  "/custom/data/Trash",
		},
		{
			name:        "without XDG_DATA_HOME",
			xdgDataHome: "",
			wantSuffix:  "/.local/share/Trash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("XDG_DATA_HOME", tt.xdgDataHome)
			got := TrashDir()
			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("TrashDir() = %q, want suffix %q", got, tt.wantSuffix)
			}
		})
	}
}

func TestEnsureTrashDirs(t *testing.T) {
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "Trash")

	// Save and restore XDG_DATA_HOME
	origXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	err := EnsureTrashDirs()
	if err != nil {
		t.Fatalf("EnsureTrashDirs() error = %v", err)
	}

	// Verify directories exist
	filesDir := filepath.Join(trashDir, "files")
	infoDir := filepath.Join(trashDir, "info")

	if _, err := os.Stat(filesDir); os.IsNotExist(err) {
		t.Errorf("EnsureTrashDirs() did not create files/ directory")
	}
	if _, err := os.Stat(infoDir); os.IsNotExist(err) {
		t.Errorf("EnsureTrashDirs() did not create info/ directory")
	}
}

func TestResolveNameCollision(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		existingFiles []string
		baseName      string
		want          string
	}{
		{
			name:          "no collision",
			existingFiles: []string{},
			baseName:      "file.txt",
			want:          "file.txt",
		},
		{
			name:          "first collision",
			existingFiles: []string{"file.txt"},
			baseName:      "file.txt",
			want:          "file.2.txt",
		},
		{
			name:          "second collision",
			existingFiles: []string{"file.txt", "file.2.txt"},
			baseName:      "file.txt",
			want:          "file.3.txt",
		},
		{
			name:          "multiple collisions",
			existingFiles: []string{"file.txt", "file.2.txt", "file.3.txt", "file.4.txt"},
			baseName:      "file.txt",
			want:          "file.5.txt",
		},
		{
			name:          "no extension",
			existingFiles: []string{"file"},
			baseName:      "file",
			want:          "file.2",
		},
		{
			name:          "directory collision",
			existingFiles: []string{"mydir"},
			baseName:      "mydir",
			want:          "mydir.2",
		},
		{
			name:          "dotfile",
			existingFiles: []string{".config"},
			baseName:      ".config",
			want:          ".config.2",
		},
		{
			name:          "double extension",
			existingFiles: []string{"file.tar.gz"},
			baseName:      "file.tar.gz",
			want:          "file.tar.2.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh subdirectory for each test
			testDir := filepath.Join(tmpDir, tt.name)
			if err := os.MkdirAll(testDir, 0755); err != nil {
				t.Fatalf("failed to create test dir: %v", err)
			}

			// Create existing files
			for _, f := range tt.existingFiles {
				path := filepath.Join(testDir, f)
				if err := os.WriteFile(path, []byte{}, 0644); err != nil {
					t.Fatalf("failed to create existing file: %v", err)
				}
			}

			got := ResolveNameCollision(testDir, tt.baseName)
			if got != tt.want {
				t.Errorf("ResolveNameCollision() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMoveToTrash(t *testing.T) {
	// Create temp directories for test
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "Trash")

	// Save and restore XDG_DATA_HOME
	origXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	tests := []struct {
		name        string
		setup       func(t *testing.T) string // returns path to file/dir to trash
		wantErr     bool
		errContains string
		verify      func(t *testing.T, originalPath string)
	}{
		{
			name: "single file",
			setup: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "testfile.txt")
				if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return path
			},
			verify: func(t *testing.T, originalPath string) {
				// Original should not exist
				if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
					t.Errorf("original file should not exist")
				}

				// Should exist in trash
				baseName := filepath.Base(originalPath)
				trashPath := filepath.Join(trashDir, "files", baseName)
				if _, err := os.Stat(trashPath); os.IsNotExist(err) {
					t.Errorf("file should exist in trash at %s", trashPath)
				}

				// Trashinfo should exist
				trashinfoPath := filepath.Join(trashDir, "info", baseName+".trashinfo")
				if _, err := os.Stat(trashinfoPath); os.IsNotExist(err) {
					t.Errorf("trashinfo should exist at %s", trashinfoPath)
				}
			},
		},
		{
			name: "directory",
			setup: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "testdir")
				if err := os.MkdirAll(path, 0755); err != nil {
					t.Fatalf("failed to create test dir: %v", err)
				}
				// Create a file inside
				if err := os.WriteFile(filepath.Join(path, "inner.txt"), []byte("inner"), 0644); err != nil {
					t.Fatalf("failed to create inner file: %v", err)
				}
				return path
			},
			verify: func(t *testing.T, originalPath string) {
				// Original should not exist
				if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
					t.Errorf("original dir should not exist")
				}

				// Should exist in trash with contents
				baseName := filepath.Base(originalPath)
				trashPath := filepath.Join(trashDir, "files", baseName)
				innerPath := filepath.Join(trashPath, "inner.txt")
				if _, err := os.Stat(innerPath); os.IsNotExist(err) {
					t.Errorf("inner file should exist in trashed dir at %s", innerPath)
				}
			},
		},
		{
			name: "nonexistent file",
			setup: func(t *testing.T) string {
				return filepath.Join(tmpDir, "nonexistent.txt")
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "symlink moves link not target",
			setup: func(t *testing.T) string {
				// Create target file
				targetPath := filepath.Join(tmpDir, "target.txt")
				if err := os.WriteFile(targetPath, []byte("target"), 0644); err != nil {
					t.Fatalf("failed to create target: %v", err)
				}

				// Create symlink
				linkPath := filepath.Join(tmpDir, "link.txt")
				if err := os.Symlink(targetPath, linkPath); err != nil {
					t.Fatalf("failed to create symlink: %v", err)
				}
				return linkPath
			},
			verify: func(t *testing.T, originalPath string) {
				// Symlink should not exist at original location
				if _, err := os.Lstat(originalPath); !os.IsNotExist(err) {
					t.Errorf("symlink should not exist at original location")
				}

				// Target should still exist
				targetPath := filepath.Join(tmpDir, "target.txt")
				if _, err := os.Stat(targetPath); os.IsNotExist(err) {
					t.Errorf("target file should still exist")
				}

				// Symlink should be in trash (as a symlink)
				baseName := filepath.Base(originalPath)
				trashPath := filepath.Join(trashDir, "files", baseName)
				info, err := os.Lstat(trashPath)
				if err != nil {
					t.Errorf("trashed symlink should exist: %v", err)
				} else if info.Mode()&os.ModeSymlink == 0 {
					t.Errorf("trashed item should be a symlink")
				}
			},
		},
		{
			name: "file with special characters",
			setup: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "file with spaces & special.txt")
				if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return path
			},
			verify: func(t *testing.T, originalPath string) {
				// Should exist in trash
				baseName := filepath.Base(originalPath)
				trashPath := filepath.Join(trashDir, "files", baseName)
				if _, err := os.Stat(trashPath); os.IsNotExist(err) {
					t.Errorf("file should exist in trash")
				}

				// Trashinfo should have URL-encoded path
				trashinfoPath := filepath.Join(trashDir, "info", baseName+".trashinfo")
				content, err := os.ReadFile(trashinfoPath)
				if err != nil {
					t.Fatalf("failed to read trashinfo: %v", err)
				}
				if !strings.Contains(string(content), "%20") {
					t.Errorf("trashinfo should contain URL-encoded path")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure fresh trash directories
			os.RemoveAll(trashDir)
			if err := EnsureTrashDirs(); err != nil {
				t.Fatalf("failed to ensure trash dirs: %v", err)
			}

			path := tt.setup(t)
			err := MoveToTrash(path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("MoveToTrash() expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("MoveToTrash() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("MoveToTrash() unexpected error: %v", err)
				return
			}

			if tt.verify != nil {
				tt.verify(t, path)
			}
		})
	}
}

func TestMoveToTrashNameCollision(t *testing.T) {
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "Trash")

	// Save and restore XDG_DATA_HOME
	origXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	// Ensure trash directories
	if err := EnsureTrashDirs(); err != nil {
		t.Fatalf("failed to ensure trash dirs: %v", err)
	}

	// Create and trash first file
	file1 := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := MoveToTrash(file1); err != nil {
		t.Fatalf("failed to trash file1: %v", err)
	}

	// Create and trash second file with same name
	file2 := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}
	if err := MoveToTrash(file2); err != nil {
		t.Fatalf("failed to trash file2: %v", err)
	}

	// Verify both exist in trash with different names
	trashFiles := filepath.Join(trashDir, "files")
	entries, err := os.ReadDir(trashFiles)
	if err != nil {
		t.Fatalf("failed to read trash files: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 files in trash, got %d", len(entries))
	}

	// Check names
	hasOriginal := false
	hasNumbered := false
	for _, e := range entries {
		if e.Name() == "file.txt" {
			hasOriginal = true
		}
		if e.Name() == "file.2.txt" {
			hasNumbered = true
		}
	}

	if !hasOriginal {
		t.Errorf("expected file.txt in trash")
	}
	if !hasNumbered {
		t.Errorf("expected file.2.txt in trash")
	}
}

func TestMoveToTrashRollbackOnFailure(t *testing.T) {
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "Trash")

	// Save and restore XDG_DATA_HOME
	origXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	// Ensure trash directories
	if err := EnsureTrashDirs(); err != nil {
		t.Fatalf("failed to ensure trash dirs: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Make trash/files/ read-only to cause move failure
	trashFiles := filepath.Join(trashDir, "files")
	os.Chmod(trashFiles, 0555)
	defer os.Chmod(trashFiles, 0755) // Restore for cleanup

	// Attempt to trash the file
	err := MoveToTrash(testFile)

	// Should fail
	if err == nil {
		t.Fatalf("expected MoveToTrash to fail")
	}

	// Original file should still exist
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("original file should still exist after failed trash")
	}

	// No .trashinfo should exist (rollback)
	trashinfoPath := filepath.Join(trashDir, "info", "testfile.txt.trashinfo")
	if _, err := os.Stat(trashinfoPath); !os.IsNotExist(err) {
		t.Errorf("trashinfo should be cleaned up after failed trash")
	}
}

func TestIsInTrash(t *testing.T) {
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "Trash")

	// Save and restore XDG_DATA_HOME
	origXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	// Ensure trash directories
	if err := EnsureTrashDirs(); err != nil {
		t.Fatalf("failed to ensure trash dirs: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "trash files directory",
			path: filepath.Join(trashDir, "files"),
			want: true,
		},
		{
			name: "file in trash",
			path: filepath.Join(trashDir, "files", "somefile.txt"),
			want: true,
		},
		{
			name: "nested in trash",
			path: filepath.Join(trashDir, "files", "dir", "file.txt"),
			want: true,
		},
		{
			name: "outside trash",
			path: filepath.Join(tmpDir, "somefile.txt"),
			want: false,
		},
		{
			name: "trash info directory (not files)",
			path: filepath.Join(trashDir, "info"),
			want: false,
		},
		{
			name: "home directory",
			path: "/home/user/Documents",
			want: false,
		},
		{
			name: "trash root",
			path: trashDir,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInTrash(tt.path)
			if got != tt.want {
				t.Errorf("IsInTrash(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestTrashFilesDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and restore XDG_DATA_HOME
	origXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	got := TrashFilesDir()
	want := filepath.Join(tmpDir, "Trash", "files")
	if got != want {
		t.Errorf("TrashFilesDir() = %q, want %q", got, want)
	}
}

func TestTrashInfoDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and restore XDG_DATA_HOME
	origXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	got := TrashInfoDir()
	want := filepath.Join(tmpDir, "Trash", "info")
	if got != want {
		t.Errorf("TrashInfoDir() = %q, want %q", got, want)
	}
}

func TestRestoreFromTrash(t *testing.T) {
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "Trash")

	// Save and restore XDG_DATA_HOME
	origXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	tests := []struct {
		name        string
		setup       func(t *testing.T) (trashName, originalPath string)
		wantErr     bool
		errContains string
		verify      func(t *testing.T, originalPath string)
	}{
		{
			name: "restore single file",
			setup: func(t *testing.T) (string, string) {
				// Create and trash a file
				originalPath := filepath.Join(tmpDir, "testfile.txt")
				if err := os.WriteFile(originalPath, []byte("test content"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				if err := MoveToTrash(originalPath); err != nil {
					t.Fatalf("failed to trash file: %v", err)
				}
				return "testfile.txt", originalPath
			},
			verify: func(t *testing.T, originalPath string) {
				// File should be restored
				content, err := os.ReadFile(originalPath)
				if err != nil {
					t.Errorf("restored file should exist: %v", err)
				}
				if string(content) != "test content" {
					t.Errorf("restored file content mismatch")
				}

				// Trashinfo should be deleted
				trashinfoPath := filepath.Join(trashDir, "info", "testfile.txt.trashinfo")
				if _, err := os.Stat(trashinfoPath); !os.IsNotExist(err) {
					t.Errorf("trashinfo should be deleted after restore")
				}

				// Trash files entry should be deleted
				trashFilePath := filepath.Join(trashDir, "files", "testfile.txt")
				if _, err := os.Stat(trashFilePath); !os.IsNotExist(err) {
					t.Errorf("trash file should be deleted after restore")
				}
			},
		},
		{
			name: "restore directory",
			setup: func(t *testing.T) (string, string) {
				// Create and trash a directory
				originalPath := filepath.Join(tmpDir, "testdir")
				if err := os.MkdirAll(originalPath, 0755); err != nil {
					t.Fatalf("failed to create test dir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(originalPath, "inner.txt"), []byte("inner"), 0644); err != nil {
					t.Fatalf("failed to create inner file: %v", err)
				}
				if err := MoveToTrash(originalPath); err != nil {
					t.Fatalf("failed to trash dir: %v", err)
				}
				return "testdir", originalPath
			},
			verify: func(t *testing.T, originalPath string) {
				// Directory should be restored with contents
				innerPath := filepath.Join(originalPath, "inner.txt")
				content, err := os.ReadFile(innerPath)
				if err != nil {
					t.Errorf("restored inner file should exist: %v", err)
				}
				if string(content) != "inner" {
					t.Errorf("restored inner file content mismatch")
				}
			},
		},
		{
			name: "restore to deleted parent directory",
			setup: func(t *testing.T) (string, string) {
				// Create parent dir and file
				parentDir := filepath.Join(tmpDir, "parent")
				if err := os.MkdirAll(parentDir, 0755); err != nil {
					t.Fatalf("failed to create parent dir: %v", err)
				}
				originalPath := filepath.Join(parentDir, "file.txt")
				if err := os.WriteFile(originalPath, []byte("content"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				if err := MoveToTrash(originalPath); err != nil {
					t.Fatalf("failed to trash file: %v", err)
				}
				// Delete parent directory
				if err := os.RemoveAll(parentDir); err != nil {
					t.Fatalf("failed to remove parent: %v", err)
				}
				return "file.txt", originalPath
			},
			verify: func(t *testing.T, originalPath string) {
				// Parent directory should be recreated and file restored
				content, err := os.ReadFile(originalPath)
				if err != nil {
					t.Errorf("restored file should exist: %v", err)
				}
				if string(content) != "content" {
					t.Errorf("restored file content mismatch")
				}
			},
		},
		{
			name: "missing trashinfo",
			setup: func(t *testing.T) (string, string) {
				// Create file in trash without trashinfo
				trashFilesDir := filepath.Join(trashDir, "files")
				os.MkdirAll(trashFilesDir, 0700)
				trashPath := filepath.Join(trashFilesDir, "orphan.txt")
				if err := os.WriteFile(trashPath, []byte("orphan"), 0644); err != nil {
					t.Fatalf("failed to create orphan: %v", err)
				}
				return "orphan.txt", ""
			},
			wantErr:     true,
			errContains: "trashinfo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure fresh trash directories
			os.RemoveAll(trashDir)
			if err := EnsureTrashDirs(); err != nil {
				t.Fatalf("failed to ensure trash dirs: %v", err)
			}

			trashName, originalPath := tt.setup(t)
			err := RestoreFromTrash(trashName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("RestoreFromTrash() expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("RestoreFromTrash() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("RestoreFromTrash() unexpected error: %v", err)
				return
			}

			if tt.verify != nil {
				tt.verify(t, originalPath)
			}
		})
	}
}

func TestEmptyTrash(t *testing.T) {
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "Trash")

	// Save and restore XDG_DATA_HOME
	origXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	// Ensure trash directories
	if err := EnsureTrashDirs(); err != nil {
		t.Fatalf("failed to ensure trash dirs: %v", err)
	}

	// Create and trash some files
	for i := 0; i < 3; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", i))
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		if err := MoveToTrash(path); err != nil {
			t.Fatalf("failed to trash file: %v", err)
		}
	}

	// Verify files are in trash
	filesDir := filepath.Join(trashDir, "files")
	entries, _ := os.ReadDir(filesDir)
	if len(entries) != 3 {
		t.Fatalf("expected 3 files in trash, got %d", len(entries))
	}

	// Empty trash
	err := EmptyTrash()
	if err != nil {
		t.Errorf("EmptyTrash() error = %v", err)
	}

	// Verify trash is empty
	entries, _ = os.ReadDir(filesDir)
	if len(entries) != 0 {
		t.Errorf("expected empty trash/files, got %d entries", len(entries))
	}

	infoDir := filepath.Join(trashDir, "info")
	entries, _ = os.ReadDir(infoDir)
	if len(entries) != 0 {
		t.Errorf("expected empty trash/info, got %d entries", len(entries))
	}
}

func TestEmptyTrashEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and restore XDG_DATA_HOME
	origXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	// Ensure trash directories
	if err := EnsureTrashDirs(); err != nil {
		t.Fatalf("failed to ensure trash dirs: %v", err)
	}

	// Empty an already empty trash should not error
	err := EmptyTrash()
	if err != nil {
		t.Errorf("EmptyTrash() on empty trash should not error: %v", err)
	}
}

func TestGetTrashItemInfo(t *testing.T) {
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "Trash")

	// Save and restore XDG_DATA_HOME
	origXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	// Ensure trash directories
	if err := EnsureTrashDirs(); err != nil {
		t.Fatalf("failed to ensure trash dirs: %v", err)
	}

	// Create and trash a file
	originalPath := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(originalPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := MoveToTrash(originalPath); err != nil {
		t.Fatalf("failed to trash file: %v", err)
	}

	// Get trash item info
	info, err := GetTrashItemInfo("testfile.txt")
	if err != nil {
		t.Fatalf("GetTrashItemInfo() error = %v", err)
	}

	if info.OriginalPath != originalPath {
		t.Errorf("OriginalPath = %q, want %q", info.OriginalPath, originalPath)
	}

	// DeletionDate should be recent
	if info.DeletionDate.IsZero() {
		t.Errorf("DeletionDate should not be zero")
	}

	// Test with nonexistent item
	_, err = GetTrashItemInfo("nonexistent.txt")
	if err == nil {
		t.Errorf("GetTrashItemInfo() should error for nonexistent item")
	}

	_ = trashDir // suppress unused variable warning
}

func TestIsInTrashPathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and restore XDG_DATA_HOME
	origXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	// Ensure trash directories
	if err := EnsureTrashDirs(); err != nil {
		t.Fatalf("failed to ensure trash dirs: %v", err)
	}

	trashFilesDir := TrashFilesDir()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "path traversal with ..",
			path: filepath.Join(trashFilesDir, "..", "info"),
			want: false,
		},
		{
			name: "double path traversal",
			path: filepath.Join(trashFilesDir, "..", "..", "etc"),
			want: false,
		},
		{
			name: "path with .. in middle",
			path: filepath.Join(trashFilesDir, "subdir", "..", "file.txt"),
			want: true, // After normalization, this is still in trash
		},
		{
			name: "valid nested path",
			path: filepath.Join(trashFilesDir, "deep", "nested", "file.txt"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInTrash(tt.path)
			if got != tt.want {
				t.Errorf("IsInTrash(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestResolveNameCollisionLimit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	baseName := "test.txt"
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	// Create enough files to test the limit behavior
	for i := 0; i <= 10; i++ {
		var name string
		if i == 0 {
			name = baseName
		} else {
			name = fmt.Sprintf("%s.%d%s", nameWithoutExt, i+1, ext)
		}
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte{}, 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	// ResolveNameCollision should find the next available number
	result := ResolveNameCollision(tmpDir, baseName)
	expected := "test.12.txt"
	if result != expected {
		t.Errorf("ResolveNameCollision() = %q, want %q", result, expected)
	}
}
