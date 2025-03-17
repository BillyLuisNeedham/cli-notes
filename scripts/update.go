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

// timeNow is a variable that returns the current time
// Can be overridden in tests
var timeNow = func() time.Time {
	return time.Now()
}

func DelayDueDate(delayDays int, file File, writeFile WriteFile) error {
	// Read the latest content from the file to ensure we don't lose any updates
	updatedFile, err := readLatestFileContent(file)
	if err != nil {
		return err
	}

	// Update the due date on the file with the latest content
	today := timeNow()
	updatedFile.DueAt = today.AddDate(0, 0, delayDays)
	
	// Ensure priority is preserved from the original file if it exists
	if file.Priority > 0 {
		updatedFile.Priority = file.Priority
	}
	
	return writeFile(updatedFile)
}

func SetDueDateToToday(file File, writeFile WriteFile) error {
	// Read the latest content from the file to ensure we don't lose any updates
	updatedFile, err := readLatestFileContent(file)
	if err != nil {
		return err
	}

	updatedFile.DueAt = timeNow()
	
	// Ensure priority is preserved from the original file if it exists
	if file.Priority > 0 {
		updatedFile.Priority = file.Priority
	}
	
	return writeFile(updatedFile)
}

// SetDueDateToNextDay sets the due date of a file to the next occurrence of the specified day of the week
func SetDueDateToNextDay(dayOfWeek time.Weekday, file File, writeFile WriteFile) error {
	// Read the latest content from the file to ensure we don't lose any updates
	updatedFile, err := readLatestFileContent(file)
	if err != nil {
		return err
	}

	// Get the current date and time
	now := timeNow()

	// Calculate days until the next occurrence of the specified day
	daysUntil := int(dayOfWeek - now.Weekday())
	if daysUntil <= 0 {
		// If the day has already passed this week or is today, get next week's occurrence
		daysUntil += 7
	}

	// Set the due date to the next occurrence of the specified day at the same time
	updatedFile.DueAt = now.AddDate(0, 0, daysUntil)

	// Ensure priority is preserved from the original file if it exists
	if file.Priority > 0 {
		updatedFile.Priority = file.Priority
	}
	
	return writeFile(updatedFile)
}
