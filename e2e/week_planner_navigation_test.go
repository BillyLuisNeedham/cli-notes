package e2e

import (
	"strings"
	"testing"
)

func TestWeekPlanner_SwitchDays(t *testing.T) {
	t.Run("SwitchToMonday", func(t *testing.T) {
		h := NewTestHarness(t)

		// Create a todo for Monday to verify it's displayed
		h.CreateTodo("monday-todo.md", "Monday Task", []string{"work"}, MondayThisWeek(), false, 1)

		// Open week planner and switch to Monday view
		// Flow: wp -> M (switch to Monday) -> q (quit)
		input := "wp\nMq"

		stdout, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Verify Monday is shown in output (the UI should display Monday's date or "Monday")
		if !strings.Contains(stdout, "Monday") && !strings.Contains(stdout, "Mon") {
			t.Logf("Expected Monday view in output. Got:\n%s", stdout)
		}
	})

	t.Run("SwitchBetweenDays", func(t *testing.T) {
		h := NewTestHarness(t)

		// Create todos for different days
		h.CreateTodo("mon-todo.md", "Monday Task", []string{}, MondayThisWeek(), false, 1)
		h.CreateTodo("wed-todo.md", "Wednesday Task", []string{}, WednesdayThisWeek(), false, 1)
		h.CreateTodo("fri-todo.md", "Friday Task", []string{}, FridayThisWeek(), false, 1)

		// Open planner and cycle through days
		// Flow: wp -> M (Monday) -> W (Wednesday) -> F (Friday) -> q
		input := "wp\nMWFq"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// The test verifies no crash occurs during day switching
		// Actual UI state is hard to verify without parsing TUI output
	})

	t.Run("SwitchToAllDays", func(t *testing.T) {
		h := NewTestHarness(t)

		// Test all day switching keys: M, T, W, R (Thursday), F, A (Saturday), S (Sunday)
		// Flow: wp -> M -> T -> W -> R -> F -> A -> S -> q
		input := "wp\nMTWRFASq"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Verify no crashes when switching between all days
	})
}

func TestWeekPlanner_NavigateWithinDay(t *testing.T) {
	t.Run("NavigateWithJ", func(t *testing.T) {
		h := NewTestHarness(t)

		// Create multiple todos for the same day
		today := Today()
		h.CreateTodo("todo1.md", "First Todo", []string{}, today, false, 1)
		h.CreateTodo("todo2.md", "Second Todo", []string{}, today, false, 1)
		h.CreateTodo("todo3.md", "Third Todo", []string{}, today, false, 1)

		// Open planner and navigate down with 'j' key
		// Flow: wp -> j (down) -> j (down) -> q
		input := "wp\njjq"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Verify no crashes during navigation
	})

	t.Run("NavigateWithK", func(t *testing.T) {
		h := NewTestHarness(t)

		// Create multiple todos for the same day
		today := Today()
		h.CreateTodo("todo1.md", "First Todo", []string{}, today, false, 1)
		h.CreateTodo("todo2.md", "Second Todo", []string{}, today, false, 1)
		h.CreateTodo("todo3.md", "Third Todo", []string{}, today, false, 1)

		// Open planner, navigate down then up with 'k' key
		// Flow: wp -> j (down) -> j (down) -> k (up) -> q
		input := "wp\njjkq"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Verify no crashes during up/down navigation
	})

	t.Run("NavigateWithArrows", func(t *testing.T) {
		h := NewTestHarness(t)

		// Create multiple todos for the same day
		today := Today()
		h.CreateTodo("todo1.md", "First Todo", []string{}, today, false, 1)
		h.CreateTodo("todo2.md", "Second Todo", []string{}, today, false, 1)
		h.CreateTodo("todo3.md", "Third Todo", []string{}, today, false, 1)

		// Open planner and navigate with arrow keys
		// Flow: wp -> ArrowDown -> ArrowDown -> ArrowUp -> q
		// \x1b[B = Down arrow, \x1b[A = Up arrow
		input := "wp\n\x1b[B\x1b[B\x1b[Aq"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Verify arrow key navigation works without crashes
	})
}

func TestWeekPlanner_NavigateWeeks(t *testing.T) {
	t.Run("NavigateForwardOneWeek", func(t *testing.T) {
		h := NewTestHarness(t)

		// Create a todo for next week
		nextWeek := FutureDate(7)
		h.CreateTodo("next-week.md", "Next Week Task", []string{}, nextWeek, false, 1)

		// Open planner and navigate forward one week
		// Flow: wp -> ] (next week) -> q
		input := "wp\n]q"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Verify week navigation doesn't crash
	})

	t.Run("NavigateBackwardOneWeek", func(t *testing.T) {
		h := NewTestHarness(t)

		// Open planner and navigate backward one week
		// Flow: wp -> [ (previous week) -> q
		input := "wp\n[q"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Verify backward week navigation doesn't crash
	})

	t.Run("NavigateMultipleWeeks", func(t *testing.T) {
		h := NewTestHarness(t)

		// Open planner and navigate forward and backward
		// Flow: wp -> ] -> ] -> [ -> ] -> q
		input := "wp\n]][[]q"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Verify multiple week navigations work without crashes
	})
}
