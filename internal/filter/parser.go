package filter

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser converts a token stream into an AST.
type Parser struct {
	tokens []Token
	pos    int
}

// Parse parses a SQL-like WHERE clause and returns an AST.
func Parse(input string) (Expression, error) {
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}

	parser := &Parser{
		tokens: tokens,
		pos:    0,
	}

	return parser.parse()
}

// parse is the entry point for parsing.
func (p *Parser) parse() (Expression, error) {
	// Skip optional WHERE keyword
	if p.check(TokenWHERE) {
		p.advance()
	}

	// Empty query
	if p.check(TokenEOF) {
		return nil, nil
	}

	expr, err := p.parseOrExpr()
	if err != nil {
		return nil, err
	}

	// Ensure we consumed all tokens
	if !p.check(TokenEOF) {
		return nil, p.errorAt(p.peek().Pos, "unexpected token", "")
	}

	return expr, nil
}

// parseOrExpr parses OR expressions (lowest precedence).
func (p *Parser) parseOrExpr() (Expression, error) {
	left, err := p.parseAndExpr()
	if err != nil {
		return nil, err
	}

	for p.check(TokenOR) {
		p.advance()
		right, err := p.parseAndExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Operator: "OR", Right: right}
	}

	return left, nil
}

// parseAndExpr parses AND expressions.
func (p *Parser) parseAndExpr() (Expression, error) {
	left, err := p.parseNotExpr()
	if err != nil {
		return nil, err
	}

	for p.check(TokenAND) {
		p.advance()
		right, err := p.parseNotExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Operator: "AND", Right: right}
	}

	return left, nil
}

// parseNotExpr parses NOT expressions (highest precedence among logical ops).
func (p *Parser) parseNotExpr() (Expression, error) {
	if p.check(TokenNOT) {
		p.advance()
		operand, err := p.parseNotExpr()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Operator: "NOT", Operand: operand}, nil
	}

	return p.parsePrimary()
}

// parsePrimary parses primary expressions (comparisons, parentheses, etc.).
func (p *Parser) parsePrimary() (Expression, error) {
	// Parenthesized expression
	if p.check(TokenLParen) {
		p.advance()
		expr, err := p.parseOrExpr()
		if err != nil {
			return nil, err
		}
		if !p.check(TokenRParen) {
			return nil, p.errorAt(p.peek().Pos, "expected ')'", "check for matching parentheses")
		}
		p.advance()
		return expr, nil
	}

	return p.parseComparison()
}

// parseComparison parses comparison expressions and special forms.
func (p *Parser) parseComparison() (Expression, error) {
	left, err := p.parseOperand()
	if err != nil {
		return nil, err
	}

	// Check for special forms based on keywords following the operand
	tok := p.peek()

	// LIKE / ILIKE
	if tok.Type == TokenLIKE || tok.Type == TokenILIKE {
		return p.parseLikeExpr(left, tok.Type == TokenILIKE, false)
	}

	// NOT LIKE
	if tok.Type == TokenNOT {
		nextTok := p.peekNext()
		if nextTok.Type == TokenLIKE {
			p.advance() // consume NOT
			return p.parseLikeExpr(left, false, true)
		}
		if nextTok.Type == TokenIN {
			p.advance() // consume NOT
			return p.parseInExpr(left, true)
		}
	}

	// IN
	if tok.Type == TokenIN {
		return p.parseInExpr(left, false)
	}

	// IS NULL / IS NOT NULL
	if tok.Type == TokenIS {
		return p.parseIsNullExpr(left)
	}

	// Comparison operators
	if isComparisonOp(tok.Type) {
		op := tok.Value
		p.advance()
		right, err := p.parseOperand()
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Left: left, Operator: op, Right: right}, nil
	}

	// Boolean column (isdir, isfile, issymlink) or function returning boolean
	// Just return the left operand as is
	return left, nil
}

// parseLikeExpr parses LIKE/ILIKE expressions.
func (p *Parser) parseLikeExpr(column Expression, caseInsensitive, negated bool) (Expression, error) {
	p.advance() // consume LIKE/ILIKE

	pattern, err := p.parseOperand()
	if err != nil {
		return nil, err
	}

	return &LikeExpr{
		Column:          column,
		Pattern:         pattern,
		CaseInsensitive: caseInsensitive,
		Negated:         negated,
	}, nil
}

