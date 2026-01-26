# Implementation Plan: Cancel Key Unification

## Overview

Unify cancel key handling across all dialog components in duofm to support both Esc and Ctrl+C for consistent UX.

## Objectives

- Standardize cancel key handling across all dialogs to support both Esc and Ctrl+C
- Remove legacy cancel keys (q in TrashDialog/SortDialog, y/n in ArchiveWarningDialog) for UX unification
- Provide consistent user experience regardless of which dialog is displayed

## Prerequisites

### Development Environment
- Go 1.21+
- duofm development environment set up

### Dependencies
- No external dependencies required
- All changes are within existing codebase

### Knowledge Requirements
- Bubble Tea Update function pattern
- Go switch/case statement syntax

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea (TUI framework)

### Design Approach

All dialogs implement a common pattern for key handling in their `Update` method. The modification involves extending the cancel key cases to include both `esc` and `ctrl+c`.

**Pattern Categories:**

| Category | Pattern | Example |
|----------|---------|---------|
| Type A | `case tea.KeyEsc:` | InputDialog, PathJumpDialog, BookmarkDialog |
| Type B | `case "esc":` | SortDialog.HandleKey |
| Type C | `case "esc", "n":` | ArchiveWarningDialog |

### Component Interaction

Each dialog component independently handles key events in its `Update` method. No inter-component dependencies exist for this change.

## Implementation Phases

### Phase 1: Dialogs using tea.KeyEsc

**Goal**: Add Ctrl+C support to dialogs that use `tea.KeyEsc` pattern

**Files to Modify**:

| File | Current Pattern | Change |
|------|-----------------|--------|
| `input_dialog.go` | `case tea.KeyEsc:` | Add `case tea.KeyEsc, tea.KeyCtrlC:` |
| `path_jump_dialog.go` | `case tea.KeyEsc:` | Add `case tea.KeyEsc, tea.KeyCtrlC:` |
| `archive_name_dialog.go` | `case tea.KeyEsc:` | Add `case tea.KeyEsc, tea.KeyCtrlC:` |
| `recursive_perm_dialog.go` | `case tea.KeyEsc:` | Add `case tea.KeyEsc, tea.KeyCtrlC:` |
| `bookmark_dialog.go` | `case tea.KeyEsc:` | Add `case tea.KeyEsc, tea.KeyCtrlC:` |
| `regex_search_dialog.go` | `case tea.KeyEsc:` | Add `case tea.KeyEsc, tea.KeyCtrlC:` |
| `rename_input_dialog.go` | `case tea.KeyEsc:` | Add `case tea.KeyEsc, tea.KeyCtrlC:` |
| `archive_progress_dialog.go` | `case tea.KeyEsc:` | Add `case tea.KeyEsc, tea.KeyCtrlC:` |
| `compression_level_dialog.go` | `case tea.KeyEsc:` | Add `case tea.KeyEsc, tea.KeyCtrlC:` |
| `permission_error_report_dialog.go` | `case tea.KeyEsc:` or combined | Add `tea.KeyCtrlC` |
| `permission_dialog.go` | `case tea.KeyEsc:` | Add `case tea.KeyEsc, tea.KeyCtrlC:` |
| `query_search_dialog.go` | `case tea.KeyEsc:` | Add `case tea.KeyEsc, tea.KeyCtrlC:` |
| `extension_rename_dialog.go` | `case tea.KeyEsc:` | Add `case tea.KeyEsc, tea.KeyCtrlC:` |
| `trash_dialog.go` | `case tea.KeyEsc:` | Add `case tea.KeyEsc, tea.KeyCtrlC:` |

**Note on trash_dialog.go**: This dialog previously had two separate cancel paths:
1. `case "q":` under `tea.KeyRunes` - Remove (UX unification)
2. `case tea.KeyEsc:` - Modify to `case tea.KeyEsc, tea.KeyCtrlC:`

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Update method | Handle key events and close dialog | Key event received | Dialog closed, cancel message sent |

**Processing Flow**:
```
1. Key event received in Update method
2. Check if key matches cancel pattern
   |-- Esc key -> Close dialog, send cancel message
   |-- Ctrl+C key -> Close dialog, send cancel message
   |-- Other keys -> Process normally
3. Return dialog state and command
```

**Implementation Steps**:

1. **Modify each dialog's Update method**
   - Locate the `case tea.KeyEsc:` statement
   - Change to `case tea.KeyEsc, tea.KeyCtrlC:`

