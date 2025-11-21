package presentation

import (
	"cli-notes/scripts"
	"errors"
	"reflect"
	"testing"

	"github.com/eiannone/keyboard"
)

func TestCommandHandler_ArrowKeys(t *testing.T) {
	// Mock file selection functions
	selectNextFile := func() scripts.File {
		return scripts.File{Name: "nextFile.md"}
	}
	
	selectPrevFile := func() scripts.File {
		return scripts.File{Name: "prevFile.md"}
	}
	
	getTasksInFile := func(file scripts.File) ([]string, error) {
		if file.Name == "nextFile.md" {
			return []string{"Task 1", "Task 2"}, nil
		} else if file.Name == "prevFile.md" {
			return []string{"Task 3", "Task 4"}, nil
		}
		return nil, errors.New("file not found")
	}
	
	onBackSpace := func() {}
	
	// Test arrow up
	currentCommand := WIPCommand{Text: "test", SelectedFile: scripts.File{}}
	result, err := CommandHandler(0, keyboard.KeyArrowUp, currentCommand, selectNextFile, selectPrevFile, getTasksInFile, onBackSpace)
	
	if err != nil {
		t.Errorf("Arrow up should not return error, got: %v", err)
	}
	
	fileSelected, ok := result.(FileSelectedWIPCommand)
	if !ok {
		t.Errorf("Expected FileSelectedWIPCommand, got: %T", result)
	}
	
	if fileSelected.SelectedFile.Name != "nextFile.md" || !reflect.DeepEqual(fileSelected.Tasks, []string{"Task 1", "Task 2"}) {
		t.Errorf("Arrow up didn't select correct file or tasks, got: %+v", fileSelected)
	}
	
	// Check that the text is cleared
	if fileSelected.Text != "" {
		t.Errorf("Arrow up should clear the text, got: %s", fileSelected.Text)
	}
	
	// Test arrow down
	result, err = CommandHandler(0, keyboard.KeyArrowDown, currentCommand, selectNextFile, selectPrevFile, getTasksInFile, onBackSpace)
	
	if err != nil {
		t.Errorf("Arrow down should not return error, got: %v", err)
	}
	
	fileSelected, ok = result.(FileSelectedWIPCommand)
	if !ok {
		t.Errorf("Expected FileSelectedWIPCommand, got: %T", result)
	}
	
	if fileSelected.SelectedFile.Name != "prevFile.md" || !reflect.DeepEqual(fileSelected.Tasks, []string{"Task 3", "Task 4"}) {
		t.Errorf("Arrow down didn't select correct file or tasks, got: %+v", fileSelected)
	}
	
	// Check that the text is cleared
	if fileSelected.Text != "" {
		t.Errorf("Arrow down should clear the text, got: %s", fileSelected.Text)
	}
}

func TestCommandHandler_ErrorHandling(t *testing.T) {
	// Mock that returns an error
	getTasksInFile := func(file scripts.File) ([]string, error) {
		return nil, errors.New("mock error")
	}
	
	currentCommand := WIPCommand{Text: "test", SelectedFile: scripts.File{}}
	_, err := CommandHandler(0, keyboard.KeyArrowUp, currentCommand, 
		func() scripts.File { return scripts.File{} }, 
		func() scripts.File { return scripts.File{} }, 
		getTasksInFile, 
		func() {})
	
	if err == nil || err.Error() != "mock error" {
		t.Errorf("Expected mock error, got: %v", err)
	}
}

func TestCommandHandler_EnterKey(t *testing.T) {
	currentCommand := WIPCommand{
		Text: "add task one, task two",
		SelectedFile: scripts.File{Name: "notes.md"},
	}
	
	result, err := CommandHandler(0, keyboard.KeyEnter, currentCommand, 
		nil, nil, nil, nil)
	
	if err != nil {
		t.Errorf("Enter key should not return error, got: %v", err)
	}
	
	completed, ok := result.(CompletedCommand)
	if !ok {
		t.Errorf("Expected CompletedCommand, got: %T", result)
	}
	
	if completed.Name != "add" || 
	   !reflect.DeepEqual(completed.Queries, []string{"task one", "task two"}) ||
	   completed.SelectedFile.Name != "notes.md" {
		t.Errorf("Enter key didn't convert WIP to completed correctly, got: %+v", completed)
	}
}

