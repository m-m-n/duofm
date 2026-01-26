# Feature: Cancel Key Unification

## Overview

Unify cancel key handling across all dialog components in duofm. Currently, some dialogs only support Esc, while others support Esc/Ctrl+C or Esc/q. This feature ensures all dialogs can be cancelled with both Esc and Ctrl+C for consistent UX.

## Objectives

- Standardize cancel key handling across all dialogs to support both Esc and Ctrl+C
- Remove legacy cancel keys (q in TrashDialog/SortDialog, y/n in ArchiveWarningDialog) to unify UX
- Provide consistent user experience regardless of which dialog is displayed

## User Stories

### US1: Cancel Dialog with Esc
As a user, I want to cancel any dialog by pressing Esc, so that I can quickly close dialogs using a familiar key.

**Acceptance Criteria:**
- [ ] All dialogs close when Esc is pressed
- [ ] No changes are applied when dialog is cancelled

### US2: Cancel Dialog with Ctrl+C
As a user, I want to cancel any dialog by pressing Ctrl+C, so that I can use the standard terminal interrupt key.

**Acceptance Criteria:**
- [ ] All dialogs close when Ctrl+C is pressed
- [ ] No changes are applied when dialog is cancelled

## Technical Requirements

### Functional Requirements
- **FR1:** All dialog components must handle both "esc" and "ctrl+c" key events for cancellation
- **FR2:** Legacy cancel keys (q in TrashDialog/SortDialog, y/n in ArchiveWarningDialog) must be removed
- **FR3:** Cancel action must not apply any pending changes from the dialog
- **FR4:** ArchiveWarningDialog must use Tab/Arrow for button selection and Enter for confirmation

### Non-Functional Requirements
- **NFR1 - Performance:** Key handling response time must be imperceptible (<50ms)
- **NFR2 - Consistency:** All dialogs must use the same pattern for cancel key handling

## Implementation Approach

### Current State Analysis

**Dialogs already supporting Esc/Ctrl+C:**
- `confirm_dialog.go`
- `error_dialog.go`
- `help_dialog.go`
- `context_menu_dialog.go`
- `permission_progress_dialog.go`

**Dialogs requiring Ctrl+C addition:**
| File | Current Keys | Action |
|------|--------------|--------|
| `input_dialog.go` | Esc | Add Ctrl+C |
| `path_jump_dialog.go` | Esc | Add Ctrl+C |
| `archive_name_dialog.go` | Esc | Add Ctrl+C |
| `recursive_perm_dialog.go` | Esc | Add Ctrl+C |
| `bookmark_dialog.go` | Esc | Add Ctrl+C |
| `regex_search_dialog.go` | Esc | Add Ctrl+C |
| `rename_input_dialog.go` | Esc | Add Ctrl+C |
| `archive_progress_dialog.go` | Esc | Add Ctrl+C |
| `compression_level_dialog.go` | Esc | Add Ctrl+C |
| `permission_error_report_dialog.go` | Esc | Add Ctrl+C |
| `permission_dialog.go` | Esc | Add Ctrl+C |
| `query_search_dialog.go` | Esc | Add Ctrl+C |
| `extension_rename_dialog.go` | Esc | Add Ctrl+C |
| `trash_dialog.go` | Esc, q | Add Ctrl+C, remove q |
| `sort_dialog.go` | Esc, q | Add Ctrl+C, remove q |
| `archive_warning_dialog.go` | Esc, n, y | Add Ctrl+C, remove n/y |

### Implementation Pattern

For each dialog, modify the `Update` method to handle both keys:

**Before (Esc only):**
```go
case "esc":
    return d, CancelDialogCmd()
```

**After (Esc and Ctrl+C):**
```go
case "esc", "ctrl+c":
    return d, CancelDialogCmd()
```

### File Structure

All changes are within `internal/ui/`:

