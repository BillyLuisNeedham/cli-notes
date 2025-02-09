package data

import "cli-notes/scripts"

type SearchedFilesStore struct {
    filesThatHaveBeenSearched []scripts.File
    selectedIndex int
}

func NewSearchedFilesStore() *SearchedFilesStore {
    return &SearchedFilesStore{
        filesThatHaveBeenSearched: make([]scripts.File, 0),
        selectedIndex: -1,
    }
}

func (s *SearchedFilesStore) SetFilesSearched(files []scripts.File) {
    s.filesThatHaveBeenSearched = files
}

func (s *SearchedFilesStore) GetFilesSearched() []scripts.File {
    return s.filesThatHaveBeenSearched
}

func (s *SearchedFilesStore) GetNextFile() *scripts.File {
    lengthOfFiles := len(s.filesThatHaveBeenSearched)
    if s.selectedIndex == -1 && lengthOfFiles > 0 {
        s.selectedIndex = lengthOfFiles - 1
        return &s.filesThatHaveBeenSearched[lengthOfFiles-1]

    } else if s.selectedIndex > 0 && lengthOfFiles > 0 {
        s.selectedIndex--
        return &s.filesThatHaveBeenSearched[s.selectedIndex]
    } else if s.selectedIndex == 0 && lengthOfFiles > 0 {
        return &s.filesThatHaveBeenSearched[0]
    } else {
        return nil
    }
}

func (s *SearchedFilesStore) GetPreviousFile() *scripts.File {
    lengthOfFiles := len(s.filesThatHaveBeenSearched)
    if s.selectedIndex == -1 && lengthOfFiles > 0 {
        s.selectedIndex = 0
        return &s.filesThatHaveBeenSearched[0]

    } else if s.selectedIndex < lengthOfFiles-1 && lengthOfFiles > 0 {
        s.selectedIndex++
        return &s.filesThatHaveBeenSearched[s.selectedIndex]
    } else if s.selectedIndex == lengthOfFiles-1 && lengthOfFiles > 0 {
        return &s.filesThatHaveBeenSearched[lengthOfFiles-1]
    } else {
        return nil
    }
} 