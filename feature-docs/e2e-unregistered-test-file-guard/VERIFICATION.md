# Verification Document: e2e-unregistered-test-file-guard

## Overview

**Feature**: e2e-unregistered-test-file-guard / **SPEC.md**: `feature-docs/e2e-unregistered-test-file-guard/SPEC.md` / **IMPLEMENTATION.md**: `feature-docs/e2e-unregistered-test-file-guard/IMPLEMENTATION.md`

## Build Verification

- Command: `project.components.duofm.build_command` in `workflow.yaml` (runs `make build` in the Go container)
- Expected: exit code 0, no errors (regression check; no Go code changes)
- Command: `make test-e2e-build`
- Expected: the E2E image builds with exit code 0
- Command: `bash -n` on `test/e2e/scripts/helpers.sh`, `test/e2e/scripts/run_all_tests.sh` and `test/e2e/scripts/runner_self_test.sh`
- Expected: exit code 0 for each (no syntax errors)

## Test Verification

- Command: `project.components.duofm.test_command` in `workflow.yaml` (`go test ./...`)
- Expected: all Go tests pass (regression check; no Go code changes)
- Coverage target: not applicable (no Go code changes)
- Command: `bash test/e2e/scripts/runner_self_test.sh` on the host from the repository root
- Expected: exit code 0, one PASS line for each of TS-1 to TS-5

### Test Scenarios from SPEC.md

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-1 | Full run on a fixture: a registered file with one passing test and an unregistered file with one test, both tests in `FULL_RUN_LIST` | Exit code non-zero; Total 3 / Passed 1 / Failed 2 (unregistered file + test called while undefined); output names the unregistered file and its test function as failures | Integration (self-test) |
| TS-2 | `--check-list` on the TS-1 fixture | Exit code 1; output names the unregistered file; no OK line | Integration (self-test) |
| TS-3 | All fixture files registered and every defined test listed | `--check-list`: OK line, exit code 0; full run: Failed 0, exit code 0 | Integration (self-test) |
| TS-4 | Unregistered file whose test is not in `FULL_RUN_LIST`; `--check-list` | Exit code 1; output has a line reporting the unregistered file (distinct from the missing-entry line) and the existing `Missing from run list` line | Integration (self-test) |
| TS-5 | `run_test` called with a name that is not defined, after loading `helpers.sh` | TESTS_RUN +1, TESTS_FAILED +1, TESTS_PASSED unchanged; the name is shown as a failure | Unit (self-test) |
| TS-6 | Real repository (15 `tests/*.sh` files, all registered): `run_all_tests.sh --check-list` on the host, and the full E2E run | `--check-list` prints the OK line and exits 0; the full run output contains no guard-origin failure line (unregistered file, undefined run-list entry, defined test not in run list) | E2E (SPEC.md AC-4) |
| TS-7 | Self-test against the pre-change scripts: in a temporary directory, place the post-change `runner_self_test.sh` next to `run_all_tests.sh` and `helpers.sh` taken from the feature's base revision, run it; then run it in the repository | Pre-change: non-zero exit, TS-1, TS-2, TS-4 and TS-5 reported FAIL; post-change: exit 0, all PASS | Integration (SPEC.md AC-5) |
| TS-8 | `make test-e2e-build && make test-e2e` | The self-test's per-scenario results appear before the runner's `duofm E2E Tests - All` header; the run's exit code equals the runner's | E2E (SPEC.md AC-6) |
| TS-9 | Built E2E image run with its default command while a stand-in self-test that exits non-zero is mounted read-only over `/e2e/scripts/runner_self_test.sh` | The runner's header never appears; exit code non-zero | E2E (SPEC.md AC-6, A-1) |
| TS-10 | Self-test run with recording stand-ins for tmux and duofm placed first on the command search path | duofm is never invoked; tmux is invoked only for session cleanup (no session is created) | Integration (SPEC.md AC-7) |
| TS-11 | Self-test write scope: run on the host and, inside the E2E image as testuser, run it directly | Host: `git status` shows no change and `test/e2e/scripts/` is unchanged; image: recursive listings (names, sizes, modification times) of `/testdata` and `/e2e/scripts` are identical before and after; no self-test directory is left in the system temporary area | Integration + E2E |
| TS-12 | Dependency check | The Dockerfile diff leaves the base image and the package installation unchanged; the self-test passes inside the E2E image (TS-8) using only tools already there | Review + E2E |
| TS-13 | Existing CLI: `--help` and `--list` output compared with the base revision; a category run inside the image (runner with a category argument); `/e2e/scripts/interactive.sh` as an explicit container command; `tests/*.sh` in the diff | `--help` / `--list` byte-identical; the category run behaves as before without the self-test; the explicit command runs without the self-test; no `tests/*.sh` file changed | Integration + E2E |

