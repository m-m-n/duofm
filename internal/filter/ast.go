package filter

import "regexp"

// Expression represents an AST node that can be evaluated.
type Expression interface {
	// exprNode is a marker method to distinguish expression types.
	exprNode()
}

// BinaryExpr represents a binary operation (AND, OR, comparisons).
type BinaryExpr struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (e *BinaryExpr) exprNode() {}

// UnaryExpr represents a unary operation (NOT).
type UnaryExpr struct {
	Operator string
	Operand  Expression
}

func (e *UnaryExpr) exprNode() {}

// ColumnRef represents a reference to a file attribute column.
type ColumnRef struct {
	Name string
}

func (e *ColumnRef) exprNode() {}

// Literal represents a constant value (string, number, timestamp, bool).
type Literal struct {
	Value interface{} // string, int64, float64, time.Time, or bool
}

func (e *Literal) exprNode() {}

// SizeLiteral represents a size value with unit (e.g., 1GiB).
type SizeLiteral struct {
	Value int64 // Size in bytes
}

func (e *SizeLiteral) exprNode() {}

// DurationLiteral represents a duration (e.g., 7d, 24h).
type DurationLiteral struct {
	Value int64  // Duration in the specified unit's base (seconds for h/m, days for d)
	Unit  string // "d", "h", or "m"
}

func (e *DurationLiteral) exprNode() {}

// FunctionCall represents a function invocation (now, year, lower, etc.).
type FunctionCall struct {
	Name string
	Args []Expression
}

func (e *FunctionCall) exprNode() {}

// InExpr represents IN/NOT IN list membership.
type InExpr struct {
	Column  Expression
	Values  []Expression
	Negated bool // true for NOT IN
}

func (e *InExpr) exprNode() {}

// IsNullExpr represents IS NULL/IS NOT NULL check.
type IsNullExpr struct {
	Column  Expression
	Negated bool // true for IS NOT NULL
}

func (e *IsNullExpr) exprNode() {}

// LikeExpr represents LIKE/ILIKE/NOT LIKE pattern matching.
type LikeExpr struct {
	Column          Expression
	Pattern         Expression
	CaseInsensitive bool // true for ILIKE
	Negated         bool // true for NOT LIKE
	// compiledRegex caches the compiled regex pattern for performance.
	// It is lazily initialized on first evaluation.
	compiledRegex *regexp.Regexp
}

func (e *LikeExpr) exprNode() {}
