package data

import (
	"cli-notes/scripts"
	"os"
	"path/filepath"
	"strings"
)

// QueryAllObjectives returns all parent objectives
func QueryAllObjectives() ([]scripts.File, error) {
	return queryAllFiles("objective-role: parent")
}

// QueryChildrenByObjectiveID returns all todos with a specific objective-id
func QueryChildrenByObjectiveID(objectiveID string, includeDone bool) ([]scripts.File, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	notesPath := filepath.Join(currentDir, DirectoryPath)
	matchingFiles := make([]scripts.File, 0)

	err = filepath.Walk(notesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := getFileIfQueryMatches(path, "objective-id:")
			if err != nil {
				return err
			}

			// Only include children, not the parent objective itself
			if file != nil && file.ObjectiveID == objectiveID && file.ObjectiveRole != "parent" {
				// Filter by done status if specified
				if includeDone || !file.Done {
					matchingFiles = append(matchingFiles, *file)
				}
			}
		}
		return nil
	})

	return matchingFiles, err
}

// GetObjectiveByID finds a parent objective by its ID
func GetObjectiveByID(objectiveID string) (*scripts.File, error) {
	objectives, err := QueryAllObjectives()
	if err != nil {
		return nil, err
	}

	for _, obj := range objectives {
		if obj.ObjectiveID == objectiveID {
			return &obj, nil
		}
	}
	return nil, nil // Not found
}

// QueryTodosWithoutObjective returns all todos not linked to any objective
func QueryTodosWithoutObjective(queries []string) ([]scripts.File, error) {
	// First get all non-done todos
	allTodos, err := QueryFilesByDone(false)
	if err != nil {
		return nil, err
	}

	// Filter by queries if provided
	var filteredTodos []scripts.File
	if len(queries) == 0 {
		filteredTodos = allTodos
	} else {
		// Apply comma-separated grep chain
		for _, todo := range allTodos {
			matchesAll := true
			for _, query := range queries {
				queryLower := strings.ToLower(query)
				// Check if query matches title, tags, or content
				titleMatch := strings.Contains(strings.ToLower(todo.Title), queryLower)
				contentMatch := strings.Contains(strings.ToLower(todo.Content), queryLower)
				tagsMatch := false
				for _, tag := range todo.Tags {
					if strings.Contains(strings.ToLower(tag), queryLower) {
						tagsMatch = true
						break
					}
				}

				if !titleMatch && !contentMatch && !tagsMatch {
					matchesAll = false
					break
				}
			}
			if matchesAll {
				filteredTodos = append(filteredTodos, todo)
			}
		}
	}

	// Filter out todos with objective-id or objective-role
	result := make([]scripts.File, 0)
	for _, todo := range filteredTodos {
		if todo.ObjectiveID == "" && todo.ObjectiveRole != "parent" {
			result = append(result, todo)
		}
	}

	return result, nil
}

// GetCompletionStats returns (complete, total) counts for an objective
func GetCompletionStats(objectiveID string) (int, int, error) {
	children, err := QueryChildrenByObjectiveID(objectiveID, true)
	if err != nil {
		return 0, 0, err
	}

	complete := 0
	for _, child := range children {
		if child.Done {
			complete++
		}
	}

	return complete, len(children), nil
}

// QueryNonFinishedObjectives returns all parent objectives that are not done
func QueryNonFinishedObjectives() ([]scripts.File, error) {
	allObjectives, err := QueryAllObjectives()
	if err != nil {
		return nil, err
	}

	nonFinished := make([]scripts.File, 0)
	for _, obj := range allObjectives {
		if !obj.Done {
			nonFinished = append(nonFinished, obj)
		}
	}

	return nonFinished, nil
}
