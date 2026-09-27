# Verification Document: runner-self-test-tmpfile-scope

## Overview
**Feature**: runner-self-test-tmpfile-scope / **SPEC.md**: `feature-docs/runner-self-test-tmpfile-scope/SPEC.md` / **IMPLEMENTATION.md**: not produced (reduced tier; single task, no file shared between tasks)

Task plan: `feature-docs/runner-self-test-tmpfile-scope/tasks/task0001.md`

## Build Verification
- Command: `project.components.duofm.build_command` in workflow.yaml (containerized `make build`)
- Expected: exit code 0, no errors. The feature changes no Go source; this is a no-regression check.

## Test Verification
- Command (primary): `bash test/e2e/scripts/runner_self_test.sh` (`project.components.runner-self-test.test_command`)
- Command (regression): `project.components.duofm.test_command` in workflow.yaml (containerized `go test ./...`)
- Expected: exit code 0 for both
- Coverage target: not applicable. No coverage tooling is configured for the bash runner-self-test component, and no Go code changes.

### Test Scenarios from SPEC.md
| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-6 | An empty sentinel directory is created inside SELF_TMPDIR and exported as TMPDIR; apply_fixture_lists is called with a runner-copy path that does not exist, so both rewrite steps fail after creating their intermediate files and no move happens; the sentinel directory is then checked | The sentinel directory exists and has no entries; the self-test prints the TS-6 line ending in ": PASS" | Self-test (automated) |
| MUT-1 | Verification-time mutation: one intermediate-file creation site in apply_fixture_lists at a time is reverted to the pre-fix form quoted in SPEC.md (directory taken from the environment), the self-test is run, and the revert is undone | For each of the two sites: TS-6 prints ": FAIL" and the self-test exits non-zero; after restoration no diff remains | Mutation (verification-time) |
| REG-1 | The self-test is run directly, and through the E2E container's default command | In both runs all six self-test scenarios report PASS, TS-6 included, and the self-test exits 0; in the container the E2E suite then runs as before | Self-test + E2E |

### Verification-time Check Procedures

**MUT-1** (SPEC.md AC-2)
1. In apply_fixture_lists, revert only the first intermediate-file creation to the pre-fix form. Run `bash test/e2e/scripts/runner_self_test.sh`. Expect a TS-6 line ending in ": FAIL" and a non-zero exit status.
2. Restore the first site. Repeat step 1 for the second creation site only.
3. Restore the second site and confirm that no diff remains against the implemented version.

**Whole-run write scope** (NFR1)
1. Create a fresh empty directory E outside the repository.
2. Run `bash test/e2e/scripts/runner_self_test.sh` with TMPDIR set to E.
3. After the script exits, E is empty (SELF_TMPDIR, created under E, was removed by the EXIT trap and nothing else was left), and `git status --porcelain` in the worktree reports no change caused by the run.

## Code Quality Verification
- Format: `project.components.duofm.format_command` in workflow.yaml (containerized `gofmt -w .`). No Go file changes; expect no resulting diff. No formatter is configured for the bash component.
- Static analysis: none configured in workflow.yaml for either component. The bash script's syntax is exercised by REG-1.

## SPEC.md Compliance
### Success Criteria
| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC-1 | With an empty directory D exported as TMPDIR, apply_fixture_lists failing after temporary-file creation leaves D empty (no tmp.XXXXXXXXXX entry) | TS-6 PASS |
| AC-2 | After the fix the self-test prints the TS-6 PASS line; reverting either creation site to the pre-fix form makes TS-6 FAIL and the self-test exit non-zero | TS-6 PASS on the implemented code; MUT-1 |
| AC-3 | `bash test/e2e/scripts/runner_self_test.sh` reports PASS for TS-1 to TS-6 and exits 0 | REG-1 (direct run) |
| AC-4 | `make test-e2e-build && make test-e2e` succeeds; the container default command runs the self-test, TS-6 included, before the E2E suite | REG-1 (container run) through the E2E command |

### Functional Requirements Coverage
| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-6 (intermediate files no longer land in TMPDIR); MUT-1 |
| FR2 | task0001 | TS-6 PASS on the implemented code; MUT-1 (TS-6 fails for either reverted site) |
| NFR1 | task0001 | TS-6 (sentinel and helper writes stay under SELF_TMPDIR); REG-1; whole-run write-scope check |
| NFR2 | task0001 | TS-6 scenario inspection: no tmux session, no duofm process, no runner-copy execution (Manual Testing) |
| NFR3 | task0001 | REG-1 (direct and container runs); diff inspection that the TS-1 to TS-5 assertions and the Dockerfile default command are unchanged (Manual Testing) |

## E2E Testing
- Command: `make test-e2e-build && make test-e2e` (`project.components.duofm.e2e_test_command`)
- [ ] The E2E image builds, and its default command runs the runner self-test, TS-6 included, as the non-root user testuser before the E2E suite; every self-test scenario reports PASS (REG-1, container run)
- [ ] Existing E2E tests pass without regression

## Manual Testing (E2E Not Possible)
- [ ] NFR2: the TS-6 scenario body starts no tmux session and no duofm process, and does not execute the runner copy (code inspection)
- [ ] NFR3: the diff leaves the assertions of TS-1 to TS-5, the self-test's exit-status rule (0 only when every scenario passes), and the Dockerfile default command chain unchanged (diff inspection)

## Verification Summary
| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Test commands (self-test, go test) | 2 | 2 | 0 | 0 |
| SPEC scenarios (TS-6, MUT-1, REG-1) | 3 | 1 (TS-6; REG-1 direct run via the self-test command) | 1 (REG-1 container run) | 1 (MUT-1) |
| Verification-time checks (whole-run write scope) | 1 | 0 | 0 | 1 |
| Code quality | 1 | 1 | 0 | 0 |
| E2E | 2 | 0 | 2 | 0 |
| Inspection (NFR2, NFR3) | 2 | 0 | 0 | 2 |
