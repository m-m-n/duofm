# Implementation Plan: Extension-Preserving Rename Space Bug Fix

## Overview

This implementation plan addresses a bug where the `R` key (extension-preserving rename) fails when renaming files with spaces in their names. The fix involves investigating the root cause and adding comprehensive test coverage for space-containing filenames.

## Root Cause (Under Investigation)

**Symptom:**
- `R` key (extension-preserving rename) fails with files containing spaces in their names
- `Shift+R` (full name rename using `InputDialog`) works correctly with the same files

**Confirmed Working:**
- Space key input in `TextInput` works correctly (users can type spaces)
- `InputDialog` (used by Shift+R) handles spaces correctly

**Investigation Focus:**
The root cause is specific to `ExtensionRenameDialog` processing, not `TextInput` space handling.

**Areas to Investigate:**
1. **ExtensionRenameDialog.validateInput()**: How base name and extension are combined
2. **ExtensionRenameDialog.createResultCmd()**: How the result filename is generated
3. **Existing files map lookup**: Whether space-containing filenames are matched correctly
4. **Dialog initialization**: Whether base name with spaces is parsed correctly

**Key Difference:**
- `InputDialog` (Shift+R): User edits the full filename directly
- `ExtensionRenameDialog` (R key): User edits only the base name, extension is appended

The bug likely exists in how `ExtensionRenameDialog` processes the base name or combines it with the extension.

**Additional Investigation Points (from Codex review):**

1. **hasEditableExtension function**: Check whether the string splitting logic handles filenames with spaces correctly. The `strings.LastIndex` or similar functions used to separate base name from extension may have edge cases with space-containing filenames.

2. **handleExtensionRenameResult**: Verify that the `fs.Rename` call method matches how it's invoked in `InputDialog`. Any differences in how the filename string is passed (e.g., quoting, escaping) could cause the space-related failure.

## Objectives

- Identify the exact root cause in `ExtensionRenameDialog` processing
- Fix the identified issue with minimal code changes
- Add comprehensive test cases for filenames with spaces
- Ensure no regression in existing functionality (including Shift+R full name rename)

## Prerequisites

### Development Environment
- Go 1.21 or later
- Make (for build automation)

### Dependencies
- github.com/charmbracelet/bubbletea - TUI framework
- github.com/charmbracelet/lipgloss - Styling

### Knowledge Requirements
- Bubble Tea message handling and update loop
- Go string manipulation (runes vs bytes)
- File system operations and validation

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea (TUI)
- **Key Libraries**:
  - bubbletea - Handling keyboard input and state updates
  - lipgloss - Rendering dialog UI

### Design Approach

The bug investigation follows a systematic approach:
1. Reproduce the issue with a failing test
2. Trace the data flow through the components
3. Identify the exact failure point
4. Implement minimal fix
5. Verify with comprehensive tests

### Component Interaction

```
User Input (R key) -> Model -> ExtensionRenameDialog
                                    |
                                    v
                              TextInput.HandleKey()
                                    |
                                    v
                              validateInput()
                                    |
                                    v
                              fs.ValidateFilename()
                                    |
                                    v
                              createResultCmd() -> extensionRenameResultMsg
```

## Implementation Phases

### Phase 1: Bug Reproduction and Root Cause Identification

**Goal**: Create failing tests that reproduce the bug and identify the exact root cause in `ExtensionRenameDialog`.

**Investigation Target**: `internal/ui/extension_rename_dialog.go` - Focus on dialog-specific processing

**Files to Modify**:
- `internal/ui/extension_rename_dialog_test.go`:
  - Add test cases for space-containing filenames to reproduce the bug

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Test_SpaceInFilename | Demonstrate bug with space-containing filename | Test file structure | Failing test identifying bug location |
| Test_DialogInitialization | Verify dialog parses space-containing names correctly | Dialog exists | Base name correctly extracted |
| Test_ResultGeneration | Verify result contains correct filename | User input provided | fullName includes spaces correctly |

**Processing Flow**:
```
1. Create test file with space in name (e.g., "My Document.txt")
2. Create ExtensionRenameDialog with space-containing name
3. Verify dialog initialization
   |-- Base name correctly extracted -> Continue
   |-- Base name incorrect -> Root cause found in initialization
4. Simulate rename operation (Enter key)
5. Check result message
   |-- fullName incorrect -> Root cause found in result generation
   |-- Validation fails -> Root cause found in validateInput
6. Document exact failure point for Phase 2 fix
```

**Implementation Steps**:

1. **Add space-containing filename test cases (extension_rename_dialog_test.go)**
   - Test dialog creation with "My Document.txt" -> base name "My Document"
   - Test Enter key generates correct result "Your Document.txt"
   - Test validation accepts space-containing names

2. **Add debug output to identify failure point**
   - Log actual vs expected values at each step
   - Identify where the space handling breaks

3. **Add validation test for spaces**
   - Verify fs.ValidateFilename accepts spaces
   - Verify duplicate check with space-containing names

**Dependencies**:
- Requires: None (starting phase)
- Blocks: Phase 2

**Testing Approach**:

