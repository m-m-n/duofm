# Implementation Plan: East Asian Width Configuration

## Overview

Configure `go-runewidth` to treat Unicode Ambiguous Width characters as width 1, matching modern terminal behavior.

## Implementation Status: ✅ Complete

This is a minimal change that has already been implemented.

## Implementation Details

### Phase 1: EastAsianWidth Configuration ✅

**Goal**: Set EastAsianWidth to false at application startup

**File Modified**:
- `cmd/duofm/main.go` - Add runewidth configuration before program initialization

**Change**:
```go
import (
    // ... existing imports
    "github.com/mattn/go-runewidth"
)

func main() {
    // Ambiguous幅文字（☆、ü、①など）を幅1として扱う
    // 多くのモダンターミナルの実際の表示に合わせる設定
    // TODO: 将来的には設定ファイルで変更可能にする
    runewidth.DefaultCondition.EastAsianWidth = false

    p := tea.NewProgram(
        // ...
    )
}
```

**Effort**: Minimal (1 line of code + import)

## Testing

### Manual Testing Checklist
- [x] File with `☆` in name displays correctly aligned (Alacritty)
- [x] File with `ü` in name displays correctly aligned (Alacritty)
- [x] Tested on Tabby terminal
- [x] Existing Japanese file names still display correctly
- [x] All existing tests pass

### Automated Tests
- Existing `TestDisplayNameWithLimit` tests continue to pass
- No additional tests required (configuration change only)

## Dependencies

- `github.com/mattn/go-runewidth` - Already in use, no version change needed

## Verification

```bash
# Build
go build ./...

# Test
go test ./...

# Manual verification
./duofm  # Navigate to directory with ☆, ü, ① in filenames
```
