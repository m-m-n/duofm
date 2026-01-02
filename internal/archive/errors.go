package archive

import (
	"context"
	"fmt"
	"time"
)

// Error codes for archive operations
const (
	ErrArchiveSourceNotFound        = "ERR_ARCHIVE_001" // Source file not found
	ErrArchivePermissionDeniedRead  = "ERR_ARCHIVE_002" // Permission denied (read)
	ErrArchivePermissionDeniedWrite = "ERR_ARCHIVE_003" // Permission denied (write)
	ErrArchiveDiskSpaceInsufficient = "ERR_ARCHIVE_004" // Disk space insufficient
	ErrArchiveUnsupportedFormat     = "ERR_ARCHIVE_005" // Unsupported format
	ErrArchiveCorrupted             = "ERR_ARCHIVE_006" // Corrupted archive
	ErrArchiveInvalidName           = "ERR_ARCHIVE_007" // Invalid archive name
	ErrArchivePathTraversal         = "ERR_ARCHIVE_008" // Path traversal detected
	ErrArchiveCompressionBomb       = "ERR_ARCHIVE_009" // Compression bomb detected
	ErrArchiveOperationCancelled    = "ERR_ARCHIVE_010" // Operation cancelled
	ErrArchiveIOError               = "ERR_ARCHIVE_011" // I/O error (retriable)
	ErrArchiveInternalError         = "ERR_ARCHIVE_012" // Internal error
)

// Retry configuration constants
const (
	DefaultMaxRetries   = 3
	DefaultRetryDelay   = 1 * time.Second
	DefaultRetryBackoff = 1.5 // Exponential backoff multiplier
)

// ArchiveError represents a structured error with code and details
type ArchiveError struct {
	Code    string // Error code (ERR_ARCHIVE_XXX)
	Message string // User-friendly message
	Details string // Technical details for logging
	Cause   error  // Underlying error (if any)
}

func (e *ArchiveError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *ArchiveError) Unwrap() error {
	return e.Cause
}

// NewArchiveError creates a new ArchiveError with the given code and message
func NewArchiveError(code, message string, cause error) *ArchiveError {
	return &ArchiveError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewArchiveErrorWithDetails creates a new ArchiveError with detailed information
func NewArchiveErrorWithDetails(code, message, details string, cause error) *ArchiveError {
	return &ArchiveError{
		Code:    code,
		Message: message,
		Details: details,
		Cause:   cause,
	}
}

// IsRetriable returns true if the error is a temporary error that can be retried
func IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	archiveErr, ok := err.(*ArchiveError)
	if !ok {
		return false
	}

	// I/O errors are typically temporary and can be retried
	return archiveErr.Code == ErrArchiveIOError
}

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries int           // Maximum number of retry attempts
	Delay      time.Duration // Initial delay between retries
	Backoff    float64       // Backoff multiplier for exponential backoff
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: DefaultMaxRetries,
		Delay:      DefaultRetryDelay,
		Backoff:    DefaultRetryBackoff,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// WithRetry executes a function with retry logic for temporary errors
func WithRetry(ctx context.Context, config RetryConfig, fn RetryableFunc) error {
	var lastErr error
	delay := config.Delay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return NewArchiveError(ErrArchiveOperationCancelled, "Operation cancelled during retry", ctx.Err())
		default:
		}

		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Only retry if the error is retriable
		if !IsRetriable(err) {
			return err
		}

		// Don't wait after the last attempt
		if attempt < config.MaxRetries {
			select {
			case <-ctx.Done():
				return NewArchiveError(ErrArchiveOperationCancelled, "Operation cancelled during retry wait", ctx.Err())
			case <-time.After(delay):
				// Apply exponential backoff for next attempt
				delay = time.Duration(float64(delay) * config.Backoff)
			}
		}
	}

	return NewArchiveErrorWithDetails(
		lastErr.(*ArchiveError).Code,
		fmt.Sprintf("Operation failed after %d retries: %s", config.MaxRetries+1, lastErr.(*ArchiveError).Message),
		fmt.Sprintf("All %d retry attempts exhausted", config.MaxRetries+1),
		lastErr,
	)
}
