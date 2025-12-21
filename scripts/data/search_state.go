package data

import (
	"cli-notes/scripts"
	"strings"

	"github.com/sahilm/fuzzy"
)

type SearchViewMode int

const (
	SearchModeTyping SearchViewMode = iota // User is typing search query
	SearchModeActions                      // Quick actions menu is shown
)

// SearchResult represents a search match with context
type SearchResult struct {
	File           scripts.File
	MatchedIndices []int  // Character indices that matched in title
	ContentSnippet string // Matched content excerpt
	SnippetLine    int    // Line number where snippet was found
}

// QuickAction represents an action that can be performed on a search result
type QuickAction struct {
	Label       string
	Description string
	Key         rune
}

// SearchState holds all state for the interactive search view
type SearchState struct {
	ViewMode      SearchViewMode
	Query         string         // Current search query
	AllNotes      []scripts.File // All loaded notes
	Results       []SearchResult // Fuzzy-matched results
	SelectedIndex int            // Currently selected result
	ScrollOffset  int            // For scrolling through results
	ActionsIndex  int            // Selected action in actions menu

	// UI dimensions (set during render)
	TermWidth  int
	TermHeight int
}

// NewSearchState initializes search state with all notes
func NewSearchState(initialQuery string) (*SearchState, error) {
	// Query all notes (empty string returns all)
	notes, err := QueryFiles("")
	if err != nil {
		return nil, err
	}

	state := &SearchState{
		ViewMode:      SearchModeTyping,
		Query:         initialQuery,
		AllNotes:      notes,
		Results:       []SearchResult{},
		SelectedIndex: 0,
		ScrollOffset:  0,
		ActionsIndex:  0,
	}

	// Perform initial search if query provided
	state.UpdateQuery(initialQuery)

	return state, nil
}

// UpdateQuery performs fuzzy search and updates results
func (s *SearchState) UpdateQuery(query string) {
	s.Query = query
	s.SelectedIndex = 0
	s.ScrollOffset = 0

	if query == "" {
		// Show all notes when no query
		s.Results = make([]SearchResult, len(s.AllNotes))
		for i, note := range s.AllNotes {
			s.Results[i] = SearchResult{
				File:           note,
				MatchedIndices: []int{},
				ContentSnippet: extractSnippet(note.Content, "", 80),
				SnippetLine:    0,
			}
		}
		return
	}

	// Build search targets: combine title, tags, and content for matching
	searchTargets := make([]string, len(s.AllNotes))
	for i, note := range s.AllNotes {
		searchTargets[i] = strings.ToLower(note.Title + " " + strings.Join(note.Tags, " ") + " " + note.Content)
	}

	queryLower := strings.ToLower(query)
	matches := fuzzy.Find(queryLower, searchTargets)

	s.Results = make([]SearchResult, len(matches))
	for i, match := range matches {
		note := s.AllNotes[match.Index]
		snippet, lineNum := extractSnippetWithQuery(note.Content, query, 80)
		s.Results[i] = SearchResult{
			File:           note,
			MatchedIndices: match.MatchedIndexes,
			ContentSnippet: snippet,
			SnippetLine:    lineNum,
		}
	}
}

// SelectNext moves selection down
func (s *SearchState) SelectNext() {
	if s.ViewMode == SearchModeActions {
		actions := s.GetAvailableActions()
		if len(actions) > 0 {
			s.ActionsIndex = (s.ActionsIndex + 1) % len(actions)
		}
		return
	}

	if len(s.Results) > 0 {
		s.SelectedIndex = (s.SelectedIndex + 1) % len(s.Results)
		s.adjustScrollOffset()
	}
}

// SelectPrevious moves selection up
func (s *SearchState) SelectPrevious() {
	if s.ViewMode == SearchModeActions {
		actions := s.GetAvailableActions()
		if len(actions) > 0 {
			s.ActionsIndex--
			if s.ActionsIndex < 0 {
				s.ActionsIndex = len(actions) - 1
			}
		}
		return
	}

	if len(s.Results) > 0 {
		s.SelectedIndex--
		if s.SelectedIndex < 0 {
			s.SelectedIndex = len(s.Results) - 1
		}
		s.adjustScrollOffset()
	}
}

