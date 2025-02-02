package scripts


type WriteFile = func(File) error

// TODO add a get files by name function here
func DelayDueDate(delayDays int, file File, writeFile WriteFile) error {
	file.DueAt = file.DueAt.AddDate(0, 0, delayDays)
	return writeFile(file)
}