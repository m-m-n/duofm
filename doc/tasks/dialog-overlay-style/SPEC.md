# SPEC: Dialog Overlay Style Improvement

## Overview

Improve the visual appearance of the background pane(s) when dialogs are displayed. Instead of completely obscuring the active pane with block characters (█), render the original file list with a gray background and dimmed text, maintaining visibility while emphasizing the dialog.

## Objectives

1. Keep the original file list visible when dialogs are displayed
2. Create a visually appealing "recessed" effect for the background pane(s)
3. Maintain dialog emphasis through contrast
4. Apply different background dimming based on dialog type (full-screen vs pane-local)

## User Stories

### US-1: View Background Content During Dialog
**As a** user
**I want to** see the file list behind a dialog
**So that** I can reference the original content while making decisions

### US-2: Visual Dialog Emphasis
**As a** user
**I want to** clearly distinguish the dialog from the background
**So that** I can focus on the dialog interaction

### US-3: Context-Aware Background Dimming
**As a** user
**I want** pane-specific dialogs to only dim that pane
**So that** I can still see the other pane clearly for reference

## Technical Requirements

### TR-1: Overlay Style Rendering

When a dialog is active, affected pane(s) must be rendered with:
- **Background color**: Dark gray (lipgloss Color "236" or equivalent)
- **Text color**: Dimmed gray (lipgloss Color "243" or equivalent)
- **Content**: Original file entries, sizes, dates preserved

### TR-2: Dialog Display Types

Dialogs are classified into two types with different behavior:

#### Full-Screen Dialogs (Center of screen, both panes dimmed)
| Dialog | Trigger | Position | Dimmed Panes |
|--------|---------|----------|--------------|
| HelpDialog | `?` key | Screen center | Both |
| ErrorDialog | Error event | Screen center | Both |

#### Pane-Local Dialogs (Center of active pane, only that pane dimmed)
| Dialog | Trigger | Position | Dimmed Panes |
|--------|---------|----------|--------------|
| ConfirmDialog | `d` key | Active pane center | Active pane only |
| ContextMenuDialog | `@` key | Active pane center | Active pane only |

### TR-3: Dialog Type Interface

Each dialog must indicate its display type:

```go
type DialogDisplayType int

const (
    DialogDisplayPane   DialogDisplayType = iota // Pane-local dialog
    DialogDisplayScreen                          // Full-screen dialog
)

// Add to Dialog interface
type Dialog interface {
    Update(msg tea.Msg) (Dialog, tea.Cmd)
    View() string
    IsActive() bool
    DisplayType() DialogDisplayType // New method
}
```

## Implementation Approach

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Model                                │
│                                                             │
│  Dialog.DisplayType() determines rendering:                 │
│                                                             │
│  DialogDisplayPane:           DialogDisplayScreen:          │
│  ┌─────────┬─────────┐       ┌─────────┬─────────┐         │
│  │ Dimmed  │ Normal  │       │ Dimmed  │ Dimmed  │         │
│  │  ┌───┐  │         │       │     ┌───────┐     │         │
│  │  │Dlg│  │         │       │     │  Dlg  │     │         │
│  │  └───┘  │         │       │     └───────┘     │         │
│  └─────────┴─────────┘       └─────────┴─────────┘         │
└─────────────────────────────────────────────────────────────┘
```

### Implementation Steps

1. **Define dialog display type** in `dialog.go`:
   ```go
   type DialogDisplayType int

   const (
       DialogDisplayPane   DialogDisplayType = iota
       DialogDisplayScreen
   )
   ```

2. **Add `DisplayType()` method to Dialog interface** in `dialog.go`

3. **Implement `DisplayType()` in each dialog**:
   - `HelpDialog`: return `DialogDisplayScreen`
   - `ErrorDialog`: return `DialogDisplayScreen`
   - `ConfirmDialog`: return `DialogDisplayPane`
   - `ContextMenuDialog`: return `DialogDisplayPane`

4. **Add `ViewDimmed()` method to Pane** in `pane.go`:
   ```go
   func (p *Pane) ViewDimmedWithDiskSpace(diskSpace string) string {
       // Same as ViewWithDiskSpace but with dimmed colors
   }
   ```

5. **Update dialog rendering in `model.go`**:
   ```go
   if m.dialog != nil && m.dialog.IsActive() {
       switch m.dialog.DisplayType() {
       case DialogDisplayScreen:
           // Dim both panes, center dialog on screen
       case DialogDisplayPane:
           // Dim active pane only, center dialog on active pane
       }
   }
   ```

### Code Changes

#### dialog.go

```go
type DialogDisplayType int

