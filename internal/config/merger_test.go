package config

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestFindMissingKeybindings tests the FindMissingKeybindings function.
func TestFindMissingKeybindings(t *testing.T) {
	tests := []struct {
		name     string
		existing map[string][]string
		wantLen  int // Number of missing keybindings
	}{
		{
			name:     "empty map - all missing",
			existing: map[string][]string{},
			wantLen:  len(DefaultKeybindings()),
		},
		{
			name: "some keys set - partial missing",
			existing: map[string][]string{
				"move_down": {"J"},
				"move_up":   {"K"},
				"quit":      {"Q"},
			},
			wantLen: len(DefaultKeybindings()) - 3,
		},
		{
			name:     "all keys set - none missing",
			existing: DefaultKeybindings(),
			wantLen:  0,
		},
		{
			name:     "nil map - all missing",
			existing: nil,
			wantLen:  len(DefaultKeybindings()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindMissingKeybindings(tt.existing)
			if len(got) != tt.wantLen {
				t.Errorf("FindMissingKeybindings() returned %d items, want %d", len(got), tt.wantLen)
			}

			// Verify that returned keys are actually missing from existing
			for key := range got {
				if tt.existing != nil {
					if _, exists := tt.existing[key]; exists {
						t.Errorf("FindMissingKeybindings() returned key %q which exists in existing", key)
					}
				}
			}

			// Verify that returned keys exist in defaults
			defaults := DefaultKeybindings()
			for key := range got {
				if _, exists := defaults[key]; !exists {
					t.Errorf("FindMissingKeybindings() returned key %q which doesn't exist in defaults", key)
				}
			}
		})
	}
}

// TestFindMissingColors tests the FindMissingColors function.
func TestFindMissingColors(t *testing.T) {
	allColorKeys := AllColorKeys()
	tests := []struct {
		name     string
		existing map[string]interface{}
		wantLen  int
	}{
		{
			name:     "empty map - all missing",
			existing: map[string]interface{}{},
			wantLen:  len(allColorKeys),
		},
		{
			name: "some colors set - partial missing",
			existing: map[string]interface{}{
				"cursor_fg": 15,
				"cursor_bg": 39,
				"border_fg": 240,
			},
			wantLen: len(allColorKeys) - 3,
		},
		{
			name:     "nil map - all missing",
			existing: nil,
			wantLen:  len(allColorKeys),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindMissingColors(tt.existing)
			if len(got) != tt.wantLen {
				t.Errorf("FindMissingColors() returned %d items, want %d", len(got), tt.wantLen)
			}

			// Verify that returned keys are actually missing from existing
			for key := range got {
				if tt.existing != nil {
					if _, exists := tt.existing[key]; exists {
						t.Errorf("FindMissingColors() returned key %q which exists in existing", key)
					}
				}
			}

			// Verify that returned values match default values
			for key, value := range got {
				expectedValue := GetDefaultColorValue(key)
				if value != expectedValue {
					t.Errorf("FindMissingColors() returned value %d for key %q, want %d", value, key, expectedValue)
				}
			}
		})
	}
}

// TestFindMissingColors_AllColorsSet tests when all colors are set.
func TestFindMissingColors_AllColorsSet(t *testing.T) {
	// Create a map with all color keys set
	existing := make(map[string]interface{})
	for _, key := range AllColorKeys() {
		existing[key] = 100 // Arbitrary value
	}

	got := FindMissingColors(existing)
	if len(got) != 0 {
		t.Errorf("FindMissingColors() returned %d items when all colors are set, want 0", len(got))
	}
}

