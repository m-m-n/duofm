package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
)

// TestNewContextMenuDialog tests dialog creation
func TestNewContextMenuDialog(t *testing.T) {
	// Save original value and restore after test
	originalValue := hasDesktop
	defer func() { hasDesktop = originalValue }()

	// Set desktop environment to true so all items are enabled
	setDesktopEnvironmentForTest(true)

	tests := []struct {
		name       string
		entry      *fs.FileEntry
		sourcePath string
		destPath   string
		wantItems  int // Expected number of menu items
	}{
		{
			name: "regular file",
			entry: &fs.FileEntry{
				Name:  "test.txt",
				IsDir: false,
			},
			sourcePath: "/source",
			destPath:   "/dest",
			wantItems:  8, // open, open_with, copy, move, delete, copy_name, copy_path, compress
		},
		{
			name: "directory",
			entry: &fs.FileEntry{
				Name:  "testdir",
				IsDir: true,
			},
			sourcePath: "/source",
			destPath:   "/dest",
			wantItems:  8, // open, open_with, copy, move, delete, copy_name, copy_path, compress
		},
		{
			name: "symlink directory",
			entry: &fs.FileEntry{
				Name:       "link",
				IsDir:      true,
				IsSymlink:  true,
				LinkTarget: "/target",
				LinkBroken: false,
			},
			sourcePath: "/source",
			destPath:   "/dest",
			wantItems:  10, // open, open_with, copy, move, delete, copy_name, copy_path, compress, enter_logical, enter_physical
		},
		{
			name: "broken symlink",
			entry: &fs.FileEntry{
				Name:       "broken_link",
				IsDir:      false,
				IsSymlink:  true,
				LinkTarget: "/nonexistent",
				LinkBroken: true,
			},
			sourcePath: "/source",
			destPath:   "/dest",
			wantItems:  8, // open, open_with, copy, move, delete, copy_name, copy_path, compress (no enter_physical for broken symlink)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewContextMenuDialog(tt.entry, tt.sourcePath, tt.destPath)

			if dialog == nil {
				t.Fatal("NewContextMenuDialog returned nil")
			}

			if !dialog.IsActive() {
				t.Error("dialog should be active after creation")
			}

			if len(dialog.items) != tt.wantItems {
				t.Errorf("got %d items, want %d items", len(dialog.items), tt.wantItems)
			}

			// With desktop environment enabled, first item (Open) is enabled,
			// so cursor should be at 0
			if dialog.cursor != 0 {
				t.Errorf("initial cursor = %d, want 0", dialog.cursor)
			}

			if dialog.currentPage != 0 {
				t.Errorf("initial currentPage = %d, want 0", dialog.currentPage)
			}
		})
	}
}

// TestBuildMenuItems_RegularFile tests menu items for regular files
func TestBuildMenuItems_RegularFile(t *testing.T) {
	// Save original value and restore after test
	originalValue := hasDesktop
	defer func() { hasDesktop = originalValue }()

	// Set desktop environment to true so Open/Open with items are enabled
	setDesktopEnvironmentForTest(true)

	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	if len(dialog.items) != 8 {
		t.Fatalf("expected 8 items, got %d", len(dialog.items))
	}

	// Check item IDs
	expectedIDs := []string{"open", "open_with", "copy", "move", "delete", "copy_name", "copy_path", "compress"}
	for i, expectedID := range expectedIDs {
		if dialog.items[i].ID != expectedID {
			t.Errorf("item[%d].ID = %s, want %s", i, dialog.items[i].ID, expectedID)
		}
		if !dialog.items[i].Enabled {
			t.Errorf("item[%d] should be enabled", i)
		}
	}
}

