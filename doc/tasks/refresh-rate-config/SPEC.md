# Feature: Refresh Rate Configuration

## Overview

Add configurable periodic auto-refresh for directory listings and disk space. The refresh interval (in seconds) is set via `refresh_rate` in `config.toml`. Default is 3 seconds. Setting to 0 disables auto-refresh. This replaces the existing fixed 5-second disk space ticker (`diskSpaceTickCmd`) with a unified timer.

## Domain Rules

- Auto-refresh reloads directory listings and disk space for both panes at the configured interval.
- Auto-refresh is suppressed while any dialog is open.
- Cursor position is preserved across auto-refreshes by matching the currently selected file name.
- Setting `refresh_rate` to 0 disables auto-refresh entirely; manual refresh (F5/Ctrl+R) remains available.
- The existing fixed 5-second disk space ticker (`diskSpaceTickCmd` / `diskSpaceUpdateMsg`) is removed and replaced by this unified auto-refresh mechanism.

## Objectives

- Automatically refresh directory listings and disk space at a configurable interval.
- Allow users to configure the refresh interval via `config.toml`.
- Support disabling auto-refresh by setting the interval to 0.
- Consolidate the existing disk space ticker into the unified auto-refresh timer.

## User Stories

### US1: Automatic Refresh
As a user, I want the directory listing and disk space to automatically update at a regular interval, so that I can see file changes made by external processes without manually pressing F5.

**Acceptance Criteria:**
- [ ] Directory listings for both panes refresh at the configured interval
- [ ] Disk space for both panes refreshes at the same interval
- [ ] Default interval is 3 seconds
- [ ] Cursor position is maintained after auto-refresh
- [ ] Auto-refresh does not occur while a dialog is open

### US2: Configure Refresh Rate
As a user, I want to configure the refresh interval in `config.toml`, so that I can balance between responsiveness and system load.

**Acceptance Criteria:**
- [ ] `refresh_rate` setting is available in `config.toml`
- [ ] Valid range is 0-60 seconds
- [ ] Out-of-range values fall back to default (3 seconds) with a warning
- [ ] Changes are applied immediately via config hot-reload

### US3: Disable Auto-Refresh
As a user, I want to disable auto-refresh by setting `refresh_rate = 0`, so that I can reduce system load on slow or remote filesystems.

**Acceptance Criteria:**
- [ ] Setting `refresh_rate = 0` disables auto-refresh
- [ ] Manual refresh (F5/Ctrl+R) still works when auto-refresh is disabled
- [ ] Changing from 0 to a positive value re-enables auto-refresh via hot-reload

## Functional Requirements

- **FR1:** Add `refresh_rate` integer parameter to `Config` struct and `rawConfig` struct.
- **FR2:** Default value for `refresh_rate` is 3 (seconds).
- **FR3:** Valid range is 0-60 (inclusive). Values outside this range use the default with a warning.
- **FR4:** When `refresh_rate` > 0, a periodic ticker sends `autoRefreshMsg` at the configured interval.
- **FR5:** On receiving `autoRefreshMsg`, both panes reload their directory listings and disk space is updated.
- **FR6:** Cursor position is preserved by remembering the selected file name and restoring it after reload.
- **FR7:** Auto-refresh is suppressed when `m.dialog != nil` (any dialog is open).
- **FR8:** When `refresh_rate` is 0, no ticker is started and no auto-refresh occurs.
- **FR9:** Config hot-reload applies the new `refresh_rate` immediately, restarting or stopping the ticker as needed.
- **FR10:** Config auto-merge adds `refresh_rate` to existing config files that lack it.
- **FR11:** Remove `diskSpaceTickCmd`, `diskSpaceUpdateMsg`, and the separate disk space ticker. Disk space updates are handled within `autoRefreshMsg`.

## Non-Functional Requirements

- **NFR1 - Performance:** Auto-refresh uses the same directory loading mechanism as manual refresh. Disk space query uses `syscall.Statfs` (kernel metadata read, no disk I/O). No significant overhead at any interval within the valid range.
- **NFR2 - Compatibility:** Existing config files without `refresh_rate` work with the default value (3 seconds).

## Configuration

### config.toml

```toml
refresh_rate = 3
```

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `refresh_rate` | integer | 3 | 0-60 | Directory auto-refresh interval in seconds. 0 disables auto-refresh. |

### Validation

| Input | Behavior |
|-------|----------|
| 0 | Disable auto-refresh |
| 1-60 | Auto-refresh at specified interval |
| < 0 | Use default (3), emit warning |
| > 60 | Use default (3), emit warning |

