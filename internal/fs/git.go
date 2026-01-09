package fs

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// GetGitBranch returns the current Git branch name for the given path.
// Returns empty string if not a Git repository, on error, or if Git is not installed.
// This function uses a 100ms timeout to ensure UI responsiveness (NFR1.1).
func GetGitBranch(path string) string {
	if path == "" {
		return ""
	}

	// Create context with 100ms timeout (NFR1.1: performance requirement)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Use exec.CommandContext to respect timeout
	// Arguments are passed separately to prevent command injection
	cmd := exec.CommandContext(ctx, "git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")

	// Capture output
	output, err := cmd.Output()
	if err != nil {
		// Return empty string on any error:
		// - Git not installed
		// - Not a Git repository
		// - Timeout exceeded
		// - Permission denied
		return ""
	}

	// Trim whitespace (output includes trailing newline)
	return strings.TrimSpace(string(output))
}
