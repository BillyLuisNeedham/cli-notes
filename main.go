package main

import (
	"bufio"
	"cli-notes/input"
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"cli-notes/scripts/presentation"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"golang.org/x/term"
)

var keyboardOpen bool

func closeKeyboard() {
	if os.Getenv("CLI_NOTES_TEST_MODE") == "true" {
		return
	}
	if keyboardOpen {
		keyboard.Close()
		keyboardOpen = false
	}
}

func reopenKeyboard() error {
	if os.Getenv("CLI_NOTES_TEST_MODE") == "true" {
		return nil
	}
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
	if os.Getenv("CLI_NOTES_TEST_MODE") == "true" {
		runTestMode(fileStore, onClose)
		return
	}

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
			data.QueryNonFinishedObjectives,
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

		case presentation.TabPressedWIPCommand:
			command = nextCommand.WIPCommand
			// Clear line and rewrite with autocompleted text
			fmt.Print("\r\033[K> " + command.Text)

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

			handleCommand(completedCommand, onClose, fileStore, nil)
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

func handleCommand(command presentation.CompletedCommand, onClose func(), fileStore *data.SearchedFilesStore, testModeReader *bufio.Reader) {

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
		var reader input.InputReader
		if testModeReader != nil {
			reader = input.NewStdinReader(testModeReader)
		} else {
			reader = &input.KeyboardReader{}
		}
		err := runWeekPlanner(reader)
		if err != nil {
			fmt.Printf("Error running week planner: %v\n", err)
			return
		}

	case "ob":
		// If there's a selected file and queries, link the note to an objective
		if command.SelectedFile.Name != "" && len(command.Queries) > 0 && command.Queries[0] != "" {
			objectiveTitle := command.Queries[0]

			// Find the objective by title
			objectives, err := data.QueryNonFinishedObjectives()
			if err != nil {
				fmt.Printf("Error querying objectives: %v\n", err)
				return
			}

			var targetObjective *scripts.File
			for _, obj := range objectives {
				if strings.EqualFold(obj.Title, objectiveTitle) {
					targetObjective = &obj
					break
				}
			}

			if targetObjective == nil {
				fmt.Printf("Objective not found: %s\n", objectiveTitle)
				return
			}

			// Check if already linked to an objective
			if command.SelectedFile.ObjectiveID != "" {
				existingParent, err := data.GetObjectiveByID(command.SelectedFile.ObjectiveID)
				if err == nil && existingParent != nil {
					fmt.Printf("WARNING: \"%s\" is currently linked to objective \"%s\"\n",
						command.SelectedFile.Title, existingParent.Title)
					fmt.Print("Re-link to \"" + targetObjective.Title + "\"? (y/n): ")

					char, _, err := keyboard.GetKey()
					if err != nil || (char != 'y' && char != 'Y') {
						fmt.Println("\nCancelled.")
						return
					}
					fmt.Println()
				}
			}

			// Link the note to the objective
			err = scripts.LinkTodoToObjective(command.SelectedFile, *targetObjective, data.WriteFile)
			if err != nil {
				fmt.Printf("Error linking to objective: %v\n", err)
				return
			}

			fmt.Printf("Linked \"%s\" to objective \"%s\"\n", command.SelectedFile.Title, targetObjective.Title)
		} else {
			// No selected file or queries - open objectives view
			var reader input.InputReader
			if testModeReader != nil {
				reader = input.NewStdinReader(testModeReader)
			} else {
				reader = &input.KeyboardReader{}
			}

			err := runObjectivesView(reader)
			if err != nil {
				fmt.Printf("Error running objectives view: %v\n", err)
				return
			}
		}

	case "tt":
		filterPerson := ""
		if len(command.Queries) > 0 {
			filterPerson = command.Queries[0]
		}

		var reader input.InputReader
		if testModeReader != nil {
			reader = input.NewStdinReader(testModeReader)
		} else {
			reader = &input.KeyboardReader{}
		}

		err := runTalkToView(filterPerson, reader)
		if err != nil {
			fmt.Printf("Error running talk-to view: %v\n", err)
			return
		}

	case "cpo":
		if command.SelectedFile.Name == "" {
			fmt.Println("No file selected")
			return
		}

		// Check if already a child - show warning
		if command.SelectedFile.ObjectiveID != "" {
			parent, err := data.GetObjectiveByID(command.SelectedFile.ObjectiveID)
			if err == nil && parent != nil {
				fmt.Printf("WARNING: \"%s\" is currently linked to:\n", command.SelectedFile.Title)
				fmt.Printf("  Parent Objective: \"%s\" (%s)\n\n", parent.Title, parent.ObjectiveID)
				fmt.Println("Converting to parent objective will:")
				fmt.Println("  • Unlink from current objective")
				fmt.Println("  • Create new objective ID")
				fmt.Println("  • Become independent parent objective")
				fmt.Print("Continue? (y/n): ")

				char, _, err := keyboard.GetKey()
				if err != nil || (char != 'y' && char != 'Y') {
					fmt.Println("\nCancelled.")
					return
				}
				fmt.Println()
			}
		}

		newObjective, err := scripts.ConvertTodoToParentObjective(command.SelectedFile, data.WriteFile)
		if err != nil {
			fmt.Printf("Error converting to parent objective: %v\n", err)
			return
		}

		fmt.Printf("Converted to parent objective.\n")
		fmt.Printf("Objective ID: %s\n", newObjective.ObjectiveID)

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

func runWeekPlanner(reader input.InputReader) error {
	// Ensure terminal is cleaned up on all exit paths
	defer func() {
		// Clear screen and reset cursor
		fmt.Print("\033[2J\033[H")
	}()

	// Get terminal size
	termWidth, termHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Fallback to default size if unable to get terminal size
		termWidth = 100
		termHeight = 30
	}

	// Check minimum terminal size
	const minWidth = 80
	const minHeight = 24
	if termWidth < minWidth || termHeight < minHeight {
		return fmt.Errorf("terminal too small. Minimum size: %dx%d (current: %dx%d)",
			minWidth, minHeight, termWidth, termHeight)
	}

	// Initialize week planner state
	state, err := data.NewWeekPlannerState()
	if err != nil {
		return fmt.Errorf("error initializing week planner: %w", err)
	}

	lastMessage := ""

	// Main week planner event loop
	for {
		// Render the UI
		display := presentation.RenderWeekView(state, termWidth, termHeight)
		fmt.Print(display)

		// Display last message if any
		if lastMessage != "" {
			fmt.Printf("\n%s\n", lastMessage)
			lastMessage = ""
		}

		// Get keyboard input
		char, key, err := reader.GetKey()
		if err != nil {
			return fmt.Errorf("error reading keyboard input: %w", err)
		}

		// Check for capital letter switch-day commands
		if presentation.IsSwitchDayKey(char) {
			targetDay, ok := presentation.ParseSwitchToDay(char)
			if ok {
				input := presentation.WeekPlannerInput{
					Action: presentation.SwitchDay,
					Day:    targetDay,
				}
				shouldExit, message, err := presentation.HandleWeekPlannerInput(state, input)
				if err != nil {
					return err
				}
				lastMessage = message

				if shouldExit {
					break
				}
				continue
			}
		}

		// Parse all other input (including lowercase move-to-day commands, special keys, etc.)
		input := presentation.ParseWeekPlannerInput(char, key)

		if input.Action == presentation.NoAction {
			continue
		}

		// Handle bulk move earlier
		if input.Action == presentation.BulkMoveEarlier {
			// Only works in normal view
			if state.ViewMode != data.NormalView {
				lastMessage = "Bulk move only available in normal view"
				continue
			}

			// Execute bulk move
			movedCount, err := state.BulkMoveEarlierTodosToCurrentDay()
			if err != nil {
				lastMessage = fmt.Sprintf("Error: %v", err)
			} else {
				dayName := data.WeekDayNames[state.SelectedDay]
				lastMessage = fmt.Sprintf("Moved %d todos to %s", movedCount, dayName)
			}
			continue
		}

		// Handle reset with confirmation (special case - needs confirmation)
		if input.Action == presentation.Reset {
			if promptResetConfirmation(state, reader) {
				err := state.Reset()
				if err != nil {
					lastMessage = fmt.Sprintf("Error resetting: %v", err)
				} else {
					lastMessage = "Plan reset from disk"
				}
			} else {
				lastMessage = "Reset cancelled"
			}
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

			// Refresh only the opened note from disk to pick up any edits
			// This preserves all unsaved moves in the weekly planner
			err := state.RefreshOpenedTodo(selectedTodo.Name)
			if err != nil {
				lastMessage = fmt.Sprintf("Error refreshing note: %v", err)
			} else {
				lastMessage = "Note refreshed"
			}
			continue
		}

		// Handle create todo (special case - needs title prompt and editor opening)
		if input.Action == presentation.CreateTodo {
			// Validate day (don't allow Earlier)
			if state.SelectedDay == data.Earlier {
				lastMessage = "Cannot create todos for earlier dates"
				continue
			}

			// Prompt for title
			dayName := data.WeekDayNames[state.SelectedDay]
			dueDate := state.Plan.GetDateForWeekDay(state.SelectedDay)
			title, err := promptForTodoTitle(dayName, dueDate, reader)

			if err != nil {
				lastMessage = fmt.Sprintf("Error: %v", err)
				continue
			}

			if title == "" {
				lastMessage = "Creation cancelled"
				continue
			}

			// Create the todo with custom due date
			file, err := scripts.CreateTodoWithDueDate(title, dueDate, data.WriteFile)
			if err != nil {
				lastMessage = fmt.Sprintf("Error creating todo: %v", err)
				continue
			}

			// Open the note in editor
			openNoteInEditor(file.Name)

			// Add to plan (in case it wasn't reloaded)
			todos := state.Plan.TodosByDay[state.SelectedDay]
			alreadyExists := false
			for _, todo := range todos {
				if todo.Name == file.Name {
					alreadyExists = true
					break
				}
			}

			if !alreadyExists {
				state.Plan.TodosByDay[state.SelectedDay] = append(
					state.Plan.TodosByDay[state.SelectedDay],
					file,
				)

				// Sort by priority
				data.SortTodosByPriority(state.Plan.TodosByDay[state.SelectedDay])
			}

			// Refresh the note from disk to pick up any edits
			err = state.RefreshOpenedTodo(file.Name)
			if err != nil {
				lastMessage = fmt.Sprintf("Error refreshing note: %v", err)
				continue
			}

			// Find and select the new todo
			todos = state.Plan.TodosByDay[state.SelectedDay]
			for i, todo := range todos {
				if todo.Name == file.Name {
					state.SelectedTodo = i
					break
				}
			}

			lastMessage = fmt.Sprintf("Created: %s", title)
			continue
		}

		shouldExit, message, err := presentation.HandleWeekPlannerInput(state, input)
		if err != nil {
			return err
		}

		// Handle quit with save prompt
		if shouldExit {
			if state.Plan.HasChanges() {
				if !promptSaveChanges(state, reader) {
					break
				}
				// User cancelled, continue the loop
				continue
			}
			break
		}

		lastMessage = message
	}

	// Screen will be cleared by defer
	return nil
}

// promptSaveChanges prompts the user to save changes before exiting
// Returns false if user wants to exit, true if user cancels exit
func promptSaveChanges(state *data.WeekPlannerState, reader input.InputReader) bool {
	fmt.Printf("\nYou have %d unsaved changes. Save before exiting? (y/n/c): ", len(state.Plan.Changes))

	for {
		char, _, err := reader.GetKey()
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
				_, _, _ = reader.GetKey() // Ignore error, just wait for key
				return true               // Return to planner to try again
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

// promptResetConfirmation prompts the user to confirm reset action
// Returns true if user confirms reset, false if cancelled
func promptResetConfirmation(state *data.WeekPlannerState, reader input.InputReader) bool {
	if state.Plan.HasChanges() {
		fmt.Printf("\nYou have %d unsaved changes. Reset and discard all changes? (y/n): ", len(state.Plan.Changes))
	} else {
		fmt.Print("\nReset and reload plan from disk? (y/n): ")
	}

	for {
		char, _, err := reader.GetKey()
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return false
		}

		switch char {
		case 'y', 'Y':
			fmt.Println("y")
			return true // Confirm reset

		case 'n', 'N':
			fmt.Println("n")
			return false // Cancel reset

		default:
			// Invalid input, keep prompting
			continue
		}
	}
}

// promptForTodoTitle prompts the user for a todo title
// Returns the title string, or empty string if cancelled
func promptForTodoTitle(dayName string, date time.Time, reader input.InputReader) (string, error) {
	dateStr := date.Format("Jan 02")
	fmt.Printf("\nCreate todo on %s (%s): ", dayName, dateStr)

	var title strings.Builder

	for {
		char, key, err := reader.GetKey()
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}

		switch key {
		case keyboard.KeyEnter:
			fmt.Println()
			if title.Len() == 0 {
				return "", nil // Empty title = cancel
			}
			return title.String(), nil

		case keyboard.KeyEsc:
			fmt.Println("\nCancelled")
			return "", nil

		case keyboard.KeyBackspace, keyboard.KeyBackspace2:
			if title.Len() > 0 {
				titleStr := title.String()
				title.Reset()
				title.WriteString(titleStr[:len(titleStr)-1])
				fmt.Print("\b \b")
			}

		case keyboard.KeySpace:
			title.WriteRune(' ')
			fmt.Print(" ")

		default:
			if char != 0 && char >= 32 && char <= 126 { // Printable ASCII
				title.WriteRune(char)
				fmt.Printf("%c", char)
			}
		}
	}
}

func runObjectivesView(reader input.InputReader) error {
	// Initialize state
	state, err := data.NewObjectivesViewState()
	if err != nil {
		return fmt.Errorf("error initializing objectives view: %w", err)
	}

	lastMessage := ""
	lastChar := rune(0) // For 'dd' and other multi-key commands

	for {
		// Render current view
		var display string
		if state.ViewMode == data.ObjectivesListView {
			display = presentation.RenderObjectivesListView(state)
		} else {
			display = presentation.RenderSingleObjectiveView(state)
		}
		fmt.Print(display)

		if lastMessage != "" {
			fmt.Printf("\n%s\n", lastMessage)
			lastMessage = ""
		}

		// Get input
		char, key, err := reader.GetKey()
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		// Handle 'dd' for delete
		if char == 'd' && lastChar == 'd' {
			lastChar = rune(0) // Reset
			if state.ViewMode == data.ObjectivesListView {
				obj := state.GetSelectedObjective()
				if obj != nil {
					_, total, _ := data.GetCompletionStats(obj.ObjectiveID)
					fmt.Printf("\nDelete objective \"%s\"?\n", obj.Title)
					if total > 0 {
						fmt.Printf("(%d linked todo(s) will be unlinked but not deleted)\n", total)
					}
					fmt.Print("(y/n): ")

					confirmChar, _, _ := reader.GetKey()
					if confirmChar == 'y' || confirmChar == 'Y' {
						err := scripts.DeleteParentObjective(*obj, data.QueryChildrenByObjectiveID, data.WriteFile)
						if err != nil {
							lastMessage = fmt.Sprintf("Error deleting: %v", err)
						} else {
							lastMessage = "Deleted successfully."
							state.Refresh()
						}
					} else {
						lastMessage = "Cancelled."
					}
				}
			}
			continue
		}

		lastChar = char

		input := presentation.ParseObjectivesInput(char, key)

		switch input.Action {
		case presentation.ObjNavigateNext:
			state.SelectNext()

		case presentation.ObjNavigatePrevious:
			state.SelectPrevious()

		case presentation.ObjOpenSelected:
			if state.ViewMode == data.ObjectivesListView {
				err := state.OpenSelectedObjective()
				if err != nil {
					lastMessage = fmt.Sprintf("Error: %v", err)
				}
			} else {
				// Open in editor
				if state.OnParent {
					openNoteInEditor(state.CurrentObjective.Name)
				} else {
					child := state.GetSelectedChild()
					if child != nil {
						openNoteInEditor(child.Name)
					}
				}
				state.Refresh()
			}

		case presentation.ObjCreateNew:
			if state.ViewMode == data.ObjectivesListView {
				// Create new objective
				fmt.Print("\nCreate new objective\nTitle: ")
				title, err := getLineInput(reader)
				if err != nil {
					lastMessage = fmt.Sprintf("Error: %v", err)
					continue
				}

				newObj, err := scripts.CreateParentObjective(title, data.WriteFile)
				if err != nil {
					lastMessage = fmt.Sprintf("Error creating objective: %v", err)
				} else {
					lastMessage = fmt.Sprintf("Created objective: \"%s\" (%s)", newObj.Title, newObj.ObjectiveID)
					state.Refresh()
				}
			} else {
				// Create new child todo
				fmt.Print("\nCreate new child todo\nTitle: ")
				title, err := getLineInput(reader)
				if err != nil {
					lastMessage = fmt.Sprintf("Error: %v", err)
					continue
				}

				newChild, err := scripts.CreateChildTodo(title, *state.CurrentObjective, data.WriteFile)
				if err != nil {
					lastMessage = fmt.Sprintf("Error creating child: %v", err)
				} else {
					lastMessage = fmt.Sprintf("Created and linked: [P%d] %s", newChild.Priority, newChild.Title)
					state.Refresh()
				}
			}

		case presentation.ObjLinkExisting:
			// Get the parent objective (either from list or single view)
			var parentObj *scripts.File
			if state.ViewMode == data.ObjectivesListView {
				parentObj = state.GetSelectedObjective()
			} else {
				parentObj = state.CurrentObjective
			}

			if parentObj != nil {
				selectedTodo, err := presentation.SearchAndLinkTodo(*parentObj, closeKeyboard, reopenKeyboard)
				if err != nil {
					lastMessage = fmt.Sprintf("Error: %v", err)
				} else if selectedTodo != nil {
					// Link the selected todo
					err := scripts.LinkTodoToObjective(*selectedTodo, *parentObj, data.WriteFile)
					if err != nil {
						lastMessage = fmt.Sprintf("Error linking: %v", err)
					} else {
						lastMessage = fmt.Sprintf("Linked \"%s\" to objective", selectedTodo.Title)
						state.Refresh()
					}
				}
			}

		case presentation.ObjQuit:
			if state.ViewMode == data.SingleObjectiveView {
				state.BackToList()
			} else {
				return nil // Exit objectives view
			}

		case presentation.ObjEditParent:
			if state.ViewMode == data.SingleObjectiveView {
				openNoteInEditor(state.CurrentObjective.Name)
				state.Refresh()
			}

		case presentation.ObjUnlinkChild:
			if state.ViewMode == data.SingleObjectiveView && !state.OnParent {
				child := state.GetSelectedChild()
				if child != nil {
					fmt.Printf("\nUnlink \"%s\" from this objective? (y/n): ", child.Title)
					confirmChar, _, _ := reader.GetKey()
					if confirmChar == 'y' || confirmChar == 'Y' {
						err := scripts.UnlinkTodoFromObjective(*child, data.WriteFile)
						if err != nil {
							lastMessage = fmt.Sprintf("Error unlinking: %v", err)
						} else {
							lastMessage = "Unlinked successfully."
							state.Refresh()
						}
					}
				}
			}

		case presentation.ObjChangeSort:
			if state.ViewMode == data.SingleObjectiveView {
				state.ToggleSortOrder()
			}

		case presentation.ObjChangeFilter:
			if state.ViewMode == data.SingleObjectiveView {
				state.CycleFilterMode()
			}

		case presentation.ObjSetPriority1, presentation.ObjSetPriority2, presentation.ObjSetPriority3:
			if state.ViewMode == data.SingleObjectiveView && !state.OnParent {
				child := state.GetSelectedChild()
				if child != nil {
					var priority scripts.Priority
					switch input.Action {
					case presentation.ObjSetPriority1:
						priority = scripts.P1
					case presentation.ObjSetPriority2:
						priority = scripts.P2
					case presentation.ObjSetPriority3:
						priority = scripts.P3
					}

					err := scripts.ChangePriority(priority, *child, data.WriteFile)
					if err != nil {
						lastMessage = fmt.Sprintf("Error changing priority: %v", err)
					} else {
						lastMessage = fmt.Sprintf("Priority changed to P%d", priority)
						state.Refresh()
					}
				}
			}

		case presentation.ObjSetDueToday:
			if state.ViewMode == data.SingleObjectiveView && !state.OnParent {
				child := state.GetSelectedChild()
				if child != nil {
					err := scripts.SetDueDateToToday(*child, data.WriteFile)
					if err != nil {
						lastMessage = fmt.Sprintf("Error setting due date: %v", err)
					} else {
						lastMessage = "Due date set to today"
						state.Refresh()
					}
				}
			}

		case presentation.ObjSetDueMonday:
			if state.ViewMode == data.SingleObjectiveView && !state.OnParent {
				child := state.GetSelectedChild()
				if child != nil {
					err := scripts.SetDueDateToNextDay(time.Monday, *child, data.WriteFile)
					if err != nil {
						lastMessage = fmt.Sprintf("Error setting due date: %v", err)
					} else {
						lastMessage = "Due date set to next Monday"
						state.Refresh()
					}
				}
			}

		case presentation.ObjSetDueTuesday:
			if state.ViewMode == data.SingleObjectiveView && !state.OnParent {
				child := state.GetSelectedChild()
				if child != nil {
					err := scripts.SetDueDateToNextDay(time.Tuesday, *child, data.WriteFile)
					if err != nil {
						lastMessage = fmt.Sprintf("Error setting due date: %v", err)
					} else {
						lastMessage = "Due date set to next Tuesday"
						state.Refresh()
					}
				}
			}

		case presentation.ObjSetDueWednesday:
			if state.ViewMode == data.SingleObjectiveView && !state.OnParent {
				child := state.GetSelectedChild()
				if child != nil {
					err := scripts.SetDueDateToNextDay(time.Wednesday, *child, data.WriteFile)
					if err != nil {
						lastMessage = fmt.Sprintf("Error setting due date: %v", err)
					} else {
						lastMessage = "Due date set to next Wednesday"
						state.Refresh()
					}
				}
			}

		case presentation.ObjSetDueThursday:
			if state.ViewMode == data.SingleObjectiveView && !state.OnParent {
				child := state.GetSelectedChild()
				if child != nil {
					err := scripts.SetDueDateToNextDay(time.Thursday, *child, data.WriteFile)
					if err != nil {
						lastMessage = fmt.Sprintf("Error setting due date: %v", err)
					} else {
						lastMessage = "Due date set to next Thursday"
						state.Refresh()
					}
				}
			}

		case presentation.ObjSetDueFriday:
			if state.ViewMode == data.SingleObjectiveView && !state.OnParent {
				child := state.GetSelectedChild()
				if child != nil {
					err := scripts.SetDueDateToNextDay(time.Friday, *child, data.WriteFile)
					if err != nil {
						lastMessage = fmt.Sprintf("Error setting due date: %v", err)
					} else {
						lastMessage = "Due date set to next Friday"
						state.Refresh()
					}
				}
			}

		case presentation.ObjSetDueSaturday:
			if state.ViewMode == data.SingleObjectiveView && !state.OnParent {
				child := state.GetSelectedChild()
				if child != nil {
					err := scripts.SetDueDateToNextDay(time.Saturday, *child, data.WriteFile)
					if err != nil {
						lastMessage = fmt.Sprintf("Error setting due date: %v", err)
					} else {
						lastMessage = "Due date set to next Saturday"
						state.Refresh()
					}
				}
			}

		case presentation.ObjSetDueSunday:
			if state.ViewMode == data.SingleObjectiveView && !state.OnParent {
				child := state.GetSelectedChild()
				if child != nil {
					err := scripts.SetDueDateToNextDay(time.Sunday, *child, data.WriteFile)
					if err != nil {
						lastMessage = fmt.Sprintf("Error setting due date: %v", err)
					} else {
						lastMessage = "Due date set to next Sunday"
						state.Refresh()
					}
				}
			}
		}
	}
}

func runTalkToView(filterPerson string, reader input.InputReader) error {
	// Get terminal dimensions
	termWidth, termHeight, _ := term.GetSize(int(os.Stdout.Fd()))
	if termWidth == 0 {
		termWidth, termHeight = 100, 30 // Default dimensions
	}

	// Initialize state
	state, err := data.NewTalkToViewState(filterPerson)
	if err != nil {
		return fmt.Errorf("error initializing talk-to view: %w", err)
	}

	// Check if any todos were found
	if len(state.AllPeople) == 0 {
		if filterPerson != "" {
			fmt.Printf("No to-talk-%s items found\n", filterPerson)
		} else {
			fmt.Println("No to-talk items found")
		}
		return nil
	}

	lastMessage := ""

	for {
		// Render current view
		display := presentation.RenderTalkToView(state, termWidth, termHeight)
		fmt.Print(display)

		if lastMessage != "" {
			fmt.Printf("\n%s\n", lastMessage)
			lastMessage = ""
		}

		// Get input
		char, key, err := reader.GetKey()
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		// Parse and handle input
		input := presentation.ParseTalkToInput(char, key, state.ViewMode, state.SearchMode)
		shouldExit, message, err := presentation.HandleTalkToInput(state, input)
		if err != nil {
			return fmt.Errorf("error handling input: %w", err)
		}

		// Handle special messages
		if strings.HasPrefix(message, "OPEN_NOTE:") {
			fileName := strings.TrimPrefix(message, "OPEN_NOTE:")
			fmt.Print("\033[2J\033[H") // Clear screen
			openNoteInEditor(fileName)
			lastMessage = "Note closed"
		} else if strings.HasPrefix(message, "CREATE_NEW_NOTE:") {
			// Prompt for note title
			fmt.Print("\nCreate new note\nTitle: ")
			title, err := getLineInput(reader)
			if err != nil {
				lastMessage = fmt.Sprintf("Error: %v", err)
				continue
			}

			// Validate non-empty input
			if strings.TrimSpace(title) == "" {
				lastMessage = "Error: Note title cannot be empty"
				continue
			}

			// Create the new note file
			newFile, err := scripts.CreateTodo(title, data.WriteFile)
			if err != nil {
				lastMessage = fmt.Sprintf("Error creating note: %v", err)
				continue
			}

			// Update state to include the new note and transition to confirmation
			state.TargetNoteName = newFile.Name
			state.IsNewNote = true
			state.ViewMode = data.ConfirmationView
			lastMessage = fmt.Sprintf("Created note: %s", newFile.Name)
		} else {
			lastMessage = message
		}

		if shouldExit {
			break
		}
	}

	// Clear screen on exit
	fmt.Print("\033[2J\033[H")

	return nil
}

func getLineInput(reader input.InputReader) (string, error) {
	var inputStr strings.Builder

	for {
		char, key, err := reader.GetKey()
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}

		switch key {
		case keyboard.KeyEnter:
			fmt.Println()
			result := strings.TrimSpace(inputStr.String())
			if result == "" {
				return "", fmt.Errorf("input cannot be empty")
			}
			return result, nil

		case keyboard.KeyEsc:
			fmt.Println("\nCancelled")
			return "", fmt.Errorf("input cancelled")

		case keyboard.KeyBackspace, keyboard.KeyBackspace2:
			if inputStr.Len() > 0 {
				str := inputStr.String()
				inputStr.Reset()
				inputStr.WriteString(str[:len(str)-1])
				fmt.Print("\b \b")
			}

		case keyboard.KeySpace:
			inputStr.WriteRune(' ')
			fmt.Print(" ")

		default:
			if char != 0 && char >= 32 && char <= 126 { // Printable ASCII
				inputStr.WriteRune(char)
				fmt.Printf("%c", char)
			}
		}
	}
}
