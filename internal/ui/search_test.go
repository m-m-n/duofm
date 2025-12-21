package ui

import (
	"testing"

	"github.com/sakura/duofm/internal/fs"
)

func TestIsSmartCaseSensitive(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    bool
	}{
		{
			name:    "all lowercase returns false",
			pattern: "abc",
			want:    false,
		},
		{
			name:    "contains uppercase returns true",
			pattern: "Abc",
			want:    true,
		},
		{
			name:    "all uppercase returns true",
			pattern: "ABC",
			want:    true,
		},
		{
			name:    "mixed case returns true",
			pattern: "aBc",
			want:    true,
		},
		{
			name:    "empty string returns false",
			pattern: "",
			want:    false,
		},
		{
			name:    "numbers only returns false",
			pattern: "123",
			want:    false,
		},
		{
			name:    "special characters only returns false",
			pattern: ".*",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSmartCaseSensitive(tt.pattern)
			if got != tt.want {
				t.Errorf("isSmartCaseSensitive(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestFilterIncremental(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "README.md"},
		{Name: "readme.txt"},
		{Name: "Makefile"},
		{Name: "main.go"},
		{Name: "model.go"},
		{Name: "CHANGELOG.md"},
	}

	tests := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			name:    "empty pattern returns all entries",
			pattern: "",
			want:    []string{"README.md", "readme.txt", "Makefile", "main.go", "model.go", "CHANGELOG.md"},
		},
		{
			name:    "lowercase pattern matches case-insensitively",
			pattern: "readme",
			want:    []string{"README.md", "readme.txt"},
		},
		{
			name:    "uppercase pattern matches case-sensitively",
			pattern: "README",
			want:    []string{"README.md"},
		},
		{
			name:    "partial match works",
			pattern: "go",
			want:    []string{"main.go", "model.go"},
		},
		{
			name:    "no match returns empty",
			pattern: "xyz",
			want:    []string{},
		},
		{
			name:    "mixed case pattern is case-sensitive",
			pattern: "Make",
			want:    []string{"Makefile"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterIncremental(entries, tt.pattern)
			if len(got) != len(tt.want) {
				t.Errorf("filterIncremental() returned %d entries, want %d", len(got), len(tt.want))
				return
			}
			for i, entry := range got {
				if entry.Name != tt.want[i] {
					t.Errorf("filterIncremental()[%d].Name = %q, want %q", i, entry.Name, tt.want[i])
				}
			}
		})
	}
}

func TestFilterRegex(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "README.md"},
		{Name: "readme.txt"},
		{Name: "Makefile"},
		{Name: "main.go"},
		{Name: "model.go"},
		{Name: "test_file.go"},
	}

	tests := []struct {
		name    string
		pattern string
		want    []string
		wantErr bool
	}{
		{
			name:    "empty pattern returns all entries",
			pattern: "",
			want:    []string{"README.md", "readme.txt", "Makefile", "main.go", "model.go", "test_file.go"},
			wantErr: false,
		},
		{
			name:    "simple pattern with lowercase is case-insensitive",
			pattern: "readme",
			want:    []string{"README.md", "readme.txt"},
			wantErr: false,
		},
		{
			name:    "pattern with uppercase is case-sensitive",
			pattern: "README",
			want:    []string{"README.md"},
			wantErr: false,
		},
		{
			name:    "regex pattern matches .go files",
			pattern: `\.go$`,
			want:    []string{"main.go", "model.go", "test_file.go"},
			wantErr: false,
		},
		{
			name:    "regex pattern with word boundary",
			pattern: `^m.*\.go$`,
			want:    []string{"main.go", "model.go"},
			wantErr: false,
		},
		{
			name:    "invalid regex returns error",
			pattern: "[invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "no match returns empty",
			pattern: "^xyz",
			want:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filterRegex(entries, tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("filterRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("filterRegex() returned %d entries, want %d", len(got), len(tt.want))
				return
			}
			for i, entry := range got {
				if entry.Name != tt.want[i] {
					t.Errorf("filterRegex()[%d].Name = %q, want %q", i, entry.Name, tt.want[i])
				}
			}
		})
	}
}

func TestSearchModeString(t *testing.T) {
	tests := []struct {
		mode SearchMode
		want string
	}{
		{SearchModeNone, "none"},
		{SearchModeIncremental, "incremental"},
		{SearchModeRegex, "regex"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("SearchMode.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
