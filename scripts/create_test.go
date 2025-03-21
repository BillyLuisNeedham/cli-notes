package scripts

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCreateTodo(t *testing.T) {
	testCases := []struct {
		name          string
		title         string
		expectedError bool
		onCreated     OnFileCreated
	}{
		{
			name:          "successful creation",
			title:         "test-todo",
			expectedError: false,
			onCreated: func(file File) error {
				return nil
			},
		},
		{
			name:          "creation error",
			title:         "error-todo",
			expectedError: true,
			onCreated: func(file File) error {
				return errors.New("mock error")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := CreateTodo(tc.title, tc.onCreated)

			if tc.expectedError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tc.expectedError {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if file.Title != tc.title {
					t.Errorf("Expected title %s, got %s", tc.title, file.Title)
				}

				if !containsTag(file.Tags, "todo") {
					t.Errorf("Expected todo tag, got %v", file.Tags)
				}

				if file.Done {
					t.Error("Expected todo to not be done")
				}
			}
		})
	}
}

func TestCreateMeeting(t *testing.T) {
	testCases := []struct {
		name          string
		title         string
		expectedError bool
		onCreated     OnFileCreated
	}{
		{
			name:          "successful creation",
			title:         "weekly-meeting",
			expectedError: false,
			onCreated: func(file File) error {
				return nil
			},
		},
		{
			name:          "creation error",
			title:         "error-meeting",
			expectedError: true,
			onCreated: func(file File) error {
				return errors.New("mock error")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := CreateMeeting(tc.title, tc.onCreated)

			if tc.expectedError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tc.expectedError {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if file.Title != tc.title {
					t.Errorf("Expected title %s, got %s", tc.title, file.Title)
				}

				if !containsTag(file.Tags, "meeting") {
					t.Errorf("Expected meeting tag, got %v", file.Tags)
				}

				if !file.Done {
					t.Error("Expected meeting to be done")
				}
			}
		})
	}
}

func TestCreateStandup(t *testing.T) {
	testCases := []struct {
		name          string
		teamNames     []string
		teamNamesErr  error
		expectedError bool
		onCreated     OnFileCreated
	}{
		{
			name:          "successful creation",
			teamNames:     []string{"Team Alpha", "Team Beta"},
			teamNamesErr:  nil,
			expectedError: false,
			onCreated: func(file File) error {
				return nil
			},
		},
		{
			name:          "team names error",
			teamNames:     nil,
			teamNamesErr:  errors.New("no team names available"),
			expectedError: true,
			onCreated: func(file File) error {
				return nil
			},
		},
		{
			name:          "creation error",
			teamNames:     []string{"Team Alpha"},
			teamNamesErr:  nil,
			expectedError: true,
			onCreated: func(file File) error {
				return errors.New("mock error")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getTeamNames := func() ([]string, error) {
				return tc.teamNames, tc.teamNamesErr
			}

			file, err := CreateStandup(getTeamNames, tc.onCreated)

			if tc.expectedError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if file.Title != "standup" {
				t.Errorf("Expected title 'standup', got %s", file.Title)
			}

			if !containsTag(file.Tags, "standup") {
				t.Errorf("Expected standup tag, got %v", file.Tags)
			}

			// Verify that content contains team names and weekdays
			for _, name := range tc.teamNames {
				if !strings.Contains(file.Content, name) {
					t.Errorf("Expected content to contain team name '%s'", name)
				}
			}

			weekdays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
			for _, day := range weekdays {
				if !strings.Contains(file.Content, day+" Plan") {
					t.Errorf("Expected content to contain '%s Plan'", day)
				}
			}

			if !strings.Contains(file.Content, "Other Points") {
				t.Error("Expected content to contain 'Other Points'")
			}

			// Verify due date is set to next Friday
			now := time.Now()
			daysUntilFriday := (5 - int(now.Weekday())) % 7
			if daysUntilFriday <= 0 {
				daysUntilFriday += 7
			}
			expectedFriday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilFriday, 0, 0, 0, 0, now.Location())

			if !sameDay(file.DueAt, expectedFriday) {
				t.Errorf("Expected due date to be next Friday (%v), got %v", expectedFriday.Format("2006-01-02"), file.DueAt.Format("2006-01-02"))
			}
		})
	}
}

