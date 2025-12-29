# Feature: Configuration File (Keybindings)

## Overview

Add configuration file support to duofm, allowing users to customize keybindings. The configuration file uses TOML format and is located at `~/.config/duofm/config.toml`. The application reads this file on startup and applies custom keybindings.

## Domain Rules

- Configuration file path: `~/.config/duofm/config.toml`
- If `XDG_CONFIG_HOME` is set, use `$XDG_CONFIG_HOME/duofm/config.toml`
- If the configuration file does not exist, auto-generate a default file on first startup
- Invalid settings trigger a warning and application continues with default values
- Duplicate key assignments trigger a warning; last definition wins

## Objectives

- Allow users to customize all keybindings via TOML configuration file
- Comply with XDG Base Directory specification
- Provide seamless first-run experience with auto-generated configuration file
- Handle configuration errors gracefully without crashing

## User Stories

- As a user, I want to use Emacs-style keybindings (Ctrl+n/Ctrl+p) instead of vim-style
- As a user, I want to assign memorable keys to frequently used operations
- As a user, I want to disable unused features to prevent accidental keystrokes
- As a user, I want to see commented examples in the auto-generated config file

## Functional Requirements

### FR-1: Configuration File Loading

The application loads the TOML configuration file on startup.

- FR-1.1: Read `~/.config/duofm/config.toml` on application startup
- FR-1.2: Prioritize `$XDG_CONFIG_HOME/duofm/config.toml` if the environment variable is set
- FR-1.3: If file does not exist, proceed to FR-2 (auto-generation)

### FR-2: Configuration File Auto-Generation

Auto-generate a default configuration file when none exists.

- FR-2.1: Create `~/.config/duofm/` directory if it does not exist
- FR-2.2: Generate `config.toml` with all default keybindings
- FR-2.3: Include comments explaining each action and its default keys
- FR-2.4: Group keybindings into logical sections (navigation, file operations, etc.)

### FR-3: Keybinding Configuration

All keybindings are customizable via the configuration file.

- FR-3.1: Define keybindings in `[keybindings]` section
- FR-3.2: Use action name as key, array of key strings as value
- FR-3.3: Support multiple keys per action (e.g., `move_down = ["j", "down"]`)
- FR-3.4: Use default value for undefined actions

### FR-4: Error Handling

Handle configuration errors gracefully.

- FR-4.1: On parse error, show warning and use default configuration
- FR-4.2: On invalid key name, show warning and use default for that action
- FR-4.3: On invalid action name, show warning and ignore
- FR-4.4: On duplicate key assignment, show warning and apply last definition

### FR-5: Modifier Key Support

Support keybindings with modifier keys.

- FR-5.1: Single alphabet keys: `"J"`, `"N"` (uppercase)
- FR-5.2: Symbol keys: `"?"`, `"@"`, `"!"`, `"~"` (use resulting character)
- FR-5.3: Special keys: `"Enter"`, `"Esc"`, `"Space"`, `"Tab"`, `"Backspace"` (PascalCase)
- FR-5.4: Function keys: `"F5"`, `"F1"`
- FR-5.5: Arrow keys: `"Up"`, `"Down"`, `"Left"`, `"Right"`
- FR-5.6: Ctrl modifier: `"Ctrl+H"`, `"Ctrl+R"`, `"Ctrl+="`
- FR-5.7: Shift modifier: `"Shift+N"` (for uppercase N with modifier key held)
- FR-5.8: Multiple modifiers: `"Ctrl+Shift+N"`, `"Alt+Shift+X"`

**Key Format Principle**: Use the resulting character, not the physical key combination.
- Symbol keys are written as the symbol itself: `"?"` not `"Shift+/"`
- This is keyboard layout independent

### FR-6: Action Disabling

Allow disabling specific actions.

- FR-6.1: Empty array disables the action: `help = []`
- FR-6.2: Disabled actions do not respond to any key press

### FR-7: Help Dialog Key Format Consistency

Help dialog must use the same key format as configuration file.

