# Feature: Configurable Enter Key Behavior

## Overview

This feature allows users to configure the behavior of the Enter key when opening files in duofm. Currently, pressing Enter on a file opens it with less (pager), which is the same behavior as the V key. This feature separates the Enter key behavior from V and allows configuration through the config file.

## Objectives

- Separate Enter key behavior from V key (view) behavior
- Allow users to specify the application to open files with Enter key via config file
- Support three modes: `less` (default), `xdg-open`, and custom path
- Automatically determine execution mode (foreground/background) based on setting

## User Stories

### US1: Open File with Default Pager
As a user, I want to press Enter on a file to open it with less (default), so that I can quickly view file contents.

**Acceptance Criteria:**
- [ ] When `enter_behavior` is not set, pressing Enter opens file with less
- [ ] When `enter_behavior = "less"`, pressing Enter opens file with less
- [ ] duofm pauses and resumes after less exits

### US2: Open File with System Default Application
As a user, I want to press Enter to open files with xdg-open, so that files open in their associated applications.

**Acceptance Criteria:**
- [ ] When `enter_behavior = "xdg-open"`, pressing Enter launches xdg-open
- [ ] xdg-open runs in background (duofm remains interactive)
- [ ] Error is shown if xdg-open fails to start

### US3: Open File with Custom Application
As a user, I want to specify a custom application to open files, so that I can use my preferred viewer.

**Acceptance Criteria:**
- [ ] When `enter_behavior = "path:/usr/bin/vim"`, pressing Enter opens file with vim
- [ ] Custom application runs in foreground (duofm pauses)
- [ ] Error is shown if the specified path doesn't exist or isn't executable

## Technical Requirements

### Functional Requirements
- **FR1:** Add `enter_behavior` configuration option to config.toml
- **FR2:** Parse configuration value and determine behavior type
- **FR3:** Modify `handleEnter()` to use configured behavior
- **FR4:** Support auto-merge for missing `enter_behavior` in existing config files
- **FR5:** Show warning for invalid configuration values and fall back to default

### Non-Functional Requirements
- **NFR1 - Performance:** Configuration parsing occurs once at startup
- **NFR2 - Compatibility:** Existing config files without `enter_behavior` work as before (less)
- **NFR3 - Usability:** Invalid config values produce clear warning messages

## Implementation Approach

### Architecture

**Component Structure:**
```
cmd/duofm/
└── main.go        # Pass EnterBehavior to UI

internal/config/
├── config.go      # Add EnterBehavior to Config struct
├── defaults.go    # Add DefaultEnterBehavior()
├── enter.go       # [NEW] EnterBehavior type and parsing
├── enter_test.go  # [NEW] Tests for enter behavior
├── merger.go      # Update to support enter_behavior
└── generator.go   # Update default config template

internal/ui/
├── model.go       # Add enterBehavior field to Model
├── exec.go        # Add openWithCustomForeground()
└── model_update_keyboard.go  # Modify handleEnter()
```

**Data Flow:**
```
config.toml → LoadConfig() → Config.EnterBehavior → Model.enterBehavior
                                                           ↓
User presses Enter → handleEnter() → Check enterBehavior type
                                           ↓
                    ┌──────────────────────┼──────────────────────┐
                    ↓                      ↓                      ↓
              EnterBehaviorLess    EnterBehaviorXDGOpen    EnterBehaviorCustom
                    ↓                      ↓                      ↓
              openWithViewer()       openWithXDG()      openWithCustomForeground()
              (foreground)          (background)         (foreground)
```

### Data Structures

#### EnterBehaviorType

```go
// EnterBehaviorType represents the type of enter key behavior.
type EnterBehaviorType int

const (
    EnterBehaviorLess EnterBehaviorType = iota
    EnterBehaviorXDGOpen
    EnterBehaviorCustom
)
```

#### EnterBehavior

```go
// EnterBehavior represents the configured enter key behavior.
type EnterBehavior struct {
    Type       EnterBehaviorType
    CustomPath string // Only used when Type == EnterBehaviorCustom
}
```

