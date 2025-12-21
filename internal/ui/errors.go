package ui

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// formatDirectoryError はディレクトリアクセスエラーをユーザー向けメッセージにフォーマット
func formatDirectoryError(err error, path string) string {
	if err == nil {
		return ""
	}

	// syscallエラーを検出
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		var errno syscall.Errno
		if errors.As(pathErr.Err, &errno) {
			switch errno {
			case syscall.EACCES:
				return fmt.Sprintf("Permission denied: %s", path)
			case syscall.ENOENT:
				return fmt.Sprintf("No such directory: %s", path)
			case syscall.EIO:
				return fmt.Sprintf("I/O error: %s", path)
			}
		}
	}

	// os.IsNotExist でも判定
	if os.IsNotExist(err) {
		return fmt.Sprintf("No such directory: %s", path)
	}

	if os.IsPermission(err) {
		return fmt.Sprintf("Permission denied: %s", path)
	}

	// その他のエラー
	return fmt.Sprintf("Cannot access: %s", path)
}
