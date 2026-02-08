package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigPath_Default(t *testing.T) {
	// Unset XDG_CONFIG_HOME to test default behavior
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Unsetenv("XDG_CONFIG_HOME")
	defer func() {
		if oldXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", oldXDG)
		}
	}()

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() returned error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "duofm", "config.toml")
	if path != expected {
		t.Errorf("GetConfigPath() = %q, want %q", path, expected)
	}
}

func TestGetConfigPath_WithXDG(t *testing.T) {
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if oldXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", oldXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() returned error: %v", err)
	}

	expected := filepath.Join(tmpDir, "duofm", "config.toml")
	if path != expected {
		t.Errorf("GetConfigPath() = %q, want %q", path, expected)
	}
}

func TestLoadConfig_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent", "config.toml")

	cfg, warnings := LoadConfig(configPath)

	// Should return default config without error
	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Should have no warnings for missing file
	if len(warnings) != 0 {
		t.Errorf("LoadConfig() returned %d warnings, want 0", len(warnings))
	}

	// Should have default keybindings
	if len(cfg.Keybindings) == 0 {
		t.Error("LoadConfig() returned empty keybindings, want defaults")
	}
}

func TestLoadConfig_ValidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `[keybindings]
move_down = ["J", "Down"]
move_up = ["K", "Up"]
help = ["?"]
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, warnings := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	if len(warnings) != 0 {
		t.Errorf("LoadConfig() returned %d warnings, want 0: %v", len(warnings), warnings)
	}

	// Check parsed values
	if keys, ok := cfg.Keybindings["move_down"]; !ok || len(keys) != 2 {
		t.Errorf("move_down = %v, want [J, Down]", keys)
	}
}

func TestLoadConfig_ParseError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `[keybindings
invalid toml`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, warnings := LoadConfig(configPath)

	// Should return default config
	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Should have parse error warning
	if len(warnings) == 0 {
		t.Error("LoadConfig() returned no warnings, want parse error warning")
	}

	// Should use default keybindings
	if len(cfg.Keybindings) == 0 {
		t.Error("LoadConfig() returned empty keybindings on parse error")
	}
}

func TestLoadConfig_MissingKeybindingsSection(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `# Empty config file
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, warnings := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// No warning for missing section, just use defaults
	if len(warnings) != 0 {
		t.Errorf("LoadConfig() returned warnings for missing section: %v", warnings)
	}

	// Should have default keybindings
	if len(cfg.Keybindings) == 0 {
		t.Error("LoadConfig() returned empty keybindings")
	}
}

