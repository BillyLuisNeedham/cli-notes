package presentation

import (
	"github.com/eiannone/keyboard"
)

type ObjectivesAction int

const (
	ObjNoAction ObjectivesAction = iota
	ObjNavigateNext
	ObjNavigatePrevious
	ObjOpenSelected
	ObjCreateNew
	ObjLinkExisting
	ObjDelete
	ObjQuit
	ObjEditParent
	ObjUnlinkChild
	ObjChangeSort
	ObjChangeFilter
	ObjBack
	ObjSetPriority1
	ObjSetPriority2
	ObjSetPriority3
	ObjSetDueToday
	ObjSetDueDate
	ObjSetDueMonday
	ObjSetDueTuesday
	ObjSetDueWednesday
	ObjSetDueThursday
	ObjSetDueFriday
	ObjSetDueSaturday
	ObjSetDueSunday
)

type ObjectivesInput struct {
	Action ObjectivesAction
	Char   rune // Store the character for multi-key commands
}

// ParseObjectivesInput parses keyboard input for objectives views
func ParseObjectivesInput(char rune, key keyboard.Key) ObjectivesInput {
	switch key {
	case keyboard.KeyEnter:
		return ObjectivesInput{Action: ObjOpenSelected}
	case keyboard.KeyEsc:
		return ObjectivesInput{Action: ObjQuit}
	}

	switch char {
	case 'j':
		return ObjectivesInput{Action: ObjNavigateNext}
	case 'k':
		return ObjectivesInput{Action: ObjNavigatePrevious}
	case 'o':
		return ObjectivesInput{Action: ObjOpenSelected}
	case 'n':
		return ObjectivesInput{Action: ObjCreateNew}
	case 'l':
		return ObjectivesInput{Action: ObjLinkExisting}
	case 'd':
		// d key - could be start of 'dd' for delete or 'd' for set due date
		// Caller needs to track this
		return ObjectivesInput{Action: ObjDelete, Char: 'd'}
	case 'q':
		return ObjectivesInput{Action: ObjQuit}
	case 'e':
		return ObjectivesInput{Action: ObjEditParent}
	case 'u':
		return ObjectivesInput{Action: ObjUnlinkChild}
	case 's':
		return ObjectivesInput{Action: ObjChangeSort}
	case 'f':
		return ObjectivesInput{Action: ObjChangeFilter}
	case 'p':
		// p key for priority - caller needs to get next digit
		return ObjectivesInput{Action: ObjNoAction, Char: 'p'}
	case '1':
		return ObjectivesInput{Action: ObjSetPriority1}
	case '2':
		return ObjectivesInput{Action: ObjSetPriority2}
	case '3':
		return ObjectivesInput{Action: ObjSetPriority3}
	case 't':
		return ObjectivesInput{Action: ObjSetDueToday}
	case 'm':
		return ObjectivesInput{Action: ObjSetDueMonday}
	case 'w':
		return ObjectivesInput{Action: ObjSetDueWednesday}
	case 'r':
		return ObjectivesInput{Action: ObjSetDueThursday} // th(u)rsday
	case 'h':
		return ObjectivesInput{Action: ObjSetDueThursday}
	default:
		return ObjectivesInput{Action: ObjNoAction}
	}
}

// ParseDayInput parses two-letter day abbreviations
func ParseDayInput(first, second rune) ObjectivesAction {
	combined := string([]rune{first, second})

	switch combined {
	case "tu":
		return ObjSetDueTuesday
	case "th":
		return ObjSetDueThursday
	case "sa":
		return ObjSetDueSaturday
	case "su":
		return ObjSetDueSunday
	default:
		return ObjNoAction
	}
}
