package scripts

import (
	"testing"
	"time"
)

func TestSortTodosByPriorityAndDueDate(t *testing.T) {
	// Create test data with varying priorities and due dates
	now := time.Now()

	p1Soon := File{
		Name:      "p1-soon.md",
		Priority:  P1,
		DueAt:     now.AddDate(0, 0, 1), // Due tomorrow
		Title:     "P1 Due Soon",
		CreatedAt: now.AddDate(0, 0, -5), // Created 5 days ago
	}

	p1Later := File{
		Name:      "p1-later.md",
		Priority:  P1,
		DueAt:     now.AddDate(0, 0, 5), // Due in 5 days
		Title:     "P1 Due Later",
		CreatedAt: now.AddDate(0, 0, -10), // Created 10 days ago
	}

	p1Overdue := File{
		Name:      "p1-overdue.md",
		Priority:  P1,
		DueAt:     now.AddDate(0, 0, -2), // Due 2 days ago (overdue)
		Title:     "P1 Overdue",
		CreatedAt: now.AddDate(0, 0, -3), // Created 3 days ago
	}

	p1OverdueLongAgo := File{
		Name:      "p1-overdue-long-ago.md",
		Priority:  P1,
		DueAt:     now.AddDate(0, 0, -2), // Due 2 days ago (same as p1Overdue)
		Title:     "P1 Overdue Created Earlier",
		CreatedAt: now.AddDate(0, 0, -20), // Created 20 days ago (older than p1Overdue)
	}

	p2Soon := File{
		Name:      "p2-soon.md",
		Priority:  P2,
		DueAt:     now, // Due today
		Title:     "P2 Due Soon",
		CreatedAt: now.AddDate(0, 0, -2), // Created 2 days ago
	}

	p2Later := File{
		Name:      "p2-later.md",
		Priority:  P2,
		DueAt:     now.AddDate(0, 0, 10), // Due in 10 days
		Title:     "P2 Due Later",
		CreatedAt: now.AddDate(0, 0, -15), // Created 15 days ago
	}

	p2Overdue := File{
		Name:      "p2-overdue.md",
		Priority:  P2,
		DueAt:     now.AddDate(0, 0, -1), // Due yesterday (overdue)
		Title:     "P2 Overdue",
		CreatedAt: now.AddDate(0, 0, -5), // Created 5 days ago
	}

	p3Soon := File{
		Name:      "p3-soon.md",
		Priority:  P3,
		DueAt:     now.AddDate(0, 0, 2), // Due in 2 days
		Title:     "P3 Due Soon",
		CreatedAt: now.AddDate(0, 0, -1), // Created yesterday
	}

	p3Overdue := File{
		Name:      "p3-overdue.md",
		Priority:  P3,
		DueAt:     now.AddDate(0, 0, -3), // Due 3 days ago (overdue)
		Title:     "P3 Overdue",
		CreatedAt: now.AddDate(0, 0, -7), // Created 7 days ago
	}

	p3NoDueDate := File{
		Name:      "p3-no-due-date.md",
		Priority:  P3,
		DueAt:     now.AddDate(101, 0, 0), // Far future date (represents no due date)
		Title:     "P3 No Due Date",
		CreatedAt: now.AddDate(0, 0, -20), // Created 20 days ago
	}

	// Input is unsorted
	unsortedTodos := []File{
		p3NoDueDate,
		p2Later,
		p1Soon,
		p3Soon,
		p1Later,
		p2Soon,
		p3Overdue,
		p1Overdue,
		p2Overdue,
		p1OverdueLongAgo,
	}

	// Sort the todos
	sortedTodos := SortTodosByPriorityAndDueDate(unsortedTodos)

	// Test that the correct number of todos were returned
	if len(sortedTodos) != len(unsortedTodos) {
		t.Errorf("Expected %d todos, got %d", len(unsortedTodos), len(sortedTodos))
	}

	// Test that P1 tasks come before P2 tasks which come before P3 tasks
	lastPriority := P1
	for i, todo := range sortedTodos {
		if todo.Priority < lastPriority {
			t.Errorf("Position %d: Priority out of order. Found P%d after P%d",
				i, todo.Priority, lastPriority)
		}
		lastPriority = todo.Priority
	}

	// Group todos by priority for more detailed checks
	p1Todos := getFilesByPriority(sortedTodos, P1)
	p2Todos := getFilesByPriority(sortedTodos, P2)
	p3Todos := getFilesByPriority(sortedTodos, P3)

	// Test that within each priority group, overdue tasks come first
	testPriorityGroupOrder(t, p1Todos, "P1")
	testPriorityGroupOrder(t, p2Todos, "P2")
	testPriorityGroupOrder(t, p3Todos, "P3")

	// Test specific ordering cases
	// Check that p1OverdueLongAgo comes before p1Overdue (same due date, older creation date)
	foundP1OverdueLongAgo := false
	foundP1Overdue := false
	p1OverdueLongAgoIndex := -1
	p1OverdueIndex := -1

	for i, todo := range sortedTodos {
		if todo.Name == p1OverdueLongAgo.Name {
			foundP1OverdueLongAgo = true
			p1OverdueLongAgoIndex = i
		}
		if todo.Name == p1Overdue.Name {
			foundP1Overdue = true
			p1OverdueIndex = i
		}
	}

	if foundP1OverdueLongAgo && foundP1Overdue && p1OverdueLongAgoIndex > p1OverdueIndex {
		t.Errorf("Expected p1-overdue-long-ago.md (created %v) to come before p1-overdue.md (created %v) due to older creation date",
			p1OverdueLongAgo.CreatedAt.Format("2006-01-02"),
			p1Overdue.CreatedAt.Format("2006-01-02"))
	}

	// Test with empty slice
	emptySlice := []File{}
	sortedEmpty := SortTodosByPriorityAndDueDate(emptySlice)
	if len(sortedEmpty) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(sortedEmpty))
	}
}

