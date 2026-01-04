# Implementation Plan: Fix Permission Dialog Freeze Bug

## Overview

This implementation plan addresses a critical bug where the application becomes unresponsive after closing permission-related dialogs with the Esc key. The root cause is that affected dialogs set their `active` flag to false but fail to notify the Model to clear the dialog reference, resulting in all subsequent key inputs being delegated to an inactive dialog that ignores them.

## Objectives

- Fix the dialog freeze bug in PermissionDialog, RecursivePermDialog, and InputDialog
- Investigate and fix similar issues in remaining dialogs (12 dialogs in Category C)
- Implement comprehensive E2E tests to verify post-dialog keyboard input functionality
- Document dialog implementation best practices to prevent regression

## Prerequisites

### Development Environment
- Go 1.21 or later
- Make for build automation
- Working duofm development environment

### Dependencies

**External Dependencies:**
- `github.com/charmbracelet/bubbletea` (already installed)

**Internal Components:**
- Dialog interface (`internal/ui/dialog.go`)
- Model update system (`internal/ui/model_update_keyboard.go`)
- Message definitions (`internal/ui/messages.go`)
- Dialog implementations in `internal/ui/*_dialog.go`

### Knowledge Requirements
- Understanding of Bubble Tea's Elm Architecture (Model-Update-View pattern)
- Knowledge of Go message passing patterns
- Familiarity with duofm's dialog lifecycle

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea (github.com/charmbracelet/bubbletea)
- **Testing**: Go's built-in `testing` package

### Design Approach

**Message-Based Dialog Cancellation Pattern (Option 1 - Selected):**

This approach follows the existing pattern used by ConfirmDialog. When a dialog is canceled, it sends a cancellation message to the Model, which then clears the dialog reference.

**Rationale:**
- Explicit and maintainable
- Consistent with existing ConfirmDialog implementation
- Idiomatic Bubble Tea architecture (message-driven)
- Easier to debug and test
- Clear separation of concerns

**Alternative Approach (Rejected):**
Polling-based approach where Model checks `dialog.IsActive()` after each update. This was rejected because it is less explicit, not idiomatic in Bubble Tea, and has potential timing issues.

### Component Interaction

**Dialog Lifecycle Flow:**

```
User Action (Esc) → Dialog.Update() → Cancel Message → Model Handler → Clear Dialog Reference
                                                                      ↓
                                                         Resume Normal Key Handling
```

**Correct Message Flow:**

```
1. Dialog receives Esc key
2. Dialog sets active = false
3. Dialog returns cancellation message (tea.Cmd)
4. Model's Update() receives cancellation message
5. Model's message handler clears m.dialog = nil
6. Subsequent key inputs go to normal handlers
```

**Buggy Flow (Current State):**

```
1. Dialog receives Esc key
2. Dialog sets active = false
3. Dialog returns nil (NO MESSAGE SENT) ← BUG
4. Model's m.dialog remains non-nil ← BUG
5. Subsequent key inputs delegated to inactive dialog
6. Inactive dialog ignores inputs (early return)
7. Application appears frozen ← SYMPTOM
```

## Implementation Phases

### Phase 1: Fix Critical Dialogs

**Goal**: Fix the reported freeze bug in the three critical dialogs (PermissionDialog, RecursivePermDialog, InputDialog) to restore normal application behavior after dialog cancellation.

**Files to Modify**:

1. **`internal/ui/permission_dialog.go`**:
   - Modify Esc key handling to return cancellation message instead of nil

2. **`internal/ui/recursive_perm_dialog.go`**:
   - Modify Esc key handling at both step 1 and step 2 to return cancellation message

3. **`internal/ui/input_dialog.go`**:
   - Modify Esc key handling to return cancellation message with `cancelled` flag

4. **`internal/ui/model_permission.go`**:
   - Add message type definitions for permission dialog cancellations
   - Add message handlers to clear `m.dialog` when cancellation messages are received

5. **`internal/ui/model.go`**:
   - Modify `handleInputDialogResult()` to check for `cancelled` flag and return early

6. **`internal/ui/messages.go`**:
   - Add `cancelled bool` field to `inputDialogResultMsg` struct

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| PermissionDialog.Update() | Handle Esc key, send cancel message | Dialog is active | Dialog inactive, cancel message sent |
| RecursivePermDialog.Update() | Handle Esc key at any step, send cancel message | Dialog is active | Dialog inactive, cancel message sent |
| InputDialog.Update() | Handle Esc key, send cancel result with cancelled=true | Dialog is active | Dialog inactive, cancel result sent |
| Model.handlePermissionMessages() | Process permission dialog cancel messages | Cancel message received | m.dialog = nil |
| Model.handleInputDialogResult() | Process input dialog results, check cancelled flag | inputDialogResultMsg received | m.dialog = nil if cancelled=true |

**Processing Flow**:

```
PermissionDialog Cancellation:
1. User presses Esc
2. PermissionDialog.Update() processes Esc key
   ├─ Set d.active = false
   └─ Return permissionDialogCancelMsg
3. Model.Update() receives permissionDialogCancelMsg
4. Model.handlePermissionMessages() processes message
   ├─ Check message type is permissionDialogCancelMsg
   └─ Set m.dialog = nil
5. Normal key handling resumes

RecursivePermDialog Cancellation (similar flow):
1. User presses Esc (at step 1 or step 2)
2. RecursivePermDialog.Update() processes Esc
   ├─ Set d.active = false
   └─ Return recursivePermDialogCancelMsg
3. Model processes message and clears dialog

InputDialog Cancellation:
1. User presses Esc
2. InputDialog.Update() processes Esc
   ├─ Set d.active = false
   └─ Return inputDialogResultMsg{cancelled: true}
3. Model.handleInputDialogResult() checks cancelled flag
   ├─ If cancelled=true → return early (no-op)
   └─ m.dialog already nil'd
```

