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

// getEditor returns the editor command from $EDITOR or "vim" as fallback
func getEditor() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return "vim"
	}
	return editor
}

// getPager returns the pager command from $PAGER or "less" as fallback
func getPager() string {
	pager := os.Getenv("PAGER")
	if pager == "" {
		return "less"
	}
	return pager
}

// openWithViewer opens the file with pager ($PAGER or less)
func openWithViewer(path, workDir string) tea.Cmd {
	c := exec.Command(getPager(), path)
	c.Dir = workDir
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return execFinishedMsg{err: err}
	})
}

// openWithEditor opens the file with editor ($EDITOR or vim)
func openWithEditor(path, workDir string) tea.Cmd {
	c := exec.Command(getEditor(), path)
	c.Dir = workDir
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