- FR-7.1: All keys in help dialog use PascalCase format
- FR-7.2: Key display matches configuration file format exactly
- FR-7.3: Examples: `J/K` instead of `j/k`, `Ctrl+H` instead of `ctrl+h`

## Non-Functional Requirements

- NFR-1.1: Configuration file is read only once at startup
- NFR-1.2: Loading completes within 100ms
- NFR-2.1: Generated default config file is under 100 lines
- NFR-2.2: Config file includes section headers and comments for readability
- NFR-3.1: Application works correctly without configuration file (backward compatible)

## Interface Contract

### Configuration File Format

```toml
# duofm configuration file
# Customize keybindings by modifying the values below
# Key format: Uppercase letters, symbols as-is, PascalCase for special keys
# Example: "J", "?", "Enter", "Ctrl+H"

[keybindings]
# Navigation
move_down = ["J", "Down"]
move_up = ["K", "Up"]
move_left = ["H", "Left"]
move_right = ["L", "Right"]
enter = ["Enter"]

# File operations
copy = ["C"]
move = ["M"]
delete = ["D"]
rename = ["R"]
new_file = ["N"]
new_directory = ["Shift+N"]
mark = ["Space"]

# Display
toggle_info = ["I"]
toggle_hidden = ["Ctrl+H"]
sort = ["S"]
help = ["?"]

# Navigation extended
home = ["~"]
prev_dir = ["-"]
refresh = ["F5", "Ctrl+R"]
sync_pane = ["="]

# Search
search = ["/"]
regex_search = ["Ctrl+F"]

# External applications
view = ["V"]
edit = ["E"]
shell_command = ["!"]
context_menu = ["@"]

# Application
quit = ["Q"]
escape = ["Esc"]
```

### Supported Actions

| Action | Description | Default Keys |
|--------|-------------|--------------|
| `move_down` | Move cursor down | `["J", "Down"]` |
| `move_up` | Move cursor up | `["K", "Up"]` |
| `move_left` | Move to left pane / parent directory | `["H", "Left"]` |
| `move_right` | Move to right pane / enter directory | `["L", "Right"]` |
| `enter` | Enter directory / open file | `["Enter"]` |
| `copy` | Copy selected file(s) | `["C"]` |
| `move` | Move selected file(s) | `["M"]` |
| `delete` | Delete selected file(s) | `["D"]` |
| `rename` | Rename selected file | `["R"]` |
| `new_file` | Create new file | `["N"]` |
| `new_directory` | Create new directory | `["Shift+N"]` |
| `mark` | Toggle file mark | `["Space"]` |
| `toggle_info` | Toggle info display mode | `["I"]` |
| `toggle_hidden` | Toggle hidden files visibility | `["Ctrl+H"]` |
| `sort` | Open sort dialog | `["S"]` |
| `help` | Show help | `["?"]` |
| `home` | Go to home directory | `["~"]` |
| `prev_dir` | Go to previous directory | `["-"]` |
| `refresh` | Refresh view | `["F5", "Ctrl+R"]` |
| `sync_pane` | Synchronize panes | `["="]` |
| `search` | Incremental search | `["/"]` |
| `regex_search` | Regex search | `["Ctrl+F"]` |
| `view` | View file with pager | `["V"]` |
| `edit` | Edit file with editor | `["E"]` |
| `shell_command` | Execute shell command | `["!"]` |
| `context_menu` | Open context menu | `["@"]` |
| `quit` | Quit application | `["Q"]` |
| `escape` | Cancel / escape | `["Esc"]` |

### Supported Key Formats

Keys are written as the resulting character (keyboard layout independent).

