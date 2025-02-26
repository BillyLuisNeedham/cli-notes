package data

import (
	"cli-notes/scripts/config"
	"reflect"
	"testing"
)

// SetupTestTeamNames is a helper function to set up test team names
// and return a function to restore the original names afterward
func setupTestTeamNames(names []string) func() {
	original := config.TEAM_NAMES
	config.TEAM_NAMES = names
	return func() {
		config.TEAM_NAMES = original
	}
}

func TestGetTeamNames_Success(t *testing.T) {
	// Set up test data
	testNames := []string{"Team A", "Team B", "Team C"}
	cleanup := setupTestTeamNames(testNames)
	defer cleanup()

	// Call the function
	names, err := GetTeamNames()

	// Assert expectations
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	if !reflect.DeepEqual(names, testNames) {
		t.Errorf("Expected team names %v, but got %v", testNames, names)
	}
}

func TestGetTeamNames_EmptyList(t *testing.T) {
	// Set up empty team names
	cleanup := setupTestTeamNames([]string{})
	defer cleanup()

	// Call the function
	names, err := GetTeamNames()

	// Assert expectations
	if err == nil {
		t.Error("Expected an error for empty team names, but got nil")
	}

	if err.Error() != "team names are empty" {
		t.Errorf("Expected error message 'team names are empty', but got '%s'", err.Error())
	}

	if names != nil {
		t.Errorf("Expected nil names, but got %v", names)
	}
}

func TestGetTeamNames_SingleTeam(t *testing.T) {
	// Set up test data with a single team
	testNames := []string{"Solo Team"}
	cleanup := setupTestTeamNames(testNames)
	defer cleanup()

	// Call the function
	names, err := GetTeamNames()

	// Assert expectations
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	if len(names) != 1 || names[0] != "Solo Team" {
		t.Errorf("Expected team name ['Solo Team'], but got %v", names)
	}
} 