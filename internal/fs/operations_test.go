package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(string) (string, string)
		wantErr bool
	}{
		{
			name: "通常のファイルコピー",
			setup: func(tmpDir string) (string, string) {
				srcFile := filepath.Join(tmpDir, "source.txt")
				dstDir := filepath.Join(tmpDir, "dest")
				os.WriteFile(srcFile, []byte("test content"), 0644)
				os.Mkdir(dstDir, 0755)
				return srcFile, dstDir
			},
			wantErr: false,
		},
		{
			name: "存在しないファイルのコピー",
			setup: func(tmpDir string) (string, string) {
				srcFile := filepath.Join(tmpDir, "nonexistent.txt")
				dstDir := filepath.Join(tmpDir, "dest")
				os.Mkdir(dstDir, 0755)
				return srcFile, dstDir
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			src, dst := tt.setup(tmpDir)

			err := CopyFile(src, dst)

			if tt.wantErr {
				if err == nil {
					t.Error("CopyFile() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("CopyFile() error = %v", err)
			}

			// コピー先ファイルの存在確認
			dstFile := filepath.Join(dst, filepath.Base(src))
			if _, err := os.Stat(dstFile); err != nil {
				t.Errorf("Destination file should exist: %v", err)
			}

			// 内容の確認
			srcContent, _ := os.ReadFile(src)
			dstContent, _ := os.ReadFile(dstFile)
			if string(srcContent) != string(dstContent) {
				t.Error("File content should match")
			}
		})
	}
}

func TestCopyDirectory(t *testing.T) {
	t.Run("ディレクトリの再帰的コピー", func(t *testing.T) {
		tmpDir := t.TempDir()

		// ソースディレクトリ構造を作成
		srcDir := filepath.Join(tmpDir, "source")
		os.Mkdir(srcDir, 0755)
		os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644)
		os.Mkdir(filepath.Join(srcDir, "subdir"), 0755)
		os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0644)

		// コピー先ディレクトリ（既存）
		dstDir := filepath.Join(tmpDir, "dest")
		os.Mkdir(dstDir, 0755)

		err := CopyDirectory(srcDir, dstDir)
		if err != nil {
			t.Fatalf("CopyDirectory() error = %v", err)
		}

		// 検証（destディレクトリ内にsourceディレクトリがコピーされる）
		expectedDst := filepath.Join(dstDir, "source")
		if _, err := os.Stat(filepath.Join(expectedDst, "file1.txt")); err != nil {
			t.Error("file1.txt should be copied")
		}
		if _, err := os.Stat(filepath.Join(expectedDst, "subdir", "file2.txt")); err != nil {
			t.Error("subdir/file2.txt should be copied")
		}
	})

	t.Run("空のディレクトリのコピー", func(t *testing.T) {
		tmpDir := t.TempDir()

		srcDir := filepath.Join(tmpDir, "empty")
		os.Mkdir(srcDir, 0755)

		dstDir := filepath.Join(tmpDir, "dest")
		os.Mkdir(dstDir, 0755)

		err := CopyDirectory(srcDir, dstDir)
		if err != nil {
			t.Fatalf("CopyDirectory() error = %v", err)
		}

		expectedDst := filepath.Join(dstDir, "empty")
		if _, err := os.Stat(expectedDst); err != nil {
			t.Error("Empty directory should be copied")
		}
	})
}

func TestCopy(t *testing.T) {
	t.Run("ファイルのコピー", func(t *testing.T) {
		tmpDir := t.TempDir()

		srcFile := filepath.Join(tmpDir, "file.txt")
		os.WriteFile(srcFile, []byte("test"), 0644)

		dstDir := filepath.Join(tmpDir, "dest")
		os.Mkdir(dstDir, 0755)

		err := Copy(srcFile, dstDir)
		if err != nil {
			t.Fatalf("Copy() error = %v", err)
		}

		dstFile := filepath.Join(dstDir, "file.txt")
		if _, err := os.Stat(dstFile); err != nil {
			t.Error("File should be copied")
		}
	})

	t.Run("ディレクトリのコピー", func(t *testing.T) {
		tmpDir := t.TempDir()

		srcDir := filepath.Join(tmpDir, "dir")
		os.Mkdir(srcDir, 0755)
		os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("test"), 0644)

		dstDir := filepath.Join(tmpDir, "dest")
		os.Mkdir(dstDir, 0755)

		err := Copy(srcDir, dstDir)
		if err != nil {
			t.Fatalf("Copy() error = %v", err)
		}

		expectedDst := filepath.Join(dstDir, "dir")
		if _, err := os.Stat(filepath.Join(expectedDst, "file.txt")); err != nil {
			t.Error("Directory should be copied")
		}
	})
}

