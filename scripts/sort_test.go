package scripts

import (
	"testing"
	"time"
)

func TestSortTodosByPriorityAndDueDate(t *testing.T) {
	// Create test data with varying priorities and due dates
	now := time.Now()

	p1Soon := File{
		Name:     "p1-soon.md",
		Priority: P1,
		DueAt:    now.AddDate(0, 0, 1), // Due tomorrow
		Title:    "P1 Due Soon",
	}

	p1Later := File{
		Name:     "p1-later.md",
		Priority: P1,
		DueAt:    now.AddDate(0, 0, 5), // Due in 5 days
		Title:    "P1 Due Later",
	}

	p2Soon := File{
		Name:     "p2-soon.md",
		Priority: P2,
		DueAt:    now, // Due today
		Title:    "P2 Due Soon",
	}

	p2Later := File{
		Name:     "p2-later.md",
		Priority: P2,
		DueAt:    now.AddDate(0, 0, 10), // Due in 10 days
		Title:    "P2 Due Later",
	}

	p3Soon := File{
		Name:     "p3-soon.md",
		Priority: P3,
		DueAt:    now.AddDate(0, 0, -1), // Due yesterday (overdue)
		Title:    "P3 Due Soon",
	}

	p3NoDueDate := File{
		Name:     "p3-no-due-date.md",
		Priority: P3,
		DueAt:    now.AddDate(101, 0, 0), // Far future date (represents no due date)
		Title:    "P3 No Due Date",
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

	// Expected order: sorted first by priority, then by due date
	expectedOrder := []File{
		p1Soon,      // P1 due tomorrow
		p1Later,     // P1 due in 5 days
		p2Soon,      // P2 due today
		p2Later,     // P2 due in 10 days
		p3Soon,      // P3 due yesterday
		p3NoDueDate, // P3 no due date (far future)
	}

	// Sort the todos
	sortedTodos := SortTodosByPriorityAndDueDate(unsortedTodos)

	// Verify the sorting order
	if len(sortedTodos) != len(expectedOrder) {
		t.Errorf("Expected %d todos, got %d", len(expectedOrder), len(sortedTodos))
	}

	for i, expected := range expectedOrder {
		if sortedTodos[i].Name != expected.Name {
			t.Errorf("Position %d: expected %s, got %s", i, expected.Name, sortedTodos[i].Name)
		}
	}

	// Test with empty slice
	emptySlice := []File{}
	sortedEmpty := SortTodosByPriorityAndDueDate(emptySlice)
	if len(sortedEmpty) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(sortedEmpty))
	}
}
