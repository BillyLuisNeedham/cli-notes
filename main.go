package main

import (
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"cli-notes/scripts/presentation"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
)

var keyboardOpen bool

func closeKeyboard() {
	if keyboardOpen {
		keyboard.Close()
		keyboardOpen = false
	}
}

func reopenKeyboard() error {
	if !keyboardOpen {
		err := keyboard.Open()
		if err != nil {
			return err
		}
		keyboardOpen = true
	}
	return nil
}

func main() {
	closeChannel := make(chan bool)
	var searchedFilesStore = data.NewSearchedFilesStore()

	// restore if using git for backup
	// go scripts.MonitorDirectorySize("./notes", func() {
	// 	scripts.PushChangesToGit("./notes")
	// })

	go setupCommandScanner(searchedFilesStore, func() {
		closeChannel <- true
	})

	<-closeChannel
	fmt.Println("Exiting...")
}

func setupCommandScanner(fileStore *data.SearchedFilesStore, onClose func()) {
	err := reopenKeyboard()
	if err != nil {
		panic(err)
	}
	defer closeKeyboard()

	command := presentation.WIPCommand{}

	fmt.Print("> ")
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}

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

			fmt.Println(command.SelectedFile.Name)

			for _, task := range nextCommand.Tasks {
				fmt.Printf("%v", task)
			}
			fmt.Println("")

		case presentation.CompletedCommand:
			completedCommand := nextCommand

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

			handleCommand(completedCommand, onClose, fileStore)
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

