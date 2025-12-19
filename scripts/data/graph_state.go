package data

import (
	"cli-notes/scripts"
)

// GraphNode represents a node in the link graph
type GraphNode struct {
	File      scripts.File
	IsCenter  bool
	Direction string // "in" for backlinks, "out" for outgoing links, "center" for the selected node
}

// GraphViewState holds the state for the graph view
type GraphViewState struct {
	CenterNode  scripts.File   // The node at the center of the graph
	OutLinks    []scripts.File // Notes this file links to
	BackLinks   []scripts.File // Notes that link to this file
	SelectedIdx int            // Currently selected node in the list
	Nodes       []GraphNode    // All nodes for navigation
}

// NewGraphViewState creates a new graph view state centered on the given file
func NewGraphViewState(centerFile scripts.File) (*GraphViewState, error) {
	state := &GraphViewState{
		CenterNode: centerFile,
	}

	err := state.Refresh()
	if err != nil {
		return nil, err
	}

	return state, nil
}

// Refresh reloads the links from disk
func (s *GraphViewState) Refresh() error {
	// Get outgoing links
	outLinks, err := GetLinksFrom(s.CenterNode.Name)
	if err != nil {
		return err
	}
	s.OutLinks = outLinks

	// Get backlinks
	backLinks, err := GetBacklinks(s.CenterNode.Name)
	if err != nil {
		return err
	}
	s.BackLinks = backLinks

	// Build nodes list for navigation
	s.buildNodesList()

	return nil
}

// buildNodesList builds the list of navigable nodes
func (s *GraphViewState) buildNodesList() {
	s.Nodes = make([]GraphNode, 0)

	// Add backlinks first (incoming)
	for _, file := range s.BackLinks {
		s.Nodes = append(s.Nodes, GraphNode{
			File:      file,
			Direction: "in",
		})
	}

	// Add center node
	s.Nodes = append(s.Nodes, GraphNode{
		File:      s.CenterNode,
		IsCenter:  true,
		Direction: "center",
	})

	// Add outlinks (outgoing)
	for _, file := range s.OutLinks {
		s.Nodes = append(s.Nodes, GraphNode{
			File:      file,
			Direction: "out",
		})
	}

	// Clamp selected index
	if s.SelectedIdx >= len(s.Nodes) {
		s.SelectedIdx = len(s.Nodes) - 1
	}
	if s.SelectedIdx < 0 {
		s.SelectedIdx = 0
	}
}

// SelectNext moves selection to next node
func (s *GraphViewState) SelectNext() {
	if len(s.Nodes) > 0 {
		s.SelectedIdx = (s.SelectedIdx + 1) % len(s.Nodes)
	}
}

// SelectPrevious moves selection to previous node
func (s *GraphViewState) SelectPrevious() {
	if len(s.Nodes) > 0 {
		s.SelectedIdx--
		if s.SelectedIdx < 0 {
			s.SelectedIdx = len(s.Nodes) - 1
		}
	}
}

// GetSelectedNode returns the currently selected node
func (s *GraphViewState) GetSelectedNode() *GraphNode {
	if s.SelectedIdx >= 0 && s.SelectedIdx < len(s.Nodes) {
		return &s.Nodes[s.SelectedIdx]
	}
	return nil
}

// NavigateToSelected moves the center to the selected node
func (s *GraphViewState) NavigateToSelected() error {
	node := s.GetSelectedNode()
	if node == nil || node.IsCenter {
		return nil
	}

	// Load the full file
	file, err := LoadFileByName(node.File.Name)
	if err != nil {
		return err
	}

	s.CenterNode = file
	s.SelectedIdx = 0 // Reset selection

	return s.Refresh()
}
