# Implementation Plan: Fix Cursor Position After File Deletion

## Overview

Fix the cursor positioning behavior after file deletion to maintain user context. Currently, the cursor always jumps to the top of the list (position 0) after deletion, causing poor UX especially in large directories. The fix will move the cursor to the next file after deletion, or to the previous file when deleting the last file.

## Objectives

- Maintain user work context after file deletion by keeping cursor near deletion point
- Implement cursor positioning that matches standard file manager behavior
- Ensure smooth workflow for multiple consecutive delete operations
- Handle all edge cases gracefully without errors or crashes

## Prerequisites

### Development Environment
- Go 1.21 or later
- Make for build automation
- Terminal for testing TUI behavior

### Dependencies
- `github.com/charmbracelet/bubbletea` (v0.25.0+) - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling (already in use)
- Go standard library: `path/filepath`, `os`

### Knowledge Requirements
- Understanding of Bubble Tea's Model-Update-View architecture
- Familiarity with duofm's Pane structure and cursor management
- Go slice operations and bounds checking

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea (TUI framework based on The Elm Architecture)
- **Key Libraries**:
  - `internal/ui` - UI components and state management
  - `internal/fs` - File system operations (unchanged)

### Design Approach

**Principle**: Preserve cursor context by remembering the deletion point and calculating the optimal new cursor position after directory reload.

**Key Insight**: When file(s) are deleted from a list:
- **Single file deletion**: Use the cursor position as reference
  - If cursor index < new list size: Cursor stays at same index (next file takes its place)
  - If cursor index >= new list size: Cursor moves to last file (new list size - 1)
- **Multiple file deletion**: Use the smallest index among marked files as reference
  - The cursor will be positioned at the location where the first marked file was
  - This provides predictable behavior when deleting multiple files
- If all files deleted: Only ".." remains → cursor goes to 0

### Component Interaction

```
Model (model_update.go)
  ↓ User confirms deletion
  ↓ Remember cursor position
executeDeleteOperation()
  ↓ Delete file(s)
  ↓ Call LoadDirectory()
  ↓ Calculate new cursor position
  ↓ Set cursor and adjust scroll
Pane (pane.go)
  ↓ Apply new cursor
  ↓ Ensure visibility
adjustScroll()
```

## Implementation Phases

### Phase 1: Add Cursor Position Calculation Logic

**Goal**: Implement the core logic to calculate optimal cursor position after deletion

**Files to Create**: None

**Files to Modify**:
- `internal/ui/pane.go`:
  - Add method `calculateCursorAfterDeletion(deletedIndex int, deletedCount int) int`

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| calculateCursorAfterDeletion | Calculate optimal cursor position after file deletion | Directory reloaded, entries updated | Returns valid cursor index [0, len(entries)-1] |

**Processing Flow**:
```
1. Receive deleted index and deletion count
   - For single file: deleted index = cursor position before deletion
   - For multiple files: deleted index = smallest index of marked files
2. Check if deleted index is still valid in new entries list
   ├─ Index < len(entries) → Return same index (next file took its place)
   └─ Index >= len(entries) → Proceed to step 3
3. Check if any entries remain
   ├─ len(entries) > 0 → Return len(entries)-1 (last file)
   └─ len(entries) == 0 → Return 0 (parent directory only)
```

**Implementation Steps**:

1. **Add calculateCursorAfterDeletion method to Pane**
   - Responsibility: Determine new cursor position based on deletion point and remaining entries
   - Key considerations:
     - Handle single file deletion (deletedCount = 1)
     - Handle multiple file deletion (deletedCount > 1)
     - Ensure returned index is always within valid bounds
     - Return 0 for empty directory (only parent directory remains)

**Dependencies**:
- Requires: None (independent utility method)
- Blocks: Phase 2 (integration with delete operation)

**Testing Approach**:

*Unit Tests*:
- Test cursor calculation for middle file deletion (index < len(entries)-1)
- Test cursor calculation for last file deletion (index == len(entries)-1)
- Test cursor calculation for multiple file deletion
- Test cursor calculation when all files deleted (empty directory)
- Test cursor calculation for edge case (only 2 entries: ".." and one file)

*Manual Testing*:
- N/A for this phase (pure logic, no UI changes yet)

**Acceptance Criteria**:
- [ ] `calculateCursorAfterDeletion` returns correct index for all test cases
- [ ] Method handles edge cases without panics or out-of-bounds errors
- [ ] All unit tests pass with 100% coverage for this method
- [ ] Documentation clearly explains the logic and edge cases

**Estimated Effort**: 小 (1-2 hours)