// TestBuildMenuItems_Symlink tests menu items for symlink directories
func TestBuildMenuItems_Symlink(t *testing.T) {
	entry := &fs.FileEntry{
		Name:       "link",
		IsDir:      true,
		IsSymlink:  true,
		LinkTarget: "/target",
		LinkBroken: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	if len(dialog.items) != 10 {
		t.Fatalf("expected 10 items, got %d", len(dialog.items))
	}

	// Check that symlink-specific items exist
	foundEnterLogical := false
	foundEnterPhysical := false
	for _, item := range dialog.items {
		if item.ID == "enter_logical" {
			foundEnterLogical = true
			if !item.Enabled {
				t.Error("enter_logical should be enabled")
			}
		}
		if item.ID == "enter_physical" {
			foundEnterPhysical = true
			if !item.Enabled {
				t.Error("enter_physical should be enabled for non-broken symlink")
			}
		}
	}

	if !foundEnterLogical {
		t.Error("enter_logical item not found")
	}
	if !foundEnterPhysical {
		t.Error("enter_physical item not found")
	}
}

// TestBuildMenuItems_BrokenSymlink tests that broken symlinks disable physical navigation
func TestBuildMenuItems_BrokenSymlink(t *testing.T) {
	entry := &fs.FileEntry{
		Name:       "broken_link",
		IsDir:      false,
		IsSymlink:  true,
		LinkTarget: "/nonexistent",
		LinkBroken: true,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	// Broken symlink should not have enter_physical option or it should be disabled
	for _, item := range dialog.items {
		if item.ID == "enter_physical" {
			if item.Enabled {
				t.Error("enter_physical should be disabled for broken symlink")
			}
		}
	}
}

// TestContextMenuDialog_View tests rendering
func TestContextMenuDialog_View(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	view := dialog.View()

	if view == "" {
		t.Error("View() should return non-empty string")
	}

	// Check that view contains menu items
	if !strings.Contains(view, "Copy") {
		t.Error("View should contain 'Copy'")
	}
	if !strings.Contains(view, "Move") {
		t.Error("View should contain 'Move'")
	}
	if !strings.Contains(view, "Delete") {
		t.Error("View should contain 'Delete'")
	}

	// Check for numbering
	if !strings.Contains(view, "1.") {
		t.Error("View should contain numbered items")
	}
}

// TestContextMenuDialog_IsActive tests IsActive method
func TestContextMenuDialog_IsActive(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	if !dialog.IsActive() {
		t.Error("dialog should be active initially")
	}

	// Simulate deactivation
	dialog.active = false

	if dialog.IsActive() {
		t.Error("dialog should not be active after deactivation")
	}

	view := dialog.View()
	if view != "" {
		t.Error("inactive dialog should return empty view")
	}
}

// TestCalculateWidth tests width calculation
func TestCalculateWidth(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	if dialog.width < dialog.minWidth {
		t.Errorf("width %d is less than minWidth %d", dialog.width, dialog.minWidth)
	}

	if dialog.width > dialog.maxWidth {
		t.Errorf("width %d is greater than maxWidth %d", dialog.width, dialog.maxWidth)
	}
}

// TestGetCurrentPageItems tests pagination helpers
func TestGetCurrentPageItems(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	items := dialog.getCurrentPageItems()

	if len(items) != 8 {
		t.Errorf("getCurrentPageItems returned %d items, want 8", len(items))
	}

	// All items should be on first page
	if dialog.currentPage != 0 {
		t.Errorf("currentPage = %d, want 0", dialog.currentPage)
	}
}

// TestGetTotalPages tests page count calculation
func TestGetTotalPages(t *testing.T) {
	tests := []struct {
		name          string
		itemCount     int
		itemsPerPage  int
		expectedPages int
	}{
		{"less than one page", 3, 9, 1},
		{"exactly one page", 9, 9, 1},
		{"two pages", 10, 9, 2},
		{"three pages", 20, 9, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := &ContextMenuDialog{
				items:        make([]MenuItem, tt.itemCount),
				itemsPerPage: tt.itemsPerPage,
			}

			pages := dialog.getTotalPages()

			if pages != tt.expectedPages {
				t.Errorf("getTotalPages() = %d, want %d", pages, tt.expectedPages)
			}
		})
	}
}

// TestMenuItem_Structure tests MenuItem struct
func TestMenuItem_Structure(t *testing.T) {
	actionCalled := false
	item := MenuItem{
		ID:    "test",
		Label: "Test Item",
		Action: func() error {
			actionCalled = true
			return nil
		},
		Enabled: true,
	}

	if item.ID != "test" {
		t.Errorf("ID = %s, want test", item.ID)
	}

	if item.Label != "Test Item" {
		t.Errorf("Label = %s, want Test Item", item.Label)
	}

	if !item.Enabled {
		t.Error("item should be enabled")
	}

	// Test action execution
	err := item.Action()
	if err != nil {
		t.Errorf("Action returned error: %v", err)
	}

	if !actionCalled {
		t.Error("Action was not called")
	}
}

// TestUpdate_NavigationJK tests j/k cursor movement
func TestUpdate_NavigationJK(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	// Initial cursor should be at 0
	if dialog.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", dialog.cursor)
	}

	// Press 'j' to move down
	updatedDialog, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 1 {
		t.Errorf("after 'j', cursor = %d, want 1", dialog.cursor)
	}

	// Press 'j' again
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 2 {
		t.Errorf("after second 'j', cursor = %d, want 2", dialog.cursor)
	}

	// Press 'j' again to reach item 3
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 3 {
		t.Errorf("after third 'j', cursor = %d, want 3", dialog.cursor)
	}

	// Press 'j' to continue - should reach item 4
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 4 {
		t.Errorf("after fourth 'j', cursor = %d, want 4", dialog.cursor)
	}

	// Press 'j' to continue - should reach item 5
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 5 {
		t.Errorf("after fifth 'j', cursor = %d, want 5", dialog.cursor)
	}

	// Press 'j' to continue - should reach item 6
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 6 {
		t.Errorf("after sixth 'j', cursor = %d, want 6", dialog.cursor)
	}

	// Press 'j' to continue - should reach item 7
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 7 {
		t.Errorf("after seventh 'j', cursor = %d, want 7", dialog.cursor)
	}

	// Press 'j' at last item - should wrap to 0
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 0 {
		t.Errorf("after 'j' at last item, cursor = %d, want 0 (wrap)", dialog.cursor)
	}

	// Press 'k' to move up - should wrap to last item
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 7 {
		t.Errorf("after 'k' at first item, cursor = %d, want 7 (wrap)", dialog.cursor)
	}

	// Press 'k' to move up
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 6 {
		t.Errorf("after 'k', cursor = %d, want 6", dialog.cursor)
	}

	// Press 'k' again
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 5 {
		t.Errorf("after second 'k', cursor = %d, want 5", dialog.cursor)
	}
}

