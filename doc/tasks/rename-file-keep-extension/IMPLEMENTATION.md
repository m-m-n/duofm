# Implementation Plan: Rename File Keep Extension

## Overview

Modify the existing `R` keybinding to preserve file extensions during rename operations and introduce `Shift+R` for full filename editing. When renaming files with extensions, users edit only the base name while the extension remains fixed and visible.

## Objectives

- Improve user efficiency by preserving file extensions during rename operations
- Maintain backward compatibility by providing full filename editing via `Shift+R`
- Handle edge cases appropriately (extensionless files, hidden files, directories)
- Provide clear visual feedback showing which part of the filename is editable

## Prerequisites

### Development Environment
- Go 1.21 or later
- Make (for build automation)

### Dependencies
- Existing Bubble Tea framework (`github.com/charmbracelet/bubbletea`)
- Existing Lip Gloss styling library (`github.com/charmbracelet/lipgloss`)
- Existing dialog infrastructure (`BaseDialog`, `DialogStyles`, `TextInput`)

### Knowledge Requirements
- Bubble Tea message/update architecture
- Existing dialog implementation patterns in duofm
- Action/Keybinding system in duofm

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea
- **Key Libraries**:
  - `bubbletea` - TUI framework
  - `lipgloss` - Terminal styling

### Design Approach

Extend the existing dialog and action system with minimal changes:
1. Add new `ActionRenameFullName` action for `Shift+R`
2. Create new `ExtensionRenameDialog` for extension-preserving rename
3. Modify `handleRenameUI()` to detect file type and choose appropriate dialog
4. Add new `handleRenameFullNameUI()` handler for `Shift+R`

### Component Interaction

```
User Input (R / Shift+R)
    |
    v
keybindingMap.GetAction()
    |
    +---> ActionRename -------> handleRenameUI()
    |                               |
    |                               +---> hasEditableExtension()
    |                               |         |
    |                               |         +---> [has ext] --> ExtensionRenameDialog
    |                               |         |
    |                               |         +---> [no ext]  --> InputDialog (full name)
    |                               |
    +---> ActionRenameFullName --> handleRenameFullNameUI()
                                        |
                                        +---> InputDialog (full name)
```

## Implementation Phases

### Phase 1: Core Components

**Goal**: Implement extension detection logic and new action constant

**Files to Create**:
- `internal/ui/extension_util.go` - Extension detection utility function

**Files to Modify**:
- `internal/ui/actions.go` - Add `ActionRenameFullName` constant and mappings
- `internal/config/defaults.go` - Add `rename_full_name` keybinding

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| `hasEditableExtension()` | Determine if file should use extension-preserving mode | Valid filename and isDir flag | Returns (baseName, extension, hasExt) |
| `ActionRenameFullName` | Action constant for full filename rename | None | Action registered in action maps |

**Processing Flow** (hasEditableExtension):
```
1. Receive filename and isDir flag
   +-- If isDir --> return (name, "", false)
2. Check if hidden file (starts with ".")
   +-- If hidden:
   |     +-- Remove leading dot, check for extension in remainder
   |     +-- If extension found:
   |     |     +-- If ext == "." --> return (name, "", false) [trailing dot]
   |     |     +-- Else --> return ("." + base, ext, true)
   |     +-- If no extension --> return (name, "", false)
   +-- If not hidden:
         +-- Check for extension using last dot
         +-- If extension found:
         |     +-- If ext == "." --> return (name, "", false) [trailing dot]
         |     +-- Else --> return (base, ext, true)
         +-- If no extension --> return (name, "", false)
```

**Note**: Files ending with a trailing dot (e.g., `file.`) are treated as extensionless because the dot does not represent a meaningful extension.

**Implementation Steps**:

1. **Add ActionRenameFullName constant**
   - Add to Action enum in actions.go
   - Add to actionNames and nameToAction maps

2. **Add keybinding configuration**
   - Add "rename_full_name": {"Shift+R"} to DefaultKeybindings()
   - Add "rename_full_name" to AllActions()