**Risks and Mitigation**:
- **Risk**: Edge cases not covered (e.g., concurrent modifications)
  - **Mitigation**: Comprehensive test suite covering all deletion scenarios
- **Risk**: Index calculation error leading to out-of-bounds access
  - **Mitigation**: Strict bounds checking in all branches

---

### Phase 2: Integrate with Delete Operation

**Goal**: Modify the delete operation flow to preserve and restore cursor position

**Files to Create**: None

**Files to Modify**:
- `internal/ui/model_update.go`:
  - Modify `executeDeleteOperation()` to remember cursor position before deletion
  - Call `calculateCursorAfterDeletion()` after `LoadDirectory()`
  - Set new cursor position and adjust scroll

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| executeDeleteOperation (modified) | Execute file deletion and restore cursor context | Valid file selection or marked files | Files deleted, directory reloaded, cursor at appropriate position |

**Processing Flow**:
```
1. Get active pane and marked files list
2. Determine deletion type (single vs multiple)
   ├─ Multiple files marked → Branch A
   └─ Single file selected → Branch B

Branch A (Multiple files):
3a. Find smallest index among marked files (minMarkedIndex)
   - Iterate through entries to find first marked file position
4a. Count marked files (deletedCount)
5a. Delete each marked file
   ├─ Error occurs → Show error dialog, skip to step 8a
   └─ Success → Continue
6a. Clear marks
7a. LoadDirectory()
8a. Calculate new cursor: calculateCursorAfterDeletion(minMarkedIndex, deletedCount)
9a. Set cursor and ensure visible

Branch B (Single file):
3b. Verify entry exists and is not parent directory
4b. Remember current cursor position (cursorBeforeDeletion)
5b. Delete single file
   ├─ Error occurs → Show error dialog, return
   └─ Success → Continue
6b. LoadDirectory()
7b. Calculate new cursor: calculateCursorAfterDeletion(cursorBeforeDeletion, 1)
8b. Set cursor and ensure visible
```

**Implementation Steps**:

