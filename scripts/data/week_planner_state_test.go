package data

import (
	"cli-notes/scripts"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupWeekPlannerTest creates a temporary test environment for week planner tests
func setupWeekPlannerTest(t *testing.T) (string, string) {
	// Get original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "week_planner_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create notes subdirectory
	err = os.Mkdir(filepath.Join(tempDir, "notes"), 0755)
	if err != nil {
		t.Fatalf("Failed to create notes directory: %v", err)
	}

	// Change to temp directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	return tempDir, origDir
}

// cleanupWeekPlannerTest cleans up the test environment
func cleanupWeekPlannerTest(t *testing.T, tempDir, origDir string) {
	os.Chdir(origDir)
	os.RemoveAll(tempDir)
}

// createTestTodo creates a test todo file
func createTestTodo(t *testing.T, title string, dueDate time.Time, priority scripts.Priority) scripts.File {
	content := "- [ ] Test task"
	file := scripts.File{
		Name:     title + ".md",
		Title:    title,
		Content:  content,
		DueAt:    dueDate,
		Priority: priority,
		Done:     false,
	}
	return file
}

// TestBulkMoveEarlierTodosToCurrentDay_Success tests successful bulk move
func TestBulkMoveEarlierTodosToCurrentDay_Success(t *testing.T) {
	tempDir, origDir := setupWeekPlannerTest(t)
	defer cleanupWeekPlannerTest(t, tempDir, origDir)

	// Create a week plan starting from Monday
	monday := time.Date(2024, 11, 18, 0, 0, 0, 0, time.UTC)
	plan := &WeekPlan{
		StartDate:  monday,
		TodosByDay: make(map[WeekDay][]scripts.File),
		Changes:    []PlanChange{},
		UndoStack:  []PlanChange{},
		RedoStack:  []PlanChange{},
	}

	// Add todos to Earlier (before Monday)
	plan.TodosByDay[Earlier] = []scripts.File{
		createTestTodo(t, "Earlier Task 1", monday.AddDate(0, 0, -5), scripts.P1),
		createTestTodo(t, "Earlier Task 2", monday.AddDate(0, 0, -3), scripts.P2),
	}

	// Add todos to Monday
	plan.TodosByDay[Monday] = []scripts.File{
		createTestTodo(t, "Monday Task 1", monday, scripts.P1),
	}

	// Add todos to Tuesday and Wednesday (should not be moved)
	tuesday := monday.AddDate(0, 0, 1)
	plan.TodosByDay[Tuesday] = []scripts.File{
		createTestTodo(t, "Tuesday Task 1", tuesday, scripts.P2),
	}

	// Create state and select Tuesday
	state := &WeekPlannerState{
		Plan:         plan,
		SelectedDay:  Tuesday,
		SelectedTodo: 0,
		ViewMode:     NormalView,
	}

	// Execute bulk move
	movedCount, err := state.BulkMoveEarlierTodosToCurrentDay()

	// Assert no error
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Assert correct count (2 from Earlier + 1 from Monday)
	if movedCount != 3 {
		t.Errorf("Expected 3 todos moved, got: %d", movedCount)
	}

	// Assert Earlier is empty
	if len(plan.TodosByDay[Earlier]) != 0 {
		t.Errorf("Expected Earlier to be empty, got %d todos", len(plan.TodosByDay[Earlier]))
	}

	// Assert Monday is empty
	if len(plan.TodosByDay[Monday]) != 0 {
		t.Errorf("Expected Monday to be empty, got %d todos", len(plan.TodosByDay[Monday]))
	}

	// Assert Tuesday has all moved todos (original 1 + moved 3 = 4)
	if len(plan.TodosByDay[Tuesday]) != 4 {
		t.Errorf("Expected Tuesday to have 4 todos, got %d", len(plan.TodosByDay[Tuesday]))
	}

	// Assert selected todo was reset to 0
	if state.SelectedTodo != 0 {
		t.Errorf("Expected SelectedTodo to be 0, got: %d", state.SelectedTodo)
	}

	// Assert changes were recorded in undo stack
	if len(plan.UndoStack) != 3 {
		t.Errorf("Expected 3 entries in undo stack, got: %d", len(plan.UndoStack))
	}
}

// TestBulkMoveEarlierTodosToCurrentDay_NoTodos tests when there are no todos to move
func TestBulkMoveEarlierTodosToCurrentDay_NoTodos(t *testing.T) {
	tempDir, origDir := setupWeekPlannerTest(t)
	defer cleanupWeekPlannerTest(t, tempDir, origDir)

	// Create a week plan with todos only on Tuesday and after
	monday := time.Date(2024, 11, 18, 0, 0, 0, 0, time.UTC)
	tuesday := monday.AddDate(0, 0, 1)

	plan := &WeekPlan{
		StartDate:  monday,
		TodosByDay: make(map[WeekDay][]scripts.File),
		Changes:    []PlanChange{},
		UndoStack:  []PlanChange{},
		RedoStack:  []PlanChange{},
	}

	// Add todos only to Tuesday
	plan.TodosByDay[Tuesday] = []scripts.File{
		createTestTodo(t, "Tuesday Task 1", tuesday, scripts.P1),
	}

	// Create state and select Tuesday
	state := &WeekPlannerState{
		Plan:         plan,
		SelectedDay:  Tuesday,
		SelectedTodo: 0,
		ViewMode:     NormalView,
	}

	// Execute bulk move
	movedCount, err := state.BulkMoveEarlierTodosToCurrentDay()

	// Assert error is returned
	if err == nil {
		t.Fatal("Expected error for no todos to move, got nil")
	}

	// Assert error message
	expectedError := "no earlier todos to move"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got: '%s'", expectedError, err.Error())
	}

	// Assert count is 0
	if movedCount != 0 {
		t.Errorf("Expected 0 todos moved, got: %d", movedCount)
	}
}

