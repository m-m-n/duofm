# Implementation Plan: fix-batch-cancel-cursor-flake

## Overview

The pane's marked-file queries (`GetMarkedFiles` / `GetMarkedFilePaths`) return marked entries in the pane's display order, with marked names that are absent from the display list appended in ascending name order. Unit tests pin that order, and the E2E test `test_batch_cancel_cursor` is hardened with a bounded wait for the overwrite dialog and post-cancel state checks. The whole feature is delivered by a single task (task0001).

## Technology Stack

- **Language**: Go (existing module; no toolchain change)
- **Unit tests**: the existing Go test setup of `internal/ui`
- **E2E**: the existing Bash + tmux harness under `test/e2e/scripts/`, run through `make test-e2e-build && make test-e2e`
- **New dependencies**: none. No license check is required; `project.license` (MIT) is unaffected.

## Layer Structure

| Layer | Responsibility | Change in this feature |
|-------|----------------|------------------------|
| `internal/ui` Pane (mark set + display list) | Owns the mark set and the filtered, sorted display list; answers marked-file queries | Ordering of the two marked-file queries only |
| `internal/ui` callers (batch copy/move, delete, trash, archive, permission, context menu) | Consume the marked-name / marked-path lists as given | None (they receive a deterministic order instead of an arbitrary one) |
| `test/e2e/scripts/tests` | Black-box checks through the terminal UI and the filesystem | `test_batch_cancel_cursor` only |

Dependency directions are unchanged. No layer gains a new dependency.

## Shared Components

None across tasks: this feature has exactly one task. The ordering contract of the two marked-file queries is pinned in `tasks/task0001.md` (Design), and the E2E hardening that relies on it lives in the same task.

## Conventions

- Exported signatures of `GetMarkedFiles` and `GetMarkedFilePaths` stay as they are (NFR1). Only the order of the returned elements changes; the returned set and count stay the same (NFR2).
- Go code stays gofmt-clean and go vet-clean (NFR4).
- New test code follows the existing file layout: unit tests go into the existing mark test file; E2E helper functions stay local to the E2E script that uses them (the shared E2E helper file is not modified).

## Cross-task Design Decisions

### D1: Single-task decomposition

- **Decision**: the ordering fix, its unit tests, and the E2E hardening form one task (task0001).
- **Rationale**: the hardened E2E assertions (first file moved, second file left in the source, destination content unchanged) hold deterministically only when the ordering fix is present. Tasks run in parallel in isolated worktrees, so a separate E2E task would run against the pre-fix, order-dependent behavior and fail intermittently inside its own worktree. The change spans three files.
- **Affected tasks**: task0001.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Every caller of the marked-file queries (batch copy/move, delete, trash, archive, permission, context menu) now receives display order instead of an arbitrary order | Certain | Low | Set and count are unchanged (NFR2); callers do not depend on a specific order; the full unit and E2E suites run in verify |
| The order unit test passes by chance against an unordered implementation | Low | Medium | At least three marked names, marked in the reverse of display order, with the check repeated many times in one run (task0001 AC-1) |
| The E2E bounded wait is too short in a slow container run | Low | Medium (false failure) | A bound of several seconds with short polling; a timeout is recorded as a failed assertion and never hangs the run |
| A new E2E helper is picked up by the runner as a test | Low | Low | Helper naming rule in task0001 (the runner treats functions whose definition line starts with the test prefix as tests) |

## Open Questions

- None. The number of repeated full E2E runs is decided in the verify phase (SPEC 14.1 A-repeat-e2e).
