package scripts

import (
	"time"
)

type WriteFile = func(File) error

func DelayDueDate(delayDays int, file File, writeFile WriteFile) error {
	today := time.Now()
	file.DueAt = today.AddDate(0, 0, delayDays)
	return writeFile(file)
}

func SetDueDateToToday(file File, writeFile WriteFile) error {
	file.DueAt = time.Now()
	return writeFile(file)
}