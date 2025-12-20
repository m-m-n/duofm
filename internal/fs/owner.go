package fs

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"syscall"
)

// GetFileOwnerGroup は指定されたパスのファイルの所有者とグループ名を返す
// Unix/Linux: syscall.Stat_tからUIDとGIDを取得し、user.LookupId/LookupGroupIdで名前解決
// Windows: 制限的なサポート（"N/A"を返す）
// エラー時はプレースホルダー（"unknown"）を返す
func GetFileOwnerGroup(path string) (owner, group string, err error) {
	// ファイル情報を取得
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return "", "", fmt.Errorf("failed to get file info: %w", err)
	}

	// Windowsの場合は制限的なサポート
	if runtime.GOOS == "windows" {
		return "N/A", "N/A", nil
	}

	// Unix/Linux: syscall.Stat_tから所有者情報を取得
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return "unknown", "unknown", nil
	}

	// UIDから所有者名を取得
	uid := strconv.Itoa(int(stat.Uid))
	u, err := user.LookupId(uid)
	if err != nil {
		owner = uid // 名前が解決できない場合はUIDを使用
	} else {
		owner = u.Username
	}

	// GIDからグループ名を取得
	gid := strconv.Itoa(int(stat.Gid))
	g, err := user.LookupGroupId(gid)
	if err != nil {
		group = gid // 名前が解決できない場合はGIDを使用
	} else {
		group = g.Name
	}

	return owner, group, nil
}
