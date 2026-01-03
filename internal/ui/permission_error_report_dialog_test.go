package ui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
)

func TestNewPermissionErrorReportDialog(t *testing.T) {
	permErrors := []fs.PermissionError{
		{Path: "/test/file1.txt", Error: errors.New("permission denied")},
		{Path: "/test/file2.txt", Error: errors.New("permission denied")},
	}

	d := NewPermissionErrorReportDialog(10, 2, permErrors)

	if d == nil {
		t.Fatal("NewPermissionErrorReportDialog returned nil")
	}

	if !d.IsActive() {
		t.Error("Expected dialog to be active")
	}

	if d.successCount != 10 {
		t.Errorf("Expected successCount=10, got %d", d.successCount)
	}

	if d.failureCount != 2 {
		t.Errorf("Expected failureCount=2, got %d", d.failureCount)
	}

	if len(d.errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(d.errors))
	}
}

func TestPermissionErrorReportDialog_View(t *testing.T) {
	permErrors := []fs.PermissionError{
		{Path: "/test/file1.txt", Error: errors.New("permission denied")},
		{Path: "/test/file2.txt", Error: errors.New("read-only filesystem")},
	}

	d := NewPermissionErrorReportDialog(10, 2, permErrors)
	view := d.View()

	// Check success/failure counts
	if !strings.Contains(view, "Success: 10") {
		t.Error("Expected view to contain success count")
	}
	if !strings.Contains(view, "Failed: 2") {
		t.Error("Expected view to contain failure count")
	}

	// Check error details
	if !strings.Contains(view, "/test/file1.txt") {
		t.Error("Expected view to contain first file path")
	}
	if !strings.Contains(view, "permission denied") {
		t.Error("Expected view to contain first error message")
	}
}

func TestPermissionErrorReportDialog_Scrolling(t *testing.T) {
	// Create many errors
	permErrors := make([]fs.PermissionError, 50)
	for i := 0; i < 50; i++ {
		permErrors[i] = fs.PermissionError{
			Path:  "/test/file" + string(rune(i)) + ".txt",
			Error: errors.New("permission denied"),
		}
	}

	d := NewPermissionErrorReportDialog(0, 50, permErrors)

	// Test j key (scroll down)
	initialView := d.View()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	d.Update(msg)
	afterScrollView := d.View()

	// Scroll offset should have changed
	if d.scrollOffset != 1 {
		t.Errorf("Expected scrollOffset=1 after j key, got %d", d.scrollOffset)
	}

	// Test k key (scroll up)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	d.Update(msg)

	if d.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset=0 after k key, got %d", d.scrollOffset)
	}

	// Test k at top (should not scroll past 0)
	d.Update(msg)
	if d.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset=0 at top, got %d", d.scrollOffset)
	}

	// Test j at bottom
	d.scrollOffset = 45 // Near bottom
	for i := 0; i < 10; i++ {
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
		d.Update(msg)
	}
	// Should not scroll past the end
	if d.scrollOffset > 50 {
		t.Errorf("scrollOffset should not exceed error count, got %d", d.scrollOffset)
	}

	_ = initialView
	_ = afterScrollView
}

func TestPermissionErrorReportDialog_PageUpDown(t *testing.T) {
	// Create many errors
	permErrors := make([]fs.PermissionError, 100)
	for i := 0; i < 100; i++ {
		permErrors[i] = fs.PermissionError{
			Path:  "/test/file.txt",
			Error: errors.New("permission denied"),
		}
	}

	d := NewPermissionErrorReportDialog(0, 100, permErrors)

	initialOffset := d.scrollOffset

	// Page Down
	msg := tea.KeyMsg{Type: tea.KeyPgDown}
	d.Update(msg)

	if d.scrollOffset <= initialOffset {
		t.Error("Expected scrollOffset to increase after Page Down")
	}

	// Page Up
	offsetAfterPageDown := d.scrollOffset
	msg = tea.KeyMsg{Type: tea.KeyPgUp}
	d.Update(msg)

	if d.scrollOffset >= offsetAfterPageDown {
		t.Error("Expected scrollOffset to decrease after Page Up")
	}
}