// parseInExpr parses IN/NOT IN expressions.
func (p *Parser) parseInExpr(column Expression, negated bool) (Expression, error) {
	p.advance() // consume IN

	if !p.check(TokenLParen) {
		return nil, p.errorAt(p.peek().Pos, "expected '(' after IN", "IN requires a list of values in parentheses")
	}
	p.advance() // consume (

	values := make([]Expression, 0)
	for {
		if p.check(TokenRParen) {
			break
		}

		val, err := p.parseOperand()
		if err != nil {
			return nil, err
		}
		values = append(values, val)

		if p.check(TokenComma) {
			p.advance()
		} else {
			break
		}
	}

	if len(values) == 0 {
		return nil, p.errorAt(p.peek().Pos, "expected at least one value in IN list", "")
	}

	if !p.check(TokenRParen) {
		return nil, p.errorAt(p.peek().Pos, "expected ')' to close IN list", "")
	}
	p.advance() // consume )

	return &InExpr{
		Column:  column,
		Values:  values,
		Negated: negated,
	}, nil
}

// parseIsNullExpr parses IS NULL/IS NOT NULL expressions.
func (p *Parser) parseIsNullExpr(column Expression) (Expression, error) {
	p.advance() // consume IS

	negated := false
	if p.check(TokenNOT) {
		negated = true
		p.advance()
	}

	if !p.check(TokenNULL) {
		return nil, p.errorAt(p.peek().Pos, "expected NULL after IS", "use IS NULL or IS NOT NULL")
	}
	p.advance()

	return &IsNullExpr{
		Column:  column,
		Negated: negated,
	}, nil
}

// parseOperand parses an operand (column, literal, function, or expression with minus).
func (p *Parser) parseOperand() (Expression, error) {
	tok := p.peek()

	switch tok.Type {
	case TokenString:
		p.advance()
		return &Literal{Value: tok.Value}, nil

	case TokenNumber:
		return p.parseNumberOrSize()

	case TokenDuration:
		return p.parseDuration()

	case TokenIdent:
		// Could be a column reference or function call
		return p.parseIdentOrFunction()

	case TokenLParen:
		// Parenthesized expression
		p.advance()
		expr, err := p.parseOrExpr()
		if err != nil {
			return nil, err
		}
		if !p.check(TokenRParen) {
			return nil, p.errorAt(p.peek().Pos, "expected ')'", "")
		}
		p.advance()
		return p.maybeParseMinusAfter(expr)

	case TokenEOF:
		return nil, p.errorAt(tok.Pos, "unexpected end of input", "expected a value or column name")

	default:
		return nil, p.errorAt(tok.Pos, fmt.Sprintf("unexpected token: %s", tok.Type), "")
	}
}

// parseNumberOrSize parses a number, optionally followed by a size unit.
func (p *Parser) parseNumberOrSize() (Expression, error) {
	numTok := p.peek()
	p.advance()

	// Check for size unit
	if p.check(TokenSizeUnit) {
		unitTok := p.peek()
		p.advance()

		num, err := strconv.ParseFloat(numTok.Value, 64)
		if err != nil {
			return nil, p.errorAt(numTok.Pos, "invalid number", "")
		}

		bytes := convertToBytes(num, unitTok.Value)
		return &SizeLiteral{Value: bytes}, nil
	}

	// Plain number
	if strings.Contains(numTok.Value, ".") {
		num, err := strconv.ParseFloat(numTok.Value, 64)
		if err != nil {
			return nil, p.errorAt(numTok.Pos, "invalid number", "")
		}
		return &Literal{Value: num}, nil
	}

	num, err := strconv.ParseInt(numTok.Value, 10, 64)
	if err != nil {
		return nil, p.errorAt(numTok.Pos, "invalid number", "")
	}
	return &Literal{Value: num}, nil
}

