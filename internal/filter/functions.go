package filter

import (
	"fmt"
	"strings"
	"time"
)

// builtinFunction represents a built-in function implementation.
type builtinFunction func(args []interface{}) (interface{}, error)

// builtinFunctions maps function names to their implementations.
var builtinFunctions = map[string]builtinFunction{
	"now":   fnNow,
	"year":  fnYear,
	"month": fnMonth,
	"day":   fnDay,
	"lower": fnLower,
	"upper": fnUpper,
}

// callFunction calls a built-in function with the given arguments.
func callFunction(name string, args []interface{}) (interface{}, error) {
	fn, ok := builtinFunctions[name]
	if !ok {
		return nil, fmt.Errorf("unknown function: %s", name)
	}
	return fn(args)
}

// fnNow returns the current timestamp in local time.
func fnNow(args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("now() takes no arguments, got %d", len(args))
	}
	return time.Now(), nil
}

// fnYear extracts the year from a timestamp.
func fnYear(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("year() takes 1 argument, got %d", len(args))
	}

	t, err := coerceToTime(args[0])
	if err != nil {
		return nil, fmt.Errorf("year(): %w", err)
	}
	return int64(t.Year()), nil
}

// fnMonth extracts the month (1-12) from a timestamp.
func fnMonth(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("month() takes 1 argument, got %d", len(args))
	}

	t, err := coerceToTime(args[0])
	if err != nil {
		return nil, fmt.Errorf("month(): %w", err)
	}
	return int64(t.Month()), nil
}

// fnDay extracts the day of month (1-31) from a timestamp.
func fnDay(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("day() takes 1 argument, got %d", len(args))
	}

	t, err := coerceToTime(args[0])
	if err != nil {
		return nil, fmt.Errorf("day(): %w", err)
	}
	return int64(t.Day()), nil
}

// fnLower converts a string to lowercase.
func fnLower(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("lower() takes 1 argument, got %d", len(args))
	}

	if IsNull(args[0]) {
		return NullValue{}, nil
	}

	s, err := coerceToString(args[0])
	if err != nil {
		return nil, fmt.Errorf("lower(): %w", err)
	}
	return strings.ToLower(s), nil
}

// fnUpper converts a string to uppercase.
func fnUpper(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("upper() takes 1 argument, got %d", len(args))
	}

	if IsNull(args[0]) {
		return NullValue{}, nil
	}

	s, err := coerceToString(args[0])
	if err != nil {
		return nil, fmt.Errorf("upper(): %w", err)
	}
	return strings.ToUpper(s), nil
}
