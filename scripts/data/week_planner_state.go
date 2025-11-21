package data

import (
	"cli-notes/scripts"
	"fmt"
	"time"
)

// ViewMode represents the current view mode of the week planner
type ViewMode int

const (
	NormalView ViewMode = iota
	ExpandedEarlierView
)

// WeekPlannerState holds the current state of the week planner
type WeekPlannerState struct {
	Plan         *WeekPlan
	SelectedDay  WeekDay
	SelectedTodo int      // Index of selected todo within the day
	ViewMode     ViewMode // Current view mode
	ScrollOffset int      // Scroll offset for expanded view
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
		ViewMode:     NormalView,
		ScrollOffset: 0,
	}, nil
}

// MoveSelectedTodoLeft moves the selected todo to the previous day
func (wps *WeekPlannerState) MoveSelectedTodoLeft() error {
	if wps.SelectedDay == Earlier {
		return fmt.Errorf("cannot move tasks from Earlier (read-only)")
	}
	if wps.SelectedDay == Monday {
		return fmt.Errorf("cannot move left from Monday")
	}

	return wps.moveSelectedTodo(wps.SelectedDay - 1)
}

// MoveSelectedTodoRight moves the selected todo to the next day
func (wps *WeekPlannerState) MoveSelectedTodoRight() error {
	if wps.SelectedDay == Earlier {
		return fmt.Errorf("cannot move tasks from Earlier (read-only)")
	}
	if wps.SelectedDay == NextMonday {
		return fmt.Errorf("cannot move right from Next Monday")
	}

	return wps.moveSelectedTodo(wps.SelectedDay + 1)
}

// MoveSelectedTodoToNextMonday moves the selected todo to the overflow (next Monday)
func (wps *WeekPlannerState) MoveSelectedTodoToNextMonday() error {
	if wps.SelectedDay == Earlier {
		return fmt.Errorf("cannot move tasks from Earlier (read-only)")
	}
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
	wps.AdjustSelectionAfterMove()

	return nil
}

// AdjustSelectionAfterMove adjusts the selected todo index after a move
func (wps *WeekPlannerState) AdjustSelectionAfterMove() {
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
	if wps.SelectedDay > Earlier {
		wps.SelectedDay--
		wps.SelectedTodo = 0
	}
}

// Undo reverses the last action
func (wps *WeekPlannerState) Undo() bool {
	success := wps.Plan.Undo()
	if success {
		wps.AdjustSelectionAfterMove()
	}
	return success
}

// Redo reapplies the last undone action
func (wps *WeekPlannerState) Redo() bool {
	success := wps.Plan.Redo()
	if success {
		wps.AdjustSelectionAfterMove()
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

// RefreshOpenedTodo reloads a single todo file from disk while preserving
// all unsaved changes in the weekly planner. This is used after opening
// a note in an external editor to pick up any edits without losing moves.
func (wps *WeekPlannerState) RefreshOpenedTodo(fileName string) error {
	err := wps.Plan.RefreshTodo(fileName)
	if err != nil {
		return err
	}

	// Adjust selection if the refreshed todo is no longer on the current day
	// (e.g., if it was deleted from disk)
	todos := wps.Plan.TodosByDay[wps.SelectedDay]
	if wps.SelectedTodo >= len(todos) {
		if len(todos) > 0 {
			wps.SelectedTodo = len(todos) - 1
		} else {
			wps.SelectedTodo = 0
		}
	}

	return nil
}

// Save writes all changes to disk
func (wps *WeekPlannerState) Save() error {
	return wps.Plan.SaveChanges()
}

// NavigateToPreviousWeek loads the week plan for the previous week
func (wps *WeekPlannerState) NavigateToPreviousWeek() error {
	// Move back 7 days to get to previous week
	previousWeekStart := wps.Plan.StartDate.AddDate(0, 0, -7)
	plan, err := LoadWeekTodos(previousWeekStart)
	if err != nil {
		return err
	}

	wps.Plan = plan
	wps.SelectedTodo = 0
	// Keep the same day of week selected if possible
	return nil
}

// NavigateToNextWeek loads the week plan for the next week
func (wps *WeekPlannerState) NavigateToNextWeek() error {
	// Move forward 7 days to get to next week
	nextWeekStart := wps.Plan.StartDate.AddDate(0, 0, 7)
	plan, err := LoadWeekTodos(nextWeekStart)
	if err != nil {
		return err
	}

	wps.Plan = plan
	wps.SelectedTodo = 0
	// Keep the same day of week selected if possible
	return nil
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

// EnterExpandedEarlierView switches to expanded Earlier view
func (wps *WeekPlannerState) EnterExpandedEarlierView() {
	wps.ViewMode = ExpandedEarlierView
	wps.SelectedDay = Earlier
	wps.ScrollOffset = 0
	// Preserve SelectedTodo index
}

// ExitExpandedEarlierView returns to normal view
func (wps *WeekPlannerState) ExitExpandedEarlierView() {
	wps.ViewMode = NormalView
	wps.ScrollOffset = 0
	// Preserve SelectedTodo and SelectedDay
}

// AdjustScrollOffset updates scroll offset based on selected todo and viewport
func (wps *WeekPlannerState) AdjustScrollOffset(viewportHeight int) {
	todos := wps.Plan.TodosByDay[Earlier]
	if len(todos) == 0 {
		wps.ScrollOffset = 0
		return
	}

	// Ensure selected todo is visible
	if wps.SelectedTodo < wps.ScrollOffset {
		wps.ScrollOffset = wps.SelectedTodo
	} else if wps.SelectedTodo >= wps.ScrollOffset+viewportHeight {
		wps.ScrollOffset = wps.SelectedTodo - viewportHeight + 1
	}

	// Ensure scroll offset is valid
	maxOffset := len(todos) - viewportHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if wps.ScrollOffset > maxOffset {
		wps.ScrollOffset = maxOffset
	}
	if wps.ScrollOffset < 0 {
		wps.ScrollOffset = 0
	}
}

// MoveEarlierTodoToMonday moves a todo from Earlier to Monday (only in expanded view)
func (wps *WeekPlannerState) MoveEarlierTodoToMonday() error {
	if wps.SelectedDay != Earlier {
		return fmt.Errorf("not on Earlier day")
	}
	if wps.ViewMode != ExpandedEarlierView {
		return fmt.Errorf("can only move Earlier todos in expanded view")
	}

	return wps.moveSelectedTodo(Monday)
}

// BulkMoveEarlierTodosToCurrentDay moves all todos from earlier days to the selected day
// Returns the number of todos moved, or an error
func (wps *WeekPlannerState) BulkMoveEarlierTodosToCurrentDay() (int, error) {
	targetDay := wps.SelectedDay

	// Validation
	if targetDay == Earlier {
		return 0, fmt.Errorf("cannot bulk move to Earlier")
	}

	movedCount := 0

	// Iterate through all days before the target
	for day := Earlier; day < targetDay; day++ {
		// Get snapshot of todos (to avoid modification during iteration)
		todos := make([]scripts.File, len(wps.Plan.TodosByDay[day]))
		copy(todos, wps.Plan.TodosByDay[day])

		// Move all todos from this day
		for _, todo := range todos {
			wps.Plan.MoveTodo(todo, day, targetDay)
			movedCount++
		}
	}

	if movedCount == 0 {
		return 0, fmt.Errorf("no earlier todos to move")
	}

	// Reset selected todo to 0 (moved todos added to end of target day)
	wps.SelectedTodo = 0

	return movedCount, nil
}