| Format | Examples | Notes |
|--------|----------|-------|
| Alphabet (uppercase) | `"J"`, `"K"`, `"N"` | Always uppercase |
| Symbols | `"?"`, `"@"`, `"!"`, `"~"`, `"/"`, `"-"`, `"="` | Use resulting character |
| Special keys | `"Enter"`, `"Esc"`, `"Space"`, `"Tab"`, `"Backspace"` | PascalCase |
| Arrow keys | `"Up"`, `"Down"`, `"Left"`, `"Right"` | PascalCase |
| Function keys | `"F1"` through `"F12"` | Uppercase F |
| Ctrl modifier | `"Ctrl+H"`, `"Ctrl+R"`, `"Ctrl+="` | With resulting character |
| Shift+alphabet | `"Shift+N"` | For uppercase with Shift held |
| Alt modifier | `"Alt+X"`, `"Alt+Enter"` | |
| Multiple modifiers | `"Ctrl+Shift+N"`, `"Alt+Shift+X"` | |

### Error Conditions

| Condition | Behavior |
|-----------|----------|
| File does not exist | Auto-generate default config file |
| TOML parse error | Warning in status bar, use defaults |
| Unknown action name | Warning in status bar, ignore entry |
| Invalid key format | Warning in status bar, use default for action |
| Duplicate key mapping | Warning in status bar, last definition wins |

### Warning Message Examples

- `Warning: config parse error at line 15, using defaults`
- `Warning: unknown action "foobar" in config, ignored`
- `Warning: invalid key "Ctrl++" in config, using default for copy`
- `Warning: key "J" assigned to both move_down and custom_action`

## Test Scenarios

### Unit Tests (config package)

- [ ] LoadConfig returns default values when file does not exist
- [ ] LoadConfig correctly parses valid TOML configuration
- [ ] LoadConfig handles missing keybindings section gracefully
- [ ] LoadConfig handles empty keybindings array (action disabled)
- [ ] LoadConfig handles single key string conversion to array
- [ ] ParseKey correctly parses single character keys
- [ ] ParseKey correctly parses ctrl modifier keys
- [ ] ParseKey correctly parses function keys
- [ ] ParseKey returns error for invalid key format
- [ ] GenerateDefaultConfig creates valid TOML file
- [ ] GetConfigPath returns XDG_CONFIG_HOME path when set
- [ ] GetConfigPath returns ~/.config/duofm path as fallback
- [ ] ValidateKeybindings detects duplicate key assignments
- [ ] ValidateKeybindings detects unknown action names

### Integration Tests

- [ ] Application starts with no config file and generates default
- [ ] Application loads custom keybindings from config file
- [ ] Custom keybindings override default keybindings
- [ ] Invalid config results in warning and default keybindings
- [ ] Disabled action (empty array) does not respond to key press

### E2E Tests

- [ ] First launch creates ~/.config/duofm/config.toml
- [ ] Custom move_down = ["Ctrl+N"] works correctly
- [ ] Multiple keys for same action both work (e.g., refresh with F5 and Ctrl+R)
- [ ] Disabled action (help = []) does not show help dialog
- [ ] Modified config takes effect after restart

## Success Criteria

- [ ] Configuration file is read from `~/.config/duofm/config.toml`
- [ ] `XDG_CONFIG_HOME` environment variable is respected
- [ ] Default config file is auto-generated on first run
- [ ] Generated config file includes commented examples
- [ ] All actions can be remapped via `[keybindings]` section
- [ ] Multiple keys can be assigned to single action
- [ ] Actions can be disabled with empty array
- [ ] Modifier keys (ctrl, shift) work correctly
- [ ] Parse errors show warning and fall back to defaults
- [ ] Duplicate key warnings are displayed
- [ ] Application works without config file (backward compatible)
- [ ] Existing features work correctly after implementation
- [ ] All unit tests pass
- [ ] All E2E tests pass

## Dependencies

- TOML parsing library (BurntSushi/toml or pelletier/go-toml)
- Existing keybinding infrastructure in `internal/ui/keys.go`
- Existing model key handling in `internal/ui/model.go`

## Constraints

- Configuration changes require application restart to take effect
- Key sequences (e.g., "gg" for top) are not supported in this implementation
- Only keybindings are configurable; other settings (colors, behavior) are out of scope
