package scripts

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CreateParentObjective creates a new objective with parent role
func CreateParentObjective(title string, onFileCreated OnFileCreated) (File, error) {
	objectiveID, err := GenerateObjectiveID()
	if err != nil {
		return File{}, err
	}

	now := time.Now()
	date := now.Format("2006-01-02")
	name := fmt.Sprintf("%s-%s.md", title, date)

	content := fmt.Sprintf("# %s\n\n## Overview\n\n## Key Requirements\n\n", title)

	newFile := File{
		Name:          name,
		Title:         title,
		Tags:          []string{"objective"},
		CreatedAt:     now,
		DueAt:         time.Time{}, // No due date for parent
		Done:          false,
		Content:       content,
		Priority:      P2,
		ObjectiveRole: "parent",
		ObjectiveID:   objectiveID,
	}

	if err := onFileCreated(newFile); err != nil {
		return File{}, err
	}

	return newFile, nil
}

// ConvertTodoToParentObjective converts an existing todo to a parent objective
func ConvertTodoToParentObjective(file File, writeFile WriteFile) (File, error) {
	// Generate new objective ID
	objectiveID, err := GenerateObjectiveID()
	if err != nil {
		return File{}, err
	}

	// Read latest content to avoid losing updates
	updatedFile, err := readLatestFileContent(file)
	if err != nil {
		return File{}, err
	}

	// If already has objective-id (is a child), that will be replaced
	updatedFile.ObjectiveRole = "parent"
	updatedFile.ObjectiveID = objectiveID

	// Add "objective" tag if not present
	hasObjectiveTag := false
	for _, tag := range updatedFile.Tags {
		if tag == "objective" {
			hasObjectiveTag = true
			break
		}
	}
	if !hasObjectiveTag {
		updatedFile.Tags = append(updatedFile.Tags, "objective")
	}

	if err := writeFile(updatedFile); err != nil {
		return File{}, err
	}

	return updatedFile, nil
}

// CreateChildTodo creates a new todo linked to an objective
func CreateChildTodo(title string, parentObjective File, onFileCreated OnFileCreated) (File, error) {
	now := time.Now()
	date := now.Format("2006-01-02")
	name := fmt.Sprintf("%s-%s.md", title, date)

	content := fmt.Sprintf("# %s", title)

	// Inherit tags from parent (exclude "objective" tag)
	childTags := []string{"todo"}
	for _, tag := range parentObjective.Tags {
		if tag != "objective" {
			// Check if child already has this tag
			hasTag := false
			for _, childTag := range childTags {
				if childTag == tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				childTags = append(childTags, tag)
			}
		}
	}

	newFile := File{
		Name:        name,
		Title:       title,
		Tags:        childTags,
		CreatedAt:   now,
		DueAt:       now,
		Done:        false,
		Content:     content,
		Priority:    P2,
		ObjectiveID: parentObjective.ObjectiveID,
	}

	if err := onFileCreated(newFile); err != nil {
		return File{}, err
	}

	return newFile, nil
}

// LinkTodoToObjective links an existing todo to an objective
func LinkTodoToObjective(todo File, parentObjective File, writeFile WriteFile) error {
	// Prevent linking parent objectives as children
	if todo.ObjectiveRole == "parent" {
		return fmt.Errorf("cannot link parent objectives as children")
	}

	// Read latest content
	updatedTodo, err := readLatestFileContent(todo)
	if err != nil {
		return err
	}

	// Set objective ID
	updatedTodo.ObjectiveID = parentObjective.ObjectiveID

	// Inherit tags from parent (only tags child doesn't already have)
	childTagSet := make(map[string]bool)
	for _, tag := range updatedTodo.Tags {
		childTagSet[tag] = true
	}

	for _, parentTag := range parentObjective.Tags {
		if parentTag != "objective" && !childTagSet[parentTag] {
			updatedTodo.Tags = append(updatedTodo.Tags, parentTag)
		}
	}

	return writeFile(updatedTodo)
}

// UnlinkTodoFromObjective removes objective association
func UnlinkTodoFromObjective(todo File, writeFile WriteFile) error {
	updatedTodo, err := readLatestFileContent(todo)
	if err != nil {
		return err
	}

	updatedTodo.ObjectiveID = ""

	return writeFile(updatedTodo)
}

// DeleteParentObjective deletes a parent and unlinks all children
func DeleteParentObjective(parent File, getChildrenFunc func(string, bool) ([]File, error), writeFile WriteFile) error {
	if parent.ObjectiveRole != "parent" {
		return fmt.Errorf("file is not a parent objective")
	}

	// Get all children (including done)
	children, err := getChildrenFunc(parent.ObjectiveID, true)
	if err != nil {
		return err
	}

	// Unlink all children
	for _, child := range children {
		if err := UnlinkTodoFromObjective(child, writeFile); err != nil {
			return fmt.Errorf("failed to unlink child %s: %w", child.Name, err)
		}
	}

	// Delete the parent file
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	filePath := filepath.Join(currentDir, "notes", parent.Name)
	return os.Remove(filePath)
}
