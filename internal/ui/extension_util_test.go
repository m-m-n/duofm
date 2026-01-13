package ui

import (
	"testing"
)

func TestHasEditableExtension(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		isDir         bool
		wantBaseName  string
		wantExtension string
		wantHasExt    bool
	}{
		// Regular files with extensions
		{
			name:          "regular file with simple extension",
			filename:      "document.txt",
			isDir:         false,
			wantBaseName:  "document",
			wantExtension: ".txt",
			wantHasExt:    true,
		},
		{
			name:          "file with tar.gz extension",
			filename:      "archive.tar.gz",
			isDir:         false,
			wantBaseName:  "archive.tar",
			wantExtension: ".gz",
			wantHasExt:    true,
		},
		{
			name:          "file with multiple dots",
			filename:      "my.file.name.pdf",
			isDir:         false,
			wantBaseName:  "my.file.name",
			wantExtension: ".pdf",
			wantHasExt:    true,
		},

		// Extensionless files
		{
			name:          "Makefile (no extension)",
			filename:      "Makefile",
			isDir:         false,
			wantBaseName:  "Makefile",
			wantExtension: "",
			wantHasExt:    false,
		},
		{
			name:          "LICENSE (no extension)",
			filename:      "LICENSE",
			isDir:         false,
			wantBaseName:  "LICENSE",
			wantExtension: "",
			wantHasExt:    false,
		},
		{
			name:          "README (no extension)",
			filename:      "README",
			isDir:         false,
			wantBaseName:  "README",
			wantExtension: "",
			wantHasExt:    false,
		},

		// Hidden files without extension
		{
			name:          "hidden file without extension (.bashrc)",
			filename:      ".bashrc",
			isDir:         false,
			wantBaseName:  ".bashrc",
			wantExtension: "",
			wantHasExt:    false,
		},
		{
			name:          "hidden file without extension (.gitignore)",
			filename:      ".gitignore",
			isDir:         false,
			wantBaseName:  ".gitignore",
			wantExtension: "",
			wantHasExt:    false,
		},
		{
			name:          "hidden file without extension (.profile)",
			filename:      ".profile",
			isDir:         false,
			wantBaseName:  ".profile",
			wantExtension: "",
			wantHasExt:    false,
		},

		// Hidden files with extension
		{
			name:          "hidden file with extension (.config.json)",
			filename:      ".config.json",
			isDir:         false,
			wantBaseName:  ".config",
			wantExtension: ".json",
			wantHasExt:    true,
		},
		{
			name:          "hidden file with extension (.env.local)",
			filename:      ".env.local",
			isDir:         false,
			wantBaseName:  ".env",
			wantExtension: ".local",
			wantHasExt:    true,
		},
		{
			name:          "hidden file with extension (.foo.bar)",
			filename:      ".foo.bar",
			isDir:         false,
			wantBaseName:  ".foo",
			wantExtension: ".bar",
			wantHasExt:    true,
		},
		{
			name:          "hidden file with double extension (.config.toml.bak)",
			filename:      ".config.toml.bak",
			isDir:         false,
			wantBaseName:  ".config.toml",
			wantExtension: ".bak",
			wantHasExt:    true,
		},

		// Directories
		{
			name:          "directory without dot",
			filename:      "src",
			isDir:         true,
			wantBaseName:  "src",
			wantExtension: "",
			wantHasExt:    false,
		},
		{
			name:          "directory with dot in name",
			filename:      "node_modules",
			isDir:         true,
			wantBaseName:  "node_modules",
			wantExtension: "",
			wantHasExt:    false,
		},
		{
			name:          "directory with extension-like name",
			filename:      "backup.old",
			isDir:         true,
			wantBaseName:  "backup.old",
			wantExtension: "",
			wantHasExt:    false,
		},

		// Edge cases
		{
			name:          "file with only extension (.txt)",
			filename:      ".txt",
			isDir:         false,
			wantBaseName:  ".txt",
			wantExtension: "",
			wantHasExt:    false,
		},
		{
			name:          "file ending with dot (file.)",
			filename:      "file.",
			isDir:         false,
			wantBaseName:  "file.",
			wantExtension: "",
			wantHasExt:    false,
		},
		{
			name:          "file with multiple consecutive dots (file..txt)",
			filename:      "file..txt",
			isDir:         false,
			wantBaseName:  "file.",
			wantExtension: ".txt",
			wantHasExt:    true,
		},
		{
			name:          "file with only dots (..)",
			filename:      "..",
			isDir:         false,
			wantBaseName:  "..",
			wantExtension: "",
			wantHasExt:    false,
		},
		{
			name:          "file with triple dots (...)",
			filename:      "...",
			isDir:         false,
			wantBaseName:  "...",
			wantExtension: "",
			wantHasExt:    false,
		},
		{
			name:          "hidden file ending with dot (.config.)",
			filename:      ".config.",
			isDir:         false,
			wantBaseName:  ".config.",
			wantExtension: "",
			wantHasExt:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBaseName, gotExtension, gotHasExt := hasEditableExtension(tt.filename, tt.isDir)

			if gotBaseName != tt.wantBaseName {
				t.Errorf("hasEditableExtension() baseName = %q, want %q", gotBaseName, tt.wantBaseName)
			}
			if gotExtension != tt.wantExtension {
				t.Errorf("hasEditableExtension() extension = %q, want %q", gotExtension, tt.wantExtension)
			}
			if gotHasExt != tt.wantHasExt {
				t.Errorf("hasEditableExtension() hasExt = %v, want %v", gotHasExt, tt.wantHasExt)
			}
		})
	}
}
