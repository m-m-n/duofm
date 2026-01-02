package archive

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ExtractionMethod specifies how to extract an archive
type ExtractionMethod int

const (
	ExtractDirect      ExtractionMethod = iota // Extract directly (single root directory)
	ExtractToDirectory                         // Extract to new directory (multiple root items)
)

// ExtractionStrategy contains the extraction method and optional directory name
type ExtractionStrategy struct {
	Method        ExtractionMethod
	DirectoryName string // Used when Method is ExtractToDirectory
}

// ArchiveMetadata contains information about an archive for security checks
type ArchiveMetadata struct {
	ArchiveSize   int64    // Size of the archive file in bytes
	ExtractedSize int64    // Total size of extracted contents in bytes
	FileCount     int      // Number of files in the archive
	Files         []string // List of file paths in the archive
}

// SmartExtractor analyzes archive structure and determines extraction strategy
type SmartExtractor struct {
	tarExecutor      *TarExecutor
	zipExecutor      *ZipExecutor
	sevenZipExecutor *SevenZipExecutor
}

// NewSmartExtractor creates a new SmartExtractor instance
func NewSmartExtractor() *SmartExtractor {
	return &SmartExtractor{
		tarExecutor:      NewTarExecutor(),
		zipExecutor:      NewZipExecutor(),
		sevenZipExecutor: NewSevenZipExecutor(),
	}
}

// AnalyzeStructure examines an archive and determines the best extraction strategy
func (s *SmartExtractor) AnalyzeStructure(ctx context.Context, archivePath string, format ArchiveFormat) (*ExtractionStrategy, error) {
	var contents []string
	var err error

	// List archive contents based on format
	switch format {
	case FormatTar, FormatTarGz, FormatTarBz2, FormatTarXz:
		contents, err = s.tarExecutor.ListContents(ctx, format, archivePath)
	case FormatZip:
		contents, err = s.zipExecutor.ListContents(ctx, archivePath)
	case Format7z:
		contents, err = s.sevenZipExecutor.ListContents(ctx, archivePath)
	default:
		return nil, NewArchiveError(ErrArchiveUnsupportedFormat, "Unsupported archive format", nil)
	}

	if err != nil {
		return nil, err
	}

	return s.analyzeContents(contents), nil
}

// GetArchiveMetadata retrieves metadata about an archive for security checks
func (s *SmartExtractor) GetArchiveMetadata(ctx context.Context, archivePath string, format ArchiveFormat) (*ArchiveMetadata, error) {
	var stdout string
	var err error
	var executor = NewCommandExecutor()

	// Get archive contents with size information
	switch format {
	case FormatTar, FormatTarGz, FormatTarBz2, FormatTarXz:
		var flags string
		switch format {
		case FormatTar:
			flags = "-tvf"
		case FormatTarGz:
			flags = "-tzvf"
		case FormatTarBz2:
			flags = "-tjvf"
		case FormatTarXz:
			flags = "-tJvf"
		default:
			flags = "-tvf"
		}
		// Sanitize archive path to prevent option injection (paths starting with -)
		safePath := SanitizePathForCommand(archivePath)
		stdout, _, err = executor.ExecuteCommand(ctx, "tar", flags, safePath)
		if err != nil {
			return nil, err
		}
		return s.parseTarOutput(archivePath, stdout)

	case FormatZip:
		// Sanitize archive path to prevent option injection
		safePath := SanitizePathForCommand(archivePath)
		stdout, _, err = executor.ExecuteCommand(ctx, "unzip", "-l", safePath)
		if err != nil {
			return nil, err
		}
		return s.parseZipOutput(archivePath, stdout)

	case Format7z:
		// Sanitize archive path to prevent option injection
		safePath := SanitizePathForCommand(archivePath)
		stdout, _, err = executor.ExecuteCommand(ctx, "7z", "l", safePath)
		if err != nil {
			return nil, err
		}
		return s.parse7zOutput(archivePath, stdout)

	default:
		return nil, NewArchiveError(ErrArchiveUnsupportedFormat, "Unsupported archive format", nil)
	}
}

