package scripts

import (
	"fmt"
	"os"
	"os/exec"
)

func PushChangesToGit(directoryPath string) {

	// Store the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current directory: %v\n", err)
		return
	}

	// Change the current working directory to the specified directory
	err = os.Chdir(directoryPath)
	if err != nil {
		fmt.Printf("Failed to change directory: %v\n", err)
		return
	}

	// Execute "git pull"
	cmdPull := exec.Command("git", "pull")
	_, err = cmdPull.Output()
	if err != nil {
		fmt.Printf("Failed to execute 'git pull': %v\n", err)
		return
	}

	// Execute "git add ."
	cmdAdd := exec.Command("git", "add", ".")
	_, err = cmdAdd.Output()
	if err != nil {
		fmt.Printf("Failed to execute 'git add .': %v\n", err)
		return
	}

	// Execute "git commit -m 'auto-sync'"
	cmdCommit := exec.Command("git", "commit", "-m", "auto-sync")
	_, err = cmdCommit.Output()
	if err != nil {
		fmt.Printf("Failed to execute 'git commit': %v\n", err)
		return
	}

	// Execute "git push"
	cmdPush := exec.Command("git", "push")
	_, err = cmdPush.Output()
	if err != nil {
		fmt.Printf("Failed to execute 'git push': %v\n", err)
		return
	}

	// Change back to the original directory
	err = os.Chdir(currentDir)
	if err != nil {
		fmt.Printf("Failed to change back to the original directory: %v\n", err)
		return
	}
}