**Dependencies**:
- Requires: None
- Blocks: Phase 3 (Testing)

**Testing Approach**:

*Unit Tests*:
- Test dialog cancels on Esc key
- Test dialog cancels on Ctrl+C key
- Test cancel action does not apply changes

**Acceptance Criteria**:
- [ ] All 14 dialogs modified to handle tea.KeyCtrlC
- [ ] Existing Esc behavior unchanged
- [ ] Build succeeds without errors

**Estimated Effort**: Small (1-2 days)

---

### Phase 2: Dialogs using string-based key matching

**Goal**: Add Ctrl+C support to dialogs using string-based key matching

**Files to Modify**:

| File | Current Pattern | Change |
|------|-----------------|--------|
| `sort_dialog.go` | `case "esc":` in handleDialogKey + q key handler | Add "ctrl+c", remove q key |
| `archive_warning_dialog.go` | `case "esc", "n":` + y key | Change to `case "esc", "ctrl+c":`, remove n/y |

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| SortDialog.handleDialogKey | Handle string-based key events | Key string received | Dialog cancelled if escape/ctrl+c |
| ArchiveWarningDialog.Update | Handle key events for warnings | Key event received | Dialog closed, cancel choice sent |

**Processing Flow**:
```
1. Key event converted to string in switch statement
2. Match against cancel key patterns
   |-- "esc" -> Cancel dialog
   |-- "ctrl+c" -> Cancel dialog
3. Execute cancel behavior
```

**Implementation Steps**:

1. **Modify SortDialog**
   - Find `case "esc":` in handleDialogKey method, change to `case "esc", "ctrl+c":`
   - Remove `if key == "q"` handler at the beginning of HandleKey
   - Update footer to show "Esc:quit" instead of "q:quit"

2. **Modify ArchiveWarningDialog.Update**
   - Find `case "esc", "n":` in Update method, change to `case "esc", "ctrl+c":`
   - Remove `case "y":` handler
   - Update footer to show "[Enter] Confirm [Esc] Cancel" instead of "[y] Continue [n/Esc] Cancel"

**Dependencies**:
- Requires: None (can run in parallel with Phase 1)
- Blocks: Phase 3 (Testing)

**Testing Approach**:

*Unit Tests*:
- Test SortDialog cancels on Esc and Ctrl+C keys
- Test ArchiveWarningDialog cancels on Esc and Ctrl+C keys
- Test ArchiveWarningDialog confirms with Enter (after button selection)

**Acceptance Criteria**:
- [ ] SortDialog handles ctrl+c in handleDialogKey, q key removed
- [ ] ArchiveWarningDialog handles ctrl+c in Update, n/y keys removed
- [ ] Footer text updated to reflect new key bindings

**Estimated Effort**: Small (< 1 day)

---

### Phase 3: Testing and Verification

**Goal**: Verify all dialogs correctly handle both Esc and Ctrl+C

**Files to Create/Modify**:
- Update existing dialog tests if present
- No new test files required (existing test patterns sufficient)

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Test functions | Verify cancel key behavior | Dialogs modified | All tests pass |

**Processing Flow**:
```
1. Run existing test suite
2. Verify no regressions
3. Manual testing of each dialog type
```

**Implementation Steps**:

1. **Run full test suite**
   - Execute `go test ./...`
   - Verify all existing tests pass

2. **Manual verification**
   - Test each dialog with both Esc and Ctrl+C
   - Verify cancel behavior matches expectations

**Dependencies**:
- Requires: Phase 1 and Phase 2 complete
- Blocks: None

**Testing Approach**:

*Unit Tests*:
- Existing tests should continue to pass

*Manual Testing*:
- [ ] Open each dialog type and verify Esc closes it
- [ ] Open each dialog type and verify Ctrl+C closes it
- [ ] Verify ArchiveWarningDialog can select buttons with Tab/Arrow and confirm with Enter

**Acceptance Criteria**:
- [ ] All unit tests pass
- [ ] Manual testing confirms expected behavior
- [ ] No regressions in existing functionality

**Estimated Effort**: Small (< 1 day)

---

## Complete File Structure

