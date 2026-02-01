# Implementation Plan: MIME Type Based Enter Behavior

## Overview

This feature extends the existing `enter_behavior` configuration to support MIME type-based file opening. When `enter_behavior = "mime:"` is set, the system uses the `[enter_behavior_mime]` section to determine which application to use based on the file's MIME type.

## Objectives

- Add `enter_behavior = "mime:"` as a new configuration option
- Support `[enter_behavior_mime]` section for MIME type to command mapping
- Support wildcard matching (e.g., `text/*`, `image/*`)
- Implement command fallback when the first command fails
- Fall back to default pager ($PAGER or less) when no MIME type matches

## Prerequisites

### Development Environment
- Go 1.21 or later
- Make (for build automation)

### Dependencies
- github.com/charmbracelet/bubbletea (existing)
- github.com/BurntSushi/toml (existing)
- Standard library: `mime`, `path/filepath`, `strings`, `os/exec`

### Knowledge Requirements
- Understanding of Go's `mime` package for MIME type detection
- Familiarity with existing `enter_behavior` implementation in `internal/config/enter.go`
- Understanding of `tea.ExecProcess` for foreground command execution

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea
- **Key Libraries**:
  - `mime` - MIME type detection from file extension
  - `os/exec` - Command execution and validation

### Design Approach
- Extend existing `EnterBehaviorType` enum with new `EnterBehaviorMIME` type
- Add new `MIMEBehaviorConfig` struct to hold MIME type mappings
- Implement matching algorithm with exact match priority over wildcard
- Reuse existing `openWithViewer()` for fallback behavior

### Component Interaction

```
Config Loading Flow:
LoadConfig() → Parse enter_behavior → If "mime:" → Parse [enter_behavior_mime] section
                                                    ↓
                                            MIMEBehaviorConfig stored in Config
                                                    ↓
                                            Passed to Model via NewModelWithConfig()

Runtime Flow:
User presses Enter → handleEnter() → enterBehavior.Type == EnterBehaviorMIME
                                              ↓
                                    GetMIMEType(filename)
                                              ↓
                                    mimeBehavior.FindMatchingRule(mimeType)
                                              ↓
                            ┌─────────────────┴─────────────────┐
                            ↓                                   ↓
                      Match found                          No match
                            ↓                                   ↓
                  openWithMIME()                        openWithViewer()
                            ↓
                  Try commands in order
                            ↓
            ┌───────────────┴───────────────┐
            ↓                               ↓
      Command success                 Command fails
            ↓                               ↓
          Done                        Try next command
                                            ↓
                                  All failed → openWithViewer()
```

## Implementation Phases

### Phase 1: Core MIME Parsing

**Goal**: Implement MIME configuration parsing and type definition

**Files to Create**:
- `internal/config/mime.go` - MIME behavior parsing and matching logic
- `internal/config/mime_test.go` - Unit tests for MIME behavior

**Files to Modify**:
- `internal/config/enter.go`:
  - Add `EnterBehaviorMIME` constant
  - Update `ParseEnterBehavior()` to handle "mime:" value
  - Update `String()` method

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| EnterBehaviorMIME | New constant for MIME-based behavior | N/A | Constant defined |
| MIMEBehaviorConfig | Hold MIME type to command mappings | N/A | Struct with Rules map |
| ParseMIMEBehavior | Parse [enter_behavior_mime] section | Raw TOML map | Parsed config with warnings |

**Processing Flow**:
```
1. Input: raw map[string][]string from TOML
2. Validate each entry
   ├─ Empty key → Generate warning, skip
   └─ Empty command array → Generate warning, skip
3. Store valid entries in Rules map
4. Return MIMEBehaviorConfig and warnings
```

**Implementation Steps**:

1. **Add EnterBehaviorMIME constant**
   - Add new constant to EnterBehaviorType enum
   - Value: `EnterBehaviorMIME` after `EnterBehaviorCustom`

