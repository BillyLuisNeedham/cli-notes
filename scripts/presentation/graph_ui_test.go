package presentation

import (
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"strings"
	"testing"
	"time"
)

// ============================================
// RenderGraphView Tests
// ============================================

func TestRenderGraphView_ShowsTitle(t *testing.T) {
	state := &data.GraphViewState{
		CenterNode: scripts.File{
			Name:  "test.md",
			Title: "Test Note",
		},
		OutLinks:  []scripts.File{},
		BackLinks: []scripts.File{},
		Nodes: []data.GraphNode{{
			File:      scripts.File{Name: "test.md", Title: "Test Note"},
			IsCenter:  true,
			Direction: "center",
		}},
	}

	output := RenderGraphView(state, 80, 24)

	if !strings.Contains(output, "LINK GRAPH VIEW") {
		t.Error("Expected output to contain 'LINK GRAPH VIEW' title")
	}
}

func TestRenderGraphView_ShowsCenterNode(t *testing.T) {
	state := &data.GraphViewState{
		CenterNode: scripts.File{
			Name:  "center.md",
			Title: "My Center Note",
		},
		OutLinks:  []scripts.File{},
		BackLinks: []scripts.File{},
		Nodes: []data.GraphNode{{
			File:      scripts.File{Name: "center.md", Title: "My Center Note"},
			IsCenter:  true,
			Direction: "center",
		}},
	}

	output := RenderGraphView(state, 80, 24)

	if !strings.Contains(output, "My Center Note") {
		t.Error("Expected output to contain center note title")
	}

	if !strings.Contains(output, "[CENTER]") {
		t.Error("Expected output to contain [CENTER] label")
	}
}

func TestRenderGraphView_ShowsBacklinks(t *testing.T) {
	state := &data.GraphViewState{
		CenterNode: scripts.File{
			Name:  "center.md",
			Title: "Center",
		},
		OutLinks: []scripts.File{},
		BackLinks: []scripts.File{{
			Name:  "source.md",
			Title: "Source Note",
		}},
		Nodes: []data.GraphNode{
			{File: scripts.File{Name: "source.md", Title: "Source Note"}, Direction: "in"},
			{File: scripts.File{Name: "center.md", Title: "Center"}, IsCenter: true, Direction: "center"},
		},
	}

	output := RenderGraphView(state, 80, 24)

	if !strings.Contains(output, "Backlinks") {
		t.Error("Expected output to contain 'Backlinks' section")
	}

	if !strings.Contains(output, "Source Note") {
		t.Error("Expected output to contain backlink note title")
	}
}

func TestRenderGraphView_ShowsOutlinks(t *testing.T) {
	state := &data.GraphViewState{
		CenterNode: scripts.File{
			Name:  "center.md",
			Title: "Center",
		},
		OutLinks: []scripts.File{{
			Name:  "target.md",
			Title: "Target Note",
		}},
		BackLinks: []scripts.File{},
		Nodes: []data.GraphNode{
			{File: scripts.File{Name: "center.md", Title: "Center"}, IsCenter: true, Direction: "center"},
			{File: scripts.File{Name: "target.md", Title: "Target Note"}, Direction: "out"},
		},
	}

	output := RenderGraphView(state, 80, 24)

	if !strings.Contains(output, "Outgoing links") {
		t.Error("Expected output to contain 'Outgoing links' section")
	}

	if !strings.Contains(output, "Target Note") {
		t.Error("Expected output to contain outlink note title")
	}
}

func TestRenderGraphView_ShowsNoLinksMessage(t *testing.T) {
	state := &data.GraphViewState{
		CenterNode: scripts.File{
			Name:  "lonely.md",
			Title: "Lonely Note",
		},
		OutLinks:  []scripts.File{},
		BackLinks: []scripts.File{},
		Nodes: []data.GraphNode{{
			File:      scripts.File{Name: "lonely.md", Title: "Lonely Note"},
			IsCenter:  true,
			Direction: "center",
		}},
	}

	output := RenderGraphView(state, 80, 24)

	if !strings.Contains(output, "No links found") {
		t.Error("Expected output to show 'No links found' message")
	}
}

func TestRenderGraphView_ShowsHelpText(t *testing.T) {
	state := &data.GraphViewState{
		CenterNode: scripts.File{
			Name:  "test.md",
			Title: "Test",
		},
		Nodes: []data.GraphNode{{
			File:      scripts.File{Name: "test.md", Title: "Test"},
			IsCenter:  true,
			Direction: "center",
		}},
	}

	output := RenderGraphView(state, 80, 24)

	// Check for navigation help
	if !strings.Contains(output, "j/k") {
		t.Error("Expected output to contain j/k navigation hint")
	}

	if !strings.Contains(output, "q = quit") {
		t.Error("Expected output to contain quit instruction")
	}
}

// ============================================
// Selection Highlighting Tests
// ============================================

func TestRenderGraphView_HighlightsSelectedNode(t *testing.T) {
	state := &data.GraphViewState{
		CenterNode: scripts.File{
			Name:  "center.md",
			Title: "Center",
		},
		OutLinks: []scripts.File{{
			Name:  "target.md",
			Title: "Target",
		}},
		BackLinks:   []scripts.File{},
		SelectedIdx: 0, // Center is selected
		Nodes: []data.GraphNode{
			{File: scripts.File{Name: "center.md", Title: "Center"}, IsCenter: true, Direction: "center"},
			{File: scripts.File{Name: "target.md", Title: "Target"}, Direction: "out"},
		},
	}

	output := RenderGraphView(state, 80, 24)

	// Selected node should have ">" indicator
	if !strings.Contains(output, "> ") {
		t.Error("Expected output to contain '>' selection indicator")
	}
}

