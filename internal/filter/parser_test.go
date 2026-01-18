package filter

import (
	"testing"
)

func TestParser_SimpleComparison(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkAST func(t *testing.T, expr Expression)
	}{
		{
			name:  "size > 100",
			input: "size > 100",
			checkAST: func(t *testing.T, expr Expression) {
				bin, ok := expr.(*BinaryExpr)
				if !ok {
					t.Fatalf("expected BinaryExpr, got %T", expr)
				}
				if bin.Operator != ">" {
					t.Errorf("expected operator '>', got %q", bin.Operator)
				}
				col, ok := bin.Left.(*ColumnRef)
				if !ok {
					t.Fatalf("expected ColumnRef for left, got %T", bin.Left)
				}
				if col.Name != "size" {
					t.Errorf("expected column 'size', got %q", col.Name)
				}
				lit, ok := bin.Right.(*Literal)
				if !ok {
					t.Fatalf("expected Literal for right, got %T", bin.Right)
				}
				if lit.Value != int64(100) {
					t.Errorf("expected value 100, got %v", lit.Value)
				}
			},
		},
		{
			name:  "name = 'test.txt'",
			input: "name = 'test.txt'",
			checkAST: func(t *testing.T, expr Expression) {
				bin, ok := expr.(*BinaryExpr)
				if !ok {
					t.Fatalf("expected BinaryExpr, got %T", expr)
				}
				if bin.Operator != "=" {
					t.Errorf("expected operator '=', got %q", bin.Operator)
				}
				col, ok := bin.Left.(*ColumnRef)
				if !ok {
					t.Fatalf("expected ColumnRef for left, got %T", bin.Left)
				}
				if col.Name != "name" {
					t.Errorf("expected column 'name', got %q", col.Name)
				}
				lit, ok := bin.Right.(*Literal)
				if !ok {
					t.Fatalf("expected Literal for right, got %T", bin.Right)
				}
				if lit.Value != "test.txt" {
					t.Errorf("expected value 'test.txt', got %v", lit.Value)
				}
			},
		},
		{
			name:  "size != 0",
			input: "size != 0",
			checkAST: func(t *testing.T, expr Expression) {
				bin, ok := expr.(*BinaryExpr)
				if !ok {
					t.Fatalf("expected BinaryExpr, got %T", expr)
				}
				if bin.Operator != "!=" {
					t.Errorf("expected operator '!=', got %q", bin.Operator)
				}
			},
		},
		{
			name:  "size <> 0",
			input: "size <> 0",
			checkAST: func(t *testing.T, expr Expression) {
				bin, ok := expr.(*BinaryExpr)
				if !ok {
					t.Fatalf("expected BinaryExpr, got %T", expr)
				}
				if bin.Operator != "<>" {
					t.Errorf("expected operator '<>', got %q", bin.Operator)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.checkAST(t, expr)
		})
	}
}

func TestParser_SizeLiterals(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedSize int64
	}{
		{"1KiB", "size > 1KiB", 1024},
		{"1MiB", "size > 1MiB", 1024 * 1024},
		{"1GiB", "size > 1GiB", 1024 * 1024 * 1024},
		{"1TiB", "size > 1TiB", 1024 * 1024 * 1024 * 1024},
		{"1KB", "size > 1KB", 1000},
		{"1MB", "size > 1MB", 1000 * 1000},
		{"1GB", "size > 1GB", 1000 * 1000 * 1000},
		{"1TB", "size > 1TB", 1000 * 1000 * 1000 * 1000},
		{"100MiB", "size > 100MiB", 100 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			bin, ok := expr.(*BinaryExpr)
			if !ok {
				t.Fatalf("expected BinaryExpr, got %T", expr)
			}
			sizeLit, ok := bin.Right.(*SizeLiteral)
			if !ok {
				t.Fatalf("expected SizeLiteral for right, got %T", bin.Right)
			}
			if sizeLit.Value != tt.expectedSize {
				t.Errorf("expected size %d, got %d", tt.expectedSize, sizeLit.Value)
			}
		})
	}
}

func TestParser_ANDExpression(t *testing.T) {
	input := "size > 100 AND type = 'file'"
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bin, ok := expr.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", expr)
	}
	if bin.Operator != "AND" {
		t.Errorf("expected operator 'AND', got %q", bin.Operator)
	}

	// Left side should be size > 100
	leftBin, ok := bin.Left.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr for left, got %T", bin.Left)
	}
	if leftBin.Operator != ">" {
		t.Errorf("expected left operator '>', got %q", leftBin.Operator)
	}

	// Right side should be type = 'file'
	rightBin, ok := bin.Right.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr for right, got %T", bin.Right)
	}
	if rightBin.Operator != "=" {
		t.Errorf("expected right operator '=', got %q", rightBin.Operator)
	}
}

func TestParser_ORExpression(t *testing.T) {
	input := "ext = 'txt' OR ext = 'md'"
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bin, ok := expr.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", expr)
	}
	if bin.Operator != "OR" {
		t.Errorf("expected operator 'OR', got %q", bin.Operator)
	}
}

