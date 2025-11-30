package presentation

import (
	"cli-notes/scripts/data"
	"fmt"

	"github.com/eiannone/keyboard"
)

// TalkToAction represents an action in the Talk-To interface
type TalkToAction int

const (
	TTNoAction TalkToAction = iota
	// Person/Todo selection
	TTNavigateNext
	TTNavigatePrevious
	TTToggleSelection
	TTSelectAll
	TTSelectNone
	TTConfirm
	TTBack
	TTQuit
	// Note selection
	TTFindNote
	TTCreateNewNote
	// Search modal
	TTSearchType         // Append character to query
	TTSearchBackspace    // Remove last character
	TTSearchToggleMode   // Esc: toggle INSERT/NORMAL
	TTSearchEnterInsert  // i: enter INSERT mode
	TTSearchNavigateNext // j or arrow down
	TTSearchNavigatePrev // k or arrow up
	TTSearchConfirm      // Enter: select note
	TTSearchCancel       // q in NORMAL mode
	// Success view
	TTUndo
	TTOpenNote
	TTReturnToPerson
)

// TalkToInput represents parsed input with an action and optional data
type TalkToInput struct {
	Action TalkToAction
	Char   rune // For text input in search mode
}

// ParseTalkToInput parses keyboard input based on the current view mode
func ParseTalkToInput(char rune, key keyboard.Key, viewMode data.TalkToViewMode, searchMode data.SearchMode) TalkToInput {
	switch viewMode {
	case data.PersonSelectionView:
		return parsePersonSelectionInput(char, key)
	case data.TodoSelectionView:
		return parseTodoSelectionInput(char, key)
	case data.NoteSelectionView:
		return parseNoteSelectionInput(char, key)
	case data.NoteSearchModalView:
		return parseSearchModalInput(char, key, searchMode)
	case data.ConfirmationView:
		return parseConfirmationInput(char, key)
	case data.SuccessView:
		return parseSuccessInput(char, key)
	}

	return TalkToInput{Action: TTNoAction}
}

// parsePersonSelectionInput handles input in person selection view
func parsePersonSelectionInput(char rune, key keyboard.Key) TalkToInput {
	switch key {
	case keyboard.KeyArrowDown:
		return TalkToInput{Action: TTNavigateNext}
	case keyboard.KeyArrowUp:
		return TalkToInput{Action: TTNavigatePrevious}
	case keyboard.KeyEnter:
		return TalkToInput{Action: TTConfirm}
	case keyboard.KeyEsc:
		return TalkToInput{Action: TTQuit}
	}

	switch char {
	case 'j':
		return TalkToInput{Action: TTNavigateNext}
	case 'k':
		return TalkToInput{Action: TTNavigatePrevious}
	case 'q':
		return TalkToInput{Action: TTQuit}
	}

	return TalkToInput{Action: TTNoAction}
}

// parseTodoSelectionInput handles input in todo selection view
func parseTodoSelectionInput(char rune, key keyboard.Key) TalkToInput {
	switch key {
	case keyboard.KeyArrowDown:
		return TalkToInput{Action: TTNavigateNext}
	case keyboard.KeyArrowUp:
		return TalkToInput{Action: TTNavigatePrevious}
	case keyboard.KeyEnter:
		return TalkToInput{Action: TTConfirm}
	case keyboard.KeyEsc:
		return TalkToInput{Action: TTBack}
	case keyboard.KeySpace:
		return TalkToInput{Action: TTToggleSelection}
	}

	switch char {
	case 'j':
		return TalkToInput{Action: TTNavigateNext}
	case 'k':
		return TalkToInput{Action: TTNavigatePrevious}
	case ' ':
		return TalkToInput{Action: TTToggleSelection}
	case 'a':
		return TalkToInput{Action: TTSelectAll}
	case 'n':
		return TalkToInput{Action: TTSelectNone}
	case 'q':
		return TalkToInput{Action: TTBack}
	}

	return TalkToInput{Action: TTNoAction}
}

