// Package archive provides archive creation and extraction functionality
// with support for multiple formats (tar, tar.gz, tar.bz2, tar.xz, zip, 7z).
//
// The package implements comprehensive security measures including:
//   - Path traversal prevention in archive entries
//   - Symlink validation to prevent escape attacks
//   - Compression bomb detection via ratio analysis
//   - TOCTOU (Time-of-check to time-of-use) protection with file hashing
//   - Disk space verification before operations
//
// Usage:
//
//	controller := archive.NewArchiveController()
//	taskID, err := controller.CreateArchive(sources, output, archive.FormatZip, 6)
//	if err != nil {
//	    // handle error
//	}
//	controller.WaitForTask(taskID)
//	status := controller.GetTaskStatus(taskID)
package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ArchiveController provides high-level archive operations with task management
type ArchiveController struct {
	taskManager      *TaskManager
	tarExecutor      *TarExecutor
	zipExecutor      *ZipExecutor
	sevenZipExecutor *SevenZipExecutor
	smartExtractor   *SmartExtractor
}

// NewArchiveController creates a new ArchiveController instance
func NewArchiveController() *ArchiveController {
	return &ArchiveController{
		taskManager:      NewTaskManager(),
		tarExecutor:      NewTarExecutor(),
		zipExecutor:      NewZipExecutor(),
		sevenZipExecutor: NewSevenZipExecutor(),
		smartExtractor:   NewSmartExtractor(),
	}
}

// CreateArchive initiates archive creation as a background task
func (ac *ArchiveController) CreateArchive(sources []string, output string, format ArchiveFormat, level int) (string, error) {
	// Validate inputs
	if len(sources) == 0 {
		return "", NewArchiveError(ErrArchiveSourceNotFound, "No source files specified", nil)
	}

	for _, src := range sources {
		if _, err := os.Stat(src); os.IsNotExist(err) {
			return "", NewArchiveError(ErrArchiveSourceNotFound, fmt.Sprintf("Source not found: %s", src), err)
		}
	}

	// Check if format is available
	if !IsFormatAvailable(format) {
		return "", NewArchiveError(ErrArchiveUnsupportedFormat, fmt.Sprintf("Format %s is not available", format.String()), nil)
	}

	// Start background task
	taskID := ac.taskManager.StartTask(fmt.Sprintf("compress-%s", format.String()), func(ctx context.Context, progress chan<- *ProgressUpdate) error {
		return ac.compress(ctx, sources, output, format, level, progress)
	})

	return taskID, nil
}

// compress performs the actual compression
func (ac *ArchiveController) compress(ctx context.Context, sources []string, output string, format ArchiveFormat, level int, progress chan<- *ProgressUpdate) error {
	// Validate compression level
	if err := ValidateCompressionLevel(level); err != nil {
		return err
	}

	// Validate sources
	if err := ValidateSources(sources); err != nil {
		return err
	}

	// Calculate total size for disk space check
	var totalSize int64
	for _, src := range sources {
		totalSize += calculateTotalSize(src)
	}

	// Get output directory for disk space check
	outputDir := filepath.Dir(output)
	insufficient, diskErr := CheckDiskSpace(outputDir, totalSize)
	if diskErr != nil {
		// Log warning but continue - disk space check is best-effort
		// Could not check disk space, proceed with caution
	} else if insufficient {
		return NewArchiveError(ErrArchiveDiskSpaceInsufficient, "Not enough disk space for archive creation", nil)
	}

	// Send initial progress
	if progress != nil {
		progress <- &ProgressUpdate{
			ProcessedFiles: 0,
			TotalFiles:     len(sources),
			ProcessedBytes: 0,
			TotalBytes:     totalSize,
			CurrentFile:    "",
			StartTime:      time.Now(),
			Operation:      "compress",
			ArchivePath:    output,
		}
	}

	// Perform compression based on format
	var err error
	switch format {
	case FormatTar, FormatTarGz, FormatTarBz2, FormatTarXz:
		err = ac.tarExecutor.Compress(ctx, format, sources, output, level, progress)
	case FormatZip:
		err = ac.zipExecutor.Compress(ctx, sources, output, level, progress)
	case Format7z:
		err = ac.sevenZipExecutor.Compress(ctx, sources, output, level, progress)
	default:
		err = NewArchiveError(ErrArchiveUnsupportedFormat, "Unsupported format", nil)
	}

	return err
}

