package archive

import (
	"os"
	"testing"
)

// writeTestFile is a helper function for tests
func writeTestFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func TestSmartExtractor_AnalyzeStructure(t *testing.T) {
	tests := []struct {
		name     string
		contents []string
		want     ExtractionMethod
	}{
		{
			name:     "single root directory",
			contents: []string{"mydir/", "mydir/file1.txt", "mydir/file2.txt"},
			want:     ExtractDirect,
		},
		{
			name:     "multiple root items",
			contents: []string{"file1.txt", "file2.txt", "dir/"},
			want:     ExtractToDirectory,
		},
		{
			name:     "single file",
			contents: []string{"file.txt"},
			want:     ExtractDirect,
		},
		{
			name:     "nested structure with single root",
			contents: []string{"root/", "root/sub/", "root/sub/file.txt"},
			want:     ExtractDirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewSmartExtractor()
			strategy := extractor.analyzeContents(tt.contents)
			if strategy.Method != tt.want {
				t.Errorf("analyzeContents() method = %v, want %v", strategy.Method, tt.want)
			}
		})
	}
}

func TestSmartExtractor_GetRootItems(t *testing.T) {
	tests := []struct {
		name     string
		contents []string
		wantLen  int
	}{
		{
			name:     "single root directory",
			contents: []string{"mydir/", "mydir/file1.txt", "mydir/subdir/", "mydir/subdir/file2.txt"},
			wantLen:  1,
		},
		{
			name:     "multiple root items",
			contents: []string{"file1.txt", "file2.txt", "dir/", "dir/file3.txt"},
			wantLen:  3,
		},
		{
			name:     "single file",
			contents: []string{"file.txt"},
			wantLen:  1,
		},
		{
			name:     "empty",
			contents: []string{},
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewSmartExtractor()
			roots := extractor.getRootItems(tt.contents)
			if len(roots) != tt.wantLen {
				t.Errorf("getRootItems() returned %d items, want %d", len(roots), tt.wantLen)
			}
		})
	}
}

func TestSmartExtractor_ParseTarOutput(t *testing.T) {
	extractor := NewSmartExtractor()

	// Sample tar -tvf output
	output := `-rw-r--r-- user/group      1024 2024-01-01 10:00 file1.txt
-rw-r--r-- user/group      2048 2024-01-01 10:00 file2.txt
drwxr-xr-x user/group         0 2024-01-01 10:00 dir/
-rw-r--r-- user/group       512 2024-01-01 10:00 dir/file3.txt`

	metadata, err := extractor.parseTarOutput("/test/archive.tar", output)
	if err != nil {
		t.Fatalf("parseTarOutput() error = %v", err)
	}

	expectedSize := int64(1024 + 2048 + 0 + 512)
	if metadata.ExtractedSize != expectedSize {
		t.Errorf("ExtractedSize = %d, want %d", metadata.ExtractedSize, expectedSize)
	}

	if metadata.FileCount != 4 {
		t.Errorf("FileCount = %d, want %d", metadata.FileCount, 4)
	}
}

func TestSmartExtractor_ParseZipOutput(t *testing.T) {
	extractor := NewSmartExtractor()

	// Sample unzip -l output
	output := `Archive:  archive.zip
  Length      Date    Time    Name
---------  ---------- -----   ----
     1024  2024-01-01 10:00   file1.txt
     2048  2024-01-01 10:00   file2.txt
        0  2024-01-01 10:00   dir/
      512  2024-01-01 10:00   dir/file3.txt
---------                     -------
     3584                     4 files`

	metadata, err := extractor.parseZipOutput("/test/archive.zip", output)
	if err != nil {
		t.Fatalf("parseZipOutput() error = %v", err)
	}

	expectedSize := int64(1024 + 2048 + 0 + 512)
	if metadata.ExtractedSize != expectedSize {
		t.Errorf("ExtractedSize = %d, want %d", metadata.ExtractedSize, expectedSize)
	}

	if metadata.FileCount != 4 {
		t.Errorf("FileCount = %d, want %d", metadata.FileCount, 4)
	}
}

