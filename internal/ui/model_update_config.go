package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/config"
)

// configStartupErrorMsg is sent from Init when startup config has errors.
type configStartupErrorMsg struct {
	result *config.ConfigLoadResult
}

// handleConfigMessages processes config-related messages.
func (m Model) handleConfigMessages(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case config.ConfigFileChangedMsg:
		return m.handleConfigFileChanged()

	case configErrorDialogResultMsg:
		return m.handleConfigErrorDialogResult(msg)

	case config.ConfigWatchLostMsg:
		m.statusMessage = "Config file watch lost. Restart to re-enable."
		m.isStatusError = true
		return m, statusMessageClearCmd(10 * time.Second), true

	case configStartupErrorMsg:
		return m.handleConfigStartupError(msg)
	}

	return m, nil, false
}

// handleConfigFileChanged handles the config file change event.
func (m Model) handleConfigFileChanged() (Model, tea.Cmd, bool) {
	if m.configPath == "" {
		return m, nil, true
	}

	result := config.LoadConfigDetailed(m.configPath)

	if !result.HasErrors() {
		// Success - apply the new config
		m.applyConfig(result.Config)
		m.statusMessage = "Config reloaded"
		m.isStatusError = false
		return m, statusMessageClearCmd(3 * time.Second), true
	}

	// Error - show dialog or queue
	if m.dialog != nil {
		// Another dialog is open, queue the error
		m.pendingConfigError = result
		return m, nil, true
	}

	// Show error dialog
	m.pendingReloadResult = result
	errorMsg, details := formatConfigError(result)
	dialog := NewConfigErrorDialogForReload(errorMsg, details)
	m.dialog = dialog
	return m, nil, true
}

// handleConfigErrorDialogResult handles the user's choice in the config error dialog.
func (m Model) handleConfigErrorDialogResult(msg configErrorDialogResultMsg) (Model, tea.Cmd, bool) {
	m.dialog = nil

	switch msg.choice {
	case ConfigErrorChoiceFix:
		if m.pendingReloadResult != nil && m.configPath != "" {
			// Suppress watcher to avoid self-triggered reload
			if m.configWatcher != nil {
				m.configWatcher.SuppressFor(500 * time.Millisecond)
			}

			// Repair the config file
			if err := config.RepairConfig(m.configPath, m.pendingReloadResult); err != nil {
				m.statusMessage = fmt.Sprintf("Failed to repair config: %v", err)
				m.isStatusError = true
				m.pendingReloadResult = nil
				return m, statusMessageClearCmd(5 * time.Second), true
			}

			// Reload the repaired config
			reloadResult := config.LoadConfigDetailed(m.configPath)
			m.applyConfig(reloadResult.Config)
			m.pendingReloadResult = nil
			m.statusMessage = "Config repaired and reloaded"
			m.isStatusError = false
		}

		// Check for pending config errors
		cmd := m.checkPendingConfigError()
		if cmd != nil {
			return m, tea.Batch(statusMessageClearCmd(3*time.Second), cmd), true
		}
		return m, statusMessageClearCmd(3 * time.Second), true

	case ConfigErrorChoiceKeep:
		m.pendingReloadResult = nil
		cmd := m.checkPendingConfigError()
		return m, cmd, true

	case ConfigErrorChoiceQuit:
		return m, tea.Quit, true
	}

	return m, nil, true
}

// handleConfigStartupError handles startup config errors.
func (m Model) handleConfigStartupError(msg configStartupErrorMsg) (Model, tea.Cmd, bool) {
	m.pendingReloadResult = msg.result
	errorMsg, details := formatConfigError(msg.result)
	dialog := NewConfigErrorDialog(errorMsg, details)
	m.dialog = dialog
	return m, nil, true
}

// checkPendingConfigError shows a queued config error dialog if one exists.
func (m *Model) checkPendingConfigError() tea.Cmd {
	if m.dialog == nil && m.pendingConfigError != nil {
		m.pendingReloadResult = m.pendingConfigError
		m.pendingConfigError = nil
		errorMsg, details := formatConfigError(m.pendingReloadResult)
		dialog := NewConfigErrorDialogForReload(errorMsg, details)
		m.dialog = dialog
	}
	return nil
}

// formatConfigError formats a ConfigLoadResult into error and detail strings.
func formatConfigError(result *config.ConfigLoadResult) (string, string) {
	if result.HasSyntaxErr {
		errorMsg := fmt.Sprintf("Syntax error at line %d", result.SyntaxErrLine)
		details := result.SyntaxErrMsg
		return errorMsg, details
	}

	// Value errors
	var fields []string
	for _, e := range result.Errors {
		fields = append(fields, fmt.Sprintf("  - %s: %s", e.Field, e.Message))
	}
	errorMsg := fmt.Sprintf("Invalid values in %d field(s)", len(result.Errors))
	details := strings.Join(fields, "\n")
	return errorMsg, details
}
