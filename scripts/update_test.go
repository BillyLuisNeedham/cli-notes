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
