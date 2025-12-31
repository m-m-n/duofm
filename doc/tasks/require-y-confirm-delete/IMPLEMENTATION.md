# Implementation Plan: Require Y Key for Delete Confirmation

## Overview

This bugfix modifies the delete confirmation dialog to accept only the `y` key for confirmation, removing the `Enter` key as a confirmation option to prevent accidental file deletion.

## Objectives

- Remove Enter key as a confirmation option in ConfirmDialog
- Maintain all other key behaviors (y confirms, n/Esc/Ctrl+C cancel)
- Update dialog display to show only `[y] Yes  [n] No`
- Ensure all tests pass and add new test coverage

## Prerequisites

### Development Environment
- Go 1.21 or later
- Make for build automation
- tmux (for E2E tests)

### Dependencies
- github.com/charmbracelet/bubbletea - TUI framework (already in use)
- github.com/charmbracelet/lipgloss - Styling (already in use)

### Knowledge Requirements
- Understanding of Bubble Tea's Update-View architecture
- Familiarity with Go table-driven testing
- Basic understanding of tmux for E2E test execution

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea (TUI framework)
- **Styling**: Lip Gloss
- **Testing**: Go's built-in `testing` package + tmux-based E2E tests

### Design Approach

The change is minimal and surgical:
1. Modify the key handling logic in ConfirmDialog.Update() to ignore Enter key
2. Dialog display already shows `[y] Yes  [n] No` (no change needed)
3. Add comprehensive test coverage for the new behavior

### Component Interaction

**ConfirmDialog** receives key messages from Bubble Tea runtime:
- Processes key messages in Update() method
- Returns DialogResult via tea.Cmd when user confirms or cancels
- Main UI model receives dialogResultMsg and acts accordingly

## Implementation Phases

### Phase 1: Modify ConfirmDialog Key Handling

**Goal**: Remove Enter key from confirmation options in ConfirmDialog

**Files to Modify**:
- `internal/ui/confirm_dialog.go`:
  - Modify Update() method to remove "enter" from confirmation case
  - Keep "y" as the only confirmation key

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| ConfirmDialog.Update() | Process key messages and return appropriate DialogResult | Dialog is active, valid tea.KeyMsg received | Returns dialog state and optional command based on key |
| dialogResultMsg | Carry confirmation/cancellation result to parent | Dialog closed with user decision | Parent model receives result for action execution |

**Processing Flow**:

```
User Input → Bubble Tea Runtime
    ↓
ConfirmDialog.Update(tea.KeyMsg)
    ├─ Key == "y" → Return DialogResult{Confirmed: true}
    ├─ Key == "n" OR "esc" OR "ctrl+c" → Return DialogResult{Cancelled: true}
    └─ Key == "enter" OR other → Ignore, return nil cmd (dialog stays active)
```

**Implementation Steps**:

1. **Modify key handling in Update() method**
   - Remove "enter" from the case statement that confirms deletion
   - Keep only "y" in the confirmation case
   - Enter key will fall through to default behavior (no action)
   - Key considerations:
     - Preserve exact behavior for all other keys
     - Ensure dialog remains active when Enter is pressed
     - Maintain thread safety (none required, Bubble Tea is single-threaded)

2. **Verify dialog display remains unchanged**
   - Confirm View() method already shows `[y] Yes  [n] No`
   - No code changes needed for display

**Dependencies**:
- Requires: None (standalone change)
- Blocks: Phase 2 (tests depend on implementation)

**Testing Approach**:

*Unit Tests* (implemented in Phase 2):
- Test that pressing "y" confirms deletion
- Test that pressing "enter" does NOT confirm deletion (dialog stays active)
- Test that pressing "n", "esc", "ctrl+c" cancel deletion
- Test that other keys are ignored

*Integration Tests* (E2E, implemented in Phase 2):
- Test delete operation via `d` key with Enter press (file should NOT be deleted)
- Test delete operation via `d` key with `y` press (file should be deleted)

