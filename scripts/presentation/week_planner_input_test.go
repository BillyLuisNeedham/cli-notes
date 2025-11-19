package presentation

import (
	"cli-notes/scripts/data"
	"os"
	"testing"
)

func TestHandleWeekPlannerInput_ToggleExpandedEarlier(t *testing.T) {
	// Setup
	// Create temporary notes directory for the test
	if err := os.MkdirAll("notes", 0755); err != nil {
		t.Fatalf("Failed to create notes directory: %v", err)
	}
	defer os.RemoveAll("notes")

	state, err := data.NewWeekPlannerState()
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	// Initial state should be NormalView
	if state.ViewMode != data.NormalView {
		t.Errorf("Expected initial ViewMode to be NormalView, got %v", state.ViewMode)
	}

	// Test 1: Press 'e' to enter ExpandedEarlierView
	input := WeekPlannerInput{Action: ToggleExpandedEarlier}
	shouldExit, message, err := HandleWeekPlannerInput(state, input)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if shouldExit {
		t.Error("Expected shouldExit to be false")
	}
	if state.ViewMode != data.ExpandedEarlierView {
		t.Errorf("Expected ViewMode to be ExpandedEarlierView after pressing 'e', got %v", state.ViewMode)
	}
	if message != "Expanded Earlier view" {
		t.Errorf("Expected message 'Expanded Earlier view', got '%s'", message)
	}

	// Test 2: Press 'e' again to exit ExpandedEarlierView (The Fix)
	shouldExit, message, err = HandleWeekPlannerInput(state, input)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if shouldExit {
		t.Error("Expected shouldExit to be false")
	}
	if state.ViewMode != data.NormalView {
		t.Errorf("Expected ViewMode to be NormalView after pressing 'e' again, got %v", state.ViewMode)
	}
	if message != "Returned to normal view" {
		t.Errorf("Expected message 'Returned to normal view', got '%s'", message)
	}
}
