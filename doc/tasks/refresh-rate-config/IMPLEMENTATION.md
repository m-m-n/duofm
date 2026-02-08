# Implementation Plan: Refresh Rate Configuration

## Overview

Add configurable periodic auto-refresh for directory listings and disk space. Replaces the existing fixed 5-second disk space ticker with a unified timer controlled by `refresh_rate` in `config.toml`.

## Implementation Order

Changes are ordered by dependency: config layer first, then UI layer, then entry point.

### Step 1: Add `RefreshRate` to Config struct and rawConfig

**File:** `internal/config/config.go`

**Changes:**
1. Add `RefreshRate int` field to `Config` struct with TOML tag `refresh_rate`.
2. Add `RefreshRate *int` field to `rawConfig` struct with TOML tag `refresh_rate`.
3. Add `DefaultRefreshRate = 3` constant.
4. In `LoadConfig()`, add validation and loading logic for `refresh_rate` after `history_limit` loading:
   - If `raw.RefreshRate != nil`, validate range 0-60.
   - If out of range (< 0 or > 60), use default and append warning.
   - Otherwise, use the provided value.
5. In `defaultConfig()`, set `RefreshRate: DefaultRefreshRate`.

**Validation logic:**
```go
if raw.RefreshRate != nil {
    rate := *raw.RefreshRate
    if rate < 0 || rate > 60 {
        warnings = append(warnings, fmt.Sprintf("Warning: refresh_rate %d out of range (0-60), using default %d", rate, DefaultRefreshRate))
    } else {
        cfg.RefreshRate = rate
    }
}
```

### Step 2: Add `RefreshRate` validation to `reload.go`

**File:** `internal/config/reload.go`

**Changes:**
1. In `buildConfigFromRaw()`, add `refresh_rate` loading with validation:
   - If `raw.RefreshRate != nil`, validate range 0-60.
   - If out of range, record a `ConfigError` and keep default.
   - Otherwise, set `cfg.RefreshRate`.

### Step 3: Add `refresh_rate` to config auto-merge

**File:** `internal/config/merger.go`

**Changes:**
1. Add `RefreshRate *int` field to `mergeResult` struct.
2. Add `IsMissingRefreshRate()` function.
3. In `MergeConfig()`, check if `refresh_rate` is missing and add to `mergeResult`.
4. In `mergeResult.hasContent()`, add `m.RefreshRate != nil` check.
5. In `generateMergedFile()`, insert `refresh_rate` alongside `history_limit` (root-level key before first section).

### Step 4: Replace disk space ticker with auto-refresh ticker in messages

**File:** `internal/ui/messages.go`

**Changes:**
1. Remove `diskSpaceUpdateMsg` struct.
2. Remove `diskSpaceTickCmd()` function.
3. Add `autoRefreshMsg` struct.
4. Add `autoRefreshTickCmd(interval time.Duration) tea.Cmd` function.

```go
// autoRefreshMsg triggers periodic auto-refresh (directory listings + disk space).
type autoRefreshMsg struct{}

// autoRefreshTickCmd returns a command that sends autoRefreshMsg after the specified interval.
func autoRefreshTickCmd(interval time.Duration) tea.Cmd {
    return tea.Tick(interval, func(t time.Time) tea.Msg {
        return autoRefreshMsg{}
    })
}
```

### Step 5: Add refresh rate to Model and update Init/handlers

**File:** `internal/ui/model.go`

**Changes:**
1. Add `refreshRate int` field to `Model` struct (in Configuration section).
2. Update `NewModelWithConfig()` signature: add `refreshRate int` parameter.
3. Update `NewModel()` to pass `config.DefaultRefreshRate` to `NewModelWithConfig()`.
4. Set `refreshRate` in model initialization.
5. In `Init()`, start auto-refresh ticker if `refreshRate > 0`:
   ```go
   if m.refreshRate > 0 {
       cmds = append(cmds, autoRefreshTickCmd(time.Duration(m.refreshRate) * time.Second))
   }
   ```
6. Ensure `Init()` always returns `tea.Batch(cmds...)` (simplify existing conditional).

**Note:** `NewModel()` is used extensively in tests (~100+ calls). Since it wraps `NewModelWithConfig`, updating `NewModel()` to pass `DefaultRefreshRate` ensures all existing tests continue to work without modification.

### Step 6: Handle `autoRefreshMsg` in Update, remove `diskSpaceUpdateMsg` handler

**File:** `internal/ui/model_update.go`

**Changes:**
1. In `handleSystemMessages()`, replace `diskSpaceUpdateMsg` case with `autoRefreshMsg`:
   ```go
   case autoRefreshMsg:
       if m.refreshRate <= 0 {
           return m, nil
       }
       // Suppress during dialog
       if m.dialog != nil {
           return m, autoRefreshTickCmd(time.Duration(m.refreshRate) * time.Second)
       }
       // Refresh both panes (preserves cursor)
       if m.leftPane != nil {
           m.leftPane.RefreshDirectoryPreserveCursor()
       }
       if m.rightPane != nil {
           m.rightPane.RefreshDirectoryPreserveCursor()
       }
       // Update disk space
       m.updateDiskSpace()
       return m, autoRefreshTickCmd(time.Duration(m.refreshRate) * time.Second)
   ```