func TestSmartExtractor_Parse7zOutput(t *testing.T) {
	extractor := NewSmartExtractor()

	// Sample 7z l output
	output := `7-Zip 21.07 (x64) : Copyright (c) 1999-2021 Igor Pavlov : 2021-12-26

Listing archive: archive.7z

--
Path = archive.7z
Type = 7z

   Date      Time    Attr         Size   Compressed  Name
------------------- ----- ------------ ------------  ------------------------
2024-01-01 10:00:00 ....A         1024          500  file1.txt
2024-01-01 10:00:00 ....A         2048         1000  file2.txt
2024-01-01 10:00:00 D....            0            0  dir
2024-01-01 10:00:00 ....A          512          250  dir/file3.txt
------------------- ----- ------------ ------------  ------------------------
2024-01-01 10:00:00               3584         1750  4 files`

	metadata, err := extractor.parse7zOutput("/test/archive.7z", output)
	if err != nil {
		t.Fatalf("parse7zOutput() error = %v", err)
	}

	expectedSize := int64(1024 + 2048 + 0 + 512)
	if metadata.ExtractedSize != expectedSize {
		t.Errorf("ExtractedSize = %d, want %d", metadata.ExtractedSize, expectedSize)
	}

	if metadata.FileCount != 4 {
		t.Errorf("FileCount = %d, want %d", metadata.FileCount, 4)
	}
}

func TestArchiveMetadata_Structure(t *testing.T) {
	metadata := ArchiveMetadata{
		ArchiveSize:   1000,
		ExtractedSize: 5000,
		FileCount:     10,
		Files:         []string{"file1.txt", "file2.txt"},
	}

	if metadata.ArchiveSize != 1000 {
		t.Errorf("ArchiveSize = %d, want %d", metadata.ArchiveSize, 1000)
	}
	if metadata.ExtractedSize != 5000 {
		t.Errorf("ExtractedSize = %d, want %d", metadata.ExtractedSize, 5000)
	}
	if metadata.FileCount != 10 {
		t.Errorf("FileCount = %d, want %d", metadata.FileCount, 10)
	}
	if len(metadata.Files) != 2 {
		t.Errorf("len(Files) = %d, want %d", len(metadata.Files), 2)
	}
}

func TestSmartExtractor_ParseTarOutput_PathTraversal(t *testing.T) {
	extractor := NewSmartExtractor()

	// Sample tar output with path traversal attack
	output := `-rw-r--r-- user/group      1024 2024-01-01 10:00 ../../../etc/passwd`

	_, err := extractor.parseTarOutput("/test/archive.tar", output)
	if err == nil {
		t.Error("parseTarOutput() should return error for path traversal")
	}
}

func TestSmartExtractor_ParseZipOutput_PathTraversal(t *testing.T) {
	extractor := NewSmartExtractor()

	// Sample unzip -l output with path traversal attack
	output := `Archive:  archive.zip
  Length      Date    Time    Name
---------  ---------- -----   ----
     1024  2024-01-01 10:00   ../../../etc/passwd
---------                     -------
     1024                     1 files`

	_, err := extractor.parseZipOutput("/test/archive.zip", output)
	if err == nil {
		t.Error("parseZipOutput() should return error for path traversal")
	}
}

func TestSmartExtractor_Parse7zOutput_PathTraversal(t *testing.T) {
	extractor := NewSmartExtractor()

	// Sample 7z l output with path traversal attack
	output := `7-Zip 21.07 (x64) : Copyright (c) 1999-2021 Igor Pavlov : 2021-12-26

Listing archive: archive.7z

--
Path = archive.7z
Type = 7z

   Date      Time    Attr         Size   Compressed  Name
------------------- ----- ------------ ------------  ------------------------
2024-01-01 10:00:00 ....A         1024          500  ../../../etc/passwd
------------------- ----- ------------ ------------  ------------------------
2024-01-01 10:00:00               1024          500  1 files`

	_, err := extractor.parse7zOutput("/test/archive.7z", output)
	if err == nil {
		t.Error("parse7zOutput() should return error for path traversal")
	}
}

func TestSmartExtractor_ParseTarOutput_Empty(t *testing.T) {
	extractor := NewSmartExtractor()

	metadata, err := extractor.parseTarOutput("/test/archive.tar", "")
	if err != nil {
		t.Fatalf("parseTarOutput() error = %v", err)
	}

	if metadata.FileCount != 0 {
		t.Errorf("FileCount = %d, want 0", metadata.FileCount)
	}
	if metadata.ExtractedSize != 0 {
		t.Errorf("ExtractedSize = %d, want 0", metadata.ExtractedSize)
	}
}

func TestSmartExtractor_ParseZipOutput_Empty(t *testing.T) {
	extractor := NewSmartExtractor()

	output := `Archive:  archive.zip
  Length      Date    Time    Name
---------  ---------- -----   ----
---------                     -------
        0                     0 files`

	metadata, err := extractor.parseZipOutput("/test/archive.zip", output)
	if err != nil {
		t.Fatalf("parseZipOutput() error = %v", err)
	}

	if metadata.FileCount != 0 {
		t.Errorf("FileCount = %d, want 0", metadata.FileCount)
	}
}