func TestCreateSevenQuestions(t *testing.T) {
	testCases := []struct {
		name          string
		title         string
		expectedError bool
		onCreated     OnFileCreated
	}{
		{
			name:          "successful creation",
			title:         "project-plan",
			expectedError: false,
			onCreated: func(file File) error {
				return nil
			},
		},
		{
			name:          "creation error",
			title:         "error-plan",
			expectedError: true,
			onCreated: func(file File) error {
				return errors.New("mock error")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := CreateSevenQuestions(tc.title, tc.onCreated)

			if tc.expectedError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tc.expectedError {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if file.Title != tc.title {
					t.Errorf("Expected title %s, got %s", tc.title, file.Title)
				}

				if !containsTag(file.Tags, "plan") {
					t.Errorf("Expected plan tag, got %v", file.Tags)
				}

				// Check for all seven questions in content
				expectedQuestions := []string{
					"What is the situation and how does it affect me?",
					"What have I been told to and why?",
					"What effects do I need to achieve and what direction must I give?",
					"Where can I best accomplish each action or effect?",
					"What resources do I need to accomplish each action or effect?",
					"When and where do these actions take place in relation to each other?",
					"What control measures do I need to impose?",
				}

				for _, question := range expectedQuestions {
					if !strings.Contains(file.Content, question) {
						t.Errorf("Expected content to contain question: '%s'", question)
					}
				}
			}
		})
	}
}

func TestCreateFile(t *testing.T) {
	testCases := []struct {
		name          string
		title         string
		tags          []string
		content       string
		dueAt         time.Time
		done          bool
		expectedError bool
		onCreated     OnFileCreated
	}{
		{
			name:          "basic file creation",
			title:         "test-file",
			tags:          []string{"test", "example"},
			content:       "Test content",
			dueAt:         time.Now().AddDate(0, 0, 7), // one week from now
			done:          false,
			expectedError: false,
			onCreated: func(file File) error {
				return nil
			},
		},
		{
			name:          "empty content uses default",
			title:         "empty-content",
			tags:          []string{"test"},
			content:       "",
			dueAt:         time.Now(),
			done:          false,
			expectedError: false,
			onCreated: func(file File) error {
				return nil
			},
		},
		{
			name:          "creation error",
			title:         "error-file",
			tags:          []string{"test"},
			content:       "Test content",
			dueAt:         time.Now(),
			done:          false,
			expectedError: true,
			onCreated: func(file File) error {
				return errors.New("mock error")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := createFile(tc.title, tc.tags, tc.content, tc.dueAt, tc.done, tc.onCreated)

			if tc.expectedError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check file properties
			if file.Title != tc.title {
				t.Errorf("Expected title %s, got %s", tc.title, file.Title)
			}

			if !equalStringSlices(file.Tags, tc.tags) {
				t.Errorf("Expected tags %v, got %v", tc.tags, file.Tags)
			}

			if file.Done != tc.done {
				t.Errorf("Expected done status %v, got %v", tc.done, file.Done)
			}

			// Check filename format
			dateStr := time.Now().Format("2006-01-02")
			expectedName := tc.title + "-" + dateStr + ".md"
			if file.Name != expectedName {
				t.Errorf("Expected filename %s, got %s", expectedName, file.Name)
			}

			// Check content
			expectedContent := tc.content
			if tc.content == "" {
				expectedContent = "# " + tc.title
			}
			if file.Content != expectedContent {
				t.Errorf("Expected content %s, got %s", expectedContent, file.Content)
			}

			// Check priority is set to default P2
			if file.Priority != P2 {
				t.Errorf("Expected priority %v, got %v", P2, file.Priority)
			}
		})
	}
}

