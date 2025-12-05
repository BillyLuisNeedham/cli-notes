package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Phase 1: Happy Path Workflows

func TestTalkTo_CompleteWorkflowWithPersonFilter(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Create 3 todos with to-talk-alice tags
	h.CreateTodoWithTalkToTag("todo1.md", "Review PR", "alice", []string{}, Today(), 1)
	h.CreateTodoWithTalkToTag("todo2.md", "Discuss API design", "alice", []string{}, Today(), 2)
	h.CreateTodoWithTalkToTag("todo3.md", "Plan sprint", "alice", []string{}, Today(), 1)

	// Input: tt alice -> Enter (select person) -> space j space (select 2 todos) -> Enter -> n (create new note) -> meeting-alice\n -> Enter (confirm) -> q (quit)
	input := "tt alice\n\n j \n\nnmeeting-alice\n\nq"

	stdout, stderr, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	// Verify: New file created with 2 todos
	expectedFilename := fmt.Sprintf("meeting-alice-%s.md", Today())
	h.AssertFileExists(expectedFilename)

	// Verify todos inserted into new file (without tags)
	h.VerifyTodoInFile(expectedFilename, "Review PR")
	h.VerifyTodoInFile(expectedFilename, "Discuss API design")

	// Verify tags removed from inserted todos
	h.VerifyFileNotContains(expectedFilename, "to-talk-alice")

	// Verify source todos marked complete
	content1 := h.ReadFileContent("todo1.md")
	content2 := h.ReadFileContent("todo2.md")
	if !strings.Contains(content1, "- [x]") {
		t.Errorf("Expected todo1.md to be marked complete, got:\n%s", content1)
	}
	if !strings.Contains(content2, "- [x]") {
		t.Errorf("Expected todo2.md to be marked complete, got:\n%s", content2)
	}
}

func TestTalkTo_CompleteWorkflowWithoutFilter(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Create todos with different people tags and an existing target note
	h.CreateTodoWithTalkToTag("alice-todo.md", "Fix authentication bug", "alice", []string{}, Today(), 1)
	h.CreateTodoWithTalkToTag("bob-todo.md", "Update documentation", "bob", []string{}, Today(), 2)

	// Create existing target note
	existingNote := "meeting-notes.md"
	h.CreateTodo(existingNote, "Meeting Notes", []string{}, Today(), false, 1)

	// Input: tt -> j (navigate to bob) -> Enter (select bob) -> Enter (proceed with auto-selected) -> f (find) -> "meeting-notes" (search) -> Esc -> Enter (select) -> Enter (confirm) -> q
	input := "tt\nj\n\nfmeeting-notes\x1b\n\n\nq"

	stdout, stderr, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	// Verify: Todo inserted into existing note
	h.VerifyTodoInFile(existingNote, "Update documentation")

	// Verify tag removed
	h.VerifyFileNotContains(existingNote, "to-talk-bob")

	// Verify source marked complete
	content := h.ReadFileContent("bob-todo.md")
	if !strings.Contains(content, "- [x]") {
		t.Errorf("Expected bob-todo.md to be marked complete, got:\n%s", content)
	}
}

func TestTalkTo_CreateNewNoteWorkflow(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Create multiple todos for one person
	h.CreateTodoWithTalkToTag("standup1.md", "Complete feature X", "team", []string{}, Today(), 1)
	h.CreateTodoWithTalkToTag("standup2.md", "Review code", "team", []string{}, Today(), 2)
	h.CreateTodoWithTalkToTag("standup3.md", "Update tests", "team", []string{}, Today(), 1)

	// Input: tt -> Enter (select first person) -> a (select all) -> Enter -> c (create new) -> "standup-notes" -> Enter -> q
	input := "tt\n\na\n\nnstandup-notes\n\nq"

	stdout, stderr, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	// Verify: New note created
	expectedFilename := fmt.Sprintf("standup-notes-%s.md", Today())
	h.AssertFileExists(expectedFilename)

	// Verify all todos moved
	h.VerifyTodoInFile(expectedFilename, "Complete feature X")
	h.VerifyTodoInFile(expectedFilename, "Review code")
	h.VerifyTodoInFile(expectedFilename, "Update tests")

	// Verify all source files marked complete
	for _, file := range []string{"standup1.md", "standup2.md", "standup3.md"} {
		content := h.ReadFileContent(file)
		if !strings.Contains(content, "- [x]") {
			t.Errorf("Expected %s to be marked complete, got:\n%s", file, content)
		}
	}
}

// Phase 2: File Modification Verification

