package presentation

import (
	"cli-notes/scripts"
	"fmt"
)

func PrintAllFiles(files []scripts.File) {
	for _, file := range files {
		fmt.Printf("%v due: %v\n", file.Name, file.DueAt.Format("2006-01-02"))
	}
}