3. **Implement hasEditableExtension function**
   - Create extension_util.go with the detection logic
   - Handle directories, hidden files, and regular files

**Dependencies**:
- Requires: None (foundational phase)
- Blocks: Phase 2 (ExtensionRenameDialog), Phase 3 (handler integration)

**Testing Approach**:

*Unit Tests*:
- Test hasEditableExtension with various file types:
  - Regular files with extensions (document.txt, archive.tar.gz)
  - Extensionless files (Makefile, LICENSE)
  - Hidden files without extension (.bashrc, .gitignore)
  - Hidden files with extension (.config.json, .foo.bar)
  - Directories
  - Edge cases (.txt, file., file..txt)

**Acceptance Criteria**:
- [ ] ActionRenameFullName is recognized by ActionFromName()
- [ ] "rename_full_name" is in DefaultKeybindings() with "Shift+R"
- [ ] hasEditableExtension correctly classifies all test cases
- [ ] Build succeeds with no errors

**Estimated Effort**: Small (1-2 days)

---

### Phase 2: Extension Rename Dialog

**Goal**: Create new dialog component for extension-preserving rename

**Files to Create**:
- `internal/ui/extension_rename_dialog.go` - New dialog component
- `internal/ui/extension_rename_dialog_test.go` - Unit tests

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| `ExtensionRenameDialog` | Dialog for extension-preserving rename | Valid directory path, original name, base name, extension | User can edit base name, extension fixed |
| `extensionRenameResultMsg` | Message to communicate dialog result | Dialog closed | Contains newName, oldName, dirPath, cancelled |
| `NewExtensionRenameDialog()` | Constructor for the dialog | Valid parameters | Initialized dialog with validation ready |

**Processing Flow** (ExtensionRenameDialog.Update):
```
1. Receive message (tea.Msg)
   +-- If tea.KeyMsg:
   |     +-- If Enter:
   |     |     +-- If hasError --> do nothing
   |     |     +-- Else --> close dialog, return result message
   |     +-- If Esc:
   |     |     +-- Close dialog, return cancelled message
   |     +-- Otherwise:
   |           +-- Delegate to TextInput
   |           +-- If input changed --> validate input
   +-- If extensionRenameResultMsg:
   |     +-- Route to handleExtensionRenameResult
   +-- Otherwise:
         +-- Return unchanged model (ignore other messages)
```

**Implementation Steps**:

1. **Define dialog struct**
   - Embed BaseDialog
   - Add fields: title, textInput, extension, dirPath, originalName, existingFiles, hasError, errorMessage, styles

2. **Define extensionRenameResultMsg**
   - Fields: newName (full name with extension), oldName, dirPath, cancelled

3. **Implement NewExtensionRenameDialog constructor**
   - Load existing files from directory for duplicate check
   - Initialize TextInput with base name
   - Set up dialog styles

4. **Implement Update method**
   - Handle Enter (validate and confirm), Esc (cancel), text input
   - Real-time validation on input change

5. **Implement View method**
   - Render title with extension info
   - Render input field with fixed extension display
   - Render error message if validation fails
   - Render footer with keybinding hints

6. **Implement validation**
   - Empty input check
   - Duplicate filename check (base + extension)
   - Invalid character check (reuse fs.ValidateFilename)

**UI Layout**:
```
+------------------------------------------------+
| Rename (extension: .txt):                       |
|                                                 |
| +--------------------------------+ .txt         |
| |document                        |              |
| +--------------------------------+              |
|                                                 |
| Enter: Confirm  Esc: Cancel                     |
+------------------------------------------------+
```

**UI Width Handling**:
- Dialog width is dynamically calculated based on terminal width
- Maximum dialog width: min(terminal width - 4, 60 characters)
- Input field width: dialog width - padding - extension display width
- If filename + extension exceeds input field width:
  - Input field scrolls horizontally (built-in TextInput behavior)
  - Extension remains visible at fixed position
- Very long extensions (>15 chars): Consider truncation with "..." display
- Minimum usable input width: 20 characters (if not achievable, use full-width layout)

