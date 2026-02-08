package config

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// mergeResult holds the missing configuration items to be merged.
type mergeResult struct {
	Keybindings         map[string][]string
	Colors              map[string]int
	HistoryLimit        *int    // nil means not missing
	RefreshRate         *int    // nil means not missing
	EnterBehavior       *string // nil means not missing
	EnterBehaviorMIME   bool    // true means section is missing
	MIMEFallbackMissing bool    // true means section exists but fallback key is absent
}

// hasContent returns true if there are any missing items to merge.
func (m mergeResult) hasContent() bool {
	return len(m.Keybindings) > 0 || len(m.Colors) > 0 || m.HistoryLimit != nil || m.RefreshRate != nil || m.EnterBehavior != nil || m.EnterBehaviorMIME || m.MIMEFallbackMissing
}

// FindMissingKeybindings returns keybindings that exist in defaults but not in config.
// It compares the existing keybindings with DefaultKeybindings() and returns
// only the keys that are missing from the existing map.
func FindMissingKeybindings(existing map[string][]string) map[string][]string {
	defaults := DefaultKeybindings()
	missing := make(map[string][]string)

	for key, value := range defaults {
		if existing == nil {
			missing[key] = value
			continue
		}
		if _, exists := existing[key]; !exists {
			missing[key] = value
		}
	}

	return missing
}

// FindMissingColors returns color settings that exist in defaults but not in config.
// Uses AllColorKeys() to enumerate all color keys and GetDefaultColorValue() to get default values.
func FindMissingColors(existing map[string]interface{}) map[string]int {
	colorKeys := AllColorKeys()
	missing := make(map[string]int)

	for _, key := range colorKeys {
		if existing == nil {
			missing[key] = GetDefaultColorValue(key)
			continue
		}
		if _, exists := existing[key]; !exists {
			missing[key] = GetDefaultColorValue(key)
		}
	}

	return missing
}

// IsMissingHistoryLimit returns true if history_limit is not set in config.
// A nil pointer indicates the value was not set in the config file.
func IsMissingHistoryLimit(historyLimit *int) bool {
	return historyLimit == nil
}

// IsMissingRefreshRate returns true if refresh_rate is not set in config.
// A nil pointer indicates the value was not set in the config file.
func IsMissingRefreshRate(refreshRate *int) bool {
	return refreshRate == nil
}

// IsMissingEnterBehavior returns true if enter_behavior is not set in config.
// A nil pointer indicates the value was not set in the config file.
func IsMissingEnterBehavior(enterBehavior *string) bool {
	return enterBehavior == nil
}

// IsMissingEnterBehaviorMIME returns true if [enter_behavior_mime] section is not present in config.
// A nil map indicates the section was not defined in the config file.
func IsMissingEnterBehaviorMIME(enterBehaviorMIME map[string][]string) bool {
	return enterBehaviorMIME == nil
}

// formatKeybinding formats a keybinding entry as TOML.
// Example: move_down = ["J", "Down"]
func formatKeybinding(key string, values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	return fmt.Sprintf("%s = [%s]", key, strings.Join(quoted, ", "))
}

// MergeConfig merges missing configuration items into the existing config file.
// It reads the existing file, appends missing items with their default values,
// and writes the complete content back to the file.
// Returns nil if no items were missing or if the merge was successful.
// Returns an error if the file could not be read or written.
func MergeConfig(path string, existing *rawConfig) error {
	// Collect missing items
	result := mergeResult{
		Keybindings: FindMissingKeybindings(existing.Keybindings),
		Colors:      FindMissingColors(existing.Colors),
	}

	// Check if history_limit is missing
	if IsMissingHistoryLimit(existing.HistoryLimit) {
		defaultLimit := DefaultHistoryLimit
		result.HistoryLimit = &defaultLimit
	}

	// Check if refresh_rate is missing
	if IsMissingRefreshRate(existing.RefreshRate) {
		defaultRate := DefaultRefreshRate
		result.RefreshRate = &defaultRate
	}

	// Check if enter_behavior is missing
	if IsMissingEnterBehavior(existing.EnterBehavior) {
		defaultEnterBehavior := DefaultEnterBehavior().String()
		result.EnterBehavior = &defaultEnterBehavior
	}

	// Get file info to preserve permissions
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}
	fileMode := fileInfo.Mode()

	// Read existing file content
	existingContent, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Check if [enter_behavior_mime] section is missing
	if IsMissingEnterBehaviorMIME(existing.EnterBehaviorMIME) {
		// Section missing (regardless of commented placeholder) - add active section with fallback
		result.EnterBehaviorMIME = true
	} else {
		// Section exists - check if fallback key is present
		if _, hasFallback := existing.EnterBehaviorMIME["fallback"]; !hasFallback {
			result.MIMEFallbackMissing = true
		}
	}

	// Generate the merged content
	mergedContent := generateMergedFile(string(existingContent), result)

	// Write the merged content back to the file, preserving original permissions
	if err := os.WriteFile(path, []byte(mergedContent), fileMode); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// generateMergedFile generates the complete merged file content.
