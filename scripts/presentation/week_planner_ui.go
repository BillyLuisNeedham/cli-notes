package presentation

import (
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"fmt"
	"strings"
)

// uiDimensions holds the calculated UI dimensions
type uiDimensions struct {
	terminalWidth   int
	terminalHeight  int
	leftPanelWidth  int
	rightPanelWidth int
}

// calculateDimensions calculates panel widths based on terminal size
func calculateDimensions(termWidth, termHeight int) uiDimensions {
	// Calculate panel widths proportionally (~60% left, ~40% right)
	leftWidth := int(float64(termWidth) * 0.6)
	rightWidth := termWidth - leftWidth - 3 // -3 for borders

	return uiDimensions{
		terminalWidth:   termWidth,
		terminalHeight:  termHeight,
		leftPanelWidth:  leftWidth,
		rightPanelWidth: rightWidth,
	}
}

// RenderWeekView renders the complete week planner UI
func RenderWeekView(state *data.WeekPlannerState, termWidth, termHeight int) string {
	// Route to appropriate view based on ViewMode
	if state.ViewMode == data.ExpandedEarlierView {
		return RenderExpandedEarlierView(state, termWidth, termHeight)
	}

	// Normal view
	dims := calculateDimensions(termWidth, termHeight)
	var output strings.Builder

	// Clear screen (ANSI escape code)
	output.WriteString("\033[2J\033[H")

	// Render top border
	output.WriteString("┌" + strings.Repeat("─", dims.terminalWidth-2) + "┐\n")

	// Render header
	output.WriteString(renderHeader(state, dims))

	// Render day tabs
	output.WriteString(renderDayTabs(state, dims))

	// Render controls help
	output.WriteString(renderControlsBar(state, dims))

	// Render main content split
	output.WriteString(renderSplitBorder(dims))

	// Render content rows (left panel + right panel)
	contentLines := renderContent(state, dims)
	for _, line := range contentLines {
		output.WriteString(line)
	}

	// Render bottom border
	output.WriteString("└" + strings.Repeat("─", dims.terminalWidth-2) + "┘\n")

	return output.String()
}

// renderHeader renders the title and status bar
func renderHeader(state *data.WeekPlannerState, dims uiDimensions) string {
	plan := state.Plan
	startDate := plan.StartDate.Format("Jan 02")
	endDate := plan.EndDate.Format("Jan 02, 2006")
	title := fmt.Sprintf("WEEK PLANNER (%s - %s)", startDate, endDate)

	changesIndicator := ""
	if plan.HasChanges() {
		changesIndicator = fmt.Sprintf("[*] Changes: %d", len(plan.Changes))
	} else {
		changesIndicator = "No changes"
	}

	// Center the title, right-align the changes
	titlePadding := (dims.terminalWidth - len(title) - len(changesIndicator) - 4) / 2
	if titlePadding < 1 {
		titlePadding = 1
	}

	line := fmt.Sprintf("│ %s%s%s%s%s │\n",
		strings.Repeat(" ", titlePadding),
		title,
		strings.Repeat(" ", titlePadding),
		strings.Repeat(" ", dims.terminalWidth-len(title)-len(changesIndicator)-4-2*titlePadding),
		changesIndicator,
	)

	return line
}

// renderDayTabs renders the day selection tabs
func renderDayTabs(state *data.WeekPlannerState, dims uiDimensions) string {
	var tabs strings.Builder
	tabs.WriteString("│ ")

	days := []data.WeekDay{
		data.Earlier, data.Monday, data.Tuesday, data.Wednesday, data.Thursday,
		data.Friday, data.Saturday, data.Sunday, data.NextMonday,
	}

	for _, day := range days {
		count := state.Plan.GetTodoCount(day)
		shortName := data.WeekDayShortNames[day]

		var tab string
		if day == state.SelectedDay {
			// Selected day - highlighted with brackets
			tab = fmt.Sprintf("[%s(%d)]", shortName, count)
		} else {
			tab = fmt.Sprintf(" %s(%d) ", shortName, count)
		}

		tabs.WriteString(tab)
		if day != data.NextMonday {
			tabs.WriteString(" ")
		}
	}

	// Pad to full width
	currentLen := tabs.Len() - 2 // -2 for initial "│ "
	padding := dims.terminalWidth - currentLen - 4 // -4 for initial "│ " and ending " │"
	tabs.WriteString(strings.Repeat(" ", padding))
	tabs.WriteString(" │\n")

	return tabs.String()
}

