# Implementation Plan: Per-Directory Sort Settings

## Overview

Implement persistent per-directory sort settings with automatic application on directory navigation, and add sort state display to the pane header.

## Objectives

- Persist sort settings (field + order) per directory in a dedicated TOML file
- Automatically apply saved sort settings when entering a directory
- Display current sort state in header Line2
- Manage storage with LRU eviction at 1000 entries

## Prerequisites

### Development Environment
- Go 1.21+
- Existing duofm project builds and passes all tests

### Dependencies
- `github.com/BurntSushi/toml` (already in use)
- `internal/config` package (for config directory resolution)

### Knowledge Requirements
- Bubble Tea message-based architecture
- Existing sort system (SortConfig, SortDialog, sort dialog result messages)
- Pane navigation flow (async directory loading)

## Architecture Overview

### Design Approach

A new `DirSortStore` component in `internal/config` handles persistence. The UI layer (`Model`) owns the store instance and coordinates between sort dialog, navigation, and the store. The store operates as a simple in-memory map with TOML serialization.

### Component Interaction

```
Sort Dialog ──confirmed──► Model ──save──► DirSortStore ──write──► dir_sort.toml
                                                ▲
Navigation ──enter dir──► Model ──lookup──► DirSortStore
                            │
                            └──apply sort──► Pane
```

## Implementation Phases

### Phase 1: DirSortStore (Storage Layer)

**Goal**: Create the storage component that loads, saves, and manages per-directory sort settings with LRU eviction.

**Files to Create**:
- `internal/config/dir_sort_store.go` - Store implementation
- `internal/config/dir_sort_store_test.go` - Store tests

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| DirSortStore | Manage in-memory map of directory→sort settings with TOML persistence | Config directory path known | Settings accessible by directory path |
| DirSortEntry | Hold sort field (string), order (string), and last access time for one directory | Valid field and order string values | Serializable to TOML |
| Load | Read dir_sort.toml into memory map | File may or may not exist | Map populated (or empty on error) |
| Save | Write current map to dir_sort.toml | Map in valid state | File written (or error silently ignored) |
| Get | Look up sort setting for a directory, update last_access | Directory path provided | Returns field string, order string, found flag; updates access time |
| Set | Store sort setting for a directory, evict LRU if over limit | Valid field/order strings | Setting stored, file written, oldest evicted if >1000 |

**Dependency Design Note**:
DirSortStore uses string values ("name"/"size"/"date", "asc"/"desc") for field and order, NOT the SortConfig/SortField/SortOrder types from `internal/ui`. This avoids circular dependency between `config` and `ui` packages. The UI layer is responsible for converting between strings and SortConfig types.

**Processing Flow**:
```
Load:
1. Resolve file path from config directory
2. Read and parse TOML file
   ├─ File missing → return empty map
   ├─ Parse error → log warning, return empty map
   └─ Success → populate map, skip entries with invalid field/order values

Get(dirPath):
1. Look up dirPath in map
   ├─ Not found → return nil, false
   └─ Found → update last_access, return sort config, true

Set(dirPath, sortConfig):
1. Check if map size >= 1000 and entry is new
   └─ Yes → find and remove entry with oldest last_access
2. Store/overwrite entry with current timestamp
3. Save map to file (silently ignore errors)
```

**Testing Approach**:

| Scenario | Expected Result | Type |
|----------|-----------------|------|
| Save and retrieve sort setting | Correct field/order returned | Unit |
| Load from valid TOML file | Map populated correctly | Unit |
| Load with missing file | Empty map, no error | Unit |
| Load with corrupted file | Empty map, no crash | Unit |
| Lookup non-existing directory | nil, false | Unit |
| LRU eviction at 1000 entries | Oldest entry removed | Unit |
| LRU preserves 999 most recent | Correct entries retained | Unit |
| Last access updated on Get | Timestamp newer than before | Unit |
| Invalid field/order values in TOML | Entry skipped, others loaded | Unit |
| Special characters in path (spaces, unicode) | Correct lookup | Unit |
| Root directory "/" as key | Works correctly | Unit |

