package presentation

import (
	"os/exec"
)


func OpenNoteInEditor(filePath string) error {
	err := exec.Command("cursor", filePath).Run()
	if err != nil {
		return err
	}
	return nil
	}