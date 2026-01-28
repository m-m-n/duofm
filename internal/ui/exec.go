package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/config"
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

// openWithCustomForeground opens a file with a custom application in foreground.
// The application path should be an absolute path or available in PATH.
// This function validates the application path at runtime using exec.LookPath().
// If the application is not found or not executable, an error message is returned.
func openWithCustomForeground(application, file, workDir string) tea.Cmd {
	// Validate application path at runtime
	_, err := exec.LookPath(application)
	if err != nil {
		// Return a command that sends an error message
		return func() tea.Msg {
			return execFinishedMsg{err: fmt.Errorf("executable not found: %s", application)}
		}
	}

	c := exec.Command(application, file)
	c.Dir = workDir
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return execFinishedMsg{err: err}
	})
}

// openWithXDG launches xdg-open with the specified file in background
func openWithXDG(file, workDir string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("xdg-open", file)
		cmd.Dir = workDir

		// Start process in background (don't wait for completion)
		err := cmd.Start()
		if err != nil {
			return openWithFinishedMsg{err: err}
		}

		// Reap process in background to avoid zombies
		go func() {
			_ = cmd.Wait() // Ignore error - we only care about successful launch
		}()

		return openWithFinishedMsg{err: nil}
	}
}

// openWithCustom launches a custom application with files in background
func openWithCustom(application string, files []string, workDir string) tea.Cmd {
	return func() tea.Msg {
		// Parse application string (may include options)
		parts := strings.Fields(application)
		if len(parts) == 0 {
			return openWithFinishedMsg{err: fmt.Errorf("application field cannot be empty")}
		}

		// First element is command, rest are options
		command := parts[0]
		args := parts[1:]

		// Append files as separate arguments
		args = append(args, files...)

		cmd := exec.Command(command, args...)
		cmd.Dir = workDir

		// Start process in background (don't wait for completion)
		err := cmd.Start()
		if err != nil {
			return openWithFinishedMsg{err: err}
		}

		// Reap process in background to avoid zombies
		go func() {
			_ = cmd.Wait() // Ignore error - we only care about successful launch
		}()

		return openWithFinishedMsg{err: nil}
	}
}

// openWithMIME opens a file based on MIME type configuration.
// It detects the MIME type from the filename, finds matching rules,
// and tries each configured command in order until one succeeds.
// If no match or all commands fail, falls back to pager.
func openWithMIME(filePath, workDir string, mimeCfg config.MIMEBehaviorConfig) tea.Cmd {
	// Detect MIME type from filename
	filename := filepath.Base(filePath)
	mimeType := config.GetMIMEType(filename)

	// Find matching rule
	commands, found := mimeCfg.FindMatchingRule(mimeType)
	if !found || len(commands) == 0 {
		// No match - fall back to pager
		return openWithViewer(filePath, workDir)
	}

	// Try to find the first available command
	for _, cmd := range commands {
		_, err := exec.LookPath(cmd)
		if err == nil {
			// Command found - use it
			c := exec.Command(cmd, filePath)
			c.Dir = workDir
			return tea.ExecProcess(c, func(err error) tea.Msg {
				return execFinishedMsg{err: err}
			})
		}
	}

	// All commands not found - fall back to pager
	return openWithViewer(filePath, workDir)
}
