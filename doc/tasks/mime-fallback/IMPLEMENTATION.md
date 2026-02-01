# Implementation Plan: MIME Fallback Configuration

## Overview

This feature adds a `fallback` key to the `[enter_behavior_mime]` configuration section, allowing users to specify which command(s) to use when no MIME rule matches a file's type. The default fallback is `["xdg-open"]`, and it is auto-merged into existing configs that lack it.

## Objectives

- Add `Fallback []string` field to `MIMEBehaviorConfig`
- Parse `fallback` key from `[enter_behavior_mime]` section, separate from MIME rules
- Use fallback commands in `openWithMIME` when no MIME rule matches or all MIME commands fail
- Auto-merge `fallback = ["xdg-open"]` into existing configs via config merger
- Update default config template to include `fallback`

## Prerequisites

### Development Environment
- Go 1.21 or later
- Make (for build automation)

### Dependencies
- github.com/charmbracelet/bubbletea (existing)
- github.com/BurntSushi/toml (existing)
- Standard library: `os/exec`, `strings`, `fmt`

### Knowledge Requirements
- Understanding of existing MIME enter behavior (`internal/config/mime.go`, `internal/ui/exec.go`)
- Understanding of config merge logic (`internal/config/merger.go`)
- Understanding of `openWithMIME` return contract: `(tea.Cmd, string)`

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea
- **Key Libraries**:
  - `os/exec` - Command validation via LookPath
  - `strings` - Key analysis (contains `/`)

### Design Approach
- Extend existing `MIMEBehaviorConfig` struct with a `Fallback` field
- Use explicit key identification in `ParseMIMEBehavior`: `"fallback"` is extracted to its own field; keys containing `/` are MIME patterns; unknown keys generate a warning
- Extend `openWithMIME` fallback chain: MIME rule commands -> fallback commands -> pager
- Extend config merger to detect and insert missing `fallback` key

### Component Interaction

```
Config Loading Flow:
LoadConfig() -> Parse [enter_behavior_mime] section
                         |
                  ParseMIMEBehavior(raw)
                         |
                  Classify each key by name
                         |
              +----------+---------+---------+
              |          |                   |
        Key="fallback"   Key contains "/"    Other keys
              |          |                   |
        Store in       Store in           Generate warning,
        Fallback       Rules              skip
              |          |
              +----> MIMEBehaviorConfig <----+
                     {Rules, Fallback}

Runtime Flow (openWithMIME):
1. Detect MIME type from filename
2. Find matching MIME rule
   |
   +-- Match found -> Try MIME rule commands
   |                        |
   |                   +----+----+
   |                   |         |
   |              Cmd found   All fail
   |                   |         |
   |              Execute    Fall through
   |                              |
   +-- No match ----------------->+
                                  |
                           Try fallback commands
                                  |
                           +------+------+
                           |             |
                      Cmd found      All fail
                           |             |
                      Execute     openWithViewer()

Config Merge Flow:
MergeConfig() -> Check [enter_behavior_mime] section
                         |
              +----------+-----------+
              |                      |
        Section missing      Section exists
              |                      |
     Add full section       Check "fallback" key
     (with fallback)                 |
                           +---------+---------+
                           |                   |
                    Key missing          Key present
                           |                   |
                  Append fallback        No change
```

## Implementation Phases

### Phase 1: Extend ParseMIMEBehavior and MIMEBehaviorConfig

**Goal**: Add `Fallback` field to `MIMEBehaviorConfig` and extract `fallback` key from TOML map during parsing

**Files to Modify**:
- `internal/config/mime.go`:
  - Add `Fallback []string` field to `MIMEBehaviorConfig`
  - Update `ParseMIMEBehavior` to separate `fallback` from MIME rules
- `internal/config/mime_test.go`:
  - Add tests for fallback parsing

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| MIMEBehaviorConfig.Fallback | Store fallback command list | N/A | Field available on struct |
| ParseMIMEBehavior (updated) | Separate fallback key from MIME rules | Raw TOML map (may contain `fallback`) | `Fallback` field populated; `Rules` contains only MIME patterns (keys with `/`) |

