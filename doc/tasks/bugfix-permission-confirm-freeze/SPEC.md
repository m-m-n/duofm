# Permission Dialog Confirm Freeze Bug Fix Specification

## 1. Bug Description

### 1.1 Summary
When a user opens a permission change dialog, enters a valid permission value, and presses Enter to confirm, the application becomes completely unresponsive and cannot accept any keyboard input.

### 1.2 Affected Components
- `PermissionDialog`: Single file/directory permission changes
- `RecursivePermDialog`: Recursive permission changes (when "Recursively" option is selected for directories)
- `handlePermissionOperationComplete`: Permission operation completion handler
- `handleRecursivePermissionComplete`: Recursive permission operation completion handler

### 1.3 Related Previous Fix
Commit `c8fbfeb` fixed a similar freeze bug that occurred when cancelling dialogs with the Esc key. This bug follows the same pattern but occurs during confirmation (Enter key).

### 1.4 Severity
**Critical** - The application becomes completely unusable and requires force termination.

## 2. Reproduction Steps

### 2.1 Single File Permission Change
1. Launch duofm
2. Navigate to and select any file or directory
3. Press `p` key to open the permission change dialog
4. Enter a valid 3-digit octal permission value (e.g., `644`)
5. Press Enter to confirm
6. **Expected**: Permission is changed, dialog closes, normal operation resumes
7. **Actual**: Application freezes, no keyboard input is accepted

### 2.2 Recursive Permission Change
1. Launch duofm
2. Select a directory
3. Press `p` key to open the permission change dialog
4. Use Tab/Space to select "Recursively (all contents)"
5. Enter a permission value and press Enter (Step 1: directory permissions)
6. Enter another permission value and press Enter (Step 2: file permissions)
7. **Expected**: Permissions are changed recursively, normal operation resumes
8. **Actual**: Application freezes

## 3. Root Cause Analysis

### 3.1 Dialog System Architecture

The duofm dialog system stores the active dialog in `Model.dialog`. Key inputs are processed in `handleKeyInput`, which delegates to the dialog when one exists:

```go
// internal/ui/model_update_keyboard.go
func (m Model) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // When dialog is open, delegate to dialog
    if m.dialog != nil {
        var cmd tea.Cmd
        m.dialog, cmd = m.dialog.Update(msg)
        return m, cmd
    }
    // ... normal key handling
}
```

### 3.2 The Problem

When Enter is pressed in `PermissionDialog.Update`:

```go
// internal/ui/permission_dialog.go
case tea.KeyEnter:
    if err := fsops.ValidatePermissionMode(d.inputValue); err != nil {
        d.errorMsg = err.Error()
        return d, nil
    }
    d.Close()  // Sets d.active = false
    if d.onConfirm != nil {
        recursive := d.recursiveOption == 1
        return d, d.onConfirm(d.inputValue, recursive)  // Returns command
    }
```

The `onConfirm` callback returns `permissionOperationCompleteMsg`. This message is handled by `handlePermissionOperationComplete`:

```go
// internal/ui/model_permission.go
func (m Model) handlePermissionOperationComplete(msg permissionOperationCompleteMsg) (tea.Model, tea.Cmd) {
    if msg.success {
        m.statusMessage = fmt.Sprintf("Permission changed: %s", filepath.Base(msg.path))
        // ... refresh pane
        return m, statusMessageClearCmd(3 * time.Second)
    }
    // ... error handling
    return m, statusMessageClearCmd(5 * time.Second)
}
```

**Critical Bug**: `m.dialog = nil` is never set in `handlePermissionOperationComplete`.

### 3.3 Freeze Mechanism

1. Dialog's `Close()` sets `active = false`
2. `Model.dialog` still holds reference to the inactive dialog
3. Subsequent key inputs are delegated to the inactive dialog
4. Inactive dialog's `Update()` immediately returns without processing:
   ```go
   func (d *PermissionDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
       if !d.IsActive() {
           return d, nil  // Does nothing
       }
       // ...
   }
   ```
5. All keyboard input is silently ignored, causing apparent freeze

