package scripts

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

// TODO refactor this code to just be domain code then hook it into main.go

type GetFilesByIsDone func(isDone bool) ([]File, error)
type GetFilesByQuery func(query string) ([]File, error)
type GetFilesByTag func(tags []string) ([]File, error)
// TODO simplify this when done refactoring
type DateQuery func(dueDate string, dueDateParsed time.Time) bool
type GetFilesByDateQuery func(dateQuery DateQuery) ([]File, error)

func GetTodos(getFilesByIsDone GetFilesByIsDone) ([]File, error) {
	return getFilesByIsDone(false)
}

func QueryOpenTodos(queries []string, getFilesByIsDone GetFilesByIsDone) ([]File, error) {
	if len(queries) < 1 {
		return make([]File, 0), nil
	}

	todos, err := getFilesByIsDone(false)
	if err != nil {
		return nil, err
	}

	var matchingTodos = make([]File, 0)

	for _, todo := range todos {
		for _, query := range queries {
			if fileMatchesQuery(todo, query) {
				matchingTodos = append(matchingTodos, todo)
				break
			}

		}
	}
	return matchingTodos, nil
}

func QueryFiles(queries []string, getFilesByQuery GetFilesByQuery) ([]File, error) {
	if len(queries) < 1 {
		return make([]File, 0), nil
	}

	var filesToQuery []File = nil
	matchingFiles := make([]File, 0)

	for _, query := range queries {

		if filesToQuery == nil {
			filesToQuery = make([]File, 0)
			files, err := getFilesByQuery(query)
			if err != nil {
				return nil, err
			}
			filesToQuery = files
		}

			for _, file := range filesToQuery {

				if fileMatchesQuery(file, query) {
					matchingFiles = append(matchingFiles, file)
					break
				}

		}

	}

	return matchingFiles, nil
}

func SearchNotesByTags(tags []string, getFilesByTag GetFilesByTag) ([]File, error) {
	return getFilesByTag(tags)
}


func GetUncompletedTasksInFiles(files []File) ([]string, error) {
	if len(files) == 0 {
		return make([]string, 0), nil
	}

	tasks := make([]string, 0)

	for _, file := range files {
		
		scanner := bufio.NewScanner(strings.NewReader(file.Content))
		lineNumber := 1

		for scanner.Scan() {
			line := scanner.Text()

			if strings.Contains(line, "- [ ] ") {
				task := fmt.Sprintf("%s : %s: %d\n", file.Name, line, lineNumber)
				tasks = append(tasks, task)
			}

			lineNumber++
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	
	return tasks, nil
}

func GetOverdueTodos(getFiles GetFilesByDateQuery) ([]File, error) {
	today := time.Now().Format("2006-01-02")
	return getFiles(func(dueDate string, _ time.Time) bool {
		return dueDate <= today
	})
}

func GetSoonTodos(getFiles GetFilesByDateQuery) ([]File, error) {
	today := time.Now()
	oneWeekFromNow := today.AddDate(0, 0, 7)
	todayStr := today.Format("2006-01-02")

	return getFiles(func(dueDate string, dueDateParsed time.Time) bool {
		return dueDate <= todayStr || dueDateParsed.Before(oneWeekFromNow) || dueDateParsed.Equal(oneWeekFromNow)
	})
}

func GetTodosWithNoDueDate(getFiles GetFilesByDateQuery) ([]File, error) {
	today := time.Now()
	oneHundredYearsFromNow := today.AddDate(100, 0, 0)

	return getFiles(func(dueDate string, dueDateParsed time.Time) bool {
		return dueDateParsed.After(oneHundredYearsFromNow)
	})

}


func fileMatchesQuery(todo File, query string) bool {
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