package e2e

import (
	"strings"
	"testing"
)

func TestPriorityShortcuts(t *testing.T) {
	t.Run("SingleKeySetsPriority1", func(t *testing.T) {
		h := NewTestHarness(t)
		dateStr := Today()
		h.CreateTodo("test-todo-"+dateStr+".md", "Test Todo", []string{"work"}, dateStr, false, 0)

		// Test: List todos, select one, press '1' to set priority to P1
		// Flow: gt -> ArrowDown to select -> 1 -> Enter
		input := "gt\n\x1b[B1\n"

		_, _, err := h.RunCommand(input)
		if err != nil {
			// Command might timeout or exit, that's ok
		}

		// Verify the todo now has priority 1
		fm := h.ParseFrontmatter("test-todo-" + dateStr + ".md")
		if fm.Priority != 1 {
			t.Errorf("Expected priority to be set to 1, got %d", fm.Priority)
		}
	})

	t.Run("SingleKeySetsPriority2", func(t *testing.T) {
		h := NewTestHarness(t)
		dateStr := Today()
		h.CreateTodo("test-todo-"+dateStr+".md", "Test Todo", []string{"work"}, dateStr, false, 0)

		// Test: List todos, select one, press '2' to set priority to P2
		input := "gt\n\x1b[B2\n"

		_, _, err := h.RunCommand(input)
		if err != nil {
			// Command might timeout or exit, that's ok
		}

		// Verify the todo now has priority 2
		fm := h.ParseFrontmatter("test-todo-" + dateStr + ".md")
		if fm.Priority != 2 {
			t.Errorf("Expected priority to be set to 2, got %d", fm.Priority)
		}
	})

	t.Run("SingleKeySetsPriority3", func(t *testing.T) {
		h := NewTestHarness(t)
		dateStr := Today()
		h.CreateTodo("test-todo-"+dateStr+".md", "Test Todo", []string{"work"}, dateStr, false, 0)

		// Test: List todos, select one, press '3' to set priority to P3
		input := "gt\n\x1b[B3\n"

		_, _, err := h.RunCommand(input)
		if err != nil {
			// Command might timeout or exit, that's ok
		}

		// Verify the todo now has priority 3
		fm := h.ParseFrontmatter("test-todo-" + dateStr + ".md")
		if fm.Priority != 3 {
			t.Errorf("Expected priority to be set to 3, got %d", fm.Priority)
		}
	})

	t.Run("NumbersInCommandTextStillWork", func(t *testing.T) {
		h := NewTestHarness(t)
		dateStr := Today()

		// When typing a command that contains numbers (like "gt1"),
		// the numbers should be part of the command text, not execute as shortcuts

		// Create a todo with "1" in the title to test the query
		h.CreateTodo("task1-"+dateStr+".md", "Task 1", []string{"test"}, dateStr, false, 0)

		// Typing "gt1" should execute "gt 1" query, searching for "1"
		// It should NOT be interpreted as "gt" followed by priority shortcut "1"
		stdout, _, err := h.RunCommand("gt1\n")
		if err != nil {
			// Might fail if command is unknown, but we expect it to work
		}

		// The command might not find results (gt1 is searching for "1"),
		// but the important thing is it attempted to execute "gt 1" query
		// rather than erroring or behaving like a priority shortcut
		// We verify this indirectly: if it was treated as a priority shortcut,
		// it would have printed "No file selected"
		if strings.Contains(stdout, "No file selected") {
			t.Errorf("gt1 was incorrectly interpreted as priority shortcut. Got:\n%s", stdout)
		}
	})
}
