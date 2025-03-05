package scripts

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type WriteFile = func(File) error

// Make the function a variable so it can be overridden in tests
var readLatestFileContent = func(file File) (File, error) {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return file, err
	}

	// Construct the file path
	filePath := filepath.Join(currentDir, "notes", file.Name)

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return file, err
	}

	// Open the file
	f, err := os.Open(filePath)
	if err != nil {
		return file, err
	}
	defer f.Close()

	// Read the file and extract content without frontmatter
	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	firstFrontmatterDelimiter := false
	var contentBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Check for frontmatter delimiters "---"
		if line == "---" {
			if !firstFrontmatterDelimiter {
				// First encountered frontmatter delimiter
				firstFrontmatterDelimiter = true
				inFrontmatter = true
				continue
			} else if inFrontmatter {
				// End of frontmatter
				inFrontmatter = false
				continue
			}
		}

		// Skip frontmatter lines but include everything else
		if !inFrontmatter {
			contentBuilder.WriteString(line)
			contentBuilder.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return file, err
	}

	// Create a new file struct with the updated content
	updatedFile := file
	updatedFile.Content = contentBuilder.String()

	return updatedFile, nil
}

func DelayDueDate(delayDays int, file File, writeFile WriteFile) error {
	// Read the latest content from the file to ensure we don't lose any updates
	updatedFile, err := readLatestFileContent(file)
	if err != nil {
		return err
	}

	// Update the due date on the file with the latest content
	today := time.Now()
	updatedFile.DueAt = today.AddDate(0, 0, delayDays)
	return writeFile(updatedFile)
}

func SetDueDateToToday(file File, writeFile WriteFile) error {
	// Read the latest content from the file to ensure we don't lose any updates
	updatedFile, err := readLatestFileContent(file)
	if err != nil {
		return err
	}

	updatedFile.DueAt = time.Now()
	return writeFile(updatedFile)
}
