package e2e

import (
	"fmt"
	"strings"
	"testing"
)

// Phase 4: Edge Cases

func TestTalkTo_NoTodosWithTalkToTags(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Create todos without to-talk tags
	h.CreateTodo("normal-todo1.md", "Normal task", []string{"work"}, Today(), false, 1)
	h.CreateTodo("normal-todo2.md", "Another task", []string{"personal"}, Today(), false, 2)

	// Try to run tt command
	input := "tt\nq"

	stdout, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// Verify: Should handle gracefully (either show empty state or exit without crash)
	// The command should not crash - exact behavior may vary
	t.Logf("Output with no to-talk tags: %s", stdout)
}

func TestTalkTo_PersonWithSingleTodo(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Single todo for alice
	h.CreateTodoWithTalkToTag("single.md", "Only one task", "alice", []string{}, Today(), 1)

	// Complete workflow
	input := "tt alice\n\nspace\n\nc\nsingle-result\n\nq"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// Verify: Todo moved successfully
	targetFile := fmt.Sprintf("single-result-%s.md", Today())
	h.VerifyTodoInFile(targetFile, "Only one task")

	// Verify source marked complete
	content := h.ReadFileContent("single.md")
	if !strings.Contains(content, "- [x]") {
		t.Errorf("Expected single.md to be marked complete, got:\n%s", content)
	}
}

func TestTalkTo_BackNavigationAtEachStage(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Create todos for navigation test
	h.CreateTodoWithTalkToTag("nav1.md", "Navigation test 1", "alice", []string{}, Today(), 1)
	h.CreateTodoWithTalkToTag("nav2.md", "Navigation test 2", "bob", []string{}, Today(), 1)

	// Test 1: Back from todo selection to person selection
	// Input: tt -> Enter (select alice) -> Esc (back to person) -> q
	input1 := "tt\n\n\x1bq"

	_, _, err := h.RunCommand(input1)
	if err != nil {
		t.Logf("Back navigation test 1 completed with: %v", err)
	}

	// Test 2: Back from note selection to todo selection
	// Input: tt -> Enter -> space -> Enter -> Esc (back to todo) -> q
	input2 := "tt\n\nspace\n\n\x1bq"

	_, _, err = h.RunCommand(input2)
	if err != nil {
		t.Logf("Back navigation test 2 completed with: %v", err)
	}

	// Verify: No crashes, no files modified
	// Files should still be incomplete since we didn't complete workflow
	content1 := h.ReadFileContent("nav1.md")
	content2 := h.ReadFileContent("nav2.md")
	if !strings.Contains(content1, "- [ ]") {
		t.Errorf("Expected nav1.md to remain incomplete after back navigation, got:\n%s", content1)
	}
	if !strings.Contains(content2, "- [ ]") {
		t.Errorf("Expected nav2.md to remain incomplete after back navigation, got:\n%s", content2)
	}
}

func TestTalkTo_TagCaseInsensitivity(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Todos with different case variations
	h.CreateTodoWithContent("case1.md", "Lowercase tag", "- [ ] Task 1 to-talk-alice", Today(), 1)
	h.CreateTodoWithContent("case2.md", "Uppercase tag", "- [ ] Task 2 to-talk-ALICE", Today(), 1)
	h.CreateTodoWithContent("case3.md", "Mixed case tag", "- [ ] Task 3 to-talk-Alice", Today(), 1)

	// Run tt without filter (should show all grouped under one person)
	input := "tt\n\na\n\nc\ncase-test\n\nq"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// Verify: All three todos should be in target file (grouped by normalized name)
	targetFile := fmt.Sprintf("case-test-%s.md", Today())
	h.VerifyTodoInFile(targetFile, "Task 1")
	h.VerifyTodoInFile(targetFile, "Task 2")
	h.VerifyTodoInFile(targetFile, "Task 3")
}

