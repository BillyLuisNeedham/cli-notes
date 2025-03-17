package scripts

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDelayDueDatePreservesContent(t *testing.T) {
	// Setup: Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "notes-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initial content
	initialContent := `---
title: Test Delay Bug
date-created: 2023-03-05
tags: [test, bug]
date-due: 2023-03-05
done: false
---

# Test Delay Bug

This is the initial content.`

	// Create initial file on disk
	fileName := "test-delay-bug.md"
	filePath := filepath.Join(tempDir, fileName)
	err = os.WriteFile(filePath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}

	// Step 1: Read the file as it would happen in the application
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Initial File object as it would be loaded by the application
	initialFileObj := File{
		Name:      fileName,
		Title:     "Test Delay Bug",
		CreatedAt: time.Date(2023, 3, 5, 0, 0, 0, 0, time.UTC),
		DueAt:     time.Date(2023, 3, 5, 0, 0, 0, 0, time.UTC),
		Tags:      []string{"test", "bug"},
		Done:      false,
	}

	// Read content from file - this simulates file selection
	var selectedFileContent strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		selectedFileContent.WriteString(line)
		selectedFileContent.WriteString("\n")
	}
	initialFileObj.Content = selectedFileContent.String()

	// Step 2: Simulate user editing the file in the editor by writing to disk
	updatedContent := initialContent + "\nThis is new content that was added."
	err = os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write updated file: %v", err)
	}

	// Function to mock reading the latest content from a file
	// This is what our new readLatestFileContent function should do
	originalReadLatestFileContent := readLatestFileContent
	defer func() { readLatestFileContent = originalReadLatestFileContent }()

	readLatestFileContent = func(file File) (File, error) {
		updatedFile := file

		// Read the content from our test file
		content, err := os.ReadFile(filePath)
		if err != nil {
			return file, err
		}

		updatedFile.Content = string(content)
		return updatedFile, nil
	}

	// Function to mock writing to file
	finalContent := ""
	mockWriteFile := func(file File) error {
		finalContent = file.Content
		meta := `---
title: Test Delay Bug
date-created: 2023-03-05
tags: [test, bug]
date-due: 2023-03-06
done: false
---

`
		content := strings.TrimLeft(file.Content, "\n")
		return os.WriteFile(filePath, []byte(meta+content), 0644)
	}

	// The bug should now be fixed - initialFileObj will get updated content
	err = DelayDueDate(1, initialFileObj, mockWriteFile)
	if err != nil {
		t.Fatalf("DelayDueDate returned an error: %v", err)
	}

	// Check if the update was preserved
	if strings.Contains(finalContent, "This is new content that was added.") {
		t.Log("Bug fixed: The update to the note was preserved")
	} else {
		t.Error("Bug still exists: The update to the note was lost when delaying it")
	}
}

func TestSetDueDateToNextDay(t *testing.T) {
	// Setup a mock file
	mockFile := File{
		Name:      "test-next-day.md",
		Title:     "Test Next Day",
		Tags:      []string{"test"},
		CreatedAt: time.Date(2023, 3, 5, 0, 0, 0, 0, time.UTC),
		DueAt:     time.Date(2023, 3, 5, 0, 0, 0, 0, time.UTC),
		Done:      false,
		Content:   "# Test Next Day\n\nThis is test content.",
	}

	// Mock the readLatestFileContent function
	originalReadLatestFileContent := readLatestFileContent
	defer func() { readLatestFileContent = originalReadLatestFileContent }()

	readLatestFileContent = func(file File) (File, error) {
		return file, nil
	}

	// Test cases for different days of the week
	tests := []struct {
		name      string
		mockNow   time.Time
		dayOfWeek time.Weekday
		expected  time.Time
	}{
		{
			name:      "Monday from Sunday",
			mockNow:   time.Date(2023, 7, 9, 12, 0, 0, 0, time.UTC), // Sunday
			dayOfWeek: time.Monday,
			expected:  time.Date(2023, 7, 10, 12, 0, 0, 0, time.UTC), // Next day (Monday)
		},
		{
			name:      "Monday from Monday",
			mockNow:   time.Date(2023, 7, 10, 12, 0, 0, 0, time.UTC), // Monday
			dayOfWeek: time.Monday,
			expected:  time.Date(2023, 7, 17, 12, 0, 0, 0, time.UTC), // Next week's Monday
		},
		{
			name:      "Friday from Wednesday",
			mockNow:   time.Date(2023, 7, 12, 12, 0, 0, 0, time.UTC), // Wednesday
			dayOfWeek: time.Friday,
			expected:  time.Date(2023, 7, 14, 12, 0, 0, 0, time.UTC), // Coming Friday
		},
		{
			name:      "Wednesday from Thursday",
			mockNow:   time.Date(2023, 7, 13, 12, 0, 0, 0, time.UTC), // Thursday
			dayOfWeek: time.Wednesday,
			expected:  time.Date(2023, 7, 19, 12, 0, 0, 0, time.UTC), // Next week's Wednesday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original timeNow function and restore after test
			originalTimeNow := timeNow
			defer func() { timeNow = originalTimeNow }()

			// Override timeNow for this test
			timeNow = func() time.Time {
				return tt.mockNow
			}

			var updatedFile File
			mockWriteFile := func(file File) error {
				updatedFile = file
				return nil
			}

			err := SetDueDateToNextDay(tt.dayOfWeek, mockFile, mockWriteFile)
			if err != nil {
				t.Fatalf("SetDueDateToNextDay returned an error: %v", err)
			}

			// Check if the due date was set correctly
			if !updatedFile.DueAt.Equal(tt.expected) {
				t.Errorf("Expected due date %v, got %v", tt.expected, updatedFile.DueAt)
			}
		})
	}
}

