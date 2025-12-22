package presentation

import (
	"cli-notes/scripts/data"
	"fmt"
	"strings"
)

// searchDimensions holds the calculated UI dimensions for search view
type searchDimensions struct {
	terminalWidth   int
	terminalHeight  int
	leftPanelWidth  int
	rightPanelWidth int
	visibleResults  int
}

// calculateSearchDimensions calculates panel widths based on terminal size
func calculateSearchDimensions(termWidth, termHeight int) searchDimensions {
	// Calculate panel widths proportionally (~50% each)
	leftWidth := int(float64(termWidth) * 0.5)
	rightWidth := termWidth - leftWidth - 3 // -3 for borders

	// Calculate visible results (account for 11 fixed UI lines: 7 above content + 3 below + 1 for prompt)
	visibleResults := termHeight - 11
	if visibleResults < 5 {
		visibleResults = 5
	}

	return searchDimensions{
		terminalWidth:   termWidth,
		terminalHeight:  termHeight,
		leftPanelWidth:  leftWidth,
		rightPanelWidth: rightWidth,
		visibleResults:  visibleResults,
	}
}

// RenderSearchView renders the full-screen search interface
func RenderSearchView(state *data.SearchState, termWidth, termHeight int) string {
	var output strings.Builder

	state.TermWidth = termWidth
	state.TermHeight = termHeight

	dims := calculateSearchDimensions(termWidth, termHeight)

	// Clear screen
	output.WriteString("\033[2J\033[H")

	// Top border
	output.WriteString("┌" + strings.Repeat("─", termWidth-2) + "┐\n")

	// Header
	headerLeft := " SEARCH NOTES"
	headerRight := "q:Quit "
	headerPadding := termWidth - len(headerLeft) - len(headerRight) - 2
	if headerPadding < 0 {
		headerPadding = 0
	}
	output.WriteString(fmt.Sprintf("│%s%s%s│\n", headerLeft, strings.Repeat(" ", headerPadding), headerRight))

	// Separator
	output.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")

	// Search input line
	inputLine := fmt.Sprintf(" > %s_", state.Query)
	inputPadding := termWidth - len([]rune(inputLine)) - 2
	if inputPadding < 0 {
		inputPadding = 0
		// Truncate query if too long
		maxQueryLen := termWidth - 6
		if len(state.Query) > maxQueryLen {
			inputLine = fmt.Sprintf(" > ...%s_", state.Query[len(state.Query)-maxQueryLen+3:])
			inputPadding = 0
		}
	}
	output.WriteString(fmt.Sprintf("│%s%s│\n", inputLine, strings.Repeat(" ", inputPadding)))

	// Separator with match count and filter mode
	var filterLabel string
	switch state.FilterMode {
	case data.ShowAll:
		filterLabel = "All"
	case data.ShowIncompleteOnly:
		filterLabel = "Open"
	case data.ShowCompleteOnly:
		filterLabel = "Done"
	}
	matchCount := fmt.Sprintf(" %d matches | %s ", len(state.Results), filterLabel)
	separatorLen := termWidth - len(matchCount) - 2
	leftSep := separatorLen / 2
	rightSep := separatorLen - leftSep
	output.WriteString(fmt.Sprintf("├%s%s%s┤\n", strings.Repeat("─", leftSep), matchCount, strings.Repeat("─", rightSep)))

	// Split panel header
	output.WriteString("│" + padRight(" Results", dims.leftPanelWidth) + "│" + padRight(" Preview", dims.rightPanelWidth) + "│\n")
	output.WriteString("├" + strings.Repeat("─", dims.leftPanelWidth) + "┼" + strings.Repeat("─", dims.rightPanelWidth) + "┤\n")

	// Build panel content
	leftLines := buildSearchResultsPanel(state, dims)
	rightLines := buildSearchPreviewPanel(state, dims)

	// Render content rows
	contentRows := dims.visibleResults
	for i := 0; i < contentRows; i++ {
		leftContent := ""
		rightContent := ""
		if i < len(leftLines) {
			leftContent = leftLines[i]
		}
		if i < len(rightLines) {
			rightContent = rightLines[i]
		}
		output.WriteString(renderSearchSplitLine(leftContent, rightContent, dims))
	}

	// Bottom panel border
	output.WriteString("├" + strings.Repeat("─", dims.leftPanelWidth) + "┴" + strings.Repeat("─", dims.rightPanelWidth) + "┤\n")

	// Controls - mode-specific
	var controls string
	switch state.ViewMode {
	case data.SearchModeInsert:
		controls = " [INSERT] Type to search | Esc/Enter:Normal | ↑↓:Navigate"
	case data.SearchModeNormal:
		controls = " [NORMAL] i:Ins j/k:Nav f:Flt d:Done 1-3:Pri t:Today l:Link L:Graph o:Obj O:View q:Quit"
	case data.SearchModeActions:
		controls = " [ACTIONS] j/k:Navigate  Enter:Execute  Esc:Back"
	}
	controlsPadding := termWidth - len([]rune(controls)) - 2
	if controlsPadding < 0 {
		controlsPadding = 0
	}
	output.WriteString(fmt.Sprintf("│%s%s│\n", controls, strings.Repeat(" ", controlsPadding)))

	// Bottom border
	output.WriteString("└" + strings.Repeat("─", termWidth-2) + "┘\n")

	// Render actions overlay if in actions mode
	if state.ViewMode == data.SearchModeActions {
		output.WriteString(renderActionsOverlay(state, dims))
	}

	return output.String()
}

