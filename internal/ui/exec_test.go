package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sakura/duofm/internal/config"
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

func TestShellCommandFinishedMsgFields(t *testing.T) {
	msg := shellCommandFinishedMsg{
		err:     nil,
		command: "ls -la",
		workDir: "/tmp",
	}
	if msg.command != "ls -la" {
		t.Errorf("expected command 'ls -la', got %q", msg.command)
	}
	if msg.workDir != "/tmp" {
		t.Errorf("expected workDir '/tmp', got %q", msg.workDir)
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
			logFile := filepath.Join(tmpDir, "test.log")
			cmd := executeShellCommand(tt.command, tt.workDir, logFile)
			if cmd == nil {
				t.Error("executeShellCommand() returned nil command")
			}
		})
	}
}

func TestGetEditor(t *testing.T) {
	// lookPathFound simulates all commands being available
	lookPathFound := func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}
	// lookPathNoVim simulates vim missing, vi available
	lookPathNoVim := func(file string) (string, error) {
		if file == "vim" {
			return "", exec.ErrNotFound
		}
		return "/usr/bin/" + file, nil
	}
	// lookPathNone simulates neither vim nor vi available
	lookPathNone := func(file string) (string, error) {
		return "", exec.ErrNotFound
	}

	tests := []struct {
		name       string
		envValue   string
		setEnv     bool
		lookPathFn func(string) (string, error)
		want       string
	}{
		{
			name:       "EDITOR set to nano",
			envValue:   "nano",
			setEnv:     true,
			lookPathFn: lookPathFound,
			want:       "nano",
		},
		{
			name:       "EDITOR set to emacs",
			envValue:   "emacs",
			setEnv:     true,
			lookPathFn: lookPathFound,
			want:       "emacs",
		},
		{
			name:       "EDITOR not set, vim available",
			setEnv:     false,
			lookPathFn: lookPathFound,
			want:       "vim",
		},
		{
			name:       "EDITOR set to empty string, vim available",
			envValue:   "",
			setEnv:     true,
			lookPathFn: lookPathFound,
			want:       "vim",
		},
		{
			name:       "EDITOR not set, vim unavailable, vi available",
			setEnv:     false,
			lookPathFn: lookPathNoVim,
			want:       "vi",
		},
		{
			name:       "EDITOR empty, vim unavailable, vi available",
			envValue:   "",
			setEnv:     true,
			lookPathFn: lookPathNoVim,
			want:       "vi",
		},
		{
			name:       "EDITOR not set, neither vim nor vi available",
			setEnv:     false,
			lookPathFn: lookPathNone,
			want:       "vi",
		},
		{
			name:       "EDITOR with spaces returns as-is",
			envValue:   "vim -u NONE",
			setEnv:     true,
			lookPathFn: lookPathFound,
			want:       "vim -u NONE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original lookPathFn
			origLookPath := lookPathFn
			defer func() { lookPathFn = origLookPath }()
			lookPathFn = tt.lookPathFn

			// Save original EDITOR value
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
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "PAGER set to moar",
			envValue: "moar",
			setEnv:   true,
			wantCmd:  "moar",
			wantArgs: nil,
		},
		{
			name:     "PAGER set to cat",
			envValue: "cat",
			setEnv:   true,
			wantCmd:  "cat",
			wantArgs: nil,
		},
		{
			name:    "PAGER not set",
			setEnv:  false,
			wantCmd: "less",
		},
		{
			name:     "PAGER set to empty string",
			envValue: "",
			setEnv:   true,
			wantCmd:  "less",
		},
		{
			name:     "PAGER with arguments",
			envValue: "less -R",
			setEnv:   true,
			wantCmd:  "less",
			wantArgs: []string{"-R"},
		},
		{
			name:     "PAGER with multiple arguments",
			envValue: "less -R -N",
			setEnv:   true,
			wantCmd:  "less",
			wantArgs: []string{"-R", "-N"},
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

			gotCmd, gotArgs := getPager()
			if gotCmd != tt.wantCmd {
				t.Errorf("getPager() cmd = %q, want %q", gotCmd, tt.wantCmd)
			}
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("getPager() args len = %d, want %d", len(gotArgs), len(tt.wantArgs))
			} else {
				for i, arg := range gotArgs {
					if arg != tt.wantArgs[i] {
						t.Errorf("getPager() args[%d] = %q, want %q", i, arg, tt.wantArgs[i])
					}
				}
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

func TestOpenWithCustomForeground(t *testing.T) {
	tests := []struct {
		name        string
		application string
		file        string
		workDir     string
		wantNil     bool
	}{
		{
			name:        "valid absolute path application",
			application: "/bin/cat",
			file:        "/tmp/test.txt",
			workDir:     "/tmp",
			wantNil:     false,
		},
		{
			name:        "valid command name in PATH",
			application: "cat",
			file:        "/tmp/test.txt",
			workDir:     "/tmp",
			wantNil:     false,
		},
		{
			name:        "non-existent application path",
			application: "/nonexistent/application",
			file:        "/tmp/test.txt",
			workDir:     "/tmp",
			wantNil:     false, // Returns a command that will produce an error message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := openWithCustomForeground(tt.application, tt.file, tt.workDir)
			if (cmd == nil) != tt.wantNil {
				t.Errorf("openWithCustomForeground() returned nil = %v, want nil = %v", cmd == nil, tt.wantNil)
			}
		})
	}
}

func TestOpenWithCustomForegroundReturnsCmd(t *testing.T) {
	// Create a temporary file and directory
	tmpDir, err := os.MkdirTemp("", "test_custom_workdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	f, err := os.CreateTemp(tmpDir, "test_file")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Test that openWithCustomForeground returns a non-nil command
	cmd := openWithCustomForeground("cat", f.Name(), tmpDir)
	if cmd == nil {
		t.Error("openWithCustomForeground() returned nil command")
	}
}

func TestOpenWithMIME(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_mime")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name          string
		filename      string
		mimeConfig    config.MIMEBehaviorConfig
		wantNil       bool
		wantStatusMsg bool
	}{
		{
			name:     "text file with text/* rule",
			filename: "test.txt",
			mimeConfig: config.MIMEBehaviorConfig{
				Rules: map[string][]string{
					"text/*": {"cat"},
				},
			},
			wantNil:       false,
			wantStatusMsg: false,
		},
		{
			name:     "exact MIME match takes priority",
			filename: "test.txt",
			mimeConfig: config.MIMEBehaviorConfig{
				Rules: map[string][]string{
					"text/plain": {"head"},
					"text/*":     {"cat"},
				},
			},
			wantNil:       false,
			wantStatusMsg: false,
		},
		{
			name:          "no matching rule falls back to pager",
			filename:      "test.xyz",
			mimeConfig:    config.MIMEBehaviorConfig{},
			wantNil:       false,
			wantStatusMsg: false,
		},
		{
			name:     "empty rules falls back to pager",
			filename: "test.txt",
			mimeConfig: config.MIMEBehaviorConfig{
				Rules: map[string][]string{},
			},
			wantNil:       false,
			wantStatusMsg: false,
		},
		{
			name:     "command with options",
			filename: "test.txt",
			mimeConfig: config.MIMEBehaviorConfig{
				Rules: map[string][]string{
					"text/*": {"head -n 20"},
				},
			},
			wantNil:       false,
			wantStatusMsg: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the test file
			filePath := tmpDir + "/" + tt.filename
			f, err := os.Create(filePath)
			if err != nil {
				t.Fatal(err)
			}
			f.Close()

			cmd, statusMsg := openWithMIME(filePath, tmpDir, tt.mimeConfig)
			if (cmd == nil) != tt.wantNil {
				t.Errorf("openWithMIME() returned nil = %v, want nil = %v", cmd == nil, tt.wantNil)
			}
			if tt.wantStatusMsg && statusMsg == "" {
				t.Error("openWithMIME() expected status message but got empty")
			}
			if !tt.wantStatusMsg && statusMsg != "" {
				t.Errorf("openWithMIME() unexpected status message: %q", statusMsg)
			}
		})
	}
}

