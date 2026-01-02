package archive

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestArchiveError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ArchiveError
		expected string
	}{
		{
			name: "error without cause",
			err: &ArchiveError{
				Code:    ErrArchiveSourceNotFound,
				Message: "source file not found",
			},
			expected: "ERR_ARCHIVE_001: source file not found",
		},
		{
			name: "error with cause",
			err: &ArchiveError{
				Code:    ErrArchiveIOError,
				Message: "failed to read file",
				Cause:   errors.New("permission denied"),
			},
			expected: "ERR_ARCHIVE_011: failed to read file (caused by: permission denied)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestArchiveError_Unwrap(t *testing.T) {
	tests := []struct {
		name     string
		err      *ArchiveError
		hasCause bool
	}{
		{
			name: "error without cause",
			err: &ArchiveError{
				Code:    ErrArchiveSourceNotFound,
				Message: "source file not found",
			},
			hasCause: false,
		},
		{
			name: "error with cause",
			err: &ArchiveError{
				Code:    ErrArchiveIOError,
				Message: "failed to read file",
				Cause:   errors.New("underlying error"),
			},
			hasCause: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cause := tt.err.Unwrap()
			if tt.hasCause && cause == nil {
				t.Error("Unwrap() returned nil, expected cause")
			}
			if !tt.hasCause && cause != nil {
				t.Errorf("Unwrap() = %v, expected nil", cause)
			}
		})
	}
}

func TestIsRetriable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "non-archive error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name: "IO error is retriable",
			err: &ArchiveError{
				Code:    ErrArchiveIOError,
				Message: "I/O error",
			},
			expected: true,
		},
		{
			name: "internal error is not retriable",
			err: &ArchiveError{
				Code:    ErrArchiveInternalError,
				Message: "internal error",
			},
			expected: false,
		},
		{
			name: "source not found is not retriable",
			err: &ArchiveError{
				Code:    ErrArchiveSourceNotFound,
				Message: "not found",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetriable(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetriable() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != DefaultMaxRetries {
		t.Errorf("MaxRetries = %d, want %d", config.MaxRetries, DefaultMaxRetries)
	}
	if config.Delay != DefaultRetryDelay {
		t.Errorf("Delay = %v, want %v", config.Delay, DefaultRetryDelay)
	}
	if config.Backoff != DefaultRetryBackoff {
		t.Errorf("Backoff = %v, want %v", config.Backoff, DefaultRetryBackoff)
	}
}

func TestWithRetry_Success(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries: 3,
		Delay:      10 * time.Millisecond,
		Backoff:    1.5,
	}

	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	err := WithRetry(ctx, config, fn)
	if err != nil {
		t.Errorf("WithRetry() error = %v, want nil", err)
	}
	if callCount != 1 {
		t.Errorf("Function called %d times, want 1", callCount)
	}
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries: 3,
		Delay:      10 * time.Millisecond,
		Backoff:    1.5,
	}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return &ArchiveError{Code: ErrArchiveIOError, Message: "temporary error"}
		}
		return nil
	}

	err := WithRetry(ctx, config, fn)
	if err != nil {
		t.Errorf("WithRetry() error = %v, want nil", err)
	}
	if callCount != 3 {
		t.Errorf("Function called %d times, want 3", callCount)
	}
}

func TestWithRetry_NonRetriableError(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries: 3,
		Delay:      10 * time.Millisecond,
		Backoff:    1.5,
	}

	callCount := 0
	fn := func() error {
		callCount++
		return &ArchiveError{Code: ErrArchiveSourceNotFound, Message: "file not found"}
	}

	err := WithRetry(ctx, config, fn)
	if err == nil {
		t.Error("WithRetry() expected error, got nil")
	}
	if callCount != 1 {
		t.Errorf("Function called %d times, want 1 (no retries for non-retriable)", callCount)
	}
}

func TestWithRetry_AllRetriesFailed(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries: 2,
		Delay:      10 * time.Millisecond,
		Backoff:    1.5,
	}

	callCount := 0
	fn := func() error {
		callCount++
		return &ArchiveError{Code: ErrArchiveIOError, Message: "persistent I/O error"}
	}

	err := WithRetry(ctx, config, fn)
	if err == nil {
		t.Error("WithRetry() expected error after all retries, got nil")
	}
	// 1 initial + 2 retries = 3 total calls
	if callCount != 3 {
		t.Errorf("Function called %d times, want 3", callCount)
	}

	archiveErr, ok := err.(*ArchiveError)
	if !ok {
		t.Errorf("Expected ArchiveError, got %T", err)
	} else if archiveErr.Code != ErrArchiveIOError {
		t.Errorf("Error code = %s, want %s", archiveErr.Code, ErrArchiveIOError)
	}
}

func TestWithRetry_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := RetryConfig{
		MaxRetries: 3,
		Delay:      1 * time.Second, // Long delay to test cancellation
		Backoff:    1.5,
	}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount == 1 {
			// Cancel after first call
			cancel()
		}
		return &ArchiveError{Code: ErrArchiveIOError, Message: "I/O error"}
	}

	err := WithRetry(ctx, config, fn)
	if err == nil {
		t.Error("WithRetry() expected error on cancel, got nil")
	}

	archiveErr, ok := err.(*ArchiveError)
	if !ok {
		t.Errorf("Expected ArchiveError, got %T", err)
	} else if archiveErr.Code != ErrArchiveOperationCancelled {
		t.Errorf("Error code = %s, want %s", archiveErr.Code, ErrArchiveOperationCancelled)
	}
}

func TestWithRetry_ContextAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := RetryConfig{
		MaxRetries: 3,
		Delay:      10 * time.Millisecond,
		Backoff:    1.5,
	}

	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	err := WithRetry(ctx, config, fn)
	if err == nil {
		t.Error("WithRetry() expected error on pre-cancelled context, got nil")
	}
	if callCount != 0 {
		t.Errorf("Function called %d times, want 0 (context already cancelled)", callCount)
	}
}
