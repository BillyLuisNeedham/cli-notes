package scripts

/*
Areas for improvement in get.go:

1. Case sensitivity in containsTag: ✅ FIXED
   - The containsTag function now handles case-insensitive matching consistently with fileMatchesQuery
   - Both functions now convert strings to lowercase before comparing

2. QueryAllFiles and QueryFiles inconsistency: ✅ FIXED
   - Both QueryAllFiles and QueryFiles now use AND logic consistently
   - All queries must match for a file to be included in the results

3. Date handling in GetSoonTodos: ✅ FIXED
   - Now only uses time.Time comparison for consistency and clarity
   - Removed confusing mix of string and time.Time comparisons

4. Magic number in GetTodosWithNoDueDate:
   - Uses 100 years as a magic number to determine files with no due date
   - A more explicit approach or a dedicated field would be clearer

5. Dependency injection approach:
   - While the code uses dependency injection for testability, the function signatures are complex
   - Consider simplifying with interfaces or a more standard repository pattern
*/

import (
	"testing"
	"time"
)

// TestGetTodos tests the GetTodos function
func TestGetTodos(t *testing.T) {
	// Setup mock function
	getFilesByIsDoneMock := func(isDone bool) ([]File, error) {
		if !isDone {
			return []File{
				{Name: "todo1.md", Done: false},
				{Name: "todo2.md", Done: false},
			}, nil
		}
		return []File{}, nil
	}

	// Call the function
	result, err := GetTodos(getFilesByIsDoneMock)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 todos, got %d", len(result))
	}
}

// TestQueryOpenTodos tests the QueryOpenTodos function
func TestQueryOpenTodos(t *testing.T) {
	// Setup mock data
	mockTodos := []File{
		{Name: "todo1.md", Title: "Shopping", Content: "Get groceries", Tags: []string{"personal"}, Done: false},
		{Name: "todo2.md", Title: "Work", Content: "Finish report", Tags: []string{"work"}, Done: false},
		{Name: "todo3.md", Title: "Exercise", Content: "Go for a run", Tags: []string{"health"}, Done: false},
	}

	// Setup mock function
	getFilesByIsDoneMock := func(isDone bool) ([]File, error) {
		if !isDone {
			return mockTodos, nil
		}
		return []File{}, nil
	}

	// Test cases
	testCases := []struct {
		name     string
		queries  []string
		expected int
	}{
		{"No queries", []string{}, 0},
		{"Single query match title", []string{"work"}, 1},
		{"Single query match content", []string{"groceries"}, 1},
		{"Single query match tag", []string{"health"}, 1},
		{"No matches", []string{"invalid"}, 0},
		{"Multiple queries - all should match", []string{"work", "report"}, 1},
		{"Multiple queries - no matches", []string{"work", "invalid"}, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := QueryOpenTodos(tc.queries, getFilesByIsDoneMock)

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if len(result) != tc.expected {
				t.Errorf("Expected %d todos, got %d", tc.expected, len(result))
			}
		})
	}
}

// TestQueryAllFiles tests the QueryAllFiles function
func TestQueryAllFiles(t *testing.T) {
	// Setup mock data
	mockFiles := []File{
		{Name: "file1.md", Title: "Note 1", Content: "Content 1", Tags: []string{"tag1"}},
		{Name: "file2.md", Title: "Note 2", Content: "Content 2", Tags: []string{"tag2"}},
	}

	// Setup mock function
	getFilesByQueryMock := func(query string) ([]File, error) {
		if query == "note" {
			return mockFiles, nil
		}
		return []File{}, nil
	}

	// Test cases
	testCases := []struct {
		name     string
		queries  []string
		expected int
	}{
		{"No queries", []string{}, 0},
		{"Single query", []string{"note"}, 2},
		{"Multiple queries - first returns results, second filters", []string{"note", "Note 1"}, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := QueryAllFiles(tc.queries, getFilesByQueryMock)

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if len(result) != tc.expected {
				t.Errorf("Expected %d files, got %d", tc.expected, len(result))
			}
		})
	}
}

// TestQueryFiles tests the QueryFiles function
func TestQueryFiles(t *testing.T) {
	// Setup mock data
	mockFiles := []File{
		{Name: "file1.md", Title: "Note 1", Content: "Content 1", Tags: []string{"tag1"}},
		{Name: "file2.md", Title: "Note 2", Content: "Content 2", Tags: []string{"tag2"}},
		{Name: "file3.md", Title: "Important Note", Content: "Priority task", Tags: []string{"important", "priority"}},
	}

	// Test cases
	testCases := []struct {
		name     string
		queries  []string
		expected int
	}{
		{"No queries", []string{}, 0},
		{"Single query match title", []string{"Note 1"}, 1},
		{"Single query match content", []string{"Priority"}, 1},
		{"Single query match tag", []string{"tag2"}, 1},
		{"Multiple queries - AND logic", []string{"Important", "Priority"}, 1},
		{"Multiple queries - no match", []string{"Important", "Missing"}, 0},
		{"No matches", []string{"invalid"}, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := QueryFiles(tc.queries, mockFiles)

			if len(result) != tc.expected {
				t.Errorf("Expected %d files, got %d", tc.expected, len(result))
			}
		})
	}
}

