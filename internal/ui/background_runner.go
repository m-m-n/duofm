package ui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
)

// BackgroundRunner manages the lifecycle of a single background process.
// It starts a shell command, streams output line-by-line via callbacks,
// and supports cancellation via context.
type BackgroundRunner struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	running bool
	pane    PanePosition
	command string
}

// NewBackgroundRunner creates a new BackgroundRunner.
func NewBackgroundRunner() *BackgroundRunner {
	return &BackgroundRunner{}
}

// Start launches a background command via /bin/sh -c.
// onOutput is called for each line of stdout/stderr.
// onDone is called when the command exits with the exit error (nil on success).
// Returns an error if a command is already running.
func (r *BackgroundRunner) Start(command, workDir string, pane PanePosition,
	onOutput func(line string), onDone func(err error)) error {

	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("background command already running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	cmd.Dir = workDir
	// Set process group so we can kill the entire tree
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Combine stdout and stderr into one pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		r.mu.Unlock()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout pipe

	if err := cmd.Start(); err != nil {
		cancel()
		r.mu.Unlock()
		return fmt.Errorf("failed to start command: %w", err)
	}

	r.cmd = cmd
	r.cancel = cancel
	r.running = true
	r.pane = pane
	r.command = command
	r.mu.Unlock()

	// Scanner goroutine: reads lines and calls onOutput
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			onOutput(line)
		}
		// Drain any remaining data
		io.Copy(io.Discard, stdout)

		// Wait for process to exit
		exitErr := cmd.Wait()

		r.mu.Lock()
		r.running = false
		r.mu.Unlock()

		onDone(exitErr)
	}()

	return nil
}

// Cancel terminates the background process group.
func (r *BackgroundRunner) Cancel() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running || r.cancel == nil {
		return
	}

	// Kill the entire process group
	if r.cmd != nil && r.cmd.Process != nil {
		syscall.Kill(-r.cmd.Process.Pid, syscall.SIGKILL)
	}

	r.cancel()
}

// IsRunning returns whether a background command is currently executing.
func (r *BackgroundRunner) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

// Pane returns the pane position that launched the command.
func (r *BackgroundRunner) Pane() PanePosition {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.pane
}

// Command returns the command string being executed.
func (r *BackgroundRunner) Command() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.command
}
