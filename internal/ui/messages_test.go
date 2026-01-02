package ui

import (
	"testing"
	"time"
)

func TestDiskSpaceTickCmd(t *testing.T) {
	cmd := diskSpaceTickCmd()
	if cmd == nil {
		t.Error("diskSpaceTickCmd() should return a command")
	}
}

func TestStatusMessageClearCmd_Duration(t *testing.T) {
	cmd := statusMessageClearCmd(100 * time.Millisecond)
	if cmd == nil {
		t.Error("statusMessageClearCmd() should return a command")
	}
}

func TestCtrlCTimeoutCmd(t *testing.T) {
	cmd := ctrlCTimeoutCmd(100 * time.Millisecond)
	if cmd == nil {
		t.Error("ctrlCTimeoutCmd() should return a command")
	}
}

func TestDirectoryLoadCompleteMsg(t *testing.T) {
	msg := directoryLoadCompleteMsg{
		paneID:              LeftPane,
		panePath:            "/test/path",
		isHistoryNavigation: true,
	}

	if msg.paneID != LeftPane {
		t.Errorf("paneID = %v, want %v", msg.paneID, LeftPane)
	}

	if msg.panePath != "/test/path" {
		t.Errorf("panePath = %s, want /test/path", msg.panePath)
	}

	if !msg.isHistoryNavigation {
		t.Error("isHistoryNavigation should be true")
	}
}

func TestArchiveProgressUpdateMsg(t *testing.T) {
	msg := archiveProgressUpdateMsg{
		taskID:         "task-123",
		progress:       0.5,
		processedFiles: 50,
		totalFiles:     100,
		currentFile:    "file.txt",
		elapsedTime:    30 * time.Second,
	}

	if msg.taskID != "task-123" {
		t.Errorf("taskID = %s, want task-123", msg.taskID)
	}

	if msg.progress != 0.5 {
		t.Errorf("progress = %f, want 0.5", msg.progress)
	}

	if msg.processedFiles != 50 {
		t.Errorf("processedFiles = %d, want 50", msg.processedFiles)
	}

	if msg.totalFiles != 100 {
		t.Errorf("totalFiles = %d, want 100", msg.totalFiles)
	}
}

func TestArchiveOperationCompleteMsg(t *testing.T) {
	msg := archiveOperationCompleteMsg{
		taskID:      "task-456",
		success:     true,
		cancelled:   false,
		archivePath: "/path/to/archive.tar.gz",
	}

	if msg.taskID != "task-456" {
		t.Errorf("taskID = %s, want task-456", msg.taskID)
	}

	if !msg.success {
		t.Error("success should be true")
	}

	if msg.cancelled {
		t.Error("cancelled should be false")
	}

	if msg.archivePath != "/path/to/archive.tar.gz" {
		t.Errorf("archivePath = %s, want /path/to/archive.tar.gz", msg.archivePath)
	}
}

func TestCompressionLevelResultMsg(t *testing.T) {
	msg := compressionLevelResultMsg{
		level:     6,
		cancelled: false,
	}

	if msg.level != 6 {
		t.Errorf("level = %d, want 6", msg.level)
	}

	if msg.cancelled {
		t.Error("cancelled should be false")
	}
}

func TestArchiveNameResultMsg(t *testing.T) {
	msg := archiveNameResultMsg{
		name:      "myarchive.tar.gz",
		cancelled: false,
	}

	if msg.name != "myarchive.tar.gz" {
		t.Errorf("name = %s, want myarchive.tar.gz", msg.name)
	}

	if msg.cancelled {
		t.Error("cancelled should be false")
	}
}

func TestExtractSecurityCheckMsg(t *testing.T) {
	msg := extractSecurityCheckMsg{
		archivePath:   "/path/to/archive.tar.gz",
		destDir:       "/path/to/dest",
		archiveSize:   1024,
		extractedSize: 10240,
		availableSize: 100000,
		compressionOK: true,
		diskSpaceOK:   true,
		ratio:         10.0,
	}

	if msg.archivePath != "/path/to/archive.tar.gz" {
		t.Errorf("archivePath = %s, want /path/to/archive.tar.gz", msg.archivePath)
	}

	if !msg.compressionOK {
		t.Error("compressionOK should be true")
	}

	if !msg.diskSpaceOK {
		t.Error("diskSpaceOK should be true")
	}

	if msg.ratio != 10.0 {
		t.Errorf("ratio = %f, want 10.0", msg.ratio)
	}
}
