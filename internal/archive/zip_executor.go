package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ZipExecutor handles zip archive operations
type ZipExecutor struct {
	executor *CommandExecutor
}

// NewZipExecutor creates a new ZipExecutor instance
func NewZipExecutor() *ZipExecutor {
	return &ZipExecutor{
		executor: NewCommandExecutor(),
	}
}

// ParseZipCompressOutput parses a line of zip verbose output and returns the filename.
// zip -v output format: "  adding: filename (deflated 45%)" or "  adding: filename (stored 0%)"
// Returns the filename or empty string if parsing fails.
func ParseZipCompressOutput(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}

	// Look for "adding:" prefix
	const prefix = "adding:"
	idx := strings.Index(line, prefix)
	if idx == -1 {
		return ""
	}

	// Extract the part after "adding:"
	rest := strings.TrimSpace(line[idx+len(prefix):])
	if rest == "" {
		return ""
	}

	// The filename ends at the first " (" which precedes compression info
	parenIdx := strings.Index(rest, " (")
	if parenIdx > 0 {
		return strings.TrimSpace(rest[:parenIdx])
	}

	// Fallback: return the whole rest if no parenthesis found
	return rest
}

// sanitizeZipPath prevents option injection by prefixing paths starting with - with ./
func sanitizeZipPath(path string) string {
	if strings.HasPrefix(path, "-") {
		return "./" + path
	}
	return path
}

// ParseZipExtractOutput parses a line of unzip verbose output and returns the filename.
// unzip output format: "  inflating: filename" or "   creating: dirname/"
// Returns the filename or empty string if parsing fails.
func ParseZipExtractOutput(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}

	// Look for "inflating:" or "creating:" or "extracting:"
	for _, prefix := range []string{"inflating:", "creating:", "extracting:"} {
		idx := strings.Index(line, prefix)
		if idx != -1 {
			return strings.TrimSpace(line[idx+len(prefix):])
		}
	}

	return ""
}

