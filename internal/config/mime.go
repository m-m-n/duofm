package config

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"
)

// MIMEBehaviorConfig holds MIME type to command mappings.
type MIMEBehaviorConfig struct {
	// Rules maps MIME type patterns to command lists.
	// Key: MIME type pattern (e.g., "text/plain", "image/*")
	// Value: List of commands to try in order
	Rules map[string][]string
}

// ParseMIMEBehavior parses the [enter_behavior_mime] section from TOML.
// It validates each entry and generates warnings for invalid configurations.
//
// Invalid entries are skipped:
//   - Empty MIME type key
//   - Empty command array
//
// Returns the parsed config and any warning messages.
func ParseMIMEBehavior(raw map[string][]string) (MIMEBehaviorConfig, []string) {
	var warnings []string
	config := MIMEBehaviorConfig{
		Rules: make(map[string][]string),
	}

	if raw == nil {
		return config, warnings
	}

	for mimeType, commands := range raw {
		// Validate MIME type key
		if mimeType == "" {
			warnings = append(warnings, "empty MIME type key in enter_behavior_mime, skipping")
			continue
		}

		// Validate command array
		if len(commands) == 0 {
			warnings = append(warnings, fmt.Sprintf("empty command list for MIME type '%s', skipping", mimeType))
			continue
		}

		// Store valid entry
		config.Rules[mimeType] = commands
	}

	return config, warnings
}

// GetMIMEType returns the MIME type for the given filename based on its extension.
// If the extension is unknown or the file has no extension, returns "application/octet-stream".
// Any MIME type parameters (e.g., "; charset=utf-8") are stripped from the result.
func GetMIMEType(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "application/octet-stream"
	}

	// mime.TypeByExtension is case-insensitive in Go
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "application/octet-stream"
	}

	// Strip parameters (e.g., "; charset=utf-8")
	if idx := strings.Index(mimeType, ";"); idx != -1 {
		mimeType = strings.TrimSpace(mimeType[:idx])
	}

	return mimeType
}

// MatchesMIMEPattern checks if the given MIME type matches the pattern.
// Supports exact match (e.g., "text/plain") and wildcard match (e.g., "text/*").
func MatchesMIMEPattern(mimeType, pattern string) bool {
	if mimeType == "" || pattern == "" {
		return false
	}

	// Exact match
	if mimeType == pattern {
		return true
	}

	// Wildcard match (e.g., "text/*")
	if strings.HasSuffix(pattern, "/*") {
		typePrefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(mimeType, typePrefix+"/")
	}

	return false
}

// FindMatchingRule finds the commands for the given MIME type.
// Priority: exact match > wildcard match.
// Returns the commands and whether a match was found.
func (c *MIMEBehaviorConfig) FindMatchingRule(mimeType string) ([]string, bool) {
	if c.Rules == nil || mimeType == "" {
		return nil, false
	}

	// Priority 1: Exact match
	if cmds, ok := c.Rules[mimeType]; ok {
		return cmds, true
	}

	// Priority 2: Wildcard match
	// Extract type prefix (e.g., "text" from "text/plain")
	parts := strings.SplitN(mimeType, "/", 2)
	if len(parts) != 2 {
		return nil, false
	}

	wildcardPattern := parts[0] + "/*"
	if cmds, ok := c.Rules[wildcardPattern]; ok {
		return cmds, true
	}

	return nil, false
}
