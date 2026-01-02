package archive

import (
	"context"
	"fmt"
	"sync"
)

// TaskFunc is a function that performs an archive operation
type TaskFunc func(ctx context.Context, progress chan<- *ProgressUpdate) error

// TaskState represents the state of a task
type TaskState int

const (
	TaskStatePending TaskState = iota
	TaskStateRunning
	TaskStateCompleted
	TaskStateCancelled
	TaskStateFailed
)

// TaskStatus represents the status of a task
type TaskStatus struct {
	ID        string
	Running   bool
	Completed bool
	Cancelled bool
	Error     error
	State     TaskState
	Progress  *ProgressUpdate
}

// TaskManager manages background archive tasks
type TaskManager struct {
	tasks  map[string]*taskInfo
	mu     sync.RWMutex
	nextID int
}

type taskInfo struct {
	id       string
	cancel   context.CancelFunc
	status   *TaskStatus
	done     chan struct{}
	progress *ProgressUpdate
}

// NewTaskManager creates a new TaskManager instance
func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks:  make(map[string]*taskInfo),
		nextID: 1,
	}
}

// StartTask starts a new background task and returns its ID
func (tm *TaskManager) StartTask(name string, fn TaskFunc) string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Generate task ID
	taskID := fmt.Sprintf("%s-%d", name, tm.nextID)
	tm.nextID++

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create task info
	info := &taskInfo{
		id:     taskID,
		cancel: cancel,
		status: &TaskStatus{
			ID:        taskID,
			Running:   true,
			Completed: false,
			Cancelled: false,
			Error:     nil,
			State:     TaskStateRunning,
			Progress:  nil,
		},
		done: make(chan struct{}),
	}

	tm.tasks[taskID] = info

	// Start task in goroutine
	go tm.runTask(info, ctx, fn)

	return taskID
}

// runTask executes the task function with panic recovery
func (tm *TaskManager) runTask(info *taskInfo, ctx context.Context, fn TaskFunc) {
	defer close(info.done)

	// Create progress channel
	progress := make(chan *ProgressUpdate, 10)

	// Start goroutine to consume progress updates
	go func() {
		for p := range progress {
			tm.mu.Lock()
			info.progress = p
			info.status.Progress = p
			tm.mu.Unlock()
		}
	}()

	// Execute task with panic recovery
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Convert panic to error
				err = NewArchiveError(ErrArchiveInternalError,
					fmt.Sprintf("Task panicked: %v", r), nil)
			}
		}()
		err = fn(ctx, progress)
	}()

	// Close progress channel when task completes
	close(progress)

	// Update status
	tm.mu.Lock()
	defer tm.mu.Unlock()

	info.status.Running = false
	info.status.Completed = true
	info.status.Error = err

	// Determine final state
	if ctx.Err() == context.Canceled {
		info.status.Cancelled = true
		info.status.State = TaskStateCancelled
	} else if err != nil {
		info.status.State = TaskStateFailed
	} else {
		info.status.State = TaskStateCompleted
	}
}

// CancelTask cancels a running task
func (tm *TaskManager) CancelTask(taskID string) error {
	tm.mu.RLock()
	info, exists := tm.tasks[taskID]
	tm.mu.RUnlock()

	if !exists {
		return NewArchiveError(ErrArchiveInternalError, "Task not found", nil)
	}

	// Cancel the context
	info.cancel()

	return nil
}

// GetTaskStatus returns the status of a task
func (tm *TaskManager) GetTaskStatus(taskID string) *TaskStatus {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	info, exists := tm.tasks[taskID]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	status := &TaskStatus{
		ID:        info.status.ID,
		Running:   info.status.Running,
		Completed: info.status.Completed,
		Cancelled: info.status.Cancelled,
		Error:     info.status.Error,
		State:     info.status.State,
	}

	// Copy progress if available
	if info.progress != nil {
		status.Progress = &ProgressUpdate{
			ProcessedFiles: info.progress.ProcessedFiles,
			TotalFiles:     info.progress.TotalFiles,
			ProcessedBytes: info.progress.ProcessedBytes,
			TotalBytes:     info.progress.TotalBytes,
			CurrentFile:    info.progress.CurrentFile,
			StartTime:      info.progress.StartTime,
			Operation:      info.progress.Operation,
			ArchivePath:    info.progress.ArchivePath,
		}
	}

	return status
}

// WaitForTask waits for a task to complete
func (tm *TaskManager) WaitForTask(taskID string) {
	tm.mu.RLock()
	info, exists := tm.tasks[taskID]
	tm.mu.RUnlock()

	if !exists {
		return
	}

	<-info.done
}

// CleanupTask removes a completed task from the manager
func (tm *TaskManager) CleanupTask(taskID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.tasks, taskID)
}
