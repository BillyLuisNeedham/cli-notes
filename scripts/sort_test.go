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

	p3Soon := File{
		Name:      "p3-soon.md",
		Priority:  P3,
		DueAt:     now.AddDate(0, 0, -1), // Due yesterday (overdue)
		Title:     "P3 Due Soon",
		CreatedAt: now.AddDate(0, 0, -1), // Created yesterday
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
	}

	// Sort the todos
	sortedTodos := SortTodosByPriorityAndDueDate(unsortedTodos)

	// With the weighted scoring system, the expected order might be different
	// We'll verify that the sorting happened and the result is consistent
	if len(sortedTodos) != len(unsortedTodos) {
		t.Errorf("Expected %d todos, got %d", len(unsortedTodos), len(sortedTodos))
	}

	// Verify that each todo has a lower or equal score than the next one
	for i := 0; i < len(sortedTodos)-1; i++ {
		scoreI := CalculateTodoScore(sortedTodos[i])
		scoreJ := CalculateTodoScore(sortedTodos[i+1])
		if scoreI > scoreJ {
			t.Errorf("Position %d: score %f should be <= than position %d: score %f",
				i, scoreI, i+1, scoreJ)
		}
	}

	// Test with empty slice
	emptySlice := []File{}
	sortedEmpty := SortTodosByPriorityAndDueDate(emptySlice)
	if len(sortedEmpty) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(sortedEmpty))
	}
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