**Acceptance Criteria**:
- [ ] DirSortStore loads and saves TOML file correctly
- [ ] LRU eviction works at 1000 entry boundary
- [ ] File I/O errors do not crash the application
- [ ] Invalid entries in TOML are skipped gracefully

**Estimated Effort**: 小

---

### Phase 2: Header Sort Display

**Goal**: Display the current sort state (e.g., "Name ↑") in header Line2 between mark info and free space info.

**Files to Modify**:
- `internal/ui/pane_render.go` - Modify `renderHeaderLine2` to include sort info
- `internal/ui/pane_render_test.go` - Add tests for sort info display

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| renderHeaderLine2 (modified) | Include sort config string in header layout | Pane has valid sortConfig | Header shows "Marked X/Y Z B  Name ↑  W Free" |

**Processing Flow**:
```
renderHeaderLine2:
1. Build mark info string (existing)
2. Build sort info string from pane's sortConfig.String()
3. Build free space string (existing)
4. Layout: mark info (left) + sort info (center) + free space (right)
   └─ Distribute padding to position sort info between the other two
```

**Testing Approach**:

| Scenario | Expected Result | Type |
|----------|-----------------|------|
| Header contains sort info for default sort | "Name ↑" visible in output | Unit |
| Header contains sort info for Size desc | "Size ↓" visible in output | Unit |
| Header contains sort info for Date asc | "Date ↑" visible in output | Unit |
| Sort info positioned between mark info and free space | Correct layout | Unit |
| Narrow width still shows sort info | Sort info visible or gracefully truncated | Unit |

**Acceptance Criteria**:
- [ ] Header Line2 always shows current sort state
- [ ] All 6 sort config combinations display correctly
- [ ] Layout works across typical pane widths

**Estimated Effort**: 小

---

### Phase 3: Integration (Save on Confirm, Auto-Apply on Navigate)

**Goal**: Connect DirSortStore to the sort dialog confirmation flow and directory navigation, so settings are saved and automatically applied.

**Files to Modify**:
- `internal/ui/model.go` - Refactor NewModelWithConfig to Options pattern; add DirSortStore field
- `internal/ui/model_update_dialog.go` - Save to store on sort dialog confirmation
- `internal/ui/pane_navigation.go` - Apply saved sort config on directory entry
- `cmd/duofm/main.go` - Initialize DirSortStore and pass via ModelOptions

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| ModelOptions | Hold all configuration for Model initialization as a struct | Fields have zero-value defaults | Replaces long parameter list in NewModelWithConfig |
| Model.dirSortStore | Hold reference to DirSortStore | Set via ModelOptions | Available for save/lookup |
| sortDialogResultMsg handler (modified) | Save confirmed sort setting to store | Dialog confirmed, store available | Setting persisted for current directory |
| Navigation methods (modified) | Look up and apply saved sort config before loading directory | Store available, target path known | Pane sortConfig set to saved or default |

**NewModelWithConfig Refactoring**:
Introduce a `ModelOptions` struct to replace the current long parameter list. This prevents signature changes from affecting 100+ test calls for future additions. All existing tests using `NewModel()` remain unchanged as it wraps `NewModelWithConfig` with defaults.

**Navigation Methods to Modify**:
All methods that change the pane's current directory:
- `EnterDirectoryAsync` - Enter subdirectory or parent
- `MoveToParentAsync` - Navigate to parent directory
- `NavigateHistoryBackAsync` / `NavigateHistoryForwardAsync` - History navigation
- `NavigateToPreviousAsync` - Toggle previous directory (cd -)
- `NavigateToHomeAsync` - Navigate to home directory
- `ChangeDirectoryAsync` - Shell command cd
- `SyncTo` - Sync opposite pane to current directory

**Processing Flow**:
```
Sort Dialog Confirmed:
1. Existing handler applies sort config to pane (unchanged)
2. After applying, save to DirSortStore with current pane path
   └─ Store handles LRU eviction and file write internally

Directory Navigation (all methods listed above):
1. Before starting async load, look up target path in DirSortStore
   ├─ Found → set pane's sortConfig to saved value
   └─ Not found → set pane's sortConfig to DefaultSortConfig()
2. Proceed with async load using the (potentially updated) sortConfig
```

**Testing Approach**:

