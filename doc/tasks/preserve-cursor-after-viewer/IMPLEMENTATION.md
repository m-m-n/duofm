# Implementation Plan: Preserve Cursor Position After External Viewer

## Overview

Implement cursor position preservation when returning from external applications (less/vim). Currently, the cursor resets to position 0 after external commands complete. This change will preserve the cursor on the same file, or reset to position 0 if the file no longer exists.

## Objectives

- Preserve cursor position after external viewer (less) exits
- Preserve cursor position after external editor (vim) exits
- Reset cursor to beginning if selected file is deleted during external app execution

## Prerequisites

- Existing `execFinishedMsg` handler in `model.go`
- Existing `LoadDirectory()` and `adjustScroll()` methods in `pane.go`
- Understanding of current cursor management

## Architecture Overview

The solution adds a new method `RefreshDirectoryPreserveCursor()` that:
1. Saves the currently selected file name
2. Reloads directory contents
3. Searches for the saved file name in new entries
4. Restores cursor to that position, or resets to 0 if not found

Note: The existing `Refresh()` method has similar functionality but falls back to the previous index when file is not found. Our requirement specifies resetting to 0 instead.

## Implementation Phases

### Phase 1: Add RefreshDirectoryPreserveCursor Method

**Goal**: Create a new method in Pane that reloads directory while preserving cursor by filename.

**Files to Modify**:
- `internal/ui/pane.go` - Add `RefreshDirectoryPreserveCursor()` method

**Implementation Steps**:

1. Add `RefreshDirectoryPreserveCursor()` method after `LoadDirectory()` (around line 118)
   - Save current selected file name using `SelectedEntry()`
   - Call `fs.ReadDirectory()` to reload entries
   - Apply sorting with `SortEntries()`
   - Apply hidden file filter if needed
   - Search for saved filename in new entries
   - Set cursor to found index, or 0 if not found
   - Call `adjustScroll()` to ensure visibility
   - Clear marks (consistent with LoadDirectory behavior)

**Code to Add**:
```go
// RefreshDirectoryPreserveCursor reloads directory contents while preserving cursor position.
// If the previously selected file no longer exists, cursor resets to the beginning.
func (p *Pane) RefreshDirectoryPreserveCursor() error {
    // Store current selected file name
    var selectedName string
    if entry := p.SelectedEntry(); entry != nil {
        selectedName = entry.Name
    }

    // Reload directory entries
    entries, err := fs.ReadDirectory(p.path)
    if err != nil {
        return err
    }

    entries = SortEntries(entries, p.sortConfig)

    // Filter hidden files
    if !p.showHidden {
        entries = filterHiddenFiles(entries)
    }

    p.allEntries = entries
    p.entries = entries
    p.filterPattern = ""
    p.filterMode = SearchModeNone

    // Find the previously selected file in new entries
    newCursor := 0 // Default to beginning if file not found
    if selectedName != "" {
        for i, e := range entries {
            if e.Name == selectedName {
                newCursor = i
                break
            }
        }
    }

    p.cursor = newCursor
    p.adjustScroll()

    // Clear marks on refresh (same as LoadDirectory)
    p.markedFiles = make(map[string]bool)

    return nil
}
```

**Dependencies**: None

**Testing**:
- Unit test: cursor preserved when file exists
- Unit test: cursor reset to 0 when file deleted
- Unit test: empty directory handling

**Estimated Effort**: Small

---

### Phase 2: Update execFinishedMsg Handler

**Goal**: Use the new method in the external command completion handler.

**Files to Modify**:
- `internal/ui/model.go` - Update `execFinishedMsg` case

**Implementation Steps**:

1. Locate `execFinishedMsg` handler (around line 387)
2. Replace `LoadDirectory()` calls with `RefreshDirectoryPreserveCursor()`

**Code Change**:
```go
// Before:
case execFinishedMsg:
    m.getActivePane().LoadDirectory()
    m.getInactivePane().LoadDirectory()

// After:
case execFinishedMsg:
    m.getActivePane().RefreshDirectoryPreserveCursor()
    m.getInactivePane().RefreshDirectoryPreserveCursor()
```

**Dependencies**: Phase 1 must be completed

**Testing**:
- Manual test: view file with less, exit, verify cursor position
- Manual test: edit file with vim, exit, verify cursor position
- Manual test: delete file in vim, exit, verify cursor at beginning

**Estimated Effort**: Small

---

### Phase 3: Add Unit Tests

**Goal**: Add comprehensive unit tests for the new method.

