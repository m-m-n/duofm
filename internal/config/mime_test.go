package config

import (
	"strings"
	"testing"
)

func TestParseMIMEBehavior(t *testing.T) {
	tests := []struct {
		name            string
		input           map[string][]string
		expectedRules   int
		expectWarnings  bool
		warningContains string
	}{
		{
			name: "valid single rule",
			input: map[string][]string{
				"text/plain": {"less", "cat"},
			},
			expectedRules:  1,
			expectWarnings: false,
		},
		{
			name: "valid multiple rules",
			input: map[string][]string{
				"text/plain":    {"less"},
				"image/*":       {"feh", "eog"},
				"application/*": {"xdg-open"},
			},
			expectedRules:  3,
			expectWarnings: false,
		},
		{
			name:           "empty map",
			input:          map[string][]string{},
			expectedRules:  0,
			expectWarnings: false,
		},
		{
			name:           "nil map",
			input:          nil,
			expectedRules:  0,
			expectWarnings: false,
		},
		{
			name: "empty key generates warning",
			input: map[string][]string{
				"":          {"less"},
				"text/html": {"less"},
			},
			expectedRules:   1,
			expectWarnings:  true,
			warningContains: "empty MIME type",
		},
		{
			name: "empty command array generates warning",
			input: map[string][]string{
				"text/plain": {},
				"text/html":  {"less"},
			},
			expectedRules:   1,
			expectWarnings:  true,
			warningContains: "empty command list",
		},
		{
			name: "multiple warnings",
			input: map[string][]string{
				"":           {"less"},
				"text/plain": {},
			},
			expectedRules:  0,
			expectWarnings: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, warnings := ParseMIMEBehavior(tt.input)

			if len(result.Rules) != tt.expectedRules {
				t.Errorf("Rules count = %d, want %d", len(result.Rules), tt.expectedRules)
			}

			if tt.expectWarnings {
				if len(warnings) == 0 {
					t.Error("expected warnings but got none")
				} else if tt.warningContains != "" {
					found := false
					for _, w := range warnings {
						if strings.Contains(w, tt.warningContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("warnings %v should contain %q", warnings, tt.warningContains)
					}
				}
			} else {
				if len(warnings) > 0 {
					t.Errorf("unexpected warnings: %v", warnings)
				}
			}
		})
	}
}

func TestParseMIMEBehavior_RuleContent(t *testing.T) {
	input := map[string][]string{
		"text/plain": {"less", "cat"},
		"image/*":    {"feh"},
	}

	result, _ := ParseMIMEBehavior(input)

	// Check text/plain rule
	if cmds, ok := result.Rules["text/plain"]; !ok {
		t.Error("expected text/plain rule")
	} else if len(cmds) != 2 {
		t.Errorf("text/plain commands = %d, want 2", len(cmds))
	} else if cmds[0] != "less" || cmds[1] != "cat" {
		t.Errorf("text/plain commands = %v, want [less cat]", cmds)
	}

	// Check image/* rule
	if cmds, ok := result.Rules["image/*"]; !ok {
		t.Error("expected image/* rule")
	} else if len(cmds) != 1 || cmds[0] != "feh" {
		t.Errorf("image/* commands = %v, want [feh]", cmds)
	}
}

func TestParseEnterBehavior_MIME(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedType  EnterBehaviorType
		expectWarning bool
	}{
		{
			name:          "mime: value",
			input:         "mime:",
			expectedType:  EnterBehaviorMIME,
			expectWarning: false,
		},
		{
			name:          "mime: with whitespace",
			input:         "  mime:  ",
			expectedType:  EnterBehaviorMIME,
			expectWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, warning := ParseEnterBehavior(tt.input)

			if result.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", result.Type, tt.expectedType)
			}

			if tt.expectWarning && warning == "" {
				t.Error("expected warning but got none")
			}
			if !tt.expectWarning && warning != "" {
				t.Errorf("unexpected warning: %q", warning)
			}
		})
	}
}

