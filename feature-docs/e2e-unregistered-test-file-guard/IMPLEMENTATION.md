# Implementation Plan: e2e-unregistered-test-file-guard

## Overview

Make the E2E runner count tests that live in `tests/*.sh` files not registered in `TEST_FILES` (instead of silently dropping them), prove it with a bash self-test of the runner, and run that self-test from the E2E image's default command before the suite.

## Technology Stack

- **Language**: Bash (the existing E2E scripts under `test/e2e/scripts/`)
- **Container**: the existing E2E image defined by `test/e2e/Dockerfile` (base image, installed packages and build steps unchanged)
- **New dependencies**: none (NFR3). No license record is needed; `project.license` (MIT) is unaffected.

## Layer Structure

| Layer | Files | Responsibility |
|-------|-------|----------------|
| Test definitions | `test/e2e/scripts/tests/*.sh` | Define `test_` functions; load `helpers.sh` relative to their own location |
| Shared helpers | `test/e2e/scripts/helpers.sh` | Tally counters (TESTS_RUN / TESTS_PASSED / TESTS_FAILED), assertions, `run_test`, `print_summary` |
| Runner | `test/e2e/scripts/run_all_tests.sh` | Registry (`TEST_FILES`), run list (`FULL_RUN_LIST`), run-list guard, CLI modes (full run, category, `--list`, `--help`, `--check-list`) |
| Runner self-test | `test/e2e/scripts/runner_self_test.sh` (new) | Verify the runner and helpers against fixtures, using copies placed in its own temporary directory |
| Image entry | `test/e2e/Dockerfile` default command | Run the self-test, then the runner |

Allowed dependency directions:

- Image entry → Runner self-test, Runner (invocation by path only).
- Runner self-test → Runner, Shared helpers (it reads them to make copies and exercises only the copies; it never modifies or executes the originals in place).
- Runner → Shared helpers, Test definitions.
- Nothing in Runner, Shared helpers or Test definitions depends on the Runner self-test.

## Shared Components

| Component | Responsibility | Contract (pre/postcondition) | Used by tasks |
|-----------|----------------|------------------------------|---------------|
| Runner self-test entry point: repository path `test/e2e/scripts/runner_self_test.sh`, in-image path `/e2e/scripts/runner_self_test.sh` | Run scenarios TS-1 to TS-5 against copies of the runner and helpers that sit next to it | **Pre**: invoked with no arguments; any working directory; any user that can create a directory in the system temporary area (in the image: testuser with WORKDIR `/testdata`); `run_all_tests.sh` and `helpers.sh` are in the same directory as the self-test. **Post**: exit status 0 if and only if every scenario passed; non-zero on any scenario failure and on any setup failure; one result line per scenario on standard output; nothing written outside a temporary directory the self-test created itself; no tmux session and no duofm process started; its own directory is left unchanged. | task0001 (creates), task0002 (invokes) |
| Runner full-run entry point: repository path `test/e2e/scripts/run_all_tests.sh`, in-image path `/e2e/scripts/run_all_tests.sh`, no arguments | Execute the full E2E run and print the summary | **Pre**: invoked with no arguments (unchanged). **Post**: exit status 0 if and only if the summary's Failed count is 0 (unchanged rule; after this feature the Failed count also includes unregistered files and undefined names passed to `run_test`). | task0001 (changes what is counted), task0002 (invokes) |

## Conventions

- Scripts stay plain bash and use only commands already present in the E2E image (NFR3). No new package, tool or image layer is added.
- The exit status is the only pass/fail signal between the self-test and the image's default command; output text is for humans only.
- New script files follow the header-comment style of the sibling scripts (purpose, usage). Comments are in English.

## Cross-task Design Decisions

### D1: Self-test location and name

The self-test is `test/e2e/scripts/runner_self_test.sh`, outside `tests/`, so the runner's scan of `tests/*.sh` (definition set and unregistered-file detection) never sees it. The existing Dockerfile step that copies `test/e2e/scripts` into `/e2e/scripts` and recursively marks it executable already places it in the image at `/e2e/scripts/runner_self_test.sh`; no copy step is added.

Affected tasks: task0001 (creates the file at this path), task0002 (references the in-image path).

### D2: Gate order in the E2E image

The image's default command runs the self-test first. Only when it exits 0 does the runner run (no arguments); the container's exit status is then the runner's. When the self-test exits non-zero, or cannot be executed, the runner is not run and the container exits non-zero (A-1). The change stays a default command that an explicit command given to the container replaces entirely, so existing explicit invocations (for example `/e2e/scripts/interactive.sh`, or the runner with a category argument) run without the self-test. The Makefile is not changed: `make test-e2e` runs the image without a command, so the default command applies.

Affected tasks: task0002 (implements), task0001 (its self-test must honor the exit-status contract above).

### D3: The self-test ships with the runner fix

The self-test's scenarios are the TDD tests of the runner and helper changes, so both live in the same task. The image wiring is a separate task that depends only on the path in D1 and the exit-status contract in Shared Components. There is no placeholder: task0001 creates the real file at the pinned path and task0002 references that path; the integrated behavior is checked in the verify phase (VERIFICATION.md TS-8, TS-9).

Affected tasks: task0001, task0002.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| The self-test passes vacuously (fixture lists not applied to the runner copy, or tolerant output parsing) | Medium | High | task0001: fail closed when fixture lists cannot be applied, exact count assertions, distinctive failure phrases asserted; VERIFICATION.md TS-7 runs the self-test against the pre-fix scripts and expects failure |
| Summary counts are wrapped in terminal color sequences and are misread | Medium | Medium | task0001 Test Notes: normalize output before extracting counts; an unreadable count is a scenario failure |
| task0002's worktree does not contain the self-test yet (tasks run in parallel) | High | Low | task0002 verifies the command sequencing with stand-in scripts; the integrated run is VERIFICATION.md TS-8 / TS-9 |
| The default-command change breaks explicit container commands documented in `test/README.md` | Low | Medium | D2: remain an overridable default command; task0002 acceptance criterion |
| New guard code aborts the runner under its exit-on-error mode | Low | High | task0001: new checks run only in condition context or with non-failing statements; covered by the TS-1 / TS-3 full-run scenarios |
| The real repository gains a false positive | Low | Medium | VERIFICATION.md TS-6 (`--check-list` on the real 15 files) |

## Open Questions

- None.
