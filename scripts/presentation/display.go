package presentation

import (
	"cli-notes/scripts"
	"fmt"
)

func PrintAllFiles(files []scripts.File) {
	var currentPriority scripts.Priority
	var isFirstFile = true

	for _, file := range files {
		// Add line breaks between different priority groups
		if !isFirstFile && file.Priority != currentPriority {
			fmt.Println() // Add an empty line between priority groups
		}

		fmt.Printf("%v  due: %v", file.Name, file.DueAt.Format("2006-01-02"))

		// Print priority if available
		if file.Priority > 0 {
			fmt.Printf("  P%d", file.Priority)
		}

		fmt.Println()

		currentPriority = file.Priority
		isFirstFile = false
	}
}
