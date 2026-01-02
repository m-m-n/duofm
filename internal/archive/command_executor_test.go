package archive

import (
	"context"
	"testing"
	"time"
)

func TestNewCommandExecutor(t *testing.T) {
	executor := NewCommandExecutor()
	if executor == nil {
		t.Fatal("NewCommandExecutor() returned nil")
	}
}

func TestCommandExecutor_ExecuteCommand(t *testing.T) {
	executor := NewCommandExecutor()
	ctx := context.Background()

	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
	}{
		{
			name:    "successful command - echo",
			command: "echo",
			args:    []string{"hello"},
			wantErr: false,
		},
		{
			name:    "successful command - ls with directory",
			command: "ls",
			args:    []string{"/tmp"},
			wantErr: false,
		},
		{
			name:    "nonexistent command",
			command: "definitely_not_a_real_command_12345",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "command with error exit code",
			command: "ls",
			args:    []string{"/definitely/not/a/real/path/12345"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := executor.ExecuteCommand(ctx, tt.command, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandExecutor_ExecuteCommandWithTimeout(t *testing.T) {
	executor := NewCommandExecutor()

	t.Run("command completes before timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, _, err := executor.ExecuteCommand(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("ExecuteCommand() unexpected error: %v", err)
		}
	})

	t.Run("command times out", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, _, err := executor.ExecuteCommand(ctx, "sleep", "5")
		if err == nil {
			t.Error("ExecuteCommand() expected timeout error, got nil")
		}
	})
}

func TestCommandExecutor_ExecuteCommandWithCancellation(t *testing.T) {
	executor := NewCommandExecutor()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	_, _, err := executor.ExecuteCommand(ctx, "sleep", "10")
	if err == nil {
		t.Error("ExecuteCommand() expected cancellation error, got nil")
	}
}

func TestCommandExecutor_ExecuteCommandWithProgress(t *testing.T) {
	executor := NewCommandExecutor()
	ctx := context.Background()

	t.Run("captures all lines", func(t *testing.T) {
		var lines []string
		handler := func(line string) error {
			lines = append(lines, line)
			return nil
		}

		// Use printf to output multiple lines
		stderr, err := executor.ExecuteCommandWithProgress(ctx, "", handler, "sh", "-c", "printf 'line1\nline2\nline3\n'")
		if err != nil {
			t.Errorf("ExecuteCommandWithProgress() error = %v, stderr = %s", err, stderr)
		}

		if len(lines) != 3 {
			t.Errorf("ExecuteCommandWithProgress() captured %d lines, want 3: %v", len(lines), lines)
		}

		expected := []string{"line1", "line2", "line3"}
		for i, exp := range expected {
			if i < len(lines) && lines[i] != exp {
				t.Errorf("line[%d] = %q, want %q", i, lines[i], exp)
			}
		}
	})

	t.Run("nil handler works", func(t *testing.T) {
		stderr, err := executor.ExecuteCommandWithProgress(ctx, "", nil, "echo", "hello")
		if err != nil {
			t.Errorf("ExecuteCommandWithProgress() with nil handler error = %v, stderr = %s", err, stderr)
		}
	})

	t.Run("handles command error", func(t *testing.T) {
		var lines []string
		handler := func(line string) error {
			lines = append(lines, line)
			return nil
		}

		_, err := executor.ExecuteCommandWithProgress(ctx, "", handler, "ls", "/definitely/not/a/real/path/12345")
		if err == nil {
			t.Error("ExecuteCommandWithProgress() expected error for nonexistent path")
		}
	})

	t.Run("respects working directory", func(t *testing.T) {
		var lines []string
		handler := func(line string) error {
			lines = append(lines, line)
			return nil
		}

		stderr, err := executor.ExecuteCommandWithProgress(ctx, "/tmp", handler, "pwd")
		if err != nil {
			t.Errorf("ExecuteCommandWithProgress() error = %v, stderr = %s", err, stderr)
		}

		if len(lines) == 0 {
			t.Error("ExecuteCommandWithProgress() captured no lines")
		} else if lines[0] != "/tmp" {
			t.Errorf("pwd output = %q, want /tmp", lines[0])
		}
	})
}