// parseNoteSelectionInput handles input in note selection view
func parseNoteSelectionInput(char rune, key keyboard.Key) TalkToInput {
	switch key {
	case keyboard.KeyEsc:
		return TalkToInput{Action: TTBack}
	}

	switch char {
	case 'f':
		return TalkToInput{Action: TTFindNote}
	case 'n':
		return TalkToInput{Action: TTCreateNewNote}
	case 'q':
		return TalkToInput{Action: TTBack}
	}

	return TalkToInput{Action: TTNoAction}
}

// parseSearchModalInput handles input in search modal
func parseSearchModalInput(char rune, key keyboard.Key, searchMode data.SearchMode) TalkToInput {
	// Handle special keys first
	switch key {
	case keyboard.KeyEsc:
		return TalkToInput{Action: TTSearchToggleMode}
	case keyboard.KeyEnter:
		return TalkToInput{Action: TTSearchConfirm}
	case keyboard.KeyBackspace, keyboard.KeyBackspace2:
		return TalkToInput{Action: TTSearchBackspace}
	case keyboard.KeyArrowDown:
		return TalkToInput{Action: TTSearchNavigateNext}
	case keyboard.KeyArrowUp:
		return TalkToInput{Action: TTSearchNavigatePrev}
	}

	// In INSERT mode, all printable characters are typed into the search query
	if searchMode == data.InsertMode {
		if char >= 32 && char <= 126 {
			return TalkToInput{Action: TTSearchType, Char: char}
		}
		return TalkToInput{Action: TTNoAction}
	}

	// In NORMAL mode, handle vim-style keys
	switch char {
	case 'i':
		return TalkToInput{Action: TTSearchEnterInsert}
	case 'j':
		return TalkToInput{Action: TTSearchNavigateNext}
	case 'k':
		return TalkToInput{Action: TTSearchNavigatePrev}
	case 'q':
		return TalkToInput{Action: TTSearchCancel}
	}

	return TalkToInput{Action: TTNoAction}
}

// parseConfirmationInput handles input in confirmation view
func parseConfirmationInput(char rune, key keyboard.Key) TalkToInput {
	switch key {
	case keyboard.KeyEnter:
		return TalkToInput{Action: TTConfirm}
	case keyboard.KeyEsc:
		return TalkToInput{Action: TTBack}
	}

	switch char {
	case 'y', 'Y':
		return TalkToInput{Action: TTConfirm}
	case 'c', 'C':
		return TalkToInput{Action: TTBack}
	case 'q':
		return TalkToInput{Action: TTQuit}
	}

	return TalkToInput{Action: TTNoAction}
}

// parseSuccessInput handles input in success view
func parseSuccessInput(char rune, key keyboard.Key) TalkToInput {
	switch key {
	case keyboard.KeyEnter:
		return TalkToInput{Action: TTOpenNote}
	case keyboard.KeyEsc:
		return TalkToInput{Action: TTQuit}
	}

	switch char {
	case 'u':
		return TalkToInput{Action: TTUndo}
	case 'r':
		return TalkToInput{Action: TTReturnToPerson}
	case 'q':
		return TalkToInput{Action: TTQuit}
	}

	return TalkToInput{Action: TTNoAction}
}