// renderControlsBar renders the controls help bar
func renderControlsBar(state *data.WeekPlannerState, dims uiDimensions) string {
	controls := "j/k:Sel │ h/l:Move │ mtwrfas:Day │ MTWRFAS:MoveTo │ e:Earlier │ ^S:Save │ u:Undo │ q:Quit"
	padding := dims.terminalWidth - len(controls) - 4 // -4 for "│ " and " │"
	if padding < 0 {
		padding = 0
	}
	return fmt.Sprintf("│ %s%s │\n", controls, strings.Repeat(" ", padding))
}

// renderSplitBorder renders the border between header and content with split
func renderSplitBorder(dims uiDimensions) string {
	return fmt.Sprintf("├%s┬%s┤\n",
		strings.Repeat("─", dims.leftPanelWidth),
		strings.Repeat("─", dims.rightPanelWidth),
	)
}

// renderContent renders the main content area (both panels)
func renderContent(state *data.WeekPlannerState, dims uiDimensions) []string {
	lines := make([]string, 0)

	// Render panel titles
	leftTitle := fmt.Sprintf("  %s (%d todos)",
		data.WeekDayNames[state.SelectedDay],
		state.Plan.GetTodoCount(state.SelectedDay))
	rightTitle := "  WEEK OVERVIEW"

	lines = append(lines, renderSplitLine(leftTitle, rightTitle, dims))
	lines = append(lines, renderSplitLine("", "", dims))

	// Get todos for selected day
	todos := state.Plan.TodosByDay[state.SelectedDay]

	// Render todos on left, overview on right
	maxLines := 15 // Maximum number of content lines

	for i := 0; i < maxLines; i++ {
		var leftContent string
		var rightContent string

		// Left panel: todos
		if i < len(todos) {
			todo := todos[i]
			isSelected := i == state.SelectedTodo

			// Truncate title if too long
			maxTitleLen := dims.leftPanelWidth - 10 // Account for priority and selector
			title := todo.Title
			if len(title) > maxTitleLen {
				title = title[:maxTitleLen-3] + "..."
			}

			selector := "  "
			if isSelected {
				selector = "► "
			}

			leftContent = fmt.Sprintf("%s[P%d] %s", selector, todo.Priority, title)
		} else if i == len(todos) && len(state.Plan.Changes) > 0 {
			// Show recent changes after todos list
			leftContent = ""
		} else if i > len(todos) && i <= len(todos)+2 && len(state.Plan.Changes) > 0 {
			// Show last 2 changes
			changeIdx := len(state.Plan.Changes) - (i - len(todos))
			if changeIdx >= 0 && changeIdx < len(state.Plan.Changes) {
				change := state.Plan.Changes[changeIdx]
				fromDay := data.WeekDayShortNames[change.FromDay]
				toDay := data.WeekDayShortNames[change.ToDay]

				maxChangeLen := dims.leftPanelWidth - 4
				changeText := fmt.Sprintf("Moved %s→%s: %s", fromDay, toDay, change.Todo.Title)
				if len(changeText) > maxChangeLen {
					changeText = changeText[:maxChangeLen-3] + "..."
				}
				leftContent = fmt.Sprintf("  %s", changeText)
			} else {
				leftContent = ""
			}
		} else {
			leftContent = ""
		}

		// Right panel: week overview (Earlier + 7 days + NextMonday)
		if i < 10 { // Earlier + 7 days + 1 for next monday + 1 for blank
			if i == 0 {
				rightContent = renderOverviewLine(state, data.Earlier)
			} else if i == 1 {
				rightContent = renderOverviewLine(state, data.Monday)
			} else if i == 2 {
				rightContent = renderOverviewLine(state, data.Tuesday)
			} else if i == 3 {
				rightContent = renderOverviewLine(state, data.Wednesday)
			} else if i == 4 {
				rightContent = renderOverviewLine(state, data.Thursday)
			} else if i == 5 {
				rightContent = renderOverviewLine(state, data.Friday)
			} else if i == 6 {
				rightContent = renderOverviewLine(state, data.Saturday)
			} else if i == 7 {
				rightContent = renderOverviewLine(state, data.Sunday)
			} else if i == 8 {
				rightContent = ""
			} else if i == 9 {
				rightContent = renderOverviewLine(state, data.NextMonday)
			}
		} else if i == 11 {
			rightContent = "  Legend:"
		} else if i == 12 {
			rightContent = "  ▓ Priority 1 (urgent)"
		} else if i == 13 {
			rightContent = "  █ Priority 2 (normal)"
		} else if i == 14 {
			rightContent = "  ░ Priority 3 (low)"
		} else {
			rightContent = ""
		}

		lines = append(lines, renderSplitLine(leftContent, rightContent, dims))
	}

	// Add control hints at bottom of left panel
	lines = append(lines, renderSplitLine("", "", dims))
	lines = append(lines, renderSplitLine("  Controls:", "", dims))
	lines = append(lines, renderSplitLine("  • j/k Select todo", "", dims))
	lines = append(lines, renderSplitLine("  • h/l Move to prev/next day", "", dims))
	lines = append(lines, renderSplitLine("  • Enter Open note", "", dims))
	lines = append(lines, renderSplitLine("  • m/t/w/r/f/a/s Switch to day", "", dims))
	lines = append(lines, renderSplitLine("  • M/T/W/R/F/A/S Move todo to day", "", dims))
	lines = append(lines, renderSplitLine("  • e Show earlier todos", "", dims))
	lines = append(lines, renderSplitLine("  • Ctrl+S Save changes", "", dims))

	return lines
}