func TestDelayDueDatePreservesPriority(t *testing.T) {
	// Save original functions and restore after test
	originalReadLatestFileContent := readLatestFileContent
	originalTimeNow := timeNow
	defer func() { 
		readLatestFileContent = originalReadLatestFileContent 
		timeNow = originalTimeNow
	}()

	// Test with a specific date
	mockTime := time.Date(2023, 5, 15, 10, 0, 0, 0, time.UTC)
	timeNow = func() time.Time {
		return mockTime
	}

	// Setup a file with priority 1
	originalFile := File{
		Name:      "test-priority.md",
		Title:     "Test Priority Preservation",
		Tags:      []string{"test", "priority"},
		CreatedAt: time.Date(2023, 5, 10, 0, 0, 0, 0, time.UTC),
		DueAt:     time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
		Done:      false,
		Content:   "# Test Priority\n\nThis tests that priority is preserved.",
		Priority:  P1, // Set to highest priority
	}

	// Mock the readLatestFileContent function to return a file without priority
	readLatestFileContent = func(file File) (File, error) {
		// Return file without setting priority, to simulate the bug
		return File{
			Name:      file.Name,
			Title:     file.Title,
			Tags:      file.Tags,
			CreatedAt: file.CreatedAt,
			DueAt:     file.DueAt,
			Done:      file.Done,
			Content:   file.Content,
			// Priority intentionally not set to mimic bug
		}, nil
	}

	var writtenFile File
	mockWriteFile := func(file File) error {
		writtenFile = file
		return nil
	}

	// Test delaying by 1 day
	err := DelayDueDate(1, originalFile, mockWriteFile)
	if err != nil {
		t.Fatalf("DelayDueDate returned an error: %v", err)
	}

	// Verify due date was updated correctly
	expectedDueDate := mockTime.AddDate(0, 0, 1)
	if !writtenFile.DueAt.Equal(expectedDueDate) {
		t.Errorf("Expected due date %v, got %v", expectedDueDate, writtenFile.DueAt)
	}

	// Verify priority was preserved
	if writtenFile.Priority != P1 {
		t.Errorf("Expected priority P1, got %v", writtenFile.Priority)
	}
	
	// Test with SetDueDateToToday
	err = SetDueDateToToday(originalFile, mockWriteFile)
	if err != nil {
		t.Fatalf("SetDueDateToToday returned an error: %v", err)
	}
	
	// Verify priority was preserved
	if writtenFile.Priority != P1 {
		t.Errorf("Expected priority P1, got %v", writtenFile.Priority)
	}
	
	// Test with SetDueDateToNextDay
	err = SetDueDateToNextDay(time.Monday, originalFile, mockWriteFile)
	if err != nil {
		t.Fatalf("SetDueDateToNextDay returned an error: %v", err)
	}
	
	// Verify priority was preserved
	if writtenFile.Priority != P1 {
		t.Errorf("Expected priority P1, got %v", writtenFile.Priority)
	}
}

func TestReadLatestFileContentReadsPriority(t *testing.T) {
	// Setup: Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "notes-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initial content with priority: 1
	fileContent := `---
title: Test Priority Reading
date-created: 2023-03-05
tags: [test, priority]
priority: 1
date-due: 2023-03-05
done: false
---

# Test Priority Reading

This file has priority 1 in its metadata.`

	// Create file on disk
	fileName := "test-priority-reading.md"
	filePath := filepath.Join(tempDir, fileName)
	err = os.WriteFile(filePath, []byte(fileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Save original function and replace it with our test implementation
	originalReadLatestFileContent := readLatestFileContent
	defer func() { readLatestFileContent = originalReadLatestFileContent }()

	// Create a test notes directory
	notesDir := filepath.Join(tempDir, "notes")
	err = os.Mkdir(notesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create notes directory: %v", err)
	}

	// Copy test file to notes directory
	noteFilePath := filepath.Join(notesDir, fileName)
	noteFileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	err = os.WriteFile(noteFilePath, noteFileContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write to notes directory: %v", err)
	}

	// Create a custom readLatestFileContent that uses our temp directory
	readLatestFileContent = func(file File) (File, error) {
		// Override current directory to our temp directory
		oldWd, err := os.Getwd()
		if err != nil {
			return file, err
		}
		defer os.Chdir(oldWd) // Restore original working directory
		
		err = os.Chdir(tempDir)
		if err != nil {
			return file, err
		}
		
		// Now call the original implementation we just modified
		return originalReadLatestFileContent(file)
	}

	// Create a test file object without priority
	testFile := File{
		Name:      fileName,
		Title:     "Test Priority Reading",
		Tags:      []string{"test", "priority"},
		CreatedAt: time.Date(2023, 3, 5, 0, 0, 0, 0, time.UTC),
		DueAt:     time.Date(2023, 3, 5, 0, 0, 0, 0, time.UTC),
		Done:      false,
		// Priority intentionally not set
	}

	// Call readLatestFileContent
	updatedFile, err := readLatestFileContent(testFile)
	if err != nil {
		t.Fatalf("readLatestFileContent returned an error: %v", err)
	}

	// Verify priority was correctly read from file
	if updatedFile.Priority != P1 {
		t.Errorf("Expected priority P1, got %v", updatedFile.Priority)
	}
}
