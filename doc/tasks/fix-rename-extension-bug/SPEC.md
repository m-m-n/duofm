# Bug Fix: Extension-Preserving Rename with Spaces in Filename

## Overview

This bug fix addresses an issue where the `R` key (extension-preserving rename) fails when renaming files with spaces in their names. The root cause is in the `ExtensionRenameDialog` component where base name and extension are concatenated using simple string concatenation.

## Objectives

- Fix the extension-preserving rename functionality to handle filenames with spaces
- Add comprehensive test cases for space-containing filenames
- Ensure no regression in existing functionality

## Problem Analysis

### Affected Code

**File:** `internal/ui/extension_rename_dialog.go`

```go
// Line 82: In validateInput()
fullName := d.textInput.Value + d.extension

// Line 141: In createResultCmd()
fullName := d.textInput.Value + d.extension
```

### Behavior Comparison

| Operation | Component | Handles Spaces |
|-----------|-----------|----------------|
| R key (extension-preserving) | ExtensionRenameDialog | BUG |
| Shift+R (full name) | InputDialog | OK |

### Root Cause Investigation

After code review, the string concatenation itself (`d.textInput.Value + d.extension`) should work correctly in Go for any string including those with spaces. The actual root cause needs further investigation:

1. **TextInput component**: The `TextInput.Value` property may have issues handling spaces
2. **Validation logic**: The `fs.ValidateFilename()` function or other validation may reject spaces
3. **Key handling**: Space key might be intercepted or handled incorrectly in the dialog

### Potential Issues to Check

1. **TextInput.HandleKey()**: Does it properly handle space key input?
2. **fs.ValidateFilename()**: Does it have any restrictions on spaces?
3. **existingFiles map lookup**: Are filenames with spaces being matched correctly?

## Technical Requirements

### Functional Requirements
- **FR1:** Files with spaces in names can be renamed using R key
- **FR2:** Validation correctly handles filenames with spaces
- **FR3:** Result message contains correct filename with spaces
- **FR4:** Existing behavior for files without spaces is unchanged

### Non-Functional Requirements
- **NFR1 - Performance:** No performance degradation
- **NFR2 - Compatibility:** No breaking changes to existing API

## Implementation Approach

### Investigation Steps

1. **Reproduce the bug**: Create a test file with spaces and attempt R key rename
2. **Add debug logging**: Track the actual values at each step
3. **Verify TextInput**: Confirm space handling in TextInput component
4. **Check validation**: Verify fs.ValidateFilename() accepts spaces

### Potential Fix Areas

#### Area 1: TextInput Space Handling

Check `internal/ui/text_input.go` for space key handling:

```go
func (t *TextInput) HandleKey(msg tea.KeyMsg) bool {
    switch msg.Type {
    case tea.KeySpace:
        // Ensure space is handled correctly
        t.Value = t.Value[:t.CursorPos] + " " + t.Value[t.CursorPos:]
        t.CursorPos++
        return true
    // ... other cases
    }
}
```

#### Area 2: Validation Function

Verify `internal/fs/operations.go` ValidateFilename() accepts spaces:

```go
func ValidateFilename(name string) error {
    if name == "" {
        return fmt.Errorf("file name cannot be empty")
    }
    if strings.Contains(name, "/") {
        return fmt.Errorf("invalid file name: path separator not allowed")
    }
    // Spaces should NOT be rejected here
    return nil
}
```

#### Area 3: Existing Files Map Lookup

Verify the duplicate check handles spaces correctly:

```go
// In validateInput()
fullName := d.textInput.Value + d.extension
// This comparison should work with spaces
if fullName != d.originalName && d.existingFiles[fullName] {
    // ...
}
```

### File Structure

```
internal/ui/
├── extension_rename_dialog.go       # Primary investigation target
├── extension_rename_dialog_test.go  # Add space-handling tests
└── text_input.go                    # Check space key handling
```

## Test Scenarios

### New Unit Tests for extension_rename_dialog_test.go

#### Test: Space in Filename Dialog Creation

```go
func TestExtensionRenameDialog_SpaceInFilename(t *testing.T) {
    tmpDir := t.TempDir()

    t.Run("creates dialog with space in base name", func(t *testing.T) {
        dialog := NewExtensionRenameDialog(tmpDir, "My Document.txt", "My Document", ".txt")

        if dialog == nil {
            t.Fatal("NewExtensionRenameDialog returned nil")
        }

        if dialog.Input() != "My Document" {
            t.Errorf("Input() = %q, want %q", dialog.Input(), "My Document")
        }

        if dialog.extension != ".txt" {
            t.Errorf("extension = %q, want %q", dialog.extension, ".txt")
        }
    })
}
```

#### Test: Enter with Space in Base Name

```go
t.Run("Enter with space in base name generates correct filename", func(t *testing.T) {
    dialog := NewExtensionRenameDialog(tmpDir, "My Document.txt", "My Document", ".txt")
    dialog.SetInput("Your Document")

    _, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

    if cmd == nil {
        t.Fatal("Expected cmd, got nil")
    }

    msg := cmd()
    result, ok := msg.(extensionRenameResultMsg)
    if !ok {
        t.Fatalf("Expected extensionRenameResultMsg, got %T", msg)
    }

    if result.newName != "Your Document.txt" {
        t.Errorf("newName = %q, want %q", result.newName, "Your Document.txt")
    }
})
```

#### Test: Multiple Spaces

