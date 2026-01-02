package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/archive"
)

func TestNewArchiveProgressDialog(t *testing.T) {
	dialog := NewArchiveProgressDialog("compress", "test.tar.gz")

	if dialog == nil {
		t.Fatal("NewArchiveProgressDialog() returned nil")
	}

	if !dialog.IsActive() {
		t.Error("NewArchiveProgressDialog() should be active by default")
	}

	if dialog.operation != "compress" {
		t.Errorf("NewArchiveProgressDialog() operation = %q, want %q", dialog.operation, "compress")
	}

	if dialog.archivePath != "test.tar.gz" {
		t.Errorf("NewArchiveProgressDialog() archivePath = %q, want %q", dialog.archivePath, "test.tar.gz")
	}
}

func TestArchiveProgressDialog_UpdateProgress(t *testing.T) {
	dialog := NewArchiveProgressDialog("compress", "test.tar.gz")

	progress := &archive.ProgressUpdate{
		ProcessedFiles: 50,
		TotalFiles:     100,
		ProcessedBytes: 1024 * 1024 * 50,  // 50 MB
		TotalBytes:     1024 * 1024 * 100, // 100 MB
		CurrentFile:    "file50.txt",
		StartTime:      time.Now().Add(-10 * time.Second),
		Operation:      "compress",
		ArchivePath:    "test.tar.gz",
	}

	dialog.UpdateProgress(progress)

	if dialog.progress != progress {
		t.Error("UpdateProgress() did not update progress")
	}
}

func TestArchiveProgressDialog_Percentage(t *testing.T) {
	dialog := NewArchiveProgressDialog("compress", "test.tar.gz")

	progress := &archive.ProgressUpdate{
		ProcessedFiles: 75,
		TotalFiles:     100,
		StartTime:      time.Now(),
	}

	dialog.UpdateProgress(progress)

	// The dialog should display the percentage from the progress
	if progress.Percentage() != 75 {
		t.Errorf("Progress percentage = %d, want 75", progress.Percentage())
	}
}

func TestArchiveProgressDialog_Cancel(t *testing.T) {
	cancelCalled := false
	dialog := NewArchiveProgressDialog("compress", "test.tar.gz")
	dialog.SetOnCancel(func() {
		cancelCalled = true
	})

	// Press Escape to cancel
	_, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !cancelCalled {
		t.Error("Cancel callback should be called when Escape is pressed")
	}
}

func TestArchiveProgressDialog_Complete(t *testing.T) {
	dialog := NewArchiveProgressDialog("compress", "test.tar.gz")

	if !dialog.IsActive() {
		t.Error("Dialog should be active initially")
	}

	dialog.Complete()

	if dialog.IsActive() {
		t.Error("Dialog should be inactive after Complete()")
	}
}

func TestArchiveProgressDialog_SetActive(t *testing.T) {
	dialog := NewArchiveProgressDialog("compress", "test.tar.gz")

	dialog.SetActive(false)
	if dialog.IsActive() {
		t.Error("SetActive(false) should deactivate dialog")
	}

	dialog.SetActive(true)
	if !dialog.IsActive() {
		t.Error("SetActive(true) should activate dialog")
	}
}

func TestArchiveProgressDialog_DisplayType(t *testing.T) {
	dialog := NewArchiveProgressDialog("compress", "test.tar.gz")
	if dialog.DisplayType() != DialogDisplayScreen {
		t.Errorf("DisplayType() = %v, want DialogDisplayScreen", dialog.DisplayType())
	}
}

func TestArchiveProgressDialog_View(t *testing.T) {
	t.Run("active dialog with compress operation", func(t *testing.T) {
		dialog := NewArchiveProgressDialog("compress", "test.tar.gz")
		view := dialog.View()

		if view == "" {
			t.Error("View should not be empty for active dialog")
		}

		if !containsStr(view, "Compressing Archive") {
			t.Error("View should contain 'Compressing Archive' for compress operation")
		}

		if !containsStr(view, "test.tar.gz") {
			t.Error("View should contain archive path")
		}

		if !containsStr(view, "Starting...") {
			t.Error("View should contain 'Starting...' when no progress")
		}

		if !containsStr(view, "[Esc] Cancel") {
			t.Error("View should contain cancel hint")
		}
	})

	t.Run("active dialog with extract operation", func(t *testing.T) {
		dialog := NewArchiveProgressDialog("extract", "archive.zip")
		view := dialog.View()

		if !containsStr(view, "Extracting Archive") {
			t.Error("View should contain 'Extracting Archive' for extract operation")
		}
	})

	t.Run("active dialog with progress", func(t *testing.T) {
		dialog := NewArchiveProgressDialog("compress", "test.tar.gz")
		progress := &archive.ProgressUpdate{
			ProcessedFiles: 25,
			TotalFiles:     100,
			ProcessedBytes: 1024 * 1024 * 25,
			TotalBytes:     1024 * 1024 * 100,
			CurrentFile:    "current_file.txt",
			StartTime:      time.Now().Add(-30 * time.Second),
			Operation:      "compress",
			ArchivePath:    "test.tar.gz",
		}
		dialog.UpdateProgress(progress)

		view := dialog.View()

		if !containsStr(view, "25%") {
			t.Error("View should contain percentage")
		}

		if !containsStr(view, "25/100") {
			t.Error("View should contain file count")
		}

		if !containsStr(view, "current_file.txt") {
			t.Error("View should contain current file name")
		}
	})

	t.Run("inactive dialog shows empty", func(t *testing.T) {
		dialog := NewArchiveProgressDialog("compress", "test.tar.gz")
		dialog.Complete()
		view := dialog.View()

		if view != "" {
			t.Error("View should be empty for inactive dialog")
		}
	})

	t.Run("long current file is truncated", func(t *testing.T) {
		dialog := NewArchiveProgressDialog("compress", "test.tar.gz")
		longFileName := "very/long/path/to/a/file/that/exceeds/fifty/characters/in/length.txt"
		progress := &archive.ProgressUpdate{
			ProcessedFiles: 50,
			TotalFiles:     100,
			CurrentFile:    longFileName,
			StartTime:      time.Now().Add(-10 * time.Second),
		}
		dialog.UpdateProgress(progress)

		view := dialog.View()

		// The view should contain a truncated version with "..."
		if !containsStr(view, "...") {
			t.Error("Long file name should be truncated with ...")
		}
	})
}

func TestArchiveProgressDialog_Update_Inactive(t *testing.T) {
	dialog := NewArchiveProgressDialog("compress", "test.tar.gz")
	dialog.Complete()

	updated, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if updated == nil {
		t.Error("Updated dialog should not be nil")
	}

	if cmd != nil {
		t.Error("Command should be nil for inactive dialog")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0, "00:00"},
		{30 * time.Second, "00:30"},
		{60 * time.Second, "01:00"},
		{90 * time.Second, "01:30"},
		{5*time.Minute + 45*time.Second, "05:45"},
		{10*time.Minute + 10*time.Second, "10:10"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %s, want %s", tt.duration, result, tt.expected)
			}
		})
	}
}

// containsStr checks if s contains substr
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
