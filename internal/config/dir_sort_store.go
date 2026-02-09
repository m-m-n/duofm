package config

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	dirSortFileName   = "dir_sort.toml"
	maxDirSortEntries = 1000
)

// validFields and validOrders define the accepted string values for sort settings.
var (
	validFields = map[string]bool{"name": true, "size": true, "date": true}
	validOrders = map[string]bool{"asc": true, "desc": true}
)

// DirSortEntry holds sort settings for a single directory.
type DirSortEntry struct {
	Field      string    `toml:"field"`
	Order      string    `toml:"order"`
	LastAccess time.Time `toml:"last_access"`
}

// dirSortFile represents the TOML file structure.
type dirSortFile struct {
	Dirs map[string]*DirSortEntry `toml:"dirs"`
}

// DirSortStore manages per-directory sort settings with TOML persistence.
type DirSortStore struct {
	configDir string
	entries   map[string]*DirSortEntry
}

// NewDirSortStore creates a new DirSortStore for the given config directory.
func NewDirSortStore(configDir string) *DirSortStore {
	return &DirSortStore{
		configDir: configDir,
		entries:   make(map[string]*DirSortEntry),
	}
}

// Load reads dir_sort.toml from disk into the in-memory map.
// If the file does not exist, the map remains empty.
// If parsing fails, a warning is logged and the map remains empty.
func (s *DirSortStore) Load() error {
	filePath := filepath.Join(s.configDir, dirSortFileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		log.Printf("Warning: failed to read %s: %v", filePath, err)
		return nil
	}

	var file dirSortFile
	if _, err := toml.Decode(string(data), &file); err != nil {
		log.Printf("Warning: failed to parse %s: %v", filePath, err)
		return nil
	}

	if file.Dirs == nil {
		return nil
	}

	// Clear existing entries before loading from disk
	s.entries = make(map[string]*DirSortEntry)

	for dirPath, entry := range file.Dirs {
		if !validFields[entry.Field] || !validOrders[entry.Order] {
			continue
		}
		// Enforce max entry cap on load
		if len(s.entries) >= maxDirSortEntries {
			break
		}
		s.entries[dirPath] = &DirSortEntry{
			Field:      entry.Field,
			Order:      entry.Order,
			LastAccess: entry.LastAccess,
		}
	}

	return nil
}

// Save writes the in-memory map to dir_sort.toml.
// Errors are silently ignored.
func (s *DirSortStore) Save() {
	filePath := filepath.Join(s.configDir, dirSortFileName)

	if err := os.MkdirAll(s.configDir, 0755); err != nil {
		return
	}

	file := dirSortFile{Dirs: s.entries}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(file); err != nil {
		return
	}

	// Atomic write: write to temp file then rename
	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, buf.Bytes(), 0644); err != nil {
		return
	}
	os.Rename(tmpFile, filePath)
}

// Get looks up sort settings for a directory path.
// Returns the field, order, and whether the entry was found.
// Updates last_access on successful lookup.
func (s *DirSortStore) Get(dirPath string) (field string, order string, found bool) {
	entry, ok := s.entries[dirPath]
	if !ok {
		return "", "", false
	}
	entry.LastAccess = time.Now()
	return entry.Field, entry.Order, true
}

// Set stores sort settings for a directory path.
// If the map is at capacity and the entry is new, the oldest entry is evicted.
// The map is saved to disk after modification.
func (s *DirSortStore) Set(dirPath, field, order string) {
	// If entry is new and at capacity, evict oldest
	if _, exists := s.entries[dirPath]; !exists && len(s.entries) >= maxDirSortEntries {
		s.evictOldest()
	}

	s.entries[dirPath] = &DirSortEntry{
		Field:      field,
		Order:      order,
		LastAccess: time.Now(),
	}

	s.Save()
}

// Len returns the number of entries in the store.
func (s *DirSortStore) Len() int {
	return len(s.entries)
}

// GetLastAccess returns the last access time for a directory entry.
// Returns zero time if the entry does not exist.
func (s *DirSortStore) GetLastAccess(dirPath string) time.Time {
	entry, ok := s.entries[dirPath]
	if !ok {
		return time.Time{}
	}
	return entry.LastAccess
}

// setWithTime stores sort settings with an explicit timestamp (for testing).
func (s *DirSortStore) setWithTime(dirPath, field, order string, t time.Time) {
	if _, exists := s.entries[dirPath]; !exists && len(s.entries) >= maxDirSortEntries {
		s.evictOldest()
	}

	s.entries[dirPath] = &DirSortEntry{
		Field:      field,
		Order:      order,
		LastAccess: t,
	}
}

// evictOldest removes the entry with the oldest last_access time.
func (s *DirSortStore) evictOldest() {
	var oldestPath string
	var oldestTime time.Time
	first := true

	for path, entry := range s.entries {
		if first || entry.LastAccess.Before(oldestTime) {
			oldestPath = path
			oldestTime = entry.LastAccess
			first = false
		}
	}

	if !first {
		delete(s.entries, oldestPath)
	}
}
