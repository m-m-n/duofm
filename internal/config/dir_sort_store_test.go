package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDirSortStore_SetAndGet(t *testing.T) {
	dir := t.TempDir()
	store := NewDirSortStore(dir)

	store.Set("/home/user/Downloads", "size", "desc")

	field, order, found := store.Get("/home/user/Downloads")
	if !found {
		t.Fatal("expected entry to be found")
	}
	if field != "size" {
		t.Errorf("expected field 'size', got %q", field)
	}
	if order != "desc" {
		t.Errorf("expected order 'desc', got %q", order)
	}
}

func TestDirSortStore_GetNotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewDirSortStore(dir)

	_, _, found := store.Get("/nonexistent")
	if found {
		t.Error("expected entry not to be found")
	}
}

func TestDirSortStore_Overwrite(t *testing.T) {
	dir := t.TempDir()
	store := NewDirSortStore(dir)

	store.Set("/home/user/Downloads", "name", "asc")
	store.Set("/home/user/Downloads", "date", "desc")

	field, order, found := store.Get("/home/user/Downloads")
	if !found {
		t.Fatal("expected entry to be found")
	}
	if field != "date" || order != "desc" {
		t.Errorf("expected date/desc, got %s/%s", field, order)
	}
}

func TestDirSortStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	// Save
	store1 := NewDirSortStore(dir)
	store1.Set("/home/user/Downloads", "size", "desc")
	store1.Set("/home/user/Documents", "name", "asc")

	// Load in new store
	store2 := NewDirSortStore(dir)
	err := store2.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	field, order, found := store2.Get("/home/user/Downloads")
	if !found {
		t.Fatal("expected Downloads entry to be found after load")
	}
	if field != "size" || order != "desc" {
		t.Errorf("expected size/desc, got %s/%s", field, order)
	}

	field, order, found = store2.Get("/home/user/Documents")
	if !found {
		t.Fatal("expected Documents entry to be found after load")
	}
	if field != "name" || order != "asc" {
		t.Errorf("expected name/asc, got %s/%s", field, order)
	}
}

func TestDirSortStore_LoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	store := NewDirSortStore(dir)

	err := store.Load()
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}

	_, _, found := store.Get("/anything")
	if found {
		t.Error("expected empty map after loading missing file")
	}
}

func TestDirSortStore_LoadCorruptedFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "dir_sort.toml")
	os.WriteFile(filePath, []byte("{{invalid toml content"), 0644)

	store := NewDirSortStore(dir)
	err := store.Load()
	if err != nil {
		t.Fatalf("expected no error for corrupted file, got: %v", err)
	}

	_, _, found := store.Get("/anything")
	if found {
		t.Error("expected empty map after loading corrupted file")
	}
}

func TestDirSortStore_InvalidFieldOrderSkipped(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "dir_sort.toml")
	content := `[dirs]

[dirs."/valid"]
field = "name"
order = "asc"
last_access = 2026-02-09T10:00:00+09:00

[dirs."/invalid-field"]
field = "invalid"
order = "asc"
last_access = 2026-02-09T10:00:00+09:00

[dirs."/invalid-order"]
field = "name"
order = "invalid"
last_access = 2026-02-09T10:00:00+09:00
`
	os.WriteFile(filePath, []byte(content), 0644)

	store := NewDirSortStore(dir)
	store.Load()

	_, _, found := store.Get("/valid")
	if !found {
		t.Error("expected valid entry to be loaded")
	}

	_, _, found = store.Get("/invalid-field")
	if found {
		t.Error("expected invalid-field entry to be skipped")
	}

	_, _, found = store.Get("/invalid-order")
	if found {
		t.Error("expected invalid-order entry to be skipped")
	}
}

