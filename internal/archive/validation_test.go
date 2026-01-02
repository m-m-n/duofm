package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateCompressionLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   int
		wantErr bool
	}{
		{
			name:    "minimum level",
			level:   0,
			wantErr: false,
		},
		{
			name:    "default level",
			level:   6,
			wantErr: false,
		},
		{
			name:    "maximum level",
			level:   9,
			wantErr: false,
		},
		{
			name:    "below minimum",
			level:   -1,
			wantErr: true,
		},
		{
			name:    "above maximum",
			level:   10,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCompressionLevel(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCompressionLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSources(t *testing.T) {
	// Create temp directory with test files
	tempDir := t.TempDir()
	testFile1 := filepath.Join(tempDir, "file1.txt")
	testFile2 := filepath.Join(tempDir, "file2.txt")
	testDir := filepath.Join(tempDir, "subdir")

	os.WriteFile(testFile1, []byte("test1"), 0644)
	os.WriteFile(testFile2, []byte("test2"), 0644)
	os.Mkdir(testDir, 0755)

	tests := []struct {
		name    string
		sources []string
		wantErr bool
	}{
		{
			name:    "valid file sources",
			sources: []string{testFile1, testFile2},
			wantErr: false,
		},
		{
			name:    "valid directory source",
			sources: []string{testDir},
			wantErr: false,
		},
		{
			name:    "mixed file and directory",
			sources: []string{testFile1, testDir},
			wantErr: false,
		},
		{
			name:    "empty sources",
			sources: []string{},
			wantErr: true,
		},
		{
			name:    "nil sources",
			sources: nil,
			wantErr: true,
		},
		{
			name:    "non-existent file",
			sources: []string{"/nonexistent/path/file.txt"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSources(tt.sources)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSources_PermissionDenied(t *testing.T) {
	// Skip if running as root (root can read any file)
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tempDir := t.TempDir()
	noReadFile := filepath.Join(tempDir, "noread.txt")

	// Create file with no read permission
	os.WriteFile(noReadFile, []byte("test"), 0000)
	defer os.Chmod(noReadFile, 0644) // Restore for cleanup

	err := ValidateSources([]string{noReadFile})
	if err == nil {
		t.Error("ValidateSources() expected error for unreadable file, got nil")
	}
}