func TestEnterBehaviorString_MIME(t *testing.T) {
	behavior := EnterBehavior{Type: EnterBehaviorMIME}
	expected := "mime:"

	result := behavior.String()
	if result != expected {
		t.Errorf("String() = %q, want %q", result, expected)
	}
}

func TestMIMEBehaviorConfig_Empty(t *testing.T) {
	cfg := MIMEBehaviorConfig{}

	if cfg.Rules != nil {
		t.Errorf("Empty config should have nil Rules, got %v", cfg.Rules)
	}
}

func TestGetMIMEType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "text plain",
			filename: "file.txt",
			expected: "text/plain",
		},
		{
			name:     "image png",
			filename: "image.png",
			expected: "image/png",
		},
		{
			name:     "image jpeg",
			filename: "photo.jpg",
			expected: "image/jpeg",
		},
		{
			name:     "html file",
			filename: "index.html",
			expected: "text/html",
		},
		{
			name:     "json file",
			filename: "data.json",
			expected: "application/json",
		},
		{
			name:     "unknown extension",
			filename: "file.xyz123",
			expected: "application/octet-stream",
		},
		{
			name:     "no extension",
			filename: "Makefile",
			expected: "application/octet-stream",
		},
		{
			name:     "hidden file no extension",
			filename: ".gitignore",
			expected: "application/octet-stream",
		},
		{
			name:     "hidden file with extension",
			filename: ".config.json",
			expected: "application/json",
		},
		{
			name:     "uppercase extension",
			filename: "FILE.TXT",
			expected: "text/plain",
		},
		{
			name:     "mixed case extension",
			filename: "file.Json",
			expected: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMIMEType(tt.filename)

			// MIME types may have parameters like "; charset=utf-8"
			// We only care about the base type
			if !strings.HasPrefix(result, tt.expected) && result != tt.expected {
				t.Errorf("GetMIMEType(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestMatchesMIMEPattern(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		pattern  string
		expected bool
	}{
		{
			name:     "exact match",
			mimeType: "text/plain",
			pattern:  "text/plain",
			expected: true,
		},
		{
			name:     "exact match no match",
			mimeType: "text/plain",
			pattern:  "text/html",
			expected: false,
		},
		{
			name:     "wildcard match",
			mimeType: "image/png",
			pattern:  "image/*",
			expected: true,
		},
		{
			name:     "wildcard match jpeg",
			mimeType: "image/jpeg",
			pattern:  "image/*",
			expected: true,
		},
		{
			name:     "wildcard no match",
			mimeType: "text/plain",
			pattern:  "image/*",
			expected: false,
		},
		{
			name:     "text wildcard",
			mimeType: "text/html",
			pattern:  "text/*",
			expected: true,
		},
		{
			name:     "application wildcard",
			mimeType: "application/json",
			pattern:  "application/*",
			expected: true,
		},
		{
			name:     "empty mime type",
			mimeType: "",
			pattern:  "text/*",
			expected: false,
		},
		{
			name:     "empty pattern",
			mimeType: "text/plain",
			pattern:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchesMIMEPattern(tt.mimeType, tt.pattern)
			if result != tt.expected {
				t.Errorf("MatchesMIMEPattern(%q, %q) = %v, want %v",
					tt.mimeType, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestFindMatchingRule(t *testing.T) {
	cfg := MIMEBehaviorConfig{
		Rules: map[string][]string{
			"text/plain":       {"less", "cat"},
			"text/*":           {"bat"},
			"image/*":          {"feh", "eog"},
			"application/pdf":  {"zathura"},
			"application/json": {"jq"},
		},
	}

	tests := []struct {
		name           string
		mimeType       string
		expectFound    bool
		expectedCmds   []string
		expectWildcard bool
	}{
		{
			name:         "exact match text/plain",
			mimeType:     "text/plain",
			expectFound:  true,
			expectedCmds: []string{"less", "cat"},
		},
		{
			name:         "exact match application/pdf",
			mimeType:     "application/pdf",
			expectFound:  true,
			expectedCmds: []string{"zathura"},
		},
		{
			name:           "wildcard match text/html (exact text/plain exists)",
			mimeType:       "text/html",
			expectFound:    true,
			expectedCmds:   []string{"bat"},
			expectWildcard: true,
		},
		{
			name:           "wildcard match image/png",
			mimeType:       "image/png",
			expectFound:    true,
			expectedCmds:   []string{"feh", "eog"},
			expectWildcard: true,
		},
		{
			name:        "no match for audio",
			mimeType:    "audio/mp3",
			expectFound: false,
		},
		{
			name:        "no match for video",
			mimeType:    "video/mp4",
			expectFound: false,
		},
		{
			name:        "empty mime type",
			mimeType:    "",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmds, found := cfg.FindMatchingRule(tt.mimeType)

			if found != tt.expectFound {
				t.Errorf("FindMatchingRule(%q) found = %v, want %v",
					tt.mimeType, found, tt.expectFound)
			}

			if tt.expectFound {
				if len(cmds) != len(tt.expectedCmds) {
					t.Errorf("FindMatchingRule(%q) cmds = %v, want %v",
						tt.mimeType, cmds, tt.expectedCmds)
				} else {
					for i, cmd := range cmds {
						if cmd != tt.expectedCmds[i] {
							t.Errorf("FindMatchingRule(%q) cmds[%d] = %q, want %q",
								tt.mimeType, i, cmd, tt.expectedCmds[i])
						}
					}
				}
			}
		})
	}
}

func TestFindMatchingRule_ExactPriority(t *testing.T) {
	// Test that exact match takes priority over wildcard
	cfg := MIMEBehaviorConfig{
		Rules: map[string][]string{
			"text/plain": {"exact-viewer"},
			"text/*":     {"wildcard-viewer"},
		},
	}

	cmds, found := cfg.FindMatchingRule("text/plain")
	if !found {
		t.Fatal("expected to find rule for text/plain")
	}

	if len(cmds) != 1 || cmds[0] != "exact-viewer" {
		t.Errorf("expected exact match to take priority, got %v", cmds)
	}
}

func TestFindMatchingRule_EmptyConfig(t *testing.T) {
	cfg := MIMEBehaviorConfig{}

	cmds, found := cfg.FindMatchingRule("text/plain")
	if found {
		t.Errorf("expected no match for empty config, got %v", cmds)
	}
}

func TestFindMatchingRule_NilRules(t *testing.T) {
	cfg := MIMEBehaviorConfig{Rules: nil}

	cmds, found := cfg.FindMatchingRule("text/plain")
	if found {
		t.Errorf("expected no match for nil rules, got %v", cmds)
	}
}

func TestParseMIMEBehavior_Fallback(t *testing.T) {
	tests := []struct {
		name             string
		input            map[string][]string
		expectedRules    int
		expectedFallback []string
		expectWarnings   bool
		warningContains  string
	}{
		{
			name: "fallback extracted from rules",
			input: map[string][]string{
				"text/*":   {"less"},
				"fallback": {"xdg-open"},
			},
			expectedRules:    1,
			expectedFallback: []string{"xdg-open"},
			expectWarnings:   false,
		},
		{
			name: "fallback not in rules map",
			input: map[string][]string{
				"fallback": {"xdg-open"},
			},
			expectedRules:    0,
			expectedFallback: []string{"xdg-open"},
			expectWarnings:   false,
		},
		{
			name: "empty fallback array generates warning",
			input: map[string][]string{
				"fallback": {},
			},
			expectedRules:    0,
			expectedFallback: nil,
			expectWarnings:   true,
			warningContains:  "empty command list for fallback",
		},
		{
			name: "missing fallback results in nil",
			input: map[string][]string{
				"text/*": {"less"},
			},
			expectedRules:    1,
			expectedFallback: nil,
			expectWarnings:   false,
		},
		{
			name: "fallback only no MIME rules",
			input: map[string][]string{
				"fallback": {"xdg-open"},
			},
			expectedRules:    0,
			expectedFallback: []string{"xdg-open"},
			expectWarnings:   false,
		},
		{
			name: "multiple fallback commands",
			input: map[string][]string{
				"fallback": {"xdg-open", "open"},
			},
			expectedRules:    0,
			expectedFallback: []string{"xdg-open", "open"},
			expectWarnings:   false,
		},
		{
			name: "unknown key generates warning",
			input: map[string][]string{
				"text":     {"less"},
				"fallback": {"xdg-open"},
			},
			expectedRules:    0,
			expectedFallback: []string{"xdg-open"},
			expectWarnings:   true,
			warningContains:  "unknown key",
		},
		{
			name: "fallback with MIME rules and unknown key",
			input: map[string][]string{
				"text/*":   {"less"},
				"fallback": {"xdg-open"},
				"badkey":   {"vim"},
			},
			expectedRules:    1,
			expectedFallback: []string{"xdg-open"},
			expectWarnings:   true,
			warningContains:  "unknown key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, warnings := ParseMIMEBehavior(tt.input)

			// Check rules count
			if len(result.Rules) != tt.expectedRules {
				t.Errorf("Rules count = %d, want %d", len(result.Rules), tt.expectedRules)
			}

			// Verify fallback is NOT in Rules map
			if _, ok := result.Rules["fallback"]; ok {
				t.Error("fallback key should not be in Rules map")
			}

			// Check Fallback field
			if tt.expectedFallback == nil {
				if result.Fallback != nil {
					t.Errorf("Fallback = %v, want nil", result.Fallback)
				}
			} else {
				if len(result.Fallback) != len(tt.expectedFallback) {
					t.Errorf("Fallback length = %d, want %d", len(result.Fallback), len(tt.expectedFallback))
				} else {
					for i, cmd := range result.Fallback {
						if cmd != tt.expectedFallback[i] {
							t.Errorf("Fallback[%d] = %q, want %q", i, cmd, tt.expectedFallback[i])
						}
					}
				}
			}

			// Check warnings
			if tt.expectWarnings {
				if len(warnings) == 0 {
					t.Error("expected warnings but got none")
				} else if tt.warningContains != "" {
					found := false
					for _, w := range warnings {
						if strings.Contains(w, tt.warningContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("warnings %v should contain %q", warnings, tt.warningContains)
					}
				}
			} else {
				if len(warnings) > 0 {
					t.Errorf("unexpected warnings: %v", warnings)
				}
			}
		})
	}
}

func TestParseMIMEBehavior_FallbackContent(t *testing.T) {
	input := map[string][]string{
		"text/plain": {"less", "cat"},
		"image/*":    {"feh"},
		"fallback":   {"xdg-open", "open"},
	}

	result, warnings := ParseMIMEBehavior(input)

	if len(warnings) > 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	// Check that MIME rules are preserved
	if cmds, ok := result.Rules["text/plain"]; !ok {
		t.Error("expected text/plain rule")
	} else if len(cmds) != 2 || cmds[0] != "less" || cmds[1] != "cat" {
		t.Errorf("text/plain commands = %v, want [less cat]", cmds)
	}

	if cmds, ok := result.Rules["image/*"]; !ok {
		t.Error("expected image/* rule")
	} else if len(cmds) != 1 || cmds[0] != "feh" {
		t.Errorf("image/* commands = %v, want [feh]", cmds)
	}

	// Check that fallback is NOT in Rules
	if _, ok := result.Rules["fallback"]; ok {
		t.Error("fallback should not be in Rules")
	}

	// Check Fallback field
	if len(result.Fallback) != 2 {
		t.Fatalf("Fallback length = %d, want 2", len(result.Fallback))
	}
	if result.Fallback[0] != "xdg-open" || result.Fallback[1] != "open" {
		t.Errorf("Fallback = %v, want [xdg-open open]", result.Fallback)
	}
}