1. **Modify executeDeleteOperation for multiple file deletion**
   - Find the smallest index among marked files (first marked file position)
   - After LoadDirectory(), calculate new cursor position based on this index
   - Apply new cursor position using SetCursor()
   - Ensure cursor visibility using EnsureCursorVisible()
   - Key considerations:
     - Handle deletion errors gracefully (don't skip cursor restoration)
     - deletedCount should be accurate count of files to be deleted
     - Use minMarkedIndex instead of current cursor position for multiple deletion

2. **Modify executeDeleteOperation for single file deletion**
   - Remember cursor position before deletion
   - After successful deletion and LoadDirectory(), calculate new cursor
   - Apply new cursor position
   - Ensure cursor visibility
   - Key considerations:
     - Only restore cursor if deletion succeeds
     - Handle case where selected entry is parent directory (no deletion occurs)

**Dependencies**:
- Requires: Phase 1 (calculateCursorAfterDeletion method)
- Blocks: Phase 3 (testing)

**Testing Approach**:

*Unit Tests*:
- Mock file system to test delete operation with cursor restoration
- Test single file deletion at various positions
- Test multiple file deletion scenarios
- Test error handling during deletion (cursor should still be set)

*Integration Tests*:
- Create temporary directory with test files
- Execute delete operations and verify cursor position
- Test with various entry counts (2, 5, 10, 100 entries)

*Manual Testing*:
- N/A (E2E tests in Phase 3)

**Acceptance Criteria**:
- [ ] Single file deletion moves cursor to next file (or previous if last)
- [ ] Multiple file deletion positions cursor appropriately
- [ ] Error during deletion still results in valid cursor position
- [ ] Cursor is always visible after deletion (no off-screen cursor)
- [ ] All integration tests pass

**Estimated Effort**: 小 (2-3 hours)

**Risks and Mitigation**:
- **Risk**: Race condition between LoadDirectory() and cursor setting
  - **Mitigation**: LoadDirectory() is synchronous in delete flow, no race possible
- **Risk**: Error handling path doesn't restore cursor
  - **Mitigation**: Test error scenarios explicitly

---

### Phase 3: Comprehensive Testing and Refinement

**Goal**: Ensure all edge cases are covered and behavior matches specification

**Files to Create**:
- `internal/ui/pane_delete_test.go` - Dedicated test file for delete cursor behavior (optional, can add to existing test files)

**Files to Modify**:
- `internal/ui/model_update_test.go`:
  - Add tests for executeDeleteOperation cursor behavior
- `internal/ui/pane_test.go`:
  - Add tests for calculateCursorAfterDeletion

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Test Suite | Verify cursor positioning for all deletion scenarios | Implementation complete | 100% coverage of cursor positioning logic |

**Processing Flow**:
```
1. Setup test environment (temp directories, mock files)
2. For each test scenario:
   ├─ Setup initial state (cursor position, file list)
   ├─ Execute deletion
   ├─ Verify cursor position matches expected
   └─ Verify cursor is visible (within scroll bounds)
3. Cleanup test environment
```

**Implementation Steps**:

1. **Write unit tests for calculateCursorAfterDeletion**
   - Test scenarios from SPEC.md Test Section (Tests 1-4, Edge Cases 1-2)
   - Use table-driven test pattern for multiple scenarios
   - Key considerations:
     - Cover all branches of the calculation logic
     - Test boundary conditions (0, 1, 2 entries)

2. **Write integration tests for executeDeleteOperation**
   - Test full delete flow with real temporary files
   - Verify cursor position after delete and reload
   - Test both single and multiple file deletion
   - Key considerations:
     - Create actual files on filesystem for realistic test
     - Clean up temporary files after test
     - Test with various initial cursor positions

3. **Write E2E tests for user-visible behavior**
   - Test cursor movement is visually correct
   - Test scroll position adjusts appropriately
   - Test consecutive deletion operations
   - Key considerations:
     - May require manual testing or screenshot comparison
     - Focus on user scenarios from SPEC.md (US1-US4)

4. **Performance testing**
   - Verify cursor calculation completes within 1ms (NFR1)
   - Test with large directories (1000+ files)
   - Key considerations:
     - Use benchmarking to measure performance
     - Ensure no regression in directory load time

**Dependencies**:
- Requires: Phase 2 (implementation complete)
- Blocks: None (final phase)

**Testing Approach**:

*Unit Tests*:
- Table-driven tests for calculateCursorAfterDeletion covering all scenarios
- Mock-based tests for executeDeleteOperation flow

*Integration Tests*:
- Temporary filesystem tests for delete operations
- Verify cursor + scroll state after deletion

*E2E Tests*:
- Manual verification of UI behavior
- Automated tests using tmux/expect if available

*Performance Tests*:
- Benchmark calculateCursorAfterDeletion with various entry counts
- Ensure < 1ms execution time

**Acceptance Criteria**:
- [ ] All unit tests pass with 80%+ coverage
- [ ] All integration tests pass
- [ ] Manual testing confirms expected behavior for all user stories (US1-US4)
- [ ] Performance benchmark shows < 1ms calculation time
- [ ] No regressions in existing delete functionality
- [ ] Edge cases handled gracefully (empty directory, 2 entries, etc.)

**Estimated Effort**: 中 (4-6 hours)

**Risks and Mitigation**:
- **Risk**: Test environment setup complexity
  - **Mitigation**: Use Go's t.TempDir() for automatic cleanup
- **Risk**: E2E tests are flaky or difficult to automate
  - **Mitigation**: Prioritize unit and integration tests, use manual testing for E2E

---

## Complete File Structure

```
duofm/
├── cmd/duofm/
│   └── main.go                    # Entry point (unchanged)
├── internal/
│   ├── ui/
│   │   ├── model_update.go        # MODIFY: executeDeleteOperation() with cursor restoration
│   │   ├── model_update_test.go   # MODIFY: Add cursor behavior tests
│   │   ├── pane.go                # MODIFY: Add calculateCursorAfterDeletion()
│   │   ├── pane_test.go           # MODIFY: Add cursor calculation tests
│   │   ├── model.go               # Existing model (unchanged)
│   │   ├── pane_navigation.go     # Existing navigation (unchanged)
│   │   └── ...                    # Other UI files (unchanged)
│   ├── fs/
│   │   └── operations.go          # File operations (unchanged)
│   └── config/
│       └── config.go              # Configuration (unchanged)
├── doc/
│   └── tasks/
│       └── fix-cursor-position-after-delete/
│           ├── SPEC.md            # Specification (reference)
│           ├── 要件定義書.md       # Requirements (reference)
│           ├── IMPLEMENTATION.md  # This file
│           └── VERIFICATION.md    # Verification document
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

**File Descriptions**:
- `pane.go`: Contains Pane struct and cursor management. Added calculateCursorAfterDeletion() as a pure calculation method.
- `model_update.go`: Contains update message handlers including executeDeleteOperation(). Modified to preserve and restore cursor position.
- `pane_test.go`: Unit tests for Pane methods. Added comprehensive tests for cursor calculation.
- `model_update_test.go`: Tests for Model update logic. Added tests for delete operation with cursor preservation.

## Testing Strategy

### Unit Testing

**Approach**:
- Use Go's built-in `testing` package
- Table-driven tests for multiple scenarios
- No file system dependencies (use in-memory data)

**Test Coverage Goals**:
- calculateCursorAfterDeletion: 100% coverage (critical logic)
- executeDeleteOperation: 80%+ coverage (integration with existing flow)

**Key Test Areas**:

1. **calculateCursorAfterDeletion (pane_test.go)**
   - Middle file deletion: deletedIndex=2, entries=5 → cursor=2
   - Last file deletion: deletedIndex=5, entries=5 → cursor=4
   - Multiple file deletion: various combinations
   - Empty directory: deletedIndex=any, entries=1 → cursor=0
   - Edge case: 2 entries (.. + 1 file), delete file → cursor=0

2. **executeDeleteOperation (model_update_test.go)**
   - Single file deletion with cursor preservation
   - Multiple file deletion with cursor preservation
   - Error handling (deletion fails, cursor still set)
   - Parent directory selected (no deletion, cursor unchanged)

### Integration Testing

**Scenarios**:
1. Create temporary directory with 10 test files
2. Delete file at various positions (beginning, middle, end)
3. Verify cursor position matches expected after reload
4. Delete multiple marked files, verify cursor position
5. Delete all files, verify cursor on parent directory

**Approach**:
- Use `t.TempDir()` for automatic cleanup
- Create real files for realistic testing
- Verify both cursor position and scroll offset

### Manual Testing Checklist

Based on spec test scenarios and user stories:

**US1: Delete File in Middle of List**
- [ ] Navigate to directory with 10+ files
- [ ] Move cursor to middle file (e.g., 5th file)
- [ ] Press 'd', confirm with 'y'
- [ ] Verify cursor is now on what was the 6th file (now 5th position)
- [ ] Verify scroll position keeps cursor visible

**US2: Delete Last File in List**
- [ ] Navigate to directory with 5+ files
- [ ] Move cursor to last file (press 'G')
- [ ] Press 'd', confirm with 'y'
- [ ] Verify cursor is now on new last file (was second-to-last)
- [ ] Verify cursor is visible

**US3: Delete Multiple Marked Files**
- [ ] Navigate to directory with 10+ files
- [ ] Mark 3 consecutive files in the middle (Space key)
- [ ] Press 'd', confirm with 'y'
- [ ] Verify cursor is near previous position (on next available file)
- [ ] Verify cursor is visible

**US4: Delete All Files in Directory**
- [ ] Navigate to directory with 2-3 files
- [ ] Mark all files (excluding ..)
- [ ] Press 'd', confirm with 'y'
- [ ] Verify cursor is on ".." (parent directory)
- [ ] Verify application remains stable

**Edge Cases**:
- [ ] Delete first real file (after "..") → cursor stays at same index
- [ ] Delete file when only 2 entries exist ("..." + 1 file) → cursor on ".."
- [ ] Delete operation fails (permission denied) → cursor still positioned appropriately
- [ ] Large directory (100+ files) → delete performs smoothly, cursor positioned correctly

## Dependencies

### External Dependencies

| Package | Version | Purpose | Installation |
|---------|---------|---------|--------------|
| github.com/charmbracelet/bubbletea | v0.25.0+ | TUI framework | Already installed |
| github.com/charmbracelet/lipgloss | v0.9.1+ | Styling | Already installed |

### Internal Dependencies

**Implementation Order** (respecting dependencies):
1. Phase 1: calculateCursorAfterDeletion (no dependencies)
2. Phase 2: Modify executeDeleteOperation (depends on Phase 1)
3. Phase 3: Testing (depends on Phase 2)

**Component Dependencies**:
- `calculateCursorAfterDeletion` (pane.go) - Independent, no dependencies
- `executeDeleteOperation` (model_update.go) - Depends on:
  - Pane.LoadDirectory() (existing)
  - Pane.SetCursor() (existing)
  - Pane.EnsureCursorVisible() (existing)
  - Pane.calculateCursorAfterDeletion() (new, from Phase 1)

## Risk Assessment

### Technical Risks

1. **Off-by-One Errors in Cursor Calculation**
   - **Risk**: Incorrect index calculation leads to cursor out of bounds
   - **Likelihood**: Medium
   - **Impact**: High (application panic)
   - **Mitigation**:
     - Strict bounds checking in calculateCursorAfterDeletion
     - Comprehensive unit tests covering all edge cases
     - Defensive programming: always clamp to [0, len(entries)-1]

2. **Unexpected Behavior with Concurrent Directory Changes**
   - **Risk**: Directory modified externally between deletion and reload
   - **Likelihood**: Low
   - **Impact**: Medium (cursor on wrong file)
   - **Mitigation**:
     - LoadDirectory() provides current state, calculation based on that
     - Worst case: cursor on different file, but still valid and visible
     - User can re-navigate as needed

3. **Performance Degradation in Large Directories**
   - **Risk**: Cursor calculation adds noticeable delay
   - **Likelihood**: Very Low
   - **Impact**: Low (minor UX issue)
   - **Mitigation**:
     - Calculation is O(1) (simple arithmetic, no iteration)
     - Benchmark to verify < 1ms requirement
     - No file system operations in calculation

### Implementation Risks

1. **Breaking Existing Delete Functionality**
   - **Risk**: Modification introduces regression
   - **Likelihood**: Low
   - **Impact**: High (core feature broken)
   - **Mitigation**:
     - Comprehensive testing of existing delete scenarios
     - Code review before merge
     - Gradual rollout with testing

2. **Incomplete Edge Case Coverage**
   - **Risk**: Rare scenario not handled, causing crash
   - **Likelihood**: Medium
   - **Impact**: Medium (poor UX in edge case)
   - **Mitigation**:
     - Systematic edge case enumeration
     - Table-driven tests for all scenarios
     - Manual testing of unusual situations

## Performance Considerations

1. **Cursor Calculation**
   - Simple arithmetic operations: O(1) time complexity
   - No file system access
   - Expected time: < 0.1ms (well under 1ms requirement)

2. **Directory Reload**
   - Performance dominated by LoadDirectory() (existing implementation)
   - Cursor calculation adds negligible overhead
   - No change to overall delete operation performance

3. **Scroll Adjustment**
   - adjustScroll() is existing, already optimized
   - Called once per delete operation (no change)

## Security Considerations

- **Authentication**: Not applicable - local file operation
- **Authorization**: Inherits existing file system permissions
- **Input Validation**:
  - Cursor index validated through bounds checking
  - deletedIndex clamped to valid range
  - No user input for cursor calculation (internal logic only)
- **Data Protection**: No sensitive data involved
- **Error Handling**: Deletion errors handled gracefully, cursor still positioned

## Open Questions

None - all requirements clarified through SPEC.md and requirements document.

## Future Enhancements

Items deferred to later releases (not in current scope):

### Phase 2 Features (from original spec):
- Remember cursor position across application restarts
- Configurable cursor positioning behavior (always next, always previous, smart)
- Undo delete operation (would need cursor to return to deleted file position)

### Not in Current Spec:
- Cursor memory during filter changes
- Cursor hints (show preview of where cursor will go before confirming delete)

## Success Criteria

### Functional Completeness
- [ ] All user stories (US1-US4) are implemented and verified
- [ ] All edge cases handled without errors
- [ ] Error handling works correctly (deletion fails, cursor still set)

### Quality Metrics
- [ ] Test coverage: 80%+ for modified code, 100% for calculateCursorAfterDeletion
- [ ] No critical bugs in manual testing
- [ ] Code follows Go best practices and duofm conventions

### Performance Metrics
- [ ] Cursor calculation < 1ms (verified by benchmark)
- [ ] No regression in delete operation performance
- [ ] UI responsive after deletion (< 100ms perceived delay)

### User Experience
- [ ] Cursor movement feels natural and predictable
- [ ] Work context preserved after deletion
- [ ] No confusing behavior in edge cases
- [ ] Matches behavior of standard file managers (MC, ranger)

## References

- **Specification**: `doc/tasks/fix-cursor-position-after-delete/SPEC.md`
- **Requirements**: `doc/tasks/fix-cursor-position-after-delete/要件定義書.md`
- **Bubble Tea Documentation**: https://github.com/charmbracelet/bubbletea
- **Go Testing**: https://go.dev/doc/tutorial/add-a-test
- **Related Code**:
  - `internal/ui/pane.go` - Cursor management methods
  - `internal/ui/model_update.go` - Delete operation flow
  - `internal/ui/pane_navigation.go` - Similar cursor preservation logic (MoveToParent)

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - Review IMPLEMENTATION.md and VERIFICATION.md
   - Confirm approach aligns with requirements
   - Address any questions or concerns

2. **Environment Setup**
   - Verify Go 1.21+ installed
   - Verify dependencies up to date (`go mod tidy`)
   - Ensure test environment working (`make test`)

3. **Begin Implementation**
   - Start with Phase 1 (calculateCursorAfterDeletion)
   - Follow TDD approach (write tests first where possible)
   - Commit after each phase completion

4. **Integration**
   - Run full test suite (`make test`)
   - Manual testing of all user scenarios
   - Code review and merge to main branch
