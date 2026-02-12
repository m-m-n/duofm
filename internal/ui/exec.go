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

// getPager returns the pager command and its arguments from $PAGER or "less" as fallback.
// Supports $PAGER values with arguments (e.g., "less -R").
func getPager() (string, []string) {
	pager := os.Getenv("PAGER")
	if pager == "" {
		return "less", nil
	}
	parts := strings.Fields(pager)
	if len(parts) == 0 {
		return "less", nil
	}
	return parts[0], parts[1:]
}

// openWithViewer opens the file with pager ($PAGER or less)
func openWithViewer(path, workDir string) tea.Cmd {
	pagerCmd, pagerArgs := getPager()
	args := make([]string, len(pagerArgs)+1)
	copy(args, pagerArgs)
	args[len(pagerArgs)] = path
	c := exec.Command(pagerCmd, args...)
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
	err     error
	command string
	workDir string
}

// executeShellCommand executes a shell command in the specified directory.
// Output is displayed on terminal and appended to logFile via tee.
// After the command finishes, waits 2 seconds before returning to TUI.
func executeShellCommand(command, workDir, logFile string) tea.Cmd {
	wrapped := fmt.Sprintf(
		"set -o pipefail; { %s; } 2>&1 | tee -a %q; _exit=$?; sleep 2; exit $_exit",
		command, logFile)
	shellCmd := exec.Command("/bin/sh", "-c", wrapped)
	shellCmd.Dir = workDir
	return tea.ExecProcess(shellCmd, func(err error) tea.Msg {
		return shellCommandFinishedMsg{err: err, command: command, workDir: workDir}
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

// tryCommands tries each command string in order via LookPath.
// Returns the tea.Cmd to execute if a command is found, or nil if all fail.
// Failed command names are appended to notFoundCmds.
func tryCommands(commands []string, filePath, workDir string, notFoundCmds *[]string) tea.Cmd {
	for _, cmdStr := range commands {
		// Parse command string (may include options like "vim -R")
		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		_, err := exec.LookPath(command)
		if err == nil {
			// Command found - use it
			args := append(parts[1:], filePath)
			c := exec.Command(command, args...)
			c.Dir = workDir
			return tea.ExecProcess(c, func(err error) tea.Msg {
				return execFinishedMsg{err: err}
			})
		}
		*notFoundCmds = append(*notFoundCmds, command)
	}
	return nil
}

// tryCommandsBackground tries each command string in order via LookPath,
// launching in background mode (without blocking the terminal).
// Suitable for GUI launchers like xdg-open.
// Returns the tea.Cmd to execute if a command is found, or nil if all fail.
// Failed command names are appended to notFoundCmds.
func tryCommandsBackground(commands []string, filePath, workDir string, notFoundCmds *[]string) tea.Cmd {
	for _, cmdStr := range commands {
		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		_, err := exec.LookPath(command)
		if err == nil {
			args := append(parts[1:], filePath)
			return func() tea.Msg {
				cmd := exec.Command(command, args...)
				cmd.Dir = workDir
				err := cmd.Start()
				if err != nil {
					return execFinishedMsg{err: err}
				}
				go func() {
					_ = cmd.Wait()
				}()
				return execFinishedMsg{err: nil}
			}
		}
		*notFoundCmds = append(*notFoundCmds, command)
	}
	return nil
}

// openWithMIME opens a file based on MIME type configuration.
// It detects the MIME type from the filename, finds matching rules,
// and tries each configured command in order until one is found via LookPath.
// If no match or all MIME commands fail, tries fallback commands.
// If all fallback commands also fail, falls back to pager.
// Returns the command to execute and a status message (empty if a command was found).
// Commands may include options (e.g., "vim -R") which are split by whitespace.
func openWithMIME(filePath, workDir string, mimeCfg config.MIMEBehaviorConfig) (tea.Cmd, string) {
	// Detect MIME type from filename
	filename := filepath.Base(filePath)
	mimeType := config.GetMIMEType(filename)

	var notFoundCmds []string

	// Find matching rule and try MIME rule commands
	commands, found := mimeCfg.FindMatchingRule(mimeType)
	if found && len(commands) > 0 {
		if cmd := tryCommands(commands, filePath, workDir, &notFoundCmds); cmd != nil {
			return cmd, ""
		}
	}

	// Try fallback commands (background execution for GUI launchers like xdg-open)
	if len(mimeCfg.Fallback) > 0 {
		if cmd := tryCommandsBackground(mimeCfg.Fallback, filePath, workDir, &notFoundCmds); cmd != nil {
			return cmd, ""
		}
	}

	// All commands failed - fall back to pager
	if len(notFoundCmds) > 0 {
		statusMsg := fmt.Sprintf("All configured commands failed (%s), using pager", strings.Join(notFoundCmds, ", "))
		return openWithViewer(filePath, workDir), statusMsg
	}

	// No commands were configured at all - silent fallback to pager
	return openWithViewer(filePath, workDir), ""
}
