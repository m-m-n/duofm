package ui

import (
	"path/filepath"
	"strings"
)

// hasEditableExtension determines if a file should use extension-preserving rename mode.
// Returns:
//   - baseName: the editable part of the filename
//   - extension: the fixed extension part (e.g., ".txt")
//   - hasExt: true if extension-preserving mode should be used
//
// Rules:
//   - Directories: always return full name, no extension
//   - Hidden files (starting with "."): apply rules to the part after the leading dot
//   - Regular files: use last dot as separator
//   - Files ending with a trailing dot are treated as extensionless
func hasEditableExtension(name string, isDir bool) (baseName, extension string, hasExt bool) {
	// Directories never have editable extensions
	if isDir {
		return name, "", false
	}

	// Check if it's a hidden file (starts with .)
	isHidden := strings.HasPrefix(name, ".")

	if isHidden {
		// For hidden files, remove the leading dot and check the remainder
		// .bashrc -> bashrc (no dot) -> no extension -> full edit
		// .config.json -> config.json (has dot) -> extension .json -> editable: .config
		// .foo.bar -> foo.bar (has dot) -> extension .bar -> editable: .foo

		// Handle special cases like "." or ".."
		if len(name) <= 1 {
			return name, "", false
		}

		nameWithoutLeadingDot := name[1:]

		// Handle cases like ".." or "..."
		if nameWithoutLeadingDot == "." || nameWithoutLeadingDot == ".." {
			return name, "", false
		}

		ext := filepath.Ext(nameWithoutLeadingDot)

		// No extension in the part after leading dot
		if ext == "" {
			return name, "", false
		}

		// ext equals the whole nameWithoutLeadingDot means it's like ".txt" where
		// nameWithoutLeadingDot = "txt" and ext = "" (already handled above)
		// But filepath.Ext(".txt") returns ".txt" for file ".txt", so we need to check
		// if ext is "."+nameWithoutLeadingDot meaning the whole thing is just an extension
		if "."+nameWithoutLeadingDot == ext || nameWithoutLeadingDot == ext[1:] {
			// The part after leading dot is itself an extension-like pattern
			// e.g., name=".txt", nameWithoutLeadingDot="txt", ext=""
			// Actually filepath.Ext("txt") = "", so this case won't hit
			return name, "", false
		}

		// Check for trailing dot (e.g., ".config.")
		// filepath.Ext("config.") returns "."
		if ext == "." {
			return name, "", false
		}

		// Has extension: construct editable base (with leading dot) and extension
		baseWithoutLeadingDot := strings.TrimSuffix(nameWithoutLeadingDot, ext)
		return "." + baseWithoutLeadingDot, ext, true
	}

	// Regular file - check for extension
	ext := filepath.Ext(name)

	// No extension
	if ext == "" {
		return name, "", false
	}

	// Name is just ".something" (hidden file starting with dot, but we're not in hidden branch)
	// This case is actually handled in the hidden branch above, but for safety:
	if ext == name {
		return name, "", false
	}

	// Trailing dot (e.g., "file.")
	// filepath.Ext("file.") returns "."
	if ext == "." {
		return name, "", false
	}

	base := strings.TrimSuffix(name, ext)
	return base, ext, true
}