// TestUpdate_NavigationNumeric tests numeric key (1-9) direct selection
func TestUpdate_NavigationNumeric(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	tests := []struct {
		key         string
		shouldClose bool
	}{
		{"1", true},  // Valid item (open)
		{"2", true},  // Valid item (open_with)
		{"3", true},  // Valid item (copy)
		{"4", true},  // Valid item (move)
		{"5", true},  // Valid item (delete)
		{"6", true},  // Valid item (copy_name)
		{"7", true},  // Valid item (copy_path)
		{"8", true},  // Valid item (compress)
		{"9", false}, // Invalid (only 8 items)
	}

	for _, tt := range tests {
		t.Run("key_"+tt.key, func(t *testing.T) {
			dialog := NewContextMenuDialog(entry, "/source", "/dest")
			updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			dialog = updatedDialog.(*ContextMenuDialog)

			if tt.shouldClose {
				if dialog.IsActive() {
					t.Error("dialog should be closed after valid numeric selection")
				}
				if cmd == nil {
					t.Error("cmd should not be nil for valid selection")
				}
			} else {
				if !dialog.IsActive() {
					t.Error("dialog should still be active for invalid numeric selection")
				}
			}
		})
	}
}

// TestUpdate_Enter tests Enter key action execution
func TestUpdate_Enter(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	// Press Enter on first item
	updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.IsActive() {
		t.Error("dialog should be closed after Enter")
	}

	if cmd == nil {
		t.Error("cmd should not be nil after Enter")
	}

	// Execute the command to get the result message
	msg := cmd()
	if msg == nil {
		t.Fatal("cmd() returned nil message")
	}

	result, ok := msg.(contextMenuResultMsg)
	if !ok {
		t.Fatal("cmd() did not return contextMenuResultMsg")
	}

	if result.cancelled {
		t.Error("result should not be cancelled after Enter")
	}

	if result.action == nil {
		t.Error("result.action should not be nil after Enter")
	}

	// Check actionID is set correctly (first item is "open")
	if result.actionID != "open" {
		t.Errorf("result.actionID = %s, want 'open'", result.actionID)
	}
}

// TestUpdate_Enter_Delete tests that delete action returns correct actionID
func TestUpdate_Enter_Delete(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	// Move to delete item (index 4: open, open_with, copy, move, delete)
	dialog.cursor = 4

	// Press Enter on delete item
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("cmd should not be nil after Enter")
	}

	msg := cmd()
	result, ok := msg.(contextMenuResultMsg)
	if !ok {
		t.Fatal("cmd() did not return contextMenuResultMsg")
	}

	if result.actionID != "delete" {
		t.Errorf("result.actionID = %s, want 'delete'", result.actionID)
	}
}

// TestUpdate_NumericKey_ActionID tests that numeric key selection returns correct actionID
func TestUpdate_NumericKey_ActionID(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	tests := []struct {
		key        string
		expectedID string
	}{
		{"1", "open"},
		{"2", "open_with"},
		{"3", "copy"},
		{"4", "move"},
		{"5", "delete"},
		{"6", "copy_name"},
		{"7", "copy_path"},
		{"8", "compress"},
	}

	for _, tt := range tests {
		t.Run("key_"+tt.key, func(t *testing.T) {
			dialog := NewContextMenuDialog(entry, "/source", "/dest")
			_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})

			if cmd == nil {
				t.Fatal("cmd should not be nil")
			}

			msg := cmd()
			result, ok := msg.(contextMenuResultMsg)
			if !ok {
				t.Fatal("cmd() did not return contextMenuResultMsg")
			}

			if result.actionID != tt.expectedID {
				t.Errorf("result.actionID = %s, want %s", result.actionID, tt.expectedID)
			}
		})
	}
}

// TestUpdate_Esc tests Esc key cancellation
func TestUpdate_Esc(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	// Press Esc
	updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.IsActive() {
		t.Error("dialog should be closed after Esc")
	}

	if cmd == nil {
		t.Error("cmd should not be nil after Esc")
	}

	// Execute the command to get the result message
	msg := cmd()
	if msg == nil {
		t.Fatal("cmd() returned nil message")
	}

	result, ok := msg.(contextMenuResultMsg)
	if !ok {
		t.Fatal("cmd() did not return contextMenuResultMsg")
	}

	if !result.cancelled {
		t.Error("result should be cancelled after Esc")
	}

	if result.action != nil {
		t.Error("result.action should be nil after cancellation")
	}
}

// TestUpdate_ArrowKeys tests arrow key navigation
func TestUpdate_ArrowKeys(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	// Test down arrow
	updatedDialog, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyDown})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 1 {
		t.Errorf("after down arrow, cursor = %d, want 1", dialog.cursor)
	}

	// Test up arrow
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyUp})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 0 {
		t.Errorf("after up arrow, cursor = %d, want 0", dialog.cursor)
	}
}