// adjustScrollOffset ensures selected item is visible
func (s *SearchState) adjustScrollOffset() {
	visibleRows := s.TermHeight - 10 // Account for header, input, status, controls
	if visibleRows < 1 {
		visibleRows = 10
	}

	if s.SelectedIndex < s.ScrollOffset {
		s.ScrollOffset = s.SelectedIndex
	} else if s.SelectedIndex >= s.ScrollOffset+visibleRows {
		s.ScrollOffset = s.SelectedIndex - visibleRows + 1
	}
}

// GetSelectedResult returns the currently selected search result
func (s *SearchState) GetSelectedResult() *SearchResult {
	if len(s.Results) == 0 || s.SelectedIndex >= len(s.Results) {
		return nil
	}
	return &s.Results[s.SelectedIndex]
}

// EnterActionsMode transitions to actions menu for selected result
func (s *SearchState) EnterActionsMode() {
	if len(s.Results) > 0 {
		s.ViewMode = SearchModeActions
		s.ActionsIndex = 0
	}
}

// ExitActionsMode returns to search/results mode
func (s *SearchState) ExitActionsMode() {
	s.ViewMode = SearchModeTyping
}

// GetAvailableActions returns quick actions for selected result
func (s *SearchState) GetAvailableActions() []QuickAction {
	result := s.GetSelectedResult()
	if result == nil {
		return nil
	}

	actions := []QuickAction{
		{Label: "Edit note", Description: "Open in editor", Key: 'o'},
	}

	// Add mark done/undone based on current state
	if result.File.Done {
		actions = append(actions, QuickAction{Label: "Mark incomplete", Description: "Unmark as done", Key: 'd'})
	} else {
		actions = append(actions, QuickAction{Label: "Mark done", Description: "Mark as completed", Key: 'd'})
	}

	// Priority actions
	actions = append(actions, QuickAction{Label: "Priority: P1", Description: "Set high priority", Key: '1'})
	actions = append(actions, QuickAction{Label: "Priority: P2", Description: "Set medium priority", Key: '2'})
	actions = append(actions, QuickAction{Label: "Priority: P3", Description: "Set low priority", Key: '3'})

	// Due date actions
	actions = append(actions, QuickAction{Label: "Due: Today", Description: "Set due today", Key: 't'})
	actions = append(actions, QuickAction{Label: "Due: Monday", Description: "Set due next Monday", Key: 'm'})

	// Link to objective (only for non-objectives)
	if result.File.ObjectiveRole != "parent" && result.File.ObjectiveID == "" {
		actions = append(actions, QuickAction{Label: "Link to objective", Description: "Associate with an objective", Key: 'l'})
	}

	return actions
}

// GetSelectedAction returns the currently selected action
func (s *SearchState) GetSelectedAction() *QuickAction {
	actions := s.GetAvailableActions()
	if len(actions) == 0 || s.ActionsIndex >= len(actions) {
		return nil
	}
	return &actions[s.ActionsIndex]
}

// AddChar adds a character to the query
func (s *SearchState) AddChar(c rune) {
	s.Query += string(c)
	s.UpdateQuery(s.Query)
}

// DeleteChar removes the last character from the query
func (s *SearchState) DeleteChar() {
	if len(s.Query) > 0 {
		runes := []rune(s.Query)
		s.Query = string(runes[:len(runes)-1])
		s.UpdateQuery(s.Query)
	}
}

// ClearQuery clears the entire query
func (s *SearchState) ClearQuery() {
	s.Query = ""
	s.UpdateQuery("")
}

// extractSnippet extracts a snippet from content
func extractSnippet(content, query string, maxLen int) string {
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "---") {
			if len(line) > maxLen {
				return line[:maxLen] + "..."
			}
			return line
		}
	}
	return ""
}

// extractSnippetWithQuery finds the best matching line for the query
func extractSnippetWithQuery(content, query string, maxLen int) (string, int) {
	if content == "" {
		return "", 0
	}

	queryLower := strings.ToLower(query)
	lines := strings.Split(content, "\n")

	// First, try to find a line containing the query
	for i, line := range lines {
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, queryLower) {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "---") {
				if len(line) > maxLen {
					return line[:maxLen] + "...", i + 1
				}
				return line, i + 1
			}
		}
	}

	// Fall back to first non-empty line
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "---") {
			if len(line) > maxLen {
				return line[:maxLen] + "...", i + 1
			}
			return line, i + 1
		}
	}

	return "", 0
}
