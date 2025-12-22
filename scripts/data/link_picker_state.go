package data

import (
	"cli-notes/scripts"
	"strings"

	"github.com/sahilm/fuzzy"
)

type LinkPickerMode int

const (
	LinkPickerModeInsert LinkPickerMode = iota // User is typing search query
	LinkPickerModeNormal                       // Command mode (j/k navigate)
)

// LinkPickerState holds state for the link picker view
type LinkPickerState struct {
	Mode          LinkPickerMode
	Query         string
	AllNotes      []scripts.File
	FilteredNotes []scripts.File
	SelectedIndex int
}

// NewLinkPickerState initializes link picker state
func NewLinkPickerState(notes []scripts.File) *LinkPickerState {
	state := &LinkPickerState{
		Mode:          LinkPickerModeInsert, // Start in insert mode for immediate typing
		Query:         "",
		AllNotes:      notes,
		FilteredNotes: notes,
		SelectedIndex: 0,
	}
	return state
}

// EnterInsertMode switches to insert mode for typing
func (s *LinkPickerState) EnterInsertMode() {
	s.Mode = LinkPickerModeInsert
}

// EnterNormalMode switches to normal/command mode
func (s *LinkPickerState) EnterNormalMode() {
	s.Mode = LinkPickerModeNormal
}

// SelectNext moves selection down
func (s *LinkPickerState) SelectNext() {
	if len(s.FilteredNotes) > 0 {
		s.SelectedIndex = (s.SelectedIndex + 1) % len(s.FilteredNotes)
	}
}

// SelectPrevious moves selection up
func (s *LinkPickerState) SelectPrevious() {
	if len(s.FilteredNotes) > 0 {
		s.SelectedIndex = (s.SelectedIndex - 1 + len(s.FilteredNotes)) % len(s.FilteredNotes)
	}
}

// AddChar adds a character to the query and updates filtered results
func (s *LinkPickerState) AddChar(c rune) {
	s.Query += string(c)
	s.updateFilter()
}

// DeleteChar removes the last character from the query
func (s *LinkPickerState) DeleteChar() {
	if len(s.Query) > 0 {
		runes := []rune(s.Query)
		s.Query = string(runes[:len(runes)-1])
		s.updateFilter()
	}
}

// updateFilter applies fuzzy search and updates filtered notes
func (s *LinkPickerState) updateFilter() {
	s.SelectedIndex = 0

	if s.Query == "" {
		s.FilteredNotes = s.AllNotes
		return
	}

	// Build search targets combining title and tags
	searchTargets := make([]string, len(s.AllNotes))
	for i, note := range s.AllNotes {
		searchTargets[i] = strings.ToLower(note.Title + " " + strings.Join(note.Tags, " "))
	}

	queryLower := strings.ToLower(s.Query)
	matches := fuzzy.Find(queryLower, searchTargets)

	s.FilteredNotes = make([]scripts.File, len(matches))
	for i, match := range matches {
		s.FilteredNotes[i] = s.AllNotes[match.Index]
	}
}

// GetSelectedNote returns the currently selected note
func (s *LinkPickerState) GetSelectedNote() *scripts.File {
	if len(s.FilteredNotes) == 0 || s.SelectedIndex >= len(s.FilteredNotes) {
		return nil
	}
	selected := s.FilteredNotes[s.SelectedIndex]
	return &selected
}

// ClampSelectedIndex ensures selected index is within bounds
func (s *LinkPickerState) ClampSelectedIndex() {
	if s.SelectedIndex >= len(s.FilteredNotes) {
		s.SelectedIndex = len(s.FilteredNotes) - 1
	}
	if s.SelectedIndex < 0 {
		s.SelectedIndex = 0
	}
}