// HandleTalkToInput processes the parsed input and updates state
// Returns (shouldExit, message, error)
func HandleTalkToInput(state *data.TalkToViewState, input TalkToInput) (bool, string, error) {
	switch input.Action {
	case TTNoAction:
		return false, "", nil

	case TTNavigateNext:
		state.SelectNext()
		return false, "", nil

	case TTNavigatePrevious:
		state.SelectPrevious()
		return false, "", nil

	case TTConfirm:
		return handleConfirm(state)

	case TTBack:
		return handleBack(state)

	case TTQuit:
		return true, "", nil

	case TTToggleSelection:
		state.ToggleCurrentSelection()
		return false, "", nil

	case TTSelectAll:
		state.SelectAll()
		return false, "", nil

	case TTSelectNone:
		state.SelectNone()
		return false, "", nil

	case TTFindNote:
		return handleFindNote(state)

	case TTCreateNewNote:
		return handleCreateNewNote(state)

	case TTSearchType:
		if state.ViewMode == data.NoteSearchModalView && state.SearchMode == data.InsertMode {
			err := state.AppendToSearchQuery(input.Char)
			if err != nil {
				return false, "", err
			}
		}
		return false, "", nil

	case TTSearchBackspace:
		if state.ViewMode == data.NoteSearchModalView && state.SearchMode == data.InsertMode {
			err := state.BackspaceSearchQuery()
			if err != nil {
				return false, "", err
			}
		}
		return false, "", nil

	case TTSearchToggleMode:
		if state.ViewMode == data.NoteSearchModalView {
			state.ToggleSearchMode()
		}
		return false, "", nil

	case TTSearchEnterInsert:
		if state.ViewMode == data.NoteSearchModalView {
			state.EnterInsertMode()
		}
		return false, "", nil

	case TTSearchNavigateNext:
		if state.ViewMode == data.NoteSearchModalView {
			state.SelectNext()
		}
		return false, "", nil

	case TTSearchNavigatePrev:
		if state.ViewMode == data.NoteSearchModalView {
			state.SelectPrevious()
		}
		return false, "", nil

	case TTSearchConfirm:
		if state.ViewMode == data.NoteSearchModalView {
			err := state.SelectSearchResult()
			if err != nil {
				return false, fmt.Sprintf("Error: %v", err), nil
			}
		}
		return false, "", nil

	case TTSearchCancel:
		if state.ViewMode == data.NoteSearchModalView {
			state.CancelSearch()
		}
		return false, "", nil

	case TTUndo:
		if state.ViewMode == data.SuccessView {
			err := state.UndoLastMove()
			if err != nil {
				return false, fmt.Sprintf("Undo failed: %v", err), nil
			}
			// After undo, return to person selection
			err = state.BackToPersonSelection()
			if err != nil {
				return false, "", err
			}
		}
		return false, "", nil

	case TTOpenNote:
		if state.ViewMode == data.SuccessView {
			// Signal to open note (handled in main.go)
			return false, "OPEN_NOTE:" + state.TargetNoteName, nil
		}
		return false, "", nil

	case TTReturnToPerson:
		if state.ViewMode == data.SuccessView {
			err := state.BackToPersonSelection()
			if err != nil {
				return false, "", err
			}
		}
		return false, "", nil

	default:
		return false, "", nil
	}
}

// handleConfirm handles the Enter/Confirm action based on current view
func handleConfirm(state *data.TalkToViewState) (bool, string, error) {
	switch state.ViewMode {
	case data.PersonSelectionView:
		// Select person and enter todo selection
		selectedPerson := state.GetSelectedPerson()
		if selectedPerson == nil {
			return false, "No person selected", nil
		}

		err := state.EnterTodoSelection(state.PersonIndex)
		if err != nil {
			return false, "", err
		}
		return false, "", nil

	case data.TodoSelectionView:
		// Enter note selection if at least one todo is selected
		if state.GetSelectedCount() == 0 {
			return false, "No todos selected", nil
		}

		err := state.EnterNoteSelection()
		if err != nil {
			return false, "", err
		}
		return false, "", nil

	case data.ConfirmationView:
		// Execute the move
		err := state.ExecuteMove()
		if err != nil {
			return false, fmt.Sprintf("Move failed: %v", err), nil
		}
		return false, "", nil

	default:
		return false, "", nil
	}
}

// handleBack handles the back/cancel action based on current view
func handleBack(state *data.TalkToViewState) (bool, string, error) {
	switch state.ViewMode {
	case data.TodoSelectionView:
		// Go back to person selection
		err := state.BackToPersonSelection()
		if err != nil {
			return false, "", err
		}
		return false, "", nil

	case data.NoteSelectionView:
		// Go back to todo selection
		state.BackToTodoSelection()
		return false, "", nil

	case data.ConfirmationView:
		// Go back to todo selection
		state.BackToTodoSelection()
		return false, "", nil

	default:
		// In person selection, quit
		return true, "", nil
	}
}

// handleFindNote opens the search modal
func handleFindNote(state *data.TalkToViewState) (bool, string, error) {
	err := state.EnterSearchModal()
	if err != nil {
		return false, "", fmt.Errorf("failed to enter search modal: %w", err)
	}
	return false, "", nil
}

// handleCreateNewNote prompts for new note creation
func handleCreateNewNote(state *data.TalkToViewState) (bool, string, error) {
	return false, "CREATE_NEW_NOTE:", nil
}