**Implementation Steps**:

1. **Define Cancellation Messages**
   - Add `permissionDialogCancelMsg` struct in `model_permission.go`
   - Add `recursivePermDialogCancelMsg` struct in `model_permission.go`
   - Add `cancelled bool` field to `inputDialogResultMsg` in `messages.go`

2. **Fix PermissionDialog**
   - Locate Esc key handling in `permission_dialog.go` (around line 106-108)
   - Change return from `(d, nil)` to `(d, func() tea.Msg { return permissionDialogCancelMsg{} })`
   - Add message handler in `handlePermissionMessages()` to clear dialog

3. **Fix RecursivePermDialog**
   - Locate Esc key handling in `recursive_perm_dialog.go` (around line 83-85)
   - Change return from `(d, nil)` to `(d, func() tea.Msg { return recursivePermDialogCancelMsg{} })`
   - Add message handler in `handlePermissionMessages()` to clear dialog

4. **Fix InputDialog**
   - Locate Esc key handling in `input_dialog.go` (around line 65-67)
   - Change return to include `cancelled: true` in inputDialogResultMsg
   - Update `handleInputDialogResult()` to check cancelled flag and return early

5. **Add Message Handlers**
   - In `model_permission.go`, add handlers for permission cancel messages
   - Ensure handlers set `m.dialog = nil` and return `(m, nil, true)`

**Dependencies**:
- Requires: Understanding of existing dialog message patterns (ConfirmDialog)
- Blocks: Phase 2 (investigation of other dialogs)

**Testing Approach**:

*Unit Tests* (`internal/ui/permission_dialog_test.go`, `recursive_perm_dialog_test.go`, `input_dialog_test.go`):

| Test Case | Description | Precondition | Expected Outcome |
|-----------|-------------|--------------|------------------|
| EscKeyCancellation | Esc key deactivates dialog and sends message | Dialog active | active=false, cancel message returned |
| InactiveIgnoresInput | Inactive dialog ignores all input | Dialog inactive | No action, nil command |
| ValidInputConfirm | Valid input triggers confirmation callback | Dialog active, valid input entered | Callback executed, dialog inactive |
| RecursiveEscAtStep1 | Esc at step 1 cancels dialog | Dialog at step 1 | Cancel message sent |
| RecursiveEscAtStep2 | Esc at step 2 cancels dialog | Dialog at step 2 | Cancel message sent |

*Integration Tests* (`internal/ui/model_test.go`):

| Test Scenario | Description | Steps | Expected Outcome |
|---------------|-------------|-------|------------------|
| PermissionDialogCancellation | Model clears dialog on cancel message | 1. Open dialog 2. Send Esc 3. Process cancel message | m.dialog = nil |
| PostDialogKeyboardInput | Keyboard works after dialog cancel | 1. Open/close dialog 2. Press 'j' | Cursor moves down |
| MultipleDialogSequence | Sequential dialog open/close | Open dialog A, close, open dialog B, close | Both close cleanly |

**Acceptance Criteria**:
- [ ] PermissionDialog closes on Esc and sends permissionDialogCancelMsg
- [ ] RecursivePermDialog closes on Esc at any step and sends recursivePermDialogCancelMsg
- [ ] InputDialog closes on Esc and sends inputDialogResultMsg with cancelled=true
- [ ] Model.handlePermissionMessages() clears m.dialog on cancel messages
- [ ] Model.handleInputDialogResult() handles cancelled flag correctly
- [ ] All unit tests pass with coverage ≥95% on modified functions
- [ ] All integration tests pass
- [ ] Manual testing confirms keyboard input works after dialog cancellation
- [ ] Manual testing confirms Ctrl+C (double press) works after dialog cancellation

**Estimated Effort**: 中 (3-5 days)

**Risks and Mitigation**:

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Message type conflicts with existing messages | Low | Medium | Review existing message types before creating new ones; use descriptive names |
| Breaking existing dialog behavior | Low | High | Thorough testing of dialog workflows; verify ConfirmDialog still works |
| Incomplete message handling | Medium | High | Add comprehensive unit tests; verify all code paths |

---

### Phase 2: Investigate and Fix Remaining Dialogs

**Goal**: Audit all Category C dialogs (12 dialogs) to identify and fix any similar Esc key handling issues, ensuring no other dialogs suffer from the same bug.

**Files to Investigate**:

1. `internal/ui/rename_input_dialog.go`
2. `internal/ui/bookmark_dialog.go`
3. `internal/ui/archive_name_dialog.go`
4. `internal/ui/compression_level_dialog.go`
5. `internal/ui/archive_progress_dialog.go`
6. `internal/ui/error_dialog.go`
7. `internal/ui/help_dialog.go`
8. `internal/ui/sort_dialog.go`
9. `internal/ui/compress_format_dialog.go`
10. `internal/ui/overwrite_dialog.go`
11. `internal/ui/context_menu_dialog.go`
12. `internal/ui/archive_conflict_dialog.go`
13. `internal/ui/archive_warning_dialog.go`

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| DialogAuditor | Examine Esc key handling in each dialog | Dialog source code available | Classification: correct/buggy/needs-fix |
| DialogFixer | Apply message-based cancel pattern to buggy dialogs | Buggy dialog identified | Esc handling fixed |
| MessageHandlerCreator | Add corresponding message handlers to Model | Cancel message type defined | Model clears dialog on cancel |

**Processing Flow**:

```
For Each Dialog in Category C:
1. Read dialog source code
2. Locate Esc key handling (tea.KeyEsc case)
3. Examine return value
   ├─ If returns cancel message → Category A (correct)
   ├─ If returns nil after setting active=false → Category B (buggy)
   └─ If no Esc handling → Investigate further
4. If Category B:
   ├─ Define cancel message type
   ├─ Modify Esc handling to return message
   └─ Add message handler in Model
5. Document findings
```