```
internal/ui/
├── input_dialog.go              # Add ctrl+c
├── path_jump_dialog.go          # Add ctrl+c
├── archive_name_dialog.go       # Add ctrl+c
├── recursive_perm_dialog.go     # Add ctrl+c
├── bookmark_dialog.go           # Add ctrl+c
├── regex_search_dialog.go       # Add ctrl+c
├── rename_input_dialog.go       # Add ctrl+c
├── archive_progress_dialog.go   # Add ctrl+c
├── compression_level_dialog.go  # Add ctrl+c
├── permission_error_report_dialog.go  # Add ctrl+c
├── permission_dialog.go         # Add ctrl+c
├── query_search_dialog.go       # Add ctrl+c
├── extension_rename_dialog.go   # Add ctrl+c
├── trash_dialog.go              # Add ctrl+c, remove q
├── sort_dialog.go               # Add ctrl+c, remove q
└── archive_warning_dialog.go    # Add ctrl+c, remove y/n
```

## Test Scenarios

### Unit Tests
- [ ] Test each dialog cancels on "esc" key
- [ ] Test each dialog cancels on "ctrl+c" key
- [ ] Test ArchiveWarningDialog confirms with Enter (after selecting Continue)
- [ ] Test ArchiveWarningDialog cancels with Enter (after selecting Cancel)

### E2E Tests
- [ ] Open each dialog type and verify Esc closes it
- [ ] Open each dialog type and verify Ctrl+C closes it

### Test Cases by Dialog

| Dialog | Test Esc | Test Ctrl+C | Test q |
|--------|----------|-------------|--------|
| InputDialog | Yes | Yes | No |
| PathJumpDialog | Yes | Yes | No |
| ArchiveNameDialog | Yes | Yes | No |
| RecursivePermDialog | Yes | Yes | No |
| BookmarkDialog | Yes | Yes | No |
| RegexSearchDialog | Yes | Yes | No |
| RenameInputDialog | Yes | Yes | No |
| ArchiveProgressDialog | Yes | Yes | No |
| CompressionLevelDialog | Yes | Yes | No |
| PermissionErrorReportDialog | Yes | Yes | No |
| PermissionDialog | Yes | Yes | No |
| QuerySearchDialog | Yes | Yes | No |
| ExtensionRenameDialog | Yes | Yes | No |
| TrashDialog | Yes | Yes | No |
| SortDialog | Yes | Yes | No |
| ArchiveWarningDialog | Yes | Yes | No |
| ConfirmDialog | Yes (existing) | Yes (existing) | No |
| ErrorDialog | Yes (existing) | Yes (existing) | No |
| HelpDialog | Yes (existing) | Yes (existing) | No |
| ContextMenuDialog | Yes (existing) | Yes (existing) | No |
| PermissionProgressDialog | Yes (existing) | Yes (existing) | No |

## Success Criteria

- [ ] All 16 dialogs requiring changes are updated
- [ ] All dialogs respond to both Esc and Ctrl+C for cancellation
- [ ] Legacy cancel keys (q, y, n) are removed
- [ ] Unit tests pass for all dialog cancel scenarios
- [ ] E2E tests verify cancel behavior works correctly

## Implementation Checklist

- [ ] `input_dialog.go` - Add "ctrl+c" case
- [ ] `path_jump_dialog.go` - Add "ctrl+c" case
- [ ] `archive_name_dialog.go` - Add "ctrl+c" case
- [ ] `recursive_perm_dialog.go` - Add "ctrl+c" case
- [ ] `bookmark_dialog.go` - Add "ctrl+c" case
- [ ] `regex_search_dialog.go` - Add "ctrl+c" case
- [ ] `rename_input_dialog.go` - Add "ctrl+c" case
- [ ] `archive_progress_dialog.go` - Add "ctrl+c" case
- [ ] `compression_level_dialog.go` - Add "ctrl+c" case
- [ ] `permission_error_report_dialog.go` - Add "ctrl+c" case
- [ ] `permission_dialog.go` - Add "ctrl+c" case
- [ ] `query_search_dialog.go` - Add "ctrl+c" case
- [ ] `extension_rename_dialog.go` - Add "ctrl+c" case
- [ ] `trash_dialog.go` - Add "ctrl+c" case, remove "q"
- [ ] `sort_dialog.go` - Add "ctrl+c" case, remove "q"
- [ ] `archive_warning_dialog.go` - Add "ctrl+c" case, remove "y"/"n"
- [ ] Write/update unit tests
- [ ] Run full test suite