| Scenario | Expected Result | Type |
|----------|-----------------|------|
| Sort dialog confirm saves to store | Store contains entry for directory | Integration |
| Sort dialog cancel does not save | Store unchanged | Integration |
| Navigate to directory with saved setting | Pane sort config matches saved | Integration |
| Navigate to directory without saved setting | Pane sort config is default | Integration |
| All navigation methods apply saved sort | EnterDirectoryAsync, MoveToParentAsync, HistoryBack, HistoryForward, NavigateToPrevious, NavigateToHome, ChangeDirectoryAsync, SyncTo all work | Integration |
| Application restart preserves settings | Load after save returns same data | Integration |

**Acceptance Criteria**:
- [ ] Sort dialog confirmation persists setting to dir_sort.toml
- [ ] Directory navigation applies saved sort settings automatically
- [ ] Default sort applied when no saved setting exists
- [ ] All navigation methods (enter, history back/forward, cd -, home) handled
- [ ] Existing sort functionality (live preview, cancel restore) unchanged

**Estimated Effort**: 小

---

## Complete File Structure

```
internal/
├── config/
│   ├── dir_sort_store.go        # NEW: Per-directory sort settings store
│   ├── dir_sort_store_test.go   # NEW: Store tests
│   ├── config.go                # Unchanged
│   └── path.go                  # Unchanged (provides GetConfigDir)
├── ui/
│   ├── model.go                 # MODIFIED: Refactor to ModelOptions pattern, add dirSortStore field
│   ├── model_update_dialog.go   # MODIFIED: Save on sort dialog confirm
│   ├── pane_navigation.go       # MODIFIED: Apply saved sort on navigation
│   ├── pane_render.go           # MODIFIED: Add sort info to header Line2
│   ├── pane_render_test.go      # MODIFIED: Add header sort display tests
│   ├── sort.go                  # Unchanged
│   └── sort_dialog.go           # Unchanged
cmd/
└── duofm/
    └── main.go                  # MODIFIED: Initialize DirSortStore
```

## Testing Strategy

### Unit Testing

**Test Coverage Goals**:
- DirSortStore: 90%+ (critical data persistence)
- Header rendering: 80%+ (display logic)

**Key Test Areas**:

| Area | Scenarios | Priority |
|------|-----------|----------|
| Store CRUD | Save, Get, overwrite, delete (LRU) | High |
| Store persistence | Load from file, save to file, missing file, corrupted file | High |
| LRU eviction | Boundary at 1000, oldest removed, access time updated | High |
| Header display | All sort configs displayed, layout correct | Medium |
| Integration | Dialog confirm → save, navigate → apply | High |

### Manual Testing (E2E Not Possible)

- [ ] Start application, change sort in Downloads, restart, navigate to Downloads → sort is preserved
- [ ] Verify header Line2 shows sort info (visual check)
- [ ] Navigate between directories with different saved sorts → each directory shows its sort

## Dependencies

### Component Dependencies

**Implementation Order** (respecting dependencies):
1. Phase 1: DirSortStore (no internal dependencies, only config package)
2. Phase 2: Header display (no dependency on Phase 1)
3. Phase 3: Integration (depends on Phase 1, Phase 2 is independent)

Phases 1 and 2 can be implemented in parallel.

## Risk Assessment

### Technical Risks

1. **TOML key escaping for directory paths**
   - **Risk**: Paths with special characters (spaces, brackets) may need escaping in TOML table keys
   - **Likelihood**: Medium
   - **Impact**: Low (TOML handles quoted keys)
   - **Mitigation**: Use TOML library's built-in encoding which handles key escaping

2. **Concurrent file access**
   - **Risk**: Left and right panes could trigger simultaneous saves
   - **Likelihood**: Low (saves happen on sort dialog confirm which is serialized via Bubble Tea)
   - **Impact**: Low
   - **Mitigation**: Bubble Tea's single-threaded Update ensures serialization

## Open Questions

- None (all questions resolved during requirements gathering)

## References

- **Specification**: `doc/tasks/per-directory-sort/SPEC.md`
- **Existing sort implementation**: `internal/ui/sort.go`
- **Config path resolution**: `internal/config/path.go`
