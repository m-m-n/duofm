package ui

import (
	"strings"
	"testing"
)

func TestRenderHeaderLine1(t *testing.T) {
	tests := []struct {
		name              string
		setupPane         func() *Pane
		expectContains    []string
		expectNotContains []string
	}{
		{
			name: "with branch - sufficient space",
			setupPane: func() *Pane {
				p := &Pane{
					path:      "/home/user/project",
					gitBranch: "main",
					width:     80,
					theme:     DefaultTheme(),
				}
				return p
			},
			expectContains: []string{"/home/user/project", "[main]"},
		},
		{
			name: "without branch - non-git directory",
			setupPane: func() *Pane {
				p := &Pane{
					path:      "/tmp",
					gitBranch: "",
					width:     80,
					theme:     DefaultTheme(),
				}
				return p
			},
			expectContains:    []string{"/tmp"},
			expectNotContains: []string{"[", "]"},
		},
		{
			name: "with branch and hidden indicator",
			setupPane: func() *Pane {
				p := &Pane{
					path:       "/home/user/project",
					gitBranch:  "feature/test",
					showHidden: true,
					width:      80,
					theme:      DefaultTheme(),
				}
				return p
			},
			expectContains: []string{"[H]", "[feature/test]"},
		},
		{
			name: "long branch name",
			setupPane: func() *Pane {
				p := &Pane{
					path:      "/home/user/project",
					gitBranch: "feature/JIRA-12345-very-long-description",
					width:     80,
					theme:     DefaultTheme(),
				}
				return p
			},
			expectContains: []string{"[feature/JIRA-12345-very-long-description]"},
		},
		{
			name: "branch with special characters",
			setupPane: func() *Pane {
				p := &Pane{
					path:      "/home/user/project",
					gitBranch: "feature/test[bracket]",
					width:     80,
					theme:     DefaultTheme(),
				}
				return p
			},
			expectContains: []string{"[feature/test[bracket]]"},
		},
		{
			name: "detached HEAD state",
			setupPane: func() *Pane {
				p := &Pane{
					path:      "/home/user/project",
					gitBranch: "HEAD",
					width:     80,
					theme:     DefaultTheme(),
				}
				return p
			},
			expectContains: []string{"[HEAD]"},
		},
		{
			name: "long path truncation with branch",
			setupPane: func() *Pane {
				p := &Pane{
					path:      "/home/user/very/long/path/to/some/deep/nested/directory",
					gitBranch: "main",
					width:     50, // Narrow width to force truncation
					theme:     DefaultTheme(),
				}
				return p
			},
			expectContains:    []string{"...", "[main]"},
			expectNotContains: []string{"nested/directory"}, // Path should be truncated
		},
		{
			name: "very narrow width - branch only",
			setupPane: func() *Pane {
				p := &Pane{
					path:      "/home/user/project",
					gitBranch: "main",
					width:     15, // Very narrow
					theme:     DefaultTheme(),
				}
				return p
			},
			expectContains: []string{"[main]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane := tt.setupPane()
			result := pane.renderHeaderLine1()

			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected header to contain %q, got %q", expected, result)
				}
			}

			for _, notExpected := range tt.expectNotContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("expected header NOT to contain %q, got %q", notExpected, result)
				}
			}
		})
	}
}

