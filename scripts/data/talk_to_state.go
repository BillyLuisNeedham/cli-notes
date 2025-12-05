package data

import (
	"cli-notes/scripts"
	"fmt"
	"sort"
	"strings"
	"time"
)

// TalkToViewMode represents the current view in the Talk-To interface
type TalkToViewMode int

const (
	PersonSelectionView TalkToViewMode = iota
	TodoSelectionView
	NoteSelectionView
	NoteSearchModalView
	ConfirmationView
	SuccessView
)

// SearchMode represents the input mode in the note search modal
type SearchMode int

const (
	InsertMode SearchMode = iota
	NormalMode
)

// PersonInfo contains information about a person with pending to-talk todos
type PersonInfo struct {
	Name  string // Normalized lowercase name
	Count int    // Number of incomplete todos
}

// TodoWithMeta represents a todo with metadata for the talk-to feature
type TodoWithMeta struct {
	File       scripts.File
	TodoLine   string                // Full line with checkbox
	LineNumber int                   // 1-indexed line number in the file
	SourceFile string                // Filename
	Subtasks   []scripts.SubtaskInfo // Nested subtasks
}

// LineModification tracks a change to a specific line for rollback
type LineModification struct {
	LineNumber int
	OldContent string
	NewContent string
}

// MoveChange represents a batch move operation that can be undone
type MoveChange struct {
	Timestamp            time.Time
	Person               string
	Todos                []TodoWithMeta
	TargetNote           string
	TargetInsertionPoint int                              // Line number where todos were inserted
	SourceModifications  map[string][]LineModification    // Filename -> list of line changes
}

// TalkToViewState manages the state for the Talk-To interactive view
type TalkToViewState struct {
	ViewMode TalkToViewMode

	// Person selection
	AllPeople   []PersonInfo
	PersonIndex int

	// Todo selection
	SelectedPerson   string
	AvailableTodos   []TodoWithMeta
	SelectedTodos    []bool // Parallel array for checkbox state
	TodoIndex        int
	TodoScrollOffset int

	// Note search modal
	SearchMode    SearchMode
	SearchQuery   string
	SearchResults []scripts.File
	SearchIndex   int

	// Target note
	TargetNoteName string
	IsNewNote      bool

	// Undo tracking
	Changes   []MoveChange
	UndoStack []MoveChange

	LastMessage string
}

// NewTalkToViewState initializes a new Talk-To view state
// If filterPerson is provided, it auto-enters TodoSelectionView for that person
func NewTalkToViewState(filterPerson string) (*TalkToViewState, error) {
	// Scan all to-talk todos
	todosByPerson, err := ScanAllTalkToTodos()
	if err != nil {
		return nil, err
	}

	// Build AllPeople list
	people := make([]PersonInfo, 0)
	for person, todos := range todosByPerson {
		people = append(people, PersonInfo{
			Name:  person,
			Count: len(todos),
		})
	}

	// Sort alphabetically
	sort.Slice(people, func(i, j int) bool {
		return people[i].Name < people[j].Name
	})

	state := &TalkToViewState{
		ViewMode:    PersonSelectionView,
		AllPeople:   people,
		PersonIndex: 0,
		UndoStack:   []MoveChange{},
		Changes:     []MoveChange{},
	}

	// If filter person provided, auto-enter TodoSelectionView
	if filterPerson != "" {
		normalizedFilter := strings.ToLower(strings.TrimSpace(filterPerson))
		for i, person := range people {
			if person.Name == normalizedFilter {
				state.EnterTodoSelection(i)
				break
			}
		}
	}

	return state, nil
}

// SelectNext moves selection to the next item (with wrap-around)
func (s *TalkToViewState) SelectNext() {
	switch s.ViewMode {
	case PersonSelectionView:
		if len(s.AllPeople) > 0 {
			s.PersonIndex = (s.PersonIndex + 1) % len(s.AllPeople)
		}
	case TodoSelectionView:
		if len(s.AvailableTodos) > 0 {
			s.TodoIndex = (s.TodoIndex + 1) % len(s.AvailableTodos)
			s.adjustScrollOffset()
		}
	case NoteSearchModalView:
		if len(s.SearchResults) > 0 {
			s.SearchIndex = (s.SearchIndex + 1) % len(s.SearchResults)
		}
	}
}

