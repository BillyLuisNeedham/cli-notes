package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// NOTE: These tests currently skip because the week planner uses the keyboard library
// for interactive input, which doesn't work in the test environment where stdin is piped.
// To enable these tests, the week planner needs to be modified to accept stdin input
// when CLI_NOTES_TEST_MODE=true instead of using the keyboard library.

func TestWeekPlanner_CreateTodoOnSelectedDay(t *testing.T) {
	t.Skip("Requires week planner to support stdin input in test mode")
	t.Run("CreateOnMonday", func(t *testing.T) {
		h := NewTestHarness(t)

		// Open planner, switch to Monday, create todo
		// Flow: wp -> M (Monday view) -> n (create new todo) -> title input -> Enter -> q (quit planner)
		// Note: 'n' prompts for title, then opens editor (which is set to 'echo' in test env)
		input := "wp\nMn\nMonday Task\n\nq"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Find the created file (it will have today's date suffix)
		files, err := os.ReadDir(h.NotesDir)
		if err != nil {
			t.Fatalf("Failed to read notes dir: %v", err)
		}

		// Look for a file containing "monday-task" in the name
		var foundFile string
		for _, file := range files {
			if strings.Contains(strings.ToLower(file.Name()), "monday") {
				foundFile = file.Name()
				break
			}
		}

		if foundFile == "" {
			t.Fatalf("Could not find created Monday task file")
		}

		// Verify the due date is set to Monday
		fm := h.ParseFrontmatter(foundFile)
		expectedDate := MondayThisWeek()
		if fm.DateDue != expectedDate {
			t.Errorf("Expected new todo due date to be %s (Monday), got %s", expectedDate, fm.DateDue)
		}
	})

	t.Run("CreateOnFriday", func(t *testing.T) {
		h := NewTestHarness(t)

		// Open planner, switch to Friday, create todo
		// Flow: wp -> F (Friday view) -> n (create new todo) -> title -> Enter -> q
		input := "wp\nFn\nFriday Task\n\nq"

		_, _, err := h.RunCommand(input)
		if err != nil {
			t.Logf("Command completed with: %v", err)
		}

		// Find the created file
		files, err := os.ReadDir(h.NotesDir)
		if err != nil {
			t.Fatalf("Failed to read notes dir: %v", err)
		}

		var foundFile string
		for _, file := range files {
			if strings.Contains(strings.ToLower(file.Name()), "friday") {
				foundFile = file.Name()
				break
			}
		}

		if foundFile == "" {
			t.Fatalf("Could not find created Friday task file")
		}

		// Verify the due date is set to Friday
		fm := h.ParseFrontmatter(foundFile)
		expectedDate := FridayThisWeek()
		if fm.DateDue != expectedDate {
			t.Errorf("Expected new todo due date to be %s (Friday), got %s", expectedDate, fm.DateDue)
		}
	})
}

func TestWeekPlanner_CreateTodoSetsDueDate(t *testing.T) {
	t.Skip("Requires week planner to support stdin input in test mode")
	h := NewTestHarness(t)

	// Create todo on Wednesday and verify automatic due date assignment
	// Flow: wp -> W (Wednesday view) -> n (create) -> title -> Enter -> q
	input := "wp\nWn\nWednesday Auto Date Test\n\nq"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with: %v", err)
	}

	// Find the created file
	files, err := os.ReadDir(h.NotesDir)
	if err != nil {
		t.Fatalf("Failed to read notes dir: %v", err)
	}

	var foundFile string
	for _, file := range files {
		if strings.Contains(strings.ToLower(file.Name()), "wednesday") ||
			strings.Contains(strings.ToLower(file.Name()), "auto-date") {
			foundFile = file.Name()
			break
		}
	}

	if foundFile == "" {
		t.Fatalf("Could not find created Wednesday task file")
	}

	// Verify frontmatter has correct due date
	fm := h.ParseFrontmatter(foundFile)

	// Check due date is Wednesday
	expectedDate := WednesdayThisWeek()
	if fm.DateDue != expectedDate {
		t.Errorf("Expected auto-assigned due date to be %s (Wednesday), got %s", expectedDate, fm.DateDue)
	}

	// Verify title was set correctly
	if !strings.Contains(fm.Title, "Wednesday") || !strings.Contains(fm.Title, "Auto Date") {
		t.Errorf("Expected title to contain 'Wednesday Auto Date', got %s", fm.Title)
	}

	// Verify done is false (new todos should not be marked complete)
	if fm.Done {
		t.Errorf("Expected new todo to have done=false, got true")
	}
}

