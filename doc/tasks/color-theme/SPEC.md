# Feature: Color Theme Configuration

## Overview

Add color theme support to duofm, allowing users to customize all UI element colors through the configuration file (config.toml). Users can define colors in the `[colors]` section using ANSI 256-color codes (0-255).

## Domain Rules

- Color configuration is defined in `[colors]` section of config.toml
- Colors are specified as ANSI 256-color codes (integers 0-255)
- Missing color definitions fall back to default values
- Invalid color values (out of range, non-integer) trigger a warning and use default value
- Color changes require application restart to take effect

## Objectives

- Enable users to customize all UI element colors via configuration file
- Use ANSI 256-color codes for maximum terminal compatibility
- Maintain backward compatibility with existing configurations
- Provide sensible default colors matching current behavior

## User Stories

- As a user, I want to match duofm colors to my terminal theme
- As a user, I want to adjust colors for better visibility on my display
- As a user, I want to customize cursor and selection colors
- As a user, I want to differentiate file types by color

## Functional Requirements

### FR-1: Color Configuration Loading

Load color configuration from config.toml on startup.

- FR-1.1: Parse `[colors]` section from configuration file
- FR-1.2: Use default values for undefined color settings
- FR-1.3: Merge user settings with defaults (partial override supported)

### FR-2: Supported Color Format

Support ANSI 256-color codes only.

- FR-2.1: ANSI 256-color codes as integers (0-255)
- FR-2.2: Valid range is 0 to 255 inclusive
- FR-2.3: Out-of-range values trigger warning and use default
- FR-2.4: Non-integer values trigger warning and use default

### FR-3: Pane Colors

Configurable colors for pane elements.

#### Cursor (Current Row)
- FR-3.1: `cursor_fg` - Cursor row foreground color
- FR-3.2: `cursor_bg` - Cursor row background (active pane)
- FR-3.3: `cursor_bg_inactive` - Cursor row background (inactive pane)

#### Marked Rows
- FR-3.4: `mark_fg` - Marked row foreground (active pane)
- FR-3.5: `mark_fg_inactive` - Marked row foreground (inactive pane)
- FR-3.6: `mark_bg` - Marked row background (active pane)
- FR-3.7: `mark_bg_inactive` - Marked row background (inactive pane)

#### Cursor + Marked Rows
- FR-3.8: `cursor_mark_fg` - Cursor+marked row foreground
- FR-3.9: `cursor_mark_bg` - Cursor+marked row background (active pane)
- FR-3.10: `cursor_mark_bg_inactive` - Cursor+marked row background (inactive pane)

#### Path Display
- FR-3.11: `path_fg` - Path text color (active pane)
- FR-3.12: `path_fg_inactive` - Path text color (inactive pane)

#### Header (Column Names)
- FR-3.13: `header_fg` - Header color (active pane)
- FR-3.14: `header_fg_inactive` - Header color (inactive pane)

#### Pane Structure
- FR-3.15: `border_fg` - Border/separator color
- FR-3.16: `dimmed_bg` - Dimmed background color
- FR-3.17: `dimmed_fg` - Dimmed foreground color

### FR-4: File Type Colors

Configurable colors for different file types.

- FR-4.1: `directory_fg` - Directory name color
- FR-4.2: `symlink_fg` - Symbolic link color
- FR-4.3: `executable_fg` - Executable file color

### FR-5: Dialog Colors

Configurable colors for dialog elements.

- FR-5.1: `dialog_title_fg` - Dialog title color
- FR-5.2: `dialog_border_fg` - Dialog border color
- FR-5.3: `dialog_selected_fg` - Selected item foreground
- FR-5.4: `dialog_selected_bg` - Selected item background
- FR-5.5: `dialog_footer_fg` - Footer/hint text color

### FR-6: Input Field Colors

Configurable colors for input elements.

- FR-6.1: `input_fg` - Input field text color
- FR-6.2: `input_bg` - Input field background color
- FR-6.3: `input_border_fg` - Input field border color

### FR-7: Minibuffer Colors

Configurable colors for minibuffer.

- FR-7.1: `minibuffer_fg` - Minibuffer text color
- FR-7.2: `minibuffer_bg` - Minibuffer background color

### FR-8: Status and Error Colors

Configurable colors for status messages and errors.

- FR-8.1: `error_fg` - Error message color
- FR-8.2: `error_border_fg` - Error dialog border color
- FR-8.3: `warning_fg` - Warning message color
- FR-8.4: `status_fg` - Status bar text color
- FR-8.5: `status_bg` - Status bar background color

### FR-9: Default Theme

Provide default colors matching current behavior.

- FR-9.1: Default colors match current hardcoded values
- FR-9.2: Auto-generated config includes commented `[colors]` section
- FR-9.3: Default config shows all available color options

### FR-10: Help Dialog Enhancement

Extend the existing help dialog with scrolling and color palette reference.

#### Scrolling Support
- FR-10.1: Add scrolling to help dialog (j/k for line-by-line)
- FR-10.2: Space for page down, Shift+Space for page up
- FR-10.3: Display scroll position visually (page indicator)

