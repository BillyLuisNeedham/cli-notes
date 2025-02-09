package main

import (
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"cli-notes/scripts/presentation"
	"fmt"
	"github.com/eiannone/keyboard"
	"strconv"
	"strings"
)

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
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	// TODO extract this logic into some form of class
	// can have the command as an interface
	// one can handle just text
	// one can handle text with a file to pass
	command := presentation.WIPCommand{}

	fmt.Print("> ")
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}

		nextCommand := presentation.CommandHandler(
			char,
			key,
			command,
			func() *scripts.File { return searchRecentFilesPrintIfNotFound(fileStore.GetNextFile) },
			func() *scripts.File { return searchRecentFilesPrintIfNotFound(fileStore.GetPreviousFile) },
			func() { fmt.Print("\b \b") },
		)

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

		case presentation.CompletedCommand:
			completedCommand := nextCommand
			if completedCommand.SelectedFile != nil && completedCommand.Name == "" {
				command = presentation.WIPCommand{
					Text: fmt.Sprintf("o %v", completedCommand.SelectedFile.Name),
				}
				fmt.Println(command.SelectedFile.Name)
			}
			if completedCommand.SelectedFile != nil && completedCommand.Name != "" {
				command = presentation.WIPCommand{
					Text: fmt.Sprintf("%v %v", completedCommand.Name, completedCommand.SelectedFile.Name),
				}
				fmt.Println("")
			}
			handleCommand(command.Text, onClose, fileStore)
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

func handleCommand(command string, onClose func(), fileStore *data.SearchedFilesStore) {
	parts := strings.Fields(command)

	if len(parts) == 0 {
		return
	}

	fmt.Println()
	switch parts[0] {
	case "gt":
		if len(parts) < 2 {
			files, err := scripts.GetTodos(data.QueryFilesByDone)
			if err != nil {
				fmt.Printf("Error getting todos: %v\n", err)
			}
			onFilesFetched(files, fileStore)
		} else {
			queries := getQueries(parts)
			files, err := scripts.QueryOpenTodos(queries, data.QueryFilesByDone)
			if err != nil {
				fmt.Printf("Error querying open todos: %v\n", err)
				return
			}
			onFilesFetched(files, fileStore)
		}

	case "gta":
		if len(parts) < 2 {
			fmt.Println("Please provide a tags to query")
			return
		}
		files, err := scripts.SearchNotesByTags(parts[1:], data.QueryNotesByTags)
		if err != nil {
			fmt.Printf("Error getting notes by tags: %v\n", err)
		}
		onFilesFetched(files, fileStore)

	case "gq":
		if len(parts) < 2 {
			fmt.Println("Please provide a query to search")
			return
		}

		queries := getQueries(parts)
		files, err := scripts.QueryAllFiles(queries, data.QueryFiles)
		if err != nil {
			fmt.Printf("Error querying notes: %v", err)
		}
		onFilesFetched(files, fileStore)

	case "gqa":
		if len(parts) < 2 {
			fmt.Println("Please provide a query to search")
			return
		}

		previousFiles := fileStore.GetFilesSearched()
		if len(previousFiles) == 0 {
			fmt.Println("No files have been queried")
		} else {
			queries := getQueries(parts)
			files := scripts.QueryFiles(queries, previousFiles)
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
		handleCreateFile("todo", parts, scripts.CreateTodo)

	case "cm":
		handleCreateFile("meeting", parts, scripts.CreateMeeting)

	case "cp":
		handleCreateFile("plan", parts, scripts.CreateSevenQuestions)

	case "o":
		if len(parts) < 2 {
			fmt.Println("Please provide a file name to open")
			return
		}
		title := parts[1]
		openNoteInEditor(title)

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
		// TODO build logic to update a files date by x days
		fmt.Println(parts)
		if len(parts) < 3 {
			fmt.Println("Please provide an amount of days to delay and a file name")
			return
		}
		delayDaysCommand := parts[1]
		delayDays, err := strconv.Atoi(delayDaysCommand)
		if err != nil {
			fmt.Printf("Error converting days: %v", err)
			fmt.Println("")
		}
		fmt.Println(delayDays)

		// scripts.DelayDueDate(delayDays, )

	default:
		fmt.Println("Unknown command.")
	}
}

func onFilesFetched(files []scripts.File, fileStore *data.SearchedFilesStore) {
	fileStore.SetFilesSearched(files)
	presentation.PrintAllFileNames(files)
}

func searchRecentFilesPrintIfNotFound(search func() *scripts.File) *scripts.File {
	file := search()
	if file == nil {
		fmt.Println("No files have been searched yet.")
		fmt.Println("")
		return nil
	}

	return file
}

func getQueries(commandParts []string) []string {
	queryString := strings.Join(commandParts[1:], " ")
	queries := strings.Split(queryString, ",")

	for i, q := range queries {
		queries[i] = strings.TrimSpace(q)
	}

	return queries
}

func openNoteInEditor(fileName string) {
	filePath := "notes/" + fileName
	err := presentation.OpenNoteInEditor(filePath)
	if err != nil {
		fmt.Printf("Error opening note in editor: %v", err)
	}
}

func handleCreateFile(fileType string, parts []string, createFn func(string, scripts.OnFileCreated) (scripts.File, error)) {
	if len(parts) < 2 {
		fmt.Printf("Please provide a title for the new %s\n", fileType)
		return
	}
	title := strings.Join(parts[1:], "-")
	file, err := createFn(title, data.WriteFile)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}
	openNoteInEditor(file.Name)
}
