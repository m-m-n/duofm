# Implementation Plan: Directory Bookmark

## Overview

Implement directory bookmark functionality allowing users to save, manage, and quickly jump to frequently accessed directories. The feature integrates with the existing configuration file system and UI dialog patterns.

## Objectives

- Enable users to bookmark directories with custom aliases
- Provide a bookmark manager dialog for viewing, editing, and deleting bookmarks
- Persist bookmarks in the configuration file across sessions
- Integrate with existing keybinding and dialog infrastructure

## Prerequisites

### Development Environment

- Go 1.21+
- Existing duofm development environment

### Dependencies

- BurntSushi/toml (already used for config)
- Bubble Tea, Lip Gloss (already used for UI)

### Knowledge Requirements

- Understanding of duofm's Dialog interface pattern
- Familiarity with config package structure
- TOML array format (`[[section]]`)

## Architecture Overview

### Technology Stack

- **Language**: Go 1.21+
- **Framework**: Bubble Tea
- **Styling**: Lip Gloss
- **Configuration**: BurntSushi/toml

### Design Approach

The bookmark feature follows the existing duofm architecture patterns:

1. **Config Layer**: Bookmark data model and persistence in `internal/config`
2. **UI Layer**: Two dialog components in `internal/ui`
   - BookmarkDialog: List-based manager for viewing/navigating bookmarks
   - InputDialog: Reuse existing component for alias input
3. **Integration**: Wire into Model and keybinding system

### Component Interaction

```
User Input (B/Shift+B)
        │
        ▼
   Model.Update()
        │
        ├─ b key ──────► BookmarkDialog (list view)
        │                      │
        │                      ├─ Enter ──► Jump to directory
        │                      ├─ d ──────► Delete bookmark
        │                      └─ e ──────► Edit alias (InputDialog)
        │
        └─ Shift+B ────► InputDialog (add bookmark)
                               │
                               └─ Enter ──► Save to config
```

## Implementation Phases

### Phase 1: Bookmark Data Model and Configuration

**Goal**: Define bookmark data structure and implement config file read/write for bookmarks

**Files to Create**:

- `internal/config/bookmark.go` - Bookmark data model and operations
- `internal/config/bookmark_test.go` - Unit tests

**Files to Modify**:

- `internal/config/config.go`:
  - Add Bookmarks field to Config struct
  - Integrate bookmark loading in LoadConfig
- `internal/config/generator.go`:
  - Add bookmarks section comment to default config template

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Bookmark struct | Hold single bookmark data (name, path) | - | Contains valid name and path |
| LoadBookmarks | Parse bookmarks from TOML config | Config file exists or not | Returns bookmark slice (empty if none) |
| SaveBookmarks | Write bookmarks to config file | Valid bookmarks slice | Config file updated with bookmarks |
| AddBookmark | Add new bookmark to beginning of list | Path not already bookmarked | Bookmark prepended to list |
| RemoveBookmark | Remove bookmark by index | Valid index | Bookmark removed from list |
| UpdateBookmarkAlias | Change bookmark alias | Valid index, non-empty alias | Alias updated |
| IsPathBookmarked | Check if path exists in bookmarks | - | Returns true/false |

**Processing Flow**:

```
LoadBookmarks:
1. Parse TOML config file
2. Extract [[bookmarks]] array entries
   ├─ Entry valid → Add to result slice
   └─ Entry invalid → Log warning, skip entry
3. Return bookmark slice

SaveBookmarks:
1. Read existing config file content
2. Remove existing [[bookmarks]] sections
3. Append new [[bookmarks]] entries
4. Write updated content to file
```

**Implementation Steps**:

1. **Define Bookmark struct**
   - Name (alias) and Path fields
   - TOML struct tags for serialization

2. **Implement LoadBookmarks**
   - Parse TOML array format
   - Handle missing or empty bookmarks section
   - Validate entries (skip invalid)

3. **Implement SaveBookmarks**
   - Preserve existing config sections (keybindings, colors)
   - Write bookmarks in correct TOML array format
   - Handle file I/O errors

4. **Implement bookmark operations**
   - Add, remove, update, check duplicate

**Dependencies**:

