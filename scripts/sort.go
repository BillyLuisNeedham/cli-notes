package scripts

import (
	"sort"
	"time"
)

// CalculateTodoScore calculates a weighted score for a todo based on:
// - Days until due (50% weight)
// - Days since creation (20% weight)
// - Manual priority (30% weight)
// Lower scores indicate higher priority
func CalculateTodoScore(todo File) float64 {
	now := time.Now()

	// Calculate days until due
	daysUntilDue := 0.0
	if todo.DueAt.Year() > 2100 {
		// No due date, assign a high value
		daysUntilDue = 365.0
	} else {
		// Calculate days until due (can be negative for overdue tasks)
		daysUntilDue = todo.DueAt.Sub(now).Hours() / 24.0
	}

	// Calculate days since creation
	daysSinceCreation := 0.0
	if !todo.CreatedAt.IsZero() {
		daysSinceCreation = now.Sub(todo.CreatedAt).Hours() / 24.0
	}

	// Convert priority to a numeric value (P1=1, P2=2, P3=3)
	priorityValue := float64(todo.Priority)

	// Calculate the weighted score
	// Note: We're using the raw priority value (lower is higher priority)
	// For days until due, lower values (or negative for overdue) mean higher priority
	// For days since creation, higher values mean the task has been waiting longer
	score := (daysUntilDue * 0.5) + (daysSinceCreation * 0.2) + (priorityValue * 0.3)

	return score
}

// SortTodosByPriorityAndDueDate sorts a slice of File (todos) using a weighted scoring system
// that considers days until due date (50%), days since creation (20%), and manual priority (30%).
// It modifies the slice in place and also returns it for convenience.
func SortTodosByPriorityAndDueDate(todos []File) []File {
	sort.Slice(todos, func(i, j int) bool {
		scoreI := CalculateTodoScore(todos[i])
		scoreJ := CalculateTodoScore(todos[j])

		// Lower scores come first (higher priority)
		return scoreI < scoreJ
	})

	return todos
}

// SortTodosByDueDate sorts todos by due date with overdue first,
// then by ascending due date, with no due date last
func SortTodosByDueDate(todos []File) []File {
	now := time.Now()

	sort.Slice(todos, func(i, j int) bool {
		// Check if todo i is overdue but j is not
		iOverdue := !todos[i].DueAt.IsZero() && todos[i].DueAt.Before(now)
		jOverdue := !todos[j].DueAt.IsZero() && todos[j].DueAt.Before(now)

		// Both overdue or both not overdue, sort by due date
		if iOverdue == jOverdue {
			// If i has no due date, it comes after j
			if todos[i].DueAt.Year() > 2100 {
				return false
			}
			// If j has no due date, i comes before j
			if todos[j].DueAt.Year() > 2100 {
				return true
			}
			// Both have due dates, sort by date
			return todos[i].DueAt.Before(todos[j].DueAt)
		}

		// Overdue comes first
		return iOverdue && !jOverdue
	})

	return todos
}
