# Verification Document: fix-e2e-failures

## Overview

**Feature**: fix-e2e-failures / **SPEC.md**: `feature-docs/fix-e2e-failures/SPEC.md` / **IMPLEMENTATION.md**: `feature-docs/fix-e2e-failures/IMPLEMENTATION.md`

Only E2E shell scripts change: `test/e2e/scripts/run_all_tests.sh` and `test/e2e/scripts/tests/{archive,file_operation,mark,sort,shell,bookmark,history}_tests.sh`.

**Environment note**: the E2E command needs Docker. Docker is not available in the environment this plan was written in. Items marked "E2E" need a host with Docker. All other items run without Docker.

## Build Verification

- Command: `podman run --rm --userns=keep-id -v "$PWD":"$PWD" -w "$PWD" -v /home/sakura/go/src/duofm/.git:/home/sakura/go/src/duofm/.git:ro -v duofm-gomod:/gomod:U -v duofm-gocache:/gocache:U -e GOMODCACHE=/gomod -e GOCACHE=/gocache -e GOTOOLCHAIN=local -e DISPLAY=:0 -e GIT_CONFIG_COUNT=2 -e GIT_CONFIG_KEY_0=init.defaultBranch -e GIT_CONFIG_VALUE_0=main -e GIT_CONFIG_KEY_1=safe.directory -e GIT_CONFIG_VALUE_1=/home/sakura/go/src/duofm docker.io/library/golang:1.25.5-trixie make build`
- Expected: exit code 0, no errors

## Test Verification

- Command: `podman run --rm --userns=keep-id -v "$PWD":"$PWD" -w "$PWD" -v /home/sakura/go/src/duofm/.git:/home/sakura/go/src/duofm/.git:ro -v duofm-gomod:/gomod:U -v duofm-gocache:/gocache:U -e GOMODCACHE=/gomod -e GOCACHE=/gocache -e GOTOOLCHAIN=local -e DISPLAY=:0 -e GIT_CONFIG_COUNT=1 -e GIT_CONFIG_KEY_0=init.defaultBranch -e GIT_CONFIG_VALUE_0=main docker.io/library/golang:1.25.5-trixie go test ./...`
- Coverage target: N/A (no Go source changes)

### Test Scenarios from SPEC.md

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS1 | Run the full E2E suite (`make test-e2e-build && make test-e2e`) | Summary shows Failed: 0. Every previously failing test in the Archive, File Operation, Mark, Sort, Shell, Bookmark and History groups passes | E2E |
| TS2 | Add an undefined name to the run list in a scratch copy of `test/e2e/scripts`, then (a) run `run_all_tests.sh --check-list` on the copy, and (b) on a Docker host, run the full suite with the modified runner | (a) non-zero exit, the name is printed. (b) the name is not invoked, and it adds 1 to Total and 1 to Failed | Static (a) / E2E (b) |
| TS3 | Remove one defined test from the run list in a scratch copy, then (a) `--check-list` on the copy, and (b) the full suite with the modified runner on a Docker host | (a) non-zero exit, the missing function and its file are printed. (b) the run exits non-zero and names the missing test | Static (a) / E2E (b) |
| TS4 | Sort dialog open, press `q`. Then dropdown expanded, press `q` (`test_sort_dialog_q_cancel`, `test_sort_dialog_q_cancel_with_dropdown` in the full run) | The dialog is still shown in both cases, and both tests pass | E2E |
| TS5 | After the full run, list `/testdata/user_owned` in the same container | None of the archive-test or rename-test fixtures or outputs remain (sources, `*.tar`/`*.tar.gz` outputs from these tests, `extract_dest`, `conflictdir`, `before_rename.txt`, `after_rename.txt`, `navren_before.txt`, `navren_after.txt`) | E2E |
| TS6 | `go test ./...` (Test Verification command) | Passes | Unit |
| TS7 | Bash syntax check (`bash -n`) on `run_all_tests.sh` and every edited `tests/*.sh`; shellcheck on the same files if installed | Syntax check passes for every file; shellcheck reports no findings on changed lines | Static |
| TS8 | `test/e2e/scripts/run_all_tests.sh --check-list` on the committed scripts, and inspection of the run list | Exit 0. The run list has 120 entries, includes the 9 previously unrun tests, and has none of `test_sort_dialog_hl_navigation`, `test_sort_dialog_jk_navigation`, `test_sort_dialog_confirm` | Static |
| TS9 | Diff of the integration branch against the implement base commit | No Go source, `test/e2e/Dockerfile`, `test/e2e/scripts/helpers.sh` or `Makefile` changes. `test/e2e/scripts/run_tests.sh` is unchanged and still a symlink to `run_all_tests.sh` | Static |
| TS10 | On a Docker host, run each category alone in a fresh container (archive, file-ops, mark, sort, shell, bookmark, history) through `run_all_tests.sh <category>` | Each category passes with Failed: 0 | E2E |
| TS11 | On a Docker host, in a fresh container, create extra files in `/testdata/user_owned` that sort before `del1.txt` (for example `aaa_leftover.txt`), then run the mark category | `test_batch_delete_marked_files` passes | E2E |
| TS12 | In the TS1 summary, compare Total with Passed + Failed | Total equals Passed + Failed | E2E |