// TestSearchNotesByTags tests the SearchNotesByTags function
func TestSearchNotesByTags(t *testing.T) {
	// Setup mock data
	mockFiles := []File{
		{Name: "file1.md", Tags: []string{"tag1", "common"}},
		{Name: "file2.md", Tags: []string{"tag2", "common"}},
	}

	// Setup mock function
	getFilesByTagMock := func(tags []string) ([]File, error) {
		if contains(tags, "tag1") {
			return []File{mockFiles[0]}, nil
		}
		if contains(tags, "common") {
			return mockFiles, nil
		}
		return []File{}, nil
	}

	// Test cases
	testCases := []struct {
		name     string
		tags     []string
		expected int
	}{
		{"Single tag match one file", []string{"tag1"}, 1},
		{"Single tag match multiple files", []string{"common"}, 2},
		{"No matches", []string{"invalid"}, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := SearchNotesByTags(tc.tags, getFilesByTagMock)

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if len(result) != tc.expected {
				t.Errorf("Expected %d files, got %d", tc.expected, len(result))
			}
		})
	}
}

// Helper function for the tests
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// TestGetUncompletedTasksInFiles tests the GetUncompletedTasksInFiles function
func TestGetUncompletedTasksInFiles(t *testing.T) {
	// Setup test cases
	testCases := []struct {
		name     string
		files    []File
		expected int
	}{
		{
			name:     "Empty files",
			files:    []File{},
			expected: 0,
		},
		{
			name: "Files with no tasks",
			files: []File{
				{Name: "file1.md", Content: "Just some content\nNo tasks here"},
			},
			expected: 0,
		},
		{
			name: "Files with uncompleted tasks",
			files: []File{
				{Name: "file1.md", Content: "- [ ] Task 1\n- [x] Completed task\n- [ ] Task 2"},
			},
			expected: 2,
		},
		{
			name: "Multiple files with tasks",
			files: []File{
				{Name: "file1.md", Content: "- [ ] Task 1"},
				{Name: "file2.md", Content: "- [ ] Task 2\n- [ ] Task 3"},
			},
			expected: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GetUncompletedTasksInFiles(tc.files)

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if len(result) != tc.expected {
				t.Errorf("Expected %d tasks, got %d", tc.expected, len(result))
			}
		})
	}
}

// TestDateFunctions tests the date-related functions
func TestDateFunctions(t *testing.T) {
	// Current date for testing
	now := time.Now()

	// Setup test files
	overdue := File{Name: "overdue.md", DueAt: now.AddDate(0, 0, -1)}
	today := File{Name: "today.md", DueAt: now}
	thisWeek := File{Name: "thisWeek.md", DueAt: now.AddDate(0, 0, 5)}
	future := File{Name: "future.md", DueAt: now.AddDate(0, 1, 0)}
	noDueDate := File{Name: "noDueDate.md", DueAt: now.AddDate(101, 0, 0)}

	// Mock function for GetOverdueTodos
	getFilesForOverdue := func(dateQuery DateQuery) ([]File, error) {
		var results []File
		files := []File{overdue, today, thisWeek, future, noDueDate}

		for _, file := range files {
			dueDateStr := file.DueAt.Format("2006-01-02")
			if dateQuery(dueDateStr, file.DueAt) {
				results = append(results, file)
			}
		}

		return results, nil
	}

	// Test GetOverdueTodos
	t.Run("GetOverdueTodos", func(t *testing.T) {
		results, err := GetOverdueTodos(getFilesForOverdue)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Should include overdue and today
		if len(results) != 2 {
			t.Errorf("Expected 2 files, got %d", len(results))
		}
	})

	// Test GetSoonTodos
	t.Run("GetSoonTodos", func(t *testing.T) {
		results, err := GetSoonTodos(getFilesForOverdue)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Should include overdue, today, and thisWeek
		if len(results) != 3 {
			t.Errorf("Expected 3 files, got %d", len(results))
		}
	})

	// Test GetTodosWithNoDueDate
	t.Run("GetTodosWithNoDueDate", func(t *testing.T) {
		results, err := GetTodosWithNoDueDate(getFilesForOverdue)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Should only include noDueDate
		if len(results) != 1 {
			t.Errorf("Expected 1 file, got %d", len(results))
		}
	})
}