func TestLoadConfig_EmptyArray(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `[keybindings]
help = []
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, warnings := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	if len(warnings) != 0 {
		t.Errorf("LoadConfig() returned warnings: %v", warnings)
	}

	// help should be empty array (disabled action)
	if keys, ok := cfg.Keybindings["help"]; !ok {
		t.Error("help key not found in keybindings")
	} else if len(keys) != 0 {
		t.Errorf("help = %v, want []", keys)
	}
}

func TestGenerateDefaultConfig_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nested", "dir", "config.toml")

	err := GenerateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() returned error: %v", err)
	}

	// Check directory was created
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Directory %s was not created", dir)
	}

	// Check file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("File %s was not created", configPath)
	}
}

func TestGenerateDefaultConfig_ValidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	err := GenerateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() returned error: %v", err)
	}

	// Try to parse the generated file
	cfg, warnings := LoadConfig(configPath)
	if cfg == nil {
		t.Fatal("Generated config could not be loaded")
	}

	if len(warnings) != 0 {
		t.Errorf("Generated config has warnings: %v", warnings)
	}

	// Check that all 28 actions are present
	actions := AllActions()
	for _, action := range actions {
		if _, ok := cfg.Keybindings[action]; !ok {
			t.Errorf("Action %q not found in generated config", action)
		}
	}
}

func TestGenerateDefaultConfig_Under150Lines(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	err := GenerateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() returned error: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated config: %v", err)
	}

	lines := 0
	for _, b := range content {
		if b == '\n' {
			lines++
		}
	}
	// Add 1 for the last line if it doesn't end with newline
	if len(content) > 0 && content[len(content)-1] != '\n' {
		lines++
	}

	// Limit increased from 100 to 150 to accommodate [colors] section
	if lines > 150 {
		t.Errorf("Generated config has %d lines, want <= 150", lines)
	}
}

func TestGenerateDefaultConfig_HasComments(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	err := GenerateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() returned error: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated config: %v", err)
	}

	contentStr := string(content)

	// Check for section comments
	if !contains(contentStr, "# Navigation") {
		t.Error("Generated config missing Navigation section comment")
	}
	if !contains(contentStr, "# File operations") {
		t.Error("Generated config missing File operations section comment")
	}
	if !contains(contentStr, "[keybindings]") {
		t.Error("Generated config missing [keybindings] section")
	}
	// Check for [colors] section
	if !contains(contentStr, "[colors]") {
		t.Error("Generated config missing [colors] section")
	}
	if !contains(contentStr, "# Color Theme Configuration") {
		t.Error("Generated config missing Color Theme Configuration comment")
	}
	if !contains(contentStr, "cursor_fg") {
		t.Error("Generated config missing cursor_fg color example")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Tests for HistoryLimit configuration
func TestLoadConfig_HistoryLimitDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `[keybindings]
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, _ := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Default history_limit should be 20000
	if cfg.HistoryLimit != 20000 {
		t.Errorf("HistoryLimit = %d, want 20000", cfg.HistoryLimit)
	}
}

func TestLoadConfig_HistoryLimitExplicit(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `history_limit = 5000

[keybindings]
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, _ := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	if cfg.HistoryLimit != 5000 {
		t.Errorf("HistoryLimit = %d, want 5000", cfg.HistoryLimit)
	}
}

func TestLoadConfig_HistoryLimitZero(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `history_limit = 0

[keybindings]
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, _ := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// history_limit = 0 should disable history
	if cfg.HistoryLimit != 0 {
		t.Errorf("HistoryLimit = %d, want 0", cfg.HistoryLimit)
	}
}

func TestLoadConfig_HistoryLimitFileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent", "config.toml")

	cfg, _ := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Default history_limit should be 20000
	if cfg.HistoryLimit != 20000 {
		t.Errorf("HistoryLimit = %d, want 20000", cfg.HistoryLimit)
	}
}

// Tests for RefreshRate configuration
func TestLoadConfig_RefreshRateDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `[keybindings]
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, _ := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	if cfg.RefreshRate != DefaultRefreshRate {
		t.Errorf("RefreshRate = %d, want %d", cfg.RefreshRate, DefaultRefreshRate)
	}
}

func TestLoadConfig_RefreshRateExplicit(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `refresh_rate = 5

[keybindings]
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, _ := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	if cfg.RefreshRate != 5 {
		t.Errorf("RefreshRate = %d, want 5", cfg.RefreshRate)
	}
}

func TestLoadConfig_RefreshRateZero(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `refresh_rate = 0

[keybindings]
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, _ := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	if cfg.RefreshRate != 0 {
		t.Errorf("RefreshRate = %d, want 0", cfg.RefreshRate)
	}
}

func TestLoadConfig_RefreshRateBoundary(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
		hasWarn  bool
	}{
		{"min boundary 1", 1, 1, false},
		{"max boundary 60", 60, 60, false},
		{"below min -1", -1, DefaultRefreshRate, true},
		{"above max 61", 61, DefaultRefreshRate, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")

			content := fmt.Sprintf("refresh_rate = %d\n\n[keybindings]\nquit = [\"Q\"]\n", tt.value)
			if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			cfg, warnings := LoadConfig(configPath)

			if cfg == nil {
				t.Fatal("LoadConfig() returned nil config")
			}

			if cfg.RefreshRate != tt.expected {
				t.Errorf("RefreshRate = %d, want %d", cfg.RefreshRate, tt.expected)
			}

			hasWarning := false
			for _, w := range warnings {
				if strings.Contains(w, "refresh_rate") {
					hasWarning = true
					break
				}
			}
			if tt.hasWarn && !hasWarning {
				t.Error("Expected warning for out-of-range refresh_rate")
			}
			if !tt.hasWarn && hasWarning {
				t.Error("Unexpected warning for valid refresh_rate")
			}
		})
	}
}

// Tests for EnterBehavior configuration
func TestLoadConfig_EnterBehaviorDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `[keybindings]
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, _ := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Default enter_behavior should be "less"
	if cfg.EnterBehavior.Type != EnterBehaviorLess {
		t.Errorf("EnterBehavior.Type = %v, want EnterBehaviorLess", cfg.EnterBehavior.Type)
	}
}

