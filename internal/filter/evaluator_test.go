package filter

import (
	"io/fs"
	"testing"
	"time"

	internalfs "github.com/sakura/duofm/internal/fs"
)

// createTestEntry creates a FileEntry for testing.
func createTestEntry(name string, size int64, isDir bool, modTime time.Time) internalfs.FileEntry {
	mode := fs.FileMode(0644)
	if isDir {
		mode = fs.FileMode(0755) | fs.ModeDir
	}
	return internalfs.FileEntry{
		Name:        name,
		IsDir:       isDir,
		Size:        size,
		ModTime:     modTime,
		Permissions: mode,
		Owner:       "user",
		Group:       "group",
		IsSymlink:   false,
		LinkTarget:  "",
		LinkBroken:  false,
	}
}

func TestEvaluator_SizeComparison(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "size > 100 - true",
			query:    "size > 100",
			entry:    createTestEntry("test.txt", 200, false, time.Now()),
			expected: true,
		},
		{
			name:     "size > 100 - false",
			query:    "size > 100",
			entry:    createTestEntry("test.txt", 50, false, time.Now()),
			expected: false,
		},
		{
			name:     "size = 0",
			query:    "size = 0",
			entry:    createTestEntry("empty.txt", 0, false, time.Now()),
			expected: true,
		},
		{
			name:     "size >= 1GiB - true",
			query:    "size >= 1GiB",
			entry:    createTestEntry("large.bin", 1024*1024*1024, false, time.Now()),
			expected: true,
		},
		{
			name:     "size >= 1GiB - false",
			query:    "size >= 1GiB",
			entry:    createTestEntry("small.txt", 1024*1024, false, time.Now()),
			expected: false,
		},
		{
			name:     "size < 1MiB",
			query:    "size < 1MiB",
			entry:    createTestEntry("small.txt", 512*1024, false, time.Now()),
			expected: true,
		},
		{
			name:     "size > 1GB (decimal)",
			query:    "size > 1GB",
			entry:    createTestEntry("large.bin", 1500000000, false, time.Now()),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_StringEquality(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "name = 'test.txt' - true",
			query:    "name = 'test.txt'",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "name = 'test.txt' - false",
			query:    "name = 'test.txt'",
			entry:    createTestEntry("other.txt", 100, false, time.Now()),
			expected: false,
		},
		{
			name:     "type = 'file'",
			query:    "type = 'file'",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "type = 'dir'",
			query:    "type = 'dir'",
			entry:    createTestEntry("mydir", 0, true, time.Now()),
			expected: true,
		},
		{
			name:     "ext = 'txt'",
			query:    "ext = 'txt'",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "ext = 'go'",
			query:    "ext = 'go'",
			entry:    createTestEntry("main.go", 500, false, time.Now()),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_LIKEPattern(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "LIKE with % prefix",
			query:    "name LIKE '%.txt'",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "LIKE with % prefix - no match",
			query:    "name LIKE '%.txt'",
			entry:    createTestEntry("test.go", 100, false, time.Now()),
			expected: false,
		},
		{
			name:     "LIKE with % suffix",
			query:    "name LIKE 'test%'",
			entry:    createTestEntry("test_file.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "LIKE with % both sides",
			query:    "name LIKE '%report%'",
			entry:    createTestEntry("annual_report_2024.pdf", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "LIKE with _ single char",
			query:    "name LIKE 'file_.txt'",
			entry:    createTestEntry("file1.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "LIKE with _ - no match (multiple chars)",
			query:    "name LIKE 'file_.txt'",
			entry:    createTestEntry("file12.txt", 100, false, time.Now()),
			expected: false,
		},
		{
			name:     "LIKE with multiple _",
			query:    "name LIKE '___.go'",
			entry:    createTestEntry("foo.go", 100, false, time.Now()),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_ILIKEPattern(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "ILIKE case insensitive",
			query:    "name ILIKE '%test%'",
			entry:    createTestEntry("TEST_FILE.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "ILIKE with uppercase pattern",
			query:    "name ILIKE '%TEST%'",
			entry:    createTestEntry("test_file.txt", 100, false, time.Now()),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_NOTLIKEPattern(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "NOT LIKE - true",
			query:    "name NOT LIKE '%.bak'",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "NOT LIKE - false",
			query:    "name NOT LIKE '%.bak'",
			entry:    createTestEntry("test.bak", 100, false, time.Now()),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_DateComparison(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	lastWeek := now.AddDate(0, 0, -10)

	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "mtime > now() - 7d (recent file)",
			query:    "mtime > now() - 7d",
			entry:    createTestEntry("recent.txt", 100, false, yesterday),
			expected: true,
		},
		{
			name:     "mtime > now() - 7d (old file)",
			query:    "mtime > now() - 7d",
			entry:    createTestEntry("old.txt", 100, false, lastWeek),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_DateFunctions(t *testing.T) {
	// Create a file with known modification time
	modTime := time.Date(2024, 12, 25, 10, 30, 0, 0, time.Local)

	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "year(mtime) = 2024",
			query:    "year(mtime) = 2024",
			entry:    createTestEntry("test.txt", 100, false, modTime),
			expected: true,
		},
		{
			name:     "month(mtime) = 12",
			query:    "month(mtime) = 12",
			entry:    createTestEntry("test.txt", 100, false, modTime),
			expected: true,
		},
		{
			name:     "day(mtime) = 25",
			query:    "day(mtime) = 25",
			entry:    createTestEntry("test.txt", 100, false, modTime),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_StringFunctions(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "lower(name) = 'readme.md'",
			query:    "lower(name) = 'readme.md'",
			entry:    createTestEntry("README.md", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "upper(ext) = 'TXT'",
			query:    "upper(ext) = 'TXT'",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_BooleanColumns(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "isdir - true",
			query:    "isdir",
			entry:    createTestEntry("mydir", 0, true, time.Now()),
			expected: true,
		},
		{
			name:     "isdir - false",
			query:    "isdir",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: false,
		},
		{
			name:     "isfile - true",
			query:    "isfile",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "NOT isdir",
			query:    "NOT isdir",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:  "issymlink",
			query: "issymlink",
			entry: func() internalfs.FileEntry {
				e := createTestEntry("link", 0, false, time.Now())
				e.IsSymlink = true
				e.LinkTarget = "/tmp/target"
				return e
			}(),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_INOperator(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "ext IN ('jpg', 'png') - match",
			query:    "ext IN ('jpg', 'png', 'gif')",
			entry:    createTestEntry("photo.jpg", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "ext IN ('jpg', 'png') - no match",
			query:    "ext IN ('jpg', 'png', 'gif')",
			entry:    createTestEntry("doc.pdf", 100, false, time.Now()),
			expected: false,
		},
		{
			name:     "type NOT IN ('dir', 'symlink')",
			query:    "type NOT IN ('dir', 'symlink')",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_NULLHandling(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "ext IS NULL - file with no extension",
			query:    "ext IS NULL",
			entry:    createTestEntry(".gitignore", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "ext IS NULL - file with extension",
			query:    "ext IS NULL",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: false,
		},
		{
			name:     "ext IS NOT NULL - file with extension",
			query:    "ext IS NOT NULL",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "ext IS NOT NULL - file with no extension",
			query:    "ext IS NOT NULL",
			entry:    createTestEntry(".gitignore", 100, false, time.Now()),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_LogicalCombinations(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "AND - both true",
			query:    "size > 0 AND type = 'file'",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "AND - first false",
			query:    "size > 1000 AND type = 'file'",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: false,
		},
		{
			name:     "OR - first true",
			query:    "ext = 'txt' OR ext = 'md'",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "OR - second true",
			query:    "ext = 'txt' OR ext = 'md'",
			entry:    createTestEntry("README.md", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "OR - both false",
			query:    "ext = 'txt' OR ext = 'md'",
			entry:    createTestEntry("main.go", 100, false, time.Now()),
			expected: false,
		},
		{
			name:     "complex: (size > 1GiB OR name LIKE '%.mp4') AND NOT isdir",
			query:    "(size > 1GiB OR name LIKE '%.mp4') AND NOT isdir",
			entry:    createTestEntry("video.mp4", 500*1024*1024, false, time.Now()),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_EmptyQuery(t *testing.T) {
	entry := createTestEntry("test.txt", 100, false, time.Now())
	result, err := Evaluate("", entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty query should match all entries
	if !result {
		t.Error("expected empty query to match all entries")
	}
}

func TestEvaluator_ExtColumn(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected any // string or NullValue
	}{
		{"simple extension", "test.txt", "txt"},
		{"double extension", "archive.tar.gz", "gz"},
		{"no extension", "Makefile", NullValue{}},
		{"dot file", ".gitignore", NullValue{}},
		{"dot file with ext", ".bashrc", NullValue{}},
		{"hidden with ext", ".config.yaml", "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := createTestEntry(tt.filename, 100, false, time.Now())
			ext := getExtension(entry.Name)

			switch expected := tt.expected.(type) {
			case string:
				if IsNull(ext) {
					t.Errorf("expected %q, got NULL", expected)
				} else if ext != expected {
					t.Errorf("expected %q, got %q", expected, ext)
				}
			case NullValue:
				if !IsNull(ext) {
					t.Errorf("expected NULL, got %q", ext)
				}
			}
		})
	}
}

// Additional tests for coverage improvement

func TestTokenType_String(t *testing.T) {
	tests := []struct {
		token    TokenType
		expected string
	}{
		{TokenEOF, "EOF"},
		{TokenIdent, "IDENT"},
		{TokenString, "STRING"},
		{TokenNumber, "NUMBER"},
		{TokenSizeUnit, "SIZE_UNIT"},
		{TokenDuration, "DURATION"},
		{TokenLParen, "LPAREN"},
		{TokenRParen, "RPAREN"},
		{TokenComma, "COMMA"},
		{TokenMinus, "MINUS"},
		{TokenEQ, "EQ"},
		{TokenNE, "NE"},
		{TokenNE2, "NE2"},
		{TokenLT, "LT"},
		{TokenGT, "GT"},
		{TokenLE, "LE"},
		{TokenGE, "GE"},
		{TokenAND, "AND"},
		{TokenOR, "OR"},
		{TokenNOT, "NOT"},
		{TokenLIKE, "LIKE"},
		{TokenILIKE, "ILIKE"},
		{TokenIN, "IN"},
		{TokenIS, "IS"},
		{TokenNULL, "NULL"},
		{TokenWHERE, "WHERE"},
		{TokenType(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.token.String(); got != tt.expected {
				t.Errorf("TokenType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEvaluator_TimeComparison(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	weekAgo := now.Add(-7 * 24 * time.Hour)

	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "mtime > week ago - recent file",
			query:    "mtime > now() - 7d",
			entry:    createTestEntry("recent.txt", 100, false, yesterday),
			expected: true,
		},
		{
			name:     "mtime < now - true",
			query:    "mtime < now()",
			entry:    createTestEntry("past.txt", 100, false, yesterday),
			expected: true,
		},
		{
			name:     "mtime >= specific date",
			query:    "mtime >= '2020-01-01'",
			entry:    createTestEntry("new.txt", 100, false, now),
			expected: true,
		},
		{
			name:     "mtime <= week ago",
			query:    "mtime <= now() - 7d",
			entry:    createTestEntry("old.txt", 100, false, weekAgo.Add(-time.Hour)),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_StringComparison(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "name < 'z' - true",
			query:    "name < 'zzz'",
			entry:    createTestEntry("abc.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "name > 'a' - true",
			query:    "name > 'a'",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "name <= 'test.txt' - equal",
			query:    "name <= 'test.txt'",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
		{
			name:     "name >= 'test.txt' - equal",
			query:    "name >= 'test.txt'",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_Coercion(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "size comparison with float",
			query:    "size > 99.5",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_DateTimeFunctions(t *testing.T) {
	// Create entry with known date
	knownTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.Local)
	entry := createTestEntry("test.txt", 100, false, knownTime)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"year function", "year(mtime) = 2024", true},
		{"month function", "month(mtime) = 6", true},
		{"day function", "day(mtime) = 15", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_ErrorCases(t *testing.T) {
	entry := createTestEntry("test.txt", 100, false, time.Now())

	tests := []struct {
		name  string
		query string
	}{
		{"invalid column", "invalidcol = 'test'"},
		{"invalid function", "invalidfunc() = 1"},
		{"unclosed string", "name = 'test"},
		{"unexpected token", "size > > 100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.query, entry)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestEvaluator_FileTypes(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "type = 'dir' for directory",
			query:    "type = 'dir'",
			entry:    createTestEntry("mydir", 0, true, now),
			expected: true,
		},
		{
			name:     "type = 'file' for file",
			query:    "type = 'file'",
			entry:    createTestEntry("test.txt", 100, false, now),
			expected: true,
		},
		{
			name:     "isdir for directory",
			query:    "isdir",
			entry:    createTestEntry("mydir", 0, true, now),
			expected: true,
		},
		{
			name:     "isfile for file",
			query:    "isfile",
			entry:    createTestEntry("test.txt", 100, false, now),
			expected: true,
		},
		{
			name:     "NOT isdir for file",
			query:    "NOT isdir",
			entry:    createTestEntry("test.txt", 100, false, now),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_DurationSubtraction(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		query    string
		entry    internalfs.FileEntry
		expected bool
	}{
		{
			name:     "now() - 1h",
			query:    "mtime > now() - 1h",
			entry:    createTestEntry("recent.txt", 100, false, now.Add(-30*time.Minute)),
			expected: true,
		},
		{
			name:     "now() - 30m",
			query:    "mtime > now() - 30m",
			entry:    createTestEntry("recent.txt", 100, false, now.Add(-15*time.Minute)),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCoercion_Functions(t *testing.T) {
	// Test coerceToInt64
	t.Run("coerceToInt64", func(t *testing.T) {
		testCases := []struct {
			input    any
			expected int64
			hasError bool
		}{
			{int64(42), 42, false},
			{int(42), 42, false},
			{float64(42.9), 42, false},
			{"42", 42, false},
			{"invalid", 0, true},
			{struct{}{}, 0, true},
		}

		for _, tc := range testCases {
			result, err := coerceToInt64(tc.input)
			if tc.hasError {
				if err == nil {
					t.Errorf("expected error for input %v", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %v: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("expected %d, got %d for input %v", tc.expected, result, tc.input)
				}
			}
		}
	})

	// Test coerceToFloat64
	t.Run("coerceToFloat64", func(t *testing.T) {
		testCases := []struct {
			input    any
			expected float64
			hasError bool
		}{
			{float64(42.5), 42.5, false},
			{int64(42), 42.0, false},
			{int(42), 42.0, false},
			{"42.5", 42.5, false},
			{"invalid", 0, true},
			{struct{}{}, 0, true},
		}

		for _, tc := range testCases {
			result, err := coerceToFloat64(tc.input)
			if tc.hasError {
				if err == nil {
					t.Errorf("expected error for input %v", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %v: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("expected %f, got %f for input %v", tc.expected, result, tc.input)
				}
			}
		}
	})

	// Test coerceToString
	t.Run("coerceToString", func(t *testing.T) {
		testCases := []struct {
			input    any
			expected string
			hasError bool
		}{
			{"hello", "hello", false},
			{int64(42), "42", false},
			{int(42), "42", false},
			{float64(3.14), "3.14", false},
			{true, "true", false},
			{false, "false", false},
			{struct{}{}, "", true},
		}

		for _, tc := range testCases {
			result, err := coerceToString(tc.input)
			if tc.hasError {
				if err == nil {
					t.Errorf("expected error for input %v", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %v: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("expected %q, got %q for input %v", tc.expected, result, tc.input)
				}
			}
		}
	})

	// Test coerceToBool
	t.Run("coerceToBool", func(t *testing.T) {
		testCases := []struct {
			input    any
			expected bool
			hasError bool
		}{
			{true, true, false},
			{false, false, false},
			{int64(1), true, false},
			{int64(0), false, false},
			{int(1), true, false},
			{int(0), false, false},
			{"true", true, false},
			{"false", false, false},
			{"1", true, false},
			{"0", false, false},
			{"yes", true, false},
			{"no", false, false},
			{"invalid", false, true},
			{struct{}{}, false, true},
		}

		for _, tc := range testCases {
			result, err := coerceToBool(tc.input)
			if tc.hasError {
				if err == nil {
					t.Errorf("expected error for input %v", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %v: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("expected %v, got %v for input %v", tc.expected, result, tc.input)
				}
			}
		}
	})

	// Test coerceToTime
	t.Run("coerceToTime", func(t *testing.T) {
		now := time.Now()
		testCases := []struct {
			input    any
			hasError bool
		}{
			{now, false},
			{"2024-01-15", false},
			{"2024-01-15 10:30:00", false},
			{"invalid-date", true},
			{struct{}{}, true},
		}

		for _, tc := range testCases {
			_, err := coerceToTime(tc.input)
			if tc.hasError {
				if err == nil {
					t.Errorf("expected error for input %v", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %v: %v", tc.input, err)
				}
			}
		}
	})
}

func TestParser_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		hasError bool
	}{
		{"WHERE keyword", "WHERE size > 100", false},
		{"parenthesized expression", "(size > 100)", false},
		{"nested parentheses", "((size > 100))", false},
		{"complex expression", "(size > 100 AND name LIKE '%.txt') OR isdir", false},
		{"float number", "size > 99.5", false},
		{"ILIKE case insensitive", "name ILIKE '%.TXT'", false},
		{"NOT IN", "ext NOT IN ('exe', 'dll')", false},
		{"missing closing paren", "(size > 100", true},
		{"empty IN list - parse error", "ext IN ()", true},
		{"duration literal", "mtime > now() - 7d", false},
	}

	entry := createTestEntry("test.txt", 100, false, time.Now())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.query, entry)
			if tt.hasError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestEvaluator_SymlinkEntry(t *testing.T) {
	now := time.Now()
	entry := internalfs.FileEntry{
		Name:        "link.txt",
		IsDir:       false,
		Size:        0,
		ModTime:     now,
		Permissions: 0777,
		Owner:       "user",
		Group:       "group",
		IsSymlink:   true,
		LinkTarget:  "/some/target",
		LinkBroken:  false,
	}

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"issymlink for symlink", "issymlink", true},
		{"type = 'symlink'", "type = 'symlink'", true},
		{"NOT isfile", "NOT isfile", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_NotLike(t *testing.T) {
	entry := createTestEntry("document.pdf", 100, false, time.Now())

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"NOT LIKE matching pattern", "name NOT LIKE '%.txt'", true},
		{"NOT LIKE non-matching pattern", "name NOT LIKE '%.pdf'", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_OrExpression(t *testing.T) {
	entry := createTestEntry("test.txt", 100, false, time.Now())

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"OR both false", "size > 1000 OR isdir", false},
		{"OR first true", "size < 1000 OR isdir", true},
		{"OR second true", "size > 1000 OR isfile", true},
		{"OR both true", "size < 1000 OR isfile", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_PermColumn(t *testing.T) {
	entry := createTestEntry("test.txt", 100, false, time.Now())

	// perm column returns permission bits as string
	result, err := Evaluate("perm = '-rw-r--r--'", entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected perm comparison to match")
	}
}

func TestEvaluator_OwnerGroupColumns(t *testing.T) {
	entry := createTestEntry("test.txt", 100, false, time.Now())

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"owner = 'user'", "owner = 'user'", true},
		{"group = 'group'", "group = 'group'", true},
		{"owner != 'root'", "owner != 'root'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLexer_AllSizeUnits(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"100KiB", "KiB"},
		{"100MiB", "MiB"},
		{"100GiB", "GiB"},
		{"100TiB", "TiB"},
		{"100KB", "KB"},
		{"100MB", "MB"},
		{"100GB", "GB"},
		{"100TB", "TB"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Should have NUMBER, SIZE_UNIT, EOF
			if len(tokens) != 3 {
				t.Fatalf("expected 3 tokens, got %d", len(tokens))
			}
			if tokens[1].Type != TokenSizeUnit {
				t.Errorf("expected SIZE_UNIT, got %v", tokens[1].Type)
			}
			if tokens[1].Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tokens[1].Value)
			}
		})
	}
}

func TestEvaluator_FunctionArgErrors(t *testing.T) {
	entry := createTestEntry("test.txt", 100, false, time.Now())

	// Test function with wrong argument type
	tests := []struct {
		name  string
		query string
	}{
		{"year with non-time", "year('notadate') = 2024"},
		{"month with non-time", "month('notadate') = 6"},
		{"day with non-time", "day('notadate') = 15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.query, entry)
			if err == nil {
				t.Error("expected error for invalid function argument")
			}
		})
	}
}

func TestFilter_CompileAndMatch(t *testing.T) {
	// Test compiling and reusing a query
	query, err := CompileQuery("size > 50 AND ext = 'txt'")
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	entry1 := createTestEntry("test.txt", 100, false, time.Now())
	entry2 := createTestEntry("test.txt", 30, false, time.Now())
	entry3 := createTestEntry("test.pdf", 100, false, time.Now())

	if !query.Match(entry1) {
		t.Error("expected entry1 to match")
	}
	if query.Match(entry2) {
		t.Error("expected entry2 not to match (size too small)")
	}
	if query.Match(entry3) {
		t.Error("expected entry3 not to match (wrong extension)")
	}
}

func TestFilter_ValidateQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		hasError bool
	}{
		{"valid query", "size > 100", false},
		{"invalid column", "badcol = 'test'", true},
		{"syntax error", "size >", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQuery(tt.query)
			if tt.hasError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestEvaluator_LowerUpperFunctions(t *testing.T) {
	entry := createTestEntry("Test.TXT", 100, false, time.Now())

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"lower function", "lower(name) = 'test.txt'", true},
		{"upper function", "upper(name) = 'TEST.TXT'", true},
		{"lower in LIKE", "lower(name) LIKE '%.txt'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_CompareIntegers(t *testing.T) {
	entry := createTestEntry("test.txt", 100, false, time.Now())

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"size < 200", "size < 200", true},
		{"size > 50", "size > 50", true},
		{"size <= 100", "size <= 100", true},
		{"size >= 100", "size >= 100", true},
		{"size = 100", "size = 100", true},
		{"size != 200", "size != 200", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_NowFunction(t *testing.T) {
	entry := createTestEntry("test.txt", 100, false, time.Now().Add(-time.Hour))

	// mtime should be less than now()
	result, err := Evaluate("mtime < now()", entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected mtime < now() to be true")
	}
}

func TestParser_OperandTypes(t *testing.T) {
	entry := createTestEntry("test.txt", 100, false, time.Now())

	tests := []struct {
		name     string
		query    string
		hasError bool
	}{
		{"string literal", "name = 'test.txt'", false},
		{"number literal", "size > 50", false},
		{"duration literal in comparison", "mtime > now() - 7d", false},
		{"function call", "lower(name) = 'test.txt'", false},
		{"column reference", "ext = 'txt'", false},
		{"function with args", "year(mtime) = 2024", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.query, entry)
			if tt.hasError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestEvaluator_InExprWithMultipleTypes(t *testing.T) {
	tests := []struct {
		name     string
		entry    internalfs.FileEntry
		query    string
		expected bool
	}{
		{
			name:     "ext IN list - match",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			query:    "ext IN ('txt', 'pdf', 'doc')",
			expected: true,
		},
		{
			name:     "ext IN list - no match",
			entry:    createTestEntry("test.go", 100, false, time.Now()),
			query:    "ext IN ('txt', 'pdf', 'doc')",
			expected: false,
		},
		{
			name:     "name IN list",
			entry:    createTestEntry("README.md", 100, false, time.Now()),
			query:    "name IN ('README.md', 'LICENSE', 'CHANGELOG.md')",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_IsNullVariations(t *testing.T) {
	tests := []struct {
		name     string
		entry    internalfs.FileEntry
		query    string
		expected bool
	}{
		{
			name:     "ext IS NULL for no extension",
			entry:    createTestEntry("Makefile", 100, false, time.Now()),
			query:    "ext IS NULL",
			expected: true,
		},
		{
			name:     "ext IS NOT NULL for extension",
			entry:    createTestEntry("test.txt", 100, false, time.Now()),
			query:    "ext IS NOT NULL",
			expected: true,
		},
		{
			name:     "ext IS NULL for dot file",
			entry:    createTestEntry(".gitignore", 100, false, time.Now()),
			query:    "ext IS NULL",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.query, tt.entry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
