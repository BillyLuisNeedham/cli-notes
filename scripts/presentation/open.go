package presentation

import (
	"os"
	"os/exec"
)


func OpenNoteInEditor(filePath string) error {
	cmd := exec.Command("nvim", filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
	}