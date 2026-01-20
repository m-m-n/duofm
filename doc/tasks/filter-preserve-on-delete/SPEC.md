# Feature: Preserve Filter State on File Deletion

## Overview

When a user deletes files while a filter is active, the filter is unexpectedly cleared. This specification defines the fix to preserve filter state during delete operations.

## Objectives

- Preserve filter pattern and mode after file deletion
- Maintain user's working context during delete operations
- Support all filter modes: incremental, regex, and SQL-like

## User Stories

### US1: Delete Single File with Filter Active
As a user, I want the filter to remain active after deleting a file, so that I can continue working within my filtered view.

**Acceptance Criteria:**
- [ ] Filter pattern is preserved after single file deletion
- [ ] Filter mode is preserved after single file deletion
- [ ] Deleted file is removed from the filtered list
- [ ] Cursor position is correctly adjusted

### US2: Delete Multiple Marked Files with Filter Active
As a user, I want the filter to remain active after deleting multiple marked files, so that I can efficiently manage filtered files.

**Acceptance Criteria:**
- [ ] Filter pattern is preserved after multiple file deletion
- [ ] Filter mode is preserved after multiple file deletion
- [ ] All deleted files are removed from the filtered list
- [ ] Marks are cleared after deletion

## Technical Requirements

### Functional Requirements
- **FR1:** Delete operation must preserve `filterPattern` field value
- **FR2:** Delete operation must preserve `filterMode` field value
- **FR3:** After reload, filter must be re-applied to updated entry list
- **FR4:** On deletion error, filter state must remain unchanged

### Non-Functional Requirements
- **NFR1 - Performance:** Filter re-application must be imperceptible (< 10ms)
- **NFR2 - Compatibility:** Existing delete behavior without filter must be unchanged

## Root Cause Analysis

### Current Behavior
1. `executeDeleteOperation()` calls `activePane.LoadDirectory()` after successful deletion
2. `LoadDirectory()` (pane.go:110-112) resets filter state:
   ```go
   p.filterPattern = ""
   p.filterMode = SearchModeNone
   ```

### Problem Location
- File: `internal/ui/pane.go`
- Lines: 110-112
- Method: `LoadDirectory()`

## Implementation Approach

### Solution: Create ReloadDirectoryWithFilter Method with Shared Helper

Add a new method that reloads directory contents while preserving and re-applying the current filter. To avoid code duplication, extract common reload logic into a shared helper function.

### Helper Function

```go
// loadEntriesFromDisk reads directory entries from filesystem, applies sort and hidden file filter.
// This is the shared logic between LoadDirectory() and ReloadDirectoryWithFilter().
func (p *Pane) loadEntriesFromDisk() ([]fs.FileEntry, error)
```

### Method Signature

```go
// ReloadDirectoryWithFilter reloads directory contents while preserving filter state.
// If a filter was active, it is re-applied to the new entries.
func (p *Pane) ReloadDirectoryWithFilter() error
```

### Refactoring Plan

Both `LoadDirectory()` and `ReloadDirectoryWithFilter()` share the following logic:
1. Read directory from filesystem
2. Apply sort configuration
3. Filter hidden files

This common logic is extracted into `loadEntriesFromDisk()` helper function.

**LoadDirectory() refactored:**
```
1. Call loadEntriesFromDisk() -> entries
2. Update allEntries = entries
3. Clear filter state (filterPattern = "", filterMode = SearchModeNone)
4. Set entries = allEntries
5. Reset cursor and scroll
6. Clear marked entries
7. Update Git branch
```

**ReloadDirectoryWithFilter():**
```
1. Save current filter state (pattern, mode)
2. Clear marked entries
3. Call loadEntriesFromDisk() -> entries
4. Update allEntries = entries
5. If filter was active:
   a. Re-apply filter to new allEntries
6. If filter was not active:
   a. Set entries = allEntries
7. Adjust cursor position
8. Reset scroll position
```

