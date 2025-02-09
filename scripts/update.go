package scripts


type WriteFile = func(File) error

func DelayDueDate(delayDays int, file File, writeFile WriteFile) error {
	file.DueAt = file.DueAt.AddDate(0, 0, delayDays)
	return writeFile(file)
}