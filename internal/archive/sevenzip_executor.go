package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SevenZipExecutor handles 7z archive operations
type SevenZipExecutor struct {
	executor *CommandExecutor
}

// NewSevenZipExecutor creates a new SevenZipExecutor instance
func NewSevenZipExecutor() *SevenZipExecutor {
	return &SevenZipExecutor{
		executor: NewCommandExecutor(),
	}
}

// sanitize7zPath prevents option injection by prefixing paths starting with - with ./
func sanitize7zPath(path string) string {
	if strings.HasPrefix(path, "-") {
		return "./" + path
	}
	return path
}

// Parse7zCompressOutput parses a line of 7z verbose output and returns the filename.
// 7z output format during compression: "+ filename" or "Compressing  filename"
// Returns the filename or empty string if parsing fails.
func Parse7zCompressOutput(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}

	// 7z uses "+" prefix for files being added
	if strings.HasPrefix(line, "+ ") {
		return strings.TrimSpace(line[2:])
	}

	// Also check for "Compressing" prefix
	const compressingPrefix = "Compressing"
	if strings.HasPrefix(line, compressingPrefix) {
		rest := strings.TrimSpace(line[len(compressingPrefix):])
		return rest
	}

	return ""
}

// Parse7zExtractOutput parses a line of 7z extract output and returns the filename.
// 7z output format during extraction: "- filename" or "Extracting  filename"
// Returns the filename or empty string if parsing fails.
func Parse7zExtractOutput(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}

	// 7z uses "- " prefix for files being extracted
	if strings.HasPrefix(line, "- ") {
		return strings.TrimSpace(line[2:])
	}

	// Also check for "Extracting" prefix
	const extractingPrefix = "Extracting"
	if strings.HasPrefix(line, extractingPrefix) {
		rest := strings.TrimSpace(line[len(extractingPrefix):])
		return rest
	}

	return ""
}

// Compress creates a 7z archive
func (e *SevenZipExecutor) Compress(ctx context.Context, sources []string, output string, level int, progressChan chan<- *ProgressUpdate) error {
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

	// Get base directory for running 7z command
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
	// (7z needs to create a new file, not write to an existing one)
	tempFile, err := os.CreateTemp(outputDir, ".duofm-archive-*.tmp")
	if err != nil {
		return NewArchiveError(ErrArchiveIOError, "Failed to create temporary file", err)
	}
	tempOutput := tempFile.Name()
	tempFile.Close()
	os.Remove(tempOutput) // Remove empty file so 7z can create it fresh

	// Build arguments: 7z a -mx=level output sources...
	// Run 7z in baseDir to preserve relative paths (using cmd.Dir for thread safety)
	// Use -- to separate options from filenames to prevent option injection
	args := []string{"a", fmt.Sprintf("-mx=%d", level), tempOutput, "--"}
	// Sanitize filenames that start with - to prevent option injection
	for _, src := range relSources {
		args = append(args, sanitize7zPath(src))
	}

	// Use streaming progress if channel is provided
	if progressChan != nil {
		processedFiles := 0
		lineHandler := func(line string) error {
			filename := Parse7zCompressOutput(line)
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

		stderr, err := e.executor.ExecuteCommandWithProgress(ctx, baseDir, lineHandler, "7z", args...)
		if err != nil {
			os.Remove(tempOutput)
			return NewArchiveErrorWithDetails(
				ErrArchiveInternalError,
				"Failed to create 7z archive",
				stderr,
				err,
			)
		}
	} else {
		// Fallback to non-streaming execution
		_, stderr, err := e.executor.ExecuteCommandInDir(ctx, baseDir, "7z", args...)
		if err != nil {
			os.Remove(tempOutput)
			return NewArchiveErrorWithDetails(
				ErrArchiveInternalError,
				"Failed to create 7z archive",
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
func (e *SevenZipExecutor) calculateSize(path string) int64 {
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
func (e *SevenZipExecutor) countFilesAndSize(path string) (int, int64) {
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

// Extract extracts a 7z archive
func (e *SevenZipExecutor) Extract(ctx context.Context, archivePath string, destDir string, progressChan chan<- *ProgressUpdate) error {
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

	// 7z x archive -o<destDir>
	// Sanitize archive path to prevent option injection
	safePath := SanitizePathForCommand(absArchivePath)
	args := []string{"x", "-o" + absDestDir, "-y", safePath}

	// Use streaming progress if channel is provided
	if progressChan != nil {
		processedFiles := 0
		lineHandler := func(line string) error {
			filename := Parse7zExtractOutput(line)
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

		stderr, err := e.executor.ExecuteCommandWithProgress(ctx, "", lineHandler, "7z", args...)
		if err != nil {
			return NewArchiveErrorWithDetails(
				ErrArchiveInternalError,
				"Failed to extract 7z archive",
				stderr,
				err,
			)
		}
	} else {
		// Fallback to non-streaming execution
		_, stderr, err := e.executor.ExecuteCommand(ctx, "7z", args...)
		if err != nil {
			return NewArchiveErrorWithDetails(
				ErrArchiveInternalError,
				"Failed to extract 7z archive",
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

// countArchiveFiles counts the number of files in a 7z archive
func (e *SevenZipExecutor) countArchiveFiles(ctx context.Context, archivePath string) int {
	contents, err := e.ListContents(ctx, archivePath)
	if err != nil {
		return 0
	}
	return len(contents)
}

// ListContents lists the contents of a 7z archive
func (e *SevenZipExecutor) ListContents(ctx context.Context, archivePath string) ([]string, error) {
	// 7z l archive
	// Sanitize archive path to prevent option injection
	safePath := SanitizePathForCommand(archivePath)
	stdout, stderr, err := e.executor.ExecuteCommand(ctx, "7z", "l", safePath)
	if err != nil {
		return nil, NewArchiveErrorWithDetails(
			ErrArchiveInternalError,
			"Failed to list 7z archive contents",
			stderr,
			err,
		)
	}

	// Parse 7z l output
	// Format:
	//   Date      Time    Attr         Size   Compressed  Name
	//   ------------------- ----- ------------ ------------  ------------------------
	//   2024-01-01 10:00:00 ....A          100           50  file.txt
	lines := strings.Split(stdout, "\n")
	var contents []string
	inFileList := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Start of file list (after separator line)
		if strings.Contains(line, "---") {
			inFileList = true
			continue
		}

		if !inFileList {
			continue
		}

		// End of file list (another separator or empty)
		if strings.Contains(line, "---") || strings.HasPrefix(line, "Errors:") {
			break
		}

		// Parse file entry
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			// Last field is filename
			filename := strings.Join(fields[5:], " ")
			contents = append(contents, filename)
		}
	}

	return contents, nil
}
