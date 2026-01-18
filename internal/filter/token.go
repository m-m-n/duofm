// Package filter provides SQL-like WHERE clause filtering for file entries.
package filter

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// TokenEOF indicates end of input
	TokenEOF TokenType = iota
	// TokenIdent represents identifiers (column names)
	TokenIdent
	// TokenString represents string literals ('...')
	TokenString
	// TokenNumber represents numeric literals
	TokenNumber
	// TokenSizeUnit represents size unit suffixes (KiB, MiB, GB, etc.)
	TokenSizeUnit
	// TokenDuration represents duration suffixes (7d, 1h, 30m)
	TokenDuration
	// TokenLParen represents left parenthesis
	TokenLParen
	// TokenRParen represents right parenthesis
	TokenRParen
	// TokenComma represents comma
	TokenComma
	// TokenMinus represents minus sign
	TokenMinus
	// TokenEQ represents equals operator
	TokenEQ
	// TokenNE represents not equals operator (!=)
	TokenNE
	// TokenNE2 represents not equals operator (<>)
	TokenNE2
	// TokenLT represents less than operator
	TokenLT
	// TokenGT represents greater than operator
	TokenGT
	// TokenLE represents less than or equal operator
	TokenLE
	// TokenGE represents greater than or equal operator
	TokenGE
	// TokenAND represents AND keyword
	TokenAND
	// TokenOR represents OR keyword
	TokenOR
	// TokenNOT represents NOT keyword
	TokenNOT
	// TokenLIKE represents LIKE keyword
	TokenLIKE
	// TokenILIKE represents ILIKE keyword
	TokenILIKE
	// TokenIN represents IN keyword
	TokenIN
	// TokenIS represents IS keyword
	TokenIS
	// TokenNULL represents NULL keyword
	TokenNULL
	// TokenWHERE represents WHERE keyword
	TokenWHERE
)

// String returns the string representation of TokenType.
func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenIdent:
		return "IDENT"
	case TokenString:
		return "STRING"
	case TokenNumber:
		return "NUMBER"
	case TokenSizeUnit:
		return "SIZE_UNIT"
	case TokenDuration:
		return "DURATION"
	case TokenLParen:
		return "LPAREN"
	case TokenRParen:
		return "RPAREN"
	case TokenComma:
		return "COMMA"
	case TokenMinus:
		return "MINUS"
	case TokenEQ:
		return "EQ"
	case TokenNE:
		return "NE"
	case TokenNE2:
		return "NE2"
	case TokenLT:
		return "LT"
	case TokenGT:
		return "GT"
	case TokenLE:
		return "LE"
	case TokenGE:
		return "GE"
	case TokenAND:
		return "AND"
	case TokenOR:
		return "OR"
	case TokenNOT:
		return "NOT"
	case TokenLIKE:
		return "LIKE"
	case TokenILIKE:
		return "ILIKE"
	case TokenIN:
		return "IN"
	case TokenIS:
		return "IS"
	case TokenNULL:
		return "NULL"
	case TokenWHERE:
		return "WHERE"
	default:
		return "UNKNOWN"
	}
}

// Token represents a lexical token with its type, value, and position.
type Token struct {
	Type  TokenType
	Value string
	Pos   int // Position in input (0-based)
}
