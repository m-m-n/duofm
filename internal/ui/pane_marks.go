package ui

import (
	"path/filepath"
	"sort"
)

// ToggleMark toggles the mark on the currently selected file
// Returns false if the current entry is a parent directory
func (p *Pane) ToggleMark() bool {
	entry := p.SelectedEntry()
	if entry == nil || entry.IsParentDir() {
		return false
	}

	if p.markedFiles[entry.Name] {
		delete(p.markedFiles, entry.Name)
	} else {
		p.markedFiles[entry.Name] = true
	}
	return true
}

// ClearMarks removes all marks
func (p *Pane) ClearMarks() {
	p.markedFiles = make(map[string]bool)
}

// IsMarked returns whether a file is marked
func (p *Pane) IsMarked(filename string) bool {
	return p.markedFiles[filename]
}

// GetMarkedFiles returns the marked filenames in the pane's display order.
// Marked names that are not present in the current display list (e.g.
// hidden by an active filter) are appended afterward, sorted ascending by
// name. The result always has exactly one entry per marked name.
func (p *Pane) GetMarkedFiles() []string {
	result := make([]string, 0, len(p.markedFiles))
	emitted := make(map[string]bool, len(p.markedFiles))

	for _, entry := range p.entries {
		if p.markedFiles[entry.Name] && !emitted[entry.Name] {
			result = append(result, entry.Name)
			emitted[entry.Name] = true
		}
	}

	remaining := make([]string, 0, len(p.markedFiles)-len(emitted))
	for name := range p.markedFiles {
		if !emitted[name] {
			remaining = append(remaining, name)
		}
	}
	sort.Strings(remaining)

	return append(result, remaining...)
}

// GetMarkedFilePaths returns the full paths of the marked files. The order
// and length always match GetMarkedFiles, since the paths are derived from
// it rather than computed independently.
func (p *Pane) GetMarkedFilePaths() []string {
	names := p.GetMarkedFiles()
	result := make([]string, 0, len(names))
	for _, name := range names {
		result = append(result, filepath.Join(p.path, name))
	}
	return result
}

// CalculateMarkInfo returns mark statistics
func (p *Pane) CalculateMarkInfo() MarkInfo {
	info := MarkInfo{}
	for name := range p.markedFiles {
		info.Count++
		// Find the entry to get size
		for _, entry := range p.allEntries {
			if entry.Name == name && !entry.IsDir {
				info.TotalSize += entry.Size
				break
			}
		}
	}
	return info
}

// MarkCount returns the number of marked files
func (p *Pane) MarkCount() int {
	return len(p.markedFiles)
}

// HasMarkedFiles returns whether there are any marked files
func (p *Pane) HasMarkedFiles() bool {
	return len(p.markedFiles) > 0
}