// buildSearchResultsPanel builds the left panel content (results list)
func buildSearchResultsPanel(state *data.SearchState, dims searchDimensions) []string {
	var lines []string

	if len(state.Results) == 0 {
		lines = append(lines, " No results found")
		return lines
	}

	// Calculate visible range
	visibleEnd := state.ScrollOffset + dims.visibleResults
	if visibleEnd > len(state.Results) {
		visibleEnd = len(state.Results)
	}

	for i := state.ScrollOffset; i < visibleEnd; i++ {
		result := state.Results[i]

		// Selection indicator
		indicator := "  "
		if i == state.SelectedIndex {
			indicator = "► "
		}

		// Priority indicator
		priorityStr := ""
		if result.File.Priority > 0 && result.File.Priority <= 3 {
			priorityStr = fmt.Sprintf("[P%d] ", result.File.Priority)
		}

		// Done indicator
		doneStr := ""
		if result.File.Done {
			doneStr = "✓ "
		}

		// Build title line
		titleLine := fmt.Sprintf("%s%s%s%s", indicator, doneStr, priorityStr, result.File.Title)

		// Truncate if needed
		maxLen := dims.leftPanelWidth - 1
		titleRunes := []rune(titleLine)
		if len(titleRunes) > maxLen {
			titleLine = string(titleRunes[:maxLen-3]) + "..."
		}

		lines = append(lines, titleLine)
	}

	return lines
}

// buildSearchPreviewPanel builds the right panel content (preview)
func buildSearchPreviewPanel(state *data.SearchState, dims searchDimensions) []string {
	var lines []string

	result := state.GetSelectedResult()
	if result == nil {
		lines = append(lines, " (no selection)")
		return lines
	}

	// Title
	titleLine := " " + strings.ToUpper(result.File.Title)
	titleRunes := []rune(titleLine)
	if len(titleRunes) > dims.rightPanelWidth-1 {
		titleLine = string(titleRunes[:dims.rightPanelWidth-4]) + "..."
	}
	lines = append(lines, titleLine)

	// Tags
	if len(result.File.Tags) > 0 {
		tagsLine := " Tags: " + strings.Join(result.File.Tags, ", ")
		tagsRunes := []rune(tagsLine)
		if len(tagsRunes) > dims.rightPanelWidth-1 {
			tagsLine = string(tagsRunes[:dims.rightPanelWidth-4]) + "..."
		}
		lines = append(lines, tagsLine)
	}

	// Due date
	if !result.File.DueAt.IsZero() && result.File.DueAt.Year() < 2100 {
		lines = append(lines, fmt.Sprintf(" Due: %s", result.File.DueAt.Format("2006-01-02")))
	}

	// Priority
	if result.File.Priority > 0 {
		lines = append(lines, fmt.Sprintf(" Priority: P%d", result.File.Priority))
	}

	// Status
	if result.File.Done {
		lines = append(lines, " Status: Complete")
	} else {
		lines = append(lines, " Status: Open")
	}

	// Separator
	lines = append(lines, " "+strings.Repeat("─", dims.rightPanelWidth-3))

	// Content snippet
	if result.ContentSnippet != "" {
		snippetPrefix := " "
		if result.SnippetLine > 0 {
			snippetPrefix = fmt.Sprintf(" L%d: ", result.SnippetLine)
		}

		snippetLine := snippetPrefix + result.ContentSnippet
		snippetRunes := []rune(snippetLine)
		if len(snippetRunes) > dims.rightPanelWidth-1 {
			snippetLine = string(snippetRunes[:dims.rightPanelWidth-4]) + "..."
		}
		lines = append(lines, snippetLine)
	}

	return lines
}