### Implementation Logic

```
1. Save current filter state (pattern, mode)
2. Clear all marked entries
3. Call loadEntriesFromDisk() to get entries
4. Update allEntries
5. If filter was active:
   a. Re-apply filter to new allEntries
6. If filter was not active:
   a. Set entries = allEntries
7. Adjust cursor position
8. Reset scroll position (adjustScroll)
```

### Data Flow

```
Delete File
    │
    ▼
ReloadDirectoryWithFilter()
    │
    ├── savedPattern = p.filterPattern
    ├── savedMode = p.filterMode
    │
    ├── Clear marked entries (p.marked = make(...))
    │
    ├── loadEntriesFromDisk()  ◄── Shared helper
    │       │
    │       ├── Read directory from filesystem
    │       ├── Apply sort
    │       └── Filter hidden files
    │
    ├── Update p.allEntries
    │
    ├── (if savedPattern != "")
    │   └── ApplyFilter(savedPattern, savedMode)
    │
    ├── Adjust cursor
    └── Reset scroll (adjustScroll)
```

### Files to Modify

| File | Change |
|------|--------|
| `internal/ui/pane.go` | Add `loadEntriesFromDisk()` helper, refactor `LoadDirectory()` to use it |
| `internal/ui/pane_filter.go` | Add `ReloadDirectoryWithFilter()` method using `loadEntriesFromDisk()` |
| `internal/ui/model_update.go` | Replace `LoadDirectory()` with `ReloadDirectoryWithFilter()` in `executeDeleteOperation()` |

### Code Changes

#### pane.go - New Helper Function

```go
// loadEntriesFromDisk reads directory entries from filesystem, applies sort and hidden file filter.
// This is the shared logic between LoadDirectory() and ReloadDirectoryWithFilter().
func (p *Pane) loadEntriesFromDisk() ([]fs.FileEntry, error) {
    entries, err := fs.ReadDirectory(p.path)
    if err != nil {
        return nil, err
    }

    entries = SortEntries(entries, p.sortConfig)

    // Filter hidden files
    if !p.showHidden {
        entries = filterHiddenFiles(entries)
    }

    return entries, nil
}
```

#### pane.go - Refactored LoadDirectory

```go
// LoadDirectory loads directory entries (synchronous version).
// Uses loadEntriesFromDisk() helper for shared logic.
func (p *Pane) LoadDirectory() error {
    entries, err := p.loadEntriesFromDisk()
    if err != nil {
        return err
    }

    // allEntriesにすべてのエントリを保存
    p.allEntries = entries
    // フィルタをクリアして全エントリを表示
    p.entries = entries
    p.filterPattern = ""
    p.filterMode = SearchModeNone
    p.cursor = 0
    p.scrollOffset = 0
    // Clear marks on directory change
    p.markedFiles = make(map[string]bool)

    // Gitブランチを更新
    p.gitBranch = fs.GetGitBranch(p.path)

    return nil
}
```

#### pane_filter.go - New Method

```go
// ReloadDirectoryWithFilter reloads directory contents while preserving filter state.
// If a filter was active, it is re-applied to the new entries.
// Uses loadEntriesFromDisk() helper for shared logic with LoadDirectory().
func (p *Pane) ReloadDirectoryWithFilter() error {
    // Save current filter state
    savedPattern := p.filterPattern
    savedMode := p.filterMode

    // Clear marked entries (deleted files no longer exist)
    p.markedFiles = make(map[string]bool)

    // Reload directory entries using shared helper
    entries, err := p.loadEntriesFromDisk()
    if err != nil {
        return err
    }

    // Update allEntries
    p.allEntries = entries

    // Re-apply filter if it was active
    if savedPattern != "" {
        return p.ApplyFilter(savedPattern, savedMode)
    }

    // No filter - show all entries
    // Note: filterPattern and filterMode are already empty, no need to reset
    p.entries = entries

    // Adjust cursor if out of bounds
    if p.cursor >= len(p.entries) {
        p.cursor = max(0, len(p.entries)-1)
    }
    p.adjustScroll()

    return nil
}
```

