# Verification Checklist: Clipboard Copy

## Overview

**Feature**: Clipboard Copy
**SPEC.md**: `doc/tasks/clipboard-copy/SPEC.md`
**IMPLEMENTATION.md**: `doc/tasks/clipboard-copy/IMPLEMENTATION.md`
**Date**: 2026-02-05
**Status**: Implementation Complete
**All Tests**: PASS

## Automated Verification

### Build Status

- [x] `go build ./...` completes without errors
- [x] `go build -o /tmp/duofm ./cmd/duofm` produces binary

### Test Execution

- [x] `go test ./...` all tests pass
- [x] `go test -cover ./internal/clipboard/...` coverage 81.5%
- [x] `go test -cover ./internal/ui/...` coverage 76.3%

### Code Quality

- [x] `gofmt -l .` returns no unformatted files
- [x] `go vet ./...` reports no issues

## File Structure Verification

### Files Created

- [x] `internal/clipboard/clipboard.go` - OSC 52 + external command clipboard write (96 lines)
- [x] `internal/clipboard/clipboard_test.go` - Clipboard module unit tests (300 lines)
- [x] `internal/ui/model_clipboard_test.go` - Model integration tests (176 lines)

### Files Modified

- [x] `internal/ui/context_menu_dialog.go` - Added `buildClipboardMenuItems` method and integration into `buildMenuItems`
- [x] `internal/ui/context_menu_dialog_test.go` - Added 7 clipboard tests, updated existing item counts
- [x] `internal/ui/model_update.go` - Added `copy_name`/`copy_path` handling, `clipboardResultMsg` handling, `clipboardWriteCmd`
- [x] `internal/ui/messages.go` - Added `clipboardResultMsg` type

## SPEC.md Compliance

### Functional Requirements Coverage

| ID | Requirement | Phase | Status |
|----|-------------|-------|--------|
| FR1 | Add "Copy file name" menu item with action ID `copy_name` | Phase 2 | PASS |
| FR2 | Add "Copy full path" menu item with action ID `copy_path` | Phase 2 | PASS |
| FR3 | `copy_name` copies the file name to clipboard | Phase 1, 3 | PASS |
| FR4 | `copy_path` copies the absolute path to clipboard | Phase 1, 3 | PASS |
| FR5 | Both items disabled when cursor is on `..` | Phase 2 | PASS |
| FR6 | Both items disabled when marked files exist | Phase 2 | PASS |
| FR7 | Clipboard write uses OSC 52 as primary method | Phase 1 | PASS |
| FR8 | Fallback to external commands: wl-copy, xclip, xsel | Phase 1 | PASS |
| FR9 | Success: display `Copied: {text}` for 3 seconds | Phase 3 | PASS |
| FR10 | Failure: display `Copy failed: {error}` for 3 seconds | Phase 3 | PASS |

### Non-Functional Requirements Coverage

| ID | Requirement | How to Verify |
|----|-------------|---------------|
| NFR1 | Clipboard copy completes within 100ms | Clipboard write is async via `tea.Cmd`; UI responds instantly |
| NFR2 | Works on Linux with OSC 52 terminals and fallback commands | Manual testing on target environments |

### Success Criteria from SPEC.md

| ID | Criterion | Status |
|----|-----------|--------|
| SC-1 | Both context menu items appear and function correctly | PASS |
| SC-2 | OSC 52 escape sequence is emitted | PASS |
| SC-3 | External command fallback works when clipboard tools are available | PASS |
| SC-4 | Status bar feedback is displayed on success and failure | PASS |
| SC-5 | Parent directory and marked files correctly disable the items | PASS |
| SC-6 | All unit tests pass | PASS |
| SC-7 | No regression in existing context menu functionality | PASS |

## Unit Test Scenarios from SPEC.md

