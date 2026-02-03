package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestRepairConfig_SyntaxError_RemovesErrorLines(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `history_limit = 5000

[keybindings]
quit = ["q"]

!!!syntax error here
more bad content
`
	os.WriteFile(configPath, []byte(content), 0644)

	result := LoadConfigDetailed(configPath)
	if !result.HasSyntaxErr {
		t.Fatal("Expected syntax error")
	}

	err := RepairConfig(configPath, result)
	if err != nil {
		t.Fatalf("RepairConfig failed: %v", err)
	}

	// Read repaired file
	repaired, _ := os.ReadFile(configPath)
	repairedStr := string(repaired)

	// Should not contain the error lines
	if strings.Contains(repairedStr, "!!!syntax error") {
		t.Error("Repaired file should not contain syntax error line")
	}

	// Should still contain valid config before the error
	if !strings.Contains(repairedStr, "history_limit = 5000") {
		t.Error("Repaired file should preserve history_limit")
	}

	if !strings.Contains(repairedStr, `quit = ["q"]`) {
		t.Error("Repaired file should preserve keybinding")
	}
}

func TestRepairConfig_SyntaxError_ValidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `history_limit = 5000

[keybindings]
quit = ["q"]

!!!syntax error
`
	os.WriteFile(configPath, []byte(content), 0644)

	result := LoadConfigDetailed(configPath)
	err := RepairConfig(configPath, result)
	if err != nil {
		t.Fatalf("RepairConfig failed: %v", err)
	}

	// Verify the repaired file is valid TOML
	repaired, _ := os.ReadFile(configPath)
	var raw rawConfig
	if _, err := toml.Decode(string(repaired), &raw); err != nil {
		t.Errorf("Repaired file is not valid TOML: %v", err)
	}
}

func TestRepairConfig_ValueError_ReplacesInvalidValues(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `history_limit = 5000
enter_behavior = "invalid_value"

[keybindings]
quit = ["q"]

[colors]
cursor_fg = 999
cursor_bg = 39
`
	os.WriteFile(configPath, []byte(content), 0644)

	result := LoadConfigDetailed(configPath)
	if result.HasSyntaxErr {
		t.Fatal("Expected no syntax error")
	}
	if len(result.Errors) == 0 {
		t.Fatal("Expected value errors")
	}

	err := RepairConfig(configPath, result)
	if err != nil {
		t.Fatalf("RepairConfig failed: %v", err)
	}

	// Read repaired file and verify it's valid
	repaired, _ := os.ReadFile(configPath)
	repairedStr := string(repaired)

	// Should not contain the invalid value
	if strings.Contains(repairedStr, "invalid_value") {
		t.Error("Repaired file should not contain invalid enter_behavior value")
	}

	// Should still contain valid settings
	if !strings.Contains(repairedStr, "history_limit = 5000") {
		t.Error("Repaired file should preserve valid history_limit")
	}
	if !strings.Contains(repairedStr, "cursor_bg = 39") {
		t.Error("Repaired file should preserve valid cursor_bg")
	}
}

func TestRepairConfig_ValueError_PreservesOtherSettings(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `history_limit = 5000
enter_behavior = "invalid_value"

[keybindings]
quit = ["q"]
move_down = ["j", "Down"]
`
	os.WriteFile(configPath, []byte(content), 0644)

	result := LoadConfigDetailed(configPath)
	err := RepairConfig(configPath, result)
	if err != nil {
		t.Fatalf("RepairConfig failed: %v", err)
	}

	repaired, _ := os.ReadFile(configPath)
	repairedStr := string(repaired)

	// keybindings should be fully preserved
	if !strings.Contains(repairedStr, `quit = ["q"]`) {
		t.Error("Repaired file should preserve quit keybinding")
	}
	if !strings.Contains(repairedStr, "move_down") {
		t.Error("Repaired file should preserve move_down keybinding")
	}
}

func TestRepairConfig_PreservesPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `!!!syntax error`
	os.WriteFile(configPath, []byte(content), 0640)

	result := LoadConfigDetailed(configPath)
	err := RepairConfig(configPath, result)
	if err != nil {
		t.Fatalf("RepairConfig failed: %v", err)
	}

	info, _ := os.Stat(configPath)
	if info.Mode().Perm() != 0640 {
		t.Errorf("Expected permissions 0640, got %v", info.Mode().Perm())
	}
}

func TestRepairSyntaxError(t *testing.T) {
	content := `line1 = "a"
line2 = "b"
!!!error
line4 = "d"
`
	result := repairSyntaxError(content, 3)
	if strings.Contains(result, "!!!error") {
		t.Error("Should remove error line and below")
	}
	if !strings.Contains(result, `line1 = "a"`) {
		t.Error("Should preserve line1")
	}
	if !strings.Contains(result, `line2 = "b"`) {
		t.Error("Should preserve line2")
	}
}

func TestRepairValueErrors(t *testing.T) {
	content := `history_limit = 5000
enter_behavior = "bad_value"

[colors]
cursor_fg = 999
cursor_bg = 39
`
	errors := []ConfigError{
		{Field: "enter_behavior", Message: "invalid", Line: 2},
		{Field: "colors.cursor_fg", Message: "out of range", Line: 5},
	}

	result := repairValueErrors(content, errors)

	if strings.Contains(result, "bad_value") {
		t.Error("Should replace bad enter_behavior value")
	}
	if strings.Contains(result, "999") {
		t.Error("Should replace out-of-range color value")
	}
	if !strings.Contains(result, "cursor_bg = 39") {
		t.Error("Should preserve valid cursor_bg")
	}
}
