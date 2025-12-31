package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBookmark(t *testing.T) {
	t.Run("NewBookmark creates bookmark with name and path", func(t *testing.T) {
		b := Bookmark{Name: "Projects", Path: "/path/to/projects"}
		if b.Name != "Projects" {
			t.Errorf("expected Name 'Projects', got '%s'", b.Name)
		}
		if b.Path != "/path/to/projects" {
			t.Errorf("expected Path '/path/to/projects', got '%s'", b.Path)
		}
	})
}

func TestLoadBookmarks(t *testing.T) {
	t.Run("returns empty slice when config file does not exist", func(t *testing.T) {
		bookmarks, warnings := LoadBookmarks("/nonexistent/config.toml")
		if len(bookmarks) != 0 {
			t.Errorf("expected empty slice, got %d bookmarks", len(bookmarks))
		}
		if len(warnings) != 0 {
			t.Errorf("expected no warnings, got %v", warnings)
		}
	})

	t.Run("returns empty slice when no bookmarks section", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")
		content := `[keybindings]
move_down = ["J"]
`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		bookmarks, warnings := LoadBookmarks(configPath)
		if len(bookmarks) != 0 {
			t.Errorf("expected empty slice, got %d bookmarks", len(bookmarks))
		}
		if len(warnings) != 0 {
			t.Errorf("expected no warnings, got %v", warnings)
		}
	})

	t.Run("parses valid bookmarks", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")
		content := `[[bookmarks]]
name = "Projects"
path = "/path/to/projects"

[[bookmarks]]
name = "Downloads"
path = "/path/to/downloads"
`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		bookmarks, warnings := LoadBookmarks(configPath)
		if len(warnings) != 0 {
			t.Errorf("expected no warnings, got %v", warnings)
		}
		if len(bookmarks) != 2 {
			t.Fatalf("expected 2 bookmarks, got %d", len(bookmarks))
		}
		if bookmarks[0].Name != "Projects" {
			t.Errorf("expected first bookmark name 'Projects', got '%s'", bookmarks[0].Name)
		}
		if bookmarks[0].Path != "/path/to/projects" {
			t.Errorf("expected first bookmark path '/path/to/projects', got '%s'", bookmarks[0].Path)
		}
		if bookmarks[1].Name != "Downloads" {
			t.Errorf("expected second bookmark name 'Downloads', got '%s'", bookmarks[1].Name)
		}
	})

	t.Run("skips invalid entries and returns warning", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")
		content := `[[bookmarks]]
name = "Valid"
path = "/valid/path"

[[bookmarks]]
name = ""
path = "/missing/name"

[[bookmarks]]
name = "MissingPath"
path = ""
`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		bookmarks, warnings := LoadBookmarks(configPath)
		if len(bookmarks) != 1 {
			t.Errorf("expected 1 valid bookmark, got %d", len(bookmarks))
		}
		if len(warnings) != 2 {
			t.Errorf("expected 2 warnings, got %d", len(warnings))
		}
	})
}

func TestSaveBookmarks(t *testing.T) {
	t.Run("saves bookmarks to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")
		// Create initial config with keybindings
		initial := `[keybindings]
move_down = ["J"]
`
		if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
			t.Fatal(err)
		}

		bookmarks := []Bookmark{
			{Name: "Projects", Path: "/path/to/projects"},
			{Name: "Downloads", Path: "/path/to/downloads"},
		}

		err := SaveBookmarks(configPath, bookmarks)
		if err != nil {
			t.Fatalf("SaveBookmarks failed: %v", err)
		}

		// Verify saved content
		loaded, warnings := LoadBookmarks(configPath)
		if len(warnings) != 0 {
			t.Errorf("unexpected warnings: %v", warnings)
		}
		if len(loaded) != 2 {
			t.Fatalf("expected 2 bookmarks, got %d", len(loaded))
		}
		if loaded[0].Name != "Projects" {
			t.Errorf("expected 'Projects', got '%s'", loaded[0].Name)
		}
	})

	t.Run("preserves existing config sections", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")
		initial := `[keybindings]
move_down = ["J"]
move_up = ["K"]

[colors]
cursor_fg = 15
`
		if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
			t.Fatal(err)
		}

		bookmarks := []Bookmark{
			{Name: "Test", Path: "/test"},
		}

		err := SaveBookmarks(configPath, bookmarks)
		if err != nil {
			t.Fatalf("SaveBookmarks failed: %v", err)
		}

		// Read the file and verify other sections are preserved
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}

		// Check that keybindings section exists
		if !strings.Contains(string(content), "[keybindings]") {
			t.Error("keybindings section was not preserved")
		}
		if !strings.Contains(string(content), "move_down") {
			t.Error("move_down keybinding was not preserved")
		}
	})

	t.Run("creates file if it does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "subdir", "config.toml")

		bookmarks := []Bookmark{
			{Name: "Test", Path: "/test"},
		}

		err := SaveBookmarks(configPath, bookmarks)
		if err != nil {
			t.Fatalf("SaveBookmarks failed: %v", err)
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}
	})
}

