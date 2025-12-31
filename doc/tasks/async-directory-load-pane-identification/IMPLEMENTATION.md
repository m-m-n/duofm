# Implementation Plan: Async Directory Load Pane Identification Bug Fix

## Overview

Modify the async directory load mechanism to include an explicit pane identifier, ensuring that `directoryLoadCompleteMsg` is applied to the correct pane regardless of whether both panes display the same path.

## Objectives

- Add `paneID` field to `directoryLoadCompleteMsg` to identify the target pane
- Update `LoadDirectoryAsync()` to accept pane identification
- Modify all async navigation functions to pass pane information
- Update the message handler to use pane identifier instead of path matching

## Prerequisites

- Understanding of Bubble Tea message passing mechanism
- Familiarity with current pane navigation implementation

## Architecture Overview

The fix involves three main changes:

1. **Message Structure**: Add `PanePosition` field to identify target pane
2. **Async Function Signature**: Include pane identifier in load command
3. **Handler Logic**: Use explicit identifier instead of path matching

## Implementation Phases

### Phase 1: Message Structure Modification

**Goal**: Add pane identifier to the directory load completion message

**Files to Modify**:
- `internal/ui/messages.go` - Add paneID field to directoryLoadCompleteMsg

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| directoryLoadCompleteMsg | Carry async load result with pane ID | panePath exists | paneID + panePath + entries/err |

**Implementation Steps**:
1. Add `paneID PanePosition` field to `directoryLoadCompleteMsg` struct
2. Place the new field as the first field for clarity

**Estimated Effort**: Small

---

### Phase 2: Async Load Function Update

**Goal**: Enable `LoadDirectoryAsync` to accept and propagate pane identifier

**Files to Modify**:
- `internal/ui/pane.go` - Update LoadDirectoryAsync signature

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| LoadDirectoryAsync | Start async load with pane ID | paneID, path, sortConfig | Returns tea.Cmd with correct paneID |

**Processing Flow**:
```
LoadDirectoryAsync(paneID, path, sortConfig)
  → ReadDirectory(path)
  → SortEntries(entries)
  → Return directoryLoadCompleteMsg{paneID, path, entries, err}
```

**Implementation Steps**:
1. Update `LoadDirectoryAsync` signature to accept `paneID PanePosition` as first parameter
2. Update the returned `directoryLoadCompleteMsg` to include `paneID`

**Dependencies**:
- Phase 1 must be completed

**Estimated Effort**: Small

---

### Phase 3: Navigation Function Updates

**Goal**: Update all async navigation functions to pass pane identifier

**Files to Modify**:
- `internal/ui/pane.go` - Update all async navigation methods
- `internal/ui/model.go` - Update call sites that invoke MoveToParentAsync

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Pane.paneID | Store pane's identity | Assigned at initialization | Returns correct PanePosition |
| EnterDirectoryAsync | Navigate into directory with pane ID | Pane has paneID | Calls LoadDirectoryAsync with paneID |
| MoveToParentAsync | Navigate to parent with pane ID | Pane has paneID | Calls LoadDirectoryAsync with paneID |
| NavigateToHomeAsync | Navigate to home with pane ID | Pane has paneID | Calls LoadDirectoryAsync with paneID |
| NavigateToPreviousAsync | Navigate to previous with pane ID | Pane has paneID | Calls LoadDirectoryAsync with paneID |
| ChangeDirectoryAsync | Navigate to path with pane ID | Pane has paneID | Calls LoadDirectoryAsync with paneID |

**Processing Flow**:
```
Pane.NavigateToHomeAsync()
  → Get home directory
  → Record previous path
  → Start loading state
  → Call LoadDirectoryAsync(p.paneID, home, sortConfig)
```

**Implementation Steps**:
1. Add `paneID PanePosition` field to `Pane` struct
2. Update `NewPane()` to accept and store pane identifier
3. Update `Model.Init()` to pass correct pane identifiers when creating panes
4. Update each async navigation method to pass `p.paneID` to `LoadDirectoryAsync`
5. Update `model.go` call sites that directly call `MoveToParentAsync()` for arrow key navigation

**Dependencies**:
- Phase 2 must be completed

**Estimated Effort**: Medium

---

### Phase 4: Handler Logic Update

**Goal**: Modify handler to use pane identifier for target selection

**Files to Modify**:
- `internal/ui/model.go` - Update directoryLoadCompleteMsg handler

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| directoryLoadCompleteMsg handler | Apply result to correct pane | msg.paneID valid | Target pane updated correctly |

**Processing Flow**:
```
Receive directoryLoadCompleteMsg
  → If paneID == LeftPane → targetPane = leftPane
  → If paneID == RightPane → targetPane = rightPane
  → Apply entries/error to targetPane
  → Clear loading state
```

**Implementation Steps**:
1. Replace path-based pane detection with paneID-based detection
2. Use `msg.paneID` to directly select target pane
3. Remove the current path/pendingPath matching logic

**Dependencies**:
- Phases 1-3 must be completed

**Testing**:
- Verify left pane navigation still works
- Verify right pane navigation works when both panes show same path
- Verify error handling applies to correct pane

**Estimated Effort**: Small

---

## File Structure

```
internal/ui/
├── messages.go     # Add paneID field to directoryLoadCompleteMsg
├── pane.go         # Add paneID to Pane, update async functions
├── model.go        # Update handler logic and pane initialization
└── model_test.go   # Add tests for same-path scenarios
```

## Testing Strategy

### Unit Tests

- Verify `LoadDirectoryAsync` returns correct `paneID` in message
- Verify each async navigation method passes correct `paneID`
- Verify handler applies changes to correct pane based on `paneID`

### Integration Tests

- Both panes at same path, right pane navigates out and back
- Both panes at same path, left pane navigates out and back
- Error during navigation applies to correct pane

### Manual Testing Checklist
- [ ] Start duofm in home directory (both panes at ~)
- [ ] Switch to right pane with `l`
- [ ] Move to parent with `l`
- [ ] Press `~` to return home - verify loading completes
- [ ] Press `-` to toggle back and forth
- [ ] Repeat test with left pane

## Dependencies

### External Libraries
- None required

### Internal Dependencies
- `PanePosition` type already exists in model.go (LeftPane, RightPane)
- All changes are internal refactoring

## Risk Assessment

### Technical Risks
- **Missed Call Site**: Some async load invocation might be missed
  - Mitigation: Use grep to find all `LoadDirectoryAsync` calls, update all
- **Test Coverage**: Existing tests may not cover same-path scenarios
  - Mitigation: Add specific test cases for the bug scenario

### Implementation Risks
- **Regression in Normal Navigation**: Changes might break normal navigation
  - Mitigation: Run all existing tests, perform manual smoke testing

## Performance Considerations

- Adding a single `PanePosition` field (int type) has negligible memory impact
- No additional processing required; just passing an identifier

## Security Considerations

- No security implications; this is purely internal state management

## Open Questions

None - the specification is clear and complete.

## References

- Specification: `doc/tasks/async-directory-load-pane-identification/SPEC.md`
- Bug Report: `tmp/BUG_RIGHT_PANE_LOADING.md`
- Related Feature: `doc/tasks/navigation-enhancement/SPEC.md`
