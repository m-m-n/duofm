# Feature: Dynamic Version Display in Toolbar

## Overview

Enable the TUI toolbar to display the dynamically injected version from build-time ldflags, matching the `--version` CLI output. Currently, the toolbar shows a hardcoded "v0.1.0" while `--version` correctly shows the git-tag-derived version.

## Domain Rules

- **Single Source of Truth**: Version string must be defined in exactly one location
- **Build-time Injection**: Version is set via Go's `-ldflags -X` mechanism at compile time
- **Consistency**: CLI `--version` and toolbar must always show identical version strings

## Objectives

- Eliminate hardcoded version strings from the codebase
- Share version information between main package and UI package
- Maintain existing build process compatibility

## User Stories

- As a user, I want to see the correct version in the TUI toolbar, so that I know which version I'm running
- As a developer, I want version to be defined in one place, so that updates don't require changes in multiple files

## Functional Requirements

- **FR1.1**: Toolbar displays version string from build-time injected variable
- **FR1.2**: `--version` CLI option displays same version as toolbar
- **FR1.3**: Version variable uses "dev" as default when not set via ldflags
- **FR1.4**: Title bar format remains "duofm <version>" (e.g., "duofm v1.0.0")

## Non-Functional Requirements

- **NFR1.1**: No changes to existing Makefile ldflags injection pattern
- **NFR1.2**: No circular import dependencies between packages
- **NFR1.3**: Build must succeed with standard `make build` command

## Interface Contract

### Version Variable

- **Type**: `string`
- **Default Value**: `"dev"`
- **Injected Value Format**: Git tag (e.g., "v1.0.0") or "dev-<short-commit>"
- **Display Format in Toolbar**: `"duofm <version>"` (e.g., "duofm v1.0.0" or "duofm dev-a1b2c3d")

### Ldflags Injection

- **Current Pattern**: `-X main.version=$(VERSION)`
- **Target Pattern**: `-X github.com/sakura/duofm/internal/version.Version=$(VERSION)`

## Dependencies

- Makefile: Update ldflags to reference new package path
- main.go: Import version package
- internal/ui/model.go: Import version package and replace hardcoded strings

## Test Scenarios

- [ ] Build with `make build` and run `./duofm --version` shows git tag or dev-commit
- [ ] TUI toolbar displays same version as `--version` output
- [ ] Build without git tag shows "dev-<commit>" format
- [ ] Default `go build` (no ldflags) shows "dev"

## Success Criteria

- [ ] No hardcoded version strings remain in source code
- [ ] `duofm --version` matches toolbar version exactly
- [ ] Existing Makefile workflow continues to work
- [ ] All existing tests pass

## Constraints

- Must use Go's `-ldflags -X` mechanism (no version files or config)
- Cannot create import cycles (version package must be leaf dependency)
- Version package path must be valid for ldflags injection
