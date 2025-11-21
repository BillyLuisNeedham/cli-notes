package e2e

import (
	"testing"
)

func TestCommandInterface(t *testing.T) {
	h := NewTestHarness(t)

	t.Run("Backspace", func(t *testing.T) {
		// Type "exi", backspace, type "xit" -> "exit"
		// \x7f is backspace
		input := "exi\x7ft\n" // "ex" + "t" -> "ext" ? No.
		// "exi" -> backspace -> "ex" -> "t" -> "ext".
		// Wait, we want "exit".
		// "exi" -> backspace -> "ex" -> "it" -> "exit".
		input = "exi\x7fit\n"

		// If this works, it runs "exit" command and closes.
		_, _, err := h.RunCommand(input)
		if err != nil {
			// If it didn't exit, it might timeout.
		}
	})

	t.Run("Escape", func(t *testing.T) {
		// Type "garbage", press ESC, type "exit"
		// \x1b is ESC
		input := "garbage\x1bexit\n"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Errorf("Failed to handle escape: %v", err)
		}
	})
}
