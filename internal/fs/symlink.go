package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetSymlinkInfo はシンボリックリンクの情報を取得する
// 返り値:
//   - isSymlink: シンボリックリンクかどうか
//   - target: リンク先のパス（シンボリックリンクの場合）
//   - isBroken: リンク切れかどうか（シンボリックリンクの場合）
//   - isTargetDir: リンク先がディレクトリかどうか（シンボリックリンクの場合）
//   - err: エラー
func GetSymlinkInfo(path string) (isSymlink bool, target string, isBroken bool, isTargetDir bool, err error) {
	// ファイル情報を取得（シンボリックリンクを追跡しない）
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false, "", false, false, fmt.Errorf("failed to get file info: %w", err)
	}

	// シンボリックリンクでない場合
	if fileInfo.Mode()&os.ModeSymlink == 0 {
		return false, "", false, false, nil
	}

	// シンボリックリンクの場合、リンク先を取得
	linkTarget, err := os.Readlink(path)
	if err != nil {
		return true, "", false, false, fmt.Errorf("failed to read symlink: %w", err)
	}

	// 相対パスの場合は絶対パスに変換
	if !filepath.IsAbs(linkTarget) {
		dir := filepath.Dir(path)
		linkTarget = filepath.Join(dir, linkTarget)
	}

	// リンク先が存在するか確認
	targetInfo, err := os.Stat(linkTarget)
	if err != nil {
		// リンク切れ
		return true, linkTarget, true, false, nil
	}

	// リンク先がディレクトリかどうか確認
	isTargetDir = targetInfo.IsDir()

	return true, linkTarget, false, isTargetDir, nil
}
