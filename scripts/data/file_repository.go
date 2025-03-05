package data

import (
	"bufio"
	"cli-notes/scripts"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const DirectoryPath = "/notes"
const dateFormat = "2006-01-02"

type metaData struct {
	Key   string
	Value string
}

func WriteFile(newFile scripts.File) error {
	meta := []metaData{
		{
			Key:   "title",
			Value: newFile.Title,
		},
		{
			Key:   "date-created",
			Value: timeToString(newFile.CreatedAt),
		},
		{
			Key:   "tags",
			Value: fmt.Sprintf("%v", newFile.Tags),
		},
		{
			Key:   "date-due",
			Value: timeToString(newFile.DueAt),
		},
		{
			Key:   "done",
			Value: fmt.Sprintf("%v", newFile.Done),
		},
	}

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory path:", err)
		return err
	}

	notesPath := filepath.Join(currentDir, DirectoryPath)
	filePath := filepath.Join(notesPath, newFile.Name)

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteString("---\n")
	if err != nil {
		fmt.Println("Error writing top of meta tags:", err)
		return err
	}

	for _, m := range meta {
		_, err = fmt.Fprintf(file, "%s: %s\n", m.Key, m.Value)
		if err != nil {
			fmt.Println("Error writing meta data:", err)
			return err
		}
	}

	_, err = file.WriteString("---\n\n")
	if err != nil {
		fmt.Println("Error writing bottom of meta tags:", err)
		return err
	}

	_, err = file.WriteString(newFile.Content)
	if err != nil {
		fmt.Println("Error writing file content:", err)
		return err
	}

	return nil

}

func QueryFilesByDone(isDone bool) ([]scripts.File, error) {

	var query = fmt.Sprintf("done: %v", isDone)

	return queryAllFiles(query)
}

func QueryFiles(query string) ([]scripts.File, error) {
	return queryAllFiles(query)
}

