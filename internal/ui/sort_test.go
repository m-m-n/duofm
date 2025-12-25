package ui

import (
	"testing"
	"time"

	"github.com/sakura/duofm/internal/fs"
)

func TestSortConfig_String(t *testing.T) {
	tests := []struct {
		name   string
		config SortConfig
		want   string
	}{
		{
			name:   "Name ascending",
			config: SortConfig{Field: SortByName, Order: SortAsc},
			want:   "Name ↑",
		},
		{
			name:   "Name descending",
			config: SortConfig{Field: SortByName, Order: SortDesc},
			want:   "Name ↓",
		},
		{
			name:   "Size ascending",
			config: SortConfig{Field: SortBySize, Order: SortAsc},
			want:   "Size ↑",
		},
		{
			name:   "Size descending",
			config: SortConfig{Field: SortBySize, Order: SortDesc},
			want:   "Size ↓",
		},
		{
			name:   "Date ascending",
			config: SortConfig{Field: SortByDate, Order: SortAsc},
			want:   "Date ↑",
		},
		{
			name:   "Date descending",
			config: SortConfig{Field: SortByDate, Order: SortDesc},
			want:   "Date ↓",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.String()
			if got != tt.want {
				t.Errorf("SortConfig.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSortEntries_ParentDirAlwaysFirst(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "file1.txt", IsDir: false},
		{Name: "..", IsDir: true},
		{Name: "dir1", IsDir: true},
	}

	// すべてのソートモードで親ディレクトリが先頭にあることを確認
	configs := []SortConfig{
		{Field: SortByName, Order: SortAsc},
		{Field: SortByName, Order: SortDesc},
		{Field: SortBySize, Order: SortAsc},
		{Field: SortBySize, Order: SortDesc},
		{Field: SortByDate, Order: SortAsc},
		{Field: SortByDate, Order: SortDesc},
	}

	for _, config := range configs {
		t.Run(config.String(), func(t *testing.T) {
			result := SortEntries(entries, config)
			if len(result) == 0 {
				t.Fatal("Empty result")
			}
			if result[0].Name != ".." {
				t.Errorf("Parent directory not first: got %q", result[0].Name)
			}
		})
	}
}

func TestSortEntries_DirectoriesBeforeFiles(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "file1.txt", IsDir: false},
		{Name: "aaa_dir", IsDir: true},
		{Name: "zzz_file.txt", IsDir: false},
		{Name: "bbb_dir", IsDir: true},
	}

	configs := []SortConfig{
		{Field: SortByName, Order: SortAsc},
		{Field: SortByName, Order: SortDesc},
	}

	for _, config := range configs {
		t.Run(config.String(), func(t *testing.T) {
			result := SortEntries(entries, config)

			// ディレクトリがファイルより先に来ることを確認
			dirsDone := false
			for _, e := range result {
				if e.IsDir {
					if dirsDone {
						t.Errorf("Directory %q found after files", e.Name)
					}
				} else {
					dirsDone = true
				}
			}
		})
	}
}

func TestSortEntries_ByName(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "charlie.txt", IsDir: false},
		{Name: "alpha.txt", IsDir: false},
		{Name: "bravo.txt", IsDir: false},
	}

	t.Run("Ascending", func(t *testing.T) {
		config := SortConfig{Field: SortByName, Order: SortAsc}
		result := SortEntries(entries, config)

		expected := []string{"alpha.txt", "bravo.txt", "charlie.txt"}
		for i, name := range expected {
			if result[i].Name != name {
				t.Errorf("Position %d: got %q, want %q", i, result[i].Name, name)
			}
		}
	})

	t.Run("Descending", func(t *testing.T) {
		config := SortConfig{Field: SortByName, Order: SortDesc}
		result := SortEntries(entries, config)

		expected := []string{"charlie.txt", "bravo.txt", "alpha.txt"}
		for i, name := range expected {
			if result[i].Name != name {
				t.Errorf("Position %d: got %q, want %q", i, result[i].Name, name)
			}
		}
	})
}

func TestSortEntries_BySize(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "medium.txt", IsDir: false, Size: 500},
		{Name: "small.txt", IsDir: false, Size: 100},
		{Name: "large.txt", IsDir: false, Size: 1000},
	}

	t.Run("Ascending", func(t *testing.T) {
		config := SortConfig{Field: SortBySize, Order: SortAsc}
		result := SortEntries(entries, config)

		expected := []string{"small.txt", "medium.txt", "large.txt"}
		for i, name := range expected {
			if result[i].Name != name {
				t.Errorf("Position %d: got %q, want %q", i, result[i].Name, name)
			}
		}
	})

	t.Run("Descending", func(t *testing.T) {
		config := SortConfig{Field: SortBySize, Order: SortDesc}
		result := SortEntries(entries, config)

		expected := []string{"large.txt", "medium.txt", "small.txt"}
		for i, name := range expected {
			if result[i].Name != name {
				t.Errorf("Position %d: got %q, want %q", i, result[i].Name, name)
			}
		}
	})
}

