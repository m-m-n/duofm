package ui

import (
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// execFinishedMsg is sent when external command completes
type execFinishedMsg struct {
	err error
}

// openWithViewer opens the file with less
func openWithViewer(path string) tea.Cmd {
	c := exec.Command("less", path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return execFinishedMsg{err: err}
	})
}

// openWithEditor opens the file with vim
func openWithEditor(path string) tea.Cmd {
	c := exec.Command("vim", path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return execFinishedMsg{err: err}
	})
}

// checkReadPermission verifies the file can be read
func checkReadPermission(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

// shellCommandFinishedMsg is sent when shell command completes
type shellCommandFinishedMsg struct {
	err error
}

// executeShellCommand executes a shell command in the specified directory
func executeShellCommand(command, workDir string) tea.Cmd {
	shellCmd := exec.Command("/bin/sh", "-c", command+"; echo; echo 'Press Enter to continue...'; read _")
	shellCmd.Dir = workDir
	return tea.ExecProcess(shellCmd, func(err error) tea.Msg {
		return shellCommandFinishedMsg{err: err}
	})
}
