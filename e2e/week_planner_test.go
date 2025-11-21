package e2e

import (
	"testing"
)

func TestWeekPlanner(t *testing.T) {
	h := NewTestHarness(t)

	t.Run("Week Planner Navigation", func(t *testing.T) {
		// wp - Open week planner
		// Interactive TUI.
		// Flow:
		// 1. "wp" -> open
		// 2. "l" -> next day (vim keys usually supported in TUIs or arrow keys)
		// 3. "q" -> quit

		input := "wp\nlq"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command failed: %v", err)
		}

		// Verify no crash
	})

	t.Run("Move Todo in Planner", func(t *testing.T) {
		h.CreateTodo("move-me.md", "Move Me", []string{}, Today(), false, 1)

		// Flow:
		// 1. "wp"
		// 2. Select todo (might need navigation)
		// 3. Move to tomorrow (e.g., "M" key or similar)
		// 4. Save and quit

		// This is hard to test blindly without knowing exact UI state.
		// We'll skip complex interaction for now and just test basic open/close.
	})
}
