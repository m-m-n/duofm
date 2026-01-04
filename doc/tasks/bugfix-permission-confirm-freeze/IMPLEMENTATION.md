# Permission Dialog Confirm Freeze Bug Fix - Implementation Plan

## 1. Implementation Phases

### Phase 1: Single Permission Dialog Fix
**Goal**: Fix the freeze when confirming single file/directory permission changes

**Implementation Steps**:
1. Add `m.dialog = nil` at the beginning of `handlePermissionOperationComplete`
2. Ensure dialog reference is cleared regardless of success or failure

**Estimated Effort**: Small

### Phase 2: Recursive Permission Dialog Fix
**Goal**: Fix the freeze when confirming recursive permission changes

**Implementation Steps**:
1. Add `m.dialog = nil` in `handleRecursivePermissionComplete` when operation completes without errors
2. Keep existing logic that sets `m.dialog = errorDialog` when showing error report

**Estimated Effort**: Small

### Phase 3: Batch Permission Dialog Fix
**Goal**: Fix the freeze when confirming batch permission changes (especially small batches)

**Implementation Steps**:
1. Modify `handleBatchPermissionComplete` to unconditionally clear `m.dialog` at the beginning
2. Remove the conditional check `if _, ok := m.dialog.(*PermissionProgressDialog); ok`
3. Keep the logic that sets `m.dialog = errorDialog` when showing error report

**Bug Found**: The current implementation only clears `PermissionProgressDialog`. For small batches (under ProgressThreshold), `PermissionDialog` remains in `m.dialog`, causing freeze.

**Estimated Effort**: Small

### Phase 4: Testing
**Goal**: Ensure all confirmation paths work correctly

**Implementation Steps**:
1. Add unit tests for Enter key handling in dialogs
2. Add integration tests for confirmation flows
3. Perform manual testing for post-dialog keyboard navigation

**Estimated Effort**: Medium

## 2. Components and Contracts

### 2.1 handlePermissionOperationComplete

**File**: `internal/ui/model_permission.go`

**Current Behavior**:
- Receives `permissionOperationCompleteMsg`
- Updates status message based on success/failure
- Refreshes active pane
- Returns status message clear command

**Required Change**:
- Clear `m.dialog` reference at the beginning of the handler
- Must clear dialog regardless of operation success or failure

**Contract**:
- Input: `permissionOperationCompleteMsg` with path, success, and error fields
- Output: Updated model with `m.dialog = nil` and appropriate status message
- Side Effect: Active pane refresh

### 2.2 handleRecursivePermissionComplete

**File**: `internal/ui/model_permission.go`

**Current Behavior**:
- Receives `recursivePermissionCompleteMsg`
- Refreshes active pane
- If errors exist: creates and sets `PermissionErrorReportDialog`
- If successful: sets status message

**Required Change**:
- Add `m.dialog = nil` before setting success status message (no-error path)
- Do NOT clear dialog when showing error report dialog (existing logic already handles this)

**Contract**:
- Input: `recursivePermissionCompleteMsg` with path, successCount, and errors
- Output:
  - If errors: `m.dialog` set to new `PermissionErrorReportDialog`
  - If no errors: `m.dialog = nil` with success status message
- Side Effect: Active pane refresh

### 2.3 handleBatchPermissionComplete

**File**: `internal/ui/model_permission.go`

**Current Behavior** (BUG):
- Only clears `PermissionProgressDialog` (type check)
- Does NOT clear `PermissionDialog` for small batches
- Clears marks
- Refreshes active pane
- If errors: shows error report dialog
- If successful: sets status message

**Required Change**:
- Unconditionally clear `m.dialog` at the beginning of the handler
- This handles both `PermissionDialog` (small batches) and `PermissionProgressDialog` (large batches)

**Contract**:
- Input: `batchPermissionCompleteMsg` with counts and errors
- Output:
  - If errors: `m.dialog` set to new `PermissionErrorReportDialog`
  - If no errors: `m.dialog = nil`
- Side Effect: Marks cleared, active pane refresh

## 3. Dependencies and Implementation Order

```
Phase 1: handlePermissionOperationComplete fix
    ↓
Phase 2: handleRecursivePermissionComplete fix
    ↓
Phase 3: handleBatchPermissionComplete fix
    ↓
Phase 4: Unit Tests
    ↓
Phase 5: Integration Tests
    ↓
Phase 6: Manual Testing
```

### Dependency Notes

1. **Phase 1 has no dependencies** - Can be implemented and tested independently

2. **Phase 2 depends on understanding Phase 1** - Same pattern but with conditional logic

3. **Phase 3 requires Phases 1-2** - To establish consistent patterns before applying the same fix

4. **Phase 4-6 require Phases 1-3** - All fixes must be in place before comprehensive testing

### File Dependencies

| File | Modifies | Depends On |
|------|----------|------------|
| `model_permission.go` | Handlers | `messages.go`, `permission_dialog.go` |
| `dialog_integration_test.go` | Tests | `model_permission.go`, `permission_dialog.go` |

## 4. Risk Mitigation

### Low-Risk Changes
- Adding `m.dialog = nil` follows established pattern from Esc key fix (commit c8fbfeb)
- Limited scope of changes minimizes regression risk

### Testing Strategy
1. Run existing dialog cancellation tests to verify no regression
2. Add new confirmation-specific tests
3. Manual testing to verify end-user behavior

### Rollback Plan
- Changes are isolated to handler functions
- Can revert individual handler changes if issues arise
- Existing cancel message handling remains unchanged
