package archive

import (
	"context"
	"testing"
	"time"
)

func TestNewTaskManager(t *testing.T) {
	tm := NewTaskManager()
	if tm == nil {
		t.Fatal("NewTaskManager() returned nil")
	}
}

func TestTaskManager_StartTask(t *testing.T) {
	tm := NewTaskManager()

	taskID := tm.StartTask("test-task", func(ctx context.Context, progress chan<- *ProgressUpdate) error {
		// Simulate task
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	if taskID == "" {
		t.Error("StartTask() returned empty task ID")
	}
}

func TestTaskManager_CancelTask(t *testing.T) {
	tm := NewTaskManager()

	taskStarted := make(chan bool)
	taskID := tm.StartTask("cancel-test", func(ctx context.Context, progress chan<- *ProgressUpdate) error {
		taskStarted <- true
		// Wait for context cancellation
		<-ctx.Done()
		return ctx.Err()
	})

	// Wait for task to start
	<-taskStarted

	// Cancel task
	err := tm.CancelTask(taskID)
	if err != nil {
		t.Errorf("CancelTask() error = %v", err)
	}

	// Wait a bit for cancellation to complete
	time.Sleep(100 * time.Millisecond)
}

func TestTaskManager_GetTaskStatus(t *testing.T) {
	tm := NewTaskManager()

	taskID := tm.StartTask("status-test", func(ctx context.Context, progress chan<- *ProgressUpdate) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	status := tm.GetTaskStatus(taskID)
	if status == nil {
		t.Error("GetTaskStatus() returned nil for existing task")
	}
}

func TestTaskManager_CleanupTask(t *testing.T) {
	tm := NewTaskManager()

	// Start a task
	taskID := tm.StartTask("cleanup-test", func(ctx context.Context, progress chan<- *ProgressUpdate) error {
		return nil
	})

	// Wait for task to complete
	tm.WaitForTask(taskID)

	// Verify task exists
	status := tm.GetTaskStatus(taskID)
	if status == nil {
		t.Fatal("GetTaskStatus() returned nil before cleanup")
	}

	// Cleanup the task
	tm.CleanupTask(taskID)

	// Verify task no longer exists
	status = tm.GetTaskStatus(taskID)
	if status != nil {
		t.Error("GetTaskStatus() should return nil after cleanup")
	}
}

func TestTaskManager_CleanupTask_NonExistent(t *testing.T) {
	tm := NewTaskManager()

	// Cleanup a non-existent task should not panic
	tm.CleanupTask("non-existent-task")

	// Verify manager still works
	taskID := tm.StartTask("test", func(ctx context.Context, progress chan<- *ProgressUpdate) error {
		return nil
	})
	if taskID == "" {
		t.Error("StartTask() should work after cleanup of non-existent task")
	}
}
