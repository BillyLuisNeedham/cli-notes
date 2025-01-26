package presentation

import (
	"cli-notes/scripts"
	"fmt"
)

func PrintAllFileNames(files []scripts.File) {
	for _, file := range files {
		fmt.Println(file.Name)
	}
}