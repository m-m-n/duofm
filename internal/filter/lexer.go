package filter

import (
	"fmt"
	"strings"
)

// Lexer tokenizes SQL-like WHERE clause input.
type Lexer struct {
	input  string
	pos    int
	tokens []Token
}

// NewLexer creates a new Lexer for the given input.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		tokens: make([]Token, 0),
	}
}

// Tokenize converts the input string into a slice of tokens.
func (l *Lexer) Tokenize() ([]Token, error) {
	for !l.isAtEnd() {
		if err := l.scanToken(); err != nil {
			return nil, err
		}
	}
	l.tokens = append(l.tokens, Token{Type: TokenEOF, Value: "", Pos: l.pos})
	return l.tokens, nil
}

// isAtEnd returns true if we've reached the end of input.
func (l *Lexer) isAtEnd() bool {
	return l.pos >= len(l.input)
}

// peek returns the current character without advancing.
func (l *Lexer) peek() byte {
	if l.isAtEnd() {
		return 0
	}
	return l.input[l.pos]
}

// peekNext returns the next character without advancing.
func (l *Lexer) peekNext() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

// advance consumes and returns the current character.
func (l *Lexer) advance() byte {
	ch := l.input[l.pos]
	l.pos++
	return ch
}

// scanToken scans a single token from the input.
func (l *Lexer) scanToken() error {
	l.skipWhitespace()
	if l.isAtEnd() {
		return nil
	}

	startPos := l.pos
	ch := l.peek()

	// Single character tokens
	switch ch {
	case '(':
		l.advance()
		l.tokens = append(l.tokens, Token{Type: TokenLParen, Value: "(", Pos: startPos})
		return nil
	case ')':
		l.advance()
		l.tokens = append(l.tokens, Token{Type: TokenRParen, Value: ")", Pos: startPos})
		return nil
	case ',':
		l.advance()
		l.tokens = append(l.tokens, Token{Type: TokenComma, Value: ",", Pos: startPos})
		return nil
	case '-':
		l.advance()
		l.tokens = append(l.tokens, Token{Type: TokenMinus, Value: "-", Pos: startPos})
		return nil
	case '=':
		l.advance()
		l.tokens = append(l.tokens, Token{Type: TokenEQ, Value: "=", Pos: startPos})
		return nil
	case '!':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			l.tokens = append(l.tokens, Token{Type: TokenNE, Value: "!=", Pos: startPos})
		} else {
			return l.errorAt(startPos, "unexpected character '!'")
		}
		return nil
	case '<':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			l.tokens = append(l.tokens, Token{Type: TokenLE, Value: "<=", Pos: startPos})
		} else if l.peek() == '>' {
			l.advance()
			l.tokens = append(l.tokens, Token{Type: TokenNE2, Value: "<>", Pos: startPos})
		} else {
			l.tokens = append(l.tokens, Token{Type: TokenLT, Value: "<", Pos: startPos})
		}
		return nil
	case '>':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			l.tokens = append(l.tokens, Token{Type: TokenGE, Value: ">=", Pos: startPos})
		} else {
			l.tokens = append(l.tokens, Token{Type: TokenGT, Value: ">", Pos: startPos})
		}
		return nil
	case '\'':
		return l.scanString()
	}

	// Numbers and durations/size units
	if isDigit(ch) {
		return l.scanNumber()
	}

	// Identifiers and keywords
	if isAlpha(ch) || ch == '_' {
		return l.scanIdentifier()
	}

	return l.errorAt(startPos, fmt.Sprintf("unexpected character '%c'", ch))
}

// skipWhitespace skips over whitespace characters.
func (l *Lexer) skipWhitespace() {
	for !l.isAtEnd() && isWhitespace(l.peek()) {
		l.advance()
	}
}

