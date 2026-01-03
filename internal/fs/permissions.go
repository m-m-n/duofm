package fs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
)

// ProgressThreshold is the minimum number of files to trigger progress dialog
// According to FR7.7, progress dialog should be shown for batches with 10+ items
const ProgressThreshold = 10

// ValidatePermissionMode validates octal permission string (000-777)
func ValidatePermissionMode(mode string) error {
	// Check length
	if len(mode) != 3 {
		return errors.New("permission must be exactly 3 digits")
	}

	// Check each digit is 0-7
	for i, c := range mode {
		if c < '0' || c > '7' {
			return fmt.Errorf("invalid digit at position %d: must be 0-7", i+1)
		}
	}

	return nil
}

// ParsePermissionMode converts octal string to fs.FileMode
func ParsePermissionMode(mode string) (fs.FileMode, error) {
	// Validate first
	if err := ValidatePermissionMode(mode); err != nil {
		return 0, err
	}

	// Parse octal string
	perm, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse permission: %w", err)
	}

	return fs.FileMode(perm), nil
}

// FormatSymbolic converts fs.FileMode to symbolic string
// isDir: true for directories (shows 'd' prefix)
func FormatSymbolic(mode fs.FileMode, isDir bool) string {
	// Get permission bits only
	perm := mode.Perm()

	var result [10]rune

	// File type
	if isDir {
		result[0] = 'd'
	} else {
		result[0] = '-'
	}

	// Owner permissions
	if perm&0400 != 0 {
		result[1] = 'r'
	} else {
		result[1] = '-'
	}
	if perm&0200 != 0 {
		result[2] = 'w'
	} else {
		result[2] = '-'
	}
	if perm&0100 != 0 {
		result[3] = 'x'
	} else {
		result[3] = '-'
	}

	// Group permissions
	if perm&0040 != 0 {
		result[4] = 'r'
	} else {
		result[4] = '-'
	}
	if perm&0020 != 0 {
		result[5] = 'w'
	} else {
		result[5] = '-'
	}
	if perm&0010 != 0 {
		result[6] = 'x'
	} else {
		result[6] = '-'
	}

	// Other permissions
	if perm&0004 != 0 {
		result[7] = 'r'
	} else {
		result[7] = '-'
	}
	if perm&0002 != 0 {
		result[8] = 'w'
	} else {
		result[8] = '-'
	}
	if perm&0001 != 0 {
		result[9] = 'x'
	} else {
		result[9] = '-'
	}

	return string(result[:])
}

// ChangePermission changes permission of a single file/directory
func ChangePermission(path string, mode fs.FileMode) error {
	return os.Chmod(path, mode)
}

// PermissionError represents a permission change failure
type PermissionError struct {
	Path  string
	Error error
}

// ChangePermissionRecursive recursively changes permissions in a directory tree
// dirMode: permission for directories
// fileMode: permission for files
// Returns: success count, errors
func ChangePermissionRecursive(rootPath string, dirMode, fileMode fs.FileMode) (successCount int, errors []PermissionError) {
	return ChangePermissionRecursiveWithProgress(rootPath, dirMode, fileMode, nil)
}

// ChangePermissionRecursiveWithProgress recursively changes permissions with progress callback
// dirMode: permission for directories
// fileMode: permission for files
// progressCallback: called periodically with (processed, total, currentPath)
// Returns: success count, errors
func ChangePermissionRecursiveWithProgress(
	rootPath string,
	dirMode, fileMode fs.FileMode,
	progressCallback func(processed, total int, currentPath string),
) (successCount int, errors []PermissionError) {
	errors = make([]PermissionError, 0)

	// First, count total files for progress calculation
	totalFiles := 0
	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || (info != nil && info.Mode()&os.ModeSymlink != 0) {
			return nil
		}
		totalFiles++
		return nil
	})

	processed := 0

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errors = append(errors, PermissionError{Path: path, Error: err})
			return nil // Continue processing other files
		}

		// Skip symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		// Determine which mode to apply
		var targetMode fs.FileMode
		if info.IsDir() {
			targetMode = dirMode
		} else {
			targetMode = fileMode
		}

		// Change permission
		if err := os.Chmod(path, targetMode); err != nil {
			errors = append(errors, PermissionError{Path: path, Error: err})
			return nil // Continue processing
		}

		successCount++
		processed++

		// Call progress callback if provided
		if progressCallback != nil {
			progressCallback(processed, totalFiles, path)
		}

		return nil
	})

	// If Walk itself failed, add that error
	if err != nil {
		errors = append(errors, PermissionError{Path: rootPath, Error: err})
	}

	return successCount, errors
}
