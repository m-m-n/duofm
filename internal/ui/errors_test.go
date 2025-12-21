package ui

import (
	"errors"
	"os"
	"syscall"
	"testing"
)

func TestFormatDirectoryError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		path     string
		expected string
	}{
		{
			name:     "nil error returns empty string",
			err:      nil,
			path:     "/test/path",
			expected: "",
		},
		{
			name:     "EACCES permission denied",
			err:      &os.PathError{Op: "open", Path: "/test", Err: syscall.EACCES},
			path:     "/test/path",
			expected: "Permission denied: /test/path",
		},
		{
			name:     "ENOENT no such file or directory",
			err:      &os.PathError{Op: "open", Path: "/test", Err: syscall.ENOENT},
			path:     "/test/path",
			expected: "No such directory: /test/path",
		},
		{
			name:     "EIO I/O error",
			err:      &os.PathError{Op: "read", Path: "/test", Err: syscall.EIO},
			path:     "/test/path",
			expected: "I/O error: /test/path",
		},
		{
			name:     "os.IsNotExist error",
			err:      os.ErrNotExist,
			path:     "/nonexistent/path",
			expected: "No such directory: /nonexistent/path",
		},
		{
			name:     "os.IsPermission error",
			err:      os.ErrPermission,
			path:     "/restricted/path",
			expected: "Permission denied: /restricted/path",
		},
		{
			name:     "generic error",
			err:      errors.New("some unknown error"),
			path:     "/some/path",
			expected: "Cannot access: /some/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDirectoryError(tt.err, tt.path)
			if result != tt.expected {
				t.Errorf("formatDirectoryError() = %q, want %q", result, tt.expected)
			}
		})
	}
}
