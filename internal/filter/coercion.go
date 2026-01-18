package filter

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// coerceToInt64 converts a value to int64.
func coerceToInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case string:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert %q to int64", val)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

// coerceToFloat64 converts a value to float64.
func coerceToFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int64:
		return float64(val), nil
	case int:
		return float64(val), nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert %q to float64", val)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// coerceToString converts a value to string.
func coerceToString(v interface{}) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case int64:
		return strconv.FormatInt(val, 10), nil
	case int:
		return strconv.Itoa(val), nil
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(val), nil
	default:
		return "", fmt.Errorf("cannot convert %T to string", v)
	}
}

// coerceToBool converts a value to bool.
func coerceToBool(v interface{}) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case int64:
		return val != 0, nil
	case int:
		return val != 0, nil
	case string:
		lower := strings.ToLower(val)
		switch lower {
		case "true", "1", "yes":
			return true, nil
		case "false", "0", "no":
			return false, nil
		default:
			return false, fmt.Errorf("cannot convert %q to bool", val)
		}
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

// coerceToTime converts a value to time.Time.
// Supports ISO 8601 format: YYYY-MM-DD or YYYY-MM-DD HH:MM:SS
// All dates are interpreted in local timezone.
func coerceToTime(v interface{}) (time.Time, error) {
	switch val := v.(type) {
	case time.Time:
		return val, nil
	case string:
		return parseDateTime(val)
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to time", v)
	}
}

// parseDateTime parses a date/time string in ISO 8601 format.
func parseDateTime(s string) (time.Time, error) {
	// Try with time
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err == nil {
		return t, nil
	}

	// Try date only
	t, err = time.ParseInLocation("2006-01-02", s, time.Local)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid date format: %q (expected YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)", s)
}

// NullValue represents a SQL NULL value.
type NullValue struct{}

// IsNull checks if a value is NULL.
func IsNull(v interface{}) bool {
	_, ok := v.(NullValue)
	return ok
}
