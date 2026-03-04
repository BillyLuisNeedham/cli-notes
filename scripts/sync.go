package scripts

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const SYNC_DELAY_TIME_SECONDS = 30

func StartBackupSync(srcDir string) {
	if os.Getenv("CLI_NOTES_TEST_MODE") == "true" {
		return
	}

	syncToBackup(srcDir)

	for {
		time.Sleep(SYNC_DELAY_TIME_SECONDS * time.Second)
		syncToBackup(srcDir)
	}
}

func RunFinalSync(srcDir string) {
	if os.Getenv("CLI_NOTES_TEST_MODE") == "true" {
		return
	}

	syncToBackup(srcDir)
}

func syncToBackup(srcDir string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Backup sync error: %v\n", err)
		return
	}

	dstDir := filepath.Join(homeDir, "Documents", "notes")
	_, err = SyncNotesToBackup(srcDir, dstDir)
	if err != nil {
		fmt.Printf("Backup sync error: %v\n", err)
	}
}

func SyncNotesToBackup(srcDir, dstDir string) (int, error) {
	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		return 0, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Track which source files exist for stale cleanup
	sourceFiles := make(map[string]bool)
	copied := 0

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		sourceFiles[relPath] = true
		dstPath := filepath.Join(dstDir, relPath)

		if shouldSkipCopy(path, dstPath, info) {
			return nil
		}

		err = copyFile(path, dstPath)
		if err != nil {
			return fmt.Errorf("failed to copy %s: %w", relPath, err)
		}
		copied++

		return nil
	})
	if err != nil {
		return copied, err
	}

	// Remove stale files from destination
	err = filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		relPath, err := filepath.Rel(dstDir, path)
		if err != nil {
			return err
		}

		if !sourceFiles[relPath] {
			return os.Remove(path)
		}

		return nil
	})

	return copied, err
}

func shouldSkipCopy(srcPath, dstPath string, srcInfo os.FileInfo) bool {
	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		return false // destination doesn't exist, need to copy
	}

	return dstInfo.Size() == srcInfo.Size() && !dstInfo.ModTime().Before(srcInfo.ModTime())
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
}