func TestDirSortStore_LRUEviction(t *testing.T) {
	dir := t.TempDir()
	store := NewDirSortStore(dir)

	// Add 1000 entries with explicit timestamps
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 1000; i++ {
		path := filepath.Join("/dir", fmt.Sprintf("%04d", i))
		store.setWithTime(path, "name", "asc", baseTime.Add(time.Duration(i)*time.Second))
	}

	// Verify 1000 entries exist
	if store.Len() != 1000 {
		t.Fatalf("expected 1000 entries, got %d", store.Len())
	}

	// Add 1001st entry - oldest should be evicted
	store.Set("/new-dir", "size", "desc")

	if store.Len() != 1000 {
		t.Fatalf("expected 1000 entries after eviction, got %d", store.Len())
	}

	// The first entry (oldest) should be evicted
	_, _, found := store.Get("/dir/0000")
	if found {
		t.Error("expected oldest entry to be evicted")
	}

	// The new entry should exist
	_, _, found = store.Get("/new-dir")
	if !found {
		t.Error("expected new entry to exist")
	}
}

func TestDirSortStore_LRUPreservesRecent(t *testing.T) {
	dir := t.TempDir()
	store := NewDirSortStore(dir)

	// Add 1000 entries with explicit timestamps
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 1000; i++ {
		path := filepath.Join("/dir", fmt.Sprintf("%04d", i))
		store.setWithTime(path, "name", "asc", baseTime.Add(time.Duration(i)*time.Second))
	}

	// Access entry 0000 to make it recent (Get updates last_access to time.Now())
	store.Get("/dir/0000")

	// Add new entry - should evict entry 0001 (now the oldest)
	store.Set("/new-dir", "size", "desc")

	// 0000 should still exist (was accessed recently)
	_, _, found := store.Get("/dir/0000")
	if !found {
		t.Error("expected recently accessed entry to be preserved")
	}

	// 0001 should be evicted (now the oldest)
	_, _, found = store.Get("/dir/0001")
	if found {
		t.Error("expected oldest untouched entry to be evicted")
	}
}

func TestDirSortStore_LastAccessUpdatedOnGet(t *testing.T) {
	dir := t.TempDir()
	store := NewDirSortStore(dir)

	store.Set("/home/user/Downloads", "name", "asc")

	// Record time before Get
	time.Sleep(time.Millisecond)
	beforeGet := time.Now()
	time.Sleep(time.Millisecond)

	store.Get("/home/user/Downloads")

	// Verify last_access was updated
	lastAccess := store.GetLastAccess("/home/user/Downloads")
	if lastAccess.Before(beforeGet) {
		t.Error("expected last_access to be updated after Get")
	}
}

func TestDirSortStore_SpecialCharactersInPath(t *testing.T) {
	dir := t.TempDir()
	store := NewDirSortStore(dir)

	paths := []string{
		"/home/user/My Documents",
		"/home/user/日本語",
		"/home/user/path with spaces/and more",
	}

	for _, p := range paths {
		store.Set(p, "size", "desc")
	}

	for _, p := range paths {
		field, order, found := store.Get(p)
		if !found {
			t.Errorf("expected entry for path %q to be found", p)
			continue
		}
		if field != "size" || order != "desc" {
			t.Errorf("path %q: expected size/desc, got %s/%s", p, field, order)
		}
	}
}

func TestDirSortStore_RootDirectory(t *testing.T) {
	dir := t.TempDir()
	store := NewDirSortStore(dir)

	store.Set("/", "date", "asc")

	field, order, found := store.Get("/")
	if !found {
		t.Fatal("expected root directory entry to be found")
	}
	if field != "date" || order != "asc" {
		t.Errorf("expected date/asc, got %s/%s", field, order)
	}
}

func TestDirSortStore_FileWriteErrorNocrash(t *testing.T) {
	// Use a path that cannot be written to
	store := NewDirSortStore("/nonexistent/path/that/does/not/exist")

	// Should not panic
	store.Set("/home/user/test", "name", "asc")

	// In-memory should still work
	field, order, found := store.Get("/home/user/test")
	if !found {
		t.Error("expected in-memory entry to exist despite file write error")
	}
	if field != "name" || order != "asc" {
		t.Errorf("expected name/asc, got %s/%s", field, order)
	}
}