```
internal/ui/
+-- input_dialog.go              # Add tea.KeyCtrlC
+-- path_jump_dialog.go          # Add tea.KeyCtrlC
+-- archive_name_dialog.go       # Add tea.KeyCtrlC
+-- recursive_perm_dialog.go     # Add tea.KeyCtrlC
+-- bookmark_dialog.go           # Add tea.KeyCtrlC
+-- regex_search_dialog.go       # Add tea.KeyCtrlC
+-- rename_input_dialog.go       # Add tea.KeyCtrlC
+-- archive_progress_dialog.go   # Add tea.KeyCtrlC
+-- compression_level_dialog.go  # Add tea.KeyCtrlC
+-- permission_error_report_dialog.go  # Add tea.KeyCtrlC
+-- permission_dialog.go         # Add tea.KeyCtrlC
+-- query_search_dialog.go       # Add tea.KeyCtrlC
+-- extension_rename_dialog.go   # Add tea.KeyCtrlC
+-- trash_dialog.go              # Add tea.KeyCtrlC, remove q
+-- sort_dialog.go               # Add "ctrl+c" to handleDialogKey, remove q
+-- archive_warning_dialog.go    # Add "ctrl+c" to Update, remove n/y
```

**File Descriptions**:
- Each dialog file contains an `Update` method that handles key events
- Changes are isolated to the switch/case statements for cancel keys
- No structural changes to dialog components

## Testing Strategy

### Unit Testing

**Approach**:
- Use Go's built-in `testing` package
- Test key event handling in Update methods

**Test Coverage Goals**:
- Cancel key handling: 100% coverage for modified code paths

**Key Test Areas**:
1. **Cancel Key Handling**
   - Test Esc key closes dialog
   - Test Ctrl+C key closes dialog
   - Test cancel does not apply changes

2. **ArchiveWarningDialog Button Selection**
   - Test Tab/Arrow navigation between buttons
   - Test Enter confirms selected button

### Manual Testing Checklist

| Dialog | Test Esc | Test Ctrl+C | Test q | Test other |
|--------|----------|-------------|--------|------------|
| InputDialog | Yes | Yes | - | - |
| PathJumpDialog | Yes | Yes | - | - |
| ArchiveNameDialog | Yes | Yes | - | - |
| RecursivePermDialog | Yes | Yes | - | - |
| BookmarkDialog | Yes | Yes | - | - |
| RegexSearchDialog | Yes | Yes | - | - |
| RenameInputDialog | Yes | Yes | - | - |
| ArchiveProgressDialog | Yes | Yes | - | - |
| CompressionLevelDialog | Yes | Yes | - | - |
| PermissionErrorReportDialog | Yes | Yes | - | - |
| PermissionDialog | Yes | Yes | - | - |
| QuerySearchDialog | Yes | Yes | - | - |
| ExtensionRenameDialog | Yes | Yes | - | - |
| TrashDialog | Yes | Yes | - | - |
| SortDialog | Yes | Yes | - | - |
| ArchiveWarningDialog | Yes | Yes | - | Enter, Tab |

## Dependencies

### External Dependencies

None required.

### Internal Dependencies

**Implementation Order**:
1. Phase 1 and Phase 2 (can be done in parallel)
2. Phase 3 (depends on Phase 1 and 2)

## Risk Assessment

### Technical Risks

1. **tea.KeyCtrlC constant availability**
   - **Risk**: tea.KeyCtrlC may not be the correct constant name
   - **Likelihood**: Low (standard Bubble Tea constant)
   - **Impact**: Low (easy to verify and fix)
   - **Mitigation**: Verify constant name in Bubble Tea documentation

### Implementation Risks

1. **Scope Creep**
   - **Risk**: Adding features beyond spec
   - **Mitigation**: Stick to spec, only modify cancel key handling

## Performance Considerations

None. Key handling is already O(1) and this change does not affect performance.

## Security Considerations

None. This change only affects UI key handling.

## Open Questions

### From Specification:
- None

### Implementation-Specific:
- [x] `model_update_dialog.go` - Confirmed this is a message handler, not a dialog component. No changes needed.

## Success Metrics

### Functional Completeness
- [ ] All 16 dialogs handle both Esc and Ctrl+C for cancellation
- [ ] Existing key bindings (q, n, y) continue to work
- [ ] No regressions in cancel behavior

### Quality Metrics
- [ ] All tests pass
- [ ] Code follows existing patterns

## References

- **Specification**: `doc/tasks/cancel-key-unification/SPEC.md`
- **Requirements**: `doc/tasks/cancel-key-unification/要件定義書.md`
- **Bubble Tea Key Types**: github.com/charmbracelet/bubbletea

## Next Steps

1. Review and confirm this implementation plan
2. Begin implementation with Phase 1
3. Complete Phase 2 (can be done in parallel)
4. Run Phase 3 testing and verification
5. Commit changes with appropriate commit message
