package data

import (
	"cli-notes/scripts"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestHelper struct for managing the test environment
type testHelper struct {
	tempDir string
	origDir string
}

// Setup creates a temporary test directory and redirects the DirectoryPath constant
func setupTest(t *testing.T) *testHelper {
	th := &testHelper{}

	// Save original working directory
	var err error
	th.origDir, err = os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create temporary directory
	th.tempDir, err = os.MkdirTemp("", "file_repository_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create notes subdirectory in temp dir
	err = os.Mkdir(filepath.Join(th.tempDir, "notes"), 0755)
	if err != nil {
		t.Fatalf("Failed to create notes directory: %v", err)
	}

	// Change to temp directory
	err = os.Chdir(th.tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	return th
}

// Cleanup removes temporary test directory and restores original directory
func (th *testHelper) cleanup(t *testing.T) {
	// Return to original directory
	err := os.Chdir(th.origDir)
	if err != nil {
		t.Fatalf("Failed to return to original directory: %v", err)
	}

	// Remove temp directory
	err = os.RemoveAll(th.tempDir)
	if err != nil {
		t.Fatalf("Failed to remove temp directory: %v", err)
	}
}

// Helper to create test files
func createTestFile(t *testing.T, file scripts.File) {
	// Set default priority if not set
	if file.Priority == 0 {
		file.Priority = scripts.P2
	}

	err := WriteFile(file)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
}

func TestWriteFile(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Test file to create
	now := time.Now()
	dueDate := now.AddDate(0, 0, 7)
	testFile := scripts.File{
		Name:      "test1.md",
		Title:     "Test File 1",
		CreatedAt: now,
		DueAt:     dueDate,
		Tags:      []string{"test", "unit-test"},
		Done:      false,
		Content:   "This is test content",
		Priority:  scripts.P2,
	}

	// Call WriteFile
	err := WriteFile(testFile)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Check if file exists
	currentDir, _ := os.Getwd()
	filePath := filepath.Join(currentDir, DirectoryPath, testFile.Name)
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		t.Fatal("WriteFile did not create the file")
	}

	// Read the file content to verify
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	// Check basic content
	contentStr := string(content)
	contentLines := strings.Split(contentStr, "\n")

	expectedContent := []string{
		"---",
		"title: Test File 1",
		"date-created: " + timeToString(now),
		"tags: [test unit-test]",
		"priority: 2",
		"date-due: " + timeToString(dueDate),
		"done: false",
		"---",
		"",
		"This is test content",
	}

	for _, line := range expectedContent {
		if !contains(contentLines, line) {
			t.Errorf("File content missing expected line: %s", line)
		}
	}
}

func TestQueryFilesByDone(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	now := time.Now()

	// Create test files
	createTestFile(t, scripts.File{
		Name:      "test-done.md",
		Title:     "Test Done",
		CreatedAt: now,
		DueAt:     now,
		Tags:      []string{"test"},
		Done:      true,
		Content:   "This is a completed task",
		Priority:  scripts.P1,
	})

	createTestFile(t, scripts.File{
		Name:      "test-not-done.md",
		Title:     "Test Not Done",
		CreatedAt: now,
		DueAt:     now,
		Tags:      []string{"test"},
		Done:      false,
		Content:   "This is an incomplete task",
		Priority:  scripts.P3,
	})

	// Test querying done files
	doneFiles, err := QueryFilesByDone(true)
	if err != nil {
		t.Fatalf("QueryFilesByDone(true) failed: %v", err)
	}

	if len(doneFiles) != 1 {
		t.Errorf("Expected 1 done file, got %d", len(doneFiles))
	}

	if len(doneFiles) > 0 && doneFiles[0].Name != "test-done.md" {
		t.Errorf("Expected file name test-done.md, got %s", doneFiles[0].Name)
	}

	// Test querying todo files
	todoFiles, err := QueryFilesByDone(false)
	if err != nil {
		t.Fatalf("QueryFilesByDone(false) failed: %v", err)
	}

	if len(todoFiles) != 1 {
		t.Errorf("Expected 1 todo file, got %d", len(todoFiles))
	}

	if len(todoFiles) > 0 && todoFiles[0].Name != "test-not-done.md" {
		t.Errorf("Expected file name test-not-done.md, got %s", todoFiles[0].Name)
	}
}

func TestQueryTodosWithDateCriteria(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	nextWeek := now.AddDate(0, 0, 7)

	// Create test files with different due dates
	createTestFile(t, scripts.File{
		Name:      "overdue.md",
		Title:     "Overdue Task",
		CreatedAt: yesterday,
		DueAt:     yesterday,
		Tags:      []string{"test"},
		Done:      false,
		Content:   "This task is overdue",
	})

	createTestFile(t, scripts.File{
		Name:      "due-today.md",
		Title:     "Due Today Task",
		CreatedAt: yesterday,
		DueAt:     now,
		Tags:      []string{"test"},
		Done:      false,
		Content:   "This task is due today",
	})

	createTestFile(t, scripts.File{
		Name:      "due-tomorrow.md",
		Title:     "Due Tomorrow Task",
		CreatedAt: yesterday,
		DueAt:     tomorrow,
		Tags:      []string{"test"},
		Done:      false,
		Content:   "This task is due tomorrow",
	})

	createTestFile(t, scripts.File{
		Name:      "due-next-week.md",
		Title:     "Due Next Week Task",
		CreatedAt: yesterday,
		DueAt:     nextWeek,
		Tags:      []string{"test"},
		Done:      false,
		Content:   "This task is due next week",
	})

	// Test for overdue tasks
	overdueTasks, err := QueryTodosWithDateCriteria(func(dueDate string, dueDateParsed time.Time) bool {
		// Normalize to start of day to ignore time components
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return dueDateParsed.Before(today)
	})

	if err != nil {
		t.Fatalf("QueryTodosWithDateCriteria for overdue tasks failed: %v", err)
	}

	if len(overdueTasks) != 1 {
		t.Errorf("Expected 1 overdue task, got %d", len(overdueTasks))
	}

	// Test for today's tasks
	todaysTasks, err := QueryTodosWithDateCriteria(func(dueDate string, dueDateParsed time.Time) bool {
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return dueDateParsed.Equal(today)
	})

	if err != nil {
		t.Fatalf("QueryTodosWithDateCriteria for today's tasks failed: %v", err)
	}

	if len(todaysTasks) != 1 {
		t.Errorf("Expected 1 task due today, got %d", len(todaysTasks))
	}

	// Test for future tasks (tomorrow and beyond)
	futureTasks, err := QueryTodosWithDateCriteria(func(dueDate string, dueDateParsed time.Time) bool {
		return dueDateParsed.After(now)
	})

	if err != nil {
		t.Fatalf("QueryTodosWithDateCriteria for future tasks failed: %v", err)
	}

	if len(futureTasks) != 2 {
		t.Errorf("Expected 2 future tasks, got %d", len(futureTasks))
	}
}

func TestQueryNotesByTags(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	now := time.Now()

	// Create test files with different tags
	createTestFile(t, scripts.File{
		Name:      "work.md",
		Title:     "Work Note",
		CreatedAt: now,
		Tags:      []string{"work", "important"},
		Content:   "Work related note",
	})

	createTestFile(t, scripts.File{
		Name:      "personal.md",
		Title:     "Personal Note",
		CreatedAt: now,
		Tags:      []string{"personal", "important"},
		Content:   "Personal note",
	})

	createTestFile(t, scripts.File{
		Name:      "project.md",
		Title:     "Project Note",
		CreatedAt: now,
		Tags:      []string{"work", "project", "important"},
		Content:   "Project related note",
	})

	// Test querying by single tag
	workNotes, err := QueryNotesByTags([]string{"work"})
	if err != nil {
		t.Fatalf("QueryNotesByTags failed: %v", err)
	}

	if len(workNotes) != 2 {
		t.Errorf("Expected 2 work notes, got %d", len(workNotes))
	}

	// Test querying by multiple tags
	workProjectNotes, err := QueryNotesByTags([]string{"work", "project"})
	if err != nil {
		t.Fatalf("QueryNotesByTags failed: %v", err)
	}

	if len(workProjectNotes) != 1 {
		t.Errorf("Expected 1 work project note, got %d", len(workProjectNotes))
	}

	// Test querying by common tag
	importantNotes, err := QueryNotesByTags([]string{"important"})
	if err != nil {
		t.Fatalf("QueryNotesByTags failed: %v", err)
	}

	if len(importantNotes) != 3 {
		t.Errorf("Expected 3 important notes, got %d", len(importantNotes))
	}
}

func TestContains(t *testing.T) {
	testCases := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{[]string{"apple", "banana", "cherry"}, "banana", true},
		{[]string{"apple", "banana", "cherry"}, "orange", false},
		{[]string{" apple ", "banana", "cherry"}, "apple", true},
		{[]string{"apple", "banana", "cherry"}, " banana ", true},
		{[]string{}, "anything", false},
	}

	for _, tc := range testCases {
		result := contains(tc.slice, tc.item)
		if result != tc.expected {
			t.Errorf("contains(%v, %s) = %v; expected %v",
				tc.slice, tc.item, result, tc.expected)
		}
	}
}

func TestTimeToString(t *testing.T) {
	testTime := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
	expected := "2023-05-15"
	result := timeToString(testTime)

	if result != expected {
		t.Errorf("timeToString(%v) = %s; expected %s", testTime, result, expected)
	}
}

func TestQueryCompletedTodosByDateRange(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	now := time.Now()

	// Create test files with different dates
	// In range (completed todos)
	createTestFile(t, scripts.File{
		Name:      "completed1.md",
		Title:     "Completed Task 1",
		CreatedAt: now,
		DueAt:     time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
		Tags:      []string{"test"},
		Done:      true,
		Content:   "This is a completed task in range",
	})

	createTestFile(t, scripts.File{
		Name:      "completed2.md",
		Title:     "Completed Task 2",
		CreatedAt: now,
		DueAt:     time.Date(2023, 1, 20, 0, 0, 0, 0, time.UTC),
		Tags:      []string{"test"},
		Done:      true,
		Content:   "This is another completed task in range",
	})

	// Edge cases - inclusive boundaries
	createTestFile(t, scripts.File{
		Name:      "completed_start.md",
		Title:     "Completed Task Start",
		CreatedAt: now,
		DueAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), // Start of range
		Tags:      []string{"test"},
		Done:      true,
		Content:   "This is a completed task at start of range",
	})

	createTestFile(t, scripts.File{
		Name:      "completed_end.md",
		Title:     "Completed Task End",
		CreatedAt: now,
		DueAt:     time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC), // End of range
		Tags:      []string{"test"},
		Done:      true,
		Content:   "This is a completed task at end of range",
	})

	// Date range query note (should be excluded even though it's in range and completed)
	createTestFile(t, scripts.File{
		Name:      "date_range_query.md",
		Title:     "Date Range Query 2023-01-01 - 2023-01-31",
		CreatedAt: now,
		DueAt:     time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC), // In range
		Tags:      []string{"date-range-query"},
		Done:      true,
		Content:   "This is a date range query note that should be excluded",
	})

	// Out of range (completed todos)
	createTestFile(t, scripts.File{
		Name:      "completed_before.md",
		Title:     "Completed Task Before",
		CreatedAt: now,
		DueAt:     time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC), // Before range
		Tags:      []string{"test"},
		Done:      true,
		Content:   "This is a completed task before range",
	})

	createTestFile(t, scripts.File{
		Name:      "completed_after.md",
		Title:     "Completed Task After",
		CreatedAt: now,
		DueAt:     time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC), // After range
		Tags:      []string{"test"},
		Done:      true,
		Content:   "This is a completed task after range",
	})

	// In range but not completed
	createTestFile(t, scripts.File{
		Name:      "not_completed.md",
		Title:     "Not Completed Task",
		CreatedAt: now,
		DueAt:     time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC), // In range
		Tags:      []string{"test"},
		Done:      false,
		Content:   "This is not a completed task",
	})

	// Define date range for test
	startDate := "2023-01-01"
	endDate := "2023-01-31"

	// Test the function
	files, err := QueryCompletedTodosByDateRange(func(dueDate string, dueDateParsed time.Time) bool {
		return dueDate >= startDate && dueDate <= endDate
	})

	if err != nil {
		t.Fatalf("QueryCompletedTodosByDateRange failed: %v", err)
	}

	// Should find 4 completed todos in the date range (excluding the date range query note)
	if len(files) != 4 {
		t.Fatalf("Expected 4 files, got %d", len(files))
	}

	// Verify the expected files were returned
	expectedFiles := map[string]bool{
		"completed1.md":      true,
		"completed2.md":      true,
		"completed_start.md": true,
		"completed_end.md":   true,
	}

	for _, file := range files {
		if !expectedFiles[file.Name] {
			t.Errorf("Unexpected file in results: %s", file.Name)
		}

		// Verify each file is marked as done
		if !file.Done {
			t.Errorf("File %s should be marked as done", file.Name)
		}

		// Verify no date range query notes are included
		if isDateRangeQueryNote(&file) {
			t.Errorf("File %s is a date range query note but was included in results", file.Name)
		}
	}

	// Verify files outside the range weren't included
	unexpectedFiles := []string{
		"completed_before.md",
		"completed_after.md",
		"not_completed.md",
		"date_range_query.md", // This should be excluded explicitly
	}

	for _, fileName := range unexpectedFiles {
		found := false
		for _, file := range files {
			if file.Name == fileName {
				found = true
				break
			}
		}

		if found {
			t.Errorf("File %s should not be in results", fileName)
		}
	}
}
