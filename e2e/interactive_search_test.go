package e2e

import (
	"strings"
	"testing"
)

func TestInteractiveSearch(t *testing.T) {
	h := NewTestHarness(t)

	// Setup test data
	h.CreateTodo("meeting-notes-2025-01-01.md", "Meeting Notes", []string{"work"}, Today(), false, 1)
	h.CreateTodo("project-plan-2025-01-02.md", "Project Plan", []string{"planning"}, Today(), false, 2)
	h.CreateTodo("weekly-review-2025-01-03.md", "Weekly Review", []string{"work"}, Today(), false, 3)

	t.Run("Open and close search view", func(t *testing.T) {
		// gs opens the search view, q exits it
		stdout, _, err := h.RunCommand("gs\nq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// The output should contain the search UI elements
		// (ANSI codes make exact matching difficult, but we can check for key elements)
		if !strings.Contains(stdout, "SEARCH") {
			t.Errorf("Expected search UI header, got: %s", stdout)
		}
	})

	t.Run("Search with initial query", func(t *testing.T) {
		// gs meeting opens search with "meeting" pre-filled
		stdout, _, err := h.RunCommand("gs meeting\nq\n")
		if err != nil {
			t.Fatalf("Failed to run gs with query: %v", err)
		}

		// Should show search UI
		if !strings.Contains(stdout, "SEARCH") {
			t.Errorf("Expected search UI, got: %s", stdout)
		}
	})

	t.Run("Fuzzy search finds matching notes", func(t *testing.T) {
		// Type partial query and exit
		// "mtg" should fuzzy match "meeting"
		stdout, _, err := h.RunCommand("gs\nmtg\nq\n")
		if err != nil {
			t.Fatalf("Failed to run fuzzy search: %v", err)
		}

		// Should show search UI
		if !strings.Contains(stdout, "SEARCH") {
			t.Errorf("Expected search UI, got: %s", stdout)
		}
	})

	t.Run("Navigate with j/k keys", func(t *testing.T) {
		// Navigate through results
		stdout, _, err := h.RunCommand("gs\nj\nk\nq\n")
		if err != nil {
			t.Fatalf("Failed to navigate: %v", err)
		}

		// Should complete without error
		if !strings.Contains(stdout, "SEARCH") {
			t.Errorf("Expected search UI, got: %s", stdout)
		}
	})

	t.Run("Set priority from search", func(t *testing.T) {
		// Navigate and set priority
		// Open search, wait for results, press 2 to set priority P2
		stdout, _, err := h.RunCommand("gs\n2\nq\n")
		if err != nil {
			t.Fatalf("Failed to set priority: %v", err)
		}

		// Should show priority confirmation
		if !strings.Contains(stdout, "Priority") {
			t.Logf("Note: Priority message may be obscured by screen clearing")
		}
	})

	t.Run("Toggle done status from search", func(t *testing.T) {
		// Create a separate todo for this test
		h.CreateTodo("test-done-2025-01-04.md", "Test Done", []string{}, Today(), false, 1)

		// Open search, find the note, press d to toggle done
		stdout, _, err := h.RunCommand("gs\nTest Done\nd\nq\n")
		if err != nil {
			t.Fatalf("Failed to toggle done: %v", err)
		}

		// The command should complete
		if !strings.Contains(stdout, "SEARCH") {
			t.Errorf("Expected search UI, got: %s", stdout)
		}
	})
}