**注意:** EnterBehaviorはゼロ値がデフォルト（less）として機能するため、非ポインタ型として扱う。

#### Config Update

```go
type Config struct {
    Keybindings   map[string][]string `toml:"keybindings"`
    Colors        *ColorConfig
    HistoryLimit  int           `toml:"history_limit"`
    EnterBehavior EnterBehavior // New field (non-pointer, uses zero value as default)
}
```

### Configuration Format

#### config.toml Addition

```toml
# Enter key behavior when opening files
# Options:
#   "less"     - Open with pager (foreground, default)
#   "xdg-open" - Open with system default app (background)
#   "path:/path/to/app" - Open with custom app (foreground)
enter_behavior = "less"
```

### API Design

#### New Functions in internal/config/enter.go

```go
// ParseEnterBehavior parses the enter_behavior config value.
// Returns the parsed EnterBehavior and any warning message.
// Valid values: "less", "xdg-open", "path:/path/to/app"
// Invalid values return default (less) with a warning.
//
// Processing:
//   - Input is trimmed (strings.TrimSpace) before parsing
//   - For "path:" format, PATH existence is NOT validated here
//   - Validation occurs at runtime in openWithCustomForeground() using exec.LookPath()
//
// Warning messages are used as logical identifiers:
//   - "invalid enter_behavior value '...', using default" - unknown value
//   - "empty path in enter_behavior, using default" - path: with no path
func ParseEnterBehavior(value string) (EnterBehavior, string)

// DefaultEnterBehavior returns the default enter behavior (less).
func DefaultEnterBehavior() EnterBehavior

// String returns the string representation of EnterBehavior.
func (e EnterBehavior) String() string
```

#### New Function in internal/ui/exec.go

```go
// openWithCustomForeground opens a file with a custom application in foreground.
// The application path should be an absolute path or available in PATH.
// This function validates the application path at runtime using exec.LookPath().
// If the application is not found or not executable, an error is returned.
func openWithCustomForeground(application, file, workDir string) tea.Cmd
```

### File Structure Changes

```
cmd/
└── duofm/
    └── main.go             # Modified: Pass EnterBehavior to UI

internal/
├── config/
│   ├── config.go           # Modified: Add EnterBehavior to Config
│   ├── defaults.go         # Modified: Add DefaultEnterBehavior()
│   ├── enter.go            # NEW: EnterBehavior type and parsing
│   ├── enter_test.go       # NEW: Tests
│   ├── generator.go        # Modified: Update default config template
│   └── merger.go           # Modified: Support enter_behavior merge
└── ui/
    ├── model.go            # Modified: Store enterBehavior in Model
    ├── exec.go             # Modified: Add openWithCustomForeground()
    └── model_update_keyboard.go  # Modified: Update handleEnter()
```

## Test Scenarios

### Unit Tests

#### config/enter_test.go

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| ParseEnterBehavior_Less | `"less"` | Type=EnterBehaviorLess, warning="" |
| ParseEnterBehavior_XDGOpen | `"xdg-open"` | Type=EnterBehaviorXDGOpen, warning="" |
| ParseEnterBehavior_CustomPath | `"path:/usr/bin/vim"` | Type=EnterBehaviorCustom, CustomPath="/usr/bin/vim" |
| ParseEnterBehavior_EmptyString | `""` | Type=EnterBehaviorLess, warning contains "invalid" |
| ParseEnterBehavior_Whitespace | `"  less  "` | Type=EnterBehaviorLess, warning="" (TrimSpace適用) |
| ParseEnterBehavior_InvalidValue | `"unknown"` | Type=EnterBehaviorLess, warning contains "invalid" |
| ParseEnterBehavior_PathOnly | `"path:"` | Type=EnterBehaviorLess, warning contains "empty" |
| ParseEnterBehavior_PathWithSpaces | `"path:/path/to/my app"` | Type=EnterBehaviorCustom, CustomPath="/path/to/my app" |

#### config/merger_test.go (additions)

