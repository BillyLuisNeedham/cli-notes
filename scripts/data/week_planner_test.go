package data

import (
	"cli-notes/scripts"
	"os"
	"testing"
	"time"
)

func TestRefreshTodo_CompletedTodoIsRemoved(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create a week plan with one incomplete todo on Monday
	plan := NewWeekPlan(time.Now())
	todo := scripts.File{
		Name:     "test-todo.md",
		Title:    "Test Todo",
		DueAt:    plan.GetDateForWeekDay(Monday),
		Done:     false,
		Priority: scripts.P2,
		Content:  "Test content",
	}
	createTestFile(t, todo)

	// Add todo to the plan
	plan.TodosByDay[Monday] = []scripts.File{todo}

	// Verify todo is in the plan
	if len(plan.TodosByDay[Monday]) != 1 {
		t.Fatalf("Expected 1 todo on Monday, got %d", len(plan.TodosByDay[Monday]))
	}

	// Mark the todo as complete on disk
	todo.Done = true
	err := WriteFile(todo)
	if err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// Refresh the todo
	err = plan.RefreshTodo(todo.Name)
	if err != nil {
		t.Fatalf("RefreshTodo failed: %v", err)
	}

	// Verify todo was removed from the plan
	if len(plan.TodosByDay[Monday]) != 0 {
		t.Errorf("Expected completed todo to be removed from plan, but %d todos remain", len(plan.TodosByDay[Monday]))
	}
}

func TestRefreshTodo_CompletedTodoWithUnsavedMove(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create a week plan
	plan := NewWeekPlan(time.Now())
	todo := scripts.File{
		Name:     "moved-todo.md",
		Title:    "Moved Todo",
		DueAt:    plan.GetDateForWeekDay(Monday), // Original due date
		Done:     false,
		Priority: scripts.P1,
		Content:  "Test content",
	}
	createTestFile(t, todo)

	// Add todo to Monday
	plan.TodosByDay[Monday] = []scripts.File{todo}

	// Simulate unsaved move to Tuesday
	plan.MoveTodo(todo, Monday, Tuesday)

	// Verify todo is now on Tuesday with unsaved move
	if len(plan.TodosByDay[Tuesday]) != 1 {
		t.Fatalf("Expected 1 todo on Tuesday after move, got %d", len(plan.TodosByDay[Tuesday]))
	}
	if len(plan.Changes) != 1 {
		t.Fatalf("Expected 1 change recorded, got %d", len(plan.Changes))
	}

	// Mark the todo as complete on disk (but keep original Monday due date)
	todo.Done = true
	err := WriteFile(todo)
	if err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// Refresh the todo
	err = plan.RefreshTodo(todo.Name)
	if err != nil {
		t.Fatalf("RefreshTodo failed: %v", err)
	}

	// Verify todo was removed from Tuesday (where it was moved to)
	if len(plan.TodosByDay[Tuesday]) != 0 {
		t.Errorf("Expected completed todo to be removed from Tuesday, but %d todos remain", len(plan.TodosByDay[Tuesday]))
	}

	// Verify todo is not on Monday either
	if len(plan.TodosByDay[Monday]) != 0 {
		t.Errorf("Expected completed todo to not be on Monday, but %d todos found", len(plan.TodosByDay[Monday]))
	}

	// Note: Changes history is preserved (by design - shows what was moved)
	// The todo is simply removed from display
}

func TestRefreshTodo_IncompleteTodoRemainsInPlan(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create a week plan
	plan := NewWeekPlan(time.Now())
	todo := scripts.File{
		Name:     "incomplete-todo.md",
		Title:    "Still Incomplete",
		DueAt:    plan.GetDateForWeekDay(Wednesday),
		Done:     false,
		Priority: scripts.P2,
		Content:  "Test content",
	}
	createTestFile(t, todo)

	// Add todo to Wednesday
	plan.TodosByDay[Wednesday] = []scripts.File{todo}

	// Update title on disk but keep it incomplete
	todo.Title = "Updated Title"
	err := WriteFile(todo)
	if err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// Refresh the todo
	err = plan.RefreshTodo(todo.Name)
	if err != nil {
		t.Fatalf("RefreshTodo failed: %v", err)
	}

	// Verify todo is still in the plan
	if len(plan.TodosByDay[Wednesday]) != 1 {
		t.Errorf("Expected incomplete todo to remain in plan, but got %d todos", len(plan.TodosByDay[Wednesday]))
	}

	// Verify title was updated
	if plan.TodosByDay[Wednesday][0].Title != "Updated Title" {
		t.Errorf("Expected title to be updated to 'Updated Title', got '%s'", plan.TodosByDay[Wednesday][0].Title)
	}
}

