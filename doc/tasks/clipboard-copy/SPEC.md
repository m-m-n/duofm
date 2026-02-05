# Feature: Clipboard Copy

## Overview

Add "Copy file name" and "Copy full path" items to the context menu. The selected file's name or absolute path is written to the system clipboard using OSC 52 with external command fallback.

## Domain Rules

- Only the file under the cursor is copied (single file only).
- Parent directory entry (`..`) disables copy items.
- When marked files exist (batch mode), copy items are disabled.
- OSC 52 is attempted first; external commands are used as fallback.

## Objectives

- Copy the file name of the cursor entry to the clipboard via context menu.
- Copy the full path of the cursor entry to the clipboard via context menu.

## User Stories

### US1: Copy File Name
As a user, I want to copy the file name to the clipboard from the context menu, so that I can paste it elsewhere.

**Acceptance Criteria:**
- [ ] Context menu contains "Copy file name" item
- [ ] Selecting the item copies the file name to the clipboard
- [ ] Status bar shows `Copied: {filename}` on success
- [ ] Status bar message clears after 3 seconds

### US2: Copy Full Path
As a user, I want to copy the full path to the clipboard from the context menu, so that I can paste it elsewhere.

**Acceptance Criteria:**
- [ ] Context menu contains "Copy full path" item
- [ ] Selecting the item copies the absolute path to the clipboard
- [ ] Status bar shows `Copied: {fullpath}` on success
- [ ] Status bar message clears after 3 seconds

## Functional Requirements

- **FR1:** Add "Copy file name" menu item to context menu with action ID `copy_name`.
- **FR2:** Add "Copy full path" menu item to context menu with action ID `copy_path`.
- **FR3:** `copy_name` copies the file name (e.g., `document.txt`) to the clipboard.
- **FR4:** `copy_path` copies the absolute path (e.g., `/home/user/document.txt`) to the clipboard.
- **FR5:** Both items are disabled when the cursor is on `..` (parent directory).
- **FR6:** Both items are disabled when marked files exist (markCount > 0).
- **FR7:** Clipboard write uses OSC 52 escape sequence as the primary method.
- **FR8:** If OSC 52 is not available, fall back to external commands in this order: `wl-copy`, `xclip -selection clipboard`, `xsel --clipboard --input`.
- **FR9:** On success, display `Copied: {text}` in the status bar for 3 seconds.
- **FR10:** On failure, display `Copy failed: {error}` in the status bar for 3 seconds.

## Non-Functional Requirements

- **NFR1 - Performance:** Clipboard copy operation completes within 100ms.
- **NFR2 - Compatibility:** Works on Linux with OSC 52-capable terminals and as fallback with xclip/xsel/wl-copy.

## Interface Contract

### Input/Output Specification

**Copy file name:**
- Input: `fs.FileEntry.Name` (string)
- Output: File name string written to system clipboard

**Copy full path:**
- Input: Active pane path (string) + `fs.FileEntry.Name` (string)
- Output: `filepath.Join(panePath, entryName)` written to system clipboard

### Preconditions
- Cursor is on a valid file/directory entry (not `..`).
- No marked files exist.

### Postconditions
- System clipboard contains the copied text.
- Status bar displays result message.

### Error Conditions
- No clipboard tool available and OSC 52 not supported: display error in status bar.
- External command execution failure: display error in status bar.
- `/dev/tty` open failure and no external command available: display `Copy failed: no clipboard method available` in status bar.

## Context Menu Integration

The two new items are added to `buildMenuItems()` in `context_menu_dialog.go`.

**Position:** After the existing file operation items (copy/move/delete) and before compress.

**Menu items:**

| Order | ID | Label | Enabled Condition |
|-------|----|-------|-------------------|
| (after delete) | `copy_name` | Copy file name | `markCount == 0 && !entry.IsParentDir()` |
| (after copy_name) | `copy_path` | Copy full path | `markCount == 0 && !entry.IsParentDir()` |

## Clipboard Implementation

### OSC 52

Write the OSC 52 escape sequence to `/dev/tty`:

```
\033]52;c;{base64-encoded-text}\a
```

Where `{base64-encoded-text}` is the base64 encoding of the text to copy.

`/dev/tty` に直接書き込む理由: duofm は Bubble Tea の `tea.WithAltScreen()` で代替画面バッファを使用しており、`os.Stdout` への直接書き込みは Bubble Tea のレンダリングと競合する。`/dev/tty` への書き込みは Bubble Tea の出力管理を迂回しつつ、ターミナルにエスケープシーケンスを正しく送信できる。

### External Command Fallback

Detection order:
1. `wl-copy` - Wayland clipboard
2. `xclip -selection clipboard` - X11 clipboard (xclip)
3. `xsel --clipboard --input` - X11 clipboard (xsel)

The text is written to the command's stdin via pipe.

### Fallback Strategy

1. Attempt OSC 52 via `/dev/tty` first (best-effort; no reliable way to detect terminal support).
2. Additionally attempt external command if available (belt-and-suspenders approach).
3. If no external command is found and OSC 52 was attempted (regardless of `/dev/tty` write result), treat as success (OSC 52 may have worked silently).
4. If external command execution fails, report error.
5. If `/dev/tty` open fails and no external command is available, report error (`Copy failed: no clipboard method available`).

## Result Handling in Model

In `handleContextMenuResult()`:

- `copy_name`: Get `entry.Name`, call clipboard write, set status message.
- `copy_path`: Get `filepath.Join(activePane.Path(), entry.Name)`, call clipboard write, set status message.

No pane refresh is needed (no filesystem change).

## Test Scenarios

### Unit Tests
- [ ] `copy_name` menu item exists in context menu
- [ ] `copy_path` menu item exists in context menu
- [ ] Both items disabled when entry is parent directory
- [ ] Both items disabled when marked files exist
- [ ] Both items enabled for regular files
- [ ] Both items enabled for directories (non-parent)
- [ ] OSC 52 escape sequence is correctly formatted
- [ ] Base64 encoding is correct for ASCII file names
- [ ] Base64 encoding is correct for Unicode file names
- [ ] External command detection finds `wl-copy` first
- [ ] External command detection falls back to `xclip`
- [ ] External command detection falls back to `xsel`
- [ ] Error returned when no clipboard method available and external command fails

### Integration Tests
- [ ] Selecting "Copy file name" from context menu sets status message
- [ ] Selecting "Copy full path" from context menu sets status message
- [ ] Error case: status bar shows error message on clipboard failure

### E2E Tests
- [ ] Open context menu, select "Copy file name", verify status bar message
- [ ] Open context menu, select "Copy full path", verify status bar message

## Success Criteria

- [ ] Both context menu items appear and function correctly
- [ ] OSC 52 escape sequence is emitted
- [ ] External command fallback works when clipboard tools are available
- [ ] Status bar feedback is displayed on success and failure
- [ ] Parent directory and marked files correctly disable the items
- [ ] All unit tests pass
- [ ] No regression in existing context menu functionality

## Dependencies

**Internal:**
- `internal/ui/context_menu_dialog.go` - Menu item addition
- `internal/ui/model_update.go` - Result handling for `copy_name` and `copy_path`

**External:**
- `encoding/base64` (Go standard library) - For OSC 52
- `os/exec` (Go standard library) - For external command detection and execution

## Constraints

- Linux only (consistent with existing project scope).
- OSC 52 support depends on the terminal emulator; there is no way to detect support programmatically.

## Open Questions

None.
