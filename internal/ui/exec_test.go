package ui

import (
	"os"
	"testing"
)

func TestCheckReadPermission(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() string // returns file path
		cleanup func(string)
		wantErr bool
	}{
		{
			name: "readable file",
			setup: func() string {
				f, _ := os.CreateTemp("", "test")
				f.Close()
				return f.Name()
			},
			cleanup: func(path string) {
				os.Remove(path)
			},
			wantErr: false,
		},
		{
			name: "non-existent file",
			setup: func() string {
				return "/nonexistent/file/path"
			},
			cleanup: func(string) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			defer tt.cleanup(path)

			err := checkReadPermission(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkReadPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecFinishedMsg(t *testing.T) {
	// Test that execFinishedMsg can carry error information
	msg := execFinishedMsg{err: nil}
	if msg.err != nil {
		t.Error("expected nil error")
	}

	msg = execFinishedMsg{err: os.ErrNotExist}
	if msg.err != os.ErrNotExist {
		t.Error("expected ErrNotExist")
	}
}

func TestOpenWithViewerReturnsCmd(t *testing.T) {
	// Create a temporary file to use as a test target
	tmpDir, err := os.MkdirTemp("", "test_view_dir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	f, err := os.CreateTemp(tmpDir, "test_view")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Test that openWithViewer returns a non-nil command
	cmd := openWithViewer(f.Name(), tmpDir)
	if cmd == nil {
		t.Error("openWithViewer() returned nil command")
	}
}

func TestOpenWithEditorReturnsCmd(t *testing.T) {
	// Create a temporary file to use as a test target
	tmpDir, err := os.MkdirTemp("", "test_edit_dir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	f, err := os.CreateTemp(tmpDir, "test_edit")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Test that openWithEditor returns a non-nil command
	cmd := openWithEditor(f.Name(), tmpDir)
	if cmd == nil {
		t.Error("openWithEditor() returned nil command")
	}
}

func TestShellCommandFinishedMsg(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "success case - nil error",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "error case - command error",
			err:     os.ErrNotExist,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := shellCommandFinishedMsg{err: tt.err}
			if (msg.err != nil) != tt.wantErr {
				t.Errorf("shellCommandFinishedMsg.err = %v, wantErr %v", msg.err, tt.wantErr)
			}
		})
	}
}

func TestExecuteShellCommandReturnsCmd(t *testing.T) {
	// Create a temporary directory to use as working directory
	tmpDir, err := os.MkdirTemp("", "test_shell")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		command string
		workDir string
	}{
		{
			name:    "simple command",
			command: "echo hello",
			workDir: tmpDir,
		},
		{
			name:    "command with pipe",
			command: "ls -la | head -5",
			workDir: tmpDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := executeShellCommand(tt.command, tt.workDir)
			if cmd == nil {
				t.Error("executeShellCommand() returned nil command")
			}
		})
	}
}

func TestGetEditor(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		setEnv   bool
		want     string
	}{
		{
			name:     "EDITOR set to nano",
			envValue: "nano",
			setEnv:   true,
			want:     "nano",
		},
		{
			name:     "EDITOR set to emacs",
			envValue: "emacs",
			setEnv:   true,
			want:     "emacs",
		},
		{
			name:   "EDITOR not set",
			setEnv: false,
			want:   "vim",
		},
		{
			name:     "EDITOR set to empty string",
			envValue: "",
			setEnv:   true,
			want:     "vim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			originalValue, originalSet := os.LookupEnv("EDITOR")
			defer func() {
				if originalSet {
					os.Setenv("EDITOR", originalValue)
				} else {
					os.Unsetenv("EDITOR")
				}
			}()

			// Set test value
			if tt.setEnv {
				os.Setenv("EDITOR", tt.envValue)
			} else {
				os.Unsetenv("EDITOR")
			}

			got := getEditor()
			if got != tt.want {
				t.Errorf("getEditor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetPager(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		setEnv   bool
		want     string
	}{
		{
			name:     "PAGER set to moar",
			envValue: "moar",
			setEnv:   true,
			want:     "moar",
		},
		{
			name:     "PAGER set to cat",
			envValue: "cat",
			setEnv:   true,
			want:     "cat",
		},
		{
			name:   "PAGER not set",
			setEnv: false,
			want:   "less",
		},
		{
			name:     "PAGER set to empty string",
			envValue: "",
			setEnv:   true,
			want:     "less",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			originalValue, originalSet := os.LookupEnv("PAGER")
			defer func() {
				if originalSet {
					os.Setenv("PAGER", originalValue)
				} else {
					os.Unsetenv("PAGER")
				}
			}()

			// Set test value
			if tt.setEnv {
				os.Setenv("PAGER", tt.envValue)
			} else {
				os.Unsetenv("PAGER")
			}

			got := getPager()
			if got != tt.want {
				t.Errorf("getPager() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOpenWithViewerWithWorkDir(t *testing.T) {
	// Create a temporary file and directory
	tmpDir, err := os.MkdirTemp("", "test_view_workdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	f, err := os.CreateTemp(tmpDir, "test_file")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Test that openWithViewer returns a non-nil command with workDir
	cmd := openWithViewer(f.Name(), tmpDir)
	if cmd == nil {
		t.Error("openWithViewer() returned nil command")
	}
}

func TestOpenWithEditorWithWorkDir(t *testing.T) {
	// Create a temporary file and directory
	tmpDir, err := os.MkdirTemp("", "test_edit_workdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	f, err := os.CreateTemp(tmpDir, "test_file")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Test that openWithEditor returns a non-nil command with workDir
	cmd := openWithEditor(f.Name(), tmpDir)
	if cmd == nil {
		t.Error("openWithEditor() returned nil command")
	}
}
