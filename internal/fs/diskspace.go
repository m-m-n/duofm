package fs

import (
	"fmt"
	"runtime"
	"syscall"
)

// GetDiskSpace は指定されたパスを含むパーティションの空き容量と総容量を返す
// 返り値:
//   - freeBytes: 空き容量（バイト）
//   - totalBytes: 総容量（バイト）
//   - err: エラー
func GetDiskSpace(path string) (freeBytes, totalBytes uint64, err error) {
	switch runtime.GOOS {
	case "windows":
		return getDiskSpaceWindows(path)
	default:
		// Unix/Linux/macOS
		return getDiskSpaceUnix(path)
	}
}

// getDiskSpaceUnix はUnix系OSでディスク容量を取得
func getDiskSpaceUnix(path string) (freeBytes, totalBytes uint64, err error) {
	var stat syscall.Statfs_t
	err = syscall.Statfs(path, &stat)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get disk space: %w", err)
	}

	// 利用可能な空き容量（非特権ユーザーが使用可能）
	freeBytes = stat.Bavail * uint64(stat.Bsize)
	// 総容量
	totalBytes = stat.Blocks * uint64(stat.Bsize)

	return freeBytes, totalBytes, nil
}

// getDiskSpaceWindows はWindowsでディスク容量を取得
func getDiskSpaceWindows(path string) (freeBytes, totalBytes uint64, err error) {
	// Windowsの実装はsyscall.GetDiskFreeSpaceExを使用
	// 簡易実装: ダミー値を返す（将来的にはWindows APIを呼び出す）
	return 0, 0, fmt.Errorf("Windows disk space detection not yet implemented")
}
