package filter

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sakura/duofm/internal/fs"
)

// Evaluate parses a query and evaluates it against a FileEntry.
func Evaluate(query string, entry fs.FileEntry) (bool, error) {
	expr, err := Parse(query)
	if err != nil {
		return false, err
	}

	// Empty query matches all entries
	if expr == nil {
		return true, nil
	}

	eval := &Evaluator{entry: entry}
	return eval.eval(expr)
}

// Evaluator evaluates AST expressions against a FileEntry.
type Evaluator struct {
	entry fs.FileEntry
}

// NewEvaluator creates a new Evaluator for a FileEntry.
func NewEvaluator(entry fs.FileEntry) *Evaluator {
	return &Evaluator{entry: entry}
}

// Eval evaluates an expression and returns a boolean result.
func (e *Evaluator) Eval(expr Expression) (bool, error) {
	return e.eval(expr)
}

// eval is the internal evaluation method.
func (e *Evaluator) eval(expr Expression) (bool, error) {
	switch ex := expr.(type) {
	case *BinaryExpr:
		return e.evalBinary(ex)
	case *UnaryExpr:
		return e.evalUnary(ex)
	case *ColumnRef:
		return e.evalColumnAsBool(ex)
	case *LikeExpr:
		return e.evalLike(ex)
	case *InExpr:
		return e.evalIn(ex)
	case *IsNullExpr:
		return e.evalIsNull(ex)
	default:
		return false, fmt.Errorf("unexpected expression type: %T", expr)
	}
}

// evalBinary evaluates a binary expression.
func (e *Evaluator) evalBinary(expr *BinaryExpr) (bool, error) {
	switch expr.Operator {
	case "AND":
		left, err := e.eval(expr.Left)
		if err != nil {
			return false, err
		}
		if !left {
			return false, nil // Short-circuit
		}
		return e.eval(expr.Right)

	case "OR":
		left, err := e.eval(expr.Left)
		if err != nil {
			return false, err
		}
		if left {
			return true, nil // Short-circuit
		}
		return e.eval(expr.Right)

	case "-":
		// Subtraction for time calculations (now() - 7d)
		// This should not be called directly as boolean
		return false, fmt.Errorf("subtraction cannot be evaluated as boolean")

	default:
		// Comparison operators
		return e.evalComparison(expr)
	}
}

// evalComparison evaluates a comparison expression.
func (e *Evaluator) evalComparison(expr *BinaryExpr) (bool, error) {
	leftVal, err := e.evalValue(expr.Left)
	if err != nil {
		return false, err
	}

	rightVal, err := e.evalValue(expr.Right)
	if err != nil {
		return false, err
	}

	// NULL comparisons return false
	if IsNull(leftVal) || IsNull(rightVal) {
		return false, nil
	}

	return compare(leftVal, expr.Operator, rightVal)
}

// evalValue evaluates an expression and returns its value.
func (e *Evaluator) evalValue(expr Expression) (interface{}, error) {
	switch ex := expr.(type) {
	case *Literal:
		return ex.Value, nil
	case *SizeLiteral:
		return ex.Value, nil
	case *DurationLiteral:
		return e.evalDuration(ex)
	case *ColumnRef:
		return e.getColumnValue(ex.Name)
	case *FunctionCall:
		return e.evalFunction(ex)
	case *BinaryExpr:
		if ex.Operator == "-" {
			return e.evalSubtraction(ex)
		}
		// For other binary ops, evaluate as boolean and return
		result, err := e.evalBinary(ex)
		return result, err
	default:
		return nil, fmt.Errorf("cannot evaluate %T as value", expr)
	}
}

// evalDuration converts a DurationLiteral to a time.Duration.
func (e *Evaluator) evalDuration(dur *DurationLiteral) (time.Duration, error) {
	switch dur.Unit {
	case "d":
		return time.Duration(dur.Value) * 24 * time.Hour, nil
	case "h":
		return time.Duration(dur.Value) * time.Hour, nil
	case "m":
		return time.Duration(dur.Value) * time.Minute, nil
	default:
		return 0, fmt.Errorf("unknown duration unit: %s", dur.Unit)
	}
}