func TestTalkTo_SourceTodosMarkedCompleteWithCheckbox(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Create 3 incomplete todos
	h.CreateTodoWithTalkToTag("task1.md", "First task", "alice", []string{}, Today(), 1)
	h.CreateTodoWithTalkToTag("task2.md", "Second task", "alice", []string{}, Today(), 1)
	h.CreateTodoWithTalkToTag("task3.md", "Third task", "alice", []string{}, Today(), 1)

	// Complete workflow: select all and create note
	input := "tt alice\n\na\n\nntask-results\n\nq"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// Verify: All source lines changed from [ ] to [x]
	for _, file := range []string{"task1.md", "task2.md", "task3.md"} {
		content := h.ReadFileContent(file)
		if !strings.Contains(content, "- [x]") {
			t.Errorf("Expected %s to contain '- [x]' (marked complete), got:\n%s", file, content)
		}
		if strings.Contains(content, "- [ ]") && strings.Contains(content, "to-talk-alice") {
			t.Errorf("Expected %s to have todo marked complete (no unchecked boxes with tag), got:\n%s", file, content)
		}
	}
}

func TestTalkTo_TagsRemovedOnInsert(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Todo with multiple tags in the title
	todoContent := "- [ ] Fix bug to-talk-alice to-talk-bob some additional notes"
	h.CreateTodoWithContent("multi-tag.md", "Multi Tag Task", todoContent, Today(), 1)

	// Complete workflow
	input := "tt alice\n\n \n\nnresults\n\nq"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// Verify: Tags removed in inserted todo
	expectedFilename := fmt.Sprintf("results-%s.md", Today())
	h.VerifyTodoInFile(expectedFilename, "Fix bug")
	h.VerifyTodoInFile(expectedFilename, "some additional notes")

	// Verify: No to-talk-X tags in target
	h.VerifyFileNotContains(expectedFilename, "to-talk-alice")
	h.VerifyFileNotContains(expectedFilename, "to-talk-bob")
}

func TestTalkTo_TodosInsertedAtCorrectLocation(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Create target note with existing content
	existingContent := `---
title: Existing Note
date-created: 2025-11-28
done: false
priority: 1
---

## Existing Section

Some existing content here.
- Existing item 1
- Existing item 2`

	h.CreateTestFile("existing.md", existingContent)

	// Create todo with tag
	h.CreateTodoWithTalkToTag("new-task.md", "New task to insert", "alice", []string{}, Today(), 1)

	// Complete workflow: find existing note
	// tt alice -> Enter (proceed with auto-selected) -> f (find) -> "existing" (search) -> Esc -> Enter (select) -> Enter (confirm) -> Enter (execute) -> q
	input := "tt alice\n\nfexisting\x1b\n\n\nq"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// Verify: New todo inserted after frontmatter, before existing content
	content := h.ReadFileContent("existing.md")
	lines := strings.Split(content, "\n")

	// Find the inserted todo
	todoLineIndex := -1
	existingSectionIndex := -1
	for i, line := range lines {
		if strings.Contains(line, "New task to insert") {
			todoLineIndex = i
		}
		if strings.Contains(line, "## Existing Section") {
			existingSectionIndex = i
		}
	}

	if todoLineIndex == -1 {
		t.Errorf("Expected to find inserted todo in existing.md")
	}
	if existingSectionIndex == -1 {
		t.Errorf("Expected to find existing section in existing.md")
	}

	// Todo should be before existing section
	if todoLineIndex >= existingSectionIndex {
		t.Errorf("Expected inserted todo (line %d) to be before existing section (line %d)", todoLineIndex, existingSectionIndex)
	}
}

func TestTalkTo_PreserveTodoFormatting(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Todo with special formatting (links, code, bold)
	specialContent := "- [ ] Fix [API bug](https://example.com/issue/123) using `fetch()` **URGENT** to-talk-alice"
	h.CreateTodoWithContent("special.md", "Special Formatting", specialContent, Today(), 1)

	// Complete workflow
	input := "tt alice\n\n \n\nnformatted\n\nq"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// Verify: Formatting preserved (except tags removed)
	expectedFilename := fmt.Sprintf("formatted-%s.md", Today())
	h.VerifyFileContains(expectedFilename, "[API bug](https://example.com/issue/123)")
	h.VerifyFileContains(expectedFilename, "`fetch()`")
	h.VerifyFileContains(expectedFilename, "**URGENT**")
	h.VerifyFileNotContains(expectedFilename, "to-talk-alice")
}

