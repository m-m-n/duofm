package fs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateTrashinfo(t *testing.T) {
	tests := []struct {
		name           string
		originalPath   string
		deletionTime   time.Time
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:         "simple path",
			originalPath: "/home/user/Documents/file.txt",
			deletionTime: time.Date(2026, 1, 25, 10, 30, 0, 0, time.Local),
			wantContains: []string{
				"[Trash Info]",
				"Path=/home/user/Documents/file.txt",
				"DeletionDate=2026-01-25T10:30:00",
			},
		},
		{
			name:         "path with spaces",
			originalPath: "/home/user/My Documents/my file.txt",
			deletionTime: time.Date(2026, 1, 25, 10, 30, 0, 0, time.Local),
			wantContains: []string{
				"[Trash Info]",
				"Path=/home/user/My%20Documents/my%20file.txt",
				"DeletionDate=2026-01-25T10:30:00",
			},
		},
		{
			name:         "path with japanese characters",
			originalPath: "/home/user/ドキュメント/ファイル.txt",
			deletionTime: time.Date(2026, 1, 25, 10, 30, 0, 0, time.Local),
			wantContains: []string{
				"[Trash Info]",
				"DeletionDate=2026-01-25T10:30:00",
			},
			wantNotContain: []string{
				"Path=/home/user/ドキュメント", // Should be URL encoded
			},
		},
		{
			name:         "path with special characters",
			originalPath: "/home/user/test%file&name.txt",
			deletionTime: time.Date(2026, 1, 25, 10, 30, 0, 0, time.Local),
			wantContains: []string{
				"[Trash Info]",
				"DeletionDate=2026-01-25T10:30:00",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := GenerateTrashinfo(tt.originalPath, tt.deletionTime)

			for _, want := range tt.wantContains {
				if !strings.Contains(content, want) {
					t.Errorf("GenerateTrashinfo() missing %q\ngot:\n%s", want, content)
				}
			}

			for _, notWant := range tt.wantNotContain {
				if strings.Contains(content, notWant) {
					t.Errorf("GenerateTrashinfo() should not contain %q\ngot:\n%s", notWant, content)
				}
			}
		})
	}
}

func TestParseTrashinfo(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantPath    string
		wantDate    time.Time
		wantErr     bool
		errContains string
	}{
		{
			name: "valid trashinfo",
			content: `[Trash Info]
Path=/home/user/Documents/file.txt
DeletionDate=2026-01-25T10:30:00
`,
			wantPath: "/home/user/Documents/file.txt",
			wantDate: time.Date(2026, 1, 25, 10, 30, 0, 0, time.Local),
			wantErr:  false,
		},
		{
			name: "url encoded path",
			content: `[Trash Info]
Path=/home/user/My%20Documents/my%20file.txt
DeletionDate=2026-01-25T10:30:00
`,
			wantPath: "/home/user/My Documents/my file.txt",
			wantDate: time.Date(2026, 1, 25, 10, 30, 0, 0, time.Local),
			wantErr:  false,
		},
		{
			name: "missing header",
			content: `Path=/home/user/file.txt
DeletionDate=2026-01-25T10:30:00
`,
			wantErr:     true,
			errContains: "invalid trashinfo",
		},
		{
			name: "missing path",
			content: `[Trash Info]
DeletionDate=2026-01-25T10:30:00
`,
			wantErr:     true,
			errContains: "missing Path",
		},
		{
			name: "missing deletion date",
			content: `[Trash Info]
Path=/home/user/file.txt
`,
			wantErr:     true,
			errContains: "missing DeletionDate",
		},
		{
			name: "invalid date format",
			content: `[Trash Info]
Path=/home/user/file.txt
DeletionDate=2026/01/25 10:30:00
`,
			wantErr:     true,
			errContains: "invalid DeletionDate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with content
			tmpDir := t.TempDir()
			trashinfoPath := filepath.Join(tmpDir, "test.trashinfo")
			if err := os.WriteFile(trashinfoPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			info, err := ParseTrashinfo(trashinfoPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseTrashinfo() expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseTrashinfo() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTrashinfo() unexpected error: %v", err)
				return
			}

			if info.OriginalPath != tt.wantPath {
				t.Errorf("ParseTrashinfo() path = %q, want %q", info.OriginalPath, tt.wantPath)
			}

			if !info.DeletionDate.Equal(tt.wantDate) {
				t.Errorf("ParseTrashinfo() date = %v, want %v", info.DeletionDate, tt.wantDate)
			}
		})
	}
}

func TestURLEncodePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "simple path",
			path: "/home/user/file.txt",
			want: "/home/user/file.txt",
		},
		{
			name: "path with space",
			path: "/home/user/my file.txt",
			want: "/home/user/my%20file.txt",
		},
		{
			name: "path with multiple spaces",
			path: "/home/user/my documents/my file.txt",
			want: "/home/user/my%20documents/my%20file.txt",
		},
		{
			name: "path with percent",
			path: "/home/user/50%off.txt",
			want: "/home/user/50%25off.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := urlEncodePath(tt.path)
			if got != tt.want {
				t.Errorf("urlEncodePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestURLDecodePath(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
		want    string
		wantErr bool
	}{
		{
			name:    "simple path",
			encoded: "/home/user/file.txt",
			want:    "/home/user/file.txt",
		},
		{
			name:    "encoded space",
			encoded: "/home/user/my%20file.txt",
			want:    "/home/user/my file.txt",
		},
		{
			name:    "encoded percent",
			encoded: "/home/user/50%25off.txt",
			want:    "/home/user/50%off.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := urlDecodePath(tt.encoded)
			if (err != nil) != tt.wantErr {
				t.Errorf("urlDecodePath(%q) error = %v, wantErr %v", tt.encoded, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("urlDecodePath(%q) = %q, want %q", tt.encoded, got, tt.want)
			}
		})
	}
}
