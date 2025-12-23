package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CopyFile はファイルをコピー
func CopyFile(src, dst string) error {
	// ソースファイルを開く
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// ソースの情報を取得
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// 宛先パスを決定
	dstPath := dst
	if dstInfo, err := os.Stat(dst); err == nil && dstInfo.IsDir() {
		dstPath = filepath.Join(dst, filepath.Base(src))
	}

	// 宛先ファイルを作成
	destFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// コピー実行
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// CopyDirectory はディレクトリを再帰的にコピー
func CopyDirectory(src, dst string) error {
	// ソースディレクトリの情報を取得
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	// 宛先パスを決定
	dstPath := dst
	if dstInfo, err := os.Stat(dst); err == nil && dstInfo.IsDir() {
		dstPath = filepath.Join(dst, filepath.Base(src))
	}

	// 宛先ディレクトリを作成
	if err := os.MkdirAll(dstPath, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// ディレクトリの内容を読み込む
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// 各エントリを再帰的にコピー
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dstPath, entry.Name())

		if entry.IsDir() {
			if err := CopyDirectory(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// Copy はファイルまたはディレクトリをコピー
func Copy(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source not found: %w", err)
	}

	if srcInfo.IsDir() {
		return CopyDirectory(src, dst)
	}
	return CopyFile(src, dst)
}

// MoveFile はファイルを移動
func MoveFile(src, dst string) error {
	// 宛先パスを決定
	dstPath := dst
	if dstInfo, err := os.Stat(dst); err == nil && dstInfo.IsDir() {
		dstPath = filepath.Join(dst, filepath.Base(src))
	}

	// os.Rename を試す（同一ファイルシステム内）
	err := os.Rename(src, dstPath)
	if err == nil {
		return nil
	}

	// クロスデバイス移動の場合はコピー→削除
	if err := Copy(src, dst); err != nil {
		return fmt.Errorf("failed to copy during move: %w", err)
	}

	if err := Delete(src); err != nil {
		return fmt.Errorf("failed to delete source after copy: %w", err)
	}

	return nil
}

// DeleteFile はファイルを削除
func DeleteFile(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// DeleteDirectory はディレクトリを再帰的に削除
func DeleteDirectory(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to delete directory: %w", err)
	}
	return nil
}

// Delete はファイルまたはディレクトリを削除
func Delete(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	if info.IsDir() {
		return DeleteDirectory(path)
	}
	return DeleteFile(path)
}

// IsDirectory は指定されたパスがディレクトリかどうかを判定
func IsDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// CreateFile は空のファイルを作成
func CreateFile(path string) error {
	// ファイルが既に存在するか確認
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists: %s", filepath.Base(path))
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	return file.Close()
}

// CreateDirectory は新しいディレクトリを作成
func CreateDirectory(path string) error {
	// ファイルまたはディレクトリが既に存在するか確認
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("directory already exists: %s", filepath.Base(path))
	}

	if err := os.Mkdir(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// Rename はファイルまたはディレクトリを同一ディレクトリ内でリネーム
func Rename(oldPath, newName string) error {
	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, newName)

	// リネーム先が既に存在するか確認
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("file already exists: %s", newName)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}
	return nil
}

// ValidateFilename はファイル名が有効かどうかを検証
func ValidateFilename(name string) error {
	if name == "" {
		return fmt.Errorf("file name cannot be empty")
	}
	if strings.Contains(name, "/") {
		return fmt.Errorf("invalid file name: path separator not allowed")
	}
	return nil
}
