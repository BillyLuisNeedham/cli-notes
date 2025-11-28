package e2e

import (
	"testing"
)

func TestWeekPlanner_MoveToSpecificDay(t *testing.T) {
	t.Run("MoveToMonday", func(t *testing.T) {
		h := NewTestHarness(t)

		// Create a todo due today
		today := Today()
		filename := "move-test.md"
		h.CreateTodo(filename, "Task to Move", []string{"work"}, today, false, 1)

		// Open planner and move to Monday using 'm' key
		// Flow: wp -> m (move to Monday) -> Ctrl+S (save) -> q (quit)
		// \x13 = Ctrl+S
		input := "wp\nm\x13q"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Verify due date was updated to Monday
		fm := h.ParseFrontmatter(filename)
		expectedDate := MondayThisWeek()
		if fm.DateDue != expectedDate {
			t.Errorf("Expected due date to be %s (Monday), got %s", expectedDate, fm.DateDue)
		}
	})

	t.Run("MoveToWednesday", func(t *testing.T) {
		h := NewTestHarness(t)

		// Create a todo due Monday
		filename := "wednesday-test.md"
		h.CreateTodo(filename, "Move to Wednesday", []string{}, MondayThisWeek(), false, 1)

		// Move to Wednesday
		// Flow: wp -> M (switch to Monday) -> w (move to Wednesday) -> Ctrl+S -> q
		input := "wp\nMw\x13q"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Verify due date is now Wednesday
		fm := h.ParseFrontmatter(filename)
		expectedDate := WednesdayThisWeek()
		if fm.DateDue != expectedDate {
			t.Errorf("Expected due date to be %s (Wednesday), got %s", expectedDate, fm.DateDue)
		}
	})

	t.Run("MoveToFriday", func(t *testing.T) {
		h := NewTestHarness(t)

		// Create a todo due today
		filename := "friday-test.md"
		h.CreateTodo(filename, "Move to Friday", []string{}, Today(), false, 1)

		// Move to Friday
		// Flow: wp -> f (move to Friday) -> Ctrl+S -> q
		input := "wp\nf\x13q"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Verify due date is now Friday
		fm := h.ParseFrontmatter(filename)
		expectedDate := FridayThisWeek()
		if fm.DateDue != expectedDate {
			t.Errorf("Expected due date to be %s (Friday), got %s", expectedDate, fm.DateDue)
		}
	})
}

func TestWeekPlanner_MoveWithArrows(t *testing.T) {
	t.Run("MoveToNextDay", func(t *testing.T) {
		h := NewTestHarness(t)

		// Create a todo due Monday
		filename := "arrow-next.md"
		h.CreateTodo(filename, "Move Next Day", []string{}, MondayThisWeek(), false, 1)

		// Open planner on Monday and move to next day (Tuesday) with 'l' key
		// Flow: wp -> M (switch to Monday) -> l (move to next day) -> Ctrl+S -> q
		input := "wp\nMl\x13q"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Verify due date is now Tuesday
		fm := h.ParseFrontmatter(filename)
		expectedDate := TuesdayThisWeek()
		if fm.DateDue != expectedDate {
			t.Errorf("Expected due date to be %s (Tuesday), got %s", expectedDate, fm.DateDue)
		}
	})

	t.Run("MoveToPreviousDay", func(t *testing.T) {
		h := NewTestHarness(t)

		// Create a todo due Wednesday
		filename := "arrow-prev.md"
		h.CreateTodo(filename, "Move Previous Day", []string{}, WednesdayThisWeek(), false, 1)

		// Open planner on Wednesday and move to previous day (Tuesday) with 'h' key
		// Flow: wp -> W (switch to Wednesday) -> h (move to previous day) -> Ctrl+S -> q
		input := "wp\nWh\x13q"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Verify due date is now Tuesday
		fm := h.ParseFrontmatter(filename)
		expectedDate := TuesdayThisWeek()
		if fm.DateDue != expectedDate {
			t.Errorf("Expected due date to be %s (Tuesday), got %s", expectedDate, fm.DateDue)
		}
	})
}

func TestWeekPlanner_MoveToNextMonday(t *testing.T) {
	h := NewTestHarness(t)

	// Create a todo due today
	filename := "next-monday.md"
	h.CreateTodo(filename, "Move to Next Monday", []string{}, Today(), false, 1)

	// Move to Next Monday bucket with 'N' key
	// Flow: wp -> N (move to Next Monday) -> Ctrl+S -> q
	input := "wp\nN\x13q"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with: %v", err)
	}

	// Verify due date is next Monday
	fm := h.ParseFrontmatter(filename)
	expectedDate := NextMonday()
	if fm.DateDue != expectedDate {
		t.Errorf("Expected due date to be %s (Next Monday), got %s", expectedDate, fm.DateDue)
	}
}

func TestWeekPlanner_UndoMove(t *testing.T) {
	h := NewTestHarness(t)

	// Create a todo due Monday
	filename := "undo-test.md"
	originalDate := MondayThisWeek()
	h.CreateTodo(filename, "Test Undo", []string{}, originalDate, false, 1)

	// Move to Wednesday, then undo
	// Flow: wp -> M (switch to Monday) -> w (move to Wednesday) -> u (undo) -> Ctrl+S -> q
	input := "wp\nMwu\x13q"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with: %v", err)
	}

	// Verify due date is back to original (Monday)
	fm := h.ParseFrontmatter(filename)
	if fm.DateDue != originalDate {
		t.Errorf("Expected due date to be reverted to %s (Monday), got %s", originalDate, fm.DateDue)
	}
}

func TestWeekPlanner_SaveMoves(t *testing.T) {
	h := NewTestHarness(t)

	// Create two todos on Monday (with different priorities)
	file1 := "save-test-1.md"
	file2 := "save-test-2.md"
	h.CreateTodo(file1, "Save Test 1", []string{}, MondayThisWeek(), false, 1)
	h.CreateTodo(file2, "Save Test 2", []string{}, MondayThisWeek(), false, 2)

	// Move both todos and save
	// Flow: wp -> M (Monday view) -> w (move to Wed) -> j (next todo) -> f (move to Fri) -> Ctrl+S -> q
	input := "wp\nMwjf\x13q"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with: %v", err)
	}

	// Verify first todo moved to Wednesday
	fm1 := h.ParseFrontmatter(file1)
	expectedDate1 := WednesdayThisWeek()
	if fm1.DateDue != expectedDate1 {
		t.Errorf("Expected first todo due date to be %s (Wednesday), got %s", expectedDate1, fm1.DateDue)
	}

	// Verify second todo moved to Friday
	fm2 := h.ParseFrontmatter(file2)
	expectedDate2 := FridayThisWeek()
	if fm2.DateDue != expectedDate2 {
		t.Errorf("Expected second todo due date to be %s (Friday), got %s", expectedDate2, fm2.DateDue)
	}
}

func TestWeekPlanner_ResetMoves(t *testing.T) {
	h := NewTestHarness(t)

	// Create a todo due Monday
	filename := "reset-test.md"
	originalDate := MondayThisWeek()
	h.CreateTodo(filename, "Test Reset", []string{}, originalDate, false, 1)

	// Move to Friday, then reset (discard changes)
	// Flow: wp -> M (Monday view) -> f (move to Friday) -> x (reset) -> q
	input := "wp\nMfxq"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with: %v", err)
	}

	// Verify due date is unchanged (still Monday) because we reset without saving
	fm := h.ParseFrontmatter(filename)
	if fm.DateDue != originalDate {
		t.Errorf("Expected due date to remain %s (Monday) after reset, got %s", originalDate, fm.DateDue)
	}
}