**Processing Flow**:
```
1. Input: raw map[string][]string from TOML
2. For each key in map:
   +-- Key is empty -> Generate warning, skip
   +-- Key is "fallback" -> Validate command array, store in Fallback (see step 3)
   +-- Key contains "/" -> Validate as MIME rule, store in Rules
   +-- Otherwise (unknown key) -> Generate warning ("unknown key '%s' in enter_behavior_mime, skipping"), skip
3. Validation for fallback:
   +-- Empty array -> Generate warning ("empty command list for fallback, skipping")
   +-- Non-empty -> Store in Fallback field
4. Return MIMEBehaviorConfig{Rules, Fallback} and warnings
```

**Key Identification Strategy**:
- The `fallback` key is identified by its **exact name**, not by the absence of `/`
- Keys containing `/` are treated as MIME type patterns
- Any other key generates a warning and is skipped
- This explicit approach prevents misclassification of typos or future keys

**Implementation Steps**:

1. **Add Fallback field to MIMEBehaviorConfig**
   - New field: `Fallback []string`
   - Holds the command list from the `fallback` key

2. **Update ParseMIMEBehavior to separate keys**
   - Check each key by name: `"fallback"` is extracted to `Fallback` field
   - Keys containing `/` are treated as MIME patterns and stored in `Rules`
   - Unknown keys (not `"fallback"`, not containing `/`, not empty) generate a warning and are skipped
   - Generate warning for empty fallback array
   - Skip `fallback` when populating `Rules` map

3. **Add unit tests for fallback parsing**
   - Key considerations:
     - Existing tests must continue to pass unchanged
     - New tests cover fallback extraction, empty fallback, missing fallback, and fallback-only config

**Dependencies**:
- Requires: None (modifies existing code)
- Blocks: Phase 2

**Testing Approach**:

*Unit Tests*:

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Fallback extracted | `{"text/*": ["less"], "fallback": ["xdg-open"]}` | Rules has `text/*`; Fallback is `["xdg-open"]`; no warnings |
| Fallback not in Rules | `{"fallback": ["xdg-open"]}` | Rules is empty; Fallback is `["xdg-open"]` |
| Empty fallback array | `{"fallback": []}` | Fallback is nil/empty; warning generated |
| Missing fallback | `{"text/*": ["less"]}` | Fallback is nil/empty; no warnings |
| Fallback only (no MIME rules) | `{"fallback": ["xdg-open"]}` | Rules empty; Fallback populated |
| Multiple fallback commands | `{"fallback": ["xdg-open", "open"]}` | Fallback has 2 entries |
| Unknown key generates warning | `{"text": ["less"], "fallback": ["xdg-open"]}` | Warning for `text` key; `fallback` extracted; `text` not in Rules |
| Existing tests unchanged | All existing test inputs | Same results as before |

**Acceptance Criteria**:
- [ ] `MIMEBehaviorConfig` has `Fallback []string` field
- [ ] `ParseMIMEBehavior` extracts `fallback` to `Fallback` field
- [ ] `fallback` key is NOT included in `Rules` map
- [ ] MIME rules (keys with `/`) remain in `Rules` map
- [ ] Empty `fallback` array generates warning
- [ ] Missing `fallback` results in nil/empty `Fallback`
- [ ] All existing MIME tests continue to pass

**Estimated Effort**: 小 (0.5 days)

---

### Phase 2: Update openWithMIME for Fallback

**Goal**: When no MIME rule matches or all MIME rule commands fail, try fallback commands before falling back to pager

**Files to Modify**:
- `internal/ui/exec.go`:
  - Update `openWithMIME` to use `Fallback` commands
- `internal/ui/exec_test.go`:
  - Add tests for fallback command execution

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| openWithMIME (updated) | Try fallback commands between MIME failure and pager | MIMEBehaviorConfig with Fallback field | File opened via fallback command or pager |