**Acceptance Criteria**:
- [ ] Enter key no longer triggers confirmation in ConfirmDialog
- [ ] y key still confirms deletion correctly
- [ ] n, Esc, Ctrl+C keys still cancel correctly
- [ ] Dialog remains active when Enter is pressed
- [ ] Dialog display shows `[y] Yes  [n] No` (already correct)

**Estimated Effort**: 小 (< 1 hour for code change)

**Risks and Mitigation**:
- **Risk**: Breaking other key handling logic
  - **Mitigation**: Comprehensive unit tests, careful case statement modification
- **Risk**: Affecting other dialogs unintentionally
  - **Mitigation**: Change is isolated to ConfirmDialog only; other dialogs are separate types

---

### Phase 2: Add and Update Test Coverage

**Goal**: Ensure comprehensive test coverage for the new Enter key behavior

**Files to Modify**:
- `internal/ui/dialog_test.go`:
  - Add new unit test for Enter key behavior
  - Verify existing tests still pass

**Files to Create**:
- `test/e2e/scripts/tests/delete_confirmation_tests.sh`:
  - Add E2E test for delete with Enter key (file should remain)
  - Add E2E test for delete with y key (file should be deleted)

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| TestConfirmDialogEnterIgnored | Verify Enter key does not close dialog or trigger confirmation | ConfirmDialog created and active | Test passes if dialog remains active and no cmd returned |
| E2E delete test | Verify delete behavior with Enter vs y key in real application | duofm running in tmux, test file exists | Test passes if file behavior matches expectation |

**Processing Flow**:

```
Unit Test Flow:
1. Create ConfirmDialog
2. Send Enter key message
3. Verify dialog.IsActive() == true
4. Verify returned cmd == nil
5. Send y key message
6. Verify dialog.IsActive() == false
7. Verify returned cmd contains Confirmed=true

E2E Test Flow:
1. Start duofm in tmux session
2. Navigate to test file
3. Press 'd' to trigger delete dialog
4. Press 'Enter' key
5. Verify file still exists
6. Press 'd' again
7. Press 'y' key
8. Verify file is deleted
```

**Implementation Steps**:

1. **Add unit test for Enter key behavior**
   - Create test function `TestConfirmDialogEnterIgnored`
   - Follow existing test pattern from dialog_test.go
   - Verify dialog stays active and returns nil cmd
   - Key considerations:
     - Use tea.KeyMsg{Type: tea.KeyEnter} for Enter key
     - Check both dialog.IsActive() and returned cmd

2. **Add E2E test for delete confirmation behavior**
   - Create new test file or add to file_operation_tests.sh
   - Test scenario: delete file, press Enter, verify file exists
   - Test scenario: delete file, press y, verify file deleted
   - Key considerations:
     - Use proper sleep timing for UI updates
     - Clean up test files after execution
     - Handle both success and failure cases

3. **Run all existing tests**
   - Ensure no regression in existing functionality
   - Fix any broken tests if found

**Dependencies**:
- Requires: Phase 1 (code changes must be complete)
- Blocks: None (this is the final phase)

**Testing Approach**:

*Unit Tests*:
- Run with `go test ./internal/ui/...`
- Verify all tests pass
- Check test coverage with `go test -cover ./internal/ui/...`

*E2E Tests*:
- Run with test framework in test/e2e/
- Verify delete confirmation behavior in real application
- Test both positive (y key) and negative (Enter key) cases

**Acceptance Criteria**:
- [ ] New unit test `TestConfirmDialogEnterIgnored` passes
- [ ] All existing unit tests pass
- [ ] E2E test for Enter key (file NOT deleted) passes
- [ ] E2E test for y key (file deleted) passes
- [ ] Test coverage for ConfirmDialog remains high (>80%)

**Estimated Effort**: 小 (1-2 hours for test implementation)

**Risks and Mitigation**:
- **Risk**: E2E tests may be flaky due to timing
  - **Mitigation**: Use appropriate sleep values, verify screen capture reliability
- **Risk**: Test coverage gaps
  - **Mitigation**: Review spec test scenarios, ensure all cases covered

---

## Complete File Structure