// TestBulkMoveEarlierTodosToCurrentDay_CannotMoveToEarlier tests moving to Earlier
func TestBulkMoveEarlierTodosToCurrentDay_CannotMoveToEarlier(t *testing.T) {
	tempDir, origDir := setupWeekPlannerTest(t)
	defer cleanupWeekPlannerTest(t, tempDir, origDir)

	// Create a week plan
	monday := time.Date(2024, 11, 18, 0, 0, 0, 0, time.UTC)
	plan := &WeekPlan{
		StartDate:  monday,
		TodosByDay: make(map[WeekDay][]scripts.File),
		Changes:    []PlanChange{},
		UndoStack:  []PlanChange{},
		RedoStack:  []PlanChange{},
	}

	// Add a todo to Earlier
	plan.TodosByDay[Earlier] = []scripts.File{
		createTestTodo(t, "Earlier Task 1", monday.AddDate(0, 0, -5), scripts.P1),
	}

	// Create state and select Earlier
	state := &WeekPlannerState{
		Plan:         plan,
		SelectedDay:  Earlier,
		SelectedTodo: 0,
		ViewMode:     NormalView,
	}

	// Execute bulk move
	movedCount, err := state.BulkMoveEarlierTodosToCurrentDay()

	// Assert error is returned
	if err == nil {
		t.Fatal("Expected error for moving to Earlier, got nil")
	}

	// Assert error message
	expectedError := "cannot bulk move to Earlier"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got: '%s'", expectedError, err.Error())
	}

	// Assert count is 0
	if movedCount != 0 {
		t.Errorf("Expected 0 todos moved, got: %d", movedCount)
	}
}