| Test Case | Description |
|-----------|-------------|
| MergeConfig_MissingEnterBehavior | Verify enter_behavior is added to existing config |
| MergeConfig_ExistingEnterBehavior | Verify existing enter_behavior is preserved |

### Integration Tests

| Test Case | Description |
|-----------|-------------|
| LoadConfig_WithEnterBehavior | Load config with enter_behavior and verify parsing |
| LoadConfig_WithoutEnterBehavior | Load config without enter_behavior, verify default |

### E2E Tests

| Test Case | Description |
|-----------|-------------|
| Enter_DefaultBehavior | Press Enter on file, verify less-like behavior |
| Enter_WithCustomConfig | Test with xdg-open config (verify no TUI freeze) |

### Edge Cases
- [ ] Empty config file: Default behavior (less) is used
- [ ] Invalid value format: Warning shown, default used
- [ ] `path:` with empty path: Warning shown, default used
- [ ] Path with spaces: Path is correctly parsed and used
- [ ] Non-existent custom path: Error shown when Enter is pressed

### Regression Tests
- [ ] V key still opens with less (unchanged)
- [ ] E key still opens with editor (unchanged)
- [ ] Enter on directory still navigates into it (unchanged)
- [ ] Enter on parent directory (..) still navigates up (unchanged)

## Security Considerations

- **Input Validation:** `path:` values are validated to be non-empty
- **Path Safety:** Custom paths are used as-is (user responsibility)
- **No Shell Injection:** File paths are passed as arguments, not through shell

## Error Handling

### Error Codes

警告・エラーメッセージは論理識別子として機能し、テストでは文字列の部分一致で検証する。

| Identifier | Description | Message Pattern |
|------------|-------------|-----------------|
| WARN_INVALID_ENTER | Invalid enter_behavior value | 文字列に "invalid" を含む |
| WARN_EMPTY_PATH | Empty path in path: format | 文字列に "empty" を含む |
| ERR_EXEC_FAILED | Failed to execute application | "Cannot open file: {error}" |
| ERR_NOT_FOUND | Custom application not found | "Cannot open file: executable not found" |

**実装詳細:**
- 警告は文字列として返却（専用のエラー型は使用しない）
- テストでは `strings.Contains(warning, "invalid")` などで検証

### Error Flow

```
Invalid Config Value → Log Warning → Use Default (less) → Continue

Execution Error → Show Status Message → Clear After 5 Seconds
```

## Performance Optimization

### Performance Goals
- Configuration parsing: < 1ms additional overhead
- Enter key response: No measurable difference from current implementation

### Implementation Notes
- Parse `enter_behavior` once during `LoadConfig()`
- Store parsed `EnterBehavior` in `Model` struct
- `handleEnter()` performs simple switch on enum type

## Success Criteria

- [ ] All functional requirements are implemented and tested
- [ ] All test scenarios pass
- [ ] Backward compatibility: Existing configs work without changes
- [ ] Default behavior unchanged when `enter_behavior` is not set
- [ ] V key behavior unchanged
- [ ] Code review completed
- [ ] Documentation updated in config template

## Open Questions

- [x] Should `path:` support relative paths? → No, only absolute paths or PATH commands
- [x] Default execution mode for `path:`? → Foreground (like less)

## Implementation Phases

### Phase 1: Core Implementation
**Goals:** Implement basic functionality

**Deliverables:**
- EnterBehavior type and parsing (enter.go)
- Config struct update
- handleEnter() modification
- Unit tests for parsing

### Phase 2: Config Integration
**Goals:** Integrate with config system

**Deliverables:**
- Default config template update (generator.go)
- Config merger update (merger.go)
- LoadConfig integration
- Integration tests

### Phase 3: Testing & Polish
**Goals:** Ensure quality and documentation

**Deliverables:**
- E2E tests
- Help text update (if needed)
- Edge case handling verification

## References

- Requirements Document: `doc/tasks/config-enter-behavior/要件定義書.md`
- Existing exec implementation: `internal/ui/exec.go`
- Config implementation: `internal/config/config.go`
- Config merger: `internal/config/merger.go`
- Key handler: `internal/ui/model_update_keyboard.go`
