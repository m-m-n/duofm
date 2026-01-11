package ui

import (
	"testing"
	"time"
)

func TestNewDiskSpaceMonitor(t *testing.T) {
	m := NewDiskSpaceMonitor()

	if m == nil {
		t.Fatal("NewDiskSpaceMonitor() returned nil")
	}

	if m.checkInterval != 5*time.Second {
		t.Errorf("checkInterval = %v, want 5s", m.checkInterval)
	}
}

func TestDiskSpaceMonitor_LeftSpace(t *testing.T) {
	m := NewDiskSpaceMonitor()
	m.leftSpace = 1024 * 1024 * 1024 // 1GB

	if got := m.LeftSpace(); got != 1024*1024*1024 {
		t.Errorf("LeftSpace() = %d, want %d", got, 1024*1024*1024)
	}
}

func TestDiskSpaceMonitor_RightSpace(t *testing.T) {
	m := NewDiskSpaceMonitor()
	m.rightSpace = 2048 * 1024 * 1024 // 2GB

	if got := m.RightSpace(); got != 2048*1024*1024 {
		t.Errorf("RightSpace() = %d, want %d", got, 2048*1024*1024)
	}
}

func TestDiskSpaceMonitor_LastCheckTime(t *testing.T) {
	m := NewDiskSpaceMonitor()
	now := time.Now()
	m.lastCheckTime = now

	if got := m.LastCheckTime(); !got.Equal(now) {
		t.Errorf("LastCheckTime() = %v, want %v", got, now)
	}
}

func TestDiskSpaceMonitor_NeedsRefresh_Initial(t *testing.T) {
	m := NewDiskSpaceMonitor()

	// Initially, lastCheckTime is zero, so it should need refresh
	if !m.NeedsRefresh() {
		t.Error("NeedsRefresh() = false, want true for initial state")
	}
}

func TestDiskSpaceMonitor_NeedsRefresh_Recent(t *testing.T) {
	m := NewDiskSpaceMonitor()
	m.lastCheckTime = time.Now()

	// Just checked, should not need refresh
	if m.NeedsRefresh() {
		t.Error("NeedsRefresh() = true, want false for recent check")
	}
}

func TestDiskSpaceMonitor_NeedsRefresh_Old(t *testing.T) {
	m := NewDiskSpaceMonitor()
	m.lastCheckTime = time.Now().Add(-10 * time.Second)

	// Check was 10 seconds ago, interval is 5 seconds
	if !m.NeedsRefresh() {
		t.Error("NeedsRefresh() = false, want true for old check")
	}
}

func TestDiskSpaceMonitor_Update(t *testing.T) {
	m := NewDiskSpaceMonitor()
	tmpDir := t.TempDir()

	before := m.lastCheckTime
	m.Update(tmpDir, tmpDir)

	// lastCheckTime should be updated
	if !m.lastCheckTime.After(before) {
		t.Error("Update() did not update lastCheckTime")
	}

	// Space values should be set (non-zero for valid paths)
	if m.leftSpace == 0 {
		t.Error("Update() did not set leftSpace")
	}
	if m.rightSpace == 0 {
		t.Error("Update() did not set rightSpace")
	}
}

func TestDiskSpaceMonitor_Update_EmptyPaths(t *testing.T) {
	m := NewDiskSpaceMonitor()
	m.leftSpace = 100
	m.rightSpace = 200

	m.Update("", "")

	// Values should remain unchanged for empty paths
	if m.leftSpace != 100 {
		t.Errorf("Update(\"\", \"\") changed leftSpace to %d", m.leftSpace)
	}
	if m.rightSpace != 200 {
		t.Errorf("Update(\"\", \"\") changed rightSpace to %d", m.rightSpace)
	}
}

func TestDiskSpaceMonitor_UpdateLeft(t *testing.T) {
	m := NewDiskSpaceMonitor()
	tmpDir := t.TempDir()

	m.UpdateLeft(tmpDir)

	if m.leftSpace == 0 {
		t.Error("UpdateLeft() did not set leftSpace")
	}
}

func TestDiskSpaceMonitor_UpdateLeft_EmptyPath(t *testing.T) {
	m := NewDiskSpaceMonitor()
	m.leftSpace = 100

	m.UpdateLeft("")

	if m.leftSpace != 100 {
		t.Errorf("UpdateLeft(\"\") changed leftSpace to %d", m.leftSpace)
	}
}

func TestDiskSpaceMonitor_UpdateRight(t *testing.T) {
	m := NewDiskSpaceMonitor()
	tmpDir := t.TempDir()

	m.UpdateRight(tmpDir)

	if m.rightSpace == 0 {
		t.Error("UpdateRight() did not set rightSpace")
	}
}

func TestDiskSpaceMonitor_UpdateRight_EmptyPath(t *testing.T) {
	m := NewDiskSpaceMonitor()
	m.rightSpace = 200

	m.UpdateRight("")

	if m.rightSpace != 200 {
		t.Errorf("UpdateRight(\"\") changed rightSpace to %d", m.rightSpace)
	}
}