func TestCreateDateRangeQueryNote(t *testing.T) {
	// Setup test data
	startDate := "2023-01-01"
	endDate := "2023-01-31"
	now := time.Now()

	// Create sample completed notes for the date range
	files := []File{
		{
			Name:      "note1.md",
			Title:     "First Task",
			Tags:      []string{"work"},
			CreatedAt: now,
			DueAt:     now,
			Done:      true,
			Content:   "---\ntitle: First Task\ndate-created: " + now.Format("2006-01-02") + "\n---\n\n# First Task\n\nThis is task 1",
		},
		{
			Name:      "note2.md",
			Title:     "Second Task",
			Tags:      []string{"personal"},
			CreatedAt: now,
			DueAt:     now,
			Done:      true,
			Content:   "---\ntitle: Second Task\ndate-created: " + now.Format("2006-01-02") + "\n---\n\n# Second Task\n\nThis is task 2",
		},
	}

	testCases := []struct {
		name          string
		files         []File
		expectedError bool
		onCreated     OnFileCreated
		validate      func(t *testing.T, file File)
	}{
		{
			name:          "successful creation",
			files:         files,
			expectedError: false,
			onCreated: func(file File) error {
				return nil
			},
			validate: func(t *testing.T, file File) {
				expectedTitle := fmt.Sprintf("Date Range Query %s - %s", startDate, endDate)
				if file.Title != expectedTitle {
					t.Errorf("Expected title %s, got %s", expectedTitle, file.Title)
				}

				if !containsTag(file.Tags, "date-range-query") {
					t.Errorf("Expected date-range-query tag, got %v", file.Tags)
				}

				// Check that content includes information from both notes
				if !strings.Contains(file.Content, "First Task") {
					t.Error("Content should include the title of the first task")
				}

				if !strings.Contains(file.Content, "Second Task") {
					t.Error("Content should include the title of the second task")
				}

				if !strings.Contains(file.Content, "This is task 1") {
					t.Error("Content should include the content of the first task")
				}

				if !strings.Contains(file.Content, "This is task 2") {
					t.Error("Content should include the content of the second task")
				}
			},
		},
		{
			name:          "creation error",
			files:         files,
			expectedError: true,
			onCreated: func(file File) error {
				return errors.New("mock error")
			},
			validate: func(t *testing.T, file File) {
				// No validation needed in error case
			},
		},
		{
			name:          "empty files",
			files:         []File{},
			expectedError: false,
			onCreated: func(file File) error {
				return nil
			},
			validate: func(t *testing.T, file File) {
				// Should still create a note, just without any task content
				expectedTitle := fmt.Sprintf("Date Range Query %s - %s", startDate, endDate)
				if file.Title != expectedTitle {
					t.Errorf("Expected title %s, got %s", expectedTitle, file.Title)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := CreateDateRangeQueryNote(startDate, endDate, tc.files, tc.onCreated)

			if tc.expectedError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tc.expectedError {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				tc.validate(t, file)
			}
		})
	}
}

func TestExtractContentWithoutFrontmatter(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "content with frontmatter",
			content: `---
title: Test Note
date-created: 2023-01-01
tags: [test]
---

# Test Note

This is the content.`,
			expected: "\n# Test Note\n\nThis is the content.",
		},
		{
			name:     "content without frontmatter",
			content:  "# Test Note\n\nThis is the content.",
			expected: "# Test Note\n\nThis is the content.",
		},
		{
			name: "multiple frontmatter delimiters - only first section removed",
			content: `---
title: Test Note
---
date-created: 2023-01-01
---

# Test Note

This is the content.`,
			expected: "date-created: 2023-01-01\n---\n\n# Test Note\n\nThis is the content.",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
		{
			name: "incomplete frontmatter",
			content: `---
title: Test Note

# Some content without closing frontmatter`,
			expected: `---
title: Test Note

# Some content without closing frontmatter`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractContentWithoutFrontmatter(tc.content)

			if result != tc.expected {
				t.Errorf("Expected:\n%q\n\nGot:\n%q", tc.expected, result)
			}
		})
	}
}

// Helper functions for the tests
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sameDay(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
