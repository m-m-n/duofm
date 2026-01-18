package filter

import (
	"os"
	"testing"
	"time"

	"github.com/sakura/duofm/internal/fs"
)

func TestFilterSQLLike_Basic(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "small.txt", Size: 100, IsDir: false},
		{Name: "large.txt", Size: 2 * 1024 * 1024, IsDir: false},
		{Name: "folder", Size: 0, IsDir: true},
		{Name: "image.jpg", Size: 500 * 1024, IsDir: false},
	}

	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{"empty query returns all", "", 4},
		{"size filter", "size > 1MiB", 1},
		{"isdir filter", "isdir", 1},
		{"NOT isdir filter", "NOT isdir", 3},
		{"ext filter", "ext = 'txt'", 2},
		{"name LIKE", "name LIKE '%.txt'", 2},
		{"combined AND", "size > 100 AND NOT isdir", 2},
		{"combined OR", "ext = 'txt' OR ext = 'jpg'", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterSQLLike(entries, tt.query)
			if err != nil {
				t.Errorf("FilterSQLLike(%q) error = %v", tt.query, err)
				return
			}
			if len(result) != tt.expected {
				t.Errorf("FilterSQLLike(%q) = %d entries, want %d", tt.query, len(result), tt.expected)
			}
		})
	}
}

func TestFilterSQLLike_Error(t *testing.T) {
	entries := []fs.FileEntry{
		{Name: "test.txt", Size: 100},
	}

	tests := []struct {
		name  string
		query string
	}{
		{"unknown column", "unknown_col = 1"},
		{"unclosed paren", "(size > 1"},
		{"invalid syntax", "size > > 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FilterSQLLike(entries, tt.query)
			if err == nil {
				t.Errorf("FilterSQLLike(%q) expected error, got nil", tt.query)
			}
		})
	}
}

func TestCompileQuery_Reuse(t *testing.T) {
	query := "size > 100 AND NOT isdir"
	compiled, err := CompileQuery(query)
	if err != nil {
		t.Fatalf("CompileQuery(%q) error = %v", query, err)
	}

	entries := []fs.FileEntry{
		{Name: "small.txt", Size: 50, IsDir: false},
		{Name: "large.txt", Size: 200, IsDir: false},
		{Name: "folder", Size: 500, IsDir: true},
	}

	// Test Match
	if compiled.Match(entries[0]) {
		t.Errorf("Match(small.txt) = true, want false")
	}
	if !compiled.Match(entries[1]) {
		t.Errorf("Match(large.txt) = false, want true")
	}
	if compiled.Match(entries[2]) {
		t.Errorf("Match(folder) = true, want false")
	}

	// Test Filter
	result := compiled.Filter(entries)
	if len(result) != 1 {
		t.Errorf("Filter() = %d entries, want 1", len(result))
	}

	// Test RawQuery
	if compiled.RawQuery() != query {
		t.Errorf("RawQuery() = %q, want %q", compiled.RawQuery(), query)
	}
}

func TestValidateQuery(t *testing.T) {
	tests := []struct {
		query   string
		wantErr bool
	}{
		{"size > 1MiB", false},
		{"name LIKE '%.txt'", false},
		{"ext IN ('go', 'md')", false},
		{"unknown_col = 1", true},
		{"size > > 1", true},
		{"", false}, // empty is valid
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			err := ValidateQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQuery(%q) error = %v, wantErr = %v", tt.query, err, tt.wantErr)
			}
		})
	}
}

