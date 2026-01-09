# Feature: Show Active Git Branch

## Overview

Git管理下のディレクトリを表示している場合に、現在のGitブランチ名をペインのヘッダー1行目の右端に表示する機能。

**Key Benefits:**
- 現在のブランチを常に把握できる
- 誤ったブランチでの作業を防止
- 左右ペインで異なるリポジトリの状態を同時に確認可能

## Objectives

- Git管理下のディレクトリでブランチ名を `[branch]` 形式で表示
- ブランチ名をヘッダー1行目の右端に右寄せで配置
- Git管理外のディレクトリでは何も表示しない
- 各ペインが独立してブランチ情報を管理・表示
- ブランチ取得のパフォーマンスを100ms以内に維持

## User Stories

### US1: ブランチ名の確認
As a 開発者 navigating a Git repository, I want to see the current branch name, so that I know which branch I'm working on.

**Acceptance Criteria:**
- [ ] Git管理下のディレクトリで `[main]` のようにブランチ名が表示される
- [ ] ブランチ名はヘッダー1行目の右端に配置される
- [ ] ディレクトリ変更時にブランチ名が更新される
- [ ] 表示が100ms以内に更新される

### US2: 非Gitディレクトリでの動作
As a ユーザー, I want Git-unmanaged directories to show no branch info, so that the display remains clean.

**Acceptance Criteria:**
- [ ] Git管理外のディレクトリではブランチ表示がない
- [ ] エラーメッセージやプレースホルダーが表示されない
- [ ] Git未インストール環境でもエラーなく動作する

### US3: デュアルペイン操作
As a 開発者, I want each pane to show its own branch independently, so that I can work with multiple repositories.

**Acceptance Criteria:**
- [ ] 左右のペインが独立してブランチを表示
- [ ] 一方のペインのディレクトリ変更が他方に影響しない

## Technical Requirements

### Functional Requirements

- **FR1.1:** System SHALL display current Git branch name in `[branch]` format
- **FR1.2:** System SHALL position branch name at the right end of header line 1
- **FR1.3:** System SHALL use right-alignment for branch display
- **FR1.4:** System SHALL truncate path (not branch) when space is insufficient
- **FR1.5:** System SHALL hide branch display for non-Git directories
- **FR1.6:** System SHALL update branch when directory changes
- **FR1.7:** System SHALL maintain independent branch state per pane

### Non-Functional Requirements

- **NFR1.1 - Performance:** Branch retrieval SHALL complete within 100ms
- **NFR1.2 - Reliability:** Git command failure SHALL be handled gracefully (no display, no error)
- **NFR1.3 - Compatibility:** SHALL work without Git installed (graceful degradation)
- **NFR1.4 - Maintainability:** SHALL follow existing code patterns in pane_render.go

## Implementation Approach

### Architecture

```
Directory Change
    ↓
Pane.SetPath() or Pane.NavigateTo()
    ↓
GetGitBranch(path) → branch name or empty string
    ↓
Store in Pane.gitBranch
    ↓
viewInternal() renders path + branch
```

**Components Involved:**
1. **Git Utilities** (`internal/fs/git.go`): New file for Git operations
2. **Pane State** (`internal/ui/pane.go`): Add `gitBranch` field
3. **Pane Rendering** (`internal/ui/pane_render.go`): Modify `viewInternal()` for branch display

### Data Flow

#### Branch Retrieval Flow
```
SetPath(newPath)
    ↓
branch := fs.GetGitBranch(newPath)
    ↓
p.gitBranch = branch
    ↓
viewInternal() called on next render
    ↓
formatHeaderLine1() includes branch if non-empty
```

### API Design

#### Git Utilities

```go
// In internal/fs/git.go

// GetGitBranch returns the current Git branch name for the given path.
// Returns empty string if not a Git repository or on error.
func GetGitBranch(path string) string {
    cmd := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
    output, err := cmd.Output()
    if err != nil {
        return ""
    }
    return strings.TrimSpace(string(output))
}
```

#### Pane State

```go
// In internal/ui/pane.go

type Pane struct {
    // ... existing fields
    gitBranch string  // Current Git branch name (empty if not in Git repo)
}

// Update SetPath to retrieve branch
func (p *Pane) SetPath(path string) error {
    // ... existing logic
    p.gitBranch = fs.GetGitBranch(path)
    return nil
}
```

#### Header Rendering

```go
// In internal/ui/pane_render.go

func (p *Pane) viewInternal(diskSpace uint64, minibuffer *Minibuffer) string {
    // ...

    // Format header line 1: path (left) + branch (right)
    displayPath := p.formatPath()
    if p.showHidden {
        displayPath = "[H] " + displayPath
    }
    if p.IsFiltered() {
        filterIndicator := p.formatFilterIndicator()
        displayPath = filterIndicator + " " + displayPath
    }

    // Branch display (right-aligned)
    branchDisplay := ""
    if p.gitBranch != "" {
        branchDisplay = "[" + p.gitBranch + "]"
    }

    // Calculate layout
    availableWidth := p.width - 2  // padding
    branchWidth := runewidth.StringWidth(branchDisplay)
    pathMaxWidth := availableWidth - branchWidth - 1  // 1 for separator space

    // Truncate path if necessary
    if runewidth.StringWidth(displayPath) > pathMaxWidth {
        displayPath = truncateString(displayPath, pathMaxWidth)
    }

    // Build header line with padding
    pathWidth := runewidth.StringWidth(displayPath)
    padding := availableWidth - pathWidth - branchWidth
    if padding < 0 {
        padding = 0
    }
    headerLine1 := displayPath + strings.Repeat(" ", padding) + branchDisplay

    // ... rest of rendering
}
```

### File Structure

