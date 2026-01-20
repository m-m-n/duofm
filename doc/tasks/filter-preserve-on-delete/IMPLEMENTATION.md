# Implementation Plan: Preserve Filter State on File Deletion

## Overview

When a user deletes files while a filter is active, the filter is unexpectedly cleared because `LoadDirectory()` resets filter state. This implementation adds a new method `ReloadDirectoryWithFilter()` that preserves filter state during delete operations.

## Objectives

- Preserve filter pattern and mode after file deletion
- Maintain user's working context during delete operations
- Support all filter modes: incremental, regex, and SQL-like

## Prerequisites

### Development Environment
- Go 1.21 or later
- Make (for build automation)

### Dependencies
- `ApplyFilter()` method in `pane_filter.go` (existing)
- `fs.ReadDirectory()` for filesystem access (existing)
- `SortEntries()` for sorting (existing)
- `filterHiddenFiles()` for hidden file filtering (existing)

### Knowledge Requirements
- Bubble Tea architecture (message-based updates)
- Pane filter state management (`filterPattern`, `filterMode`)
- Delete operation flow in `executeDeleteOperation()`

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea (github.com/charmbracelet/bubbletea)
- **Key Libraries**:
  - Lip Gloss - Styling

### Design Approach

The solution creates a new method `ReloadDirectoryWithFilter()` that:
1. Saves current filter state before reloading
2. Reloads directory entries from filesystem
3. Re-applies the saved filter to new entries

To avoid code duplication, common reload logic is extracted into a shared helper function `loadEntriesFromDisk()` that both `LoadDirectory()` and `ReloadDirectoryWithFilter()` use.

**Shared Helper Function:**
- `loadEntriesFromDisk()`: Reads directory from filesystem, applies sort, filters hidden files
- Used by both `LoadDirectory()` (clears filter) and `ReloadDirectoryWithFilter()` (preserves filter)

### Component Interaction

```
executeDeleteOperation()
    |
    +-- Delete file(s) from filesystem
    |
    +-- Call ReloadDirectoryWithFilter()
           |
           +-- Save filterPattern, filterMode
           |
           +-- loadEntriesFromDisk()  <-- Shared helper
           |       |
           |       +-- Read directory from filesystem
           |       +-- Apply sort configuration
           |       +-- Filter hidden files
           |
           +-- Update allEntries
           |
           +-- If filter was active:
           |       +-- Call ApplyFilter()
           |
           +-- Adjust cursor position

LoadDirectory()  (refactored)
    |
    +-- loadEntriesFromDisk()  <-- Same shared helper
    |       |
    |       +-- Read directory from filesystem
    |       +-- Apply sort configuration
    |       +-- Filter hidden files
    |
    +-- Update allEntries
    +-- Clear filter state
    +-- Reset cursor and scroll
    +-- Clear marks
    +-- Update Git branch
```

## Implementation Phases

### Phase 1: Add Shared Helper and ReloadDirectoryWithFilter Method

**Goal**: Extract common reload logic into a shared helper function and create a method that reloads directory contents while preserving and re-applying filter state.

**Files to Create**:
- None

**Files to Modify**:
- `internal/ui/pane.go` - Add `loadEntriesFromDisk()` helper function, refactor `LoadDirectory()` to use it
- `internal/ui/pane_filter.go` - Add `ReloadDirectoryWithFilter()` method using the shared helper

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| loadEntriesFromDisk | Read directory, apply sort, filter hidden files | Valid directory path | Sorted, filtered entries returned |
| ReloadDirectoryWithFilter | Reload directory entries while preserving filter state | Valid directory path | Entries updated, filter re-applied if active |

**Processing Flow**:

**loadEntriesFromDisk() helper:**
```
1. Read directory entries from filesystem
   +-- Error -> Return error
2. Apply sort configuration to entries
3. Filter hidden files if showHidden is false
4. Return entries
```

**ReloadDirectoryWithFilter():**
```
1. Save current filter state
   +-- Store filterPattern
   +-- Store filterMode
2. Clear marked entries (deleted files no longer exist)
3. Call loadEntriesFromDisk()
   +-- Error -> Return error, filter state unchanged
4. Update allEntries with new entries
5. Check if filter was active
   +-- Pattern not empty -> Call ApplyFilter() with saved state
       +-- Success: ApplyFilter handles cursor adjustment and scroll
       +-- Error (e.g., invalid regex): Log error, show all entries (graceful degradation)
           +-- Set entries = allEntries
           +-- Clear filter state (filterPattern = "", filterMode = SearchModeNone)
           +-- Adjust cursor position if out of bounds
           +-- Reset scroll position
   +-- Pattern empty -> Set entries = allEntries, reset filter state
       +-- Adjust cursor position if out of bounds (only for non-filtered case)
       +-- Reset scroll position (only for non-filtered case)
6. Return nil (success)
```

