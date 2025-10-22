package presentation

import (
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"fmt"
	"strings"
)

const (
	// Terminal dimensions (adjust as needed)
	terminalWidth = 90
	leftPanelWidth = 54
	rightPanelWidth = terminalWidth - leftPanelWidth - 3 // -3 for borders
)

// RenderWeekView renders the complete week planner UI
func RenderWeekView(state *data.WeekPlannerState) string {
	var output strings.Builder

	// Clear screen (ANSI escape code)
	output.WriteString("\033[2J\033[H")

	// Render top border
	output.WriteString("┌" + strings.Repeat("─", terminalWidth-2) + "┐\n")

	// Render header
	output.WriteString(renderHeader(state))

	// Render day tabs
	output.WriteString(renderDayTabs(state))

	// Render controls help
	output.WriteString(renderControlsBar(state))

	// Render main content split
	output.WriteString(renderSplitBorder())

	// Render content rows (left panel + right panel)
	contentLines := renderContent(state)
	for _, line := range contentLines {
		output.WriteString(line)
	}

	// Render bottom border
	output.WriteString("└" + strings.Repeat("─", terminalWidth-2) + "┘\n")

	return output.String()
}

// renderHeader renders the title and status bar
func renderHeader(state *data.WeekPlannerState) string {
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
	titlePadding := (terminalWidth - len(title) - len(changesIndicator) - 4) / 2
	if titlePadding < 1 {
		titlePadding = 1
	}

	line := fmt.Sprintf("│ %s%s%s%s%s │\n",
		strings.Repeat(" ", titlePadding),
		title,
		strings.Repeat(" ", titlePadding),
		strings.Repeat(" ", terminalWidth-len(title)-len(changesIndicator)-4-2*titlePadding),
		changesIndicator,
	)

	return line
}

// renderDayTabs renders the day selection tabs
func renderDayTabs(state *data.WeekPlannerState) string {
	var tabs strings.Builder
	tabs.WriteString("│ ")

	days := []data.WeekDay{
		data.Monday, data.Tuesday, data.Wednesday, data.Thursday,
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
	padding := terminalWidth - currentLen - 3 // -3 for " │\n"
	tabs.WriteString(strings.Repeat(" ", padding))
	tabs.WriteString("│\n")

	return tabs.String()
}

// renderControlsBar renders the controls help bar
func renderControlsBar(state *data.WeekPlannerState) string {
	controls := "j/k:Select │ h/l:Move │ n:Next Mon │ u:Undo r:Redo │ s:Save x:Reset q:Quit"
	padding := terminalWidth - len(controls) - 3
	return fmt.Sprintf("│ %s%s │\n", controls, strings.Repeat(" ", padding))
}

// renderSplitBorder renders the border between header and content with split
func renderSplitBorder() string {
	return fmt.Sprintf("├%s┬%s┤\n",
		strings.Repeat("─", leftPanelWidth),
		strings.Repeat("─", rightPanelWidth),
	)
}

// renderContent renders the main content area (both panels)
func renderContent(state *data.WeekPlannerState) []string {
	lines := make([]string, 0)

	// Render panel titles
	leftTitle := fmt.Sprintf("  %s (%d todos)",
		data.WeekDayNames[state.SelectedDay],
		state.Plan.GetTodoCount(state.SelectedDay))
	rightTitle := "  WEEK OVERVIEW"

	lines = append(lines, renderSplitLine(leftTitle, rightTitle))
	lines = append(lines, renderSplitLine("", ""))

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
			maxTitleLen := leftPanelWidth - 10 // Account for priority and selector
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

				maxChangeLen := leftPanelWidth - 4
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

		// Right panel: week overview
		if i < 9 { // 7 days + 1 for next monday + 1 for legend
			if i == 0 {
				rightContent = renderOverviewLine(state, data.Monday)
			} else if i == 1 {
				rightContent = renderOverviewLine(state, data.Tuesday)
			} else if i == 2 {
				rightContent = renderOverviewLine(state, data.Wednesday)
			} else if i == 3 {
				rightContent = renderOverviewLine(state, data.Thursday)
			} else if i == 4 {
				rightContent = renderOverviewLine(state, data.Friday)
			} else if i == 5 {
				rightContent = renderOverviewLine(state, data.Saturday)
			} else if i == 6 {
				rightContent = renderOverviewLine(state, data.Sunday)
			} else if i == 7 {
				rightContent = ""
			} else if i == 8 {
				rightContent = renderOverviewLine(state, data.NextMonday)
			}
		} else if i == 10 {
			rightContent = "  Legend:"
		} else if i == 11 {
			rightContent = "  ▓ Priority 1 (urgent)"
		} else if i == 12 {
			rightContent = "  █ Priority 2 (normal)"
		} else if i == 13 {
			rightContent = "  ░ Priority 3 (low)"
		} else {
			rightContent = ""
		}

		lines = append(lines, renderSplitLine(leftContent, rightContent))
	}

	// Add control hints at bottom of left panel
	lines = append(lines, renderSplitLine("", ""))
	lines = append(lines, renderSplitLine("  Controls:", ""))
	lines = append(lines, renderSplitLine("  • j/k Select todo", ""))
	lines = append(lines, renderSplitLine("  • h/l Move to prev/next day", ""))
	lines = append(lines, renderSplitLine("  • Enter Open note", ""))
	lines = append(lines, renderSplitLine("  • Tab Switch day", ""))
	lines = append(lines, renderSplitLine("  • m/tu/w/th/f/sa/su Day shortcuts", ""))

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

	return fmt.Sprintf("  %-4s %s%s(%d) %s",
		dayName,
		bars.String(),
		strings.Repeat(" ", 15-bars.Len()),
		count,
		indicator,
	)
}

// renderSplitLine renders a line split between left and right panels
func renderSplitLine(leftContent, rightContent string) string {
	// Pad left content to left panel width
	leftPadding := leftPanelWidth - len(leftContent)
	if leftPadding < 0 {
		leftPadding = 0
		leftContent = leftContent[:leftPanelWidth]
	}

	// Pad right content to right panel width
	rightPadding := rightPanelWidth - len(rightContent)
	if rightPadding < 0 {
		rightPadding = 0
		rightContent = rightContent[:rightPanelWidth]
	}

	return fmt.Sprintf("│%s%s│%s%s│\n",
		leftContent,
		strings.Repeat(" ", leftPadding),
		rightContent,
		strings.Repeat(" ", rightPadding),
	)
}
