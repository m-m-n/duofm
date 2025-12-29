# Implementation Plan: Color Theme Configuration

## Overview

Add color theme support to duofm, allowing users to customize all UI element colors through `[colors]` section in config.toml using ANSI 256-color codes (0-255). Also extend the help dialog with scrolling and a color palette reference.

## Objectives

- Enable customization of all 35 UI color settings via config.toml
- Maintain backward compatibility with existing configurations
- Extend help dialog with scrolling and color palette reference
- Centralize color definitions for maintainability

## Prerequisites

- Existing configuration file infrastructure (`internal/config`)
- lipgloss library for color rendering
- Existing UI rendering code (`internal/ui`)

## Architecture Overview

### Color System Design

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│  config.toml   │────▶│ ColorConfig  │────▶│   Theme     │
│  [colors]      │     │  (parsing)   │     │ (rendering) │
└─────────────────┘     └──────────────┘     └─────────────┘
                               │
                               ▼
                        ┌──────────────┐
                        │  Defaults    │
                        │  (fallback)  │
                        └──────────────┘
```

### Data Flow

1. `main.go` loads config including `[colors]` section
2. `ColorConfig` struct holds parsed color values with defaults
3. `Theme` struct provides lipgloss.Color accessors for UI components
4. UI components use Theme instead of hardcoded colors

---

## Implementation Phases

### Phase 1: Color Configuration Infrastructure

**Goal**: Create color configuration parsing and storage in config package

**Files to Create/Modify**:
- `internal/config/colors.go` - New file for color configuration types and defaults
- `internal/config/config.go` - Extend Config struct and LoadConfig function

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| `ColorConfig` | Holds all 35 color values as int (0-255) | None | All fields have valid default or user values |
| `DefaultColors()` | Returns default ColorConfig matching current hardcoded values | None | Returns complete ColorConfig |
| `LoadColors()` | Parses [colors] section, merges with defaults | TOML file parsed | ColorConfig with user overrides |
| `ValidateColor()` | Validates color value is in range 0-255 | Color value provided | Returns validated value or error |

**Processing Flow**:
```
LoadConfig() → Parse TOML → Extract [colors] section →
Validate each value (0-255) → Merge with defaults → Return ColorConfig
```

**Implementation Steps**:
1. Define `ColorConfig` struct with all 35 color fields as `int`
2. Implement `DefaultColors()` returning current hardcoded values
3. Extend `rawConfig` to include `Colors map[string]interface{}`
4. Implement color value validation (0-255 range check)
5. Implement merging logic: user values override defaults
6. Generate warnings for invalid/unknown color keys

**Dependencies**:
- Existing TOML parsing infrastructure

**Testing**:
- Unit tests for color parsing (valid, invalid, out-of-range)
- Unit tests for default values
- Unit tests for merging logic

**Estimated Effort**: Small

---

### Phase 2: Theme System

**Goal**: Create Theme struct that provides lipgloss colors to UI components

**Files to Create/Modify**:
- `internal/ui/theme.go` - New file for Theme struct and accessors
- `internal/ui/styles.go` - Remove hardcoded colors, use Theme

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| `Theme` | Converts ColorConfig to lipgloss.Color values | Valid ColorConfig | Ready-to-use lipgloss colors |
| `NewTheme()` | Creates Theme from ColorConfig | ColorConfig loaded | Theme instance |
| `DefaultTheme()` | Creates Theme with default colors | None | Theme with defaults |

**Processing Flow**:
```
ColorConfig → NewTheme() → Theme with lipgloss.Color fields
```

**Implementation Steps**:
1. Define `Theme` struct with lipgloss.Color fields for all 35 colors
2. Implement `NewTheme(cfg *config.ColorConfig) *Theme`
3. Implement `DefaultTheme()` for backward compatibility
4. Add global theme variable or pass theme through Model

**Dependencies**:
- Phase 1 (ColorConfig)

**Testing**:
- Unit tests for Theme creation
- Verify all colors convert correctly

**Estimated Effort**: Small

---

### Phase 3: UI Components Migration

**Goal**: Replace hardcoded colors with Theme accessors across all UI components

**Files to Modify**:
- `internal/ui/model.go` - Store Theme, pass to components
- `internal/ui/pane.go` - Use Theme for pane colors
- `internal/ui/help_dialog.go` - Use Theme for dialog colors
- `internal/ui/confirm_dialog.go` - Use Theme for dialog colors
- `internal/ui/input_dialog.go` - Use Theme for input colors
- `internal/ui/error_dialog.go` - Use Theme for error colors
- `internal/ui/sort_dialog.go` - Use Theme for dialog colors
- `internal/ui/context_menu_dialog.go` - Use Theme for menu colors
- `internal/ui/overwrite_dialog.go` - Use Theme for dialog colors
- `internal/ui/rename_input_dialog.go` - Use Theme for input colors
- `internal/ui/minibuffer.go` - Use Theme for minibuffer colors

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| `Model.theme` | Holds Theme for all UI components | Theme initialized | Theme accessible to all views |

**Processing Flow**:
```
Model initialization → Create Theme from ColorConfig →
Store in Model → Pass to View methods → Render with Theme colors
```

**Implementation Steps**:
1. Add `theme *Theme` field to Model struct
2. Initialize Theme in `NewModelWithConfig()`
3. Update `pane.go`:
   - Replace hardcoded cursor colors with `theme.CursorFg`, `theme.CursorBg`, etc.
   - Replace file type colors with `theme.DirectoryFg`, `theme.SymlinkFg`, etc.
   - Replace dimmed colors with `theme.DimmedBg`, `theme.DimmedFg`
4. Update each dialog file:
   - Replace title color with `theme.DialogTitleFg`
   - Replace border color with `theme.DialogBorderFg`
   - Replace selection colors with `theme.DialogSelectedFg/Bg`
5. Update input components:
   - Replace input colors with `theme.InputFg`, `theme.InputBg`, etc.
6. Update error dialog:
   - Replace error colors with `theme.ErrorFg`, `theme.ErrorBorderFg`
7. Remove unused color variables from `styles.go`

**Dependencies**:
- Phase 2 (Theme)

**Testing**:
- Integration test: verify application renders correctly with default theme
- Integration test: verify custom colors are applied

**Estimated Effort**: Medium

---

### Phase 4: Help Dialog Enhancement

**Goal**: Add scrolling and color palette reference to help dialog

**Files to Modify**:
- `internal/ui/help_dialog.go` - Add scrolling and color palette section

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| `HelpDialog.scrollOffset` | Tracks current scroll position | Dialog active | Valid scroll offset |
| `HelpDialog.contentHeight` | Total content height in lines | Content generated | Height calculated |
| `HelpDialog.viewHeight` | Visible area height | Dialog dimensions set | View height set |
| `colorPaletteContent()` | Generates color palette reference text | None | Formatted palette string |
| `colorToHex()` | Converts color code 16-255 to #rrggbb | Valid color code | Hex string |

**Processing Flow**:
```
User presses ? → HelpDialog created →
Generate keybindings + color palette content →
Handle j/k/Space/Shift+Space for scrolling →
Render visible portion with page indicator
```

**Implementation Steps**:
1. Add scroll state fields to HelpDialog:
   - `scrollOffset int`
   - `viewHeight int` (calculated from terminal height)
2. Update Update() to handle scroll keys:
   - `j`: scroll down 1 line
   - `k`: scroll up 1 line
   - `space`: scroll down 1 page
   - `shift+space`: scroll up 1 page
3. Generate color palette content:
   - Section header: "Color Palette Reference"
   - Colors 0-15: Show with "Terminal-dependent" label
   - Colors 16-231: Show as grid with #rrggbb values
   - Colors 232-255: Show as grayscale with #rrggbb values
4. Implement `colorToHex()` for 16-255 calculation:
   - 16-231: 6x6x6 cube formula
   - 232-255: grayscale formula
5. Render color samples using lipgloss background
6. Add page indicator `[1/3]` in corner
7. Update footer with scroll key hints

**Dependencies**:
- Phase 3 (Theme for colors)

**Testing**:
- Unit tests for colorToHex() calculations
- Unit tests for scroll position bounds
- E2E test for scroll behavior

**Estimated Effort**: Medium

---

### Phase 5: Config Generator Update

**Goal**: Include [colors] section in auto-generated config file

**Files to Modify**:
- `internal/config/generator.go` - Add colors section to template

**Implementation Steps**:
1. Add `[colors]` section to `defaultConfigTemplate`
2. Include all 35 color keys with default values
3. Add comments explaining each category and color format
4. Include note about 256-color palette reference in help (?)

**Dependencies**:
- Phase 1 (color key names defined)

**Testing**:
- Verify generated config includes all color keys
- Verify comments are readable

**Estimated Effort**: Small

---

### Phase 6: Integration and Application Startup

**Goal**: Wire everything together in main.go and Model initialization

**Files to Modify**:
- `cmd/duofm/main.go` - Pass ColorConfig to UI
- `internal/ui/model.go` - Accept ColorConfig in NewModelWithConfig

**Processing Flow**:
```
main() → LoadConfig() → Extract ColorConfig →
NewModelWithConfig(keybindings, colors, warnings) →
Model creates Theme → UI uses Theme colors
```

**Implementation Steps**:
1. Update `NewModelWithConfig()` signature to accept `*config.ColorConfig`
2. Create Theme from ColorConfig in Model initialization
3. Store color-related warnings with other config warnings
4. Display warnings in status bar on startup

**Dependencies**:
- All previous phases

**Testing**:
- Full integration test with custom config
- Verify warnings display correctly

**Estimated Effort**: Small

---

## File Structure

```
internal/
├── config/
│   ├── config.go       # Extended with ColorConfig loading
│   ├── colors.go       # NEW: ColorConfig, DefaultColors, validation
│   ├── defaults.go     # Existing keybinding defaults
│   ├── generator.go    # Extended with [colors] section
│   └── ...
├── ui/
│   ├── theme.go        # NEW: Theme struct and accessors
│   ├── styles.go       # Simplified, references Theme
│   ├── model.go        # Stores Theme, passes to components
│   ├── pane.go         # Uses Theme for colors
│   ├── help_dialog.go  # Extended with scrolling and palette
│   ├── confirm_dialog.go
│   ├── input_dialog.go
│   ├── error_dialog.go
│   ├── sort_dialog.go
│   ├── context_menu_dialog.go
│   ├── overwrite_dialog.go
│   ├── rename_input_dialog.go
│   └── minibuffer.go
└── ...
```

## Testing Strategy

### Unit Tests

**config package**:
- `colors_test.go`:
  - Test `DefaultColors()` returns all 35 keys with expected values
  - Test color validation (0, 255, -1, 256, "string")
  - Test LoadColors with missing section (uses defaults)
  - Test LoadColors with partial overrides
  - Test LoadColors with invalid values (warnings generated)

**ui package**:
- `theme_test.go`:
  - Test `NewTheme()` creates valid lipgloss colors
  - Test `DefaultTheme()` matches hardcoded values
- `help_dialog_test.go`:
  - Test `colorToHex()` for color cube (16-231)
  - Test `colorToHex()` for grayscale (232-255)
  - Test scroll bounds (min/max)
  - Test page scrolling calculations

### Integration Tests

- Application starts with no [colors] section
- Application loads custom colors correctly
- Invalid colors generate warnings but don't crash
- Theme colors propagate to all UI components

### Manual Testing Checklist

- [ ] Default config generated includes [colors] section
- [ ] Custom cursor_bg changes cursor display
- [ ] Custom directory_fg changes directory color
- [ ] All dialog borders use configured color
- [ ] Help dialog scrolls with j/k
- [ ] Help dialog page scrolls with Space/Shift+Space
- [ ] Color palette shows in help dialog
- [ ] Color samples display correctly
- [ ] Invalid color shows warning in status bar

## Dependencies

### External Libraries
- `github.com/BurntSushi/toml` - Already used for config parsing
- `github.com/charmbracelet/lipgloss` - Already used for styling

### Internal Dependencies
- Phase 1 must complete before Phase 2
- Phase 2 must complete before Phase 3
- Phase 3 must complete before Phase 4
- Phase 5 can run in parallel with Phase 3-4
- Phase 6 requires all other phases

## Risk Assessment

### Technical Risks

- **Color value type in TOML**: TOML integers may parse differently
  - Mitigation: Use `interface{}` for initial parse, then convert to int

- **Terminal color support**: Some terminals may not support 256 colors
  - Mitigation: Application already uses 256-color codes; no change needed

### Implementation Risks

- **Large number of UI file changes**: Risk of missing color references
  - Mitigation: Use grep to find all `lipgloss.Color` usages before starting

- **Help dialog scrolling complexity**: May affect dialog display logic
  - Mitigation: Implement scrolling as isolated functionality

## Performance Considerations

- Color parsing happens once at startup (NFR-1.1)
- No runtime impact on UI rendering (colors are resolved once)
- Help dialog content generated once when opened

## Security Considerations

- Color values are integers only; no string execution risk
- Invalid values fall back to defaults; no crash vectors

## Open Questions

None - all requirements are clear from specification.

## References

- Specification: `doc/tasks/color-theme/SPEC.md`
- Requirements: `doc/tasks/color-theme/要件定義書.md`
- Existing config implementation: `internal/config/`
- lipgloss documentation: https://github.com/charmbracelet/lipgloss