func TestLoadConfig_EnterBehaviorLess(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `enter_behavior = "less"

[keybindings]
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, warnings := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Should have no warnings for valid value
	for _, w := range warnings {
		if strings.Contains(w, "enter_behavior") {
			t.Errorf("Unexpected warning about enter_behavior: %s", w)
		}
	}

	if cfg.EnterBehavior.Type != EnterBehaviorLess {
		t.Errorf("EnterBehavior.Type = %v, want EnterBehaviorLess", cfg.EnterBehavior.Type)
	}
}

func TestLoadConfig_EnterBehaviorXDGOpen(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `enter_behavior = "xdg-open"

[keybindings]
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, _ := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	if cfg.EnterBehavior.Type != EnterBehaviorXDGOpen {
		t.Errorf("EnterBehavior.Type = %v, want EnterBehaviorXDGOpen", cfg.EnterBehavior.Type)
	}
}

func TestLoadConfig_EnterBehaviorCustom(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `enter_behavior = "path:/usr/bin/vim"

[keybindings]
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, _ := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	if cfg.EnterBehavior.Type != EnterBehaviorCustom {
		t.Errorf("EnterBehavior.Type = %v, want EnterBehaviorCustom", cfg.EnterBehavior.Type)
	}

	if cfg.EnterBehavior.CustomPath != "/usr/bin/vim" {
		t.Errorf("EnterBehavior.CustomPath = %q, want /usr/bin/vim", cfg.EnterBehavior.CustomPath)
	}
}

func TestLoadConfig_EnterBehaviorInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `enter_behavior = "unknown"

[keybindings]
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, warnings := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Should have warning for invalid value
	hasWarning := false
	for _, w := range warnings {
		if strings.Contains(w, "enter_behavior") && strings.Contains(w, "invalid") {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("Expected warning for invalid enter_behavior value")
	}

	// Should fall back to default (less)
	if cfg.EnterBehavior.Type != EnterBehaviorLess {
		t.Errorf("EnterBehavior.Type = %v, want EnterBehaviorLess (default)", cfg.EnterBehavior.Type)
	}
}

func TestLoadConfig_EnterBehaviorFileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent", "config.toml")

	cfg, _ := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Default enter_behavior should be "less"
	if cfg.EnterBehavior.Type != EnterBehaviorLess {
		t.Errorf("EnterBehavior.Type = %v, want EnterBehaviorLess", cfg.EnterBehavior.Type)
	}
}