2. In `handleWindowSize()`, replace `diskSpaceTickCmd()` with `autoRefreshTickCmd()`:
   - On initial ready: start auto-refresh ticker if `m.refreshRate > 0`.

### Step 7: Apply refresh rate on hot-reload

**File:** `internal/ui/model.go` (where `applyConfig` is defined)

**Changes:**
1. In `applyConfig()`, save old rate, update refresh rate, and return a `tea.Cmd` if a new ticker needs to be started:
   ```go
   m.refreshRate = cfg.RefreshRate
   ```
2. In `handleConfigFileChanged()` (`model_update_config.go`), after `applyConfig()`, if `refreshRate` changed from 0 to a positive value, return `autoRefreshTickCmd` to restart the ticker. When an existing ticker is running and interval changes, it will naturally pick up the new value on the next `autoRefreshMsg`. When changing to 0, the next `autoRefreshMsg` handler will detect `refreshRate <= 0` and stop.

**Critical edge case:** When `refreshRate` was 0 (ticker stopped), changing to a positive value via hot-reload requires explicitly starting a new ticker, since no `autoRefreshMsg` will arrive to read the updated value.

### Step 8: Update main.go to pass refresh rate

**File:** `cmd/duofm/main.go`

**Changes:**
1. Update `NewModelWithConfig()` call to pass `cfg.RefreshRate`.
2. Update the fallback `ConfigLoadResult` to include `RefreshRate: config.DefaultRefreshRate`.

### Step 9: Update tests

**Files:**
- `internal/config/config_test.go` - Add tests for `refresh_rate` loading and validation.
- `internal/config/merger_test.go` - Add tests for `refresh_rate` auto-merge.
- `internal/ui/messages_test.go` - Replace `TestDiskSpaceTickCmd` with `TestAutoRefreshTickCmd`.
- `internal/ui/model_dialog_msg_test.go` - Update `diskSpaceUpdateMsg` test to `autoRefreshMsg`.

## File Change Summary

| File | Type | Description |
|------|------|-------------|
| `internal/config/config.go` | Modify | Add `RefreshRate` field, validation, default constant |
| `internal/config/reload.go` | Modify | Add `refresh_rate` validation in `buildConfigFromRaw` |
| `internal/config/merger.go` | Modify | Add `refresh_rate` to auto-merge |
| `internal/ui/messages.go` | Modify | Replace `diskSpaceUpdateMsg`/`diskSpaceTickCmd` with `autoRefreshMsg`/`autoRefreshTickCmd` |
| `internal/ui/model.go` | Modify | Add `refreshRate` field, update `NewModelWithConfig`, update `Init`, update `applyConfig` |
| `internal/ui/model_update.go` | Modify | Replace `diskSpaceUpdateMsg` handler with `autoRefreshMsg` handler |
| `cmd/duofm/main.go` | Modify | Pass `RefreshRate` to `NewModelWithConfig` |
| `internal/config/config_test.go` | Modify | Add `refresh_rate` tests |
| `internal/config/merger_test.go` | Modify | Add `refresh_rate` merge test |
| `internal/ui/messages_test.go` | Modify | Update ticker test |
| `internal/ui/model_dialog_msg_test.go` | Modify | Update message handler test |

## Test Strategy

### Unit Tests (config)
- Default config has `RefreshRate = 3`
- Explicit `refresh_rate = 5` is loaded correctly
- `refresh_rate = 0` disables auto-refresh
- `refresh_rate = -1` falls back to default with warning
- `refresh_rate = 61` falls back to default with warning
- `refresh_rate = 1` boundary accepted
- `refresh_rate = 60` boundary accepted
- Missing `refresh_rate` in config file uses default

### Unit Tests (UI)
- `autoRefreshTickCmd` returns a non-nil command
- `autoRefreshMsg` handler refreshes both panes and disk space
- `autoRefreshMsg` handler skips refresh when dialog is open
- `autoRefreshMsg` handler does nothing when `refreshRate = 0`

### Integration Tests
- Hot-reload applies new refresh rate (positive → positive)
- Hot-reload from 0 to positive starts new ticker
- Hot-reload from positive to 0 stops ticker
- Config auto-merge adds missing `refresh_rate`

## Risk Assessment

- **Low risk:** Config changes follow the established `history_limit` pattern exactly.
- **Low risk:** Replacing `diskSpaceTickCmd` with `autoRefreshTickCmd` is a straightforward message type swap.
- **Medium risk:** `RefreshDirectoryPreserveCursor` is already used in production for manual refresh; using it for auto-refresh should be safe but needs verification that it handles concurrent directory changes gracefully.
