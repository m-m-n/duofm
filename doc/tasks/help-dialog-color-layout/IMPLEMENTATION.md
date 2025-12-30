# Implementation Plan: Help Dialog Color Palette Layout Improvement

## Overview

Modify the help dialog's color palette rendering functions to display 4 colors per line instead of 6, preventing the dialog from stretching beyond its intended width.

## Objectives

- Fix dialog stretching issue caused by long color palette lines
- Maintain readability and `number=HEX` format

## Prerequisites

- None (isolated change to existing functions)

## Architecture Overview

This is a simple modification to two existing functions in `internal/ui/help_dialog.go`:
- `renderColorCube()`: Changes loop structure from 6 columns to 4 columns
- `renderGrayscale()`: Changes loop structure from 6 columns to 4 columns

No architectural changes required.

## Implementation Phases

### Phase 1: Modify Color Cube Rendering

**Goal**: Display color cube (16-231) with 4 colors per line

**Files to Modify**:
- `internal/ui/help_dialog.go` - `renderColorCube()` function

**Key Changes**:

| Function | Current | After |
|----------|---------|-------|
| `renderColorCube()` | 36 rows × 6 cols = 216 colors | 54 rows × 4 cols = 216 colors |

**Processing Flow**:
```
For each row (0 to 53):
  → Render 4 colors (row*4 + 16, row*4 + 17, row*4 + 18, row*4 + 19)
  → Append line to result
```

**Implementation Steps**:
1. Change outer loop from `row < 36` to `row < 54`
2. Change inner loop from `col < 6` to `col < 4`
3. Update color number calculation: `16 + row*4 + col`

**Testing**:
- Run existing tests to verify no regression
- Manual verification of output line count (should be 54 lines)

**Estimated Effort**: Small

---

### Phase 2: Modify Grayscale Rendering

**Goal**: Display grayscale (232-255) with 4 colors per line

**Files to Modify**:
- `internal/ui/help_dialog.go` - `renderGrayscale()` function

**Key Changes**:

| Function | Current | After |
|----------|---------|-------|
| `renderGrayscale()` | 4 rows × 6 cols = 24 colors | 6 rows × 4 cols = 24 colors |

**Processing Flow**:
```
For each row (0 to 5):
  → Render 4 colors (row*4 + 232, row*4 + 233, row*4 + 234, row*4 + 235)
  → Append line to result
```

**Implementation Steps**:
1. Change outer loop from `row < 4` to `row < 6`
2. Change inner loop from `col < 6` to `col < 4`
3. Update color number calculation: `232 + row*4 + col`

**Testing**:
- Run existing tests to verify no regression
- Manual verification of output line count (should be 6 lines)

**Estimated Effort**: Small

---

## File Structure

No new files. Modifications to:

```
internal/ui/
└── help_dialog.go    # Modify renderColorCube() and renderGrayscale()
```

## Testing Strategy

### Unit Tests

Existing tests cover:
- `TestHelpDialogContentHasColorPalette`: Verifies section headers exist
- `TestColorCubeToHex`: Verifies hex conversion (no changes needed)
- `TestGrayscaleToHex`: Verifies hex conversion (no changes needed)

No new tests required as logic is unchanged, only loop bounds.

### Manual Testing Checklist

- [ ] Launch duofm and press `?` to open help dialog
- [ ] Scroll to Color Palette section
- [ ] Verify color cube shows 4 colors per line
- [ ] Verify grayscale shows 4 colors per line
- [ ] Verify dialog width is not stretched
- [ ] Verify all 216 cube colors (16-231) are displayed
- [ ] Verify all 24 grayscale colors (232-255) are displayed

## Dependencies

### External Libraries
- None

### Internal Dependencies
- None (self-contained change)

## Risk Assessment

### Technical Risks
- **None identified**: This is a simple loop bound change with no logic modifications

### Implementation Risks
- **Line count verification**: Ensure all colors are still displayed
  - Mitigation: Verify 54 lines for cube, 6 lines for grayscale

## Performance Considerations

- More lines to render (54+6 vs 36+4), but negligible impact as content is pre-built once

## References

- Specification: `doc/tasks/help-dialog-color-layout/SPEC.md`
- Source file: `internal/ui/help_dialog.go:240-284`
