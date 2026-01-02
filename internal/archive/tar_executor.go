package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TarExecutor handles tar-based archive operations
type TarExecutor struct {
	executor *CommandExecutor
}

// NewTarExecutor creates a new TarExecutor instance
func NewTarExecutor() *TarExecutor {
	return &TarExecutor{
		executor: NewCommandExecutor(),
	}
}

// BuildCompressArgs builds command line arguments for tar compression
func (e *TarExecutor) BuildCompressArgs(format ArchiveFormat, sources []string, output string, level int) []string {
	var flags string

	switch format {
	case FormatTar:
		flags = "-cvf"
	case FormatTarGz:
		flags = "-czvf"
	case FormatTarBz2:
		flags = "-cjvf"
	case FormatTarXz:
		flags = "-cJvf"
	default:
		flags = "-cvf"
	}

	args := []string{flags, output}
	args = append(args, sources...)

	return args
}

// BuildExtractArgs builds command line arguments for tar extraction
func (e *TarExecutor) BuildExtractArgs(format ArchiveFormat, archivePath string, destDir string) []string {
	var flags string

	switch format {
	case FormatTar:
		flags = "-xvf"
	case FormatTarGz:
		flags = "-xzvf"
	case FormatTarBz2:
		flags = "-xjvf"
	case FormatTarXz:
		flags = "-xJvf"
	default:
		flags = "-xvf"
	}

	// Use --no-same-permissions to apply umask (NFR2.4), --no-same-owner to prevent
	// setuid/setgid bit preservation issues, -C for destination directory
	// Archive path is sanitized via sanitizePathForCommand in Extract()
	return []string{flags, archivePath, "--no-same-permissions", "--no-same-owner", "-C", destDir}
}

// ParseTarOutput parses a line of tar verbose output and returns the filename.
// tar -v output format: each line is a filename (possibly with path)
// Returns the filename or empty string if parsing fails.
func ParseTarOutput(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	// tar verbose output is simply the filename per line
	return line
}

