package archive

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// ValidatePath checks if a path is safe (no path traversal)
func ValidatePath(path string) error {
	// Reject absolute paths
	if filepath.IsAbs(path) {
		return NewArchiveError(ErrArchivePathTraversal, "Absolute paths are not allowed in archives", nil)
	}

	// Normalize backslashes to forward slashes for cross-platform safety
	normalizedPath := strings.ReplaceAll(path, "\\", "/")

	// Clean the path
	cleaned := filepath.Clean(normalizedPath)

	// Check for parent directory references by splitting path
	parts := strings.Split(cleaned, string(filepath.Separator))
	for _, part := range parts {
		if part == ".." {
			return NewArchiveError(ErrArchivePathTraversal, "Path traversal detected (.. in path)", nil)
		}
	}

	// Check if cleaned path would escape the current directory
	if strings.HasPrefix(cleaned, "..") || strings.HasPrefix(cleaned, "/") {
		return NewArchiveError(ErrArchivePathTraversal, "Path would escape extraction directory", nil)
	}

	return nil
}

// CalculateFileHash calculates SHA256 hash of a file for TOCTOU protection
func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// VerifyFileHash verifies that file hash hasn't changed (TOCTOU protection)
func VerifyFileHash(filePath string, expectedHash string) error {
	currentHash, err := CalculateFileHash(filePath)
	if err != nil {
		return NewArchiveError(ErrArchiveIOError, "Failed to verify file integrity", err)
	}

	if currentHash != expectedHash {
		return NewArchiveError(ErrArchiveInternalError, "Archive was modified during processing", nil)
	}

	return nil
}

// CheckCompressionRatio checks if the compression ratio is suspicious (potential zip bomb)
func CheckCompressionRatio(archiveSize int64, extractedSize int64) bool {
	if archiveSize == 0 {
		return false // Cannot calculate ratio
	}

	// Calculate ratio (extracted / archived)
	ratio := float64(extractedSize) / float64(archiveSize)

	// Warn if ratio > 1:1000
	return ratio > 1000.0
}

// GetAvailableDiskSpace returns available disk space in bytes for the given path
func GetAvailableDiskSpace(path string) int64 {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return -1
	}

	// Available blocks * block size
	return int64(stat.Bavail) * int64(stat.Bsize)
}

// CheckDiskSpace checks if there is enough disk space for extraction
// Returns (insufficient bool, err error) where:
// - insufficient=true means not enough disk space
// - err != nil means disk space could not be checked
func CheckDiskSpace(path string, required int64) (bool, error) {
	available := GetAvailableDiskSpace(path)
	if available < 0 {
		return false, NewArchiveError(ErrArchiveIOError, "Failed to check available disk space", nil)
	}

	// Return true if required > available (insufficient space)
	return required > available, nil
}

// SanitizePathForCommand prevents option injection by prefixing paths starting with - with ./
// This is used when passing file paths to external commands like tar, unzip, 7z
func SanitizePathForCommand(path string) string {
	if strings.HasPrefix(path, "-") {
		return "./" + path
	}
	return path
}

// ValidateFileName checks if a filename is valid
func ValidateFileName(name string) error {
	if name == "" {
		return NewArchiveError(ErrArchiveInvalidName, "Filename cannot be empty", nil)
	}

	// Check for null bytes and control characters
	for _, c := range name {
		if c == 0 {
			return NewArchiveError(ErrArchiveInvalidName, "Filename contains null character", nil)
		}
		if c < 32 && c != '\t' {
			return NewArchiveError(ErrArchiveInvalidName, "Filename contains control characters", nil)
		}
	}

	return nil
}

// ValidateExtractedSymlinks scans an extraction directory and validates all symlinks
// to prevent path traversal attacks via malicious symlink targets
func ValidateExtractedSymlinks(extractDir string) error {
	absExtractDir, err := filepath.Abs(extractDir)
	if err != nil {
		return NewArchiveError(ErrArchiveIOError, "Failed to get absolute path of extraction directory", err)
	}

	return filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only check symlinks
		if info.Mode()&os.ModeSymlink == 0 {
			return nil
		}

		// Read the symlink target
		linkTarget, err := os.Readlink(path)
		if err != nil {
			return NewArchiveError(ErrArchiveIOError, "Failed to read symlink: "+path, err)
		}

		// Reject absolute path symlinks
		if filepath.IsAbs(linkTarget) {
			// Remove the dangerous symlink
			os.Remove(path)
			return NewArchiveError(ErrArchivePathTraversal,
				"Archive contains absolute path symlink: "+path, nil)
		}

		// Check for path traversal in symlink target
		if err := ValidatePath(linkTarget); err != nil {
			// Remove the dangerous symlink
			os.Remove(path)
			return NewArchiveError(ErrArchivePathTraversal,
				"Symlink target contains path traversal: "+path, nil)
		}

		// Verify resolved target is within extraction directory
		absLinkPath, _ := filepath.Abs(path)
		linkDir := filepath.Dir(absLinkPath)
		resolvedTarget, _ := filepath.Abs(filepath.Join(linkDir, linkTarget))

		if !strings.HasPrefix(resolvedTarget, absExtractDir+string(filepath.Separator)) &&
			resolvedTarget != absExtractDir {
			// Remove the dangerous symlink
			os.Remove(path)
			return NewArchiveError(ErrArchivePathTraversal,
				"Symlink points outside extraction directory: "+path, nil)
		}

		return nil
	})
}
