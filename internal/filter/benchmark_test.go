package filter

import (
	"os"
	"testing"
	"time"

	"github.com/sakura/duofm/internal/fs"
)

// generateTestEntries creates a slice of test FileEntry for benchmarking.
func generateTestEntries(count int) []fs.FileEntry {
	entries := make([]fs.FileEntry, count)
	extensions := []string{"txt", "go", "md", "jpg", "png", "pdf", ""}
	types := []struct {
		isDir     bool
		isSymlink bool
	}{
		{false, false}, // file
		{true, false},  // dir
		{false, true},  // symlink
	}

	for i := 0; i < count; i++ {
		ext := extensions[i%len(extensions)]
		t := types[i%len(types)]
		name := ""
		if ext == "" {
			name = "file"
		} else {
			name = "file." + ext
		}

		entries[i] = fs.FileEntry{
			Name:        name,
			Size:        int64(i * 1024), // varying sizes
			ModTime:     time.Now().Add(-time.Duration(i) * time.Hour),
			IsDir:       t.isDir,
			IsSymlink:   t.isSymlink,
			Permissions: os.FileMode(0644),
			Owner:       "user",
			Group:       "group",
		}
	}
	return entries
}

// BenchmarkLexer_SimpleQuery benchmarks lexing a simple query.
func BenchmarkLexer_SimpleQuery(b *testing.B) {
	query := "size > 1MiB"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexer := NewLexer(query)
		_, _ = lexer.Tokenize()
	}
}

// BenchmarkLexer_ComplexQuery benchmarks lexing a complex query.
func BenchmarkLexer_ComplexQuery(b *testing.B) {
	query := "(size > 1GiB OR name LIKE '%.mp4') AND NOT isdir AND mtime > now() - 7d AND ext IN ('mp4', 'mkv', 'avi')"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexer := NewLexer(query)
		_, _ = lexer.Tokenize()
	}
}

// BenchmarkParse_SimpleQuery benchmarks parsing a simple query.
func BenchmarkParse_SimpleQuery(b *testing.B) {
	query := "size > 1MiB"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(query)
	}
}

// BenchmarkParse_ComplexQuery benchmarks parsing a complex query.
func BenchmarkParse_ComplexQuery(b *testing.B) {
	query := "(size > 1GiB OR name LIKE '%.mp4') AND NOT isdir AND mtime > now() - 7d"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(query)
	}
}

// BenchmarkCompileQuery benchmarks compiling a query (parse + wrap).
func BenchmarkCompileQuery(b *testing.B) {
	query := "size > 1MiB AND NOT isdir"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CompileQuery(query)
	}
}

// BenchmarkEvaluate_Simple benchmarks evaluating a simple query against one entry.
func BenchmarkEvaluate_Simple(b *testing.B) {
	entry := fs.FileEntry{
		Name:    "test.txt",
		Size:    2 * 1024 * 1024, // 2 MiB
		ModTime: time.Now(),
	}
	query := "size > 1MiB"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Evaluate(query, entry)
	}
}

// BenchmarkCompiledQuery_Match benchmarks matching with a pre-compiled query.
func BenchmarkCompiledQuery_Match(b *testing.B) {
	entry := fs.FileEntry{
		Name:    "test.txt",
		Size:    2 * 1024 * 1024, // 2 MiB
		ModTime: time.Now(),
	}
	compiled, _ := CompileQuery("size > 1MiB")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = compiled.Match(entry)
	}
}

// BenchmarkFilter_10K benchmarks filtering 10,000 entries.
func BenchmarkFilter_10K(b *testing.B) {
	entries := generateTestEntries(10000)
	query := "size > 1MiB AND NOT isdir"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FilterSQLLike(entries, query)
	}
}

// BenchmarkFilter_100K benchmarks filtering 100,000 entries.
func BenchmarkFilter_100K(b *testing.B) {
	entries := generateTestEntries(100000)
	query := "size > 1MiB AND NOT isdir"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FilterSQLLike(entries, query)
	}
}

// BenchmarkFilter_WithLIKE benchmarks filtering with LIKE pattern.
func BenchmarkFilter_WithLIKE(b *testing.B) {
	entries := generateTestEntries(10000)
	query := "name LIKE '%.txt'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FilterSQLLike(entries, query)
	}
}

// BenchmarkFilter_WithIN benchmarks filtering with IN clause.
func BenchmarkFilter_WithIN(b *testing.B) {
	entries := generateTestEntries(10000)
	query := "ext IN ('txt', 'go', 'md')"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FilterSQLLike(entries, query)
	}
}

// BenchmarkFilter_Complex benchmarks filtering with a complex query.
func BenchmarkFilter_Complex(b *testing.B) {
	entries := generateTestEntries(10000)
	query := "(size > 1MiB OR name LIKE '%.go') AND NOT isdir AND mtime > now() - 7d"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FilterSQLLike(entries, query)
	}
}

// BenchmarkFilter_CompiledReuse benchmarks filtering with pre-compiled query.
func BenchmarkFilter_CompiledReuse(b *testing.B) {
	entries := generateTestEntries(10000)
	compiled, _ := CompileQuery("size > 1MiB AND NOT isdir")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = compiled.Filter(entries)
	}
}
