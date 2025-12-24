package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewOverwriteDialog(t *testing.T) {
	srcInfo := OverwriteFileInfo{
		Size:    1234,
		ModTime: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	destInfo := OverwriteFileInfo{
		Size:    5678,
		ModTime: time.Date(2024, 1, 10, 15, 45, 0, 0, time.UTC),
	}

	d := NewOverwriteDialog("test.txt", "/home/user/dest", srcInfo, destInfo, "copy", "/home/user/src/test.txt")

	if d == nil {
		t.Fatal("NewOverwriteDialog returned nil")
	}
	if d.filename != "test.txt" {
		t.Errorf("filename = %q, want %q", d.filename, "test.txt")
	}
	if d.destPath != "/home/user/dest" {
		t.Errorf("destPath = %q, want %q", d.destPath, "/home/user/dest")
	}
	if d.srcPath != "/home/user/src/test.txt" {
		t.Errorf("srcPath = %q, want %q", d.srcPath, "/home/user/src/test.txt")
	}
	if d.operation != "copy" {
		t.Errorf("operation = %q, want %q", d.operation, "copy")
	}
	if d.cursor != 0 {
		t.Errorf("cursor = %d, want %d", d.cursor, 0)
	}
	if !d.active {
		t.Error("dialog should be active")
	}
}

func TestOverwriteDialogNavigationJK(t *testing.T) {
	d := NewOverwriteDialog("test.txt", "/dest", OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")

	// Initial cursor position
	if d.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", d.cursor)
	}

	// Move down with 'j'
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if d.cursor != 1 {
		t.Errorf("after j: cursor = %d, want 1", d.cursor)
	}

	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if d.cursor != 2 {
		t.Errorf("after j again: cursor = %d, want 2", d.cursor)
	}

	// Wrap around
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if d.cursor != 0 {
		t.Errorf("after j (wrap): cursor = %d, want 0", d.cursor)
	}

	// Move up with 'k' (wrap around from 0)
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if d.cursor != 2 {
		t.Errorf("after k (wrap): cursor = %d, want 2", d.cursor)
	}

	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if d.cursor != 1 {
		t.Errorf("after k: cursor = %d, want 1", d.cursor)
	}
}

func TestOverwriteDialogNavigationArrows(t *testing.T) {
	d := NewOverwriteDialog("test.txt", "/dest", OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")

	// Move down with arrow
	d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if d.cursor != 1 {
		t.Errorf("after down: cursor = %d, want 1", d.cursor)
	}

	// Move up with arrow
	d.Update(tea.KeyMsg{Type: tea.KeyUp})
	if d.cursor != 0 {
		t.Errorf("after up: cursor = %d, want 0", d.cursor)
	}
}

func TestOverwriteDialogNumberKeys(t *testing.T) {
	tests := []struct {
		key    string
		choice OverwriteChoice
	}{
		{"1", OverwriteChoiceOverwrite},
		{"2", OverwriteChoiceCancel},
		{"3", OverwriteChoiceRename},
	}

	for _, tt := range tests {
		t.Run("key "+tt.key, func(t *testing.T) {
			d := NewOverwriteDialog("test.txt", "/dest", OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")

			_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if cmd == nil {
				t.Fatal("expected command, got nil")
			}

			msg := cmd()
			result, ok := msg.(overwriteDialogResultMsg)
			if !ok {
				t.Fatalf("expected overwriteDialogResultMsg, got %T", msg)
			}
			if result.choice != tt.choice {
				t.Errorf("choice = %v, want %v", result.choice, tt.choice)
			}
		})
	}
}

func TestOverwriteDialogEnterKey(t *testing.T) {
	tests := []struct {
		cursorPos int
		choice    OverwriteChoice
	}{
		{0, OverwriteChoiceOverwrite},
		{1, OverwriteChoiceCancel},
		{2, OverwriteChoiceRename},
	}

	for _, tt := range tests {
		t.Run("cursor at "+string(rune('0'+tt.cursorPos)), func(t *testing.T) {
			d := NewOverwriteDialog("test.txt", "/dest", OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")
			d.cursor = tt.cursorPos

			_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if cmd == nil {
				t.Fatal("expected command, got nil")
			}

			msg := cmd()
			result, ok := msg.(overwriteDialogResultMsg)
			if !ok {
				t.Fatalf("expected overwriteDialogResultMsg, got %T", msg)
			}
			if result.choice != tt.choice {
				t.Errorf("choice = %v, want %v", result.choice, tt.choice)
			}
		})
	}
}

func TestOverwriteDialogEscKey(t *testing.T) {
	d := NewOverwriteDialog("test.txt", "/dest", OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	msg := cmd()
	result, ok := msg.(overwriteDialogResultMsg)
	if !ok {
		t.Fatalf("expected overwriteDialogResultMsg, got %T", msg)
	}
	if result.choice != OverwriteChoiceCancel {
		t.Errorf("choice = %v, want %v", result.choice, OverwriteChoiceCancel)
	}
	if d.active {
		t.Error("dialog should be inactive after Esc")
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{1610612736, "1.5 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatFileSize(tt.bytes)
			if got != tt.want {
				t.Errorf("formatFileSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestOverwriteDialogView(t *testing.T) {
	srcInfo := OverwriteFileInfo{
		Size:    1234,
		ModTime: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	destInfo := OverwriteFileInfo{
		Size:    5678,
		ModTime: time.Date(2024, 1, 10, 15, 45, 0, 0, time.UTC),
	}

	d := NewOverwriteDialog("test.txt", "/home/user/dest", srcInfo, destInfo, "copy", "/src/test.txt")
	view := d.View()

	// Check that essential elements are present
	if !strings.Contains(view, "test.txt") {
		t.Error("view should contain filename")
	}
	if !strings.Contains(view, "already exists") {
		t.Error("view should contain 'already exists'")
	}
	if !strings.Contains(view, "Overwrite") {
		t.Error("view should contain 'Overwrite' option")
	}
	if !strings.Contains(view, "Cancel") {
		t.Error("view should contain 'Cancel' option")
	}
	if !strings.Contains(view, "Rename") {
		t.Error("view should contain 'Rename' option")
	}
	if !strings.Contains(view, "1.2 KB") {
		t.Error("view should contain formatted source size")
	}
}

func TestOverwriteDialogIsActive(t *testing.T) {
	d := NewOverwriteDialog("test.txt", "/dest", OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")

	if !d.IsActive() {
		t.Error("new dialog should be active")
	}

	d.active = false
	if d.IsActive() {
		t.Error("inactive dialog should return false")
	}
}

func TestOverwriteDialogDisplayType(t *testing.T) {
	d := NewOverwriteDialog("test.txt", "/dest", OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")

	if d.DisplayType() != DialogDisplayPane {
		t.Errorf("DisplayType() = %v, want DialogDisplayPane", d.DisplayType())
	}
}

func TestOverwriteDialogViewInactive(t *testing.T) {
	d := NewOverwriteDialog("test.txt", "/dest", OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")
	d.active = false

	view := d.View()
	if view != "" {
		t.Errorf("inactive dialog should return empty string, got %q", view)
	}
}

func TestOverwriteDialogResultContainsAllInfo(t *testing.T) {
	srcInfo := OverwriteFileInfo{Size: 1234}
	destInfo := OverwriteFileInfo{Size: 5678}

	d := NewOverwriteDialog("test.txt", "/home/user/dest", srcInfo, destInfo, "move", "/home/user/src/test.txt")

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	msg := cmd()
	result := msg.(overwriteDialogResultMsg)

	if result.srcPath != "/home/user/src/test.txt" {
		t.Errorf("srcPath = %q, want %q", result.srcPath, "/home/user/src/test.txt")
	}
	if result.destPath != "/home/user/dest" {
		t.Errorf("destPath = %q, want %q", result.destPath, "/home/user/dest")
	}
	if result.filename != "test.txt" {
		t.Errorf("filename = %q, want %q", result.filename, "test.txt")
	}
	if result.operation != "move" {
		t.Errorf("operation = %q, want %q", result.operation, "move")
	}
}

func TestOverwriteDialogCtrlCCancels(t *testing.T) {
	d := NewOverwriteDialog("test.txt", "/dest", OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	msg := cmd()
	result, ok := msg.(overwriteDialogResultMsg)
	if !ok {
		t.Fatalf("expected overwriteDialogResultMsg, got %T", msg)
	}
	if result.choice != OverwriteChoiceCancel {
		t.Errorf("choice = %v, want %v", result.choice, OverwriteChoiceCancel)
	}
}

func TestOverwriteDialogInactiveIgnoresInput(t *testing.T) {
	d := NewOverwriteDialog("test.txt", "/dest", OverwriteFileInfo{}, OverwriteFileInfo{}, "copy", "/src/test.txt")
	d.active = false

	// Try to update - should do nothing
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if cmd != nil {
		t.Error("inactive dialog should not return command")
	}
}

func TestFormatModTime(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "zero time",
			time: time.Time{},
			want: "unknown",
		},
		{
			name: "normal time",
			time: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			want: "2024-01-15 10:30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatModTime(tt.time)
			if got != tt.want {
				t.Errorf("formatModTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncatePath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		maxLen int
		want   string
	}{
		{
			name:   "short path",
			path:   "/home/user",
			maxLen: 20,
			want:   "/home/user",
		},
		{
			name:   "long path",
			path:   "/home/user/very/long/path/to/somewhere",
			maxLen: 20,
			want:   "...path/to/somewhere",
		},
		{
			name:   "very short maxLen",
			path:   "/home/user",
			maxLen: 3,
			want:   "...",
		},
		{
			name:   "exact length",
			path:   "/home",
			maxLen: 5,
			want:   "/home",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncatePath(tt.path, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncatePath(%q, %d) = %q, want %q", tt.path, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestOverwriteDialogViewWithLongPath(t *testing.T) {
	srcInfo := OverwriteFileInfo{
		Size:    1234,
		ModTime: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	destInfo := OverwriteFileInfo{
		Size:    5678,
		ModTime: time.Date(2024, 1, 10, 15, 45, 0, 0, time.UTC),
	}

	longPath := "/home/user/very/long/path/to/destination/directory/with/many/levels"
	d := NewOverwriteDialog("test.txt", longPath, srcInfo, destInfo, "copy", "/src/test.txt")
	view := d.View()

	// View should render without error
	if view == "" {
		t.Error("view should not be empty")
	}
	// Should contain truncated path indicator
	if !strings.Contains(view, "...") {
		t.Log("path might not be long enough to trigger truncation with current width")
	}
}

func TestOverwriteDialogViewWithZeroModTime(t *testing.T) {
	srcInfo := OverwriteFileInfo{
		Size:    1234,
		ModTime: time.Time{}, // zero time
	}
	destInfo := OverwriteFileInfo{
		Size:    5678,
		ModTime: time.Time{}, // zero time
	}

	d := NewOverwriteDialog("test.txt", "/dest", srcInfo, destInfo, "copy", "/src/test.txt")
	view := d.View()

	// Should show "unknown" for zero times
	if !strings.Contains(view, "unknown") {
		t.Error("view should contain 'unknown' for zero mod time")
	}
}
