package data

import (
	"cli-notes/scripts"
	"fmt"
	"os"
	"strings"
)

// ScanAllTalkToTodos scans all incomplete todos for to-talk-X tags
// Returns a map of person name -> list of todos
func ScanAllTalkToTodos() (map[string][]TodoWithMeta, error) {
	// Load all incomplete files
	files, err := QueryFilesByDone(false)
	if err != nil {
		return nil, err
	}

	todosByPerson := make(map[string][]TodoWithMeta)

	// Scan each file
	for _, file := range files {
		// Read file content to parse individual lines
		content, err := os.ReadFile("notes/" + file.Name)
		if err != nil {
			continue // Skip files that can't be read
		}

		lines := strings.Split(string(content), "\n")

		// Scan each line for to-talk tags
		for i, line := range lines {
			// Only process incomplete todo lines
			if !strings.Contains(line, "- [ ] ") {
				continue
			}

			// Check for to-talk tags
			people := scripts.ParseTodoLine(line)
			if len(people) == 0 {
				continue
			}

			// Extract subtasks if any
			subtasks := scripts.ExtractSubtasks(lines, i)

			// Add this todo to each person's list
			for _, person := range people {
				todo := TodoWithMeta{
					File:       file,
					TodoLine:   line,
					LineNumber: i + 1, // Convert to 1-indexed
					SourceFile: file.Name,
					Subtasks:   subtasks,
				}

				todosByPerson[person] = append(todosByPerson[person], todo)
			}
		}
	}

	return todosByPerson, nil
}

// MarkTodoLineComplete marks a specific todo line as complete in a file
// Changes "- [ ]" to "- [x]" at the specified line number
func MarkTodoLineComplete(file scripts.File, lineNumber int) error {
	return modifyTodoLine(file, lineNumber, "- [ ]", "- [x]")
}

// MarkTodoLineIncomplete marks a specific todo line as incomplete in a file
// Changes "- [x]" to "- [ ]" at the specified line number (for undo)
func MarkTodoLineIncomplete(file scripts.File, lineNumber int) error {
	return modifyTodoLine(file, lineNumber, "- [x]", "- [ ]")
}

// modifyTodoLine changes a specific pattern in a line
func modifyTodoLine(file scripts.File, lineNumber int, oldPattern, newPattern string) error {
	// Read file content
	content, err := os.ReadFile("notes/" + file.Name)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", file.Name, err)
	}

	lines := strings.Split(string(content), "\n")

	// Validate line number (1-indexed)
	if lineNumber < 1 || lineNumber > len(lines) {
		return fmt.Errorf("invalid line number %d (file has %d lines)", lineNumber, len(lines))
	}

	// Convert to 0-indexed
	lineIndex := lineNumber - 1

	// Get the line content
	oldLine := lines[lineIndex]

	// VERIFY 1: Check that the old pattern exists BEFORE attempting replacement
	if !strings.Contains(oldLine, oldPattern) {
		return fmt.Errorf(
			"pattern %q not found in line %d of %s\nLine content: %q",
			oldPattern, lineNumber, file.Name, oldLine,
		)
	}

	// Perform the replacement
	newLine := strings.Replace(oldLine, oldPattern, newPattern, 1)

	// VERIFY 2: Ensure the replacement actually changed something
	if oldLine == newLine {
		return fmt.Errorf(
			"replacement failed in line %d of %s\nLine: %q",
			lineNumber, file.Name, oldLine,
		)
	}

	// VERIFY 3: Ensure the new pattern is now present
	if !strings.Contains(newLine, newPattern) {
		return fmt.Errorf(
			"verification failed: new pattern %q not found after replacement in line %d of %s",
			newPattern, lineNumber, file.Name,
		)
	}

	lines[lineIndex] = newLine

	// Write back to file
	newContent := strings.Join(lines, "\n")
	err = os.WriteFile("notes/"+file.Name, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", file.Name, err)
	}

	return nil
}

// InsertTodosIntoNote inserts todos at the top of a note (after title/frontmatter)
// Returns the insertion point line number for undo
func InsertTodosIntoNote(targetFileName string, todos []TodoWithMeta) (int, error) {
	// Validate target file exists
	_, err := LoadFileByName(targetFileName)
	if err != nil {
		return 0, fmt.Errorf("failed to load target file: %w", err)
	}

	// Read current content
	content, err := os.ReadFile("notes/" + targetFileName)
	if err != nil {
		return 0, fmt.Errorf("failed to read target file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find insertion point: after frontmatter (if present) or after "# Title" line
	insertionPoint := 0

	// Check if file starts with frontmatter (---)
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		// Find closing frontmatter delimiter
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				insertionPoint = i + 1
				break
			}
		}
		// Skip any empty lines after frontmatter
		for insertionPoint < len(lines) && strings.TrimSpace(lines[insertionPoint]) == "" {
			insertionPoint++
		}
	} else {
		// No frontmatter - find "# Title" line
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "# ") {
				insertionPoint = i + 1
				break
			}
		}
	}

	// Build todo lines to insert
	todoLines := []string{}

	// Add blank line before todos if there's content after title
	if insertionPoint < len(lines) && strings.TrimSpace(lines[insertionPoint]) != "" {
		todoLines = append(todoLines, "")
	}

	// Add each todo (with tags removed) and its subtasks
	for _, todo := range todos {
		cleanedLine := scripts.RemoveTalkToTags(todo.TodoLine)
		todoLines = append(todoLines, cleanedLine)

		// Add subtasks
		for _, subtask := range todo.Subtasks {
			todoLines = append(todoLines, subtask.Line)
		}
	}

	// Add blank line after todos
	todoLines = append(todoLines, "")

	// Insert into lines
	newLines := make([]string, 0, len(lines)+len(todoLines))
	newLines = append(newLines, lines[:insertionPoint]...)
	newLines = append(newLines, todoLines...)
	newLines = append(newLines, lines[insertionPoint:]...)

	// Write back to file
	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile("notes/"+targetFileName, []byte(newContent), 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to write target file: %w", err)
	}

	return insertionPoint, nil
}

// RemoveTodosFromNote removes N lines starting from a specific point (for undo)
func RemoveTodosFromNote(fileName string, insertionPoint int, lineCount int) error {
	// Read file content
	content, err := os.ReadFile("notes/" + fileName)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Validate bounds
	if insertionPoint < 0 || insertionPoint >= len(lines) {
		return fmt.Errorf("invalid insertion point")
	}

	endPoint := insertionPoint + lineCount
	if endPoint > len(lines) {
		endPoint = len(lines)
	}

	// Remove lines
	newLines := append(lines[:insertionPoint], lines[endPoint:]...)

	// Write back
	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile("notes/"+fileName, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
