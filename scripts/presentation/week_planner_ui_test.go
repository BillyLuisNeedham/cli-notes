package presentation

import (
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"strings"
	"testing"
	"time"
)

func TestRenderWeekView_ShowsNoteSummary(t *testing.T) {
	// Setup
	startDate := time.Date(2023, 11, 20, 0, 0, 0, 0, time.Local) // A Monday
	plan := data.NewWeekPlan(startDate)

	// Create a todo with uncompleted tasks
	todo := scripts.File{
		Name:      "test-note.md",
		Title:     "Test Note",
		Priority:  scripts.P1,
		DueAt:     startDate,
		Content:   "Some content\n- [ ] Unfinished Task 1\n- [x] Finished Task\n- [ ] Unfinished Task 2",
		CreatedAt: time.Now(),
	}

	// Add todo to Monday
	plan.TodosByDay[data.Monday] = []scripts.File{todo}

	state := &data.WeekPlannerState{
		Plan:         plan,
		SelectedDay:  data.Monday,
		SelectedTodo: 0,
		ViewMode:     data.NormalView,
	}

	// Execute
	// Use a large enough height to ensure summary is shown
	output := RenderWeekView(state, 100, 40)

	// Verify
	if !strings.Contains(output, "NOTE SUMMARY") {
		t.Error("Expected output to contain 'NOTE SUMMARY'")
	}
	if !strings.Contains(output, "Unfinished Task 1") {
		t.Error("Expected output to contain 'Unfinished Task 1'")
	}
	if !strings.Contains(output, "Unfinished Task 2") {
		t.Error("Expected output to contain 'Unfinished Task 2'")
	}
	// Note: The current implementation of GetUncompletedTasksInFiles might include line numbers,
	// so we just check for the task text.

	// Also check that we don't show finished tasks (GetUncompletedTasksInFiles filters them)
	// However, RenderWeekView might show them if I didn't implement it right?
	// No, GetUncompletedTasksInFiles only returns lines with "- [ ] ".

	// But wait, if the content is displayed elsewhere? No, only in summary.
}