func TestSmartExtractor_Parse7zOutput_Empty(t *testing.T) {
	extractor := NewSmartExtractor()

	output := `7-Zip 21.07 (x64) : Copyright (c) 1999-2021 Igor Pavlov : 2021-12-26

Listing archive: archive.7z

--
Path = archive.7z
Type = 7z

   Date      Time    Attr         Size   Compressed  Name
------------------- ----- ------------ ------------  ------------------------
------------------- ----- ------------ ------------  ------------------------
                                    0            0  0 files`

	metadata, err := extractor.parse7zOutput("/test/archive.7z", output)
	if err != nil {
		t.Fatalf("parse7zOutput() error = %v", err)
	}

	if metadata.FileCount != 0 {
		t.Errorf("FileCount = %d, want 0", metadata.FileCount)
	}
}

func TestSmartExtractor_AnalyzeContents_Empty(t *testing.T) {
	extractor := NewSmartExtractor()
	strategy := extractor.analyzeContents([]string{})

	if strategy.Method != ExtractDirect {
		t.Errorf("analyzeContents() method = %v, want ExtractDirect for empty archive", strategy.Method)
	}
}

func TestSmartExtractor_GetRootItems_WhitespaceHandling(t *testing.T) {
	extractor := NewSmartExtractor()

	tests := []struct {
		name     string
		contents []string
		wantLen  int
	}{
		{
			name:     "items with whitespace",
			contents: []string{"  file1.txt  ", "  file2.txt  "},
			wantLen:  2,
		},
		{
			name:     "empty strings",
			contents: []string{"", "   ", "file.txt"},
			wantLen:  1,
		},
		{
			name:     "trailing slashes",
			contents: []string{"dir/", "dir/file.txt"},
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roots := extractor.getRootItems(tt.contents)
			if len(roots) != tt.wantLen {
				t.Errorf("getRootItems() returned %d items, want %d", len(roots), tt.wantLen)
			}
		})
	}
}

func TestSmartExtractor_GetFileSize(t *testing.T) {
	extractor := NewSmartExtractor()

	// Test non-existent file
	size := extractor.getFileSize("/nonexistent/file.txt")
	if size != 0 {
		t.Errorf("getFileSize() = %d, want 0 for non-existent file", size)
	}
}

func TestNewSmartExtractor(t *testing.T) {
	extractor := NewSmartExtractor()
	if extractor == nil {
		t.Fatal("NewSmartExtractor() returned nil")
	}
	if extractor.tarExecutor == nil {
		t.Error("tarExecutor is nil")
	}
	if extractor.zipExecutor == nil {
		t.Error("zipExecutor is nil")
	}
	if extractor.sevenZipExecutor == nil {
		t.Error("sevenZipExecutor is nil")
	}
}

func TestExtractionStrategy_Methods(t *testing.T) {
	tests := []struct {
		name   string
		method ExtractionMethod
	}{
		{name: "ExtractDirect", method: ExtractDirect},
		{name: "ExtractToDirectory", method: ExtractToDirectory},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := &ExtractionStrategy{Method: tt.method}
			if strategy.Method != tt.method {
				t.Errorf("ExtractionStrategy.Method = %v, want %v", strategy.Method, tt.method)
			}
		})
	}
}

