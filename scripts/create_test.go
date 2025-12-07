package scripts

import (
	"strings"
	"testing"
)

func TestCreateTodoWithCheckboxes_NoCheckboxes(t *testing.T) {
	var createdFile File
	onFileCreated := func(f File) error {
		createdFile = f
		return nil
	}

	title := "simple-todo"
	_, err := CreateTodoWithCheckboxes(title, []string{}, onFileCreated)
	if err != nil {
		t.Fatalf("Error creating todo: %v", err)
	}

	// Verify title
	if createdFile.Title != title {
		t.Errorf("Expected title '%s', got '%s'", title, createdFile.Title)
	}

	// Verify content has title heading
	expectedContent := "# simple-todo"
	if createdFile.Content != expectedContent {
		t.Errorf("Expected content '%s', got '%s'", expectedContent, createdFile.Content)
	}

	// Verify no checkboxes
	if strings.Contains(createdFile.Content, "- [ ]") {
		t.Error("Content should not contain checkboxes")
	}
}

func TestCreateTodoWithCheckboxes_SingleCheckbox(t *testing.T) {
	var createdFile File
	onFileCreated := func(f File) error {
		createdFile = f
		return nil
	}

	title := "todo-with-one"
	checkboxItems := []string{"first task"}
	_, err := CreateTodoWithCheckboxes(title, checkboxItems, onFileCreated)
	if err != nil {
		t.Fatalf("Error creating todo: %v", err)
	}

	// Verify content has title heading
	if !strings.Contains(createdFile.Content, "# todo-with-one") {
		t.Error("Content should contain title heading")
	}

	// Verify checkbox
	if !strings.Contains(createdFile.Content, "- [ ] first task") {
		t.Error("Content should contain checkbox item")
	}
}

func TestCreateTodoWithCheckboxes_MultipleCheckboxes(t *testing.T) {
	var createdFile File
	onFileCreated := func(f File) error {
		createdFile = f
		return nil
	}

	title := "todo-with-many"
	checkboxItems := []string{"task one", "task two", "task three"}
	_, err := CreateTodoWithCheckboxes(title, checkboxItems, onFileCreated)
	if err != nil {
		t.Fatalf("Error creating todo: %v", err)
	}

	// Verify content has title heading
	if !strings.Contains(createdFile.Content, "# todo-with-many") {
		t.Error("Content should contain title heading")
	}

	// Verify all checkboxes
	if !strings.Contains(createdFile.Content, "- [ ] task one") {
		t.Error("Content should contain first checkbox item")
	}
	if !strings.Contains(createdFile.Content, "- [ ] task two") {
		t.Error("Content should contain second checkbox item")
	}
	if !strings.Contains(createdFile.Content, "- [ ] task three") {
		t.Error("Content should contain third checkbox item")
	}
}

func TestCreateTodoWithCheckboxes_WhitespaceHandling(t *testing.T) {
	var createdFile File
	onFileCreated := func(f File) error {
		createdFile = f
		return nil
	}

	title := "whitespace-test"
	checkboxItems := []string{"  leading space", "trailing space  ", "  both  "}
	_, err := CreateTodoWithCheckboxes(title, checkboxItems, onFileCreated)
	if err != nil {
		t.Fatalf("Error creating todo: %v", err)
	}

	// Verify whitespace is trimmed
	if !strings.Contains(createdFile.Content, "- [ ] leading space") {
		t.Error("Leading whitespace should be trimmed")
	}
	if !strings.Contains(createdFile.Content, "- [ ] trailing space") {
		t.Error("Trailing whitespace should be trimmed")
	}
	if !strings.Contains(createdFile.Content, "- [ ] both") {
		t.Error("Both leading and trailing whitespace should be trimmed")
	}

	// Ensure untrimmed versions are not present
	if strings.Contains(createdFile.Content, "- [ ]   leading") {
		t.Error("Leading whitespace should not be in content")
	}
}

func TestCreateTodoWithCheckboxes_EmptyItemsFiltered(t *testing.T) {
	var createdFile File
	onFileCreated := func(f File) error {
		createdFile = f
		return nil
	}

	title := "empty-filter"
	checkboxItems := []string{"valid", "", "  ", "also valid"}
	_, err := CreateTodoWithCheckboxes(title, checkboxItems, onFileCreated)
	if err != nil {
		t.Fatalf("Error creating todo: %v", err)
	}

	// Verify valid items are present
	if !strings.Contains(createdFile.Content, "- [ ] valid") {
		t.Error("Content should contain 'valid' checkbox item")
	}
	if !strings.Contains(createdFile.Content, "- [ ] also valid") {
		t.Error("Content should contain 'also valid' checkbox item")
	}

	// Count checkboxes - should be exactly 2
	checkboxCount := strings.Count(createdFile.Content, "- [ ]")
	if checkboxCount != 2 {
		t.Errorf("Expected 2 checkboxes, got %d", checkboxCount)
	}
}

func TestCreateTodoWithCheckboxes_VerifyTags(t *testing.T) {
	var createdFile File
	onFileCreated := func(f File) error {
		createdFile = f
		return nil
	}

	title := "tag-test"
	_, err := CreateTodoWithCheckboxes(title, []string{"item"}, onFileCreated)
	if err != nil {
		t.Fatalf("Error creating todo: %v", err)
	}

	// Verify todo tag is present
	hasTodoTag := false
	for _, tag := range createdFile.Tags {
		if tag == "todo" {
			hasTodoTag = true
			break
		}
	}
	if !hasTodoTag {
		t.Error("Expected 'todo' tag")
	}
}

func TestCreateTodoWithCheckboxes_ContentFormat(t *testing.T) {
	var createdFile File
	onFileCreated := func(f File) error {
		createdFile = f
		return nil
	}

	title := "format-test"
	checkboxItems := []string{"item one", "item two"}
	_, err := CreateTodoWithCheckboxes(title, checkboxItems, onFileCreated)
	if err != nil {
		t.Fatalf("Error creating todo: %v", err)
	}

	// Verify format: title, blank line, checkboxes
	lines := strings.Split(createdFile.Content, "\n")
	if len(lines) < 4 {
		t.Fatalf("Expected at least 4 lines, got %d", len(lines))
	}

	if lines[0] != "# format-test" {
		t.Errorf("Expected first line to be title heading, got '%s'", lines[0])
	}
	if lines[1] != "" {
		t.Errorf("Expected second line to be blank, got '%s'", lines[1])
	}
	if lines[2] != "- [ ] item one" {
		t.Errorf("Expected third line to be first checkbox, got '%s'", lines[2])
	}
	if lines[3] != "- [ ] item two" {
		t.Errorf("Expected fourth line to be second checkbox, got '%s'", lines[3])
	}
}