```go
t.Run("handles multiple spaces in filename", func(t *testing.T) {
    dialog := NewExtensionRenameDialog(tmpDir, "My  Long  Document.txt", "My  Long  Document", ".txt")
    dialog.SetInput("Another  Long  Name")

    _, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

    msg := cmd()
    result := msg.(extensionRenameResultMsg)

    if result.newName != "Another  Long  Name.txt" {
        t.Errorf("newName = %q, want %q", result.newName, "Another  Long  Name.txt")
    }
})
```

#### Test: Leading and Trailing Spaces

```go
t.Run("handles leading and trailing spaces", func(t *testing.T) {
    dialog := NewExtensionRenameDialog(tmpDir, " Document .txt", " Document ", ".txt")
    dialog.SetInput(" New Name ")

    _, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

    msg := cmd()
    result := msg.(extensionRenameResultMsg)

    if result.newName != " New Name .txt" {
        t.Errorf("newName = %q, want %q", result.newName, " New Name .txt")
    }
})
```

#### Test: Validation with Spaces

```go
t.Run("validates filename with spaces correctly", func(t *testing.T) {
    dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
    dialog.SetInput("new name with spaces")

    // Should not have error for valid name with spaces
    if dialog.hasError {
        t.Errorf("Expected no error for valid name with spaces, got: %s", dialog.errorMessage)
    }
})
```

#### Test: Duplicate Check with Spaces

```go
t.Run("detects duplicate with space in name", func(t *testing.T) {
    // Create existing file with space
    os.WriteFile(filepath.Join(tmpDir, "existing file.txt"), []byte("test"), 0644)

    dialog := NewExtensionRenameDialog(tmpDir, "document.txt", "document", ".txt")
    dialog.SetInput("existing file")

    if !dialog.hasError {
        t.Error("Expected hasError = true for duplicate filename")
    }

    if dialog.errorMessage != "File already exists" {
        t.Errorf("errorMessage = %q, want %q", dialog.errorMessage, "File already exists")
    }
})
```

#### Test: Space Around Dot Separator (TS-7)

```go
t.Run("handles space around dot separator", func(t *testing.T) {
    // Test case: "My Doc .txt" - space before the dot separator
    dialog := NewExtensionRenameDialog(tmpDir, "My Doc .txt", "My Doc ", ".txt")
    dialog.SetInput("Your Doc ")

    _, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

    msg := cmd()
    result := msg.(extensionRenameResultMsg)

    if result.newName != "Your Doc .txt" {
        t.Errorf("newName = %q, want %q", result.newName, "Your Doc .txt")
    }
})
```

#### Test: Hidden File with Space (TS-8)

```go
t.Run("handles hidden file with space in name", func(t *testing.T) {
    // Test case: ".my doc.txt" - hidden file with space in the visible name
    dialog := NewExtensionRenameDialog(tmpDir, ".my doc.txt", ".my doc", ".txt")
    dialog.SetInput(".your doc")

    _, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

    msg := cmd()
    result := msg.(extensionRenameResultMsg)

    if result.newName != ".your doc.txt" {
        t.Errorf("newName = %q, want %q", result.newName, ".your doc.txt")
    }
})
```

### Integration Tests

- [ ] Complete rename flow with space-containing filename
- [ ] Rename from spaced name to another spaced name
- [ ] Rename from spaced name to non-spaced name
- [ ] Rename from non-spaced name to spaced name

### E2E Tests

- [ ] Scenario: Rename "My Document.txt" to "Your Document.txt" using R key
- [ ] Scenario: Type a name with spaces in the dialog input field

## Error Handling

### Expected Error Cases

| Case | Expected Behavior |
|------|-------------------|
| Empty input after removing spaces | Show "File name cannot be empty" |
| Duplicate filename with spaces | Show "File already exists" |
| Path separator in input | Show "Invalid file name" |

### Error Flow

```
User types filename with spaces
    |
    v
validateInput() called
    |
    v
fullName = baseName + extension (spaces preserved)
    |
    v
[Validation checks]
    |-- Empty check: PASS (spaces don't make it empty)
    |-- Duplicate check: Compare with existingFiles map
    |-- Invalid chars check: fs.ValidateFilename()
    |
    v
[All pass] -> hasError = false
[Any fail] -> hasError = true, show error
```

## Success Criteria

- [ ] "My Document.txt" can be renamed to "Your Document.txt" using R key
- [ ] All new test cases pass
- [ ] All existing test cases still pass
- [ ] No performance regression
- [ ] Shift+R (full name rename) behavior unchanged

## Implementation Checklist

### Phase 1: Investigation
- [ ] Run existing tests to establish baseline
- [ ] Add debug logging to track values
- [ ] Identify exact failure point

### Phase 2: Fix
- [ ] Implement fix based on investigation
- [ ] Add new test cases
- [ ] Verify fix resolves the issue

### Phase 3: Verification
- [ ] Run all unit tests
- [ ] Run E2E tests
- [ ] Manual testing with various space scenarios
- [ ] Code review

## References

- Original feature specification: `doc/tasks/rename-file-keep-extension/SPEC.md`
- Requirements document: `doc/tasks/fix-rename-extension-bug/要件定義書.md`
- Source file: `internal/ui/extension_rename_dialog.go`
- Test file: `internal/ui/extension_rename_dialog_test.go`
- TextInput component: `internal/ui/text_input.go`
- Validation function: `internal/fs/operations.go`
