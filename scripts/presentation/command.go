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
	onBackSpace func(),
) Command {
	switch key {
	case keyboard.KeyArrowUp:
		return FileSelectedWIPCommand{
			WIPCommand: WIPCommand{
				Text:         "",
				SelectedFile: selectNextFile(),
			},
		}

	case keyboard.KeyArrowDown:
		return FileSelectedWIPCommand{
			WIPCommand: WIPCommand{
				Text:         "",
				SelectedFile: selectPrevFile(),
			},
		}

	case keyboard.KeyEnter:
		return toCompletedCommand(currentCommand)

	case keyboard.KeyBackspace, keyboard.KeyBackspace2:
		text := currentCommand.Text
		if len(text) > 0 {
			text = text[:len(text)-1]
			return BackSpacedWIPCommand{
				WIPCommand: WIPCommand{
					Text:         text,
					SelectedFile: currentCommand.SelectedFile,
				},
			}
		} else {
			return currentCommand
		}

	case keyboard.KeySpace:
		return SpacedWIPCommand{
			WIPCommand: WIPCommand{
				Text:         currentCommand.Text + " ",
				SelectedFile: currentCommand.SelectedFile,
			},
		}

	case keyboard.KeyEsc:
		return ResetCommand{}

	default:
		return WIPCommand{
			Text:         currentCommand.Text + string(char),
			SelectedFile: currentCommand.SelectedFile,
		}
	}
}

func toCompletedCommand(wip WIPCommand) CompletedCommand {
	parts := strings.Split(wip.Text, " ")
	name := parts[0]

	remaining := strings.Join(parts[1:], " ")

	queries := strings.Split(remaining, ",")

	for i, query := range queries {
		queries[i] = strings.TrimSpace(query)
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