## Code Quality Verification

- Format: `project.components.duofm.format_command` in `workflow.yaml` (`gofmt -w .`) — must produce no diff (no Go code changes)
- Static analysis: `bash -n` on the three scripts (see Build Verification)

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC-1 | Reproduction procedure ends with non-zero exit, Failed at least 1, and the unregistered file and its test function shown as failures | TS-1 |
| AC-2 | `--check-list` in the same state exits 1 and names the unregistered file | TS-2 |
| AC-3 | `run_test` with an undefined name adds exactly 1 to TESTS_RUN and TESTS_FAILED and leaves TESTS_PASSED unchanged | TS-5 |
| AC-4 | Current repository: `--check-list` prints OK and exits 0; no new guard-origin failures in the full run | TS-3, TS-6 |
| AC-5 | Self-test fails against the pre-change scripts and passes after the change | TS-7 |
| AC-6 | Every `make test-e2e-build && make test-e2e` runs the self-test; a failing self-test makes `make test-e2e` exit non-zero | TS-8, TS-9 |
| AC-7 | No tmux session and no duofm process is started during the self-test | TS-10 |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-1, TS-5 |
| FR2 | task0001 | TS-1, TS-2, TS-3, TS-4, TS-6 |
| FR3 | task0001 | TS-1, TS-3, TS-6 |
| FR4 | task0001 | TS-2, TS-3, TS-4, TS-6 |
| FR5 | task0001 | TS-1, TS-2, TS-3, TS-4, TS-5 (executed by the self-test), TS-7 |
| FR6 | task0002 | TS-8, TS-9 |
| NFR1 | task0001 | TS-10 |
| NFR2 | task0001, task0002 | TS-11 |
| NFR3 | task0001, task0002 | TS-12 |
| NFR4 | task0001, task0002 | TS-13 |

## E2E Testing

Framework: the project's Docker + tmux E2E suite. Command: `make test-e2e-build && make test-e2e`.

- [ ] Existing E2E tests pass without regression (no new failures compared with the base revision)
- [ ] TS-6: the full run output contains no guard-origin failure line
- [ ] TS-8: the self-test runs before the runner on every `make test-e2e`
- [ ] TS-9: a failing self-test stops the run before the runner, with a non-zero exit code
- [ ] TS-11 (image part): `/testdata` and `/e2e/scripts` unchanged by a self-test run as testuser
- [ ] TS-13 (image part): category run and `/e2e/scripts/interactive.sh` as explicit commands run without the self-test

## Manual Testing (E2E Not Possible)

- [ ] The self-test's per-scenario output is readable: each result line names the scenario ID and PASS or FAIL, and a failure shows what was expected and what was observed
- [ ] The design step was skipped, so there is no mockup comparison

## Performance / Security Verification (if applicable)

- NFR2 write scope: TS-11
- Performance: not applicable

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | Go build, E2E image build, bash syntax check | 2 | 1 | 0 |
| Tests | Go tests, self-test (TS-1 to TS-5) | 6 | 0 | 0 |
| Integrated scenarios | TS-6 to TS-13 | 4 (TS-7, TS-10, TS-11 host part, TS-13 host part) | 5 (TS-6, TS-8, TS-9, TS-11 image part, TS-13 image part) | 1 (TS-12 diff review) |
| Code quality | gofmt, bash syntax check | 2 | 0 | 0 |
| Manual | self-test output readability | 0 | 0 | 1 |