func QueryTodosWithDateCriteria(dateCheck func(dueDate string, dueDateParsed time.Time) bool) ([]scripts.File, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory path:", err)
		return nil, err
	}

	notesPath := filepath.Join(currentDir, DirectoryPath)
	matchingFiles := make([]scripts.File, 0)

	err = filepath.Walk(notesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			isATodo := false
			dueDate := ""

			for scanner.Scan() {
				line := scanner.Text()

				if strings.Contains(line, "done: false") {
					isATodo = true
				}

				if strings.Contains(line, "date-due:") {
					dueDate = strings.TrimSpace(strings.TrimPrefix(line, "date-due:"))
				}

				if isATodo && dueDate != "" {
					break
				}
			}

			if err := scanner.Err(); err != nil {
				return err
			}

			if isATodo && dueDate != "" {
				dueDateParsed, err := time.Parse(dateFormat, dueDate)
				if err != nil {
					return err
				}

				if dateCheck(dueDate, dueDateParsed) {
					matchingFile, err := getFileIfQueryMatches(path, "date-due:")
					if err != nil {
						return err
					}
					matchingFiles = append(matchingFiles, *matchingFile)
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return matchingFiles, err
}

func QueryNotesByTags(tags []string) ([]scripts.File, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	matchingNotes := make([]scripts.File, 0)
	notesPath := filepath.Join(currentDir, DirectoryPath)

	err = filepath.Walk(notesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			allTagsFound := false

			for scanner.Scan() {
				line := scanner.Text()

				if strings.HasPrefix(line, "tags:") {
					tagsLine := strings.TrimPrefix(line, "tags:")
					tagsLine = strings.TrimSpace(tagsLine)
					tagsLine = strings.Trim(tagsLine, "[]")

					// Changed: Split by spaces instead of commas, and handle both formats
					var fileTags []string
					if strings.Contains(tagsLine, ",") {
						// Handle comma-separated format
						parts := strings.Split(tagsLine, ",")
						for _, p := range parts {
							fileTags = append(fileTags, strings.TrimSpace(p))
						}
					} else {
						// Handle space-separated format
						fileTags = strings.Fields(tagsLine)
					}

					// Check if all query tags are in the file tags
					allTagsFound = true
					for _, tag := range tags {
						if !contains(fileTags, tag) {
							allTagsFound = false
							break
						}
					}

					if allTagsFound {
						matchingFile, err := getFileIfQueryMatches(path, "tags:")
						if err != nil {
							return err
						}
						matchingNotes = append(matchingNotes, *matchingFile)
					}
					break
				}
			}

			if err := scanner.Err(); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return matchingNotes, nil
}

func timeToString(time time.Time) string {
	return time.Format(dateFormat)
}

func queryAllFiles(lineQuery string) ([]scripts.File, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory path:", err)
		return nil, err
	}

	notesPath := filepath.Join(currentDir, DirectoryPath)

	var matchingFiles = make([]scripts.File, 0)

	err = filepath.Walk(notesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the current path is a file
		if !info.IsDir() {
			file, err := getFileIfQueryMatches(path, lineQuery)

			if err != nil {
				return err
			}

			if file != nil {
				matchingFiles = append(matchingFiles, *file)
			}
		}

		return nil
	})

	// Check for any errors during directory traversal
	if err != nil {
		fmt.Println("Error walking through files:", err)
		return nil, err
	}

	return matchingFiles, err
}

func getFileIfQueryMatches(path, lineQuery string) (*scripts.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Initialize file struct
	result := &scripts.File{
		Name: filepath.Base(path),
	}

	// State tracking
	inMetadata := false
	var content strings.Builder
	lineQuery = strings.ToLower(lineQuery)
	foundMatch := false

	for scanner.Scan() {
		line := scanner.Text()
		lowerLine := strings.ToLower(line)

		// Check for metadata section
		if line == "---" {
			if !inMetadata {
				inMetadata = true
				continue
			} else {
				inMetadata = false
				continue
			}
		}

		// Parse metadata
		if inMetadata {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "title":
				result.Title = value
			case "tags":
				// Remove brackets and split by comma
				value = strings.Trim(value, "[]")
				if value != "" {
					tags := strings.Split(value, ",")
					for i, tag := range tags {
						tags[i] = strings.TrimSpace(tag)
					}
					result.Tags = tags
				}
			case "date-created":
				result.CreatedAt, _ = time.Parse(dateFormat, value)
			case "date-due":
				parsedTime, err := time.Parse(dateFormat, value)
				if err != nil {
					// Set to a far future date if parsing fails
					result.DueAt = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
				} else {
					result.DueAt = parsedTime
				}
			case "done":
				result.Done = value == "true"
			}
		} else {
			// Append to content
			content.WriteString(line)
			content.WriteString("\n")
		}

		// Check for query match
		if strings.Contains(lowerLine, lineQuery) {
			foundMatch = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if !foundMatch {
		return nil, nil
	}

	result.Content = content.String()
	return result, nil
}

func contains(slice []string, item string) bool {
	for _, val := range slice {
		trimmedVal := strings.TrimSpace(val)
		trimmedItem := strings.TrimSpace(item)

		if trimmedVal == trimmedItem {
			return true
		}
	}
	return false
}

// Helper function to check if a file is a date range query note
func isDateRangeQueryNote(file *scripts.File) bool {
	if file == nil || file.Tags == nil {
		return false
	}

	for _, tag := range file.Tags {
		if tag == "date-range-query" {
			return true
		}
	}

	return false
}

func QueryCompletedTodosByDateRange(dateCheck func(dueDate string, dueDateParsed time.Time) bool) ([]scripts.File, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory path:", err)
		return nil, err
	}

	notesPath := filepath.Join(currentDir, DirectoryPath)
	matchingFiles := make([]scripts.File, 0)

	err = filepath.Walk(notesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			isCompletedTodo := false
			dueDate := ""

			for scanner.Scan() {
				line := scanner.Text()

				if strings.Contains(line, "done: true") {
					isCompletedTodo = true
				}

				if strings.Contains(line, "date-due:") {
					dueDate = strings.TrimSpace(strings.TrimPrefix(line, "date-due:"))
				}

				if isCompletedTodo && dueDate != "" {
					break
				}
			}

			if err := scanner.Err(); err != nil {
				return err
			}

			if isCompletedTodo && dueDate != "" {
				dueDateParsed, err := time.Parse(dateFormat, dueDate)
				if err != nil {
					return err
				}

				if dateCheck(dueDate, dueDateParsed) {
					matchingFile, err := getFileIfQueryMatches(path, "done: true")
					if err != nil {
						return err
					}

					// Only include the file if it's not a date range query note
					if matchingFile != nil && !isDateRangeQueryNote(matchingFile) {
						matchingFiles = append(matchingFiles, *matchingFile)
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return matchingFiles, err
}
