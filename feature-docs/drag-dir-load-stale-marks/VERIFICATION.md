# Verification Document: drag-dir-load-stale-marks

## Overview

**Feature**: drag-dir-load-stale-marks / **SPEC.md**: `feature-docs/drag-dir-load-stale-marks/SPEC.md` / **IMPLEMENTATION.md**: not produced (reduced tier; a single task, no file shared between tasks)

## Build Verification

- Command (workflow.yaml `project.components.duofm.build_command`): `podman run --rm --userns=keep-id -v "$PWD":"$PWD" -w "$PWD" -v /home/sakura/go/src/duofm/.git:/home/sakura/go/src/duofm/.git:ro -v duofm-gomod:/gomod:U -v duofm-gocache:/gocache:U -e GOMODCACHE=/gomod -e GOCACHE=/gocache -e GOTOOLCHAIN=local -e DISPLAY=:0 -e GIT_CONFIG_COUNT=2 -e GIT_CONFIG_KEY_0=init.defaultBranch -e GIT_CONFIG_VALUE_0=main -e GIT_CONFIG_KEY_1=safe.directory -e GIT_CONFIG_VALUE_1=/home/sakura/go/src/duofm docker.io/library/golang:1.25.5-trixie make build`
- Expected: exit code 0, no errors

## Test Verification

- Command (workflow.yaml `project.components.duofm.test_command`): `podman run --rm --userns=keep-id -v "$PWD":"$PWD" -w "$PWD" -v /home/sakura/go/src/duofm/.git:/home/sakura/go/src/duofm/.git:ro -v duofm-gomod:/gomod:U -v duofm-gocache:/gocache:U -e GOMODCACHE=/gomod -e GOCACHE=/gocache -e GOTOOLCHAIN=local -e DISPLAY=:0 -e GIT_CONFIG_COUNT=1 -e GIT_CONFIG_KEY_0=init.defaultBranch -e GIT_CONFIG_VALUE_0=main docker.io/library/golang:1.25.5-trixie go test ./...`
- Coverage target: no project-wide percentage. Every branch of the changed decision in the directory load completion handler is exercised: success for the drag pane (TS-1, TS-2), success for the other pane (TS-3), error (TS-4), stale completion (TS-5).

### Test Scenarios from SPEC.md

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-1 | FR1, FR5 / AC1, AC6: press an entry in the left pane (session armed, no motion); send a successful completion for the left pane with a different entry list (no error; pane path matches or pending path empty); then move to a different row and release | Right after the completion the drag session is neither armed nor active; after the motion and release the left pane's mark set equals the set captured right after the completion | Unit |
| TS-2 | FR1, FR4, FR5 / AC2, AC6: press and move in the left pane (a range is marked); send a successful completion for the left pane with a different entry list; then move to a different row and release | Marks present before the completion remain; the session is reset; the later motion and release do not change the mark set | Unit |
| TS-3 | FR2 / AC3: press in the left pane; send a successful completion for the right pane; move in the left pane | The left session stays armed; the range from the original anchor is marked in the left pane | Unit |
| TS-4 | FR3 / AC4: press in the left pane; send a completion for the left pane that carries an error; move | The session is unchanged; the range from the original anchor is marked | Unit |
| TS-5 | FR3 / AC5: press in the left pane; set the left pane's pending path; send a completion for the left pane with a different pane path | The session and the left pane's entries are unchanged | Unit |
| TS-6 | NFR1, NFR2, NFR3 / AC7: run the full unit test suite (test command above) and the E2E suites (`make test-e2e-build && make test-e2e`, including `mouse_tests.sh` and `mark_tests.sh`) | All tests pass | Unit / E2E |
| TS-7 | NFR3 / AC7: run the format command and `go vet ./...` | The format command leaves no diff; `go vet` reports no findings | Static |

## Code Quality Verification

- Format (workflow.yaml `project.components.duofm.format_command`): `podman run --rm --userns=keep-id -v "$PWD":"$PWD" -w "$PWD" -v /home/sakura/go/src/duofm/.git:/home/sakura/go/src/duofm/.git:ro -v duofm-gomod:/gomod:U -v duofm-gocache:/gocache:U -e GOMODCACHE=/gomod -e GOCACHE=/gocache -e GOTOOLCHAIN=local -e DISPLAY=:0 -e GIT_CONFIG_COUNT=1 -e GIT_CONFIG_KEY_0=init.defaultBranch -e GIT_CONFIG_VALUE_0=main docker.io/library/golang:1.25.5-trixie gofmt -w .` — expected: `git status --porcelain` shows no change afterwards
- Static analysis: `go vet ./...` in the same container as the test command (the test command with its trailing `go test ./...` replaced by `go vet ./...`) — expected: exit code 0, no findings

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| SC-1 | All functional requirements are implemented and tested | Functional Requirements Coverage below; TS-1 to TS-5 pass |
| SC-2 | All test scenarios pass | TS-1 to TS-7 |
| SC-3 | Code review is completed | review phase status in workflow.yaml |
| SC-4 | The reproduction steps no longer produce the defect | M-1 |
| SC-5 | A test detects recurrence | TS-1 and TS-2 fail against the pre-fix handler and pass after the fix (task0001 AC-6) |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-1, TS-2 |
| FR2 | task0001 | TS-3 |
| FR3 | task0001 | TS-4, TS-5 |
| FR4 | task0001 | TS-2 |
| FR5 | task0001 | TS-1, TS-2 (fail before the fix, pass after) |
| NFR1 | task0001 | TS-6 |
| NFR2 | task0001 | TS-6 |
| NFR3 | task0001 | TS-6, TS-7 |

## E2E Testing

Existing shell-based E2E suites, run with `make test-e2e-build && make test-e2e`. No new E2E scenario for this feature.

- [ ] TS-6 (E2E part): the existing E2E suites, including `mouse_tests.sh` and `mark_tests.sh`, pass without regression

## Manual Testing (E2E Not Possible)

- [ ] M-1: Open a directory whose load takes long enough to press before it completes. Press on an entry in that pane before the load completes. After the load completes, move the pointer to another row and release. Expected: the gesture marks no file, and marks that existed before the press are unchanged.

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Unit tests (new) | 5 (TS-1 to TS-5) | 5 | 0 | 0 |
| Regression suites | 1 (TS-6) | 1 | 1 | 0 |
| Static checks | 1 (TS-7) | 1 | 0 | 0 |
| Manual | 1 (M-1) | 0 | 0 | 1 |
