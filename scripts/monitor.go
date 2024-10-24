package scripts

import (
	"os"
	"path/filepath"
	"time"
)

const MONITOR_DELAY_TIME_SECONDS = 30

var directoryPreviousSize int64 = 0

func MonitorDirectorySize(
	dirPath string,
	onDirectorySizeChanged func(),
) {
	for {
		fileSizeBytes, err := getDirectorySizeInBytes(dirPath)
		if err != nil {
			panic("Error getting directory size: ")
		}

		if directorySizeHasChanged(fileSizeBytes) {
			onDirectorySizeChanged()
		}

		time.Sleep(MONITOR_DELAY_TIME_SECONDS * time.Second)
	}
}

func directorySizeHasChanged(fileSizeBytes int64) bool {
	hasChanged := false

	if directoryPreviousSize != 0 && fileSizeBytes != directoryPreviousSize {
		hasChanged = true
	}

	directoryPreviousSize = fileSizeBytes

	return hasChanged
}

func getDirectorySizeInBytes(dirPath string) (int64, error) {
	var size int64
	err := filepath.Walk(dirPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == ".md" {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
