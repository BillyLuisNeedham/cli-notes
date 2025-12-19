package data

import (
	"cli-notes/scripts"
	"testing"
	"time"
)

// ============================================
// NewGraphViewState Tests
// ============================================

func TestNewGraphViewState_LoadsLinksAndBacklinks(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create center note with outgoing link
	createTestFile(t, scripts.File{
		Name:      "center.md",
		Title:     "Center Note",
		CreatedAt: time.Now(),
		Content:   "# Center Note\n\n[[Target Note]]",
	})

	// Create target note (for outlink)
	createTestFile(t, scripts.File{
		Name:      "target.md",
		Title:     "Target Note",
		CreatedAt: time.Now(),
		Content:   "# Target Note",
	})

	// Create source note that links to center (for backlink)
	createTestFile(t, scripts.File{
		Name:      "source.md",
		Title:     "Source Note",
		CreatedAt: time.Now(),
		Content:   "[[Center Note]]",
	})

	centerFile, err := LoadFileByName("center.md")
	if err != nil {
		t.Fatalf("Failed to load center file: %v", err)
	}

	state, err := NewGraphViewState(centerFile)
	if err != nil {
		t.Fatalf("NewGraphViewState failed: %v", err)
	}

	// Should have 1 outlink
	if len(state.OutLinks) != 1 {
		t.Errorf("Expected 1 outlink, got %d", len(state.OutLinks))
	}

	// Should have 1 backlink
	if len(state.BackLinks) != 1 {
		t.Errorf("Expected 1 backlink, got %d", len(state.BackLinks))
	}

	// Center node should be set
	if state.CenterNode.Name != "center.md" {
		t.Errorf("Expected center node 'center.md', got '%s'", state.CenterNode.Name)
	}
}

func TestGraphViewState_BuildsNodesList(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create interconnected notes
	createTestFile(t, scripts.File{
		Name:      "center.md",
		Title:     "Center",
		CreatedAt: time.Now(),
		Content:   "[[Out1]] [[Out2]]",
	})

	createTestFile(t, scripts.File{
		Name:      "out1.md",
		Title:     "Out1",
		CreatedAt: time.Now(),
		Content:   "Target 1",
	})

	createTestFile(t, scripts.File{
		Name:      "out2.md",
		Title:     "Out2",
		CreatedAt: time.Now(),
		Content:   "Target 2",
	})

	createTestFile(t, scripts.File{
		Name:      "back1.md",
		Title:     "Back1",
		CreatedAt: time.Now(),
		Content:   "[[Center]]",
	})

	centerFile, _ := LoadFileByName("center.md")
	state, err := NewGraphViewState(centerFile)
	if err != nil {
		t.Fatalf("NewGraphViewState failed: %v", err)
	}

	// Nodes should be: backlinks + center + outlinks = 1 + 1 + 2 = 4
	expectedNodes := len(state.BackLinks) + 1 + len(state.OutLinks)
	if len(state.Nodes) != expectedNodes {
		t.Errorf("Expected %d nodes, got %d", expectedNodes, len(state.Nodes))
	}

	// First node should be backlink (direction "in")
	if len(state.Nodes) > 0 && state.Nodes[0].Direction != "in" {
		t.Errorf("Expected first node direction 'in', got '%s'", state.Nodes[0].Direction)
	}

	// Center node should be in the middle
	centerIdx := len(state.BackLinks)
	if centerIdx < len(state.Nodes) {
		if !state.Nodes[centerIdx].IsCenter {
			t.Error("Expected center node at backlinks position")
		}
		if state.Nodes[centerIdx].Direction != "center" {
			t.Errorf("Expected center direction, got '%s'", state.Nodes[centerIdx].Direction)
		}
	}
}

// ============================================
// Navigation Tests
// ============================================

func TestGraphViewState_SelectNext(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "center.md",
		Title:     "Center",
		CreatedAt: time.Now(),
		Content:   "[[A]] [[B]]",
	})

	createTestFile(t, scripts.File{
		Name:      "a.md",
		Title:     "A",
		CreatedAt: time.Now(),
		Content:   "",
	})

	createTestFile(t, scripts.File{
		Name:      "b.md",
		Title:     "B",
		CreatedAt: time.Now(),
		Content:   "",
	})

	centerFile, _ := LoadFileByName("center.md")
	state, _ := NewGraphViewState(centerFile)

	initialIdx := state.SelectedIdx

	state.SelectNext()
	if state.SelectedIdx != initialIdx+1 {
		t.Errorf("Expected index %d after SelectNext, got %d", initialIdx+1, state.SelectedIdx)
	}

	// Keep selecting until we wrap around (need len-1 more to get back to start)
	for i := 0; i < len(state.Nodes)-1; i++ {
		state.SelectNext()
	}

	// Should wrap back to start
	if state.SelectedIdx != initialIdx {
		t.Errorf("Expected to wrap around to %d, got %d", initialIdx, state.SelectedIdx)
	}
}

func TestGraphViewState_SelectPrevious(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "center.md",
		Title:     "Center",
		CreatedAt: time.Now(),
		Content:   "[[A]]",
	})

	createTestFile(t, scripts.File{
		Name:      "a.md",
		Title:     "A",
		CreatedAt: time.Now(),
		Content:   "",
	})

	centerFile, _ := LoadFileByName("center.md")
	state, _ := NewGraphViewState(centerFile)

	// From 0, going previous should wrap to end
	state.SelectPrevious()
	if state.SelectedIdx != len(state.Nodes)-1 {
		t.Errorf("Expected to wrap to %d, got %d", len(state.Nodes)-1, state.SelectedIdx)
	}
}