func TestMoveFile(t *testing.T) {
	t.Run("同一ファイルシステム内の移動", func(t *testing.T) {
		tmpDir := t.TempDir()

		srcFile := filepath.Join(tmpDir, "source.txt")
		os.WriteFile(srcFile, []byte("test"), 0644)

		dstDir := filepath.Join(tmpDir, "dest")
		os.Mkdir(dstDir, 0755)

		err := MoveFile(srcFile, dstDir)
		if err != nil {
			t.Fatalf("MoveFile() error = %v", err)
		}

		// ソースファイルが存在しないことを確認
		if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
			t.Error("Source file should not exist after move")
		}

		// 宛先ファイルが存在することを確認
		dstFile := filepath.Join(dstDir, "source.txt")
		if _, err := os.Stat(dstFile); err != nil {
			t.Error("Destination file should exist")
		}
	})

	t.Run("存在しないファイルの移動", func(t *testing.T) {
		tmpDir := t.TempDir()

		srcFile := filepath.Join(tmpDir, "nonexistent.txt")
		dstDir := filepath.Join(tmpDir, "dest")
		os.Mkdir(dstDir, 0755)

		err := MoveFile(srcFile, dstDir)
		if err == nil {
			t.Error("MoveFile() should return error for nonexistent file")
		}
	})
}

func TestDeleteFile(t *testing.T) {
	t.Run("ファイルの削除", func(t *testing.T) {
		tmpDir := t.TempDir()

		testFile := filepath.Join(tmpDir, "test.txt")
		os.WriteFile(testFile, []byte("test"), 0644)

		err := DeleteFile(testFile)
		if err != nil {
			t.Fatalf("DeleteFile() error = %v", err)
		}

		// ファイルが存在しないことを確認
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("File should be deleted")
		}
	})

	t.Run("存在しないファイルの削除", func(t *testing.T) {
		tmpDir := t.TempDir()

		testFile := filepath.Join(tmpDir, "nonexistent.txt")

		err := DeleteFile(testFile)
		if err == nil {
			t.Error("DeleteFile() should return error for nonexistent file")
		}
	})
}

func TestDeleteDirectory(t *testing.T) {
	t.Run("空のディレクトリの削除", func(t *testing.T) {
		tmpDir := t.TempDir()

		testDir := filepath.Join(tmpDir, "testdir")
		os.Mkdir(testDir, 0755)

		err := DeleteDirectory(testDir)
		if err != nil {
			t.Fatalf("DeleteDirectory() error = %v", err)
		}

		if _, err := os.Stat(testDir); !os.IsNotExist(err) {
			t.Error("Directory should be deleted")
		}
	})

	t.Run("ファイルを含むディレクトリの削除", func(t *testing.T) {
		tmpDir := t.TempDir()

		testDir := filepath.Join(tmpDir, "testdir")
		os.Mkdir(testDir, 0755)
		os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("test"), 0644)
		os.Mkdir(filepath.Join(testDir, "subdir"), 0755)

		err := DeleteDirectory(testDir)
		if err != nil {
			t.Fatalf("DeleteDirectory() error = %v", err)
		}

		if _, err := os.Stat(testDir); !os.IsNotExist(err) {
			t.Error("Directory should be deleted")
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("ファイルの削除", func(t *testing.T) {
		tmpDir := t.TempDir()

		testFile := filepath.Join(tmpDir, "test.txt")
		os.WriteFile(testFile, []byte("test"), 0644)

		err := Delete(testFile)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("File should be deleted")
		}
	})

	t.Run("ディレクトリの削除", func(t *testing.T) {
		tmpDir := t.TempDir()

		testDir := filepath.Join(tmpDir, "testdir")
		os.Mkdir(testDir, 0755)
		os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("test"), 0644)

		err := Delete(testDir)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		if _, err := os.Stat(testDir); !os.IsNotExist(err) {
			t.Error("Directory should be deleted")
		}
	})

	t.Run("存在しないパスの削除", func(t *testing.T) {
		tmpDir := t.TempDir()

		nonexistent := filepath.Join(tmpDir, "nonexistent")

		err := Delete(nonexistent)
		if err == nil {
			t.Error("Delete() should return error for nonexistent path")
		}
	})
}

func TestCopyFilePermissions(t *testing.T) {
	t.Run("ファイルのパーミッション保持", func(t *testing.T) {
		tmpDir := t.TempDir()

		srcFile := filepath.Join(tmpDir, "source.txt")
		os.WriteFile(srcFile, []byte("test"), 0600)

		dstDir := filepath.Join(tmpDir, "dest")
		os.Mkdir(dstDir, 0755)

		err := CopyFile(srcFile, dstDir)
		if err != nil {
			t.Fatalf("CopyFile() error = %v", err)
		}

		// パーミッションの確認
		dstFile := filepath.Join(dstDir, "source.txt")
		dstInfo, _ := os.Stat(dstFile)
		srcInfo, _ := os.Stat(srcFile)

		if dstInfo.Mode() != srcInfo.Mode() {
			t.Error("File permissions should be preserved")
		}
	})
}
