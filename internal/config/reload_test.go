package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDetailed_NormalFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `history_limit = 5000
enter_behavior = "less"

[keybindings]
quit = ["q"]

[colors]
cursor_fg = 15
`
	os.WriteFile(configPath, []byte(content), 0644)

	result := LoadConfigDetailed(configPath)

	if result.HasErrors() {
		t.Errorf("Expected no errors, got HasSyntaxErr=%v, Errors=%v", result.HasSyntaxErr, result.Errors)
	}

	if result.Config == nil {
		t.Fatal("Expected Config to be non-nil")
	}

	if result.Config.HistoryLimit != 5000 {
		t.Errorf("Expected HistoryLimit=5000, got %d", result.Config.HistoryLimit)
	}
}

func TestLoadConfigDetailed_FileNotExist(t *testing.T) {
	result := LoadConfigDetailed("/nonexistent/config.toml")

	if result.HasErrors() {
		t.Error("Expected no errors for non-existent file")
	}

	if result.Config == nil {
		t.Fatal("Expected default Config, got nil")
	}

	if result.Config.HistoryLimit != DefaultHistoryLimit {
		t.Errorf("Expected default HistoryLimit=%d, got %d", DefaultHistoryLimit, result.Config.HistoryLimit)
	}
}

func TestLoadConfigDetailed_SyntaxError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `history_limit = 5000

[keybindings]
quit = ["q"]

[colors]
cursor_fg = ???invalid
cursor_bg = 39
`
	os.WriteFile(configPath, []byte(content), 0644)

	result := LoadConfigDetailed(configPath)

	if !result.HasSyntaxErr {
		t.Error("Expected HasSyntaxErr=true")
	}

	if result.SyntaxErrLine <= 0 {
		t.Errorf("Expected positive SyntaxErrLine, got %d", result.SyntaxErrLine)
	}

	if result.Config == nil {
		t.Fatal("Expected Config to be non-nil (with defaults for broken parts)")
	}
}

func TestLoadConfigDetailed_SyntaxError_PartialParse(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Error is on line 5, so lines 1-4 should be parsed
	content := `history_limit = 3000

[keybindings]
quit = ["q"]
!!!syntax error here
cursor_bg = 39
`
	os.WriteFile(configPath, []byte(content), 0644)

	result := LoadConfigDetailed(configPath)

	if !result.HasSyntaxErr {
		t.Error("Expected HasSyntaxErr=true")
	}

	if result.Config == nil {
		t.Fatal("Expected Config to be non-nil")
	}

	// history_limit=3000 is before the error line, so it should be parsed
	if result.Config.HistoryLimit != 3000 {
		t.Errorf("Expected partial parse to get HistoryLimit=3000, got %d", result.Config.HistoryLimit)
	}
}

func TestLoadConfigDetailed_SyntaxError_PartialParseFallbackToDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Syntax error on line 2 means no useful content can be parsed before it.
	// partialParse with upToLine=2 will try to parse only line 1 which is valid,
	// but for a case where the error is on line 1, we get full defaults.
	content := `!!!syntax error on first line
history_limit = 5000
`
	os.WriteFile(configPath, []byte(content), 0644)

	result := LoadConfigDetailed(configPath)

	if !result.HasSyntaxErr {
		t.Error("Expected HasSyntaxErr=true")
	}

	if result.Config == nil {
		t.Fatal("Expected Config (default fallback) to be non-nil")
	}

	// Syntax error on line 1 means nothing can be parsed
	// Config should be defaults
	if result.Config.HistoryLimit != DefaultHistoryLimit {
		t.Errorf("Expected default HistoryLimit=%d, got %d", DefaultHistoryLimit, result.Config.HistoryLimit)
	}
}

func TestLoadConfigDetailed_ValueError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `history_limit = 5000
enter_behavior = "invalid_value"

[keybindings]
quit = ["q"]

[colors]
cursor_fg = 999
`
	os.WriteFile(configPath, []byte(content), 0644)

	result := LoadConfigDetailed(configPath)

	if result.HasSyntaxErr {
		t.Error("Expected no syntax error for value errors")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected value errors to be present")
	}

	// Check that error fields are correct
	hasEnterBehaviorError := false
	hasColorError := false
	for _, e := range result.Errors {
		if e.Field == "enter_behavior" {
			hasEnterBehaviorError = true
		}
		if e.Field == "colors.cursor_fg" {
			hasColorError = true
		}
	}

	if !hasEnterBehaviorError {
		t.Error("Expected error for enter_behavior field")
	}
	if !hasColorError {
		t.Error("Expected error for colors.cursor_fg field")
	}
}

func TestLoadConfigDetailed_ValueError_NormalFieldsPreserved(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `history_limit = 3000
enter_behavior = "invalid_value"

[keybindings]
quit = ["q"]
`
	os.WriteFile(configPath, []byte(content), 0644)

	result := LoadConfigDetailed(configPath)

	if result.Config == nil {
		t.Fatal("Expected Config to be non-nil")
	}

	// history_limit should be preserved (it's valid)
	if result.Config.HistoryLimit != 3000 {
		t.Errorf("Expected HistoryLimit=3000, got %d", result.Config.HistoryLimit)
	}

	// enter_behavior should be reset to default (it's invalid)
	if result.Config.EnterBehavior.Type != EnterBehaviorLess {
		t.Errorf("Expected default EnterBehavior, got %v", result.Config.EnterBehavior)
	}
}

func TestLoadConfigDetailed_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	os.WriteFile(configPath, []byte(""), 0644)

	result := LoadConfigDetailed(configPath)

	if result.HasErrors() {
		t.Error("Expected no errors for empty file")
	}

	if result.Config == nil {
		t.Fatal("Expected Config to be non-nil")
	}

	if result.Config.HistoryLimit != DefaultHistoryLimit {
		t.Errorf("Expected default HistoryLimit, got %d", result.Config.HistoryLimit)
	}
}

func TestPartialParse_ValidContent(t *testing.T) {
	content := `history_limit = 3000

[keybindings]
quit = ["q"]

[colors]
cursor_fg = 15
cursor_bg = 39
extra line that will be cut
`
	raw, err := partialParse(content, 8) // cut at line 8, keeping lines 1-7

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if raw == nil {
		t.Fatal("Expected non-nil rawConfig")
	}

	if raw.HistoryLimit == nil || *raw.HistoryLimit != 3000 {
		t.Error("Expected HistoryLimit=3000 from partial parse")
	}
}

func TestPartialParse_InvalidCut(t *testing.T) {
	// Cutting in the middle of a multiline string produces invalid TOML
	content := `multi = """
line1
line2
"""
`
	_, err := partialParse(content, 3) // cut at line 3, in the middle of multiline

	if err == nil {
		t.Error("Expected error for invalid partial TOML")
	}
}