### 3.4 Comparison with Previous Fix

The Esc key cancellation fix (c8fbfeb) works because:
1. `PermissionDialog.Update` returns `permissionDialogCancelMsg`
2. `handlePermissionMessages` handles this message and sets `m.dialog = nil`

The confirmation path lacks this dialog cleanup step.

## 4. Fix Requirements

### 4.1 Functional Requirements

#### FR1: PermissionDialog Confirmation Handling
- **FR1.1**: After permission change operation completes, `m.dialog` MUST be set to `nil`
- **FR1.2**: Dialog reference MUST be cleared regardless of success or failure
- **FR1.3**: Normal keyboard operations MUST function correctly after dialog closes

#### FR2: RecursivePermDialog Confirmation Handling
- **FR2.1**: After recursive permission change completes, `m.dialog` MUST be set to `nil`
- **FR2.2**: Exception: When error report dialog is displayed, `m.dialog` should be set to the new dialog
- **FR2.3**: Error report dialog transition MUST be handled correctly

#### FR3: Batch Permission Change Handling
- **FR3.1**: After batch permission change completes, dialog reference MUST be properly managed
- **FR3.2**: Transition from progress dialog to error report dialog MUST work correctly
- **FR3.3**: Small batch operations (under ProgressThreshold, no progress dialog) MUST also clear `m.dialog`

### 4.2 Non-Functional Requirements

#### NFR1: Backward Compatibility
- Esc key cancellation behavior MUST remain unchanged
- Other dialogs (ErrorDialog, HelpDialog, etc.) MUST not be affected

#### NFR2: Code Consistency
- Follow the same pattern established in the Esc cancellation fix
- Adhere to dialog best practices documented in `doc/development/DIALOG_BEST_PRACTICES.md`

## 5. Implementation Notes

### 5.1 Files to Modify

#### `internal/ui/model_permission.go`

**handlePermissionOperationComplete**:
```go
func (m Model) handlePermissionOperationComplete(msg permissionOperationCompleteMsg) (tea.Model, tea.Cmd) {
    m.dialog = nil  // ADD THIS LINE

    if msg.success {
        m.statusMessage = fmt.Sprintf("Permission changed: %s", filepath.Base(msg.path))
        // ...
    }
    // ...
}
```

**handleRecursivePermissionComplete**:
```go
func (m Model) handleRecursivePermissionComplete(msg recursivePermissionCompleteMsg) (tea.Model, tea.Cmd) {
    // ... refresh pane logic

    if len(msg.errors) > 0 {
        errorDialog := NewPermissionErrorReportDialog(...)
        m.dialog = errorDialog  // Sets new dialog
        return m, nil
    }

    // ADD: Clear dialog when no errors
    m.dialog = nil

    m.statusMessage = fmt.Sprintf("Recursive permissions changed: %d files successful", msg.successCount)
    // ...
}
```

**handleBatchPermissionComplete** (FIX REQUIRED):
The current implementation only clears the progress dialog:
```go
if _, ok := m.dialog.(*PermissionProgressDialog); ok {
    m.dialog = nil  // Only clears PermissionProgressDialog
}
```

**Bug**: For small batches (under ProgressThreshold), `PermissionDialog` remains in `m.dialog`, causing freeze.

**Required Fix**:
```go
func (m Model) handleBatchPermissionComplete(msg batchPermissionCompleteMsg) (tea.Model, tea.Cmd) {
    // Clear any permission-related dialog (both PermissionDialog and PermissionProgressDialog)
    m.dialog = nil  // ADD THIS LINE - unconditionally clear dialog first

    // Clear marks after batch operation (even if some failed)
    m.getActivePane().ClearMarks()

    // ... rest of the function

    // If there are errors, show error report dialog
    if len(msg.errors) > 0 {
        errorDialog := NewPermissionErrorReportDialog(...)
        m.dialog = errorDialog  // Set new error dialog
        return m, nil
    }
    // ...
}
```

### 5.2 Alternative Approach (Not Recommended)

