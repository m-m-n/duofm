// Package clipboard provides clipboard write functionality using OSC 52
// escape sequences and external command fallback (wl-copy, xclip, xsel).
package clipboard

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// clipboardCmd represents an external clipboard command with its arguments.
type clipboardCmd struct {
	name string
	args []string
}

// findClipboardCommandFunc is the function used to detect external clipboard commands.
// It can be overridden in tests.
var findClipboardCommandFunc = findClipboardCommand

// buildOSC52Sequence generates an OSC 52 escape sequence for the given text.
// The text is base64-encoded and wrapped in the OSC 52 format: \033]52;c;{base64}\a
func buildOSC52Sequence(text string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	return fmt.Sprintf("\033]52;c;%s\a", encoded)
}

// writeOSC52 writes the OSC 52 escape sequence for the given text to the writer.
func writeOSC52(w io.Writer, text string) error {
	seq := buildOSC52Sequence(text)
	_, err := io.WriteString(w, seq)
	return err
}

// findClipboardCommand detects the first available external clipboard command.
// Detection order: wl-copy > xclip > xsel
func findClipboardCommand() *clipboardCmd {
	commands := []clipboardCmd{
		{name: "wl-copy", args: nil},
		{name: "xclip", args: []string{"-selection", "clipboard"}},
		{name: "xsel", args: []string{"--clipboard", "--input"}},
	}

	for _, cmd := range commands {
		if _, err := exec.LookPath(cmd.name); err == nil {
			return &clipboardCmd{name: cmd.name, args: cmd.args}
		}
	}
	return nil
}

// execClipboardCommand executes an external command with the given text piped to stdin.
// The command is run with the provided context for timeout control.
func execClipboardCommand(ctx context.Context, name string, args []string, text string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// WriteToClipboard writes the given text to the system clipboard.
//
// Strategy:
//  1. If ttyWriter is non-nil, write OSC 52 escape sequence (best-effort).
//  2. Detect and execute external clipboard command (wl-copy, xclip, xsel) with 5s timeout.
//  3. If no external command found and OSC 52 was attempted (ttyWriter non-nil), return success.
//  4. If ttyWriter is nil and no external command found, return error.
func WriteToClipboard(text string, ttyWriter io.Writer) error {
	osc52Attempted := false

	// Step 1: Attempt OSC 52 via provided writer
	if ttyWriter != nil {
		osc52Attempted = true
		// Best-effort: write error is logged but not fatal
		writeOSC52(ttyWriter, text)
	}

	// Step 2: Try external command
	extCmd := findClipboardCommandFunc()
	if extCmd != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := execClipboardCommand(ctx, extCmd.name, extCmd.args, text); err != nil {
			return fmt.Errorf("clipboard command failed: %w", err)
		}
		return nil
	}

	// Step 3: No external command available
	if osc52Attempted {
		return nil // OSC 52 may have worked
	}

	// Step 4: No clipboard method available
	return fmt.Errorf("no clipboard method available")
}