**Processing Flow**:
```
openWithMIME (updated):
1. Detect MIME type from filename
2. Find matching MIME rule
   +-- Match found -> Try MIME rule commands in order (existing logic)
   |                        +-- Command found -> Execute, return
   |                        +-- All fail -> Fall through to step 3
   +-- No match -> Fall through to step 3
3. Check if Fallback is configured and non-empty
   +-- No fallback -> openWithViewer (pager), return
   +-- Has fallback -> Try fallback commands in order
4. For each fallback command:
   +-- Parse command string (split by whitespace for options)
   +-- Validate via LookPath
   +-- Found -> Execute with filepath appended, return
   +-- Not found -> Track in notFoundCmds, try next
5. All fallback commands not found -> openWithViewer with status message
```

**Command Format**:
- Fallback commands use the same format as MIME rule commands
- A command string like `"vim -R"` is split by whitespace: command=`vim`, args=`["-R", filepath]`
- File path is appended as the last argument

**Status Message Behavior**:

| Scenario | Status Message |
|----------|---------------|
| MIME match found, command available | Empty (no message) |
| No MIME match, fallback command available | Empty (no message) |
| All MIME commands fail, fallback command available | Empty (no message) |
| All MIME + fallback commands fail | "All configured commands failed ({all names combined}), using pager" |
| No MIME match, all fallback commands fail | "All configured commands failed ({fallback names}), using pager" |
| No MIME match, no fallback configured | Empty (silent fallback to pager) |

**Note**: The `{names}` in the status message includes all tried command names from both MIME rules and fallback, collected into a single `notFoundCmds` slice throughout the function.

**Implementation Steps**:

1. **Refactor openWithMIME fallback path**
   - After MIME rule commands are exhausted (or no match found), check `mimeCfg.Fallback`
   - If Fallback is non-empty, iterate and try each command via LookPath
   - If a fallback command is found, execute it and return
   - If all fallback commands fail, fall back to pager with combined status message

2. **Add unit tests for fallback execution**
   - Key considerations:
     - Tests use `MIMEBehaviorConfig` with `Fallback` field populated
     - Test both "no MIME match" and "all MIME commands fail" paths into fallback

**Dependencies**:
- Requires: Phase 1 (Fallback field exists)
- Blocks: None

**Testing Approach**:

*Unit Tests*:

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| No MIME match + fallback has valid command | Unknown file type, fallback=["cat"] | Command returned, no status message |
| No MIME match + all fallback commands missing | Unknown file type, fallback=["nonexist1", "nonexist2"] | Pager returned, status message with command names |
| No MIME match + no fallback | Unknown file type, fallback=nil | Pager returned, no status message |
| All MIME commands fail + fallback has valid command | Matched MIME rule with missing commands, fallback=["cat"] | Fallback command returned, no status message |
| MIME match found + fallback configured | Matched MIME rule, fallback=["xdg-open"] | MIME command used, fallback not tried |
| Fallback command with options | fallback=["head -n 20"] | Command parsed with options, filepath appended |
| Fallback tries commands in order | fallback=["nonexist", "cat"] | Second command (cat) used |

**Acceptance Criteria**:
- [ ] No MIME match + fallback configured -> tries fallback commands
- [ ] No MIME match + fallback command found -> executes fallback command
- [ ] No MIME match + all fallback commands not found -> falls back to pager
- [ ] No MIME match + no fallback configured -> falls back to pager (silent)
- [ ] All MIME rule commands not found + fallback configured -> tries fallback commands
- [ ] MIME match found -> uses MIME rule (fallback not used)
- [ ] Fallback commands support options (whitespace splitting)
- [ ] Status message includes failed command names when all commands fail

**Estimated Effort**: 小 (1 day)

---

### Phase 3: Config Merger and Template Update

**Goal**: Auto-merge `fallback = ["xdg-open"]` into existing configs and update default template

**Files to Modify**:
- `internal/config/merger.go`:
  - Add fallback detection and insertion logic
- `internal/config/merger_test.go`:
  - Add tests for fallback merge scenarios
- `internal/config/generator.go`:
  - Update `defaultConfigTemplate` to include `fallback`

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| MergeConfig (updated) | Detect missing `fallback` in MIME section and insert it | Config file on disk | `fallback` key present in file |
| generateMergedFile (updated) | Insert `fallback` into MIME section content | File content string | Updated content with `fallback` |
| defaultConfigTemplate (updated) | Include `fallback` in template | N/A | Template has `fallback` entry |