// scanString scans a string literal enclosed in single quotes.
func (l *Lexer) scanString() error {
	startPos := l.pos
	l.advance() // consume opening quote

	var sb strings.Builder
	for !l.isAtEnd() {
		ch := l.peek()
		if ch == '\'' {
			l.advance()
			// Check for escaped quote ('')
			if l.peek() == '\'' {
				sb.WriteByte('\'')
				l.advance()
				continue
			}
			// End of string
			l.tokens = append(l.tokens, Token{Type: TokenString, Value: sb.String(), Pos: startPos})
			return nil
		}
		sb.WriteByte(ch)
		l.advance()
	}

	return l.errorAt(startPos, "unterminated string literal")
}

// scanNumber scans a numeric literal, possibly followed by a size unit or duration suffix.
func (l *Lexer) scanNumber() error {
	startPos := l.pos

	// Consume digits
	for !l.isAtEnd() && isDigit(l.peek()) {
		l.advance()
	}

	// Check for decimal point
	if l.peek() == '.' && isDigit(l.peekNext()) {
		l.advance() // consume '.'
		for !l.isAtEnd() && isDigit(l.peek()) {
			l.advance()
		}
	}

	numberValue := l.input[startPos:l.pos]

	// Check for size unit or duration suffix
	if !l.isAtEnd() && isAlpha(l.peek()) {
		suffixStart := l.pos
		// Peek ahead to see what suffix follows
		suffix := l.peekSuffix()

		if isDurationSuffix(suffix) {
			// Duration: Nd, Nh, Nm
			l.advance() // consume single char suffix
			durationValue := numberValue + suffix
			l.tokens = append(l.tokens, Token{Type: TokenDuration, Value: durationValue, Pos: startPos})
			return nil
		}

		if isSizeUnit(suffix) {
			// Size unit: consume it
			for i := 0; i < len(suffix); i++ {
				l.advance()
			}
			l.tokens = append(l.tokens, Token{Type: TokenNumber, Value: numberValue, Pos: startPos})
			l.tokens = append(l.tokens, Token{Type: TokenSizeUnit, Value: suffix, Pos: suffixStart})
			return nil
		}
	}

	l.tokens = append(l.tokens, Token{Type: TokenNumber, Value: numberValue, Pos: startPos})
	return nil
}

// peekSuffix peeks at a potential suffix (identifier characters) without advancing.
func (l *Lexer) peekSuffix() string {
	pos := l.pos
	var sb strings.Builder
	for pos < len(l.input) && isAlpha(l.input[pos]) {
		sb.WriteByte(l.input[pos])
		pos++
	}
	return sb.String()
}

// isDurationSuffix checks if the suffix is a valid duration suffix.
func isDurationSuffix(s string) bool {
	return s == "d" || s == "h" || s == "m"
}

// isSizeUnit checks if the string is a valid size unit.
func isSizeUnit(s string) bool {
	switch s {
	case "KiB", "MiB", "GiB", "TiB", "KB", "MB", "GB", "TB":
		return true
	}
	return false
}

// scanIdentifier scans an identifier or keyword.
func (l *Lexer) scanIdentifier() error {
	startPos := l.pos

	for !l.isAtEnd() && (isAlphaNumeric(l.peek()) || l.peek() == '_') {
		l.advance()
	}

	value := l.input[startPos:l.pos]
	tokenType := l.lookupKeyword(value)
	l.tokens = append(l.tokens, Token{Type: tokenType, Value: value, Pos: startPos})
	return nil
}

// lookupKeyword returns the token type for a keyword or TokenIdent for identifiers.
func (l *Lexer) lookupKeyword(value string) TokenType {
	upper := strings.ToUpper(value)
	switch upper {
	case "AND":
		return TokenAND
	case "OR":
		return TokenOR
	case "NOT":
		return TokenNOT
	case "LIKE":
		return TokenLIKE
	case "ILIKE":
		return TokenILIKE
	case "IN":
		return TokenIN
	case "IS":
		return TokenIS
	case "NULL":
		return TokenNULL
	case "WHERE":
		return TokenWHERE
	default:
		return TokenIdent
	}
}

// errorAt creates an error at the given position.
func (l *Lexer) errorAt(pos int, message string) error {
	return fmt.Errorf("lexer error at position %d: %s", pos, message)
}

// Helper functions

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isAlphaNumeric(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