func TestFilterSQLLike_AllColumns(t *testing.T) {
	// Create a test entry with all fields populated
	entry := fs.FileEntry{
		Name:        "README.md",
		Size:        1024,
		ModTime:     time.Date(2024, 6, 15, 10, 30, 0, 0, time.Local),
		IsDir:       false,
		IsSymlink:   false,
		Permissions: os.FileMode(0644),
		Owner:       "user",
		Group:       "staff",
	}
	entries := []fs.FileEntry{entry}

	tests := []struct {
		query   string
		matches bool
	}{
		// name column
		{"name = 'README.md'", true},
		{"name = 'other.txt'", false},

		// size column
		{"size = 1024", true},
		{"size > 1023", true},
		{"size < 1025", true},
		{"size = 1KiB", true},

		// mtime column
		{"year(mtime) = 2024", true},
		{"month(mtime) = 6", true},
		{"day(mtime) = 15", true},

		// type column
		{"type = 'file'", true},
		{"type = 'dir'", false},

		// ext column
		{"ext = 'md'", true},
		{"ext = 'txt'", false},

		// perm column
		{"perm = '-rw-r--r--'", true},

		// owner column
		{"owner = 'user'", true},
		{"owner = 'other'", false},

		// group column
		{"group = 'staff'", true},

		// isdir column
		{"isdir", false},
		{"NOT isdir", true},

		// isfile column
		{"isfile", true},

		// issymlink column
		{"issymlink", false},
		{"NOT issymlink", true},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result, err := FilterSQLLike(entries, tt.query)
			if err != nil {
				t.Errorf("FilterSQLLike(%q) error = %v", tt.query, err)
				return
			}
			if (len(result) == 1) != tt.matches {
				t.Errorf("FilterSQLLike(%q) matches = %v, want %v", tt.query, len(result) == 1, tt.matches)
			}
		})
	}
}

func TestFilterSQLLike_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		entries []fs.FileEntry
		query   string
		count   int
	}{
		{
			name:    "empty entries",
			entries: []fs.FileEntry{},
			query:   "size > 0",
			count:   0,
		},
		{
			name: "whitespace-only query",
			entries: []fs.FileEntry{
				{Name: "test.txt", Size: 100},
			},
			query: "   ",
			count: 1,
		},
		{
			name: "zero size file",
			entries: []fs.FileEntry{
				{Name: "empty.txt", Size: 0},
				{Name: "full.txt", Size: 100},
			},
			query: "size = 0",
			count: 1,
		},
		{
			name: "hidden file",
			entries: []fs.FileEntry{
				{Name: ".gitignore", Size: 100},
				{Name: "README.md", Size: 100},
			},
			query: "name LIKE '.%'",
			count: 1,
		},
		{
			name: "file with no extension",
			entries: []fs.FileEntry{
				{Name: "Makefile", Size: 100},
				{Name: "README.md", Size: 100},
			},
			query: "ext IS NULL",
			count: 1,
		},
		{
			name: "Unicode filename",
			entries: []fs.FileEntry{
				{Name: "日本語.txt", Size: 100},
				{Name: "english.txt", Size: 100},
			},
			query: "name LIKE '%日本語%'",
			count: 1,
		},
		{
			name: "deeply nested parentheses",
			entries: []fs.FileEntry{
				{Name: "test.txt", Size: 100, IsDir: false},
			},
			query: "((((size > 0 AND NOT isdir))))",
			count: 1,
		},
		{
			name: "many OR conditions",
			entries: []fs.FileEntry{
				{Name: "a.txt", Size: 100},
				{Name: "b.go", Size: 100},
				{Name: "c.md", Size: 100},
				{Name: "d.rs", Size: 100},
			},
			query: "ext = 'txt' OR ext = 'go' OR ext = 'md'",
			count: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterSQLLike(tt.entries, tt.query)
			if err != nil {
				t.Errorf("FilterSQLLike(%q) error = %v", tt.query, err)
				return
			}
			if len(result) != tt.count {
				t.Errorf("FilterSQLLike(%q) = %d entries, want %d", tt.query, len(result), tt.count)
			}
		})
	}
}

func TestCompiledQuery_EmptyAST(t *testing.T) {
	// Test with empty query that produces nil AST
	compiled, err := CompileQuery("")
	if err != nil {
		t.Fatalf("CompileQuery('') error = %v", err)
	}

	entry := fs.FileEntry{Name: "test.txt", Size: 100}

	// Empty query should match all entries
	if !compiled.Match(entry) {
		t.Error("Empty query Match() = false, want true")
	}

	entries := []fs.FileEntry{entry, {Name: "other.txt", Size: 200}}
	result := compiled.Filter(entries)
	if len(result) != 2 {
		t.Errorf("Empty query Filter() = %d entries, want 2", len(result))
	}
}
