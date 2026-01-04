# Feature: Fix Permission Dialog Freeze Bug

## Overview

A critical bug has been identified where the application becomes completely unresponsive after closing the permission change dialog with the Esc key. Users cannot use any keyboard input, including Ctrl+C to quit, forcing them to use `pkill duofm` to terminate the application. This specification defines the technical approach to fix this bug and prevent similar issues in other dialogs.

## Objectives

- Fix the permission dialog freeze bug to allow normal operation after canceling dialogs
- Identify and fix similar issues in other dialogs throughout the application
- Add comprehensive E2E tests that verify dialog closure and post-dialog behavior
- Establish best practices for dialog implementation to prevent regression

## User Stories

### US1: Cancel Permission Change Dialog
As a user, I want to cancel the permission change dialog with Esc key, so that I can return to normal file manager operations without changing any permissions.

**Acceptance Criteria:**
- [ ] Pressing Esc key closes the permission dialog
- [ ] After closing, all keyboard inputs work normally
- [ ] After closing, Ctrl+C (double press) can quit the application
- [ ] File permissions remain unchanged after canceling

### US2: Cancel Recursive Permission Dialog
As a user, I want to cancel the recursive permission change dialog at any step, so that I can abort the operation without affecting file permissions.

**Acceptance Criteria:**
- [ ] Pressing Esc at step 1 (directory permissions) closes the dialog
- [ ] Pressing Esc at step 2 (file permissions) closes the dialog
- [ ] After closing, all keyboard inputs work normally
- [ ] No permissions are changed after canceling

### US3: Cancel Input Dialog
As a user, I want to cancel any input dialog with Esc key, so that I can abort operations like file creation or renaming.

**Acceptance Criteria:**
- [ ] Pressing Esc closes the input dialog
- [ ] After closing, all keyboard inputs work normally
- [ ] No files are created or renamed after canceling

## Technical Requirements

### Functional Requirements
- **FR1:** Dialog Esc key handling must properly deactivate the dialog and notify the Model
- **FR2:** Model must clear the dialog reference (set to nil) when a dialog is closed
- **FR3:** All dialogs must implement consistent cancellation behavior
- **FR4:** Post-dialog keyboard input must function normally
- **FR5:** Ctrl+C quit mechanism must work after any dialog operation

### Non-Functional Requirements
- **NFR1 - Performance:** Dialog opening and closing must complete within 100ms
- **NFR2 - Reliability:** Dialog cancellation must succeed 100% of the time
- **NFR3 - Maintainability:** Dialog implementation patterns must be documented for future development
- **NFR4 - Testability:** All dialogs must have E2E tests verifying post-closure behavior

## Implementation Approach

### Root Cause Analysis

**Current Buggy Behavior:**

```
User presses Esc
    ↓
PermissionDialog.Update() sets active = false (line 107)
    ↓
Returns (d, nil) - NO MESSAGE SENT
    ↓
Model.dialog remains non-nil
    ↓
handleKeyInput() (line 21-24) delegates ALL key inputs to dialog
    ↓
Dialog.Update() returns early because !active (line 83-85)
    ↓
KEY INPUT IS IGNORED → FREEZE
```

**Root Cause:**
1. `PermissionDialog.Update()` sets `active = false` on Esc but does NOT send a message to Model
2. Model never receives notification that dialog should be cleared
3. `Model.dialog` remains non-nil
4. All subsequent key inputs are delegated to the inactive dialog
5. The inactive dialog ignores all inputs (early return at line 83-85)
6. Result: Application appears frozen, even Ctrl+C doesn't work

**Correct Behavior (as seen in ConfirmDialog):**

```
User presses Esc
    ↓
ConfirmDialog.Update() sets active = false (line 44)
    ↓
Returns dialogResultMsg with Cancelled=true (line 45-48)
    ↓
Model receives dialogResultMsg
    ↓
handleConfirmDialogResult() sets m.dialog = nil (line 395)
    ↓
Next key input goes to normal handlers
    ↓
WORKS CORRECTLY
```

### Architecture

**Dialog Lifecycle Pattern:**