**LoadDirectory() refactored:**
```
1. Call loadEntriesFromDisk()
   +-- Error -> Return error
2. Update allEntries = entries
3. Clear filter state (filterPattern = "", filterMode = SearchModeNone)
4. Set entries = allEntries
5. Reset cursor = 0, scrollOffset = 0
6. Clear marked entries
7. Update Git branch
8. Return nil (success)
```

**Cursor Adjustment Responsibility**:
| Case | Cursor Adjustment | Scroll Adjustment |
|------|-------------------|-------------------|
| Filter active | ApplyFilter() handles | ApplyFilter() handles |
| No filter | ReloadDirectoryWithFilter() handles | ReloadDirectoryWithFilter() handles |

**Implementation Steps**:

1. **Add loadEntriesFromDisk helper to pane.go**
   - Read directory from filesystem
   - Apply sort configuration
   - Filter hidden files
   - Return sorted, filtered entries

2. **Refactor LoadDirectory to use helper**
   - Call loadEntriesFromDisk() instead of inline logic
   - Keep existing filter clearing and state reset logic

3. **Add ReloadDirectoryWithFilter method to pane_filter.go**
   - Save filter state at start
   - Clear marked entries
   - Call loadEntriesFromDisk() to reload entries
   - Re-apply filter using existing ApplyFilter()
   - Handle cursor adjustment
   - Reset scroll position

**Dependencies**:
- Requires: None (uses existing methods)
- Blocks: Phase 2

**Testing Approach**:

*Unit Tests*:
- Test loadEntriesFromDisk returns sorted, filtered entries
- Test LoadDirectory uses shared helper correctly
- Test filter preservation with incremental search
- Test filter preservation with regex search
- Test filter preservation with SQL-like search
- Test behavior when no filter is active
- Test cursor adjustment after reload

**Acceptance Criteria**:
- [ ] loadEntriesFromDisk correctly sorts and filters entries
- [ ] LoadDirectory uses loadEntriesFromDisk (no code duplication)
- [ ] ReloadDirectoryWithFilter uses loadEntriesFromDisk (no code duplication)
- [ ] ReloadDirectoryWithFilter preserves filterPattern
- [ ] ReloadDirectoryWithFilter preserves filterMode
- [ ] Marked entries are cleared after reload
- [ ] Filter is re-applied to updated entries
- [ ] Cursor position is adjusted correctly
- [ ] Scroll position is reset appropriately

**Estimated Effort**: Small (1-2 days)

---

### Phase 2: Update Delete Operation

**Goal**: Replace `LoadDirectory()` calls with `ReloadDirectoryWithFilter()` in delete operation.

**Files to Modify**:
- `internal/ui/model_update.go` - Modify `executeDeleteOperation()`

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| executeDeleteOperation | Execute file deletion and reload directory | File(s) selected for deletion | File(s) deleted, filter preserved |

**Processing Flow**:
```
1. Delete file(s) from filesystem
2. If deletion successful:
   +-- Call ReloadDirectoryWithFilter() instead of LoadDirectory()
   +-- Calculate new cursor position
3. If deletion failed:
   +-- Show error dialog
   +-- Do not modify filter state
```

**Implementation Steps**:

1. **Replace LoadDirectory calls in executeDeleteOperation**
   - Single file deletion: Replace `activePane.LoadDirectory()` with `activePane.ReloadDirectoryWithFilter()`
   - Multiple file deletion: Replace `activePane.LoadDirectory()` with `activePane.ReloadDirectoryWithFilter()`

**Dependencies**:
- Requires: Phase 1 (ReloadDirectoryWithFilter method)
- Blocks: None

**Testing Approach**:

*Unit Tests*:
- Test single file deletion preserves incremental filter
- Test single file deletion preserves regex filter
- Test single file deletion preserves SQL-like filter
- Test multiple file deletion preserves filter
- Test deletion without filter (no regression)

*Integration Tests*:
- Test end-to-end delete with filter active

**Acceptance Criteria**:
- [ ] Single file deletion preserves filter
- [ ] Multiple file deletion preserves filter
- [ ] Deleted files are removed from filtered list
- [ ] Cursor position is correctly adjusted
- [ ] No regression in delete without filter

**Estimated Effort**: Small (1-2 days)

**Risks and Mitigation**:
- **Risk**: Cursor position calculation may not account for filtered entries
  - **Mitigation**: Use existing calculateCursorAfterDeletion which operates on p.entries (filtered list)

---

## Complete File Structure

