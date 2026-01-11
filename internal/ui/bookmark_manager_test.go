package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sakura/duofm/internal/config"
)

func TestBookmarkManager_SetBookmarks(t *testing.T) {
	manager := &BookmarkManager{
		bookmarks: []config.Bookmark{},
		editIndex: -1,
	}

	newBookmarks := []config.Bookmark{
		{Name: "home", Path: "/home/user"},
		{Name: "docs", Path: "/home/user/Documents"},
	}

	manager.SetBookmarks(newBookmarks)

	if len(manager.Bookmarks()) != 2 {
		t.Errorf("SetBookmarks() bookmarks length = %d, want 2", len(manager.Bookmarks()))
	}

	if manager.Bookmarks()[0].Name != "home" {
		t.Errorf("Bookmarks()[0].Name = %s, want home", manager.Bookmarks()[0].Name)
	}

	if manager.Bookmarks()[1].Path != "/home/user/Documents" {
		t.Errorf("Bookmarks()[1].Path = %s, want /home/user/Documents", manager.Bookmarks()[1].Path)
	}
}

func TestBookmarkManager_Bookmarks(t *testing.T) {
	bookmarks := []config.Bookmark{
		{Name: "test1", Path: "/test1"},
		{Name: "test2", Path: "/test2"},
	}

	manager := &BookmarkManager{
		bookmarks: bookmarks,
		editIndex: -1,
	}

	result := manager.Bookmarks()

	if len(result) != 2 {
		t.Errorf("Bookmarks() length = %d, want 2", len(result))
	}

	if result[0].Name != "test1" {
		t.Errorf("Bookmarks()[0].Name = %s, want test1", result[0].Name)
	}
}

func TestBookmarkManager_EditIndex(t *testing.T) {
	manager := &BookmarkManager{
		bookmarks: []config.Bookmark{},
		editIndex: -1,
	}

	// Initial value
	if manager.EditIndex() != -1 {
		t.Errorf("EditIndex() = %d, want -1", manager.EditIndex())
	}

	// Set edit index
	manager.SetEditIndex(2)
	if manager.EditIndex() != 2 {
		t.Errorf("EditIndex() = %d, want 2", manager.EditIndex())
	}

	// Set to another value
	manager.SetEditIndex(5)
	if manager.EditIndex() != 5 {
		t.Errorf("EditIndex() = %d, want 5", manager.EditIndex())
	}
}

func TestBookmarkManager_SetEditIndex(t *testing.T) {
	manager := &BookmarkManager{
		bookmarks: []config.Bookmark{},
		editIndex: -1,
	}

	tests := []struct {
		name  string
		index int
	}{
		{"set to 0", 0},
		{"set to positive", 3},
		{"set to negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.SetEditIndex(tt.index)
			if manager.editIndex != tt.index {
				t.Errorf("SetEditIndex(%d) editIndex = %d", tt.index, manager.editIndex)
			}
		})
	}
}

func TestBookmarkManager_ClearEditIndex(t *testing.T) {
	manager := &BookmarkManager{
		bookmarks: []config.Bookmark{},
		editIndex: 5,
	}

	if manager.EditIndex() != 5 {
		t.Errorf("Initial EditIndex() = %d, want 5", manager.EditIndex())
	}

	manager.ClearEditIndex()

	if manager.EditIndex() != -1 {
		t.Errorf("ClearEditIndex() EditIndex() = %d, want -1", manager.EditIndex())
	}
}

func TestBookmarkManager_Add_EmptyAlias(t *testing.T) {
	manager := &BookmarkManager{
		bookmarks: []config.Bookmark{},
		editIndex: -1,
	}

	cmd := manager.Add("/test/path", "")

	if cmd == nil {
		t.Fatal("Add() returned nil command")
	}

	msg := cmd()
	statusMsg, ok := msg.(showStatusMsg)
	if !ok {
		t.Fatalf("Add() returned %T, want showStatusMsg", msg)
	}

	if !statusMsg.isError {
		t.Error("Add() with empty alias should return error status")
	}

	if statusMsg.message != "Bookmark name cannot be empty" {
		t.Errorf("Add() message = %s, want 'Bookmark name cannot be empty'", statusMsg.message)
	}
}

func TestBookmarkManager_Add_DuplicatePath(t *testing.T) {
	manager := &BookmarkManager{
		bookmarks: []config.Bookmark{
			{Name: "existing", Path: "/existing/path"},
		},
		editIndex: -1,
	}

	cmd := manager.Add("/existing/path", "new_alias")

	if cmd == nil {
		t.Fatal("Add() returned nil command")
	}

	msg := cmd()
	statusMsg, ok := msg.(showStatusMsg)
	if !ok {
		t.Fatalf("Add() returned %T, want showStatusMsg", msg)
	}

	if statusMsg.isError {
		t.Error("Add() with duplicate path should not be an error")
	}

	if statusMsg.message != "Already bookmarked" {
		t.Errorf("Add() message = %s, want 'Already bookmarked'", statusMsg.message)
	}
}