## Implementation Approach

### New Message Type

```go
// autoRefreshMsg triggers periodic auto-refresh (directory listings + disk space).
type autoRefreshMsg struct{}
```

### Ticker Command

```go
func autoRefreshTickCmd(interval time.Duration) tea.Cmd {
    return tea.Tick(interval, func(t time.Time) tea.Msg {
        return autoRefreshMsg{}
    })
}
```

### Removed Message Types

- `diskSpaceUpdateMsg` - replaced by `autoRefreshMsg`
- `diskSpaceTickCmd()` - replaced by `autoRefreshTickCmd()`

### Config Changes

Add to `Config` struct:
```go
RefreshRate int `toml:"refresh_rate"` // 0 = disabled, 1-60 = seconds
```

Add to `rawConfig` struct:
```go
RefreshRate *int `toml:"refresh_rate"`
```

Default constant:
```go
const DefaultRefreshRate = 3
```

### Update Handler

In `Model.Update()`, handle `autoRefreshMsg`:
1. If `m.activeDialog != nil`, skip refresh and re-schedule ticker.
2. Reload both panes' directory listings.
3. Update disk space via `DiskSpaceMonitor.Update()`.
4. Preserve cursor position by file name.
5. Re-schedule next tick.

Remove the existing `diskSpaceUpdateMsg` handler.

### Hot-Reload

In `applyConfig()`, update the refresh rate and restart the ticker if the value changed.

## Affected Files

| File | Change |
|------|--------|
| `internal/config/config.go` | Add `RefreshRate` to `Config` and `rawConfig`; load and validate in `LoadConfig`; add `DefaultRefreshRate` constant |
| `internal/config/reload.go` | Add `refresh_rate` validation in `buildConfigFromRaw` |
| `internal/config/merger.go` | Add `refresh_rate` to auto-merge list |
| `internal/ui/messages.go` | Add `autoRefreshMsg` and `autoRefreshTickCmd`; remove `diskSpaceUpdateMsg` and `diskSpaceTickCmd` |
| `internal/ui/model.go` | Store refresh rate; initialize ticker in `Init()`; update `applyConfig` for hot-reload |
| `internal/ui/model_update.go` | Handle `autoRefreshMsg`; remove `diskSpaceUpdateMsg` handler |
| `cmd/duofm/main.go` | Pass `RefreshRate` to `NewModelWithConfig` |

## Test Scenarios

### Unit Tests
- [ ] Default config has `refresh_rate` = 3
- [ ] `refresh_rate = 0` disables auto-refresh (no ticker started)
- [ ] `refresh_rate = 5` produces 5-second interval
- [ ] `refresh_rate = -1` falls back to default with warning
- [ ] `refresh_rate = 61` falls back to default with warning
- [ ] `refresh_rate = 60` is accepted (boundary)
- [ ] `refresh_rate = 1` is accepted (boundary)

### Integration Tests
- [ ] Auto-refresh reloads directory contents when files change externally
- [ ] Cursor position is preserved after auto-refresh
- [ ] Auto-refresh is suppressed during dialog display
- [ ] Hot-reload applies new refresh rate immediately
- [ ] Changing from 0 to positive value enables auto-refresh
- [ ] Changing from positive value to 0 disables auto-refresh

## Success Criteria

- [ ] Directory listings and disk space auto-refresh at the configured interval
- [ ] `refresh_rate` is configurable via `config.toml`
- [ ] Default value is 3 seconds
- [ ] Setting to 0 disables auto-refresh
- [ ] Out-of-range values fall back to default with warning
- [ ] Cursor position is preserved
- [ ] Dialogs suppress auto-refresh
- [ ] Config hot-reload applies changes immediately
- [ ] Existing `diskSpaceTickCmd` / `diskSpaceUpdateMsg` are removed
- [ ] All unit tests pass
- [ ] No regression in existing functionality

## Dependencies

**Internal:**
- `internal/config/` - Configuration loading and hot-reload
- `internal/ui/` - Bubble Tea model and update loop

**External:**
- `github.com/charmbracelet/bubbletea` - `tea.Tick` for periodic timer

## Constraints

- Minimum resolution is 1 second (integer seconds only).
- Auto-refresh reads the filesystem for directory listings; on slow or network filesystems, users should increase the interval or disable it. Disk space queries (`syscall.Statfs`) are negligible in cost.

## Open Questions

None.
