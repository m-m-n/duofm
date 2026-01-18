package filter

import (
	"testing"
)

func TestLexer_SimpleTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "single identifier",
			input: "size",
			expected: []Token{
				{Type: TokenIdent, Value: "size", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 4},
			},
		},
		{
			name:  "comparison operator >",
			input: ">",
			expected: []Token{
				{Type: TokenGT, Value: ">", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 1},
			},
		},
		{
			name:  "comparison operator <",
			input: "<",
			expected: []Token{
				{Type: TokenLT, Value: "<", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 1},
			},
		},
		{
			name:  "comparison operator =",
			input: "=",
			expected: []Token{
				{Type: TokenEQ, Value: "=", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 1},
			},
		},
		{
			name:  "comparison operator !=",
			input: "!=",
			expected: []Token{
				{Type: TokenNE, Value: "!=", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 2},
			},
		},
		{
			name:  "comparison operator <>",
			input: "<>",
			expected: []Token{
				{Type: TokenNE2, Value: "<>", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 2},
			},
		},
		{
			name:  "comparison operator >=",
			input: ">=",
			expected: []Token{
				{Type: TokenGE, Value: ">=", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 2},
			},
		},
		{
			name:  "comparison operator <=",
			input: "<=",
			expected: []Token{
				{Type: TokenLE, Value: "<=", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 2},
			},
		},
		{
			name:  "number",
			input: "100",
			expected: []Token{
				{Type: TokenNumber, Value: "100", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 3},
			},
		},
		{
			name:  "parentheses",
			input: "()",
			expected: []Token{
				{Type: TokenLParen, Value: "(", Pos: 0},
				{Type: TokenRParen, Value: ")", Pos: 1},
				{Type: TokenEOF, Value: "", Pos: 2},
			},
		},
		{
			name:  "comma",
			input: ",",
			expected: []Token{
				{Type: TokenComma, Value: ",", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 1},
			},
		},
		{
			name:  "minus",
			input: "-",
			expected: []Token{
				{Type: TokenMinus, Value: "-", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tokens) != len(tt.expected) {
				t.Fatalf("expected %d tokens, got %d", len(tt.expected), len(tokens))
			}
			for i, tok := range tokens {
				if tok.Type != tt.expected[i].Type {
					t.Errorf("token %d: expected type %v, got %v", i, tt.expected[i].Type, tok.Type)
				}
				if tok.Value != tt.expected[i].Value {
					t.Errorf("token %d: expected value %q, got %q", i, tt.expected[i].Value, tok.Value)
				}
				if tok.Pos != tt.expected[i].Pos {
					t.Errorf("token %d: expected pos %d, got %d", i, tt.expected[i].Pos, tok.Pos)
				}
			}
		})
	}
}