// Compress creates a zip archive
func (e *ZipExecutor) Compress(ctx context.Context, sources []string, output string, level int, progressChan chan<- *ProgressUpdate) error {
	// Validate output path - must be absolute to ensure safety
	absOutput, err := filepath.Abs(output)
	if err != nil {
		return NewArchiveError(ErrArchiveInternalError, fmt.Sprintf("Failed to get absolute output path: %s", output), err)
	}
	outputDir := filepath.Dir(absOutput)

	// Ensure output directory exists and is accessible
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return NewArchiveError(ErrArchivePermissionDeniedWrite, fmt.Sprintf("Output directory does not exist: %s", outputDir), err)
	}

	// Convert sources to absolute paths
	absSources := make([]string, len(sources))
	for i, src := range sources {
		abs, err := filepath.Abs(src)
		if err != nil {
			return NewArchiveError(ErrArchiveInternalError, fmt.Sprintf("Failed to get absolute path: %s", src), err)
		}
		absSources[i] = abs
	}

	// Get base directory for running zip command
	// Using cmd.Dir instead of os.Chdir() avoids race conditions
	baseDir := filepath.Dir(absSources[0])

	// Convert sources to relative paths from baseDir
	relSources := make([]string, len(absSources))
	for i, abs := range absSources {
		rel, err := filepath.Rel(baseDir, abs)
		if err != nil {
			relSources[i] = filepath.Base(abs)
		} else {
			relSources[i] = rel
		}
	}

	// Calculate total files and size for progress
	var totalFiles int
	var totalBytes int64
	for _, src := range absSources {
		count, size := e.countFilesAndSize(src)
		totalFiles += count
		totalBytes += size
	}

	// Send initial progress
	if progressChan != nil {
		progressChan <- &ProgressUpdate{
			ProcessedFiles: 0,
			TotalFiles:     totalFiles,
			ProcessedBytes: 0,
			TotalBytes:     totalBytes,
			CurrentFile:    "",
			Operation:      "compress",
			ArchivePath:    absOutput,
		}
	}

	// Create temporary output file for atomic operation
	// Use os.CreateTemp to generate unique filename, then remove the empty file
	// (zip needs to create a new file, not write to an existing one)
	tempFile, err := os.CreateTemp(outputDir, ".duofm-archive-*.tmp")
	if err != nil {
		return NewArchiveError(ErrArchiveIOError, "Failed to create temporary file", err)
	}
	tempOutput := tempFile.Name()
	tempFile.Close()
	os.Remove(tempOutput) // Remove empty file so zip can create it fresh

	// Build arguments: zip -r -level output sources...
	// Run zip in baseDir to preserve relative paths (using cmd.Dir for thread safety)
	// Sanitize filenames that start with - to prevent option injection
	args := []string{"-r", fmt.Sprintf("-%d", level), tempOutput}
	for _, src := range relSources {
		args = append(args, sanitizeZipPath(src))
	}

	// Use streaming progress if channel is provided
	if progressChan != nil {
		processedFiles := 0
		lineHandler := func(line string) error {
			filename := ParseZipCompressOutput(line)
			if filename != "" {
				processedFiles++
				progressChan <- &ProgressUpdate{
					ProcessedFiles: processedFiles,
					TotalFiles:     totalFiles,
					ProcessedBytes: 0,
					TotalBytes:     totalBytes,
					CurrentFile:    filename,
					Operation:      "compress",
					ArchivePath:    absOutput,
				}
			}
			return nil
		}

		stderr, err := e.executor.ExecuteCommandWithProgress(ctx, baseDir, lineHandler, "zip", args...)
		if err != nil {
			os.Remove(tempOutput)
			return NewArchiveErrorWithDetails(
				ErrArchiveInternalError,
				"Failed to create zip archive",
				stderr,
				err,
			)
		}
	} else {
		// Fallback to non-streaming execution
		_, stderr, err := e.executor.ExecuteCommandInDir(ctx, baseDir, "zip", args...)
		if err != nil {
			os.Remove(tempOutput)
			return NewArchiveErrorWithDetails(
				ErrArchiveInternalError,
				"Failed to create zip archive",
				stderr,
				err,
			)
		}
	}

	// Atomic rename from temp to final output
	if err := os.Rename(tempOutput, absOutput); err != nil {
		os.Remove(tempOutput)
		return NewArchiveError(ErrArchiveIOError, "Failed to finalize archive", err)
	}

	// Send completion progress
	if progressChan != nil {
		progressChan <- &ProgressUpdate{
			ProcessedFiles: totalFiles,
			TotalFiles:     totalFiles,
			ProcessedBytes: totalBytes,
			TotalBytes:     totalBytes,
			CurrentFile:    "",
			Operation:      "compress",
			ArchivePath:    absOutput,
		}
	}

	return nil
}

// calculateSize calculates the total size of a file or directory
func (e *ZipExecutor) calculateSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// countFilesAndSize counts the total number of files and total size in a path
func (e *ZipExecutor) countFilesAndSize(path string) (int, int64) {
	var count int
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil {
			count++
			if !info.IsDir() {
				size += info.Size()
			}
		}
		return nil
	})
	return count, size
}

