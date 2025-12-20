package ui

import (
	"fmt"
	"io/fs"
	"time"
)

// FormatSize はバイト数を人間が読みやすいサイズ文字列に変換する（1024進法）
// 例: 1024 → "1.0 KiB", 1536 → "1.5 KiB"
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	div := int64(unit)
	exp := 0

	for n := bytes / unit; n >= unit && exp < len(units)-2; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp+1])
}

// FormatTimestamp は時刻をISO 8601形式の文字列に変換する
// フォーマット: "2024-12-17 22:28"
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

// FormatPermissions はファイルモードをUnix形式のパーミッション文字列に変換する
// 例: 0755 → "rwxr-xr-x"
func FormatPermissions(mode fs.FileMode) string {
	perm := mode.Perm()

	var result [9]byte

	// Owner (user)
	if perm&0400 != 0 {
		result[0] = 'r'
	} else {
		result[0] = '-'
	}
	if perm&0200 != 0 {
		result[1] = 'w'
	} else {
		result[1] = '-'
	}
	if perm&0100 != 0 {
		result[2] = 'x'
	} else {
		result[2] = '-'
	}

	// Group
	if perm&0040 != 0 {
		result[3] = 'r'
	} else {
		result[3] = '-'
	}
	if perm&0020 != 0 {
		result[4] = 'w'
	} else {
		result[4] = '-'
	}
	if perm&0010 != 0 {
		result[5] = 'x'
	} else {
		result[5] = '-'
	}

	// Others
	if perm&0004 != 0 {
		result[6] = 'r'
	} else {
		result[6] = '-'
	}
	if perm&0002 != 0 {
		result[7] = 'w'
	} else {
		result[7] = '-'
	}
	if perm&0001 != 0 {
		result[8] = 'x'
	} else {
		result[8] = '-'
	}

	return string(result[:])
}

// CalculateColumnWidths は端末幅に基づいて各カラムの幅を計算する
// 返り値:
//   - nameWidth: ファイル名カラムの幅
//   - hasSpace: 詳細情報を表示する余裕があるか
func CalculateColumnWidths(terminalWidth int) (nameWidth int, hasSpace bool) {
	// 各カラムの固定幅
	const (
		sizeWidth        = 10 // "1.2 MiB" など
		timestampWidth   = 16 // "2024-12-17 22:28"
		permissionsWidth = 9  // "rwxr-xr-x"
		ownerWidth       = 10 // 所有者名
		groupWidth       = 10 // グループ名
		padding          = 2  // カラム間のパディング
	)

	// 基本情報モード（名前 + サイズ + タイムスタンプ）に必要な最小幅
	basicModeWidth := sizeWidth + timestampWidth + padding*3

	// 詳細情報モード（名前 + パーミッション + 所有者 + グループ）に必要な最小幅
	detailModeWidth := permissionsWidth + ownerWidth + groupWidth + padding*4

	// 端末幅が狭い場合は名前のみ
	if terminalWidth < 60 {
		nameWidth = terminalWidth - 5 // マージンを考慮
		hasSpace = false
		return
	}

	// 基本情報モードまたは詳細情報モードを表示できる
	hasSpace = true

	// ファイル名に残りの幅を割り当て
	// 基本情報モードと詳細情報モードのうち、大きい方を考慮
	requiredWidth := basicModeWidth
	if detailModeWidth > requiredWidth {
		requiredWidth = detailModeWidth
	}

	nameWidth = terminalWidth - requiredWidth
	if nameWidth < 20 {
		nameWidth = 20 // 最小幅を保証
	}

	return nameWidth, hasSpace
}
