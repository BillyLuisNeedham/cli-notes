package presentation

import (
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"fmt"
	"strings"
)

// objectivesDimensions holds the calculated UI dimensions for objectives view
type objectivesDimensions struct {
	terminalWidth   int
	terminalHeight  int
	leftPanelWidth  int
	rightPanelWidth int
}

// calculateObjectivesDimensions calculates panel widths based on terminal size
func calculateObjectivesDimensions(termWidth, termHeight int) objectivesDimensions {
	// Calculate panel widths proportionally (~60% left, ~40% right)
	leftWidth := int(float64(termWidth) * 0.6)
	rightWidth := termWidth - leftWidth - 3 // -3 for borders

	return objectivesDimensions{
		terminalWidth:   termWidth,
		terminalHeight:  termHeight,
		leftPanelWidth:  leftWidth,
		rightPanelWidth: rightWidth,
	}
}

// renderObjectivesSplitLine renders a line split between left and right panels
func renderObjectivesSplitLine(leftContent, rightContent string, dims objectivesDimensions) string {
	// Use rune count for proper length calculation with multi-byte characters
	leftLen := len([]rune(leftContent))
	rightLen := len([]rune(rightContent))

	// Pad left content to left panel width
	leftPadding := dims.leftPanelWidth - leftLen
	if leftPadding < 0 {
		leftPadding = 0
		// Truncate using runes
		leftRunes := []rune(leftContent)
		leftContent = string(leftRunes[:dims.leftPanelWidth])
	}

	// Pad right content to right panel width
	rightPadding := dims.rightPanelWidth - rightLen
	if rightPadding < 0 {
		rightPadding = 0
		// Truncate using runes
		rightRunes := []rune(rightContent)
		rightContent = string(rightRunes[:dims.rightPanelWidth])
	}

	return fmt.Sprintf("│%s%s│%s%s│\n",
		leftContent,
		strings.Repeat(" ", leftPadding),
		rightContent,
		strings.Repeat(" ", rightPadding),
	)
}

// RenderObjectivesListView renders the objectives list view
func RenderObjectivesListView(state *data.ObjectivesViewState) string {
	var output strings.Builder

	// Clear screen
	output.WriteString("\033[2J\033[H")

	// Header
	output.WriteString("================================\n")
	output.WriteString("OBJECTIVES\n")
	output.WriteString("================================\n\n")

	// Tab header
	activeCount, completedCount := state.GetListCounts()
	if state.ListFilterMode == data.ListShowActive {
		output.WriteString(fmt.Sprintf("[ACTIVE (%d)]  COMPLETED (%d)\n", activeCount, completedCount))
	} else {
		output.WriteString(fmt.Sprintf(" ACTIVE (%d)  [COMPLETED (%d)]\n", activeCount, completedCount))
	}
	output.WriteString("────────────────────────────────\n\n")

	if len(state.Objectives) == 0 {
		if state.ListFilterMode == data.ListShowActive {
			output.WriteString("No active objectives.\n\n")
		} else {
			output.WriteString("No completed objectives.\n\n")
		}
		output.WriteString("Press 'n' to create a new objective, 'f' to switch tab.\n")
	} else {
		for i, obj := range state.Objectives {
			// Get completion stats
			complete, total, err := data.GetCompletionStats(obj.ObjectiveID)
			if err != nil {
				complete, total = 0, 0
			}

			// Selection indicator
			if i == state.SelectedIndex {
				output.WriteString("> ")
			} else {
				output.WriteString("  ")
			}

			// Objective title and completion
			output.WriteString(fmt.Sprintf("%s (%d/%d complete)\n", obj.Title, complete, total))
		}
	}

	output.WriteString("\n")
	output.WriteString("j/k=navigate, f=switch tab, o=open, n=create, l=link, dd=delete, q=quit\n")

	return output.String()
}

// RenderSingleObjectiveView renders a single objective with its children in split screen layout
func RenderSingleObjectiveView(state *data.ObjectivesViewState, termWidth, termHeight int) string {
	var output strings.Builder

	if state.CurrentObjective == nil {
		return "Error: No objective selected\n"
	}

	dims := calculateObjectivesDimensions(termWidth, termHeight)

	// Get uncompleted tasks from selected child
	var uncompletedTasks []string
	selectedChild := state.GetSelectedChild()
	if selectedChild != nil {
		tasks, err := scripts.GetUncompletedTasksInFiles([]scripts.File{*selectedChild})
		if err == nil {
			uncompletedTasks = tasks
		}
	}

	// Clear screen
	output.WriteString("\033[2J\033[H")

	// Render top border
	output.WriteString("┌" + strings.Repeat("─", termWidth-2) + "┐\n")

	// Render header
	title := fmt.Sprintf("OBJECTIVE: %s [%s]",
		state.CurrentObjective.Title,
		state.CurrentObjective.ObjectiveID)
	titlePadding := (termWidth - len(title) - 4) / 2
	if titlePadding < 1 {
		titlePadding = 1
	}
	output.WriteString(fmt.Sprintf("│%s%s%s│\n",
		strings.Repeat(" ", titlePadding),
		title,
		strings.Repeat(" ", termWidth-len(title)-titlePadding-2)))

	// Render split border
	output.WriteString("├" + strings.Repeat("─", dims.leftPanelWidth) + "┬" + strings.Repeat("─", dims.rightPanelWidth) + "┤\n")

	// Build left panel content lines
	leftLines := buildLeftPanelLines(state, dims)

	// Build right panel content lines
	rightLines := buildRightPanelLines(selectedChild, uncompletedTasks, dims)

	// Render content rows (left panel + right panel)
	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	// Ensure minimum height
	if maxLines < termHeight-8 {
		maxLines = termHeight - 8
	}

	for i := 0; i < maxLines; i++ {
		leftContent := ""
		rightContent := ""
		if i < len(leftLines) {
			leftContent = leftLines[i]
		}
		if i < len(rightLines) {
			rightContent = rightLines[i]
		}
		output.WriteString(renderObjectivesSplitLine(leftContent, rightContent, dims))
	}

	// Render bottom border
	output.WriteString("├" + strings.Repeat("─", dims.leftPanelWidth) + "┴" + strings.Repeat("─", dims.rightPanelWidth) + "┤\n")

	// Render controls
	controls := "  j/k=navigate, o=open, n=new child, l=link, e=edit, u=unlink, s=sort, f=filter, q=back"
	controlsLen := len([]rune(controls))
	controlsPadding := termWidth - controlsLen - 2
	if controlsPadding < 0 {
		controlsPadding = 0
	}
	output.WriteString(fmt.Sprintf("│%s%s│\n", controls, strings.Repeat(" ", controlsPadding)))

	output.WriteString("└" + strings.Repeat("─", termWidth-2) + "┘\n")

	return output.String()
}

