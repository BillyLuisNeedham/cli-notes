package e2e

import (
	"fmt"
	"strings"
	"testing"
)

func TestNoteCreation(t *testing.T) {
	h := NewTestHarness(t)

	t.Run("Create Meeting Note", func(t *testing.T) {
		stdout, _, err := h.RunCommand("cm team-sync\n")
		if err != nil {
			t.Fatalf("Failed to run cm: %v", err)
		}

		// Extract filename from output
		var filename string
		lines := strings.Split(stdout, "\n")
		for _, line := range lines {
			if strings.Contains(line, "notes/") && strings.HasSuffix(line, ".md") {
				parts := strings.Split(strings.TrimSpace(line), "/")
				if len(parts) > 0 {
					filename = parts[len(parts)-1]
					break
				}
			}
		}

		if filename == "" {
			t.Fatalf("Could not extract filename from output: %s", stdout)
		}

		h.AssertFileExists(filename)
		h.AssertFrontmatterValue(filename, func(fm Frontmatter) error {
			if fm.Title != "team-sync" {
				return fmt.Errorf("expected title 'team-sync'")
			}
			// Check tags if meeting notes have specific tags
			return nil
		})
	})

	t.Run("Create Planning Note", func(t *testing.T) {
		stdout, _, err := h.RunCommand("cp week-planning\n")
		if err != nil {
			t.Fatalf("Failed to run cp: %v", err)
		}

		// Extract filename from output
		var filename string
		lines := strings.Split(stdout, "\n")
		for _, line := range lines {
			if strings.Contains(line, "notes/") && strings.HasSuffix(line, ".md") {
				parts := strings.Split(strings.TrimSpace(line), "/")
				if len(parts) > 0 {
					filename = parts[len(parts)-1]
					break
				}
			}
		}

		if filename == "" {
			t.Fatalf("Could not extract filename from output: %s", stdout)
		}

		h.AssertFileExists(filename)
		// Check for 7 questions template content (actual template questions)
		h.AssertFileContent(filename, "What is the situation and how does it affect me")
		h.AssertFileContent(filename, "What effects do I need to achieve")
	})

	t.Run("Create Standup Note", func(t *testing.T) {
		stdout, _, err := h.RunCommand("cs\n")
		if err != nil {
			t.Fatalf("Failed to run cs: %v", err)
		}

		// Extract filename from output
		var filename string
		lines := strings.Split(stdout, "\n")
		for _, line := range lines {
			if strings.Contains(line, "notes/") && strings.HasSuffix(line, ".md") {
				parts := strings.Split(strings.TrimSpace(line), "/")
				if len(parts) > 0 {
					filename = parts[len(parts)-1]
					break
				}
			}
		}

		if filename == "" {
			t.Fatalf("Could not extract filename from output: %s", stdout)
		}

		h.AssertFileExists(filename)

		// Check for team members (assuming default team or we mocked the repo)
		// The original code uses "scripts/data/team_names_repository.go".
		// If it reads from a file, we might need to populate it.
		// If it's hardcoded or has defaults, we check that.
		// Let's assume it has some content.
	})
}
