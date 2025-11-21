package presentation

import (
	"os"
	"os/exec"
)

func OpenNoteInEditor(filePath string, onKeyboardClose func(), onKeyboardReopen func()) error {
	onKeyboardClose()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}
	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	onKeyboardReopen()

	if err != nil {
		return err
	}
	return nil
}
