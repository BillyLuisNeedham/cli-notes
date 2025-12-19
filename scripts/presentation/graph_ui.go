package presentation

import (
	"cli-notes/scripts/data"
	"fmt"
	"strings"
)

const (
	boxWidth = 30
)

// maxInt returns the larger of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RenderGraphView renders the ASCII graph view
func RenderGraphView(state *data.GraphViewState, termWidth, termHeight int) string {
	var sb strings.Builder

	// Clear screen and reset cursor
	sb.WriteString("\033[2J\033[H")

	// Title
	sb.WriteString("╔══════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                         LINK GRAPH VIEW                              ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════════════════════════╝\n\n")

	// Render the graph
	renderGraph(&sb, state)

	// Help
	sb.WriteString("\n───────────────────────────────────────────────────────────────────────\n")
	sb.WriteString("j/k = navigate nodes │ Enter = go to selected │ o = open in editor │ q = quit\n")

	return sb.String()
}

// renderGraph renders the ASCII graph layout
func renderGraph(sb *strings.Builder, state *data.GraphViewState) {
	centerTitle := truncateTitle(state.CenterNode.Title, boxWidth-4)

	// Backlinks section (incoming)
	if len(state.BackLinks) > 0 {
		sb.WriteString("  ┌── Backlinks (notes linking TO this) ──────────────────────────────┐\n")
		sb.WriteString("  │                                                                    │\n")

		for i, link := range state.BackLinks {
			nodeIdx := i // Backlinks are first in the nodes list
			selected := nodeIdx == state.SelectedIdx

			title := truncateTitle(link.Title, boxWidth-4)
			renderNodeBox(sb, title, selected, "│  ")
			sb.WriteString("  │\n")
		}

		sb.WriteString("  │                              │                                     │\n")
		sb.WriteString("  │                              ▼                                     │\n")
		sb.WriteString("  └────────────────────────────────────────────────────────────────────┘\n")
	}

	// Center node
	sb.WriteString("\n")
	centerIdx := len(state.BackLinks) // Center is after backlinks
	centerSelected := centerIdx == state.SelectedIdx
	renderCenterBox(sb, centerTitle, centerSelected)
	sb.WriteString("\n")

	// Outlinks section (outgoing)
	if len(state.OutLinks) > 0 {
		sb.WriteString("  ┌── Outgoing links (this note links TO) ────────────────────────────┐\n")
		sb.WriteString("  │                              │                                     │\n")
		sb.WriteString("  │                              ▼                                     │\n")
		sb.WriteString("  │                                                                    │\n")

		for i, link := range state.OutLinks {
			nodeIdx := len(state.BackLinks) + 1 + i // After backlinks and center
			selected := nodeIdx == state.SelectedIdx

			title := truncateTitle(link.Title, boxWidth-4)
			renderNodeBox(sb, title, selected, "│  ")
			sb.WriteString("  │\n")
		}

		sb.WriteString("  └────────────────────────────────────────────────────────────────────┘\n")
	}

	// No links message
	if len(state.BackLinks) == 0 && len(state.OutLinks) == 0 {
		sb.WriteString("\n  (No links found for this note)\n")
	}
}

// renderCenterBox renders the center node with special styling
func renderCenterBox(sb *strings.Builder, title string, selected bool) {
	selector := "  "
	if selected {
		selector = "> "
	}

	// Top border
	sb.WriteString(fmt.Sprintf("                       %s╔%s╗\n", selector, strings.Repeat("═", boxWidth)))

	// CENTER label line - center the [CENTER] text
	centerLabel := "[CENTER]"
	centerPadding := maxInt(0, (boxWidth-len(centerLabel))/2)
	centerRightPad := maxInt(0, boxWidth-len(centerLabel)-centerPadding)
	sb.WriteString(fmt.Sprintf("                       %s║%s%s%s║\n",
		selector,
		strings.Repeat(" ", centerPadding),
		centerLabel,
		strings.Repeat(" ", centerRightPad)))

	// Note title line
	titlePadding := maxInt(0, (boxWidth-len(title))/2)
	titleRightPad := maxInt(0, boxWidth-len(title)-titlePadding)
	sb.WriteString(fmt.Sprintf("                       %s║%s%s%s║\n",
		selector,
		strings.Repeat(" ", titlePadding),
		title,
		strings.Repeat(" ", titleRightPad)))

	// Bottom border
	sb.WriteString(fmt.Sprintf("                       %s╚%s╝\n", selector, strings.Repeat("═", boxWidth)))
}

// renderNodeBox renders a regular node box
func renderNodeBox(sb *strings.Builder, title string, selected bool, prefix string) {
	selector := "  "
	if selected {
		selector = "> "
	}

	// Calculate padding with safety checks
	titlePadding := maxInt(0, (boxWidth-len(title)-2)/2)
	titleRightPad := maxInt(0, boxWidth-len(title)-2-titlePadding)

	// Single line box
	sb.WriteString(fmt.Sprintf("%s%s┌%s┐  ", prefix, selector, strings.Repeat("─", boxWidth-2)))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s%s│%s%s%s│  ", prefix, selector, strings.Repeat(" ", titlePadding), title, strings.Repeat(" ", titleRightPad)))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s%s└%s┘  ", prefix, selector, strings.Repeat("─", boxWidth-2)))
}

// truncateTitle truncates a title to fit in the box
func truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen-3] + "..."
}