```
duofm/
├── internal/
│   └── ui/
│       ├── confirm_dialog.go      # MODIFIED: Remove "enter" from confirmation case
│       └── dialog_test.go          # MODIFIED: Add TestConfirmDialogEnterIgnored
├── test/
│   └── e2e/
│       └── scripts/
│           └── tests/
│               └── file_operation_tests.sh  # MODIFIED: Add delete confirmation tests
└── doc/
    └── tasks/
        └── require-y-confirm-delete/
            ├── SPEC.md
            └── IMPLEMENTATION.md   # This file
```

**File Descriptions**:
- `confirm_dialog.go`: Contains ConfirmDialog type and Update() method; single line change to remove "enter" from case statement
- `dialog_test.go`: Contains unit tests for all dialog types; add new test function for Enter key behavior
- `file_operation_tests.sh`: E2E tests for file operations; add delete confirmation scenarios

## Testing Strategy

### Unit Testing

**Approach**:
- Use Go's built-in `testing` package
- Table-driven tests for multiple key scenarios (if needed)
- Follow existing test patterns in dialog_test.go
- Mock nothing (ConfirmDialog is self-contained)

**Test Coverage Goals**:
- ConfirmDialog Update() method: 100% coverage (all key cases tested)
- Overall ui package: maintain existing coverage (>80%)

**Key Test Areas**:

1. **Enter Key Behavior** (NEW)
   - Pressing Enter keeps dialog active
   - Pressing Enter returns nil cmd
   - No DialogResult generated

2. **Confirmation Behavior** (EXISTING, verify no regression)
   - Pressing y closes dialog
   - Pressing y returns DialogResult{Confirmed: true}

3. **Cancellation Behavior** (EXISTING, verify no regression)
   - Pressing n/Esc/Ctrl+C closes dialog
   - Returns DialogResult{Cancelled: true}

### Integration Testing

**E2E Test Scenarios**:

| Scenario | Steps | Expected Result |
|----------|-------|-----------------|
| Delete with Enter | 1. Navigate to file<br>2. Press 'd'<br>3. Press 'Enter' | File remains, dialog still active |
| Delete with y | 1. Navigate to file<br>2. Press 'd'<br>3. Press 'y' | File deleted, dialog closed |
| Delete with n | 1. Navigate to file<br>2. Press 'd'<br>3. Press 'n' | File remains, dialog closed (EXISTING) |
| Delete with Esc | 1. Navigate to file<br>2. Press 'd'<br>3. Press 'Esc' | File remains, dialog closed (EXISTING) |

**Approach**:
- Use tmux for terminal simulation
- Create temporary test files in /testdata/user_owned/
- Capture screen output to verify behavior
- Clean up test files after execution

### Manual Testing Checklist

Based on spec test scenarios:
- [ ] Delete file with 'd' key, press Enter → file not deleted, dialog stays open
- [ ] Press 'y' after Enter → file deleted successfully
- [ ] Delete file with 'd' key, press 'y' directly → file deleted
- [ ] Delete file with 'd' key, press 'n' → file not deleted, dialog closed
- [ ] Delete file with 'd' key, press 'Esc' → file not deleted, dialog closed
- [ ] Delete file with 'd' key, press 'Ctrl+C' → file not deleted, dialog closed
- [ ] Context menu delete, press Enter → file not deleted
- [ ] Context menu delete, press 'y' → file deleted

## Dependencies

### External Dependencies

None (all required packages already in use)

### Internal Dependencies

**Implementation Order** (respecting dependencies):
1. Phase 1: Modify ConfirmDialog (no dependencies)
2. Phase 2: Add tests (depends on Phase 1 implementation)

**Component Dependencies**:
- ConfirmDialog depends on: Bubble Tea framework (tea.Msg, tea.Cmd)
- dialog_test.go depends on: ConfirmDialog implementation
- E2E tests depend on: Full application build with modified ConfirmDialog

## Risk Assessment

### Technical Risks

1. **Unintended Behavior Change**
   - **Risk**: Breaking other key handling in ConfirmDialog
   - **Likelihood**: Low (change is minimal and isolated)
   - **Impact**: Medium (could affect delete safety)
   - **Mitigation**:
     - Comprehensive unit tests for all key cases
     - E2E tests verify real-world behavior
     - Code review before merge

