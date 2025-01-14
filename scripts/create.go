package scripts

import (
	"fmt"
	"os"
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

// TODO refactor this into domain code
func CreateStandup() {
	title := "standup"
	meta := []MetaData{}
	tags := []string{"standup"}

	// Handle both return values
	filePath, err := createNote(title, meta, tags)
	if err != nil {
		return
	}

	// Append the standup template content
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file for appending:", err)
		return
	}
	defer file.Close()

	content := `## Me

### Todays plan

### Blockers

`
	_, err = file.WriteString(content)
	if err != nil {
		fmt.Println("Error writing standup content:", err)
		return
	}

	OpenNoteInEditor(filePath)
}
