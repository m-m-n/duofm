# Feature: Async Directory Load Pane Identification Bug Fix

## Overview

Fix a bug where the pane identification logic in `directoryLoadCompleteMsg` handler incorrectly identifies the target pane when both panes display the same path. This causes the right pane to become stuck in "Loading directory..." state when navigating to a directory that the left pane is already displaying.

## Domain Rules

- **DR-1**: Each async directory load operation MUST apply completion handling only to the pane that initiated it
- **DR-2**: Pane identification MUST be performed using an explicit pane identifier, not by path matching
- **DR-3**: Each pane's state MUST be managed independently, even when both panes display the same path

## Objectives

- Ensure correct pane identification during async directory load completion
- Guarantee proper operation when both panes display the same path
- Maintain backward compatibility with existing navigation behavior

## User Stories

### US-1: Right Pane Home Directory Navigation
As a user, when I navigate to the home directory in the right pane, I expect the directory to load and display correctly, regardless of what the left pane is currently displaying.

### US-2: Previous Directory Navigation
As a user, when I press `-` to return to the previous directory, I expect the operation to complete correctly in the pane where I initiated it, even if the previous directory is the same as the other pane's current directory.

## Functional Requirements

### FR-1: Pane Identifier in Directory Load Message
- FR-1.1: Add a `paneID` field of type `PanePosition` to `directoryLoadCompleteMsg`
- FR-1.2: The `paneID` field SHALL contain either `LeftPane` or `RightPane`

### FR-2: Pane Identifier in Async Load Initiation
- FR-2.1: All async directory load functions MUST accept or determine the originating pane
- FR-2.2: The completion message MUST include the correct pane identifier

### FR-3: Pane Identification Logic in Handler
- FR-3.1: The `directoryLoadCompleteMsg` handler MUST determine the target pane based on `paneID`, not path
- FR-3.2: When `paneID` is `LeftPane`, the handler MUST apply changes to the left pane
- FR-3.3: When `paneID` is `RightPane`, the handler MUST apply changes to the right pane

## Non-Functional Requirements

### NFR-1: Backward Compatibility
- NFR-1.1: Existing navigation functionality MUST NOT be altered
- NFR-1.2: Existing key bindings MUST NOT be changed

### NFR-2: Performance
- NFR-2.1: The pane identifier processing MUST NOT impact key response time (< 50ms)

## Interface Contract

### Input/Output Specification

#### directoryLoadCompleteMsg (Modified)
```go
type directoryLoadCompleteMsg struct {
    paneID        PanePosition   // NEW: LeftPane or RightPane
    panePath      string
    entries       []fs.FileEntry
    err           error
    attemptedPath string
}
```

### Preconditions
- The async load operation was initiated by one specific pane
- The pane identifier was correctly captured at initiation time

### Postconditions
- The `loading` flag is cleared only on the pane identified by `paneID`
- The entries are updated only on the pane identified by `paneID`
- The other pane's state remains unchanged

### State Transitions

```
Load Initiated:
  pane.loading = true
  pane.pendingPath = targetPath

Load Completed (Success):
  pane.loading = false
  pane.pendingPath = ""
  pane.entries = loadedEntries
  pane.cursor = 0
  pane.scrollOffset = 0

Load Completed (Error):
  pane.loading = false
  pane.restorePreviousPath()
  statusBar = errorMessage
```

### Error Conditions
- If an invalid pane identifier is received, the message SHOULD be ignored
- Error handling for directory access failures remains unchanged

## Dependencies

- Existing `PanePosition` type definition (LeftPane, RightPane)
- Existing async directory loading infrastructure
- Bubble Tea framework for message handling

## Test Scenarios

### Basic Functionality
- [ ] TS-1: Left pane async directory load completes successfully
- [ ] TS-2: Right pane async directory load completes successfully

### Same Path Scenarios
- [ ] TS-3: Both panes showing home directory; right pane navigates to parent then back to home
- [ ] TS-4: Both panes showing same directory; left pane navigates successfully
- [ ] TS-5: Right pane uses `~` to navigate home while left pane displays home
- [ ] TS-6: Right pane uses `-` to return to previous directory (same as left pane's current)

### Error Scenarios
- [ ] TS-7: Async navigation to non-existent directory applies error to correct pane
- [ ] TS-8: Async navigation to permission-denied directory applies error to correct pane

## Success Criteria

- [ ] SC-1: Following the reproduction steps, the right pane successfully displays the home directory
- [ ] SC-2: Both panes can navigate independently when displaying the same path
- [ ] SC-3: All existing unit tests pass
- [ ] SC-4: All new tests pass

## Constraints

- The API change requires updating all call sites of async load functions
- Pane information must be passed through the async callback chain