// TestUpdate_LeftRightArrowKeys tests left/right arrow key pagination
func TestUpdate_LeftRightArrowKeys(t *testing.T) {
	// Create a dialog with many items to enable pagination
	base := NewBaseDialog(DialogDisplayPane)
	dialog := &ContextMenuDialog{
		BaseDialog:   base,
		items:        make([]MenuItem, 20), // 20 items, more than one page
		cursor:       0,
		currentPage:  0,
		itemsPerPage: 9,
		minWidth:     40,
		maxWidth:     60,
	}

	// Fill items
	for i := range dialog.items {
		dialog.items[i] = MenuItem{
			ID:      "item",
			Label:   "Item",
			Enabled: true,
		}
	}

	// Test right arrow (next page)
	updatedDialog, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyRight})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.currentPage != 1 {
		t.Errorf("after right arrow, currentPage = %d, want 1", dialog.currentPage)
	}

	// Test left arrow (previous page)
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyLeft})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.currentPage != 0 {
		t.Errorf("after left arrow, currentPage = %d, want 0", dialog.currentPage)
	}

	// Test left arrow at first page (should stay at 0)
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyLeft})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.currentPage != 0 {
		t.Errorf("left arrow at first page: currentPage = %d, want 0", dialog.currentPage)
	}
}

// MockPane is a test double for Pane that records ChangeDirectory calls
type MockPane struct {
	LastChangedDir string
	ChangeError    error
}

func (m *MockPane) ChangeDirectory(path string) error {
	m.LastChangedDir = path
	return m.ChangeError
}

// TestEnterPhysical_NavigatesToLinkTarget tests that enter_physical navigates to the link target itself
func TestEnterPhysical_NavigatesToLinkTarget(t *testing.T) {
	tests := []struct {
		name       string
		linkTarget string
		sourcePath string
		wantDir    string
	}{
		{
			name:       "absolute path link target",
			linkTarget: "/usr/share",
			sourcePath: "/home/user",
			wantDir:    "/usr/share", // Should navigate to /usr/share, NOT /usr
		},
		{
			name:       "relative path link target",
			linkTarget: "../share",
			sourcePath: "/usr/bin",
			wantDir:    "/usr/share", // ../share from /usr/bin = /usr/share
		},
		{
			name:       "relative path with dot components",
			linkTarget: "./subdir/../target",
			sourcePath: "/home/user",
			wantDir:    "/home/user/target",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &fs.FileEntry{
				Name:       "testlink",
				IsDir:      true,
				IsSymlink:  true,
				LinkTarget: tt.linkTarget,
				LinkBroken: false,
			}

			mockPane := &MockPane{}
			dialog := NewContextMenuDialogWithMockPane(entry, tt.sourcePath, "/dest", mockPane)

			// Find and execute the enter_physical action
			var enterPhysicalAction func() error
			for _, item := range dialog.items {
				if item.ID == "enter_physical" {
					enterPhysicalAction = item.Action
					break
				}
			}

			if enterPhysicalAction == nil {
				t.Fatal("enter_physical action not found")
			}

			err := enterPhysicalAction()
			if err != nil {
				t.Fatalf("enter_physical action returned error: %v", err)
			}

			if mockPane.LastChangedDir != tt.wantDir {
				t.Errorf("ChangeDirectory called with %q, want %q", mockPane.LastChangedDir, tt.wantDir)
			}
		})
	}
}

// TestEnterPhysical_ChainedSymlink tests that chained symlinks follow one level only
func TestEnterPhysical_ChainedSymlink(t *testing.T) {
	// Setup: link1 -> /tmp/link2 (where link2 is also a symlink)
	// Expected: Navigate to /tmp/link2 (not the final target)
	entry := &fs.FileEntry{
		Name:       "link1",
		IsDir:      true,
		IsSymlink:  true,
		LinkTarget: "/tmp/link2", // This is also a symlink, but we should only follow one level
		LinkBroken: false,
	}

	mockPane := &MockPane{}
	dialog := NewContextMenuDialogWithMockPane(entry, "/home/user", "/dest", mockPane)

	// Find and execute the enter_physical action
	var enterPhysicalAction func() error
	for _, item := range dialog.items {
		if item.ID == "enter_physical" {
			enterPhysicalAction = item.Action
			break
		}
	}

	if enterPhysicalAction == nil {
		t.Fatal("enter_physical action not found")
	}

	err := enterPhysicalAction()
	if err != nil {
		t.Fatalf("enter_physical action returned error: %v", err)
	}

	// Should navigate to /tmp/link2 directly, not follow the chain
	if mockPane.LastChangedDir != "/tmp/link2" {
		t.Errorf("ChangeDirectory called with %q, want %q", mockPane.LastChangedDir, "/tmp/link2")
	}
}

