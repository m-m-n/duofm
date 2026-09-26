package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

// --- AC-1, AC-2, AC-3: header line 1 stays on one screen row ---

func TestRenderHeaderLine1_NoBranch_BoundedWidthNoWrapEllipsis(t *testing.T) {
	const width = 40
	const wantMaxWidth = 36 // width - 4

	longASCIIPath := "/home/user/very/long/path/to/some/deeply/nested/directory/that/is/wider/than/the/pane"
	longFullWidthPath := "/home/ユーザー/とても/長い/パス/を/持つ/深く/ネストされた/ディレクトリ"

	tests := []struct {
		name  string
		setup func() *Pane
	}{
		{
			name: "ASCII path wider than the pane",
			setup: func() *Pane {
				return &Pane{path: longASCIIPath, width: width, theme: DefaultTheme()}
			},
		},
		{
			name: "same path with hidden-files and filter indicators",
			setup: func() *Pane {
				return &Pane{
					path:          longASCIIPath,
					width:         width,
					theme:         DefaultTheme(),
					showHidden:    true,
					filterPattern: "abc",
					filterMode:    SearchModeIncremental,
				}
			},
		},
		{
			name: "path with full-width characters wider than the pane",
			setup: func() *Pane {
				return &Pane{path: longFullWidthPath, width: width, theme: DefaultTheme()}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane := tt.setup()
			result := pane.renderHeaderLine1()

			if w := runewidth.StringWidth(result); w > wantMaxWidth {
				t.Errorf("display width = %d, want <= %d (result: %q)", w, wantMaxWidth, result)
			}
			if strings.ContainsAny(result, "\n\r") {
				t.Errorf("result contains a line break: %q", result)
			}
			if !strings.HasSuffix(result, "...") {
				t.Errorf("result does not end with \"...\": %q", result)
			}
		})
	}
}

func TestRenderHeaderLine1_BranchOnly_BoundedWidthNoLineBreak(t *testing.T) {
	pane := &Pane{
		path:      "/home/user/project",
		gitBranch: "feature/a-very-long-branch-name-that-does-not-fit",
		width:     15,
		theme:     DefaultTheme(),
	}
	result := pane.renderHeaderLine1()

	const wantMaxWidth = 11 // width - 4
	if w := runewidth.StringWidth(result); w > wantMaxWidth {
		t.Errorf("display width = %d, want <= %d (result: %q)", w, wantMaxWidth, result)
	}
	if strings.ContainsAny(result, "\n\r") {
		t.Errorf("result contains a line break: %q", result)
	}
}

func TestRenderHeaderLine1_NarrowWidths_NoPanicZeroWidth(t *testing.T) {
	for width := 0; width <= 4; width++ {
		for _, branch := range []string{"", "main"} {
			t.Run(fmt.Sprintf("width=%d/branch=%q", width, branch), func(t *testing.T) {
				pane := &Pane{
					path:      "/home/user/some/long/path",
					gitBranch: branch,
					width:     width,
					theme:     DefaultTheme(),
				}

				var result string
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Fatalf("renderHeaderLine1 panicked: %v", r)
						}
					}()
					result = pane.renderHeaderLine1()
				}()

				if w := runewidth.StringWidth(result); w != 0 {
					t.Errorf("display width = %d, want 0 (result: %q)", w, result)
				}
			})
		}
	}
}

func TestTruncateStringWithEllipsis_ZeroAndNegativeWidths_NoPanic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
	}{
		{"zero width, non-empty input", "hello", 0},
		{"negative width, non-empty input", "hello", -1},
		{"very negative width", "hello", -100},
		{"empty input, zero width", "", 0},
		{"empty input, negative width", "", -5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("truncateStringWithEllipsis panicked: %v", r)
					}
				}()
				result = truncateStringWithEllipsis(tt.input, tt.maxWidth)
			}()
			if w := runewidth.StringWidth(result); w != 0 {
				t.Errorf("display width = %d, want 0 (result: %q)", w, result)
			}
		})
	}
}

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

// --- shared helpers for narrow-width panic regression tests (task0001) ---

