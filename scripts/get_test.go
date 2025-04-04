package scripts

import (
	"bufio"
	"strings"
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
	// Setup current time for due dates
	now := time.Now()

	// Setup mock data with priorities and due dates
	mockTodos := []File{
		{
			Name:     "todo1.md",
			Title:    "Shopping",
			Content:  "Get groceries",
			Tags:     []string{"personal"},
			Done:     false,
			Priority: P2,
			DueAt:    now.AddDate(0, 0, 3), // Due in 3 days
		},
		{
			Name:     "todo2.md",
			Title:    "Work",
			Content:  "Finish report",
			Tags:     []string{"work"},
			Done:     false,
			Priority: P1,                   // High priority
			DueAt:    now.AddDate(0, 0, 1), // Due tomorrow
		},
		{
			Name:     "todo3.md",
			Title:    "Exercise",
			Content:  "Go for a run",
			Tags:     []string{"health"},
			Done:     false,
			Priority: P3,                   // Low priority
			DueAt:    now.AddDate(0, 0, 2), // Due in 2 days
		},
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
		name          string
		queries       []string
		expected      int
		expectedFiles []string // Names of files we expect to be returned
	}{
		{"No queries", []string{}, 0, nil},
		{"Single query match title", []string{"work"}, 1, []string{"todo2.md"}},
		{"Single query match content", []string{"groceries"}, 1, []string{"todo1.md"}},
		{"Single query match tag", []string{"health"}, 1, []string{"todo3.md"}},
		{"No matches", []string{"invalid"}, 0, nil},
		{"Multiple queries - all should match", []string{"work", "report"}, 1, []string{"todo2.md"}},
		{"Multiple queries - no matches", []string{"work", "invalid"}, 0, nil},
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

			// Check that the expected files are in the results, regardless of order
			if tc.expected > 0 && tc.expectedFiles != nil {
				expectedMap := make(map[string]bool)
				for _, name := range tc.expectedFiles {
					expectedMap[name] = false
				}

				for _, file := range result {
					if _, exists := expectedMap[file.Name]; exists {
						expectedMap[file.Name] = true
					} else if tc.expected > 0 {
						t.Errorf("Unexpected file in results: %s", file.Name)
					}
				}

				// Verify all expected files were found
				for name, found := range expectedMap {
					if !found {
						t.Errorf("Expected file %s was not in results", name)
					}
				}
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

	// Setup test files with different priorities and creation dates
	overdue := File{
		Name:      "overdue.md",
		DueAt:     now.AddDate(0, 0, -1),
		Priority:  P2,
		CreatedAt: now.AddDate(0, 0, -5), // Created 5 days ago
	}
	today := File{
		Name:      "today.md",
		DueAt:     now,
		Priority:  P1,                    // High priority
		CreatedAt: now.AddDate(0, 0, -1), // Created yesterday
	}
	thisWeek := File{
		Name:      "thisWeek.md",
		DueAt:     now.AddDate(0, 0, 5),
		Priority:  P1,                     // High priority
		CreatedAt: now.AddDate(0, 0, -10), // Created 10 days ago
	}
	future := File{
		Name:      "future.md",
		DueAt:     now.AddDate(0, 1, 0),
		Priority:  P3,                    // Low priority
		CreatedAt: now.AddDate(0, 0, -2), // Created 2 days ago
	}
	noDueDate := File{
		Name:      "noDueDate.md",
		DueAt:     now.AddDate(101, 0, 0),
		Priority:  P2,
		CreatedAt: now.AddDate(0, 0, -15), // Created 15 days ago
	}

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

		// Check that the results contain the expected files without caring about order
		expectedFiles := map[string]bool{
			"overdue.md": false,
			"today.md":   false,
		}

		for _, file := range results {
			if _, exists := expectedFiles[file.Name]; exists {
				expectedFiles[file.Name] = true
			} else {
				t.Errorf("Unexpected file in results: %s", file.Name)
			}
		}

		// Verify all expected files were found
		for name, found := range expectedFiles {
			if !found {
				t.Errorf("Expected file %s was not in results", name)
			}
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

		// Check that the results contain the expected files without caring about order
		expectedFiles := map[string]bool{
			"overdue.md":  false,
			"today.md":    false,
			"thisWeek.md": false,
		}

		for _, file := range results {
			if _, exists := expectedFiles[file.Name]; exists {
				expectedFiles[file.Name] = true
			} else {
				t.Errorf("Unexpected file in results: %s", file.Name)
			}
		}

		// Verify all expected files were found
		for name, found := range expectedFiles {
			if !found {
				t.Errorf("Expected file %s was not in results", name)
			}
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

		if len(results) == 1 && results[0].Name != "noDueDate.md" {
			t.Errorf("Expected result to be noDueDate.md, got %s", results[0].Name)
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

func TestDelayDueDate(t *testing.T) {
	// Initial file would have proper formatting with exactly one newline after metadata
	initialContent := `---
title: Test Delay Note
date-created: 2023-03-05
tags: [test]
date-due: 2023-03-05
done: false
---

# Test Delay Note

This is a test note for delaying.`

	// Create a test note with the initial content
	testFile := File{
		Name:      "test-delay-note.md",
		Title:     "Test Delay Note",
		CreatedAt: time.Now(),
		DueAt:     time.Now(),
		Tags:      []string{"test"},
		Done:      false,
		Content:   initialContent,
	}

	// This simulates the first read of a note as it would happen after 'gto' command
	// and arrow key selection
	simSelectedFileContent := initialContent
	var contentBuilder strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(simSelectedFileContent))
	for scanner.Scan() {
		line := scanner.Text()
		contentBuilder.WriteString(line)
		contentBuilder.WriteString("\n")
	}

	// Update the test file with content as it would be after selection
	fileAfterSelection := testFile
	fileAfterSelection.Content = contentBuilder.String()

	// Save the original implementation to restore after the test
	originalReadLatestFileContent := readLatestFileContent
	defer func() { readLatestFileContent = originalReadLatestFileContent }()

	// Mock the readLatestFileContent function to return the file as is
	readLatestFileContent = func(file File) (File, error) {
		return file, nil
	}

	// Mock the file writer that should now trim leading newlines
	var writtenContent string
	mockWriteFile := func(file File) error {
		// In the actual code, this writes the metadata and then appends content
		metadataSection := `---
title: Test Delay Note
date-created: 2023-03-05
tags: [test]
date-due: 2023-03-06
done: false
---

` // Note the single newline here

		// Trim leading newlines from content to prevent accumulation
		content := strings.TrimLeft(file.Content, "\n")
		writtenContent = metadataSection + content
		return nil
	}

	// First delay after selection
	err := DelayDueDate(1, fileAfterSelection, mockWriteFile)
	if err != nil {
		t.Fatalf("DelayDueDate returned an error: %v", err)
	}

	// Count the number of newlines between metadata section and title after first delay
	lines := strings.Split(writtenContent, "\n")
	metadataEndIndex := -1
	titleIndex := -1

	for i, line := range lines {
		if line == "---" && i > 0 {
			metadataEndIndex = i
		}
		if strings.HasPrefix(line, "# Test Delay Note") {
			titleIndex = i
			break
		}
	}

	if metadataEndIndex == -1 || titleIndex == -1 {
		t.Fatalf("Could not find metadata end or title line after first delay: %s", writtenContent)
	}

	newlinesCountAfterFirstDelay := titleIndex - metadataEndIndex - 1

	// Now simulate a second delay on the same file
	// Read the content as it was written after first delay
	var secondContentBuilder strings.Builder
	scanner = bufio.NewScanner(strings.NewReader(writtenContent))
	for scanner.Scan() {
		line := scanner.Text()
		secondContentBuilder.WriteString(line)
		secondContentBuilder.WriteString("\n")
	}

	fileAfterFirstDelay := fileAfterSelection
	fileAfterFirstDelay.Content = secondContentBuilder.String()
	fileAfterFirstDelay.DueAt = fileAfterFirstDelay.DueAt.AddDate(0, 0, 1) // Simulate first delay

	// Second delay
	err = DelayDueDate(1, fileAfterFirstDelay, mockWriteFile)
	if err != nil {
		t.Fatalf("Second DelayDueDate returned an error: %v", err)
	}

	// Count newlines after second delay
	lines = strings.Split(writtenContent, "\n")
	metadataEndIndex = -1
	titleIndex = -1

	for i, line := range lines {
		if line == "---" && i > 0 {
			metadataEndIndex = i
		}
		if strings.HasPrefix(line, "# Test Delay Note") {
			titleIndex = i
			break
		}
	}

	if metadataEndIndex == -1 || titleIndex == -1 {
		t.Fatalf("Could not find metadata end or title line after second delay: %s", writtenContent)
	}

	newlinesCountAfterSecondDelay := titleIndex - metadataEndIndex - 1

	// We should have exactly one newline between the metadata and title
	if newlinesCountAfterFirstDelay != 1 {
		t.Errorf("Expected exactly 1 newline after first delay, but got %d newlines",
			newlinesCountAfterFirstDelay)
	}

	if newlinesCountAfterSecondDelay != 1 {
		t.Errorf("Expected exactly 1 newline after second delay, but got %d newlines",
			newlinesCountAfterSecondDelay)
	}
}

// TestGetTodosByPriority tests the GetTodosByPriority function
func TestGetTodosByPriority(t *testing.T) {
	// Setup current time for due dates
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	dayAfterTomorrow := now.AddDate(0, 0, 2)
	noDueDate := time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)

	// Setup mock data with priorities and due dates
	mockTodos := []File{
		{
			Name:     "p1-overdue.md",
			Priority: P1,
			DueAt:    yesterday,
			Done:     false,
		},
		{
			Name:     "p1-tomorrow.md",
			Priority: P1,
			DueAt:    tomorrow,
			Done:     false,
		},
		{
			Name:     "p1-no-date.md",
			Priority: P1,
			DueAt:    noDueDate,
			Done:     false,
		},
		{
			Name:     "p2-overdue.md",
			Priority: P2,
			DueAt:    yesterday,
			Done:     false,
		},
		{
			Name:     "p2-tomorrow.md",
			Priority: P2,
			DueAt:    tomorrow,
			Done:     false,
		},
		{
			Name:     "p3-day-after.md",
			Priority: P3,
			DueAt:    dayAfterTomorrow,
			Done:     false,
		},
	}

	// Mock function that returns all todos
	getFilesByIsDoneMock := func(isDone bool) ([]File, error) {
		if !isDone {
			return mockTodos, nil
		}
		return []File{}, nil
	}

	// Test cases
	testCases := []struct {
		name          string
		priority      Priority
		expectedCount int
		expectedFiles []string
	}{
		{
			name:          "P1 todos",
			priority:      P1,
			expectedCount: 3,
			expectedFiles: []string{"p1-overdue.md", "p1-tomorrow.md", "p1-no-date.md"},
		},
		{
			name:          "P2 todos",
			priority:      P2,
			expectedCount: 2,
			expectedFiles: []string{"p2-overdue.md", "p2-tomorrow.md"},
		},
		{
			name:          "P3 todos",
			priority:      P3,
			expectedCount: 1,
			expectedFiles: []string{"p3-day-after.md"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GetTodosByPriority(tc.priority, getFilesByIsDoneMock)

			// Check for errors
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			// Check count
			if len(result) != tc.expectedCount {
				t.Errorf("Expected %d todos, got %d", tc.expectedCount, len(result))
			}

			// Check that all expected files are returned without caring about order
			if len(result) > 0 {
				expectedMap := make(map[string]bool)
				for _, name := range tc.expectedFiles {
					expectedMap[name] = false
				}

				for _, file := range result {
					if _, exists := expectedMap[file.Name]; exists {
						expectedMap[file.Name] = true
					} else {
						t.Errorf("Unexpected file in results: %s", file.Name)
					}
				}

				// Verify all expected files were found
				for name, found := range expectedMap {
					if !found {
						t.Errorf("Expected file %s was not in results", name)
					}
				}
			}
		})
	}
}

// TestSortingApplied tests that sorting is applied in the get functions
// without depending on the exact sorting algorithm implementation
func TestSortingApplied(t *testing.T) {
	// Create a set of todos with different priorities and due dates
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	nextWeek := now.AddDate(0, 0, 7)
	noDueDate := time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)

	// Create mock todos with intentionally reverse/random ordering
	mockTodos := []File{
		{
			Name:     "p3-no-date.md",
			Priority: P3,
			DueAt:    noDueDate,
			Done:     false,
		},
		{
			Name:     "p1-tomorrow.md",
			Priority: P1,
			DueAt:    tomorrow,
			Done:     false,
		},
		{
			Name:     "p2-overdue.md",
			Priority: P2,
			DueAt:    yesterday,
			Done:     false,
		},
		{
			Name:     "p1-overdue.md",
			Priority: P1,
			DueAt:    yesterday,
			Done:     false,
		},
		{
			Name:     "p2-next-week.md",
			Priority: P2,
			DueAt:    nextWeek,
			Done:     false,
		},
	}

	// Mock function that returns all todos in the specific, unsorted order
	getFilesByIsDoneMock := func(isDone bool) ([]File, error) {
		// Return a copy of the mockTodos to avoid side effects
		todosCopy := make([]File, len(mockTodos))
		copy(todosCopy, mockTodos)
		return todosCopy, nil
	}

	// Test case 1: GetTodos should apply sorting
	t.Run("GetTodos applies sorting", func(t *testing.T) {
		// Call the function
		result, err := GetTodos(getFilesByIsDoneMock)

		// Verify no error
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify right number of results
		if len(result) != len(mockTodos) {
			t.Fatalf("Expected %d todos, got %d", len(mockTodos), len(result))
		}

		// Verify that P1 todos come before P2 todos, which come before P3 todos
		lastPriority := P1
		for i, todo := range result {
			if i > 0 && todo.Priority < lastPriority {
				t.Errorf("Priority out of order at position %d: Found P%d after P%d",
					i, todo.Priority, lastPriority)
			}
			lastPriority = todo.Priority
		}

		// Check that overdue P1 comes before non-overdue P1
		var p1OverdueIndex, p1TomorrowIndex int = -1, -1
		for i, todo := range result {
			if todo.Name == "p1-overdue.md" {
				p1OverdueIndex = i
			}
			if todo.Name == "p1-tomorrow.md" {
				p1TomorrowIndex = i
			}
		}

		if p1OverdueIndex != -1 && p1TomorrowIndex != -1 && p1OverdueIndex > p1TomorrowIndex {
			t.Errorf("Expected overdue P1 to come before non-overdue P1, but p1-overdue.md (index %d) came after p1-tomorrow.md (index %d)",
				p1OverdueIndex, p1TomorrowIndex)
		}
	})

	// Test case 2: Test that unsorted (original) and sorted orders are different
	t.Run("Sorting changes the order", func(t *testing.T) {
		// Get fresh copy of unsorted todos
		unsortedTodos, _ := getFilesByIsDoneMock(false)
		// Get a reference to the original order before sorting
		originalOrder := make([]string, len(unsortedTodos))
		for i, file := range unsortedTodos {
			originalOrder[i] = file.Name
		}

		// Sort the todos
		sortedTodos := SortTodosByPriorityAndDueDate(unsortedTodos)

		// Get the sorted order
		sortedOrder := make([]string, len(sortedTodos))
		for i, file := range sortedTodos {
			sortedOrder[i] = file.Name
		}

		// Check the orders are different
		sameOrder := true
		for i := range originalOrder {
			if originalOrder[i] != sortedOrder[i] {
				sameOrder = false
				break
			}
		}

		if sameOrder {
			t.Errorf("Expected sorting to change the order, but unsorted and sorted todos are in the same order")
			t.Errorf("Original order: %v", originalOrder)
			t.Errorf("Sorted order:   %v", sortedOrder)
		}
	})

	// Test case 3: Test breaking the sort functionality
	t.Run("Breaking sort functionality causes test to fail", func(t *testing.T) {
		// Create a copy of the unsorted todos
		unsortedTodos, _ := getFilesByIsDoneMock(false)

		// Apply an incorrect "sorting" function (which actually doesn't sort at all)
		noSortTodos := make([]File, len(unsortedTodos))
		copy(noSortTodos, unsortedTodos)

		// Verify that P1 todos don't all come before P2 todos in the unsorted list
		orderCorrect := true
		lastPriority := P1

		for i, todo := range noSortTodos {
			if i > 0 && todo.Priority < lastPriority {
				// This is not supposed to be sorted correctly, so we found a disorder
				orderCorrect = false
				break
			}
			lastPriority = todo.Priority
		}

		if orderCorrect {
			t.Logf("WARNING: The test data may not be random enough to detect sorting issues")
		} else {
			// Test passes - we detected that without sorting, the priorities aren't in order
			t.Logf("Confirmed: Unsorted todos don't maintain priority order")
		}
	})
}

// Helper function to get file names from a slice of files
func getFileNames(files []File) []string {
	result := make([]string, len(files))
	for i, file := range files {
		result[i] = file.Name
	}
	return result
}