- Requires: Existing config infrastructure
- Blocks: Phase 2 (UI needs data model)

**Testing Approach**:

*Unit Tests*:

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Load empty config | No bookmarks section | Empty slice |
| Load valid bookmarks | Valid [[bookmarks]] entries | Populated slice |
| Load invalid entry | Entry missing name/path | Warning, entry skipped |
| Save bookmarks | Bookmark slice | Valid TOML written |
| Add duplicate path | Existing path | Error returned |
| Remove by index | Valid index | Bookmark removed |

**Acceptance Criteria**:

- [ ] Bookmark struct defined with TOML tags
- [ ] LoadBookmarks correctly parses [[bookmarks]] array
- [ ] SaveBookmarks writes valid TOML format
- [ ] Existing config sections preserved after save
- [ ] Duplicate path detection works
- [ ] All unit tests pass

**Estimated Effort**: 小 (1-2 days)

---

### Phase 2: Bookmark Manager Dialog

**Goal**: Create a dialog for viewing, selecting, editing, and deleting bookmarks

**Files to Create**:

- `internal/ui/bookmark_dialog.go` - Bookmark manager dialog component
- `internal/ui/bookmark_dialog_test.go` - Unit tests

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| BookmarkDialog struct | Hold dialog state (bookmarks, cursor, mode) | - | Valid dialog state |
| NewBookmarkDialog | Create dialog with bookmark list | Bookmarks loaded | Dialog ready for display |
| Update | Handle keyboard input (j/k/Enter/D/E/Esc) | Dialog active | State updated, commands returned |
| View | Render bookmark list with two-line format | - | Styled string output |
| checkPathExists | Verify bookmark path exists on filesystem | - | Sets exists flag per bookmark |

**Processing Flow**:

```
BookmarkDialog Update:
1. Receive key message
   ├─ j/k → Move cursor within bounds
   ├─ Enter → If path exists, return jump command
   ├─ d → Return delete command with current index
   ├─ e → Return edit command with current bookmark
   └─ Esc → Deactivate dialog, return nil

View Rendering:
1. For each bookmark:
   ├─ Line 1: Alias (with warning emoji if path not exists)
   └─ Line 2: Path (wrapped if exceeds width)
2. Highlight current cursor position
3. Render footer with key hints
```

**Implementation Steps**:

1. **Define BookmarkDialog struct**
   - Bookmarks slice, cursor position, width
   - Path existence cache for warning indicators

2. **Implement Update for navigation**
   - j/k keys for cursor movement
   - Boundary checking (wrap or clamp)

3. **Implement Update for actions**
   - Enter: Jump command (if path exists)
   - D: Delete command
   - E: Edit mode trigger
   - Esc: Close dialog

4. **Implement View rendering**
   - Two-line format per bookmark
   - Path wrapping for long paths
   - Warning emoji for non-existent paths
   - Cursor highlighting
   - Footer with key hints

5. **Implement path existence check**
   - Check each bookmark path on dialog open
   - Cache results for display

**Dependencies**:

- Requires: Phase 1 (Bookmark data model)
- Blocks: Phase 3 (Integration)

**Testing Approach**:

*Unit Tests*:

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Navigate down | j key | Cursor incremented |
| Navigate up | k key | Cursor decremented |
| Navigate wrap | j at last item | Cursor to 0 |
| Select valid bookmark | Enter on existing path | Jump message |
| Select invalid bookmark | Enter on non-existent path | No action |
| Delete bookmark | d key | Delete message with index |
| Edit bookmark | e key | Edit message with bookmark |
| Close dialog | Esc | Dialog deactivated |
| Empty bookmarks view | No bookmarks | "No bookmarks" message |

**Acceptance Criteria**:

- [ ] j/k navigation works correctly
- [ ] Two-line display format renders properly
- [ ] Long paths wrap correctly
- [ ] Warning indicator shows for non-existent paths
- [ ] Enter jumps only to existing paths
- [ ] D triggers delete action
- [ ] E triggers edit mode
- [ ] Esc closes dialog
- [ ] Empty state handled gracefully
- [ ] All unit tests pass

**Estimated Effort**: 中 (3-5 days)

---

### Phase 3: Add Bookmark Dialog and Integration