// Compress creates a tar archive
func (e *TarExecutor) Compress(ctx context.Context, format ArchiveFormat, sources []string, output string, level int, progressChan chan<- *ProgressUpdate) error {
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

	// Convert sources to absolute paths and validate
	absSources := make([]string, len(sources))
	for i, src := range sources {
		abs, err := filepath.Abs(src)
		if err != nil {
			return NewArchiveError(ErrArchiveInternalError, fmt.Sprintf("Failed to get absolute path: %s", src), err)
		}
		absSources[i] = abs
	}

	// Get base directory for -C option (avoid os.Chdir for race safety)
	baseDir := filepath.Dir(absSources[0])

	// Convert sources to relative paths from baseDir
	relSources := make([]string, len(absSources))
	for i, abs := range absSources {
		rel, err := filepath.Rel(baseDir, abs)
		if err != nil {
			return NewArchiveError(ErrArchiveInternalError, fmt.Sprintf("Cannot compute relative path for: %s", abs), err)
		}
		relSources[i] = rel
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
	// Use os.CreateTemp to avoid collision with concurrent operations
	tempFile, err := os.CreateTemp(outputDir, ".duofm-archive-*.tmp")
	if err != nil {
		return NewArchiveError(ErrArchiveIOError, "Failed to create temporary file", err)
	}
	tempOutput := tempFile.Name()
	tempFile.Close() // Close immediately, tar will write to it

	// Build args with -C option to avoid os.Chdir
	args := e.buildCompressArgsWithDir(format, relSources, tempOutput, baseDir)

	// Use streaming progress if channel is provided
	if progressChan != nil {
		processedFiles := 0
		lineHandler := func(line string) error {
			filename := ParseTarOutput(line)
			if filename != "" {
				processedFiles++
				progressChan <- &ProgressUpdate{
					ProcessedFiles: processedFiles,
					TotalFiles:     totalFiles,
					ProcessedBytes: 0, // tar doesn't provide byte progress per file
					TotalBytes:     totalBytes,
					CurrentFile:    filename,
					Operation:      "compress",
					ArchivePath:    absOutput,
				}
			}
			return nil
		}

		stderr, err := e.executor.ExecuteCommandWithProgress(ctx, "", lineHandler, "tar", args...)
		if err != nil {
			os.Remove(tempOutput)
			return NewArchiveErrorWithDetails(
				ErrArchiveInternalError,
				"Failed to create archive",
				stderr,
				err,
			)
		}
	} else {
		// Fallback to non-streaming execution
		_, stderr, err := e.executor.ExecuteCommand(ctx, "tar", args...)
		if err != nil {
			os.Remove(tempOutput)
			return NewArchiveErrorWithDetails(
				ErrArchiveInternalError,
				"Failed to create archive",
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

// buildCompressArgsWithDir builds command line arguments for tar compression with -C option
func (e *TarExecutor) buildCompressArgsWithDir(format ArchiveFormat, sources []string, output string, baseDir string) []string {
	var flags string

	switch format {
	case FormatTar:
		flags = "-cvf"
	case FormatTarGz:
		flags = "-czvf"
	case FormatTarBz2:
		flags = "-cjvf"
	case FormatTarXz:
		flags = "-cJvf"
	default:
		flags = "-cvf"
	}

	// Use -- to separate options from filenames to prevent option injection
	args := []string{flags, output, "-C", baseDir, "--"}
	// Sanitize filenames that start with - to prevent option injection
	for _, src := range sources {
		args = append(args, sanitizePathForCommand(src))
	}

	return args
}

// sanitizePathForCommand is a local alias for SanitizePathForCommand
func sanitizePathForCommand(path string) string {
	return SanitizePathForCommand(path)
}

// calculateSize calculates the total size of a file or directory
func (e *TarExecutor) calculateSize(path string) int64 {
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
func (e *TarExecutor) countFilesAndSize(path string) (int, int64) {
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

// Extract extracts a tar archive
func (e *TarExecutor) Extract(ctx context.Context, format ArchiveFormat, archivePath string, destDir string, progressChan chan<- *ProgressUpdate) error {
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
	totalFiles := e.countArchiveFiles(ctx, format, absArchivePath)
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

	args := e.BuildExtractArgs(format, absArchivePath, absDestDir)

	// Use streaming progress if channel is provided
	if progressChan != nil {
		processedFiles := 0
		lineHandler := func(line string) error {
			filename := ParseTarOutput(line)
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

		stderr, err := e.executor.ExecuteCommandWithProgress(ctx, "", lineHandler, "tar", args...)
		if err != nil {
			return NewArchiveErrorWithDetails(
				ErrArchiveInternalError,
				"Failed to extract archive",
				stderr,
				err,
			)
		}
	} else {
		// Fallback to non-streaming execution
		_, stderr, err := e.executor.ExecuteCommand(ctx, "tar", args...)
		if err != nil {
			return NewArchiveErrorWithDetails(
				ErrArchiveInternalError,
				"Failed to extract archive",
				stderr,
				err,
			)
		}
	}

	// Validate symlinks after extraction (security check for path traversal)
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

// countArchiveFiles counts the number of files in a tar archive
func (e *TarExecutor) countArchiveFiles(ctx context.Context, format ArchiveFormat, archivePath string) int {
	contents, err := e.ListContents(ctx, format, archivePath)
	if err != nil {
		return 0
	}
	return len(contents)
}

// ListContents lists the contents of a tar archive
func (e *TarExecutor) ListContents(ctx context.Context, format ArchiveFormat, archivePath string) ([]string, error) {
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

	stdout, stderr, err := e.executor.ExecuteCommand(ctx, "tar", flags, archivePath)
	if err != nil {
		return nil, NewArchiveErrorWithDetails(
			ErrArchiveInternalError,
			"Failed to list archive contents",
			stderr,
			err,
		)
	}

	// Parse tar output - format: permissions user group size date time filename
	lines := strings.Split(stdout, "\n")
	var contents []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			// Last field is the filename
			filename := fields[len(fields)-1]
			contents = append(contents, filename)
		}
	}

	return contents, nil
}