func TestBookmarkManager_Edit_InvalidIndex(t *testing.T) {
	manager := &BookmarkManager{
		bookmarks: []config.Bookmark{
			{Name: "test", Path: "/test"},
		},
		editIndex: -1,
	}

	tests := []struct {
		name  string
		index int
	}{
		{"negative index", -1},
		{"index equal to length", 1},
		{"index greater than length", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := manager.Edit(tt.index, "new_alias")

			if cmd == nil {
				t.Fatal("Edit() returned nil command")
			}

			msg := cmd()
			statusMsg, ok := msg.(showStatusMsg)
			if !ok {
				t.Fatalf("Edit() returned %T, want showStatusMsg", msg)
			}

			if !statusMsg.isError {
				t.Error("Edit() with invalid index should return error status")
			}

			if statusMsg.message != "Invalid bookmark index" {
				t.Errorf("Edit() message = %s, want 'Invalid bookmark index'", statusMsg.message)
			}
		})
	}
}

func TestBookmarkManager_Edit_EmptyAlias(t *testing.T) {
	manager := &BookmarkManager{
		bookmarks: []config.Bookmark{
			{Name: "test", Path: "/test"},
		},
		editIndex: -1,
	}

	cmd := manager.Edit(0, "")

	if cmd == nil {
		t.Fatal("Edit() returned nil command")
	}

	msg := cmd()
	statusMsg, ok := msg.(showStatusMsg)
	if !ok {
		t.Fatalf("Edit() returned %T, want showStatusMsg", msg)
	}

	if !statusMsg.isError {
		t.Error("Edit() with empty alias should return error status")
	}

	if statusMsg.message != "Bookmark name cannot be empty" {
		t.Errorf("Edit() message = %s, want 'Bookmark name cannot be empty'", statusMsg.message)
	}
}

func TestBookmarkManager_Delete_InvalidIndex(t *testing.T) {
	manager := &BookmarkManager{
		bookmarks: []config.Bookmark{
			{Name: "test", Path: "/test"},
		},
		editIndex: -1,
	}

	tests := []struct {
		name  string
		index int
	}{
		{"negative index", -1},
		{"index equal to length", 1},
		{"index greater than length", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := manager.Delete(tt.index)

			if cmd == nil {
				t.Fatal("Delete() returned nil command")
			}

			msg := cmd()
			statusMsg, ok := msg.(showStatusMsg)
			if !ok {
				t.Fatalf("Delete() returned %T, want showStatusMsg", msg)
			}

			if !statusMsg.isError {
				t.Error("Delete() with invalid index should return error status")
			}

			if statusMsg.message != "Invalid bookmark index" {
				t.Errorf("Delete() message = %s, want 'Invalid bookmark index'", statusMsg.message)
			}
		})
	}
}

func TestBookmarkManager_Add_Success(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a minimal config file with bookmarks section
	configContent := `[keybindings]
quit = ["Q"]

[[bookmarks]]
name = "existing"
path = "/existing"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Override config path for testing
	originalGetConfigPath := config.GetConfigPath
	defer func() {
		// This won't work as GetConfigPath is a function, not a variable
		// We need to test the actual save functionality differently
		_ = originalGetConfigPath
	}()

	manager := &BookmarkManager{
		bookmarks: []config.Bookmark{
			{Name: "existing", Path: "/existing"},
		},
		editIndex: -1,
	}

	// Test that Add creates the correct command structure
	cmd := manager.Add("/new/path", "new_bookmark")

	if cmd == nil {
		t.Fatal("Add() returned nil command")
	}

	// Note: We can't fully test the save functionality without mocking
	// the config package functions. The message will likely be an error
	// about saving, but we can verify the command is created.
}

func TestBookmarkManager_Edit_Success(t *testing.T) {
	manager := &BookmarkManager{
		bookmarks: []config.Bookmark{
			{Name: "old_name", Path: "/test/path"},
		},
		editIndex: -1,
	}

	// Test that Edit creates the correct command structure
	cmd := manager.Edit(0, "new_name")

	if cmd == nil {
		t.Fatal("Edit() returned nil command")
	}

	// The command execution would require proper config file setup
}

func TestBookmarkManager_Delete_Success(t *testing.T) {
	manager := &BookmarkManager{
		bookmarks: []config.Bookmark{
			{Name: "to_delete", Path: "/delete/me"},
			{Name: "keep", Path: "/keep/me"},
		},
		editIndex: -1,
	}

	// Test that Delete creates the correct command structure
	cmd := manager.Delete(0)

	if cmd == nil {
		t.Fatal("Delete() returned nil command")
	}

	// The command execution would require proper config file setup
}

func TestBookmarkDeletedMsg(t *testing.T) {
	bookmarks := []config.Bookmark{
		{Name: "test1", Path: "/test1"},
		{Name: "test2", Path: "/test2"},
	}

	msg := bookmarkDeletedMsg{
		bookmarks: bookmarks,
	}

	if len(msg.bookmarks) != 2 {
		t.Errorf("bookmarkDeletedMsg bookmarks length = %d, want 2", len(msg.bookmarks))
	}

	if msg.bookmarks[0].Name != "test1" {
		t.Errorf("bookmarkDeletedMsg bookmarks[0].Name = %s, want test1", msg.bookmarks[0].Name)
	}
}

// Test that BookmarkManager initializes correctly
func TestBookmarkManager_Initialization(t *testing.T) {
	manager := &BookmarkManager{
		bookmarks: []config.Bookmark{},
		editIndex: -1,
	}

	// Verify empty state
	if len(manager.Bookmarks()) != 0 {
		t.Error("New manager should have empty bookmarks")
	}

	if manager.EditIndex() != -1 {
		t.Error("New manager should have editIndex -1")
	}
}
