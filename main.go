package main

import (
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"cli-notes/scripts/presentation"
	"cli-notes/scripts/presentation/searched_files_store"
	"fmt"
	"strings"

	"github.com/eiannone/keyboard"
)

func main() {
	closeChannel := make(chan bool)

	// restore if using git for backup
	// go scripts.MonitorDirectorySize("./notes", func() {
	// 	scripts.PushChangesToGit("./notes")
	// })

	go setupCommandScanner(func() {
		closeChannel <- true
	})

	<-closeChannel
	fmt.Println("Exiting...")
}

func setupCommandScanner(onClose func()) {
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	var command string
	fmt.Print("> ")
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}

		if key == keyboard.KeyArrowUp {

			command = searchRecentFilesPrintAndReturnNewCommand(searched_files_store.GetNextFile)
		} else if key == keyboard.KeyArrowDown {

			command = searchRecentFilesPrintAndReturnNewCommand(searched_files_store.GetPreviousFile)
		} else if key == keyboard.KeyEnter {

			handleCommand(command, onClose)
			fmt.Print("> ")
			command = ""
		} else if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {

			if len(command) > 0 {
				command = command[:len(command)-1]
				fmt.Print("\b \b")
			}
		} else if key == keyboard.KeySpace {

			command += " "
			fmt.Print(" ")
		} else if key == keyboard.KeyEsc {
			fmt.Println("Screw that line")
			fmt.Println("> ")
			command = ""
		} else {

			command += string(char)
			fmt.Print(string(char))
		}
	}
}

func handleCommand(command string, onClose func()) {
	parts := strings.Fields(command)

	if len(parts) == 0 {
		return
	}

	fmt.Println()
	switch parts[0] {

	case "gt,":
		if len(parts) < 2 {
			fmt.Println("Please provide a query to search")
			return
		}

		queryString := strings.Join(parts[1:], " ")
		queries := strings.Split(queryString, ",")

		for i, q := range queries {
			queries[i] = strings.TrimSpace(q)
		}
		files, err := scripts.QueryOpenTodos(queries, data.QueryFilesByDone)
		if err != nil {
			fmt.Printf("Error querying open todos: %v\n", err)
			return
		}
		searched_files_store.SetFilesSearched(files)
		presentation.PrintAllFileNames(files)
		

	// case "gt":
	// 	scripts.GetTodos()

	// case "gta":
	// 	if len(parts) < 2 {
	// 		fmt.Println("Please provide a tags to query")
	// 		return
	// 	}
	// 	scripts.SearchNotesByTags(parts[1:])

	// case "gq":
	// 	if len(parts) < 2 {
	// 		fmt.Println("Please provide a query to search")
	// 		return
	// 	}

	// 	query := strings.Join(parts[1:], " ")
	// 	scripts.SearchAllFilesPrintWhenMatch(query)

	// case "gqa":
	// 	if len(parts) < 2 {
	// 		fmt.Println("Please provide a query to search")
	// 		return
	// 	}

	// 	query := strings.Join(parts[1:], " ")
	// 	scripts.SearchLastFilesSearchedForQueryPrintWhenMatch(query)

	// case "gqa,":
	// 	if len(parts) < 2 {
	// 		fmt.Println("Please provide a query to search")
	// 		return
	// 	}
	// 	queryString := strings.Join(parts[1:], " ")
	// 	queries := strings.Split(queryString, ",")

	// 	// Trim whitespace from each query
	// 	for i, q := range queries {
	// 		queries[i] = strings.TrimSpace(q)
	// 	}

	// 	scripts.QueryPreviouslySearchedFiles(queries)

	// case "gq,":
	// 	if len(parts) < 2 {
	// 		fmt.Println("Please provide a query to search")
	// 		return
	// 	}

	// 	queryString := strings.Join(parts[1:], " ")
	// 	queries := strings.Split(queryString, ",")

	// 	// Trim whitespace from each query
	// 	for i, q := range queries {
	// 		queries[i] = strings.TrimSpace(q)
	// 	}

	// 	scripts.QueryFiles(queries)

	// case "gat":
	// 	scripts.SearchPreviousFilesForUncompletedTasks()

	case "ct":
		if len(parts) < 2 {
			fmt.Println("Please provide a title for the new todo")
			return
		}
		title := strings.Join(parts[1:], "-")
		file, err := scripts.CreateTodo(title, data.WriteFile)
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			return
		}
		filePath := "notes/" + file.Name
		scripts.OpenNoteInEditor(filePath)

	case "cm":
		if len(parts) < 2 {
			fmt.Println("Please provide a title for the new meeting")
			return
		}
		title := strings.Join(parts[1:], "-")
		file, err := scripts.CreateMeeting(title, data.WriteFile)
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			return
		}
		filePath := "notes/" + file.Name
		scripts.OpenNoteInEditor(filePath)

	case "o":
		if len(parts) < 2 {
			fmt.Println("Please provide a file name to open")
			return
		}
		title := parts[1]
		fileName := "notes/" + title
		scripts.OpenNoteInEditor(fileName)

	case "exit", "quit", "q":
		onClose()
		return

	case "cs":
		file, err := scripts.CreateStandup(data.GetTeamNames, data.WriteFile)
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			return
		}
		filePath := "notes/" + file.Name
		scripts.OpenNoteInEditor(filePath)

	// case "gto":  
	// 	scripts.GetOverdueTodos()

	// case "gtnd":
	// 	scripts.GetTodosWithNoDueDate()
	
	// case "gts":
	// 	scripts.GetSoonTodos()

	default:
		fmt.Println("Unknown command.")
	}
}

func searchRecentFilesPrintAndReturnNewCommand(search func() *scripts.File) string {
	file := search()
	if file == nil {
		fmt.Println("No files have been searched yet.")
		fmt.Println("")
		return ""
	}

		fmt.Println(file.Name)
		return "o " + file.Name
}