// TestFileMatchesQuery tests the fileMatchesQuery function
func TestFileMatchesQuery(t *testing.T) {
	// Setup test file
	file := File{
		Name:    "project-notes.md",
		Title:   "Project Planning Notes",
		Content: "This is content about the project.",
		Tags:    []string{"project", "planning", "important"},
	}

	// Test cases
	testCases := []struct {
		name     string
		query    string
		expected bool
	}{
		{"Match in name", "project", true},
		{"Match in title", "Planning", true},
		{"Match in content", "content", true},
		{"Match in tags", "important", true},
		{"Case insensitive", "PROJECT", true},
		{"No match", "invalid", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := fileMatchesQuery(file, tc.query)

			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestContainsTag tests the containsTag function
func TestContainsTag(t *testing.T) {
	// Setup tags
	tags := []string{"project", "planning", "important"}

	// Test cases
	testCases := []struct {
		name     string
		query    string
		expected bool
	}{
		{"Exact match", "project", true},
		{"Partial match", "plan", true},
		{"Case insensitive - lowercase query", "important", true},
		{"Case insensitive - mixed case query", "pLaNnInG", true},
		{"No match", "invalid", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := containsTag(tags, tc.query)

			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// Helper function to check if a file has a specific tag
func fileHasTag(file File, tagName string) bool {
	for _, tag := range file.Tags {
		if tag == tagName {
			return true
		}
	}
	return false
}

// TestGetCompletedTodosByDateRange tests the GetCompletedTodosByDateRange function
func TestGetCompletedTodosByDateRange(t *testing.T) {
	// Setup dates for testing
	startDate := "2023-01-01"
	endDate := "2023-01-31"

	// Setup test files with different dates
	inRangeDate1 := "2023-01-15"
	inRangeDate2 := "2023-01-31" // Edge case - end date inclusive
	inRangeDate3 := "2023-01-01" // Edge case - start date inclusive
	beforeRangeDate := "2022-12-31"
	afterRangeDate := "2023-02-01"

	// Parse dates for the mock function
	parseDate := func(dateStr string) time.Time {
		t, _ := time.Parse("2006-01-02", dateStr)
		return t
	}

	// Setup mock function
	getFilesByDateRangeQueryMock := func(dateQuery DateQuery) ([]File, error) {
		// Create test files with various dates
		testFiles := []File{
			{Name: "in-range-1.md", DueAt: parseDate(inRangeDate1), Done: true, Tags: []string{"task"}},
			{Name: "in-range-2.md", DueAt: parseDate(inRangeDate2), Done: true, Tags: []string{"task"}},
			{Name: "in-range-3.md", DueAt: parseDate(inRangeDate3), Done: true, Tags: []string{"task"}},
			{Name: "before-range.md", DueAt: parseDate(beforeRangeDate), Done: true, Tags: []string{"task"}},
			{Name: "after-range.md", DueAt: parseDate(afterRangeDate), Done: true, Tags: []string{"task"}},
			// Add a date range query note that is in range and completed
			{Name: "date-range-query.md", DueAt: parseDate(inRangeDate1), Done: true, Tags: []string{"date-range-query"}},
		}

		// Filter files based on date query and exclude date range query notes
		var matchingFiles []File
		for _, file := range testFiles {
			if dateQuery(file.DueAt.Format("2006-01-02"), file.DueAt) {
				// Exclude date range query notes
				if !fileHasTag(file, "date-range-query") {
					matchingFiles = append(matchingFiles, file)
				}
			}
		}

		return matchingFiles, nil
	}

	// Call the function
	result, err := GetCompletedTodosByDateRange(startDate, endDate, getFilesByDateRangeQueryMock)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 todos in date range, got %d", len(result))
	}

	// Verify the correct files were returned
	fileNames := make([]string, len(result))
	for i, file := range result {
		fileNames[i] = file.Name
	}

	expectedFileNames := []string{"in-range-1.md", "in-range-2.md", "in-range-3.md"}
	for _, name := range expectedFileNames {
		if !containsString(fileNames, name) {
			t.Errorf("Expected file %s to be in results, but it wasn't", name)
		}
	}

	unexpectedFileNames := []string{"before-range.md", "after-range.md", "date-range-query.md"}
	for _, name := range unexpectedFileNames {
		if containsString(fileNames, name) {
			t.Errorf("File %s should not be in results, but it was", name)
		}
	}
}

// Helper function to check if a string slice contains a specific string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
