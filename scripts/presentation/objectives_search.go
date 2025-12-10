package presentation

import (
	"cli-notes/input"
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"fmt"
	"strings"

	"github.com/eiannone/keyboard"
)

// SearchAndLinkTodo presents a search interface to find and link todos
func SearchAndLinkTodo(parentObjective scripts.File, reader input.InputReader) (*scripts.File, error) {
	fmt.Printf("\nLink existing todo to: %s\n", parentObjective.Title)
	fmt.Print("Enter search query (comma-separated): ")

	// Get search terms using keyboard input (supports escape to cancel)
	queryStr, err := readSearchQuery(reader)
	if err != nil {
		return nil, err
	}
	if queryStr == "" {
		return nil, nil // User cancelled
	}

	queries := strings.Split(queryStr, ",")
	for i, q := range queries {
		queries[i] = strings.TrimSpace(q)
	}

	// Search for matching todos
	fmt.Printf("\nSearching for todos containing: %v...\n\n", queries)

	results, err := data.QueryTodosWithoutObjective(queries)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no matching todos found")
	}

	// Display results with selection
	selectedIndex := 0

	for {
		fmt.Print("\033[2J\033[H") // Clear screen

		fmt.Printf("Search results for: %v\n", queries)
		fmt.Println("─────────────────────────────────")

		for i, todo := range results {
			if i == selectedIndex {
				fmt.Print("> ")
			} else {
				fmt.Print("  ")
			}

			fmt.Printf("%s (P%d", todo.Title, todo.Priority)
			if !todo.DueAt.IsZero() && todo.DueAt.Year() < 2100 {
				fmt.Printf(", due: %s", todo.DueAt.Format("2006-01-02"))
			} else {
				fmt.Print(", no due date")
			}
			fmt.Print(")")

			if todo.ObjectiveRole == "parent" {
				fmt.Print(" [PARENT]")
			} else if todo.ObjectiveID != "" {
				// This shouldn't happen with our query, but just in case
				fmt.Print(" [LINKED]")
			}
			fmt.Println()
		}

		fmt.Println("\nj/k=navigate, Enter=link, Esc=cancel")

		char, key, err := reader.GetKey()
		if err != nil {
			return nil, err
		}

		switch {
		case char == 'j':
			selectedIndex = (selectedIndex + 1) % len(results)
		case char == 'k':
			selectedIndex--
			if selectedIndex < 0 {
				selectedIndex = len(results) - 1
			}
		case key == keyboard.KeyEnter:
			selected := results[selectedIndex]
			if selected.ObjectiveRole == "parent" {
				fmt.Println("\nError: Cannot link parent objectives as children.")
				fmt.Println("Press any key to continue...")
				reader.GetKey()
				continue
			}
			return &selected, nil
		case key == keyboard.KeyEsc:
			return nil, nil
		}
	}
}

// readSearchQuery reads search query input character by character
// Returns empty string if user presses Escape to cancel
func readSearchQuery(reader input.InputReader) (string, error) {
	var inputStr strings.Builder

	for {
		char, key, err := reader.GetKey()
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}

		switch key {
		case keyboard.KeyEnter:
			fmt.Println()
			return strings.TrimSpace(inputStr.String()), nil

		case keyboard.KeyEsc:
			fmt.Println("\nCancelled")
			return "", nil

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