// TestUpdate_HLKeys tests h/l key pagination
func TestUpdate_HLKeys(t *testing.T) {
	// Create a dialog with many items to enable pagination
	base := NewBaseDialog(DialogDisplayPane)
	dialog := &ContextMenuDialog{
		BaseDialog:   base,
		items:        make([]MenuItem, 20), // 20 items, more than one page
		cursor:       0,
		currentPage:  0,
		itemsPerPage: 9,
		minWidth:     40,
		maxWidth:     60,
	}

	// Fill items
	for i := range dialog.items {
		dialog.items[i] = MenuItem{
			ID:      "item",
			Label:   "Item",
			Enabled: true,
		}
	}

	// Test 'l' key (next page)
	updatedDialog, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.currentPage != 1 {
		t.Errorf("after 'l' key, currentPage = %d, want 1", dialog.currentPage)
	}

	// Test 'h' key (previous page)
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.currentPage != 0 {
		t.Errorf("after 'h' key, currentPage = %d, want 0", dialog.currentPage)
	}
}

// TestUpdate_CtrlC tests Ctrl+C cancellation
func TestUpdate_CtrlC(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	// Press Ctrl+C
	updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.IsActive() {
		t.Error("dialog should be closed after Ctrl+C")
	}

	if cmd == nil {
		t.Error("cmd should not be nil after Ctrl+C")
	}

	// Execute the command to get the result message
	msg := cmd()
	if msg == nil {
		t.Fatal("cmd() returned nil message")
	}

	result, ok := msg.(contextMenuResultMsg)
	if !ok {
		t.Fatal("cmd() did not return contextMenuResultMsg")
	}

	if !result.cancelled {
		t.Error("result should be cancelled after Ctrl+C")
	}

	if result.action != nil {
		t.Error("result.action should be nil after cancellation")
	}
}

// TestContextMenuDialog_OpenMenuItemPresent tests that "Open" menu item appears at position 1
func TestContextMenuDialog_OpenMenuItemPresent(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/test/source", "/test/dest")
	items := dialog.items

	if len(items) < 2 {
		t.Fatal("Expected at least 2 menu items, got", len(items))
	}

	// "Open" should be at position 0 (displayed as "1.")
	if items[0].ID != "open" {
		t.Errorf("Expected first item ID to be 'open', got '%s'", items[0].ID)
	}

	if items[0].Label != "Open" {
		t.Errorf("Expected first item label to be 'Open', got '%s'", items[0].Label)
	}
}

// TestContextMenuDialog_OpenEnabledWhenNoFilesMarked tests that "Open" is enabled when markCount == 0
// and desktop environment is available.
func TestContextMenuDialog_OpenEnabledWhenNoFilesMarked(t *testing.T) {
	// Save original value and restore after test
	originalValue := hasDesktop
	defer func() { hasDesktop = originalValue }()

	// Set desktop environment to true
	setDesktopEnvironmentForTest(true)

	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	// Create dialog without marked files
	dialog := NewContextMenuDialog(entry, "/test/source", "/test/dest")

	// "Open" should be at position 0 and enabled
	if !dialog.items[0].Enabled {
		t.Error("Expected 'Open' to be enabled when no files are marked and desktop environment is available")
	}
}

// TestContextMenuDialog_OpenDisabledWhenFilesMarked tests that "Open" is disabled when markCount > 0
func TestContextMenuDialog_OpenDisabledWhenFilesMarked(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	// Create pane with marked files
	pane := &Pane{}
	pane.markedFiles = map[string]bool{
		"file1.txt": true,
		"file2.txt": true,
	}

	dialog := NewContextMenuDialogWithPane(entry, "/test/source", "/test/dest", pane)

	// "Open" should be at position 0 and disabled
	if dialog.items[0].ID != "open" {
		t.Fatalf("Expected first item to be 'open', got '%s'", dialog.items[0].ID)
	}

	if dialog.items[0].Enabled {
		t.Error("Expected 'Open' to be disabled when multiple files are marked")
	}
}

// TestContextMenuDialog_OpenWithMenuItemPresent tests that "Open with ..." menu item appears at position 2
func TestContextMenuDialog_OpenWithMenuItemPresent(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/test/source", "/test/dest")
	items := dialog.items

	if len(items) < 2 {
		t.Fatalf("Expected at least 2 menu items, got %d", len(items))
	}

	// "Open with ..." should be at position 1 (displayed as "2.")
	if items[1].ID != "open_with" {
		t.Errorf("Expected second item ID to be 'open_with', got '%s'", items[1].ID)
	}

	if items[1].Label != "Open with ..." {
		t.Errorf("Expected second item label to be 'Open with ...', got '%s'", items[1].Label)
	}
}