// TestBulkMoveEarlierTodosToCurrentDay_ToNextMonday tests moving to NextMonday
func TestBulkMoveEarlierTodosToCurrentDay_ToNextMonday(t *testing.T) {
	tempDir, origDir := setupWeekPlannerTest(t)
	defer cleanupWeekPlannerTest(t, tempDir, origDir)

	// Create a week plan
	monday := time.Date(2024, 11, 18, 0, 0, 0, 0, time.UTC)
	plan := &WeekPlan{
		StartDate:  monday,
		TodosByDay: make(map[WeekDay][]scripts.File),
		Changes:    []PlanChange{},
		UndoStack:  []PlanChange{},
		RedoStack:  []PlanChange{},
	}

	// Add todos to Earlier and all weekdays
	plan.TodosByDay[Earlier] = []scripts.File{
		createTestTodo(t, "Earlier Task 1", monday.AddDate(0, 0, -5), scripts.P1),
	}
	plan.TodosByDay[Monday] = []scripts.File{
		createTestTodo(t, "Monday Task 1", monday, scripts.P1),
	}
	plan.TodosByDay[Tuesday] = []scripts.File{
		createTestTodo(t, "Tuesday Task 1", monday.AddDate(0, 0, 1), scripts.P2),
	}
	plan.TodosByDay[Wednesday] = []scripts.File{
		createTestTodo(t, "Wednesday Task 1", monday.AddDate(0, 0, 2), scripts.P2),
	}
	plan.TodosByDay[Thursday] = []scripts.File{
		createTestTodo(t, "Thursday Task 1", monday.AddDate(0, 0, 3), scripts.P1),
	}
	plan.TodosByDay[Friday] = []scripts.File{
		createTestTodo(t, "Friday Task 1", monday.AddDate(0, 0, 4), scripts.P2),
	}
	plan.TodosByDay[Saturday] = []scripts.File{
		createTestTodo(t, "Saturday Task 1", monday.AddDate(0, 0, 5), scripts.P3),
	}
	plan.TodosByDay[Sunday] = []scripts.File{
		createTestTodo(t, "Sunday Task 1", monday.AddDate(0, 0, 6), scripts.P1),
	}

	// Create state and select NextMonday
	state := &WeekPlannerState{
		Plan:         plan,
		SelectedDay:  NextMonday,
		SelectedTodo: 0,
		ViewMode:     NormalView,
	}

	// Execute bulk move
	movedCount, err := state.BulkMoveEarlierTodosToCurrentDay()

	// Assert no error
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Assert correct count (8 todos: 1 Earlier + 7 weekdays)
	if movedCount != 8 {
		t.Errorf("Expected 8 todos moved, got: %d", movedCount)
	}

	// Assert all earlier days are empty
	for day := Earlier; day < NextMonday; day++ {
		if len(plan.TodosByDay[day]) != 0 {
			t.Errorf("Expected %s to be empty, got %d todos", WeekDayNames[day], len(plan.TodosByDay[day]))
		}
	}

	// Assert NextMonday has all 8 todos
	if len(plan.TodosByDay[NextMonday]) != 8 {
		t.Errorf("Expected NextMonday to have 8 todos, got %d", len(plan.TodosByDay[NextMonday]))
	}
}

// TestBulkMoveEarlierTodosToCurrentDay_Undo tests undo functionality
func TestBulkMoveEarlierTodosToCurrentDay_Undo(t *testing.T) {
	tempDir, origDir := setupWeekPlannerTest(t)
	defer cleanupWeekPlannerTest(t, tempDir, origDir)

	// Create a week plan
	monday := time.Date(2024, 11, 18, 0, 0, 0, 0, time.UTC)
	plan := &WeekPlan{
		StartDate:  monday,
		TodosByDay: make(map[WeekDay][]scripts.File),
		Changes:    []PlanChange{},
		UndoStack:  []PlanChange{},
		RedoStack:  []PlanChange{},
	}

	// Add todos to Earlier and Monday
	earlierTodo1 := createTestTodo(t, "Earlier Task 1", monday.AddDate(0, 0, -5), scripts.P1)
	earlierTodo2 := createTestTodo(t, "Earlier Task 2", monday.AddDate(0, 0, -3), scripts.P2)
	mondayTodo1 := createTestTodo(t, "Monday Task 1", monday, scripts.P1)

	plan.TodosByDay[Earlier] = []scripts.File{earlierTodo1, earlierTodo2}
	plan.TodosByDay[Monday] = []scripts.File{mondayTodo1}

	// Create state and select Tuesday
	state := &WeekPlannerState{
		Plan:         plan,
		SelectedDay:  Tuesday,
		SelectedTodo: 0,
		ViewMode:     NormalView,
	}

	// Execute bulk move
	movedCount, err := state.BulkMoveEarlierTodosToCurrentDay()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify todos were moved
	if len(plan.TodosByDay[Tuesday]) != 3 {
		t.Errorf("Expected Tuesday to have 3 todos after bulk move, got %d", len(plan.TodosByDay[Tuesday]))
	}

	// Undo all moves
	for i := 0; i < movedCount; i++ {
		success := state.Undo()
		if !success {
			t.Errorf("Undo %d failed", i+1)
		}
	}

	// Assert todos are back in original positions
	if len(plan.TodosByDay[Earlier]) != 2 {
		t.Errorf("Expected Earlier to have 2 todos after undo, got %d", len(plan.TodosByDay[Earlier]))
	}
	if len(plan.TodosByDay[Monday]) != 1 {
		t.Errorf("Expected Monday to have 1 todo after undo, got %d", len(plan.TodosByDay[Monday]))
	}
	if len(plan.TodosByDay[Tuesday]) != 0 {
		t.Errorf("Expected Tuesday to be empty after undo, got %d", len(plan.TodosByDay[Tuesday]))
	}
}