2. **Update ParseEnterBehavior**
   - Handle "mime:" string value
   - Return `EnterBehavior{Type: EnterBehaviorMIME}`

3. **Create MIMEBehaviorConfig type**
   - Define struct with `Rules map[string][]string`
   - MIME type pattern as key, command list as value

4. **Implement ParseMIMEBehavior**
   - Iterate input map, validate entries
   - Generate warnings for invalid entries
   - Return populated config

**Dependencies**:
- Requires: None (new code)
- Blocks: Phase 2, Phase 3

**Testing Approach**:

*Unit Tests*:
- Test ParseEnterBehavior with "mime:" input
- Test ParseMIMEBehavior with valid/invalid configurations
- Test warning generation for edge cases

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| ParseEnterBehavior_MIME | `"mime:"` | Type=EnterBehaviorMIME |
| ParseMIMEBehavior_Valid | Valid config | Parsed rules, no warnings |
| ParseMIMEBehavior_EmptyKey | `"" = ["less"]` | Warning generated |
| ParseMIMEBehavior_EmptyArray | `"text/*" = []` | Warning generated |

**Acceptance Criteria**:
- [ ] EnterBehaviorMIME constant defined
- [ ] ParseEnterBehavior returns EnterBehaviorMIME for "mime:"
- [ ] String() returns "mime:" for EnterBehaviorMIME
- [ ] MIMEBehaviorConfig struct defined
- [ ] ParseMIMEBehavior validates and parses correctly
- [ ] All unit tests pass

**Estimated Effort**: 小 (1-2 days)

---

### Phase 2: MIME Detection and Matching

**Goal**: Implement MIME type detection from filename and pattern matching

**Files to Modify**:
- `internal/config/mime.go`:
  - Add GetMIMEType function
  - Add FindMatchingRule method
  - Add MatchesMIMEPattern function

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| GetMIMEType | Detect MIME type from filename | Valid filename | MIME type string |
| FindMatchingRule | Find commands for MIME type | Initialized config | Commands or nil |
| MatchesMIMEPattern | Check pattern match | MIME type and pattern | Boolean match result |

