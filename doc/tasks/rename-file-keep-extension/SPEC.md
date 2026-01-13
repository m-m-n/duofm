# Feature: Rename File Keep Extension

## Overview

This feature modifies the existing `R` keybinding behavior to preserve file extensions during rename operations and introduces `Shift+R` for full filename editing. When renaming files with extensions, users will edit only the base name while the extension remains fixed and visible, reducing keystrokes and preventing accidental extension changes.

## Objectives

- Improve user efficiency by preserving file extensions during rename operations
- Maintain backward compatibility by providing full filename editing via `Shift+R`
- Handle edge cases appropriately (extensionless files, hidden files, directories)
- Provide clear visual feedback showing which part of the filename is editable

## User Stories

### US1: Extension-Preserving Rename
As a user, I want to rename files while keeping their extensions, so that I can quickly change filenames without re-typing or accidentally modifying extensions.

**Acceptance Criteria:**
- [ ] Pressing `R` on a file with extension shows a dialog with fixed extension display
- [ ] The input field contains only the base name (without extension)
- [ ] The extension is displayed to the right of the input field (not editable)
- [ ] Pressing `Enter` renames the file with the new base name + original extension

### US2: Full Filename Rename
As a user, I want to edit the entire filename including extension when needed, so that I can change file types or handle special cases.

**Acceptance Criteria:**
- [ ] Pressing `Shift+R` on any file/directory shows a dialog for full filename editing
- [ ] The input field contains the complete filename (including extension)
- [ ] This works consistently for all file types

### US3: Extensionless File Handling
As a user, I want extensionless files like `Makefile` to be renamed using full filename editing, so that the rename behavior is intuitive.

**Acceptance Criteria:**
- [ ] Pressing `R` on `Makefile`, `LICENSE`, etc. shows full filename dialog
- [ ] The behavior is identical to `Shift+R`

### US4: Hidden File Handling
As a user, I want hidden files to be handled by applying extension detection rules to the part after the leading dot.

**Acceptance Criteria:**
- [ ] `.bashrc` (leading dot removed → `bashrc`, no extension) uses full filename editing with `R`
- [ ] `.config.json` (leading dot removed → `config.json`, extension `.json`) uses extension-preserving mode with `R`
- [ ] `.foo.bar` (leading dot removed → `foo.bar`, extension `.bar`) uses extension-preserving mode with `R`
- [ ] Leading dot is always preserved in all rename operations
- [ ] `Shift+R` always uses full filename editing

### US5: Directory Handling
As a user, I want directories to always use full name editing since they don't have extensions.

**Acceptance Criteria:**
- [ ] Pressing `R` on a directory shows full filename dialog
- [ ] Pressing `Shift+R` on a directory shows full filename dialog (same behavior)

## Technical Requirements

### Functional Requirements
- **FR1:** `R` key triggers extension-preserving rename for files with extensions
- **FR2:** `R` key triggers full rename for extensionless files, hidden files (without extension), and directories
- **FR3:** `Shift+R` key triggers full rename for all file types
- **FR4:** Extension detection uses the last dot as separator (e.g., `archive.tar.gz` -> ext: `.gz`)
- **FR5:** Hidden files are identified by names starting with `.`
- **FR6:** For hidden files, extension detection is applied to the part after removing the leading dot
- **FR6.1:** Leading dot is always preserved during rename
- **FR6.2:** `.bashrc` (after removing leading dot: `bashrc`) has no extension -> full edit mode
- **FR6.3:** `.config.json` (after removing leading dot: `config.json`) has extension `.json` -> editable part is `.config`
- **FR6.4:** `.foo.bar` (after removing leading dot: `foo.bar`) has extension `.bar` -> editable part is `.foo`
- **FR7:** Validation includes empty input check and duplicate filename check
- **FR8:** Help dialog displays both keybindings (`R` and `Shift+R`)

### Non-Functional Requirements
- **NFR1 - Performance:** Dialog display latency < 50ms
- **NFR2 - Usability:** Clear visual separation between editable and fixed parts
- **NFR3 - Maintainability:** Reuse existing dialog framework (`BaseDialog`)

## Implementation Approach

### Architecture

