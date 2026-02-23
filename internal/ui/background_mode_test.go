package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestModelForBg(t *testing.T) Model {
	t.Helper()
	m := NewModel()
	// Simulate window size so panes initialize
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	result, _ := m.Update(msg)
	return result.(Model)
}

func TestBgMode_ExclaimToggle(t *testing.T) {
	m := newTestModelForBg(t)
	m.startShellCommandMode()

	// Press ! to enter background mode
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("!")}
	result, _ := m.handleShellCommandInput(key)
	rm := result.(Model)

	if !rm.bgMode {
		t.Error("expected bgMode=true after first !")
	}
	if !rm.shellCommandMode {
		t.Error("expected shellCommandMode still active")
	}
}

func TestBgMode_SecondExclaimAppendsChar(t *testing.T) {
	m := newTestModelForBg(t)
	m.startShellCommandMode()

	// Enter bg mode
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("!")}
	result, _ := m.handleShellCommandInput(key)
	rm := result.(Model)

	// Second ! should append character
	result2, _ := rm.handleShellCommandInput(key)
	rm2 := result2.(Model)

	if !rm2.bgMode {
		t.Error("bgMode should still be true")
	}
	if rm2.minibuffer.Input() != "!" {
		t.Errorf("expected input '!', got %q", rm2.minibuffer.Input())
	}
}

func TestBgMode_BackspaceOnEmptyExits(t *testing.T) {
	m := newTestModelForBg(t)
	m.startShellCommandMode()

	// Enter bg mode
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("!")}
	result, _ := m.handleShellCommandInput(key)
	rm := result.(Model)

	// Backspace on empty input should exit bgMode
	bsKey := tea.KeyMsg{Type: tea.KeyBackspace}
	result2, _ := rm.handleShellCommandInput(bsKey)
	rm2 := result2.(Model)

	if rm2.bgMode {
		t.Error("expected bgMode=false after backspace on empty")
	}
	if !rm2.shellCommandMode {
		t.Error("expected still in shellCommandMode after backspace exit")
	}
}

func TestBgMode_BackspaceOnNonEmptyDeletesChar(t *testing.T) {
	m := newTestModelForBg(t)
	m.startShellCommandMode()

	// Enter bg mode
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("!")}
	result, _ := m.handleShellCommandInput(key)
	rm := result.(Model)

	// Type some text
	textKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	result2, _ := rm.handleShellCommandInput(textKey)
	rm2 := result2.(Model)

	// Backspace should delete character, not exit bgMode
	bsKey := tea.KeyMsg{Type: tea.KeyBackspace}
	result3, _ := rm2.handleShellCommandInput(bsKey)
	rm3 := result3.(Model)

	if !rm3.bgMode {
		t.Error("bgMode should still be true after backspace on non-empty input")
	}
}

func TestBgMode_EscapeCancels(t *testing.T) {
	m := newTestModelForBg(t)
	m.startShellCommandMode()

	// Enter bg mode
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("!")}
	result, _ := m.handleShellCommandInput(key)
	rm := result.(Model)

	// Escape should cancel both bgMode and shellCommandMode
	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	result2, _ := rm.handleShellCommandInput(escKey)
	rm2 := result2.(Model)

	if rm2.bgMode {
		t.Error("expected bgMode=false after Escape")
	}
	if rm2.shellCommandMode {
		t.Error("expected shellCommandMode=false after Escape")
	}
}

func TestBgMode_ShellCommandBlockedDuringBgExec(t *testing.T) {
	m := newTestModelForBg(t)

	// Simulate a running background command
	m.bgRunner.Start("sleep 60", "/tmp", LeftPane,
		func(line string) {},
		func(err error) {},
	)
	defer m.bgRunner.Cancel()

	// Try to enter shell command mode via action
	result, _ := m.handleAction(ActionShellCommand)
	rm := result.(Model)

	if rm.shellCommandMode {
		t.Error("should not enter shell command mode while bg running")
	}
	if rm.statusMessage != "Background command running" {
		t.Errorf("expected warning message, got %q", rm.statusMessage)
	}
}

func TestBgMode_EmptyEnterCancels(t *testing.T) {
	m := newTestModelForBg(t)
	m.startShellCommandMode()

	// Enter bg mode
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("!")}
	result, _ := m.handleShellCommandInput(key)
	rm := result.(Model)

	// Enter with empty input should cancel
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	result2, _ := rm.handleShellCommandInput(enterKey)
	rm2 := result2.(Model)

	if rm2.shellCommandMode {
		t.Error("expected shellCommandMode=false after empty Enter")
	}
	if rm2.bgMode {
		t.Error("expected bgMode=false after empty Enter")
	}
}

func TestBgOutputMsg_AppendsToBuffer(t *testing.T) {
	m := newTestModelForBg(t)
	// Set up channels so waitForBgEvent works
	m.bgOutputCh = make(chan string, 10)
	m.bgDoneCh = make(chan error, 1)
	m.bgCommand = "test"
	m.bgWorkDir = "/tmp"

	msg := bgOutputMsg{line: "hello output"}
	result, cmd := m.Update(msg)
	rm := result.(Model)

	lines := rm.bgOutputBuffer.Lines()
	if len(lines) != 1 || lines[0] != "hello output" {
		t.Errorf("expected [hello output], got %v", lines)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd (should wait for next event)")
	}
}

