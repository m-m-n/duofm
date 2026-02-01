# Feature: MIME Type Based Enter Behavior

## Overview

This feature extends the existing `enter_behavior` configuration to support MIME type-based file opening. When `enter_behavior = "mime:"` is set, the system uses the `[enter_behavior_mime]` section to determine which application to use based on the file's MIME type.

## Objectives

- Add `enter_behavior = "mime:"` as a new configuration option
- Support `[enter_behavior_mime]` section for MIME type to command mapping
- Support wildcard matching (e.g., `text/*`, `image/*`)
- Implement command fallback when the first command fails
- Fall back to default pager ($PAGER or less) when no MIME type matches

## User Stories

### US1: Open Files Based on MIME Type
As a user, I want to configure different applications for different file types, so that images open in an image viewer and text files open in a pager.

**Acceptance Criteria:**
- [ ] When `enter_behavior = "mime:"` is set, the system checks `[enter_behavior_mime]`
- [ ] Files are opened with the command matching their MIME type
- [ ] Wildcard patterns like `image/*` match all image subtypes

### US2: Fallback When No Match
As a user, I want files without matching MIME configurations to open with the default pager, so that I can still view any file.

**Acceptance Criteria:**
- [ ] When no MIME type matches, the file opens with $PAGER or less
- [ ] Behavior is identical to when `enter_behavior` is not set

### US3: Command Fallback on Failure
As a user, I want the system to try alternative commands if the first one fails, so that I have backup viewers configured.

**Acceptance Criteria:**
- [ ] Commands are specified as arrays: `["primary", "fallback1", "fallback2"]`
- [ ] If the first command is not found, the next is tried
- [ ] If all commands fail, falls back to default pager

## Technical Requirements

### Functional Requirements
- **FR1:** Parse `enter_behavior = "mime:"` as a new behavior type
- **FR2:** Parse `[enter_behavior_mime]` section as map[string][]string
- **FR3:** Determine file MIME type using `mime.TypeByExtension()`
- **FR4:** Support exact MIME type matching (e.g., `application/pdf`)
- **FR5:** Support wildcard matching (e.g., `text/*`)
- **FR6:** Prioritize exact match over wildcard match
- **FR7:** Execute commands in foreground mode
- **FR8:** Implement command fallback when execution fails
- **FR9:** Fall back to default pager when no MIME type matches

### Non-Functional Requirements
- **NFR1 - Performance:** MIME type detection < 1ms
- **NFR2 - Compatibility:** Existing `enter_behavior` values continue to work
- **NFR3 - Usability:** Clear warning messages for invalid configurations

## Implementation Approach

### Architecture

**Component Structure:**
```
internal/config/
├── config.go        # Modified: Add MIMEBehavior to Config and rawConfig
├── enter.go         # Modified: Add EnterBehaviorMIME type
├── mime.go          # NEW: MIME behavior parsing and matching
└── mime_test.go     # NEW: Tests for MIME behavior

internal/ui/
├── exec.go          # Modified: Add openWithMIME()
└── model_update_keyboard.go  # Modified: Handle EnterBehaviorMIME
```

**Data Flow:**
```
config.toml → LoadConfig() → Config.MIMEBehavior → Model.mimeBehavior
                                                          ↓
User presses Enter on file → handleEnter() → Check enterBehavior.Type
                                                    ↓
                                              EnterBehaviorMIME
                                                    ↓
                                         GetMIMEType(filename)
                                                    ↓
                                         FindMatchingRule(mimeType)
                                                    ↓
                              ┌─────────────────────┴─────────────────────┐
                              ↓                                           ↓
                        Match found                                  No match
                              ↓                                           ↓
                    Execute commands[]                           openWithViewer()
                              ↓                                    (default pager)
                    ┌─────────┴─────────┐
                    ↓                   ↓
              Command success     Command fails
                    ↓                   ↓
                  Done            Try next command
                                        ↓
                              All failed → openWithViewer()
```

### Data Structures

#### EnterBehaviorType Addition

```go
const (
    EnterBehaviorLess EnterBehaviorType = iota
    EnterBehaviorXDGOpen
    EnterBehaviorCustom
    EnterBehaviorMIME  // NEW
)
```

#### MIMEBehaviorConfig

```go
// MIMEBehaviorConfig holds MIME type to command mappings.
type MIMEBehaviorConfig struct {
    Rules map[string][]string // MIME type pattern → command list
}
```

#### Config Update

```go
type Config struct {
    Keybindings   map[string][]string `toml:"keybindings"`
    Colors        *ColorConfig
    HistoryLimit  int                 `toml:"history_limit"`
    EnterBehavior EnterBehavior
    MIMEBehavior  MIMEBehaviorConfig  // NEW
}
```

#### rawConfig Update