func TestOpenWithMIME_CommandNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_mime_notfound")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a text file
	filePath := tmpDir + "/test.txt"
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Configure with a non-existent command followed by a valid one
	mimeConfig := config.MIMEBehaviorConfig{
		Rules: map[string][]string{
			"text/*": {"nonexistent_command_xyz", "cat"},
		},
	}

	// Should return a command (will try first, fail LookPath, then try second)
	cmd, statusMsg := openWithMIME(filePath, tmpDir, mimeConfig)
	if cmd == nil {
		t.Error("openWithMIME() should return a command even with first command not found")
	}
	// Second command (cat) should be found, so no status message
	if statusMsg != "" {
		t.Errorf("openWithMIME() unexpected status message: %q", statusMsg)
	}
}

func TestOpenWithMIME_AllCommandsNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_mime_allfail")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a text file
	filePath := tmpDir + "/test.txt"
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Configure with only non-existent commands
	mimeConfig := config.MIMEBehaviorConfig{
		Rules: map[string][]string{
			"text/*": {"nonexistent_cmd_1", "nonexistent_cmd_2"},
		},
	}

	// Should fall back to pager with status message
	cmd, statusMsg := openWithMIME(filePath, tmpDir, mimeConfig)
	if cmd == nil {
		t.Error("openWithMIME() should fall back to pager when all commands fail")
	}
	if statusMsg == "" {
		t.Error("openWithMIME() should return status message when all commands fail")
	}
	// Status message should mention the failed commands
	if !strings.Contains(statusMsg, "nonexistent_cmd_1") || !strings.Contains(statusMsg, "nonexistent_cmd_2") {
		t.Errorf("status message should contain failed command names, got: %q", statusMsg)
	}
}