// evalSubtraction evaluates a subtraction expression (e.g., now() - 7d).
func (e *Evaluator) evalSubtraction(expr *BinaryExpr) (interface{}, error) {
	leftVal, err := e.evalValue(expr.Left)
	if err != nil {
		return nil, err
	}

	rightVal, err := e.evalValue(expr.Right)
	if err != nil {
		return nil, err
	}

	// time.Time - time.Duration
	if t, ok := leftVal.(time.Time); ok {
		if d, ok := rightVal.(time.Duration); ok {
			return t.Add(-d), nil
		}
	}

	return nil, fmt.Errorf("cannot subtract %T from %T", rightVal, leftVal)
}

// evalFunction evaluates a function call.
func (e *Evaluator) evalFunction(fn *FunctionCall) (interface{}, error) {
	// Evaluate arguments
	args := make([]interface{}, len(fn.Args))
	for i, arg := range fn.Args {
		val, err := e.evalValue(arg)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	return callFunction(fn.Name, args)
}

// evalUnary evaluates a unary expression (NOT).
func (e *Evaluator) evalUnary(expr *UnaryExpr) (bool, error) {
	if expr.Operator != "NOT" {
		return false, fmt.Errorf("unknown unary operator: %s", expr.Operator)
	}

	result, err := e.eval(expr.Operand)
	if err != nil {
		return false, err
	}
	return !result, nil
}

// evalColumnAsBool evaluates a column reference as a boolean (for isdir, isfile, etc.).
func (e *Evaluator) evalColumnAsBool(col *ColumnRef) (bool, error) {
	val, err := e.getColumnValue(col.Name)
	if err != nil {
		return false, err
	}

	if IsNull(val) {
		return false, nil
	}

	return coerceToBool(val)
}

// evalLike evaluates a LIKE/ILIKE expression.
func (e *Evaluator) evalLike(expr *LikeExpr) (bool, error) {
	colVal, err := e.evalValue(expr.Column)
	if err != nil {
		return false, err
	}

	// NULL handling
	if IsNull(colVal) {
		return false, nil
	}

	colStr, err := coerceToString(colVal)
	if err != nil {
		return false, err
	}

	// Use cached compiled regex if available, otherwise compile and cache
	if expr.compiledRegex == nil {
		patternVal, err := e.evalValue(expr.Pattern)
		if err != nil {
			return false, err
		}

		if IsNull(patternVal) {
			return false, nil
		}

		patternStr, err := coerceToString(patternVal)
		if err != nil {
			return false, err
		}

		re, err := compileLikePattern(patternStr, expr.CaseInsensitive)
		if err != nil {
			return false, err
		}
		expr.compiledRegex = re
	}

	match := expr.compiledRegex.MatchString(colStr)

	if expr.Negated {
		return !match, nil
	}
	return match, nil
}

// compileLikePattern converts a SQL LIKE pattern to a compiled regex.
func compileLikePattern(pattern string, caseInsensitive bool) (*regexp.Regexp, error) {
	// Convert SQL LIKE pattern to regex
	var sb strings.Builder
	sb.WriteString("^")

	for i := 0; i < len(pattern); i++ {
		ch := pattern[i]
		switch ch {
		case '%':
			sb.WriteString(".*")
		case '_':
			sb.WriteString(".")
		case '.', '\\', '^', '$', '+', '?', '{', '}', '[', ']', '|', '(':
			sb.WriteByte('\\')
			sb.WriteByte(ch)
		default:
			sb.WriteByte(ch)
		}
	}

	sb.WriteString("$")
	regexPattern := sb.String()

	if caseInsensitive {
		regexPattern = "(?i)" + regexPattern
	}

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid LIKE pattern: %w", err)
	}

	return re, nil
}

// evalIn evaluates an IN/NOT IN expression.
func (e *Evaluator) evalIn(expr *InExpr) (bool, error) {
	colVal, err := e.evalValue(expr.Column)
	if err != nil {
		return false, err
	}

	if IsNull(colVal) {
		return false, nil
	}

	for _, valExpr := range expr.Values {
		val, err := e.evalValue(valExpr)
		if err != nil {
			return false, err
		}

		if IsNull(val) {
			continue
		}

		equal, err := compare(colVal, "=", val)
		if err != nil {
			return false, err
		}
		if equal {
			if expr.Negated {
				return false, nil
			}
			return true, nil
		}
	}

	// Not found in list
	if expr.Negated {
		return true, nil
	}
	return false, nil
}

