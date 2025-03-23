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

// SortTodosByPriorityAndDueDate sorts a slice of File (todos) according to the following rules:
//  1. Priority order: P1 > P2 > P3 (P1 tasks always come first)
//  2. Within each priority group:
//     a. Overdue tasks come first
//     b. Then tasks sorted by due date (soonest first)
//     c. Tasks with no due date come last within their priority group
//     d. When tasks have the same due date, sort by creation date (oldest first)
//
// It modifies the slice in place and also returns it for convenience.
func SortTodosByPriorityAndDueDate(todos []File) []File {
	now := time.Now()

	// Custom sorting function that follows the priority rules
	sort.Slice(todos, func(i, j int) bool {
		// First, sort by priority (P1 < P2 < P3)
		if todos[i].Priority != todos[j].Priority {
			return todos[i].Priority < todos[j].Priority
		}

		// Within the same priority, check for no due date
		iNoDueDate := todos[i].DueAt.Year() > 2100
		jNoDueDate := todos[j].DueAt.Year() > 2100

		// If one has no due date and the other does, the task with a due date comes first
		if iNoDueDate != jNoDueDate {
			return !iNoDueDate // Return true if i has a due date and j doesn't
		}

		// Both have due dates or both don't have due dates
		// Check overdue status
		iOverdue := todos[i].DueAt.Before(now)
		jOverdue := todos[j].DueAt.Before(now)

		// If one is overdue and the other isn't, the overdue task comes first
		if iOverdue != jOverdue {
			return iOverdue
		}

		// If both have the same overdue status and both have due dates, sort by due date
		if !iNoDueDate && !jNoDueDate && !todos[i].DueAt.Equal(todos[j].DueAt) {
			// Sort by due date (earlier comes first)
			return todos[i].DueAt.Before(todos[j].DueAt)
		}

		// If due dates are equal or both have no due date, sort by creation date
		// Older creation dates (tasks created longer ago) come first
		iHasCreationDate := !todos[i].CreatedAt.IsZero()
		jHasCreationDate := !todos[j].CreatedAt.IsZero()

		// If both have creation dates, older comes first
		if iHasCreationDate && jHasCreationDate {
			// Note: Tasks created earlier (older) should come before those created later
			return todos[i].CreatedAt.Before(todos[j].CreatedAt)
		}

		// If only one has a creation date, it comes first
		if iHasCreationDate {
			return true
		}
		if jHasCreationDate {
			return false
		}

		// If neither has a creation date, keep original order
		return i < j
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