**Dependencies**:
- Requires: Phase 1 (hasEditableExtension for understanding the design)
- Blocks: Phase 3 (handler integration)

**Testing Approach**:

*Unit Tests*:
- Dialog initialization with correct values
- Input field contains base name only
- Extension is displayed but not part of input
- Enter generates correct full filename (base + extension)
- Esc cancels dialog
- Empty input shows error, blocks Enter
- Duplicate filename shows error, blocks Enter
- Invalid characters show error

**Acceptance Criteria**:
- [ ] Dialog displays extension separately from input
- [ ] Base name is pre-filled and editable
- [ ] Validation works for empty, duplicate, invalid inputs
- [ ] Result message contains correct full filename
- [ ] Dialog follows existing style conventions

**Estimated Effort**: Medium (2-3 days)

---

### Phase 3: Handler Integration

**Goal**: Connect new components to existing action handling system

**Files to Modify**:
- `internal/ui/model_update_keyboard.go` - Modify handleRenameUI, add handleRenameFullNameUI
- `internal/ui/model_update.go` - Add handleExtensionRenameResult
- `internal/ui/messages.go` - Import extensionRenameResultMsg handling if needed

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| `handleRenameUI()` (modified) | Choose appropriate dialog based on file type | Selected entry is not nil and not parent dir | Correct dialog is opened |
| `handleRenameFullNameUI()` (new) | Always open full filename dialog | Selected entry is not nil and not parent dir | InputDialog opened with full name |
| `handleExtensionRenameResult()` | Process extension rename dialog result | extensionRenameResultMsg received | File renamed, pane refreshed |

**Processing Flow** (handleRenameUI):
```
1. Get selected entry from active pane
   +-- If nil or parent dir --> return (no action)
2. Call hasEditableExtension(entry.Name, entry.IsDir)
   +-- If hasExt:
   |     +-- Create ExtensionRenameDialog with baseName, ext
   +-- If not hasExt:
         +-- Create InputDialog with full name (same as Shift+R)
3. Set dialog to model
```

**Processing Flow** (handleExtensionRenameResult):
```
1. Clear dialog from model
2. Check if cancelled
   +-- If cancelled --> return (no action)
3. Validate the new name
4. Execute rename operation (fs.Rename)
   +-- If error --> show error message
5. Refresh active pane
```

**Implementation Steps**:

1. **Add ActionRenameFullName case to handleAction**
   - Route to handleRenameFullNameUI

2. **Modify handleRenameUI**
   - Call hasEditableExtension to determine mode
   - Open appropriate dialog based on result

3. **Implement handleRenameFullNameUI**
   - Always open InputDialog with full filename
   - Reuse existing InputDialog with SetInput

4. **Add extensionRenameResultMsg handling in Update**
   - Add case in message switch
   - Route to handleExtensionRenameResult

5. **Implement handleExtensionRenameResult**
   - Similar to existing rename result handling
   - Execute rename and refresh pane

**Dependencies**:
- Requires: Phase 1 (ActionRenameFullName, hasEditableExtension), Phase 2 (ExtensionRenameDialog)
- Blocks: Phase 4 (help dialog update)

**Testing Approach**:

*Unit Tests*:
- R on .txt file opens ExtensionRenameDialog
- R on Makefile opens InputDialog
- R on .bashrc opens InputDialog (no ext after removing leading dot)
- R on .config.json opens ExtensionRenameDialog
- R on directory opens InputDialog
- Shift+R on any file opens InputDialog with full name
- Rename result handling works correctly

*Integration Tests*:
- Complete rename flow: R -> type name -> Enter -> file renamed
- Rename with error -> correction -> success
- Rename cancel with Esc

**Acceptance Criteria**:
- [ ] R key triggers correct dialog based on file type
- [ ] Shift+R always triggers full filename dialog
- [ ] Rename operation completes successfully
- [ ] Pane refreshes after rename
- [ ] Error handling works correctly

**Estimated Effort**: Medium (2-3 days)

---

