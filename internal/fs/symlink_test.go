package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetSymlinkInfo(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "test_symlink_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// テスト用のファイルを作成
	targetFile := filepath.Join(tmpDir, "target.txt")
	if err := os.WriteFile(targetFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// テスト用のディレクトリを作成
	targetDir := filepath.Join(tmpDir, "target_dir")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	// ファイルへのシンボリックリンクを作成
	linkToFile := filepath.Join(tmpDir, "link_to_file")
	if err := os.Symlink(targetFile, linkToFile); err != nil {
		t.Fatalf("Failed to create symlink to file: %v", err)
	}

	// ディレクトリへのシンボリックリンクを作成
	linkToDir := filepath.Join(tmpDir, "link_to_dir")
	if err := os.Symlink(targetDir, linkToDir); err != nil {
		t.Fatalf("Failed to create symlink to directory: %v", err)
	}

	// 存在しないファイルへのシンボリックリンクを作成（リンク切れ）
	brokenLink := filepath.Join(tmpDir, "broken_link")
	if err := os.Symlink("/nonexistent/path", brokenLink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	tests := []struct {
		name          string
		path          string
		wantIsLink    bool
		wantBroken    bool
		wantTargetDir bool
		wantErr       bool
	}{
		{
			name:          "regular file",
			path:          targetFile,
			wantIsLink:    false,
			wantBroken:    false,
			wantTargetDir: false,
			wantErr:       false,
		},
		{
			name:          "symlink to file",
			path:          linkToFile,
			wantIsLink:    true,
			wantBroken:    false,
			wantTargetDir: false,
			wantErr:       false,
		},
		{
			name:          "symlink to directory",
			path:          linkToDir,
			wantIsLink:    true,
			wantBroken:    false,
			wantTargetDir: true,
			wantErr:       false,
		},
		{
			name:          "broken symlink",
			path:          brokenLink,
			wantIsLink:    true,
			wantBroken:    true,
			wantTargetDir: false,
			wantErr:       false,
		},
		{
			name:          "non-existing file",
			path:          "/nonexistent/file",
			wantIsLink:    false,
			wantBroken:    false,
			wantTargetDir: false,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isLink, target, isBroken, isTargetDir, err := GetSymlinkInfo(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSymlinkInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if isLink != tt.wantIsLink {
				t.Errorf("GetSymlinkInfo() isLink = %v, want %v", isLink, tt.wantIsLink)
			}
			if isBroken != tt.wantBroken {
				t.Errorf("GetSymlinkInfo() isBroken = %v, want %v", isBroken, tt.wantBroken)
			}
			if isTargetDir != tt.wantTargetDir {
				t.Errorf("GetSymlinkInfo() isTargetDir = %v, want %v", isTargetDir, tt.wantTargetDir)
			}

			if tt.wantIsLink && !tt.wantBroken {
				if target == "" {
					t.Errorf("GetSymlinkInfo() target is empty for valid symlink")
				}
				t.Logf("Target: %s", target)
			}
		})
	}
}