// SelectPrevious moves selection to the previous item (with wrap-around)
func (s *TalkToViewState) SelectPrevious() {
	switch s.ViewMode {
	case PersonSelectionView:
		if len(s.AllPeople) > 0 {
			s.PersonIndex = (s.PersonIndex - 1 + len(s.AllPeople)) % len(s.AllPeople)
		}
	case TodoSelectionView:
		if len(s.AvailableTodos) > 0 {
			s.TodoIndex = (s.TodoIndex - 1 + len(s.AvailableTodos)) % len(s.AvailableTodos)
			s.adjustScrollOffset()
		}
	case NoteSearchModalView:
		if len(s.SearchResults) > 0 {
			s.SearchIndex = (s.SearchIndex - 1 + len(s.SearchResults)) % len(s.SearchResults)
		}
	}
}

// adjustScrollOffset keeps the selected todo visible in the viewport
func (s *TalkToViewState) adjustScrollOffset() {
	if s.TodoScrollOffset > s.TodoIndex {
		s.TodoScrollOffset = s.TodoIndex
	}
}

// EnterTodoSelection transitions to TodoSelectionView for a specific person
func (s *TalkToViewState) EnterTodoSelection(personIndex int) error {
	if personIndex < 0 || personIndex >= len(s.AllPeople) {
		return fmt.Errorf("invalid person index: %d", personIndex)
	}

	person := s.AllPeople[personIndex]
	s.SelectedPerson = person.Name

	// Load todos for this person
	todosByPerson, err := ScanAllTalkToTodos()
	if err != nil {
		return err
	}

	todos := todosByPerson[person.Name]
	s.AvailableTodos = todos

	// Initialize selection state (all selected by default)
	s.SelectedTodos = make([]bool, len(todos))
	for i := range s.SelectedTodos {
		s.SelectedTodos[i] = true
	}

	s.TodoIndex = 0
	s.TodoScrollOffset = 0
	s.ViewMode = TodoSelectionView

	return nil
}

// ToggleCurrentSelection toggles the checkbox for the current todo
func (s *TalkToViewState) ToggleCurrentSelection() {
	if s.ViewMode == TodoSelectionView && s.TodoIndex >= 0 && s.TodoIndex < len(s.SelectedTodos) {
		s.SelectedTodos[s.TodoIndex] = !s.SelectedTodos[s.TodoIndex]
	}
}

// SelectAll selects all todos in the current list
func (s *TalkToViewState) SelectAll() {
	if s.ViewMode == TodoSelectionView {
		for i := range s.SelectedTodos {
			s.SelectedTodos[i] = true
		}
	}
}

// SelectNone deselects all todos in the current list
func (s *TalkToViewState) SelectNone() {
	if s.ViewMode == TodoSelectionView {
		for i := range s.SelectedTodos {
			s.SelectedTodos[i] = false
		}
	}
}

// GetSelectedCount returns the number of selected todos
func (s *TalkToViewState) GetSelectedCount() int {
	count := 0
	for _, selected := range s.SelectedTodos {
		if selected {
			count++
		}
	}
	return count
}

// EnterNoteSelection transitions to NoteSelectionView
// Returns error if no todos are selected
func (s *TalkToViewState) EnterNoteSelection() error {
	if s.GetSelectedCount() == 0 {
		return fmt.Errorf("no todos selected")
	}
	s.ViewMode = NoteSelectionView
	return nil
}

// BackToPersonSelection returns to the PersonSelectionView
func (s *TalkToViewState) BackToPersonSelection() error {
	// Reload todos to reflect any changes
	todosByPerson, err := ScanAllTalkToTodos()
	if err != nil {
		return err
	}

	// Rebuild people list
	people := make([]PersonInfo, 0)
	for person, todos := range todosByPerson {
		people = append(people, PersonInfo{
			Name:  person,
			Count: len(todos),
		})
	}

	// Sort alphabetically
	sort.Slice(people, func(i, j int) bool {
		return people[i].Name < people[j].Name
	})

	s.AllPeople = people
	s.PersonIndex = 0
	s.ViewMode = PersonSelectionView

	return nil
}

// BackToTodoSelection returns to the TodoSelectionView
func (s *TalkToViewState) BackToTodoSelection() {
	s.ViewMode = TodoSelectionView
}

