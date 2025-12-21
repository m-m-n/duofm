package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDirectory(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(string) error
		wantMinEntries int
		wantErr        bool
	}{
		{
			name: "空のディレクトリを読み込む",
			setupFunc: func(tmpDir string) error {
				return nil
			},
			wantMinEntries: 1, // 親ディレクトリのみ
			wantErr:        false,
		},
		{
			name: "ファイルとディレクトリを含むディレクトリを読み込む",
			setupFunc: func(tmpDir string) error {
				if err := os.Mkdir(filepath.Join(tmpDir, "dir1"), 0755); err != nil {
					return err
				}
				if err := os.Mkdir(filepath.Join(tmpDir, "dir2"), 0755); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644); err != nil {
					return err
				}
				return nil
			},
			wantMinEntries: 4, // 親ディレクトリ + 2 dirs + 1 file
			wantErr:        false,
		},
		{
			name: "存在しないディレクトリを読み込む",
			setupFunc: func(tmpDir string) error {
				return nil
			},
			wantMinEntries: 0,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if err := tt.setupFunc(tmpDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			testPath := tmpDir
			if tt.wantErr {
				testPath = filepath.Join(tmpDir, "nonexistent")
			}

			entries, err := ReadDirectory(testPath)

			if tt.wantErr {
				if err == nil {
					t.Error("ReadDirectory() should return error for nonexistent directory")
				}
				return
			}

			if err != nil {
				t.Fatalf("ReadDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(entries) < tt.wantMinEntries {
				t.Errorf("ReadDirectory() returned %d entries, want at least %d", len(entries), tt.wantMinEntries)
			}

			// 親ディレクトリが最初にあることを確認
			if len(entries) > 0 && !entries[0].IsParentDir() {
				t.Error("First entry should be parent directory")
			}
		})
	}
}

func TestReadDirectoryRootPath(t *testing.T) {
	// ルートディレクトリは親ディレクトリエントリを含まない
	entries, err := ReadDirectory("/")
	if err != nil {
		t.Fatalf("ReadDirectory('/') failed: %v", err)
	}

	if len(entries) == 0 {
		t.Error("Root directory should have entries")
	}

	// ルートの場合は親ディレクトリエントリがないはず
	if len(entries) > 0 && entries[0].IsParentDir() {
		t.Error("Root directory should not have parent directory entry")
	}
}

func TestHomeDirectory(t *testing.T) {
	home, err := HomeDirectory()
	if err != nil {
		t.Fatalf("HomeDirectory() failed: %v", err)
	}

	if home == "" {
		t.Error("HomeDirectory() should return non-empty path")
	}

	// ホームディレクトリが存在することを確認
	if _, err := os.Stat(home); err != nil {
		t.Errorf("Home directory does not exist: %v", err)
	}
}

func TestCurrentDirectory(t *testing.T) {
	cwd, err := CurrentDirectory()
	if err != nil {
		t.Fatalf("CurrentDirectory() failed: %v", err)
	}

	if cwd == "" {
		t.Error("CurrentDirectory() should return non-empty path")
	}

	// カレントディレクトリが存在することを確認
	if _, err := os.Stat(cwd); err != nil {
		t.Errorf("Current directory does not exist: %v", err)
	}
}

func TestFileEntryProperties(t *testing.T) {
	tmpDir := t.TempDir()

	// テストファイルとディレクトリを作成
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	entries, err := ReadDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ReadDirectory() failed: %v", err)
	}

	// ファイルとディレクトリのプロパティを確認
	foundFile := false
	foundDir := false

	for _, entry := range entries {
		if entry.IsParentDir() {
			continue
		}

		if entry.Name == "test.txt" {
			foundFile = true
			if entry.IsDir {
				t.Error("test.txt should not be a directory")
			}
			if entry.Size != int64(len(testContent)) {
				t.Errorf("test.txt size = %d, want %d", entry.Size, len(testContent))
			}
		}

		if entry.Name == "testdir" {
			foundDir = true
			if !entry.IsDir {
				t.Error("testdir should be a directory")
			}
		}
	}

	if !foundFile {
		t.Error("test.txt not found in entries")
	}

	if !foundDir {
		t.Error("testdir not found in entries")
	}
}

func TestReadDirectory_ParentDirMetadata(t *testing.T) {
	// Create a subdirectory to test parent metadata
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	entries, err := ReadDirectory(subDir)
	if err != nil {
		t.Fatalf("ReadDirectory() failed: %v", err)
	}

	// Find parent directory entry
	var parentEntry *FileEntry
	for i := range entries {
		if entries[i].IsParentDir() {
			parentEntry = &entries[i]
			break
		}
	}

	if parentEntry == nil {
		t.Fatal("Parent directory entry not found")
	}

	// Verify ModTime is not zero value
	if parentEntry.ModTime.IsZero() {
		t.Error("Parent directory ModTime should not be zero")
	}

	// Verify Permissions is set
	if parentEntry.Permissions == 0 {
		t.Error("Parent directory Permissions should be set")
	}

	// Verify Owner/Group are populated
	if parentEntry.Owner == "" {
		t.Error("Parent directory Owner should not be empty")
	}
	if parentEntry.Group == "" {
		t.Error("Parent directory Group should not be empty")
	}
}
