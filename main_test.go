package main

import (
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"cli-notes/scripts/presentation"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestHandleCommand(t *testing.T) {
	// Setup
	fileStore := data.NewSearchedFilesStore()
	closeCalled := false
	onClose := func() { closeCalled = true }

	testCases := []struct {
		name     string
		command  presentation.CompletedCommand
		wantErr  bool
		validate func(t *testing.T)
	}{
		{
			name: "gt with no queries",
			command: presentation.CompletedCommand{
				Name:    "gt",
				Queries: []string{},
			},
			validate: func(t *testing.T) {
				files := fileStore.GetFilesSearched()
				if len(files) == 0 {
					t.Error("Expected files to be stored")
				}
			},
		},
		{
			name: "exit command",
			command: presentation.CompletedCommand{
				Name: "exit",
			},
			validate: func(t *testing.T) {
				if !closeCalled {
					t.Error("Expected close function to be called")
				}
			},
		},
		{
			name: "d command without file",
			command: presentation.CompletedCommand{
				Name:    "d",
				Queries: []string{"5"},
			},
			validate: func(t *testing.T) {
				// No validation needed - just checking it doesn't panic
			},
		},
		{
			name: "d command with file but no days",
			command: presentation.CompletedCommand{
				Name:         "d",
				SelectedFile: scripts.File{Name: "test.md"},
			},
			validate: func(t *testing.T) {
				// No validation needed - just checking it doesn't panic
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handleCommand(tc.command, onClose, fileStore)
			tc.validate(t)
		})
	}
}

func TestHandleCreateFile(t *testing.T) {
	testCases := []struct {
		name      string
		fileType  string
		queries   []string
		createFn  func(string, scripts.OnFileCreated) (scripts.File, error)
		wantError bool
	}{
		{
			name:     "empty queries",
			fileType: "todo",
			queries:  []string{},
			createFn: func(string, scripts.OnFileCreated) (scripts.File, error) {
				return scripts.File{}, nil
			},
			wantError: true,
		},
		{
			name:     "successful creation",
			fileType: "todo",
			queries:  []string{"test", "todo"},
			createFn: func(title string, onCreated scripts.OnFileCreated) (scripts.File, error) {
				if title != "test-todo" {
					t.Errorf("Expected title 'test-todo', got %s", title)
				}
				return scripts.File{Name: "test.md"}, nil
			},
			wantError: false,
		},
		{
			name:     "creation error",
			fileType: "todo",
			queries:  []string{"test"},
			createFn: func(string, scripts.OnFileCreated) (scripts.File, error) {
				return scripts.File{}, errors.New("creation failed")
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handleCreateFile(tc.fileType, tc.queries, tc.createFn)
			// Test passes if it doesn't panic
		})
	}
}

func TestSearchRecentFilesPrintIfNotFound(t *testing.T) {
	testCases := []struct {
		name       string
		searchFunc func() *scripts.File
		wantEmpty  bool
	}{
		{
			name: "found file",
			searchFunc: func() *scripts.File {
				return &scripts.File{
					Name:      "test.md",
					CreatedAt: time.Now(),
				}
			},
			wantEmpty: false,
		},
		{
			name: "no file found",
			searchFunc: func() *scripts.File {
				return nil
			},
			wantEmpty: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := searchRecentFilesPrintIfNotFound(tc.searchFunc)

			if tc.wantEmpty && result.Name != "" {
				t.Errorf("Expected empty file, got name: %s", result.Name)
			}
			if !tc.wantEmpty && result.Name == "" {
				t.Error("Expected non-empty file, got empty")
			}
		})
	}
}

func TestIsValidDate(t *testing.T) {
	testCases := []struct {
		name     string
		date     string
		expected bool
	}{
		{
			name:     "valid date",
			date:     "2023-01-01",
			expected: true,
		},
		{
			name:     "valid date with different month",
			date:     "2023-02-15",
			expected: true,
		},
		{
			name:     "invalid format - missing leading zeros",
			date:     "2023-1-1",
			expected: false,
		},
		{
			name:     "invalid format - American style",
			date:     "01/01/2023",
			expected: false,
		},
		{
			name:     "invalid format - text month",
			date:     "2023-Jan-01",
			expected: false,
		},
		{
			name:     "invalid date - out of range",
			date:     "2023-13-01",
			expected: false,
		},
		{
			name:     "empty string",
			date:     "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidDate(tc.date)
			if result != tc.expected {
				t.Errorf("Expected isValidDate(%s) to be %v, got %v", tc.date, tc.expected, result)
			}
		})
	}
}

func TestHandleCommandGD(t *testing.T) {
	// Setup
	fileStore := data.NewSearchedFilesStore()
	onClose := func() {}

	testCases := []struct {
		name    string
		command presentation.CompletedCommand
		wantErr bool
	}{
		{
			name: "gd with no queries",
			command: presentation.CompletedCommand{
				Name:    "gd",
				Queries: []string{},
			},
			wantErr: true,
		},
		{
			name: "gd with one query",
			command: presentation.CompletedCommand{
				Name:    "gd",
				Queries: []string{"2023-01-01"},
			},
			wantErr: true,
		},
		{
			name: "gd with invalid date format",
			command: presentation.CompletedCommand{
				Name:    "gd",
				Queries: []string{"01/01/2023", "01/31/2023"},
			},
			wantErr: true,
		},
		{
			name: "gd with valid dates",
			command: presentation.CompletedCommand{
				Name:    "gd",
				Queries: []string{"2023-01-01", "2023-01-31"},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture output to check for error messages
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call the function
			handleCommand(tc.command, onClose, fileStore)

			// Restore stdout
			w.Close()
			os.Stdout = originalStdout

			// Read captured output
			var buf strings.Builder
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatalf("Failed to read captured output: %v", err)
			}
			output := buf.String()

			// Check if error message was printed
			if tc.wantErr {
				if !strings.Contains(output, "Please provide") &&
					!strings.Contains(output, "Invalid date format") {
					t.Errorf("Expected error message but got: %s", output)
				}
			} else {
				// Note: In the actual test, the file creation will likely fail since we're in a test environment
				// without proper file system setup, but we're just testing that it attempts to proceed.
				if strings.Contains(output, "Please provide") ||
					strings.Contains(output, "Invalid date format") {
					t.Errorf("Didn't expect error message but got: %s", output)
				}
			}
		})
	}
}