// renderOverviewLine renders one line of the week overview
func renderOverviewLine(state *data.WeekPlannerState, day data.WeekDay) string {
	dayName := data.WeekDayShortNames[day]
	count := state.Plan.GetTodoCount(day)
	todos := state.Plan.TodosByDay[day]

	// Build bar chart
	var bars strings.Builder
	for _, todo := range todos {
		switch todo.Priority {
		case scripts.P1:
			bars.WriteString("▓")
		case scripts.P2:
			bars.WriteString("█")
		case scripts.P3:
			bars.WriteString("░")
		}
	}

	// Add current day indicator
	indicator := " "
	if day == state.SelectedDay {
		indicator = "◄"
	}

	// Calculate spacing, ensuring it never goes negative
	spacing := 15 - bars.Len()
	if spacing < 0 {
		spacing = 0
	}

	return fmt.Sprintf("  %-4s %s%s(%d) %s",
		dayName,
		bars.String(),
		strings.Repeat(" ", spacing),
		count,
		indicator,
	)
}

// renderSplitLine renders a line split between left and right panels
func renderSplitLine(leftContent, rightContent string, dims uiDimensions) string {
	// Pad left content to left panel width
	leftPadding := dims.leftPanelWidth - len(leftContent)
	if leftPadding < 0 {
		leftPadding = 0
		leftContent = leftContent[:dims.leftPanelWidth]
	}

	// Pad right content to right panel width
	rightPadding := dims.rightPanelWidth - len(rightContent)
	if rightPadding < 0 {
		rightPadding = 0
		rightContent = rightContent[:dims.rightPanelWidth]
	}

	return fmt.Sprintf("│%s%s│%s%s│\n",
		leftContent,
		strings.Repeat(" ", leftPadding),
		rightContent,
		strings.Repeat(" ", rightPadding),
	)
}

