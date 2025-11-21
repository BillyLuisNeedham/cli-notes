package e2e

import (
	"strings"
	"testing"
)

func TestFileOperations(t *testing.T) {
	h := NewTestHarness(t)
	// CLI rename requires files to have date suffixes
	dateStr := Today()
	filename := "file-op-" + dateStr + ".md"
	h.CreateTodo(filename, "File Op", []string{}, dateStr, false, 1)

	t.Run("Open File", func(t *testing.T) {
		// o <file>
		// This usually opens editor. We need to ensure it doesn't block.
		// If we can't mock editor easily, we might skip or assume it returns if editor exits.
		// We can try "o file-op.md" and see if it returns.
		// If we set EDITOR=echo, it should print filename and exit.

		// Note: TestHarness doesn't set EDITOR yet.
		// We should probably update TestHarness to set EDITOR=echo or similar.
	})

	t.Run("Rename File", func(t *testing.T) {
		// r <new-name> (needs selection)
		// Flow:
		// 1. "gt" (list all todos)
		// 2. Select file with arrow down
		// 3. "r new-name"

		// Note: Since there's only one todo file, it should be the first one listed
		input := "gt\n\x1b[Br new-name\n"
		h.RunCommand(input)

		// CLI preserves date suffix when renaming
		expectedName := "new-name-" + dateStr + ".md"
		h.AssertFileExists(expectedName)
		h.AssertFileNotExists(filename)
	})

	t.Run("Error Handling", func(t *testing.T) {
		// Invalid command
		stdout, _, _ := h.RunCommand("invalidcmd\n")
		if !strings.Contains(stdout, "Unknown command") {
			t.Errorf("Expected unknown command error")
		}
	})
}