## Code Quality Verification

- Format: `podman run --rm --userns=keep-id -v "$PWD":"$PWD" -w "$PWD" -v /home/sakura/go/src/duofm/.git:/home/sakura/go/src/duofm/.git:ro -v duofm-gomod:/gomod:U -v duofm-gocache:/gocache:U -e GOMODCACHE=/gomod -e GOCACHE=/gocache -e GOTOOLCHAIN=local -e DISPLAY=:0 -e GIT_CONFIG_COUNT=1 -e GIT_CONFIG_KEY_0=init.defaultBranch -e GIT_CONFIG_VALUE_0=main docker.io/library/golang:1.25.5-trixie gofmt -w .` (expected: no diff, since no Go changes)
- Static analysis (shell): TS7

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC1 | The full run `make test-e2e-build && make test-e2e` reports Failed: 0 | TS1 |
| AC2 | The test summary includes the 9 tests that did not run before | TS8 (list), TS1 (the 9 names appear as "Running" lines in the full-run output) |
| AC3 | An undefined test function name in the run list is reported as a failure | TS2 |
| AC4 | A defined `test_` function missing from the run list fails the run | TS3 |
| AC5 | Both q tests pass only while the sort dialog stays open after `q`, normal and with dropdown | TS4, plus review of the two tests (presence checks only) |
| AC6 | After the full run, the archive and rename test files are gone from `/testdata/user_owned` | TS5 |
| AC7 | No Go, Dockerfile or run_tests.sh changes, and `go test ./...` passes | TS9, TS6 |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0002 | TS1, TS5, TS7, TS10 |
| FR2 | task0003 | TS1, TS5, TS7, TS10 |
| FR3 | task0003 | TS1, TS7, TS10, TS11 |
| FR4 | task0004 | TS1, TS4, TS7 |
| FR5 | task0005 | TS1, TS7 |
| FR6 | task0005 | TS1, TS7 |
| FR7 | task0005 | TS1, TS7 |
| FR8 | task0001 | TS2, TS3, TS7, TS8 |
| FR9 | task0002, task0003, task0004, task0005 | TS5, TS12 |
| NFR1 | task0001, task0002, task0003, task0004, task0005 | TS9 |
| NFR2 | task0001 | TS9 |
| NFR3 | task0001, task0002, task0003, task0004, task0005 | TS6 |
| NFR4 | task0002, task0003 | TS5, TS10, TS11 |

## E2E Testing

Framework: tmux-driven Bash scripts in `test/e2e/scripts`, run in the image from `test/e2e/Dockerfile`.
Run command: `make test-e2e-build && make test-e2e`

- [ ] TS1 full run: Failed: 0
- [ ] TS2 (b) undefined name counted as a failure in the full run
- [ ] TS3 (b) missing defined test fails the full run and is named
- [ ] TS4 sort `q` tests pass
- [ ] TS5 no archive or rename fixtures left in `/testdata/user_owned`
- [ ] TS10 each edited category passes alone
- [ ] TS11 batch delete passes with leftover files present
- [ ] TS12 Total = Passed + Failed

Static items that need no Docker: TS2 (a), TS3 (a), TS6, TS7, TS8, TS9.

## Manual Testing (E2E Not Possible)

- [ ] Review the edited tests for checks that pass regardless of behavior (FR9). The review scope is the tests named in task0002 to task0005.
- [ ] If Docker is unavailable at verify time, record TS1, TS2 (b), TS3 (b), TS4, TS5 and TS10 to TS12 as not executed and run them on a host with Docker.

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build / unit | TS6 + build | 2 | 0 | 0 |
| Static (shell, scope) | TS2a, TS3a, TS7, TS8, TS9 | 5 | 0 | 0 |
| E2E | TS1, TS2b, TS3b, TS4, TS5, TS10, TS11, TS12 | 0 | 8 | 0 |
| Review | FR9 strictness review | 0 | 0 | 1 |