// parseTarOutput parses tar -tvf output and extracts size information
func (s *SmartExtractor) parseTarOutput(archivePath string, output string) (*ArchiveMetadata, error) {
	lines := strings.Split(output, "\n")
	var totalSize int64
	var files []string

	for _, line := range lines {
		if line == "" {
			continue
		}
		// Format: permissions user/group size date time filename
		// Example: -rw-r--r-- user/group   12345 2024-01-01 10:00 file.txt
		// Symlink: lrwxrwxrwx user/group       0 2024-01-01 10:00 link -> target
		// Hardlink: -rw-r--r-- user/group       0 2024-01-01 10:00 hardlink link to target
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			permissions := fields[0]

			// Size is the 3rd field (index 2)
			size, err := strconv.ParseInt(fields[2], 10, 64)
			if err == nil {
				totalSize += size
			}

			// Filename is everything after the 5th field (to handle spaces in filenames)
			filename := strings.Join(fields[5:], " ")

			// Check for symlinks (permissions start with 'l')
			if strings.HasPrefix(permissions, "l") {
				// Symlink format: "link -> target"
				parts := strings.Split(filename, " -> ")
				if len(parts) == 2 {
					target := parts[1]
					// Reject absolute path symlinks
					if filepath.IsAbs(target) {
						return nil, NewArchiveError(ErrArchivePathTraversal,
							"Archive contains absolute path symlink: "+filename, nil)
					}
					// Validate symlink target for path traversal
					if err := ValidatePath(target); err != nil {
						return nil, NewArchiveError(ErrArchivePathTraversal,
							"Symlink target contains path traversal: "+filename, nil)
					}
					filename = parts[0] // Use only the link name for file list
				}
			}

			// Check for hardlinks (contains " link to ")
			// Format: "hardlink link to target"
			if strings.Contains(filename, " link to ") {
				parts := strings.SplitN(filename, " link to ", 2)
				if len(parts) == 2 {
					target := parts[1]
					// Reject absolute path hardlinks
					if filepath.IsAbs(target) {
						return nil, NewArchiveError(ErrArchivePathTraversal,
							"Archive contains absolute path hardlink: "+filename, nil)
					}
					// Validate hardlink target for path traversal
					if err := ValidatePath(target); err != nil {
						return nil, NewArchiveError(ErrArchivePathTraversal,
							"Hardlink target contains path traversal: "+filename, nil)
					}
					filename = parts[0] // Use only the link name for file list
				}
			}

			// Validate path for security (path traversal check)
			if err := ValidatePath(filename); err != nil {
				return nil, err
			}

			files = append(files, filename)
		}
	}

	// Get archive file size
	archiveSize := s.getFileSize(archivePath)

	return &ArchiveMetadata{
		ArchiveSize:   archiveSize,
		ExtractedSize: totalSize,
		FileCount:     len(files),
		Files:         files,
	}, nil
}

// parseZipOutput parses unzip -l output and extracts size information
func (s *SmartExtractor) parseZipOutput(archivePath string, output string) (*ArchiveMetadata, error) {
	lines := strings.Split(output, "\n")
	var totalSize int64
	var files []string
	inFileList := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip header lines
		if strings.Contains(line, "Length") || strings.Contains(line, "---") {
			inFileList = true
			continue
		}

		if !inFileList {
			continue
		}

		// Stop at footer
		if strings.Contains(line, "files") && strings.Contains(line, "bytes") {
			break
		}

		// Parse file entry
		// Format: Length Date Time Name
		// Example:     1024  2024-01-01 10:00   file with spaces.txt
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			// First field is size
			size, err := strconv.ParseInt(fields[0], 10, 64)
			if err == nil {
				totalSize += size
			}
			// Filename is everything after the 3rd field (to handle spaces in filenames)
			filename := strings.Join(fields[3:], " ")

			// Validate path for security (path traversal check)
			if err := ValidatePath(filename); err != nil {
				return nil, err
			}

			files = append(files, filename)
		}
	}

	// Get archive file size
	archiveSize := s.getFileSize(archivePath)

	return &ArchiveMetadata{
		ArchiveSize:   archiveSize,
		ExtractedSize: totalSize,
		FileCount:     len(files),
		Files:         files,
	}, nil
}

// parse7zOutput parses 7z l output and extracts size information
func (s *SmartExtractor) parse7zOutput(archivePath string, output string) (*ArchiveMetadata, error) {
	lines := strings.Split(output, "\n")
	var totalSize int64
	var files []string
	inFileList := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Start of file list (after separator line)
		if strings.Contains(line, "---") {
			inFileList = !inFileList
			continue
		}

		if !inFileList {
			continue
		}

		// Parse file entry
		// Format: Date Time Attr Size Compressed Name
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			// Size is the 4th field (index 3)
			size, err := strconv.ParseInt(fields[3], 10, 64)
			if err == nil {
				totalSize += size
			}
			// Filename is everything after the 5th field
			filename := strings.Join(fields[5:], " ")

			// Validate path for security (path traversal check)
			if err := ValidatePath(filename); err != nil {
				return nil, err
			}

			files = append(files, filename)
		}
	}

	// Get archive file size
	archiveSize := s.getFileSize(archivePath)

	return &ArchiveMetadata{
		ArchiveSize:   archiveSize,
		ExtractedSize: totalSize,
		FileCount:     len(files),
		Files:         files,
	}, nil
}

// getFileSize returns the size of a file in bytes
func (s *SmartExtractor) getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// analyzeContents determines extraction strategy based on archive contents
func (s *SmartExtractor) analyzeContents(contents []string) *ExtractionStrategy {
	if len(contents) == 0 {
		return &ExtractionStrategy{
			Method: ExtractDirect,
		}
	}

	rootItems := s.getRootItems(contents)

	// If only one root item, extract directly
	if len(rootItems) == 1 {
		return &ExtractionStrategy{
			Method: ExtractDirect,
		}
	}

	// Multiple root items - need to create directory
	return &ExtractionStrategy{
		Method: ExtractToDirectory,
	}
}

// getRootItems returns unique root-level items in the archive
func (s *SmartExtractor) getRootItems(contents []string) []string {
	rootMap := make(map[string]bool)

	for _, item := range contents {
		// Clean the path
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		// Get the root component
		parts := strings.Split(item, string(filepath.Separator))
		if len(parts) > 0 {
			root := parts[0]
			// Remove trailing slash if present
			root = strings.TrimSuffix(root, "/")
			if root != "" {
				rootMap[root] = true
			}
		}
	}

	// Convert map to slice
	roots := make([]string, 0, len(rootMap))
	for root := range rootMap {
		roots = append(roots, root)
	}

	return roots
}