func TestOpenWithMIME_FallbackNoMIMEMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_mime_fallback")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with unknown MIME type
	filePath := tmpDir + "/test.xyz"
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// No MIME match, fallback has valid command (cat)
	mimeConfig := config.MIMEBehaviorConfig{
		Rules:    map[string][]string{},
		Fallback: []string{"cat"},
	}

	cmd, statusMsg := openWithMIME(filePath, tmpDir, mimeConfig)
	if cmd == nil {
		t.Error("openWithMIME() should return command from fallback")
	}
	if statusMsg != "" {
		t.Errorf("openWithMIME() unexpected status message: %q", statusMsg)
	}
}

func TestOpenWithMIME_FallbackAllCommandsMissing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_mime_fallback_fail")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with unknown MIME type
	filePath := tmpDir + "/test.xyz"
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// No MIME match, all fallback commands missing
	mimeConfig := config.MIMEBehaviorConfig{
		Rules:    map[string][]string{},
		Fallback: []string{"nonexist1", "nonexist2"},
	}

	cmd, statusMsg := openWithMIME(filePath, tmpDir, mimeConfig)
	if cmd == nil {
		t.Error("openWithMIME() should fall back to pager")
	}
	if statusMsg == "" {
		t.Error("openWithMIME() should return status message when all fallback commands fail")
	}
	if !strings.Contains(statusMsg, "nonexist1") || !strings.Contains(statusMsg, "nonexist2") {
		t.Errorf("status message should contain failed fallback command names, got: %q", statusMsg)
	}
}

func TestOpenWithMIME_FallbackNoFallbackConfigured(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_mime_no_fallback")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with unknown MIME type
	filePath := tmpDir + "/test.xyz"
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// No MIME match, no fallback configured
	mimeConfig := config.MIMEBehaviorConfig{
		Rules: map[string][]string{},
	}

	cmd, statusMsg := openWithMIME(filePath, tmpDir, mimeConfig)
	if cmd == nil {
		t.Error("openWithMIME() should fall back to pager")
	}
	// No status message when no fallback configured (silent fallback)
	if statusMsg != "" {
		t.Errorf("openWithMIME() unexpected status message: %q", statusMsg)
	}
}