func TestSortEntries_ByDate(t *testing.T) {
	now := time.Now()
	entries := []fs.FileEntry{
		{Name: "recent.txt", IsDir: false, ModTime: now},
		{Name: "old.txt", IsDir: false, ModTime: now.Add(-24 * time.Hour)},
		{Name: "older.txt", IsDir: false, ModTime: now.Add(-48 * time.Hour)},
	}

	t.Run("Ascending (old to new)", func(t *testing.T) {
		config := SortConfig{Field: SortByDate, Order: SortAsc}
		result := SortEntries(entries, config)

		expected := []string{"older.txt", "old.txt", "recent.txt"}
		for i, name := range expected {
			if result[i].Name != name {
				t.Errorf("Position %d: got %q, want %q", i, result[i].Name, name)
			}
		}
	})

	t.Run("Descending (new to old)", func(t *testing.T) {
		config := SortConfig{Field: SortByDate, Order: SortDesc}
		result := SortEntries(entries, config)

		expected := []string{"recent.txt", "old.txt", "older.txt"}
		for i, name := range expected {
			if result[i].Name != name {
				t.Errorf("Position %d: got %q, want %q", i, result[i].Name, name)
			}
		}
	})
}

func TestSortEntries_CompleteScenario(t *testing.T) {
	now := time.Now()
	entries := []fs.FileEntry{
		{Name: "file1.txt", IsDir: false, Size: 100, ModTime: now},
		{Name: "..", IsDir: true},
		{Name: "dir_z", IsDir: true, ModTime: now},
		{Name: "file2.txt", IsDir: false, Size: 200, ModTime: now.Add(-time.Hour)},
		{Name: "dir_a", IsDir: true, ModTime: now.Add(-time.Hour)},
	}

	t.Run("Name ascending", func(t *testing.T) {
		config := SortConfig{Field: SortByName, Order: SortAsc}
		result := SortEntries(entries, config)

		// 期待される順序: .., dir_a, dir_z, file1.txt, file2.txt
		expected := []string{"..", "dir_a", "dir_z", "file1.txt", "file2.txt"}
		for i, name := range expected {
			if result[i].Name != name {
				t.Errorf("Position %d: got %q, want %q", i, result[i].Name, name)
			}
		}
	})

	t.Run("Size descending", func(t *testing.T) {
		config := SortConfig{Field: SortBySize, Order: SortDesc}
		result := SortEntries(entries, config)

		// 親ディレクトリが先頭、ディレクトリがファイルより先
		if result[0].Name != ".." {
			t.Errorf("First should be '..', got %q", result[0].Name)
		}

		// ファイル部分でサイズ降順を確認
		fileStart := 0
		for i, e := range result {
			if !e.IsDir {
				fileStart = i
				break
			}
		}
		if fileStart > 0 && len(result) > fileStart+1 {
			// ファイルが大きい順になっていることを確認
			if result[fileStart].Name != "file2.txt" || result[fileStart+1].Name != "file1.txt" {
				t.Errorf("Files not sorted by size descending")
			}
		}
	})
}

func TestSortEntries_EmptyList(t *testing.T) {
	entries := []fs.FileEntry{}
	config := SortConfig{Field: SortByName, Order: SortAsc}

	result := SortEntries(entries, config)

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d items", len(result))
	}
}

func TestSortEntries_SingleItem(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "only_file.txt", IsDir: false},
	}
	config := SortConfig{Field: SortByName, Order: SortAsc}

	result := SortEntries(entries, config)

	if len(result) != 1 || result[0].Name != "only_file.txt" {
		t.Errorf("Single item not preserved correctly")
	}
}

func TestSortEntries_OnlyParentDir(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "..", IsDir: true},
	}
	config := SortConfig{Field: SortByName, Order: SortAsc}

	result := SortEntries(entries, config)

	if len(result) != 1 || result[0].Name != ".." {
		t.Errorf("Parent dir only not preserved correctly")
	}
}

func TestDefaultSortConfig(t *testing.T) {
	config := DefaultSortConfig()

	if config.Field != SortByName {
		t.Errorf("Default Field = %v, want SortByName", config.Field)
	}
	if config.Order != SortAsc {
		t.Errorf("Default Order = %v, want SortAsc", config.Order)
	}
}