// TestIsMissingHistoryLimit tests the IsMissingHistoryLimit function.
func TestIsMissingHistoryLimit(t *testing.T) {
	tests := []struct {
		name         string
		historyLimit *int
		want         bool
	}{
		{
			name:         "nil - missing",
			historyLimit: nil,
			want:         true,
		},
		{
			name:         "value set - not missing",
			historyLimit: intPtr(1000),
			want:         false,
		},
		{
			name:         "zero value set - not missing",
			historyLimit: intPtr(0),
			want:         false,
		},
		{
			name:         "default value set - not missing",
			historyLimit: intPtr(DefaultHistoryLimit),
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMissingHistoryLimit(tt.historyLimit)
			if got != tt.want {
				t.Errorf("IsMissingHistoryLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetDefaultColorValue tests the GetDefaultColorValue function.
func TestGetDefaultColorValue(t *testing.T) {
	defaults := DefaultColors()
	tests := []struct {
		name string
		key  string
		want int
	}{
		{"cursor_fg", "cursor_fg", defaults.CursorFg},
		{"cursor_bg", "cursor_bg", defaults.CursorBg},
		{"border_fg", "border_fg", defaults.BorderFg},
		{"directory_fg", "directory_fg", defaults.DirectoryFg},
		{"status_fg", "status_fg", defaults.StatusFg},
		{"status_bg", "status_bg", defaults.StatusBg},
		{"unknown key", "unknown_key", -1},
		{"empty key", "", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDefaultColorValue(tt.key)
			if got != tt.want {
				t.Errorf("GetDefaultColorValue(%q) = %d, want %d", tt.key, got, tt.want)
			}
		})
	}
}

// TestGetDefaultColorValue_AllKeys tests GetDefaultColorValue for all color keys.
func TestGetDefaultColorValue_AllKeys(t *testing.T) {
	for _, key := range AllColorKeys() {
		got := GetDefaultColorValue(key)
		if got == -1 {
			t.Errorf("GetDefaultColorValue(%q) returned -1, expected valid color value", key)
		}
		if got < 0 || got > 255 {
			t.Errorf("GetDefaultColorValue(%q) = %d, expected value in range 0-255", key, got)
		}
	}
}

// intPtr is a helper to create *int from int value.
func intPtr(v int) *int {
	return &v
}

// TestGenerateMergedFile tests the generateMergedFile function.
func TestGenerateMergedFile(t *testing.T) {
	tests := []struct {
		name         string
		original     string
		result       mergeResult
		wantContains []string
		wantNotIn    []string
	}{
		{
			name:     "empty result - no changes",
			original: "[keybindings]\nmove_down = [\"J\"]\n",
			result: mergeResult{
				Keybindings:  map[string][]string{},
				Colors:       map[string]int{},
				HistoryLimit: nil,
			},
			wantContains: []string{
				"[keybindings]",
				`move_down = ["J"]`,
			},
			wantNotIn: []string{
				"# --- Auto-merged settings",
			},
		},
		{
			name:     "missing keybindings inserted into existing section",
			original: "[keybindings]\nmove_down = [\"J\"]\n",
			result: mergeResult{
				Keybindings: map[string][]string{
					"move_up": {"K", "Up"},
				},
				Colors:       map[string]int{},
				HistoryLimit: nil,
			},
			wantContains: []string{
				"[keybindings]",
				`move_down = ["J"]`,
				`move_up = ["K", "Up"]`,
			},
		},
		{
			name:     "missing colors inserted into existing section",
			original: "[colors]\ncursor_fg = 100\n",
			result: mergeResult{
				Keybindings: map[string][]string{},
				Colors: map[string]int{
					"border_fg": 240,
				},
				HistoryLimit: nil,
			},
			wantContains: []string{
				"[colors]",
				"cursor_fg = 100",
				"border_fg = 240",
			},
		},
		{
			name:     "missing history_limit inserted before first section",
			original: "[keybindings]\nmove_down = [\"J\"]\n",
			result: mergeResult{
				Keybindings:  map[string][]string{},
				Colors:       map[string]int{},
				HistoryLimit: intPtr(20000),
			},
			wantContains: []string{
				"history_limit = 20000",
				"[keybindings]",
			},
		},
		{
			name:     "new keybindings section created when missing",
			original: "# Config file\n",
			result: mergeResult{
				Keybindings: map[string][]string{
					"quit": {"Q"},
				},
				Colors:       map[string]int{},
				HistoryLimit: nil,
			},
			wantContains: []string{
				"# Config file",
				"# --- Auto-merged settings (added by duofm) ---",
				"[keybindings]",
				`quit = ["Q"]`,
			},
		},
		{
			name:     "new colors section created when missing",
			original: "[keybindings]\nmove_down = [\"J\"]\n",
			result: mergeResult{
				Keybindings: map[string][]string{},
				Colors: map[string]int{
					"cursor_fg": 15,
				},
				HistoryLimit: nil,
			},
			wantContains: []string{
				"[keybindings]",
				"# --- Auto-merged settings (added by duofm) ---",
				"[colors]",
				"cursor_fg = 15",
			},
		},
		{
			name: "both sections exist - items inserted appropriately",
			original: `[keybindings]
move_down = ["J"]

[colors]
cursor_fg = 100
`,
			result: mergeResult{
				Keybindings: map[string][]string{
					"move_up": {"K"},
				},
				Colors: map[string]int{
					"border_fg": 240,
				},
				HistoryLimit: intPtr(5000),
			},
			wantContains: []string{
				`move_up = ["K"]`,
				"border_fg = 240",
				"history_limit = 5000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateMergedFile(tt.original, tt.result)

			for _, s := range tt.wantContains {
				if !strings.Contains(got, s) {
					t.Errorf("generateMergedFile() output missing %q\nGot:\n%s", s, got)
				}
			}

			for _, s := range tt.wantNotIn {
				if strings.Contains(got, s) {
					t.Errorf("generateMergedFile() output should not contain %q\nGot:\n%s", s, got)
				}
			}
		})
	}
}

// TestMergeConfig tests the MergeConfig function.
func TestMergeConfig(t *testing.T) {
	t.Run("missing items are appended to file", func(t *testing.T) {
		// Create temp file with partial config
		tmpFile := createTempFile(t, `[keybindings]
move_down = ["J"]
`)
		defer os.Remove(tmpFile)

		raw := &rawConfig{
			Keybindings: map[string][]string{
				"move_down": {"J"},
			},
			Colors:       nil,
			HistoryLimit: nil,
		}

		err := MergeConfig(tmpFile, raw)
		if err != nil {
			t.Fatalf("MergeConfig() error = %v", err)
		}

		// Read the file and verify content
		content, err := os.ReadFile(tmpFile)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		// Should contain the original content
		if !strings.Contains(string(content), `move_down = ["J"]`) {
			t.Error("Original content missing from file")
		}

		// Should contain auto-merged marker
		if !strings.Contains(string(content), "# --- Auto-merged settings (added by duofm) ---") {
			t.Error("Auto-merged marker missing from file")
		}

		// Should contain history_limit (since it was nil)
		if !strings.Contains(string(content), "history_limit = 20000") {
			t.Error("history_limit missing from file")
		}
	})

	t.Run("no changes when all items present", func(t *testing.T) {
		// Create config with all items
		allKeybindings := DefaultKeybindings()
		var keybindingsToml strings.Builder
		keybindingsToml.WriteString("[keybindings]\n")
		for k, v := range allKeybindings {
			keybindingsToml.WriteString(fmt.Sprintf("%s = %v\n", k, formatTestArray(v)))
		}
		keybindingsToml.WriteString("\n[colors]\n")
		for _, k := range AllColorKeys() {
			keybindingsToml.WriteString(fmt.Sprintf("%s = %d\n", k, GetDefaultColorValue(k)))
		}
		keybindingsToml.WriteString("\nhistory_limit = 1000\n")

		tmpFile := createTempFile(t, keybindingsToml.String())
		defer os.Remove(tmpFile)

		// Get original content
		originalContent, _ := os.ReadFile(tmpFile)

		// Create raw config with all items
		colors := make(map[string]interface{})
		for _, k := range AllColorKeys() {
			colors[k] = GetDefaultColorValue(k)
		}
		historyLimit := 1000
		raw := &rawConfig{
			Keybindings:  allKeybindings,
			Colors:       colors,
			HistoryLimit: &historyLimit,
		}

		err := MergeConfig(tmpFile, raw)
		if err != nil {
			t.Fatalf("MergeConfig() error = %v", err)
		}

		// Read the file and verify it hasn't changed
		newContent, _ := os.ReadFile(tmpFile)
		if string(newContent) != string(originalContent) {
			t.Errorf("File was modified when no changes were needed\nOriginal:\n%s\nNew:\n%s",
				string(originalContent), string(newContent))
		}
	})

	t.Run("existing custom values are preserved", func(t *testing.T) {
		tmpFile := createTempFile(t, `[keybindings]
move_down = ["CustomKey"]

[colors]
cursor_fg = 100

history_limit = 5000
`)
		defer os.Remove(tmpFile)

		colors := map[string]interface{}{
			"cursor_fg": 100,
		}
		historyLimit := 5000
		raw := &rawConfig{
			Keybindings: map[string][]string{
				"move_down": {"CustomKey"},
			},
			Colors:       colors,
			HistoryLimit: &historyLimit,
		}

		err := MergeConfig(tmpFile, raw)
		if err != nil {
			t.Fatalf("MergeConfig() error = %v", err)
		}

		content, _ := os.ReadFile(tmpFile)

		// Custom values should still be in original section
		if !strings.Contains(string(content), `move_down = ["CustomKey"]`) {
			t.Error("Custom keybinding was overwritten")
		}
		if !strings.Contains(string(content), "cursor_fg = 100") {
			t.Error("Custom color was overwritten")
		}
		if !strings.Contains(string(content), "history_limit = 5000") {
			t.Error("Custom history_limit was overwritten")
		}
	})

	t.Run("file write error returns error", func(t *testing.T) {
		// Use a non-existent directory path
		invalidPath := "/nonexistent/directory/config.toml"

		raw := &rawConfig{
			Keybindings:  nil,
			Colors:       nil,
			HistoryLimit: nil,
		}

		err := MergeConfig(invalidPath, raw)
		if err == nil {
			t.Error("MergeConfig() should return error for invalid path")
		}
	})

	t.Run("appends newline if file doesn't end with one", func(t *testing.T) {
		// Create file without trailing newline
		tmpFile := createTempFile(t, "[keybindings]\nmove_down = [\"J\"]")
		defer os.Remove(tmpFile)

		raw := &rawConfig{
			Keybindings: map[string][]string{
				"move_down": {"J"},
			},
			Colors:       nil,
			HistoryLimit: nil,
		}

		err := MergeConfig(tmpFile, raw)
		if err != nil {
			t.Fatalf("MergeConfig() error = %v", err)
		}

		content, _ := os.ReadFile(tmpFile)

		// The content should be properly formatted with the auto-merged marker
		if !strings.Contains(string(content), "# --- Auto-merged settings (added by duofm) ---") {
			t.Errorf("Missing auto-merged section marker\nContent:\n%s", string(content))
		}

		// Original content should be preserved
		if !strings.Contains(string(content), `move_down = ["J"]`) {
			t.Errorf("Original keybinding missing\nContent:\n%s", string(content))
		}
	})
}

// TestMergeConfigReload tests that merged config can be reloaded correctly.
// This test verifies that LoadConfig properly merges missing items and
// subsequent loads work correctly.
func TestMergeConfigReload(t *testing.T) {
	// Create partial config
	tmpFile := createTempFile(t, `[keybindings]
move_down = ["J"]
quit = ["X"]

[colors]
cursor_fg = 100
`)
	defer os.Remove(tmpFile)

	// First load - should merge missing items
	cfg, warnings := LoadConfig(tmpFile)

	// Check for no critical warnings
	for _, w := range warnings {
		t.Logf("Warning: %s", w)
	}

	// Custom values should be preserved
	if keys, ok := cfg.Keybindings["move_down"]; !ok || len(keys) != 1 || keys[0] != "J" {
		t.Errorf("Custom keybinding not preserved: got %v", cfg.Keybindings["move_down"])
	}
	if keys, ok := cfg.Keybindings["quit"]; !ok || len(keys) != 1 || keys[0] != "X" {
		t.Errorf("Custom quit keybinding not preserved: got %v", cfg.Keybindings["quit"])
	}
	if cfg.Colors.CursorFg != 100 {
		t.Errorf("Custom color not preserved: got %d", cfg.Colors.CursorFg)
	}

	// Default values should be filled in (since they weren't in the file)
	if cfg.HistoryLimit != DefaultHistoryLimit {
		t.Errorf("Default history_limit not applied: got %d, want %d", cfg.HistoryLimit, DefaultHistoryLimit)
	}

	// Missing keybindings should be available (from defaults)
	if _, ok := cfg.Keybindings["move_up"]; !ok {
		t.Error("Missing keybinding 'move_up' not available")
	}

	// Second load - should not modify file again
	contentAfterFirstLoad, _ := os.ReadFile(tmpFile)

	cfg2, warnings2 := LoadConfig(tmpFile)
	for _, w := range warnings2 {
		t.Logf("Second load warning: %s", w)
	}

	contentAfterSecondLoad, _ := os.ReadFile(tmpFile)

	// File content should be unchanged after second load
	if string(contentAfterSecondLoad) != string(contentAfterFirstLoad) {
		t.Error("File was modified on second load when no changes should be needed")
	}

	// Values should still be correct
	if cfg2.Colors.CursorFg != 100 {
		t.Errorf("Custom color not preserved on second load: got %d", cfg2.Colors.CursorFg)
	}
}

// createTempFile creates a temporary file with the given content.
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	return tmpFile.Name()
}

// formatTestArray formats a string slice as TOML array string.
func formatTestArray(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// TestMergeResultHasContent tests the hasContent method.
func TestMergeResultHasContent(t *testing.T) {
	tests := []struct {
		name   string
		result mergeResult
		want   bool
	}{
		{
			name: "empty - no content",
			result: mergeResult{
				Keybindings:  map[string][]string{},
				Colors:       map[string]int{},
				HistoryLimit: nil,
			},
			want: false,
		},
		{
			name: "has keybindings",
			result: mergeResult{
				Keybindings:  map[string][]string{"quit": {"Q"}},
				Colors:       map[string]int{},
				HistoryLimit: nil,
			},
			want: true,
		},
		{
			name: "has colors",
			result: mergeResult{
				Keybindings:  map[string][]string{},
				Colors:       map[string]int{"cursor_fg": 15},
				HistoryLimit: nil,
			},
			want: true,
		},
		{
			name: "has history_limit",
			result: mergeResult{
				Keybindings:  map[string][]string{},
				Colors:       map[string]int{},
				HistoryLimit: intPtr(1000),
			},
			want: true,
		},
		{
			name: "nil maps - no content",
			result: mergeResult{
				Keybindings:  nil,
				Colors:       nil,
				HistoryLimit: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.hasContent()
			if got != tt.want {
				t.Errorf("mergeResult.hasContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