// mustNotPanic runs fn, recovering any panic and failing the test with the
// given name if one occurs. Style matches TestRenderHeaderLine1_NarrowWidths_NoPanicZeroWidth.
func mustNotPanic(t *testing.T, name string, fn func() string) string {
	t.Helper()
	var result string
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("%s panicked: %v", name, r)
			}
		}()
		result = fn()
	}()
	return result
}

// splitRows splits a rendered view into its rows. Every row, including the
// last, ends with "\n" (task0001 row-structure invariant), so the final
// element of strings.Split is always the empty string after the trailing
// break; this drops it.
func splitRows(s string) []string {
	rows := strings.Split(s, "\n")
	if len(rows) > 0 && rows[len(rows)-1] == "" {
		rows = rows[:len(rows)-1]
	}
	return rows
}

// --- AC-1 (FR1, NFR1): header separator row glyph count ---

func TestPaneViews_HeaderSeparatorGlyphCount(t *testing.T) {
	const height = 24
	widths := []int{0, 1, 2, 3, 4, 40}

	wantCount := func(width int) int {
		want := width - 2
		if want < 0 {
			want = 0
		}
		return want
	}

	checkSeparator := func(t *testing.T, result string, want int) {
		t.Helper()
		rows := splitRows(result)
		if len(rows) < 3 {
			t.Fatalf("expected at least 3 rows, got %d: %q", len(rows), result)
		}
		got := strings.Count(rows[2], "─")
		if got != want {
			t.Errorf("separator glyph count = %d, want %d (row: %q)", got, want, rows[2])
		}
	}

	for _, width := range widths {
		want := wantCount(width)

		t.Run(fmt.Sprintf("View/width=%d", width), func(t *testing.T) {
			pane := newFilesPane(t, LeftPane, 5, width, height, true)
			result := mustNotPanic(t, "View", pane.View)
			checkSeparator(t, result, want)
		})

		t.Run(fmt.Sprintf("ViewWithBgOutput/width=%d", width), func(t *testing.T) {
			pane := newFilesPane(t, LeftPane, 5, width, height, true)
			pane.SetBgOutputActive(true)
			buf := NewOutputBuffer(10)
			buf.Append("output line one")
			result := mustNotPanic(t, "ViewWithBgOutput", func() string {
				return pane.ViewWithBgOutput(0, buf, "ls -la", false)
			})
			checkSeparator(t, result, want)
		})

		t.Run(fmt.Sprintf("ViewDimmedWithDiskSpace/width=%d", width), func(t *testing.T) {
			pane := newFilesPane(t, LeftPane, 5, width, height, true)
			result := mustNotPanic(t, "ViewDimmedWithDiskSpace", func() string {
				return pane.ViewDimmedWithDiskSpace(0)
			})
			checkSeparator(t, result, want)
		})
	}
}

// --- AC-2 (FR2, FR5): the four pane views render without panicking at widths 0-4 ---

func TestPaneViews_NarrowWidths_NoPanic(t *testing.T) {
	const height = 24

	for width := 0; width <= 4; width++ {
		t.Run(fmt.Sprintf("width=%d/View", width), func(t *testing.T) {
			pane := newFilesPane(t, LeftPane, 5, width, height, true)
			mustNotPanic(t, "View", pane.View)
		})

		for _, focused := range []bool{false, true} {
			t.Run(fmt.Sprintf("width=%d/ViewWithBgOutput/focused=%v", width, focused), func(t *testing.T) {
				pane := newFilesPane(t, LeftPane, 5, width, height, true)
				pane.SetBgOutputActive(true)
				buf := NewOutputBuffer(10)
				buf.Append("output line one")
				buf.Append("output line two")
				mustNotPanic(t, "ViewWithBgOutput", func() string {
					return pane.ViewWithBgOutput(0, buf, "ls -la", focused)
				})
			})
		}

		t.Run(fmt.Sprintf("width=%d/ViewDimmedWithDiskSpace", width), func(t *testing.T) {
			pane := newFilesPane(t, LeftPane, 5, width, height, true)
			mustNotPanic(t, "ViewDimmedWithDiskSpace", func() string {
				return pane.ViewDimmedWithDiskSpace(0)
			})
		})

		for _, input := range []string{"", "abc"} {
			t.Run(fmt.Sprintf("width=%d/ViewWithMinibuffer/input=%q", width, input), func(t *testing.T) {
				pane := newFilesPane(t, LeftPane, 5, width, height, true)
				mb := NewMinibuffer()
				mb.SetPrompt("/: ")
				mb.SetWidth(width)
				mb.SetInput(input)
				mb.Show()
				mustNotPanic(t, "ViewWithMinibuffer", func() string {
					return pane.ViewWithMinibuffer(0, mb)
				})
			})
		}
	}
}

