package fs

import (
	"io/fs"
	"time"
)

// FileEntry はファイル/ディレクトリの情報を保持
type FileEntry struct {
	Name        string
	IsDir       bool
	Size        int64
	ModTime     time.Time
	Permissions fs.FileMode
	Owner       string // 所有者名
	Group       string // グループ名
	IsSymlink   bool   // シンボリックリンクか
	LinkTarget  string // リンク先パス（シンボリックリンクの場合）
	LinkBroken  bool   // リンク切れか（シンボリックリンクの場合）
}

// IsParentDir は親ディレクトリエントリかどうかを判定
func (e FileEntry) IsParentDir() bool {
	return e.Name == ".."
}

// DisplayName は表示用の名前を返す
// maxWidth が指定された場合、リンク先の表示を省略する
func (e FileEntry) DisplayName() string {
	name := e.Name
	if e.IsDir && !e.IsParentDir() && !e.IsSymlink {
		name += "/"
	}

	// シンボリックリンクの場合、リンク先を表示
	if e.IsSymlink && e.LinkTarget != "" {
		name = name + " -> " + e.LinkTarget
	}

	return name
}

// DisplayNameWithLimit は指定された最大幅に収まるように表示用の名前を返す
func (e FileEntry) DisplayNameWithLimit(maxWidth int) string {
	fullName := e.DisplayName()
	if len(fullName) <= maxWidth {
		return fullName
	}

	// シンボリックリンクの場合、リンク先を省略
	if e.IsSymlink && e.LinkTarget != "" {
		// "name -> ..." の形式で省略
		prefix := e.Name + " -> "
		if len(prefix)+3 < maxWidth {
			return prefix + "..."
		}
	}

	// 通常のファイル名の省略
	if maxWidth > 3 {
		return fullName[:maxWidth-3] + "..."
	}
	return fullName[:maxWidth]
}