func TestAddBookmark(t *testing.T) {
	t.Run("adds bookmark to beginning of list", func(t *testing.T) {
		bookmarks := []Bookmark{
			{Name: "Existing", Path: "/existing"},
		}

		result, err := AddBookmark(bookmarks, "New", "/new")
		if err != nil {
			t.Fatalf("AddBookmark failed: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 bookmarks, got %d", len(result))
		}
		if result[0].Name != "New" {
			t.Errorf("expected first bookmark to be 'New', got '%s'", result[0].Name)
		}
		if result[0].Path != "/new" {
			t.Errorf("expected path '/new', got '%s'", result[0].Path)
		}
	})

	t.Run("normalizes path with trailing slash", func(t *testing.T) {
		var bookmarks []Bookmark

		result, err := AddBookmark(bookmarks, "Test", "/path/to/dir/")
		if err != nil {
			t.Fatalf("AddBookmark failed: %v", err)
		}
		// filepath.Clean removes trailing slash
		if result[0].Path != "/path/to/dir" {
			t.Errorf("expected normalized path '/path/to/dir', got '%s'", result[0].Path)
		}
	})

	t.Run("returns error for duplicate path", func(t *testing.T) {
		bookmarks := []Bookmark{
			{Name: "Existing", Path: "/existing"},
		}

		_, err := AddBookmark(bookmarks, "Duplicate", "/existing")
		if err == nil {
			t.Error("expected error for duplicate path, got nil")
		}
		if err != ErrDuplicatePath {
			t.Errorf("expected ErrDuplicatePath, got %v", err)
		}
	})

	t.Run("adds to empty list", func(t *testing.T) {
		var bookmarks []Bookmark

		result, err := AddBookmark(bookmarks, "First", "/first")
		if err != nil {
			t.Fatalf("AddBookmark failed: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 bookmark, got %d", len(result))
		}
	})

	t.Run("returns error for empty alias", func(t *testing.T) {
		var bookmarks []Bookmark

		_, err := AddBookmark(bookmarks, "", "/path")
		if err == nil {
			t.Error("expected error for empty alias, got nil")
		}
		if err != ErrEmptyAlias {
			t.Errorf("expected ErrEmptyAlias, got %v", err)
		}
	})
}

func TestRemoveBookmark(t *testing.T) {
	t.Run("removes bookmark by index", func(t *testing.T) {
		bookmarks := []Bookmark{
			{Name: "First", Path: "/first"},
			{Name: "Second", Path: "/second"},
			{Name: "Third", Path: "/third"},
		}

		result, err := RemoveBookmark(bookmarks, 1)
		if err != nil {
			t.Fatalf("RemoveBookmark failed: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 bookmarks, got %d", len(result))
		}
		if result[0].Name != "First" {
			t.Errorf("expected 'First', got '%s'", result[0].Name)
		}
		if result[1].Name != "Third" {
			t.Errorf("expected 'Third', got '%s'", result[1].Name)
		}
	})

	t.Run("returns error for invalid index", func(t *testing.T) {
		bookmarks := []Bookmark{
			{Name: "Only", Path: "/only"},
		}

		_, err := RemoveBookmark(bookmarks, -1)
		if err == nil {
			t.Error("expected error for negative index")
		}

		_, err = RemoveBookmark(bookmarks, 1)
		if err == nil {
			t.Error("expected error for out of range index")
		}
	})
}

func TestUpdateBookmarkAlias(t *testing.T) {
	t.Run("updates alias at index", func(t *testing.T) {
		bookmarks := []Bookmark{
			{Name: "Old", Path: "/path"},
		}

		result, err := UpdateBookmarkAlias(bookmarks, 0, "New")
		if err != nil {
			t.Fatalf("UpdateBookmarkAlias failed: %v", err)
		}
		if result[0].Name != "New" {
			t.Errorf("expected 'New', got '%s'", result[0].Name)
		}
		if result[0].Path != "/path" {
			t.Errorf("path should not change, got '%s'", result[0].Path)
		}
	})

	t.Run("returns error for invalid index", func(t *testing.T) {
		bookmarks := []Bookmark{
			{Name: "Only", Path: "/only"},
		}

		_, err := UpdateBookmarkAlias(bookmarks, 5, "New")
		if err == nil {
			t.Error("expected error for invalid index")
		}
	})

	t.Run("returns error for empty alias", func(t *testing.T) {
		bookmarks := []Bookmark{
			{Name: "Test", Path: "/test"},
		}

		_, err := UpdateBookmarkAlias(bookmarks, 0, "")
		if err == nil {
			t.Error("expected error for empty alias")
		}
	})
}

func TestIsPathBookmarked(t *testing.T) {
	bookmarks := []Bookmark{
		{Name: "Projects", Path: "/path/to/projects"},
		{Name: "Downloads", Path: "/path/to/downloads"},
	}

	t.Run("returns true for bookmarked path", func(t *testing.T) {
		if !IsPathBookmarked(bookmarks, "/path/to/projects") {
			t.Error("expected true for bookmarked path")
		}
	})

	t.Run("returns false for non-bookmarked path", func(t *testing.T) {
		if IsPathBookmarked(bookmarks, "/not/bookmarked") {
			t.Error("expected false for non-bookmarked path")
		}
	})

	t.Run("returns false for empty list", func(t *testing.T) {
		if IsPathBookmarked(nil, "/any") {
			t.Error("expected false for empty list")
		}
	})
}

func TestDefaultAliasFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/path/to/projects", "projects"},
		{"/home/user/Downloads", "Downloads"},
		{"/", "/"},
		{"/single", "single"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := DefaultAliasFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("DefaultAliasFromPath(%s) = %s, want %s", tt.path, result, tt.expected)
			}
		})
	}
}
