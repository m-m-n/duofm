package filter

import (
	"github.com/sakura/duofm/internal/fs"
)

// FilterSQLLike filters entries using a SQL-like WHERE clause query.
// Returns filtered entries or an error if the query is invalid.
// An empty query returns all entries.
func FilterSQLLike(entries []fs.FileEntry, query string) ([]fs.FileEntry, error) {
	compiled, err := CompileQuery(query)
	if err != nil {
		return nil, err
	}

	return compiled.Filter(entries), nil
}

// CompiledQuery represents a parsed and validated SQL-like query.
// It can be reused to filter multiple entry sets efficiently.
type CompiledQuery struct {
	ast      Expression
	rawQuery string
}

// CompileQuery parses and validates a SQL-like WHERE clause query.
// Returns a CompiledQuery that can be used to filter entries.
func CompileQuery(query string) (*CompiledQuery, error) {
	ast, err := Parse(query)
	if err != nil {
		return nil, err
	}

	return &CompiledQuery{
		ast:      ast,
		rawQuery: query,
	}, nil
}

// Match checks if a single FileEntry matches the query.
func (q *CompiledQuery) Match(entry fs.FileEntry) bool {
	// Empty query matches all entries
	if q.ast == nil {
		return true
	}

	eval := NewEvaluator(entry)
	result, err := eval.Eval(q.ast)
	if err != nil {
		return false
	}
	return result
}

// Filter returns only the entries that match the query.
func (q *CompiledQuery) Filter(entries []fs.FileEntry) []fs.FileEntry {
	// Empty query returns all entries
	if q.ast == nil {
		return entries
	}

	result := make([]fs.FileEntry, 0, len(entries))
	for _, entry := range entries {
		if q.Match(entry) {
			result = append(result, entry)
		}
	}
	return result
}

// RawQuery returns the original query string.
func (q *CompiledQuery) RawQuery() string {
	return q.rawQuery
}

// ValidateQuery checks if a query is syntactically valid.
// Returns nil if valid, or an error describing the problem.
func ValidateQuery(query string) error {
	_, err := Parse(query)
	return err
}
