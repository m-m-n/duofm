package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TrashDir returns the path to the user's trash directory
// Uses XDG_DATA_HOME if set, otherwise defaults to ~/.local/share/Trash
func TrashDir() string {
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home is not available
			return ".local/share/Trash"
		}
		xdgDataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(xdgDataHome, "Trash")
}

// TrashFilesDir returns the path to the trash files directory
func TrashFilesDir() string {
	return filepath.Join(TrashDir(), "files")
}

// TrashInfoDir returns the path to the trash info directory
func TrashInfoDir() string {
	return filepath.Join(TrashDir(), "info")
}

// EnsureTrashDirs ensures the trash directories exist
// Creates ~/.local/share/Trash/files and ~/.local/share/Trash/info if they don't exist
func EnsureTrashDirs() error {
	filesDir := TrashFilesDir()
	infoDir := TrashInfoDir()

	if err := os.MkdirAll(filesDir, 0700); err != nil {
		return fmt.Errorf("failed to create trash files directory: %w", err)
	}

	if err := os.MkdirAll(infoDir, 0700); err != nil {
		return fmt.Errorf("failed to create trash info directory: %w", err)
	}

	return nil
}

// maxNameCollisionAttempts limits the number of attempts to find a unique name
// to prevent potential DoS from filesystem issues
const maxNameCollisionAttempts = 10000

// ResolveNameCollision returns a unique name for a file in the trash directory
// If the name already exists, appends .2, .3, etc. before the extension
// Returns the original name with a timestamp suffix if max attempts exceeded
func ResolveNameCollision(trashFilesDir, baseName string) string {
	targetPath := filepath.Join(trashFilesDir, baseName)

	// Check if name already exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return baseName
	}

	// Find the extension and base parts
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	// Handle dotfiles (files starting with .)
	if nameWithoutExt == "" && ext != "" {
		// This is a dotfile like ".config"
		nameWithoutExt = ext
		ext = ""
	}

	// Try incrementing counter until we find a unique name
	for counter := 2; counter <= maxNameCollisionAttempts; counter++ {
		newName := fmt.Sprintf("%s.%d%s", nameWithoutExt, counter, ext)
		targetPath = filepath.Join(trashFilesDir, newName)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			return newName
		}
	}

	// Fallback: use timestamp to guarantee uniqueness
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s.%d%s", nameWithoutExt, timestamp, ext)
}

// MoveToTrash moves a file or directory to the trash
// Creates a .trashinfo file with the original path and deletion date
func MoveToTrash(path string) error {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if file exists (use Lstat to handle symlinks)
	if _, err := os.Lstat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("source not found: %s", absPath)
	}

	// Ensure trash directories exist
	if err := EnsureTrashDirs(); err != nil {
		return fmt.Errorf("failed to ensure trash directories: %w", err)
	}

	// Get the base name and resolve collisions
	baseName := filepath.Base(absPath)
	trashName := ResolveNameCollision(TrashFilesDir(), baseName)

	// Paths in trash
	trashPath := filepath.Join(TrashFilesDir(), trashName)
	trashinfoPath := filepath.Join(TrashInfoDir(), trashName+".trashinfo")

	// Create .trashinfo file first (this is atomic on most filesystems)
	deletionTime := time.Now()
	if err := WriteTrashinfo(trashinfoPath, absPath, deletionTime); err != nil {
		return fmt.Errorf("failed to create trashinfo: %w", err)
	}

	// Move the file to trash
	// First try rename (works for same filesystem)
	err = os.Rename(absPath, trashPath)
	if err != nil {
		// If rename fails (e.g., cross-filesystem), fall back to copy+delete
		// But first, check if this is a symlink
		info, statErr := os.Lstat(absPath)
		if statErr != nil {
			// Cleanup trashinfo on failure
			os.Remove(trashinfoPath)
			return fmt.Errorf("failed to stat source: %w", statErr)
		}

		if info.Mode()&os.ModeSymlink != 0 {
			// For symlinks, read the target and recreate in trash
			target, readErr := os.Readlink(absPath)
			if readErr != nil {
				os.Remove(trashinfoPath)
				return fmt.Errorf("failed to read symlink: %w", readErr)
			}

			if symlinkErr := os.Symlink(target, trashPath); symlinkErr != nil {
				os.Remove(trashinfoPath)
				return fmt.Errorf("failed to create symlink in trash: %w", symlinkErr)
			}

			// Remove original symlink
			if removeErr := os.Remove(absPath); removeErr != nil {
				os.Remove(trashPath)
				os.Remove(trashinfoPath)
				return fmt.Errorf("failed to remove original symlink: %w", removeErr)
			}
		} else {
			// For regular files/directories, use copy+delete
			if copyErr := Copy(absPath, trashPath); copyErr != nil {
				os.Remove(trashinfoPath)
				return fmt.Errorf("failed to copy to trash: %w", copyErr)
			}

			if deleteErr := Delete(absPath); deleteErr != nil {
				// Try to clean up the copy
				os.RemoveAll(trashPath)
				os.Remove(trashinfoPath)
				return fmt.Errorf("failed to delete original after copy: %w", deleteErr)
			}
		}
	}

	return nil
}

