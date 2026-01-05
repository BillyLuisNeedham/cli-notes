package e2e

import (
	"strings"
	"testing"
)

func TestSearchModes(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: create notes with words that fuzzy vs strict would match differently
	h.CreateTestFile("note-children.md", "---\ntitle: Note about Children\ntags: [family]\ndone: false\n---\nContent about children playing")
	h.CreateTestFile("note-child.md", "---\ntitle: Child Care\ntags: [work]\ndone: false\n---\nContent about child care services")
	h.CreateTestFile("note-other.md", "---\ntitle: Other Note\ntags: [misc]\ndone: false\n---\nSomething else entirely")

	t.Run("Strict mode is default - non-substring returns no results", func(t *testing.T) {
		// gs -> type "chldcare" (not a substring - missing 'i') -> should show 0 matches in Strict mode
		// ff to show All, Enter for normal mode, q to quit
		stdout, _, err := h.RunCommand("gs\nchldcare\nffq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// Should show "Strict" mode label
		if !strings.Contains(stdout, "Strict") {
			t.Errorf("Expected 'Strict' mode label, got: %s", stdout)
		}

		// Should show 0 matches in final All view (chldcare is not exact substring)
		if !strings.Contains(stdout, "0 matches | All") {
			t.Errorf("Expected '0 matches | All' for 'chldcare' in Strict mode, got: %s", stdout)
		}
	})

	t.Run("Strict mode - exact substring matches", func(t *testing.T) {
		// gs -> type "child" (exact substring) -> should match both notes
		stdout, _, err := h.RunCommand("gs\nchild\nffq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// Should show 2 matches (both notes contain "child")
		if !strings.Contains(stdout, "2 match") {
			t.Errorf("Expected 2 matches for 'child' in Strict mode, got: %s", stdout)
		}
	})

	t.Run("Toggle to Fuzzy mode with s key", func(t *testing.T) {
		// gs -> Enter (normal mode) -> s (toggle to Fuzzy) -> i (insert) -> type "chldcare" -> should match
		// In Fuzzy mode, "chldcare" should match "child care" (letters in sequence)
		stdout, _, err := h.RunCommand("gs\n\nsichldcare\nffq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// Should show "Fuzzy" mode label
		if !strings.Contains(stdout, "Fuzzy") {
			t.Errorf("Expected 'Fuzzy' mode label after toggle, got: %s", stdout)
		}

		// Should show matches in All view (fuzzy matching finds "chldcare" pattern in "child care")
		if !strings.Contains(stdout, "matches | All | Fuzzy") {
			t.Errorf("Expected matches in All Fuzzy mode for 'chldcare', got: %s", stdout)
		}
	})

	t.Run("Space works in search query", func(t *testing.T) {
		// gs -> type "child care" (with space) -> should match
		stdout, _, err := h.RunCommand("gs\nchild care\nffq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// Query should contain space
		if !strings.Contains(stdout, "child care") {
			t.Errorf("Expected space in query 'child care', got: %s", stdout)
		}

		// Should show 1 match (only "Child Care" note has both words)
		if !strings.Contains(stdout, "1 match") {
			t.Errorf("Expected 1 match for 'child care', got: %s", stdout)
		}
	})

	t.Run("UI shows s:Srch in controls", func(t *testing.T) {
		// gs -> Enter (normal mode) -> check controls include s:Srch
		stdout, _, err := h.RunCommand("gs\n\nq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// Controls should show s:Srch
		if !strings.Contains(stdout, "s:Srch") {
			t.Errorf("Expected 's:Srch' in controls, got: %s", stdout)
		}
	})

	t.Run("Comma-separated AND logic works in Strict mode", func(t *testing.T) {
		// gs -> type "child,family" -> should match only note with both
		stdout, _, err := h.RunCommand("gs\nchild,family\nffq\n")
		if err != nil {
			t.Fatalf("Failed to run gs: %v", err)
		}

		// Should show 1 match (only children note has "child" substring AND "family" tag)
		if !strings.Contains(stdout, "1 match") {
			t.Errorf("Expected 1 match for 'child,family', got: %s", stdout)
		}
	})
}