*Unit Tests*:
- Test ExtensionRenameDialog with space in original filename
- Test ExtensionRenameDialog result generation with spaces
- Test fs.ValidateFilename with space-containing names

**Acceptance Criteria**:
- [ ] Failing test demonstrates the bug clearly
- [ ] Test identifies which component/function is failing
- [ ] Test cases cover single space, multiple spaces, leading/trailing spaces
- [ ] Root cause is identified in ExtensionRenameDialog (not TextInput)

**Estimated Effort**: Small (1-2 days)

---

### Phase 2: Bug Fix Implementation

**Goal**: Implement the fix for the root cause identified in Phase 1 within `ExtensionRenameDialog`.

**File to Modify**:
- `internal/ui/extension_rename_dialog.go` - Fix the identified issue

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| ExtensionRenameDialog | Handle extension-preserving rename | Root cause identified | Space-containing filenames work correctly |

**Processing Flow**:
```
1. Open internal/ui/extension_rename_dialog.go
2. Apply fix based on Phase 1 findings:
   - If initialization issue: Fix base name extraction
   - If validation issue: Fix validateInput()
   - If result generation issue: Fix createResultCmd()
3. Run Phase 1 failing tests
   |-- Tests pass -> Fix confirmed
   |-- Tests fail -> Review implementation
4. Run full test suite for regression check
```

**Implementation Steps**:

1. **Apply targeted fix to ExtensionRenameDialog**
   - Location: `internal/ui/extension_rename_dialog.go`
   - Fix depends on Phase 1 root cause findings
   - Likely areas:
     - `validateInput()` - Line 82
     - `createResultCmd()` - Line 141
     - Dialog initialization

2. **Verify fix doesn't break existing functionality**
   - Run existing test suite
   - Ensure non-space filenames still work

**Dependencies**:
- Requires: Phase 1 (root cause identification)
- Blocks: Phase 3

**Testing Approach**:

*Unit Tests*:
- Verify all space-related tests from Phase 1 now pass
- Run full existing test suite for regression

*Manual Testing*:
- [ ] Rename "My Document.txt" to "Your Document.txt" using R key
- [ ] Verify Shift+R (full name rename) still works with spaces

**Acceptance Criteria**:
- [ ] All Phase 1 failing tests now pass
- [ ] All existing tests still pass
- [ ] Fix is minimal and targeted
- [ ] Shift+R functionality unchanged

**Estimated Effort**: Small (1-2 days)

**Risks and Mitigation**:
- **Risk**: Fix introduces regression in non-space filenames
  - **Mitigation**: Run full test suite, test with various filename types

---

### Phase 3: Comprehensive Test Coverage

**Goal**: Add comprehensive test coverage to prevent future regressions.

**Files to Modify**:
- `internal/ui/extension_rename_dialog_test.go`:
  - Add all test scenarios from specification

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| TestExtensionRenameDialog_SpaceInFilename | Test space scenarios | Test infrastructure | All space scenarios covered |
| TestExtensionRenameDialog_EdgeCases | Test edge cases | Test infrastructure | Edge cases verified |

**Processing Flow**:
```
1. Add test for single space in filename
2. Add test for multiple consecutive spaces
3. Add test for leading/trailing spaces
4. Add test for space in duplicate filename check
5. Add test for space-only filename (should be invalid)
6. Run all tests to verify coverage
```

**Implementation Steps**:

1. **Add space-containing filename tests**
   - Single space: "My Document.txt"
   - Multiple spaces: "My  Long  Document.txt"
   - Leading space: " Document.txt"
   - Trailing space: "Document .txt"

2. **Add validation edge case tests**
   - Space-only base name (should fail empty check after trim or be valid if spaces allowed)
   - Duplicate detection with spaces
   - Mixed normal and space characters
   - Spaces around the dot separator (e.g., `"My Doc .txt"`)
   - Hidden files with spaces (e.g., `".my doc.txt"`)

3. **Add integration-style tests**
   - Full rename flow simulation
   - Multiple input changes with spaces

**Dependencies**:
- Requires: Phase 2 (fix implemented)
- Blocks: None

**Testing Approach**:

*Unit Tests* (as defined in specification):

| Test | Input | Expected Result |
|------|-------|-----------------|
| Space in base name | "My Document", ".txt" | Valid, fullName = "My Document.txt" |
| Multiple spaces | "My  Long  Document", ".txt" | Valid, fullName = "My  Long  Document.txt" |
| Leading space | " Document", ".txt" | Valid, fullName = " Document.txt" |
| Trailing space | "Document ", ".txt" | Valid, fullName = "Document .txt" |
| Duplicate with space | "existing file" (existing.txt exists) | Error: "File already exists" |
| Space-only | "   ", ".txt" | Depends on design decision |
| Space around dot | "My Doc ", ".txt" | Valid, fullName = "My Doc .txt" |
| Hidden file with space | ".my doc", ".txt" | Valid, handles hidden file format |

**Acceptance Criteria**:
- [ ] All test scenarios from SPEC.md implemented
- [ ] Test coverage for extension_rename_dialog.go > 80%
- [ ] Edge cases documented and tested

**Estimated Effort**: Small (1 day)

---

## Complete File Structure