```
┌─────────────────────────────────────┐
│         Dialog Active               │
│    (receives key inputs)            │
├─────────────────────────────────────┤
│  User Action (Esc/Enter/etc.)       │
├─────────────────────────────────────┤
│  Dialog.Update():                   │
│    1. Set active = false            │
│    2. Return cancellation message   │
├─────────────────────────────────────┤
│  Model.handleXxxMessages():         │
│    1. Receive cancellation message  │
│    2. Set m.dialog = nil            │
│    3. Clean up state                │
├─────────────────────────────────────┤
│  Dialog Closed                      │
│    (m.dialog == nil)                │
└─────────────────────────────────────┘
```

**Message Flow:**

```
Dialog → Cancel Message → Model Handler → m.dialog = nil
                                        ↓
                           Normal Key Input Handling
```

### Dialog Classification

Based on code investigation, dialogs fall into three categories:

**Category A: Correctly Implemented (No Changes Needed)**
- `ConfirmDialog` - Returns `dialogResultMsg` on Esc
- Properly implemented cancellation flow

**Category B: Buggy Implementation (Needs Fix)**
- `PermissionDialog` - Sets active=false, returns nil on Esc
- `RecursivePermDialog` - Sets active=false, returns nil on Esc
- `InputDialog` - Sets active=false, returns nil on Esc

**Category C: Requires Investigation**
- `RenameInputDialog`
- `BookmarkDialog`
- `ArchiveNameDialog`
- `CompressionLevelDialog`
- `ArchiveProgressDialog`
- `ErrorDialog`
- `HelpDialog`
- `SortDialog`
- `CompressFormatDialog`
- `OverwriteDialog`
- `ContextMenuDialog`

### Fix Strategy

#### Option 1: Message-Based Approach (Recommended)

**Pros:**
- Follows existing pattern from ConfirmDialog
- Explicit message passing makes flow clear
- Easier to debug
- Consistent with Bubble Tea architecture

**Cons:**
- Requires creating new message types or reusing existing ones
- More code changes

**Implementation:**

For each buggy dialog:
1. Define or reuse a cancellation message type
2. Return the message on Esc key
3. Handle the message in Model to set dialog = nil

Example for PermissionDialog:
```go
// In permission_dialog.go
case tea.KeyEsc:
    d.active = false
    return d, func() tea.Msg {
        return permissionDialogCancelMsg{}
    }

// In model_permission.go
type permissionDialogCancelMsg struct{}

func (m Model) handlePermissionMessages(msg tea.Msg) (Model, tea.Cmd, bool) {
    if _, ok := msg.(permissionDialogCancelMsg); ok {
        m.dialog = nil
        return m, nil, true
    }
    // ... existing code
}
```

#### Option 2: IsActive Polling Approach

**Pros:**
- Minimal code changes
- No new message types needed

**Cons:**
- Less explicit, harder to understand
- Polling is not idiomatic in Bubble Tea
- Potential timing issues

**Implementation:**
```go
// In model_update_keyboard.go
func (m Model) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    if m.dialog != nil {
        var cmd tea.Cmd
        m.dialog, cmd = m.dialog.Update(msg)

        // NEW: Check if dialog became inactive
        if !m.dialog.IsActive() {
            m.dialog = nil
        }

        return m, cmd
    }
    // ... rest of code
}
```

**Decision: Use Option 1 (Message-Based Approach)**

Reasons:
- More explicit and maintainable
- Consistent with existing ConfirmDialog pattern
- Better for debugging and testing
- Idiomatic Bubble Tea architecture

### Affected Dialogs and Fix Details

#### 1. PermissionDialog

**Current Buggy Code:**
```go
// permission_dialog.go:106-108
case tea.KeyEsc:
    d.active = false
    return d, nil  // BUG: No message sent
```

**Fixed Code:**
```go
case tea.KeyEsc:
    d.active = false
    return d, func() tea.Msg {
        return permissionDialogCancelMsg{}
    }
```

**Message Handler:**
```go
// model_permission.go
type permissionDialogCancelMsg struct{}

func (m Model) handlePermissionMessages(msg tea.Msg) (Model, tea.Cmd, bool) {
    if _, ok := msg.(permissionDialogCancelMsg); ok {
        m.dialog = nil
        return m, nil, true
    }
    // ... existing handlers
}
```

#### 2. RecursivePermDialog

**Current Buggy Code:**
```go
// recursive_perm_dialog.go:83-85
case tea.KeyEsc:
    d.active = false
    return d, nil  // BUG: No message sent
```

**Fixed Code:**
```go
case tea.KeyEsc:
    d.active = false
    return d, func() tea.Msg {
        return recursivePermDialogCancelMsg{}
    }
```