Files to create:
```
internal/
└── fs/
    ├── git.go          # New: Git utilities
    └── git_test.go     # New: Tests for Git utilities
```

Files to modify:
```
internal/
├── ui/
│   ├── pane.go             # Add gitBranch field
│   ├── pane_render.go      # Modify header rendering
│   └── pane_render_test.go # Add tests for branch display
└── fs/
    └── (git.go - new)
```

## Test Scenarios

### Unit Tests

#### Test: GetGitBranch - Git Repository
- **Setup:** Directory inside a Git repository on 'main' branch
- **Action:** Call `GetGitBranch(path)`
- **Expected:** Returns "main"

#### Test: GetGitBranch - Non-Git Directory
- **Setup:** Directory not in any Git repository
- **Action:** Call `GetGitBranch(path)`
- **Expected:** Returns empty string

#### Test: GetGitBranch - Subdirectory of Git Repo
- **Setup:** Subdirectory deep inside a Git repository
- **Action:** Call `GetGitBranch(subdir)`
- **Expected:** Returns correct branch name

#### Test: Header Rendering - With Branch
- **Setup:** Pane with gitBranch = "main", path = "~/project"
- **Action:** Render header
- **Expected:** `~/project                              [main]`

#### Test: Header Rendering - Without Branch
- **Setup:** Pane with gitBranch = "", path = "~/documents"
- **Action:** Render header
- **Expected:** `~/documents` (no branch display)

#### Test: Header Rendering - Long Path Truncation
- **Setup:** Pane with gitBranch = "feature-branch", very long path
- **Action:** Render header
- **Expected:** Path truncated, branch fully visible

#### Test: Header Rendering - Special Character Branch Name
- **Setup:** Pane with gitBranch = "feature/test[bracket]", path = "~/project"
- **Action:** Render header
- **Expected:** `~/project                     [feature/test[bracket]]` (branch displayed correctly inside outer brackets)

### E2E Tests

#### Test: Branch Display in Git Directory
```bash
test_branch_display() {
    # Navigate to Git repository
    send_keys "$SESSION" "/"
    send_keys "$SESSION" "/path/to/git/repo"
    send_keys "$SESSION" "Enter"
    sleep 0.3

    # Verify branch is displayed
    capture_screen "$SESSION"
    assert_contains "$SESSION" "[main]" "Branch should be displayed"
}
```

#### Test: No Branch in Non-Git Directory
```bash
test_no_branch_in_non_git() {
    # Navigate to non-Git directory
    send_keys "$SESSION" "/"
    send_keys "$SESSION" "/tmp"
    send_keys "$SESSION" "Enter"
    sleep 0.3

    # Verify no branch bracket
    capture_screen "$SESSION"
    assert_not_contains "$SESSION" "[" "No branch bracket in non-Git dir"
}
```

### Edge Cases

- **Edge Case 1: Git not installed**
  - Expected: Empty branch, no errors
  - Validation: Application continues normally

- **Edge Case 2: Corrupted .git directory**
  - Expected: Empty branch, no errors
  - Validation: Error is swallowed

- **Edge Case 3: Very long branch name**
  - Branch name like "feature/JIRA-12345-very-long-description"
  - Expected: Branch displayed, path truncated if necessary

- **Edge Case 4: Branch name with special characters**
  - Branch like "feature/test[bracket]"
  - Expected: Displayed correctly inside outer brackets `[[bracket]]`

- **Edge Case 5: Detached HEAD state**
  - When HEAD is not pointing to a branch (e.g., after `git checkout <commit-hash>`)
  - Git returns "HEAD" from `git rev-parse --abbrev-ref HEAD`
  - Expected: Display `[HEAD]` to indicate detached state

## Security Considerations

- **Command Injection:** Use `exec.Command` with separate arguments (no shell interpretation)
- **Path Traversal:** Git command operates within specified directory only
- **Resource Usage:** Single git process per directory change, self-limiting

## Error Handling

### Error Flow
```
Git Command Failure → Return empty string → No branch displayed → Continue
```

**No user-visible errors:** All Git-related errors result in silent fallback to no branch display.

## Performance Optimization

### Performance Goals
- Branch retrieval: < 100ms for 99% of operations
- Memory overhead: O(branch name length) per pane

### Optimization Strategies
- **Lazy Evaluation:** Only call git when directory changes
- **No Caching:** Branch can change externally, always fetch fresh
- **Fast Git Command:** `git rev-parse --abbrev-ref HEAD` is lightweight

## Success Criteria

- [ ] Branch displayed in `[branch]` format at header right
- [ ] Branch hidden for non-Git directories
- [ ] Each pane shows independent branch
- [ ] Performance under 100ms
- [ ] All unit tests pass
- [ ] E2E tests pass
- [ ] No regressions in existing functionality

## Open Questions

None - all requirements have been clarified.

## Implementation Phases

### Phase 1: Git Utilities
**Goals:** Implement branch retrieval function
**Deliverables:**
- Create `internal/fs/git.go` with `GetGitBranch()`
- Write unit tests for Git utilities

### Phase 2: Pane Integration
**Goals:** Store and display branch in pane
**Deliverables:**
- Add `gitBranch` field to Pane
- Update `SetPath()` and related methods to retrieve branch
- Modify `viewInternal()` for header layout

### Phase 3: Testing & Polish
**Goals:** Comprehensive testing and edge case handling
**Deliverables:**
- Unit tests for rendering
- E2E tests
- Test edge cases (no git, long branch, etc.)

## References

- **Existing Implementation:**
  - `internal/ui/pane_render.go` - Header rendering patterns
  - `internal/ui/pane.go` - Pane state management
- **Git Documentation:** `git rev-parse --abbrev-ref HEAD`
