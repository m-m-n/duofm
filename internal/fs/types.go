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
}

// IsParentDir は親ディレクトリエントリかどうかを判定
func (e FileEntry) IsParentDir() bool {
	return e.Name == ".."
}

// DisplayName は表示用の名前を返す
func (e FileEntry) DisplayName() string {
	if e.IsDir && !e.IsParentDir() {
		return e.Name + "/"
	}
	return e.Name
}
