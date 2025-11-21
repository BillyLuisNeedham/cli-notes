package presentation

import (
	"cli-notes/scripts"
	"strings"

	"github.com/eiannone/keyboard"
)

type Command interface {
	command()
}

type CompletedCommand struct {
	Name         string
	Queries      []string
	SelectedFile scripts.File
}

type WIPCommand struct {
	Text         string
	SelectedFile scripts.File
}

type ResetCommand struct{}

type BackSpacedWIPCommand struct {
	WIPCommand
}

type SpacedWIPCommand struct {
	WIPCommand
}

type FileSelectedWIPCommand struct {
	WIPCommand
	Tasks []string
}

func (WIPCommand) command()       {}
func (CompletedCommand) command() {}
func (ResetCommand) command()     {}

func CommandHandler(
	char rune,
	key keyboard.Key,
	currentCommand WIPCommand,
	selectNextFile func() scripts.File,
	selectPrevFile func() scripts.File,
	getTasksInFile func(scripts.File) ([]string, error),
	onBackSpace func(),
) (Command, error) {
	switch key {
	case keyboard.KeyArrowUp:
		file := selectNextFile()
		tasks, err := getTasksInFile(file)
		if err != nil {
			return nil, err
		}
		return FileSelectedWIPCommand{
			WIPCommand: WIPCommand{
				Text:         "",
				SelectedFile: file,
			},
			Tasks: tasks,
		}, nil

	case keyboard.KeyArrowDown:
		file := selectPrevFile()
		tasks, err := getTasksInFile(file)
		if err != nil {
			return nil, err
		}
		return FileSelectedWIPCommand{
			WIPCommand: WIPCommand{
				Text:         "",
				SelectedFile: file,
			},
			Tasks: tasks,
		}, nil

	case keyboard.KeyEnter:
		completed := ToCompletedCommand(currentCommand)
		return completed, nil

	case keyboard.KeyBackspace, keyboard.KeyBackspace2:
		text := currentCommand.Text
		if len(text) > 0 {
			text = text[:len(text)-1]
			return BackSpacedWIPCommand{
				WIPCommand: WIPCommand{
					Text:         text,
					SelectedFile: currentCommand.SelectedFile,
				},
			}, nil
		} else {
			return currentCommand, nil
		}

	case keyboard.KeySpace:
		return SpacedWIPCommand{
			WIPCommand: WIPCommand{
				Text:         currentCommand.Text + " ",
				SelectedFile: currentCommand.SelectedFile,
			},
		}, nil

	case keyboard.KeyEsc:
		return ResetCommand{}, nil

	default:
		// Handle j/k navigation when no command text is entered
		if currentCommand.Text == "" {
			if char == 'k' {
				// k navigates up (to next file)
				file := selectNextFile()
				tasks, err := getTasksInFile(file)
				if err != nil {
					return nil, err
				}
				return FileSelectedWIPCommand{
					WIPCommand: WIPCommand{
						Text:         "",
						SelectedFile: file,
					},
					Tasks: tasks,
				}, nil
			} else if char == 'j' {
				// j navigates down (to previous file)
				file := selectPrevFile()
				tasks, err := getTasksInFile(file)
				if err != nil {
					return nil, err
				}
				return FileSelectedWIPCommand{
					WIPCommand: WIPCommand{
						Text:         "",
						SelectedFile: file,
					},
					Tasks: tasks,
				}, nil
			} else if char == '1' || char == '2' || char == '3' {
				// Handle single-key priority commands (1/2/3)
				// When user presses 1/2/3 with no command text, set priority (like "p 1")
				// This matches the behavior in objectives view
				return CompletedCommand{
					Name:         "p",
					Queries:      []string{string(char)},
					SelectedFile: currentCommand.SelectedFile,
				}, nil
			}
		}

		return WIPCommand{
			Text:         currentCommand.Text + string(char),
			SelectedFile: currentCommand.SelectedFile,
		}, nil
	}
}

func ToCompletedCommand(wip WIPCommand) CompletedCommand {
	parts := strings.Split(wip.Text, " ")
	name := parts[0]

	remaining := strings.Join(parts[1:], " ")

	var queries []string
	
	// Special handling for gd command - use space-separated arguments
	if name == "gd" {
		// Split by spaces for date arguments
		spaceParts := strings.Fields(remaining)
		queries = make([]string, len(spaceParts))
		for i, part := range spaceParts {
			queries[i] = strings.TrimSpace(part)
		}
	} else {
		// Default behavior: split by commas for other commands
		queries = strings.Split(remaining, ",")
		for i, query := range queries {
			queries[i] = strings.TrimSpace(query)
		}
	}

	selectedFile := scripts.File{}
	if wip.SelectedFile.Name != "" {
		selectedFile = wip.SelectedFile
	}

	return CompletedCommand{
		Name:         name,
		Queries:      queries,
		SelectedFile: selectedFile,
	}
}
