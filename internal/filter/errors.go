package filter

import "fmt"

// QueryError represents a parsing or evaluation error with position and hint.
type QueryError struct {
	Message  string
	Position int
	Hint     string
}

// Error implements the error interface.
func (e *QueryError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("Error at position %d: %s\nHint: %s", e.Position, e.Message, e.Hint)
	}
	return fmt.Sprintf("Error at position %d: %s", e.Position, e.Message)
}

// NewQueryError creates a new QueryError.
func NewQueryError(pos int, message string, hint string) *QueryError {
	return &QueryError{
		Message:  message,
		Position: pos,
		Hint:     hint,
	}
}

// ValidColumns is the list of valid column names for error messages.
var ValidColumns = []string{
	"name", "size", "mtime", "type", "ext", "perm",
	"owner", "group", "isdir", "isfile", "issymlink",
}

// ValidFunctions is the list of valid function names for error messages.
var ValidFunctions = []string{
	"now", "year", "month", "day", "lower", "upper",
}

// formatValidColumns returns a formatted string of valid columns.
func formatValidColumns() string {
	return fmt.Sprintf("%v", ValidColumns)
}

// formatValidFunctions returns a formatted string of valid functions.
func formatValidFunctions() string {
	return fmt.Sprintf("%v", ValidFunctions)
}
