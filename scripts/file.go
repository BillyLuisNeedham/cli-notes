package scripts

import "time"

// Define priorities for notes
type Priority int

const (
	P1 Priority = 1 // Highest priority
	P2 Priority = 2 // Medium priority (default)
	P3 Priority = 3 // Lowest priority
)

type File struct {
	Name          string
	Title         string
	Tags          []string
	CreatedAt     time.Time
	DueAt         time.Time
	Done          bool
	Content       string
	Priority      Priority
	ObjectiveRole string // "parent" or "" (empty for non-objectives)
	ObjectiveID   string // 8-char hash linking parent and children
}
