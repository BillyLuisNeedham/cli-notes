package main

import (
	"bufio"
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"cli-notes/scripts/presentation"
	"fmt"
	"os"

	"github.com/eiannone/keyboard"
)

func runTestMode(fileStore *data.SearchedFilesStore, onClose func()) {
	reader := bufio.NewReader(os.Stdin)
	command := presentation.WIPCommand{}

	fmt.Print("> ")
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			// EOF or error, exit gracefully
			onClose()
			break
		}

		var char rune
		var key keyboard.Key

		// Simple mapping for test mode
		switch r {
		case '\n':
			key = keyboard.KeyEnter
		case '\x7f': // Backspace
			key = keyboard.KeyBackspace
		case '\x1b': // Escape or start of sequence
			// Check if there are more bytes
			if reader.Buffered() > 0 {
				next, _, _ := reader.ReadRune()
				if next == '[' {
					// Arrow keys
					dir, _, _ := reader.ReadRune()
					switch dir {
					case 'A':
						key = keyboard.KeyArrowUp
					case 'B':
						key = keyboard.KeyArrowDown
					}
				} else {
					// Just Escape
					key = keyboard.KeyEsc
				}
			} else {
				key = keyboard.KeyEsc
			}
		default:
			char = r
		}

		// fmt.Printf("DEBUG: char=%q key=%v\n", char, key)

		nextCommand, err := presentation.CommandHandler(
			char,
			key,
			command,
			func() scripts.File { return searchRecentFilesPrintIfNotFound(fileStore.GetNextFile) },
			func() scripts.File { return searchRecentFilesPrintIfNotFound(fileStore.GetPreviousFile) },
			func(file scripts.File) ([]string, error) {
				files := []scripts.File{file}
				return scripts.GetUncompletedTasksInFiles(files)
			},
			func() { fmt.Print("\b \b") },
		)

		if err != nil {
			fmt.Printf("Error processing command: %v\n", err)
			fmt.Print("> ")
			command = presentation.WIPCommand{}
			continue
		}

		switch nextCommand := nextCommand.(type) {
		case presentation.WIPCommand:
			command = nextCommand
			fmt.Print(string(char))

		case presentation.BackSpacedWIPCommand:
			command = nextCommand.WIPCommand
			fmt.Print("\b \b")

		case presentation.SpacedWIPCommand:
			command = nextCommand.WIPCommand
			fmt.Print(" ")

		case presentation.FileSelectedWIPCommand:
			command = nextCommand.WIPCommand
			// fmt.Printf("DEBUG: File selected: %s\n", command.SelectedFile.Name)

			fmt.Println(command.SelectedFile.Name)

			for _, task := range nextCommand.Tasks {
				fmt.Printf("%v", task)
			}
			fmt.Println("")

		case presentation.CompletedCommand:
			completedCommand := nextCommand
			// fmt.Printf("DEBUG: Command completed: %s args=%v file=%s\n", completedCommand.Name, completedCommand.Queries, completedCommand.SelectedFile.Name)

			if completedCommand.SelectedFile.Name != "" && completedCommand.Name == "" {
				finalCommand := presentation.CompletedCommand{
					Name:         "o",
					Queries:      completedCommand.Queries,
					SelectedFile: completedCommand.SelectedFile,
				}
				completedCommand = finalCommand

			} else {
				fmt.Println("")
			}

			handleCommand(completedCommand, onClose, fileStore, reader)
			fmt.Print("> ")
			command = presentation.WIPCommand{}

		case presentation.ResetCommand:
			command = presentation.WIPCommand{}
			fmt.Println("")
			fmt.Println("Screw that line")
			fmt.Print("> ")
		}
	}
}