// convertToBytes converts a number with a size unit to bytes.
func convertToBytes(num float64, unit string) int64 {
	var multiplier int64
	switch unit {
	case "KiB":
		multiplier = 1024
	case "MiB":
		multiplier = 1024 * 1024
	case "GiB":
		multiplier = 1024 * 1024 * 1024
	case "TiB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "KB":
		multiplier = 1000
	case "MB":
		multiplier = 1000 * 1000
	case "GB":
		multiplier = 1000 * 1000 * 1000
	case "TB":
		multiplier = 1000 * 1000 * 1000 * 1000
	default:
		multiplier = 1
	}
	return int64(num * float64(multiplier))
}

// parseDuration parses a duration literal (7d, 24h, 30m).
func (p *Parser) parseDuration() (Expression, error) {
	tok := p.peek()
	p.advance()

	// Parse "Nd", "Nh", "Nm"
	value := tok.Value
	unit := value[len(value)-1:]
	numStr := value[:len(value)-1]

	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return nil, p.errorAt(tok.Pos, "invalid duration", "")
	}

	return &DurationLiteral{Value: num, Unit: unit}, nil
}

// parseIdentOrFunction parses an identifier (column) or function call.
func (p *Parser) parseIdentOrFunction() (Expression, error) {
	tok := p.peek()
	p.advance()

	name := strings.ToLower(tok.Value)

	// Check if it's a function call
	if p.check(TokenLParen) {
		return p.parseFunctionCall(name, tok.Pos)
	}

	// It's a column reference - validate it
	if !isValidColumn(name) {
		return nil, p.errorAt(tok.Pos,
			fmt.Sprintf("unknown column: %s", name),
			fmt.Sprintf("Valid columns: %s", formatValidColumns()))
	}

	expr := &ColumnRef{Name: name}
	return p.maybeParseMinusAfter(expr)
}

// parseFunctionCall parses a function call.
func (p *Parser) parseFunctionCall(name string, pos int) (Expression, error) {
	// Validate function name
	if !isValidFunction(name) {
		return nil, p.errorAt(pos,
			fmt.Sprintf("unknown function: %s", name),
			fmt.Sprintf("Valid functions: %s", formatValidFunctions()))
	}

	p.advance() // consume (

	args := make([]Expression, 0)
	for !p.check(TokenRParen) {
		if p.check(TokenEOF) {
			return nil, p.errorAt(p.peek().Pos, "expected ')' to close function call", "")
		}

		arg, err := p.parseOperand()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		if p.check(TokenComma) {
			p.advance()
		} else {
			break
		}
	}

	if !p.check(TokenRParen) {
		return nil, p.errorAt(p.peek().Pos, "expected ')' to close function call", "")
	}
	p.advance()

	fn := &FunctionCall{Name: name, Args: args}
	return p.maybeParseMinusAfter(fn)
}

// maybeParseMinusAfter checks for a minus operator and parses the subtraction.
func (p *Parser) maybeParseMinusAfter(left Expression) (Expression, error) {
	if p.check(TokenMinus) {
		p.advance()
		right, err := p.parseOperand()
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Left: left, Operator: "-", Right: right}, nil
	}
	return left, nil
}

// Helper methods

func (p *Parser) check(typ TokenType) bool {
	return p.peek().Type == typ
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF, Value: "", Pos: -1}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekNext() Token {
	if p.pos+1 >= len(p.tokens) {
		return Token{Type: TokenEOF, Value: "", Pos: -1}
	}
	return p.tokens[p.pos+1]
}

func (p *Parser) advance() Token {
	tok := p.peek()
	p.pos++
	return tok
}

func (p *Parser) errorAt(pos int, message, hint string) error {
	return NewQueryError(pos, message, hint)
}

func isComparisonOp(typ TokenType) bool {
	switch typ {
	case TokenEQ, TokenNE, TokenNE2, TokenLT, TokenGT, TokenLE, TokenGE:
		return true
	}
	return false
}

func isValidColumn(name string) bool {
	for _, col := range ValidColumns {
		if col == name {
			return true
		}
	}
	return false
}

func isValidFunction(name string) bool {
	for _, fn := range ValidFunctions {
		if fn == name {
			return true
		}
	}
	return false
}