func TestLoadConfig_EnterBehaviorMIME(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `enter_behavior = "mime:"

[keybindings]
quit = ["Q"]

[enter_behavior_mime]
"text/plain" = ["less", "cat"]
"image/*" = ["feh", "eog"]
"application/pdf" = ["zathura"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, warnings := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Should have no unexpected warnings
	for _, w := range warnings {
		if strings.Contains(w, "enter_behavior") && strings.Contains(w, "invalid") {
			t.Errorf("Unexpected warning about enter_behavior: %s", w)
		}
	}

	if cfg.EnterBehavior.Type != EnterBehaviorMIME {
		t.Errorf("EnterBehavior.Type = %v, want EnterBehaviorMIME", cfg.EnterBehavior.Type)
	}

	// Check MIME behavior rules
	if cfg.MIMEBehavior.Rules == nil {
		t.Fatal("MIMEBehavior.Rules is nil")
	}

	if len(cfg.MIMEBehavior.Rules) != 3 {
		t.Errorf("MIMEBehavior.Rules count = %d, want 3", len(cfg.MIMEBehavior.Rules))
	}

	// Check text/plain rule
	if cmds, ok := cfg.MIMEBehavior.Rules["text/plain"]; !ok {
		t.Error("text/plain rule not found")
	} else if len(cmds) != 2 || cmds[0] != "less" || cmds[1] != "cat" {
		t.Errorf("text/plain commands = %v, want [less cat]", cmds)
	}

	// Check image/* rule
	if cmds, ok := cfg.MIMEBehavior.Rules["image/*"]; !ok {
		t.Error("image/* rule not found")
	} else if len(cmds) != 2 {
		t.Errorf("image/* commands = %v, want [feh eog]", cmds)
	}
}

func TestLoadConfig_MIMEWithoutSection(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `enter_behavior = "mime:"

[keybindings]
quit = ["Q"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, _ := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	if cfg.EnterBehavior.Type != EnterBehaviorMIME {
		t.Errorf("EnterBehavior.Type = %v, want EnterBehaviorMIME", cfg.EnterBehavior.Type)
	}

	// MIMEBehavior should have empty rules
	if cfg.MIMEBehavior.Rules == nil {
		t.Error("MIMEBehavior.Rules should not be nil")
	}

	if len(cfg.MIMEBehavior.Rules) != 0 {
		t.Errorf("MIMEBehavior.Rules count = %d, want 0", len(cfg.MIMEBehavior.Rules))
	}
}

func TestLoadConfig_MIMEInvalidEntries(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `enter_behavior = "mime:"

[keybindings]
quit = ["Q"]

[enter_behavior_mime]
"text/plain" = ["less"]
"" = ["cat"]
"image/*" = []
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, warnings := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Should have warnings for invalid entries
	warningCount := 0
	for _, w := range warnings {
		if strings.Contains(w, "empty MIME type") || strings.Contains(w, "empty command list") {
			warningCount++
		}
	}
	if warningCount != 2 {
		t.Errorf("Expected 2 warnings for invalid entries, got %d", warningCount)
	}

	// Only valid rule should be stored
	if len(cfg.MIMEBehavior.Rules) != 1 {
		t.Errorf("MIMEBehavior.Rules count = %d, want 1", len(cfg.MIMEBehavior.Rules))
	}
}

func TestLoadConfig_NonMIMEIgnoresMIMESection(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `enter_behavior = "less"

[keybindings]
quit = ["Q"]

[enter_behavior_mime]
"text/plain" = ["less"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, _ := LoadConfig(configPath)

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	if cfg.EnterBehavior.Type != EnterBehaviorLess {
		t.Errorf("EnterBehavior.Type = %v, want EnterBehaviorLess", cfg.EnterBehavior.Type)
	}

	// MIMEBehavior should be empty when enter_behavior is not "mime:"
	if cfg.MIMEBehavior.Rules != nil && len(cfg.MIMEBehavior.Rules) > 0 {
		t.Errorf("MIMEBehavior.Rules should be empty when not using mime: behavior, got %v", cfg.MIMEBehavior.Rules)
	}
}

func TestGenerateDefaultConfig_HasMIMEOption(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	err := GenerateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() returned error: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated config: %v", err)
	}

	contentStr := string(content)

	// Check for mime: option in comments
	if !contains(contentStr, "mime:") {
		t.Error("Generated config missing mime: option")
	}

	// Check for enter_behavior_mime section example
	if !contains(contentStr, "enter_behavior_mime") {
		t.Error("Generated config missing enter_behavior_mime section")
	}

	// Check for example MIME patterns
	if !contains(contentStr, "text/plain") {
		t.Error("Generated config missing text/plain example")
	}

	if !contains(contentStr, "image/*") {
		t.Error("Generated config missing image/* example")
	}
}
