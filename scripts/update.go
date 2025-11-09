package scripts

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

	// Read the file and extract content and metadata
	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	firstFrontmatterDelimiter := false
	var contentBuilder strings.Builder
	
	// Create a new file struct with original properties but with updated content and metadata
	updatedFile := file
	
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

		// Extract metadata from frontmatter
		if inFrontmatter {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "priority":
					priority, err := strconv.Atoi(value)
					if err == nil && priority >= 1 && priority <= 3 {
						updatedFile.Priority = Priority(priority)
					}
				case "objective-role":
					updatedFile.ObjectiveRole = value
				case "objective-id":
					updatedFile.ObjectiveID = value
				}
			}
		} else {
			// Skip frontmatter lines but include everything else for content
			contentBuilder.WriteString(line)
			contentBuilder.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return file, err
	}

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

// ChangePriority updates the priority of a file to the specified priority level
// Priority must be P1 (1), P2 (2), or P3 (3)
func ChangePriority(newPriority Priority, file File, writeFile WriteFile) error {
	// Validate priority is within range
	if newPriority < P1 || newPriority > P3 {
		return fmt.Errorf("invalid priority: must be 1, 2, or 3")
	}

	// Read the latest content from the file to ensure we don't lose any updates
	updatedFile, err := readLatestFileContent(file)
	if err != nil {
		return err
	}

	// Update the priority
	updatedFile.Priority = newPriority

	// Write the updated file
	return writeFile(updatedFile)
}

// RenameFile renames a file by extracting the date suffix, creating a new filename,
// updating the title in metadata and content, and renaming the file on disk
func RenameFile(newTitle string, file File, writeFile WriteFile) (File, error) {
	// Extract date suffix from current filename
	// Example: "test-2-2025-07-29.md" -> "2025-07-29"
	fileName := file.Name
	// Remove .md extension
	nameWithoutExt := strings.TrimSuffix(fileName, ".md")

	// Find the date suffix (last 10 characters should be YYYY-MM-DD format)
	if len(nameWithoutExt) < 10 {
		return File{}, fmt.Errorf("filename %s does not have a valid date suffix", fileName)
	}

	dateSuffix := nameWithoutExt[len(nameWithoutExt)-10:]
	// Validate date format
	_, err := time.Parse("2006-01-02", dateSuffix)
	if err != nil {
		return File{}, fmt.Errorf("filename %s does not have a valid date suffix: %v", fileName, err)
	}

	// Create new filename
	newFileName := fmt.Sprintf("%s-%s.md", newTitle, dateSuffix)

	// Read the latest content from the file
	updatedFile, err := readLatestFileContent(file)
	if err != nil {
		return File{}, err
	}

	// Update the title
	updatedFile.Title = newTitle

	// Update the first heading in content if it exists
	// Look for "# old-title" and replace with "# new-title"
	lines := strings.Split(updatedFile.Content, "\n")
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "# ") {
			// Replace the first heading with the new title
			lines[i] = fmt.Sprintf("# %s", newTitle)
			break
		}
	}
	updatedFile.Content = strings.Join(lines, "\n")

	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return File{}, err
	}

	// Construct file paths
	oldPath := filepath.Join(currentDir, "notes", fileName)
	newPath := filepath.Join(currentDir, "notes", newFileName)

	// Update the filename in the struct
	updatedFile.Name = newFileName

	// Write the updated content to the new file
	if err := writeFile(updatedFile); err != nil {
		return File{}, err
	}

	// Remove the old file if it's different from the new one
	if oldPath != newPath {
		if err := os.Remove(oldPath); err != nil {
			return File{}, fmt.Errorf("failed to remove old file: %v", err)
		}
	}

	return updatedFile, nil
}