func handleCommand(command presentation.CompletedCommand, onClose func(), fileStore *data.SearchedFilesStore) {

	switch command.Name {
	case "gt":
		if len(command.Queries) == 0 {
			files, err := scripts.GetTodos(data.QueryFilesByDone)
			if err != nil {
				fmt.Printf("Error getting todos: %v\n", err)
			}
			onFilesFetched(files, fileStore)
		} else {
			files, err := scripts.QueryOpenTodos(command.Queries, data.QueryFilesByDone)
			if err != nil {
				fmt.Printf("Error querying open todos: %v\n", err)
				return
			}
			onFilesFetched(files, fileStore)
		}

	case "p1":
		files, err := scripts.GetTodosByPriority(scripts.P1, data.QueryFilesByDone)
		if err != nil {
			fmt.Printf("Error getting P1 todos: %v\n", err)
			return
		}
		onFilesFetched(files, fileStore)

	case "p2":
		files, err := scripts.GetTodosByPriority(scripts.P2, data.QueryFilesByDone)
		if err != nil {
			fmt.Printf("Error getting P2 todos: %v\n", err)
			return
		}
		onFilesFetched(files, fileStore)

	case "p3":
		files, err := scripts.GetTodosByPriority(scripts.P3, data.QueryFilesByDone)
		if err != nil {
			fmt.Printf("Error getting P3 todos: %v\n", err)
			return
		}
		onFilesFetched(files, fileStore)

	case "gta":
		if len(command.Queries) == 0 {
			fmt.Println("Please provide a tags to query")
			return
		}
		files, err := scripts.SearchNotesByTags(command.Queries, data.QueryNotesByTags)
		if err != nil {
			fmt.Printf("Error getting notes by tags: %v\n", err)
		}
		onFilesFetched(files, fileStore)

	case "gq":
		if len(command.Queries) == 0 {
			fmt.Println("Please provide a query to search")
			return
		}
		files, err := scripts.QueryAllFiles(command.Queries, data.QueryFiles)
		if err != nil {
			fmt.Printf("Error querying notes: %v", err)
		}
		onFilesFetched(files, fileStore)

	case "gqa":
		if len(command.Queries) == 0 {
			fmt.Println("Please provide a query to search")
			return
		}
		previousFiles := fileStore.GetFilesSearched()
		if len(previousFiles) == 0 {
			fmt.Println("No files have been queried")
		} else {
			files := scripts.QueryFiles(command.Queries, previousFiles)
			onFilesFetched(files, fileStore)
		}

	case "gat":
		previousFiles := fileStore.GetFilesSearched()
		if len(previousFiles) == 0 {
			fmt.Println("No files have been queried")
		} else {
			tasks, err := scripts.GetUncompletedTasksInFiles(previousFiles)
			if err != nil {
				fmt.Printf("Error getting tasks: %v", err)
				return
			}
			for _, task := range tasks {
				fmt.Printf("%v\n\n", task)
			}
		}

	case "ct":
		handleCreateFile("todo", command.Queries, scripts.CreateTodo)

	case "cm":
		handleCreateFile("meeting", command.Queries, scripts.CreateMeeting)

	case "cp":
		handleCreateFile("plan", command.Queries, scripts.CreateSevenQuestions)

	case "o":
		if command.SelectedFile.Name != "" {
			openNoteInEditor(command.SelectedFile.Name)
		} else {
			fmt.Println("Please provide a file name to open")
		}

	case "exit", "quit", "q":
		onClose()
		return

	case "cs":
		file, err := scripts.CreateStandup(data.GetTeamNames, data.WriteFile)
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			return
		}
		openNoteInEditor(file.Name)

	case "gto":
		files, err := scripts.GetOverdueTodos(func(dateQuery scripts.DateQuery) ([]scripts.File, error) {
			return data.QueryTodosWithDateCriteria(dateQuery)
		})
		if err != nil {
			fmt.Printf("Error getting overdue todos: %v\n", err)
			return
		}
		onFilesFetched(files, fileStore)

	case "gtnd":
		files, err := scripts.GetTodosWithNoDueDate(func(dateQuery scripts.DateQuery) ([]scripts.File, error) {
			return data.QueryTodosWithDateCriteria(dateQuery)
		})
		if err != nil {
			fmt.Printf("Error getting todos with no due date: %v\n", err)
			return
		}
		onFilesFetched(files, fileStore)

	case "gts":
		files, err := scripts.GetSoonTodos(func(dateQuery scripts.DateQuery) ([]scripts.File, error) {
			return data.QueryTodosWithDateCriteria(dateQuery)
		})
		if err != nil {
			fmt.Printf("Error getting soon todos: %v\n", err)
			return
		}
		onFilesFetched(files, fileStore)

	case "d":
		if command.SelectedFile.Name == "" {
			fmt.Println("No file selected")
			return
		}
		if len(command.Queries) < 1 {
			fmt.Println("Please provide an amount of days to delay")
			return
		}

		delayDays, err := strconv.Atoi(command.Queries[0])
		if err != nil {
			fmt.Printf("Error converting days: %v\n", err)
			return
		}

		err = scripts.DelayDueDate(delayDays, command.SelectedFile, data.WriteFile)
		if err != nil {
			fmt.Printf("Error delaying note: %v\n", err)
			return
		}

		fmt.Printf("%v delayed by %v days\n", command.SelectedFile.Name, delayDays)

	case "t":
		if command.SelectedFile.Name == "" {
			fmt.Println("No file selected")
			return
		}
		err := scripts.SetDueDateToToday(command.SelectedFile, data.WriteFile)
		if err != nil {
			fmt.Printf("Error setting note to today: %v", err)
			return
		}

		fmt.Printf("%v due date set to today\n", command.SelectedFile.Name)

	case "m":
		if command.SelectedFile.Name == "" {
			fmt.Println("No file selected")
			return
		}
		err := scripts.SetDueDateToNextDay(time.Monday, command.SelectedFile, data.WriteFile)
		if err != nil {
			fmt.Printf("Error setting due date to next Monday: %v\n", err)
			return
		}
		fmt.Printf("%v due date set to next Monday\n", command.SelectedFile.Name)

	case "tu":
		if command.SelectedFile.Name == "" {
			fmt.Println("No file selected")
			return
		}
		err := scripts.SetDueDateToNextDay(time.Tuesday, command.SelectedFile, data.WriteFile)
		if err != nil {
			fmt.Printf("Error setting due date to next Tuesday: %v\n", err)
			return
		}
		fmt.Printf("%v due date set to next Tuesday\n", command.SelectedFile.Name)

	case "w":
		if command.SelectedFile.Name == "" {
			fmt.Println("No file selected")
			return
		}
		err := scripts.SetDueDateToNextDay(time.Wednesday, command.SelectedFile, data.WriteFile)
		if err != nil {
			fmt.Printf("Error setting due date to next Wednesday: %v\n", err)
			return
		}
		fmt.Printf("%v due date set to next Wednesday\n", command.SelectedFile.Name)

	case "th":
		if command.SelectedFile.Name == "" {
			fmt.Println("No file selected")
			return
		}
		err := scripts.SetDueDateToNextDay(time.Thursday, command.SelectedFile, data.WriteFile)
		if err != nil {
			fmt.Printf("Error setting due date to next Thursday: %v\n", err)
			return
		}
		fmt.Printf("%v due date set to next Thursday\n", command.SelectedFile.Name)

	case "f":
		if command.SelectedFile.Name == "" {
			fmt.Println("No file selected")
			return
		}
		err := scripts.SetDueDateToNextDay(time.Friday, command.SelectedFile, data.WriteFile)
		if err != nil {
			fmt.Printf("Error setting due date to next Friday: %v\n", err)
			return
		}
		fmt.Printf("%v due date set to next Friday\n", command.SelectedFile.Name)

	case "sa":
		if command.SelectedFile.Name == "" {
			fmt.Println("No file selected")
			return
		}
		err := scripts.SetDueDateToNextDay(time.Saturday, command.SelectedFile, data.WriteFile)
		if err != nil {
			fmt.Printf("Error setting due date to next Saturday: %v\n", err)
			return
		}
		fmt.Printf("%v due date set to next Saturday\n", command.SelectedFile.Name)

	case "su":
		if command.SelectedFile.Name == "" {
			fmt.Println("No file selected")
			return
		}
		err := scripts.SetDueDateToNextDay(time.Sunday, command.SelectedFile, data.WriteFile)
		if err != nil {
			fmt.Printf("Error setting due date to next Sunday: %v\n", err)
			return
		}
		fmt.Printf("%v due date set to next Sunday\n", command.SelectedFile.Name)

	case "p":
		if command.SelectedFile.Name == "" {
			fmt.Println("No file selected")
			return
		}
		if len(command.Queries) < 1 {
			fmt.Println("Please provide a priority (1, 2, or 3)")
			return
		}

		priorityNum, err := strconv.Atoi(command.Queries[0])
		if err != nil {
			fmt.Printf("Error converting priority: %v\n", err)
			return
		}

		if priorityNum < 1 || priorityNum > 3 {
			fmt.Println("Priority must be 1, 2, or 3")
			return
		}

		newPriority := scripts.Priority(priorityNum)
		err = scripts.ChangePriority(newPriority, command.SelectedFile, data.WriteFile)
		if err != nil {
			fmt.Printf("Error changing priority: %v\n", err)
			return
		}

		fmt.Printf("%v priority changed to P%d\n", command.SelectedFile.Name, priorityNum)

	case "r":
		if command.SelectedFile.Name == "" {
			fmt.Println("No file selected")
			return
		}
		if len(command.Queries) < 1 || command.Queries[0] == "" {
			fmt.Println("Please provide a new title for the file")
			return
		}

		newTitle := strings.Join(command.Queries, "-")
		renamedFile, err := scripts.RenameFile(newTitle, command.SelectedFile, data.WriteFile)
		if err != nil {
			fmt.Printf("Error renaming file: %v\n", err)
			return
		}

		// Update the file store with the renamed file
		previousFiles := fileStore.GetFilesSearched()
		for i, file := range previousFiles {
			if file.Name == command.SelectedFile.Name {
				previousFiles[i] = renamedFile
				break
			}
		}
		fileStore.SetFilesSearched(previousFiles)

		fmt.Printf("Renamed %v to %v\n", command.SelectedFile.Name, renamedFile.Name)

	case "gd":
		if len(command.Queries) != 2 {
			fmt.Println("Please provide exactly two dates in the format YYYY-MM-DD")
			return
		}

		startDate := command.Queries[0]
		endDate := command.Queries[1]

		// Validate dates
		if !isValidDate(startDate) || !isValidDate(endDate) {
			fmt.Println("Invalid date format. Please use YYYY-MM-DD")
			return
		}

		// Get all completed todos within the date range
		files, err := scripts.GetCompletedTodosByDateRange(startDate, endDate, func(dateQuery scripts.DateQuery) ([]scripts.File, error) {
			return data.QueryCompletedTodosByDateRange(dateQuery)
		})

		if err != nil {
			fmt.Printf("Error getting completed todos in date range: %v\n", err)
			return
		}

		if len(files) == 0 {
			fmt.Printf("No completed todos found between %s and %s\n", startDate, endDate)
			return
		}

		// Create a combined note
		newFile, err := scripts.CreateDateRangeQueryNote(startDate, endDate, files, data.WriteFile)
		if err != nil {
			fmt.Printf("Error creating date range query note: %v\n", err)
			return
		}

		fmt.Printf("Created date range query note: %s\n", newFile.Name)
		openNoteInEditor(newFile.Name)

	case "wp", "week":
		err := runWeekPlanner()
		if err != nil {
			fmt.Printf("Error running week planner: %v\n", err)
			return
		}

	default:
		fmt.Println("Unknown command.")
		return
	}
}

