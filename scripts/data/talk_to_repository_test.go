package data

import (
	"cli-notes/scripts"
	"os"
	"strings"
	"testing"
	"time"
)

// Helper functions
func setupTestNotes(t *testing.T) {
	os.RemoveAll("notes")
	err := os.Mkdir("notes", 0755)
	if err != nil && !os.IsExist(err) {
		t.Fatalf("Failed to create notes directory: %v", err)
	}
}

func cleanupTestNotes(t *testing.T) {
	os.RemoveAll("notes")
}

func createTestFileWithContent(t *testing.T, name, content string) scripts.File {
	err := os.WriteFile("notes/"+name, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	return scripts.File{
		Name:      name,
		Title:     "Test",
		CreatedAt: time.Now(),
		Priority:  scripts.P2,
		Done:      false,
	}
}

func TestMarkTodoLineComplete_Success(t *testing.T) {
	// Setup temporary test environment
	setupTestNotes(t)
	defer cleanupTestNotes(t)

	// Create test file with incomplete todos
	file := createTestFileWithContent(t, "test-todo.md", "- [ ] Test todo item\n- [ ] Another item")

	// Mark first todo complete
	err := MarkTodoLineComplete(file, 1)
	if err != nil {
		t.Fatalf("MarkTodoLineComplete failed: %v", err)
	}

	// Verify the modification
	content, err := os.ReadFile("notes/test-todo.md")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	lines := strings.Split(string(content), "\n")

	if !strings.Contains(lines[0], "- [x] Test todo item") {
		t.Errorf("Expected first line marked complete, got: %q", lines[0])
	}
	if !strings.Contains(lines[1], "- [ ] Another item") {
		t.Errorf("Expected second line still incomplete, got: %q", lines[1])
	}
}

func TestMarkTodoLineComplete_PatternNotFound(t *testing.T) {
	setupTestNotes(t)
	defer cleanupTestNotes(t)

	// Create file with already-complete todo
	file := createTestFileWithContent(t, "test-complete.md", "- [x] Already complete")

	// Try to mark it complete again (should fail with clear error)
	err := MarkTodoLineComplete(file, 1)

	if err == nil {
		t.Fatal("Expected error when pattern not found, got nil")
	}

	if !strings.Contains(err.Error(), "pattern \"- [ ]\" not found") {
		t.Errorf("Expected 'pattern not found' error, got: %v", err)
	}

	// Verify error includes line content for debugging
	if !strings.Contains(err.Error(), "- [x] Already complete") {
		t.Errorf("Expected error to include line content, got: %v", err)
	}
}

func TestMarkTodoLineComplete_WhitespaceVariations(t *testing.T) {
	setupTestNotes(t)
	defer cleanupTestNotes(t)

	testCases := []struct {
		name        string
		content     string
		shouldError bool
		description string
	}{
		{
			name:        "standard format",
			content:     "- [ ] Standard todo",
			shouldError: false,
			description: "Standard format should work",
		},
		{
			name:        "double space after dash",
			content:     "-  [ ] Double space",
			shouldError: true,
			description: "Should error - pattern doesn't match exactly",
		},
		{
			name:        "no space after dash",
			content:     "-[ ] No space",
			shouldError: true,
			description: "Should error - missing space",
		},
		{
			name:        "tab after dash",
			content:     "-\t[ ] Tab character",
			shouldError: true,
			description: "Should error - tab instead of space",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.RemoveAll("notes")
			os.Mkdir("notes", 0755)

			file := createTestFileWithContent(t, "test.md", tc.content)
			err := MarkTodoLineComplete(file, 1)

			if tc.shouldError && err == nil {
				t.Errorf("%s: expected error, got nil", tc.description)
			}
			if !tc.shouldError && err != nil {
				t.Errorf("%s: expected success, got error: %v", tc.description, err)
			}

			// Clean up for next iteration
			cleanupTestNotes(t)
			setupTestNotes(t)
		})
	}

	cleanupTestNotes(t)
}

func TestMarkTodoLineComplete_InvalidLineNumber(t *testing.T) {
	setupTestNotes(t)
	defer cleanupTestNotes(t)

	file := createTestFileWithContent(t, "test-lines.md", "- [ ] Only one line")

	// Test line 0 (invalid, 1-indexed)
	err := MarkTodoLineComplete(file, 0)
	if err == nil || !strings.Contains(err.Error(), "invalid line number") {
		t.Error("Expected 'invalid line number' error for line 0")
	}

	// Test line beyond file length
	err = MarkTodoLineComplete(file, 10)
	if err == nil || !strings.Contains(err.Error(), "invalid line number") {
		t.Error("Expected 'invalid line number' error for line 10")
	}
}

func TestMarkTodoLineComplete_MultipleLines(t *testing.T) {
	setupTestNotes(t)
	defer cleanupTestNotes(t)

	// Create file with multiple incomplete todos
	content := "- [ ] First todo\n- [ ] Second todo\n- [ ] Third todo"
	file := createTestFileWithContent(t, "test-multi.md", content)

	// Mark second todo complete
	err := MarkTodoLineComplete(file, 2)
	if err != nil {
		t.Fatalf("Failed to mark second line complete: %v", err)
	}

	// Verify only the second line was modified
	result, _ := os.ReadFile("notes/test-multi.md")
	lines := strings.Split(string(result), "\n")

	if !strings.Contains(lines[0], "- [ ] First todo") {
		t.Errorf("Expected first line unchanged, got: %q", lines[0])
	}
	if !strings.Contains(lines[1], "- [x] Second todo") {
		t.Errorf("Expected second line marked complete, got: %q", lines[1])
	}
	if !strings.Contains(lines[2], "- [ ] Third todo") {
		t.Errorf("Expected third line unchanged, got: %q", lines[2])
	}
}

func TestMarkTodoLineIncomplete_Success(t *testing.T) {
	setupTestNotes(t)
	defer cleanupTestNotes(t)

	// Create file with complete todo
	file := createTestFileWithContent(t, "test-undo.md", "- [x] Completed todo")

	// Mark it incomplete (undo operation)
	err := MarkTodoLineIncomplete(file, 1)
	if err != nil {
		t.Fatalf("MarkTodoLineIncomplete failed: %v", err)
	}

	// Verify it's now incomplete
	content, _ := os.ReadFile("notes/test-undo.md")
	if !strings.Contains(string(content), "- [ ] Completed todo") {
		t.Errorf("Expected todo marked incomplete, got: %q", string(content))
	}
}

func TestMarkTodoLineIncomplete_AlreadyIncomplete(t *testing.T) {
	setupTestNotes(t)
	defer cleanupTestNotes(t)

	// Create file with already-incomplete todo
	file := createTestFileWithContent(t, "test-already-incomplete.md", "- [ ] Already incomplete")

	// Try to mark it incomplete again (should error)
	err := MarkTodoLineIncomplete(file, 1)

	if err == nil {
		t.Fatal("Expected error when trying to mark already-incomplete todo, got nil")
	}

	if !strings.Contains(err.Error(), "pattern \"- [x]\" not found") {
		t.Errorf("Expected 'pattern not found' error, got: %v", err)
	}
}
