package data

import (
	"cli-notes/scripts"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const DirectoryPath = "/notes"

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

func timeToString(time time.Time) string {
	return time.Format("2006-01-02")
}
