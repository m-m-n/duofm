package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	err := GenerateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("GenerateDefaultConfig() did not create config file")
	}

	// Read the content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated config: %v", err)
	}

	contentStr := string(content)

	// Verify file contains expected sections
	if !strings.Contains(contentStr, "[keybindings]") {
		t.Error("Generated config missing [keybindings] section")
	}

	if !strings.Contains(contentStr, "[colors]") {
		t.Error("Generated config missing [colors] section")
	}

	// Verify file contains key comments
	if !strings.Contains(contentStr, "# duofm configuration file") {
		t.Error("Generated config missing header comment")
	}

	// Verify file contains keybinding examples
	keybindingExamples := []string{
		"move_down",
		"move_up",
		"copy",
		"delete",
		"quit",
	}

	for _, example := range keybindingExamples {
		if !strings.Contains(contentStr, example) {
			t.Errorf("Generated config missing keybinding example: %s", example)
		}
	}
}

func TestGenerateDefaultConfig_CreatesNestedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "level1", "level2", "config.toml")

	err := GenerateDefaultConfig(nestedPath)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() error = %v", err)
	}

	// Verify directory was created
	parentDir := filepath.Dir(nestedPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Error("GenerateDefaultConfig() did not create parent directories")
	}

	// Verify file was created
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("GenerateDefaultConfig() did not create config file in nested directory")
	}
}

func TestGenerateDefaultConfig_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create existing file
	existingContent := "existing content"
	if err := os.WriteFile(configPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Generate default config (should overwrite)
	err := GenerateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() error = %v", err)
	}

	// Verify content was overwritten
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	if string(content) == existingContent {
		t.Error("GenerateDefaultConfig() did not overwrite existing file")
	}

	if !strings.Contains(string(content), "[keybindings]") {
		t.Error("GenerateDefaultConfig() did not write expected content")
	}
}

func TestGenerateDefaultConfig_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	err := GenerateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() error = %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	// Check file permissions (0644 = -rw-r--r--)
	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Errorf("File permissions = %o, want 0644", perm)
	}
}

func TestGenerateDefaultConfig_ProducesValidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	err := GenerateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() error = %v", err)
	}

	// Try to load the generated config
	cfg, warnings := LoadConfig(configPath)

	if len(warnings) > 0 {
		t.Errorf("Generated config has warnings: %v", warnings)
	}

	// Verify config has expected structure
	if cfg.Keybindings == nil {
		t.Error("Loaded config has nil Keybindings")
	}
}

func TestGenerateDefaultConfig_ContentStructure(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	err := GenerateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() error = %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)

	// Verify section comments
	sections := []string{
		"# Navigation",
		"# File operations",
		"# Display",
		"# Navigation extended",
		"# Search",
		"# External applications",
		"# Application",
	}

	for _, section := range sections {
		if !strings.Contains(contentStr, section) {
			t.Errorf("Generated config missing section comment: %s", section)
		}
	}

	// Verify keybinding format examples
	keybindingFormats := []string{
		`move_down = ["J", "Down"]`,
		`copy = ["C"]`,
		`enter = ["Enter"]`,
		`toggle_hidden = ["Ctrl+H"]`,
	}

	for _, format := range keybindingFormats {
		if !strings.Contains(contentStr, format) {
			t.Errorf("Generated config missing keybinding format: %s", format)
		}
	}
}

func TestGenerateDefaultConfig_ColorsSectionCommented(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	err := GenerateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() error = %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)

	// Verify color settings are commented out (start with #)
	colorSettings := []string{
		"# cursor_fg",
		"# cursor_bg",
		"# directory_fg",
		"# error_fg",
	}

	for _, setting := range colorSettings {
		if !strings.Contains(contentStr, setting) {
			t.Errorf("Generated config color setting not commented: %s", setting)
		}
	}
}

func TestGenerateDefaultConfig_ErrorOnInvalidPath(t *testing.T) {
	// Try to create config in a directory that we can't write to
	// This test is platform-dependent, so we'll skip if running as root
	if os.Getuid() == 0 {
		t.Skip("Skipping test when running as root")
	}

	// Create a read-only directory
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}

	configPath := filepath.Join(readOnlyDir, "config.toml")

	err := GenerateDefaultConfig(configPath)
	if err == nil {
		// Clean up the file if it was somehow created
		os.Remove(configPath)
		t.Error("GenerateDefaultConfig() should fail on read-only directory")
	}
}

func TestDefaultConfigTemplate_NotEmpty(t *testing.T) {
	if defaultConfigTemplate == "" {
		t.Error("defaultConfigTemplate is empty")
	}

	if len(defaultConfigTemplate) < 100 {
		t.Error("defaultConfigTemplate seems too short")
	}
}