**Implementation Steps**:

1. **Audit Dialog Esc Handling**
   - For each dialog, search for `tea.KeyEsc` handling
   - Classify into: Correct (sends message), Buggy (returns nil), Unknown (no Esc handling)
   - Document findings in audit table

2. **Fix Buggy Dialogs**
   - For each buggy dialog, follow Phase 1 pattern:
     - Define cancel message type
     - Modify Esc key case to return cancel message
     - Add message handler to clear dialog

3. **Handle Special Cases**
   - Some dialogs may not need Esc handling (e.g., ErrorDialog that auto-closes)
   - Some dialogs may already be correct (similar to ConfirmDialog)
   - Document rationale for any dialogs left unchanged

4. **Create Audit Report**
   - List all dialogs with their status (fixed, correct, N/A)
   - Document any special considerations

**Dependencies**:
- Requires: Phase 1 complete (pattern established)
- Blocks: Phase 3 (E2E tests need all dialogs fixed)

**Testing Approach**:

*Unit Tests*:
- For each fixed dialog, add Esc key cancellation test
- Follow same pattern as Phase 1 unit tests

*Integration Tests*:
- Add test for each dialog type: open → cancel → verify dialog cleared
- Test post-dialog keyboard input for each dialog type

**Acceptance Criteria**:
- [ ] All 13 Category C dialogs audited and classified
- [ ] All buggy dialogs fixed with message-based cancellation
- [ ] Audit report documenting findings and decisions
- [ ] Unit tests added for all fixed dialogs
- [ ] Integration tests pass for all dialog types
- [ ] No regression in existing correct dialogs

**Estimated Effort**: 中 (3-4 days)

**Risks and Mitigation**:

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Unknown dialog interaction patterns | Medium | Medium | Thorough code review before modification; consult existing dialog implementations |
| Time-consuming audit process | High | Low | Prioritize dialogs by usage frequency; parallelize audit if possible |

---

### Phase 3: Add E2E Tests

**Goal**: Implement comprehensive end-to-end tests that verify dialog cancellation and post-dialog keyboard input behavior, as specifically requested by the user.

**Files to Create**:

1. **`internal/ui/dialog_e2e_test.go`**:
   - E2E tests for dialog cancel flows
   - Post-dialog operation tests
   - Multi-dialog sequence tests
   - Ctrl+C after dialog tests

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| E2ETestHarness | Setup complete Model with test panes | Test environment ready | Model initialized for testing |
| DialogFlowTester | Execute complete dialog open/cancel/verify flows | Model ready, dialog type specified | Flow verified |
| PostDialogBehaviorVerifier | Verify keyboard input and Ctrl+C after dialog | Dialog closed | Normal operations confirmed |

**Processing Flow**:

```
E2E Test Flow:
1. Initialize Model with test data
2. Open target dialog
   ├─ Verify dialog is non-nil
   └─ Verify dialog is active
3. Simulate Esc key press
   ├─ Update Model with Esc message
   └─ Process resulting cancel message
4. Verify dialog closed (m.dialog = nil)
5. Test normal operations
   ├─ Send cursor movement key ('j')
   ├─ Verify cursor moved
   └─ Verify no dialog opened
6. Test Ctrl+C quit mechanism
   ├─ Send first Ctrl+C
   ├─ Verify ctrlCPending = true
   ├─ Send second Ctrl+C
   └─ Verify tea.Quit returned
```

**Implementation Steps**:

1. **Create E2E Test Infrastructure**
   - Define test helper functions for Model initialization
   - Create test data (temporary directories, test files)
   - Implement test cleanup utilities

2. **Implement Permission Dialog E2E Tests**
   - Test complete flow: open PermissionDialog → cancel → verify normal ops
   - Test Ctrl+C after dialog cancellation
   - Verify file permissions unchanged after cancel

3. **Implement Recursive Permission Dialog E2E Tests**
   - Test cancel at step 1
   - Test cancel at step 2
   - Verify no permissions changed

4. **Implement Input Dialog E2E Tests**
   - Test cancel during file creation
   - Test cancel during rename
   - Verify no files created/renamed

5. **Implement Multi-Dialog Sequence Test**
   - Open and cancel multiple different dialog types in sequence
   - Verify normal operations after each cancellation
   - Verify no state corruption

6. **Implement Post-Dialog Operations Test**
   - After each dialog type cancellation, test:
     - Cursor movement (j/k keys)
     - Pane switching (Tab key)
     - File operations (still working)

**Dependencies**:
- Requires: Phase 1 and 2 complete (all dialogs fixed)
- Blocks: None (final testing phase)

**Testing Approach**:

*E2E Test Scenarios*:

| Test Name | Description | Steps | Expected Outcome |
|-----------|-------------|-------|------------------|
| PermissionDialogCancelAndContinue | Full flow from open to post-cancel ops | 1. Open dialog 2. Cancel 3. Move cursor 4. Quit with Ctrl+C | All operations work |
| RecursivePermDialogCancelStep1 | Cancel recursive dialog at first step | 1. Open 2. Cancel at step 1 3. Verify dialog closed | Dialog closed, no changes |
| RecursivePermDialogCancelStep2 | Cancel recursive dialog at second step | 1. Open 2. Complete step 1 3. Cancel at step 2 | Dialog closed, no changes |
| InputDialogCancel | Cancel input dialog | 1. Open for file creation 2. Cancel 3. Verify no file created | Dialog closed, no file |
| MultipleDialogSequence | Sequential open/cancel of different dialogs | 1. Open/cancel permission 2. Open/cancel help 3. Test ops | All work correctly |
| PostDialogCursorMovement | Cursor movement after dialog cancel | 1. Open/cancel dialog 2. Press j/k | Cursor moves normally |
| PostDialogCtrlCQuit | Ctrl+C quit after dialog | 1. Open/cancel dialog 2. Press Ctrl+C twice | Application quits |