func TestParser_NOTExpression(t *testing.T) {
	input := "NOT isdir"
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	unary, ok := expr.(*UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", expr)
	}
	if unary.Operator != "NOT" {
		t.Errorf("expected operator 'NOT', got %q", unary.Operator)
	}

	col, ok := unary.Operand.(*ColumnRef)
	if !ok {
		t.Fatalf("expected ColumnRef for operand, got %T", unary.Operand)
	}
	if col.Name != "isdir" {
		t.Errorf("expected column 'isdir', got %q", col.Name)
	}
}

func TestParser_Parentheses(t *testing.T) {
	input := "(size > 1GiB OR name LIKE '%.mp4') AND NOT isdir"
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Top level should be AND
	bin, ok := expr.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", expr)
	}
	if bin.Operator != "AND" {
		t.Errorf("expected operator 'AND', got %q", bin.Operator)
	}

	// Left should be the grouped OR expression
	leftBin, ok := bin.Left.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr for left, got %T", bin.Left)
	}
	if leftBin.Operator != "OR" {
		t.Errorf("expected left operator 'OR', got %q", leftBin.Operator)
	}

	// Right should be NOT isdir
	rightUnary, ok := bin.Right.(*UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr for right, got %T", bin.Right)
	}
	if rightUnary.Operator != "NOT" {
		t.Errorf("expected right operator 'NOT', got %q", rightUnary.Operator)
	}
}

func TestParser_OperatorPrecedence(t *testing.T) {
	// NOT > AND > OR
	// "a OR b AND NOT c" should parse as "a OR (b AND (NOT c))"
	input := "isfile OR isdir AND NOT issymlink"
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Top level should be OR
	bin, ok := expr.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", expr)
	}
	if bin.Operator != "OR" {
		t.Errorf("expected top level 'OR', got %q", bin.Operator)
	}

	// Right side of OR should be AND
	rightBin, ok := bin.Right.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr for right, got %T", bin.Right)
	}
	if rightBin.Operator != "AND" {
		t.Errorf("expected right operator 'AND', got %q", rightBin.Operator)
	}

	// Right side of AND should be NOT
	rightOfAnd, ok := rightBin.Right.(*UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr for right of AND, got %T", rightBin.Right)
	}
	if rightOfAnd.Operator != "NOT" {
		t.Errorf("expected NOT operator, got %q", rightOfAnd.Operator)
	}
}

func TestParser_LIKEPattern(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		caseInsensitive bool
		negated         bool
	}{
		{"LIKE", "name LIKE '%.txt'", false, false},
		{"ILIKE", "name ILIKE '%.TXT'", true, false},
		{"NOT LIKE", "name NOT LIKE '%.bak'", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			like, ok := expr.(*LikeExpr)
			if !ok {
				t.Fatalf("expected LikeExpr, got %T", expr)
			}
			if like.CaseInsensitive != tt.caseInsensitive {
				t.Errorf("expected caseInsensitive=%v, got %v", tt.caseInsensitive, like.CaseInsensitive)
			}
			if like.Negated != tt.negated {
				t.Errorf("expected negated=%v, got %v", tt.negated, like.Negated)
			}
		})
	}
}

func TestParser_INList(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		negated bool
		count   int
	}{
		{"IN list", "ext IN ('jpg', 'png', 'gif')", false, 3},
		{"NOT IN list", "type NOT IN ('dir', 'symlink')", true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			inExpr, ok := expr.(*InExpr)
			if !ok {
				t.Fatalf("expected InExpr, got %T", expr)
			}
			if inExpr.Negated != tt.negated {
				t.Errorf("expected negated=%v, got %v", tt.negated, inExpr.Negated)
			}
			if len(inExpr.Values) != tt.count {
				t.Errorf("expected %d values, got %d", tt.count, len(inExpr.Values))
			}
		})
	}
}

func TestParser_ISNull(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		negated bool
	}{
		{"IS NULL", "ext IS NULL", false},
		{"IS NOT NULL", "ext IS NOT NULL", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			isNull, ok := expr.(*IsNullExpr)
			if !ok {
				t.Fatalf("expected IsNullExpr, got %T", expr)
			}
			if isNull.Negated != tt.negated {
				t.Errorf("expected negated=%v, got %v", tt.negated, isNull.Negated)
			}
		})
	}
}

