package fs

import (
	"os"
	"testing"
)

func TestGetFileOwnerGroup(t *testing.T) {
	// テスト用の一時ファイルを作成
	tmpFile, err := os.CreateTemp("", "test_owner_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "existing file",
			path:    tmpFile.Name(),
			wantErr: false,
		},
		{
			name:    "non-existing file",
			path:    "/nonexistent/file/path",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, group, err := GetFileOwnerGroup(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileOwnerGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if owner == "" {
					t.Errorf("GetFileOwnerGroup() owner is empty")
				}
				if group == "" {
					t.Errorf("GetFileOwnerGroup() group is empty")
				}
				t.Logf("Owner: %s, Group: %s", owner, group)
			}
		})
	}
}

func TestGetFileOwnerGroupDirectory(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "test_owner_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	owner, group, err := GetFileOwnerGroup(tmpDir)
	if err != nil {
		t.Errorf("GetFileOwnerGroup() error = %v", err)
		return
	}

	if owner == "" {
		t.Errorf("GetFileOwnerGroup() owner is empty for directory")
	}
	if group == "" {
		t.Errorf("GetFileOwnerGroup() group is empty for directory")
	}
	t.Logf("Directory - Owner: %s, Group: %s", owner, group)
}