// Helper function to test the order within a priority group
func testPriorityGroupOrder(t *testing.T, todos []File, priorityName string) {
	now := time.Now()
	var lastIsOverdue bool
	var lastTodo File
	var lastDueDate time.Time

	for i, todo := range todos {
		isOverdue := todo.DueAt.Before(now)
		currentDueDate := todo.DueAt
		hasNoDueDate := currentDueDate.Year() > 2100
		lastHasNoDueDate := lastDueDate.Year() > 2100

		if i > 0 {
			// If previous todo was not overdue but current is, that's wrong ordering
			if !lastIsOverdue && isOverdue {
				t.Errorf("%s group: Found overdue task after non-overdue task", priorityName)
				t.Logf("Current task: %s (overdue), Previous task: %s (not overdue)",
					todo.Name, lastTodo.Name)
			}

			// If both have same overdue status
			if lastIsOverdue == isOverdue {
				// If current has no due date but previous has a due date, this is correct
				// If current has a due date but previous has no due date, this is wrong
				if !hasNoDueDate && lastHasNoDueDate {
					t.Errorf("%s group: Task with due date comes after task with no due date",
						priorityName)
					t.Logf("Current task: %s (has due date: %v)",
						todo.Name, currentDueDate.Format("2006-01-02"))
					t.Logf("Previous task: %s (no due date)", lastTodo.Name)
				} else if hasNoDueDate == lastHasNoDueDate {
					// If both have due dates or both don't have due dates
					// If due dates are effectively the same (or both no due date)
					if currentDueDate.Equal(lastDueDate) || (hasNoDueDate && lastHasNoDueDate) {
						// Check creation date ordering - older should come first
						if !todo.CreatedAt.IsZero() && !lastTodo.CreatedAt.IsZero() {
							if todo.CreatedAt.Before(lastTodo.CreatedAt) {
								t.Errorf("%s group: Tasks with same due date not ordered by creation date (oldest first)",
									priorityName)
								t.Logf("Current task: %s, Created: %v should be before",
									todo.Name, todo.CreatedAt.Format("2006-01-02"))
								t.Logf("Previous task: %s, Created: %v",
									lastTodo.Name, lastTodo.CreatedAt.Format("2006-01-02"))
							}
						}
					} else if !hasNoDueDate && !lastHasNoDueDate && lastDueDate.After(currentDueDate) {
						// Check due date ordering - earlier due date should come first
						t.Errorf("%s group: Tasks not ordered by due date (soonest first)", priorityName)
						t.Logf("Current task: %s, Due: %v should be before",
							todo.Name, currentDueDate.Format("2006-01-02"))
						t.Logf("Previous task: %s, Due: %v",
							lastTodo.Name, lastDueDate.Format("2006-01-02"))
					}
				}
			}
		}

		lastIsOverdue = isOverdue
		lastTodo = todo
		lastDueDate = currentDueDate
	}
}

// Helper function to filter todos by priority
func getFilesByPriority(todos []File, priority Priority) []File {
	result := []File{}
	for _, todo := range todos {
		if todo.Priority == priority {
			result = append(result, todo)
		}
	}
	return result
}

// TestCalculateTodoScore tests the score calculation function
func TestCalculateTodoScore(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name     string
		todo     File
		expected float64
		delta    float64 // Allowed difference due to floating point calculations
	}{
		{
			name: "High priority, due soon, created recently",
			todo: File{
				Priority:  P1,
				DueAt:     now.AddDate(0, 0, 1),  // Due tomorrow
				CreatedAt: now.AddDate(0, 0, -1), // Created yesterday
			},
			expected: 1.0*0.3 + 1.0*0.5 + 1.0*0.2, // Priority=1, DaysUntilDue=1, DaysSinceCreation=1
			delta:    0.1,
		},
		{
			name: "Low priority, overdue, created long ago",
			todo: File{
				Priority:  P3,
				DueAt:     now.AddDate(0, 0, -5),  // 5 days overdue
				CreatedAt: now.AddDate(0, 0, -30), // Created 30 days ago
			},
			expected: 3.0*0.3 + (-5.0)*0.5 + 30.0*0.2, // Priority=3, DaysUntilDue=-5, DaysSinceCreation=30
			delta:    0.1,
		},
		{
			name: "Medium priority, due far future, no creation date",
			todo: File{
				Priority:  P2,
				DueAt:     now.AddDate(0, 0, 20), // Due in 20 days
				CreatedAt: time.Time{},           // No creation date
			},
			expected: 2.0*0.3 + 20.0*0.5 + 0.0*0.2, // Priority=2, DaysUntilDue=20, DaysSinceCreation=0
			delta:    0.1,
		},
		{
			name: "Low priority, no due date, created long ago",
			todo: File{
				Priority:  P3,
				DueAt:     now.AddDate(101, 0, 0), // No due date (far future)
				CreatedAt: now.AddDate(0, 0, -60), // Created 60 days ago
			},
			expected: 3.0*0.3 + 365.0*0.5 + 60.0*0.2, // Priority=3, DaysUntilDue=365, DaysSinceCreation=60
			delta:    0.1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := CalculateTodoScore(tc.todo)
			diff := score - tc.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > tc.delta {
				t.Errorf("Expected score around %f, got %f (diff: %f)", tc.expected, score, diff)
			}
		})
	}
}