**Component Hierarchy:**
```
┌────────────────────────────────────────────────────────────┐
│                         Model                              │
├────────────────────────────────────────────────────────────┤
│  handleKeyInput() → keybindingMap.GetAction()              │
│                   ↓                                        │
│  handleAction() → ActionRename / ActionRenameFullName      │
│                   ↓                                        │
│  handleRenameUI() / handleRenameFullNameUI()               │
│                   ↓                                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         ExtensionRenameDialog (new)                  │  │
│  │         or InputDialog (existing)                    │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────┘
```

**New Components:**
- `ExtensionRenameDialog`: New dialog for extension-preserving rename
- `ActionRenameFullName`: New action for full filename rename

### Data Flow

```
User presses R/Shift+R
    ↓
handleKeyInput(msg)
    ↓
keybindingMap.GetAction(key)
    ↓
handleAction(action)
    ↓
[ActionRename]                    [ActionRenameFullName]
    ↓                                     ↓
handleRenameUI()                  handleRenameFullNameUI()
    ↓                                     ↓
determineRenameMode(entry)               always
    ↓                                     ↓
[has extension]    [no extension]   InputDialog (full name)
    ↓                    ↓
ExtensionRename    InputDialog
Dialog             (full name)
    ↓                    ↓
User input + Enter
    ↓
extensionRenameResultMsg / inputDialogResultMsg
    ↓
handleExtensionRenameResult() / handleInputDialogResult()
    ↓
fs.Rename(oldPath, newPath)
    ↓
RefreshPane()
```

### File Extension Detection Logic

```go
// hasEditableExtension determines if a file should use extension-preserving mode
func hasEditableExtension(name string, isDir bool) (baseName, extension string, hasExt bool) {
    // Directories never have editable extensions
    if isDir {
        return name, "", false
    }

    // Check if it's a hidden file (starts with .)
    isHidden := strings.HasPrefix(name, ".")

    if isHidden {
        // For hidden files, remove the leading dot and check the remainder
        // .bashrc -> bashrc (no dot) -> no extension -> full edit
        // .config.json -> config.json (has dot) -> extension .json -> editable: .config
        // .foo.bar -> foo.bar (has dot) -> extension .bar -> editable: .foo
        nameWithoutLeadingDot := name[1:]

        ext := filepath.Ext(nameWithoutLeadingDot)
        if ext == "" || ext == nameWithoutLeadingDot {
            // No extension in the part after leading dot
            return name, "", false
        }

        // Has extension: construct editable base (with leading dot) and extension
        baseWithoutLeadingDot := strings.TrimSuffix(nameWithoutLeadingDot, ext)
        return "." + baseWithoutLeadingDot, ext, true
    }

    // Regular file - check for extension
    ext := filepath.Ext(name)
    if ext == "" || ext == name {
        // No extension or name is just ".something"
        return name, "", false
    }

    base := strings.TrimSuffix(name, ext)
    return base, ext, true
}
```

### New Dialog: ExtensionRenameDialog

**File:** `internal/ui/extension_rename_dialog.go`

```go
// ExtensionRenameDialog is a rename dialog that preserves file extension
type ExtensionRenameDialog struct {
    BaseDialog
    title         string
    textInput     *TextInput
    extension     string          // Fixed extension (e.g., ".txt")
    dirPath       string          // Directory containing the file
    originalName  string          // Original filename
    existingFiles map[string]bool // Cached filenames for duplicate check
    hasError      bool
    errorMessage  string
    styles        DialogStyles
}

// extensionRenameResultMsg is sent when the dialog is confirmed or cancelled
type extensionRenameResultMsg struct {
    newName   string // Full new name (base + extension)
    oldName   string // Original filename
    dirPath   string // Directory path
    cancelled bool
}
```

**Dialog View:**
```
┌─────────────────────────────────────────────────┐
│ Rename (extension: .txt):                       │
│                                                 │
│ ┌───────────────────────────────────┐ .txt     │
│ │document                           │           │
│ └───────────────────────────────────┘           │
│                                                 │
│ Enter: Confirm  Esc: Cancel                     │
└─────────────────────────────────────────────────┘
```

### Action and Keybinding Changes

**File:** `internal/ui/actions.go`

Add new action:
```go
const (
    // ... existing actions ...
    ActionRenameFullName  // New: Full filename rename (Shift+R)
)

var actionNames = map[Action]string{
    // ... existing mappings ...
    ActionRenameFullName: "rename_full_name",
}

var nameToAction = map[string]Action{
    // ... existing mappings ...
    "rename_full_name": ActionRenameFullName,
}
```

