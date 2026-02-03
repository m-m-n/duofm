package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ConfigErrorChoice represents the user's choice in the config error dialog.
type ConfigErrorChoice int

const (
	// ConfigErrorChoiceFix repairs with default values
	ConfigErrorChoiceFix ConfigErrorChoice = iota
	// ConfigErrorChoiceQuit exits the application (startup only)
	ConfigErrorChoiceQuit
	// ConfigErrorChoiceKeep keeps previous config (hot-reload only)
	ConfigErrorChoiceKeep
)

// configErrorDialogResultMsg is the result message from the config error dialog.
type configErrorDialogResultMsg struct {
	choice    ConfigErrorChoice
	isStartup bool
}

// ConfigErrorDialog displays config errors and repair options.
type ConfigErrorDialog struct {
	BaseDialog
	title     string
	errorMsg  string
	details   string
	isStartup bool
	styles    DialogStyles
}

// NewConfigErrorDialog creates a startup config error dialog.
func NewConfigErrorDialog(errorMsg, details string) *ConfigErrorDialog {
	base := NewBaseDialog(DialogDisplayScreen)
	return &ConfigErrorDialog{
		BaseDialog: base,
		title:      "Configuration Error",
		errorMsg:   errorMsg,
		details:    details,
		isStartup:  true,
		styles:     DefaultDialogStyles(base.Width()),
	}
}

// NewConfigErrorDialogForReload creates a hot-reload config error dialog.
func NewConfigErrorDialogForReload(errorMsg, details string) *ConfigErrorDialog {
	base := NewBaseDialog(DialogDisplayScreen)
	return &ConfigErrorDialog{
		BaseDialog: base,
		title:      "Configuration Error",
		errorMsg:   errorMsg,
		details:    details,
		isStartup:  false,
		styles:     DefaultDialogStyles(base.Width()),
	}
}

// Update handles key input for the config error dialog.
func (d *ConfigErrorDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "f":
			d.Close()
			return d, func() tea.Msg {
				return configErrorDialogResultMsg{
					choice:    ConfigErrorChoiceFix,
					isStartup: d.isStartup,
				}
			}
		case "q":
			if d.isStartup {
				d.Close()
				return d, func() tea.Msg {
					return configErrorDialogResultMsg{
						choice:    ConfigErrorChoiceQuit,
						isStartup: d.isStartup,
					}
				}
			}
		case "k":
			if !d.isStartup {
				d.Close()
				return d, func() tea.Msg {
					return configErrorDialogResultMsg{
						choice:    ConfigErrorChoiceKeep,
						isStartup: d.isStartup,
					}
				}
			}
		case "esc":
			d.Close()
			if d.isStartup {
				return d, func() tea.Msg {
					return configErrorDialogResultMsg{
						choice:    ConfigErrorChoiceQuit,
						isStartup: d.isStartup,
					}
				}
			}
			return d, func() tea.Msg {
				return configErrorDialogResultMsg{
					choice:    ConfigErrorChoiceKeep,
					isStartup: d.isStartup,
				}
			}
		}
	}

	return d, nil
}

// View renders the config error dialog.
func (d *ConfigErrorDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(d.styles.Title.Render(d.title))
	b.WriteString("\n\n")

	// Error message
	b.WriteString(d.styles.Body.Render(d.errorMsg))
	b.WriteString("\n")

	// Details
	if d.details != "" {
		b.WriteString("\n")
		b.WriteString(d.styles.Body.Render(d.details))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Footer with options
	if d.isStartup {
		b.WriteString(d.styles.Footer.Render("[f] Fix with defaults  [q] Quit"))
	} else {
		b.WriteString(d.styles.Footer.Render("[f] Fix with defaults  [k] Keep previous"))
	}

	return d.styles.Box.Render(b.String())
}