func TestLexer_Keywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected TokenType
	}{
		{"AND lowercase", "and", TokenAND},
		{"AND uppercase", "AND", TokenAND},
		{"AND mixed case", "And", TokenAND},
		{"OR lowercase", "or", TokenOR},
		{"OR uppercase", "OR", TokenOR},
		{"NOT lowercase", "not", TokenNOT},
		{"NOT uppercase", "NOT", TokenNOT},
		{"LIKE lowercase", "like", TokenLIKE},
		{"LIKE uppercase", "LIKE", TokenLIKE},
		{"ILIKE lowercase", "ilike", TokenILIKE},
		{"ILIKE uppercase", "ILIKE", TokenILIKE},
		{"IN lowercase", "in", TokenIN},
		{"IN uppercase", "IN", TokenIN},
		{"IS lowercase", "is", TokenIS},
		{"IS uppercase", "IS", TokenIS},
		{"NULL lowercase", "null", TokenNULL},
		{"NULL uppercase", "NULL", TokenNULL},
		{"WHERE lowercase", "where", TokenWHERE},
		{"WHERE uppercase", "WHERE", TokenWHERE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tokens) < 1 {
				t.Fatal("expected at least one token")
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected type %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestLexer_SizeUnits(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "1GiB",
			input: "1GiB",
			expected: []Token{
				{Type: TokenNumber, Value: "1", Pos: 0},
				{Type: TokenSizeUnit, Value: "GiB", Pos: 1},
				{Type: TokenEOF, Value: "", Pos: 4},
			},
		},
		{
			name:  "100MB",
			input: "100MB",
			expected: []Token{
				{Type: TokenNumber, Value: "100", Pos: 0},
				{Type: TokenSizeUnit, Value: "MB", Pos: 3},
				{Type: TokenEOF, Value: "", Pos: 5},
			},
		},
		{
			name:  "1KiB",
			input: "1KiB",
			expected: []Token{
				{Type: TokenNumber, Value: "1", Pos: 0},
				{Type: TokenSizeUnit, Value: "KiB", Pos: 1},
				{Type: TokenEOF, Value: "", Pos: 4},
			},
		},
		{
			name:  "500TiB",
			input: "500TiB",
			expected: []Token{
				{Type: TokenNumber, Value: "500", Pos: 0},
				{Type: TokenSizeUnit, Value: "TiB", Pos: 3},
				{Type: TokenEOF, Value: "", Pos: 6},
			},
		},
		{
			name:  "1GB",
			input: "1GB",
			expected: []Token{
				{Type: TokenNumber, Value: "1", Pos: 0},
				{Type: TokenSizeUnit, Value: "GB", Pos: 1},
				{Type: TokenEOF, Value: "", Pos: 3},
			},
		},
		{
			name:  "10KB",
			input: "10KB",
			expected: []Token{
				{Type: TokenNumber, Value: "10", Pos: 0},
				{Type: TokenSizeUnit, Value: "KB", Pos: 2},
				{Type: TokenEOF, Value: "", Pos: 4},
			},
		},
		{
			name:  "1TB",
			input: "1TB",
			expected: []Token{
				{Type: TokenNumber, Value: "1", Pos: 0},
				{Type: TokenSizeUnit, Value: "TB", Pos: 1},
				{Type: TokenEOF, Value: "", Pos: 3},
			},
		},
		{
			name:  "5MiB",
			input: "5MiB",
			expected: []Token{
				{Type: TokenNumber, Value: "5", Pos: 0},
				{Type: TokenSizeUnit, Value: "MiB", Pos: 1},
				{Type: TokenEOF, Value: "", Pos: 4},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tokens) != len(tt.expected) {
				t.Fatalf("expected %d tokens, got %d", len(tt.expected), len(tokens))
			}
			for i, tok := range tokens {
				if tok.Type != tt.expected[i].Type {
					t.Errorf("token %d: expected type %v, got %v", i, tt.expected[i].Type, tok.Type)
				}
				if tok.Value != tt.expected[i].Value {
					t.Errorf("token %d: expected value %q, got %q", i, tt.expected[i].Value, tok.Value)
				}
			}
		})
	}
}

func TestLexer_Durations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "7d (days)",
			input: "7d",
			expected: []Token{
				{Type: TokenDuration, Value: "7d", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 2},
			},
		},
		{
			name:  "24h (hours)",
			input: "24h",
			expected: []Token{
				{Type: TokenDuration, Value: "24h", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 3},
			},
		},
		{
			name:  "30m (minutes)",
			input: "30m",
			expected: []Token{
				{Type: TokenDuration, Value: "30m", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tokens) != len(tt.expected) {
				t.Fatalf("expected %d tokens, got %d", len(tt.expected), len(tokens))
			}
			for i, tok := range tokens {
				if tok.Type != tt.expected[i].Type {
					t.Errorf("token %d: expected type %v, got %v", i, tt.expected[i].Type, tok.Type)
				}
				if tok.Value != tt.expected[i].Value {
					t.Errorf("token %d: expected value %q, got %q", i, tt.expected[i].Value, tok.Value)
				}
			}
		})
	}
}