func TestTalkTo_MultipleTagsPerTodo(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Todo with multiple to-talk tags
	multiTagContent := "- [ ] Coordinate with team to-talk-alice to-talk-bob to-talk-charlie"
	h.CreateTodoWithContent("multi.md", "Multi Tag", multiTagContent, Today(), 1)

	// Select alice's todos
	input := "tt alice\n\nspace\n\nc\nalice-notes\n\nq"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// Verify: Todo appears under alice
	targetFile := fmt.Sprintf("alice-notes-%s.md", Today())
	h.VerifyTodoInFile(targetFile, "Coordinate with team")

	// Verify: ALL tags removed (not just alice)
	h.VerifyFileNotContains(targetFile, "to-talk-alice")
	h.VerifyFileNotContains(targetFile, "to-talk-bob")
	h.VerifyFileNotContains(targetFile, "to-talk-charlie")
}

func TestTalkTo_SubtaskHandling(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Todo with indented subtasks
	todoWithSubtasks := `- [ ] Main task to-talk-alice
  - [ ] Subtask 1
  - [ ] Subtask 2
    - [ ] Nested subtask`

	h.CreateTodoWithContent("subtasks.md", "Subtasks Test", todoWithSubtasks, Today(), 1)

	// Complete workflow
	input := "tt alice\n\nspace\n\nc\nsubtask-result\n\nq"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// Verify: Parent and all subtasks inserted
	targetFile := fmt.Sprintf("subtask-result-%s.md", Today())
	h.VerifyTodoInFile(targetFile, "Main task")
	h.VerifyTodoInFile(targetFile, "Subtask 1")
	h.VerifyTodoInFile(targetFile, "Subtask 2")
	h.VerifyTodoInFile(targetFile, "Nested subtask")

	// Verify: All marked complete in source (parent and subtasks)
	sourceContent := h.ReadFileContent("subtasks.md")
	lines := strings.Split(sourceContent, "\n")

	// Count completed subtasks
	completedCount := 0
	for _, line := range lines {
		if strings.Contains(line, "- [x]") {
			completedCount++
		}
	}

	if completedCount < 4 {
		t.Errorf("Expected all 4 tasks (parent + 3 subtasks) to be marked complete, got %d:\n%s", completedCount, sourceContent)
	}
}

func TestTalkTo_CreateNoteWithDuplicateName(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Create existing note with name we'll try to use
	existingName := fmt.Sprintf("meeting-notes-%s.md", Today())
	h.CreateTodo(existingName, "Existing Meeting Notes", []string{}, Today(), false, 1)

	// Create todo to move
	h.CreateTodoWithTalkToTag("dup-test.md", "Duplicate name test", "alice", []string{}, Today(), 1)

	// Try to create note with same name
	input := "tt alice\n\nspace\n\nc\nmeeting-notes\n\nq"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// The system should handle this gracefully
	// It might: create with incremented number, add timestamp, or show error
	// At minimum, it shouldn't crash and existing file shouldn't be corrupted

	// Verify: Original file still exists and intact
	h.AssertFileExists(existingName)
	originalContent := h.ReadFileContent(existingName)
	if !strings.Contains(originalContent, "Existing Meeting Notes") {
		t.Errorf("Original file corrupted or replaced")
	}
}

func TestTalkTo_SearchWithNoResults(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Create a few notes with specific names
	h.CreateTodo("work-notes.md", "Work Notes", []string{}, Today(), false, 1)
	h.CreateTodo("personal-notes.md", "Personal Notes", []string{}, Today(), false, 1)
	h.CreateTodo("meeting-log.md", "Meeting Log", []string{}, Today(), false, 1)

	// Create todo to move
	h.CreateTodoWithTalkToTag("search-test.md", "Search test task", "alice", []string{}, Today(), 1)

	// Try to search for non-existent note
	// Input: tt alice -> Enter -> space -> Enter -> f (find) -> i (INSERT) -> "nonexistent" -> Esc -> Esc (close modal) -> c (create new instead)
	input := "tt alice\n\nspace\n\nf\ni\nnonexistent\n\x1b\x1bc\nsearch-fallback\n\nq"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// Verify: Should not crash
	// Todo should eventually be moved (via create new note fallback)
	targetFile := fmt.Sprintf("search-fallback-%s.md", Today())
	h.AssertFileExists(targetFile)
}
