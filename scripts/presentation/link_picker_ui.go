package presentation

import (
	"cli-notes/input"
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"fmt"
	"strings"
)

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

	state := data.NewLinkPickerState(allNotes)

	for {
		fmt.Print("\033[2J\033[H") // Clear screen

		fmt.Println("Select note to link")
		fmt.Println("─────────────────────────────────")
		fmt.Printf("Search: %s\n", state.Query)
		fmt.Println("─────────────────────────────────")

		// Clamp selected index
		state.ClampSelectedIndex()

		// Display notes (max 15)
		displayCount := min(15, len(state.FilteredNotes))
		for i := 0; i < displayCount; i++ {
			note := state.FilteredNotes[i]
			if i == state.SelectedIndex {
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

		if len(state.FilteredNotes) > displayCount {
			fmt.Printf("  ... and %d more\n", len(state.FilteredNotes)-displayCount)
		}

		if len(state.FilteredNotes) == 0 {
			fmt.Println("  (no matching notes)")
		}

		// Show mode-specific help
		if state.Mode == data.LinkPickerModeInsert {
			fmt.Println("\n[INSERT] Type to search, Esc=normal mode, Enter=select")
		} else {
			fmt.Println("\n[NORMAL] i=insert, j/k=navigate, Enter=select, q/Esc=cancel")
		}

		char, key, err := reader.GetKey()
		if err != nil {
			return nil, err
		}

		input := ParseLinkPickerInput(char, key, state.Mode)

		switch input.Action {
		case LinkPickerSelect:
			if selected := state.GetSelectedNote(); selected != nil {
				return selected, nil
			}

		case LinkPickerCancel:
			return nil, nil

		case LinkPickerEnterInsert:
			state.EnterInsertMode()

		case LinkPickerEnterNormal:
			state.EnterNormalMode()

		case LinkPickerNavigateDown:
			state.SelectNext()

		case LinkPickerNavigateUp:
			state.SelectPrevious()

		case LinkPickerAddChar:
			state.AddChar(input.Char)

		case LinkPickerDeleteChar:
			state.DeleteChar()
		}
	}
}

// SelectObjectiveWithFuzzy presents a fuzzy search interface to select an objective
func SelectObjectiveWithFuzzy(objectives []scripts.File, reader input.InputReader) (*scripts.File, error) {
	if len(objectives) == 0 {
		return nil, fmt.Errorf("no objectives found")
	}

	state := data.NewLinkPickerState(objectives)

	for {
		fmt.Print("\033[2J\033[H") // Clear screen

		fmt.Println("Select objective to link")
		fmt.Println("─────────────────────────────────")
		fmt.Printf("Search: %s\n", state.Query)
		fmt.Println("─────────────────────────────────")

		// Clamp selected index
		state.ClampSelectedIndex()

		// Display objectives (max 15)
		displayCount := min(15, len(state.FilteredNotes))
		for i := 0; i < displayCount; i++ {
			obj := state.FilteredNotes[i]
			if i == state.SelectedIndex {
				fmt.Print("> ")
			} else {
				fmt.Print("  ")
			}

			// Show title with tags
			fmt.Printf("%s", obj.Title)
			if len(obj.Tags) > 0 {
				fmt.Printf(" [%s]", strings.Join(obj.Tags, ", "))
			}
			fmt.Println()
		}

		if len(state.FilteredNotes) > displayCount {
			fmt.Printf("  ... and %d more\n", len(state.FilteredNotes)-displayCount)
		}

		if len(state.FilteredNotes) == 0 {
			fmt.Println("  (no matching objectives)")
		}

		// Show mode-specific help
		if state.Mode == data.LinkPickerModeInsert {
			fmt.Println("\n[INSERT] Type to search, Esc=normal mode, Enter=select")
		} else {
			fmt.Println("\n[NORMAL] i=insert, j/k=navigate, Enter=select, q/Esc=cancel")
		}

		char, key, err := reader.GetKey()
		if err != nil {
			return nil, err
		}

		input := ParseLinkPickerInput(char, key, state.Mode)

		switch input.Action {
		case LinkPickerSelect:
			if selected := state.GetSelectedNote(); selected != nil {
				return selected, nil
			}

		case LinkPickerCancel:
			return nil, nil

		case LinkPickerEnterInsert:
			state.EnterInsertMode()

		case LinkPickerEnterNormal:
			state.EnterNormalMode()

		case LinkPickerNavigateDown:
			state.SelectNext()

		case LinkPickerNavigateUp:
			state.SelectPrevious()

		case LinkPickerAddChar:
			state.AddChar(input.Char)

		case LinkPickerDeleteChar:
			state.DeleteChar()
		}
	}
}
