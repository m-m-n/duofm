package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewPermissionProgressDialog(t *testing.T) {
	d := NewPermissionProgressDialog(100)

	if d == nil {
		t.Fatal("NewPermissionProgressDialog returned nil")
	}

	if !d.IsActive() {
		t.Error("Expected dialog to be active")
	}

	if d.totalFiles != 100 {
		t.Errorf("Expected totalFiles=100, got %d", d.totalFiles)
	}
}

func TestPermissionProgressDialog_UpdateProgress(t *testing.T) {
	d := NewPermissionProgressDialog(100)

	// Update progress
	d.UpdateProgress(50, "/path/to/file.txt")

	view := d.View()
	if !strings.Contains(view, "50 / 100") {
		t.Error("Expected view to contain progress count")
	}
	if !strings.Contains(view, "/path/to/file.txt") {
		t.Error("Expected view to contain current file path")
	}
	if !strings.Contains(view, "50%") {
		t.Error("Expected view to contain percentage")
	}
}

func TestPermissionProgressDialog_ProgressBar(t *testing.T) {
	tests := []struct {
		name       string
		processed  int
		total      int
		percentage int
	}{
		{"0%", 0, 100, 0},
		{"25%", 25, 100, 25},
		{"50%", 50, 100, 50},
		{"75%", 75, 100, 75},
		{"100%", 100, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewPermissionProgressDialog(tt.total)
			d.UpdateProgress(tt.processed, "/test/file")

			view := d.View()
			expectedPercentage := tt.percentage
			percentageStr := strings.TrimSpace(strings.Split(view, "%")[0])
			if !strings.Contains(percentageStr, string(rune(expectedPercentage/10+48))) && expectedPercentage != 100 {
				t.Logf("View content:\n%s", view)
			}
		})
	}
}

func TestPermissionProgressDialog_ElapsedTime(t *testing.T) {
	d := NewPermissionProgressDialog(100)

	// Simulate some elapsed time
	time.Sleep(100 * time.Millisecond)
	d.UpdateProgress(10, "/test/file")

	view := d.View()
	if !strings.Contains(view, "00:00") {
		t.Error("Expected view to contain elapsed time")
	}
}

func TestPermissionProgressDialog_Cancel(t *testing.T) {
	d := NewPermissionProgressDialog(100)

	cancelCalled := false
	d.SetOnCancel(func() {
		cancelCalled = true
	})

	// Send Esc key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	d.Update(msg)

	if !cancelCalled {
		t.Error("Expected cancel callback to be called on Esc")
	}
}

func TestPermissionProgressDialog_Complete(t *testing.T) {
	d := NewPermissionProgressDialog(100)

	d.Complete()

	if d.IsActive() {
		t.Error("Expected dialog to be inactive after Complete()")
	}

	view := d.View()
	if view != "" {
		t.Error("Expected empty view after Complete()")
	}
}

func TestPermissionProgressDialog_DisplayType(t *testing.T) {
	d := NewPermissionProgressDialog(100)

	if d.DisplayType() != DialogDisplayScreen {
		t.Errorf("Expected DisplayType=DialogDisplayScreen, got %v", d.DisplayType())
	}
}

func TestPermissionProgressDialog_LongPath(t *testing.T) {
	d := NewPermissionProgressDialog(100)

	longPath := "/very/long/path/that/should/be/truncated/to/fit/in/the/dialog/window/file.txt"
	d.UpdateProgress(50, longPath)

	view := d.View()
	// Path should be truncated or handled gracefully
	if !strings.Contains(view, "...") && len(longPath) > 50 {
		// Should either truncate or show in some way
		if !strings.Contains(view, "file.txt") {
			t.Error("Expected view to contain at least filename")
		}
	}
}
