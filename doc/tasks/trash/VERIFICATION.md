# Verification Document: Trash Feature

## Overview
**Feature**: Trash (Recycle Bin)
**SPEC.md**: `doc/tasks/trash/SPEC.md`
**IMPLEMENTATION.md**: `doc/tasks/trash/IMPLEMENTATION.md`

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
go test ./internal/ui/... ./internal/fs/... -v -cover
```

### Coverage Target
- **Minimum**: 80%
- **Target**: 90%

### Test Scenarios from SPEC.md

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-01 | Delete key shows confirmation dialog | Dialog displayed with file name | Unit/Integration |
| TS-02 | Single file shows "Move 'filename.txt' to trash?" | Correct message format | Unit |
| TS-03 | Multiple files shows "Move N items to trash?" | Correct message format | Unit |
| TS-04 | Disk space warning is displayed | Warning text visible | Unit |
| TS-05 | Y key confirms and moves to trash | File moved to trash | Integration |
| TS-06 | N key cancels operation | File remains in place | Integration |
| TS-07 | Esc key cancels operation | File remains in place | Integration |
| TS-08 | T key opens trash dialog | Dialog displayed at center | Integration |
| TS-09 | Columns displayed correctly | Name, Size, Deleted, Original Path | Unit |
| TS-10 | j/k navigation works | Cursor moves up/down | Unit |
| TS-11 | Space marks item | Mark toggle works | Unit |
| TS-12 | R key in dialog restores file | File restored | Integration |
| TS-13 | R key outside dialog renames | Rename dialog opens | Integration |
| TS-14 | Shift+E empties trash | All files deleted | Integration |
| TS-15 | Esc closes trash dialog | Dialog closes | Unit |

## Code Quality Verification

### Format Check
```bash
gofmt -l ./internal/ui/move_to_trash_dialog.go ./internal/ui/trash_dialog.go ./internal/ui/model_update_trash.go
```

### Static Analysis
```bash
go vet ./internal/ui/... ./internal/fs/...
```

## File Structure Verification

### Files to Create
- `internal/ui/move_to_trash_dialog.go` - Move to Trash confirmation dialog
- `internal/ui/move_to_trash_dialog_test.go` - Confirmation dialog tests
- `internal/ui/trash_dialog.go` - Trash dialog implementation (if not exists)
- `internal/ui/trash_dialog_test.go` - Trash dialog tests (if not exists)

### Files to Modify
- `internal/ui/model_update_trash.go` - Add confirmation flow, modify handleTrash()
- `internal/ui/model_update.go` - Add trashConfirmResultMsg handler

## SPEC.md Compliance

### Success Criteria

| ID | Criterion from SPEC.md | How to Verify |
|----|------------------------|---------------|
| SC-01 | Delete key shows confirmation dialog before moving to trash (FR1.1) | Manual test: press Delete, check dialog appears |
| SC-02 | Confirmation dialog displays file name and disk space warning (FR1.2) | Unit test: verify View() output |
| SC-03 | File moves to ~/.local/share/Trash/files/ after confirmation (FR1.3) | Integration test: verify file location |
| SC-04 | .trashinfo file is created (FR1.4) | Integration test: verify file creation |
| SC-05 | Name collision handling (FR1.5) | Unit test: fs/trash_test.go |
| SC-06 | T key opens trash dialog (FR1.6) | Manual test: press T, check dialog |
| SC-07 | Trash dialog displays columns (FR1.7) | Unit test: verify View() output |
| SC-08 | j/k navigation in dialog (FR1.10) | Unit test: cursor movement |

### Functional Requirements Coverage

| Requirement | Implementation Phase | Verification |
|-------------|---------------------|--------------|
| FR1.1 | Phase 0 | Unit test + Manual |
| FR1.2 | Phase 0 | Unit test + Manual |
| FR1.3 | Existing | Integration test |
| FR1.4 | Existing | Integration test |
| FR1.5 | Existing | Unit test |
| FR1.6 | Phase 1 | Manual test |
| FR1.7 | Phase 1 | Unit test |
| FR1.8 | Existing | Integration test |
| FR1.9 | Existing | Integration test |
| FR1.10 | Phase 1 | Unit test |
| FR2.1 | Phase 3 | Integration test |
| FR2.2 | Phase 3 | Manual test |
| FR2.3 | Phase 3 | Integration test |
| FR2.4 | Phase 2 | Unit test |

## Manual Testing Checklist

### Phase 0: Move to Trash Confirmation

#### Basic Functionality
- [ ] Delete key on single file shows confirmation dialog
- [ ] Dialog title is "Move to Trash"
- [ ] Dialog shows "Move 'filename.txt' to trash?" for single file
- [ ] Dialog shows "Move N items to trash?" for multiple files
- [ ] Warning text is displayed: "File will not be permanently deleted..."
- [ ] Warning text mentions disk space
- [ ] Y key confirms and file is moved to trash
- [ ] y key (lowercase) also works
- [ ] N key cancels and file remains
- [ ] n key (lowercase) also works
- [ ] Esc key cancels and file remains
- [ ] After confirmation, file list is refreshed
- [ ] Status message shows "Moved to trash: filename"

#### Multiple Files
- [ ] Mark multiple files with Space
- [ ] Press Delete key
- [ ] Dialog shows correct count (e.g., "Move 3 items to trash?")
- [ ] Y confirms all files are moved
- [ ] N cancels all files remain
- [ ] Marks are cleared after operation

#### Edge Cases
- [ ] Delete on parent directory (..) does nothing
- [ ] Delete on empty selection does nothing
- [ ] Permission denied shows error message (not silent failure)
- [ ] Trash unavailable shows error message

### Phase 1: TrashDialog Display

#### Dialog Display
- [ ] T key opens TrashDialog at screen center
- [ ] Both panes are dimmed (DialogDisplayScreen)
- [ ] Dialog title shows "Trash" with item count [N]
- [ ] Columns are visible: Name, Size, Deleted, Original Path
- [ ] Footer shows key hints

#### Navigation
- [ ] j moves cursor down
- [ ] k moves cursor up
- [ ] Down arrow moves cursor down
- [ ] Up arrow moves cursor up
- [ ] Cursor stops at boundaries (no wrap)
- [ ] Scroll works when items exceed visible height
- [ ] Current item is highlighted

#### Close
- [ ] Esc closes dialog
- [ ] q closes dialog

### Phase 2: Mark Operations

- [ ] Space toggles mark on current item
- [ ] Marked items show * prefix
- [ ] Space moves cursor down after marking
- [ ] Multiple items can be marked
- [ ] GetMarkedItems returns correct items

### Phase 3: Restore and Empty

#### Restore
- [ ] R key on single item restores it
- [ ] Restored file appears at original location
- [ ] .trashinfo file is removed
- [ ] TrashDialog updates (item removed from list)

#### Restore with Conflict
- [ ] R key when destination exists shows RestoreConflictDialog
- [ ] Overwrite option works
- [ ] Rename option works
- [ ] Skip option works

#### Batch Restore
- [ ] Mark multiple items
- [ ] R key restores all marked items
- [ ] Items with conflicts are skipped
- [ ] Status message shows result

#### Empty Trash
- [ ] Shift+E shows EmptyTrashDialog
- [ ] Confirmation dialog shows item count
- [ ] Y empties all trash items
- [ ] N cancels
- [ ] After empty, dialog closes
- [ ] Status message shows "Trash emptied"

### Key Binding Resolution

- [ ] R key in TrashDialog triggers restore
- [ ] R key outside TrashDialog triggers rename
- [ ] No key binding conflict

### Error Handling

- [ ] Invalid .trashinfo shows error message
- [ ] Permission denied on restore shows error
- [ ] Permission denied on empty shows error
- [ ] Network errors are handled gracefully

## Performance Verification (if applicable)

### Benchmarks
- Same-filesystem trash operation < 100ms (NFR1.1)
- Trash dialog opening with 1000 files < 100ms (NFR1.2)

### Performance Tests
```bash
# Create test files
for i in $(seq 1 1000); do touch /tmp/testfile$i; done

