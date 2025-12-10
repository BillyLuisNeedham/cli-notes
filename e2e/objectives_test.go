package e2e

import (
	"fmt"
	"strings"
	"testing"
)

func TestObjectivesWorkflow(t *testing.T) {
	h := NewTestHarness(t)

	t.Run("Create Objective via UI", func(t *testing.T) {
		// Flow:
		// 1. "ob\n" -> enters objectives list view
		// 2. "n" -> triggers create new objective prompt
		// 3. "My Objective\n" -> title input + submit
		// 4. "q\n" -> quit view

		input := "ob\nnMy Objective\nq\n"

		stdout, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command output: %s", stdout)
		}

		// Verify the success message appears in output
		if !strings.Contains(stdout, "Created objective") {
			t.Errorf("Expected 'Created objective' message in output, got: %s", stdout)
		}

		if !strings.Contains(stdout, "My Objective") {
			t.Errorf("Expected objective title 'My Objective' in output, got: %s", stdout)
		}

		// Verify a file was created with the objective title
		files := h.ListFiles()
		found := false
		for _, f := range files {
			if strings.Contains(strings.ToLower(f), "my objective") {
				found = true
				// Verify it has objective-role: parent
				fm := h.ParseFrontmatter(f)
				if fm.ObjectiveRole != "parent" {
					t.Errorf("Expected objective-role 'parent', got '%s'", fm.ObjectiveRole)
				}
				if fm.ObjectiveID == "" {
					t.Errorf("Expected objective-id to be set")
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected to find a file containing 'my objective', found files: %v", files)
		}
	})

	t.Run("Convert Todo to Objective", func(t *testing.T) {
		h.CreateTodo("todo-to-convert.md", "Convert Me", []string{}, Today(), false, 1)

		// cpo - Convert to parent objective
		// Needs selected file.
		// Flow:
		// 1. "gt" -> list all todos
		// 2. Select it (ArrowDown)
		// 3. "cpo"
		// 4. Confirm if prompted (y)

		input := "gt\n\x1b[Bcpo\ny\n"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command failed: %v", err)
		}

		// Verify it has objective-role: parent
		h.AssertFrontmatterValue("todo-to-convert.md", func(fm Frontmatter) error {
			if fm.ObjectiveRole != "parent" {
				return fmt.Errorf("expected objective-role 'parent', got '%s'", fm.ObjectiveRole)
			}
			if fm.ObjectiveID == "" {
				return fmt.Errorf("expected objective-id to be set")
			}
			return nil
		})
	})
}

func TestOpenObjectiveFromRootView(t *testing.T) {
	h := NewTestHarness(t)

	// 1. Create a parent objective first (by converting a todo)
	h.CreateTodo("my-objective.md", "My Objective", []string{}, Today(), false, 1)

	// Convert to objective: gt -> select -> cpo -> confirm
	input := "gt\n\x1b[Bcpo\ny\n"
	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Fatalf("Failed to convert todo to objective: %v", err)
	}

	// Verify it's now an objective
	h.AssertFrontmatterValue("my-objective.md", func(fm Frontmatter) error {
		if fm.ObjectiveRole != "parent" {
			return fmt.Errorf("expected objective-role 'parent', got '%s'", fm.ObjectiveRole)
		}
		return nil
	})

	// 2. Test: Navigate to objective from root using "ob My Objective"
	// Then quit with 'q'
	input = "ob My Objective\nq\n"
	stdout, _, err := h.RunCommand(input)
	if err != nil {
		t.Fatalf("Failed to open objective from root: %v", err)
	}

	// 3. Verify output shows the objective view (single view, not list view)
	// The single objective view should show the objective title
	if !strings.Contains(stdout, "My Objective") {
		t.Errorf("Expected objective view to show 'My Objective', got: %s", stdout)
	}
}

