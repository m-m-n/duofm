package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/fs"
)

// TestNewContextMenuDialog tests dialog creation
func TestNewContextMenuDialog(t *testing.T) {
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
			wantItems:  3, // copy, move, delete
		},
		{
			name: "directory",
			entry: &fs.FileEntry{
				Name:  "testdir",
				IsDir: true,
			},
			sourcePath: "/source",
			destPath:   "/dest",
			wantItems:  3, // copy, move, delete
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
			wantItems:  5, // copy, move, delete, enter_logical, enter_physical
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
			wantItems:  3, // copy, move, delete (no enter_physical for broken symlink)
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
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	if len(dialog.items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(dialog.items))
	}

	// Check item IDs
	expectedIDs := []string{"copy", "move", "delete"}
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

	if len(dialog.items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(dialog.items))
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

	if len(items) != 3 {
		t.Errorf("getCurrentPageItems returned %d items, want 3", len(items))
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

	// Press 'j' at last item - should wrap to 0
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 0 {
		t.Errorf("after 'j' at last item, cursor = %d, want 0 (wrap)", dialog.cursor)
	}

	// Press 'k' to move up - should wrap to last item
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 2 {
		t.Errorf("after 'k' at first item, cursor = %d, want 2 (wrap)", dialog.cursor)
	}

	// Press 'k' to move up
	updatedDialog, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	dialog = updatedDialog.(*ContextMenuDialog)

	if dialog.cursor != 1 {
		t.Errorf("after 'k', cursor = %d, want 1", dialog.cursor)
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
		{"1", true},  // Valid item
		{"2", true},  // Valid item
		{"3", true},  // Valid item
		{"4", false}, // Invalid (only 3 items)
		{"9", false}, // Invalid (only 3 items)
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

	// Check actionID is set correctly (first item is "copy")
	if result.actionID != "copy" {
		t.Errorf("result.actionID = %s, want 'copy'", result.actionID)
	}
}

// TestUpdate_Enter_Delete tests that delete action returns correct actionID
func TestUpdate_Enter_Delete(t *testing.T) {
	entry := &fs.FileEntry{
		Name:  "test.txt",
		IsDir: false,
	}

	dialog := NewContextMenuDialog(entry, "/source", "/dest")

	// Move to delete item (index 2)
	dialog.cursor = 2

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
		{"1", "copy"},
		{"2", "move"},
		{"3", "delete"},
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
	dialog := &ContextMenuDialog{
		items:        make([]MenuItem, 20), // 20 items, more than one page
		cursor:       0,
		currentPage:  0,
		itemsPerPage: 9,
		active:       true,
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
	dialog := &ContextMenuDialog{
		items:        make([]MenuItem, 20), // 20 items, more than one page
		cursor:       0,
		currentPage:  0,
		itemsPerPage: 9,
		active:       true,
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
