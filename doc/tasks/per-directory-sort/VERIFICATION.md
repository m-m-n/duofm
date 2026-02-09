# Verification Document: Per-Directory Sort Settings

## Overview
**Feature**: Per-Directory Sort Settings
**SPEC.md**: `doc/tasks/per-directory-sort/SPEC.md`
**IMPLEMENTATION.md**: `doc/tasks/per-directory-sort/IMPLEMENTATION.md`

## Build Verification

### Build Command
```bash
go build ./...
```

### Expected Result
- Exit code: 0
- No error messages

## Test Verification

### Test Command
```bash
go test ./... -v -cover
```

### Coverage Target
- **Minimum**: 80%
- **Target**: 90% (especially for DirSortStore)

### Test Scenarios from SPEC.md

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-1 | Save sort setting for a directory | Entry stored in map with correct field/order | Unit |
| TS-2 | Load sort settings from valid TOML | Map populated with all entries | Unit |
| TS-3 | Load with missing file | Empty map, no error | Unit |
| TS-4 | Load with corrupted file | Empty map, no crash | Unit |
| TS-5 | Lookup existing directory | Returns correct SortConfig, true | Unit |
| TS-6 | Lookup non-existing directory | Returns nil/zero, false | Unit |
| TS-7 | LRU eviction at 1000 entries | Oldest entry removed when adding 1001st | Unit |
| TS-8 | LRU preserves 999 most recent | All recent entries retained after eviction | Unit |
| TS-9 | Last access time updated on read | Timestamp newer after Get call | Unit |
| TS-10 | Header Line2 includes sort info | "Name ↑" or similar present in rendered output | Unit |
| TS-11 | Sort info display for all combinations | All 6 field/order combos render correctly | Unit |
| TS-12 | Sort dialog confirm triggers save | Store has entry after dialog confirmed | Integration |
| TS-13 | Directory navigation applies saved sort | Pane sortConfig matches saved value | Integration |
| TS-14 | Directory navigation applies default | Pane sortConfig is DefaultSortConfig when no saved setting | Integration |
| TS-15 | Sort dialog cancel does not save | Store unchanged after cancel | Integration |
| TS-16 | Path with special characters | Spaces and unicode in path work correctly | Unit |
| TS-17 | Root directory "/" as key | Save and load work for root path | Unit |
| TS-18 | Invalid field/order in TOML | Invalid entry skipped, others loaded | Unit |
| TS-19 | File write error handling | No crash, error silently ignored | Unit |
| TS-20 | Simultaneous pane navigation | Both panes apply correct sort independently | Integration |
| TS-21 | All navigation methods apply saved sort (Home, ChangeDir, SyncTo, MoveToParent) | Sort config auto-applied for each method | Integration |
| TS-22 | ModelOptions refactoring: NewModel() still works | Existing tests pass unchanged | Regression |
| TS-23 | ModelOptions refactoring: NewModelWithConfig backward compatible | All callers work with new signature | Regression |

## Code Quality Verification

### Format Check
```bash
gofmt -l ./internal/config/dir_sort_store.go
```

### Static Analysis
```bash
go vet ./...
```

## File Structure Verification

### Files to Create
- `internal/config/dir_sort_store.go` - Per-directory sort settings store
- `internal/config/dir_sort_store_test.go` - Store tests

### Files to Modify
- `internal/ui/model.go` - Add dirSortStore field to Model struct
- `internal/ui/model_update_dialog.go` - Save to store on sort dialog confirmation
- `internal/ui/pane_navigation.go` - Apply saved sort config on directory navigation
- `internal/ui/pane_render.go` - Add sort info to header Line2
- `internal/ui/pane_render_test.go` - Tests for sort info display
- `cmd/duofm/main.go` - Initialize DirSortStore and pass to Model

## SPEC.md Compliance

### Success Criteria

| ID | Criterion from SPEC.md | How to Verify |
|----|------------------------|---------------|
| SC-1 | All functional requirements (FR1-FR7) implemented | Review FR coverage table below |
| SC-2 | All test scenarios pass | Run `go test ./... -v` |
| SC-3 | Header always shows current sort state | TS-10, TS-11, manual visual check |
| SC-4 | Settings survive application restart | TS-2 (load from file), manual restart test |
| SC-5 | LRU eviction at 1000 entry boundary | TS-7, TS-8 |
| SC-6 | No user-visible errors on file I/O failures | TS-3, TS-4, TS-19 |
| SC-7 | Existing sort functionality unchanged | Run existing sort_dialog_test.go, cancel_key_test.go, model_menu_test.go |

### Functional Requirements Coverage

| Requirement | Implementation Phase | Verification |
|-------------|---------------------|--------------|
| FR1: Save on sort dialog confirm | Phase 3 | TS-12, TS-15 |
| FR2: Auto-apply on navigation | Phase 3 | TS-13, TS-14 |
| FR3: Display sort in header Line2 | Phase 2 | TS-10, TS-11 |
| FR4: Load on startup | Phase 1 | TS-2, TS-3, TS-4 |
| FR5: LRU eviction at 1000 | Phase 1 | TS-7, TS-8 |
| FR6: Update last_access on read | Phase 1 | TS-9 |
| FR7: Silent file I/O error handling | Phase 1 | TS-19 |

## Manual Testing (E2E Not Possible)

Items requiring human judgment or application restart:

- [ ] Start application, open sort dialog, change to "Size ↓", confirm → verify header shows "Size ↓"
- [ ] Navigate to different directory → verify header shows "Name ↑" (default)
- [ ] Navigate back to first directory → verify header shows "Size ↓" (auto-applied)
- [ ] Restart application, navigate to same directory → verify "Size ↓" is still applied
- [ ] Verify `~/.config/duofm/dir_sort.toml` contains the saved entry
- [ ] Delete `dir_sort.toml`, restart application → verify no error, default sort used
- [ ] Check header Line2 layout is visually balanced (mark info + sort info + free space)

## Performance Verification

### Benchmarks
- Sort setting lookup: O(1) map access (no benchmark needed, verified by implementation)
- No measurable impact on directory navigation time

## Verification Summary

| Category | Items | Automated | Manual |
|----------|-------|-----------|--------|
| Build | 1 | ✅ | - |
| Tests | 23 | ✅ | - |
| Code Quality | 2 | ✅ | - |
| File Structure | 8 | ✅ | - |
| SPEC Compliance | 7 | Partial | ✅ |
| Manual Testing | 7 | - | ✅ |

**Total**: 34 automated items, 7 manual items