// TestContextMenuDialog_OpenWithEnabledWhenDesktopAvailable tests that "Open with ..." is enabled
// regardless of marked file count when desktop environment is available.
func TestContextMenuDialog_OpenWithEnabledWhenDesktopAvailable(t *testing.T) {
	// Save original value and restore after test
	originalValue := hasDesktop
	defer func() { hasDesktop = originalValue }()

	// Set desktop environment to true
	setDesktopEnvironmentForTest(true)

	tests := []struct {
		name        string
		markedCount int
	}{
		{"no marked files", 0},
		{"one marked file", 1},
		{"multiple marked files", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &fs.FileEntry{
				Name:  "test.txt",
				IsDir: false,
			}

			var pane *Pane
			if tt.markedCount > 0 {
				pane = &Pane{}
				pane.markedFiles = make(map[string]bool)
				for i := range tt.markedCount {
					pane.markedFiles[fmt.Sprintf("file%d.txt", i)] = true
				}
			}

			dialog := NewContextMenuDialogWithPane(entry, "/test/source", "/test/dest", pane)

			// "Open with ..." should be at position 1 and enabled when desktop is available
			if len(dialog.items) < 2 {
				t.Fatalf("Expected at least 2 items, got %d", len(dialog.items))
			}

			if dialog.items[1].ID != "open_with" {
				t.Fatalf("Expected second item to be 'open_with', got '%s'", dialog.items[1].ID)
			}

			if !dialog.items[1].Enabled {
				t.Errorf("Expected 'Open with ...' to be enabled with %d marked files when desktop is available", tt.markedCount)
			}
		})
	}
}

// TestBuildOpenMenuItems_DesktopEnvironment tests that Open/Open with are disabled without desktop environment
func TestBuildOpenMenuItems_DesktopEnvironment(t *testing.T) {
	tests := []struct {
		name                string
		hasDesktop          bool
		markCount           int
		wantOpenEnabled     bool
		wantOpenWithEnabled bool
	}{
		{
			name:                "desktop present, no marks",
			hasDesktop:          true,
			markCount:           0,
			wantOpenEnabled:     true,
			wantOpenWithEnabled: true,
		},
		{
			name:                "desktop present, with marks",
			hasDesktop:          true,
			markCount:           2,
			wantOpenEnabled:     false,
			wantOpenWithEnabled: true,
		},
		{
			name:                "no desktop, no marks",
			hasDesktop:          false,
			markCount:           0,
			wantOpenEnabled:     false,
			wantOpenWithEnabled: false,
		},
		{
			name:                "no desktop, with marks",
			hasDesktop:          false,
			markCount:           2,
			wantOpenEnabled:     false,
			wantOpenWithEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value and restore after test
			originalValue := hasDesktop
			defer func() { hasDesktop = originalValue }()

			// Set test value
			setDesktopEnvironmentForTest(tt.hasDesktop)

			entry := &fs.FileEntry{
				Name:  "test.txt",
				IsDir: false,
			}

			var pane *Pane
			if tt.markCount > 0 {
				pane = &Pane{}
				pane.markedFiles = make(map[string]bool)
				for i := range tt.markCount {
					pane.markedFiles[fmt.Sprintf("file%d.txt", i)] = true
				}
			}

			dialog := NewContextMenuDialogWithPane(entry, "/source", "/dest", pane)

			// Check Open item (index 0)
			if dialog.items[0].ID != "open" {
				t.Fatalf("Expected first item to be 'open', got '%s'", dialog.items[0].ID)
			}
			if dialog.items[0].Enabled != tt.wantOpenEnabled {
				t.Errorf("Open.Enabled = %v, want %v", dialog.items[0].Enabled, tt.wantOpenEnabled)
			}

			// Check Open with item (index 1)
			if dialog.items[1].ID != "open_with" {
				t.Fatalf("Expected second item to be 'open_with', got '%s'", dialog.items[1].ID)
			}
			if dialog.items[1].Enabled != tt.wantOpenWithEnabled {
				t.Errorf("OpenWith.Enabled = %v, want %v", dialog.items[1].Enabled, tt.wantOpenWithEnabled)
			}
		})
	}
}

// TestNavigationSkipsDisabledItems tests that j/k navigation skips disabled menu items (FR3).
func TestNavigationSkipsDisabledItems(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	// Manually set up items with some disabled
	dialog.items = []MenuItem{
		{ID: "item0", Label: "Item 0", Enabled: false}, // disabled
		{ID: "item1", Label: "Item 1", Enabled: false}, // disabled
		{ID: "item2", Label: "Item 2", Enabled: true},  // enabled
		{ID: "item3", Label: "Item 3", Enabled: false}, // disabled
		{ID: "item4", Label: "Item 4", Enabled: true},  // enabled
	}
	dialog.cursor = 2 // Start at first enabled item

	tests := []struct {
		name           string
		key            string
		startCursor    int
		expectedCursor int
	}{
		{
			name:           "move down from enabled to next enabled",
			key:            "j",
			startCursor:    2, // item2 (enabled)
			expectedCursor: 4, // item4 (enabled), skips item3 (disabled)
		},
		{
			name:           "move down wraps to first enabled",
			key:            "j",
			startCursor:    4, // item4 (enabled)
			expectedCursor: 2, // item2 (enabled), wraps and skips item0, item1
		},
		{
			name:           "move up from enabled to previous enabled",
			key:            "k",
			startCursor:    4, // item4 (enabled)
			expectedCursor: 2, // item2 (enabled), skips item3 (disabled)
		},
		{
			name:           "move up wraps to last enabled",
			key:            "k",
			startCursor:    2, // item2 (enabled)
			expectedCursor: 4, // item4 (enabled), wraps and skips disabled items
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.cursor = tt.startCursor

			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			dialog.Update(keyMsg)

			if dialog.cursor != tt.expectedCursor {
				t.Errorf("cursor = %d, want %d", dialog.cursor, tt.expectedCursor)
			}
		})
	}
}

