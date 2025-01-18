package scripts

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TODO refactor this code to just be domain code then hook it into main.go

type GetFilesByIsDone func(isDone bool) ([]File, error)

func GetTodos(getFilesByIsDone GetFilesByIsDone) ([]File, error) {
	return getFilesByIsDone(false)
}

func QueryOpenTodos(queries []string, getFilesByIsDone GetFilesByIsDone) ([]File, error) {
	if len(queries) < 1 {
		return nil, nil
	}

	todos, err := getFilesByIsDone(false)
	if err != nil {
		return nil, err
	}

	var matchingTodos = make([]File, 0)

	for _, todo := range todos {
		for _, query := range queries {
			if todoMatchesQuery(todo, query) {
				matchingTodos = append(matchingTodos, todo)
				break
			}
		
			return matchingTodos, nil
		}
	}
	return matchingTodos, nil
}

func todoMatchesQuery(todo File, query string) bool {
	lowerCaseQuery := strings.ToLower(query)

	lowerCaseName := strings.ToLower(todo.Name)
	lowerCaseTitle := strings.ToLower(todo.Title)
	lowerCaseContent := strings.ToLower(todo.Content)

	lowerCaseTags := make([]string, len(todo.Tags))
	for i, tag := range todo.Tags {
		lowerCaseTags[i] = strings.ToLower(tag)
	}

	return strings.Contains(lowerCaseName, lowerCaseQuery) ||
		strings.Contains(lowerCaseTitle, lowerCaseQuery) ||
		strings.Contains(lowerCaseContent, lowerCaseQuery) ||
		containsTag(lowerCaseTags, lowerCaseQuery)
}

func containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(tag, query) {
			return true
		}
	}
	return false
}

func QueryFiles(queries []string) {
	if len(queries) == 0 {
		fmt.Println("No queries provided")
		return
	}

	if len(queries) == 1 {
		searchAllFilesRunCallbackWhenMatch(queries[0], func(fileName string) {
			printFileName(fileName)
			filesThatHaveBeenSearched = append(filesThatHaveBeenSearched, fileName)
		})
		return
	}

	firstQuery := queries[0]
	remainingQueries := queries[1:]

	searchAllFilesRunCallbackWhenMatch(firstQuery, func(fileName string) {})
	QueryPreviouslySearchedFiles(remainingQueries)
}

func QueryPreviouslySearchedFiles(queries []string) {
	var matchingFiles []string

	for _, query := range queries {
		var currentMatches []string
		err := searchFilesRunCallbackIfMatch(filesThatHaveBeenSearched, query, func(fileName string) {
			currentMatches = append(currentMatches, fileName)
		})
		if err != nil {
			fmt.Printf("Error searching for query '%s': %v\n", query, err)
			return
		}

		matchingFiles = currentMatches
		filesThatHaveBeenSearched = currentMatches
	}

	filesThatHaveBeenSearched = matchingFiles

	for _, file := range matchingFiles {
		fmt.Println(file)
	}
}

func SearchNotesByTags(tags []string) {

	resetFilesThatHaveBeenSearched()

	var matchingNotes []string

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory path:", err)
		return
	}

	notesPath := filepath.Join(currentDir, "/notes")
	err = filepath.Walk(notesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the current path is a file
		if !info.IsDir() {
			// Open the file
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// Create a scanner to read the file line by line
			scanner := bufio.NewScanner(file)

			// Flag to track if all tags are found in the file
			allTagsFound := true

			// Iterate over each line in the file
			for scanner.Scan() {
				line := scanner.Text()

				// Check if the line contains the "tags:" key
				if strings.HasPrefix(line, "tags:") {
					// Extract the tags from the line
					tagsLine := strings.TrimPrefix(line, "tags:")
					tagsLine = strings.TrimSpace(tagsLine)
					tagsLine = strings.Trim(tagsLine, "[]")
					fileTags := strings.Split(tagsLine, ",")

					// Check if all the specified tags are present in the file tags
					for _, tag := range tags {
						if !contains(fileTags, tag) {
							allTagsFound = false
							break
						}
					}

					// If all tags are found, add the file to the matching notes
					if allTagsFound {
						matchingNotes = append(matchingNotes, stripFilePathFromFileNameAndAddToFoundFilesList(path))
					}
					break
				}
			}

			// Check for any errors during scanning
			if err := scanner.Err(); err != nil {
				return err
			}
		}
		return nil
	})

	// Check for any errors during directory traversal
	if err != nil {
		fmt.Println("Error walking through files:", err)
		return
	}

	for _, note := range matchingNotes {
		fmt.Println(note)
	}
}
func SearchAllFilesPrintWhenMatch(lineQuery string) {
	searchAllFilesRunCallbackWhenMatch(lineQuery, printFileName)
}

func searchAllFilesRunCallbackWhenMatch(lineQuery string, callback func(string)) {
	resetFilesThatHaveBeenSearched()
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory path:", err)
		return
	}

	notesPath := filepath.Join(currentDir, "/notes")

	err = filepath.Walk(notesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the current path is a file
		if !info.IsDir() {
			err = searchFileForQueryAndRunCallbackIfMatch(path, lineQuery, callback)

			if err != nil {
				return err
			}
		}

		return nil
	})

	// Check for any errors during directory traversal
	if err != nil {
		fmt.Println("Error walking through files:", err)
		return
	}
}

