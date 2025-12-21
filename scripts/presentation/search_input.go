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
	SearchQuit                  // q or Esc - exit search or actions
	SearchClearQuery            // Ctrl+U - clear query

	// Direct actions (from results view)
	SearchSetPriority1 // 1 key
	SearchSetPriority2 // 2 key
	SearchSetPriority3 // 3 key
	SearchToggleDone   // d key
	SearchSetDueToday  // t key
	SearchSetDueMonday // m key
)

type SearchInput struct {
	Action SearchAction
	Char   rune // For SearchAddChar
}

// ParseSearchInput parses keyboard input for search view
func ParseSearchInput(char rune, key keyboard.Key, mode data.SearchViewMode) SearchInput {
	// Handle special keys first
	switch key {
	case keyboard.KeyEnter:
		return SearchInput{Action: SearchSelect}
	case keyboard.KeyEsc:
		return SearchInput{Action: SearchQuit}
	case keyboard.KeyBackspace, keyboard.KeyBackspace2:
		return SearchInput{Action: SearchDeleteChar}
	case keyboard.KeyCtrlU:
		return SearchInput{Action: SearchClearQuery}
	case keyboard.KeyArrowUp:
		return SearchInput{Action: SearchNavigateUp}
	case keyboard.KeyArrowDown:
		return SearchInput{Action: SearchNavigateDown}
	}

	// In actions mode, only handle navigation and selection
	if mode == data.SearchModeActions {
		switch char {
		case 'j':
			return SearchInput{Action: SearchNavigateDown}
		case 'k':
			return SearchInput{Action: SearchNavigateUp}
		case 'q':
			return SearchInput{Action: SearchQuit}
		default:
			return SearchInput{Action: SearchNoAction}
		}
	}

	// In typing mode, handle character input and shortcuts
	switch char {
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
	case 'm':
		return SearchInput{Action: SearchSetDueMonday}
	default:
		// Any other printable character is added to the query
		if char >= 32 && char < 127 {
			return SearchInput{Action: SearchAddChar, Char: char}
		}
		return SearchInput{Action: SearchNoAction}
	}
}

// IsNavigationKey returns true if the character is a navigation key (j/k)
// This helps determine if we should add to query or navigate
func IsNavigationKey(char rune) bool {
	return char == 'j' || char == 'k'
}

// IsShortcutKey returns true if the character is a shortcut key
func IsShortcutKey(char rune) bool {
	switch char {
	case 'o', 'q', '1', '2', '3', 'd', 't', 'm':
		return true
	default:
		return false
	}
}