// RenderExpandedEarlierView renders a full-screen view of all Earlier todos
func RenderExpandedEarlierView(state *data.WeekPlannerState, termWidth, termHeight int) string {
	var output strings.Builder

	// Clear screen
	output.WriteString("\033[2J\033[H")

	// Render top border
	output.WriteString("┌" + strings.Repeat("─", termWidth-2) + "┐\n")

	// Render header
	plan := state.Plan
	startDate := plan.StartDate.Format("Jan 02")
	title := fmt.Sprintf("EARLIER TODOS (before %s)", startDate)

	changesIndicator := ""
	if plan.HasChanges() {
		changesIndicator = fmt.Sprintf("[*] Changes: %d", len(plan.Changes))
	} else {
		changesIndicator = "No changes"
	}

	// Center the title, right-align the changes
	titlePadding := (termWidth - len(title) - len(changesIndicator) - 4) / 2
	if titlePadding < 1 {
		titlePadding = 1
	}

	output.WriteString(fmt.Sprintf("│ %s%s%s%s%s │\n",
		strings.Repeat(" ", titlePadding),
		title,
		strings.Repeat(" ", titlePadding),
		strings.Repeat(" ", termWidth-len(title)-len(changesIndicator)-4-2*titlePadding),
		changesIndicator,
	))

	// Render controls bar
	controls := "j/k:Sel │ MTWRFAS:MoveTo │ Enter:Open │ ESC/e:Exit │ ^S:Save │ u:Undo │ q:Quit"
	padding := termWidth - len(controls) - 4
	if padding < 0 {
		padding = 0
	}
	output.WriteString(fmt.Sprintf("│ %s%s │\n", controls, strings.Repeat(" ", padding)))

	// Render separator
	output.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")

	// Calculate content area
	headerLines := 4  // Top border + header + controls + separator
	footerLines := 2  // Scroll indicator + bottom border
	contentHeight := termHeight - headerLines - footerLines

	// Get Earlier todos
	earlierTodos := state.Plan.TodosByDay[data.Earlier]
	todoCount := len(earlierTodos)

	// Adjust scroll offset to keep selected todo visible
	state.AdjustScrollOffset(contentHeight)

	// Render todos in viewport
	for i := 0; i < contentHeight; i++ {
		todoIdx := state.ScrollOffset + i

		var line string
		if todoIdx < todoCount {
			todo := earlierTodos[todoIdx]
			isSelected := todoIdx == state.SelectedTodo

			// Format: [►] [P1] Due: Jan 15 │ Task title here...
			selector := "  "
			if isSelected {
				selector = "► "
			}

			dueDate := todo.DueAt.Format("Jan 02")
			maxTitleLen := termWidth - 30 // Reserve space for selector, priority, date
			if maxTitleLen < 20 {
				maxTitleLen = 20
			}
			title := todo.Title
			if len(title) > maxTitleLen {
				title = title[:maxTitleLen-3] + "..."
			}

			line = fmt.Sprintf("%s[P%d] Due: %s │ %s", selector, todo.Priority, dueDate, title)
		} else {
			line = ""
		}

		// Pad to full width
		linePadding := termWidth - len(line) - 4
		if linePadding < 0 {
			linePadding = 0
			if len(line) > termWidth-4 {
				line = line[:termWidth-4]
			}
		}

		output.WriteString(fmt.Sprintf("│ %s%s │\n", line, strings.Repeat(" ", linePadding)))
	}

	// Render scroll indicator
	scrollInfo := ""
	if todoCount > 0 {
		visibleStart := state.ScrollOffset + 1
		visibleEnd := state.ScrollOffset + contentHeight
		if visibleEnd > todoCount {
			visibleEnd = todoCount
		}
		scrollInfo = fmt.Sprintf("Showing %d-%d of %d todos", visibleStart, visibleEnd, todoCount)
	} else {
		scrollInfo = "No Earlier todos"
	}

	scrollPadding := termWidth - len(scrollInfo) - 4
	output.WriteString(fmt.Sprintf("│ %s%s │\n", scrollInfo, strings.Repeat(" ", scrollPadding)))

	// Render bottom border
	output.WriteString("└" + strings.Repeat("─", termWidth-2) + "┘\n")

	return output.String()
}
