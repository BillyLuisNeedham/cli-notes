package presentation

import (
	"cli-notes/scripts"
	"strings"

	"github.com/sahilm/fuzzy"
)

// AutocompleteState tracks the state of autocomplete during command input
type AutocompleteState struct {
	Input      string        // The user's input text (e.g., "an")
	Candidates []scripts.File // All matching objectives
	Index      int           // Current selected index in candidates (-1 if not cycling)
}

// NewAutocompleteState creates a new autocomplete state
func NewAutocompleteState(input string, candidates []scripts.File) AutocompleteState {
	return AutocompleteState{
		Input:      input,
		Candidates: candidates,
		Index:      -1, // Not cycling yet
	}
}

// GetCurrentCompletion returns the current completion text
// If not cycling (Index == -1), returns the longest common prefix
// If cycling, returns the candidate at the current index
func (a *AutocompleteState) GetCurrentCompletion() string {
	if len(a.Candidates) == 0 {
		return a.Input
	}

	if a.Index >= 0 && a.Index < len(a.Candidates) {
		// Cycling through candidates
		return a.Candidates[a.Index].Title
	}

	// Not cycling - return longest common prefix if there are multiple candidates
	if len(a.Candidates) == 1 {
		return a.Candidates[0].Title
	}

	return longestCommonPrefix(a.Candidates)
}

// CycleNext advances to the next candidate
func (a *AutocompleteState) CycleNext() {
	if len(a.Candidates) == 0 {
		return
	}

	a.Index++
	if a.Index >= len(a.Candidates) {
		a.Index = 0 // Wrap around
	}
}

// longestCommonPrefix finds the longest common prefix among all candidate titles
func longestCommonPrefix(candidates []scripts.File) string {
	if len(candidates) == 0 {
		return ""
	}

	if len(candidates) == 1 {
		return candidates[0].Title
	}

	prefix := candidates[0].Title
	for i := 1; i < len(candidates); i++ {
		prefix = commonPrefix(prefix, candidates[i].Title)
		if prefix == "" {
			break
		}
	}

	return prefix
}

// commonPrefix finds the common prefix between two strings (case-insensitive)
func commonPrefix(a, b string) string {
	// Convert to lowercase for case-insensitive comparison
	aLower := strings.ToLower(a)
	bLower := strings.ToLower(b)

	minLen := len(aLower)
	if len(bLower) < minLen {
		minLen = len(bLower)
	}

	for i := 0; i < minLen; i++ {
		if aLower[i] != bLower[i] {
			return a[:i] // Return original case from 'a'
		}
	}

	return a[:minLen] // Return original case from 'a'
}

// FilterObjectivesByPrefix filters objectives by title prefix (case-insensitive)
func FilterObjectivesByPrefix(objectives []scripts.File, prefix string) []scripts.File {
	if prefix == "" {
		return objectives
	}

	prefixLower := strings.ToLower(prefix)
	matches := make([]scripts.File, 0)

	for _, obj := range objectives {
		titleLower := strings.ToLower(obj.Title)
		if strings.HasPrefix(titleLower, prefixLower) {
			matches = append(matches, obj)
		}
	}

	return matches
}

// FuzzyFilterObjectives filters objectives using fzf-style fuzzy matching
// Results are sorted by match quality (best first)
func FuzzyFilterObjectives(objectives []scripts.File, pattern string) []scripts.File {
	if pattern == "" {
		return objectives
	}

	patternLower := strings.ToLower(pattern)

	// Create lowercase titles for case-insensitive matching
	lowercaseTitles := make([]string, len(objectives))
	for i, obj := range objectives {
		lowercaseTitles[i] = strings.ToLower(obj.Title)
	}

	matches := fuzzy.Find(patternLower, lowercaseTitles)

	result := make([]scripts.File, len(matches))
	for i, match := range matches {
		result[i] = objectives[match.Index]
	}

	return result
}
