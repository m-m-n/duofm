package fs

import (
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestDisplayNameWithLimit(t *testing.T) {
	tests := []struct {
		name     string
		entry    FileEntry
		maxWidth int
		want     string
	}{
		// ASCII tests
		{
			name:     "ASCII name fits exactly",
			entry:    FileEntry{Name: "file.txt"},
			maxWidth: 8,
			want:     "file.txt",
		},
		{
			name:     "ASCII name with room",
			entry:    FileEntry{Name: "file.txt"},
			maxWidth: 20,
			want:     "file.txt",
		},
		{
			name:     "ASCII name truncated",
			entry:    FileEntry{Name: "verylongfilename.txt"},
			maxWidth: 10,
			want:     "verylon...", // 7 chars + ... = 10
		},
		// Japanese tests - each Japanese character has display width of 2
		{
			name:     "Japanese name fits",
			entry:    FileEntry{Name: "テスト.txt"},
			maxWidth: 20,
			want:     "テスト.txt",
		},
		{
			name:     "Japanese name truncated",
			entry:    FileEntry{Name: "日本語ファイル名.txt"},
			maxWidth: 10,
			want:     "日本語...", // 日(2) + 本(2) + 語(2) + ...(3) = 9 (fits in 10)
		},
		{
			name:     "Mixed ASCII and Japanese truncated",
			entry:    FileEntry{Name: "file_テスト.txt"},
			maxWidth: 14,
			want:     "file_テスト...", // file_(5) + テスト(6) + ...(3) = 14
		},
		{
			name:     "Mixed ASCII and Japanese fits",
			entry:    FileEntry{Name: "file_テスト.txt"},
			maxWidth: 15,
			want:     "file_テスト.txt", // display width = 5 + 6 + 4 = 15, fits exactly
		},
		// Directory tests
		{
			name:     "Japanese directory name truncated",
			entry:    FileEntry{Name: "フォルダ", IsDir: true},
			maxWidth: 8,
			want:     "フォ...", // フ(2) + ォ(2) + ...(3) = 7, フォルダ/(9) > 8
		},
		{
			name:     "Japanese directory name fits",
			entry:    FileEntry{Name: "フォルダ", IsDir: true},
			maxWidth: 9,
			want:     "フォルダ/", // フォルダ(8) + /(1) = 9
		},
		// Symlink tests
		{
			name:     "Symlink with Japanese target fits",
			entry:    FileEntry{Name: "link", IsSymlink: true, LinkTarget: "テスト"},
			maxWidth: 20,
			want:     "link -> テスト",
		},
		{
			name:     "Symlink with Japanese target truncated",
			entry:    FileEntry{Name: "link", IsSymlink: true, LinkTarget: "日本語ターゲット"},
			maxWidth: 15,
			want:     "link -> ...",
		},
		// Edge cases - current implementation behavior:
		// When maxWidth <= 3, runewidth.Truncate(s, 0, "") returns empty, then ... not added
		// When maxWidth > 3 but < displayWidth, truncate to maxWidth-3 and add ...
		{
			name:     "Very small maxWidth 3",
			entry:    FileEntry{Name: "テスト"},
			maxWidth: 3,
			want:     "テ", // maxWidth=3 is not > 3, so returns Truncate(s, 3, "") = テ (width 2, fits in 3)
		},
		{
			name:     "Very small maxWidth 2",
			entry:    FileEntry{Name: "テスト"},
			maxWidth: 2,
			want:     "テ", // runewidth.Truncate returns as much as fits
		},
		{
			name:     "maxWidth 4 with Japanese",
			entry:    FileEntry{Name: "テスト"},
			maxWidth: 4,
			want:     "...", // maxWidth > 3: Truncate(s, 1, "") = "" (can't fit テ in width 1), + ... = ...
		},
		{
			name:     "maxWidth 5 with Japanese",
			entry:    FileEntry{Name: "テスト"},
			maxWidth: 5,
			want:     "テ...", // テ(2) + ...(3) = 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.DisplayNameWithLimit(tt.maxWidth)
			if got != tt.want {
				t.Errorf("DisplayNameWithLimit(%d) = %q (width=%d), want %q (width=%d)",
					tt.maxWidth, got, runewidth.StringWidth(got), tt.want, runewidth.StringWidth(tt.want))
			}
		})
	}
}

func TestDisplayName(t *testing.T) {
	tests := []struct {
		name  string
		entry FileEntry
		want  string
	}{
		{
			name:  "Regular file",
			entry: FileEntry{Name: "file.txt"},
			want:  "file.txt",
		},
		{
			name:  "Directory adds slash",
			entry: FileEntry{Name: "dir", IsDir: true},
			want:  "dir/",
		},
		{
			name:  "Parent dir no slash",
			entry: FileEntry{Name: "..", IsDir: true},
			want:  "..",
		},
		{
			name:  "Symlink shows target",
			entry: FileEntry{Name: "link", IsSymlink: true, LinkTarget: "/path/to/target"},
			want:  "link -> /path/to/target",
		},
		{
			name:  "Symlink dir no slash",
			entry: FileEntry{Name: "linkdir", IsDir: true, IsSymlink: true, LinkTarget: "/path"},
			want:  "linkdir -> /path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.DisplayName()
			if got != tt.want {
				t.Errorf("DisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}
