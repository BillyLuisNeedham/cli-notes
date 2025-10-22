package data

import (
	"cli-notes/scripts"
	"fmt"
	"time"
)

// WeekPlannerState holds the current state of the week planner
type WeekPlannerState struct {
	Plan         *WeekPlan
	SelectedDay  WeekDay
	SelectedTodo int // Index of selected todo within the day
}

// NewWeekPlannerState creates a new week planner state for the current week
func NewWeekPlannerState() (*WeekPlannerState, error) {
	// Get current date and find this week's Monday
	now := time.Now()
	plan, err := LoadWeekTodos(now)
	if err != nil {
		return nil, err
	}

	// Default to current day of week
	currentDay := getWeekDayFromTime(now)

	return &WeekPlannerState{
		Plan:         plan,
		SelectedDay:  currentDay,
		SelectedTodo: 0,
	}, nil
}

// MoveSelectedTodoLeft moves the selected todo to the previous day
func (wps *WeekPlannerState) MoveSelectedTodoLeft() error {
	if wps.SelectedDay == Monday {
		return fmt.Errorf("cannot move left from Monday")
	}

	return wps.moveSelectedTodo(wps.SelectedDay - 1)
}

// MoveSelectedTodoRight moves the selected todo to the next day
func (wps *WeekPlannerState) MoveSelectedTodoRight() error {
	if wps.SelectedDay == NextMonday {
		return fmt.Errorf("cannot move right from Next Monday")
	}

	return wps.moveSelectedTodo(wps.SelectedDay + 1)
}

// MoveSelectedTodoToNextMonday moves the selected todo to the overflow (next Monday)
func (wps *WeekPlannerState) MoveSelectedTodoToNextMonday() error {
	return wps.moveSelectedTodo(NextMonday)
}

// moveSelectedTodo is a helper that moves the selected todo to a target day
func (wps *WeekPlannerState) moveSelectedTodo(targetDay WeekDay) error {
	todos := wps.Plan.TodosByDay[wps.SelectedDay]

	if len(todos) == 0 {
		return fmt.Errorf("no todos on current day")
	}

	if wps.SelectedTodo < 0 || wps.SelectedTodo >= len(todos) {
		return fmt.Errorf("invalid todo selection")
	}

	todo := todos[wps.SelectedTodo]
	wps.Plan.MoveTodo(todo, wps.SelectedDay, targetDay)

	// Adjust selection after move
	wps.adjustSelectionAfterMove()

	return nil
}

// adjustSelectionAfterMove adjusts the selected todo index after a move
func (wps *WeekPlannerState) adjustSelectionAfterMove() {
	todosRemaining := len(wps.Plan.TodosByDay[wps.SelectedDay])

	if todosRemaining == 0 {
		wps.SelectedTodo = 0
	} else if wps.SelectedTodo >= todosRemaining {
		wps.SelectedTodo = todosRemaining - 1
	}
}

// SelectNextTodo moves selection to the next todo on the current day
func (wps *WeekPlannerState) SelectNextTodo() {
	todos := wps.Plan.TodosByDay[wps.SelectedDay]
	if len(todos) == 0 {
		wps.SelectedTodo = 0
		return
	}

	wps.SelectedTodo++
	if wps.SelectedTodo >= len(todos) {
		wps.SelectedTodo = 0 // Wrap around
	}
}

// SelectPreviousTodo moves selection to the previous todo on the current day
func (wps *WeekPlannerState) SelectPreviousTodo() {
	todos := wps.Plan.TodosByDay[wps.SelectedDay]
	if len(todos) == 0 {
		wps.SelectedTodo = 0
		return
	}

	wps.SelectedTodo--
	if wps.SelectedTodo < 0 {
		wps.SelectedTodo = len(todos) - 1 // Wrap around
	}
}

// SwitchToDay changes the selected day and resets todo selection
func (wps *WeekPlannerState) SwitchToDay(day WeekDay) {
	wps.SelectedDay = day
	wps.SelectedTodo = 0
}

// SwitchToNextDay moves to the next day in the week
func (wps *WeekPlannerState) SwitchToNextDay() {
	if wps.SelectedDay < NextMonday {
		wps.SelectedDay++
		wps.SelectedTodo = 0
	}
}

// SwitchToPreviousDay moves to the previous day in the week
func (wps *WeekPlannerState) SwitchToPreviousDay() {
	if wps.SelectedDay > Monday {
		wps.SelectedDay--
		wps.SelectedTodo = 0
	}
}

// Undo reverses the last action
func (wps *WeekPlannerState) Undo() bool {
	success := wps.Plan.Undo()
	if success {
		wps.adjustSelectionAfterMove()
	}
	return success
}

// Redo reapplies the last undone action
func (wps *WeekPlannerState) Redo() bool {
	success := wps.Plan.Redo()
	if success {
		wps.adjustSelectionAfterMove()
	}
	return success
}

// Reset reloads the plan from disk
func (wps *WeekPlannerState) Reset() error {
	err := wps.Plan.Reset()
	if err == nil {
		wps.SelectedTodo = 0
	}
	return err
}

// Save writes all changes to disk
func (wps *WeekPlannerState) Save() error {
	return wps.Plan.SaveChanges()
}

// GetSelectedTodo returns the currently selected todo, or nil if none
func (wps *WeekPlannerState) GetSelectedTodo() *scripts.File {
	todos := wps.Plan.TodosByDay[wps.SelectedDay]
	if len(todos) == 0 || wps.SelectedTodo < 0 || wps.SelectedTodo >= len(todos) {
		return nil
	}
	return &todos[wps.SelectedTodo]
}

// GetChangeSummary returns a formatted summary of all changes
func (wps *WeekPlannerState) GetChangeSummary() []string {
	summary := make([]string, 0)
	for _, change := range wps.Plan.Changes {
		fromDay := WeekDayShortNames[change.FromDay]
		toDay := WeekDayShortNames[change.ToDay]
		summary = append(summary, fmt.Sprintf("Moved from %s: \"%s\"", fromDay, change.Todo.Title))
		summary = append(summary, fmt.Sprintf("Moved to %s: \"%s\"", toDay, change.Todo.Title))
	}
	return summary
}

// getWeekDayFromTime converts a time.Time to a WeekDay
func getWeekDayFromTime(t time.Time) WeekDay {
	switch t.Weekday() {
	case time.Monday:
		return Monday
	case time.Tuesday:
		return Tuesday
	case time.Wednesday:
		return Wednesday
	case time.Thursday:
		return Thursday
	case time.Friday:
		return Friday
	case time.Saturday:
		return Saturday
	case time.Sunday:
		return Sunday
	default:
		return Monday
	}
}