**Processing Flow**:
```
GetMIMEType:
1. Extract extension from filename
   ├─ No extension → Return "application/octet-stream"
   └─ Has extension → Continue
2. Look up MIME type via mime.TypeByExtension()
   ├─ Unknown → Return "application/octet-stream"
   └─ Known → Continue
3. Strip parameters (e.g., "; charset=utf-8")
4. Return clean MIME type

FindMatchingRule:
1. Try exact match in Rules map
   ├─ Found → Return commands
   └─ Not found → Continue
2. Extract type prefix (e.g., "text" from "text/plain")
3. Try wildcard match (e.g., "text/*")
   ├─ Found → Return commands
   └─ Not found → Return nil, false

**Matching Priority Order:**
| Priority | Pattern Type | Example | Matches |
|----------|--------------|---------|---------|
| 1 (Highest) | Exact match | `text/plain` | Only `text/plain` |
| 2 (Lowest) | Wildcard | `text/*` | Any `text/*` subtype |

Note: If both `text/plain` and `text/*` rules exist, `text/plain` files use the exact match rule.
```

**Implementation Steps**:

1. **Implement GetMIMEType**
   - Use `filepath.Ext()` to extract extension
   - Use `mime.TypeByExtension()` for lookup
   - Handle unknown extensions and parameter stripping

2. **Implement FindMatchingRule**
   - Check exact match first (priority)
   - Fall back to wildcard match
   - Return commands and found boolean

3. **Implement MatchesMIMEPattern (helper)**
   - Support exact match
   - Support `type/*` wildcard pattern

**Dependencies**:
- Requires: Phase 1 (MIMEBehaviorConfig type)
- Blocks: Phase 3

**Testing Approach**:

*Unit Tests*:

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| GetMIMEType_TextPlain | `"file.txt"` | `"text/plain"` |
| GetMIMEType_ImagePNG | `"image.png"` | `"image/png"` |
| GetMIMEType_Unknown | `"file.xyz"` | `"application/octet-stream"` |
| GetMIMEType_NoExtension | `"Makefile"` | `"application/octet-stream"` |
| FindMatchingRule_ExactMatch | `"application/pdf"` with exact rule | Commands found |
| FindMatchingRule_WildcardMatch | `"image/png"` with `"image/*"` | Commands found |
| FindMatchingRule_ExactPriority | `"text/plain"` with both rules | Exact rule commands |
| FindMatchingRule_NoMatch | `"unknown/type"` | nil, false |

**Acceptance Criteria**:
- [ ] GetMIMEType returns correct MIME types
- [ ] GetMIMEType returns "application/octet-stream" for unknown
- [ ] FindMatchingRule prioritizes exact match
- [ ] FindMatchingRule falls back to wildcard
- [ ] All unit tests pass

**Estimated Effort**: 小 (1-2 days)

---

### Phase 3: Configuration Integration

**Goal**: Integrate MIME behavior into config loading and Model

**Files to Modify**:
- `internal/config/config.go`:
  - Add `MIMEBehavior` field to Config struct
  - Add `EnterBehaviorMIME` field to rawConfig struct
  - Update LoadConfig to parse MIME section

- `internal/ui/model.go`:
  - Add `mimeBehavior` field to Model struct
  - Update NewModelWithConfig signature and initialization

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Config.MIMEBehavior | Store parsed MIME config | Config loaded | MIMEBehaviorConfig available |
| rawConfig.EnterBehaviorMIME | TOML parsing target | TOML file parsed | Raw data available |
| Model.mimeBehavior | Runtime MIME config access | Model initialized | Config accessible |

**Processing Flow**:
```
LoadConfig:
1. Parse TOML file
2. Check if enter_behavior is "mime:"
   ├─ No → Skip MIME section parsing
   └─ Yes → Continue
3. Parse [enter_behavior_mime] section
4. Call ParseMIMEBehavior with raw data
5. Store result in Config.MIMEBehavior
6. Return config with warnings

NewModelWithConfig:
1. Accept MIMEBehaviorConfig parameter
2. Store in Model.mimeBehavior field
```

**Implementation Steps**:

1. **Update Config struct**
   - Add `MIMEBehavior MIMEBehaviorConfig` field

2. **Update rawConfig struct**
   - Add `EnterBehaviorMIME map[string][]string` with TOML tag

3. **Update LoadConfig**
   - Parse MIME section when enter_behavior is "mime:"
   - Call ParseMIMEBehavior and collect warnings
   - Store in Config.MIMEBehavior

4. **Update Model and NewModelWithConfig**
   - Add mimeBehavior field to Model
   - Update function signature to accept MIMEBehaviorConfig
   - Update all NewModelWithConfig call sites

**Dependencies**:
- Requires: Phase 1, Phase 2
- Blocks: Phase 4

**Testing Approach**:

*Integration Tests*:

| Test Case | Description |
|-----------|-------------|
| LoadConfig_WithMIMEBehavior | Load config with mime: and MIME section |
| LoadConfig_MIMEWithoutSection | mime: without section uses empty rules |

**Acceptance Criteria**:
- [ ] Config struct has MIMEBehavior field
- [ ] rawConfig parses [enter_behavior_mime] section
- [ ] LoadConfig populates MIMEBehavior correctly
- [ ] Model stores and exposes mimeBehavior
- [ ] All existing tests pass
- [ ] Integration tests pass

**Estimated Effort**: 小 (1-2 days)

---

### Phase 4: Execution Integration

**Goal**: Integrate MIME-based file opening with Enter key handling

**Files to Modify**:
- `internal/ui/exec.go`:
  - Add openWithMIME function

- `internal/ui/model_update_keyboard.go`:
  - Update handleEnter to handle EnterBehaviorMIME

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| openWithMIME | Execute file with MIME-based command | File path, MIME config | Command execution |
| handleEnter (update) | Route to MIME handler | Enter pressed on file | Correct handler called |

**Processing Flow**:
```
openWithMIME:
1. Detect MIME type from filename
2. Find matching rule
   ├─ No match → Return openWithViewer()
   └─ Match found → Continue
3. For each command in list:
   a. Validate command exists (exec.LookPath)
      ├─ Not found → Show status message, try next command
      └─ Found → Execute via tea.ExecProcess
   b. If execution succeeds → Done
   c. If command startup fails (exec error) → Try next command
4. All commands failed → Show status message, return openWithViewer()

**Command Failure Definition**:
A command is considered "failed" only when:
- `exec.LookPath` returns error (command not found in PATH)

A command is NOT considered "failed" when:
- The command exits with non-zero status (application's internal error)
- The user closes the application normally
- `exec.Command().Start()` returns error (process startup failure)

Note: `Start()` failure cannot trigger fallback because Bubble Tea's `tea.ExecProcess`
releases terminal control before executing the command. Once terminal control is released,
returning to the TUI to try the next command is not feasible. Therefore, fallback is
limited to `LookPath` validation only.

handleEnter (addition):
1. Check enterBehavior.Type
   ├─ EnterBehaviorMIME → Call openWithMIME()
   └─ Other types → Existing handling
```

**Status Message Handling**:

Use existing `statusMessageClearCmd(5 * time.Second)` pattern from `internal/ui/commands.go`.

| Scenario | Message Format | Duration |
|----------|----------------|----------|
| Command not found (LookPath) | `"Command not found: {cmd}, trying next..."` | 5 seconds |
| All commands failed (LookPath) | `"All configured commands failed, using pager"` | 5 seconds |
| MIME type no match | (No message - silent fallback to pager) | - |

**Implementation**: Return `tea.Batch(statusMessageCmd(...), statusMessageClearCmd(5*time.Second), nextCmd)` to display status and continue.

**Implementation Steps**:

1. **Implement openWithMIME**
   - Accept file path, workDir, MIMEBehaviorConfig
   - Detect MIME type using GetMIMEType
   - Find matching rule using FindMatchingRule
   - Try commands in order with exec.LookPath validation
   - Fall back to openWithViewer on failure

2. **Update handleEnter**
   - Add case for EnterBehaviorMIME
   - Call openWithMIME with appropriate parameters

**Dependencies**:
- Requires: Phase 1, Phase 2, Phase 3
- Blocks: None

**Testing Approach**:

*Unit Tests*:
- Test openWithMIME with various scenarios

*Integration Tests*:
- Test Enter key behavior with MIME configuration

*Manual Testing*:
- [ ] Press Enter on .txt file with text/* rule
- [ ] Press Enter on .png file with image/* rule
- [ ] Press Enter on unknown file type (fallback to pager)
- [ ] Configure invalid command, verify fallback works

**Acceptance Criteria**:
- [ ] openWithMIME executes correct command for MIME type
- [ ] Command fallback works when first command fails
- [ ] Fallback to pager works when no match or all fail
- [ ] handleEnter routes to openWithMIME for EnterBehaviorMIME
- [ ] All tests pass

**Estimated Effort**: 中 (2-3 days)

---

### Phase 5: Documentation and Testing

**Goal**: Update configuration template and comprehensive testing

**Files to Modify**:
- `internal/config/generator.go`:
  - Update defaultConfigTemplate with MIME example

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| defaultConfigTemplate | Example configuration | N/A | MIME section documented |

**Processing Flow**:
```
1. Add "mime:" option to enter_behavior documentation
2. Add [enter_behavior_mime] section template with examples
3. Include comments explaining wildcard patterns and fallback behavior
```

**Implementation Steps**:

1. **Update defaultConfigTemplate**
   - Add "mime:" to enter_behavior options comment
   - Add commented [enter_behavior_mime] section
   - Include examples for common MIME types

**Dependencies**:
- Requires: Phase 1-4
- Blocks: None

**Testing Approach**:

*Manual Testing*:
- [ ] Generate new config file, verify MIME section present
- [ ] Verify comments are clear and accurate

**Acceptance Criteria**:
- [ ] Config template includes "mime:" option
- [ ] Config template includes [enter_behavior_mime] section
- [ ] Comments explain usage clearly
- [ ] All E2E tests pass

**Estimated Effort**: 小 (0.5 days)

---

## Complete File Structure

```
internal/
├── config/
│   ├── config.go           # Modified: Add MIMEBehavior field
│   ├── enter.go            # Modified: Add EnterBehaviorMIME constant
│   ├── mime.go             # NEW: MIME behavior parsing and matching
│   ├── mime_test.go        # NEW: Tests for MIME behavior
│   ├── generator.go        # Modified: Update config template
│   └── ...
└── ui/
    ├── exec.go             # Modified: Add openWithMIME function
    ├── model.go            # Modified: Add mimeBehavior field
    ├── model_update_keyboard.go  # Modified: Handle EnterBehaviorMIME
    └── ...
```

**File Descriptions**:

| File | Purpose |
|------|---------|
| `config/mime.go` | MIME type detection, pattern matching, configuration parsing |
| `config/mime_test.go` | Unit tests for MIME functionality |
| `config/enter.go` | EnterBehaviorMIME constant and parsing |
| `config/config.go` | MIMEBehavior field and loading logic |
| `ui/exec.go` | openWithMIME function for command execution |
| `ui/model.go` | mimeBehavior field storage |
| `ui/model_update_keyboard.go` | Enter key routing to MIME handler |
| `config/generator.go` | Config template with MIME examples |

## Testing Strategy

### Unit Testing

**Approach**:
- Use Go's built-in `testing` package
- Table-driven tests for multiple scenarios
- No external mocking libraries needed

**Test Coverage Goals**:
- Core logic (mime.go): 90%+ coverage
- Parser updates (enter.go): 90%+ coverage
- Config loading: 80%+ coverage

**Key Test Areas**:

1. **MIME Type Detection** (`internal/config/mime.go`)
   - Various file extensions
   - Unknown extensions
   - Files without extensions
   - MIME types with parameters

2. **Pattern Matching** (`internal/config/mime.go`)
   - Exact matches
   - Wildcard matches
   - Priority (exact over wildcard)
   - No matches

3. **Configuration Parsing** (`internal/config/`)
   - Valid MIME configurations
   - Invalid entries (empty key, empty array)
   - Missing section with mime: behavior

### Integration Testing

**Scenarios**:
1. Load config with enter_behavior = "mime:" and MIME section
2. Load config with enter_behavior = "mime:" without MIME section
3. Model initialization with MIME configuration

### E2E Testing

| Test Case | Description |
|-----------|-------------|
| Enter_MIMEText | Press Enter on .txt file, verify configured viewer opens |
| Enter_MIMEImage | Press Enter on .png file, verify configured viewer opens |
| Enter_MIMENoMatch | Press Enter on unknown type, verify pager opens |
| Enter_MIMECommandFallback | First command fails, verify second is tried |

### Edge Cases

- [ ] Empty `[enter_behavior_mime]` section: All files fall back to pager
- [ ] MIME type with parameters: Parameters are stripped before matching
- [ ] File with no extension: Falls back to pager
- [ ] All configured commands not found: Falls back to pager
- [ ] `enter_behavior = "mime:"` without MIME section: Falls back to pager

## Dependencies

### External Dependencies

| Package | Version | Purpose | Notes |
|---------|---------|---------|-------|
| Standard library `mime` | N/A | MIME type detection | No external dependency |
| Standard library `os/exec` | N/A | Command execution | Already used |

### Internal Dependencies

**Implementation Order** (respecting dependencies):
1. Phase 1 (no dependencies)
2. Phase 2 (depends on Phase 1)
3. Phase 3 (depends on Phase 1, 2)
4. Phase 4 (depends on Phase 1, 2, 3)
5. Phase 5 (depends on Phase 1-4)

**Component Dependencies**:
- `config/mime.go` - standalone, no internal dependencies
- `config/config.go` - depends on `config/mime.go`
- `ui/model.go` - depends on `config/config.go`
- `ui/exec.go` - depends on `config/mime.go`
- `ui/model_update_keyboard.go` - depends on `ui/exec.go`, `config/enter.go`

## Risk Assessment

### Technical Risks

1. **MIME Type Database Variability**
   - **Risk**: Go's mime package may not recognize all extensions
   - **Likelihood**: Low (covers common types)
   - **Impact**: Low (falls back to pager)
   - **Mitigation**: Document limitation, fallback behavior handles it

2. **Command Execution Failure**
   - **Risk**: Commands may fail for various reasons
   - **Likelihood**: Medium (user configuration errors)
   - **Impact**: Low (fallback to pager)
   - **Mitigation**: Clear status messages, graceful fallback

### Implementation Risks

1. **Breaking Existing Behavior**
   - **Risk**: Existing enter_behavior values may be affected
   - **Likelihood**: Low (additive change)
   - **Impact**: High (core functionality)
   - **Mitigation**: Comprehensive testing of existing values

## Performance Considerations

1. **MIME Type Detection**
   - Uses extension lookup only (no file I/O)
   - Performance: < 1ms (meets NFR1)

2. **Pattern Matching**
   - O(n) where n is number of rules
   - Typical n < 10, acceptable performance

3. **Command Validation**
   - exec.LookPath called only when executing
   - Single system call per command tried

## Security Considerations

1. **No Shell Execution**
   - Commands executed directly via exec, not through shell
   - Prevents shell injection

2. **Path Arguments**
   - File paths passed as command arguments, not interpolated
   - No path injection risk

3. **Command Validation**
   - Commands validated via exec.LookPath before execution
   - Only executable files can be run

## Open Questions

### From Specification:
- [x] Fallback behavior when no MIME type matches? -> Same as default ($PAGER or less)
- [x] Commands as array or string? -> Array, to support fallback
- [x] MIME detection method? -> Extension-based using mime.TypeByExtension()

### Implementation-Specific:
- All questions resolved in specification

## Future Enhancements

Items not in current scope:

- Content-based MIME detection (magic bytes)
- User-defined MIME type overrides
- Per-directory MIME configurations

## Success Criteria

### Functional Completeness
- [ ] `enter_behavior = "mime:"` is recognized as valid
- [ ] `[enter_behavior_mime]` section is parsed correctly
- [ ] Exact MIME type matches work
- [ ] Wildcard patterns work
- [ ] Exact matches take priority over wildcards
- [ ] Command fallback works when first command fails
- [ ] Unmatched MIME types fall back to pager
- [ ] Backward compatibility maintained

### Quality Metrics
- [ ] Test coverage meets goals (90%+ for mime.go)
- [ ] No critical bugs in manual testing
- [ ] Code follows Go best practices

### Performance Metrics
- [ ] MIME type detection < 1ms

### User Experience
- [ ] Clear warning messages for invalid configurations
- [ ] Status messages for command failures

## References

- **Specification**: `doc/tasks/mime-enter-behavior/SPEC.md`
- **Existing enter behavior**: `internal/config/enter.go`
- **Config implementation**: `internal/config/config.go`
- **Execution helpers**: `internal/ui/exec.go`
- **Go mime package**: https://pkg.go.dev/mime

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - Confirm approach and timeline
   - Address any remaining questions

2. **Begin Implementation**
   - Start with Phase 1
   - Follow TDD approach (write tests first)
   - Commit incrementally

3. **Verification**
   - Run tests after each phase
   - Manual testing with real configurations