// buildLeftPanelLines builds content lines for the left panel (parent + children)
func buildLeftPanelLines(state *data.ObjectivesViewState, dims objectivesDimensions) []string {
	var lines []string

	// Parent content header
	parentIndicator := "  "
	if state.OnParent {
		parentIndicator = "► "
	}
	lines = append(lines, parentIndicator+"[PARENT] "+state.CurrentObjective.Title)

	// Parent content lines (first few lines of content)
	content := state.CurrentObjective.Content
	contentLines := strings.Split(content, "\n")
	for i, line := range contentLines {
		if i >= 3 { // Limit content preview to 3 lines
			lines = append(lines, "    ...")
			break
		}
		lines = append(lines, "    "+line)
	}

	lines = append(lines, "") // blank line

	// Get incomplete and complete children separately
	incomplete, complete := state.GetIncompleteAndCompleteChildren()
	incompleteCnt, completeCnt := state.GetCompletionCounts()

	// Linked todos header
	lines = append(lines, fmt.Sprintf("  LINKED TODOS (%d incomplete, %d complete)", incompleteCnt, completeCnt))
	lines = append(lines, "  "+strings.Repeat("─", dims.leftPanelWidth-4))

	if len(state.Children) == 0 {
		lines = append(lines, "  No linked todos yet.")
	} else {
		// Render incomplete todos
		if state.FilterMode != data.ShowCompleteOnly && len(incomplete) > 0 {
			lines = append(lines, "  INCOMPLETE:")
			lines = append(lines, buildChildLines(incomplete, 0, state)...)
			lines = append(lines, "")
		}

		// Render complete todos
		if state.FilterMode != data.ShowIncompleteOnly && len(complete) > 0 {
			lines = append(lines, "  COMPLETE:")
			lines = append(lines, buildChildLines(complete, len(incomplete), state)...)
		}
	}

	return lines
}

// buildChildLines builds content lines for a list of children
func buildChildLines(children []scripts.File, offset int, state *data.ObjectivesViewState) []string {
	var lines []string
	for i, child := range children {
		globalIndex := offset + i
		indicator := "  "
		if !state.OnParent && globalIndex == state.ChildSelectedIndex {
			indicator = "► "
		}

		line := fmt.Sprintf("  %s[P%d] %s", indicator, child.Priority, child.Title)
		if !child.DueAt.IsZero() && child.DueAt.Year() < 2100 {
			line += fmt.Sprintf(" (due: %s)", child.DueAt.Format("2006-01-02"))
		}
		lines = append(lines, line)
	}
	return lines
}

// buildRightPanelLines builds content lines for the right panel (open tasks)
func buildRightPanelLines(selectedChild *scripts.File, uncompletedTasks []string, dims objectivesDimensions) []string {
	var lines []string

	lines = append(lines, "  OPEN TASKS")
	lines = append(lines, "  "+strings.Repeat("─", dims.rightPanelWidth-4))

	if selectedChild == nil {
		lines = append(lines, "  (select a child todo)")
		lines = append(lines, "  (to see open tasks)")
	} else if len(uncompletedTasks) == 0 {
		lines = append(lines, "  No open tasks")
	} else {
		for _, task := range uncompletedTasks {
			// Clean up task string (remove filename prefix if present)
			parts := strings.SplitN(task, ": ", 2)
			if len(parts) > 1 {
				task = parts[1]
			}
			task = strings.TrimSpace(task)
			// Truncate if needed (use rune count)
			taskRunes := []rune(task)
			if len(taskRunes) > dims.rightPanelWidth-4 {
				task = string(taskRunes[:dims.rightPanelWidth-7]) + "..."
			}
			lines = append(lines, "  "+task)
		}
	}

	return lines
}

// renderChildrenList renders a list of children with proper selection
func renderChildrenList(output *strings.Builder, children []scripts.File, offset int, state *data.ObjectivesViewState) {
	for i, child := range children {
		// Calculate the global index for this child
		globalIndex := offset + i

		// Selection indicator
		if !state.OnParent && globalIndex == state.ChildSelectedIndex {
			output.WriteString("> ")
		} else {
			output.WriteString("  ")
		}

		// Priority and title
		output.WriteString(fmt.Sprintf("[P%d] %s", child.Priority, child.Title))

		// Due date if present
		if !child.DueAt.IsZero() && child.DueAt.Year() < 2100 {
			output.WriteString(fmt.Sprintf(" (due: %s)", child.DueAt.Format("2006-01-02")))
		}

		output.WriteString("\n")
	}
}