2. **Test Flakiness**
   - **Risk**: E2E tests may fail intermittently due to timing
   - **Likelihood**: Low (existing E2E tests are stable)
   - **Impact**: Low (can re-run tests)
   - **Mitigation**:
     - Use proven sleep timings from existing tests
     - Verify screen capture reliability
     - Add retry logic if needed

### Implementation Risks

1. **Scope Creep**
   - **Risk**: Modifying other dialogs or adding unrelated features
   - **Mitigation**: Strict adherence to spec, change only ConfirmDialog

2. **Incomplete Testing**
   - **Risk**: Missing edge cases in test coverage
   - **Mitigation**: Review spec test scenarios, ensure all cases covered

## Performance Considerations

No performance impact expected:
- Change is a simple case statement modification
- No additional allocations or processing
- Dialog behavior remains O(1) for key handling

## Security Considerations

**Positive Security Impact**:
- Reduces risk of accidental file deletion
- Requires explicit confirmation with 'y' key
- Prevents destructive operations from habitual Enter pressing

**No New Security Risks**:
- Change does not introduce any vulnerabilities
- File operations still respect filesystem permissions
- Error handling remains unchanged

## Open Questions

### From Specification:

None - specification is clear and complete

### Implementation-Specific:

- [ ] Should we apply the same change to other confirmation dialogs (e.g., OverwriteDialog)?
  - **Answer**: No, per spec constraints - "Other dialogs are out of scope for this bugfix"
  - Overwrite is less destructive (can be undone by re-copying)
  - Can be addressed in future if needed

### To Clarify with User:

None - requirements are clear

## Future Enhancements

Items deferred to later phases or releases:

### Potential Future Changes:

1. **Consistent Confirmation Pattern**
   - Apply same "y-only" confirmation to OverwriteDialog
   - Standardize destructive operation confirmations across all dialogs
   - Would require separate specification and implementation plan

2. **Configurable Confirmation Keys**
   - Allow users to configure confirmation key preferences
   - Would require configuration system enhancement
   - Lower priority (current behavior is safe default)

### Not in Current Spec:

- Visual indication that Enter is ignored (e.g., brief message or animation)
- Customizable confirmation messages
- Undo functionality for delete operations

## Success Metrics

### Functional Completeness
- [ ] Enter key no longer confirms deletion in ConfirmDialog
- [ ] y key confirms deletion correctly
- [ ] n, Esc, Ctrl+C cancel deletion correctly
- [ ] Dialog display shows `[y] Yes  [n] No`

### Quality Metrics
- [ ] All existing tests pass
- [ ] New test for Enter key behavior passes
- [ ] E2E tests verify real-world behavior
- [ ] Code follows Go best practices (gofmt, go vet pass)

### Performance Metrics
- [ ] No performance regression (not applicable for this change)

### User Experience
- [ ] Prevents accidental deletion with Enter key
- [ ] Clear visual indication of available options
- [ ] No confusion about why Enter doesn't work (button hints show only y/n)

## References

- **Specification**: `doc/tasks/require-y-confirm-delete/SPEC.md`
- **Modified File**: `internal/ui/confirm_dialog.go`
- **Test File**: `internal/ui/dialog_test.go`
- **E2E Tests**: `test/e2e/scripts/tests/file_operation_tests.sh`
- **Bubble Tea Documentation**: https://github.com/charmbracelet/bubbletea
- **Go Testing**: https://go.dev/doc/tutorial/add-a-test

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - Review the single-line code change approach
   - Confirm test coverage is adequate
   - Approve proceeding with implementation

2. **Begin Implementation**
   - Start with Phase 1: Modify confirm_dialog.go
   - Test locally with manual testing
   - Proceed to Phase 2: Add test coverage

3. **Verification**
   - Run `go test ./internal/ui/...` to verify unit tests
   - Run E2E tests to verify real-world behavior
   - Manual testing with checklist above

4. **Commit and Review**
   - Create commit with descriptive message
   - Submit for code review
   - Address any feedback
   - Merge to main branch