func TestGraphViewState_GetSelectedNode(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "center.md",
		Title:     "Center",
		CreatedAt: time.Now(),
		Content:   "[[Target]]",
	})

	createTestFile(t, scripts.File{
		Name:      "target.md",
		Title:     "Target",
		CreatedAt: time.Now(),
		Content:   "",
	})

	centerFile, _ := LoadFileByName("center.md")
	state, _ := NewGraphViewState(centerFile)

	node := state.GetSelectedNode()
	if node == nil {
		t.Fatal("Expected selected node, got nil")
	}

	// Initially should be first node (center if no backlinks)
	if state.SelectedIdx != 0 {
		t.Errorf("Initial selection should be 0, got %d", state.SelectedIdx)
	}
}

// ============================================
// NavigateToSelected Tests
// ============================================

func TestGraphViewState_NavigateToSelected_ChangesCenter(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "center.md",
		Title:     "Center",
		CreatedAt: time.Now(),
		Content:   "[[Target]]",
	})

	createTestFile(t, scripts.File{
		Name:      "target.md",
		Title:     "Target",
		CreatedAt: time.Now(),
		Content:   "[[Another]]",
	})

	createTestFile(t, scripts.File{
		Name:      "another.md",
		Title:     "Another",
		CreatedAt: time.Now(),
		Content:   "",
	})

	centerFile, _ := LoadFileByName("center.md")
	state, _ := NewGraphViewState(centerFile)

	// Move selection to target node (outlink)
	for state.GetSelectedNode() != nil && state.GetSelectedNode().File.Name != "target.md" {
		state.SelectNext()
	}

	if state.GetSelectedNode() == nil || state.GetSelectedNode().File.Name != "target.md" {
		t.Fatal("Failed to select target node")
	}

	err := state.NavigateToSelected()
	if err != nil {
		t.Fatalf("NavigateToSelected failed: %v", err)
	}

	// Center should now be target
	if state.CenterNode.Name != "target.md" {
		t.Errorf("Expected center to be 'target.md', got '%s'", state.CenterNode.Name)
	}

	// Should have loaded target's links
	// Target links to Another, so should have 1 outlink
	if len(state.OutLinks) != 1 {
		t.Errorf("Expected 1 outlink after navigation, got %d", len(state.OutLinks))
	}
}

func TestGraphViewState_NavigateToSelected_IgnoresCenter(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "center.md",
		Title:     "Center",
		CreatedAt: time.Now(),
		Content:   "[[Target]]",
	})

	createTestFile(t, scripts.File{
		Name:      "target.md",
		Title:     "Target",
		CreatedAt: time.Now(),
		Content:   "",
	})

	centerFile, _ := LoadFileByName("center.md")
	state, _ := NewGraphViewState(centerFile)

	// Select center node
	for state.GetSelectedNode() != nil && !state.GetSelectedNode().IsCenter {
		state.SelectNext()
	}

	originalCenter := state.CenterNode.Name

	err := state.NavigateToSelected()
	if err != nil {
		t.Fatalf("NavigateToSelected failed: %v", err)
	}

	// Should remain unchanged when center is selected
	if state.CenterNode.Name != originalCenter {
		t.Error("Center should not change when navigating to center node")
	}
}

// ============================================
// Refresh Tests
// ============================================

func TestGraphViewState_Refresh_ReloadsLinks(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "center.md",
		Title:     "Center",
		CreatedAt: time.Now(),
		Content:   "# Center",
	})

	centerFile, _ := LoadFileByName("center.md")
	state, _ := NewGraphViewState(centerFile)

	// Initially no links
	if len(state.OutLinks) != 0 {
		t.Errorf("Expected 0 outlinks initially, got %d", len(state.OutLinks))
	}

	// Add a target and update center to link to it
	createTestFile(t, scripts.File{
		Name:      "newlink.md",
		Title:     "NewLink",
		CreatedAt: time.Now(),
		Content:   "",
	})

	// Update center file to have a link
	centerFile.Content = "[[NewLink]]"
	WriteFile(centerFile)

	// Refresh state
	err := state.Refresh()
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	// Should now have 1 outlink
	if len(state.OutLinks) != 1 {
		t.Errorf("Expected 1 outlink after refresh, got %d", len(state.OutLinks))
	}
}

// ============================================
// Edge Cases
// ============================================

func TestGraphViewState_EmptyGraph(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "lonely.md",
		Title:     "Lonely Note",
		CreatedAt: time.Now(),
		Content:   "No links here",
	})

	file, _ := LoadFileByName("lonely.md")
	state, err := NewGraphViewState(file)
	if err != nil {
		t.Fatalf("NewGraphViewState failed: %v", err)
	}

	if len(state.OutLinks) != 0 {
		t.Errorf("Expected 0 outlinks, got %d", len(state.OutLinks))
	}

	if len(state.BackLinks) != 0 {
		t.Errorf("Expected 0 backlinks, got %d", len(state.BackLinks))
	}

	// Should still have 1 node (the center)
	if len(state.Nodes) != 1 {
		t.Errorf("Expected 1 node (center), got %d", len(state.Nodes))
	}

	if !state.Nodes[0].IsCenter {
		t.Error("Only node should be the center node")
	}
}
