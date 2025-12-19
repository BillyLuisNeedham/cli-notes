package presentation

import (
	"cli-notes/input"
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"fmt"
	"strings"

	"github.com/eiannone/keyboard"
	"github.com/sahilm/fuzzy"
)

// noteList implements fuzzy.Source for note searching
type noteList []scripts.File

func (n noteList) String(i int) string {
	return n[i].Title
}

func (n noteList) Len() int {
	return len(n)
}

// SearchAndSelectNote presents a fuzzy search interface to select a note
func SearchAndSelectNote(reader input.InputReader) (*scripts.File, error) {
	// Load all notes
	allNotes, err := data.QueryFiles("")
	if err != nil {
		return nil, fmt.Errorf("error loading notes: %w", err)
	}

	if len(allNotes) == 0 {
		return nil, fmt.Errorf("no notes found")
	}

	searchQuery := ""
	selectedIndex := 0
	var filteredNotes []scripts.File

	// Initial display - show all notes
	filteredNotes = allNotes

	for {
		fmt.Print("\033[2J\033[H") // Clear screen

		fmt.Println("Select note to link")
		fmt.Println("─────────────────────────────────")
		fmt.Printf("Search: %s\n", searchQuery)
		fmt.Println("─────────────────────────────────")

		// Apply fuzzy filter if there's a search query
		if searchQuery != "" {
			matches := fuzzy.FindFrom(searchQuery, noteList(allNotes))
			filteredNotes = make([]scripts.File, len(matches))
			for i, match := range matches {
				filteredNotes[i] = allNotes[match.Index]
			}
		} else {
			filteredNotes = allNotes
		}

		// Clamp selected index
		if selectedIndex >= len(filteredNotes) {
			selectedIndex = len(filteredNotes) - 1
		}
		if selectedIndex < 0 {
			selectedIndex = 0
		}

		// Display notes (max 15)
		displayCount := min(15, len(filteredNotes))
		for i := 0; i < displayCount; i++ {
			note := filteredNotes[i]
			if i == selectedIndex {
				fmt.Print("> ")
			} else {
				fmt.Print("  ")
			}

			// Show title with metadata
			fmt.Printf("%s", note.Title)
			if len(note.Tags) > 0 {
				fmt.Printf(" [%s]", strings.Join(note.Tags, ", "))
			}
			fmt.Println()
		}

		if len(filteredNotes) > displayCount {
			fmt.Printf("  ... and %d more\n", len(filteredNotes)-displayCount)
		}

		if len(filteredNotes) == 0 {
			fmt.Println("  (no matching notes)")
		}

		fmt.Println("\nType to search, j/k=navigate, Enter=select, Esc=cancel")

		char, key, err := reader.GetKey()
		if err != nil {
			return nil, err
		}

		switch key {
		case keyboard.KeyEnter:
			if len(filteredNotes) > 0 && selectedIndex < len(filteredNotes) {
				selected := filteredNotes[selectedIndex]
				return &selected, nil
			}
			continue

		case keyboard.KeyEsc:
			return nil, nil

		case keyboard.KeyBackspace, keyboard.KeyBackspace2:
			if len(searchQuery) > 0 {
				searchQuery = searchQuery[:len(searchQuery)-1]
				selectedIndex = 0
			}

		case keyboard.KeySpace:
			searchQuery += " "
			selectedIndex = 0

		default:
			// Handle j/k navigation
			if char == 'j' && searchQuery == "" {
				if len(filteredNotes) > 0 {
					selectedIndex = (selectedIndex + 1) % len(filteredNotes)
				}
			} else if char == 'k' && searchQuery == "" {
				if len(filteredNotes) > 0 {
					selectedIndex--
					if selectedIndex < 0 {
						selectedIndex = len(filteredNotes) - 1
					}
				}
			} else if char != 0 && char >= 32 && char <= 126 {
				// Printable character - add to search query
				searchQuery += string(char)
				selectedIndex = 0
			}
		}
	}
}
