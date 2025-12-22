package data

import (
	"cli-notes/scripts"
	"testing"
	"time"
)

// ============================================
// CycleFilterMode Tests
// ============================================

func TestCycleFilterMode_CyclesCorrectly(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create a test note so search state can load
	createTestFile(t, scripts.File{
		Name:      "test.md",
		Title:     "Test Note",
		CreatedAt: time.Now(),
		Done:      false,
	})

	state, err := NewSearchState("")
	if err != nil {
		t.Fatalf("Failed to create search state: %v", err)
	}

	// Default should be ShowIncompleteOnly
	if state.FilterMode != ShowIncompleteOnly {
		t.Errorf("Expected default filter mode ShowIncompleteOnly, got %d", state.FilterMode)
	}

	// Cycle: ShowIncompleteOnly -> ShowCompleteOnly
	state.CycleFilterMode()
	if state.FilterMode != ShowCompleteOnly {
		t.Errorf("Expected ShowCompleteOnly after first cycle, got %d", state.FilterMode)
	}

	// Cycle: ShowCompleteOnly -> ShowAll
	state.CycleFilterMode()
	if state.FilterMode != ShowAll {
		t.Errorf("Expected ShowAll after second cycle, got %d", state.FilterMode)
	}

	// Cycle: ShowAll -> ShowIncompleteOnly
	state.CycleFilterMode()
	if state.FilterMode != ShowIncompleteOnly {
		t.Errorf("Expected ShowIncompleteOnly after third cycle, got %d", state.FilterMode)
	}
}

func TestCycleFilterMode_ReappliesFilter(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create mix of done and incomplete notes
	createTestFile(t, scripts.File{
		Name:      "incomplete1.md",
		Title:     "Incomplete Note 1",
		CreatedAt: time.Now(),
		Done:      false,
	})
	createTestFile(t, scripts.File{
		Name:      "complete1.md",
		Title:     "Complete Note 1",
		CreatedAt: time.Now(),
		Done:      true,
	})
	createTestFile(t, scripts.File{
		Name:      "incomplete2.md",
		Title:     "Incomplete Note 2",
		CreatedAt: time.Now(),
		Done:      false,
	})

	state, err := NewSearchState("")
	if err != nil {
		t.Fatalf("Failed to create search state: %v", err)
	}

	// Default (ShowIncompleteOnly) should show 2 incomplete notes
	if len(state.Results) != 2 {
		t.Errorf("Expected 2 results with ShowIncompleteOnly, got %d", len(state.Results))
	}

	// Cycle to ShowCompleteOnly - should show 1 complete note
	state.CycleFilterMode()
	if len(state.Results) != 1 {
		t.Errorf("Expected 1 result with ShowCompleteOnly, got %d", len(state.Results))
	}
	if len(state.Results) > 0 && !state.Results[0].File.Done {
		t.Error("Expected complete note in ShowCompleteOnly mode")
	}

	// Cycle to ShowAll - should show all 3 notes
	state.CycleFilterMode()
	if len(state.Results) != 3 {
		t.Errorf("Expected 3 results with ShowAll, got %d", len(state.Results))
	}
}

// ============================================
// applyFilterMode Tests
// ============================================

func TestApplyFilterMode_ShowAll(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "note1.md",
		Title:     "Note 1",
		CreatedAt: time.Now(),
		Done:      false,
	})

	state, err := NewSearchState("")
	if err != nil {
		t.Fatalf("Failed to create search state: %v", err)
	}

	candidates := []SearchResult{
		{File: scripts.File{Name: "a.md", Done: false}},
		{File: scripts.File{Name: "b.md", Done: true}},
		{File: scripts.File{Name: "c.md", Done: false}},
	}

	state.FilterMode = ShowAll
	result := state.applyFilterMode(candidates)

	if len(result) != 3 {
		t.Errorf("ShowAll should return all candidates, got %d", len(result))
	}
}

func TestApplyFilterMode_ShowIncompleteOnly(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "note1.md",
		Title:     "Note 1",
		CreatedAt: time.Now(),
		Done:      false,
	})

	state, err := NewSearchState("")
	if err != nil {
		t.Fatalf("Failed to create search state: %v", err)
	}

	candidates := []SearchResult{
		{File: scripts.File{Name: "a.md", Done: false}},
		{File: scripts.File{Name: "b.md", Done: true}},
		{File: scripts.File{Name: "c.md", Done: false}},
		{File: scripts.File{Name: "d.md", Done: true}},
	}

	state.FilterMode = ShowIncompleteOnly
	result := state.applyFilterMode(candidates)

	if len(result) != 2 {
		t.Errorf("ShowIncompleteOnly should return 2 incomplete notes, got %d", len(result))
	}

	for _, r := range result {
		if r.File.Done {
			t.Errorf("ShowIncompleteOnly returned a done note: %s", r.File.Name)
		}
	}
}

func TestApplyFilterMode_ShowCompleteOnly(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "note1.md",
		Title:     "Note 1",
		CreatedAt: time.Now(),
		Done:      false,
	})

	state, err := NewSearchState("")
	if err != nil {
		t.Fatalf("Failed to create search state: %v", err)
	}

	candidates := []SearchResult{
		{File: scripts.File{Name: "a.md", Done: false}},
		{File: scripts.File{Name: "b.md", Done: true}},
		{File: scripts.File{Name: "c.md", Done: false}},
		{File: scripts.File{Name: "d.md", Done: true}},
	}

	state.FilterMode = ShowCompleteOnly
	result := state.applyFilterMode(candidates)

	if len(result) != 2 {
		t.Errorf("ShowCompleteOnly should return 2 complete notes, got %d", len(result))
	}

	for _, r := range result {
		if !r.File.Done {
			t.Errorf("ShowCompleteOnly returned an incomplete note: %s", r.File.Name)
		}
	}
}