# Time trash dialog opening
time ./duofm  # Press T and measure
```

## Security Verification (if applicable)

### Security Checks
- [ ] Path traversal prevented in trash names
- [ ] URL encoding/decoding works for special characters
- [ ] Symlinks are moved (not followed)
- [ ] Permission checks prevent unauthorized access

## Implementation Status

### Phase 0: Move to Trash Confirmation
- [ ] MoveToTrashConfirmDialog created
- [ ] Unit tests written
- [ ] handleTrash() modified
- [ ] trashConfirmResultMsg handler added
- [ ] Integration tests pass
- [ ] Manual tests pass

### Phase 1: TrashDialog基本実装
- [x] TrashDialog created
- [x] Unit tests written
- [x] handleOpenTrashDialog() implemented
- [x] Navigation working
- [x] Manual tests pass

### Phase 2: マーク機能
- [x] Space mark toggle implemented
- [x] Visual indication working
- [x] Unit tests pass

### Phase 3: 復元と空にする
- [x] R key restore implemented
- [x] Shift+E empty trash implemented
- [x] Conflict handling working
- [x] Integration tests pass

## Verification Summary

| Category | Items | Automated | Manual |
|----------|-------|-----------|--------|
| Build | 1 | Yes | - |
| Tests | 15+ | Yes | - |
| Code Quality | 2 | Yes | - |
| File Structure | 4+ | Yes | - |
| SPEC Compliance | 8 | Partial | Yes |
| Phase 0 Manual | 16 | - | Yes |
| Phase 1 Manual | 15 | - | Yes |
| Phase 2 Manual | 5 | - | Yes |
| Phase 3 Manual | 14 | - | Yes |
| Performance | 2 | Manual | - |
| Security | 4 | Partial | Yes |

**Total**: 15+ automated test scenarios, 50+ manual test items

## Unit Test Checklist

### MoveToTrashConfirmDialog Tests
- [ ] TestNewMoveToTrashDialog_SingleFile
- [ ] TestNewMoveToTrashDialog_MultipleFiles
- [ ] TestMoveToTrashDialog_View_SingleFile
- [ ] TestMoveToTrashDialog_View_MultipleFiles
- [ ] TestMoveToTrashDialog_View_ContainsWarning
- [ ] TestMoveToTrashDialog_Update_YKeyConfirms
- [ ] TestMoveToTrashDialog_Update_NKeyCancels
- [ ] TestMoveToTrashDialog_Update_EscCancels

### TrashDialog Tests (existing)
- [x] TestNewTrashDialog
- [x] TestLoadTrashItems
- [x] TestTrashDialog_CursorNavigation
- [x] TestTrashDialog_Close
- [x] TestTrashDialog_View
- [x] TestTrashDialog_Scroll
- [x] TestTrashDialog_Mark
- [x] TestTrashDialog_Restore
- [x] TestTrashDialog_EmptyTrash
- [x] TestTrashItem_Size

## Conclusion

This verification document defines all test scenarios and acceptance criteria for the Trash feature implementation. Phase 0 (Move to Trash Confirmation) is the new addition to support FR1.1 and FR1.2 from SPEC.md.

**Priority Order**:
1. Phase 0 - Move to Trash Confirmation (NEW)
2. Phase 1-3 - Already implemented (verify still working)

**Next Steps**:
1. Implement Phase 0 (MoveToTrashConfirmDialog)
2. Run all verification items
3. `/sdd.6-verify` for automated verification
4. `/sdd.7-review` for code review