func onFilesFetched(files []scripts.File, fileStore *data.SearchedFilesStore) {
	fileStore.SetFilesSearched(files)
	presentation.PrintAllFiles(files)
}

func searchRecentFilesPrintIfNotFound(search func() *scripts.File) scripts.File {
	file := search()
	if file == nil {
		fmt.Println("No files have been searched yet.")
		fmt.Println("")
		return scripts.File{}
	}

	return *file
}

func openNoteInEditor(fileName string) {
	filePath := "notes/" + fileName
	err := presentation.OpenNoteInEditor(filePath, closeKeyboard, func() {
		if err := reopenKeyboard(); err != nil {
			fmt.Printf("Error reopening keyboard: %v\n", err)
		}
	})
	if err != nil {
		fmt.Printf("Error opening note in editor: %v", err)
	}
}

func handleCreateFile(fileType string, queries []string, createFn func(string, scripts.OnFileCreated) (scripts.File, error)) {
	if len(queries) < 1 {
		fmt.Printf("Please provide a title for the new %s\n", fileType)
		return
	}
	title := strings.Join(queries, "-")
	file, err := createFn(title, data.WriteFile)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}
	openNoteInEditor(file.Name)
}

func isValidDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

func runWeekPlanner() error {
	// Initialize week planner state
	state, err := data.NewWeekPlannerState()
	if err != nil {
		return fmt.Errorf("error initializing week planner: %w", err)
	}

	// Command buffer for multi-character commands (tu, th, sa, su)
	commandBuffer := ""
	lastMessage := ""

	// Main week planner event loop
	for {
		// Render the UI
		display := presentation.RenderWeekView(state)
		fmt.Print(display)

		// Display last message if any
		if lastMessage != "" {
			fmt.Printf("\n%s\n", lastMessage)
			lastMessage = ""
		}

		// Get keyboard input
		char, key, err := keyboard.GetKey()
		if err != nil {
			return fmt.Errorf("error reading keyboard input: %w", err)
		}

		// Try to build multi-character command
		if char >= 'a' && char <= 'z' && key == 0 {
			commandBuffer += string(char)

			// Check if we have a complete multi-char command
			if day, ok := presentation.ParseMultiCharCommand(commandBuffer); ok {
				input := presentation.WeekPlannerInput{
					Action: presentation.SwitchDay,
					Day:    day,
				}
				shouldExit, message, err := presentation.HandleWeekPlannerInput(state, input)
				if err != nil {
					return err
				}
				lastMessage = message
				commandBuffer = ""

				if shouldExit {
					break
				}
				continue
			}

			// Check if buffer is getting too long (invalid command)
			if len(commandBuffer) > 2 {
				commandBuffer = ""
			}

			// For single-char commands, process immediately
			if len(commandBuffer) == 1 {
				input := presentation.ParseWeekPlannerInput(rune(commandBuffer[0]), key)
				if input.Action != presentation.NoAction {
					shouldExit, message, err := presentation.HandleWeekPlannerInput(state, input)
					if err != nil {
						return err
					}

					// Handle quit with save prompt
					if shouldExit {
						if state.Plan.HasChanges() {
							if !promptSaveChanges(state) {
								break
							}
							// User cancelled, continue the loop
							commandBuffer = ""
							continue
						}
						break
					}

					lastMessage = message
					commandBuffer = ""
					continue
				}
			}

			// Wait for more chars to complete the command
			continue
		}

		// Process non-character input (arrow keys, etc.)
		commandBuffer = "" // Reset buffer for non-char input
		input := presentation.ParseWeekPlannerInput(char, key)

		if input.Action == presentation.NoAction {
			continue
		}

		// Handle opening a note (special case - needs keyboard management)
		if input.Action == presentation.OpenTodo {
			selectedTodo := state.GetSelectedTodo()
			if selectedTodo == nil {
				lastMessage = "No todo selected"
				continue
			}

			// Open the note in editor
			openNoteInEditor(selectedTodo.Name)

			// Reload the week plan from disk to pick up any changes made in the editor
			err := state.Reset()
			if err != nil {
				lastMessage = fmt.Sprintf("Error reloading week plan: %v", err)
			} else {
				lastMessage = "Week plan reloaded"
			}
			continue
		}

		shouldExit, message, err := presentation.HandleWeekPlannerInput(state, input)
		if err != nil {
			return err
		}

		// Handle quit with save prompt
		if shouldExit {
			if state.Plan.HasChanges() {
				if !promptSaveChanges(state) {
					break
				}
				// User cancelled, continue the loop
				continue
			}
			break
		}

		lastMessage = message
	}

	// Clear screen and return to main prompt
	fmt.Print("\033[2J\033[H")
	return nil
}

// promptSaveChanges prompts the user to save changes before exiting
// Returns false if user wants to exit, true if user cancels exit
func promptSaveChanges(state *data.WeekPlannerState) bool {
	fmt.Printf("\nYou have %d unsaved changes. Save before exiting? (y/n/c): ", len(state.Plan.Changes))

	for {
		char, _, err := keyboard.GetKey()
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return false
		}

		switch char {
		case 'y', 'Y':
			fmt.Println("y")
			err := state.Save()
			if err != nil {
				fmt.Printf("Error saving changes: %v\n", err)
				fmt.Println("Press any key to continue...")
				keyboard.GetKey()
				return true // Return to planner to try again
			}
			fmt.Println("Changes saved successfully!")
			return false // Exit

		case 'n', 'N':
			fmt.Println("n")
			fmt.Println("Changes discarded.")
			return false // Exit without saving

		case 'c', 'C':
			fmt.Println("c")
			fmt.Println("Cancelled. Returning to week planner...")
			time.Sleep(500 * time.Millisecond)
			return true // Return to planner

		default:
			// Invalid input, keep prompting
			continue
		}
	}
}
