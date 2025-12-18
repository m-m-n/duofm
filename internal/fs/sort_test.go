package fs

import (
	"testing"
)

func TestSortEntries(t *testing.T) {
	tests := []struct {
		name    string
		entries []FileEntry
		want    []string
	}{
		{
			name: "親ディレクトリが最初に来る",
			entries: []FileEntry{
				{Name: "file.txt", IsDir: false},
				{Name: "..", IsDir: true},
				{Name: "dir", IsDir: true},
			},
			want: []string{"..", "dir", "file.txt"},
		},
		{
			name: "ディレクトリがファイルより先に来る",
			entries: []FileEntry{
				{Name: "file.txt", IsDir: false},
				{Name: "another_dir", IsDir: true},
				{Name: "zebra_dir", IsDir: true},
			},
			want: []string{"another_dir", "zebra_dir", "file.txt"},
		},
		{
			name: "アルファベット順にソート（大文字小文字を区別しない）",
			entries: []FileEntry{
				{Name: "Zebra.txt", IsDir: false},
				{Name: "apple.txt", IsDir: false},
				{Name: "Banana.txt", IsDir: false},
			},
			want: []string{"apple.txt", "Banana.txt", "Zebra.txt"},
		},
		{
			name: "複合的なソート",
			entries: []FileEntry{
				{Name: "file3.txt", IsDir: false},
				{Name: "DirB", IsDir: true},
				{Name: "..", IsDir: true},
				{Name: "file1.txt", IsDir: false},
				{Name: "dirA", IsDir: true},
				{Name: "file2.txt", IsDir: false},
			},
			want: []string{"..", "dirA", "DirB", "file1.txt", "file2.txt", "file3.txt"},
		},
		{
			name:    "空のエントリリスト",
			entries: []FileEntry{},
			want:    []string{},
		},
		{
			name: "単一エントリ",
			entries: []FileEntry{
				{Name: "single.txt", IsDir: false},
			},
			want: []string{"single.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// エントリのコピーを作成（元のスライスを変更しないため）
			entries := make([]FileEntry, len(tt.entries))
			copy(entries, tt.entries)

			SortEntries(entries)

			// ソート結果を検証
			if len(entries) != len(tt.want) {
				t.Errorf("SortEntries() resulted in %d entries, want %d", len(entries), len(tt.want))
				return
			}

			for i, want := range tt.want {
				if entries[i].Name != want {
					t.Errorf("SortEntries() entry[%d] = %s, want %s", i, entries[i].Name, want)
				}
			}
		})
	}
}

func TestSortEntriesStability(t *testing.T) {
	// 同じ名前（大文字小文字違い）の場合の安定性テスト
	entries := []FileEntry{
		{Name: "file", IsDir: false},
		{Name: "File", IsDir: false},
		{Name: "FILE", IsDir: false},
	}

	SortEntries(entries)

	// すべてのエントリが存在することを確認
	names := make(map[string]bool)
	for _, entry := range entries {
		names[entry.Name] = true
	}

	if !names["file"] || !names["File"] || !names["FILE"] {
		t.Error("SortEntries() should preserve all entries")
	}
}

func TestSortEntriesPreservesProperties(t *testing.T) {
	// ソートがエントリのプロパティを保持することを確認
	entries := []FileEntry{
		{Name: "file2.txt", IsDir: false, Size: 200},
		{Name: "file1.txt", IsDir: false, Size: 100},
		{Name: "dir1", IsDir: true, Size: 0},
	}

	SortEntries(entries)

	// プロパティが保持されているか確認
	for _, entry := range entries {
		if entry.Name == "file1.txt" && entry.Size != 100 {
			t.Error("SortEntries() should preserve Size property")
		}
		if entry.Name == "file2.txt" && entry.Size != 200 {
			t.Error("SortEntries() should preserve Size property")
		}
		if entry.Name == "dir1" && !entry.IsDir {
			t.Error("SortEntries() should preserve IsDir property")
		}
	}
}
