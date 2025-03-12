package scripts

import (
	"fmt"
	"strings"
	"time"
)

type MetaData struct {
	Key   string
	Value string
}

type OnFileCreated = func(File) error

func CreateTodo(title string, onFileCreated OnFileCreated) (File, error) {
	now := time.Now()
	return createFile(title, []string{"todo"}, "", now, false, onFileCreated)
}

func CreateMeeting(title string, onFileCreated OnFileCreated) (File, error) {
	now := time.Now()
	return createFile(title, []string{"meeting"}, "", now, true, onFileCreated)
}

func CreateStandup(getTeamNames func() ([]string, error), onFileCreated OnFileCreated) (File, error) {
	teamNames, err := getTeamNames()
	if err != nil {
		return File{}, err
	}

	now := time.Now()
	nextFriday := now
	for nextFriday.Weekday() != time.Friday {
		nextFriday = nextFriday.Add(24 * time.Hour)
	}

	title := "standup"
	content := fmt.Sprintf("# %v\n\n", title)
	weekdays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}

	for _, name := range teamNames {
		content += fmt.Sprintf("## %s\n\n", name)
		for _, day := range weekdays {
			content += fmt.Sprintf("### %s Plan\n\n", day)
		}
		content += "\n"
	}

	content += "## Other Points"

	return createFile(title, []string{"standup"}, content, nextFriday, false, onFileCreated)
}

func CreateSevenQuestions(title string, onFileCreated OnFileCreated) (File, error) {
	now := time.Now()
	questions := []string{
		"What is the situation and how does it affect me?",
		"What have I been told to and why?",
		"What effects do I need to achieve and what direction must I give?",
		"Where can I best accomplish each action or effect?",
		"What resources do I need to accomplish each action or effect?",
		"When and where do these actions take place in relation to each other?",
		"What control measures do I need to impose?",
	}

	content := fmt.Sprintf("# %v", title)

	for _, question := range questions {
		content += fmt.Sprintf("\n\n\n## %v", question)
	}

	return createFile(title, []string{"plan"}, content, now, false, onFileCreated)
}

func CreateDateRangeQueryNote(startDate, endDate string, files []File, onFileCreated OnFileCreated) (File, error) {
	now := time.Now()
	exampleSummary := `
	Completed Tasks

🎯 People Management

- Put together the team members for the cross-functional team with Jeff 

- Onboarding 2 New Team Members into Green Lizards

🤖 Technical Execution

- Developing in the sprint at good pace and maintaining my Manager role at the same time

🎨 Jetpack Compose Migration (specific high level projects)

- Successfully prototyped Figma→Compose theme extraction using AI:

  - Automated color/typography/shape conversions

  - Validated against Figma specs

  - Zero design workflow changes required

🔧 Operations

- Created a plan to improve Android install conversion metrics from Ops Social meeting and UI/UX Show & Tell

  - Focus on improving the app experience for users

  - Chat is the first area, reported by users as the best feature but also 29% of users report issues with it

  - Will also look for improvements in the Booking flows`

	title := fmt.Sprintf("Date Range Query %s - %s", startDate, endDate)

	content := fmt.Sprintf("# %s\n\n", title)
	content += fmt.Sprintf("## Example Summary\n\n%s\n\n", exampleSummary)
	content += fmt.Sprintf("## Completed notes between %s and %s\n\n", startDate, endDate)

	for _, file := range files {
		content += fmt.Sprintf("### %s\n\n", file.Title)
		content += fmt.Sprintf("- **File**: %s\n", file.Name)
		content += fmt.Sprintf("- **Due Date**: %s\n", file.DueAt.Format("2006-01-02"))
		content += fmt.Sprintf("- **Tags**: %s\n\n", strings.Join(file.Tags, ", "))

		// Extract content without frontmatter
		fileContent := extractContentWithoutFrontmatter(file.Content)
		content += fileContent + "\n\n---\n\n"
	}

	return createFile(title, []string{"date-range-query"}, content, now, false, onFileCreated)
}

// Helper function to extract content without frontmatter
func extractContentWithoutFrontmatter(content string) string {
	// If content is empty, return empty string
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")

	// Check if content starts with frontmatter delimiter
	if len(lines) > 0 && lines[0] == "---" {
		// Find the closing frontmatter delimiter
		closingIndex := -1
		for i := 1; i < len(lines); i++ {
			if lines[i] == "---" {
				closingIndex = i
				break
			}
		}

		// If we found a closing delimiter, extract content after it
		if closingIndex > 0 && closingIndex < len(lines)-1 {
			return strings.Join(lines[closingIndex+1:], "\n")
		}
	}

	// If no frontmatter or incomplete frontmatter, return original content
	return content
}

func createFile(title string, tags []string, content string, dueAt time.Time, done bool, onFileCreated OnFileCreated) (File, error) {
	now := time.Now()
	date := now.Format("2006-01-02")
	name := fmt.Sprintf("%v-%v.md", title, date)

	if content == "" {
		content = fmt.Sprintf("# %v", title)
	}

	newFile := File{
		Name:      name,
		Title:     title,
		Tags:      tags,
		CreatedAt: now,
		DueAt:     dueAt,
		Done:      done,
		Content:   content,
	}

	if err := onFileCreated(newFile); err != nil {
		return File{}, err
	}

	return newFile, nil
}
