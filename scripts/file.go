package scripts

import "time"

type File struct {
	Name string
	Title string
	Tags []string
	CreatedAt time.Time
	DueAt time.Time
	Done bool
	Content string
}