*Test Coverage Requirements*:
- At least one E2E test per critical dialog type (PermissionDialog, RecursivePermDialog, InputDialog)
- At least one multi-dialog sequence test
- At least one post-dialog operation test per operation type (cursor, tab, quit)

**Acceptance Criteria**:
- [ ] E2E test file created with comprehensive tests
- [ ] All E2E tests pass consistently
- [ ] Tests verify dialog closure (m.dialog = nil)
- [ ] Tests verify post-dialog keyboard input works
- [ ] Tests verify Ctrl+C quit mechanism works after dialog
- [ ] Tests verify no file modifications on cancel
- [ ] Test output is clear and informative

**Estimated Effort**: 中 (3-4 days)

**Risks and Mitigation**:

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| E2E tests are flaky | Medium | Medium | Use deterministic test data; avoid timing dependencies; thorough cleanup |
| Tests are hard to maintain | Low | Medium | Use helper functions; clear test structure; good documentation |

---

### Phase 4: Documentation and Best Practices

**Goal**: Document dialog implementation best practices to prevent regression and guide future dialog development.

**Files to Create/Modify**:

1. **`doc/development/DIALOG_BEST_PRACTICES.md`**:
   - Dialog implementation guide
   - Common pitfalls and solutions
   - Code examples
   - Testing requirements

2. **`doc/CONTRIBUTING.md`** (modify):
   - Add dialog implementation section
   - Reference best practices document
   - Add to code review checklist

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| BestPracticesDoc | Provide comprehensive guide for dialog implementation | Bug fixed, patterns established | Future developers have clear guidance |
| CodeReviewChecklist | Add dialog-specific review items | Best practices documented | Reviewers catch dialog issues |
| ExampleCode | Provide working examples of correct dialog patterns | ConfirmDialog, PermissionDialog fixed | Copy-paste-ready examples available |

**Processing Flow**:

```
Documentation Creation:
1. Analyze correct dialog implementations
   ├─ ConfirmDialog (reference implementation)
   └─ Fixed PermissionDialog, RecursivePermDialog, InputDialog
2. Extract common patterns
   ├─ Message-based cancellation
   ├─ Active state management
   └─ Model integration
3. Identify common pitfalls
   ├─ Returning nil on Esc (the bug we fixed)
   ├─ Not clearing dialog in Model
   └─ Inconsistent active state
4. Document best practices
5. Add code examples
6. Create review checklist
```

**Implementation Steps**:

1. **Create Dialog Best Practices Document**
   - Introduction: Why proper dialog implementation matters
   - Architecture: Dialog lifecycle and message flow
   - Pattern: Message-based cancellation (with code example)
   - Anti-Pattern: Returning nil on Esc (with explanation)
   - Checklist: Requirements for new dialogs
   - Testing: Required tests for dialogs

2. **Provide Code Examples**
   - Minimal correct dialog implementation
   - Esc key handling pattern
   - Message handler pattern
   - Unit test example

3. **Update CONTRIBUTING.md**
   - Add "Implementing Dialogs" section
   - Reference best practices document
   - Add to "Before Submitting PR" checklist

4. **Create Code Review Checklist for Dialogs**
   - Dialog sends message on Esc
   - Model has corresponding message handler
   - Message handler clears m.dialog
   - Unit tests cover Esc handling
   - Integration tests verify Model integration

**Dependencies**:
- Requires: Phase 1 complete (patterns established)
- Blocks: None (documentation phase)

**Testing Approach**:

*No automated tests required for documentation*

*Manual Verification*:
- Review documentation for clarity and completeness
- Ask another developer to review for understandability
- Verify code examples compile and run

**Acceptance Criteria**:
- [ ] DIALOG_BEST_PRACTICES.md created with comprehensive content
- [ ] CONTRIBUTING.md updated with dialog section
- [ ] Code examples provided and verified
- [ ] Code review checklist includes dialog-specific items
- [ ] Documentation reviewed by at least one other developer
- [ ] Examples are copy-paste ready and functional

**Estimated Effort**: 小 (1-2 days)

**Risks and Mitigation**:

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Documentation becomes outdated | Medium | Low | Keep documentation close to code; reference in CONTRIBUTING.md |
| Developers don't read documentation | Medium | Medium | Make checklist visible; enforce in code review; provide examples |

---

## Complete File Structure