// calculateTotalSize calculates the total size of a file or directory
// Ignores errors and symlinks to prevent loops and permission issues
func calculateTotalSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip files/directories with errors (permission denied, etc.)
			return nil
		}
		// Skip symlinks to prevent loops
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// ExtractArchive initiates archive extraction as a background task
func (ac *ArchiveController) ExtractArchive(archivePath string, destDir string) (string, error) {
	// Validate input
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return "", NewArchiveError(ErrArchiveSourceNotFound, fmt.Sprintf("Archive not found: %s", archivePath), err)
	}

	// Detect format
	format, err := DetectFormat(archivePath)
	if err != nil {
		return "", err
	}

	// Check if format is available
	if !IsFormatAvailable(format) {
		return "", NewArchiveError(ErrArchiveUnsupportedFormat, fmt.Sprintf("Format %s is not available", format.String()), nil)
	}

	// Start background task
	taskID := ac.taskManager.StartTask(fmt.Sprintf("extract-%s", format.String()), func(ctx context.Context, progress chan<- *ProgressUpdate) error {
		return ac.extract(ctx, archivePath, destDir, format, progress)
	})

	return taskID, nil
}

// extract performs the actual extraction
func (ac *ArchiveController) extract(ctx context.Context, archivePath string, destDir string, format ArchiveFormat, progress chan<- *ProgressUpdate) error {
	// TOCTOU Protection: Calculate hash before security checks
	archiveHash, err := CalculateFileHash(archivePath)
	if err != nil {
		return NewArchiveError(ErrArchiveIOError, "Failed to calculate archive hash", err)
	}

	// Get archive metadata for security checks (includes path traversal validation)
	metadata, err := ac.smartExtractor.GetArchiveMetadata(ctx, archivePath, format)
	if err != nil {
		return err
	}

	// Check for compression bomb (suspicious compression ratio)
	if CheckCompressionRatio(metadata.ArchiveSize, metadata.ExtractedSize) {
		return NewArchiveError(ErrArchiveCompressionBomb, "Suspicious compression ratio detected (potential zip bomb)", nil)
	}

	// Check disk space for extraction
	insufficient, diskErr := CheckDiskSpace(destDir, metadata.ExtractedSize)
	if diskErr != nil {
		// Log warning but continue - disk space check is best-effort
		// Could not check disk space, proceed with caution
	} else if insufficient {
		return NewArchiveError(ErrArchiveDiskSpaceInsufficient, "Not enough disk space for extraction", nil)
	}

	// Send initial progress
	if progress != nil {
		progress <- &ProgressUpdate{
			ProcessedFiles: 0,
			TotalFiles:     metadata.FileCount,
			ProcessedBytes: 0,
			TotalBytes:     metadata.ExtractedSize,
			CurrentFile:    "",
			StartTime:      time.Now(),
			Operation:      "extract",
			ArchivePath:    archivePath,
		}
	}

	// Analyze archive structure for smart extraction
	strategy, err := ac.smartExtractor.AnalyzeStructure(ctx, archivePath, format)
	if err != nil {
		return err
	}

	// Determine extraction directory
	extractDir := destDir
	if strategy.Method == ExtractToDirectory {
		// Create directory based on archive name
		archiveName := filepath.Base(archivePath)
		// Remove extension(s)
		dirName := strings.TrimSuffix(archiveName, filepath.Ext(archiveName))
		// For tar.gz, tar.bz2, tar.xz, remove .tar as well
		if strings.HasSuffix(dirName, ".tar") {
			dirName = strings.TrimSuffix(dirName, ".tar")
		}
		extractDir = filepath.Join(destDir, dirName)
	}

	// TOCTOU Protection: Verify hash hasn't changed before extraction
	if err := VerifyFileHash(archivePath, archiveHash); err != nil {
		return err
	}

	// Perform extraction based on format
	switch format {
	case FormatTar, FormatTarGz, FormatTarBz2, FormatTarXz:
		err = ac.tarExecutor.Extract(ctx, format, archivePath, extractDir, progress)
	case FormatZip:
		err = ac.zipExecutor.Extract(ctx, archivePath, extractDir, progress)
	case Format7z:
		err = ac.sevenZipExecutor.Extract(ctx, archivePath, extractDir, progress)
	default:
		err = NewArchiveError(ErrArchiveUnsupportedFormat, "Unsupported format", nil)
	}

	return err
}

// CancelTask cancels a running task
func (ac *ArchiveController) CancelTask(taskID string) error {
	return ac.taskManager.CancelTask(taskID)
}

// GetTaskStatus returns the status of a task
func (ac *ArchiveController) GetTaskStatus(taskID string) *TaskStatus {
	return ac.taskManager.GetTaskStatus(taskID)
}

// WaitForTask waits for a task to complete
func (ac *ArchiveController) WaitForTask(taskID string) {
	ac.taskManager.WaitForTask(taskID)
}

// GetArchiveMetadata retrieves metadata about an archive for security checks
func (ac *ArchiveController) GetArchiveMetadata(archivePath string) (*ArchiveMetadata, error) {
	// Detect format
	format, err := DetectFormat(archivePath)
	if err != nil {
		return nil, err
	}

	return ac.smartExtractor.GetArchiveMetadata(context.Background(), archivePath, format)
}
