package data

import (
	"cli-notes/scripts"
	"fmt"
)

type ObjectivesViewMode int

const (
	ObjectivesListView ObjectivesViewMode = iota
	SingleObjectiveView
)

type SortOrder int

const (
	SortByDueDateThenPriority SortOrder = iota
	SortByPriorityThenDueDate
)

type FilterMode int

const (
	ShowAll FilterMode = iota
	ShowIncompleteOnly
	ShowCompleteOnly
)

type ObjectivesViewState struct {
	ViewMode ObjectivesViewMode

	// List view state
	Objectives    []scripts.File
	SelectedIndex int // Index in Objectives list

	// Single objective view state
	CurrentObjective   *scripts.File
	Children           []scripts.File
	ChildSelectedIndex int
	OnParent           bool // True if parent is selected, false if child
	SortOrder          SortOrder
	FilterMode         FilterMode
}

// NewObjectivesViewState initializes the objectives list view
func NewObjectivesViewState() (*ObjectivesViewState, error) {
	objectives, err := QueryAllObjectives()
	if err != nil {
		return nil, err
	}

	// Sort by most recently created first
	sortByCreatedAt(objectives, true)

	return &ObjectivesViewState{
		ViewMode:      ObjectivesListView,
		Objectives:    objectives,
		SelectedIndex: 0,
		SortOrder:     SortByDueDateThenPriority,
		FilterMode:    ShowAll,
	}, nil
}

// SelectNext moves selection down
func (ovs *ObjectivesViewState) SelectNext() {
	if ovs.ViewMode == ObjectivesListView {
		if len(ovs.Objectives) > 0 {
			ovs.SelectedIndex = (ovs.SelectedIndex + 1) % len(ovs.Objectives)
		}
	} else {
		// In single objective view
		if ovs.OnParent {
			// Move from parent to first child
			if len(ovs.Children) > 0 {
				ovs.OnParent = false
				ovs.ChildSelectedIndex = 0
			}
		} else {
			// Move to next child
			if len(ovs.Children) > 0 {
				ovs.ChildSelectedIndex = (ovs.ChildSelectedIndex + 1) % len(ovs.Children)
			}
		}
	}
}

// SelectPrevious moves selection up
func (ovs *ObjectivesViewState) SelectPrevious() {
	if ovs.ViewMode == ObjectivesListView {
		if len(ovs.Objectives) > 0 {
			ovs.SelectedIndex--
			if ovs.SelectedIndex < 0 {
				ovs.SelectedIndex = len(ovs.Objectives) - 1
			}
		}
	} else {
		// In single objective view
		if ovs.OnParent {
			// Already at parent, wrap to last child
			if len(ovs.Children) > 0 {
				ovs.OnParent = false
				ovs.ChildSelectedIndex = len(ovs.Children) - 1
			}
		} else if ovs.ChildSelectedIndex == 0 {
			// At first child, move to parent
			ovs.OnParent = true
		} else {
			// Move to previous child
			ovs.ChildSelectedIndex--
		}
	}
}

// OpenSelectedObjective transitions to single objective view
func (ovs *ObjectivesViewState) OpenSelectedObjective() error {
	if ovs.ViewMode != ObjectivesListView {
		return fmt.Errorf("not in list view")
	}
	if len(ovs.Objectives) == 0 {
		return fmt.Errorf("no objectives to open")
	}

	objective := ovs.Objectives[ovs.SelectedIndex]
	children, err := QueryChildrenByObjectiveID(objective.ObjectiveID, true)
	if err != nil {
		return err
	}

	ovs.CurrentObjective = &objective
	ovs.Children = children
	ovs.applySortAndFilter()
	ovs.ViewMode = SingleObjectiveView
	ovs.OnParent = true
	ovs.ChildSelectedIndex = 0

	return nil
}

// BackToList returns to objectives list view
func (ovs *ObjectivesViewState) BackToList() error {
	objectives, err := QueryAllObjectives()
	if err != nil {
		return err
	}

	// Sort by most recently created
	sortByCreatedAt(objectives, true)

	ovs.Objectives = objectives
	ovs.ViewMode = ObjectivesListView
	ovs.CurrentObjective = nil
	ovs.Children = nil
	ovs.SortOrder = SortByDueDateThenPriority
	ovs.FilterMode = ShowAll

	// Try to maintain selection if possible
	if ovs.SelectedIndex >= len(objectives) {
		ovs.SelectedIndex = 0
	}

	return nil
}

// Refresh reloads current view
func (ovs *ObjectivesViewState) Refresh() error {
	if ovs.ViewMode == ObjectivesListView {
		objectives, err := QueryAllObjectives()
		if err != nil {
			return err
		}
		sortByCreatedAt(objectives, true)
		ovs.Objectives = objectives

		// Adjust selection if out of bounds
		if ovs.SelectedIndex >= len(objectives) && len(objectives) > 0 {
			ovs.SelectedIndex = len(objectives) - 1
		}
	} else {
		if ovs.CurrentObjective == nil {
			return fmt.Errorf("no current objective")
		}
		children, err := QueryChildrenByObjectiveID(ovs.CurrentObjective.ObjectiveID, true)
		if err != nil {
			return err
		}
		ovs.Children = children
		ovs.applySortAndFilter()

		// Adjust selection if out of bounds
		if !ovs.OnParent && ovs.ChildSelectedIndex >= len(ovs.Children) && len(ovs.Children) > 0 {
			ovs.ChildSelectedIndex = len(ovs.Children) - 1
		}
	}
	return nil
}