func TestBgCommandDoneMsg_SetsClosing(t *testing.T) {
	m := newTestModelForBg(t)

	msg := bgCommandDoneMsg{err: nil, command: "echo test", workDir: "/tmp"}
	result, cmd := m.Update(msg)
	rm := result.(Model)

	if !rm.bgClosing {
		t.Error("expected bgClosing=true after done msg")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd (2-second timer)")
	}
}

func TestBgAutoCloseMsg_ClearsState(t *testing.T) {
	m := newTestModelForBg(t)
	m.bgClosing = true
	m.bgOutputBuffer.Append("some output")
	m.bgOutputFocused = true

	msg := bgAutoCloseMsg{}
	result, _ := m.Update(msg)
	rm := result.(Model)

	if rm.bgClosing {
		t.Error("expected bgClosing=false after auto-close")
	}
	if rm.bgOutputFocused {
		t.Error("expected bgOutputFocused=false after auto-close")
	}
	if rm.bgOutputBuffer.LineCount() != 0 {
		t.Error("expected buffer cleared after auto-close")
	}
}

// Phase 5 Tests: Focus, Cancellation, and Auto-Close

func TestBgFocus_TabFocusesOutputWhenBgRunning(t *testing.T) {
	m := newTestModelForBg(t)

	// Start a background command on the active pane
	m.bgRunner.Start("sleep 60", "/tmp", m.activePane,
		func(line string) {},
		func(err error) {},
	)
	defer m.bgRunner.Cancel()

	// TAB should focus the output area
	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := m.handleKeyInput(tabKey)
	rm := result.(Model)

	if !rm.bgOutputFocused {
		t.Error("expected bgOutputFocused=true after TAB with bg running on active pane")
	}
}

func TestBgFocus_TabDoesNotFocusWhenBgNotRunning(t *testing.T) {
	m := newTestModelForBg(t)

	// No bg command running; TAB should not set bgOutputFocused
	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := m.handleKeyInput(tabKey)
	rm := result.(Model)

	if rm.bgOutputFocused {
		t.Error("expected bgOutputFocused=false when no bg running")
	}
}

func TestBgFocus_TabDoesNotFocusWhenBgOnOppositPane(t *testing.T) {
	m := newTestModelForBg(t)

	// Start bg on opposite pane
	oppPane := RightPane
	if m.activePane == RightPane {
		oppPane = LeftPane
	}
	m.bgRunner.Start("sleep 60", "/tmp", oppPane,
		func(line string) {},
		func(err error) {},
	)
	defer m.bgRunner.Cancel()

	// TAB should NOT focus output (bg is on opposite pane)
	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := m.handleKeyInput(tabKey)
	rm := result.(Model)

	if rm.bgOutputFocused {
		t.Error("expected bgOutputFocused=false when bg on opposite pane")
	}
}

func TestBgFocused_CtrlCCancelsAndClearsState(t *testing.T) {
	m := newTestModelForBg(t)

	// Start bg command
	m.bgRunner.Start("sleep 60", "/tmp", m.activePane,
		func(line string) {},
		func(err error) {},
	)
	m.bgOutputFocused = true
	m.bgOutputBuffer.Append("output line")
	m.bgOutputCh = make(chan string, 10)
	m.bgDoneCh = make(chan error, 1)
	m.bgCommand = "sleep 60"
	m.bgWorkDir = "/tmp"

	// Ctrl+C should cancel and clean up
	ctrlCKey := tea.KeyMsg{Type: tea.KeyCtrlC}
	result, _ := m.handleBgOutputFocusedInput(ctrlCKey)
	rm := result.(Model)

	if rm.bgOutputFocused {
		t.Error("expected bgOutputFocused=false after Ctrl+C")
	}
	if rm.bgClosing {
		t.Error("expected bgClosing=false after Ctrl+C")
	}
	if rm.bgOutputBuffer.LineCount() != 0 {
		t.Error("expected buffer cleared after Ctrl+C")
	}
	if rm.bgCommand != "" {
		t.Error("expected bgCommand cleared after Ctrl+C")
	}
}

func TestBgFocused_TabUnfocuses(t *testing.T) {
	m := newTestModelForBg(t)
	m.bgOutputFocused = true

	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := m.handleBgOutputFocusedInput(tabKey)
	rm := result.(Model)

	if rm.bgOutputFocused {
		t.Error("expected bgOutputFocused=false after TAB in focused mode")
	}
}

func TestBgFocused_EscUnfocuses(t *testing.T) {
	m := newTestModelForBg(t)
	m.bgOutputFocused = true

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleBgOutputFocusedInput(escKey)
	rm := result.(Model)

	if rm.bgOutputFocused {
		t.Error("expected bgOutputFocused=false after Esc in focused mode")
	}
}

func TestBgFocused_OtherKeysIgnored(t *testing.T) {
	m := newTestModelForBg(t)
	m.bgOutputFocused = true

	// Regular key should be ignored
	aKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	result, _ := m.handleBgOutputFocusedInput(aKey)
	rm := result.(Model)

	// bgOutputFocused should remain true (key ignored)
	if !rm.bgOutputFocused {
		t.Error("expected bgOutputFocused=true - other keys should be ignored")
	}
}

func TestBgFocused_RoutedFromHandleKeyInput(t *testing.T) {
	m := newTestModelForBg(t)
	m.bgOutputFocused = true

	// handleKeyInput should route to handleBgOutputFocusedInput
	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleKeyInput(escKey)
	rm := result.(Model)

	if rm.bgOutputFocused {
		t.Error("expected bgOutputFocused=false - handleKeyInput should route to bg focused handler")
	}
}
