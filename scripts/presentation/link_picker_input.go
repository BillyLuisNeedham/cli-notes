package presentation

import (
	"cli-notes/scripts/data"

	"github.com/eiannone/keyboard"
)

type LinkPickerAction int

const (
	LinkPickerNoAction LinkPickerAction = iota
	LinkPickerAddChar                   // Add character to query
	LinkPickerDeleteChar                // Backspace
	LinkPickerNavigateUp                // k or arrow up
	LinkPickerNavigateDown              // j or arrow down
	LinkPickerSelect                    // Enter - select current note
	LinkPickerCancel                    // q or Esc - cancel selection
	LinkPickerEnterInsert               // i or / - enter insert mode
	LinkPickerEnterNormal               // Esc from insert - enter normal mode
)

type LinkPickerInput struct {
	Action LinkPickerAction
	Char   rune // For LinkPickerAddChar
}

// ParseLinkPickerInput parses keyboard input for link picker view
func ParseLinkPickerInput(char rune, key keyboard.Key, mode data.LinkPickerMode) LinkPickerInput {
	switch mode {
	case data.LinkPickerModeInsert:
		return parseLinkPickerInsertInput(char, key)
	case data.LinkPickerModeNormal:
		return parseLinkPickerNormalInput(char, key)
	default:
		return LinkPickerInput{Action: LinkPickerNoAction}
	}
}

// parseLinkPickerInsertInput handles input in insert mode (typing query)
func parseLinkPickerInsertInput(char rune, key keyboard.Key) LinkPickerInput {
	switch key {
	case keyboard.KeyEsc:
		return LinkPickerInput{Action: LinkPickerEnterNormal}
	case keyboard.KeyEnter:
		return LinkPickerInput{Action: LinkPickerEnterNormal} // Enter also exits insert mode
	case keyboard.KeyBackspace, keyboard.KeyBackspace2:
		return LinkPickerInput{Action: LinkPickerDeleteChar}
	case keyboard.KeyArrowUp:
		return LinkPickerInput{Action: LinkPickerNavigateUp}
	case keyboard.KeyArrowDown:
		return LinkPickerInput{Action: LinkPickerNavigateDown}
	}

	// All printable characters go to query in insert mode
	if char >= 32 && char < 127 {
		return LinkPickerInput{Action: LinkPickerAddChar, Char: char}
	}

	return LinkPickerInput{Action: LinkPickerNoAction}
}

// parseLinkPickerNormalInput handles input in normal/command mode
func parseLinkPickerNormalInput(char rune, key keyboard.Key) LinkPickerInput {
	// Handle special keys
	switch key {
	case keyboard.KeyEsc:
		return LinkPickerInput{Action: LinkPickerCancel}
	case keyboard.KeyEnter:
		return LinkPickerInput{Action: LinkPickerSelect}
	case keyboard.KeyArrowUp:
		return LinkPickerInput{Action: LinkPickerNavigateUp}
	case keyboard.KeyArrowDown:
		return LinkPickerInput{Action: LinkPickerNavigateDown}
	case keyboard.KeyBackspace, keyboard.KeyBackspace2:
		return LinkPickerInput{Action: LinkPickerDeleteChar}
	}

	// Handle character commands
	switch char {
	case 'i', '/':
		return LinkPickerInput{Action: LinkPickerEnterInsert}
	case 'j':
		return LinkPickerInput{Action: LinkPickerNavigateDown}
	case 'k':
		return LinkPickerInput{Action: LinkPickerNavigateUp}
	case 'q':
		return LinkPickerInput{Action: LinkPickerCancel}
	}

	return LinkPickerInput{Action: LinkPickerNoAction}
}