**Files to Modify**:
- `internal/ui/pane_test.go` - Add test functions

**Implementation Steps**:

1. Add `TestRefreshDirectoryPreserveCursor` test function
2. Test case: cursor preserved when file exists
3. Test case: cursor reset to 0 when selected file is deleted
4. Test case: handles empty directory
5. Test case: clears filter pattern

**Test Code**:
```go
func TestRefreshDirectoryPreserveCursor(t *testing.T) {
    tmpDir := t.TempDir()

    // Create test files
    os.WriteFile(filepath.Join(tmpDir, "aaa.txt"), []byte(""), 0644)
    os.WriteFile(filepath.Join(tmpDir, "bbb.txt"), []byte(""), 0644)
    os.WriteFile(filepath.Join(tmpDir, "ccc.txt"), []byte(""), 0644)

    pane, err := NewPane(tmpDir, 40, 20, true)
    if err != nil {
        t.Fatalf("NewPane() failed: %v", err)
    }

    t.Run("preserves cursor on same file", func(t *testing.T) {
        // Move cursor to bbb.txt (index 1, after ..)
        pane.cursor = 2 // .. is index 0, aaa.txt is 1, bbb.txt is 2

        err := pane.RefreshDirectoryPreserveCursor()
        if err != nil {
            t.Fatalf("RefreshDirectoryPreserveCursor() failed: %v", err)
        }

        // Cursor should still be on bbb.txt
        entry := pane.SelectedEntry()
        if entry == nil || entry.Name != "bbb.txt" {
            t.Errorf("Expected cursor on bbb.txt, got %v", entry)
        }
    })

    t.Run("resets cursor to 0 when file deleted", func(t *testing.T) {
        // Move cursor to ccc.txt
        for i, e := range pane.entries {
            if e.Name == "ccc.txt" {
                pane.cursor = i
                break
            }
        }

        // Delete ccc.txt
        os.Remove(filepath.Join(tmpDir, "ccc.txt"))

        err := pane.RefreshDirectoryPreserveCursor()
        if err != nil {
            t.Fatalf("RefreshDirectoryPreserveCursor() failed: %v", err)
        }

        // Cursor should be at 0
        if pane.cursor != 0 {
            t.Errorf("Expected cursor at 0, got %d", pane.cursor)
        }
    })
}
```

**Dependencies**: Phase 1 must be completed

**Testing**: Run `go test ./internal/ui/...`

**Estimated Effort**: Small

---

## File Structure

```
internal/ui/
├── pane.go           # Add RefreshDirectoryPreserveCursor() method
├── pane_test.go      # Add tests for RefreshDirectoryPreserveCursor()
└── model.go          # Update execFinishedMsg handler
```

## Testing Strategy

### Unit Tests
- `TestRefreshDirectoryPreserveCursor` - cursor preservation scenarios
- Existing tests must continue to pass

### Manual Testing Checklist
- [ ] View file with `v`, exit less with `q`, cursor stays on same file
- [ ] View file with `Enter`, exit less with `q`, cursor stays on same file
- [ ] Edit file with `e`, exit vim with `:q`, cursor stays on same file
- [ ] Delete file in vim, exit, cursor resets to beginning
- [ ] Inactive pane also preserves cursor position
- [ ] Large file list (100+ files) - cursor in middle preserved

## Dependencies

### Internal Dependencies
- Phase 2 depends on Phase 1 (method must exist before usage)
- Phase 3 depends on Phase 1 (tests require the method)

## Risk Assessment

### Technical Risks
- **Method name collision**: Low risk - `RefreshDirectoryPreserveCursor` is unique
  - Mitigation: Verified no existing method with this name

### Implementation Risks
- **Breaking existing behavior**: Low risk - `LoadDirectory()` unchanged
  - Mitigation: Only `execFinishedMsg` handler modified

## Performance Considerations

- No additional filesystem operations beyond what `LoadDirectory()` already performs
- Linear search for filename is O(n) but acceptable for typical directory sizes
- No performance regression expected

## Security Considerations

None - no new external inputs or commands introduced.

## Verification

After implementation, run:
```bash
# Run all tests
make test

# Manual verification
./duofm
# Navigate to a file, press 'v', exit less, verify cursor position
```

## References

- Specification: `doc/tasks/preserve-cursor-after-viewer/SPEC.md`
- Requirements: `doc/tasks/preserve-cursor-after-viewer/要件定義書.md`
- Related feature: `doc/tasks/open-file/SPEC.md`