**Goal**: Implement add bookmark flow and integrate all components into main Model

**Files to Modify**:

- `internal/ui/model.go`:
  - Add bookmark-related fields
  - Handle B and Shift+B keys
  - Process bookmark dialog results
- `internal/ui/keys.go`:
  - Add bookmark key constants
- `internal/config/defaults.go`:
  - Add bookmark keybindings to defaults
- `internal/ui/help_dialog.go`:
  - Add "Bookmarks" section with B and Shift+B keybindings

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Model.bookmarks | Hold current bookmarks list | - | Synced with config |
| handleBookmarkKey | Open bookmark manager dialog | b key pressed | Dialog displayed |
| handleAddBookmarkKey | Open add bookmark dialog | Shift+B pressed | InputDialog displayed |
| processBookmarkResult | Handle dialog result messages | Dialog closed | Action executed (jump/delete/edit/save) |

**Processing Flow**:

```
B Key (Open Manager):
1. Load bookmarks from config
2. Create BookmarkDialog with bookmarks
3. Set as active dialog

Shift+B Key (Add Bookmark):
1. Get current directory path from active pane
2. Check if already bookmarked
   ├─ Yes → Show "Already bookmarked" status message
   └─ No → Continue
3. Derive default alias from path basename
4. Create InputDialog with default alias
5. Set as active dialog

Bookmark Jump Result:
1. Get target path from result
2. Navigate active pane to path
3. Close dialog

Bookmark Delete Result:
1. Get index from result
2. Remove bookmark from list
3. Save bookmarks to config
4. Update dialog or close if empty

Bookmark Edit Result:
1. Get index and new alias from result
2. Update bookmark alias
3. Save bookmarks to config
4. Refresh dialog view

Add Bookmark Confirm:
1. Get alias from input
2. Create bookmark with alias and current path
3. Add to bookmarks (prepend)
4. Save to config
5. Show success status message
```

**Implementation Steps**:

1. **Add keybinding constants and defaults**
   - KeyBookmark and KeyAddBookmark constants
   - Default keybindings (B, Shift+B)

2. **Add bookmark state to Model**
   - Bookmarks slice field
   - Load bookmarks on init

3. **Implement b key handler**
   - Create and display BookmarkDialog
   - Pass current bookmarks

4. **Implement Shift+b key handler**
   - Check duplicate path
   - Create InputDialog with default alias
   - Set appropriate callback

5. **Implement result message handlers**
   - Jump: Navigate pane
   - Delete: Remove and save
   - Edit: Update and save
   - Add: Create, prepend, save

6. **Add status messages**
   - Success: "Bookmarked: {alias}"
   - Duplicate: "Already bookmarked"
   - Delete: "Removed: {alias}"
   - Error: Config save failure

**Dependencies**:

- Requires: Phase 1 (data model), Phase 2 (dialog)
- Blocks: None

**Testing Approach**:

*Integration Tests*:

| Test Case | Steps | Expected Result |
|-----------|-------|-----------------|
| Open bookmark dialog | Press b | Dialog displayed with bookmarks |
| Jump to bookmark | b → select → Enter | Pane navigated to path |
| Delete bookmark | b → select → d | Bookmark removed from list |
| Add new bookmark | Shift+b → enter alias → Enter | Bookmark added to config |
| Add duplicate | Shift+b on already bookmarked | Status shows "Already bookmarked" |
| Edit bookmark alias | b → select → e → edit → Enter | Alias updated in config |

*E2E Tests*:

- [ ] Full flow: Add bookmark → view in list → jump → delete
- [ ] Config persistence: Add bookmark → restart → bookmark exists
- [ ] Non-existent path handling: Delete directory → warning shown

**Acceptance Criteria**:

- [ ] b key opens bookmark manager
- [ ] Shift+B opens add dialog with default alias
- [ ] Jumping to bookmark navigates active pane
- [ ] Delete removes bookmark and updates config
- [ ] Edit updates alias and saves config
- [ ] Duplicate detection shows appropriate message
- [ ] Status messages display correctly
- [ ] All integration tests pass

**Estimated Effort**: 中 (3-5 days)

---

## Complete File Structure

