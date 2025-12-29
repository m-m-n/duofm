# Verification: Color Theme Configuration

## Build Status

- [ ] `go build ./...` completes without errors
- [ ] `make build` succeeds

## Test Execution

- [ ] `go test ./...` passes all tests
- [ ] Test coverage >= 80% for new code

## Code Quality

- [ ] `gofmt -l .` returns no files
- [ ] `go vet ./...` reports no issues

## Created Files

- [ ] `internal/config/colors.go` - ColorConfig struct and functions
- [ ] `internal/config/colors_test.go` - Unit tests for colors
- [ ] `internal/ui/theme.go` - Theme struct for UI
- [ ] `internal/ui/theme_test.go` - Unit tests for theme

## Modified Files

- [ ] `internal/config/config.go` - Added Colors field to Config
- [ ] `internal/config/generator.go` - Added [colors] section template
- [ ] `internal/config/config_test.go` - Updated line limit test
- [ ] `internal/ui/model.go` - Added theme field, updated NewModelWithConfig
- [ ] `internal/ui/pane.go` - Added theme field, updated color usage
- [ ] `internal/ui/help_dialog.go` - Added scrolling and color palette
- [ ] `internal/ui/help_dialog_test.go` - Added scroll and palette tests
- [ ] `internal/ui/displaymode_test.go` - Updated NewPane calls
- [ ] `internal/ui/pane_test.go` - Updated NewPane calls
- [ ] `internal/ui/pane_mark_test.go` - Updated NewPane calls
- [ ] `cmd/duofm/main.go` - Added theme initialization

## Unit Test Checklist (from SPEC.md)

### config package
- [ ] ParseColor correctly parses ANSI 256-color codes (0-255)
- [ ] ParseColor returns error for out-of-range values (< 0, > 255)
- [ ] ParseColor returns error for non-integer values
- [ ] LoadColors returns defaults when section missing
- [ ] LoadColors merges partial settings with defaults
- [ ] LoadColors handles missing color values
- [ ] DefaultColors returns all expected keys with correct values
- [ ] GenerateDefaultConfig includes [colors] section with examples

### ui package
- [ ] Theme applies cursor colors correctly
- [ ] Theme applies mark colors correctly
- [ ] Theme applies file type colors correctly
- [ ] Theme falls back to defaults for missing colors
- [ ] HelpDialog scroll position updates correctly
- [ ] HelpDialog generates correct hex values for colors 16-231
- [ ] HelpDialog generates correct hex values for colors 232-255

## Success Criteria (from SPEC.md)

- [ ] `[colors]` section in config.toml customizes all UI colors
- [ ] ANSI 256-color codes (0-255) supported
- [ ] Missing settings use default values
- [ ] Invalid colors trigger warning and use defaults
- [ ] Backward compatible with existing config files
- [ ] All pane elements customizable
- [ ] Generated default config includes [colors] examples
- [ ] Help dialog supports scrolling
- [ ] Help dialog includes color palette reference
- [ ] All unit tests pass

## Manual Testing Checklist

### Basic Configuration
1. [ ] Start duofm without config.toml - uses default colors
2. [ ] Start duofm with empty [colors] section - uses default colors
3. [ ] Start duofm with custom cursor_bg - cursor color changes
4. [ ] Start duofm with invalid color (300) - warning displayed, default used

### Pane Colors
5. [ ] Cursor row shows configured cursor_bg color (active pane)
6. [ ] Cursor row shows configured cursor_bg_inactive color (inactive pane)
7. [ ] Marked files show configured mark_bg color
8. [ ] Cursor + marked file shows cursor_mark_bg color
9. [ ] Directory names show configured directory_fg color
10. [ ] Symlinks show configured symlink_fg color
11. [ ] Executables show configured executable_fg color
12. [ ] Path bar shows configured path_fg color
13. [ ] Header shows configured header_fg color

### Help Dialog
14. [ ] Press ? to open help dialog
15. [ ] Press j/k to scroll line by line
16. [ ] Press Space/Shift+Space to scroll page by page
17. [ ] Press g to go to top, G to go to bottom
18. [ ] Scroll down to see "Color Palette Reference" section
19. [ ] Standard colors (0-15) display with "Terminal-dependent" label
20. [ ] 6x6x6 color cube (16-231) shows color samples with hex values
21. [ ] Grayscale (232-255) shows color samples with hex values
22. [ ] Page indicator [N/M] updates while scrolling

### Config Generation
23. [ ] Delete config.toml and start duofm
24. [ ] Verify generated config.toml contains [colors] section
25. [ ] Verify all 35 color options are listed with comments
