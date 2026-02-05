package clipboard

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os/exec"
	"testing"
	"time"
)

// TestBuildOSC52Sequence_ASCII tests OSC 52 sequence generation for ASCII strings.
func TestBuildOSC52Sequence_ASCII(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple filename",
			input:    "test.txt",
			expected: fmt.Sprintf("\033]52;c;%s\a", base64.StdEncoding.EncodeToString([]byte("test.txt"))),
		},
		{
			name:     "filename with spaces",
			input:    "my file.txt",
			expected: fmt.Sprintf("\033]52;c;%s\a", base64.StdEncoding.EncodeToString([]byte("my file.txt"))),
		},
		{
			name:     "full path",
			input:    "/home/user/documents/test.txt",
			expected: fmt.Sprintf("\033]52;c;%s\a", base64.StdEncoding.EncodeToString([]byte("/home/user/documents/test.txt"))),
		},
		{
			name:     "empty string",
			input:    "",
			expected: fmt.Sprintf("\033]52;c;%s\a", base64.StdEncoding.EncodeToString([]byte(""))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildOSC52Sequence(tt.input)
			if result != tt.expected {
				t.Errorf("buildOSC52Sequence(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestBuildOSC52Sequence_Unicode tests OSC 52 sequence generation for Unicode strings.
func TestBuildOSC52Sequence_Unicode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Japanese filename",
			input:    "テスト.txt",
			expected: fmt.Sprintf("\033]52;c;%s\a", base64.StdEncoding.EncodeToString([]byte("テスト.txt"))),
		},
		{
			name:     "Chinese characters",
			input:    "文档.pdf",
			expected: fmt.Sprintf("\033]52;c;%s\a", base64.StdEncoding.EncodeToString([]byte("文档.pdf"))),
		},
		{
			name:     "emoji filename",
			input:    "📁 folder",
			expected: fmt.Sprintf("\033]52;c;%s\a", base64.StdEncoding.EncodeToString([]byte("📁 folder"))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildOSC52Sequence(tt.input)
			if result != tt.expected {
				t.Errorf("buildOSC52Sequence(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestWriteOSC52 tests writing OSC 52 sequence to an io.Writer.
func TestWriteOSC52(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantErr  bool
		checkSeq bool
	}{
		{
			name:     "write ASCII text",
			text:     "test.txt",
			wantErr:  false,
			checkSeq: true,
		},
		{
			name:     "write Unicode text",
			text:     "テスト.txt",
			wantErr:  false,
			checkSeq: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writeOSC52(&buf, tt.text)

			if (err != nil) != tt.wantErr {
				t.Errorf("writeOSC52() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkSeq {
				expected := buildOSC52Sequence(tt.text)
				if buf.String() != expected {
					t.Errorf("writeOSC52() wrote %q, want %q", buf.String(), expected)
				}
			}
		})
	}
}

// TestWriteOSC52_ErrorWriter tests writeOSC52 with a failing writer.
func TestWriteOSC52_ErrorWriter(t *testing.T) {
	w := &failWriter{err: errors.New("write failed")}
	err := writeOSC52(w, "test.txt")
	if err == nil {
		t.Error("writeOSC52() should return error for failing writer")
	}
}

// failWriter is a test double that always returns an error on Write.
type failWriter struct {
	err error
}

func (w *failWriter) Write(p []byte) (int, error) {
	return 0, w.err
}

// TestFindClipboardCommand tests external clipboard command detection.
func TestFindClipboardCommand(t *testing.T) {
	// This test verifies the detection logic by checking the return structure.
	// We can't control which commands are installed, so we verify the function
	// returns valid results.
	cmd := findClipboardCommand()

	if cmd != nil {
		// If a command was found, it should have a non-empty name
		if cmd.name == "" {
			t.Error("findClipboardCommand() returned command with empty name")
		}

		// Verify it's one of the expected commands
		validCommands := map[string]bool{
			"wl-copy": true,
			"xclip":   true,
			"xsel":    true,
		}
		if !validCommands[cmd.name] {
			t.Errorf("findClipboardCommand() returned unexpected command: %s", cmd.name)
		}
	}
	// cmd == nil is also valid (no clipboard command available)
}

// TestFindClipboardCommand_DetectionOrder tests that detection follows wl-copy > xclip > xsel order.
func TestFindClipboardCommand_DetectionOrder(t *testing.T) {
	// Test the detection order by verifying the lookup sequence
	// This is tested implicitly through the function's behavior
	commands := []struct {
		name string
		args []string
	}{
		{"wl-copy", nil},
		{"xclip", []string{"-selection", "clipboard"}},
		{"xsel", []string{"--clipboard", "--input"}},
	}

	cmd := findClipboardCommand()
	if cmd == nil {
		t.Skip("No clipboard command available on this system")
	}

	// Find the first available command in expected order
	for _, expected := range commands {
		if _, err := exec.LookPath(expected.name); err == nil {
			if cmd.name != expected.name {
				t.Errorf("findClipboardCommand() = %s, want %s (first in detection order)", cmd.name, expected.name)
			}
			break
		}
	}
}

// TestExecClipboardCommand_Success tests successful external command execution.
func TestExecClipboardCommand_Success(t *testing.T) {
	// Use 'cat' as a safe command that accepts stdin and exits successfully
	if _, err := exec.LookPath("cat"); err != nil {
		t.Skip("cat command not available")
	}

	ctx := context.Background()
	err := execClipboardCommand(ctx, "cat", nil, "test text")
	if err != nil {
		t.Errorf("execClipboardCommand() with cat returned error: %v", err)
	}
}

// TestExecClipboardCommand_Failure tests external command execution failure.
func TestExecClipboardCommand_Failure(t *testing.T) {
	ctx := context.Background()
	err := execClipboardCommand(ctx, "false", nil, "test text")
	if err == nil {
		t.Error("execClipboardCommand() with 'false' should return error")
	}
}

// TestExecClipboardCommand_Timeout tests external command timeout.
func TestExecClipboardCommand_Timeout(t *testing.T) {
	if _, err := exec.LookPath("sleep"); err != nil {
		t.Skip("sleep command not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := execClipboardCommand(ctx, "sleep", []string{"10"}, "test text")
	if err == nil {
		t.Error("execClipboardCommand() should return error on timeout")
	}
}

// TestExecClipboardCommand_NonexistentCommand tests with a command that doesn't exist.
func TestExecClipboardCommand_NonexistentCommand(t *testing.T) {
	ctx := context.Background()
	err := execClipboardCommand(ctx, "/nonexistent/command", nil, "test text")
	if err == nil {
		t.Error("execClipboardCommand() should return error for nonexistent command")
	}
}

// withMockFindCmd temporarily replaces findClipboardCommandFunc for testing.
func withMockFindCmd(t *testing.T, mock func() *clipboardCmd) {
	t.Helper()
	original := findClipboardCommandFunc
	findClipboardCommandFunc = mock
	t.Cleanup(func() { findClipboardCommandFunc = original })
}

// TestWriteToClipboard_WithWriter_NoExtCmd tests WriteToClipboard with a writer and no external command.
func TestWriteToClipboard_WithWriter_NoExtCmd(t *testing.T) {
	withMockFindCmd(t, func() *clipboardCmd { return nil })

	var buf bytes.Buffer
	err := WriteToClipboard("test.txt", &buf)

	// OSC 52 attempted + no external command => success (best-effort)
	if err != nil {
		t.Errorf("WriteToClipboard() with valid writer and no ext cmd returned error: %v", err)
	}

	// Verify OSC 52 sequence was written
	expected := buildOSC52Sequence("test.txt")
	if buf.String() != expected {
		t.Errorf("WriteToClipboard() wrote %q to writer, want %q", buf.String(), expected)
	}
}

// TestWriteToClipboard_WithWriter_WithExtCmd tests WriteToClipboard with a writer and an external command.
func TestWriteToClipboard_WithWriter_WithExtCmd(t *testing.T) {
	if _, err := exec.LookPath("cat"); err != nil {
		t.Skip("cat command not available")
	}
	// Use "cat" as a mock clipboard command that reads stdin and exits 0
	withMockFindCmd(t, func() *clipboardCmd { return &clipboardCmd{name: "cat"} })

	var buf bytes.Buffer
	err := WriteToClipboard("test.txt", &buf)

	if err != nil {
		t.Errorf("WriteToClipboard() returned error: %v", err)
	}

	// OSC 52 should still have been written
	expected := buildOSC52Sequence("test.txt")
	if buf.String() != expected {
		t.Errorf("WriteToClipboard() wrote %q to writer, want %q", buf.String(), expected)
	}
}

// TestWriteToClipboard_NilWriter_NoExtCmd tests no clipboard method available.
func TestWriteToClipboard_NilWriter_NoExtCmd(t *testing.T) {
	withMockFindCmd(t, func() *clipboardCmd { return nil })

	err := WriteToClipboard("test.txt", nil)

	// No OSC 52 + no external command = error
	if err == nil {
		t.Error("WriteToClipboard() with nil writer and no external command should return error")
	}
}

// TestWriteToClipboard_NilWriter_WithExtCmd tests WriteToClipboard with nil writer but external command available.
func TestWriteToClipboard_NilWriter_WithExtCmd(t *testing.T) {
	if _, err := exec.LookPath("cat"); err != nil {
		t.Skip("cat command not available")
	}
	withMockFindCmd(t, func() *clipboardCmd { return &clipboardCmd{name: "cat"} })

	err := WriteToClipboard("test.txt", nil)

	// External command succeeds even without OSC 52
	if err != nil {
		t.Errorf("WriteToClipboard() with nil writer but ext cmd returned error: %v", err)
	}
}

// TestWriteToClipboard_ExtCmdFails tests WriteToClipboard when external command fails.
func TestWriteToClipboard_ExtCmdFails(t *testing.T) {
	withMockFindCmd(t, func() *clipboardCmd { return &clipboardCmd{name: "false"} })

	var buf bytes.Buffer
	err := WriteToClipboard("test.txt", &buf)

	// External command fails => error
	if err == nil {
		t.Error("WriteToClipboard() should return error when external command fails")
	}
}

// TestWriteToClipboard_FailedWriter_NoExtCmd tests WriteToClipboard with a failing writer and no external command.
func TestWriteToClipboard_FailedWriter_NoExtCmd(t *testing.T) {
	withMockFindCmd(t, func() *clipboardCmd { return nil })

	w := &failWriter{err: errors.New("write failed")}
	err := WriteToClipboard("test.txt", w)

	// Writer is non-nil so osc52Attempted = true, no external command => success (best-effort)
	if err != nil {
		t.Errorf("WriteToClipboard() with failed writer but non-nil should succeed: %v", err)
	}
}