func TestCommandHandler_BackspaceKey(t *testing.T) {
	// Test backspace with text
	currentCommand := WIPCommand{
		Text: "test",
		SelectedFile: scripts.File{Name: "notes.md"},
	}
	
	result, err := CommandHandler(0, keyboard.KeyBackspace, currentCommand, 
		nil, nil, nil, func() {})
	
	if err != nil {
		t.Errorf("Backspace should not return error, got: %v", err)
	}
	
	backspaced, ok := result.(BackSpacedWIPCommand)
	if !ok {
		t.Errorf("Expected BackSpacedWIPCommand, got: %T", result)
	}
	
	if backspaced.Text != "tes" {
		t.Errorf("Backspace didn't remove last character, got: %s", backspaced.Text)
	}
	
	// Test backspace with empty text
	currentCommand = WIPCommand{
		Text: "",
		SelectedFile: scripts.File{Name: "notes.md"},
	}
	
	result, err = CommandHandler(0, keyboard.KeyBackspace, currentCommand, 
		nil, nil, nil, func() {})
	
	if err != nil {
		t.Errorf("Backspace with empty text should not return error, got: %v", err)
	}
	
	// Should return unchanged WIPCommand
	_, ok = result.(WIPCommand)
	if !ok {
		t.Errorf("Expected unchanged WIPCommand, got: %T", result)
	}
}

func TestCommandHandler_SpaceKey(t *testing.T) {
	currentCommand := WIPCommand{
		Text: "test",
		SelectedFile: scripts.File{Name: "notes.md"},
	}
	
	result, err := CommandHandler(0, keyboard.KeySpace, currentCommand, 
		nil, nil, nil, nil)
	
	if err != nil {
		t.Errorf("Space key should not return error, got: %v", err)
	}
	
	spaced, ok := result.(SpacedWIPCommand)
	if !ok {
		t.Errorf("Expected SpacedWIPCommand, got: %T", result)
	}
	
	if spaced.Text != "test " {
		t.Errorf("Space key didn't add space character, got: %s", spaced.Text)
	}
}

func TestCommandHandler_EscapeKey(t *testing.T) {
	currentCommand := WIPCommand{
		Text: "test",
		SelectedFile: scripts.File{Name: "notes.md"},
	}
	
	result, err := CommandHandler(0, keyboard.KeyEsc, currentCommand, 
		nil, nil, nil, nil)
	
	if err != nil {
		t.Errorf("Escape key should not return error, got: %v", err)
	}
	
	_, ok := result.(ResetCommand)
	if !ok {
		t.Errorf("Expected ResetCommand, got: %T", result)
	}
}

func TestCommandHandler_DefaultCase(t *testing.T) {
	currentCommand := WIPCommand{
		Text: "test",
		SelectedFile: scripts.File{Name: "notes.md"},
	}
	
	result, err := CommandHandler('a', 0, currentCommand, 
		nil, nil, nil, nil)
	
	if err != nil {
		t.Errorf("Default case should not return error, got: %v", err)
	}
	
	wip, ok := result.(WIPCommand)
	if !ok {
		t.Errorf("Expected WIPCommand, got: %T", result)
	}
	
	if wip.Text != "testa" {
		t.Errorf("Default case didn't add character, got: %s", wip.Text)
	}
}

func TestToCompletedCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    WIPCommand
		expected CompletedCommand
	}{
		{
			name: "Basic command with queries",
			input: WIPCommand{
				Text: "add task one, task two",
				SelectedFile: scripts.File{Name: "notes.md"},
			},
			expected: CompletedCommand{
				Name: "add",
				Queries: []string{"task one", "task two"},
				SelectedFile: scripts.File{Name: "notes.md"},
			},
		},
		{
			name: "Command with no queries",
			input: WIPCommand{
				Text: "list",
				SelectedFile: scripts.File{Name: "notes.md"},
			},
			expected: CompletedCommand{
				Name: "list",
				Queries: []string{""},
				SelectedFile: scripts.File{Name: "notes.md"},
			},
		},
		{
			name: "Command with no selected file",
			input: WIPCommand{
				Text: "add task",
				SelectedFile: scripts.File{},
			},
			expected: CompletedCommand{
				Name: "add",
				Queries: []string{"task"},
				SelectedFile: scripts.File{},
			},
		},
		{
			name: "Command with extra spaces",
			input: WIPCommand{
				Text: "add  task one ,  task two ",
				SelectedFile: scripts.File{Name: "notes.md"},
			},
			expected: CompletedCommand{
				Name: "add",
				Queries: []string{"task one", "task two"},
				SelectedFile: scripts.File{Name: "notes.md"},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToCompletedCommand(tt.input)
			
			if result.Name != tt.expected.Name {
				t.Errorf("Expected name %s, got %s", tt.expected.Name, result.Name)
			}
			
			if !reflect.DeepEqual(result.Queries, tt.expected.Queries) {
				t.Errorf("Expected queries %v, got %v", tt.expected.Queries, result.Queries)
			}
			
			if result.SelectedFile.Name != tt.expected.SelectedFile.Name {
				t.Errorf("Expected selected file %s, got %s",
					tt.expected.SelectedFile.Name, result.SelectedFile.Name)
			}
		})
	}
}

func TestCommandHandler_SingleKeyPriorityCommands(t *testing.T) {
	tests := []struct {
		name         string
		char         rune
		expectedName string
	}{
		{
			name:         "Press '1' with no command text becomes p1",
			char:         '1',
			expectedName: "p1",
		},
		{
			name:         "Press '2' with no command text becomes p2",
			char:         '2',
			expectedName: "p2",
		},
		{
			name:         "Press '3' with no command text becomes p3",
			char:         '3',
			expectedName: "p3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentCommand := WIPCommand{
				Text:         "",
				SelectedFile: scripts.File{Name: "test.md"},
			}

			result, err := CommandHandler(tt.char, 0, currentCommand, nil, nil, nil, nil)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			completed, ok := result.(CompletedCommand)
			if !ok {
				t.Errorf("Expected CompletedCommand, got: %T", result)
			}

			if completed.Name != tt.expectedName {
				t.Errorf("Expected command name %s, got %s", tt.expectedName, completed.Name)
			}

			if completed.SelectedFile.Name != "test.md" {
				t.Errorf("Expected selected file preserved, got: %s", completed.SelectedFile.Name)
			}

			if len(completed.Queries) != 0 {
				t.Errorf("Expected empty queries, got: %v", completed.Queries)
			}
		})
	}
}

func TestCommandHandler_NumbersInCommandText(t *testing.T) {
	// Verify numbers are still added to command text when user is typing
	currentCommand := WIPCommand{
		Text:         "gt",
		SelectedFile: scripts.File{},
	}

	result, err := CommandHandler('1', 0, currentCommand, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	wip, ok := result.(WIPCommand)
	if !ok {
		t.Errorf("Expected WIPCommand when adding to existing text, got: %T", result)
	}

	if wip.Text != "gt1" {
		t.Errorf("Expected text 'gt1', got '%s'", wip.Text)
	}

	// Test with '2' and '3' as well
	result, err = CommandHandler('2', 0, currentCommand, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	wip, ok = result.(WIPCommand)
	if !ok {
		t.Errorf("Expected WIPCommand when adding to existing text, got: %T", result)
	}

	if wip.Text != "gt2" {
		t.Errorf("Expected text 'gt2', got '%s'", wip.Text)
	}
}