func TestParser_FunctionCall(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		funcName   string
		argCount   int
		funcOnLeft bool // true if function is on left side of comparison
	}{
		{"now() on right", "mtime > now()", "now", 0, false},
		{"year(mtime)", "year(mtime) = 2024", "year", 1, true},
		{"month(mtime)", "month(mtime) = 12", "month", 1, true},
		{"day(mtime)", "day(mtime) = 25", "day", 1, true},
		{"lower(name)", "lower(name) = 'readme.md'", "lower", 1, true},
		{"upper(ext)", "upper(ext) = 'TXT'", "upper", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			bin, ok := expr.(*BinaryExpr)
			if !ok {
				t.Fatalf("expected BinaryExpr, got %T", expr)
			}

			var fn *FunctionCall
			if tt.funcOnLeft {
				fn, ok = bin.Left.(*FunctionCall)
				if !ok {
					t.Fatalf("expected FunctionCall for left, got %T", bin.Left)
				}
			} else {
				fn, ok = bin.Right.(*FunctionCall)
				if !ok {
					t.Fatalf("expected FunctionCall for right, got %T", bin.Right)
				}
			}
			if fn.Name != tt.funcName {
				t.Errorf("expected function %q, got %q", tt.funcName, fn.Name)
			}
			if len(fn.Args) != tt.argCount {
				t.Errorf("expected %d args, got %d", tt.argCount, len(fn.Args))
			}
		})
	}
}

func TestParser_NowMinusDuration(t *testing.T) {
	input := "mtime > now() - 7d"
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bin, ok := expr.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", expr)
	}
	if bin.Operator != ">" {
		t.Errorf("expected operator '>', got %q", bin.Operator)
	}

	// Right side should be now() - 7d (BinaryExpr with MINUS)
	rightBin, ok := bin.Right.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr for right, got %T", bin.Right)
	}
	if rightBin.Operator != "-" {
		t.Errorf("expected operator '-', got %q", rightBin.Operator)
	}

	// Left of minus should be now()
	fn, ok := rightBin.Left.(*FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall for left of minus, got %T", rightBin.Left)
	}
	if fn.Name != "now" {
		t.Errorf("expected function 'now', got %q", fn.Name)
	}

	// Right of minus should be duration
	dur, ok := rightBin.Right.(*DurationLiteral)
	if !ok {
		t.Fatalf("expected DurationLiteral for right of minus, got %T", rightBin.Right)
	}
	if dur.Value != 7 || dur.Unit != "d" {
		t.Errorf("expected 7d, got %d%s", dur.Value, dur.Unit)
	}
}

func TestParser_WHEREKeyword(t *testing.T) {
	// WHERE keyword should be optional
	tests := []struct {
		name  string
		input string
	}{
		{"without WHERE", "size > 0"},
		{"with WHERE", "WHERE size > 0"},
		{"with where lowercase", "where size > 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			bin, ok := expr.(*BinaryExpr)
			if !ok {
				t.Fatalf("expected BinaryExpr, got %T", expr)
			}
			if bin.Operator != ">" {
				t.Errorf("expected operator '>', got %q", bin.Operator)
			}
		})
	}
}

func TestParser_BooleanColumn(t *testing.T) {
	// Boolean columns like isdir, isfile, issymlink can be used alone
	tests := []struct {
		name   string
		input  string
		column string
	}{
		{"isdir", "isdir", "isdir"},
		{"isfile", "isfile", "isfile"},
		{"issymlink", "issymlink", "issymlink"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			col, ok := expr.(*ColumnRef)
			if !ok {
				t.Fatalf("expected ColumnRef, got %T", expr)
			}
			if col.Name != tt.column {
				t.Errorf("expected column %q, got %q", tt.column, col.Name)
			}
		})
	}
}

func TestParser_Errors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "missing operand",
			input:       "size >",
			expectedErr: "unexpected end of input",
		},
		{
			name:        "unbalanced parentheses - missing close",
			input:       "(size > 100",
			expectedErr: "expected ')'",
		},
		{
			name:        "unbalanced parentheses - extra close",
			input:       "size > 100)",
			expectedErr: "unexpected token",
		},
		{
			name:        "unknown column",
			input:       "unknown > 100",
			expectedErr: "unknown column",
		},
		{
			name:        "invalid IN list - missing values",
			input:       "ext IN ()",
			expectedErr: "expected",
		},
		{
			name:        "invalid IN list - missing close paren",
			input:       "ext IN ('txt'",
			expectedErr: "expected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !containsString(err.Error(), tt.expectedErr) {
				t.Errorf("expected error containing %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestParser_ComplexQueries(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "complex filter with multiple conditions",
			input: "size > 1GiB AND year(mtime) = 2024 AND name LIKE '%.mp4'",
		},
		{
			name:  "nested parentheses",
			input: "((size > 1GiB) AND (type = 'file'))",
		},
		{
			name:  "mixed AND/OR with NOT",
			input: "isfile AND (ext = 'txt' OR ext = 'md') AND NOT name LIKE '%backup%'",
		},
		{
			name:  "ext IN with AND",
			input: "ext IN ('jpg', 'png', 'gif') AND size < 10MiB",
		},
		{
			name:  "date comparison with function",
			input: "mtime > now() - 30d AND mtime < now() - 7d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestParser_EmptyQuery(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Empty query should return nil expression
			if expr != nil {
				t.Errorf("expected nil expression for empty query, got %T", expr)
			}
		})
	}
}
