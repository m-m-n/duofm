# Permission Dialog Confirm Freeze Bug Fix - Verification Plan

## 1. Build and Test Commands

```bash
# Build
make build

# Run all tests
make test

# Run specific package tests
go test -v ./internal/ui/...

# Run with race detection
go test -race ./internal/ui/...

# Run specific test file
go test -v ./internal/ui/ -run TestPermissionDialog

# Run integration tests only
go test -v ./internal/ui/ -run Integration
```

## 2. Success Criteria (from SPEC.md)

### Acceptance Criteria
1. Permission change dialog confirmation does not cause freeze
2. Recursive permission dialog confirmation does not cause freeze
3. Batch permission change completion does not cause freeze
4. Esc key cancellation continues to work correctly
5. Navigation keys (j/k/h/l) work after dialog closes
6. Status message displays correctly after operation
7. All integration tests pass
8. Manual tests confirm post-dialog behavior

## 3. Automated Verification Items

### 3.1 Unit Tests

#### UT1: PermissionDialog Enter Key Handling
**File**: `internal/ui/permission_dialog_test.go`

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestPermissionDialog_EnterWithValidInput` | Enter key with valid 3-digit input | Returns command from onConfirm callback |
| `TestPermissionDialog_EnterWithInvalidInput` | Enter key with invalid input (e.g., "999") | Shows error message, dialog remains open |
| `TestPermissionDialog_EnterWithIncompleteInput` | Enter key with 1-2 digit input | Shows validation error, dialog remains open |

#### UT2: RecursivePermDialog Enter Key Handling
**File**: `internal/ui/recursive_perm_dialog_test.go`

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestRecursivePermDialog_Step1Enter` | Enter on step 1 with valid input | Transitions to step 2 |
| `TestRecursivePermDialog_Step2Enter` | Enter on step 2 with valid input | Returns command from onConfirm callback |
| `TestRecursivePermDialog_Step1InvalidInput` | Enter on step 1 with invalid input | Shows error, stays on step 1 |
| `TestRecursivePermDialog_Step2InvalidInput` | Enter on step 2 with invalid input | Shows error, stays on step 2 |

### 3.2 Integration Tests

**File**: `internal/ui/dialog_integration_test.go`

#### IT1: PermissionDialogConfirmationIntegration
```go
func TestPermissionDialogConfirmationIntegration(t *testing.T) {
    // Setup: Create test model and dialog with onConfirm callback
    // Action: Enter valid permission (e.g., "644"), press Enter
    // Verify: permissionOperationCompleteMsg is generated
    // Verify: After processing message, m.dialog becomes nil
    // Verify: Subsequent key press (j) is processed normally
}
```

#### IT2: RecursivePermDialogConfirmationIntegration
```go
func TestRecursivePermDialogConfirmationIntegration(t *testing.T) {
    // Setup: Create test model with RecursivePermDialog
    // Action: Complete step 1 with Enter (e.g., "755")
    // Verify: Dialog transitions to step 2
    // Action: Complete step 2 with Enter (e.g., "644")
    // Verify: recursivePermissionCompleteMsg is generated
    // Verify: After processing, m.dialog becomes nil (or error dialog)
    // Verify: Subsequent key press (j) is processed normally
}
```

#### IT3: BatchPermissionConfirmationIntegration
```go
func TestBatchPermissionConfirmationIntegration(t *testing.T) {
    // Setup: Create test model with multiple marked files
    // Action: Open permission dialog, enter value, press Enter
    // Verify: batchPermissionCompleteMsg is processed
    // Verify: m.dialog becomes nil
    // Verify: Marks are cleared
    // Verify: Navigation works
}
```

