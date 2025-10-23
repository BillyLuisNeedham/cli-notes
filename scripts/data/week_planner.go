package data

import (
	"cli-notes/scripts"
	"time"
)

// WeekDay represents a day of the week
type WeekDay int

const (
	Monday WeekDay = iota
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
	NextMonday // Overflow bucket for todos moved to next week
)

// WeekDayNames maps WeekDay to string representations
var WeekDayNames = map[WeekDay]string{
	Monday:     "Monday",
	Tuesday:    "Tuesday",
	Wednesday:  "Wednesday",
	Thursday:   "Thursday",
	Friday:     "Friday",
	Saturday:   "Saturday",
	Sunday:     "Sunday",
	NextMonday: "Next Monday",
}

// WeekDayShortNames maps WeekDay to short string representations
var WeekDayShortNames = map[WeekDay]string{
	Monday:     "Mon",
	Tuesday:    "Tue",
	Wednesday:  "Wed",
	Thursday:   "Thu",
	Friday:     "Fri",
	Saturday:   "Sat",
	Sunday:     "Sun",
	NextMonday: "Next→",
}

// WeekPlan holds todos organized by day of the week
type WeekPlan struct {
	StartDate  time.Time                // Monday of the current week
	EndDate    time.Time                // Sunday of the current week
	TodosByDay map[WeekDay][]scripts.File // Todos organized by day
	Changes    []PlanChange             // History of changes made
	UndoStack  []PlanChange             // Stack for undo operations
	RedoStack  []PlanChange             // Stack for redo operations
}

// PlanChange represents a single change in the week plan
type PlanChange struct {
	Todo      scripts.File
	FromDay   WeekDay
	ToDay     WeekDay
	Timestamp time.Time
}

// NewWeekPlan creates a new week plan for the given start date (should be a Monday)
func NewWeekPlan(startDate time.Time) *WeekPlan {
	// Ensure start date is a Monday
	weekday := startDate.Weekday()
	daysToMonday := (int(time.Monday) - int(weekday) + 7) % 7
	if daysToMonday != 0 {
		startDate = startDate.AddDate(0, 0, -daysToMonday)
	}

	// Normalize to midnight in local time for consistent date comparisons
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 0, 6) // Sunday at midnight

	return &WeekPlan{
		StartDate:  startDate,
		EndDate:    endDate,
		TodosByDay: make(map[WeekDay][]scripts.File),
		Changes:    make([]PlanChange, 0),
		UndoStack:  make([]PlanChange, 0),
		RedoStack:  make([]PlanChange, 0),
	}
}

// GetDateForWeekDay returns the date for a given weekday in this week plan
// Returns the date normalized to midnight in local time
func (wp *WeekPlan) GetDateForWeekDay(day WeekDay) time.Time {
	var date time.Time
	if day == NextMonday {
		// Next Monday is 7 days after the end of current week
		date = wp.EndDate.AddDate(0, 0, 1)
	} else {
		date = wp.StartDate.AddDate(0, 0, int(day))
	}
	// Normalize to midnight in local time
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
}

// GetWeekDayForDate returns the WeekDay for a given date within the week
// Returns -1 if the date is not in this week
func (wp *WeekPlan) GetWeekDayForDate(date time.Time) WeekDay {
	// Normalize to midnight in local time for consistent comparison with StartDate/EndDate
	normalized := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)

	if normalized.Before(wp.StartDate) || normalized.After(wp.EndDate) {
		return -1
	}

	diff := int(normalized.Sub(wp.StartDate).Hours() / 24)
	return WeekDay(diff)
}

// LoadWeekTodos loads all todos with due dates in the current week
func LoadWeekTodos(startDate time.Time) (*WeekPlan, error) {
	plan := NewWeekPlan(startDate)

	// Query all incomplete todos
	allTodos, err := QueryFilesByDone(false)
	if err != nil {
		return nil, err
	}

	// Organize todos by day
	for _, todo := range allTodos {
		// Skip todos without a due date
		if todo.DueAt.Year() == 9999 {
			continue
		}

		// Check if todo falls within this week
		day := plan.GetWeekDayForDate(todo.DueAt)
		if day >= 0 {
			plan.TodosByDay[day] = append(plan.TodosByDay[day], todo)
		}
	}

	return plan, nil
}