func TestOpenWithMIME_AllMIMEFailFallbackWorks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_mime_fail_fallback_ok")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a text file
	filePath := tmpDir + "/test.txt"
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// MIME rule matches but all commands fail, fallback has valid command
	mimeConfig := config.MIMEBehaviorConfig{
		Rules: map[string][]string{
			"text/*": {"nonexistent_mime_cmd"},
		},
		Fallback: []string{"cat"},
	}

	cmd, statusMsg := openWithMIME(filePath, tmpDir, mimeConfig)
	if cmd == nil {
		t.Error("openWithMIME() should return command from fallback")
	}
	if statusMsg != "" {
		t.Errorf("openWithMIME() unexpected status message: %q", statusMsg)
	}
}

func TestOpenWithMIME_MIMEMatchFallbackNotUsed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_mime_match_no_fallback")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a text file
	filePath := tmpDir + "/test.txt"
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// MIME rule matches and command is available, fallback should NOT be used
	mimeConfig := config.MIMEBehaviorConfig{
		Rules: map[string][]string{
			"text/*": {"cat"},
		},
		Fallback: []string{"nonexistent_fallback_cmd"},
	}

	cmd, statusMsg := openWithMIME(filePath, tmpDir, mimeConfig)
	if cmd == nil {
		t.Error("openWithMIME() should return MIME rule command")
	}
	if statusMsg != "" {
		t.Errorf("openWithMIME() unexpected status message: %q", statusMsg)
	}
}

func TestOpenWithMIME_FallbackTriesInOrder(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_mime_fallback_order")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with unknown MIME type
	filePath := tmpDir + "/test.xyz"
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// First fallback command missing, second valid
	mimeConfig := config.MIMEBehaviorConfig{
		Rules:    map[string][]string{},
		Fallback: []string{"nonexist_cmd", "cat"},
	}

	cmd, statusMsg := openWithMIME(filePath, tmpDir, mimeConfig)
	if cmd == nil {
		t.Error("openWithMIME() should return second fallback command")
	}
	if statusMsg != "" {
		t.Errorf("openWithMIME() unexpected status message: %q", statusMsg)
	}
}

func TestOpenWithMIME_FallbackWithOptions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_mime_fallback_opts")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with unknown MIME type
	filePath := tmpDir + "/test.xyz"
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Fallback command with options
	mimeConfig := config.MIMEBehaviorConfig{
		Rules:    map[string][]string{},
		Fallback: []string{"head -n 20"},
	}

	cmd, statusMsg := openWithMIME(filePath, tmpDir, mimeConfig)
	if cmd == nil {
		t.Error("openWithMIME() should return fallback command with options")
	}
	if statusMsg != "" {
		t.Errorf("openWithMIME() unexpected status message: %q", statusMsg)
	}
}

func TestOpenWithMIME_AllMIMEAndFallbackFail(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_mime_all_fail")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a text file
	filePath := tmpDir + "/test.txt"
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// MIME rule matches but all commands fail, fallback also fails
	mimeConfig := config.MIMEBehaviorConfig{
		Rules: map[string][]string{
			"text/*": {"nonexistent_mime_1", "nonexistent_mime_2"},
		},
		Fallback: []string{"nonexistent_fb_1", "nonexistent_fb_2"},
	}

	cmd, statusMsg := openWithMIME(filePath, tmpDir, mimeConfig)
	if cmd == nil {
		t.Error("openWithMIME() should fall back to pager")
	}
	if statusMsg == "" {
		t.Error("openWithMIME() should return status message when all commands fail")
	}
	// Status message should include both MIME and fallback command names
	if !strings.Contains(statusMsg, "nonexistent_mime_1") {
		t.Errorf("status message should contain MIME command name, got: %q", statusMsg)
	}
	if !strings.Contains(statusMsg, "nonexistent_fb_1") {
		t.Errorf("status message should contain fallback command name, got: %q", statusMsg)
	}
}
