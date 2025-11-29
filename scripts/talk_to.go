package scripts

import (
	"regexp"
	"strings"
)

// Regular expression to match to-talk-X tags
var talkToPattern = regexp.MustCompile(`to-talk-(\w+)`)

// ParseTodoLine extracts all person names from to-talk-X tags in a todo line
// Returns normalized (lowercase) person names
func ParseTodoLine(line string) []string {
	matches := talkToPattern.FindAllStringSubmatch(line, -1)
	people := make([]string, 0)

	for _, match := range matches {
		if len(match) > 1 {
			person := strings.ToLower(strings.TrimSpace(match[1]))
			people = append(people, person)
		}
	}

	return people
}

// RemoveTalkToTags removes all to-talk-X tags from a line
func RemoveTalkToTags(line string) string {
	result := talkToPattern.ReplaceAllString(line, "")
	// Clean up any extra whitespace left behind
	result = strings.TrimSpace(result)
	// Preserve leading checkbox formatting
	if strings.HasPrefix(line, "- [ ] ") || strings.HasPrefix(line, "- [x] ") {
		return result
	}
	return result
}

// SubtaskInfo represents a subtask with its line content, line number, and indentation level
type SubtaskInfo struct {
	Line       string
	LineNumber int // 1-indexed line number in the file
	Indent     int // Number of leading spaces/tabs
}

// ExtractSubtasks extracts nested subtasks following a parent todo
// Assumes lines is the full content split by newlines, and parentIndex is 0-indexed
func ExtractSubtasks(lines []string, parentIndex int) []SubtaskInfo {
	if parentIndex < 0 || parentIndex >= len(lines) {
		return []SubtaskInfo{}
	}

	parentIndent := getIndentLevel(lines[parentIndex])
	subtasks := []SubtaskInfo{}

	// Start from the line after the parent
	for i := parentIndex + 1; i < len(lines); i++ {
		line := lines[i]

		// Stop at blank lines
		if strings.TrimSpace(line) == "" {
			break
		}

		currentIndent := getIndentLevel(line)

		// Stop if we hit same or lower indentation level
		if currentIndent <= parentIndent {
			break
		}

		// This is a subtask - add it
		subtasks = append(subtasks, SubtaskInfo{
			Line:       line,
			LineNumber: i + 1, // Convert to 1-indexed
			Indent:     currentIndent,
		})
	}

	return subtasks
}

// getIndentLevel counts the number of leading spaces/tabs
// Tabs count as 4 spaces for consistency
func getIndentLevel(line string) int {
	indent := 0
	for _, char := range line {
		if char == ' ' {
			indent++
		} else if char == '\t' {
			indent += 4
		} else {
			break
		}
	}
	return indent
}
