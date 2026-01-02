package ui

import "path/filepath"

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

// GetMarkedFiles returns list of marked filenames
func (p *Pane) GetMarkedFiles() []string {
	result := make([]string, 0, len(p.markedFiles))
	for name := range p.markedFiles {
		result = append(result, name)
	}
	return result
}

// GetMarkedFilePaths returns list of full paths for marked files
func (p *Pane) GetMarkedFilePaths() []string {
	result := make([]string, 0, len(p.markedFiles))
	for name := range p.markedFiles {
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
