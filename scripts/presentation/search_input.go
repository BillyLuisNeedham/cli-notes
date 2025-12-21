package presentation

import (
	"cli-notes/scripts/data"

	"github.com/eiannone/keyboard"
)

type SearchAction int

const (
	SearchNoAction SearchAction = iota
	SearchAddChar               // Add character to query
	SearchDeleteChar            // Backspace
	SearchNavigateUp            // k or arrow up
	SearchNavigateDown          // j or arrow down
	SearchSelect                // Enter - open actions menu or execute action
	SearchOpenNote              // o - directly open in editor
	SearchQuit                  // q or Esc - exit search
	SearchClearQuery            // Ctrl+U - clear query
	SearchEnterInsert           // i or / - enter insert mode
	SearchEnterNormal           // Esc from insert - enter normal mode

	// Direct actions (from normal mode)
	SearchSetPriority1  // 1 key
	SearchSetPriority2  // 2 key
	SearchSetPriority3  // 3 key
	SearchToggleDone    // d key
	SearchSetDueToday   // t key
	SearchLinkNote      // l key - link to another note
	SearchLinkObjective // L key - link to objective
)

type SearchInput struct {
	Action SearchAction
	Char   rune // For SearchAddChar
}

// ParseSearchInput parses keyboard input for search view
func ParseSearchInput(char rune, key keyboard.Key, mode data.SearchViewMode) SearchInput {
	switch mode {
	case data.SearchModeInsert:
		return parseInsertModeInput(char, key)
	case data.SearchModeNormal:
		return parseNormalModeInput(char, key)
	case data.SearchModeActions:
		return parseActionsModeInput(char, key)
	default:
		return SearchInput{Action: SearchNoAction}
	}
}

// parseInsertModeInput handles input in insert mode (typing query)
func parseInsertModeInput(char rune, key keyboard.Key) SearchInput {
	switch key {
	case keyboard.KeyEsc:
		return SearchInput{Action: SearchEnterNormal}
	case keyboard.KeyEnter:
		return SearchInput{Action: SearchEnterNormal} // Enter also exits insert mode
	case keyboard.KeyBackspace, keyboard.KeyBackspace2:
		return SearchInput{Action: SearchDeleteChar}
	case keyboard.KeyCtrlU:
		return SearchInput{Action: SearchClearQuery}
	case keyboard.KeyArrowUp:
		return SearchInput{Action: SearchNavigateUp}
	case keyboard.KeyArrowDown:
		return SearchInput{Action: SearchNavigateDown}
	}

	// All printable characters go to query in insert mode
	if char >= 32 && char < 127 {
		return SearchInput{Action: SearchAddChar, Char: char}
	}

	return SearchInput{Action: SearchNoAction}
}

// parseNormalModeInput handles input in normal/command mode
func parseNormalModeInput(char rune, key keyboard.Key) SearchInput {
	// Handle special keys
	switch key {
	case keyboard.KeyEsc:
		return SearchInput{Action: SearchQuit}
	case keyboard.KeyEnter:
		return SearchInput{Action: SearchSelect}
	case keyboard.KeyArrowUp:
		return SearchInput{Action: SearchNavigateUp}
	case keyboard.KeyArrowDown:
		return SearchInput{Action: SearchNavigateDown}
	case keyboard.KeyBackspace, keyboard.KeyBackspace2:
		return SearchInput{Action: SearchDeleteChar}
	}

	// Handle character commands
	switch char {
	case 'i', '/':
		return SearchInput{Action: SearchEnterInsert}
	case 'j':
		return SearchInput{Action: SearchNavigateDown}
	case 'k':
		return SearchInput{Action: SearchNavigateUp}
	case 'o':
		return SearchInput{Action: SearchOpenNote}
	case 'q':
		return SearchInput{Action: SearchQuit}
	case '1':
		return SearchInput{Action: SearchSetPriority1}
	case '2':
		return SearchInput{Action: SearchSetPriority2}
	case '3':
		return SearchInput{Action: SearchSetPriority3}
	case 'd':
		return SearchInput{Action: SearchToggleDone}
	case 't':
		return SearchInput{Action: SearchSetDueToday}
	case 'l':
		return SearchInput{Action: SearchLinkNote}
	case 'L':
		return SearchInput{Action: SearchLinkObjective}
	}

	return SearchInput{Action: SearchNoAction}
}

// parseActionsModeInput handles input in actions menu mode
func parseActionsModeInput(char rune, key keyboard.Key) SearchInput {
	switch key {
	case keyboard.KeyEnter:
		return SearchInput{Action: SearchSelect}
	case keyboard.KeyEsc:
		return SearchInput{Action: SearchEnterNormal}
	case keyboard.KeyArrowUp:
		return SearchInput{Action: SearchNavigateUp}
	case keyboard.KeyArrowDown:
		return SearchInput{Action: SearchNavigateDown}
	}

	switch char {
	case 'j':
		return SearchInput{Action: SearchNavigateDown}
	case 'k':
		return SearchInput{Action: SearchNavigateUp}
	case 'q':
		return SearchInput{Action: SearchEnterNormal}
	}

	return SearchInput{Action: SearchNoAction}
}