```go
type rawConfig struct {
    Keybindings       map[string][]string    `toml:"keybindings"`
    Colors            map[string]interface{} `toml:"colors"`
    HistoryLimit      *int                   `toml:"history_limit"`
    EnterBehavior     *string                `toml:"enter_behavior"`
    EnterBehaviorMIME map[string][]string    `toml:"enter_behavior_mime"` // NEW
}
```

### Configuration Format

```toml
# Enter key behavior: "less", "xdg-open", "path:/path/to/app", or "mime:"
enter_behavior = "mime:"

# MIME type handlers (only used when enter_behavior = "mime:")
# Commands are tried in order; if one fails, the next is attempted.
# Use wildcard patterns like "text/*" for broad matching.
# Fallback: When no MIME type matches, opens with $PAGER or less.
[enter_behavior_mime]
"text/*" = ["less"]
"image/*" = ["feh", "xdg-open"]
"video/*" = ["mpv", "vlc"]
"audio/*" = ["mpv"]
"application/pdf" = ["zathura", "evince", "xdg-open"]
```

### API Design

#### New Functions in internal/config/mime.go

```go
// ParseMIMEBehavior parses the [enter_behavior_mime] section.
// Returns the parsed MIMEBehaviorConfig and any warnings.
func ParseMIMEBehavior(raw map[string][]string) (MIMEBehaviorConfig, []string)

// GetMIMEType returns the MIME type for a filename based on its extension.
// Returns "application/octet-stream" for unknown extensions.
func GetMIMEType(filename string) string

// FindMatchingRule finds the command list for a MIME type.
// Returns the commands and true if found, nil and false otherwise.
// Exact matches are prioritized over wildcard matches.
func (c *MIMEBehaviorConfig) FindMatchingRule(mimeType string) ([]string, bool)

// MatchesMIMEPattern checks if a MIME type matches a pattern.
// Supports exact match and wildcard patterns (e.g., "text/*").
// Note: FindMatchingRule uses optimized map lookup internally instead of this function.
// This is exported as a utility for external callers needing single-pattern matching.
func MatchesMIMEPattern(mimeType, pattern string) bool
```

#### New Function in internal/ui/exec.go

```go
// openWithMIME opens a file using MIME-type based command selection.
// Tries commands in order until one is found via LookPath.
// Falls back to pager if no command is found or no match exists.
// Returns (tea.Cmd, statusMessage). statusMessage is empty when a command is found.
// Commands may include options (e.g., "vim -R") which are split by whitespace.
func openWithMIME(filePath, workDir string, mimeCfg MIMEBehaviorConfig) (tea.Cmd, string)
```

### MIME Type Detection

Uses Go's standard `mime` package:

```go
import "mime"

func GetMIMEType(filename string) string {
    ext := filepath.Ext(filename)
    if ext == "" {
        return "application/octet-stream"
    }
    mimeType := mime.TypeByExtension(ext)
    if mimeType == "" {
        return "application/octet-stream"
    }
    // Remove parameters (e.g., "text/plain; charset=utf-8" → "text/plain")
    if idx := strings.Index(mimeType, ";"); idx != -1 {
        mimeType = strings.TrimSpace(mimeType[:idx])
    }
    return mimeType
}
```

### Matching Algorithm

```go
func (c *MIMEBehaviorConfig) FindMatchingRule(mimeType string) ([]string, bool) {
    // 1. Try exact match first
    if commands, ok := c.Rules[mimeType]; ok {
        return commands, true
    }

    // 2. Try wildcard match
    parts := strings.SplitN(mimeType, "/", 2)
    if len(parts) == 2 {
        wildcard := parts[0] + "/*"
        if commands, ok := c.Rules[wildcard]; ok {
            return commands, true
        }
    }

    return nil, false
}
```

### File Structure Changes

```
internal/
├── config/
│   ├── config.go       # Modified: Add MIMEBehavior field, update LoadConfig
│   ├── enter.go        # Modified: Add EnterBehaviorMIME constant
│   ├── mime.go         # NEW: MIME behavior parsing and matching
│   └── mime_test.go    # NEW: Tests for MIME behavior
└── ui/
    ├── exec.go         # Modified: Add openWithMIME function
    ├── model.go        # Modified: Add mimeBehavior field
    └── model_update_keyboard.go  # Modified: Handle EnterBehaviorMIME
```

## Test Scenarios

### Unit Tests

#### config/enter_test.go (additions)

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| ParseEnterBehavior_MIME | `"mime:"` | Type=EnterBehaviorMIME |

#### config/mime_test.go