// IsInTrash checks if a path is inside the trash files directory
// Uses filepath.Rel to safely check path containment, avoiding symlink traversal issues
func IsInTrash(path string) bool {
	// Clean and normalize paths
	cleanPath := filepath.Clean(path)
	trashFilesDir := filepath.Clean(TrashFilesDir())

	// Check if path equals the trash directory
	if cleanPath == trashFilesDir {
		return true
	}

	// Use filepath.Rel to safely check if path is inside trash
	// This handles edge cases like "../" in paths
	rel, err := filepath.Rel(trashFilesDir, cleanPath)
	if err != nil {
		return false
	}

	// If the relative path starts with "..", it's outside the trash directory
	return !strings.HasPrefix(rel, "..") && rel != ".."
}

// GetTrashItemInfo returns the trashinfo for an item in the trash
func GetTrashItemInfo(trashName string) (*TrashInfo, error) {
	trashinfoPath := filepath.Join(TrashInfoDir(), trashName+".trashinfo")
	return ParseTrashinfo(trashinfoPath)
}

// RestoreFromTrash restores a file from trash to its original location
func RestoreFromTrash(trashName string) error {
	// Get trashinfo
	trashinfoPath := filepath.Join(TrashInfoDir(), trashName+".trashinfo")
	info, err := ParseTrashinfo(trashinfoPath)
	if err != nil {
		return fmt.Errorf("failed to read trashinfo: %w", err)
	}

	// Source path in trash
	trashFilePath := filepath.Join(TrashFilesDir(), trashName)

	// Check if trash file exists
	if _, err := os.Lstat(trashFilePath); os.IsNotExist(err) {
		return fmt.Errorf("trash file not found: %s", trashName)
	}

	// Ensure parent directory of original path exists
	originalDir := filepath.Dir(info.OriginalPath)
	if err := os.MkdirAll(originalDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Move file back to original location
	err = os.Rename(trashFilePath, info.OriginalPath)
	if err != nil {
		// If rename fails (e.g., cross-filesystem), fall back to copy+delete
		fileInfo, statErr := os.Lstat(trashFilePath)
		if statErr != nil {
			return fmt.Errorf("failed to stat trash file: %w", statErr)
		}

		if fileInfo.Mode()&os.ModeSymlink != 0 {
			// Handle symlinks
			target, readErr := os.Readlink(trashFilePath)
			if readErr != nil {
				return fmt.Errorf("failed to read symlink: %w", readErr)
			}
			if symlinkErr := os.Symlink(target, info.OriginalPath); symlinkErr != nil {
				return fmt.Errorf("failed to create symlink at destination: %w", symlinkErr)
			}
			if removeErr := os.Remove(trashFilePath); removeErr != nil {
				os.Remove(info.OriginalPath)
				return fmt.Errorf("failed to remove trash symlink: %w", removeErr)
			}
		} else {
			// Copy and delete
			if copyErr := Copy(trashFilePath, info.OriginalPath); copyErr != nil {
				return fmt.Errorf("failed to copy from trash: %w", copyErr)
			}
			if deleteErr := Delete(trashFilePath); deleteErr != nil {
				os.RemoveAll(info.OriginalPath)
				return fmt.Errorf("failed to delete trash file: %w", deleteErr)
			}
		}
	}

	// Remove trashinfo file
	// Ignore error - the orphaned trashinfo is not critical as the restore was successful
	_ = os.Remove(trashinfoPath)

	return nil
}

// EmptyTrash permanently deletes all files in the trash
func EmptyTrash() error {
	filesDir := TrashFilesDir()
	infoDir := TrashInfoDir()

	// Ensure directories exist
	if err := EnsureTrashDirs(); err != nil {
		return err
	}

	// Delete all files in trash/files/
	filesEntries, err := os.ReadDir(filesDir)
	if err != nil {
		return fmt.Errorf("failed to read trash files directory: %w", err)
	}

	var lastErr error
	for _, entry := range filesEntries {
		path := filepath.Join(filesDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			lastErr = err
			// Continue deleting other files
		}
	}

	// Delete all files in trash/info/
	infoEntries, err := os.ReadDir(infoDir)
	if err != nil {
		return fmt.Errorf("failed to read trash info directory: %w", err)
	}

	for _, entry := range infoEntries {
		path := filepath.Join(infoDir, entry.Name())
		if err := os.Remove(path); err != nil {
			lastErr = err
			// Continue deleting other files
		}
	}

	return lastErr
}