// It preserves the original content and appends missing items in their appropriate sections.
// history_limit is inserted at the beginning (before any section) since it's a root-level key.
func generateMergedFile(original string, result mergeResult) string {
	// Parse the original content to find section positions
	lines := strings.Split(original, "\n")

	// Track which sections exist and their line ranges (start and end index)
	type sectionInfo struct {
		start int
		end   int
	}
	var keybindingsSection *sectionInfo
	var colorsSection *sectionInfo
	var enterBehaviorMIMESection *sectionInfo
	currentSection := ""
	firstSectionLine := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			// Track first section position for inserting root-level keys
			if firstSectionLine == -1 {
				firstSectionLine = i
			}

			// New section found - finalize previous section
			if currentSection == "keybindings" && keybindingsSection != nil {
				keybindingsSection.end = i - 1
			} else if currentSection == "colors" && colorsSection != nil {
				colorsSection.end = i - 1
			} else if currentSection == "enter_behavior_mime" && enterBehaviorMIMESection != nil {
				enterBehaviorMIMESection.end = i - 1
			}

			// Start new section
			sectionName := trimmed[1 : len(trimmed)-1]
			currentSection = sectionName

			if sectionName == "keybindings" {
				keybindingsSection = &sectionInfo{start: i, end: -1}
			} else if sectionName == "colors" {
				colorsSection = &sectionInfo{start: i, end: -1}
			} else if sectionName == "enter_behavior_mime" {
				enterBehaviorMIMESection = &sectionInfo{start: i, end: -1}
			}
		}
	}

	// Finalize last section
	lastLineIdx := len(lines) - 1
	// Handle trailing empty lines
	for lastLineIdx > 0 && strings.TrimSpace(lines[lastLineIdx]) == "" {
		lastLineIdx--
	}

	if currentSection == "keybindings" && keybindingsSection != nil && keybindingsSection.end == -1 {
		keybindingsSection.end = lastLineIdx
	} else if currentSection == "colors" && colorsSection != nil && colorsSection.end == -1 {
		colorsSection.end = lastLineIdx
	} else if currentSection == "enter_behavior_mime" && enterBehaviorMIMESection != nil && enterBehaviorMIMESection.end == -1 {
		enterBehaviorMIMESection.end = lastLineIdx
	}

	// Build output
	var sb strings.Builder
	insertedHistoryLimit := false
	insertedRefreshRate := false
	insertedEnterBehavior := false

	for i, line := range lines {
		// Insert root-level keys before the first section
		if i == firstSectionLine {
			if result.HistoryLimit != nil && !insertedHistoryLimit {
				sb.WriteString(fmt.Sprintf("history_limit = %d\n", *result.HistoryLimit))
				insertedHistoryLimit = true
			}
			if result.RefreshRate != nil && !insertedRefreshRate {
				sb.WriteString(fmt.Sprintf("refresh_rate = %d\n", *result.RefreshRate))
				insertedRefreshRate = true
			}
			if result.EnterBehavior != nil && !insertedEnterBehavior {
				sb.WriteString(fmt.Sprintf("enter_behavior = %q\n", *result.EnterBehavior))
				insertedEnterBehavior = true
			}
			// Add blank line after root-level keys if we inserted any
			if insertedHistoryLimit || insertedRefreshRate || insertedEnterBehavior {
				sb.WriteString("\n")
			}
		}

		sb.WriteString(line)
		if i < len(lines)-1 {
			sb.WriteString("\n")
		}

		// Insert missing keybindings at the end of keybindings section
		if keybindingsSection != nil && i == keybindingsSection.end && len(result.Keybindings) > 0 {
			sb.WriteString("\n")
			sb.WriteString(generateKeybindingsEntries(result.Keybindings))
		}

		// Insert missing colors at the end of colors section
		if colorsSection != nil && i == colorsSection.end && len(result.Colors) > 0 {
			sb.WriteString("\n")
			sb.WriteString(generateColorsEntries(result.Colors))
		}

		// Insert missing fallback at the end of enter_behavior_mime section
		if enterBehaviorMIMESection != nil && i == enterBehaviorMIMESection.end && result.MIMEFallbackMissing {
			sb.WriteString("\nfallback = [\"xdg-open\"]\n")
		}
	}

	// Ensure content ends with newline
	content := sb.String()
	if len(content) > 0 && content[len(content)-1] != '\n' {
		content += "\n"
	}

	// Append sections/items that weren't inserted into existing sections
	var appendContent strings.Builder

	// Track if we need the separator comment
	needsSeparator := false
	if keybindingsSection == nil && len(result.Keybindings) > 0 {
		needsSeparator = true
	}
	if colorsSection == nil && len(result.Colors) > 0 {
		needsSeparator = true
	}
	// Root-level keys at end only if there are no sections (file is empty or has no sections)
	if result.HistoryLimit != nil && !insertedHistoryLimit {
		needsSeparator = true
	}
	if result.RefreshRate != nil && !insertedRefreshRate {
		needsSeparator = true
	}
	if result.EnterBehavior != nil && !insertedEnterBehavior {
		needsSeparator = true
	}
	if result.EnterBehaviorMIME {
		needsSeparator = true
	}

	if needsSeparator {
		appendContent.WriteString("\n# --- Auto-merged settings (added by duofm) ---\n")
	}

	// Add root-level keys if missing and weren't inserted yet (no sections exist)
	if result.HistoryLimit != nil && !insertedHistoryLimit {
		appendContent.WriteString(fmt.Sprintf("\nhistory_limit = %d\n", *result.HistoryLimit))
	}
	if result.RefreshRate != nil && !insertedRefreshRate {
		appendContent.WriteString(fmt.Sprintf("\nrefresh_rate = %d\n", *result.RefreshRate))
	}
	if result.EnterBehavior != nil && !insertedEnterBehavior {
		appendContent.WriteString(fmt.Sprintf("\nenter_behavior = %q\n", *result.EnterBehavior))
	}

	// Add keybindings section if it doesn't exist
	if keybindingsSection == nil && len(result.Keybindings) > 0 {
		appendContent.WriteString("\n[keybindings]\n")
		appendContent.WriteString(generateKeybindingsEntries(result.Keybindings))
	}

	// Add colors section if it doesn't exist
	if colorsSection == nil && len(result.Colors) > 0 {
		appendContent.WriteString("\n[colors]\n")
		appendContent.WriteString(generateColorsEntries(result.Colors))
	}

	// Add [enter_behavior_mime] section with fallback if it doesn't exist
	if result.EnterBehaviorMIME {
		appendContent.WriteString("\n# MIME type based file opening (used when enter_behavior = \"mime:\")\n")
		appendContent.WriteString("# Format: \"MIME/type\" = [\"command1\", \"command2\", ...]\n")
		appendContent.WriteString("# Wildcard patterns supported: \"image/*\", \"text/*\", etc.\n")
		appendContent.WriteString("# Commands are tried in order until one is found in PATH.\n")
		appendContent.WriteString("# If no MIME rule matches, fallback commands are tried.\n")
		appendContent.WriteString("# If all commands fail, falls back to pager.\n")
		appendContent.WriteString("[enter_behavior_mime]\n")
		appendContent.WriteString("# \"text/plain\" = [\"bat\", \"less\"]\n")
		appendContent.WriteString("# \"text/*\" = [\"less\"]\n")
		appendContent.WriteString("# \"image/*\" = [\"feh\", \"eog\", \"xdg-open\"]\n")
		appendContent.WriteString("# \"application/pdf\" = [\"zathura\", \"evince\"]\n")
		appendContent.WriteString("fallback = [\"xdg-open\"]\n")
	}

	return content + appendContent.String()
}

// generateKeybindingsEntries generates TOML entries for keybindings.
func generateKeybindingsEntries(keybindings map[string][]string) string {
	var sb strings.Builder
	keys := make([]string, 0, len(keybindings))
	for k := range keybindings {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		values := keybindings[key]
		sb.WriteString(formatKeybinding(key, values))
		sb.WriteString("\n")
	}
	return sb.String()
}

// generateColorsEntries generates TOML entries for colors.
func generateColorsEntries(colors map[string]int) string {
	var sb strings.Builder
	keys := make([]string, 0, len(colors))
	for k := range colors {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := colors[key]
		sb.WriteString(fmt.Sprintf("%s = %d\n", key, value))
	}
	return sb.String()
}
