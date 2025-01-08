package main

import (
	"fmt"
	"cli-notes/scripts"
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

			command = searchRecentFilesPrintAndReturnNewCommand(scripts.GetLatestFileThatHasBeenSearched)
		} else if key == keyboard.KeyArrowDown {

			command = searchRecentFilesPrintAndReturnNewCommand(scripts.GetPreviousFileThatHasBeenSearched)
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

		// Trim whitespace from each query
		for i, q := range queries {
			queries[i] = strings.TrimSpace(q)
		}
		scripts.QueryOpenTodos(queries)

	case "gt":
		scripts.GetTodos()

	case "gta":
		if len(parts) < 2 {
			fmt.Println("Please provide a tags to query")
			return
		}
		scripts.SearchNotesByTags(parts[1:])

	case "gq":
		if len(parts) < 2 {
			fmt.Println("Please provide a query to search")
			return
		}

		query := strings.Join(parts[1:], " ")
		scripts.SearchAllFilesPrintWhenMatch(query)

	case "gqa":
		if len(parts) < 2 {
			fmt.Println("Please provide a query to search")
			return
		}

		query := strings.Join(parts[1:], " ")
		scripts.SearchLastFilesSearchedForQueryPrintWhenMatch(query)

	case "gqa,":
		if len(parts) < 2 {
			fmt.Println("Please provide a query to search")
			return
		}
		queryString := strings.Join(parts[1:], " ")
		queries := strings.Split(queryString, ",")

		// Trim whitespace from each query
		for i, q := range queries {
			queries[i] = strings.TrimSpace(q)
		}

		scripts.QueryPreviouslySearchedFiles(queries)

	case "gq,":
		if len(parts) < 2 {
			fmt.Println("Please provide a query to search")
			return
		}

		queryString := strings.Join(parts[1:], " ")
		queries := strings.Split(queryString, ",")

		// Trim whitespace from each query
		for i, q := range queries {
			queries[i] = strings.TrimSpace(q)
		}

		scripts.QueryFiles(queries)

	case "gat":
		scripts.SearchPreviousFilesForUncompletedTasks()

	case "ct":
		if len(parts) < 2 {
			fmt.Println("Please provide a title for the new todo")
			return
		}
		title := strings.Join(parts[1:], "-")
		scripts.CreateTodo(title)

	case "cm":
		if len(parts) < 2 {
			fmt.Println("Please provide a title for the new meeting")
			return
		}
		title := strings.Join(parts[1:], "-")
		scripts.CreateMeeting(title)

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
		scripts.CreateStandup()

	case "gto":  
		scripts.GetOverdueTodos()

	case "gtnd":
		scripts.GetTodosWithNoDueDate()
	
	case "gts":
		scripts.GetSoonTodos()

	default:
		fmt.Println("Unknown command.")
	}
}

func searchRecentFilesPrintAndReturnNewCommand(search func() string) string {
	fileName := search()
	if fileName != "" {
		fmt.Println(fileName)
		return "o " + fileName
	} else {
		fmt.Println("No files have been searched yet.")
		fmt.Println("")
		return ""
	}
}