An alternative would be to have the dialog return a confirmation message similar to cancellation:
```go
// NOT RECOMMENDED - adds unnecessary complexity
type permissionDialogConfirmMsg struct {
    mode      string
    recursive bool
}
```

This approach is not recommended because:
1. The operation already returns `permissionOperationCompleteMsg`
2. It would add redundant message handling
3. Clearing dialog reference in completion handler is simpler and consistent

## 6. Test Requirements

### 6.1 Unit Tests

#### UT1: PermissionDialog Enter Key Handling
- Verify Enter key with valid input returns command from onConfirm callback
- Verify Enter key with invalid input shows error and keeps dialog open

#### UT2: RecursivePermDialog Enter Key Handling
- Verify step 1 to step 2 transition on Enter
- Verify step 2 completion returns command from onConfirm callback

### 6.2 Integration Tests

Add to `internal/ui/dialog_integration_test.go`:

#### IT1: PermissionDialogConfirmationIntegration
```go
func TestPermissionDialogConfirmationIntegration(t *testing.T) {
    // Setup: Create test model and dialog
    // Action: Enter valid permission, press Enter
    // Verify: permissionOperationCompleteMsg is processed
    // Verify: m.dialog becomes nil
    // Verify: Subsequent key presses work (j/k navigation)
}
```

#### IT2: RecursivePermDialogConfirmationIntegration
```go
func TestRecursivePermDialogConfirmationIntegration(t *testing.T) {
    // Setup: Create test model and recursive dialog
    // Action: Complete both steps with Enter
    // Verify: recursivePermissionCompleteMsg is processed
    // Verify: m.dialog becomes nil (or error report dialog)
    // Verify: Subsequent key presses work
}
```

#### IT3: BatchPermissionConfirmationIntegration
```go
func TestBatchPermissionConfirmationIntegration(t *testing.T) {
    // Setup: Create test model with multiple marked files
    // Action: Open permission dialog, enter value, press Enter
    // Verify: batchPermissionCompleteMsg is processed
    // Verify: m.dialog becomes nil (both small and large batches)
    // Verify: Marks are cleared
    // Verify: Navigation works
}
```

#### IT4: ConfirmAndCancelSequenceIntegration
```go
func TestConfirmAndCancelSequenceIntegration(t *testing.T) {
    // First dialog: Confirm
    // Second dialog: Cancel
    // Verify both paths work correctly
}
```

#### IT5: NavigationAfterDialogConfirm
```go
func TestNavigationAfterDialogConfirm(t *testing.T) {
    // Setup: Create model with PermissionDialog
    // Action: Confirm dialog
    // Action: Press j, k, h, l keys
    // Verify: Each navigation key is processed correctly
    // Verify: Cursor moves as expected
}
```

## 7. Risk Assessment

### 7.1 Low Risk
- Change is minimal (adding `m.dialog = nil` in handler functions)
- Follows established pattern from previous fix
- Limited scope of impact

### 7.2 Considerations
- `handleRecursivePermissionComplete` conditionally displays error report dialog
  - Only clear dialog when NOT showing error report dialog
- `handleBatchPermissionComplete` already has dialog clearing logic
  - Verify it covers all edge cases

## 8. Acceptance Criteria

1. Permission change dialog confirmation does not cause freeze
2. Recursive permission dialog confirmation does not cause freeze
3. Batch permission change completion does not cause freeze
4. Esc key cancellation continues to work correctly
5. Navigation keys (j/k/h/l) work after dialog closes
6. Status message displays correctly after operation
7. All integration tests pass
8. Manual tests confirm post-dialog behavior

## 9. Timeline Estimate

- Implementation: 30 minutes
- Unit/Integration Tests: 1 hour
- Manual Testing: 30 minutes
- Code Review: 30 minutes
- Total: ~2.5 hours

## 10. References

- Previous fix commit: `c8fbfeb`
- Dialog Best Practices: `doc/development/DIALOG_BEST_PRACTICES.md`
- Previous bug fix spec: `doc/tasks/bugfix-permission-dialog-freeze/SPEC.md`