// --- AC-3 (FR4): output row count at widths 0/1 matches the width-40 baseline ---

func TestPaneViews_NarrowWidths_RowCountMatchesWidth40(t *testing.T) {
	const height = 24
	const baselineWidth = 40

	for _, width := range []int{0, 1} {
		t.Run(fmt.Sprintf("width=%d/View", width), func(t *testing.T) {
			narrow := newFilesPane(t, LeftPane, 5, width, height, true)
			base := newFilesPane(t, LeftPane, 5, baselineWidth, height, true)

			gotRows := strings.Count(mustNotPanic(t, "View", narrow.View), "\n")
			wantRows := strings.Count(base.View(), "\n")
			if gotRows != wantRows {
				t.Errorf("row count = %d, want %d", gotRows, wantRows)
			}
		})

		t.Run(fmt.Sprintf("width=%d/ViewWithBgOutput", width), func(t *testing.T) {
			narrow := newFilesPane(t, LeftPane, 5, width, height, true)
			narrow.SetBgOutputActive(true)
			base := newFilesPane(t, LeftPane, 5, baselineWidth, height, true)
			base.SetBgOutputActive(true)
			buf := NewOutputBuffer(10)
			buf.Append("output line")

			gotRows := strings.Count(mustNotPanic(t, "ViewWithBgOutput", func() string {
				return narrow.ViewWithBgOutput(0, buf, "ls -la", false)
			}), "\n")
			wantRows := strings.Count(base.ViewWithBgOutput(0, buf, "ls -la", false), "\n")
			if gotRows != wantRows {
				t.Errorf("row count = %d, want %d", gotRows, wantRows)
			}
		})

		t.Run(fmt.Sprintf("width=%d/ViewDimmedWithDiskSpace", width), func(t *testing.T) {
			narrow := newFilesPane(t, LeftPane, 5, width, height, true)
			base := newFilesPane(t, LeftPane, 5, baselineWidth, height, true)

			gotRows := strings.Count(mustNotPanic(t, "ViewDimmedWithDiskSpace", func() string {
				return narrow.ViewDimmedWithDiskSpace(0)
			}), "\n")
			wantRows := strings.Count(base.ViewDimmedWithDiskSpace(0), "\n")
			if gotRows != wantRows {
				t.Errorf("row count = %d, want %d", gotRows, wantRows)
			}
		})

		t.Run(fmt.Sprintf("width=%d/ViewWithMinibuffer", width), func(t *testing.T) {
			narrow := newFilesPane(t, LeftPane, 5, width, height, true)
			narrowMB := NewMinibuffer()
			narrowMB.SetPrompt("/: ")
			narrowMB.SetWidth(width)
			narrowMB.Show()

			base := newFilesPane(t, LeftPane, 5, baselineWidth, height, true)
			baseMB := NewMinibuffer()
			baseMB.SetPrompt("/: ")
			baseMB.SetWidth(baselineWidth)
			baseMB.Show()

			gotRows := strings.Count(mustNotPanic(t, "ViewWithMinibuffer", func() string {
				return narrow.ViewWithMinibuffer(0, narrowMB)
			}), "\n")
			wantRows := strings.Count(base.ViewWithMinibuffer(0, baseMB), "\n")
			if gotRows != wantRows {
				t.Errorf("row count = %d, want %d", gotRows, wantRows)
			}
		})
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