func TestSmartExtractor_ParseTarOutput_AllFormats(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		wantFiles   int
		wantSize    int64
		shouldError bool
	}{
		{
			name: "valid tar output with multiple files",
			output: `-rw-r--r-- user/group     1024 2024-01-01 10:00 file1.txt
-rw-r--r-- user/group     2048 2024-01-01 10:01 file2.txt
drwxr-xr-x user/group        0 2024-01-01 10:02 subdir/`,
			wantFiles:   3,
			wantSize:    3072, // 1024 + 2048 + 0
			shouldError: false,
		},
		{
			name:        "empty output",
			output:      "",
			wantFiles:   0,
			wantSize:    0,
			shouldError: false,
		},
		{
			name: "output with symlink",
			output: `-rw-r--r-- user/group     1024 2024-01-01 10:00 file.txt
lrwxrwxrwx user/group        0 2024-01-01 10:01 link -> file.txt`,
			wantFiles:   2,
			wantSize:    1024,
			shouldError: false,
		},
		{
			name: "output with valid hardlink",
			output: `-rw-r--r-- user/group     1024 2024-01-01 10:00 file.txt
-rw-r--r-- user/group        0 2024-01-01 10:01 hardlink link to file.txt`,
			wantFiles:   2,
			wantSize:    1024,
			shouldError: false,
		},
		{
			name: "hardlink with path traversal",
			output: `-rw-r--r-- user/group     1024 2024-01-01 10:00 file.txt
-rw-r--r-- user/group        0 2024-01-01 10:01 hardlink link to ../../../etc/passwd`,
			wantFiles:   0,
			wantSize:    0,
			shouldError: true,
		},
		{
			name: "hardlink with absolute path",
			output: `-rw-r--r-- user/group     1024 2024-01-01 10:00 file.txt
-rw-r--r-- user/group        0 2024-01-01 10:01 hardlink link to /etc/passwd`,
			wantFiles:   0,
			wantSize:    0,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewSmartExtractor()
			// Create a temp file to get archive size
			tempDir := t.TempDir()
			tempFile := tempDir + "/test.tar"
			if err := writeTestFile(tempFile, "test"); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			metadata, err := extractor.parseTarOutput(tempFile, tt.output)
			if tt.shouldError {
				if err == nil {
					t.Error("parseTarOutput() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("parseTarOutput() unexpected error: %v", err)
				} else {
					if len(metadata.Files) != tt.wantFiles {
						t.Errorf("parseTarOutput() files = %d, want %d", len(metadata.Files), tt.wantFiles)
					}
					if metadata.ExtractedSize != tt.wantSize {
						t.Errorf("parseTarOutput() size = %d, want %d", metadata.ExtractedSize, tt.wantSize)
					}
				}
			}
		})
	}
}

func TestSmartExtractor_ParseZipOutput_AllFormats(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		wantFiles   int
		wantSize    int64
		shouldError bool
	}{
		{
			name: "valid zip output",
			output: `Archive:  test.zip
  Length      Date    Time    Name
---------  ---------- -----   ----
     1024  2024-01-01 10:00   file1.txt
     2048  2024-01-01 10:01   file2.txt
---------                     -------
     3072                     2 files`,
			wantFiles:   2,
			wantSize:    3072,
			shouldError: false,
		},
		{
			name:        "empty output",
			output:      "",
			wantFiles:   0,
			wantSize:    0,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewSmartExtractor()
			tempDir := t.TempDir()
			tempFile := tempDir + "/test.zip"
			if err := writeTestFile(tempFile, "test"); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			metadata, err := extractor.parseZipOutput(tempFile, tt.output)
			if tt.shouldError {
				if err == nil {
					t.Error("parseZipOutput() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("parseZipOutput() unexpected error: %v", err)
				} else {
					if len(metadata.Files) != tt.wantFiles {
						t.Errorf("parseZipOutput() files = %d, want %d", len(metadata.Files), tt.wantFiles)
					}
					if metadata.ExtractedSize != tt.wantSize {
						t.Errorf("parseZipOutput() size = %d, want %d", metadata.ExtractedSize, tt.wantSize)
					}
				}
			}
		})
	}
}

func TestSmartExtractor_Parse7zOutput_AllFormats(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		wantFiles   int
		wantSize    int64
		shouldError bool
	}{
		{
			name: "valid 7z output",
			output: `7-Zip [64] 16.02 : Copyright

Listing archive: test.7z

--
Path = test.7z
Type = 7z

   Date      Time    Attr         Size   Compressed  Name
------------------- ----- ------------ ------------  ------------------------
2024-01-01 10:00:00 ....A         1024          512  file1.txt
2024-01-01 10:01:00 ....A         2048         1024  file2.txt
------------------- ----- ------------ ------------  ------------------------
                                  3072         1536  2 files`,
			wantFiles:   2,
			wantSize:    3072,
			shouldError: false,
		},
		{
			name:        "empty output",
			output:      "",
			wantFiles:   0,
			wantSize:    0,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewSmartExtractor()
			tempDir := t.TempDir()
			tempFile := tempDir + "/test.7z"
			if err := writeTestFile(tempFile, "test"); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			metadata, err := extractor.parse7zOutput(tempFile, tt.output)
			if tt.shouldError {
				if err == nil {
					t.Error("parse7zOutput() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("parse7zOutput() unexpected error: %v", err)
				} else {
					if len(metadata.Files) != tt.wantFiles {
						t.Errorf("parse7zOutput() files = %d, want %d", len(metadata.Files), tt.wantFiles)
					}
					if metadata.ExtractedSize != tt.wantSize {
						t.Errorf("parse7zOutput() size = %d, want %d", metadata.ExtractedSize, tt.wantSize)
					}
				}
			}
		})
	}
}