func TestLexer_StringLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "simple string",
			input: "'test'",
			expected: []Token{
				{Type: TokenString, Value: "test", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 6},
			},
		},
		{
			name:  "pattern with percent",
			input: "'%.txt'",
			expected: []Token{
				{Type: TokenString, Value: "%.txt", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 7},
			},
		},
		{
			name:  "escaped single quote",
			input: "'file''s name.txt'",
			expected: []Token{
				{Type: TokenString, Value: "file's name.txt", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 18},
			},
		},
		{
			name:  "multiple escaped quotes",
			input: "'it''s a test''s file'",
			expected: []Token{
				{Type: TokenString, Value: "it's a test's file", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 22},
			},
		},
		{
			name:  "empty string",
			input: "''",
			expected: []Token{
				{Type: TokenString, Value: "", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tokens) != len(tt.expected) {
				t.Fatalf("expected %d tokens, got %d", len(tt.expected), len(tokens))
			}
			for i, tok := range tokens {
				if tok.Type != tt.expected[i].Type {
					t.Errorf("token %d: expected type %v, got %v", i, tt.expected[i].Type, tok.Type)
				}
				if tok.Value != tt.expected[i].Value {
					t.Errorf("token %d: expected value %q, got %q", i, tt.expected[i].Value, tok.Value)
				}
			}
		})
	}
}

func TestLexer_ComplexQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "size > 1GiB",
			input: "size > 1GiB",
			expected: []Token{
				{Type: TokenIdent, Value: "size", Pos: 0},
				{Type: TokenGT, Value: ">", Pos: 5},
				{Type: TokenNumber, Value: "1", Pos: 7},
				{Type: TokenSizeUnit, Value: "GiB", Pos: 8},
				{Type: TokenEOF, Value: "", Pos: 11},
			},
		},
		{
			name:  "size > 1GiB AND name LIKE '%.txt'",
			input: "size > 1GiB AND name LIKE '%.txt'",
			expected: []Token{
				{Type: TokenIdent, Value: "size", Pos: 0},
				{Type: TokenGT, Value: ">", Pos: 5},
				{Type: TokenNumber, Value: "1", Pos: 7},
				{Type: TokenSizeUnit, Value: "GiB", Pos: 8},
				{Type: TokenAND, Value: "AND", Pos: 12},
				{Type: TokenIdent, Value: "name", Pos: 16},
				{Type: TokenLIKE, Value: "LIKE", Pos: 21},
				{Type: TokenString, Value: "%.txt", Pos: 26},
				{Type: TokenEOF, Value: "", Pos: 33},
			},
		},
		{
			name:  "ext IN ('jpg', 'png', 'gif')",
			input: "ext IN ('jpg', 'png', 'gif')",
			expected: []Token{
				{Type: TokenIdent, Value: "ext", Pos: 0},
				{Type: TokenIN, Value: "IN", Pos: 4},
				{Type: TokenLParen, Value: "(", Pos: 7},
				{Type: TokenString, Value: "jpg", Pos: 8},
				{Type: TokenComma, Value: ",", Pos: 13},
				{Type: TokenString, Value: "png", Pos: 15},
				{Type: TokenComma, Value: ",", Pos: 20},
				{Type: TokenString, Value: "gif", Pos: 22},
				{Type: TokenRParen, Value: ")", Pos: 27},
				{Type: TokenEOF, Value: "", Pos: 28},
			},
		},
		{
			name:  "now() - 7d",
			input: "now() - 7d",
			expected: []Token{
				{Type: TokenIdent, Value: "now", Pos: 0},
				{Type: TokenLParen, Value: "(", Pos: 3},
				{Type: TokenRParen, Value: ")", Pos: 4},
				{Type: TokenMinus, Value: "-", Pos: 6},
				{Type: TokenDuration, Value: "7d", Pos: 8},
				{Type: TokenEOF, Value: "", Pos: 10},
			},
		},
		{
			name:  "(size > 1GiB OR name LIKE '%.mp4') AND NOT isdir",
			input: "(size > 1GiB OR name LIKE '%.mp4') AND NOT isdir",
			expected: []Token{
				{Type: TokenLParen, Value: "(", Pos: 0},
				{Type: TokenIdent, Value: "size", Pos: 1},
				{Type: TokenGT, Value: ">", Pos: 6},
				{Type: TokenNumber, Value: "1", Pos: 8},
				{Type: TokenSizeUnit, Value: "GiB", Pos: 9},
				{Type: TokenOR, Value: "OR", Pos: 13},
				{Type: TokenIdent, Value: "name", Pos: 16},
				{Type: TokenLIKE, Value: "LIKE", Pos: 21},
				{Type: TokenString, Value: "%.mp4", Pos: 26},
				{Type: TokenRParen, Value: ")", Pos: 33},
				{Type: TokenAND, Value: "AND", Pos: 35},
				{Type: TokenNOT, Value: "NOT", Pos: 39},
				{Type: TokenIdent, Value: "isdir", Pos: 43},
				{Type: TokenEOF, Value: "", Pos: 48},
			},
		},
		{
			name:  "ext IS NULL",
			input: "ext IS NULL",
			expected: []Token{
				{Type: TokenIdent, Value: "ext", Pos: 0},
				{Type: TokenIS, Value: "IS", Pos: 4},
				{Type: TokenNULL, Value: "NULL", Pos: 7},
				{Type: TokenEOF, Value: "", Pos: 11},
			},
		},
		{
			name:  "WHERE size > 0",
			input: "WHERE size > 0",
			expected: []Token{
				{Type: TokenWHERE, Value: "WHERE", Pos: 0},
				{Type: TokenIdent, Value: "size", Pos: 6},
				{Type: TokenGT, Value: ">", Pos: 11},
				{Type: TokenNumber, Value: "0", Pos: 13},
				{Type: TokenEOF, Value: "", Pos: 14},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tokens) != len(tt.expected) {
				t.Fatalf("expected %d tokens, got %d\nGot: %v", len(tt.expected), len(tokens), tokens)
			}
			for i, tok := range tokens {
				if tok.Type != tt.expected[i].Type {
					t.Errorf("token %d: expected type %v, got %v", i, tt.expected[i].Type, tok.Type)
				}
				if tok.Value != tt.expected[i].Value {
					t.Errorf("token %d: expected value %q, got %q", i, tt.expected[i].Value, tok.Value)
				}
			}
		})
	}
}