#### Color Palette Section
- FR-10.4: Add color palette section after existing keybinding section
- FR-10.5: Colors 0-15 shown with "Terminal-dependent" label
- FR-10.6: Colors 16-231 shown as 6x6x6 color cube with #rrggbb values
- FR-10.7: Colors 232-255 shown as grayscale with #rrggbb values
- FR-10.8: Each color number displays a color sample (as background)

### FR-11: Error Handling

Handle color configuration errors gracefully.

- FR-11.1: Invalid color format triggers warning, uses default
- FR-11.2: Unknown color key triggers warning, is ignored
- FR-11.3: Parse errors in `[colors]` section don't affect other settings
- FR-11.4: Warnings displayed in status bar on startup

## Non-Functional Requirements

- NFR-1.1: Color configuration loaded once at startup
- NFR-1.2: Startup time impact under 10ms
- NFR-2.1: Application works correctly without `[colors]` section (backward compatible)
- NFR-2.2: Existing config files continue to work without modification
- NFR-3.1: Color definitions centralized in single location

## Interface Contract

### Configuration File Format

```toml
# Colors are specified as ANSI 256-color codes (0-255)
[colors]
# Pane - Cursor
cursor_fg = 15              # white
cursor_bg = 39              # blue (active pane)
cursor_bg_inactive = 240    # gray (inactive pane)

# Pane - Mark
mark_fg = 0                 # black (active pane)
mark_fg_inactive = 15       # white (inactive pane)
mark_bg = 136               # dark yellow (active pane)
mark_bg_inactive = 94       # darker yellow (inactive pane)

# Pane - Cursor + Mark
cursor_mark_fg = 15         # white
cursor_mark_bg = 30         # cyan (active pane)
cursor_mark_bg_inactive = 23 # dark cyan (inactive pane)

# Pane - Path
path_fg = 39                # blue (active pane)
path_fg_inactive = 240      # gray (inactive pane)

# Pane - Header
header_fg = 245             # light gray (active pane)
header_fg_inactive = 240    # gray (inactive pane)

# Pane - Structure
border_fg = 240             # gray
dimmed_bg = 236             # dark gray
dimmed_fg = 243             # medium gray

# File Types
directory_fg = 39           # blue
symlink_fg = 14             # cyan
executable_fg = 9           # red

# Dialog
dialog_title_fg = 39        # blue
dialog_border_fg = 39       # blue
dialog_selected_fg = 0      # black
dialog_selected_bg = 39     # blue
dialog_footer_fg = 240      # gray

# Input
input_fg = 15               # white
input_bg = 236              # dark gray
input_border_fg = 240       # gray

# Minibuffer
minibuffer_fg = 15          # white
minibuffer_bg = 236         # dark gray

# Error and Warning
error_fg = 196              # red
error_border_fg = 196       # red
warning_fg = 240            # gray

# Status Bar
status_fg = 15              # white
status_bg = 240             # gray
```

### Color Key Reference

| Category | Key | Description | Default |
|----------|-----|-------------|---------|
| Cursor | `cursor_fg` | Cursor row foreground | `15` |
| Cursor | `cursor_bg` | Cursor background (active) | `39` |
| Cursor | `cursor_bg_inactive` | Cursor background (inactive) | `240` |
| Mark | `mark_fg` | Marked row foreground (active) | `0` |
| Mark | `mark_fg_inactive` | Marked row foreground (inactive) | `15` |
| Mark | `mark_bg` | Marked row background (active) | `136` |
| Mark | `mark_bg_inactive` | Marked row background (inactive) | `94` |
| Cursor+Mark | `cursor_mark_fg` | Cursor+marked foreground | `15` |
| Cursor+Mark | `cursor_mark_bg` | Cursor+marked background (active) | `30` |
| Cursor+Mark | `cursor_mark_bg_inactive` | Cursor+marked background (inactive) | `23` |
| Path | `path_fg` | Path color (active) | `39` |
| Path | `path_fg_inactive` | Path color (inactive) | `240` |
| Header | `header_fg` | Header color (active) | `245` |
| Header | `header_fg_inactive` | Header color (inactive) | `240` |
| Pane | `border_fg` | Border color | `240` |
| Pane | `dimmed_bg` | Dimmed background | `236` |
| Pane | `dimmed_fg` | Dimmed foreground | `243` |
| File | `directory_fg` | Directory color | `39` |
| File | `symlink_fg` | Symlink color | `14` |
| File | `executable_fg` | Executable color | `9` |
| Dialog | `dialog_title_fg` | Dialog title | `39` |
| Dialog | `dialog_border_fg` | Dialog border | `39` |
| Dialog | `dialog_selected_fg` | Selected item foreground | `0` |
| Dialog | `dialog_selected_bg` | Selected item background | `39` |
| Dialog | `dialog_footer_fg` | Footer text | `240` |
| Input | `input_fg` | Input foreground | `15` |
| Input | `input_bg` | Input background | `236` |
| Input | `input_border_fg` | Input border | `240` |
| Mini | `minibuffer_fg` | Minibuffer foreground | `15` |
| Mini | `minibuffer_bg` | Minibuffer background | `236` |
| Error | `error_fg` | Error message | `196` |
| Error | `error_border_fg` | Error border | `196` |
| Warning | `warning_fg` | Warning message | `240` |
| Status | `status_fg` | Status bar foreground | `15` |
| Status | `status_bg` | Status bar background | `240` |