// applySortAndFilter applies current sort and filter settings to children
func (ovs *ObjectivesViewState) applySortAndFilter() {
	// Separate into incomplete and complete first
	incomplete := make([]scripts.File, 0)
	complete := make([]scripts.File, 0)

	for _, child := range ovs.Children {
		if child.Done {
			complete = append(complete, child)
		} else {
			incomplete = append(incomplete, child)
		}
	}

	// Apply filter
	var filtered []scripts.File
	if ovs.FilterMode == ShowIncompleteOnly {
		filtered = incomplete
	} else if ovs.FilterMode == ShowCompleteOnly {
		filtered = complete
	} else {
		// ShowAll: sort each group separately, then concatenate
		// Sort incomplete
		if ovs.SortOrder == SortByDueDateThenPriority {
			scripts.SortTodosByDueDate(incomplete)
		} else {
			scripts.SortTodosByPriorityAndDueDate(incomplete)
		}

		// Sort complete
		if ovs.SortOrder == SortByDueDateThenPriority {
			scripts.SortTodosByDueDate(complete)
		} else {
			scripts.SortTodosByPriorityAndDueDate(complete)
		}

		// Concatenate: incomplete first, then complete
		// This ensures the array order matches the visual display order
		filtered = append(incomplete, complete...)
		ovs.Children = filtered
		return
	}

	// For filtered modes (ShowIncompleteOnly or ShowCompleteOnly), sort the filtered list
	if ovs.SortOrder == SortByDueDateThenPriority {
		scripts.SortTodosByDueDate(filtered)
	} else {
		scripts.SortTodosByPriorityAndDueDate(filtered)
	}

	ovs.Children = filtered
}

// ToggleSortOrder switches between sort orders
func (ovs *ObjectivesViewState) ToggleSortOrder() {
	if ovs.SortOrder == SortByDueDateThenPriority {
		ovs.SortOrder = SortByPriorityThenDueDate
	} else {
		ovs.SortOrder = SortByDueDateThenPriority
	}
	ovs.applySortAndFilter()
}

// CycleFilterMode cycles through filter modes
func (ovs *ObjectivesViewState) CycleFilterMode() {
	switch ovs.FilterMode {
	case ShowAll:
		ovs.FilterMode = ShowIncompleteOnly
	case ShowIncompleteOnly:
		ovs.FilterMode = ShowCompleteOnly
	case ShowCompleteOnly:
		ovs.FilterMode = ShowAll
	}
	ovs.applySortAndFilter()
}

// GetSelectedObjective returns currently selected objective in list view
func (ovs *ObjectivesViewState) GetSelectedObjective() *scripts.File {
	if ovs.ViewMode != ObjectivesListView || len(ovs.Objectives) == 0 {
		return nil
	}
	if ovs.SelectedIndex < 0 || ovs.SelectedIndex >= len(ovs.Objectives) {
		return nil
	}
	return &ovs.Objectives[ovs.SelectedIndex]
}

// GetSelectedChild returns currently selected child in single view
func (ovs *ObjectivesViewState) GetSelectedChild() *scripts.File {
	if ovs.ViewMode != SingleObjectiveView || ovs.OnParent {
		return nil
	}
	if len(ovs.Children) == 0 || ovs.ChildSelectedIndex >= len(ovs.Children) {
		return nil
	}
	return &ovs.Children[ovs.ChildSelectedIndex]
}

// GetCompletionCounts returns incomplete and complete counts for current objective
func (ovs *ObjectivesViewState) GetCompletionCounts() (int, int) {
	if ovs.ViewMode != SingleObjectiveView || ovs.CurrentObjective == nil {
		return 0, 0
	}

	// Need to query all children, not just filtered ones
	allChildren, err := QueryChildrenByObjectiveID(ovs.CurrentObjective.ObjectiveID, true)
	if err != nil {
		return 0, 0
	}

	incomplete := 0
	complete := 0
	for _, child := range allChildren {
		if child.Done {
			complete++
		} else {
			incomplete++
		}
	}

	return incomplete, complete
}

// GetIncompleteAndCompleteChildren returns separate slices for display
func (ovs *ObjectivesViewState) GetIncompleteAndCompleteChildren() ([]scripts.File, []scripts.File) {
	incomplete := make([]scripts.File, 0)
	complete := make([]scripts.File, 0)

	for _, child := range ovs.Children {
		if child.Done {
			complete = append(complete, child)
		} else {
			incomplete = append(incomplete, child)
		}
	}

	return incomplete, complete
}

// sortByCreatedAt sorts files by creation date
// If mostRecentFirst is true, newest files come first
func sortByCreatedAt(files []scripts.File, mostRecentFirst bool) {
	if mostRecentFirst {
		// Sort descending (newest first)
		for i := 0; i < len(files)-1; i++ {
			for j := i + 1; j < len(files); j++ {
				if files[i].CreatedAt.Before(files[j].CreatedAt) {
					files[i], files[j] = files[j], files[i]
				}
			}
		}
	} else {
		// Sort ascending (oldest first)
		for i := 0; i < len(files)-1; i++ {
			for j := i + 1; j < len(files); j++ {
				if files[i].CreatedAt.After(files[j].CreatedAt) {
					files[i], files[j] = files[j], files[i]
				}
			}
		}
	}
}