| Test Case | Description |
|-----------|-------------|
| GetMIMEType_TextPlain | `.txt` → `text/plain` |
| GetMIMEType_ImagePNG | `.png` → `image/png` |
| GetMIMEType_Unknown | `.xyz` → `application/octet-stream` |
| GetMIMEType_NoExtension | `Makefile` → `application/octet-stream` |
| FindMatchingRule_ExactMatch | `application/pdf` matches `application/pdf` |
| FindMatchingRule_WildcardMatch | `image/png` matches `image/*` |
| FindMatchingRule_ExactPriority | `text/plain` exact takes priority over `text/*` |
| FindMatchingRule_NoMatch | `unknown/type` returns false |
| ParseMIMEBehavior_Valid | Valid config parses correctly |
| ParseMIMEBehavior_EmptyKey | Empty key generates warning |
| ParseMIMEBehavior_EmptyArray | Empty array generates warning |

### Integration Tests

| Test Case | Description |
|-----------|-------------|
| LoadConfig_WithMIMEBehavior | Load config with mime: and MIME section |
| LoadConfig_MIMEWithoutSection | mime: without section uses empty rules |

### E2E Tests

| Test Case | Description |
|-----------|-------------|
| Enter_MIMEText | Press Enter on .txt file, verify less opens |
| Enter_MIMEImage | Press Enter on .png file, verify configured viewer |
| Enter_MIMENoMatch | Press Enter on unknown type, verify pager opens |
| Enter_MIMECommandFallback | First command fails, second is tried |

### Edge Cases
- [ ] Empty `[enter_behavior_mime]` section: All files fall back to pager
- [ ] MIME type with parameters: Parameters are stripped before matching
- [ ] File with no extension: Falls back to pager
- [ ] All configured commands not found: Falls back to pager
- [ ] `enter_behavior = "mime:"` without MIME section: Falls back to pager

## Security Considerations

- **No Shell Execution:** Commands are executed directly via exec, not through shell
- **Path Arguments:** File paths are passed as command arguments, not interpolated
- **Command Validation:** Commands are validated via exec.LookPath before execution

## Error Handling

### Warning Messages

| Condition | Message |
|-----------|---------|
| Empty MIME type key | "empty MIME type key in enter_behavior_mime" |
| Empty command array | "empty command list for MIME type: {type}" |
| Command not found (LookPath) | Status: "Command not found: {command}, trying next..." |
| All commands failed (LookPath) | Status: "All configured commands failed, using pager" |

### Error Flow

```
Config Parse Error → Log Warning → Use Empty Rules → Fall back to pager

Command Not Found (LookPath) → Try Next Command → All Failed → Use Pager
```

## Performance Optimization

### Performance Goals
- MIME type detection: < 1ms (uses extension lookup, no file I/O)
- Pattern matching: O(n) where n is number of rules (typically < 10)

### Implementation Notes
- MIME type determined from extension only (no file content inspection)
- Rules parsed once at startup
- Pattern matching iterates through rules (acceptable for small rule sets)

## Success Criteria

- [ ] `enter_behavior = "mime:"` is recognized as valid
- [ ] `[enter_behavior_mime]` section is parsed correctly
- [ ] Exact MIME type matches work
- [ ] Wildcard patterns work
- [ ] Exact matches take priority over wildcards
- [ ] Command fallback works when first command fails
- [ ] Unmatched MIME types fall back to pager
- [ ] All unit tests pass
- [ ] Backward compatibility maintained

## Open Questions

- [x] Fallback behavior when no MIME type matches? → Same as when enter_behavior is not set ($PAGER or less)
- [x] Should commands be specified as array or string? → Array, to support fallback
- [x] MIME detection method? → Extension-based using mime.TypeByExtension()

## Implementation Phases

### Phase 1: Core MIME Parsing
**Goals:** Implement MIME configuration parsing

**Deliverables:**
- EnterBehaviorMIME constant in enter.go
- ParseEnterBehavior update for "mime:"
- MIMEBehaviorConfig type in mime.go
- ParseMIMEBehavior function
- Unit tests for parsing

### Phase 2: MIME Detection and Matching
**Goals:** Implement MIME type detection and matching

**Deliverables:**
- GetMIMEType function
- FindMatchingRule function
- MatchesMIMEPattern function
- Unit tests for detection and matching

### Phase 3: Execution Integration
**Goals:** Integrate with Enter key handling

**Deliverables:**
- openWithMIME function in exec.go
- handleEnter modification for EnterBehaviorMIME
- Command fallback logic
- Integration tests

### Phase 4: Testing & Documentation
**Goals:** Comprehensive testing and documentation

**Deliverables:**
- E2E tests
- Edge case handling
- Config template update (generator.go)

## References

- Requirements Document: `doc/tasks/mime-enter-behavior/要件定義書.md`
- Existing enter behavior: `doc/tasks/config-enter-behavior/SPEC.md`
- Config implementation: `internal/config/config.go`
- Enter behavior implementation: `internal/config/enter.go`
- Execution helpers: `internal/ui/exec.go`
- Go mime package: https://pkg.go.dev/mime
