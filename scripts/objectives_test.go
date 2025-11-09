package scripts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "objectives-test-*")
	if err != nil {
		t.Fatal(err)
	}

	notesDir := filepath.Join(tempDir, "notes")
	if err := os.Mkdir(notesDir, 0755); err != nil {
		t.Fatal(err)
	}

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

func TestCreateParentObjective(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	var createdFile File
	onFileCreated := func(f File) error {
		createdFile = f
		return nil
	}

	title := "Test Objective"
	objective, err := CreateParentObjective(title, onFileCreated)
	if err != nil {
		t.Fatalf("Error creating parent objective: %v", err)
	}

	// Verify title
	if objective.Title != title {
		t.Errorf("Expected title '%s', got '%s'", title, objective.Title)
	}

	// Verify ObjectiveRole
	if objective.ObjectiveRole != "parent" {
		t.Errorf("Expected ObjectiveRole 'parent', got '%s'", objective.ObjectiveRole)
	}

	// Verify ObjectiveID is generated and valid
	if objective.ObjectiveID == "" {
		t.Error("ObjectiveID should not be empty")
	}
	if !ValidateObjectiveID(objective.ObjectiveID) {
		t.Errorf("ObjectiveID '%s' is not valid", objective.ObjectiveID)
	}

	// Verify tags include "objective"
	hasObjectiveTag := false
	for _, tag := range objective.Tags {
		if tag == "objective" {
			hasObjectiveTag = true
			break
		}
	}
	if !hasObjectiveTag {
		t.Error("Expected 'objective' tag in tags")
	}

	// Verify content has structure
	if objective.Content == "" {
		t.Error("Content should not be empty")
	}

	// Verify priority defaults to P2
	if objective.Priority != P2 {
		t.Errorf("Expected priority P2, got P%d", objective.Priority)
	}

	// Verify callback was called
	if createdFile.Title != title {
		t.Error("onFileCreated callback was not called properly")
	}
}

func TestConvertTodoToParentObjective(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	// Mock writeFile and readLatestFileContent
	writeFile := func(f File) error {
		return nil
	}

	// Override readLatestFileContent for testing
	originalReadLatest := readLatestFileContent
	defer func() { readLatestFileContent = originalReadLatest }()
	readLatestFileContent = func(f File) (File, error) {
		return f, nil
	}

	// Create a regular todo
	todo := File{
		Name:      "test-todo-2025-11-08.md",
		Title:     "Test Todo",
		Tags:      []string{"todo", "api"},
		CreatedAt: time.Now(),
		Done:      false,
		Content:   "# Test Todo\n\nSome content",
		Priority:  P1,
	}

	// Convert to parent objective
	objective, err := ConvertTodoToParentObjective(todo, writeFile)
	if err != nil {
		t.Fatalf("Error converting to parent objective: %v", err)
	}

	// Verify ObjectiveRole is set
	if objective.ObjectiveRole != "parent" {
		t.Errorf("Expected ObjectiveRole 'parent', got '%s'", objective.ObjectiveRole)
	}

	// Verify ObjectiveID is generated
	if objective.ObjectiveID == "" {
		t.Error("ObjectiveID should not be empty")
	}

	// Verify original content is preserved
	if objective.Content != todo.Content {
		t.Error("Content should be preserved")
	}

	// Verify "objective" tag is added
	hasObjectiveTag := false
	for _, tag := range objective.Tags {
		if tag == "objective" {
			hasObjectiveTag = true
			break
		}
	}
	if !hasObjectiveTag {
		t.Error("Expected 'objective' tag to be added")
	}

	// Verify original tags are preserved
	hasApiTag := false
	for _, tag := range objective.Tags {
		if tag == "api" {
			hasApiTag = true
			break
		}
	}
	if !hasApiTag {
		t.Error("Original tags should be preserved")
	}
}