// MoveTodo moves a todo from one day to another
func (wp *WeekPlan) MoveTodo(todo scripts.File, fromDay, toDay WeekDay) {
	// Record the change
	change := PlanChange{
		Todo:      todo,
		FromDay:   fromDay,
		ToDay:     toDay,
		Timestamp: time.Now(),
	}

	// Remove from source day
	wp.removeTodoFromDay(todo, fromDay)

	// Add to target day
	wp.TodosByDay[toDay] = append(wp.TodosByDay[toDay], todo)

	// Update the todo's due date
	newDueDate := wp.GetDateForWeekDay(toDay)
	for i := range wp.TodosByDay[toDay] {
		if wp.TodosByDay[toDay][i].Name == todo.Name {
			wp.TodosByDay[toDay][i].DueAt = newDueDate
			break
		}
	}

	// Add to change history and undo stack
	wp.Changes = append(wp.Changes, change)
	wp.UndoStack = append(wp.UndoStack, change)

	// Clear redo stack when a new action is performed
	wp.RedoStack = make([]PlanChange, 0)
}

// Undo reverses the last change
func (wp *WeekPlan) Undo() bool {
	if len(wp.UndoStack) == 0 {
		return false
	}

	// Pop from undo stack
	change := wp.UndoStack[len(wp.UndoStack)-1]
	wp.UndoStack = wp.UndoStack[:len(wp.UndoStack)-1]

	// Reverse the change: move from ToDay back to FromDay
	wp.removeTodoFromDay(change.Todo, change.ToDay)
	wp.TodosByDay[change.FromDay] = append(wp.TodosByDay[change.FromDay], change.Todo)

	// Restore the original due date
	originalDueDate := wp.GetDateForWeekDay(change.FromDay)
	for i := range wp.TodosByDay[change.FromDay] {
		if wp.TodosByDay[change.FromDay][i].Name == change.Todo.Name {
			wp.TodosByDay[change.FromDay][i].DueAt = originalDueDate
			break
		}
	}

	// Push to redo stack
	wp.RedoStack = append(wp.RedoStack, change)

	return true
}

// Redo reapplies the last undone change
func (wp *WeekPlan) Redo() bool {
	if len(wp.RedoStack) == 0 {
		return false
	}

	// Pop from redo stack
	change := wp.RedoStack[len(wp.RedoStack)-1]
	wp.RedoStack = wp.RedoStack[:len(wp.RedoStack)-1]

	// Reapply the change
	wp.removeTodoFromDay(change.Todo, change.FromDay)
	wp.TodosByDay[change.ToDay] = append(wp.TodosByDay[change.ToDay], change.Todo)

	// Update the due date
	newDueDate := wp.GetDateForWeekDay(change.ToDay)
	for i := range wp.TodosByDay[change.ToDay] {
		if wp.TodosByDay[change.ToDay][i].Name == change.Todo.Name {
			wp.TodosByDay[change.ToDay][i].DueAt = newDueDate
			break
		}
	}

	// Push back to undo stack
	wp.UndoStack = append(wp.UndoStack, change)

	return true
}

// Reset reloads the plan from disk, discarding all changes
func (wp *WeekPlan) Reset() error {
	freshPlan, err := LoadWeekTodos(wp.StartDate)
	if err != nil {
		return err
	}

	*wp = *freshPlan
	return nil
}

// SaveChanges writes all modified todos back to disk
func (wp *WeekPlan) SaveChanges() error {
	// Collect all todos in the current week (these may have updated due dates)
	modifiedTodos := make(map[string]scripts.File)

	for _, todos := range wp.TodosByDay {
		for _, todo := range todos {
			modifiedTodos[todo.Name] = todo
		}
	}

	// Write each todo
	for _, todo := range modifiedTodos {
		if err := WriteFile(todo); err != nil {
			return err
		}
	}

	// Clear change history after successful save
	wp.Changes = make([]PlanChange, 0)
	wp.UndoStack = make([]PlanChange, 0)
	wp.RedoStack = make([]PlanChange, 0)

	return nil
}

// GetTodoCount returns the number of todos for a given day
func (wp *WeekPlan) GetTodoCount(day WeekDay) int {
	return len(wp.TodosByDay[day])
}

// HasChanges returns true if there are unsaved changes
func (wp *WeekPlan) HasChanges() bool {
	return len(wp.Changes) > 0
}

// removeTodoFromDay is a helper that removes a todo from a specific day
func (wp *WeekPlan) removeTodoFromDay(todo scripts.File, day WeekDay) {
	todos := wp.TodosByDay[day]
	for i, t := range todos {
		if t.Name == todo.Name {
			wp.TodosByDay[day] = append(todos[:i], todos[i+1:]...)
			break
		}
	}
}
