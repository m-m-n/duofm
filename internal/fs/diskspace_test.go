package fs

import (
	"os"
	"testing"
)

func TestGetDiskSpace(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "test_diskspace_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "temp directory",
			path:    tmpDir,
			wantErr: false,
		},
		{
			name:    "root directory",
			path:    "/",
			wantErr: false,
		},
		{
			name:    "home directory",
			path:    os.Getenv("HOME"),
			wantErr: false,
		},
		{
			name:    "non-existing directory",
			path:    "/nonexistent/directory/path",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			freeBytes, totalBytes, err := GetDiskSpace(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDiskSpace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if freeBytes == 0 {
					t.Errorf("GetDiskSpace() freeBytes = 0, expected > 0")
				}
				if totalBytes == 0 {
					t.Errorf("GetDiskSpace() totalBytes = 0, expected > 0")
				}
				if freeBytes > totalBytes {
					t.Errorf("GetDiskSpace() freeBytes (%d) > totalBytes (%d)", freeBytes, totalBytes)
				}
				t.Logf("Path: %s, Free: %d bytes, Total: %d bytes", tt.path, freeBytes, totalBytes)
			}
		})
	}
}

func TestGetDiskSpaceForFile(t *testing.T) {
	// テスト用の一時ファイルを作成
	tmpFile, err := os.CreateTemp("", "test_diskspace_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// ファイルのパスに対してもディスク容量を取得できることを確認
	freeBytes, totalBytes, err := GetDiskSpace(tmpFile.Name())
	if err != nil {
		t.Errorf("GetDiskSpace() error = %v for file path", err)
		return
	}

	if freeBytes == 0 || totalBytes == 0 {
		t.Errorf("GetDiskSpace() returned zero values for file path")
	}
	t.Logf("File path - Free: %d bytes, Total: %d bytes", freeBytes, totalBytes)
}
