package presentation

import (
	"cli-notes/scripts"
	"testing"
	"time"
)

func TestFilterObjectivesByPrefix(t *testing.T) {
	objectives := []scripts.File{
		{Title: "annual-review", Name: "annual-review.md"},
		{Title: "annual-planning", Name: "annual-planning.md"},
		{Title: "quarterly-goals", Name: "quarterly-goals.md"},
		{Title: "team-objectives", Name: "team-objectives.md"},
	}

	tests := []struct {
		name     string
		prefix   string
		expected int
	}{
		{
			name:     "Empty prefix returns all",
			prefix:   "",
			expected: 4,
		},
		{
			name:     "Prefix 'an' matches two",
			prefix:   "an",
			expected: 2,
		},
		{
			name:     "Prefix 'annual' matches two",
			prefix:   "annual",
			expected: 2,
		},
		{
			name:     "Prefix 'q' matches one",
			prefix:   "q",
			expected: 1,
		},
		{
			name:     "No match returns empty",
			prefix:   "xyz",
			expected: 0,
		},
		{
			name:     "Case insensitive matching",
			prefix:   "AN",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterObjectivesByPrefix(objectives, tt.prefix)
			if len(result) != tt.expected {
				t.Errorf("Expected %d matches, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestAutocompleteState_GetCurrentCompletion(t *testing.T) {
	objectives := []scripts.File{
		{Title: "annual-review", Name: "annual-review.md"},
		{Title: "annual-planning", Name: "annual-planning.md"},
		{Title: "quarterly-goals", Name: "quarterly-goals.md"},
	}

	t.Run("Single candidate returns title", func(t *testing.T) {
		state := NewAutocompleteState("q", []scripts.File{objectives[2]})
		result := state.GetCurrentCompletion()
		if result != "quarterly-goals" {
			t.Errorf("Expected 'quarterly-goals', got '%s'", result)
		}
	})

	t.Run("Multiple candidates return longest common prefix", func(t *testing.T) {
		state := NewAutocompleteState("an", []scripts.File{objectives[0], objectives[1]})
		result := state.GetCurrentCompletion()
		if result != "annual-" {
			t.Errorf("Expected 'annual-', got '%s'", result)
		}
	})

	t.Run("No candidates returns input", func(t *testing.T) {
		state := NewAutocompleteState("xyz", []scripts.File{})
		result := state.GetCurrentCompletion()
		if result != "xyz" {
			t.Errorf("Expected 'xyz', got '%s'", result)
		}
	})
}

func TestAutocompleteState_CycleNext(t *testing.T) {
	objectives := []scripts.File{
		{Title: "objective-a", Name: "objective-a.md"},
		{Title: "objective-b", Name: "objective-b.md"},
		{Title: "objective-c", Name: "objective-c.md"},
	}

	state := NewAutocompleteState("obj", objectives)

	// Initial state (Index = -1)
	if state.Index != -1 {
		t.Errorf("Expected initial Index to be -1, got %d", state.Index)
	}

	// First cycle
	state.CycleNext()
	if state.Index != 0 {
		t.Errorf("Expected Index to be 0 after first cycle, got %d", state.Index)
	}
	if state.GetCurrentCompletion() != "objective-a" {
		t.Errorf("Expected 'objective-a', got '%s'", state.GetCurrentCompletion())
	}

	// Second cycle
	state.CycleNext()
	if state.Index != 1 {
		t.Errorf("Expected Index to be 1 after second cycle, got %d", state.Index)
	}
	if state.GetCurrentCompletion() != "objective-b" {
		t.Errorf("Expected 'objective-b', got '%s'", state.GetCurrentCompletion())
	}

	// Third cycle
	state.CycleNext()
	if state.Index != 2 {
		t.Errorf("Expected Index to be 2 after third cycle, got %d", state.Index)
	}
	if state.GetCurrentCompletion() != "objective-c" {
		t.Errorf("Expected 'objective-c', got '%s'", state.GetCurrentCompletion())
	}

	// Fourth cycle (wrap around)
	state.CycleNext()
	if state.Index != 0 {
		t.Errorf("Expected Index to wrap to 0, got %d", state.Index)
	}
	if state.GetCurrentCompletion() != "objective-a" {
		t.Errorf("Expected 'objective-a' after wrap, got '%s'", state.GetCurrentCompletion())
	}
}

func TestLongestCommonPrefix(t *testing.T) {
	tests := []struct {
		name       string
		candidates []scripts.File
		expected   string
	}{
		{
			name: "Same prefix",
			candidates: []scripts.File{
				{Title: "test-one"},
				{Title: "test-two"},
				{Title: "test-three"},
			},
			expected: "test-",
		},
		{
			name: "Exact match with prefix",
			candidates: []scripts.File{
				{Title: "annual-review"},
				{Title: "annual-planning"},
			},
			expected: "annual-",
		},
		{
			name: "No common prefix",
			candidates: []scripts.File{
				{Title: "apple"},
				{Title: "banana"},
			},
			expected: "",
		},
		{
			name: "Single candidate",
			candidates: []scripts.File{
				{Title: "single"},
			},
			expected: "single",
		},
		{
			name:       "Empty list",
			candidates: []scripts.File{},
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := longestCommonPrefix(tt.candidates)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestCommandHandler_TabAutocomplete(t *testing.T) {
	objectives := []scripts.File{
		{
			Title:         "annual-review",
			Name:          "annual-review.md",
			ObjectiveRole: "parent",
			ObjectiveID:   "abc12345",
			Done:          false,
			CreatedAt:     time.Now(),
		},
		{
			Title:         "annual-planning",
			Name:          "annual-planning.md",
			ObjectiveRole: "parent",
			ObjectiveID:   "def67890",
			Done:          false,
			CreatedAt:     time.Now(),
		},
		{
			Title:         "quarterly-goals",
			Name:          "quarterly-goals.md",
			ObjectiveRole: "parent",
			ObjectiveID:   "ghi11111",
			Done:          false,
			CreatedAt:     time.Now(),
		},
	}

	selectedFile := scripts.File{
		Name:  "my-todo.md",
		Title: "My Todo",
	}

	getObjectives := func() ([]scripts.File, error) {
		return objectives, nil
	}

	t.Run("Tab without 'ob ' command returns unchanged", func(t *testing.T) {
		cmd := WIPCommand{
			Text:         "gt",
			SelectedFile: selectedFile,
		}

		result, err := CommandHandler(
			'\t',
			9, // KeyTab
			cmd,
			func() scripts.File { return scripts.File{} },
			func() scripts.File { return scripts.File{} },
			func(scripts.File) ([]string, error) { return nil, nil },
			func() {},
			getObjectives,
		)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		wipResult, ok := result.(WIPCommand)
		if !ok {
			t.Fatalf("Expected WIPCommand, got %T", result)
		}

		if wipResult.Text != "gt" {
			t.Errorf("Expected text to be unchanged 'gt', got '%s'", wipResult.Text)
		}
	})

	t.Run("Tab with 'ob an' autocompletes to common prefix", func(t *testing.T) {
		cmd := WIPCommand{
			Text:         "ob an",
			SelectedFile: selectedFile,
		}

		result, err := CommandHandler(
			'\t',
			9, // KeyTab
			cmd,
			func() scripts.File { return scripts.File{} },
			func() scripts.File { return scripts.File{} },
			func(scripts.File) ([]string, error) { return nil, nil },
			func() {},
			getObjectives,
		)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		tabResult, ok := result.(TabPressedWIPCommand)
		if !ok {
			t.Fatalf("Expected TabPressedWIPCommand, got %T", result)
		}

		// First Tab should cycle to first candidate
		if tabResult.Text != "ob annual-review" {
			t.Errorf("Expected 'ob annual-review', got '%s'", tabResult.Text)
		}

		if tabResult.AutocompleteState == nil {
			t.Fatal("Expected autocomplete state to be set")
		}
	})

	t.Run("Multiple Tab presses cycle through candidates", func(t *testing.T) {
		cmd := WIPCommand{
			Text:         "ob an",
			SelectedFile: selectedFile,
		}

		// First Tab
		result1, _ := CommandHandler('\t', 9, cmd, nil, nil, nil, func() {}, getObjectives)
		tabResult1 := result1.(TabPressedWIPCommand)

		if tabResult1.Text != "ob annual-review" {
			t.Errorf("First tab: Expected 'ob annual-review', got '%s'", tabResult1.Text)
		}

		// Second Tab
		result2, _ := CommandHandler('\t', 9, tabResult1.WIPCommand, nil, nil, nil, func() {}, getObjectives)
		tabResult2 := result2.(TabPressedWIPCommand)

		if tabResult2.Text != "ob annual-planning" {
			t.Errorf("Second tab: Expected 'ob annual-planning', got '%s'", tabResult2.Text)
		}

		// Third Tab (wrap around)
		result3, _ := CommandHandler('\t', 9, tabResult2.WIPCommand, nil, nil, nil, func() {}, getObjectives)
		tabResult3 := result3.(TabPressedWIPCommand)

		if tabResult3.Text != "ob annual-review" {
			t.Errorf("Third tab: Expected wrap to 'ob annual-review', got '%s'", tabResult3.Text)
		}
	})
}