func TestTruncateStringWithEllipsis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		expected string
	}{
		{
			name:     "no truncation needed",
			input:    "short",
			maxWidth: 10,
			expected: "short",
		},
		{
			name:     "exact fit",
			input:    "exact",
			maxWidth: 5,
			expected: "exact",
		},
		{
			name:     "needs truncation",
			input:    "this is a long string",
			maxWidth: 10,
			expected: "this is...",
		},
		{
			name:     "very short maxWidth",
			input:    "test",
			maxWidth: 3,
			expected: "...",
		},
		{
			name:     "maxWidth of 4",
			input:    "testing",
			maxWidth: 4,
			expected: "t...",
		},
		{
			name:     "unicode characters",
			input:    "hello world",
			maxWidth: 8,
			expected: "hello...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateStringWithEllipsis(tt.input, tt.maxWidth)

			// Validate expected output
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRenderHeaderLine1_BranchAlignment(t *testing.T) {
	// Test that branch is right-aligned
	pane := &Pane{
		path:      "/home/user",
		gitBranch: "main",
		width:     40,
		theme:     DefaultTheme(),
	}

	result := pane.renderHeaderLine1()

	// Branch should be at the end
	if !strings.HasSuffix(result, "[main]") {
		t.Errorf("expected branch to be right-aligned (end with [main]), got %q", result)
	}

	// There should be spaces between path and branch
	pathIndex := strings.Index(result, "/home/user")
	branchIndex := strings.Index(result, "[main]")

	if pathIndex == -1 || branchIndex == -1 {
		t.Fatalf("path or branch not found in result: %q", result)
	}

	// Branch should come after path
	if branchIndex <= pathIndex+len("/home/user") {
		t.Errorf("expected branch to come after path with padding, path at %d, branch at %d", pathIndex, branchIndex)
	}
}

func TestRenderHeaderLine2_SortInfo(t *testing.T) {
	tests := []struct {
		name         string
		sortConfig   SortConfig
		expectedSort string
	}{
		{
			name:         "default sort - Name ascending",
			sortConfig:   SortConfig{Field: SortByName, Order: SortAsc},
			expectedSort: "Name ↑",
		},
		{
			name:         "Name descending",
			sortConfig:   SortConfig{Field: SortByName, Order: SortDesc},
			expectedSort: "Name ↓",
		},
		{
			name:         "Size ascending",
			sortConfig:   SortConfig{Field: SortBySize, Order: SortAsc},
			expectedSort: "Size ↑",
		},
		{
			name:         "Size descending",
			sortConfig:   SortConfig{Field: SortBySize, Order: SortDesc},
			expectedSort: "Size ↓",
		},
		{
			name:         "Date ascending",
			sortConfig:   SortConfig{Field: SortByDate, Order: SortAsc},
			expectedSort: "Date ↑",
		},
		{
			name:         "Date descending",
			sortConfig:   SortConfig{Field: SortByDate, Order: SortDesc},
			expectedSort: "Date ↓",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane := &Pane{
				path:       "/home/user/test",
				width:      80,
				theme:      DefaultTheme(),
				sortConfig: tt.sortConfig,
			}

			result := pane.renderHeaderLine2(50 * 1024 * 1024 * 1024) // 50 GB

			if !strings.Contains(result, tt.expectedSort) {
				t.Errorf("expected header to contain sort info %q, got %q", tt.expectedSort, result)
			}

			// Also verify mark info and free space are present
			if !strings.Contains(result, "Marked") {
				t.Errorf("expected header to contain 'Marked', got %q", result)
			}
			if !strings.Contains(result, "Free") {
				t.Errorf("expected header to contain 'Free', got %q", result)
			}
		})
	}
}

func TestRenderHeaderLine2_SortInfoLayout(t *testing.T) {
	pane := &Pane{
		path:       "/home/user/test",
		width:      80,
		theme:      DefaultTheme(),
		sortConfig: DefaultSortConfig(),
	}

	result := pane.renderHeaderLine2(50 * 1024 * 1024 * 1024)

	// Verify order: Marked ... Name ↑ ... Free
	markedIdx := strings.Index(result, "Marked")
	sortIdx := strings.Index(result, "Name ↑")
	freeIdx := strings.Index(result, "Free")

	if markedIdx == -1 || sortIdx == -1 || freeIdx == -1 {
		t.Fatalf("missing components in header: %q", result)
	}

	if markedIdx >= sortIdx {
		t.Errorf("expected Marked before sort info, got marked=%d sort=%d", markedIdx, sortIdx)
	}
	if sortIdx >= freeIdx {
		t.Errorf("expected sort info before Free, got sort=%d free=%d", sortIdx, freeIdx)
	}
}

func TestRenderHeaderLine2_NarrowWidth(t *testing.T) {
	pane := &Pane{
		path:       "/home/user/test",
		width:      30,
		theme:      DefaultTheme(),
		sortConfig: DefaultSortConfig(),
	}

	result := pane.renderHeaderLine2(50 * 1024 * 1024 * 1024)

	// Sort info should still be visible even with narrow width
	if !strings.Contains(result, "Name ↑") {
		t.Errorf("expected sort info to be visible even at narrow width, got %q", result)
	}
}

func TestRenderHeaderLine1_NoBranchWhenEmpty(t *testing.T) {
	pane := &Pane{
		path:      "/tmp/test",
		gitBranch: "",
		width:     80,
		theme:     DefaultTheme(),
	}

	result := pane.renderHeaderLine1()

	// Should only contain path, no brackets at all
	if strings.Contains(result, "[") || strings.Contains(result, "]") {
		t.Errorf("expected no brackets when gitBranch is empty, got %q", result)
	}

	if !strings.Contains(result, "/tmp/test") {
		t.Errorf("expected path in result, got %q", result)
	}
}