// TestNavigationAllDisabledItems tests navigation when all items are disabled (guard against infinite loop).
func TestNavigationAllDisabledItems(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	// All items disabled
	dialog.items = []MenuItem{
		{ID: "item0", Label: "Item 0", Enabled: false},
		{ID: "item1", Label: "Item 1", Enabled: false},
		{ID: "item2", Label: "Item 2", Enabled: false},
	}
	dialog.cursor = 0

	// Move down - should not infinite loop, cursor stays at 0 or moves to next position
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	dialog.Update(keyMsg)

	// The cursor should have moved (or stayed), but not caused infinite loop
	// When all items are disabled, we allow cursor to move to any position
	if dialog.cursor < 0 || dialog.cursor >= len(dialog.items) {
		t.Errorf("cursor out of bounds: %d", dialog.cursor)
	}
}

// TestFindFirstEnabledItem tests the findFirstEnabledItem helper function.
func TestFindFirstEnabledItem(t *testing.T) {
	tests := []struct {
		name        string
		items       []MenuItem
		expectedIdx int
	}{
		{
			name: "first item enabled",
			items: []MenuItem{
				{ID: "item0", Label: "Item 0", Enabled: true},
				{ID: "item1", Label: "Item 1", Enabled: true},
				{ID: "item2", Label: "Item 2", Enabled: true},
			},
			expectedIdx: 0,
		},
		{
			name: "first two items disabled",
			items: []MenuItem{
				{ID: "item0", Label: "Item 0", Enabled: false},
				{ID: "item1", Label: "Item 1", Enabled: false},
				{ID: "item2", Label: "Item 2", Enabled: true},
				{ID: "item3", Label: "Item 3", Enabled: true},
			},
			expectedIdx: 2,
		},
		{
			name: "only last item enabled",
			items: []MenuItem{
				{ID: "item0", Label: "Item 0", Enabled: false},
				{ID: "item1", Label: "Item 1", Enabled: false},
				{ID: "item2", Label: "Item 2", Enabled: false},
				{ID: "item3", Label: "Item 3", Enabled: true},
			},
			expectedIdx: 3,
		},
		{
			name: "all items disabled - defensive default returns 0",
			items: []MenuItem{
				{ID: "item0", Label: "Item 0", Enabled: false},
				{ID: "item1", Label: "Item 1", Enabled: false},
				{ID: "item2", Label: "Item 2", Enabled: false},
			},
			expectedIdx: 0,
		},
		{
			name:        "empty items - defensive default returns 0",
			items:       []MenuItem{},
			expectedIdx: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := NewBaseDialog(DialogDisplayPane)
			dialog := &ContextMenuDialog{
				BaseDialog:   base,
				items:        tt.items,
				currentPage:  0,
				itemsPerPage: 9,
			}

			result := dialog.findFirstEnabledItem()
			if result != tt.expectedIdx {
				t.Errorf("findFirstEnabledItem() = %d, want %d", result, tt.expectedIdx)
			}
		})
	}
}

// TestInitialCursorPositionWithDisabledItems tests that the initial cursor is positioned
// at the first enabled item when desktop environment is unavailable.
func TestInitialCursorPositionWithDisabledItems(t *testing.T) {
	// Save original value and restore after test
	originalValue := hasDesktop
	defer func() { hasDesktop = originalValue }()

	// Set desktop environment to false (simulating headless environment)
	setDesktopEnvironmentForTest(false)

	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	// Without desktop environment:
	// - item 0 (Open) is disabled
	// - item 1 (Open with) is disabled
	// - item 2 (Copy) is enabled
	// So initial cursor should be 2

	if dialog.items[0].ID != "open" || dialog.items[0].Enabled {
		t.Error("Open should be disabled without desktop environment")
	}

	if dialog.items[1].ID != "open_with" || dialog.items[1].Enabled {
		t.Error("Open with should be disabled without desktop environment")
	}

	if dialog.items[2].ID != "copy" || !dialog.items[2].Enabled {
		t.Error("Copy should be enabled")
	}

	// Cursor should be at first enabled item (Copy at index 2)
	if dialog.cursor != 2 {
		t.Errorf("initial cursor = %d, want 2 (first enabled item 'Copy')", dialog.cursor)
	}
}