#### model_update.go - executeDeleteOperation Changes

Replace calls to `LoadDirectory()` with `ReloadDirectoryWithFilter()`:

```go
// Line 562: Single file deletion success case
activePane.ReloadDirectoryWithFilter()

// Line 591: Multiple file deletion success case
activePane.ReloadDirectoryWithFilter()
```

## Test Scenarios

### Unit Tests

#### Test: Preserve Incremental Filter on Single Delete
```go
func TestReloadDirectoryWithFilter_Incremental(t *testing.T) {
    // Setup: Create pane with entries, apply incremental filter
    // Action: Call ReloadDirectoryWithFilter()
    // Assert: filterPattern and filterMode preserved, entries filtered
}
```

#### Test: Preserve Regex Filter on Single Delete
```go
func TestReloadDirectoryWithFilter_Regex(t *testing.T) {
    // Setup: Create pane with entries, apply regex filter
    // Action: Call ReloadDirectoryWithFilter()
    // Assert: filterPattern and filterMode preserved, entries filtered
}
```

#### Test: Preserve SQLLike Filter on Single Delete
```go
func TestReloadDirectoryWithFilter_SQLLike(t *testing.T) {
    // Setup: Create pane with entries, apply SQL-like filter
    // Action: Call ReloadDirectoryWithFilter()
    // Assert: filterPattern and filterMode preserved, entries filtered
}
```

#### Test: No Filter Active
```go
func TestReloadDirectoryWithFilter_NoFilter(t *testing.T) {
    // Setup: Create pane with entries, no filter
    // Action: Call ReloadDirectoryWithFilter()
    // Assert: All entries shown, filter state remains empty
}
```

### Integration Tests

#### Test: Delete Single File with Filter
```go
func TestExecuteDeleteOperation_PreservesFilter(t *testing.T) {
    // Setup: Model with filtered pane, cursor on deletable file
    // Action: Execute delete operation
    // Assert: Filter preserved, file removed from list
}
```

#### Test: Delete Multiple Files with Filter
```go
func TestExecuteDeleteOperation_MultipleFiles_PreservesFilter(t *testing.T) {
    // Setup: Model with filtered pane, multiple marked files
    // Action: Execute delete operation
    // Assert: Filter preserved, all marked files removed
}
```

### Edge Cases

- [ ] Delete last file matching filter (entries become empty)
- [ ] Delete file that doesn't match filter (should not be possible in UI, but handle gracefully)
- [ ] Invalid regex filter after reload (should not occur, but handle error)

## Success Criteria

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Manual testing confirms filter preservation
- [ ] No regression in delete without filter
- [ ] No regression in other operations using LoadDirectory()

## Dependencies

### Internal Dependencies
- `ApplyFilter()` method in pane_filter.go
- `fs.ReadDirectory()` for filesystem access
- `SortEntries()` for sorting
- `filterHiddenFiles()` for hidden file filtering

### External Dependencies
- None

## File Structure

```
internal/ui/
├── pane.go                 # Add loadEntriesFromDisk() helper, refactor LoadDirectory()
├── pane_filter.go          # Add ReloadDirectoryWithFilter() using shared helper
├── pane_filter_test.go     # Add tests for ReloadDirectoryWithFilter()
├── model_update.go         # Modify executeDeleteOperation()
└── model_update_delete_test.go  # Add filter preservation tests
```

## Error Handling

| Error | Condition | Handling |
|-------|-----------|----------|
| Directory read error | Filesystem access fails | Return error, do not modify filter state |
| Filter apply error | Invalid regex pattern | Log error, show all entries (graceful degradation) |

## References

- Requirements: `doc/tasks/filter-preserve-on-delete/要件定義書.md`
- Related task: `doc/tasks/fix-cursor-position-after-delete/`