const (
    DialogDisplayPane   DialogDisplayType = iota
    DialogDisplayScreen
)

type Dialog interface {
    Update(msg tea.Msg) (Dialog, tea.Cmd)
    View() string
    IsActive() bool
    DisplayType() DialogDisplayType
}
```

#### help_dialog.go / error_dialog.go

```go
func (d *HelpDialog) DisplayType() DialogDisplayType {
    return DialogDisplayScreen
}
```

#### confirm_dialog.go / context_menu_dialog.go

```go
func (d *ConfirmDialog) DisplayType() DialogDisplayType {
    return DialogDisplayPane
}
```

#### pane.go

```go
func (p *Pane) ViewDimmedWithDiskSpace(diskSpace string) string {
    dimmedBg := lipgloss.Color("236")
    dimmedFg := lipgloss.Color("243")
    // Apply dimmed styles to all elements
}
```

#### model.go

```go
if m.dialog != nil && m.dialog.IsActive() {
    switch m.dialog.DisplayType() {
    case DialogDisplayScreen:
        // Both panes dimmed, dialog centered on full screen
        leftView := m.leftPane.ViewDimmedWithDiskSpace(m.leftDiskSpace)
        rightView := m.rightPane.ViewDimmedWithDiskSpace(m.rightDiskSpace)
        panes := lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
        // Overlay dialog at screen center

    case DialogDisplayPane:
        // Only active pane dimmed
        if m.activePane == LeftPane {
            leftView := m.leftPane.ViewDimmedWithDiskSpace(m.leftDiskSpace)
            // Overlay dialog on left pane
            rightView := m.rightPane.ViewWithDiskSpace(m.rightDiskSpace)
        } else {
            leftView := m.leftPane.ViewWithDiskSpace(m.leftDiskSpace)
            rightView := m.rightPane.ViewDimmedWithDiskSpace(m.rightDiskSpace)
            // Overlay dialog on right pane
        }
    }
}
```

## Dependencies

- `github.com/charmbracelet/lipgloss` - Already in use

## Test Scenarios

| ID | Scenario | Expected Result |
|----|----------|-----------------|
| T1 | Press `?` to show help dialog | Both panes dimmed, dialog at screen center |
| T2 | Press `d` on left pane | Left pane dimmed, right pane normal, dialog at left pane center |
| T3 | Press `d` on right pane | Right pane dimmed, left pane normal, dialog at right pane center |
| T4 | Press `@` on left pane | Left pane dimmed, right pane normal |
| T5 | Press `@` on right pane | Right pane dimmed, left pane normal |
| T6 | Trigger error dialog | Both panes dimmed, dialog at screen center |
| T7 | Close any dialog | All panes return to normal style immediately |

## Security Considerations

No security implications - this is a visual styling change only.

## Error Handling

No new error conditions introduced.

## Performance Optimization

- Reuse existing rendering logic with style modifications
- No additional allocations beyond style objects
- Event-driven rendering (Bubble Tea) ensures no continuous overhead

## Success Criteria

1. [ ] File list visible behind all dialogs
2. [ ] Gray background and dimmed text create "recessed" visual effect
3. [ ] Full-screen dialogs (help, error) dim both panes
4. [ ] Pane-local dialogs (confirm, context menu) dim only the active pane
5. [ ] No performance regression

## Open Questions

None - all requirements confirmed.