**Behavioral Change Note**:

The current merger adds a **commented** placeholder (`# [enter_behavior_mime]`) when the section is missing. This feature changes the behavior:
- When the section is missing: add an **active** `[enter_behavior_mime]` section with `fallback = ["xdg-open"]` and MIME rule examples as comments
- When the commented placeholder exists: replace it with an active section including `fallback`
- This is safe because `[enter_behavior_mime]` only takes effect when `enter_behavior = "mime:"` is set; otherwise the section is ignored

**Processing Flow**:
```
Config Merge - Fallback Detection:
1. Check [enter_behavior_mime] section state
   +-- Section missing entirely
   |     +-- No commented placeholder either
   |     |     -> Set EnterBehaviorMIME=true: add active section with fallback
   |     +-- Commented placeholder exists
   |           -> Set EnterBehaviorMIME=true: replace placeholder with active section including fallback
   +-- Section exists
         +-- Check if "fallback" key present in raw map
               +-- Present -> No change needed
               +-- Missing -> Set MIMEFallbackMissing=true: append fallback = ["xdg-open"] to section

Section Content Update (when section exists but fallback missing, MIMEFallbackMissing=true):
1. Track [enter_behavior_mime] section boundaries in generateMergedFile (similar to keybindings/colors)
2. Find the end of the section
3. Insert fallback = ["xdg-open"] at end of section
```

**Merge Scenarios**:

| Scenario | Current State | Action |
|----------|---------------|--------|
| Section missing, no placeholder | No MIME section at all | Add full section with `fallback` |
| Section missing, commented placeholder | `# [enter_behavior_mime]` exists | Replace placeholder with active section including `fallback` |
| Section exists, `fallback` missing | `[enter_behavior_mime]` with MIME rules only | Append `fallback = ["xdg-open"]` to section |
| Section exists, `fallback` present | Complete section | No change |

**Implementation Steps**:

1. **Extend mergeResult to track missing fallback**
   - Add `MIMEFallbackMissing bool` field to `mergeResult`: `true` when `[enter_behavior_mime]` section exists but `fallback` key is absent
   - Update `hasContent()` to include `MIMEFallbackMissing`
   - Key considerations:
     - The existing `EnterBehaviorMIME bool` flag handles the case where the entire section is missing
     - `MIMEFallbackMissing bool` handles the case where the section exists but `fallback` is absent
     - These two flags are mutually exclusive: if the section is missing, `MIMEFallbackMissing` is not set

2. **Add fallback detection in MergeConfig**
   - When `EnterBehaviorMIME` raw map is non-nil (section exists), check if `fallback` key is present
   - If absent, flag for insertion

3. **Update generateMergedFile for fallback insertion**
   - Track `[enter_behavior_mime]` section boundaries (add `enterBehaviorMIMESection *sectionInfo` similar to `keybindingsSection`/`colorsSection`)
   - When `MIMEFallbackMissing=true`: find the section end and insert `fallback = ["xdg-open"]` at end of section
   - When `EnterBehaviorMIME=true` (section missing or commented placeholder): append active section with `fallback` and comment examples
   - The active section content when created from scratch:
     ```toml
     [enter_behavior_mime]
     # "text/html" = ["vim", "-R"]
     fallback = ["xdg-open"]
     ```

4. **Update defaultConfigTemplate in generator.go**
   - Change the `[enter_behavior_mime]` section from commented-out to active
   - Include `fallback = ["xdg-open"]` as a non-commented entry
   - Keep MIME rule examples as comments

5. **Add merger tests**
   - Key considerations:
     - Test all four merge scenarios
     - Verify existing merge tests continue to pass
     - Verify idempotency (second merge makes no changes)

**Dependencies**:
- Requires: Phase 1 (Fallback field for detection)
- Blocks: None

**Testing Approach**:

*Unit Tests (merger)*:

| Test Case | Input Config | Expected Result |
|-----------|-------------|-----------------|
| Section missing, no placeholder | No MIME section | Full section with `fallback` added |
| Section exists, `fallback` missing | `[enter_behavior_mime]` with rules only | `fallback = ["xdg-open"]` appended |
| Section exists, `fallback` present | Complete section | No change |
| Commented placeholder, `fallback` missing | `# [enter_behavior_mime]` | Section with `fallback` added |
| Idempotency | Config after merge | No further changes on second merge |

*Unit Tests (generator)*:

| Test Case | Expected |
|-----------|----------|
| Template contains `[enter_behavior_mime]` | Section header present |
| Template contains `fallback` | `fallback = ["xdg-open"]` present |
| Template is valid TOML | Parses without error |

**Acceptance Criteria**:
- [ ] Missing `fallback` is detected when MIME section exists
- [ ] `fallback = ["xdg-open"]` is inserted into existing MIME section
- [ ] New MIME section includes `fallback` when created from scratch
- [ ] Default config template includes non-commented `[enter_behavior_mime]` section with `fallback`
- [ ] Existing configs are not broken by the merge
- [ ] Second merge on same file makes no changes (idempotency)
- [ ] All existing merger tests continue to pass

**Estimated Effort**: 中 (2-3 days)

**Risks and Mitigation**:
- **Risk**: Text-based TOML manipulation may incorrectly identify section boundaries
  - **Mitigation**: Reuse existing section parsing logic from `generateMergedFile`; test with various file formats

---

## Complete File Structure

```
internal/
├── config/
│   ├── mime.go             # Modified: Add Fallback field, update ParseMIMEBehavior
│   ├── mime_test.go        # Modified: Add fallback parsing tests
│   ├── merger.go           # Modified: Add fallback detection and insertion
│   ├── merger_test.go      # Modified: Add fallback merge tests
│   ├── generator.go        # Modified: Update defaultConfigTemplate
│   └── config.go           # No changes needed (uses ParseMIMEBehavior output)
└── ui/
    ├── exec.go             # Modified: Update openWithMIME fallback chain
    └── exec_test.go        # Modified: Add fallback execution tests
```

**File Descriptions**:

| File | Changes |
|------|---------|
| `config/mime.go` | Add `Fallback` field to struct; update parser to separate `fallback` from MIME rules |
| `config/mime_test.go` | Add tests for fallback extraction, empty fallback, missing fallback |
| `config/merger.go` | Detect missing `fallback` key; insert into existing MIME section or new section |
| `config/merger_test.go` | Add merge scenario tests for all fallback cases |
| `config/generator.go` | Update template: uncomment MIME section, add `fallback = ["xdg-open"]` |
| `ui/exec.go` | Extend `openWithMIME`: try Fallback commands after MIME rules fail |
| `ui/exec_test.go` | Add tests for fallback command execution scenarios |

## Testing Strategy

### Unit Testing

**Approach**:
- Use Go's built-in `testing` package
- Table-driven tests for multiple scenarios
- No external mocking libraries needed

**Test Coverage Goals**:
- MIME parsing changes: 90%+ coverage
- openWithMIME changes: 80%+ coverage
- Merger changes: 80%+ coverage

**Key Test Areas**:

1. **Fallback Parsing** (`internal/config/mime_test.go`)
   - `fallback` extracted to Fallback field
   - `fallback` not included in Rules map
   - Empty fallback array generates warning
   - Missing fallback results in nil Fallback
   - Config with only fallback (no MIME rules)

2. **Fallback Execution** (`internal/ui/exec_test.go`)
   - No MIME match -> fallback tried
   - All MIME commands fail -> fallback tried
   - All fallback commands fail -> pager
   - Fallback commands with options
   - MIME match -> fallback not used

3. **Config Merge** (`internal/config/merger_test.go`)
   - Section missing -> full section with fallback
   - Section exists, fallback missing -> fallback appended
   - Section exists, fallback present -> no change
   - Idempotency

### Manual Testing Checklist