func TestTalkTo_MultipleSourceFiles(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: 3 todos from 3 different source files
	h.CreateTodoWithTalkToTag("project-a.md", "Task from project A", "alice", []string{}, Today(), 1)
	h.CreateTodoWithTalkToTag("project-b.md", "Task from project B", "alice", []string{}, Today(), 2)
	h.CreateTodoWithTalkToTag("project-c.md", "Task from project C", "alice", []string{}, Today(), 1)

	// Complete workflow: select all -> create new note
	input := "tt alice\n\na\n\nnmulti-project\n\nq"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// Verify: All 3 source files modified correctly
	for _, file := range []string{"project-a.md", "project-b.md", "project-c.md"} {
		content := h.ReadFileContent(file)
		if !strings.Contains(content, "- [x]") {
			t.Errorf("Expected %s to be marked complete with [x], got:\n%s", file, content)
		}
	}

	// Verify: Target file has all 3 todos
	targetFile := fmt.Sprintf("multi-project-%s.md", Today())
	h.VerifyTodoInFile(targetFile, "Task from project A")
	h.VerifyTodoInFile(targetFile, "Task from project B")
	h.VerifyTodoInFile(targetFile, "Task from project C")

	// Verify: Count is correct
	count := h.CountTodosInFile(targetFile)
	if count != 3 {
		t.Errorf("Expected 3 todos in target file, got %d", count)
	}
}

// Phase 3: Undo Functionality

func TestTalkTo_UndoImmediatelyAfterMove(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Create todo and target note
	todoContent := "- [ ] Important task to-talk-alice"
	h.CreateTodoWithContent("undo-test.md", "Undo Test", todoContent, Today(), 1)

	// Complete workflow and press 'u' to undo on success screen
	input := "tt alice\n\n \n\nnundo-target\n\nu"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// Verify: Todo removed from target note
	targetFile := fmt.Sprintf("undo-target-%s.md", Today())
	targetPath := filepath.Join(h.NotesDir, targetFile)
	if _, err := os.Stat(targetPath); err == nil {
		// If file exists, verify todo was removed
		content := h.ReadFileContent(targetFile)
		if strings.Contains(content, "Important task") {
			t.Errorf("Expected todo to be removed from target after undo, but found it in:\n%s", content)
		}
	}

	// Verify: Source todo marked incomplete again
	sourceContent := h.ReadFileContent("undo-test.md")
	if !strings.Contains(sourceContent, "- [ ]") {
		t.Errorf("Expected source todo to be marked incomplete after undo, got:\n%s", sourceContent)
	}
	if strings.Contains(sourceContent, "- [x]") && strings.Contains(sourceContent, "Important task") {
		t.Errorf("Expected source todo NOT to be marked complete after undo, got:\n%s", sourceContent)
	}
}

func TestTalkTo_UndoAffectsCorrectFiles(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Move 3 todos from 3 different source files
	h.CreateTodoWithTalkToTag("undo-src1.md", "Task from file 1", "alice", []string{}, Today(), 1)
	h.CreateTodoWithTalkToTag("undo-src2.md", "Task from file 2", "alice", []string{}, Today(), 1)
	h.CreateTodoWithTalkToTag("undo-src3.md", "Task from file 3", "alice", []string{}, Today(), 1)

	// Complete workflow with all todos selected, then undo
	input := "tt alice\n\na\n\nnundo-multi\n\nu"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Logf("Command completed with error: %v", err)
	}

	// Verify: All 3 source files restored to incomplete
	for i, file := range []string{"undo-src1.md", "undo-src2.md", "undo-src3.md"} {
		content := h.ReadFileContent(file)
		if !strings.Contains(content, "- [ ]") {
			t.Errorf("Expected %s to be marked incomplete after undo, got:\n%s", file, content)
		}
		expectedTask := fmt.Sprintf("Task from file %d", i+1)
		if !strings.Contains(content, expectedTask) {
			t.Errorf("Expected %s to contain %q, got:\n%s", file, expectedTask, content)
		}
	}
}

func TestTalkTo_MultipleMoves(t *testing.T) {
	h := NewTestHarness(t)

	// Setup: Create todos for first move
	h.CreateTodoWithTalkToTag("move1.md", "First move task", "alice", []string{}, Today(), 1)

	// First move: complete workflow and return: tt alice -> Enter -> space (toggle) -> Enter -> n (create) -> first-move -> Enter -> r (return)
	input1 := "tt alice\n\n \n\nnfirst-move\n\nr"

	_, _, err := h.RunCommand(input1)
	if err != nil {
		t.Logf("First move completed with error: %v", err)
	}

	// Verify first move succeeded
	firstTarget := fmt.Sprintf("first-move-%s.md", Today())
	h.VerifyTodoInFile(firstTarget, "First move task")

	// Setup second move: Create new todo
	h.CreateTodoWithTalkToTag("move2.md", "Second move task", "bob", []string{}, Today(), 1)

	// Second move: complete workflow
	input2 := "tt bob\n\n \n\nnsecond-move\n\nq"

	_, _, err = h.RunCommand(input2)
	if err != nil {
		t.Logf("Second move completed with error: %v", err)
	}

	// Verify second move succeeded
	secondTarget := fmt.Sprintf("second-move-%s.md", Today())
	h.VerifyTodoInFile(secondTarget, "Second move task")

	// Both moves should persist (undo stack should have both)
	h.AssertFileExists(firstTarget)
	h.AssertFileExists(secondTarget)
}
