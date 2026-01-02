package ui

import (
	"errors"
	"os"

	"github.com/sakura/duofm/internal/config"
	"github.com/sakura/duofm/internal/fs"
)

// File system helper functions

// fileExists checks if a file exists.
// Returns (true, nil) if file exists, (false, nil) if not found,
// or (false, err) for permission errors or other issues.
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// removeFile removes a file
func removeFile(path string) error {
	return os.Remove(path)
}

// removeAllFiles removes a file or directory recursively
func removeAllFiles(path string) error {
	return os.RemoveAll(path)
}

// isPermissionError checks if an error is a permission error
func isPermissionError(err error) bool {
	return os.IsPermission(err)
}

// deleteFile deletes a file or directory
func deleteFile(path string) error {
	return fs.Delete(path)
}

// copyFile copies a file or directory
func copyFile(src, dest string) error {
	return fs.Copy(src, dest)
}

// moveFile moves a file
func moveFile(src, dest string) error {
	return fs.MoveFile(src, dest)
}

// Config helper functions

// removeBookmark removes a bookmark from the list
func removeBookmark(bookmarks []config.Bookmark, index int) ([]config.Bookmark, error) {
	return config.RemoveBookmark(bookmarks, index)
}

// isPathBookmarked checks if a path is already bookmarked
func isPathBookmarked(bookmarks []config.Bookmark, path string) bool {
	return config.IsPathBookmarked(bookmarks, path)
}

// defaultAliasFromPath generates a default alias from a path
func defaultAliasFromPath(path string) string {
	return config.DefaultAliasFromPath(path)
}
