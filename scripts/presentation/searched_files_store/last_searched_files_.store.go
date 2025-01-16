package searched_files_store

var filesThatHaveBeenSearched []string = make([]string, 0)
var filesThatHaveBeenSearchedSelectedIndex = -1

func SetFilesSearched(files []string) {
	filesThatHaveBeenSearched = files
}

func GetFilesSearched() []string {
	return filesThatHaveBeenSearched
}

func GetNextFile() string {
	lengthOfFiles := len(filesThatHaveBeenSearched)
	if filesThatHaveBeenSearchedSelectedIndex == -1 && lengthOfFiles > 0 {
		filesThatHaveBeenSearchedSelectedIndex = lengthOfFiles - 1
		return filesThatHaveBeenSearched[lengthOfFiles-1]

	} else if filesThatHaveBeenSearchedSelectedIndex > 0 && lengthOfFiles > 0 {
		filesThatHaveBeenSearchedSelectedIndex--
		return filesThatHaveBeenSearched[filesThatHaveBeenSearchedSelectedIndex]
	} else if filesThatHaveBeenSearchedSelectedIndex == 0 && lengthOfFiles > 0 {
		return filesThatHaveBeenSearched[0]
	} else {
		return ""
	}
}

func GetPreviousFile() string {
	lengthOfFiles := len(filesThatHaveBeenSearched)
	if filesThatHaveBeenSearchedSelectedIndex == -1 && lengthOfFiles > 0 {
		filesThatHaveBeenSearchedSelectedIndex = 0
		return filesThatHaveBeenSearched[0]

	} else if filesThatHaveBeenSearchedSelectedIndex < lengthOfFiles-1 && lengthOfFiles > 0 {
		filesThatHaveBeenSearchedSelectedIndex++
		return filesThatHaveBeenSearched[filesThatHaveBeenSearchedSelectedIndex]
	} else if filesThatHaveBeenSearchedSelectedIndex == lengthOfFiles-1 && lengthOfFiles > 0 {
		return filesThatHaveBeenSearched[lengthOfFiles-1]
	} else {
		return ""
	}
}
