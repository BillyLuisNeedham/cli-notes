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

// TODO write tests when working
func CreateTodo(title string, onFileCreated func(File) error) (File, error) {
	tags := []string{"todo"}
	now := time.Now()
	date := now.Format("2006-01-02")
	name := fmt.Sprintf("%v-%v.md", title, date)
	content := fmt.Sprintf("# %v", title)

	newFile := File{
		Name:      name,
		Title:     title,
		Tags:      tags,
		CreatedAt: now,
		DueAt:     now,
		Done:      false,
		Content:   content,
	}

	if err := onFileCreated(newFile); err != nil {
		return File{}, err
	}

	return newFile, nil
}

func CreateMeeting(title string, onFileCreated func(File) error) (File, error) {

	tags := []string{"meeting"}
	now := time.Now()
	date := now.Format("2006-01-02")
	name := fmt.Sprintf("%v-%v.md", title, date)
	content := fmt.Sprintf("# %v", title)

	newFile := File{
		Name:      name,
		Title:     title,
		Tags:      tags,
		CreatedAt: now,
		DueAt:     now,
		Done:      true,
		Content:   content,
	}

	if err := onFileCreated(newFile); err != nil {
		return File{}, err
	}

	return newFile, nil

}

func OpenNoteInEditor(filePath string) {
	err := exec.Command("cursor", filePath).Run()
	if err != nil {
		fmt.Println("Error opening file in editor:", err)
		return
	}
}

func CreateStandup( getTeamNames func() ([]string, error), onFileCreated func(File) error) (File, error) {

	teamNames, err := getTeamNames()
	if err != nil {
		return File{}, err
	}

	tags := []string{"standup"}
	now := time.Now()
	date := now.Format("2006-01-02")
	title := "Standup"
	name := fmt.Sprintf("%v-%v.md", title, date)
	content := fmt.Sprintf("# %v\n\n", title)

	nextFriday := now
	for nextFriday.Weekday() != time.Friday {
		nextFriday = nextFriday.Add(24 * time.Hour)
	}

	weekdays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	
	for _, name := range teamNames {
		content += fmt.Sprintf("## %s\n\n", name)

      for _, day := range weekdays {
		content += fmt.Sprintf("### %s Plan\n\n### %s Blockers\n\n", day, day)
		}

		content += "\n"
	}

	newFile := File{
		Name:      name,
		Title:     title,
		Tags:      tags,
		CreatedAt: now,
		DueAt:     nextFriday,
		Done:      false,
		Content:   content,
	}

	if err := onFileCreated(newFile); err != nil {
		return File{}, err
	}

	return newFile, nil
}
