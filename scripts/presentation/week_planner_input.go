package presentation

import (
	"cli-notes/scripts/data"
	"fmt"

	"github.com/eiannone/keyboard"
)

// WeekPlannerAction represents an action to take in the week planner
type WeekPlannerAction int

const (
	NoAction WeekPlannerAction = iota
	MoveLeft
	MoveRight
	MoveToNextMonday
	SelectUp
	SelectDown
	SwitchDay
	Undo
	Redo
	Save
	Reset
	Quit
	NextDay
	PreviousDay
	OpenTodo
)

// WeekPlannerInput represents a parsed input from the keyboard
type WeekPlannerInput struct {
	Action WeekPlannerAction
	Day    data.WeekDay // Used when Action is SwitchDay
}

// ParseWeekPlannerInput reads keyboard input and returns the corresponding action
func ParseWeekPlannerInput(char rune, key keyboard.Key) WeekPlannerInput {
	// Handle special keys
	switch key {
	case keyboard.KeyEnter:
		return WeekPlannerInput{Action: OpenTodo}
	case keyboard.KeyTab:
		return WeekPlannerInput{Action: NextDay}
	}

	// Handle character commands (including vim bindings)
	switch char {
	// Vim bindings for navigation
	case 'j':
		return WeekPlannerInput{Action: SelectDown}
	case 'k':
		return WeekPlannerInput{Action: SelectUp}
	case 'h':
		return WeekPlannerInput{Action: MoveLeft}
	case 'l':
		return WeekPlannerInput{Action: MoveRight}

	// Commands
	case 'u':
		return WeekPlannerInput{Action: Undo}
	case 'r':
		return WeekPlannerInput{Action: Redo}
	case 's':
		return WeekPlannerInput{Action: Save}
	case 'x':
		return WeekPlannerInput{Action: Reset}
	case 'q':
		return WeekPlannerInput{Action: Quit}
	case 'n':
		return WeekPlannerInput{Action: MoveToNextMonday}

	// Day shortcuts
	case 'm':
		return WeekPlannerInput{Action: SwitchDay, Day: data.Monday}
	}

	// Handle two-character day shortcuts
	// Note: This will need to be enhanced to handle multi-char input
	// For now, we'll use single char mappings
	return WeekPlannerInput{Action: NoAction}
}

// HandleWeekPlannerInput processes the input and updates the state
func HandleWeekPlannerInput(state *data.WeekPlannerState, input WeekPlannerInput) (shouldExit bool, message string, err error) {
	switch input.Action {
	case SelectUp:
		state.SelectPreviousTodo()
		return false, "", nil

	case SelectDown:
		state.SelectNextTodo()
		return false, "", nil

	case MoveLeft:
		err := state.MoveSelectedTodoLeft()
		if err != nil {
			return false, err.Error(), nil
		}
		return false, "Moved todo to previous day", nil

	case MoveRight:
		err := state.MoveSelectedTodoRight()
		if err != nil {
			return false, err.Error(), nil
		}
		return false, "Moved todo to next day", nil

	case MoveToNextMonday:
		err := state.MoveSelectedTodoToNextMonday()
		if err != nil {
			return false, err.Error(), nil
		}
		return false, "Moved todo to next Monday", nil

	case SwitchDay:
		state.SwitchToDay(input.Day)
		return false, fmt.Sprintf("Switched to %s", data.WeekDayNames[input.Day]), nil

	case NextDay:
		state.SwitchToNextDay()
		return false, "", nil

	case PreviousDay:
		state.SwitchToPreviousDay()
		return false, "", nil

	case Undo:
		success := state.Undo()
		if !success {
			return false, "Nothing to undo", nil
		}
		return false, "Undone", nil

	case Redo:
		success := state.Redo()
		if !success {
			return false, "Nothing to redo", nil
		}
		return false, "Redone", nil

	case Save:
		err := state.Save()
		if err != nil {
			return false, "", err
		}
		return false, "Changes saved successfully", nil

	case Reset:
		err := state.Reset()
		if err != nil {
			return false, "", err
		}
		return false, "Plan reset from disk", nil

	case Quit:
		return true, "", nil

	case NoAction:
		return false, "", nil

	default:
		return false, "", nil
	}
}

// ParseMultiCharCommand handles multi-character commands like "tu", "th", etc.
// This function should be called when building up a command buffer
func ParseMultiCharCommand(command string) (data.WeekDay, bool) {
	switch command {
	case "m":
		return data.Monday, true
	case "tu":
		return data.Tuesday, true
	case "w":
		return data.Wednesday, true
	case "th":
		return data.Thursday, true
	case "f":
		return data.Friday, true
	case "sa":
		return data.Saturday, true
	case "su":
		return data.Sunday, true
	default:
		return -1, false
	}
}