- [ ] Create config with `enter_behavior = "mime:"` and `fallback = ["xdg-open"]`, open unmatched file
- [ ] Remove `fallback` from config, verify auto-merge adds it
- [ ] Set `fallback = ["nonexistent_command"]`, verify pager fallback
- [ ] Set MIME rules + fallback, verify MIME rules take priority for matching types
- [ ] Generate fresh config, verify template includes `fallback`
- [ ] Verify existing MIME behavior unchanged for matched files

## Dependencies

### External Dependencies

No new external dependencies required.

### Internal Dependencies

**Implementation Order** (respecting dependencies):
1. Phase 1 - Fallback field and parsing (no dependencies)
2. Phase 2 - openWithMIME update (depends on Phase 1)
3. Phase 3 - Config merger and template (depends on Phase 1)

Note: Phase 2 and Phase 3 can be implemented in parallel after Phase 1.

**Component Dependencies**:
- `config/mime.go` - standalone changes
- `ui/exec.go` - depends on `config/mime.go` (Fallback field)
- `config/merger.go` - depends on `config/mime.go` (Fallback field for detection)
- `config/generator.go` - standalone (template text only)

## Risk Assessment

### Technical Risks

1. **Config Merger Text Manipulation**
   - **Risk**: Inserting `fallback` into the wrong position in the TOML file
   - **Likelihood**: Low (reuses existing section parsing)
   - **Impact**: Medium (corrupted config file)
   - **Mitigation**: Comprehensive tests with various file formats; idempotency tests

2. **Backward Compatibility**
   - **Risk**: Existing configs without `fallback` may behave differently
   - **Likelihood**: Low (auto-merge provides default)
   - **Impact**: High (unexpected behavior change)
   - **Mitigation**: Default `["xdg-open"]` is a safe choice; existing pager fallback still works as last resort

### Implementation Risks

1. **Existing Test Breakage**
   - **Risk**: Changes to `ParseMIMEBehavior` or `openWithMIME` may break existing tests
   - **Likelihood**: Low (additive changes)
   - **Impact**: Medium
   - **Mitigation**: Run full test suite after each phase; existing tests should pass without modification

## Performance Considerations

1. **Fallback Key Detection**
   - O(1) check: `strings.Contains(key, "/")` for each entry during parsing
   - No measurable impact

2. **Fallback Command Execution**
   - Same mechanism as MIME rule command execution (LookPath + exec)
   - Only triggered when MIME rules don't match or fail
   - No additional overhead in the common case

## Security Considerations

1. **Command Execution**
   - Fallback commands use the same direct exec pattern as MIME commands (no shell)
   - Validated via LookPath before execution
   - No new attack surface

## Open Questions

(None - all questions resolved in specification)

## Success Criteria

### Functional Completeness
- [ ] `fallback` key is parsed separately from MIME rules
- [ ] `MIMEBehaviorConfig` has `Fallback` field
- [ ] `openWithMIME` uses fallback commands when no MIME rule matches
- [ ] `openWithMIME` uses fallback commands when all MIME rule commands fail
- [ ] Config merger adds missing `fallback` key
- [ ] Default config template includes `fallback`
- [ ] All existing MIME behavior tests continue to pass
- [ ] Backward compatibility with existing configs maintained

### Quality Metrics
- [ ] Test coverage meets goals (80%+ for changed code)
- [ ] No critical bugs in manual testing
- [ ] Code follows existing codebase conventions

## References

- **Specification**: `doc/tasks/mime-fallback/SPEC.md`
- **Requirements**: `doc/tasks/mime-fallback/要件定義書.md`
- **Parent feature spec**: `doc/tasks/mime-enter-behavior/SPEC.md`
- **Parent feature implementation**: `doc/tasks/mime-enter-behavior/IMPLEMENTATION.md`
- **MIME behavior code**: `internal/config/mime.go`
- **Execution helpers**: `internal/ui/exec.go`
- **Config merger**: `internal/config/merger.go`
- **Config generator**: `internal/config/generator.go`

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - Confirm approach
   - Address any remaining questions

2. **Begin Implementation**
   - Start with Phase 1 (parsing)
   - Follow TDD approach (write tests first)
   - Commit incrementally

3. **Verification**
   - Run tests after each phase: `go test ./...`
   - Manual testing with real configurations