```
duofm/
├── cmd/duofm/
│   └── main.go                              # Entry point (no changes)
├── internal/ui/
│   ├── dialog.go                            # Dialog interface (no changes)
│   ├── messages.go                          # Add cancelled field to inputDialogResultMsg
│   ├── permission_dialog.go                 # Fix Esc handling (Phase 1)
│   ├── recursive_perm_dialog.go             # Fix Esc handling (Phase 1)
│   ├── input_dialog.go                      # Fix Esc handling (Phase 1)
│   ├── model_permission.go                  # Add cancel message handlers (Phase 1)
│   ├── model.go                             # Update inputDialogResult handler (Phase 1)
│   ├── model_update_keyboard.go             # No changes (bug is in dialogs, not routing)
│   ├── confirm_dialog.go                    # Reference implementation (no changes)
│   ├── rename_input_dialog.go               # Investigate and fix (Phase 2)
│   ├── bookmark_dialog.go                   # Investigate and fix (Phase 2)
│   ├── archive_name_dialog.go               # Investigate and fix (Phase 2)
│   ├── compression_level_dialog.go          # Investigate and fix (Phase 2)
│   ├── archive_progress_dialog.go           # Investigate and fix (Phase 2)
│   ├── error_dialog.go                      # Investigate (Phase 2)
│   ├── help_dialog.go                       # Investigate and fix (Phase 2)
│   ├── sort_dialog.go                       # Investigate and fix (Phase 2)
│   ├── compress_format_dialog.go            # Investigate and fix (Phase 2)
│   ├── overwrite_dialog.go                  # Investigate and fix (Phase 2)
│   ├── context_menu_dialog.go               # Investigate and fix (Phase 2)
│   ├── archive_conflict_dialog.go           # Investigate and fix (Phase 2)
│   ├── archive_warning_dialog.go            # Investigate and fix (Phase 2)
│   ├── permission_dialog_test.go            # Unit tests (Phase 1)
│   ├── recursive_perm_dialog_test.go        # Unit tests (Phase 1)
│   ├── input_dialog_test.go                 # Unit tests (Phase 1)
│   ├── model_test.go                        # Integration tests (Phase 1, 2)
│   └── dialog_e2e_test.go                   # E2E tests (Phase 3) - NEW FILE
├── doc/
│   ├── development/
│   │   └── DIALOG_BEST_PRACTICES.md         # Best practices guide (Phase 4) - NEW FILE
│   ├── tasks/
│   │   └── bugfix-permission-dialog-freeze/
│   │       ├── SPEC.md                      # This specification
│   │       └── IMPLEMENTATION.md            # This implementation plan
│   └── CONTRIBUTING.md                      # Update with dialog section (Phase 4)
├── go.mod
├── go.sum
└── Makefile
```

**File Descriptions**:

**Modified Files (Phase 1)**:
- `messages.go` - Message type definitions; add `cancelled` field to existing `inputDialogResultMsg`
- `permission_dialog.go` - Permission dialog implementation; fix Esc to return cancel message
- `recursive_perm_dialog.go` - Recursive permission dialog; fix Esc at both steps to return cancel message
- `input_dialog.go` - Input dialog for file/dir creation and rename; fix Esc to return cancelled result
- `model_permission.go` - Permission operation handlers; add cancel message handlers
- `model.go` - Main Model; update inputDialogResult handler to check cancelled flag

**Investigated/Fixed Files (Phase 2)**:
- 13 dialog files in Category C - Audit for similar issues and fix as needed

**New Files (Phase 3)**:
- `dialog_e2e_test.go` - End-to-end tests for dialog flows and post-dialog behavior

**New/Modified Files (Phase 4)**:
- `doc/development/DIALOG_BEST_PRACTICES.md` - Comprehensive dialog implementation guide (NEW)
- `doc/CONTRIBUTING.md` - Add dialog implementation section (MODIFY)

**Reference Files (No Changes)**:
- `confirm_dialog.go` - Reference implementation showing correct pattern
- `dialog.go` - Dialog interface definition
- `model_update_keyboard.go` - Key input routing (no bug here)

## Testing Strategy

### Overall Testing Approach

This bug fix requires a comprehensive testing strategy across three levels: unit, integration, and E2E. The focus is on verifying that dialogs close correctly and that keyboard input resumes normal operation afterward.

### Unit Testing

**Approach**:
- Use Go's built-in `testing` package
- Table-driven tests for multiple scenarios
- Mock Model interactions (dialogs tested in isolation)
- Focus on individual dialog behavior

**Test Coverage Goals**:
- Modified dialog files: ≥95% coverage
- New message handlers: 100% coverage
- Esc key handling paths: 100% coverage

**Key Test Areas**:

1. **Dialog Esc Key Handling** (per dialog):
   - Esc press deactivates dialog
   - Esc press returns cancel message (not nil)
   - Inactive dialog ignores subsequent input
   - Esc handling works at all dialog steps (for multi-step dialogs)

2. **Dialog Confirmation** (per dialog):
   - Valid input triggers confirmation callback
   - Callback receives correct parameters
   - Dialog deactivates after confirmation

3. **Dialog State Management**:
   - Active/inactive state transitions
   - No state corruption on multiple Esc presses

### Integration Testing

**Approach**:
- Test Model-Dialog interaction
- Simulate message passing between components
- Verify dialog lifecycle in context of Model
- Use real dialog instances with mocked file system where needed

**Test Coverage Goals**:
- All dialog cancellation flows: 100%
- Message handler paths: 100%
- Dialog cleanup (m.dialog = nil): 100%

**Key Test Scenarios**:

1. **Dialog Cancellation Integration**:
   - Model opens dialog
   - Send Esc key to Model
   - Verify cancel message generated
   - Verify Model processes message
   - Verify m.dialog = nil

2. **Post-Dialog Keyboard Input**:
   - Open and cancel dialog
   - Send normal key inputs (j, k, Tab)
   - Verify inputs processed normally
   - Verify no dialog state interference

3. **Multi-Dialog Sequence**:
   - Open different dialog types sequentially
   - Cancel each dialog
   - Verify clean state between dialogs
   - Verify no state leakage

### E2E Testing

**Approach**:
- Full user flow simulation
- Complete Model initialization with test data
- Simulate actual user interactions
- Verify end-to-end behavior

**Test Coverage Goals**:
- At least one complete flow per critical dialog
- All user stories covered (US1, US2, US3)
- Post-dialog operation verification

**Key Test Scenarios** (detailed in Phase 3):

1. **Permission Dialog Cancel Flow** (US1):
   - Open permission dialog
   - Cancel with Esc
   - Verify normal operations (cursor movement)
   - Verify Ctrl+C quit works
   - Verify file permissions unchanged

2. **Recursive Permission Dialog Cancel Flow** (US2):
   - Test cancel at step 1
   - Test cancel at step 2
   - Verify no permissions changed