// TestViewDoesNotHighlightDisabledItem tests that disabled items are not highlighted
// even if cursor happens to be on them.
func TestViewDoesNotHighlightDisabledItem(t *testing.T) {
	base := NewBaseDialog(DialogDisplayPane)
	dialog := &ContextMenuDialog{
		BaseDialog: base,
		items: []MenuItem{
			{ID: "disabled1", Label: "Disabled Item 1", Enabled: false},
			{ID: "enabled1", Label: "Enabled Item 1", Enabled: true},
		},
		cursor:       0, // Force cursor on disabled item
		currentPage:  0,
		itemsPerPage: 9,
		minWidth:     40,
		maxWidth:     60,
	}
	dialog.calculateWidth()
	dialog.styles = NewDialogStyles(dialog.Width(), ColorPrimary)

	view := dialog.View()

	// The view should still render (not crash)
	if view == "" {
		t.Error("View() should return non-empty string")
	}

	// The disabled item should be rendered with muted color (not highlighted)
	// This is a visual test, but we can at least verify the view contains both items
	if !strings.Contains(view, "Disabled Item 1") {
		t.Error("View should contain 'Disabled Item 1'")
	}
	if !strings.Contains(view, "Enabled Item 1") {
		t.Error("View should contain 'Enabled Item 1'")
	}
}

// TestClipboardMenuItems_Presence tests that copy_name and copy_path items exist in the menu.
func TestClipboardMenuItems_Presence(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	foundCopyName := false
	foundCopyPath := false
	for _, item := range dialog.items {
		if item.ID == "copy_name" {
			foundCopyName = true
			if item.Label != "Copy file name" {
				t.Errorf("copy_name label = %q, want %q", item.Label, "Copy file name")
			}
		}
		if item.ID == "copy_path" {
			foundCopyPath = true
			if item.Label != "Copy full path" {
				t.Errorf("copy_path label = %q, want %q", item.Label, "Copy full path")
			}
		}
	}

	if !foundCopyName {
		t.Error("copy_name item not found in context menu")
	}
	if !foundCopyPath {
		t.Error("copy_path item not found in context menu")
	}
}

// TestClipboardMenuItems_Position tests that clipboard items are after delete and before compress.
func TestClipboardMenuItems_Position(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	var deleteIdx, copyNameIdx, copyPathIdx, compressIdx int
	for i, item := range dialog.items {
		switch item.ID {
		case "delete":
			deleteIdx = i
		case "copy_name":
			copyNameIdx = i
		case "copy_path":
			copyPathIdx = i
		case "compress":
			compressIdx = i
		}
	}

	// copy_name should be after delete
	if copyNameIdx <= deleteIdx {
		t.Errorf("copy_name (idx=%d) should be after delete (idx=%d)", copyNameIdx, deleteIdx)
	}

	// copy_path should be after copy_name
	if copyPathIdx <= copyNameIdx {
		t.Errorf("copy_path (idx=%d) should be after copy_name (idx=%d)", copyPathIdx, copyNameIdx)
	}

	// compress should be after copy_path
	if compressIdx <= copyPathIdx {
		t.Errorf("compress (idx=%d) should be after copy_path (idx=%d)", compressIdx, copyPathIdx)
	}
}

// TestClipboardMenuItems_EnabledForRegularFile tests that clipboard items are enabled for regular files.
func TestClipboardMenuItems_EnabledForRegularFile(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	for _, item := range dialog.items {
		if item.ID == "copy_name" && !item.Enabled {
			t.Error("copy_name should be enabled for regular file")
		}
		if item.ID == "copy_path" && !item.Enabled {
			t.Error("copy_path should be enabled for regular file")
		}
	}
}

// TestClipboardMenuItems_EnabledForDirectory tests that clipboard items are enabled for directories.
func TestClipboardMenuItems_EnabledForDirectory(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "testdir",
		IsDir: true,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	for _, item := range dialog.items {
		if item.ID == "copy_name" && !item.Enabled {
			t.Error("copy_name should be enabled for directory")
		}
		if item.ID == "copy_path" && !item.Enabled {
			t.Error("copy_path should be enabled for directory")
		}
	}
}

// TestClipboardMenuItems_DisabledForParentDir tests that clipboard items are disabled for parent directory.
func TestClipboardMenuItems_DisabledForParentDir(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "..",
		IsDir: true,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	for _, item := range dialog.items {
		if item.ID == "copy_name" && item.Enabled {
			t.Error("copy_name should be disabled for parent directory")
		}
		if item.ID == "copy_path" && item.Enabled {
			t.Error("copy_path should be disabled for parent directory")
		}
	}
}

// TestClipboardMenuItems_DisabledWhenMarked tests that clipboard items are disabled when files are marked.
func TestClipboardMenuItems_DisabledWhenMarked(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	pane := &Pane{}
	pane.markedFiles = map[string]bool{
		"file1.txt": true,
		"file2.txt": true,
	}

	dialog := NewContextMenuDialogWithPane(entry, "/source", "/dest", pane)

	for _, item := range dialog.items {
		if item.ID == "copy_name" && item.Enabled {
			t.Error("copy_name should be disabled when files are marked")
		}
		if item.ID == "copy_path" && item.Enabled {
			t.Error("copy_path should be disabled when files are marked")
		}
	}
}

// TestClipboardMenuItems_ActionIsNil tests that clipboard menu items have nil Action.
func TestClipboardMenuItems_ActionIsNil(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	for _, item := range dialog.items {
		if item.ID == "copy_name" && item.Action != nil {
			t.Error("copy_name Action should be nil (handled by Model)")
		}
		if item.ID == "copy_path" && item.Action != nil {
			t.Error("copy_path Action should be nil (handled by Model)")
		}
	}
}