**File:** `internal/config/defaults.go`

Update default keybindings:
```go
func DefaultKeybindings() map[string][]string {
    return map[string][]string{
        // ... existing bindings ...
        "rename":           {"R"},        // Extension-preserving (contextual)
        "rename_full_name": {"Shift+R"},  // Full filename edit
        // ... rest of bindings ...
    }
}

func AllActions() []string {
    return []string{
        // ... existing actions ...
        "rename_full_name",
        // ... rest of actions ...
    }
}
```

### Handler Implementation

**File:** `internal/ui/model_update_keyboard.go`

```go
func (m Model) handleAction(action Action) (tea.Model, tea.Cmd) {
    switch action {
    // ... existing cases ...

    case ActionRename:
        return m.handleRenameUI()

    case ActionRenameFullName:
        return m.handleRenameFullNameUI()

    // ... rest of cases ...
    }
}

// handleRenameUI shows contextual rename dialog (extension-preserving or full)
func (m Model) handleRenameUI() (tea.Model, tea.Cmd) {
    entry := m.getActivePane().SelectedEntry()
    if entry == nil || entry.IsParentDir() {
        return m, nil
    }

    pane := m.getActivePane()
    baseName, ext, hasExt := hasEditableExtension(entry.Name, entry.IsDir)

    if hasExt {
        // Extension-preserving mode
        m.dialog = NewExtensionRenameDialog(pane.Path(), entry.Name, baseName, ext)
    } else {
        // Full filename mode (same as Shift+R)
        m.dialog = NewInputDialog("Rename to:", func(newName string) tea.Cmd {
            return m.handleRename(pane.Path(), entry.Name, newName)
        })
        m.dialog.(*InputDialog).SetInput(entry.Name)
    }
    return m, nil
}

// handleRenameFullNameUI shows full filename rename dialog
func (m Model) handleRenameFullNameUI() (tea.Model, tea.Cmd) {
    entry := m.getActivePane().SelectedEntry()
    if entry == nil || entry.IsParentDir() {
        return m, nil
    }

    pane := m.getActivePane()
    oldName := entry.Name
    m.dialog = NewInputDialog("Rename to:", func(newName string) tea.Cmd {
        return m.handleRename(pane.Path(), oldName, newName)
    })
    m.dialog.(*InputDialog).SetInput(oldName)
    return m, nil
}
```

### File Structure

```
internal/
├── ui/
│   ├── extension_rename_dialog.go      # New: Extension-preserving dialog
│   ├── extension_rename_dialog_test.go # New: Tests for dialog
│   ├── actions.go                      # Modified: Add ActionRenameFullName
│   ├── model_update_keyboard.go        # Modified: Add handlers
│   ├── keybinding_map.go               # No change (uses config)
│   ├── help_dialog.go                  # Modified: Update keybinding display
│   └── input_dialog.go                 # No change (reused for full rename)
├── config/
│   ├── defaults.go                     # Modified: Add rename_full_name
│   └── defaults_test.go                # Modified: Add tests
└── fs/
    └── operations.go                   # No change (Rename function exists)
```

## Test Scenarios

### Unit Tests

#### Extension Detection Tests
- [ ] `document.txt` -> base: `document`, ext: `.txt`, hasExt: true
- [ ] `archive.tar.gz` -> base: `archive.tar`, ext: `.gz`, hasExt: true
- [ ] `Makefile` -> base: `Makefile`, ext: ``, hasExt: false
- [ ] `LICENSE` -> base: `LICENSE`, ext: ``, hasExt: false
- [ ] `.bashrc` (leading dot removed: `bashrc`, no dot) -> base: `.bashrc`, ext: ``, hasExt: false
- [ ] `.gitignore` (leading dot removed: `gitignore`, no dot) -> base: `.gitignore`, ext: ``, hasExt: false
- [ ] `.config.json` (leading dot removed: `config.json`, has dot) -> base: `.config`, ext: `.json`, hasExt: true
- [ ] `.env.local` (leading dot removed: `env.local`, has dot) -> base: `.env`, ext: `.local`, hasExt: true
- [ ] `.foo.bar` (leading dot removed: `foo.bar`, has dot) -> base: `.foo`, ext: `.bar`, hasExt: true
- [ ] Directory `src` -> base: `src`, ext: ``, hasExt: false
- [ ] Directory `node_modules` -> base: `node_modules`, ext: ``, hasExt: false