#### IT4: ConfirmAndCancelSequenceIntegration
```go
func TestConfirmAndCancelSequenceIntegration(t *testing.T) {
    // First dialog: Confirm with Enter
    // Verify: Dialog closes, navigation works
    // Second dialog: Cancel with Esc
    // Verify: Dialog closes, navigation works
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

## 4. Manual Test Checklist

### 4.1 Permission Dialog Confirmation

- [ ] Single file permission change with `644` - dialog closes, file updated
- [ ] Single directory permission change with `755` - dialog closes, directory updated
- [ ] Invalid permission (e.g., `999`) shows error, dialog stays open
- [ ] Empty input and Enter shows validation error
- [ ] Partial input (1-2 digits) and Enter shows validation error

### 4.2 Post-Confirmation Keyboard Navigation

- [ ] After confirmation, `j` key moves cursor down
- [ ] After confirmation, `k` key moves cursor up
- [ ] After confirmation, `l` key enters directory (if directory selected)
- [ ] After confirmation, `h` key goes to parent directory
- [ ] After confirmation, Tab key switches panes
- [ ] After confirmation, `p` key opens new permission dialog

### 4.3 Recursive Permission Change

- [ ] Step 1: Enter directory permission, moves to step 2
- [ ] Step 2: Enter file permission, applies changes
- [ ] After completion, keyboard navigation works
- [ ] With errors, error report dialog is shown
- [ ] After dismissing error report, keyboard navigation works

### 4.4 Batch Permission Change

- [ ] Mark multiple files, open permission dialog
- [ ] Confirm - all files updated, marks cleared
- [ ] After completion, keyboard navigation works
- [ ] With 10+ files, progress dialog appears
- [ ] After progress complete, keyboard navigation works

### 4.5 Error Handling

- [ ] Permission error shows appropriate message
- [ ] After error dialog closes, keyboard navigation works
- [ ] Partial failure in batch shows error report
- [ ] After error report closes, keyboard navigation works

### 4.6 Regression: Esc Key Cancellation

- [ ] PermissionDialog Esc - dialog closes, navigation works
- [ ] RecursivePermDialog Esc on step 1 - dialog closes, navigation works
- [ ] RecursivePermDialog Esc on step 2 - dialog closes, navigation works
- [ ] Batch dialog Esc - dialog closes, navigation works

## 5. Coverage Targets

### Overall Coverage Goals
- **Line Coverage**: >= 80% for modified functions
- **Branch Coverage**: >= 75% for handler functions

### Specific Coverage Targets

| Function | Target Coverage |
|----------|-----------------|
| `handlePermissionOperationComplete` | >= 90% |
| `handleRecursivePermissionComplete` | >= 85% |
| `handleBatchPermissionComplete` | >= 85% |
| `PermissionDialog.Update` | >= 80% |
| `RecursivePermDialog.Update` | >= 80% |

### Coverage Commands

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./internal/ui/...

# View coverage in browser
go tool cover -html=coverage.out

# View coverage summary
go tool cover -func=coverage.out | grep -E "(handlePermission|Dialog)"
```

## 6. Test Data Requirements

### Test Files Setup
```bash
# Create test directory structure
mkdir -p /tmp/duofm-test/{dir1,dir2}
touch /tmp/duofm-test/file{1,2,3}.txt
chmod 644 /tmp/duofm-test/file*.txt
chmod 755 /tmp/duofm-test/dir*

# Create read-only file for error testing (optional)
touch /tmp/duofm-test/readonly.txt
chmod 444 /tmp/duofm-test/readonly.txt
chattr +i /tmp/duofm-test/readonly.txt  # Requires root
```

### Test Fixtures
- Regular files with default permissions (644)
- Directories with default permissions (755)
- Multiple files for batch testing
- Nested directories for recursive testing

## 7. Verification Report Template

```markdown
# Verification Report - bugfix-permission-confirm-freeze

## Date: YYYY-MM-DD
## Tester: [Name]

### Build Status
- [ ] `make build` - PASS/FAIL
- [ ] `make test` - PASS/FAIL

### Unit Tests
- [ ] UT1: PermissionDialog Enter Key - PASS/FAIL
- [ ] UT2: RecursivePermDialog Enter Key - PASS/FAIL

### Integration Tests
- [ ] IT1: PermissionDialogConfirmation - PASS/FAIL
- [ ] IT2: RecursivePermDialogConfirmation - PASS/FAIL
- [ ] IT3: BatchPermissionConfirmation - PASS/FAIL
- [ ] IT4: ConfirmAndCancelSequence - PASS/FAIL
- [ ] IT5: NavigationAfterDialogConfirm - PASS/FAIL

### Manual Tests
- [ ] All checklist items verified

### Coverage
- handlePermissionOperationComplete: XX%
- handleRecursivePermissionComplete: XX%
- handleBatchPermissionComplete: XX%

### Issues Found
- [None / List issues]

### Conclusion
- [ ] APPROVED for merge
- [ ] REQUIRES fixes
```