| ID | Scenario | Status |
|----|----------|--------|
| TS-1 | `copy_name` menu item exists in context menu | PASS |
| TS-2 | `copy_path` menu item exists in context menu | PASS |
| TS-3 | Both items disabled when entry is parent directory | PASS |
| TS-4 | Both items disabled when marked files exist | PASS |
| TS-5 | Both items enabled for regular files | PASS |
| TS-6 | Both items enabled for directories (non-parent) | PASS |
| TS-7 | OSC 52 escape sequence is correctly formatted | PASS |
| TS-8 | Base64 encoding is correct for ASCII file names | PASS |
| TS-9 | Base64 encoding is correct for Unicode file names | PASS |
| TS-10 | External command detection finds `wl-copy` first | PASS |
| TS-11 | External command detection falls back to `xclip` | PASS |
| TS-12 | External command detection falls back to `xsel` | PASS |
| TS-13 | Error returned when external command fails | PASS |
| TS-14 | External command timeout after 5 seconds | PASS |
| TS-15 | OSC 52 attempted + no external cmd = success | PASS |
| TS-16 | /dev/tty open failed + no external cmd = error | PASS |

## Integration Test Scenarios from SPEC.md

| ID | Scenario | Status |
|----|----------|--------|
| IT-1 | Selecting "Copy file name" from context menu sets status message | PASS |
| IT-2 | Selecting "Copy full path" from context menu sets status message | PASS |
| IT-3 | Error case: status bar shows error message on clipboard failure | PASS |

## Test Coverage

### Clipboard Module (`internal/clipboard/`)

| Function | Coverage |
|----------|----------|
| `buildOSC52Sequence` | 100.0% |
| `writeOSC52` | 100.0% |
| `findClipboardCommand` | 80.0% |
| `execClipboardCommand` | 100.0% |
| `WriteToClipboard` | 71.4% |
| **Total** | **81.5%** |

### File Size Check

| File | Lines | Status |
|------|-------|--------|
| `internal/clipboard/clipboard.go` | 96 | OK |
| `internal/ui/context_menu_dialog.go` | 566 | OK |
| `internal/ui/model_update.go` | 486 | OK |
| `internal/ui/messages.go` | 165 | OK |

All implementation files are within the 1000-line limit.

## E2E Testing (Docker)

### Setup

- Run: `make test-e2e`

### Test Scenarios

- [ ] Open context menu on a file, select "Copy file name", verify status bar shows `Copied: {filename}`
- [ ] Open context menu on a file, select "Copy full path", verify status bar shows `Copied: {path}`
- [ ] Open context menu on parent directory `..`, verify both clipboard items are grayed out
- [ ] Mark files, open context menu, verify both clipboard items are grayed out

## Manual Testing (E2E Not Possible)

### Items Requiring Human Judgment

- [ ] OSC 52-capable terminal: verify text is actually in system clipboard after "Copy file name"
- [ ] OSC 52-capable terminal: verify text is actually in system clipboard after "Copy full path"
- [ ] With xclip installed: verify clipboard content via `xclip -o -selection clipboard`
- [ ] With wl-copy installed (Wayland): verify clipboard content via `wl-paste`
- [ ] No clipboard tools available: verify no error displayed (OSC 52 best-effort)
- [ ] Status bar message clears after 3 seconds
- [ ] Copy operation feels instantaneous (< 100ms perceived)

## Regression Verification

### Existing Context Menu Tests

- [x] `TestNewContextMenuDialog` passes with updated item counts (6->8, 8->10)
- [x] `TestBuildMenuItems_RegularFile` passes with updated expected IDs
- [x] `TestBuildMenuItems_Symlink` passes with updated item count (8->10)
- [x] `TestBuildMenuItems_BrokenSymlink` passes
- [x] `TestUpdate_NavigationJK` passes with updated wrap positions (5->7)
- [x] `TestUpdate_NavigationNumeric` passes with updated expectations
- [x] `TestUpdate_NumericKey_ActionID` passes with updated expected IDs
- [x] `TestGetCurrentPageItems` passes with updated item count (6->8)
- [x] All other existing context menu tests pass
- [x] All existing UI tests pass
- [x] All other module tests pass (`fs`, `config`, `archive`, `filter`, `version`)

## Verification Summary

| Category | Items | Status |
|----------|-------|--------|
| Build | 2 | All PASS |
| Tests | 3 | All PASS |
| Code Quality | 2 | All PASS |
| File Structure | 7 | All PASS |
| SPEC Compliance (FR) | 10 | All PASS |
| Unit Test Scenarios | 16 | All PASS |
| Integration Test Scenarios | 3 | All PASS |
| E2E Testing | 4 | Pending |
| Manual Testing | 7 | Pending |
| Regression | 11 | All PASS |

**Automated**: 54 items, all PASS
**Pending**: 4 E2E items, 7 manual items
