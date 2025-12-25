package ui

import (
	"fmt"
	"sort"

	"github.com/sakura/duofm/internal/fs"
)

// SortField はソート対象のフィールドを表す
type SortField int

const (
	SortByName SortField = iota
	SortBySize
	SortByDate
)

// SortOrder はソート順序を表す
type SortOrder int

const (
	SortAsc SortOrder = iota
	SortDesc
)

// SortConfig はソート設定を保持
type SortConfig struct {
	Field SortField
	Order SortOrder
}

// String はソート設定の表示用文字列を返す
func (c SortConfig) String() string {
	fields := []string{"Name", "Size", "Date"}
	orders := []string{"↑", "↓"}
	return fmt.Sprintf("%s %s", fields[c.Field], orders[c.Order])
}

// DefaultSortConfig はデフォルトのソート設定を返す
func DefaultSortConfig() SortConfig {
	return SortConfig{Field: SortByName, Order: SortAsc}
}

// SortEntries はエントリを指定された設定でソートする
// 親ディレクトリ(..)は常に先頭、ディレクトリはファイルより先に配置
func SortEntries(entries []fs.FileEntry, config SortConfig) []fs.FileEntry {
	if len(entries) == 0 {
		return entries
	}

	// 親ディレクトリ、ディレクトリ、ファイルを分離
	var parentDir []fs.FileEntry
	var dirs []fs.FileEntry
	var files []fs.FileEntry

	for _, e := range entries {
		if e.Name == ".." {
			parentDir = append(parentDir, e)
		} else if e.IsDir {
			dirs = append(dirs, e)
		} else {
			files = append(files, e)
		}
	}

	// 比較関数を取得
	less := getLessFunc(config)

	// ディレクトリとファイルを個別にソート
	sort.SliceStable(dirs, func(i, j int) bool {
		return less(dirs[i], dirs[j])
	})
	sort.SliceStable(files, func(i, j int) bool {
		return less(files[i], files[j])
	})

	// 結合: 親ディレクトリ → ディレクトリ → ファイル
	result := make([]fs.FileEntry, 0, len(entries))
	result = append(result, parentDir...)
	result = append(result, dirs...)
	result = append(result, files...)

	return result
}

// getLessFunc はSortConfigに基づいて比較関数を返す
func getLessFunc(config SortConfig) func(a, b fs.FileEntry) bool {
	switch config.Field {
	case SortByName:
		if config.Order == SortAsc {
			return func(a, b fs.FileEntry) bool { return a.Name < b.Name }
		}
		return func(a, b fs.FileEntry) bool { return a.Name > b.Name }
	case SortBySize:
		if config.Order == SortAsc {
			return func(a, b fs.FileEntry) bool { return a.Size < b.Size }
		}
		return func(a, b fs.FileEntry) bool { return a.Size > b.Size }
	case SortByDate:
		if config.Order == SortAsc {
			return func(a, b fs.FileEntry) bool { return a.ModTime.Before(b.ModTime) }
		}
		return func(a, b fs.FileEntry) bool { return a.ModTime.After(b.ModTime) }
	default:
		return func(a, b fs.FileEntry) bool { return a.Name < b.Name }
	}
}
