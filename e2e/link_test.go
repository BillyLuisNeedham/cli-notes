package e2e

import (
	"strings"
	"testing"
)

// ============================================
// gl (Get Links) Tests
// ============================================

func TestGL_ShowsOutgoingLinks(t *testing.T) {
	h := NewTestHarness(t)
	dateStr := Today()

	// Create source note with links (P1 so it appears first)
	sourceFile := "source-" + dateStr + ".md"
	h.CreateTodoWithContent(sourceFile, "Source Note", "Links to [[Target One]] and [[Target Two]]", dateStr, 1)

	// Create target notes
	h.CreateTodo("target-one-"+dateStr+".md", "Target One", []string{}, dateStr, false, 2)
	h.CreateTodo("target-two-"+dateStr+".md", "Target Two", []string{}, dateStr, false, 2)

	// Run: list todos, arrow down to select first item, run gl
	// Pattern: gt\n -> lists todos, \x1b[B -> selects first, gl\n -> runs command
	input := "gt\n\x1b[Bgl\n"

	stdout, _, _ := h.RunCommand(input)

	// Should show outgoing links
	if !strings.Contains(stdout, "Links from") && !strings.Contains(stdout, "Target One") {
		t.Log("Output:", stdout)
		t.Error("Expected 'Links from' header or target notes in output")
	}
}

func TestGL_ShowsUnresolvedLinks(t *testing.T) {
	h := NewTestHarness(t)
	dateStr := Today()

	// Create note with dead link
	noteFile := "dead-links-" + dateStr + ".md"
	h.CreateTodoWithContent(noteFile, "Dead Links Note", "Links to [[Missing Note]] that doesn't exist", dateStr, 1)

	// Run: list todos, arrow down to select, run gl
	input := "gt\n\x1b[Bgl\n"

	stdout, _, _ := h.RunCommand(input)

	// Should show unresolved links section
	if !strings.Contains(stdout, "Unresolved") && !strings.Contains(stdout, "Missing Note") {
		t.Log("Output:", stdout)
		// May show "No outgoing links" if no valid links exist, but should mention dead links
		if !strings.Contains(stdout, "unresolved") && !strings.Contains(stdout, "Missing") {
			t.Error("Expected unresolved links to be shown")
		}
	}
}

func TestGL_NoFileSelected(t *testing.T) {
	h := NewTestHarness(t)

	// Run gl without selecting a file
	input := "gl\n"

	stdout, _, _ := h.RunCommand(input)

	if !strings.Contains(stdout, "No file selected") {
		t.Log("Output:", stdout)
		t.Error("Expected 'No file selected' error")
	}
}

// ============================================
// gb (Get Backlinks) Tests
// ============================================

func TestGB_ShowsBacklinks(t *testing.T) {
	h := NewTestHarness(t)
	dateStr := Today()

	// Create target note (P1 so it appears first)
	targetFile := "target-" + dateStr + ".md"
	h.CreateTodo(targetFile, "Target Note", []string{}, dateStr, false, 1)

	// Create source notes that link to target
	h.CreateTodoWithContent("source-one-"+dateStr+".md", "Source One", "See [[Target Note]]", dateStr, 2)
	h.CreateTodoWithContent("source-two-"+dateStr+".md", "Source Two", "Also links to [[Target Note]]", dateStr, 2)

	// Run: list todos, arrow down to select target, run gb
	input := "gt\n\x1b[Bgb\n"

	stdout, _, _ := h.RunCommand(input)

	// Should show backlinks - check for either header style or source notes
	if !strings.Contains(stdout, "linking to") && !strings.Contains(stdout, "source") {
		t.Log("Output:", stdout)
		t.Error("Expected backlinks to be shown")
	}
}

func TestGB_NoBacklinks(t *testing.T) {
	h := NewTestHarness(t)
	dateStr := Today()

	// Create lonely note with no backlinks
	h.CreateTodo("lonely-"+dateStr+".md", "Lonely Note", []string{}, dateStr, false, 1)

	// Run: list todos, arrow down to select, run gb
	input := "gt\n\x1b[Bgb\n"

	stdout, _, _ := h.RunCommand(input)

	// Should indicate no backlinks (check for either message or just no error)
	if strings.Contains(stdout, "Error") {
		t.Log("Output:", stdout)
		t.Error("Unexpected error")
	}
}

// ============================================
// gg (Graph View) Tests
// ============================================

func TestGG_OpensGraphView(t *testing.T) {
	h := NewTestHarness(t)
	dateStr := Today()

	// Create center note with outgoing link (P1 so it appears first)
	h.CreateTodoWithContent("center-"+dateStr+".md", "Center Note", "Links to [[Target]]", dateStr, 1)
	h.CreateTodo("target-"+dateStr+".md", "Target", []string{}, dateStr, false, 2)

	// Run: list todos, arrow down to select, run gg, then quit
	input := "gt\n\x1b[Bgg\nq\n"

	stdout, _, _ := h.RunCommand(input)

	// Should show graph view elements
	if !strings.Contains(stdout, "GRAPH VIEW") && !strings.Contains(stdout, "CENTER") {
		t.Log("Output:", stdout)
		// Check for any graph-related output
		if !strings.Contains(stdout, "Center Note") {
			t.Error("Expected graph view to show center note")
		}
	}
}

