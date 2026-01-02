package archive

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
)

// CommandExecutor provides methods to execute external CLI commands
type CommandExecutor struct{}

// NewCommandExecutor creates a new CommandExecutor instance
func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{}
}

// ExecuteCommand runs a command with the given arguments and returns stdout, stderr, and error
func (e *CommandExecutor) ExecuteCommand(ctx context.Context, command string, args ...string) (stdout string, stderr string, err error) {
	return e.ExecuteCommandInDir(ctx, "", command, args...)
}

// ExecuteCommandInDir runs a command in the specified directory and returns stdout, stderr, and error
// This is safe for concurrent use as it uses cmd.Dir instead of os.Chdir()
func (e *CommandExecutor) ExecuteCommandInDir(ctx context.Context, dir string, command string, args ...string) (stdout string, stderr string, err error) {
	cmd := exec.CommandContext(ctx, command, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if err != nil {
		// Wrap the error with archive error
		return stdout, stderr, NewArchiveErrorWithDetails(
			ErrArchiveInternalError,
			"Command execution failed",
			stderr,
			err,
		)
	}

	return stdout, stderr, nil
}

// LineHandler is a function type that processes each line of command output
// It receives the line content and returns an error if processing should stop
type LineHandler func(line string) error

// ExecuteCommandWithProgress runs a command and processes stdout line by line using the provided handler.
// This enables real-time progress updates by streaming output instead of buffering it all.
// The handler is called for each line of stdout; stderr is still buffered and returned.
func (e *CommandExecutor) ExecuteCommandWithProgress(ctx context.Context, dir string, lineHandler LineHandler, command string, args ...string) (stderr string, err error) {
	cmd := exec.CommandContext(ctx, command, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	// Get stdout pipe for streaming
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", NewArchiveError(ErrArchiveInternalError, "Failed to create stdout pipe", err)
	}

	// Buffer stderr
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", NewArchiveError(ErrArchiveInternalError, "Failed to start command", err)
	}

	// Read stdout line by line
	scanner := bufio.NewScanner(stdoutPipe)
	for scanner.Scan() {
		line := scanner.Text()
		if lineHandler != nil {
			// Call the handler for each line; ignore handler errors to continue processing (FR6.6 fallback)
			_ = lineHandler(line)
		}
	}

	// Check for scanner errors
	if scanErr := scanner.Err(); scanErr != nil && scanErr != io.EOF {
		// Non-fatal: log but continue
	}

	// Wait for command completion
	waitErr := cmd.Wait()
	stderr = errBuf.String()

	if waitErr != nil {
		return stderr, NewArchiveErrorWithDetails(
			ErrArchiveInternalError,
			"Command execution failed",
			stderr,
			waitErr,
		)
	}

	return stderr, nil
}
