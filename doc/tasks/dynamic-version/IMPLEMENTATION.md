# Implementation Plan: Dynamic Version Display in Toolbar

## Overview

Create a shared `internal/version` package to centralize version information, enabling both the CLI `--version` option and TUI toolbar to display the same dynamically injected version string.

## Objectives

- Create `internal/version` package with exported Version variable
- Update Makefile ldflags to inject version into new package
- Update main.go to use version package
- Replace hardcoded version strings in UI with version package reference

## Prerequisites

- Go development environment
- Understanding of Go's `-ldflags -X` mechanism
- Access to Makefile and source files

## Architecture Overview

```
Before:
  main.go (var version) ← ldflags injection
  model.go ("v0.1.0" hardcoded)

After:
  internal/version/version.go (var Version) ← ldflags injection
  main.go → imports version package
  model.go → imports version package
```

## Implementation Phases

### Phase 1: Create Version Package

**Goal**: Create a new package to hold the version variable

**Files to Create**:
- `internal/version/version.go` - Version variable definition

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Version variable | Hold version string | None | Exported variable accessible |

**Processing Flow**:
```
1. Create package → 2. Define exported variable with default "dev"
```

**Implementation Steps**:
1. Create `internal/version/` directory
2. Create `version.go` with exported `Version` variable
   - Default value: `"dev"`
   - Package path: `github.com/sakura/duofm/internal/version`

**Dependencies**: None

**Testing**:
- Verify package compiles
- Verify variable is exported (capital V)

**Estimated Effort**: Small

---

### Phase 2: Update Makefile

**Goal**: Change ldflags to inject version into new package

**Files to Modify**:
- `Makefile` - Update LDFLAGS variable

**Processing Flow**:
```
1. Locate LDFLAGS definition → 2. Update package path → 3. Verify VERSION variable unchanged
```

**Implementation Steps**:
1. Change ldflags from:
   - `-X main.version=$(VERSION)`
   to:
   - `-X github.com/sakura/duofm/internal/version.Version=$(VERSION)`

**Dependencies**: Phase 1 completed

**Testing**:
- `make build` succeeds
- `./duofm --version` shows expected version (will test after Phase 3)

**Estimated Effort**: Small

---

### Phase 3: Update main.go

**Goal**: Use version package instead of local variable

**Files to Modify**:
- `cmd/duofm/main.go` - Import version package, remove local variable

**Processing Flow**:
```
1. Add import → 2. Remove local version var → 3. Update --version output to use version.Version
```

**Implementation Steps**:
1. Add import for `github.com/sakura/duofm/internal/version`
2. Remove `var version = "dev"` line
3. Update `--version` handler to use `version.Version`

**Dependencies**: Phase 1, Phase 2 completed

**Testing**:
- `make build` succeeds
- `./duofm --version` shows git tag or dev-commit

**Estimated Effort**: Small

---

### Phase 4: Update UI Model

**Goal**: Replace hardcoded version in toolbar with dynamic version

**Files to Modify**:
- `internal/ui/model.go` - Import version package, replace hardcoded strings

**Locations to Update**:
- Line 898: `View()` method
- Line 949: `renderDialogScreen()` method
- Line 1005: `renderDialogPane()` method
- Line 1033: `renderSortDialogPane()` method

**Processing Flow**:
```
1. Add import → 2. Create title rendering helper or inline format → 3. Replace all 4 hardcoded locations
```

**Implementation Steps**:
1. Add import for `github.com/sakura/duofm/internal/version`
2. Replace each occurrence of `"duofm v0.1.0"` with:
   - `fmt.Sprintf("duofm %s", version.Version)` or
   - `"duofm " + version.Version`

**Dependencies**: Phase 1 completed

**Testing**:
- Build and run duofm
- Verify toolbar shows correct version
- Verify all 4 locations display consistently

**Estimated Effort**: Small

---

## File Structure

```
duofm/
├── cmd/duofm/
│   └── main.go           # Modified: import version pkg, remove local var
├── internal/
│   ├── version/
│   │   └── version.go    # NEW: exported Version variable
│   └── ui/
│       └── model.go      # Modified: import version pkg, use Version
└── Makefile              # Modified: update ldflags path
```

## Testing Strategy

### Unit Tests
- No new unit tests required (version package is trivial)

### Integration Tests
- Existing tests should continue to pass

### Manual Testing Checklist
- [ ] `make build` completes successfully
- [ ] `./duofm --version` displays git tag (e.g., "duofm v1.0.0")
- [ ] `./duofm --version` displays dev-commit when no tag (e.g., "duofm dev-a1b2c3d")
- [ ] `go build ./cmd/duofm` without ldflags shows "duofm dev"
- [ ] TUI toolbar displays same version as `--version` output
- [ ] TUI dialogs (help, sort) show same version in title bar

## Dependencies

### External Libraries
- None

### Internal Dependencies
- Phase 2 depends on Phase 1
- Phase 3 depends on Phase 1 and Phase 2
- Phase 4 depends on Phase 1

## Risk Assessment

### Technical Risks
- **Incorrect ldflags path**: Typo in package path causes version not to be injected
  - Mitigation: Copy exact module path from go.mod, test with `make build && ./duofm --version`

### Implementation Risks
- **Missing replacement location**: One of the hardcoded strings might be missed
  - Mitigation: Use grep to verify no "v0.1.0" remains after changes

## Performance Considerations
- None: Variable access has negligible overhead

## Security Considerations
- None: Version string is non-sensitive information

## Open Questions
- None

## References
- [Specification](./SPEC.md)
- [Requirements](./要件定義書.md)