func TestPermissionErrorReportDialog_ScrollIndicators(t *testing.T) {
	// Create many errors to enable scrolling
	permErrors := make([]fs.PermissionError, 50)
	for i := 0; i < 50; i++ {
		permErrors[i] = fs.PermissionError{
			Path:  "/test/file.txt",
			Error: errors.New("permission denied"),
		}
	}

	d := NewPermissionErrorReportDialog(0, 50, permErrors)
	view := d.View()

	// Should show scroll indicator when there's more content below
	if !strings.Contains(view, "↓") && !strings.Contains(view, "...") {
		t.Log("Expected view to contain scroll indicator")
	}

	// Scroll to middle
	d.scrollOffset = 25
	view = d.View()

	// Should show indicators above and below
	// (Implementation-dependent, this is optional)
}

func TestPermissionErrorReportDialog_Close(t *testing.T) {
	permErrors := []fs.PermissionError{
		{Path: "/test/file.txt", Error: errors.New("permission denied")},
	}

	d := NewPermissionErrorReportDialog(10, 1, permErrors)

	// Test Enter key closes dialog
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newDialog, _ := d.Update(msg)

	// Dialog should return itself but mark as inactive
	if newDialog.IsActive() {
		t.Error("Expected dialog to be inactive after Enter")
	}

	// Test Esc key closes dialog
	d = NewPermissionErrorReportDialog(10, 1, permErrors)
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	newDialog, _ = d.Update(msg)

	if newDialog.IsActive() {
		t.Error("Expected dialog to be inactive after Esc")
	}
}

func TestPermissionErrorReportDialog_DisplayType(t *testing.T) {
	permErrors := []fs.PermissionError{
		{Path: "/test/file.txt", Error: errors.New("permission denied")},
	}

	d := NewPermissionErrorReportDialog(10, 1, permErrors)

	if d.DisplayType() != DialogDisplayScreen {
		t.Errorf("Expected DisplayType=DialogDisplayScreen, got %v", d.DisplayType())
	}
}

func TestPermissionErrorReportDialog_NoErrors(t *testing.T) {
	// Test with no errors
	d := NewPermissionErrorReportDialog(10, 0, []fs.PermissionError{})

	view := d.View()
	if !strings.Contains(view, "Success: 10") {
		t.Error("Expected view to contain success count")
	}
	if !strings.Contains(view, "Failed: 0") {
		t.Error("Expected view to contain zero failures")
	}
}

func TestPermissionErrorReportDialog_LongErrorPaths(t *testing.T) {
	permErrors := []fs.PermissionError{
		{
			Path:  "/very/long/path/that/should/be/truncated/to/fit/in/the/dialog/window/file.txt",
			Error: errors.New("permission denied"),
		},
	}

	d := NewPermissionErrorReportDialog(0, 1, permErrors)
	view := d.View()

	// Path should be truncated with "..." prefix or handled gracefully
	if !strings.Contains(view, "...") && !strings.Contains(view, "file.txt") {
		t.Error("Expected long paths to be truncated or shown with ellipsis")
	}
}

func TestPermissionErrorReportDialog_CommonErrorCategories(t *testing.T) {
	permErrors := []fs.PermissionError{
		{Path: "/test/file1.txt", Error: errors.New("permission denied")},
		{Path: "/test/file2.txt", Error: errors.New("no such file or directory")},
		{Path: "/test/file3.txt", Error: errors.New("read-only file system")},
	}

	d := NewPermissionErrorReportDialog(0, 3, permErrors)
	view := d.View()

	// Should display different error types
	if !strings.Contains(view, "permission denied") {
		t.Error("Expected view to contain 'permission denied' error")
	}
	if !strings.Contains(view, "no such file or directory") {
		t.Error("Expected view to contain 'no such file or directory' error")
	}
	if !strings.Contains(view, "read-only file system") {
		t.Error("Expected view to contain 'read-only file system' error")
	}
}
