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
	f, err := os.CreateTemp("", "test_view")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	// Test that openWithViewer returns a non-nil command
	cmd := openWithViewer(f.Name())
	if cmd == nil {
		t.Error("openWithViewer() returned nil command")
	}
}

func TestOpenWithEditorReturnsCmd(t *testing.T) {
	// Create a temporary file to use as a test target
	f, err := os.CreateTemp("", "test_edit")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	// Test that openWithEditor returns a non-nil command
	cmd := openWithEditor(f.Name())
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
