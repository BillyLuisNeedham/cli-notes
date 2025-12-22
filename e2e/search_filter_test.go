package e2e

import (
	"strings"
	"testing"
)

func TestSearchFilter(t *testing.T) {
	h := NewTestHarness(t)

	// Setup test data: 2 incomplete and 2 complete notes
	h.CreateTodo("incomplete1.md", "Incomplete Note One", []string{}, Today(), false, 1)
	h.CreateTodo("incomplete2.md", "Incomplete Note Two", []string{}, Today(), false, 2)
	h.CreateTodo("complete1.md", "Complete Note One", []string{}, Today(), true, 1)
	h.CreateTodo("complete2.md", "Complete Note Two", []string{}, Today(), true, 2)

	t.Run("Default filter shows incomplete notes only", func(t *testing.T) {
		// gs opens search (insert mode), Enter to normal mode, q to quit
		// Using Enter instead of Escape because of input buffering issue
		stdout, _, err := h.RunCommand("gs\n\nq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// Should show "Open" filter label (default)
		if !strings.Contains(stdout, "Open") {
			t.Errorf("Expected 'Open' filter label in output, got: %s", stdout)
		}

		// Should show 2 matches (only incomplete notes)
		if !strings.Contains(stdout, "2 matches") {
			t.Errorf("Expected '2 matches' for incomplete only, got: %s", stdout)
		}
	})

	t.Run("Cycle to ShowCompleteOnly", func(t *testing.T) {
		// gs opens search (insert mode), Enter to normal mode, f cycles filter, q to quit
		// Default -> ShowCompleteOnly (first cycle)
		stdout, _, err := h.RunCommand("gs\n\nfq\n")
		if err != nil {
			t.Fatalf("Failed to run gs with filter cycle: %v", err)
		}

		// Should show "Done" filter label
		if !strings.Contains(stdout, "Done") {
			t.Errorf("Expected 'Done' filter label after cycling, got: %s", stdout)
		}

		// Should show 2 matches (complete notes)
		if !strings.Contains(stdout, "2 matches") {
			t.Errorf("Expected '2 matches' for complete only, got: %s", stdout)
		}
	})

	t.Run("Cycle to ShowAll", func(t *testing.T) {
		// Cycle twice: ShowIncompleteOnly -> ShowCompleteOnly -> ShowAll
		stdout, _, err := h.RunCommand("gs\n\nffq\n")
		if err != nil {
			t.Fatalf("Failed to run gs with filter cycles: %v", err)
		}

		// Should show "All" filter label
		if !strings.Contains(stdout, "All") {
			t.Errorf("Expected 'All' filter label after two cycles, got: %s", stdout)
		}

		// Should show 4 matches (all notes)
		if !strings.Contains(stdout, "4 matches") {
			t.Errorf("Expected '4 matches' for show all, got: %s", stdout)
		}
	})

	t.Run("Cycle back to ShowIncompleteOnly", func(t *testing.T) {
		// Cycle three times: ShowIncompleteOnly -> ShowCompleteOnly -> ShowAll -> ShowIncompleteOnly
		stdout, _, err := h.RunCommand("gs\n\nfffq\n")
		if err != nil {
			t.Fatalf("Failed to run gs with filter cycles: %v", err)
		}

		// Should show "Open" filter label again
		if !strings.Contains(stdout, "Open") {
			t.Errorf("Expected 'Open' filter label after three cycles, got: %s", stdout)
		}

		// Should show 2 matches (incomplete only)
		if !strings.Contains(stdout, "2 matches") {
			t.Errorf("Expected '2 matches' for incomplete only, got: %s", stdout)
		}
	})
}

func TestSearchFilter_WithQuery(t *testing.T) {
	h := NewTestHarness(t)

	// Setup test data with searchable titles
	h.CreateTodo("meeting-incomplete.md", "Meeting Incomplete", []string{}, Today(), false, 1)
	h.CreateTodo("meeting-complete.md", "Meeting Complete", []string{}, Today(), true, 1)
	h.CreateTodo("other-note.md", "Other Note", []string{}, Today(), false, 2)

	t.Run("Filter applies to search results", func(t *testing.T) {
		// Search for "meeting" with default filter (ShowIncompleteOnly)
		// gs with query starts in insert mode, Enter to normal mode, q to quit
		stdout, _, err := h.RunCommand("gs meeting\n\nq\n")
		if err != nil {
			t.Fatalf("Failed to run gs with query: %v", err)
		}

		// Should show 1 match (only incomplete meeting)
		if !strings.Contains(stdout, "1 match") {
			t.Errorf("Expected '1 match' for 'meeting' with incomplete filter, got: %s", stdout)
		}
	})

	t.Run("ShowAll filter with search query", func(t *testing.T) {
		// Search for "meeting", cycle to ShowAll (ff), then quit
		stdout, _, err := h.RunCommand("gs meeting\n\nffq\n")
		if err != nil {
			t.Fatalf("Failed to run gs with query and filter: %v", err)
		}

		// Should show 2 matches (both meetings)
		if !strings.Contains(stdout, "2 matches") {
			t.Errorf("Expected '2 matches' for 'meeting' with show all, got: %s", stdout)
		}
	})

	t.Run("ShowCompleteOnly filter with search query", func(t *testing.T) {
		// Search for "meeting", cycle to ShowCompleteOnly (f), then quit
		stdout, _, err := h.RunCommand("gs meeting\n\nfq\n")
		if err != nil {
			t.Fatalf("Failed to run gs with query and filter: %v", err)
		}

		// Should show 1 match (only complete meeting)
		if !strings.Contains(stdout, "1 match") {
			t.Errorf("Expected '1 match' for 'meeting' with complete filter, got: %s", stdout)
		}

		// Should show "Done" label
		if !strings.Contains(stdout, "Done") {
			t.Errorf("Expected 'Done' filter label, got: %s", stdout)
		}
	})
}

func TestSearchFilter_PersistsAfterActions(t *testing.T) {
	h := NewTestHarness(t)

	// Setup test data
	h.CreateTodo("task-one.md", "Task One", []string{}, Today(), false, 1)
	h.CreateTodo("task-two.md", "Task Two", []string{}, Today(), false, 2)
	h.CreateTodo("task-done.md", "Task Done", []string{}, Today(), true, 1)

	t.Run("Filter persists after setting priority", func(t *testing.T) {
		// Set filter to ShowAll, then set priority
		// gs -> Enter (normal mode) -> ff (ShowAll) -> 3 (set P3) -> q
		stdout, _, err := h.RunCommand("gs\n\nff3q\n")
		if err != nil {
			t.Fatalf("Failed to run gs with priority action: %v", err)
		}

		// Should still show "All" filter after priority action
		if !strings.Contains(stdout, "All") {
			t.Errorf("Expected 'All' filter to persist after priority action, got: %s", stdout)
		}
	})

	t.Run("Filter persists after toggle done", func(t *testing.T) {
		// Create fresh harness to avoid state from previous test
		h2 := NewTestHarness(t)
		h2.CreateTodo("fresh-task.md", "Fresh Task", []string{}, Today(), false, 1)
		h2.CreateTodo("fresh-done.md", "Fresh Done", []string{}, Today(), true, 1)

		// Set filter to ShowAll, then toggle done
		// gs -> Enter (normal mode) -> ff (ShowAll) -> d (toggle done) -> q
		stdout, _, err := h2.RunCommand("gs\n\nffdq\n")
		if err != nil {
			t.Fatalf("Failed to run gs with toggle done: %v", err)
		}

		// Should still show "All" filter after toggle
		if !strings.Contains(stdout, "All") {
			t.Errorf("Expected 'All' filter to persist after toggle done, got: %s", stdout)
		}
	})
}

func TestSearchFilter_EmptyResults(t *testing.T) {
	h := NewTestHarness(t)

	// Setup test data: only complete notes
	h.CreateTodo("done1.md", "Done Note One", []string{}, Today(), true, 1)
	h.CreateTodo("done2.md", "Done Note Two", []string{}, Today(), true, 2)

	t.Run("Empty results with ShowIncompleteOnly when all done", func(t *testing.T) {
		// Default filter (ShowIncompleteOnly) should show no results
		stdout, _, err := h.RunCommand("gs\n\nq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// Should show 0 matches or "No results"
		if !strings.Contains(stdout, "0 match") && !strings.Contains(stdout, "No results") {
			// Check the match count is not present or shows empty state
			if strings.Contains(stdout, "Done Note") {
				t.Errorf("Expected no incomplete notes to be shown, got: %s", stdout)
			}
		}
	})

	t.Run("Shows results when cycling to ShowCompleteOnly", func(t *testing.T) {
		// Cycle to ShowCompleteOnly should show the done notes
		stdout, _, err := h.RunCommand("gs\n\nfq\n")
		if err != nil {
			t.Fatalf("Failed to run gs with filter: %v", err)
		}

		// Should show 2 matches
		if !strings.Contains(stdout, "2 matches") {
			t.Errorf("Expected '2 matches' for complete only, got: %s", stdout)
		}
	})
}

func TestSearchFilter_FKeyInNormalMode(t *testing.T) {
	h := NewTestHarness(t)

	h.CreateTodo("test.md", "Test Note", []string{}, Today(), false, 1)

	t.Run("F key works in normal mode", func(t *testing.T) {
		// Press f in normal mode to cycle filter
		stdout, _, err := h.RunCommand("gs\n\nfq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// Should show "Done" after first cycle
		if !strings.Contains(stdout, "Done") {
			t.Errorf("Expected filter to change on f key, got: %s", stdout)
		}
	})

	t.Run("F key typed in insert mode is part of query", func(t *testing.T) {
		// Type 'f' in insert mode - it should be added to query, not cycle filter
		// Then Enter to normal mode and quit
		stdout, _, err := h.RunCommand("gs\nf\nq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// Should still show "Open" (default filter, f didn't cycle it)
		// The f character was typed as search query
		if !strings.Contains(stdout, "Open") {
			t.Errorf("Expected 'Open' filter (f should be query in insert mode), got: %s", stdout)
		}
	})
}