func TestConvertChildToParentObjective(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	writeFile := func(f File) error {
		return nil
	}

	originalReadLatest := readLatestFileContent
	defer func() { readLatestFileContent = originalReadLatest }()
	readLatestFileContent = func(f File) (File, error) {
		return f, nil
	}

	// Create a child todo with existing ObjectiveID
	childTodo := File{
		Name:        "child-todo-2025-11-08.md",
		Title:       "Child Todo",
		Tags:        []string{"todo"},
		CreatedAt:   time.Now(),
		Done:        false,
		Content:     "# Child Todo",
		Priority:    P2,
		ObjectiveID: "oldparent",
	}

	// Convert to parent objective
	objective, err := ConvertTodoToParentObjective(childTodo, writeFile)
	if err != nil {
		t.Fatalf("Error converting child to parent: %v", err)
	}

	// Verify old ObjectiveID is replaced
	if objective.ObjectiveID == "oldparent" {
		t.Error("Old ObjectiveID should be replaced")
	}

	// Verify new ObjectiveID is generated
	if objective.ObjectiveID == "" {
		t.Error("New ObjectiveID should be generated")
	}

	// Verify ObjectiveRole is set
	if objective.ObjectiveRole != "parent" {
		t.Errorf("Expected ObjectiveRole 'parent', got '%s'", objective.ObjectiveRole)
	}
}

func TestCreateChildTodo(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	onFileCreated := func(f File) error {
		return nil
	}

	// Create parent objective
	parent := File{
		Name:          "parent-2025-11-08.md",
		Title:         "Parent",
		Tags:          []string{"objective", "frontend", "api"},
		CreatedAt:     time.Now(),
		ObjectiveRole: "parent",
		ObjectiveID:   "parent123",
		Content:       "# Parent",
		Priority:      P1,
	}

	title := "Child Todo"
	child, err := CreateChildTodo(title, parent, onFileCreated)
	if err != nil {
		t.Fatalf("Error creating child todo: %v", err)
	}

	// Verify title
	if child.Title != title {
		t.Errorf("Expected title '%s', got '%s'", title, child.Title)
	}

	// Verify ObjectiveID matches parent
	if child.ObjectiveID != parent.ObjectiveID {
		t.Errorf("Expected ObjectiveID '%s', got '%s'", parent.ObjectiveID, child.ObjectiveID)
	}

	// Verify ObjectiveRole is NOT set (child is not a parent)
	if child.ObjectiveRole != "" {
		t.Errorf("Expected empty ObjectiveRole, got '%s'", child.ObjectiveRole)
	}

	// Verify tags are inherited (excluding "objective")
	hasTodoTag := false
	hasFrontendTag := false
	hasApiTag := false
	hasObjectiveTag := false

	for _, tag := range child.Tags {
		switch tag {
		case "todo":
			hasTodoTag = true
		case "frontend":
			hasFrontendTag = true
		case "api":
			hasApiTag = true
		case "objective":
			hasObjectiveTag = true
		}
	}

	if !hasTodoTag {
		t.Error("Expected 'todo' tag")
	}
	if !hasFrontendTag {
		t.Error("Expected inherited 'frontend' tag")
	}
	if !hasApiTag {
		t.Error("Expected inherited 'api' tag")
	}
	if hasObjectiveTag {
		t.Error("Should NOT inherit 'objective' tag")
	}
}

func TestLinkTodoToObjective(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	var writtenFile File
	writeFile := func(f File) error {
		writtenFile = f
		return nil
	}

	originalReadLatest := readLatestFileContent
	defer func() { readLatestFileContent = originalReadLatest }()
	readLatestFileContent = func(f File) (File, error) {
		return f, nil
	}

	// Create parent
	parent := File{
		Name:          "parent-2025-11-08.md",
		Title:         "Parent",
		Tags:          []string{"objective", "frontend"},
		ObjectiveRole: "parent",
		ObjectiveID:   "parent123",
	}

	// Create unlinked todo
	todo := File{
		Name:  "todo-2025-11-08.md",
		Title: "Todo",
		Tags:  []string{"todo", "backend"},
	}

	// Link todo to objective
	err := LinkTodoToObjective(todo, parent, writeFile)
	if err != nil {
		t.Fatalf("Error linking todo: %v", err)
	}

	// Verify ObjectiveID is set
	if writtenFile.ObjectiveID != parent.ObjectiveID {
		t.Errorf("Expected ObjectiveID '%s', got '%s'", parent.ObjectiveID, writtenFile.ObjectiveID)
	}

	// Verify tags are inherited
	hasFrontendTag := false
	hasBackendTag := false
	for _, tag := range writtenFile.Tags {
		if tag == "frontend" {
			hasFrontendTag = true
		}
		if tag == "backend" {
			hasBackendTag = true
		}
	}

	if !hasFrontendTag {
		t.Error("Expected inherited 'frontend' tag")
	}
	if !hasBackendTag {
		t.Error("Expected original 'backend' tag to be preserved")
	}
}

