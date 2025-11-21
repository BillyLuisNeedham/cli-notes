package presentation

import (
	"cli-notes/scripts"
	"cli-notes/scripts/data"
	"fmt"
	"strings"
)

// RenderObjectivesListView renders the objectives list view
func RenderObjectivesListView(state *data.ObjectivesViewState) string {
	var output strings.Builder

	// Clear screen
	output.WriteString("\033[2J\033[H")

	// Header
	output.WriteString("================================\n")
	output.WriteString("OBJECTIVES\n")
	output.WriteString("================================\n\n")

	if len(state.Objectives) == 0 {
		output.WriteString("No objectives found.\n\n")
		output.WriteString("Press 'n' to create a new objective.\n")
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
	output.WriteString("j/k=navigate, o=open, n=create, l=link, dd=delete, q=quit\n")

	return output.String()
}

// RenderSingleObjectiveView renders a single objective with its children
func RenderSingleObjectiveView(state *data.ObjectivesViewState) string {
	var output strings.Builder

	if state.CurrentObjective == nil {
		return "Error: No objective selected\n"
	}

	// Clear screen
	output.WriteString("\033[2J\033[H")

	// Header
	output.WriteString("================================\n")
	output.WriteString(fmt.Sprintf("OBJECTIVE: %s [%s]\n",
		state.CurrentObjective.Title,
		state.CurrentObjective.ObjectiveID))
	output.WriteString("================================\n\n")

	// Parent content - show selection indicator if on parent
	if state.OnParent {
		output.WriteString("> ")
	} else {
		output.WriteString("  ")
	}

	// Display parent content
	content := state.CurrentObjective.Content
	// Split content into lines and indent each line
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if i == 0 {
			output.WriteString(line)
			output.WriteString("\n")
		} else {
			output.WriteString("  ")
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	// Separator
	output.WriteString("─────────────────────────────────\n")

	// Get incomplete and complete children separately
	incomplete, complete := state.GetIncompleteAndCompleteChildren()
	incompleteCnt, completeCnt := state.GetCompletionCounts()

	output.WriteString(fmt.Sprintf("LINKED TODOS (%d incomplete, %d complete)\n",
		incompleteCnt, completeCnt))
	output.WriteString("─────────────────────────────────\n\n")

	if len(state.Children) == 0 {
		output.WriteString("No linked todos yet.\n\n")
	} else {
		// Render incomplete todos
		if state.FilterMode != data.ShowCompleteOnly && len(incomplete) > 0 {
			output.WriteString("INCOMPLETE:\n")
			renderChildrenList(&output, incomplete, 0, state)
			output.WriteString("\n")
		}

		// Render complete todos
		if state.FilterMode != data.ShowIncompleteOnly && len(complete) > 0 {
			output.WriteString("COMPLETE:\n")
			renderChildrenList(&output, complete, len(incomplete), state)
		}
	}

	output.WriteString("\n")
	output.WriteString("j/k=navigate, o=open, n=new child, l=link existing,\n")
	output.WriteString("e=edit parent, u=unlink, s=sort, f=filter,\n")
	output.WriteString("1/2/3=priority, t=due today, m/tu/w/th/f/sa/su=due day, q=back\n")

	return output.String()
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
