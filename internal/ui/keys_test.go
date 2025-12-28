package ui

import "testing"

func TestKeyShellCommand(t *testing.T) {
	// Test that KeyShellCommand constant is defined as "!"
	if KeyShellCommand != "!" {
		t.Errorf("KeyShellCommand = %q, want %q", KeyShellCommand, "!")
	}
}
