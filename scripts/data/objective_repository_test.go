package data

import (
	"cli-notes/scripts"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestObjectives(t *testing.T) (string, func()) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "objective-test-*")
	if err != nil {
		t.Fatal(err)
	}

	notesDir := filepath.Join(tempDir, "notes")
	if err := os.Mkdir(notesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		os.Chdir(originalDir)
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestQueryAllObjectives(t *testing.T) {
	_, cleanup := setupTestObjectives(t)
	defer cleanup()

	// Create test objectives
	objective1 := scripts.File{
		Name:          "objective1-2025-11-08.md",
		Title:         "Objective 1",
		Tags:          []string{"objective"},
		CreatedAt:     time.Now(),
		ObjectiveRole: "parent",
		ObjectiveID:   "abc12345",
		Content:       "# Objective 1\n\nSome content",
		Priority:      scripts.P1,
	}

	objective2 := scripts.File{
		Name:          "objective2-2025-11-08.md",
		Title:         "Objective 2",
		Tags:          []string{"objective"},
		CreatedAt:     time.Now(),
		ObjectiveRole: "parent",
		ObjectiveID:   "def67890",
		Content:       "# Objective 2\n\nMore content",
		Priority:      scripts.P2,
	}

	// Create a regular todo (should not be returned)
	regularTodo := scripts.File{
		Name:      "todo-2025-11-08.md",
		Title:     "Regular Todo",
		Tags:      []string{"todo"},
		CreatedAt: time.Now(),
		Done:      false,
		Content:   "# Regular Todo",
		Priority:  scripts.P2,
	}

	// Write files
	if err := WriteFile(objective1); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(objective2); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(regularTodo); err != nil {
		t.Fatal(err)
	}

	// Query objectives
	objectives, err := QueryAllObjectives()
	if err != nil {
		t.Fatalf("Error querying objectives: %v", err)
	}

	// Should return only 2 objectives
	if len(objectives) != 2 {
		t.Errorf("Expected 2 objectives, got %d", len(objectives))
	}

	// Verify all returned files have objective-role: parent
	for _, obj := range objectives {
		if obj.ObjectiveRole != "parent" {
			t.Errorf("Expected ObjectiveRole 'parent', got '%s'", obj.ObjectiveRole)
		}
	}
}

func TestQueryChildrenByObjectiveID(t *testing.T) {
	_, cleanup := setupTestObjectives(t)
	defer cleanup()

	objectiveID := "test1234"

	// Create parent objective
	parent := scripts.File{
		Name:          "parent-2025-11-08.md",
		Title:         "Parent Objective",
		Tags:          []string{"objective"},
		CreatedAt:     time.Now(),
		ObjectiveRole: "parent",
		ObjectiveID:   objectiveID,
		Content:       "# Parent",
		Priority:      scripts.P1,
	}

	// Create child todos
	child1 := scripts.File{
		Name:        "child1-2025-11-08.md",
		Title:       "Child 1",
		Tags:        []string{"todo"},
		CreatedAt:   time.Now(),
		Done:        false,
		ObjectiveID: objectiveID,
		Content:     "# Child 1",
		Priority:    scripts.P1,
	}

	child2 := scripts.File{
		Name:        "child2-2025-11-08.md",
		Title:       "Child 2",
		Tags:        []string{"todo"},
		CreatedAt:   time.Now(),
		Done:        true,
		ObjectiveID: objectiveID,
		Content:     "# Child 2",
		Priority:    scripts.P2,
	}

	// Create unrelated todo
	unrelated := scripts.File{
		Name:      "unrelated-2025-11-08.md",
		Title:     "Unrelated",
		Tags:      []string{"todo"},
		CreatedAt: time.Now(),
		Done:      false,
		Content:   "# Unrelated",
		Priority:  scripts.P3,
	}

	// Write files
	if err := WriteFile(parent); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(child1); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(child2); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(unrelated); err != nil {
		t.Fatal(err)
	}

	// Test: Query children excluding done
	children, err := QueryChildrenByObjectiveID(objectiveID, false)
	if err != nil {
		t.Fatalf("Error querying children: %v", err)
	}

	if len(children) != 1 {
		t.Errorf("Expected 1 incomplete child, got %d", len(children))
	}

	// Test: Query children including done
	allChildren, err := QueryChildrenByObjectiveID(objectiveID, true)
	if err != nil {
		t.Fatalf("Error querying all children: %v", err)
	}

	if len(allChildren) != 2 {
		t.Errorf("Expected 2 total children, got %d", len(allChildren))
	}
}

func TestGetObjectiveByID(t *testing.T) {
	_, cleanup := setupTestObjectives(t)
	defer cleanup()

	objectiveID := "find1234"

	// Create objective
	objective := scripts.File{
		Name:          "findme-2025-11-08.md",
		Title:         "Find Me",
		Tags:          []string{"objective"},
		CreatedAt:     time.Now(),
		ObjectiveRole: "parent",
		ObjectiveID:   objectiveID,
		Content:       "# Find Me",
		Priority:      scripts.P1,
	}

	if err := WriteFile(objective); err != nil {
		t.Fatal(err)
	}

	// Test: Find existing objective
	found, err := GetObjectiveByID(objectiveID)
	if err != nil {
		t.Fatalf("Error getting objective: %v", err)
	}

	if found == nil {
		t.Fatal("Expected to find objective, got nil")
	}

	if found.ObjectiveID != objectiveID {
		t.Errorf("Expected ObjectiveID %s, got %s", objectiveID, found.ObjectiveID)
	}

	if found.Title != "Find Me" {
		t.Errorf("Expected Title 'Find Me', got '%s'", found.Title)
	}

	// Test: Non-existent objective
	notFound, err := GetObjectiveByID("nonexist")
	if err != nil {
		t.Fatalf("Error getting non-existent objective: %v", err)
	}

	if notFound != nil {
		t.Error("Expected nil for non-existent objective")
	}
}

func TestQueryTodosWithoutObjective(t *testing.T) {
	_, cleanup := setupTestObjectives(t)
	defer cleanup()

	// Create parent objective
	parent := scripts.File{
		Name:          "parent-2025-11-08.md",
		Title:         "Parent",
		Tags:          []string{"objective"},
		CreatedAt:     time.Now(),
		ObjectiveRole: "parent",
		ObjectiveID:   "parent123",
		Content:       "# Parent",
		Priority:      scripts.P1,
	}

	// Create linked child
	linkedChild := scripts.File{
		Name:        "linked-2025-11-08.md",
		Title:       "Linked Child",
		Tags:        []string{"todo", "api"},
		CreatedAt:   time.Now(),
		Done:        false,
		ObjectiveID: "parent123",
		Content:     "# Linked Child\n\nAPI implementation",
		Priority:    scripts.P1,
	}

	// Create unlinked todos
	unlinked1 := scripts.File{
		Name:      "unlinked1-2025-11-08.md",
		Title:     "Unlinked API Todo",
		Tags:      []string{"todo", "api"},
		CreatedAt: time.Now(),
		Done:      false,
		Content:   "# Unlinked API Todo",
		Priority:  scripts.P2,
	}

	unlinked2 := scripts.File{
		Name:      "unlinked2-2025-11-08.md",
		Title:     "Unlinked Frontend Todo",
		Tags:      []string{"todo", "frontend"},
		CreatedAt: time.Now(),
		Done:      false,
		Content:   "# Unlinked Frontend Todo",
		Priority:  scripts.P3,
	}

	// Write files
	if err := WriteFile(parent); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(linkedChild); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(unlinked1); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(unlinked2); err != nil {
		t.Fatal(err)
	}

	// Test: Query all unlinked todos
	allUnlinked, err := QueryTodosWithoutObjective([]string{})
	if err != nil {
		t.Fatalf("Error querying unlinked todos: %v", err)
	}

	if len(allUnlinked) != 2 {
		t.Errorf("Expected 2 unlinked todos, got %d", len(allUnlinked))
	}

	// Test: Query with filter
	apiTodos, err := QueryTodosWithoutObjective([]string{"api"})
	if err != nil {
		t.Fatalf("Error querying API todos: %v", err)
	}

	if len(apiTodos) != 1 {
		t.Errorf("Expected 1 unlinked API todo, got %d", len(apiTodos))
	}

	if len(apiTodos) > 0 && apiTodos[0].Title != "Unlinked API Todo" {
		t.Errorf("Expected 'Unlinked API Todo', got '%s'", apiTodos[0].Title)
	}
}

func TestGetCompletionStats(t *testing.T) {
	_, cleanup := setupTestObjectives(t)
	defer cleanup()

	objectiveID := "stats123"

	// Create parent
	parent := scripts.File{
		Name:          "parent-2025-11-08.md",
		Title:         "Parent",
		Tags:          []string{"objective"},
		CreatedAt:     time.Now(),
		ObjectiveRole: "parent",
		ObjectiveID:   objectiveID,
		Content:       "# Parent",
		Priority:      scripts.P1,
	}

	// Create children - 2 complete, 3 incomplete
	children := []scripts.File{
		{
			Name:        "child1-2025-11-08.md",
			Title:       "Child 1",
			Tags:        []string{"todo"},
			CreatedAt:   time.Now(),
			Done:        true,
			ObjectiveID: objectiveID,
			Content:     "# Child 1",
			Priority:    scripts.P1,
		},
		{
			Name:        "child2-2025-11-08.md",
			Title:       "Child 2",
			Tags:        []string{"todo"},
			CreatedAt:   time.Now(),
			Done:        true,
			ObjectiveID: objectiveID,
			Content:     "# Child 2",
			Priority:    scripts.P1,
		},
		{
			Name:        "child3-2025-11-08.md",
			Title:       "Child 3",
			Tags:        []string{"todo"},
			CreatedAt:   time.Now(),
			Done:        false,
			ObjectiveID: objectiveID,
			Content:     "# Child 3",
			Priority:    scripts.P2,
		},
		{
			Name:        "child4-2025-11-08.md",
			Title:       "Child 4",
			Tags:        []string{"todo"},
			CreatedAt:   time.Now(),
			Done:        false,
			ObjectiveID: objectiveID,
			Content:     "# Child 4",
			Priority:    scripts.P2,
		},
		{
			Name:        "child5-2025-11-08.md",
			Title:       "Child 5",
			Tags:        []string{"todo"},
			CreatedAt:   time.Now(),
			Done:        false,
			ObjectiveID: objectiveID,
			Content:     "# Child 5",
			Priority:    scripts.P3,
		},
	}

	// Write files
	if err := WriteFile(parent); err != nil {
		t.Fatal(err)
	}
	for _, child := range children {
		if err := WriteFile(child); err != nil {
			t.Fatal(err)
		}
	}

	// Test completion stats
	complete, total, err := GetCompletionStats(objectiveID)
	if err != nil {
		t.Fatalf("Error getting completion stats: %v", err)
	}

	if complete != 2 {
		t.Errorf("Expected 2 complete, got %d", complete)
	}

	if total != 5 {
		t.Errorf("Expected 5 total, got %d", total)
	}
}