func SearchLastFilesSearchedForQueryPrintWhenMatch(lineQuery string) {
	if len(filesThatHaveBeenSearched) == 0 {
		fmt.Println("No files have been searched yet")
		return
	}

	filesToSearch := filesThatHaveBeenSearched
	resetFilesThatHaveBeenSearched()

	searchFilesRunCallbackIfMatch(filesToSearch, lineQuery, printFileName)
}

func searchFilesRunCallbackIfMatch(files []string, lineQuery string, callback func(string)) error {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory path:", err)
		return err
	}

	for _, file := range files {
		filePath := filepath.Join(currentDir, "/notes", file)
		err = searchFileForQueryAndRunCallbackIfMatch(filePath, lineQuery, callback)
		if err != nil {
			return err
		}
	}

	return nil
}

func SearchFilesForUncompletedTasks(filePaths []string) {
	if len(filePaths) == 0 {
		fmt.Println("No files provided to search")
		return
	}

	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", filePath, err)
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNumber := 1

		for scanner.Scan() {
			line := scanner.Text()

			if strings.Contains(line, "- [ ] ") {
				fileName := filepath.Base(filePath)
				fmt.Printf("%s : %s: %d\n", fileName, line, lineNumber)
			}

			lineNumber++
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error scanning file %s: %v\n", filePath, err)
		}
	}
}

func SearchPreviousFilesForUncompletedTasks() {
	filesToSearch := make([]string, 0)
	for _, file := range filesThatHaveBeenSearched {
		filesToSearch = append(filesToSearch, addFilePathToFileName(file))
	}

	SearchFilesForUncompletedTasks(filesToSearch)
}

func searchFileForQueryAndRunCallbackIfMatch(path, lineQuery string, callback func(string)) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	lineQuery = strings.ToLower(lineQuery)
	// Iterate over each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ToLower(line)

		if strings.Contains(line, lineQuery) {
			fileName := stripFilePathFromFileNameAndAddToFoundFilesList(path)

			// Todo replace this everywhere its currently being used
			// fmt.Println(fileName)
			callback(fileName)
			break
		}
	}

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func searchFileForQueryAndPrintIfMatch(path, lineQuery string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	lineQuery = strings.ToLower(lineQuery)
	// Iterate over each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ToLower(line)

		if strings.Contains(line, lineQuery) {
			fmt.Println(stripFilePathFromFileNameAndAddToFoundFilesList(path))
			break
		}
	}

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func stripFilePathFromFileNameAndAddToFoundFilesList(filePath string) string {
	fileName := filepath.Base(filePath)
	filesThatHaveBeenSearched = append(filesThatHaveBeenSearched, fileName)
	return fileName
}

func resetFilesThatHaveBeenSearched() {
	filesThatHaveBeenSearched = make([]string, 0)
	filesThatHaveBeenSearchedSelectedIndex = -1
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

func addFilePathToFileName(fileName string) string {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory path:", err)
		return ""
	}

	filePath := filepath.Join(currentDir, "/notes", fileName)
	return filePath
}

func printFileName(name string) {
	fmt.Println(name)
}

func searchTodosWithDateCriteria(dateCheck func(dueDate string, dueDateParsed time.Time) bool) {
	resetFilesThatHaveBeenSearched()
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory path:", err)
		return
	}

	notesPath := filepath.Join(currentDir, "/notes")

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
			fileName := filepath.Base(path)

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
				dueDateParsed, err := time.Parse("2006-01-02", dueDate)
				if err != nil {
					return err
				}

				if dateCheck(dueDate, dueDateParsed) {
					filesThatHaveBeenSearched = append(filesThatHaveBeenSearched, fileName)
					fmt.Println(fileName)
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error walking through files:", err)
		return
	}
}

func GetOverdueTodos() {
	today := time.Now().Format("2006-01-02")
	searchTodosWithDateCriteria(func(dueDate string, _ time.Time) bool {
		return dueDate <= today
	})
}

func GetSoonTodos() {
	today := time.Now()
	oneWeekFromNow := today.AddDate(0, 0, 7)
	todayStr := today.Format("2006-01-02")

	searchTodosWithDateCriteria(func(dueDate string, dueDateParsed time.Time) bool {
		return dueDate <= todayStr || dueDateParsed.Before(oneWeekFromNow) || dueDateParsed.Equal(oneWeekFromNow)
	})
}

// TODO currently it matters that date-due in on the line above done: false, I should fix this when I get time
func GetTodosWithNoDueDate() {
	resetFilesThatHaveBeenSearched()
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory path:", err)
		return
	}

	notesPath := filepath.Join(currentDir, "/notes")

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
			hasDueDate := false
			fileName := filepath.Base(path)

			for scanner.Scan() {
				line := scanner.Text()

				if strings.Contains(line, "done: false") {
					isATodo = true
				}

				if strings.Contains(line, "date-due:") {
					dueDateValue := strings.TrimSpace(strings.TrimPrefix(line, "date-due:"))
					hasDueDate = dueDateValue != ""
				}

				// If we found both fields, we can stop scanning
				if isATodo || hasDueDate {
					break
				}
			}

			if err := scanner.Err(); err != nil {
				return err
			}

			// If the todo is not done and has no due date
			if isATodo && !hasDueDate {
				filesThatHaveBeenSearched = append(filesThatHaveBeenSearched, fileName)
				fmt.Println(fileName)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error walking through files:", err)
		return
	}
}