func TestLinkTodoToObjective_PreventParentAsChild(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	writeFile := func(f File) error {
		return nil
	}

	// Create two parent objectives
	parent1 := File{
		Name:          "parent1-2025-11-08.md",
		Title:         "Parent 1",
		Tags:          []string{"objective"},
		ObjectiveRole: "parent",
		ObjectiveID:   "parent123",
	}

	parent2 := File{
		Name:          "parent2-2025-11-08.md",
		Title:         "Parent 2",
		Tags:          []string{"objective"},
		ObjectiveRole: "parent",
		ObjectiveID:   "parent456",
	}

	// Try to link parent2 as child of parent1
	err := LinkTodoToObjective(parent2, parent1, writeFile)
	if err == nil {
		t.Error("Expected error when linking parent as child, got nil")
	}

	expectedError := "cannot link parent objectives as children"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestUnlinkTodoFromObjective(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	var writtenFile File
	writeFile := func(f File) error {
		writtenFile = f
		return nil
	}

	originalReadLatest := readLatestFileContent
	defer func() { readLatestFileContent = originalReadLatest }()
	readLatestFileContent = func(f File) (File, error) {
		return f, nil
	}

	// Create linked todo
	linkedTodo := File{
		Name:        "linked-2025-11-08.md",
		Title:       "Linked Todo",
		Tags:        []string{"todo", "frontend"},
		ObjectiveID: "parent123",
	}

	// Unlink todo
	err := UnlinkTodoFromObjective(linkedTodo, writeFile)
	if err != nil {
		t.Fatalf("Error unlinking todo: %v", err)
	}

	// Verify ObjectiveID is cleared
	if writtenFile.ObjectiveID != "" {
		t.Errorf("Expected empty ObjectiveID, got '%s'", writtenFile.ObjectiveID)
	}

	// Verify other fields are preserved
	if writtenFile.Title != linkedTodo.Title {
		t.Error("Title should be preserved")
	}
}

func TestDeleteParentObjective(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create actual file for deletion test
	notesPath := filepath.Join(tempDir, "notes")
	parentPath := filepath.Join(notesPath, "parent-2025-11-08.md")
	if err := os.WriteFile(parentPath, []byte("# Parent"), 0644); err != nil {
		t.Fatal(err)
	}

	// Override readLatestFileContent to avoid file read issues
	originalReadLatest := readLatestFileContent
	defer func() { readLatestFileContent = originalReadLatest }()
	readLatestFileContent = func(f File) (File, error) {
		return f, nil
	}

	var unlinkedChildren []File
	writeFile := func(f File) error {
		unlinkedChildren = append(unlinkedChildren, f)
		return nil
	}

	// Mock getChildrenFunc
	getChildrenFunc := func(objectiveID string, includeDone bool) ([]File, error) {
		return []File{
			{
				Name:        "child1-2025-11-08.md",
				Title:       "Child 1",
				ObjectiveID: objectiveID,
			},
			{
				Name:        "child2-2025-11-08.md",
				Title:       "Child 2",
				ObjectiveID: objectiveID,
			},
		}, nil
	}

	parent := File{
		Name:          "parent-2025-11-08.md",
		Title:         "Parent",
		ObjectiveRole: "parent",
		ObjectiveID:   "parent123",
	}

	// Delete parent objective
	err := DeleteParentObjective(parent, getChildrenFunc, writeFile)
	if err != nil {
		t.Fatalf("Error deleting parent objective: %v", err)
	}

	// Verify children were unlinked
	if len(unlinkedChildren) != 2 {
		t.Errorf("Expected 2 children to be unlinked, got %d", len(unlinkedChildren))
	}

	for _, child := range unlinkedChildren {
		if child.ObjectiveID != "" {
			t.Errorf("Expected child ObjectiveID to be empty, got '%s'", child.ObjectiveID)
		}
	}

	// Verify parent file was deleted
	if _, err := os.Stat(parentPath); !os.IsNotExist(err) {
		t.Error("Expected parent file to be deleted")
	}
}

func TestDeleteParentObjective_NotParent(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	writeFile := func(f File) error {
		return nil
	}

	getChildrenFunc := func(objectiveID string, includeDone bool) ([]File, error) {
		return []File{}, nil
	}

	// Try to delete a regular todo (not a parent)
	regularTodo := File{
		Name:  "todo-2025-11-08.md",
		Title: "Regular Todo",
	}

	err := DeleteParentObjective(regularTodo, getChildrenFunc, writeFile)
	if err == nil {
		t.Error("Expected error when deleting non-parent file, got nil")
	}

	expectedError := "file is not a parent objective"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}