func TestRefreshTodo_CompletedTodoNotInPlan(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create a completed todo on disk
	plan := NewWeekPlan(time.Now())
	todo := scripts.File{
		Name:     "already-complete.md",
		Title:    "Already Complete",
		DueAt:    plan.GetDateForWeekDay(Friday),
		Done:     true, // Already complete
		Priority: scripts.P3,
		Content:  "Test content",
	}
	createTestFile(t, todo)

	// Verify plan is empty
	if len(plan.TodosByDay[Friday]) != 0 {
		t.Fatalf("Expected empty plan, got %d todos on Friday", len(plan.TodosByDay[Friday]))
	}

	// Try to refresh a todo that's not in the plan
	err := plan.RefreshTodo(todo.Name)
	if err != nil {
		t.Fatalf("RefreshTodo failed: %v", err)
	}

	// Verify todo was not added to the plan
	if len(plan.TodosByDay[Friday]) != 0 {
		t.Errorf("Expected completed todo to not be added to plan, but got %d todos", len(plan.TodosByDay[Friday]))
	}
}

func TestRefreshTodo_IncompleteTodoNotInPlanGetsAdded(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create an incomplete todo on disk that's not in the plan
	plan := NewWeekPlan(time.Now())
	todo := scripts.File{
		Name:     "new-todo.md",
		Title:    "New Todo",
		DueAt:    plan.GetDateForWeekDay(Thursday),
		Done:     false,
		Priority: scripts.P1,
		Content:  "Test content",
	}
	createTestFile(t, todo)

	// Verify plan is empty
	if len(plan.TodosByDay[Thursday]) != 0 {
		t.Fatalf("Expected empty plan, got %d todos on Thursday", len(plan.TodosByDay[Thursday]))
	}

	// Refresh the todo (existing behavior test)
	err := plan.RefreshTodo(todo.Name)
	if err != nil {
		t.Fatalf("RefreshTodo failed: %v", err)
	}

	// Verify todo was added to the plan
	if len(plan.TodosByDay[Thursday]) != 1 {
		t.Errorf("Expected incomplete todo to be added to plan, but got %d todos", len(plan.TodosByDay[Thursday]))
	}
}

func TestRefreshTodo_DeletedFileIsRemoved(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create a week plan with one todo
	plan := NewWeekPlan(time.Now())
	todo := scripts.File{
		Name:     "deleted-todo.md",
		Title:    "To Be Deleted",
		DueAt:    plan.GetDateForWeekDay(Tuesday),
		Done:     false,
		Priority: scripts.P2,
		Content:  "Test content",
	}
	createTestFile(t, todo)

	// Add todo to the plan
	plan.TodosByDay[Tuesday] = []scripts.File{todo}

	// Verify todo is in the plan
	if len(plan.TodosByDay[Tuesday]) != 1 {
		t.Fatalf("Expected 1 todo on Tuesday, got %d", len(plan.TodosByDay[Tuesday]))
	}

	// Delete the file from disk
	// Note: The file was created, so we need to actually delete it
	filePath := "notes/" + todo.Name
	err := os.Remove(filePath)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Refresh the todo
	err = plan.RefreshTodo(todo.Name)
	if err != nil {
		t.Fatalf("RefreshTodo failed: %v", err)
	}

	// Verify todo was removed from the plan
	if len(plan.TodosByDay[Tuesday]) != 0 {
		t.Errorf("Expected deleted todo to be removed from plan, but %d todos remain", len(plan.TodosByDay[Tuesday]))
	}
}