**Message Handler:**
```go
// model_permission.go
type recursivePermDialogCancelMsg struct{}

func (m Model) handlePermissionMessages(msg tea.Msg) (Model, tea.Cmd, bool) {
    if _, ok := msg.(recursivePermDialogCancelMsg); ok {
        m.dialog = nil
        return m, nil, true
    }
    // ... existing handlers
}
```

#### 3. InputDialog

**Current Buggy Code:**
```go
// input_dialog.go:65-67
case tea.KeyEsc:
    d.active = false
    return d, nil  // BUG: No message sent
```

**Fixed Code:**
```go
case tea.KeyEsc:
    d.active = false
    return d, func() tea.Msg {
        return inputDialogResultMsg{
            cancelled: true,
        }
    }
```

**Message Handler:**
Since InputDialog already uses `inputDialogResultMsg`, we need to add a `cancelled` field:

```go
// messages.go
type inputDialogResultMsg struct {
    operation string
    input     string
    oldName   string
    err       error
    cancelled bool  // NEW FIELD
}

// model.go
func (m Model) handleInputDialogResult(msg inputDialogResultMsg) (tea.Model, tea.Cmd) {
    m.dialog = nil

    if msg.cancelled {  // NEW CHECK
        return m, nil
    }

    // ... existing code
}
```

### File Structure

```
internal/ui/
├── permission_dialog.go           # Fix Esc handling
├── recursive_perm_dialog.go       # Fix Esc handling
├── input_dialog.go                # Fix Esc handling
├── model_permission.go            # Add cancel message handlers
├── model.go                       # Update inputDialogResultMsg handler
├── messages.go                    # Add cancelled field to inputDialogResultMsg
└── (other dialogs)                # Investigate and fix as needed
```

### Dependencies

**Internal Dependencies:**
- Dialog interface (`internal/ui/dialog.go`)
- Model update system (`internal/ui/model_update.go`)
- Message handling (`internal/ui/messages.go`)

**External Dependencies:**
- `github.com/charmbracelet/bubbletea` - TUI framework

## Test Scenarios

### Unit Tests

#### PermissionDialog Tests

**Test File:** `internal/ui/permission_dialog_test.go`

```go
func TestPermissionDialog_EscKeyCancellation(t *testing.T) {
    // Arrange
    dialog := NewPermissionDialog("test.txt", false, 0644)
    escMsg := tea.KeyMsg{Type: tea.KeyEsc}

    // Act
    updatedDialog, cmd := dialog.Update(escMsg)

    // Assert
    assert.False(t, updatedDialog.IsActive(), "Dialog should be inactive")
    assert.NotNil(t, cmd, "Command should be returned")

    // Execute command to get message
    msg := cmd()
    _, ok := msg.(permissionDialogCancelMsg)
    assert.True(t, ok, "Should return permissionDialogCancelMsg")
}

func TestPermissionDialog_EnterWithValidInput(t *testing.T) {
    // Arrange
    dialog := NewPermissionDialog("test.txt", false, 0644)
    callbackExecuted := false
    dialog.SetOnConfirm(func(mode string, recursive bool) tea.Cmd {
        callbackExecuted = true
        assert.Equal(t, "755", mode)
        assert.False(t, recursive)
        return nil
    })

    // Input "755"
    dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}})
    dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
    dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})

    // Act
    updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

    // Assert
    assert.False(t, updatedDialog.IsActive())
    assert.NotNil(t, cmd)
    cmd() // Execute callback
    assert.True(t, callbackExecuted)
}

func TestPermissionDialog_InactiveIgnoresInput(t *testing.T) {
    // Arrange
    dialog := NewPermissionDialog("test.txt", false, 0644)
    dialog.Update(tea.KeyMsg{Type: tea.KeyEsc}) // Deactivate

    // Act
    updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}})

    // Assert
    assert.Nil(t, cmd, "Inactive dialog should return nil command")
    assert.False(t, updatedDialog.IsActive())
}
```

#### RecursivePermDialog Tests

```go
func TestRecursivePermDialog_EscAtStep1(t *testing.T) {
    // Arrange
    dialog := NewRecursivePermDialog("testdir")

    // Act
    updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

    // Assert
    assert.False(t, updatedDialog.IsActive())
    assert.NotNil(t, cmd)

    msg := cmd()
    _, ok := msg.(recursivePermDialogCancelMsg)
    assert.True(t, ok)
}

func TestRecursivePermDialog_EscAtStep2(t *testing.T) {
    // Arrange
    dialog := NewRecursivePermDialog("testdir")

    // Complete step 1
    dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7', '5', '5'}})
    dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

    // Act - Esc at step 2
    updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

    // Assert
    assert.False(t, updatedDialog.IsActive())
    assert.NotNil(t, cmd)
}
```