func TestOpenObjectiveFromRootViewWithTabAutocomplete(t *testing.T) {
	h := NewTestHarness(t)

	// 1. Create a parent objective
	h.CreateTodo("annual-review.md", "Annual Review", []string{}, Today(), false, 1)

	// Convert to objective
	input := "gt\n\x1b[Bcpo\ny\n"
	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Fatalf("Failed to convert todo to objective: %v", err)
	}

	// 2. Test: Tab autocomplete for objective name from root
	// Type "ob Ann" then Tab to autocomplete, then Enter, then quit
	input = "ob Ann\t\nq\n"
	stdout, _, err := h.RunCommand(input)
	if err != nil {
		t.Fatalf("Failed to open objective with autocomplete: %v", err)
	}

	// 3. Verify it opened the correct objective
	if !strings.Contains(stdout, "Annual Review") {
		t.Errorf("Expected objective view to show 'Annual Review', got: %s", stdout)
	}
}

func TestObjectivesViewShowsOpenTasks(t *testing.T) {
	h := NewTestHarness(t)

	// 1. Create a parent objective with a unique ID
	objectiveID := "test1234"
	h.CreateObjective("project-objective.md", "Project Objective", objectiveID, "Project goals and tasks")

	// 2. Create a child todo with checkbox content (open tasks)
	childContent := `Tasks to complete:
- [ ] Task One
- [ ] Task Two
- [x] Completed Task`
	h.CreateLinkedTodo("child-todo.md", "Child Todo", objectiveID, childContent, Today(), 1)

	// 3. Open objectives view, navigate to the objective, then navigate to child
	// ob -> opens objectives list
	// o or Enter -> opens selected objective (single view)
	// j -> navigate down to first child
	// q -> quit
	input := "ob\no\nj\nq\n"
	stdout, _, err := h.RunCommand(input)
	if err != nil {
		t.Fatalf("Failed to open objectives view: %v", err)
	}

	// 4. Verify the output shows the open tasks from the child
	// The right panel should show "OPEN TASKS" and the uncompleted checkboxes
	if !strings.Contains(stdout, "OPEN TASKS") {
		t.Errorf("Expected to see 'OPEN TASKS' section in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Task One") {
		t.Errorf("Expected to see 'Task One' in open tasks, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Task Two") {
		t.Errorf("Expected to see 'Task Two' in open tasks, got: %s", stdout)
	}
}

func TestObjectivesViewEscapeCancelsQueryInput(t *testing.T) {
	h := NewTestHarness(t)

	// 1. Create a parent objective
	objectiveID := "esc12345"
	h.CreateObjective("escape-test-obj.md", "Escape Test Objective", objectiveID, "Test objective")

	// 2. Create an unlinked todo that could be linked
	h.CreateTodo("unlinked-todo.md", "Unlinked Todo", []string{}, Today(), false, 1)

	// 3. Open objectives view, open the objective, press 'l' to link, then press Escape
	// ob -> opens objectives list
	// o -> opens selected objective
	// l -> start link mode (prompts for search query)
	// \x1b -> escape key (cancel)
	// q -> quit
	input := "ob\no\nl\x1b\nq\n"
	stdout, _, err := h.RunCommand(input)
	if err != nil {
		t.Fatalf("Failed in escape test: %v", err)
	}

	// 4. Verify that "Cancelled" appears in output (indicates escape was handled)
	if !strings.Contains(stdout, "Cancelled") {
		t.Errorf("Expected 'Cancelled' message when pressing escape, got: %s", stdout)
	}

	// 5. Verify the todo was NOT linked (still has no objective-id or has different one)
	fm := h.ParseFrontmatter("unlinked-todo.md")
	if fm.ObjectiveID == objectiveID {
		t.Errorf("Expected todo to remain unlinked after escape, but it was linked to %s", fm.ObjectiveID)
	}
}

func TestObjectivesViewEscapeCancelsCreateChild(t *testing.T) {
	h := NewTestHarness(t)

	// 1. Create a parent objective
	objectiveID := "cresc123"
	h.CreateObjective("create-escape-obj.md", "Create Escape Test", objectiveID, "Test objective")

	// 2. Open objectives view, open the objective, press 'n' to create child, then press Escape
	// ob -> opens objectives list
	// o -> opens selected objective
	// n -> start create child mode (prompts for title)
	// \x1b -> escape key (cancel)
	// q -> quit
	input := "ob\no\nn\x1b\nq\n"
	stdout, _, err := h.RunCommand(input)
	if err != nil {
		t.Fatalf("Failed in create escape test: %v", err)
	}

	// 3. Verify that "Cancelled" appears in output (indicates escape was handled)
	if !strings.Contains(stdout, "Cancelled") {
		t.Errorf("Expected 'Cancelled' message when pressing escape during title input, got: %s", stdout)
	}
}

func TestObjectivesListViewTabSwitching(t *testing.T) {
	h := NewTestHarness(t)

	// 1. Create an active objective
	h.CreateObjective("active-obj.md", "Active Objective", "act12345", "This is active")

	// 2. Create a completed objective
	h.CreateCompletedObjective("completed-obj.md", "Completed Objective", "done1234", "This is done")

	// 3. Open objectives view - should show active tab by default
	// ob -> opens objectives list
	// q -> quit (no newline between commands in objectives view - newline = Enter = open)
	input := "ob\nq"
	stdout, _, err := h.RunCommand(input)
	if err != nil {
		t.Fatalf("Failed to open objectives view: %v", err)
	}

	// 4. Verify tab header shows [ACTIVE] (selected) and active objective is visible
	if !strings.Contains(stdout, "[ACTIVE") {
		t.Errorf("Expected [ACTIVE] tab to be selected by default, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Active Objective") {
		t.Errorf("Expected 'Active Objective' to be visible in active tab, got: %s", stdout)
	}
	// Completed objective should NOT be visible in active tab
	if strings.Contains(stdout, "Completed Objective") {
		t.Errorf("Expected 'Completed Objective' NOT to be visible in active tab, got: %s", stdout)
	}

	// 5. Test switching to completed tab with 'f' key
	// ob -> opens objectives list
	// f -> switch to completed tab
	// q -> quit (no newlines between single-char commands)
	input = "ob\nfq"
	stdout, _, err = h.RunCommand(input)
	if err != nil {
		t.Fatalf("Failed to switch to completed tab: %v", err)
	}

	// 6. Verify [COMPLETED] tab is now selected and completed objective is visible
	if !strings.Contains(stdout, "[COMPLETED") {
		t.Errorf("Expected [COMPLETED] tab to be selected after pressing 'f', got: %s", stdout)
	}
	if !strings.Contains(stdout, "Completed Objective") {
		t.Errorf("Expected 'Completed Objective' to be visible in completed tab, got: %s", stdout)
	}

	// 7. Verify the completed tab view shows Completed Objective as selectable (with > prefix)
	// This confirms it's in the completed tab, not active tab
	if !strings.Contains(stdout, "> Completed Objective") {
		t.Errorf("Expected 'Completed Objective' to be selected in completed tab, got: %s", stdout)
	}
}

func TestObjectivesListViewTabCounts(t *testing.T) {
	h := NewTestHarness(t)

	// Create 2 active objectives
	h.CreateObjective("active1.md", "Active One", "act00001", "Active 1")
	h.CreateObjective("active2.md", "Active Two", "act00002", "Active 2")

	// Create 1 completed objective
	h.CreateCompletedObjective("completed1.md", "Completed One", "done0001", "Done 1")

	// Open objectives view and check counts (no newline between single-char commands)
	input := "ob\nq"
	stdout, _, err := h.RunCommand(input)
	if err != nil {
		t.Fatalf("Failed to open objectives view: %v", err)
	}

	// Verify tab header shows correct counts: [ACTIVE (2)]  COMPLETED (1)
	if !strings.Contains(stdout, "ACTIVE (2)") {
		t.Errorf("Expected 'ACTIVE (2)' in tab header, got: %s", stdout)
	}
	if !strings.Contains(stdout, "COMPLETED (1)") {
		t.Errorf("Expected 'COMPLETED (1)' in tab header, got: %s", stdout)
	}
}
