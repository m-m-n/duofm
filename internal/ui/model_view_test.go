package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/fs"
)

func TestEntryTypeLabel(t *testing.T) {
	theme := DefaultTheme()
	m := Model{theme: theme}

	tests := []struct {
		name      string
		entry     *fs.FileEntry
		wantLabel string
		wantFg    lipgloss.Color
	}{
		{
			name:      "directory entry",
			entry:     &fs.FileEntry{Name: "docs", IsDir: true},
			wantLabel: "Directory",
			wantFg:    theme.DirectoryFg,
		},
		{
			name:      "symlink entry",
			entry:     &fs.FileEntry{Name: "link", IsSymlink: true, LinkTarget: "/tmp"},
			wantLabel: "SymbolicLink",
			wantFg:    theme.SymlinkFg,
		},
		{
			name:      "html file",
			entry:     &fs.FileEntry{Name: "index.html"},
			wantLabel: "text/html",
			wantFg:    theme.StatusFg,
		},
		{
			name:      "txt file",
			entry:     &fs.FileEntry{Name: "readme.txt"},
			wantLabel: "text/plain",
			wantFg:    theme.StatusFg,
		},
		{
			name:      "no extension file",
			entry:     &fs.FileEntry{Name: "Makefile"},
			wantLabel: "application/octet-stream",
			wantFg:    theme.StatusFg,
		},
		{
			name:      "parent directory",
			entry:     &fs.FileEntry{Name: "..", IsDir: true},
			wantLabel: "Directory",
			wantFg:    theme.DirectoryFg,
		},
		{
			name:      "symlink to directory (IsSymlink takes priority over IsDir)",
			entry:     &fs.FileEntry{Name: "link-to-dir", IsDir: true, IsSymlink: true},
			wantLabel: "SymbolicLink",
			wantFg:    theme.SymlinkFg,
		},
		{
			name:      "nil entry",
			entry:     nil,
			wantLabel: "",
			wantFg:    theme.StatusFg,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label, fg := m.entryTypeLabel(tt.entry)
			if label != tt.wantLabel {
				t.Errorf("entryTypeLabel() label = %q, want %q", label, tt.wantLabel)
			}
			if fg != tt.wantFg {
				t.Errorf("entryTypeLabel() fg = %v, want %v", fg, tt.wantFg)
			}
		})
	}
}

func TestRenderStatusBar_WithMIMEType(t *testing.T) {
	theme := DefaultTheme()

	tests := []struct {
		name           string
		entries        []fs.FileEntry
		cursor         int
		width          int
		statusMessage  string
		expectContains []string
		expectMissing  []string
	}{
		{
			name:           "regular file shows MIME type",
			entries:        []fs.FileEntry{{Name: "index.html"}},
			cursor:         0,
			width:          80,
			expectContains: []string{"1/1", "[text/html]"},
		},
		{
			name:           "directory shows [Directory]",
			entries:        []fs.FileEntry{{Name: "docs", IsDir: true}},
			cursor:         0,
			width:          80,
			expectContains: []string{"1/1", "[Directory]"},
		},
		{
			name:           "symlink shows [SymbolicLink]",
			entries:        []fs.FileEntry{{Name: "link", IsSymlink: true}},
			cursor:         0,
			width:          80,
			expectContains: []string{"1/1", "[SymbolicLink]"},
		},
		{
			name:          "empty directory - no MIME type",
			entries:       []fs.FileEntry{},
			cursor:        0,
			width:         80,
			expectMissing: []string{"[Directory]", "[SymbolicLink]", "[text/"},
		},
		{
			name:           "narrow width omits MIME type",
			entries:        []fs.FileEntry{{Name: "index.html"}},
			cursor:         0,
			width:          30,
			expectContains: []string{"1/1"},
			expectMissing:  []string{"[text/html]"},
		},
		{
			name:           "status message hides MIME type",
			entries:        []fs.FileEntry{{Name: "index.html"}},
			cursor:         0,
			width:          80,
			statusMessage:  "File copied",
			expectContains: []string{"File copied"},
			expectMissing:  []string{"[text/html]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane := &Pane{
				entries: tt.entries,
				cursor:  tt.cursor,
				width:   tt.width / 2,
				theme:   theme,
			}

			m := Model{
				leftPane:      pane,
				rightPane:     &Pane{theme: theme},
				activePane:    LeftPane,
				theme:         theme,
				width:         tt.width,
				statusMessage: tt.statusMessage,
			}

			result := m.renderStatusBar()

			// Strip ANSI escape sequences for content checking
			stripped := stripANSI(result)

			for _, s := range tt.expectContains {
				if !strings.Contains(stripped, s) {
					t.Errorf("renderStatusBar() should contain %q, got %q", s, stripped)
				}
			}

			for _, s := range tt.expectMissing {
				if strings.Contains(stripped, s) {
					t.Errorf("renderStatusBar() should NOT contain %q, got %q", s, stripped)
				}
			}
		})
	}
}

// stripANSI removes ANSI escape sequences from a string for test assertions.
func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}