```
internal/
├── config/
│   ├── bookmark.go          # NEW: Bookmark data model and operations
│   ├── bookmark_test.go     # NEW: Bookmark unit tests
│   ├── config.go            # MODIFY: Add Bookmarks field
│   ├── defaults.go          # MODIFY: Add bookmark keybindings
│   └── generator.go         # MODIFY: Add bookmarks section comment
└── ui/
    ├── bookmark_dialog.go   # NEW: Bookmark manager dialog
    ├── bookmark_dialog_test.go # NEW: Dialog unit tests
    ├── keys.go              # MODIFY: Add bookmark key constants
    └── model.go             # MODIFY: Integrate bookmark handling
```

**File Descriptions**:

- `internal/config/bookmark.go`: Bookmark struct definition, CRUD operations, config read/write
- `internal/ui/bookmark_dialog.go`: List-based dialog for bookmark management with two-line display
- Modified files integrate the new components into existing infrastructure

## Testing Strategy

### Unit Testing

**Approach**:

- Table-driven tests for all bookmark operations
- Mock filesystem for path existence checks
- Test edge cases (empty list, single item, many items)

**Test Coverage Goals**:

- Config operations: 90%+ (critical for data integrity)
- Dialog logic: 80%+ (state transitions, rendering)

**Key Test Areas**:

| Area | Test Focus |
|------|------------|
| Bookmark CRUD | Add, remove, update, duplicate detection |
| Config persistence | Load, save, preserve other sections |
| Dialog navigation | Cursor movement, boundary behavior |
| Dialog actions | Jump, delete, edit message generation |
| Path validation | Existence check, warning display |

### Manual Testing Checklist

- [ ] b key opens bookmark manager with existing bookmarks
- [ ] j/k navigates up/down in bookmark list
- [ ] Enter jumps to selected bookmark directory
- [ ] D deletes selected bookmark
- [ ] E opens edit dialog with current alias pre-filled
- [ ] Shift+B opens add bookmark dialog
- [ ] Default alias is directory basename
- [ ] Duplicate path shows "Already bookmarked" message
- [ ] Non-existent paths show warning indicator
- [ ] Cannot jump to non-existent path
- [ ] Bookmarks persist after application restart
- [ ] Long paths display correctly (wrapped)

## Dependencies

### External Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/BurntSushi/toml | existing | TOML parsing |
| github.com/charmbracelet/bubbletea | existing | TUI framework |
| github.com/charmbracelet/lipgloss | existing | Styling |

### Internal Dependencies

**Implementation Order**:

1. Phase 1 - Config (no dependencies)
2. Phase 2 - Dialog (depends on Phase 1)
3. Phase 3 - Integration (depends on Phases 1 & 2)

**Component Dependencies**:

- `bookmark_dialog.go` depends on `config/bookmark.go`
- `model.go` bookmark handling depends on both above

## Risk Assessment

### Technical Risks

1. **Config File Corruption**
   - **Risk**: Saving bookmarks might corrupt existing config sections
   - **Likelihood**: Medium
   - **Impact**: High (user loses keybindings/colors)
   - **Mitigation**:
     - Parse and rewrite full config structure
     - Test with various existing config formats
     - Create backup before save

2. **Long Path Display**
   - **Risk**: Very long paths might break layout
   - **Likelihood**: Low
   - **Impact**: Low (visual only)
   - **Mitigation**:
     - Implement proper text wrapping
     - Test with maximum path lengths

## Success Metrics

### Functional Completeness

- [ ] All functional requirements (FR-1 through FR-8) implemented
- [ ] All test scenarios pass
- [ ] Error handling works correctly

### Quality Metrics

- [ ] Test coverage meets goals (90%+ config, 80%+ dialog)
- [ ] No critical bugs in manual testing
- [ ] Code follows existing duofm patterns

### User Experience

- [ ] Dialog is intuitive (follows existing patterns)
- [ ] Warning indicators are clear
- [ ] Status messages are informative

## References

- **Specification**: `doc/tasks/bookmark/SPEC.md`
- **Similar Components**:
  - `internal/ui/context_menu_dialog.go` (list-based dialog pattern)
  - `internal/ui/input_dialog.go` (text input pattern)
  - `internal/config/config.go` (config loading pattern)
