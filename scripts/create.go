package scripts

import (
	"fmt"
	"os/exec"
	"time"
)

type MetaData struct {
	Key   string
	Value string
}

// TODO write tests for file 

func CreateTodo(title string, onFileCreated func(File) error) (File, error) {
	now := time.Now()
	return createFile(title, []string{"todo"}, "", now, false, onFileCreated)
}

func CreateMeeting(title string, onFileCreated func(File) error) (File, error) {
	now := time.Now()
	return createFile(title, []string{"meeting"}, "", now, true, onFileCreated)
}

// TODO move this to presentation layer
func OpenNoteInEditor(filePath string) {
	err := exec.Command("cursor", filePath).Run()
	if err != nil {
		fmt.Println("Error opening file in editor:", err)
		return
	}
}

func CreateStandup(getTeamNames func() ([]string, error), onFileCreated func(File) error) (File, error) {
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

	content += "\n## Other Points"

	return createFile(title, []string{"standup"}, content, nextFriday, false, onFileCreated)
}

func createFile(title string, tags []string, content string, dueAt time.Time, done bool, onFileCreated func(File) error) (File, error) {
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