# Verification Document: press-scroll-click-mark

## Overview

**Feature**: press-scroll-click-mark / **SPEC.md**: `feature-docs/press-scroll-click-mark/SPEC.md` / **IMPLEMENTATION.md**: `feature-docs/press-scroll-click-mark/IMPLEMENTATION.md`

## Build Verification

- Command: `podman run --rm --userns=keep-id -v "$PWD":"$PWD" -w "$PWD" -v /home/sakura/go/src/duofm/.git:/home/sakura/go/src/duofm/.git:ro -v duofm-gomod:/gomod:U -v duofm-gocache:/gocache:U -e GOMODCACHE=/gomod -e GOCACHE=/gocache -e GOTOOLCHAIN=local -e DISPLAY=:0 -e GIT_CONFIG_COUNT=2 -e GIT_CONFIG_KEY_0=init.defaultBranch -e GIT_CONFIG_VALUE_0=main -e GIT_CONFIG_KEY_1=safe.directory -e GIT_CONFIG_VALUE_1=/home/sakura/go/src/duofm docker.io/library/golang:1.25.5-trixie make build`
- Expected: exit code 0, no errors

## Test Verification

- Command: `podman run --rm --userns=keep-id -v "$PWD":"$PWD" -w "$PWD" -v /home/sakura/go/src/duofm/.git:/home/sakura/go/src/duofm/.git:ro -v duofm-gomod:/gomod:U -v duofm-gocache:/gocache:U -e GOMODCACHE=/gomod -e GOCACHE=/gocache -e GOTOOLCHAIN=local -e DISPLAY=:0 -e GIT_CONFIG_COUNT=1 -e GIT_CONFIG_KEY_0=init.defaultBranch -e GIT_CONFIG_VALUE_0=main docker.io/library/golang:1.25.5-trixie go test ./...`
- Expected: exit code 0, all tests pass
- Coverage target: minimum — `internal/ui` package coverage not lower than on the base branch; target — every branch of the changed entry-press handler exercised by the new tests

### Test Scenarios from SPEC.md

Fixture state (TS-1 to TS-8, TS-10): 80x24; left pane active; right pane inactive over 20 files (21 entries with ".."), right cursor 15, right scroll offset 0; background command attached to the right pane and running. Rows are computed with scroll offset 0.

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-1 | Fixture state; Model.Update with a left press on right-pane index 7's row, then a release on the same row | Right mark set unchanged, right cursor 7, right scroll offset 0, entry 7 on the pressed row; the test fails on the pre-fix code | Unit |
| TS-2 | TS-1 with the background state closing instead of running | Same result as TS-1 | Unit |
| TS-3 | Table-driven: pressed index above the shifted window (3), inside the kept range (7), below the shifted window on a pre-press visible row (16); same-row press/release | No marks added; cursor on the pressed index; scroll offset 0 for the first two rows | Unit |
| TS-4 | Press index 13 (at or beyond pre-press offset + 11 visible lines); same-row release | Cursor 13 on the bottom visible file-list row (scroll offset 3); no marks added | Unit |
| TS-5 | Right-pane entries marked before the press (inside and outside 7..12), then TS-1 | Mark set exactly equal to the pre-press set | Unit |
| TS-6 | Same-row press/release on the right pane's ".." row | No marks; cursor at index 0 on ".." | Unit |
| TS-7 | Press index 7, move to index 9's row, release. Second case: press 7, move to 9's row, move back to 7's row, release | First: pre-press set plus entries 7..9 exactly. Second: pre-press set plus entry 7 only, cursor 7 | Unit |
| TS-8 | Fixture with a directory at right-pane index 7, fake clock: press, release, second press on the same row within the double-click window | Enter action on that directory: non-nil command, right pane path becomes the directory and starts loading | Unit |
| TS-9 | Full unit suite and E2E suite (including mouse_tests.sh and mark_tests.sh) | All pass | Regression / E2E |
| TS-10 | Fixture state; keyboard move-right action switches to the right pane | Right cursor 15, right scroll offset 5 (current behavior kept) | Unit |
| TS-11 | Diff of the feature against the base branch | Pane.adjustScroll, Pane.SetBgOutputActive, Model.syncPaneBgOutputState, Model.switchToPane unchanged; `internal/ui/model.go` and `internal/ui/pane.go` not in the diff | Static (diff inspection) |
| TS-12 | gofmt list mode and go vet over the module | gofmt lists no files; go vet reports nothing | Static |

## Code Quality Verification

- Format: the format_command container invocation with `gofmt -l .` in place of `gofmt -w .` — expected: empty output
- Static analysis: the test_command container invocation with `go vet ./...` in place of `go test ./...` — expected: exit code 0, no findings

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| SC-1 | All functional requirements are implemented and tested | Functional Requirements Coverage table: every row has a task and passing scenarios |
| SC-2 | All test scenarios pass | TS-1 to TS-12 |
| SC-3 | Code review is completed | Review phase completed with no residual critical/high findings |

SPEC.md acceptance criteria mapping:

| SPEC AC | Scenario |
|---------|----------|
| AC-1 | TS-1 |
| AC-2 | TS-1 |
| AC-3 | TS-2 |
| AC-4 | TS-3 |
| AC-5 | TS-4 |
| AC-6 | TS-5 |
| AC-7 | TS-6 |
| AC-8 | TS-7 |
| AC-9 | TS-8 |
| AC-10 | TS-1 |
| AC-11 | TS-9 |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-1, TS-2, TS-3, TS-4, TS-5, TS-6 |
| FR2 | task0001 | TS-1, TS-2, TS-3, TS-4 |
| FR3 | task0001 | TS-7 |
| FR4 | task0001 | TS-8 |
| FR5 | task0001 | TS-1 (fails on the pre-fix code) |
| NFR1 | task0001 | TS-10 |
| NFR2 | task0001 | TS-10, TS-11 |
| NFR3 | task0001 | TS-9 |
| NFR4 | task0001 | TS-9, TS-12 |

## E2E Testing

- Framework: shell-script E2E under `test/e2e/`
- Command: `make test-e2e-build && make test-e2e`
- [ ] Existing E2E tests pass without regression, including `test/e2e/scripts/tests/mouse_tests.sh` and `test/e2e/scripts/tests/mark_tests.sh` (TS-9)

## Manual Testing (E2E Not Possible)

- [ ] In a real 80x24 terminal, with a background command running on the right pane, the left pane active, and the right pane's cursor 11 or more rows below its scroll offset: click (without moving the mouse) an entry of the right pane shown within the top 11 entry rows. No mark is added, the cursor moves to the clicked entry, and the list does not scroll.
- [ ] In the same state, double-click an entry shown within the top 11 entry rows of the right pane. The Enter action runs on that entry.

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Unit scenarios | 9 (TS-1 to TS-8, TS-10) | 9 | 0 | 0 |
| Regression / E2E | 1 (TS-9) | 1 | 1 | 0 |
| Static | 2 (TS-11, TS-12) | 1 | 0 | 1 (TS-11 diff inspection) |
| Manual reproduction | 2 | 0 | 0 | 2 |