// Extract extracts a zip archive
func (e *ZipExecutor) Extract(ctx context.Context, archivePath string, destDir string, progressChan chan<- *ProgressUpdate) error {
	// Get absolute paths
	absArchivePath, err := filepath.Abs(archivePath)
	if err != nil {
		return NewArchiveError(ErrArchiveInternalError, fmt.Sprintf("Failed to get absolute archive path: %s", archivePath), err)
	}
	absDestDir, err := filepath.Abs(destDir)
	if err != nil {
		return NewArchiveError(ErrArchiveInternalError, fmt.Sprintf("Failed to get absolute destination path: %s", destDir), err)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(absDestDir, 0755); err != nil {
		return NewArchiveError(ErrArchivePermissionDeniedWrite, fmt.Sprintf("Cannot create directory: %s", absDestDir), err)
	}

	// Get archive size for progress
	archiveInfo, err := os.Stat(absArchivePath)
	var archiveSize int64
	if err == nil {
		archiveSize = archiveInfo.Size()
	}

	// Count total files in archive for accurate progress
	totalFiles := e.countArchiveFiles(ctx, absArchivePath)
	if totalFiles == 0 {
		totalFiles = 1 // Fallback
	}

	// Send initial progress
	if progressChan != nil {
		progressChan <- &ProgressUpdate{
			ProcessedFiles: 0,
			TotalFiles:     totalFiles,
			ProcessedBytes: 0,
			TotalBytes:     archiveSize,
			CurrentFile:    filepath.Base(absArchivePath),
			Operation:      "extract",
			ArchivePath:    absArchivePath,
		}
	}

	// unzip archive -d destDir
	// Sanitize archive path to prevent option injection
	safePath := SanitizePathForCommand(absArchivePath)
	args := []string{"-d", absDestDir, safePath}

	// Use streaming progress if channel is provided
	if progressChan != nil {
		processedFiles := 0
		lineHandler := func(line string) error {
			filename := ParseZipExtractOutput(line)
			if filename != "" {
				processedFiles++
				progressChan <- &ProgressUpdate{
					ProcessedFiles: processedFiles,
					TotalFiles:     totalFiles,
					ProcessedBytes: 0,
					TotalBytes:     archiveSize,
					CurrentFile:    filename,
					Operation:      "extract",
					ArchivePath:    absArchivePath,
				}
			}
			return nil
		}

		stderr, err := e.executor.ExecuteCommandWithProgress(ctx, "", lineHandler, "unzip", args...)
		if err != nil {
			return NewArchiveErrorWithDetails(
				ErrArchiveInternalError,
				"Failed to extract zip archive",
				stderr,
				err,
			)
		}
	} else {
		// Fallback to non-streaming execution
		_, stderr, err := e.executor.ExecuteCommand(ctx, "unzip", args...)
		if err != nil {
			return NewArchiveErrorWithDetails(
				ErrArchiveInternalError,
				"Failed to extract zip archive",
				stderr,
				err,
			)
		}
	}

	// Validate extracted symlinks to prevent path traversal attacks
	if err := ValidateExtractedSymlinks(absDestDir); err != nil {
		return err
	}

	// Send completion progress
	if progressChan != nil {
		progressChan <- &ProgressUpdate{
			ProcessedFiles: totalFiles,
			TotalFiles:     totalFiles,
			ProcessedBytes: archiveSize,
			TotalBytes:     archiveSize,
			CurrentFile:    "",
			Operation:      "extract",
			ArchivePath:    absArchivePath,
		}
	}

	return nil
}

// ListContents lists the contents of a zip archive
func (e *ZipExecutor) ListContents(ctx context.Context, archivePath string) ([]string, error) {
	// unzip -l archive
	// Sanitize archive path to prevent option injection
	safePath := SanitizePathForCommand(archivePath)
	stdout, stderr, err := e.executor.ExecuteCommand(ctx, "unzip", "-l", safePath)
	if err != nil {
		return nil, NewArchiveErrorWithDetails(
			ErrArchiveInternalError,
			"Failed to list zip archive contents",
			stderr,
			err,
		)
	}

	// Parse unzip -l output
	// Format:
	//   Length      Date    Time    Name
	//   ---------  ---------- -----   ----
	//   0  2024-01-01 10:00   file.txt
	lines := strings.Split(stdout, "\n")
	var contents []string
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

		// Parse file entry
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			// Last field is filename
			filename := fields[len(fields)-1]
			contents = append(contents, filename)
		}
	}

	return contents, nil
}

// countArchiveFiles counts the number of files in a zip archive
func (e *ZipExecutor) countArchiveFiles(ctx context.Context, archivePath string) int {
	contents, err := e.ListContents(ctx, archivePath)
	if err != nil {
		return 0
	}
	return len(contents)
}