```
internal/
├── ui/
│   ├── extension_rename_dialog.go       # Primary fix target
│   ├── extension_rename_dialog_test.go  # Add test cases
│   └── text_input.go                    # Confirmed working (space input OK)
├── fs/
│   └── operations.go                    # ValidateFilename (likely OK)
└── doc/
    └── tasks/
        └── fix-rename-extension-bug/
            ├── SPEC.md                   # Specification
            ├── 要件定義書.md             # Requirements
            ├── IMPLEMENTATION.md         # This file
            └── VERIFICATION.md           # Verification checklist
```

**File Descriptions**:
- `extension_rename_dialog.go`: Dialog component for extension-preserving rename - **primary investigation/fix target**
- `extension_rename_dialog_test.go`: Test file, adding space-related test cases
- `text_input.go`: Reusable text input component - **confirmed working** (space input works correctly)
- `operations.go`: File system operations including ValidateFilename

## Testing Strategy

### Unit Testing

**Approach**:
- Use Go's built-in `testing` package
- Table-driven tests for multiple scenarios
- Temporary directories for file system tests

**Test Coverage Goals**:
- Core dialog logic: 80%+ coverage
- Space-related scenarios: 100% coverage

**Key Test Areas**:
1. **Dialog Initialization** (`extension_rename_dialog.go`)
   - Space-containing original filename
   - Correct base name extraction
   - Existing files map loading

2. **Input Handling** (`text_input.go`) - Confirmed Working
   - Space key input works correctly
   - InsertRunes handles space character properly

3. **Validation** (`extension_rename_dialog.go`)
   - Spaces accepted in filename
   - Duplicate check with spaces
   - Empty check behavior

4. **Result Generation** (`extension_rename_dialog.go`)
   - Correct full filename construction
   - Spaces preserved in result

### Manual Testing Checklist

Based on spec test scenarios:
- [ ] Rename "My Document.txt" to "Your Document.txt" using R key
- [ ] Type a name with spaces in the dialog input field
- [ ] Rename from spaced name to non-spaced name
- [ ] Rename from non-spaced name to spaced name
- [ ] Verify Shift+R (full name rename) still works with spaces

## Dependencies

### External Dependencies

| Package | Purpose |
|---------|---------|
| github.com/charmbracelet/bubbletea | TUI framework, keyboard handling |
| github.com/charmbracelet/lipgloss | UI styling |

### Internal Dependencies

**Implementation Order**:
1. Phase 1 (no dependencies)
2. Phase 2 (depends on Phase 1)
3. Phase 3 (depends on Phase 2)

## Risk Assessment

### Technical Risks

1. **Root Cause Misidentification**
   - **Risk**: Initial investigation identifies wrong component
   - **Likelihood**: Low (code review suggests limited scope)
   - **Impact**: Medium (delays fix)
   - **Mitigation**: Systematic investigation with debug logging

2. **Regression in Non-Space Filenames**
   - **Risk**: Fix breaks existing functionality
   - **Likelihood**: Low (targeted fix approach)
   - **Impact**: High (core functionality)
   - **Mitigation**: Run full test suite, manual verification

### Implementation Risks

1. **Incomplete Test Coverage**
   - **Risk**: Missing edge cases in tests
   - **Mitigation**: Follow spec test scenarios, add boundary tests

## Performance Considerations

- No performance impact expected (string operations are fast)
- No changes to file system operations

## Security Considerations

- Existing path traversal prevention maintained
- ValidateFilename rejects "/" in names
- No new security concerns with space handling

## Open Questions

### From Specification:
- [ ] Should space-only filenames be allowed? (Currently would fail empty check)
- [ ] Exact error message when space-related failure occurs (for debugging)

### To Clarify with User:
- [ ] Are leading/trailing spaces intentionally allowed? (Some file systems handle them differently)

## Success Metrics

### Functional Completeness
- [ ] R key rename works with space-containing filenames
- [ ] All new test cases pass
- [ ] All existing test cases pass

### Quality Metrics
- [ ] Test coverage > 80% for dialog component
- [ ] No regressions identified

### User Experience
- [ ] Rename dialog accepts space input
- [ ] Clear error messages if validation fails

## References

- **Specification**: `doc/tasks/fix-rename-extension-bug/SPEC.md`
- **Requirements**: `doc/tasks/fix-rename-extension-bug/要件定義書.md`
- **Original Feature**: `doc/tasks/rename-file-keep-extension/SPEC.md`
- **Source Files**:
  - `internal/ui/extension_rename_dialog.go`
  - `internal/ui/extension_rename_dialog_test.go`
  - `internal/ui/text_input.go`
  - `internal/fs/operations.go`

## Next Steps

After reviewing this implementation plan:

1. **Begin Investigation (Phase 1)**
   - Add failing test cases to reproduce the bug
   - Identify exact failure point

2. **Implement Fix (Phase 2)**
   - Apply targeted fix based on investigation
   - Verify fix resolves the issue

3. **Complete Testing (Phase 3)**
   - Add comprehensive test coverage
   - Run full regression suite

4. **Code Review and Merge**
   - Review changes for quality
   - Merge to main branch