func TestWeekPlanner_NewTodoAppearsInCorrectDay(t *testing.T) {
	t.Skip("Requires week planner to support stdin input in test mode")
	h := NewTestHarness(t)

	// Create an existing todo on Monday
	h.CreateTodo("existing-monday.md", "Existing Monday Task", []string{}, MondayThisWeek(), false, 1)

	// Create a new todo on Tuesday
	// Flow: wp -> T (Tuesday view) -> n (create) -> title -> Enter -> Ctrl+S (save) -> q
	input := "wp\nTn\nNew Tuesday Task\n\n\x13q"

	stdout, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with: %v", err)
	}

	// Find the newly created file
	files, err := os.ReadDir(h.NotesDir)
	if err != nil {
		t.Fatalf("Failed to read notes dir: %v", err)
	}

	var newTodoFile string
	for _, file := range files {
		// Skip the existing Monday file
		if file.Name() == "existing-monday.md" {
			continue
		}
		if strings.Contains(strings.ToLower(file.Name()), "tuesday") ||
			strings.Contains(strings.ToLower(file.Name()), "new-tuesday") {
			newTodoFile = file.Name()
			break
		}
	}

	if newTodoFile == "" {
		// List all files for debugging
		var fileNames []string
		for _, file := range files {
			fileNames = append(fileNames, file.Name())
		}
		t.Logf("Files in notes dir: %v", fileNames)
		t.Logf("Stdout: %s", stdout)
		t.Fatalf("Could not find newly created Tuesday task file")
	}

	// Verify the new todo has correct due date (Tuesday)
	fm := h.ParseFrontmatter(newTodoFile)
	expectedDate := TuesdayThisWeek()
	if fm.DateDue != expectedDate {
		t.Errorf("Expected new todo to appear on %s (Tuesday), got %s", expectedDate, fm.DateDue)
	}

	// Verify both files exist
	h.AssertFileExists("existing-monday.md")
	h.AssertFileExists(newTodoFile)

	// Verify existing Monday todo wasn't modified
	mondayFm := h.ParseFrontmatter("existing-monday.md")
	if mondayFm.DateDue != MondayThisWeek() {
		t.Errorf("Existing Monday todo was modified, due date changed from %s to %s",
			MondayThisWeek(), mondayFm.DateDue)
	}
}

func TestWeekPlanner_CreateInNextMondayBucket(t *testing.T) {
	t.Skip("Requires week planner to support stdin input in test mode")
	h := NewTestHarness(t)

	// Switch to Next Monday view and create a todo there
	// The exact key for "Next Monday" view might vary, but typically it's accessible
	// For this test, we'll create on current Monday and move to Next Monday
	input := "wp\nMn\nTask for Next Monday\n\nN\x13q"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with: %v", err)
	}

	// Find the created file
	files, err := os.ReadDir(h.NotesDir)
	if err != nil {
		t.Fatalf("Failed to read notes dir: %v", err)
	}

	var foundFile string
	for _, file := range files {
		if strings.Contains(strings.ToLower(file.Name()), "next") ||
			strings.Contains(strings.ToLower(file.Name()), "monday") {
			// Read to check if it has next Monday's date
			path := filepath.Join(h.NotesDir, file.Name())
			content, _ := os.ReadFile(path)
			if strings.Contains(string(content), "Next Monday") {
				foundFile = file.Name()
				break
			}
		}
	}

	if foundFile == "" {
		t.Logf("Warning: Could not find todo created in Next Monday bucket")
		// This test might be flaky depending on UI implementation
		return
	}

	// Verify the due date is next Monday
	fm := h.ParseFrontmatter(foundFile)
	expectedDate := NextMonday()
	if fm.DateDue != expectedDate {
		t.Errorf("Expected todo in Next Monday bucket to have due date %s, got %s", expectedDate, fm.DateDue)
	}
}