### Phase 4: Help Dialog and Polish

**Goal**: Update help dialog and ensure all edge cases are handled

**Files to Modify**:
- `internal/ui/help_dialog.go` - Update keybinding display

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| `HelpDialog.buildContent()` (modified) | Display updated keybindings | None | Help shows both R and Shift+R |

**Implementation Steps**:

1. **Update help dialog content**
   - Change R description to indicate extension-preserving behavior
   - Add Shift+R for full filename rename

2. **Review and test edge cases**
   - File with only extension (.txt)
   - File with multiple consecutive dots (file..txt)
   - File ending with dot (file.)
   - Very long filenames
   - Unicode characters

**Help Dialog Update**:
```
File Operations
  ...
  R              : rename file/directory (preserve extension)
  Shift+R        : rename file/directory (full name)
  ...
```

**Dependencies**:
- Requires: Phase 3 (all functionality working)
- Blocks: None (final phase)

**Testing Approach**:

*Manual Testing*:
- Verify help dialog shows updated keybindings
- Test all edge case scenarios
- Verify performance (dialog display < 50ms)

**Acceptance Criteria**:
- [ ] Help dialog displays both R and Shift+R keybindings
- [ ] All edge cases handled correctly
- [ ] No regression in existing functionality

**Estimated Effort**: Small (1 day)

---

## Complete File Structure

```
internal/
+-- ui/
|   +-- actions.go                      # Modified: Add ActionRenameFullName
|   +-- extension_util.go               # New: hasEditableExtension function
|   +-- extension_util_test.go          # New: Tests for extension detection
|   +-- extension_rename_dialog.go      # New: Extension-preserving dialog
|   +-- extension_rename_dialog_test.go # New: Dialog tests
|   +-- model_update_keyboard.go        # Modified: handleRenameUI, handleRenameFullNameUI
|   +-- model_update.go                 # Modified: handleExtensionRenameResult
|   +-- help_dialog.go                  # Modified: Update keybinding display
|   +-- messages.go                     # No change (or add extensionRenameResultMsg)
|   +-- input_dialog.go                 # No change (reused)
|   +-- dialog_base.go                  # No change (reused)
+-- config/
|   +-- defaults.go                     # Modified: Add rename_full_name
|   +-- defaults_test.go                # Modified: Add test for new keybinding
+-- fs/
    +-- operations.go                   # No change (Rename exists)
```

## Testing Strategy

### Unit Testing

**Approach**:
- Table-driven tests for extension detection
- Dialog state tests for ExtensionRenameDialog
- Mock-based tests for handler functions

**Test Coverage Goals**:
- Extension detection: 100% coverage
- Dialog component: 80%+ coverage
- Handler integration: 70%+ coverage

**Key Test Areas**:

1. **Extension Detection** (`internal/ui/extension_util_test.go`)
   - Regular files: document.txt, archive.tar.gz
   - Extensionless files: Makefile, LICENSE
   - Hidden files: .bashrc, .config.json, .foo.bar
   - Directories
   - Edge cases: .txt, file., file..txt

2. **ExtensionRenameDialog** (`internal/ui/extension_rename_dialog_test.go`)
   - Initialization with correct values
   - Input handling
   - Validation (empty, duplicate, invalid)
   - Result message generation

3. **Handler Integration** (`internal/ui/model_update_keyboard_test.go`)
   - Correct dialog selection based on file type
   - Action routing

### Integration Testing

**Scenarios**:
1. R on file with extension -> ExtensionRenameDialog -> rename succeeds
2. R on extensionless file -> InputDialog -> rename succeeds
3. Shift+R on any file -> InputDialog -> rename succeeds
4. Validation error -> correction -> success
5. Cancel with Esc

### Manual Testing Checklist

