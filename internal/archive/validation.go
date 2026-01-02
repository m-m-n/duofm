package archive

import (
	"os"
)

// ValidateCompressionLevel checks if compression level is in valid range (0-9)
func ValidateCompressionLevel(level int) error {
	if level < 0 || level > 9 {
		return NewArchiveError(ErrArchiveInternalError, "Compression level must be between 0 and 9", nil)
	}
	return nil
}

// ValidateSources checks if source list is valid and all sources are accessible
func ValidateSources(sources []string) error {
	if sources == nil || len(sources) == 0 {
		return NewArchiveError(ErrArchiveSourceNotFound, "No source files specified", nil)
	}

	// Check each source exists and is readable
	for _, src := range sources {
		info, err := os.Stat(src)
		if os.IsNotExist(err) {
			return NewArchiveError(ErrArchiveSourceNotFound, "Source not found: "+src, err)
		}
		if os.IsPermission(err) {
			return NewArchiveError(ErrArchivePermissionDeniedRead, "Cannot read source: "+src, err)
		}
		if err != nil {
			return NewArchiveError(ErrArchiveIOError, "Cannot access source: "+src, err)
		}

		// Try to open to verify access (works for both files and directories)
		f, err := os.Open(src)
		if err != nil {
			if os.IsPermission(err) {
				if info.IsDir() {
					return NewArchiveError(ErrArchivePermissionDeniedRead, "Cannot read directory: "+src, err)
				}
				return NewArchiveError(ErrArchivePermissionDeniedRead, "Cannot read file: "+src, err)
			}
			if info.IsDir() {
				return NewArchiveError(ErrArchiveIOError, "Cannot open directory: "+src, err)
			}
			return NewArchiveError(ErrArchiveIOError, "Cannot open file: "+src, err)
		}
		f.Close()
	}

	return nil
}
