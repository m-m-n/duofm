package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReadDirectory はディレクトリの内容を読み込む
func ReadDirectory(path string) ([]FileEntry, error) {
	// パスの正規化
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// ディレクトリの読み込み
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// FileEntry に変換
	var fileEntries []FileEntry

	// 親ディレクトリエントリを追加（ルートディレクトリ以外）
	if absPath != "/" {
		parentPath := filepath.Dir(absPath)
		parentEntry := FileEntry{
			Name:  "..",
			IsDir: true,
		}

		// 親ディレクトリの情報を取得
		if info, err := os.Stat(parentPath); err == nil {
			parentEntry.ModTime = info.ModTime()
			parentEntry.Permissions = info.Mode()
		}

		// 所有者・グループ情報を取得
		if owner, group, err := GetFileOwnerGroup(parentPath); err == nil {
			parentEntry.Owner = owner
			parentEntry.Group = group
		} else {
			parentEntry.Owner = "unknown"
			parentEntry.Group = "unknown"
		}

		fileEntries = append(fileEntries, parentEntry)
	}

	// 各エントリを処理
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // エラーは無視して次へ
		}

		entryPath := filepath.Join(absPath, entry.Name())

		// 基本情報
		fileEntry := FileEntry{
			Name:        entry.Name(),
			IsDir:       entry.IsDir(),
			Size:        info.Size(),
			ModTime:     info.ModTime(),
			Permissions: info.Mode(),
		}

		// 所有者・グループ情報を取得
		owner, group, err := GetFileOwnerGroup(entryPath)
		if err == nil {
			fileEntry.Owner = owner
			fileEntry.Group = group
		} else {
			fileEntry.Owner = "unknown"
			fileEntry.Group = "unknown"
		}

		// シンボリックリンク情報を取得
		isLink, target, isBroken, isTargetDir, err := GetSymlinkInfo(entryPath)
		if err == nil {
			fileEntry.IsSymlink = isLink
			fileEntry.LinkTarget = target
			fileEntry.LinkBroken = isBroken
			// シンボリックリンクがディレクトリを指している場合、IsDirを更新
			if isLink && !isBroken && isTargetDir {
				fileEntry.IsDir = true
			}
		}

		fileEntries = append(fileEntries, fileEntry)
	}

	return fileEntries, nil
}

// HomeDirectory はホームディレクトリのパスを返す
func HomeDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return home, nil
}

// CurrentDirectory は現在の作業ディレクトリを返す
func CurrentDirectory() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return cwd, nil
}
