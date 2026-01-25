package fs

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

// TrashinfoFormat is the time format for DeletionDate in .trashinfo files
// ISO 8601 format without timezone suffix (freedesktop.org spec)
const TrashinfoFormat = "2006-01-02T15:04:05"

// TrashInfo holds parsed information from a .trashinfo file
type TrashInfo struct {
	OriginalPath string
	DeletionDate time.Time
}

// GenerateTrashinfo generates the content of a .trashinfo file
// Path is URL-encoded, DeletionDate is in ISO 8601 format (local time, no timezone)
func GenerateTrashinfo(originalPath string, deletionTime time.Time) string {
	encodedPath := urlEncodePath(originalPath)
	dateStr := deletionTime.Format(TrashinfoFormat)

	return fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n", encodedPath, dateStr)
}

// ParseTrashinfo parses a .trashinfo file and returns the original path and deletion date
func ParseTrashinfo(trashinfoPath string) (*TrashInfo, error) {
	file, err := os.Open(trashinfoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open trashinfo: %w", err)
	}
	defer file.Close()

	var (
		hasHeader    bool
		originalPath string
		deletionDate string
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "[Trash Info]" {
			hasHeader = true
			continue
		}

		if strings.HasPrefix(line, "Path=") {
			originalPath = strings.TrimPrefix(line, "Path=")
		}
		if strings.HasPrefix(line, "DeletionDate=") {
			deletionDate = strings.TrimPrefix(line, "DeletionDate=")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read trashinfo: %w", err)
	}

	if !hasHeader {
		return nil, fmt.Errorf("invalid trashinfo: missing [Trash Info] header")
	}

	if originalPath == "" {
		return nil, fmt.Errorf("invalid trashinfo: missing Path")
	}

	if deletionDate == "" {
		return nil, fmt.Errorf("invalid trashinfo: missing DeletionDate")
	}

	// URL decode the path
	decodedPath, err := urlDecodePath(originalPath)
	if err != nil {
		return nil, fmt.Errorf("invalid trashinfo: failed to decode Path: %w", err)
	}

	// Parse the deletion date
	parsedDate, err := time.ParseInLocation(TrashinfoFormat, deletionDate, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid trashinfo: invalid DeletionDate format: %w", err)
	}

	return &TrashInfo{
		OriginalPath: decodedPath,
		DeletionDate: parsedDate,
	}, nil
}

// urlEncodePath encodes a file path for use in .trashinfo files
// Only encodes characters that need escaping while preserving path separators
func urlEncodePath(path string) string {
	// Split path into components and encode each
	parts := strings.Split(path, "/")
	encodedParts := make([]string, len(parts))

	for i, part := range parts {
		// Use url.PathEscape which encodes spaces as %20 (not +)
		// but preserves safe characters
		encodedParts[i] = url.PathEscape(part)
	}

	return strings.Join(encodedParts, "/")
}

// urlDecodePath decodes a URL-encoded path from .trashinfo files
func urlDecodePath(encoded string) (string, error) {
	// Use url.PathUnescape to decode
	decoded, err := url.PathUnescape(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode path: %w", err)
	}
	return decoded, nil
}

// WriteTrashinfo writes a .trashinfo file to the specified path
func WriteTrashinfo(trashinfoPath, originalPath string, deletionTime time.Time) error {
	content := GenerateTrashinfo(originalPath, deletionTime)
	return os.WriteFile(trashinfoPath, []byte(content), 0644)
}
