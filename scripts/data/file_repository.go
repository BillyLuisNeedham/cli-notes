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
		return nil , err
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
				result.DueAt, _ = time.Parse(dateFormat, value)
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
