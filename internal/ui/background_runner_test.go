package ui

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewBackgroundRunner(t *testing.T) {
	runner := NewBackgroundRunner()
	if runner == nil {
		t.Fatal("NewBackgroundRunner returned nil")
	}
	if runner.IsRunning() {
		t.Error("new runner should not be running")
	}
	if runner.Command() != "" {
		t.Error("new runner should have empty command")
	}
}

func TestBackgroundRunner_Start_BasicOutput(t *testing.T) {
	runner := NewBackgroundRunner()

	var mu sync.Mutex
	var lines []string
	var done bool
	var doneErr error

	err := runner.Start("echo hello", "/tmp", LeftPane,
		func(line string) {
			mu.Lock()
			lines = append(lines, line)
			mu.Unlock()
		},
		func(err error) {
			mu.Lock()
			done = true
			doneErr = err
			mu.Unlock()
		},
	)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !runner.IsRunning() {
		t.Error("runner should be running after Start")
	}
	if runner.Command() != "echo hello" {
		t.Errorf("expected command 'echo hello', got %q", runner.Command())
	}
	if runner.Pane() != LeftPane {
		t.Errorf("expected LeftPane, got %v", runner.Pane())
	}

	// Wait for completion
	timeout := time.After(5 * time.Second)
	for {
		mu.Lock()
		isDone := done
		mu.Unlock()
		if isDone {
			break
		}
		select {
		case <-timeout:
			t.Fatal("timed out waiting for command completion")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	mu.Lock()
	defer mu.Unlock()

	if doneErr != nil {
		t.Errorf("unexpected error: %v", doneErr)
	}
	if len(lines) != 1 || lines[0] != "hello" {
		t.Errorf("expected [hello], got %v", lines)
	}
}

func TestBackgroundRunner_Start_MultilineOutput(t *testing.T) {
	runner := NewBackgroundRunner()

	var mu sync.Mutex
	var lines []string
	doneCh := make(chan struct{})

	err := runner.Start("printf 'line1\\nline2\\nline3'", "/tmp", LeftPane,
		func(line string) {
			mu.Lock()
			lines = append(lines, line)
			mu.Unlock()
		},
		func(err error) {
			close(doneCh)
		},
	)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out")
	}

	mu.Lock()
	defer mu.Unlock()

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
	}
}

func TestBackgroundRunner_Start_Stderr(t *testing.T) {
	runner := NewBackgroundRunner()

	var mu sync.Mutex
	var lines []string
	doneCh := make(chan struct{})

	err := runner.Start("echo stderr_output >&2", "/tmp", LeftPane,
		func(line string) {
			mu.Lock()
			lines = append(lines, line)
			mu.Unlock()
		},
		func(err error) {
			close(doneCh)
		},
	)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out")
	}

	mu.Lock()
	defer mu.Unlock()

	if len(lines) != 1 || lines[0] != "stderr_output" {
		t.Errorf("expected stderr to be captured, got %v", lines)
	}
}

func TestBackgroundRunner_Start_WorkingDirectory(t *testing.T) {
	runner := NewBackgroundRunner()

	var mu sync.Mutex
	var lines []string
	doneCh := make(chan struct{})

	err := runner.Start("pwd", "/tmp", LeftPane,
		func(line string) {
			mu.Lock()
			lines = append(lines, line)
			mu.Unlock()
		},
		func(err error) {
			close(doneCh)
		},
	)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out")
	}

	mu.Lock()
	defer mu.Unlock()

	if len(lines) == 0 || !strings.HasPrefix(lines[0], "/tmp") {
		t.Errorf("expected working directory /tmp, got %v", lines)
	}
}

func TestBackgroundRunner_Cancel(t *testing.T) {
	runner := NewBackgroundRunner()

	doneCh := make(chan error, 1)

	err := runner.Start("sleep 60", "/tmp", LeftPane,
		func(line string) {},
		func(err error) {
			doneCh <- err
		},
	)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give the process time to start
	time.Sleep(100 * time.Millisecond)

	runner.Cancel()

	select {
	case <-doneCh:
		// Process was cancelled
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for cancellation")
	}

	if runner.IsRunning() {
		t.Error("runner should not be running after cancel")
	}
}

func TestBackgroundRunner_ConcurrentStartRejection(t *testing.T) {
	runner := NewBackgroundRunner()

	doneCh := make(chan struct{})

	err := runner.Start("sleep 60", "/tmp", LeftPane,
		func(line string) {},
		func(err error) {
			close(doneCh)
		},
	)
	if err != nil {
		t.Fatalf("first Start failed: %v", err)
	}

	// Second start should fail
	err = runner.Start("echo second", "/tmp", RightPane,
		func(line string) {},
		func(err error) {},
	)
	if err == nil {
		t.Error("expected error from concurrent Start, got nil")
	}

	runner.Cancel()
	<-doneCh
}

func TestBackgroundRunner_NotRunningAfterCompletion(t *testing.T) {
	runner := NewBackgroundRunner()

	doneCh := make(chan struct{})

	err := runner.Start("echo done", "/tmp", LeftPane,
		func(line string) {},
		func(err error) {
			close(doneCh)
		},
	)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out")
	}

	// Give a moment for internal cleanup
	time.Sleep(50 * time.Millisecond)

	if runner.IsRunning() {
		t.Error("runner should not be running after completion")
	}
}

func TestBackgroundRunner_FailedCommand(t *testing.T) {
	runner := NewBackgroundRunner()

	doneCh := make(chan error, 1)

	err := runner.Start("exit 1", "/tmp", LeftPane,
		func(line string) {},
		func(err error) {
			doneCh <- err
		},
	)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	select {
	case doneErr := <-doneCh:
		if doneErr == nil {
			t.Error("expected error for failed command")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out")
	}
}
