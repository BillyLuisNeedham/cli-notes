package input

import (
	"bufio"

	"github.com/eiannone/keyboard"
)

// InputReader abstracts keyboard input for testability
type InputReader interface {
	GetKey() (char rune, key keyboard.Key, err error)
}

// KeyboardReader reads from physical keyboard (production)
type KeyboardReader struct{}

func (k *KeyboardReader) GetKey() (rune, keyboard.Key, error) {
	return keyboard.GetKey()
}

// StdinReader reads from buffered stdin (testing)
type StdinReader struct {
	reader *bufio.Reader
}

func NewStdinReader(reader *bufio.Reader) *StdinReader {
	return &StdinReader{reader: reader}
}

func (s *StdinReader) GetKey() (rune, keyboard.Key, error) {
	r, _, err := s.reader.ReadRune()
	if err != nil {
		return 0, 0, err
	}

	char, key := mapRuneToKeyboard(r, s.reader)
	return char, key, nil
}

// mapRuneToKeyboard converts stdin runes to keyboard key constants
func mapRuneToKeyboard(r rune, reader *bufio.Reader) (char rune, key keyboard.Key) {
	switch r {
	case '\n':
		key = keyboard.KeyEnter
	case '\x13': // Ctrl+S
		key = keyboard.KeyCtrlS
	case '\x0E': // Ctrl+N
		key = keyboard.KeyCtrlN
	case '\x14': // Ctrl+T
		key = keyboard.KeyCtrlT
	case '\x17': // Ctrl+W
		key = keyboard.KeyCtrlW
	case '\x12': // Ctrl+R
		key = keyboard.KeyCtrlR
	case '\x06': // Ctrl+F
		key = keyboard.KeyCtrlF
	case '\x01': // Ctrl+A
		key = keyboard.KeyCtrlA
	case '\x15': // Ctrl+U
		key = keyboard.KeyCtrlU
	case '\x7f': // Backspace
		key = keyboard.KeyBackspace
	case '\t': // Tab
		key = keyboard.KeyTab
	case '\x1b': // Escape or ANSI sequence
		if reader.Buffered() > 0 {
			next, _, _ := reader.ReadRune()
			if next == '[' {
				// Arrow keys and special sequences
				dir, _, _ := reader.ReadRune()
				switch dir {
				case 'A':
					key = keyboard.KeyArrowUp
				case 'B':
					key = keyboard.KeyArrowDown
				case 'C':
					key = keyboard.KeyArrowRight
				case 'D':
					key = keyboard.KeyArrowLeft
				}
			} else {
				// Just Escape
				key = keyboard.KeyEsc
			}
		} else {
			key = keyboard.KeyEsc
		}
	default:
		char = r
	}
	return
}