3. **Input Dialog Cancel Flow** (US3):
   - Cancel during file creation
   - Cancel during rename
   - Verify no files created/renamed

4. **Multi-Dialog Sequence**:
   - Open and cancel multiple dialogs
   - Verify operations after each

### Manual Testing Checklist

Based on spec test scenarios and user stories:

**Critical Dialogs**:
- [ ] Open PermissionDialog with 'p', press Esc, verify UI responsive
- [ ] After closing PermissionDialog, press 'j' to move cursor down
- [ ] After closing PermissionDialog, press Ctrl+C twice to quit
- [ ] Open RecursivePermDialog (on directory), press Esc at step 1
- [ ] Open RecursivePermDialog, complete step 1, press Esc at step 2
- [ ] Open InputDialog (create file with 'n'), press Esc, verify no file created
- [ ] Open InputDialog (rename with 'r'), press Esc, verify no rename occurred

**Additional Dialogs (Phase 2)**:
- [ ] Test each Category C dialog after fix (if applicable)
- [ ] Verify no regression in ConfirmDialog

**Regression Testing**:
- [ ] Verify normal dialog confirmation (Enter key) still works
- [ ] Verify dialog UI renders correctly
- [ ] Verify dialog operations (permission changes, file creation) work when confirmed

## Dependencies

### External Dependencies

All external dependencies are already installed in the project:

| Package | Version | Purpose | Status |
|---------|---------|---------|--------|
| github.com/charmbracelet/bubbletea | v0.25.0 | TUI framework | Installed |
| github.com/charmbracelet/lipgloss | v0.9.1 | Styling | Installed |

No new external dependencies required.

### Internal Dependencies

**Component Dependencies Graph**:

```
Phase 1 (Critical Dialogs)
├─ Requires: Existing Dialog interface, Model update system
├─ Produces: Fixed dialogs, cancel messages, message handlers
└─ Blocks: Phase 2, Phase 3

Phase 2 (Remaining Dialogs)
├─ Requires: Phase 1 complete (pattern established)
├─ Produces: Audit report, additional fixes
└─ Blocks: Phase 3 (all dialogs must be fixed for comprehensive E2E tests)

Phase 3 (E2E Tests)
├─ Requires: Phase 1 and 2 complete (all dialogs fixed)
├─ Produces: E2E test suite
└─ Blocks: None (can run in parallel with Phase 4)

Phase 4 (Documentation)
├─ Requires: Phase 1 complete (pattern established)
├─ Produces: Best practices documentation
└─ Blocks: None
```

**Implementation Order** (respecting dependencies):
1. Phase 1 (mandatory, foundation for other phases)
2. Phase 2 and Phase 4 can run in parallel after Phase 1
3. Phase 3 should start after Phase 2 completes

**Critical Path**: Phase 1 → Phase 2 → Phase 3

**Component-Level Dependencies**:

| Component | Depends On | Reason |
|-----------|------------|--------|
| PermissionDialog fix | Dialog interface, tea.Msg | Must implement Dialog.Update() |
| Cancel message handlers | messages.go, model.go | Message types must be defined |
| Unit tests | Fixed dialog code | Tests verify the fix |
| Integration tests | Message handlers | Tests verify Model integration |
| E2E tests | All dialogs fixed | Tests verify complete flows |
| Documentation | Fixed implementation | Examples must reflect correct pattern |

## Risk Assessment

### Technical Risks

1. **Message Type Name Conflicts**
   - **Risk**: New cancel message types conflict with existing messages
   - **Likelihood**: Low
   - **Impact**: Medium (compilation errors, confusion)
   - **Mitigation**:
     - Review all existing message types in `messages.go` before creating new ones
     - Use descriptive, unique names (e.g., `permissionDialogCancelMsg` not just `cancelMsg`)
     - Follow existing naming conventions

2. **Breaking Existing Dialog Behavior**
   - **Risk**: Changes to dialog message handling affect correctly-working dialogs
   - **Likelihood**: Low
   - **Impact**: High (regression, breaking existing features)
   - **Mitigation**:
     - Thorough testing of all dialog types
     - Verify ConfirmDialog (reference implementation) still works
     - Run full regression test suite
     - Code review focused on changed areas

3. **Incomplete Message Handling**
   - **Risk**: Message handler doesn't cover all code paths
   - **Likelihood**: Medium
   - **Impact**: High (bug not fully fixed, edge cases remain)
   - **Mitigation**:
     - Add comprehensive unit tests for all message handler paths
     - Use table-driven tests for multiple scenarios
     - Test error cases and edge conditions

4. **Timing Issues with Message Passing**
   - **Risk**: Messages processed in unexpected order
   - **Likelihood**: Low (Bubble Tea guarantees message ordering)
   - **Impact**: Medium
   - **Mitigation**:
     - Follow Bubble Tea's message passing patterns
     - Don't rely on external state during message processing
     - Integration tests verify message flow

### Implementation Risks

1. **Scope Creep**
   - **Risk**: Attempting to refactor dialog architecture beyond bug fix scope
   - **Likelihood**: Medium
   - **Impact**: Medium (delayed timeline, increased complexity)
   - **Mitigation**:
     - Stick to specification requirements
     - Document future improvements separately (in open questions)
     - Defer architecture improvements to separate task

2. **Underestimated Complexity in Category C Dialogs**
   - **Risk**: Category C dialogs have unique patterns that don't fit message-based approach
   - **Likelihood**: Medium
   - **Impact**: Medium (requires alternative solution, delays Phase 2)
   - **Mitigation**:
     - Early audit of Category C dialogs in Phase 2
     - Document any special cases
     - Consult with team if unusual patterns found