// renderSearchSplitLine renders a line split between left and right panels
func renderSearchSplitLine(leftContent, rightContent string, dims searchDimensions) string {
	leftRunes := []rune(leftContent)
	rightRunes := []rune(rightContent)

	// Pad left content
	leftPadding := dims.leftPanelWidth - len(leftRunes)
	if leftPadding < 0 {
		leftPadding = 0
		leftContent = string(leftRunes[:dims.leftPanelWidth])
	}

	// Pad right content
	rightPadding := dims.rightPanelWidth - len(rightRunes)
	if rightPadding < 0 {
		rightPadding = 0
		rightContent = string(rightRunes[:dims.rightPanelWidth])
	}

	return fmt.Sprintf("│%s%s│%s%s│\n",
		leftContent,
		strings.Repeat(" ", leftPadding),
		rightContent,
		strings.Repeat(" ", rightPadding),
	)
}

// renderActionsOverlay renders the quick actions menu overlay
func renderActionsOverlay(state *data.SearchState, dims searchDimensions) string {
	var output strings.Builder

	actions := state.GetAvailableActions()
	if len(actions) == 0 {
		return ""
	}

	// Calculate overlay position and size
	overlayWidth := 30
	overlayHeight := len(actions) + 4 // +4 for borders and header

	// Position in center of screen
	startCol := (dims.terminalWidth - overlayWidth) / 2
	startRow := (dims.terminalHeight - overlayHeight) / 2

	// Move cursor to overlay position and draw
	output.WriteString(fmt.Sprintf("\033[%d;%dH", startRow, startCol))

	// Top border
	output.WriteString("┌" + strings.Repeat("─", overlayWidth-2) + "┐")

	// Header
	output.WriteString(fmt.Sprintf("\033[%d;%dH", startRow+1, startCol))
	header := " Quick Actions"
	output.WriteString("│" + header + strings.Repeat(" ", overlayWidth-len(header)-2) + "│")

	// Separator
	output.WriteString(fmt.Sprintf("\033[%d;%dH", startRow+2, startCol))
	output.WriteString("├" + strings.Repeat("─", overlayWidth-2) + "┤")

	// Actions
	for i, action := range actions {
		output.WriteString(fmt.Sprintf("\033[%d;%dH", startRow+3+i, startCol))

		indicator := "  "
		if i == state.ActionsIndex {
			indicator = "► "
		}

		line := fmt.Sprintf("%s%s", indicator, action.Label)
		if len(line) > overlayWidth-2 {
			line = line[:overlayWidth-5] + "..."
		}

		padding := overlayWidth - len(line) - 2
		if padding < 0 {
			padding = 0
		}

		output.WriteString("│" + line + strings.Repeat(" ", padding) + "│")
	}

	// Bottom border
	output.WriteString(fmt.Sprintf("\033[%d;%dH", startRow+3+len(actions), startCol))
	output.WriteString("└" + strings.Repeat("─", overlayWidth-2) + "┘")

	return output.String()
}

// padRight pads a string to the specified length
func padRight(s string, length int) string {
	sRunes := []rune(s)
	if len(sRunes) >= length {
		return string(sRunes[:length])
	}
	return s + strings.Repeat(" ", length-len(sRunes))
}
