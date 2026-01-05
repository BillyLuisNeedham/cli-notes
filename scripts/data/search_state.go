package data

import (
	"cli-notes/scripts"
	"strings"

	"github.com/sahilm/fuzzy"
)

type SearchViewMode int

const (
	SearchModeInsert  SearchViewMode = iota // User is typing search query (all chars go to query)
	SearchModeNormal                        // Command mode (j/k navigate, shortcuts work)
	SearchModeActions                       // Quick actions menu is shown
)

type SearchMatchMode int

const (
	MatchModeFuzzy  SearchMatchMode = iota // Fuzzy matching (letters in sequence)
	MatchModeStrict                        // Strict substring matching
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
	FilterMode    FilterMode     // Show all/incomplete only/complete only
	MatchMode     SearchMatchMode // Fuzzy or strict matching

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
		ViewMode:      SearchModeInsert, // Start in insert mode for immediate typing
		Query:         initialQuery,
		AllNotes:      notes,
		Results:       []SearchResult{},
		SelectedIndex: 0,
		ScrollOffset:  0,
		ActionsIndex:  0,
		FilterMode:    ShowIncompleteOnly, // Default to showing incomplete notes
		MatchMode:     MatchModeStrict,    // Default to strict substring matching
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

	var candidates []SearchResult

	if query == "" {
		// Show all notes when no query
		candidates = make([]SearchResult, len(s.AllNotes))
		for i, note := range s.AllNotes {
			candidates[i] = SearchResult{
				File:           note,
				MatchedIndices: []int{},
				ContentSnippet: extractSnippet(note.Content, "", 80),
				SnippetLine:    0,
			}
		}
	} else {
		// Split query by comma for AND logic (like gt command)
		queries := strings.Split(query, ",")
		for i := range queries {
			queries[i] = strings.TrimSpace(queries[i])
		}

		// Build search targets: combine title, tags, and content for matching
		searchTargets := make([]string, len(s.AllNotes))
		for i, note := range s.AllNotes {
			searchTargets[i] = strings.ToLower(note.Title + " " + strings.Join(note.Tags, " ") + " " + note.Content)
		}

		// For each note, check if it matches ALL query terms
		for i, note := range s.AllNotes {
			matchCount := 0
			for _, q := range queries {
				if q == "" {
					matchCount++
					continue
				}
				qLower := strings.ToLower(q)
				var matched bool
				if s.MatchMode == MatchModeFuzzy {
					matches := fuzzy.Find(qLower, []string{searchTargets[i]})
					matched = len(matches) > 0
				} else {
					matched = strings.Contains(searchTargets[i], qLower)
				}
				if matched {
					matchCount++
				}
			}
			// Only include if all query terms match
			if matchCount == len(queries) {
				snippet, lineNum := extractSnippetWithQuery(note.Content, queries[0], 80)
				candidates = append(candidates, SearchResult{
					File:           note,
					MatchedIndices: []int{},
					ContentSnippet: snippet,
					SnippetLine:    lineNum,
				})
			}
		}
	}

	// Apply filter mode (all/incomplete/complete)
	s.Results = s.applyFilterMode(candidates)
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

// ExitActionsMode returns to normal mode
func (s *SearchState) ExitActionsMode() {
	s.ViewMode = SearchModeNormal
}

// EnterInsertMode switches to insert mode for typing
func (s *SearchState) EnterInsertMode() {
	s.ViewMode = SearchModeInsert
}

// EnterNormalMode switches to normal/command mode
func (s *SearchState) EnterNormalMode() {
	s.ViewMode = SearchModeNormal
}

// CycleFilterMode cycles through filter modes: All -> Incomplete -> Complete -> All
func (s *SearchState) CycleFilterMode() {
	switch s.FilterMode {
	case ShowAll:
		s.FilterMode = ShowIncompleteOnly
	case ShowIncompleteOnly:
		s.FilterMode = ShowCompleteOnly
	case ShowCompleteOnly:
		s.FilterMode = ShowAll
	}
	s.UpdateQuery(s.Query) // Re-apply filter with current query
}

// CycleMatchMode toggles between fuzzy and strict matching
func (s *SearchState) CycleMatchMode() {
	if s.MatchMode == MatchModeFuzzy {
		s.MatchMode = MatchModeStrict
	} else {
		s.MatchMode = MatchModeFuzzy
	}
	s.UpdateQuery(s.Query) // Re-run search with new mode
}

// GetMatchModeLabel returns display label for current match mode
func (s *SearchState) GetMatchModeLabel() string {
	if s.MatchMode == MatchModeFuzzy {
		return "Fuzzy"
	}
	return "Strict"
}

// applyFilterMode filters results based on current filter mode
func (s *SearchState) applyFilterMode(candidates []SearchResult) []SearchResult {
	if s.FilterMode == ShowAll {
		return candidates
	}

	filtered := make([]SearchResult, 0, len(candidates))
	for _, result := range candidates {
		if s.FilterMode == ShowIncompleteOnly && !result.File.Done {
			filtered = append(filtered, result)
		} else if s.FilterMode == ShowCompleteOnly && result.File.Done {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// GetAvailableActions returns quick actions for selected result
func (s *SearchState) GetAvailableActions() []QuickAction {
	result := s.GetSelectedResult()
	if result == nil {
		return nil
	}

	actions := []QuickAction{
		{Label: "Edit note", Description: "Open in editor", Key: 'e'},
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

	// Link actions
	actions = append(actions, QuickAction{Label: "Link to note", Description: "Create link to another note", Key: 'l'})

	// Link to objective (only for non-objectives that aren't already linked)
	if result.File.ObjectiveRole != "parent" && result.File.ObjectiveID == "" {
		actions = append(actions, QuickAction{Label: "Link to objective", Description: "Associate with an objective", Key: 'o'})
	}

	// Open graph view
	actions = append(actions, QuickAction{Label: "Open graph view", Description: "View linked notes", Key: 'L'})

	// Open objectives view (only for objectives or linked notes)
	if result.File.ObjectiveRole == "parent" || result.File.ObjectiveID != "" {
		actions = append(actions, QuickAction{Label: "Open objectives view", Description: "View objective", Key: 'O'})
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