func TestLexer_Errors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "unterminated string",
			input:       "'unterminated",
			expectedErr: "unterminated string literal",
		},
		{
			name:        "unescaped single quote in string",
			input:       "'file's name'",
			expectedErr: "unterminated string literal",
		},
		{
			name:        "invalid character @",
			input:       "@invalid",
			expectedErr: "unexpected character",
		},
		{
			name:        "invalid character #",
			input:       "size # 100",
			expectedErr: "unexpected character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			_, err := lexer.Tokenize()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !containsString(err.Error(), tt.expectedErr) {
				t.Errorf("expected error containing %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestLexer_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"leading whitespace", "   size"},
		{"trailing whitespace", "size   "},
		{"multiple spaces", "size    >    100"},
		{"tabs", "size\t>\t100"},
		{"mixed whitespace", " \t size \t > \t 100 \t "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Should have meaningful tokens (not just EOF)
			if len(tokens) < 2 {
				t.Fatalf("expected at least 2 tokens, got %d", len(tokens))
			}
		})
	}
}

func TestLexer_EmptyInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tabs only", "\t\t\t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Should only have EOF token
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token (EOF), got %d", len(tokens))
			}
			if tokens[0].Type != TokenEOF {
				t.Errorf("expected EOF token, got %v", tokens[0].Type)
			}
		})
	}
}

func TestLexer_DecimalNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "decimal number",
			input: "1.5",
			expected: []Token{
				{Type: TokenNumber, Value: "1.5", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 3},
			},
		},
		{
			name:  "decimal with size unit",
			input: "1.5GiB",
			expected: []Token{
				{Type: TokenNumber, Value: "1.5", Pos: 0},
				{Type: TokenSizeUnit, Value: "GiB", Pos: 3},
				{Type: TokenEOF, Value: "", Pos: 6},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tokens) != len(tt.expected) {
				t.Fatalf("expected %d tokens, got %d", len(tt.expected), len(tokens))
			}
			for i, tok := range tokens {
				if tok.Type != tt.expected[i].Type {
					t.Errorf("token %d: expected type %v, got %v", i, tt.expected[i].Type, tok.Type)
				}
				if tok.Value != tt.expected[i].Value {
					t.Errorf("token %d: expected value %q, got %q", i, tt.expected[i].Value, tok.Value)
				}
			}
		})
	}
}

// containsString checks if s contains substr
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