// EnterSearchModal initializes the note search modal in INSERT mode
func (s *TalkToViewState) EnterSearchModal() error {
	s.ViewMode = NoteSearchModalView
	s.SearchMode = InsertMode
	s.SearchQuery = ""
	s.SearchIndex = 0

	// Initialize with all notes
	allNotes, err := QueryFilesByDone(false)
	if err != nil {
		return err
	}

	s.SearchResults = allNotes
	return nil
}

// UpdateSearchResults filters notes based on the current search query
func (s *TalkToViewState) UpdateSearchResults() error {
	allNotes, err := QueryFilesByDone(false)
	if err != nil {
		return err
	}

	if s.SearchQuery == "" {
		s.SearchResults = allNotes
		return nil
	}

	// Simple substring matching (case-insensitive)
	query := strings.ToLower(s.SearchQuery)
	filtered := make([]scripts.File, 0)

	for _, note := range allNotes {
		fileName := strings.ToLower(note.Name)
		title := strings.ToLower(note.Title)

		if strings.Contains(fileName, query) || strings.Contains(title, query) {
			filtered = append(filtered, note)
		}
	}

	s.SearchResults = filtered

	// Reset index if out of bounds
	if s.SearchIndex >= len(s.SearchResults) {
		s.SearchIndex = 0
	}

	return nil
}

// ToggleSearchMode switches between INSERT and NORMAL mode
func (s *TalkToViewState) ToggleSearchMode() {
	if s.SearchMode == InsertMode {
		s.SearchMode = NormalMode
	} else {
		s.SearchMode = InsertMode
	}
}

// EnterInsertMode switches to INSERT mode
func (s *TalkToViewState) EnterInsertMode() {
	s.SearchMode = InsertMode
}

// AppendToSearchQuery adds a character to the search query and updates results
func (s *TalkToViewState) AppendToSearchQuery(char rune) error {
	s.SearchQuery += string(char)
	return s.UpdateSearchResults()
}

// BackspaceSearchQuery removes the last character from the search query
func (s *TalkToViewState) BackspaceSearchQuery() error {
	if len(s.SearchQuery) > 0 {
		s.SearchQuery = s.SearchQuery[:len(s.SearchQuery)-1]
		return s.UpdateSearchResults()
	}
	return nil
}

// SelectSearchResult sets the target note from the search results
func (s *TalkToViewState) SelectSearchResult() error {
	if s.SearchIndex < 0 || s.SearchIndex >= len(s.SearchResults) {
		return fmt.Errorf("no note selected")
	}

	selectedNote := s.SearchResults[s.SearchIndex]
	s.TargetNoteName = selectedNote.Name
	s.IsNewNote = false
	s.ViewMode = ConfirmationView

	return nil
}

// CancelSearch returns to NoteSelectionView
func (s *TalkToViewState) CancelSearch() {
	s.ViewMode = NoteSelectionView
}

// GetSelectedPerson returns the currently selected PersonInfo
func (s *TalkToViewState) GetSelectedPerson() *PersonInfo {
	if s.PersonIndex >= 0 && s.PersonIndex < len(s.AllPeople) {
		return &s.AllPeople[s.PersonIndex]
	}
	return nil
}

// GetSelectedTodos returns only the todos that are checked for moving
func (s *TalkToViewState) GetSelectedTodos() []TodoWithMeta {
	selected := make([]TodoWithMeta, 0)
	for i, todo := range s.AvailableTodos {
		if i < len(s.SelectedTodos) && s.SelectedTodos[i] {
			selected = append(selected, todo)
		}
	}
	return selected
}

