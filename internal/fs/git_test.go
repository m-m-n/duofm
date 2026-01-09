package fs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Helper function to check if git is available
func isGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// Helper function to create a temporary git repository
func createTempGitRepo(t *testing.T, branchName string) string {
	t.Helper()

	dir := t.TempDir()

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Set user config for commits (required for some operations)
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to set git config email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to set git config name: %v", err)
	}

	// Create an initial commit (required for branch operations)
	testFile := filepath.Join(dir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to git commit: %v", err)
	}

	// Create and checkout the specified branch (if not main/master)
	if branchName != "main" && branchName != "master" {
		cmd = exec.Command("git", "checkout", "-b", branchName)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create branch %s: %v", branchName, err)
		}
	}

	return dir
}

func TestGetGitBranch(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("git is not available")
	}

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) string // returns path to test
		expectedBranch string
		expectEmpty    bool
	}{
		{
			name: "git repository root with feature branch",
			setupFunc: func(t *testing.T) string {
				return createTempGitRepo(t, "feature/test-branch")
			},
			expectedBranch: "feature/test-branch",
			expectEmpty:    false,
		},
		{
			name: "git repository root with main branch",
			setupFunc: func(t *testing.T) string {
				return createTempGitRepo(t, "main")
			},
			expectedBranch: "main",
			expectEmpty:    false,
		},
		{
			name: "subdirectory of git repository",
			setupFunc: func(t *testing.T) string {
				dir := createTempGitRepo(t, "develop")
				subdir := filepath.Join(dir, "src", "pkg")
				if err := os.MkdirAll(subdir, 0755); err != nil {
					t.Fatalf("failed to create subdir: %v", err)
				}
				return subdir
			},
			expectedBranch: "develop",
			expectEmpty:    false,
		},
		{
			name: "non-git directory",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			expectedBranch: "",
			expectEmpty:    true,
		},
		{
			name: "non-existent directory",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/path/that/does/not/exist"
			},
			expectedBranch: "",
			expectEmpty:    true,
		},
		{
			name: "branch name with special characters",
			setupFunc: func(t *testing.T) string {
				return createTempGitRepo(t, "feature/JIRA-123-fix-bug")
			},
			expectedBranch: "feature/JIRA-123-fix-bug",
			expectEmpty:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupFunc(t)
			result := GetGitBranch(path)

			if tt.expectEmpty {
				if result != "" {
					t.Errorf("expected empty string, got %q", result)
				}
			} else {
				if result != tt.expectedBranch {
					t.Errorf("expected branch %q, got %q", tt.expectedBranch, result)
				}
			}
		})
	}
}

func TestGetGitBranch_DetachedHead(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("git is not available")
	}

	dir := createTempGitRepo(t, "main")

	// Get the current commit hash
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get commit hash: %v", err)
	}
	commitHash := string(output)

	// Checkout the commit directly to create detached HEAD state
	cmd = exec.Command("git", "checkout", commitHash[:8]) // Use short hash
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to checkout commit: %v", err)
	}

	result := GetGitBranch(dir)
	if result != "HEAD" {
		t.Errorf("expected 'HEAD' for detached HEAD, got %q", result)
	}
}

func TestGetGitBranch_EmptyPath(t *testing.T) {
	result := GetGitBranch("")
	if result != "" {
		t.Errorf("expected empty string for empty path, got %q", result)
	}
}

func TestGetGitBranch_DoesNotPanic(t *testing.T) {
	// This test ensures the function handles all edge cases gracefully
	// without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetGitBranch panicked: %v", r)
		}
	}()

	testPaths := []string{
		"",
		"/",
		"/tmp",
		"/nonexistent/path",
		"relative/path",
		".",
		"..",
	}

	for _, path := range testPaths {
		_ = GetGitBranch(path)
	}
}