func TestApplyFilterMode_EmptyResults(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "note1.md",
		Title:     "Note 1",
		CreatedAt: time.Now(),
		Done:      false,
	})

	state, err := NewSearchState("")
	if err != nil {
		t.Fatalf("Failed to create search state: %v", err)
	}

	// All incomplete notes with ShowCompleteOnly should return empty
	candidates := []SearchResult{
		{File: scripts.File{Name: "a.md", Done: false}},
		{File: scripts.File{Name: "b.md", Done: false}},
	}

	state.FilterMode = ShowCompleteOnly
	result := state.applyFilterMode(candidates)

	if len(result) != 0 {
		t.Errorf("Expected empty results, got %d", len(result))
	}

	// All complete notes with ShowIncompleteOnly should return empty
	candidates = []SearchResult{
		{File: scripts.File{Name: "a.md", Done: true}},
		{File: scripts.File{Name: "b.md", Done: true}},
	}

	state.FilterMode = ShowIncompleteOnly
	result = state.applyFilterMode(candidates)

	if len(result) != 0 {
		t.Errorf("Expected empty results, got %d", len(result))
	}
}

func TestApplyFilterMode_PreservesOrder(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "note1.md",
		Title:     "Note 1",
		CreatedAt: time.Now(),
		Done:      false,
	})

	state, err := NewSearchState("")
	if err != nil {
		t.Fatalf("Failed to create search state: %v", err)
	}

	candidates := []SearchResult{
		{File: scripts.File{Name: "first.md", Done: false}},
		{File: scripts.File{Name: "second.md", Done: true}},
		{File: scripts.File{Name: "third.md", Done: false}},
		{File: scripts.File{Name: "fourth.md", Done: true}},
		{File: scripts.File{Name: "fifth.md", Done: false}},
	}

	state.FilterMode = ShowIncompleteOnly
	result := state.applyFilterMode(candidates)

	expectedOrder := []string{"first.md", "third.md", "fifth.md"}
	if len(result) != len(expectedOrder) {
		t.Fatalf("Expected %d results, got %d", len(expectedOrder), len(result))
	}

	for i, expected := range expectedOrder {
		if result[i].File.Name != expected {
			t.Errorf("At index %d: expected %s, got %s", i, expected, result[i].File.Name)
		}
	}
}

// ============================================
// NewSearchState Default Filter Tests
// ============================================

func TestNewSearchState_DefaultsToShowIncompleteOnly(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "note.md",
		Title:     "Note",
		CreatedAt: time.Now(),
		Done:      false,
	})

	state, err := NewSearchState("")
	if err != nil {
		t.Fatalf("Failed to create search state: %v", err)
	}

	if state.FilterMode != ShowIncompleteOnly {
		t.Errorf("Expected default FilterMode to be ShowIncompleteOnly, got %d", state.FilterMode)
	}
}

func TestNewSearchState_AppliesDefaultFilterOnInit(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create mix of done and incomplete notes
	createTestFile(t, scripts.File{
		Name:      "incomplete.md",
		Title:     "Incomplete Note",
		CreatedAt: time.Now(),
		Done:      false,
	})
	createTestFile(t, scripts.File{
		Name:      "complete.md",
		Title:     "Complete Note",
		CreatedAt: time.Now(),
		Done:      true,
	})

	state, err := NewSearchState("")
	if err != nil {
		t.Fatalf("Failed to create search state: %v", err)
	}

	// Default filter (ShowIncompleteOnly) should exclude the complete note
	if len(state.Results) != 1 {
		t.Errorf("Expected 1 result (incomplete only), got %d", len(state.Results))
	}

	if len(state.Results) > 0 && state.Results[0].File.Done {
		t.Error("Default filter should not include completed notes")
	}
}

// ============================================
// UpdateQuery with Filter Tests
// ============================================

func TestUpdateQuery_AppliesFilter(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "meeting-incomplete.md",
		Title:     "Meeting Incomplete",
		CreatedAt: time.Now(),
		Done:      false,
	})
	createTestFile(t, scripts.File{
		Name:      "meeting-complete.md",
		Title:     "Meeting Complete",
		CreatedAt: time.Now(),
		Done:      true,
	})
	createTestFile(t, scripts.File{
		Name:      "other.md",
		Title:     "Other Note",
		CreatedAt: time.Now(),
		Done:      false,
	})

	state, err := NewSearchState("")
	if err != nil {
		t.Fatalf("Failed to create search state: %v", err)
	}

	// Search for "meeting" with default filter (ShowIncompleteOnly)
	state.UpdateQuery("meeting")
	if len(state.Results) != 1 {
		t.Errorf("Expected 1 result (meeting incomplete only), got %d", len(state.Results))
	}

	// Change to ShowAll and search again
	state.FilterMode = ShowAll
	state.UpdateQuery("meeting")
	if len(state.Results) != 2 {
		t.Errorf("Expected 2 results (both meetings), got %d", len(state.Results))
	}

	// Change to ShowCompleteOnly
	state.FilterMode = ShowCompleteOnly
	state.UpdateQuery("meeting")
	if len(state.Results) != 1 {
		t.Errorf("Expected 1 result (meeting complete only), got %d", len(state.Results))
	}
	if len(state.Results) > 0 && !state.Results[0].File.Done {
		t.Error("ShowCompleteOnly should only return done notes")
	}
}
