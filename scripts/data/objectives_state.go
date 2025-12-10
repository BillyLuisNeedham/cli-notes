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

type ListFilterMode int

const (
	ListShowActive ListFilterMode = iota
	ListShowCompleted
)

type ObjectivesViewState struct {
	ViewMode ObjectivesViewMode

	// List view state
	AllObjectives  []scripts.File     // All objectives (unfiltered)
	Objectives     []scripts.File     // Filtered objectives for display
	SelectedIndex  int                // Index in Objectives list
	ListFilterMode ListFilterMode     // Active vs Completed tab

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

	state := &ObjectivesViewState{
		ViewMode:       ObjectivesListView,
		AllObjectives:  objectives,
		SelectedIndex:  0,
		ListFilterMode: ListShowActive,
		SortOrder:      SortByDueDateThenPriority,
		FilterMode:     ShowAll,
	}

	// Apply filter to populate Objectives from AllObjectives
	state.applyListFilter()

	return state, nil
}

// NewSingleObjectiveViewStateForObjective initializes the state directly in SingleObjectiveView mode
// for a specific objective. This is used when navigating to an objective from the root view.
func NewSingleObjectiveViewStateForObjective(objective scripts.File) (*ObjectivesViewState, error) {
	// Query all objectives for BackToList functionality
	objectives, err := QueryAllObjectives()
	if err != nil {
		return nil, err
	}

	state := &ObjectivesViewState{
		ViewMode:           SingleObjectiveView,
		AllObjectives:      objectives,
		SelectedIndex:      0,
		ListFilterMode:     ListShowActive,
		CurrentObjective:   &objective,
		ChildSelectedIndex: 0,
		OnParent:           true,
		SortOrder:          SortByDueDateThenPriority,
		FilterMode:         ShowAll,
	}

	// Apply list filter to populate Objectives
	state.applyListFilter()

	// Find the index of this objective in the filtered list (for BackToList selection)
	for i, obj := range state.Objectives {
		if obj.ObjectiveID == objective.ObjectiveID {
			state.SelectedIndex = i
			break
		}
	}

	// Load children for the objective
	children, err := QueryChildrenByObjectiveID(objective.ObjectiveID, true)
	if err != nil {
		return nil, err
	}
	state.Children = children

	state.applySortAndFilter()

	return state, nil
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

	ovs.AllObjectives = objectives
	ovs.ViewMode = ObjectivesListView
	ovs.CurrentObjective = nil
	ovs.Children = nil
	ovs.SortOrder = SortByDueDateThenPriority
	ovs.FilterMode = ShowAll

	// Apply list filter to populate Objectives
	ovs.applyListFilter()

	return nil
}

// Refresh reloads current view
func (ovs *ObjectivesViewState) Refresh() error {
	if ovs.ViewMode == ObjectivesListView {
		objectives, err := QueryAllObjectives()
		if err != nil {
			return err
		}
		ovs.AllObjectives = objectives
		ovs.applyListFilter()
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

// applyListFilter filters AllObjectives based on ListFilterMode
func (ovs *ObjectivesViewState) applyListFilter() {
	active := make([]scripts.File, 0)
	completed := make([]scripts.File, 0)

	for _, obj := range ovs.AllObjectives {
		if obj.Done {
			completed = append(completed, obj)
		} else {
			active = append(active, obj)
		}
	}

	// Sort each group by creation date (most recent first)
	sortByCreatedAt(active, true)
	sortByCreatedAt(completed, true)

	if ovs.ListFilterMode == ListShowActive {
		ovs.Objectives = active
	} else {
		ovs.Objectives = completed
	}

	// Reset selection if out of bounds
	if ovs.SelectedIndex >= len(ovs.Objectives) {
		if len(ovs.Objectives) > 0 {
			ovs.SelectedIndex = 0
		} else {
			ovs.SelectedIndex = 0
		}
	}
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

// CycleListFilterMode toggles between Active and Completed tabs in list view
func (ovs *ObjectivesViewState) CycleListFilterMode() {
	if ovs.ListFilterMode == ListShowActive {
		ovs.ListFilterMode = ListShowCompleted
	} else {
		ovs.ListFilterMode = ListShowActive
	}
	ovs.applyListFilter()
}

// GetListCounts returns (activeCount, completedCount) for tab headers
func (ovs *ObjectivesViewState) GetListCounts() (int, int) {
	active := 0
	completed := 0
	for _, obj := range ovs.AllObjectives {
		if obj.Done {
			completed++
		} else {
			active++
		}
	}
	return active, completed
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
