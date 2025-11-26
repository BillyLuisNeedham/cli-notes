package presentation

import (
	"cli-notes/scripts"
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
	MoveTodoToDay
	Undo
	Redo
	Save
	Reset
	Quit
	NextDay
	PreviousDay
	OpenTodo
	PreviousWeek
	NextWeek
	ToggleExpandedEarlier
	ExitExpandedView
	BulkMoveEarlier
	SetPriority1
	SetPriority2
	SetPriority3
	CreateTodo
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
	case keyboard.KeyEsc:
		return WeekPlannerInput{Action: ExitExpandedView}
	case keyboard.KeyCtrlS:
		return WeekPlannerInput{Action: Save}
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

	// Week navigation
	case '[':
		return WeekPlannerInput{Action: PreviousWeek}
	case ']':
		return WeekPlannerInput{Action: NextWeek}

	// Commands
	case 'u':
		return WeekPlannerInput{Action: Undo}
	case 'x':
		return WeekPlannerInput{Action: Reset}
	case 'q':
		return WeekPlannerInput{Action: Quit}
	case 'n':
		return WeekPlannerInput{Action: CreateTodo}
	case 'N':
		return WeekPlannerInput{Action: MoveToNextMonday}
	case 'e':
		return WeekPlannerInput{Action: ToggleExpandedEarlier}
	case 'b':
		return WeekPlannerInput{Action: BulkMoveEarlier}

	// Day shortcuts (lowercase = move todo to day)
	case 'm':
		return WeekPlannerInput{Action: MoveTodoToDay, Day: data.Monday}
	case 't':
		return WeekPlannerInput{Action: MoveTodoToDay, Day: data.Tuesday}
	case 'w':
		return WeekPlannerInput{Action: MoveTodoToDay, Day: data.Wednesday}
	case 'r':
		return WeekPlannerInput{Action: MoveTodoToDay, Day: data.Thursday}
	case 'f':
		return WeekPlannerInput{Action: MoveTodoToDay, Day: data.Friday}
	case 'a':
		return WeekPlannerInput{Action: MoveTodoToDay, Day: data.Saturday}
	case 's':
		return WeekPlannerInput{Action: MoveTodoToDay, Day: data.Sunday}

	// Priority shortcuts (1/2/3)
	case '1':
		return WeekPlannerInput{Action: SetPriority1}
	case '2':
		return WeekPlannerInput{Action: SetPriority2}
	case '3':
		return WeekPlannerInput{Action: SetPriority3}
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
		// Special handling for expanded Earlier view
		if state.ViewMode == data.ExpandedEarlierView && state.SelectedDay == data.Earlier {
			err := state.MoveEarlierTodoToMonday()
			if err != nil {
				return false, err.Error(), nil
			}
			return false, "Moved todo to Monday", nil
		}
		// Normal move right behavior
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

	case MoveTodoToDay:
		selectedTodo := state.GetSelectedTodo()
		if selectedTodo == nil {
			return false, "No todo selected", nil
		}

		sourceDay := state.SelectedDay
		targetDay := input.Day

		// Don't move if already on target day
		if sourceDay == targetDay {
			return false, "", nil
		}

		// Perform the move
		state.Plan.MoveTodo(*selectedTodo, sourceDay, targetDay)

		// Special handling for ExpandedEarlierView when moving from Earlier
		if state.ViewMode == data.ExpandedEarlierView && sourceDay == data.Earlier {
			// Stay in expanded view, stay on Earlier day
			// Just adjust selection index since a todo was removed from Earlier list
			state.AdjustSelectionAfterMove()
			return false, fmt.Sprintf("Moved todo to %s", data.WeekDayNames[targetDay]), nil
		}

		// Normal behavior: Exit expanded view if we're in it before switching days
		if state.ViewMode == data.ExpandedEarlierView {
			state.ExitExpandedEarlierView()
		}

		// Stay on the current day and adjust selection after the move
		state.AdjustSelectionAfterMove()

		return false, fmt.Sprintf("Moved todo to %s", data.WeekDayNames[targetDay]), nil

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

	case PreviousWeek:
		err := state.NavigateToPreviousWeek()
		if err != nil {
			return false, "", err
		}
		return false, "Navigated to previous week", nil

	case NextWeek:
		err := state.NavigateToNextWeek()
		if err != nil {
			return false, "", err
		}
		return false, "Navigated to next week", nil

	case ToggleExpandedEarlier:
		if state.ViewMode == data.NormalView {
			state.EnterExpandedEarlierView()
			return false, "Expanded Earlier view", nil
		} else if state.ViewMode == data.ExpandedEarlierView {
			state.ExitExpandedEarlierView()
			return false, "Returned to normal view", nil
		}
		return false, "", nil

	case ExitExpandedView:
		if state.ViewMode == data.ExpandedEarlierView {
			state.ExitExpandedEarlierView()
			return false, "Returned to normal view", nil
		}
		return false, "", nil

	case SetPriority1, SetPriority2, SetPriority3:
		selectedTodo := state.GetSelectedTodo()
		if selectedTodo == nil {
			return false, "No todo selected", nil
		}

		var priority scripts.Priority
		switch input.Action {
		case SetPriority1:
			priority = scripts.P1
		case SetPriority2:
			priority = scripts.P2
		case SetPriority3:
			priority = scripts.P3
		}

		err := state.ChangeTodoPriority(selectedTodo, priority)
		if err != nil {
			return false, fmt.Sprintf("Error changing priority: %v", err), nil
		}

		return false, fmt.Sprintf("Priority changed to P%d", priority), nil

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

// ParseSwitchToDay parses a single uppercase character into a WeekDay for switching view
// Returns the target day and true if valid, or -1 and false if invalid
func ParseSwitchToDay(char rune) (data.WeekDay, bool) {
	switch char {
	case 'M':
		return data.Monday, true
	case 'T':
		return data.Tuesday, true
	case 'W':
		return data.Wednesday, true
	case 'R':
		return data.Thursday, true
	case 'F':
		return data.Friday, true
	case 'A':
		return data.Saturday, true
	case 'S':
		return data.Sunday, true
	default:
		return -1, false
	}
}

// IsSwitchDayKey checks if the character is a valid switch-day command
func IsSwitchDayKey(char rune) bool {
	return char == 'M' || char == 'T' || char == 'W' ||
		char == 'R' || char == 'F' || char == 'A' || char == 'S'
}
