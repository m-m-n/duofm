# Feature: East Asian Width Configuration

## Overview

Configure Unicode Ambiguous Width characters (`☆`, `ü`, `①`, `→`, etc.) to be treated as width 1, matching the actual display behavior of modern terminals.

## Objectives

- Fix display alignment issues caused by Ambiguous Width characters
- Match the calculation to actual terminal rendering behavior
- Ensure consistent column alignment in file lists

## User Stories

- As a user with files containing symbols like `☆` or `①`, I want them to display correctly aligned in the file list
- As a user with files containing European characters like `ü`, I want the columns to align properly

## Technical Requirements

### TR-1: EastAsianWidth Configuration

Set `runewidth.DefaultCondition.EastAsianWidth = false` at application startup in `cmd/duofm/main.go`.

```go
func main() {
    // Treat Ambiguous Width characters as width 1
    // Matches actual display in modern terminals
    runewidth.DefaultCondition.EastAsianWidth = false

    // ... rest of initialization
}
```

### TR-2: Affected Characters

Characters classified as "Ambiguous" in Unicode East Asian Width:

| Character | Unicode | Before (EAW=true) | After (EAW=false) |
|-----------|---------|-------------------|-------------------|
| `☆` | U+2606 | width 2 | width 1 |
| `★` | U+2605 | width 2 | width 1 |
| `ü` | U+00FC | width 2 | width 1 |
| `①` | U+2460 | width 2 | width 1 |
| `→` | U+2192 | width 2 | width 1 |
| `■` | U+25A0 | width 2 | width 1 |
| `○` | U+25CB | width 2 | width 1 |

Note: Full-width characters (Japanese, Chinese, Korean) remain width 2 as expected.

## Implementation Approach

### Files to Modify

| File | Change |
|------|--------|
| `cmd/duofm/main.go` | Add `runewidth.DefaultCondition.EastAsianWidth = false` |

### Dependencies

- `github.com/mattn/go-runewidth` - Already in use

## Test Scenarios

- [ ] File with `☆` in name displays correctly aligned
- [ ] File with `ü` in name displays correctly aligned
- [ ] Existing Japanese file names still display correctly
- [ ] Column alignment is consistent across all file types
- [ ] Tested on Alacritty terminal
- [ ] Tested on Tabby terminal

## Success Criteria

- [ ] Ambiguous Width characters treated as width 1
- [ ] File list columns properly aligned with these characters
- [ ] No regression in Japanese/Chinese/Korean character handling
- [ ] All existing tests pass

## Out of Scope

- Configuration file support for toggling this setting
- Per-terminal auto-detection
