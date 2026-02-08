# Verification Plan: Refresh Rate Configuration

## 1. Build Verification

- [ ] `make build` succeeds without errors
- [ ] `go vet ./...` reports no issues
- [ ] No new compiler warnings

## 2. Unit Tests

### 2.1 Config Loading (`internal/config/config_test.go`)

- [ ] `TestLoadConfig_RefreshRateDefault` - Default config has `RefreshRate = 3`
- [ ] `TestLoadConfig_RefreshRateExplicit` - `refresh_rate = 5` loads correctly
- [ ] `TestLoadConfig_RefreshRateZero` - `refresh_rate = 0` loads as 0 (disabled)
- [ ] `TestLoadConfig_RefreshRateNegative` - `refresh_rate = -1` falls back to default, emits warning
- [ ] `TestLoadConfig_RefreshRateOverMax` - `refresh_rate = 61` falls back to default, emits warning
- [ ] `TestLoadConfig_RefreshRateBoundaryMin` - `refresh_rate = 1` accepted
- [ ] `TestLoadConfig_RefreshRateBoundaryMax` - `refresh_rate = 60` accepted
- [ ] `TestLoadConfig_RefreshRateFileNotExists` - Missing file returns default `RefreshRate = 3`

### 2.2 Config Detailed Loading (`internal/config/reload_test.go`)

- [ ] `TestLoadConfigDetailed_RefreshRateInvalid` - Invalid values produce `ConfigError`

### 2.3 Config Auto-Merge (`internal/config/merger_test.go`)

- [ ] `TestIsMissingRefreshRate` - Returns true when nil, false when set
- [ ] `TestMergeConfig_RefreshRate` - Missing `refresh_rate` is auto-merged into config file

### 2.4 UI Messages (`internal/ui/messages_test.go`)

- [ ] `TestAutoRefreshTickCmd` - Returns non-nil command
- [ ] No references to `diskSpaceTickCmd` or `diskSpaceUpdateMsg` remain

### 2.5 UI Update Handler (`internal/ui/model_dialog_msg_test.go`)

- [ ] `TestHandleAutoRefreshMsg` - Handler refreshes panes and reschedules ticker
- [ ] `TestHandleAutoRefreshMsg_DialogOpen` - Handler skips refresh when dialog is open, still reschedules
- [ ] `TestHandleAutoRefreshMsg_Disabled` - Handler does nothing when `refreshRate = 0`

### 2.6 Hot-Reload Edge Cases

- [ ] Hot-reload from `refreshRate > 0` to different `refreshRate > 0` - next tick uses new interval
- [ ] Hot-reload from `refreshRate > 0` to `refreshRate = 0` - ticker stops on next `autoRefreshMsg`
- [ ] Hot-reload from `refreshRate = 0` to `refreshRate > 0` - new ticker is started

## 3. Existing Test Regression

- [ ] `make test` - All existing tests pass
- [ ] No test references to removed `diskSpaceUpdateMsg` / `diskSpaceTickCmd`

## 4. Manual Verification

### 4.1 Basic Auto-Refresh

1. Start duofm with default config (`refresh_rate = 3`)
2. In another terminal, create a file in the current directory: `touch /tmp/testfile_$$`
3. Within 3 seconds, verify the new file appears in the pane
4. Delete the file: `rm /tmp/testfile_$$`
5. Within 3 seconds, verify the file disappears from the pane

### 4.2 Disk Space Update

1. Start duofm with `refresh_rate = 3`
2. Verify disk space is displayed in the status bar
3. Create a large file in another terminal
4. Within 3 seconds, verify disk space value updates

### 4.3 Disable Auto-Refresh

1. Set `refresh_rate = 0` in `config.toml`
2. Start duofm
3. Create a file in another terminal
4. Wait 10 seconds - file should NOT appear automatically
5. Press F5 (manual refresh) - file should appear

### 4.4 Hot-Reload

1. Start duofm with `refresh_rate = 10`
2. Edit `config.toml` and change to `refresh_rate = 1`
3. Verify status bar shows "Config reloaded"
4. Create files rapidly - they should appear within ~1 second

### 4.5 Dialog Suppression

1. Start duofm with `refresh_rate = 1`
2. Open any dialog (e.g., help dialog with `?`)
3. Create a file in another terminal
4. File should NOT cause visual glitches or close the dialog
5. Close dialog - file should appear on next auto-refresh

### 4.6 Cursor Preservation

1. Start duofm with `refresh_rate = 1`
2. Navigate cursor to a specific file
3. Wait for auto-refresh
4. Cursor should remain on the same file

### 4.7 Validation

1. Set `refresh_rate = -5` in `config.toml`
2. Start duofm
3. Verify warning appears and default (3 seconds) is used
4. Set `refresh_rate = 100` in `config.toml`
5. Hot-reload should show error dialog

## 5. Code Quality

- [ ] No leftover references to `diskSpaceUpdateMsg` or `diskSpaceTickCmd` in source code
- [ ] No leftover references in test code
- [ ] `gofmt` formatting is correct
- [ ] New code follows existing patterns (matches `history_limit` pattern for config)

## 6. Success Criteria

All items above must pass. Specifically:
- [ ] All unit tests pass (`make test`)
- [ ] Build succeeds (`make build`)
- [ ] Auto-refresh works at configured interval
- [ ] Disk space updates are included in auto-refresh
- [ ] `refresh_rate = 0` disables auto-refresh
- [ ] Hot-reload applies new refresh rate
- [ ] Dialog suppression works correctly
- [ ] Cursor position is preserved
- [ ] Out-of-range values produce appropriate warnings/errors
- [ ] Existing `diskSpaceTickCmd` is completely removed