#### InputDialog Tests

```go
func TestInputDialog_EscKeyCancellation(t *testing.T) {
    // Arrange
    dialog := NewInputDialog("Enter name:", func(input string) tea.Cmd {
        return nil
    })

    // Act
    updatedDialog, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})

    // Assert
    assert.False(t, updatedDialog.IsActive())
    assert.NotNil(t, cmd)

    msg := cmd()
    result, ok := msg.(inputDialogResultMsg)
    assert.True(t, ok)
    assert.True(t, result.cancelled)
}
```

### Integration Tests

#### Model-Dialog Integration Tests

```go
func TestModel_PermissionDialogCancellation(t *testing.T) {
    // Arrange
    m := NewModel()
    m.ready = true
    setupTestPanes(&m)

    // Open permission dialog
    m, _ = m.handlePermission()
    require.NotNil(t, m.dialog, "Dialog should be open")

    // Act - Send Esc key
    m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

    // Execute command to send cancel message
    if cmd != nil {
        msg := cmd()
        m, _ = m.Update(msg)
    }

    // Assert
    assert.Nil(t, m.dialog, "Dialog should be closed")
}

func TestModel_PostDialogKeyboardInput(t *testing.T) {
    // Arrange
    m := NewModel()
    m.ready = true
    setupTestPanes(&m)

    initialCursor := m.getActivePane().cursor

    // Open and close dialog
    m, _ = m.handlePermission()
    m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
    if cmd != nil {
        msg := cmd()
        m, _ = m.Update(msg)
    }

    // Act - Press 'j' to move cursor down
    m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

    // Assert
    assert.Nil(t, m.dialog, "Dialog should remain closed")
    assert.Greater(t, m.getActivePane().cursor, initialCursor, "Cursor should have moved")
}
```

### E2E Tests

#### Full User Flow Tests

```go
func TestE2E_PermissionDialogCancelAndContinue(t *testing.T) {
    // Test Plan:
    // 1. Open permission dialog
    // 2. Cancel with Esc
    // 3. Verify normal operations work
    // 4. Verify Ctrl+C works

    // Arrange
    m := NewModel()
    m.ready = true
    setupTestPanes(&m)

    // Act 1: Open dialog
    m, _ = m.handlePermission()
    assert.NotNil(t, m.dialog, "Step 1: Dialog should open")

    // Act 2: Cancel with Esc
    m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
    if cmd != nil {
        m, _ = m.Update(cmd())
    }
    assert.Nil(t, m.dialog, "Step 2: Dialog should close")

    // Act 3: Normal operations
    initialCursor := m.getActivePane().cursor
    m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
    assert.Greater(t, m.getActivePane().cursor, initialCursor, "Step 3: Cursor should move")

    // Act 4: Ctrl+C (first press)
    m, _ = m.Update(tea.KeyMsg{String: "ctrl+c"})
    assert.True(t, m.ctrlCPending, "Step 4a: First Ctrl+C should set pending")

    // Act 5: Ctrl+C (second press)
    m, cmd = m.Update(tea.KeyMsg{String: "ctrl+c"})
    assert.Equal(t, tea.Quit, cmd, "Step 4b: Second Ctrl+C should quit")
}

func TestE2E_MultipleDialogsCancelSequence(t *testing.T) {
    // Test opening and canceling multiple different dialogs
    m := NewModel()
    m.ready = true
    setupTestPanes(&m)

    // Dialog 1: Permission dialog
    m, _ = m.handlePermission()
    assert.NotNil(t, m.dialog)
    m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
    if cmd != nil {
        m, _ = m.Update(cmd())
    }
    assert.Nil(t, m.dialog)

    // Dialog 2: Help dialog
    m.dialog = NewHelpDialog()
    m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
    if cmd != nil {
        m, _ = m.Update(cmd())
    }
    assert.Nil(t, m.dialog)

    // Verify normal operations still work
    m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
    assert.Nil(t, m.dialog, "No dialog should be open after sequence")
}
```

### Test Coverage Requirements