// ExecuteMove performs the atomic move operation
// Inserts todos into target note and marks them complete in source files
func (s *TalkToViewState) ExecuteMove() error {
	selectedTodos := s.GetSelectedTodos()
	if len(selectedTodos) == 0 {
		return fmt.Errorf("no todos selected")
	}

	if s.TargetNoteName == "" {
		return fmt.Errorf("no target note selected")
	}

	// Track all modifications for rollback
	sourceModifications := make(map[string][]LineModification)

	// Step 1: Insert todos into target note
	insertionPoint, err := InsertTodosIntoNote(s.TargetNoteName, selectedTodos)
	if err != nil {
		return fmt.Errorf("failed to insert todos: %w", err)
	}

	// Step 2: Mark todos complete in source files (with rollback on error)
	for _, todo := range selectedTodos {
		// Mark main todo line complete
		oldContent := todo.TodoLine
		err := MarkTodoLineComplete(todo.File, todo.LineNumber)
		if err != nil {
			// Rollback: remove inserted todos
			_ = RemoveTodosFromNote(s.TargetNoteName, insertionPoint, countTodoLines(selectedTodos))
			// Rollback: restore previously modified lines
			rollbackSourceModifications(sourceModifications)
			return fmt.Errorf("failed to mark todo complete at %s:%d: %w", todo.SourceFile, todo.LineNumber, err)
		}

		// Track modification
		sourceModifications[todo.SourceFile] = append(sourceModifications[todo.SourceFile], LineModification{
			LineNumber: todo.LineNumber,
			OldContent: oldContent,
			NewContent: strings.Replace(oldContent, "- [ ]", "- [x]", 1),
		})

		// Mark subtasks complete
		for _, subtask := range todo.Subtasks {
			oldSubContent := subtask.Line
			err := MarkTodoLineComplete(todo.File, subtask.LineNumber)
			if err != nil {
				// Rollback all changes
				_ = RemoveTodosFromNote(s.TargetNoteName, insertionPoint, countTodoLines(selectedTodos))
				rollbackSourceModifications(sourceModifications)
				return fmt.Errorf("failed to mark subtask complete at %s:%d: %w", todo.SourceFile, subtask.LineNumber, err)
			}

			// Track subtask modification
			sourceModifications[todo.SourceFile] = append(sourceModifications[todo.SourceFile], LineModification{
				LineNumber: subtask.LineNumber,
				OldContent: oldSubContent,
				NewContent: strings.Replace(oldSubContent, "- [ ]", "- [x]", 1),
			})
		}
	}

	// Success! Add to undo stack
	change := MoveChange{
		Timestamp:            time.Now(),
		Person:               s.SelectedPerson,
		Todos:                selectedTodos,
		TargetNote:           s.TargetNoteName,
		TargetInsertionPoint: insertionPoint,
		SourceModifications:  sourceModifications,
	}

	s.UndoStack = append(s.UndoStack, change)
	s.ViewMode = SuccessView

	return nil
}

// UndoLastMove reverses the last move operation
func (s *TalkToViewState) UndoLastMove() error {
	if len(s.UndoStack) == 0 {
		return fmt.Errorf("no moves to undo")
	}

	// Pop from undo stack
	lastChange := s.UndoStack[len(s.UndoStack)-1]
	s.UndoStack = s.UndoStack[:len(s.UndoStack)-1]

	// Step 1: Remove todos from target note
	lineCount := countTodoLines(lastChange.Todos)
	err := RemoveTodosFromNote(lastChange.TargetNote, lastChange.TargetInsertionPoint, lineCount)
	if err != nil {
		return fmt.Errorf("failed to remove todos from target: %w", err)
	}

	// Step 2: Mark todos incomplete in source files
	for fileName, modifications := range lastChange.SourceModifications {
		// Load the file
		file, err := LoadFileByName(fileName)
		if err != nil {
			return fmt.Errorf("failed to load source file %s: %w", fileName, err)
		}

		// Mark each line incomplete (in reverse order to maintain line numbers)
		for _, mod := range modifications {
			err := MarkTodoLineIncomplete(file, mod.LineNumber)
			if err != nil {
				return fmt.Errorf("failed to mark line incomplete at %s:%d: %w", fileName, mod.LineNumber, err)
			}
		}
	}

	s.LastMessage = "Undo successful"
	return nil
}

// countTodoLines counts the total number of lines in a list of todos (including subtasks and padding)
func countTodoLines(todos []TodoWithMeta) int {
	count := 0
	for _, todo := range todos {
		count++ // Main todo line
		count += len(todo.Subtasks)
	}
	// Add padding lines (blank line before and after)
	if count > 0 {
		count += 2
	}
	return count
}

// rollbackSourceModifications reverts source file modifications
func rollbackSourceModifications(mods map[string][]LineModification) {
	for fileName, modifications := range mods {
		file, err := LoadFileByName(fileName)
		if err != nil {
			continue
		}

		for _, mod := range modifications {
			_ = MarkTodoLineIncomplete(file, mod.LineNumber)
		}
	}
}
