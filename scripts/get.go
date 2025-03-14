package scripts

import (
	"bufio"
	"fmt"
	"sort"
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

// CalculateTodoScore calculates a weighted score for a todo based on:
// - Days until due (50% weight)
// - Days since creation (20% weight)
// - Manual priority (30% weight)
// Lower scores indicate higher priority
func CalculateTodoScore(todo File) float64 {
	now := time.Now()

	// Calculate days until due
	daysUntilDue := 0.0
	if todo.DueAt.Year() > 2100 {
		// No due date, assign a high value
		daysUntilDue = 365.0
	} else {
		// Calculate days until due (can be negative for overdue tasks)
		daysUntilDue = todo.DueAt.Sub(now).Hours() / 24.0
	}

	// Calculate days since creation
	daysSinceCreation := 0.0
	if !todo.CreatedAt.IsZero() {
		daysSinceCreation = now.Sub(todo.CreatedAt).Hours() / 24.0
	}

	// Convert priority to a numeric value (P1=1, P2=2, P3=3)
	priorityValue := float64(todo.Priority)

	// Calculate the weighted score
	// Note: We're using the raw priority value (lower is higher priority)
	// For days until due, lower values (or negative for overdue) mean higher priority
	// For days since creation, higher values mean the task has been waiting longer
	score := (daysUntilDue * 0.5) + (daysSinceCreation * 0.2) + (priorityValue * 0.3)

	return score
}

// SortTodosByPriorityAndDueDate sorts a slice of File (todos) using a weighted scoring system
// that considers days until due date (50%), days since creation (20%), and manual priority (30%).
// It modifies the slice in place and also returns it for convenience.
func SortTodosByPriorityAndDueDate(todos []File) []File {
	sort.Slice(todos, func(i, j int) bool {
		scoreI := CalculateTodoScore(todos[i])
		scoreJ := CalculateTodoScore(todos[j])

		// Lower scores come first (higher priority)
		return scoreI < scoreJ
	})

	return todos
}

func GetTodos(getFilesByIsDone GetFilesByIsDone) ([]File, error) {
	todos, err := getFilesByIsDone(false)
	if err != nil {
		return nil, err
	}

	return SortTodosByPriorityAndDueDate(todos), nil
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
		matchCount := 0
		for _, query := range queries {
			if fileMatchesQuery(todo, query) {
				matchCount++
			}
		}
		if matchCount == len(queries) {
			matchingTodos = append(matchingTodos, todo)
		}
	}

	return SortTodosByPriorityAndDueDate(matchingTodos), nil
}

func QueryAllFiles(queries []string, getFilesByQuery GetFilesByQuery) ([]File, error) {
	if len(queries) < 1 {
		return make([]File, 0), nil
	}

	files, err := getFilesByQuery(queries[0])
	if err != nil {
		return nil, err
	}

	if len(queries) > 1 {
		matchingFiles := QueryFiles(queries[1:], files)
		return matchingFiles, nil
	} else {
		// Sort by priority and due date if these are todos
		hasTodos := false
		for _, file := range files {
			if !file.DueAt.IsZero() {
				hasTodos = true
				break
			}
		}

		if hasTodos {
			return SortTodosByPriorityAndDueDate(files), nil
		}
		return files, nil
	}
}

func QueryFiles(queries []string, files []File) []File {
	if len(queries) < 1 {
		return make([]File, 0)
	}

	matchingFiles := make([]File, 0)

	// Check all files against all queries (AND logic)
	for _, file := range files {
		matchesAllQueries := true
		for _, query := range queries {
			if !fileMatchesQuery(file, query) {
				matchesAllQueries = false
				break
			}
		}
		if matchesAllQueries {
			matchingFiles = append(matchingFiles, file)
		}
	}

	// Sort by priority and due date if these are todos
	// We can determine if files are todos by checking if at least one file has a defined DueAt time
	hasTodos := false
	for _, file := range matchingFiles {
		if !file.DueAt.IsZero() {
			hasTodos = true
			break
		}
	}

	if hasTodos {
		return SortTodosByPriorityAndDueDate(matchingFiles)
	}

	return matchingFiles
}

func SearchNotesByTags(tags []string, getFilesByTag GetFilesByTag) ([]File, error) {
	files, err := getFilesByTag(tags)
	if err != nil {
		return nil, err
	}

	// Sort by priority and due date if these are todos
	hasTodos := false
	for _, file := range files {
		if !file.DueAt.IsZero() {
			hasTodos = true
			break
		}
	}

	if hasTodos {
		return SortTodosByPriorityAndDueDate(files), nil
	}

	return files, nil
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
	files, err := getFiles(func(dueDate string, _ time.Time) bool {
		return dueDate <= today
	})
	if err != nil {
		return nil, err
	}

	return SortTodosByPriorityAndDueDate(files), nil
}

func GetSoonTodos(getFiles GetFilesByDateQuery) ([]File, error) {
	now := time.Now()
	oneWeekFromNow := now.AddDate(0, 0, 7)

	files, err := getFiles(func(dueDate string, dueDateParsed time.Time) bool {
		// Only use time.Time comparison for consistency
		// A todo is considered "soon" if it's due within a week
		return dueDateParsed.Before(oneWeekFromNow) || dueDateParsed.Equal(oneWeekFromNow)
	})
	if err != nil {
		return nil, err
	}

	return SortTodosByPriorityAndDueDate(files), nil
}

func GetTodosWithNoDueDate(getFiles GetFilesByDateQuery) ([]File, error) {
	today := time.Now()
	oneHundredYearsFromNow := today.AddDate(100, 0, 0)

	files, err := getFiles(func(dueDate string, dueDateParsed time.Time) bool {
		return dueDateParsed.After(oneHundredYearsFromNow)
	})
	if err != nil {
		return nil, err
	}

	return SortTodosByPriorityAndDueDate(files), nil
}

func GetCompletedTodosByDateRange(startDate, endDate string, getFilesByDateRangeQuery GetFilesByDateQuery) ([]File, error) {
	files, err := getFilesByDateRangeQuery(func(dueDate string, dueDateParsed time.Time) bool {
		// Check if the dueDate is within the range (inclusive)
		return dueDate >= startDate && dueDate <= endDate
	})
	if err != nil {
		return nil, err
	}

	return SortTodosByPriorityAndDueDate(files), nil
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
	lowerQuery := strings.ToLower(query)
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), lowerQuery) {
			return true
		}
	}
	return false
}