- Unit test coverage: ≥ 95% for modified files
- Integration test coverage: All dialog cancel flows
- E2E test coverage: At least one complete user flow per affected dialog

**Modified Files Requiring Tests:**
1. `permission_dialog.go` - Unit tests for Esc handling
2. `recursive_perm_dialog.go` - Unit tests for Esc handling at both steps
3. `input_dialog.go` - Unit tests for Esc handling
4. `model_permission.go` - Integration tests for message handling
5. `model.go` - Integration tests for inputDialogResultMsg handling

## Error Handling

### Error Scenarios

| Scenario | Current Behavior | Expected Behavior | Error Handling |
|----------|-----------------|-------------------|----------------|
| Dialog already inactive when Esc pressed | Returns nil, no effect | No-op, return immediately | Early return if !d.active |
| Multiple Esc presses | Repeated inactive state setting | First Esc closes, subsequent ignored | Check active state before processing |
| Message handler receives unexpected message type | Not handled | Ignored, passed to next handler | Return (m, nil, false) |

### Error Messages

No user-facing error messages are needed for this bug fix. The fix is transparent to users - dialogs will simply close correctly.

## Performance Optimization

### Performance Goals
- Dialog closure response time: < 10ms (imperceptible to user)
- Memory cleanup: Dialog struct garbage collected within 1 GC cycle
- No memory leaks from unclosed dialogs

### Optimization Strategies
- Use simple message types (empty structs where possible) to minimize allocation
- Ensure all dialog references are properly nil'd to allow GC
- No additional goroutines or background processing needed

## Security Considerations

This bug fix has no security implications. It addresses a usability and stability issue without touching any security-sensitive code paths.

## Success Criteria

- [ ] PermissionDialog closes correctly on Esc key press
- [ ] RecursivePermDialog closes correctly on Esc key press at any step
- [ ] InputDialog closes correctly on Esc key press
- [ ] After closing any dialog with Esc, all keyboard inputs work normally
- [ ] After closing any dialog, Ctrl+C (double press) can quit the application
- [ ] No files are modified when dialogs are canceled
- [ ] All unit tests pass with ≥95% coverage
- [ ] All integration tests pass
- [ ] All E2E tests pass
- [ ] No regression in existing dialog behavior (ConfirmDialog, etc.)
- [ ] User-reported freeze bug is resolved

## Open Questions

- [x] **Q1:** Should we investigate all dialogs or only fix reported ones?
  - **A:** Investigate all dialogs in Category C to prevent similar issues

- [ ] **Q2:** Should we create a common base dialog struct with correct Esc handling?
  - **Decision Needed:** Would be ideal for future, but out of scope for this bug fix. Add to technical debt backlog.

- [ ] **Q3:** Should we add a linter rule to catch this pattern?
  - **Decision Needed:** Could prevent regression. Consider as follow-up task.

## Implementation Phases

### Phase 1: Fix Critical Dialogs (Priority: High)
**Goals:** Fix the reported bug and most critical dialogs
**Deliverables:**
- Fix PermissionDialog Esc handling
- Fix RecursivePermDialog Esc handling
- Fix InputDialog Esc handling
- Add message handlers in Model
- Unit tests for all three dialogs
- Integration tests for Model-Dialog interaction

**Estimated Effort:** 4-6 hours

### Phase 2: Investigate Remaining Dialogs (Priority: Medium)
**Goals:** Ensure no other dialogs have similar issues
**Deliverables:**
- Audit all Category C dialogs
- Fix any identified issues
- Document findings

**Estimated Effort:** 2-4 hours

### Phase 3: E2E Tests (Priority: Medium)
**Goals:** Add comprehensive E2E tests
**Deliverables:**
- E2E test for permission dialog cancel flow
- E2E test for multiple dialog sequence
- E2E test for post-dialog operations
- E2E test for Ctrl+C after dialog

**Estimated Effort:** 3-5 hours

### Phase 4: Documentation (Priority: Low)
**Goals:** Document best practices to prevent regression
**Deliverables:**
- Dialog implementation guide
- Code review checklist updates
- CONTRIBUTING.md updates

**Estimated Effort:** 2 hours

## References

- Bubble Tea Documentation: https://github.com/charmbracelet/bubbletea
- The Elm Architecture: https://guide.elm-lang.org/architecture/
- Correct Dialog Implementation Example: `internal/ui/confirm_dialog.go`
- Bug Report Issue: User reported via feedback
