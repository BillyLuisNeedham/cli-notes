package scripts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const GIT_COMMIT_INTERVAL_SECONDS = 60

func StartGitVersioning(dirPath string) {
	if os.Getenv("CLI_NOTES_TEST_MODE") == "true" {
		return
	}

	err := InitGitRepo(dirPath)
	if err != nil {
		fmt.Printf("Git versioning init error: %v\n", err)
		return
	}

	for {
		time.Sleep(GIT_COMMIT_INTERVAL_SECONDS * time.Second)
		err := CommitChanges(dirPath)
		if err != nil {
			fmt.Printf("Git versioning error: %v\n", err)
		}
	}
}

func RunFinalGitCommit(dirPath string) {
	if os.Getenv("CLI_NOTES_TEST_MODE") == "true" {
		return
	}

	err := CommitChanges(dirPath)
	if err != nil {
		fmt.Printf("Git final commit error: %v\n", err)
	}
}

func InitGitRepo(dirPath string) error {
	gitDir := filepath.Join(dirPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return nil
	}

	err := runGit(dirPath, "init")
	if err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	gitignorePath := filepath.Join(dirPath, ".gitignore")
	err = os.WriteFile(gitignorePath, []byte("*\n!*.md\n!.gitignore\n"), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	err = runGit(dirPath, "add", ".")
	if err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	err = runGit(dirPath, "commit", "-m", "initial commit")
	if err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	return nil
}

func CommitChanges(dirPath string) error {
	err := runGit(dirPath, "add", ".")
	if err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	if !hasChanges(dirPath) {
		return nil
	}

	msg := fmt.Sprintf("auto: %s", time.Now().Format("2006-01-02 15:04:05"))
	err = runGit(dirPath, "commit", "-m", msg)
	if err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	return nil
}

func runGit(dirPath string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dirPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, output)
	}
	return nil
}

func hasChanges(dirPath string) bool {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = dirPath
	err := cmd.Run()
	return err != nil
}