3. **Test Environment Setup Complexity**
   - **Risk**: E2E tests require complex setup that's hard to maintain
   - **Likelihood**: Medium
   - **Impact**: Low (tests work but are fragile)
   - **Mitigation**:
     - Use helper functions for common setup
     - Use temporary directories for filesystem tests
     - Thorough cleanup after each test

4. **Documentation Not Used**
   - **Risk**: Developers don't reference best practices documentation
   - **Likelihood**: Medium
   - **Impact**: Medium (regression in future development)
   - **Mitigation**:
     - Make documentation easy to find (link from CONTRIBUTING.md)
     - Add to code review checklist
     - Provide copy-paste-ready examples

## Performance Considerations

### Performance Goals

- **Dialog closure response time**: < 10ms (imperceptible to user)
- **Memory cleanup**: Dialog struct garbage collected within 1 GC cycle
- **No memory leaks**: All dialog references properly cleared
- **No additional overhead**: Fix should not add measurable performance cost

### Performance Analysis

**Current Buggy Behavior**:
- Dialog remains in memory even after "close" (m.dialog still references it)
- Inactive dialog still processes (and ignores) all key inputs
- Potential minor memory leak (dialog never GC'd while app runs)

**Expected Performance After Fix**:
- Dialog reference cleared immediately (m.dialog = nil)
- No key input processing for non-existent dialog
- Dialog garbage collected in next GC cycle
- Cleaner memory profile

### Optimization Strategies

1. **Use Empty Struct Messages**
   - Cancel messages have no data: `type permissionDialogCancelMsg struct{}`
   - Zero allocation cost (empty structs are optimized by Go compiler)
   - No GC pressure

2. **Ensure Proper Dialog Cleanup**
   - Always set `m.dialog = nil` in message handlers
   - No dangling references to prevent GC
   - No goroutines or timers tied to dialog lifecycle

3. **No Additional Processing**
   - Message-based approach adds one message to Bubble Tea's message queue
   - Message processing is already part of normal update cycle
   - No additional goroutines or background work

4. **Verify No Performance Regression**
   - Benchmark dialog open/close cycle before and after fix
   - Measure memory usage with multiple dialog open/close cycles
   - Verify no increase in CPU usage

### Performance Testing

While not strictly required for this bug fix, consider:

```go
// Benchmark dialog cancellation performance
func BenchmarkPermissionDialogCancel(b *testing.B) {
    for i := 0; i < b.N; i++ {
        d := NewPermissionDialog("test.txt", false, 0644)
        d.Update(tea.KeyMsg{Type: tea.KeyEsc})
    }
}
```

**Performance Acceptance Criteria**:
- Dialog cancellation completes in < 10ms (verified by manual testing)
- No memory leaks detected after 100 dialog open/close cycles
- No performance regression compared to current (working) dialogs

## Security Considerations

### Security Impact Assessment

**Impact Level**: None (this is a usability/stability bug fix)

**Security-Relevant Aspects**:
- No changes to file system operations
- No changes to permission checks
- No changes to input validation
- No network or external system interaction
- No changes to privilege levels

### Security Invariants to Maintain

While this fix has no security impact, we must ensure existing security properties are preserved:

1. **File Permissions Unchanged on Cancel**
   - When PermissionDialog or RecursivePermDialog is canceled, no file permissions should be modified
   - Verified by test scenarios

2. **Input Validation Still Applied**
   - InputDialog cancellation doesn't bypass validation
   - If dialog is re-opened, validation still applies

3. **No Privilege Escalation**
   - Dialog cancellation doesn't create any new code paths that could bypass permission checks

**Security Testing**:
- Verify file permissions are unchanged after dialog cancellation (manual test)
- Verify no files created/modified when dialogs are canceled (E2E test)
- Run existing security-related tests to ensure no regression

## Success Criteria

### Functional Success Criteria

- [ ] **FR1**: PermissionDialog closes correctly on Esc key press
- [ ] **FR2**: RecursivePermDialog closes correctly on Esc key press at step 1
- [ ] **FR3**: RecursivePermDialog closes correctly on Esc key press at step 2
- [ ] **FR4**: InputDialog closes correctly on Esc key press
- [ ] **FR5**: After closing any dialog with Esc, all keyboard inputs work normally
- [ ] **FR6**: After closing any dialog, Ctrl+C (double press) can quit the application
- [ ] **FR7**: No files are modified when dialogs are canceled
- [ ] **FR8**: All Category C dialogs investigated and issues fixed (Phase 2)

### Technical Success Criteria

- [ ] **TC1**: All unit tests pass with ≥95% coverage on modified dialog files
- [ ] **TC2**: All integration tests pass
- [ ] **TC3**: All E2E tests pass
- [ ] **TC4**: No regression in existing dialog behavior (ConfirmDialog remains functional)
- [ ] **TC5**: Code follows existing patterns and conventions
- [ ] **TC6**: Message types are well-named and documented

### Quality Success Criteria

- [ ] **QC1**: User-reported freeze bug is resolved (verified by manual testing)
- [ ] **QC2**: Code is reviewed and approved
- [ ] **QC3**: Documentation is complete and accurate (Phase 4)
- [ ] **QC4**: Best practices guide is clear and useful (Phase 4)
- [ ] **QC5**: No new bugs introduced

### User Story Acceptance Criteria

**US1: Cancel Permission Change Dialog**
- [ ] Pressing Esc key closes the permission dialog
- [ ] After closing, all keyboard inputs work normally
- [ ] After closing, Ctrl+C (double press) can quit the application
- [ ] File permissions remain unchanged after canceling

**US2: Cancel Recursive Permission Dialog**
- [ ] Pressing Esc at step 1 (directory permissions) closes the dialog
- [ ] Pressing Esc at step 2 (file permissions) closes the dialog
- [ ] After closing, all keyboard inputs work normally
- [ ] No permissions are changed after canceling

**US3: Cancel Input Dialog**
- [ ] Pressing Esc closes the input dialog
- [ ] After closing, all keyboard inputs work normally
- [ ] No files are created or renamed after canceling

## Open Questions

### From Specification

**Resolved Questions**:
- **Q1**: Should we investigate all dialogs or only fix reported ones?
  - **A**: Investigate all dialogs in Category C to prevent similar issues (Phase 2)

**Pending Questions**:

- [ ] **Q2**: Should we create a common base dialog struct with correct Esc handling?
  - **Status**: Out of scope for this bug fix
  - **Recommendation**: Add to technical debt backlog as future improvement
  - **Impact**: Would prevent regression but requires significant refactoring

- [ ] **Q3**: Should we add a linter rule to catch this pattern?
  - **Status**: Consider as follow-up task
  - **Recommendation**: Custom linter rule to detect `case tea.KeyEsc:` that returns nil
  - **Impact**: Would prevent future bugs but requires linter development

### Implementation-Specific Questions

- [ ] **Q4**: Should ErrorDialog have Esc handling?
  - **Context**: ErrorDialog may be designed to require explicit acknowledgment (Enter key)
  - **Decision Needed**: Clarify expected behavior for error dialogs

- [ ] **Q5**: Should HelpDialog use same cancel message pattern?
  - **Context**: HelpDialog may have different lifecycle (informational vs. operational)
  - **Decision Needed**: Verify current HelpDialog Esc handling and align with pattern if needed

- [ ] **Q6**: Should we add metrics/logging for dialog cancellations?
  - **Context**: Could help understand user behavior and catch issues early
  - **Decision Needed**: Out of scope for bug fix, but useful for future monitoring

## Future Enhancements

Items identified during implementation planning but deferred to future releases:

### Phase 5+ Features (Not in Current Scope)

1. **Common Base Dialog Implementation**
   - Create `BaseDialog` struct with correct Esc handling built-in
   - All dialogs embed BaseDialog
   - Reduces code duplication
   - Prevents regression
   - **Effort**: Large (requires refactoring all dialogs)

2. **Dialog State Machine**
   - Formal state machine for dialog lifecycle
   - State: Opening → Active → Closing → Closed
   - Prevents invalid state transitions
   - **Effort**: Medium

3. **Dialog Testing Framework**
   - Reusable test helpers for dialog testing
   - Common test cases (Esc, Enter, inactive state)
   - Reduces test code duplication
   - **Effort**: Small

4. **Custom Linter Rule**
   - Detect dialogs that return nil on Esc
   - Enforce message-based cancellation pattern
   - Integrate into CI pipeline
   - **Effort**: Medium (requires linter development)

5. **Dialog Metrics/Telemetry**
   - Track dialog open/close events
   - Track cancellation vs. confirmation rates
   - Help identify UX issues
   - **Effort**: Small

### Not in Current Spec

**Architectural Improvements**:
- Dialog manager service (centralized dialog state)
- Dialog transition animations
- Dialog stacking (multiple dialogs)

**UX Improvements**:
- Visual feedback for Esc key press
- Undo/redo for dialog operations
- Dialog history (reopen last dialog)

## References

- **Specification**: `/home/sakura/cache/worktrees/duofm/bugfix-permission-dialog-freeze/doc/tasks/bugfix-permission-dialog-freeze/SPEC.md`
- **Bubble Tea Documentation**: https://github.com/charmbracelet/bubbletea
- **The Elm Architecture**: https://guide.elm-lang.org/architecture/
- **Correct Dialog Implementation Example**: `internal/ui/confirm_dialog.go`
- **Go Testing Documentation**: https://go.dev/doc/tutorial/add-a-test
- **Effective Go**: https://go.dev/doc/effective_go

**Related Code Files**:
- Dialog interface: `internal/ui/dialog.go`
- Message definitions: `internal/ui/messages.go`
- Model update routing: `internal/ui/model_update_keyboard.go`
- Permission handlers: `internal/ui/model_permission.go`

## Next Steps

After reviewing this implementation plan:

### 1. Review and Approval
- [ ] Review implementation plan with team
- [ ] Confirm approach and architecture decisions
- [ ] Resolve open questions (Q2-Q6)
- [ ] Approve estimated effort and timeline

### 2. Environment Setup
- [ ] Verify development environment is ready
- [ ] Ensure all dependencies are installed (`go mod download`)
- [ ] Create feature branch from main
- [ ] Verify tests run successfully (`make test`)

### 3. Begin Implementation

**Start with Phase 1 (Critical Dialogs)**:

1. Read reference implementation (`confirm_dialog.go`)
2. Review existing message types in `messages.go`
3. Fix PermissionDialog Esc handling
4. Add unit tests for PermissionDialog
5. Fix RecursivePermDialog Esc handling
6. Add unit tests for RecursivePermDialog
7. Fix InputDialog Esc handling and message type
8. Add unit tests for InputDialog
9. Add integration tests for Model-Dialog interaction
10. Verify all Phase 1 acceptance criteria

**Phase 1 Complete → Proceed to Phase 2**

### 4. Continuous Integration
- [ ] Run tests before each commit (`make test`)
- [ ] Verify test coverage (`go test -cover ./...`)
- [ ] Run linters if available (`go vet ./...`)
- [ ] Commit incrementally with clear messages

### 5. Testing and Validation
- [ ] Run full test suite after each phase
- [ ] Perform manual testing using checklist
- [ ] Verify no regression in existing features
- [ ] Test on actual terminal (not just tests)

### 6. Documentation and Completion
- [ ] Complete Phase 4 documentation
- [ ] Update CONTRIBUTING.md
- [ ] Verify all success criteria met
- [ ] Prepare for code review and merge
