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

	createNote(title, meta, tags)
}

func CreateMeeting(title string) {

	meta := []MetaData{}

	tags := []string{"meeting"}

	createNote(title, meta, tags)

}

func OpenNoteInEditor(filePath string) {
	err := exec.Command("cursor", filePath).Run()
	if err != nil {
		fmt.Println("Error opening file in editor:", err)
		return
	}
}

func createNote(title string, meta []MetaData, tags []string, ) {
	date := time.Now().Format("2006-01-02")

	// Get the current directory path
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory path:", err)
		return
	}

	// Create the file path in the parent directory
	fileName := fmt.Sprintf("%s-%s.md", title, date)
	notesPath := filepath.Join(currentDir, "/notes")
	filePath := filepath.Join(notesPath, fileName)

	// Create the Markdown file
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Write the YAML Front Matter
	_, err = file.WriteString("---\n")
	if err != nil {
		fmt.Println("Error writing top of meta tags:", err)
		return
	}
	_, err = fmt.Fprintf(file, "title: %s\n", title)
	if err != nil {
		fmt.Println("Error writing title:", err)
		return
	}
	_, err = fmt.Fprintf(file, "date-created: %s\n", date)
	if err != nil {
		fmt.Println("Error writing date:", err)
		return
	}
	_, err = fmt.Fprintf(file, "tags: %v\n", tags)
	if err != nil {
		fmt.Println("Error writing tags:", err)
		return
	}
	for _, m := range meta {
		_, err = fmt.Fprintf(file, "%s: %s\n", m.Key, m.Value)
		if err != nil {
			fmt.Println("Error writing meta data:", err)
			return
		}
	}

	_, err = file.WriteString("---\n\n")
	if err != nil {
		fmt.Println("Error writing bottom of meta tags:", err)
		return
	}

	// Write the main content
	todoTitle := fmt.Sprintf("# %s\n\n", title)
	_, err = file.WriteString(todoTitle)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("Markdown file created successfully at:", filePath)	
	OpenNoteInEditor(filePath)
}
