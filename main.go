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
			fmt.Println(completedCommand)
			if completedCommand.SelectedFile != nil && completedCommand.Name == "" {
				command = presentation.WIPCommand{
					Text: fmt.Sprintf("o %v", completedCommand.SelectedFile.Name),
				}
				fmt.Println(command.SelectedFile.Name)
			}
			if completedCommand.SelectedFile != nil && completedCommand.Name != "" {
				queries := strings.Join(completedCommand.Queries, " ")
				command = presentation.WIPCommand{
					Text: fmt.Sprintf("%v %v %v", completedCommand.Name, queries, completedCommand.SelectedFile.Name),
				}
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

func handleCommand(completedCommand presentation.CompletedCommand, onClose func(), fileStore *data.SearchedFilesStore) {
	switch completedCommand.Name {
	case "gt":
		if len(completedCommand.Queries) == 0 {
			files, err := scripts.GetTodos(data.QueryFilesByDone)
			if err != nil {
				fmt.Printf("Error getting todos: %v\n", err)
			}
			onFilesFetched(files, fileStore)
		} else {
			files, err := scripts.QueryOpenTodos(completedCommand.Queries, data.QueryFilesByDone)
			if err != nil {
				fmt.Printf("Error querying open todos: %v\n", err)
				return
			}
			onFilesFetched(files, fileStore)
		}

	case "gta":
		if len(completedCommand.Queries) == 0 {
			fmt.Println("Please provide a tags to query")
			return
		}
		files, err := scripts.SearchNotesByTags(completedCommand.Queries, data.QueryNotesByTags)
		if err != nil {
			fmt.Printf("Error getting notes by tags: %v\n", err)
		}
		onFilesFetched(files, fileStore)

	case "gq":
		if len(completedCommand.Queries) == 0 {
			fmt.Println("Please provide a query to search")
			return
		}
		files, err := scripts.QueryAllFiles(completedCommand.Queries, data.QueryFiles)
		if err != nil {
			fmt.Printf("Error querying notes: %v", err)
		}
		onFilesFetched(files, fileStore)

	case "gqa":
		if len(completedCommand.Queries) == 0 {
			fmt.Println("Please provide a query to search")
			return
		}
		previousFiles := fileStore.GetFilesSearched()
		if len(previousFiles) == 0 {
			fmt.Println("No files have been queried")
		} else {
			files := scripts.QueryFiles(completedCommand.Queries, previousFiles)
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
		handleCreateFile("todo", completedCommand.Queries, scripts.CreateTodo)

	case "cm":
		handleCreateFile("meeting", completedCommand.Queries, scripts.CreateMeeting)

	case "cp":
		handleCreateFile("plan", completedCommand.Queries, scripts.CreateSevenQuestions)

	case "o":
		if completedCommand.SelectedFile != nil {
			openNoteInEditor(completedCommand.SelectedFile.Name)
		} else if len(completedCommand.Queries) > 0 {
			openNoteInEditor(completedCommand.Queries[0])
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
		if len(completedCommand.Queries) < 1 {
			fmt.Println("Please provide an amount of days to delay and a file name")
			return
		}
		delayDays, err := strconv.Atoi(completedCommand.Queries[0])
		if err != nil {
			fmt.Printf("Error converting days: %v\n", err)
			return
		}
		fmt.Println(delayDays)
		// TODO: Implement delay functionality
		// err = scripts.DelayDueDate(delayDays, )

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

func openNoteInEditor(fileName string) {
	filePath := "notes/" + fileName
	err := presentation.OpenNoteInEditor(filePath)
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
