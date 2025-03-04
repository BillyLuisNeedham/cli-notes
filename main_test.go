package main

import (
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"cli-notes/scripts/presentation"
	"errors"
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