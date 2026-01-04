# Dialog Implementation Best Practices

This document provides guidelines for implementing dialogs in duofm following the Bubble Tea framework and the established message-based cancellation pattern.

## Table of Contents

1. [Overview](#overview)
2. [Dialog Lifecycle](#dialog-lifecycle)
3. [Message-Based Cancellation Pattern](#message-based-cancellation-pattern)
4. [Common Pitfalls](#common-pitfalls)
5. [Implementation Checklist](#implementation-checklist)
6. [Code Examples](#code-examples)
7. [Testing Requirements](#testing-requirements)

## Overview

Dialogs in duofm follow the Bubble Tea (Elm Architecture) pattern where all state changes occur through messages. Proper dialog implementation is critical to ensure the application remains responsive after dialog operations.

**Why This Matters:**

A common bug pattern occurs when dialogs set their `active` flag to `false` but fail to notify the Model. This leaves the Model with a non-nil dialog reference that delegates all inputs to an inactive dialog, effectively freezing the application.

## Dialog Lifecycle

### Correct Flow

```
1. User Action → Open Dialog
   ├─ Model sets m.dialog = NewDialog()
   └─ Dialog.active = true

2. User Action (Esc) → Cancel Dialog
   ├─ Dialog sets d.active = false
   ├─ Dialog returns cancellation message (tea.Cmd)
   └─ Return (d, cmd)

3. Model receives cancellation message
   ├─ Model processes message in Update()
   ├─ Model sets m.dialog = nil
   └─ Normal key handling resumes
```

### Buggy Flow (AVOID)

```
1. User Action (Esc) → Cancel Dialog
   ├─ Dialog sets d.active = false
   ├─ Dialog returns nil          ← BUG: No message sent
   └─ Return (d, nil)

2. Model state
   ├─ m.dialog remains non-nil     ← BUG: Dialog not cleared
   └─ All inputs delegated to inactive dialog

3. Inactive dialog
   ├─ Early return on all input    ← BUG: Inputs ignored
   └─ Application appears frozen   ← SYMPTOM
```

## Message-Based Cancellation Pattern

### Step 1: Define Cancellation Message

```go
// Define a cancellation message type
type myDialogCancelMsg struct{}
```

### Step 2: Implement Esc Key Handling in Dialog

```go
func (d *MyDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			d.active = false
			// CRITICAL: Return cancellation message, not nil
			return d, func() tea.Msg {
				return myDialogCancelMsg{}
			}
		// ... other key handling
		}
	}

	return d, nil
}
```

### Step 3: Add Message Handler in Model

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case myDialogCancelMsg:
		// CRITICAL: Clear dialog reference
		m.dialog = nil
		return m, nil

	// ... other message handling
	}

	return m, nil
}
```

## Common Pitfalls

### Pitfall 1: Returning nil on Esc

**Problem:**
```go
case tea.KeyEsc:
	d.active = false
	return d, nil  // ← WRONG: Dialog not cleared from Model
```

**Solution:**
```go
case tea.KeyEsc:
	d.active = false
	return d, func() tea.Msg {
		return myDialogCancelMsg{}
	}
```

### Pitfall 2: Forgetting Message Handler

**Problem:**
Dialog sends cancel message, but Model has no handler to process it.

**Solution:**
Always add corresponding message handler in `Model.Update()` or appropriate sub-handler.

### Pitfall 3: Not Checking Cancelled Flag

For dialogs using result messages with a `cancelled` field:

**Problem:**
```go
func (m Model) handleDialogResult(msg dialogResultMsg) (tea.Model, tea.Cmd) {
	m.dialog = nil

	// Process result even if cancelled
	return m.processResult(msg.value)  // ← WRONG
}
```

**Solution:**
```go
func (m Model) handleDialogResult(msg dialogResultMsg) (tea.Model, tea.Cmd) {
	m.dialog = nil

	// Check cancelled flag first
	if msg.cancelled {
		return m, nil  // Early return, no action
	}

	return m.processResult(msg.value)
}
```

### Pitfall 4: Inconsistent active State

**Problem:**
Dialog claims to be active but has already been cancelled.

**Solution:**
Always set `d.active = false` BEFORE returning cancellation message.

## Implementation Checklist

Use this checklist when implementing a new dialog or reviewing existing ones:

### Dialog Implementation

- [ ] **Cancellation Message Defined**: Created a unique message type (e.g., `myDialogCancelMsg`)
- [ ] **Esc Key Handling**: Esc key case exists in `Update()`
- [ ] **active Flag Set**: `d.active = false` called before returning
- [ ] **Message Returned**: Returns cancellation message command, NOT nil
- [ ] **Inactive Check**: Early return if `!d.active` at start of `Update()`
- [ ] **DisplayType Implemented**: Returns correct `DialogDisplayType`
- [ ] **View Returns Empty**: `View()` returns empty string when `!d.active`

### Model Integration

- [ ] **Message Handler Added**: Model has handler for cancellation message
- [ ] **Dialog Cleared**: Handler sets `m.dialog = nil`
- [ ] **Result Handling**: If using result message, checks `cancelled` flag

### Testing

- [ ] **Unit Test - Esc Key**: Test verifies Esc deactivates and returns message
- [ ] **Unit Test - Inactive**: Test verifies inactive dialog ignores input
- [ ] **Integration Test**: Test verifies Model clears dialog on cancel message
- [ ] **Integration Test - Result**: If applicable, test cancelled flag handling

## Code Examples

### Example 1: Simple Cancel-Only Dialog

```go
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message type
type myDialogCancelMsg struct{}

// Dialog struct
type MyDialog struct {
	title  string
	active bool
}

// Constructor
func NewMyDialog(title string) *MyDialog {
	return &MyDialog{
		title:  title,
		active: true,
	}
}

// Update handles input
func (d *MyDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			d.active = false
			return d, func() tea.Msg {
				return myDialogCancelMsg{}
			}
		}
	}

	return d, nil
}

// View renders the dialog
func (d *MyDialog) View() string {
	if !d.active {
		return ""
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2)

	return style.Render(d.title + "\n\nPress Esc to close")
}

// IsActive returns active state
func (d *MyDialog) IsActive() bool {
	return d.active
}

// DisplayType returns display type
func (d *MyDialog) DisplayType() DialogDisplayType {
	return DialogDisplayScreen
}
```

**Model Handler:**

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case myDialogCancelMsg:
		m.dialog = nil
		return m, nil

	// ... other handlers
	}

	return m, nil
}
```

### Example 2: Dialog with Result Message

```go
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Result message with cancelled flag
type myDialogResultMsg struct {
	value     string
	cancelled bool
}

type MyResultDialog struct {
	input  string
	active bool
}

func NewMyResultDialog() *MyResultDialog {
	return &MyResultDialog{
		active: true,
	}
}

func (d *MyResultDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			d.active = false
			return d, func() tea.Msg {
				return myDialogResultMsg{
					cancelled: true,
				}
			}

		case tea.KeyEnter:
			d.active = false
			return d, func() tea.Msg {
				return myDialogResultMsg{
					value:     d.input,
					cancelled: false,
				}
			}

		case tea.KeyRunes:
			d.input += string(msg.Runes)
			return d, nil
		}
	}

	return d, nil
}

// ... View, IsActive, DisplayType methods
```

**Model Handler:**

```go
func (m Model) handleMyDialogResult(msg myDialogResultMsg) (tea.Model, tea.Cmd) {
	m.dialog = nil

	// Check cancelled flag first
	if msg.cancelled {
		return m, nil
	}

	// Process result
	return m.processValue(msg.value)
}
```

## Testing Requirements

### Unit Tests

Every dialog must have unit tests covering:

1. **Esc Key Cancellation**
   ```go
   func TestMyDialog_EscKey(t *testing.T) {
       d := NewMyDialog("Test")

       _, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEsc})

       if d.active {
           t.Error("Dialog should be inactive after Esc")
       }

       if cmd == nil {
           t.Error("Dialog should return cancel message")
       }

       msg := cmd()
       if _, ok := msg.(myDialogCancelMsg); !ok {
           t.Errorf("Message type = %T, want myDialogCancelMsg", msg)
       }
   }
   ```

2. **Inactive State Ignores Input**
   ```go
   func TestMyDialog_InactiveIgnoresInput(t *testing.T) {
       d := NewMyDialog("Test")
       d.active = false

       _, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})

       if cmd != nil {
           t.Error("Inactive dialog should return nil command")
       }
   }
   ```

### Integration Tests

Test Model-Dialog integration:

```go
func TestMyDialogCancellationIntegration(t *testing.T) {
	m := createTestModel(t)

	// Open dialog
	m.dialog = NewMyDialog("Test")

	// Send Esc
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if cmd == nil {
		t.Fatal("Dialog should return cancel message")
	}

	// Process cancel message
	msg := cmd()
	updated, _ = m.Update(msg)
	m = updated.(Model)

	// Verify dialog cleared
	if m.dialog != nil {
		t.Error("Model should clear dialog after cancel")
	}
}
```

## Reference Implementations

Good examples to study in the codebase:

1. **ConfirmDialog** (`internal/ui/confirm_dialog.go`)
   - Clean implementation of message-based cancellation
   - Reference implementation for cancel-only dialogs

2. **PermissionDialog** (`internal/ui/permission_dialog.go`)
   - Fixed implementation following this guide
   - Good example of standalone cancel message

3. **InputDialog** (`internal/ui/input_dialog.go`)
   - Fixed implementation using result message with cancelled flag
   - Good example of result-based cancellation

## Review Checklist for Code Reviewers

When reviewing PRs that add or modify dialogs:

- [ ] Dialog sends message on Esc (not nil)
- [ ] Model has corresponding message handler
- [ ] Message handler clears `m.dialog`
- [ ] Unit tests cover Esc key behavior
- [ ] Integration tests verify Model integration
- [ ] Inactive dialog ignores input (early return)
- [ ] View returns empty string when inactive

## FAQ

### Q: When should I use a standalone cancel message vs. result message with cancelled flag?

**A:** Use standalone cancel message (e.g., `myDialogCancelMsg`) when:
- Dialog has no result value
- Cancellation and confirmation are clearly separate actions

Use result message with `cancelled` flag when:
- Dialog always produces a result structure
- You want to handle both cases in the same handler
- Callback-based dialogs (InputDialog pattern)

### Q: Can I use the same message type for multiple dialogs?

**A:** No. Each dialog should have its own message type to avoid conflicts and improve clarity. Use descriptive names like `permissionDialogCancelMsg`, not generic names like `cancelMsg`.

### Q: What if my dialog has multiple steps?

**A:** Each step that can be cancelled should follow the same pattern. The RecursivePermDialog is a good example - it can be cancelled at step 1 or step 2, and both send the same cancellation message.

### Q: Should ErrorDialog have Esc handling?

**A:** Yes, even informational dialogs should have Esc handling for consistency. Users expect Esc to close any dialog.

## Conclusion

Following these best practices ensures:
- Application remains responsive after dialog cancellation
- Consistent user experience across all dialogs
- Easier maintenance and debugging
- Prevention of common dialog-related bugs

When in doubt, refer to ConfirmDialog as the reference implementation.

## Related Documentation

- [Dialog Interface](../../internal/ui/dialog.go)
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Contributing Guidelines](../CONTRIBUTING.md)
