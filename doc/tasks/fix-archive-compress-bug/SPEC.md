# Bugfix: Context Menu Compress Action Not Triggering

## Overview

Fix a bug where selecting "Compress" from the context menu (`@` key) does nothing instead of opening the CompressFormatDialog.

**Bug Type**: Logic Error
**Severity**: High (Core functionality broken)
**Affected Version**: Current (commit d0e21b4)

---

## Problem Statement

When a user selects "Compress" from the context menu, the expected CompressFormatDialog does not appear. The UI simply returns to the normal state without any action.

### Root Cause

In `model.go`, the message handler for `contextMenuResultMsg` checks `result.action != nil` before processing action IDs. However, the "compress" menu item in `context_menu_dialog.go` sets `Action: nil`, causing the compress handling code to be unreachable.

**context_menu_dialog.go:176-180**
```go
items = append(items, MenuItem{
    ID:      "compress",
    Label:   compressLabel,
    Action:  nil,  // Action is nil
    Enabled: true,
})
```

**model.go:166-196**
```go
if result.action != nil {  // compress fails this check
    // ...
    if result.actionID == "compress" {  // Never reached
        m.dialog = NewCompressFormatDialog()
        return m, nil
    }
}
```

---

## Solution

### Approach

Restructure the `contextMenuResultMsg` handler in `model.go` to check `actionID` before checking `action != nil`. This allows action-ID-based operations (compress, delete, extract) to work regardless of whether an Action function is provided.

### Code Changes

**File**: `internal/ui/model.go`

**Before** (problematic structure):
```go
if result.action != nil {
    // All actionID checks are inside this block
    if result.actionID == "delete" { ... }
    if result.actionID == "compress" { ... }
    if result.actionID == "extract" { ... }
    // ...
}
```

**After** (fixed structure):
```go
// Handle actionID-based operations first (regardless of action function)
if result.actionID == "delete" { ... return }
if result.actionID == "compress" { ... return }
if result.actionID == "extract" { ... return }
if result.actionID == "copy" || result.actionID == "move" { ... return }

// Handle custom action functions
if result.action != nil {
    if err := result.action(); err != nil { ... }
    // ...
}
```

---

## Functional Requirements

### FR-1: Compress Menu Action
When the user selects "Compress" from the context menu, the system shall display the CompressFormatDialog.

**Verification**:
- Select a file or directory
- Press `@` to open context menu
- Select "Compress"
- Verify CompressFormatDialog appears with format options

### FR-2: Multi-file Compress Action
When multiple files are marked and the user selects "Compress N files", the system shall display the CompressFormatDialog.

**Verification**:
- Mark multiple files with Space key
- Press `@` to open context menu
- Select "Compress N files"
- Verify CompressFormatDialog appears

### FR-3: Other Menu Actions Preserved
All other context menu actions (Delete, Copy, Move, Extract, etc.) shall continue to work correctly.

**Verification**:
- Test Delete action with confirmation dialog
- Test Copy action with file conflict check
- Test Move action with file conflict check
- Test Extract action for archive files

---

## Non-Functional Requirements

### NFR-1: Minimal Code Changes
The fix shall be implemented with minimal changes to existing code structure.

### NFR-2: Test Compatibility
All existing tests shall pass after the fix.

### NFR-3: No Performance Impact
The fix shall not introduce any performance regression.

---

## Test Scenarios

### Unit Tests

1. **Test Compress Action Triggers Dialog**
   - Send contextMenuResultMsg with actionID="compress" and action=nil
   - Verify m.dialog is CompressFormatDialog

2. **Test Delete Action Still Works**
   - Send contextMenuResultMsg with actionID="delete" and action=nil
   - Verify confirmation dialog appears

3. **Test Extract Action Still Works**
   - Send contextMenuResultMsg with actionID="extract" and action function
   - Verify extraction process starts

4. **Test Copy/Move Actions Still Work**
   - Send contextMenuResultMsg with actionID="copy"/"move"
   - Verify file conflict check is triggered

### E2E Tests

1. **E2E: Compress Single Directory**
   - Start duofm
   - Navigate to a directory
   - Press `@`, select "Compress"
   - Verify format selection dialog appears
   - Select format, verify compression flow continues

2. **E2E: Compress Multiple Files**
   - Start duofm
   - Mark 3 files with Space
   - Press `@`, verify "Compress 3 files" option
   - Select it, verify format selection dialog appears

---

## Affected Files

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/ui/model.go` | Modify | Restructure contextMenuResultMsg handler |
| `internal/ui/model_test.go` | Modify | Add/update test cases for compress action |

---

## Acceptance Criteria

- [ ] Selecting "Compress" from context menu opens CompressFormatDialog
- [ ] Selecting "Compress N files" opens CompressFormatDialog
- [ ] Delete action works correctly (shows confirmation dialog)
- [ ] Copy action works correctly (checks file conflicts)
- [ ] Move action works correctly (checks file conflicts)
- [ ] Extract action works correctly (extracts archives)
- [ ] All existing tests pass
- [ ] New test case added for compress action

---

## Implementation Notes

The fix requires careful restructuring to ensure:

1. **Delete handling**: Currently inside `result.action != nil` block but action is nil. Move outside.
2. **Compress handling**: Move outside `result.action != nil` block.
3. **Extract handling**: Has action function but is ID-based. Can stay or move.
4. **Copy/Move handling**: Has action function and is ID-based. Can stay or move.
5. **Generic action handling**: For non-ID-based actions, keep inside `result.action != nil` block.

The safest approach is to check all known actionIDs first, then fall through to generic action handling.

---

## References

- Original Archive SPEC: `doc/tasks/archive/SPEC.md` (FR10: Context Menu Integration)
- Implementation: `doc/tasks/archive/IMPLEMENTATION.md`
- Verification Report: `doc/tasks/archive/VERIFICATION_REPORT_FINAL.md`

---

**Last Updated**: 2026-01-02
**Status**: Ready for Implementation