```
internal/ui/
+-- pane.go                 # Add loadEntriesFromDisk() helper, refactor LoadDirectory()
+-- pane_filter.go          # Add ReloadDirectoryWithFilter() using shared helper
+-- pane_filter_test.go     # Add tests for ReloadDirectoryWithFilter() (new file)
+-- model_update.go         # Modify executeDeleteOperation()
+-- model_update_delete_test.go  # Add filter preservation tests
```

**File Descriptions**:
- `pane.go`: Contains loadEntriesFromDisk() helper and refactored LoadDirectory()
- `pane_filter.go`: Contains filter-related methods including new ReloadDirectoryWithFilter()
- `pane_filter_test.go`: Unit tests for ReloadDirectoryWithFilter() method
- `model_update.go`: Contains executeDeleteOperation() that will use new method
- `model_update_delete_test.go`: Integration tests for delete operation with filter

## Testing Strategy

### Unit Testing

**Approach**:
- Use Go's built-in `testing` package
- Table-driven tests for filter mode variations
- Use temporary directories for filesystem tests

**Test Coverage Goals**:
- ReloadDirectoryWithFilter: 90%+ coverage
- Delete operation filter preservation: 80%+ coverage

**Key Test Areas**:

1. **loadEntriesFromDisk** (`internal/ui/pane_test.go`)
   - Returns sorted entries
   - Filters hidden files when showHidden is false
   - Returns all files when showHidden is true

2. **ReloadDirectoryWithFilter** (`internal/ui/pane_filter_test.go`)
   - Incremental filter preservation
   - Regex filter preservation
   - SQL-like filter preservation
   - No filter case
   - Cursor adjustment

3. **Delete Operation** (`internal/ui/model_update_delete_test.go`)
   - Filter preservation after single delete
   - Filter preservation after multiple delete
   - No regression without filter

### Integration Testing

**Scenarios**:
1. Delete file while incremental search active
2. Delete multiple marked files while filter active
3. Delete last matching file (entries become empty)

### Manual Testing Checklist

- [ ] Apply incremental filter, delete file, verify filter persists
- [ ] Apply regex filter, delete file, verify filter persists
- [ ] Apply SQL-like filter, delete file, verify filter persists
- [ ] Mark multiple files with filter, delete all, verify filter persists
- [ ] Delete without filter, verify normal behavior

## Dependencies

### External Dependencies
None

### Internal Dependencies

**Implementation Order**:
1. Phase 1: ReloadDirectoryWithFilter method
2. Phase 2: Update executeDeleteOperation

**Component Dependencies**:
- `ReloadDirectoryWithFilter()` depends on `ApplyFilter()`
- `executeDeleteOperation()` depends on `ReloadDirectoryWithFilter()`

## Risk Assessment

### Technical Risks

1. **Filter Pattern Invalid After File System Change**
   - **Risk**: Edge case where filter pattern becomes invalid
   - **Likelihood**: Low (pattern doesn't depend on files)
   - **Impact**: Low
   - **Mitigation**: ApplyFilter already handles invalid patterns gracefully

### Implementation Risks

1. **Regression in Other Operations**
   - **Risk**: Other operations using LoadDirectory might be affected
   - **Likelihood**: None (only modifying delete operation)
   - **Mitigation**: Only change calls in executeDeleteOperation

## Performance Considerations

- ReloadDirectoryWithFilter has same performance as LoadDirectory
- ApplyFilter is O(n) where n is number of entries
- No performance regression expected

## Error Handling

| Error | Condition | Handling |
|-------|-----------|----------|
| Directory read error | Filesystem access fails | Return error, filter state unchanged |
| Filter apply error | Invalid regex pattern | Log error, show all entries (graceful degradation) |

## Open Questions

### From Specification:
- None (specification is complete)

### Implementation-Specific:
- None

## Success Metrics

### Functional Completeness
- [ ] ReloadDirectoryWithFilter method implemented
- [ ] executeDeleteOperation updated
- [ ] All test scenarios pass

### Quality Metrics
- [ ] Test coverage meets goals (80%+ core logic)
- [ ] No regressions in existing tests

## References

- **Specification**: `doc/tasks/filter-preserve-on-delete/SPEC.md`
- **Requirements**: `doc/tasks/filter-preserve-on-delete/要件定義書.md`
- **Related Code**:
  - `internal/ui/pane_filter.go` - ApplyFilter, ClearFilter
  - `internal/ui/model_update.go` - executeDeleteOperation

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - Confirm approach
   - Address any questions

2. **Begin Implementation**
   - Start with Phase 1 (ReloadDirectoryWithFilter)
   - Write tests first (TDD approach)
   - Proceed to Phase 2

3. **Verification**
   - Run automated tests
   - Complete manual testing checklist