// ============================================
// truncateTitle Tests
// ============================================

func TestTruncateTitle_ShortTitle(t *testing.T) {
	title := "Short"
	result := truncateTitle(title, 20)

	if result != "Short" {
		t.Errorf("Expected 'Short', got '%s'", result)
	}
}

func TestTruncateTitle_ExactLength(t *testing.T) {
	title := "Exact Length"
	result := truncateTitle(title, 12) // same length as title

	if result != title {
		t.Errorf("Expected '%s', got '%s'", title, result)
	}
}

func TestTruncateTitle_LongTitle(t *testing.T) {
	title := "This is a very long title that needs truncation"
	result := truncateTitle(title, 20)

	if len(result) != 20 {
		t.Errorf("Expected length 20, got %d", len(result))
	}

	if !strings.HasSuffix(result, "...") {
		t.Error("Expected truncated title to end with '...'")
	}
}

func TestTruncateTitle_EmptyTitle(t *testing.T) {
	result := truncateTitle("", 20)

	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

// ============================================
// maxInt Tests
// ============================================

func TestMaxInt_FirstLarger(t *testing.T) {
	result := maxInt(10, 5)
	if result != 10 {
		t.Errorf("Expected 10, got %d", result)
	}
}

func TestMaxInt_SecondLarger(t *testing.T) {
	result := maxInt(5, 10)
	if result != 10 {
		t.Errorf("Expected 10, got %d", result)
	}
}

func TestMaxInt_Equal(t *testing.T) {
	result := maxInt(7, 7)
	if result != 7 {
		t.Errorf("Expected 7, got %d", result)
	}
}

func TestMaxInt_WithNegative(t *testing.T) {
	result := maxInt(-5, 0)
	if result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}
}

func TestMaxInt_BothNegative(t *testing.T) {
	result := maxInt(-10, -5)
	if result != -5 {
		t.Errorf("Expected -5, got %d", result)
	}
}

// ============================================
// Edge Case Tests
// ============================================

func TestRenderCenterBox_LongTitle(t *testing.T) {
	// Test that long titles don't cause panics
	var sb strings.Builder

	longTitle := "This is a very long title that exceeds the box width by quite a bit"
	truncated := truncateTitle(longTitle, boxWidth-4)

	// This should not panic
	renderCenterBox(&sb, truncated, false)

	output := sb.String()
	if output == "" {
		t.Error("Expected non-empty output")
	}
}

func TestRenderNodeBox_PaddingNeverNegative(t *testing.T) {
	// Test with title exactly at box width (edge case that previously caused panic)
	var sb strings.Builder

	// Title that when padded would have caused negative values
	title := strings.Repeat("A", boxWidth-2) // max possible without truncation

	// This should not panic
	renderNodeBox(&sb, title, false, "")

	output := sb.String()
	if output == "" {
		t.Error("Expected non-empty output")
	}

	// Verify no negative padding (no truncated content)
	if strings.Contains(output, "strings: negative Repeat count") {
		t.Error("Detected panic message in output")
	}
}

func TestRenderNodeBox_VeryLongTitle(t *testing.T) {
	var sb strings.Builder

	// Very long title that needs truncation
	title := strings.Repeat("X", 100)
	truncated := truncateTitle(title, boxWidth-4)

	// This should not panic
	renderNodeBox(&sb, truncated, false, "│  ")

	output := sb.String()
	if !strings.Contains(output, "...") {
		// After truncation, should have ellipsis
		if len(truncated) > boxWidth-4 {
			t.Error("Title should be truncated")
		}
	}
}

func TestRenderGraphView_WithRealState(t *testing.T) {
	// Integration test with a more realistic state
	state := &data.GraphViewState{
		CenterNode: scripts.File{
			Name:      "project-notes.md",
			Title:     "Project Notes",
			CreatedAt: time.Now(),
		},
		OutLinks: []scripts.File{
			{Name: "meeting.md", Title: "Weekly Meeting"},
			{Name: "design.md", Title: "Design Decisions"},
		},
		BackLinks: []scripts.File{
			{Name: "index.md", Title: "Home Page"},
		},
		SelectedIdx: 1, // Center selected (backlinks=1, center is at index 1)
		Nodes: []data.GraphNode{
			{File: scripts.File{Name: "index.md", Title: "Home Page"}, Direction: "in"},
			{File: scripts.File{Name: "project-notes.md", Title: "Project Notes"}, IsCenter: true, Direction: "center"},
			{File: scripts.File{Name: "meeting.md", Title: "Weekly Meeting"}, Direction: "out"},
			{File: scripts.File{Name: "design.md", Title: "Design Decisions"}, Direction: "out"},
		},
	}

	output := RenderGraphView(state, 120, 40)

	// Should contain all titles
	if !strings.Contains(output, "Project Notes") {
		t.Error("Missing center node title")
	}
	if !strings.Contains(output, "Weekly Meeting") {
		t.Error("Missing outlink title")
	}
	if !strings.Contains(output, "Home Page") {
		t.Error("Missing backlink title")
	}

	// Should have both sections
	if !strings.Contains(output, "Backlinks") {
		t.Error("Missing backlinks section")
	}
	if !strings.Contains(output, "Outgoing") {
		t.Error("Missing outgoing section")
	}
}
