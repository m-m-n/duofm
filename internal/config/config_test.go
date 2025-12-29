package config

import (
	"os"
	"path/filepath"
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

func TestGenerateDefaultConfig_Under100Lines(t *testing.T) {
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

	if lines > 100 {
		t.Errorf("Generated config has %d lines, want <= 100", lines)
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