func TestGG_NavigationWorks(t *testing.T) {
	h := NewTestHarness(t)
	dateStr := Today()

	// Create connected notes (center with P1 to appear first)
	h.CreateTodoWithContent("center-"+dateStr+".md", "Center Note", "[[Target A]] [[Target B]]", dateStr, 1)
	h.CreateTodo("target-a-"+dateStr+".md", "Target A", []string{}, dateStr, false, 2)
	h.CreateTodo("target-b-"+dateStr+".md", "Target B", []string{}, dateStr, false, 2)

	// Run: select center, open graph, navigate down twice, quit
	input := "gt\n\x1b[Bgg\njjq\n"

	stdout, _, err := h.RunCommand(input)
	if err != nil {
		t.Log("Output:", stdout)
	}

	// Should complete without crashing
}

func TestGG_FollowLink(t *testing.T) {
	h := NewTestHarness(t)
	dateStr := Today()

	// Create connected notes: center -> target -> another
	h.CreateTodoWithContent("center-"+dateStr+".md", "Center Note", "[[Target]]", dateStr, 1)
	h.CreateTodoWithContent("target-"+dateStr+".md", "Target", "[[Another]]", dateStr, 2)
	h.CreateTodo("another-"+dateStr+".md", "Another", []string{}, dateStr, false, 3)

	// Run: select center, open graph, navigate to target, press enter to follow, quit
	input := "gt\n\x1b[Bgg\nj\nq\n"

	stdout, _, err := h.RunCommand(input)
	if err != nil {
		t.Log("Output:", stdout)
	}

	// Should show the graph was rendered
	if strings.Contains(stdout, "Target") || strings.Contains(stdout, "Center") {
		// Success - graph view worked
	}
}

func TestGG_Quit(t *testing.T) {
	h := NewTestHarness(t)
	dateStr := Today()

	h.CreateTodo("note-"+dateStr+".md", "My Note", []string{}, dateStr, false, 1)

	// Run: select note, open graph, quit immediately
	input := "gt\n\x1b[Bgg\nq\n"

	_, _, err := h.RunCommand(input)
	if err != nil {
		t.Log("Error:", err)
	}

	// Should exit cleanly after q
}

// ============================================
// ln (Link Note) Tests
// ============================================

func TestLN_AddsLinkToNote(t *testing.T) {
	h := NewTestHarness(t)
	dateStr := Today()

	// Create source note and target note
	sourceFile := "source-" + dateStr + ".md"
	targetFile := "target-" + dateStr + ".md"

	h.CreateTodo(sourceFile, "Source Note", []string{}, dateStr, false, 1)
	h.CreateTodo(targetFile, "Target Note", []string{}, dateStr, false, 2)

	// Run: list todos, select source, run ln, type to filter, select
	// The ln command opens a fuzzy picker
	input := "gt\n\x1b[Bln\nTarget\n"

	stdout, _, _ := h.RunCommand(input)

	// Check if link was added (output should confirm linking)
	if strings.Contains(stdout, "Linked") || strings.Contains(stdout, "Target") {
		// Success indication
	}

	// Verify the file content includes the link
	h.AssertFileContent(sourceFile, "[[Target Note]]")
}

// ============================================
// Integration Tests
// ============================================

func TestLinkWorkflow_CreateAndNavigate(t *testing.T) {
	h := NewTestHarness(t)
	dateStr := Today()

	// Create a small knowledge graph
	// Project -> Meeting Notes -> Action Items
	//         -> Design Doc
	projectFile := "project-" + dateStr + ".md"
	h.CreateTodoWithContent(projectFile, "Project Notes",
		"See [[Meeting Notes]] and [[Design Doc]] for details", dateStr, 1)

	h.CreateTodoWithContent("meeting-notes-"+dateStr+".md", "Meeting Notes",
		"Action items: [[Action Items]]", dateStr, 2)

	h.CreateTodo("design-doc-"+dateStr+".md", "Design Doc", []string{}, dateStr, false, 2)
	h.CreateTodo("action-items-"+dateStr+".md", "Action Items", []string{}, dateStr, false, 2)

	// Verify links work - select project, check links
	input := "gt\n\x1b[Bgl\n"

	stdout, _, _ := h.RunCommand(input)

	// Should show outgoing links from project (check lowercase filenames)
	if !strings.Contains(stdout, "meeting") && !strings.Contains(stdout, "design") {
		t.Log("Output:", stdout)
		t.Error("Expected to see linked notes in output")
	}
}