From SPEC.md test scenarios:
- [ ] `document.txt` -> base: `document`, ext: `.txt`
- [ ] `archive.tar.gz` -> base: `archive.tar`, ext: `.gz`
- [ ] `Makefile` -> full edit mode
- [ ] `.bashrc` -> full edit mode
- [ ] `.config.json` -> base: `.config`, ext: `.json`
- [ ] `.foo.bar` -> base: `.foo`, ext: `.bar`
- [ ] Directory -> full edit mode
- [ ] Shift+R on any -> full edit mode
- [ ] Empty input -> error
- [ ] Duplicate name -> error

## Dependencies

### External Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/charmbracelet/bubbletea | (existing) | TUI framework |
| github.com/charmbracelet/lipgloss | (existing) | Terminal styling |

### Internal Dependencies

**Implementation Order** (respecting dependencies):
1. Phase 1 (actions, keybindings, extension util) - no dependencies
2. Phase 2 (ExtensionRenameDialog) - depends on Phase 1 patterns
3. Phase 3 (handler integration) - depends on Phases 1 and 2
4. Phase 4 (help dialog, polish) - depends on Phase 3

**Component Dependencies**:
- `extension_rename_dialog.go` depends on `dialog_base.go`, `extension_util.go`
- `model_update_keyboard.go` depends on `extension_util.go`, `extension_rename_dialog.go`
- `help_dialog.go` has no new dependencies

## Risk Assessment

### Technical Risks

1. **Extension Detection Edge Cases**
   - **Risk**: Unusual filename patterns may not be handled correctly
   - **Likelihood**: Medium
   - **Impact**: Low (fallback to full edit mode)
   - **Mitigation**: Comprehensive unit tests for edge cases

2. **Dialog Interaction Consistency**
   - **Risk**: New dialog may not match existing dialog patterns
   - **Likelihood**: Low
   - **Impact**: Medium (UX inconsistency)
   - **Mitigation**: Follow existing InputDialog and RenameInputDialog patterns

### Implementation Risks

1. **Keybinding Conflicts**
   - **Risk**: Shift+R may conflict with user customizations
   - **Likelihood**: Low
   - **Impact**: Low (configurable)
   - **Mitigation**: Document in help dialog, respect user config

## Performance Considerations

1. **Dialog Display**
   - Target: < 50ms display latency
   - Existing dialogs meet this target
   - No additional performance concerns expected

2. **File Listing for Validation**
   - Load existing files once on dialog open
   - Cache in dialog struct for duplicate checking

## Security Considerations

1. **Input Validation**
   - Reuse existing fs.ValidateFilename
   - Prevent path separator in filename
   - Block null characters

2. **Path Handling**
   - Use filepath.Join for safe path construction
   - No directory traversal possible (rename within same directory)

## Open Questions

### From Specification
- None (all requirements confirmed in SPEC.md)

### Implementation-Specific
- None (design follows existing patterns)

## Success Metrics

### Functional Completeness
- [ ] All functional requirements (FR1-FR8) implemented
- [ ] All user story acceptance criteria met
- [ ] All test scenarios pass

### Quality Metrics
- [ ] Test coverage: 80%+ for new code
- [ ] No critical bugs in manual testing
- [ ] Code follows Go best practices

### Performance Metrics
- [ ] Dialog display < 50ms
- [ ] No UI lag during input

### User Experience
- [ ] Clear visual separation between editable and fixed parts
- [ ] Intuitive behavior based on file type
- [ ] Consistent with existing rename functionality

## References

- **Specification**: `doc/tasks/rename-file-keep-extension/SPEC.md`
- **Requirements**: `doc/tasks/rename-file-keep-extension/要件定義書.md`
- **Existing Input Dialog**: `internal/ui/input_dialog.go`
- **Existing Rename Dialog**: `internal/ui/rename_input_dialog.go`
- **Action Definitions**: `internal/ui/actions.go`
- **Keybinding Configuration**: `internal/config/defaults.go`

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - Verify plan aligns with specification
   - Confirm phase breakdown is appropriate

2. **Begin Implementation**
   - Start with Phase 1 (foundational components)
   - Follow TDD approach where appropriate
   - Commit incrementally after each phase

3. **Verification**
   - Run verification checks after each phase
   - Execute `/sdd.6-verify` after completion
