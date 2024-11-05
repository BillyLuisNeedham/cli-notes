package scripts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type MetaData struct {
	Key   string
	Value string
}

func CreateTodo(title string) {

	meta := []MetaData{
		{
			Key:   "date-due",
			Value: "",
		},
		{
			Key: "done",
			Value: "false",
		},
	}

	tags := []string{"todo"}

	filePath, err := createNote(title, meta, tags)
	if err != nil {
		fmt.Println("Error creating note:", err)
		return
	}

	OpenNoteInEditor(filePath)
}

func CreateMeeting(title string) {

	meta := []MetaData{}

	tags := []string{"meeting"}

	filePath, err := createNote(title, meta, tags)
	if err != nil {
		fmt.Println("Error creating note:", err)
		return
	}

	OpenNoteInEditor(filePath)

}

func OpenNoteInEditor(filePath string) {
	err := exec.Command("cursor", filePath).Run()
	if err != nil {
		fmt.Println("Error opening file in editor:", err)
		return
	}
}

func createNote(title string, meta []MetaData, tags []string) (string, error) {
	date := time.Now().Format("2006-01-02")

	// Get the current directory path
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory path:", err)
		return "", err
	}

	// Create the file path in the parent directory
	fileName := fmt.Sprintf("%s-%s.md", title, date)
	notesPath := filepath.Join(currentDir, "/notes")
	filePath := filepath.Join(notesPath, fileName)

	// Create the Markdown file
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return "", err
	}
	defer file.Close()

	// Write the YAML Front Matter
	_, err = file.WriteString("---\n")
	if err != nil {
		fmt.Println("Error writing top of meta tags:", err)
		return "", err
	}
	_, err = fmt.Fprintf(file, "title: %s\n", title)
	if err != nil {
		fmt.Println("Error writing title:", err)
		return "", err
	}
	_, err = fmt.Fprintf(file, "date-created: %s\n", date)
	if err != nil {
		fmt.Println("Error writing date:", err)
		return "", err
	}
	_, err = fmt.Fprintf(file, "tags: %v\n", tags)
	if err != nil {
		fmt.Println("Error writing tags:", err)
		return "", err
	}
	for _, m := range meta {
		_, err = fmt.Fprintf(file, "%s: %s\n", m.Key, m.Value)
		if err != nil {
			fmt.Println("Error writing meta data:", err)
			return "", err
		}
	}

	_, err = file.WriteString("---\n\n")
	if err != nil {
		fmt.Println("Error writing bottom of meta tags:", err)
		return "", err
	}

	// Write the main content
	todoTitle := fmt.Sprintf("# %s\n\n", title)
	_, err = file.WriteString(todoTitle)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return "", err
	}

	fmt.Println("Markdown file created successfully at:", filePath)	
	return filePath, nil
}

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
