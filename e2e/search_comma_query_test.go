package e2e

import (
	"strings"
	"testing"
)

func TestSearchCommaQueries(t *testing.T) {
	h := NewTestHarness(t)

	// Setup test data: notes with different tag combinations
	h.CreateTestFile("note-apples.md", "---\ntitle: Note Apples\ntags: [fruit]\ndone: false\n---\nContent about apples")
	h.CreateTestFile("note-bananas.md", "---\ntitle: Note Bananas\ntags: [fruit]\ndone: false\n---\nContent about bananas")
	h.CreateTestFile("note-both.md", "---\ntitle: Note Both\ntags: [fruit]\ndone: false\n---\nContent about apples and bananas")

	t.Run("Single query returns all matches", func(t *testing.T) {
		// gs apples -> should match note-apples and note-both
		// Cycle to ShowAll first with ff to include all notes
		stdout, _, err := h.RunCommand("gs apples\n\nffq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// Should show 2 matches (apples and both)
		if !strings.Contains(stdout, "2 match") {
			t.Errorf("Expected 2 matches for 'apples', got: %s", stdout)
		}
	})

	t.Run("Comma-separated queries use AND logic from command line", func(t *testing.T) {
		// gs apples, bananas -> should only match note-both
		stdout, _, err := h.RunCommand("gs apples, bananas\n\nffq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// Should show 1 match (only note-both has both terms)
		if !strings.Contains(stdout, "1 match") {
			t.Errorf("Expected 1 match for 'apples, bananas' (AND logic), got: %s", stdout)
		}
	})

	t.Run("Comma-separated queries typed in search view", func(t *testing.T) {
		// gs -> type "apples, bananas" in insert mode -> should only match note-both
		// Enter to go to normal mode, ff to show all, q to quit
		stdout, _, err := h.RunCommand("gs\napples, bananas\nffq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// Should show 1 match (only note-both has both terms)
		if !strings.Contains(stdout, "1 match") {
			t.Errorf("Expected 1 match for typed 'apples, bananas' (AND logic), got: %s", stdout)
		}
	})
}
