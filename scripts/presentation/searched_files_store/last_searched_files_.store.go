package searched_files_store

import "cli-notes/scripts"

var filesThatHaveBeenSearched = make([]scripts.File, 0)
var filesThatHaveBeenSearchedSelectedIndex = -1

func SetFilesSearched(files []scripts.File) {
	filesThatHaveBeenSearched = files
}

func GetFilesSearched() []scripts.File {
	return filesThatHaveBeenSearched
}

func GetNextFile() *scripts.File {
	lengthOfFiles := len(filesThatHaveBeenSearched)
	if filesThatHaveBeenSearchedSelectedIndex == -1 && lengthOfFiles > 0 {
		filesThatHaveBeenSearchedSelectedIndex = lengthOfFiles - 1
		return &filesThatHaveBeenSearched[lengthOfFiles-1]

	} else if filesThatHaveBeenSearchedSelectedIndex > 0 && lengthOfFiles > 0 {
		filesThatHaveBeenSearchedSelectedIndex--
		return &filesThatHaveBeenSearched[filesThatHaveBeenSearchedSelectedIndex]
	} else if filesThatHaveBeenSearchedSelectedIndex == 0 && lengthOfFiles > 0 {
		return &filesThatHaveBeenSearched[0]
	} else {
		return nil
	}
}

func GetPreviousFile() *scripts.File {
	lengthOfFiles := len(filesThatHaveBeenSearched)
	if filesThatHaveBeenSearchedSelectedIndex == -1 && lengthOfFiles > 0 {
		filesThatHaveBeenSearchedSelectedIndex = 0
		return &filesThatHaveBeenSearched[0]

	} else if filesThatHaveBeenSearchedSelectedIndex < lengthOfFiles-1 && lengthOfFiles > 0 {
		filesThatHaveBeenSearchedSelectedIndex++
		return &filesThatHaveBeenSearched[filesThatHaveBeenSearchedSelectedIndex]
	} else if filesThatHaveBeenSearchedSelectedIndex == lengthOfFiles-1 && lengthOfFiles > 0 {
		return &filesThatHaveBeenSearched[lengthOfFiles-1]
	} else {
		return nil
	}
}