### Error Conditions

| Condition | Behavior |
|-----------|----------|
| No `[colors]` section | Use all default colors |
| Out-of-range value (< 0 or > 255) | Warning, use default for that key |
| Non-integer value | Warning, use default for that key |
| Unknown color key | Warning, ignore the key |

### Warning Message Examples

- `Warning: color value 300 out of range for cursor_bg, using default`
- `Warning: invalid color value "blue" for path_fg, using default`
- `Warning: unknown color key "invalid_key" in config, ignored`

### Help Dialog Display (Extended)

The help dialog now includes scrolling and a color palette section:

```
┌─ Help ────────────────────────────────────────────┐
│                                              [1/3]│
│ Keybindings                                       │
│                                                   │
│ Navigation                                        │
│   J/K/Up/Down    : move cursor down/up           │
│   H/L/Left/Right : move to left/right pane       │
│   ...                                             │
│                                                   │
│ [j/k] [Space/Shift+Space: page] [?/Esc: close]   │
└───────────────────────────────────────────────────┘
```

After scrolling to the color palette section:

```
┌─ Help ────────────────────────────────────────────┐
│                                              [3/3]│
│ Color Palette Reference                           │
│                                                   │
│ Standard Colors (0-15): Terminal-dependent        │
│ ██ 0  ██ 1  ██ 2  ██ 3  ██ 4  ██ 5  ██ 6  ██ 7  │
│ ██ 8  ██ 9  ██10  ██11  ██12  ██13  ██14  ██15  │
│                                                   │
│ 6x6x6 Color Cube (16-231):                       │
│  16=#000000  17=#00005f  18=#000087  ...         │
│                                                   │
│ Grayscale (232-255):                             │
│ 232=#080808  233=#121212  234=#1c1c1c  ...       │
│                                                   │
│ [j/k] [Space/Shift+Space: page] [?/Esc: close]   │
└───────────────────────────────────────────────────┘
```

- Color samples displayed as filled blocks with background color
- Page indicator shows current position (e.g., [1/3])
- Shows hex values for colors 16-255

## Test Scenarios

### Unit Tests (config package)

- [ ] ParseColor correctly parses ANSI 256-color codes (0-255)
- [ ] ParseColor returns error for out-of-range values (< 0, > 255)
- [ ] ParseColor returns error for non-integer values
- [ ] LoadColors returns defaults when section missing
- [ ] LoadColors merges partial settings with defaults
- [ ] LoadColors handles missing color values
- [ ] DefaultColors returns all expected keys with correct values
- [ ] GenerateDefaultConfig includes [colors] section with examples

### Unit Tests (ui package)

- [ ] Theme applies cursor colors correctly
- [ ] Theme applies mark colors correctly
- [ ] Theme applies file type colors correctly
- [ ] Theme applies dialog colors correctly
- [ ] Theme falls back to defaults for missing colors
- [ ] HelpDialog scroll position updates correctly
- [ ] HelpDialog generates correct hex values for colors 16-231
- [ ] HelpDialog generates correct hex values for colors 232-255

### Integration Tests

- [ ] Application starts with no [colors] section
- [ ] Application loads custom colors from config
- [ ] Custom colors override defaults correctly
- [ ] Invalid colors result in warning and defaults
- [ ] Multiple color keys can be customized

### E2E Tests

- [ ] Custom cursor_bg changes cursor display
- [ ] Custom directory_fg changes directory display
- [ ] Custom dialog colors apply to all dialogs
- [ ] All 256 color codes (0-255) display correctly
- [ ] Help dialog scrolls with j/k keys
- [ ] Help dialog shows color palette section

## Success Criteria

- [ ] `[colors]` section in config.toml customizes all UI colors
- [ ] ANSI 256-color codes (0-255) supported
- [ ] Missing settings use default values
- [ ] Invalid colors trigger warning and use defaults
- [ ] Backward compatible with existing config files
- [ ] All pane elements customizable
- [ ] All dialog elements customizable
- [ ] All file type colors customizable
- [ ] Generated default config includes [colors] examples
- [ ] Help dialog supports scrolling
- [ ] Help dialog includes color palette reference
- [ ] All unit tests pass
- [ ] All integration tests pass

## Dependencies

- Existing configuration file infrastructure (internal/config)
- lipgloss library for color rendering
- Existing UI rendering code (internal/ui)

## Constraints

- Color changes require application restart
- Preset themes (Gruvbox, etc.) not included
- HEX color format (#rrggbb) not supported (256-color codes only)
- Color names ("red", "blue") not supported
- Font styles (bold, italic) not configurable via this feature
