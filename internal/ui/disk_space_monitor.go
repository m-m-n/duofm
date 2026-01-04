package ui

import (
	"time"

	"github.com/sakura/duofm/internal/fs"
)

// DiskSpaceMonitor monitors disk space for both panes.
// It caches disk space values and refreshes them periodically.
type DiskSpaceMonitor struct {
	leftSpace      uint64    // Free space in left pane's filesystem
	rightSpace     uint64    // Free space in right pane's filesystem
	lastCheckTime  time.Time // Time of last disk space check
	checkInterval  time.Duration
}

// NewDiskSpaceMonitor creates a new disk space monitor with default settings.
func NewDiskSpaceMonitor() *DiskSpaceMonitor {
	return &DiskSpaceMonitor{
		checkInterval: 5 * time.Second,
	}
}

// LeftSpace returns the cached free space for the left pane.
func (m *DiskSpaceMonitor) LeftSpace() uint64 {
	return m.leftSpace
}

// RightSpace returns the cached free space for the right pane.
func (m *DiskSpaceMonitor) RightSpace() uint64 {
	return m.rightSpace
}

// LastCheckTime returns the time of the last disk space check.
func (m *DiskSpaceMonitor) LastCheckTime() time.Time {
	return m.lastCheckTime
}

// NeedsRefresh returns true if the disk space cache should be refreshed.
func (m *DiskSpaceMonitor) NeedsRefresh() bool {
	return time.Since(m.lastCheckTime) > m.checkInterval
}

// Update refreshes disk space for both paths.
func (m *DiskSpaceMonitor) Update(leftPath, rightPath string) {
	if leftPath != "" {
		if freeBytes, _, err := fs.GetDiskSpace(leftPath); err == nil {
			m.leftSpace = freeBytes
		}
	}

	if rightPath != "" {
		if freeBytes, _, err := fs.GetDiskSpace(rightPath); err == nil {
			m.rightSpace = freeBytes
		}
	}

	m.lastCheckTime = time.Now()
}

// UpdateLeft refreshes disk space for the left pane only.
func (m *DiskSpaceMonitor) UpdateLeft(path string) {
	if path != "" {
		if freeBytes, _, err := fs.GetDiskSpace(path); err == nil {
			m.leftSpace = freeBytes
		}
	}
}

// UpdateRight refreshes disk space for the right pane only.
func (m *DiskSpaceMonitor) UpdateRight(path string) {
	if path != "" {
		if freeBytes, _, err := fs.GetDiskSpace(path); err == nil {
			m.rightSpace = freeBytes
		}
	}
}