#### ExtensionRenameDialog Tests
- [ ] Test dialog initialization with correct base name and extension
- [ ] Test input field contains base name only
- [ ] Test extension is displayed but not editable
- [ ] Test Enter key generates correct full filename
- [ ] Test Esc key cancels dialog
- [ ] Test empty input validation
- [ ] Test duplicate filename validation
- [ ] Test invalid character validation

#### Action Handler Tests
- [ ] Test `R` on `.txt` file opens ExtensionRenameDialog
- [ ] Test `R` on `Makefile` opens InputDialog
- [ ] Test `R` on `.bashrc` opens InputDialog
- [ ] Test `R` on `.config.json` opens ExtensionRenameDialog
- [ ] Test `R` on directory opens InputDialog
- [ ] Test `Shift+R` on `.txt` file opens InputDialog
- [ ] Test `Shift+R` on directory opens InputDialog

### Integration Tests
- [ ] Test complete rename flow: R -> type new name -> Enter
- [ ] Test rename with validation error then correction
- [ ] Test rename cancel with Esc
- [ ] Test keybinding configuration override

### E2E Tests
- [ ] Scenario 1: Rename `document.txt` to `report.txt` using R key
- [ ] Scenario 2: Rename `Makefile` to `Makefile.old` using R key (fallback to full)
- [ ] Scenario 3: Rename any file to completely new name using Shift+R
- [ ] Scenario 4: Attempt to rename to existing filename (expect error)

### Edge Cases
- [ ] File with only extension (`.txt`) -> treat as no extension
- [ ] File with multiple consecutive dots (`file..txt`) -> ext: `.txt`
- [ ] File ending with dot (`file.`) -> no extension
- [ ] Very long filename that exceeds dialog width
- [ ] Unicode characters in filename

## Security Considerations

- **Input Validation:** Prevent directory traversal attacks (`../` in filename)
- **Filename Restrictions:** Block invalid characters (`/`, `\0`)
- **Permission Check:** Verify write permission before rename operation
- **Race Condition:** Handle concurrent modifications gracefully

## Error Handling

### Error Codes

| Error | Description | User Message |
|-------|-------------|--------------|
| ERR_EMPTY | Empty input | "File name cannot be empty" |
| ERR_EXISTS | File already exists | "File already exists" |
| ERR_INVALID | Invalid filename | "Invalid filename" |
| ERR_PERMISSION | Permission denied | "Permission denied" |
| ERR_IO | I/O error | "Rename failed: [details]" |

### Error Flow

```
User submits rename
    ↓
Validate input (empty, invalid chars)
    ↓ [invalid]
Show inline error, block Enter
    ↓ [valid]
Check for duplicate
    ↓ [exists]
Show inline error, block Enter
    ↓ [unique]
Execute rename
    ↓ [permission/IO error]
Show error dialog
    ↓ [success]
Refresh pane, close dialog
```

## Success Criteria

- [ ] All functional requirements (FR1-FR8) are implemented
- [ ] All user story acceptance criteria are met
- [ ] All unit tests pass with 80%+ coverage
- [ ] All integration tests pass
- [ ] Help dialog updated with new keybindings
- [ ] No regression in existing rename functionality
- [ ] Performance: dialog display < 50ms

## Open Questions

- None (all requirements confirmed)

## Implementation Phases

### Phase 1: Core Implementation
**Goals:** Implement basic extension-preserving rename
**Deliverables:**
- `hasEditableExtension()` function
- `ExtensionRenameDialog` component
- `ActionRenameFullName` action
- Updated keybindings

### Phase 2: Integration
**Goals:** Connect new dialog to existing system
**Deliverables:**
- Updated `handleRenameUI()` with mode detection
- New `handleRenameFullNameUI()` handler
- Message handling for new dialog

### Phase 3: Polish
**Goals:** Complete testing and documentation
**Deliverables:**
- Comprehensive unit tests
- Updated help dialog
- Edge case handling

## References

- 要件定義書: `doc/tasks/rename-file-keep-extension/要件定義書.md`
- Existing input dialog: `internal/ui/input_dialog.go`
- Existing rename dialog (for copy/move): `internal/ui/rename_input_dialog.go`
- Keybinding configuration: `internal/config/defaults.go`
- Action definitions: `internal/ui/actions.go`