// evalIsNull evaluates an IS NULL/IS NOT NULL expression.
func (e *Evaluator) evalIsNull(expr *IsNullExpr) (bool, error) {
	colVal, err := e.evalValue(expr.Column)
	if err != nil {
		return false, err
	}

	isNull := IsNull(colVal)
	if expr.Negated {
		return !isNull, nil
	}
	return isNull, nil
}

// getColumnValue retrieves the value of a column from the FileEntry.
func (e *Evaluator) getColumnValue(name string) (interface{}, error) {
	switch name {
	case "name":
		return e.entry.Name, nil
	case "size":
		return e.entry.Size, nil
	case "mtime":
		return e.entry.ModTime, nil
	case "type":
		return getFileType(e.entry), nil
	case "ext":
		return getExtension(e.entry.Name), nil
	case "perm":
		return e.entry.Permissions.String(), nil
	case "owner":
		return e.entry.Owner, nil
	case "group":
		return e.entry.Group, nil
	case "isdir":
		return e.entry.IsDir, nil
	case "isfile":
		return !e.entry.IsDir && !e.entry.IsSymlink, nil
	case "issymlink":
		return e.entry.IsSymlink, nil
	default:
		return nil, fmt.Errorf("unknown column: %s", name)
	}
}

// getFileType returns the type string for a FileEntry.
func getFileType(entry fs.FileEntry) string {
	if entry.IsSymlink {
		return "symlink"
	}
	if entry.IsDir {
		return "dir"
	}
	return "file"
}

// getExtension returns the file extension or NullValue if none.
// For files like ".gitignore", returns NullValue (no extension).
// For files like "archive.tar.gz", returns "gz" (last extension only).
func getExtension(name string) interface{} {
	// Find the last dot
	lastDot := strings.LastIndex(name, ".")

	// No dot found, or dot is at the start (hidden file like .gitignore)
	if lastDot <= 0 {
		return NullValue{}
	}

	// Extract extension (without the dot)
	ext := name[lastDot+1:]

	// Empty extension (file ends with dot)
	if ext == "" {
		return NullValue{}
	}

	return ext
}

// compare compares two values with the given operator.
func compare(left interface{}, op string, right interface{}) (bool, error) {
	// Try numeric comparison first
	if leftNum, err := coerceToInt64(left); err == nil {
		if rightNum, err := coerceToInt64(right); err == nil {
			return compareInt64(leftNum, op, rightNum)
		}
	}

	// Try time comparison
	if leftTime, ok := left.(time.Time); ok {
		rightTime, err := coerceToTime(right)
		if err == nil {
			return compareTime(leftTime, op, rightTime)
		}
	}

	// Fall back to string comparison
	leftStr, err := coerceToString(left)
	if err != nil {
		return false, err
	}
	rightStr, err := coerceToString(right)
	if err != nil {
		return false, err
	}

	return compareString(leftStr, op, rightStr)
}

// compareInt64 compares two int64 values.
func compareInt64(left int64, op string, right int64) (bool, error) {
	switch op {
	case "=":
		return left == right, nil
	case "!=", "<>":
		return left != right, nil
	case "<":
		return left < right, nil
	case ">":
		return left > right, nil
	case "<=":
		return left <= right, nil
	case ">=":
		return left >= right, nil
	default:
		return false, fmt.Errorf("unknown comparison operator: %s", op)
	}
}

// compareTime compares two time.Time values.
func compareTime(left time.Time, op string, right time.Time) (bool, error) {
	switch op {
	case "=":
		return left.Equal(right), nil
	case "!=", "<>":
		return !left.Equal(right), nil
	case "<":
		return left.Before(right), nil
	case ">":
		return left.After(right), nil
	case "<=":
		return !left.After(right), nil
	case ">=":
		return !left.Before(right), nil
	default:
		return false, fmt.Errorf("unknown comparison operator: %s", op)
	}
}

// compareString compares two string values.
func compareString(left string, op string, right string) (bool, error) {
	switch op {
	case "=":
		return left == right, nil
	case "!=", "<>":
		return left != right, nil
	case "<":
		return left < right, nil
	case ">":
		return left > right, nil
	case "<=":
		return left <= right, nil
	case ">=":
		return left >= right, nil
	default:
		return false, fmt.Errorf("unknown comparison operator: %s", op)
	}
}
