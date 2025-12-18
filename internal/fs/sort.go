package fs

import (
	"sort"
	"strings"
)

// SortEntries はエントリをソート（ディレクトリ優先、アルファベット順）
func SortEntries(entries []FileEntry) {
	sort.Slice(entries, func(i, j int) bool {
		// 親ディレクトリ (..) は常に最初
		if entries[i].IsParentDir() {
			return true
		}
		if entries[j].IsParentDir() {
			return false
		}

		// ディレクトリとファイルを分離
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir // ディレクトリを先に
		}

		// 同じタイプ内では名前でソート（大文字小文字を区別しない）
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
}
