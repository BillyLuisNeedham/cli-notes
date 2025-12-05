package presentation

import (
	"cli-notes/scripts/data"
	"fmt"
	"strings"
)

// RenderTalkToView is the main dispatcher for rendering different view modes
func RenderTalkToView(state *data.TalkToViewState, termWidth, termHeight int) string {
	switch state.ViewMode {
	case data.PersonSelectionView:
		return RenderPersonSelection(state, termWidth, termHeight)
	case data.TodoSelectionView:
		return RenderTodoSelection(state, termWidth, termHeight)
	case data.NoteSelectionView:
		return RenderNoteSelection(state, termWidth, termHeight)
	case data.NoteSearchModalView:
		return RenderNoteSearchModal(state, termWidth, termHeight)
	case data.ConfirmationView:
		return RenderConfirmation(state, termWidth, termHeight)
	case data.SuccessView:
		return RenderSuccess(state, termWidth, termHeight)
	}

	return ""
}

// RenderPersonSelection renders the person selection view
func RenderPersonSelection(state *data.TalkToViewState, termWidth, termHeight int) string {
	var builder strings.Builder

	builder.WriteString("\033[2J\033[H")

	builder.WriteString("┌" + strings.Repeat("─", termWidth-2) + "┐\n")
	builder.WriteString("│ TALK TO - Select Person" + strings.Repeat(" ", termWidth-27) + "│\n")
	builder.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")

	contentHeight := termHeight - 9 // Header (3 lines) + footer (2 lines) + border (1 line) + message area (3 lines)
	visibleStart := 0
	visibleEnd := len(state.AllPeople)

	showScrollUp := false
	showScrollDown := false

	if len(state.AllPeople) > contentHeight {
		visibleStart = state.PersonIndex - contentHeight/2
		if visibleStart < 0 {
			visibleStart = 0
		}
		visibleEnd = visibleStart + contentHeight

		if visibleStart > 0 {
			showScrollUp = true
			visibleStart++
			contentHeight--
		}

		if visibleEnd > len(state.AllPeople) {
			visibleEnd = len(state.AllPeople)
			visibleStart = visibleEnd - contentHeight
			if visibleStart < 0 {
				visibleStart = 0
			}
		} else {
			showScrollDown = true
			contentHeight--
		}
	}

	if showScrollUp {
		builder.WriteString("│   ↑ (more above)" + strings.Repeat(" ", termWidth-20) + "│\n")
	}

	for i := visibleStart; i < visibleEnd; i++ {
		person := state.AllPeople[i]

		indicator := "  "
		if i == state.PersonIndex {
			indicator = "> "
		}

		itemText := fmt.Sprintf("%s (", person.Name)
		if person.Count == 1 {
			itemText += "1 item)"
		} else {
			itemText += fmt.Sprintf("%d items)", person.Count)
		}

		line := fmt.Sprintf("│ %s%s", indicator, itemText)
		padding := termWidth - runeCount(line) - 1
		if padding < 0 {
			padding = 0
		}
		line += strings.Repeat(" ", padding) + "│\n"

		builder.WriteString(line)
	}

	if showScrollDown {
		builder.WriteString("│   ↓ (more below)" + strings.Repeat(" ", termWidth-20) + "│\n")
	}

	currentLines := visibleEnd - visibleStart
	if showScrollUp {
		currentLines++
	}
	if showScrollDown {
		currentLines++
	}

	for i := currentLines; i < contentHeight+1; i++ {
		builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")
	}

	builder.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")
	builder.WriteString("│ j/k=navigate • Enter=select • q=quit" + strings.Repeat(" ", termWidth-41) + "│\n")
	builder.WriteString("└" + strings.Repeat("─", termWidth-2) + "┘\n")

	return builder.String()
}

// RenderTodoSelection renders the todo selection view
func RenderTodoSelection(state *data.TalkToViewState, termWidth, termHeight int) string {
	var builder strings.Builder

	builder.WriteString("\033[2J\033[H")

	totalCount := len(state.AvailableTodos)
	selectedCount := state.GetSelectedCount()

	personName := strings.ToUpper(state.SelectedPerson)
	headerText := fmt.Sprintf("TALK TO %s - Select Items (%d found, %d selected)",
		personName, totalCount, selectedCount)

	builder.WriteString("┌" + strings.Repeat("─", termWidth-2) + "┐\n")

	padding := termWidth - 2 - runeCount(headerText)
	if padding < 0 {
		headerText = truncateString(headerText, termWidth-4)
		padding = 0
	}
	builder.WriteString("│ " + headerText + strings.Repeat(" ", padding-1) + "│\n")
	builder.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")

	contentHeight := termHeight - 9

	for i := 0; i < len(state.AvailableTodos) && i < contentHeight; i++ {
		todo := state.AvailableTodos[i]

		cursor := "  "
		if i == state.TodoIndex {
			cursor = "> "
		}

		checkbox := "[ ] "
		if i < len(state.SelectedTodos) && state.SelectedTodos[i] {
			checkbox = "[x] "
		}

		todoText := todo.TodoLine
		todoText = strings.TrimPrefix(todoText, "- [ ] ")
		todoText = strings.TrimSpace(todoText)

		line := fmt.Sprintf("│ %s%s%s", cursor, checkbox, todoText)

		lineLen := runeCount(line)
		if lineLen > termWidth-1 {
			line = truncateString(line, termWidth-4) + "│\n"
		} else {
			line += strings.Repeat(" ", termWidth-lineLen-1) + "│\n"
		}

		builder.WriteString(line)

		sourceText := fmt.Sprintf("(from: %s)", todo.SourceFile)
		sourceLine := "│     " + sourceText
		sourceLen := runeCount(sourceLine)
		if sourceLen > termWidth-1 {
			sourceLine = truncateString(sourceLine, termWidth-4) + "│\n"
		} else {
			sourceLine += strings.Repeat(" ", termWidth-sourceLen-1) + "│\n"
		}
		builder.WriteString(sourceLine)
	}

	usedLines := len(state.AvailableTodos) * 2 // Each todo uses 2 lines
	if usedLines > contentHeight {
		usedLines = contentHeight
	}

	for i := usedLines; i < contentHeight; i++ {
		builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")
	}

	builder.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")
	builder.WriteString("│ j/k=nav • space=toggle • a=all • n=none • Enter=continue • q=back" +
		strings.Repeat(" ", termWidth-71) + "│\n")
	builder.WriteString("└" + strings.Repeat("─", termWidth-2) + "┘\n")

	return builder.String()
}

// RenderNoteSelection renders the note selection prompt
func RenderNoteSelection(state *data.TalkToViewState, termWidth, termHeight int) string {
	var builder strings.Builder

	builder.WriteString("\033[2J\033[H")

	builder.WriteString("┌" + strings.Repeat("─", termWidth-2) + "┐\n")
	builder.WriteString("│ SELECT TARGET NOTE" + strings.Repeat(" ", termWidth-22) + "│\n")
	builder.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")

	selectedCount := state.GetSelectedCount()
	message := fmt.Sprintf("%d items ready to move", selectedCount)

	availableHeight := termHeight - 9

	for i := 0; i < availableHeight/2; i++ {
		builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")
	}

	// Center message
	msgPadding := (termWidth - 2 - runeCount(message)) / 2
	builder.WriteString("│" + strings.Repeat(" ", msgPadding) + message +
		strings.Repeat(" ", termWidth-2-msgPadding-runeCount(message)) + "│\n")

	for i := availableHeight/2 + 1; i < availableHeight; i++ {
		builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")
	}

	builder.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")
	builder.WriteString("│ f=find existing note • n=create new note • q=back" +
		strings.Repeat(" ", termWidth-55) + "│\n")
	builder.WriteString("└" + strings.Repeat("─", termWidth-2) + "┘\n")

	return builder.String()
}

// RenderNoteSearchModal renders the search modal with INSERT/NORMAL modes
func RenderNoteSearchModal(state *data.TalkToViewState, termWidth, termHeight int) string {
	var builder strings.Builder

	builder.WriteString("\033[2J\033[H")

	builder.WriteString("┌" + strings.Repeat("─", termWidth-2) + "┐\n")
	builder.WriteString("│ SELECT TARGET NOTE" + strings.Repeat(" ", termWidth-22) + "│\n")
	builder.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")

	modalWidth := termWidth - 10
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalWidth > 70 {
		modalWidth = 70
	}

	modalHeight := 16

	modalLeft := (termWidth - modalWidth) / 2
	modalTop := (termHeight - modalHeight) / 2

	for i := 3; i < modalTop; i++ {
		builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")
	}

	builder.WriteString("│" + strings.Repeat(" ", modalLeft-1))
	builder.WriteString("┌─ FIND NOTE " + strings.Repeat("─", modalWidth-15) + "┐")
	builder.WriteString(strings.Repeat(" ", termWidth-modalLeft-modalWidth-1) + "│\n")

	modeIndicator := "[INSERT MODE]"
	if state.SearchMode == data.NormalMode {
		modeIndicator = "[NORMAL MODE]"
	}
	modeSpaces := strings.Repeat(" ", modalWidth-2-runeCount(modeIndicator))
	builder.WriteString("│" + strings.Repeat(" ", modalLeft-1) + "│ ")
	builder.WriteString(modeSpaces)
	builder.WriteString(modeIndicator)
	builder.WriteString(" │")
	builder.WriteString(strings.Repeat(" ", termWidth-modalLeft-modalWidth-1) + "│\n")

	queryLine := fmt.Sprintf(" Search: %s_", state.SearchQuery)
	queryPadding := strings.Repeat(" ", max(0, modalWidth-2-runeCount(queryLine)))
	builder.WriteString("│" + strings.Repeat(" ", modalLeft-1) + "│")
	builder.WriteString(queryLine)
	builder.WriteString(queryPadding)
	builder.WriteString("│")
	builder.WriteString(strings.Repeat(" ", termWidth-modalLeft-modalWidth-1) + "│\n")

	builder.WriteString("│" + strings.Repeat(" ", modalLeft-1))
	builder.WriteString("├" + strings.Repeat("─", modalWidth-2) + "┤")
	builder.WriteString(strings.Repeat(" ", termWidth-modalLeft-modalWidth-1) + "│\n")

	resultsHeight := modalHeight - 6 // Header + mode + query + separator + footer + border
	startIdx := 0
	endIdx := len(state.SearchResults)

	if len(state.SearchResults) > resultsHeight {
		// Center selection in viewport, or anchor to top/bottom
		startIdx = max(0, state.SearchIndex-resultsHeight/2)
		if startIdx+resultsHeight > len(state.SearchResults) {
			startIdx = max(0, len(state.SearchResults)-resultsHeight)
		}
		endIdx = min(len(state.SearchResults), startIdx+resultsHeight)
	}

	visibleResults := state.SearchResults[startIdx:endIdx]
	visibleIndex := state.SearchIndex - startIdx

	if startIdx > 0 {
		indicator := "  ↑ (more above)"
		padding := strings.Repeat(" ", modalWidth-2-runeCount(indicator))
		builder.WriteString("│" + strings.Repeat(" ", modalLeft-1) + "│ ")
		builder.WriteString(indicator)
		builder.WriteString(padding)
		builder.WriteString(" │")
		builder.WriteString(strings.Repeat(" ", termWidth-modalLeft-modalWidth-1) + "│\n")
	}

	for i, result := range visibleResults {
		cursor := "  "
		if i == visibleIndex {
			cursor = "> "
		}

		resultText := cursor + result.Name
		if runeCount(resultText) > modalWidth-4 {
			resultText = truncateString(resultText, modalWidth-7)
		}

		builder.WriteString("│" + strings.Repeat(" ", modalLeft-1) + "│ ")
		builder.WriteString(resultText)
		builder.WriteString(strings.Repeat(" ", modalWidth-2-runeCount(resultText)))
		builder.WriteString(" │")
		builder.WriteString(strings.Repeat(" ", termWidth-modalLeft-modalWidth-1) + "│\n")
	}

	if endIdx < len(state.SearchResults) {
		indicator := "  ↓ (more below)"
		padding := strings.Repeat(" ", modalWidth-2-runeCount(indicator))
		builder.WriteString("│" + strings.Repeat(" ", modalLeft-1) + "│ ")
		builder.WriteString(indicator)
		builder.WriteString(padding)
		builder.WriteString(" │")
		builder.WriteString(strings.Repeat(" ", termWidth-modalLeft-modalWidth-1) + "│\n")
	}

	usedLines := len(visibleResults)
	if startIdx > 0 {
		usedLines++
	}
	if endIdx < len(state.SearchResults) {
		usedLines++
	}
	for i := usedLines; i < resultsHeight; i++ {
		builder.WriteString("│" + strings.Repeat(" ", modalLeft-1) + "│ " +
			strings.Repeat(" ", modalWidth-2) + " │" +
			strings.Repeat(" ", termWidth-modalLeft-modalWidth-1) + "│\n")
	}

	builder.WriteString("│" + strings.Repeat(" ", modalLeft-1) + "│ ")
	footerText := "Esc=toggle mode • i=insert • j/k=nav • Enter=select • q=cancel"
	if runeCount(footerText) > modalWidth-4 {
		footerText = truncateString(footerText, modalWidth-7)
	}
	builder.WriteString(footerText)
	builder.WriteString(strings.Repeat(" ", modalWidth-2-runeCount(footerText)))
	builder.WriteString(" │")
	builder.WriteString(strings.Repeat(" ", termWidth-modalLeft-modalWidth-1) + "│\n")

	builder.WriteString("│" + strings.Repeat(" ", modalLeft-1))
	builder.WriteString("└" + strings.Repeat("─", modalWidth-2) + "┘")
	builder.WriteString(strings.Repeat(" ", termWidth-modalLeft-modalWidth-1) + "│\n")

	currentLine := modalTop + modalHeight
	fillEnd := termHeight - 20 // EXTREME: Leave lots of room to see header
	for i := currentLine; i < fillEnd; i++ {
		builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")
	}

	builder.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")
	builder.WriteString("│ [Finding note...]" + strings.Repeat(" ", termWidth-22) + "│\n")
	builder.WriteString("└" + strings.Repeat("─", termWidth-2) + "┘\n")

	return builder.String()
}

// RenderConfirmation renders the confirmation view
func RenderConfirmation(state *data.TalkToViewState, termWidth, termHeight int) string {
	var builder strings.Builder

	builder.WriteString("\033[2J\033[H")

	builder.WriteString("┌" + strings.Repeat("─", termWidth-2) + "┐\n")
	builder.WriteString("│ CONFIRM MOVE" + strings.Repeat(" ", termWidth-16) + "│\n")
	builder.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")
	builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")

	// Summary line
	selectedTodos := state.GetSelectedTodos()
	summaryLine := fmt.Sprintf("Moving %d items to: %s", len(selectedTodos), state.TargetNoteName)
	builder.WriteString("│ " + summaryLine + strings.Repeat(" ", termWidth-3-runeCount(summaryLine)) + "│\n")
	builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")

	maxItems := termHeight - 18 // Leave room for header, footer, explanation, and message area
	for i, todo := range selectedTodos {
		if i >= maxItems {
			builder.WriteString("│   ... and " + fmt.Sprintf("%d", len(selectedTodos)-i) + " more" +
				strings.Repeat(" ", termWidth-16-len(fmt.Sprintf("%d", len(selectedTodos)-i))) + "│\n")
			break
		}

		todoText := strings.TrimPrefix(todo.TodoLine, "- [ ] ")
		todoText = strings.TrimSpace(todoText)
		if runeCount(todoText) > termWidth-12 {
			todoText = truncateString(todoText, termWidth-15)
		}

		line := fmt.Sprintf("│   • %s", todoText)
		padding := termWidth - runeCount(line) - 1
		if padding < 0 {
			padding = 0
		}
		builder.WriteString(line + strings.Repeat(" ", padding) + "│\n")

		sourceText := fmt.Sprintf("(from: %s)", todo.SourceFile)
		sourceLine := "│     " + sourceText
		sourcePadding := termWidth - runeCount(sourceLine) - 1
		if sourcePadding < 0 {
			sourcePadding = 0
		}
		builder.WriteString(sourceLine + strings.Repeat(" ", sourcePadding) + "│\n")
	}

	builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")

	builder.WriteString("│ These items will be:" + strings.Repeat(" ", termWidth-23) + "│\n")
	builder.WriteString("│   • Added to top of " + state.TargetNoteName +
		strings.Repeat(" ", termWidth-25-runeCount(state.TargetNoteName)) + "│\n")
	builder.WriteString("│   • Marked as complete in their original notes" +
		strings.Repeat(" ", termWidth-52) + "│\n")
	builder.WriteString("│   • Tracked for undo" + strings.Repeat(" ", termWidth-24) + "│\n")
	builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")

	builder.WriteString("│ Continue?" + strings.Repeat(" ", termWidth-13) + "│\n")

	currentLines := 12 + (len(selectedTodos) * 2)
	if len(selectedTodos) > maxItems {
		currentLines = 12 + (maxItems * 2) + 1
	}
	for i := currentLines; i < termHeight-6; i++ {
		builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")
	}

	builder.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")
	builder.WriteString("│ y/Enter=execute • c=cancel, back to selection • q=quit to main" +
		strings.Repeat(" ", termWidth-69) + "│\n")
	builder.WriteString("└" + strings.Repeat("─", termWidth-2) + "┘\n")

	return builder.String()
}

// RenderSuccess renders the success view
func RenderSuccess(state *data.TalkToViewState, termWidth, termHeight int) string {
	var builder strings.Builder

	builder.WriteString("\033[2J\033[H")

	builder.WriteString("┌" + strings.Repeat("─", termWidth-2) + "┐\n")
	builder.WriteString("│ MOVE COMPLETED" + strings.Repeat(" ", termWidth-18) + "│\n")
	builder.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")
	builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")
	builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")

	selectedTodos := state.GetSelectedTodos()
	successMsg := fmt.Sprintf("   ✓ Successfully moved %d items to %s", len(selectedTodos), state.TargetNoteName)
	builder.WriteString("│" + successMsg + strings.Repeat(" ", termWidth-2-runeCount(successMsg)) + "│\n")
	builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")
	builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")

	builder.WriteString("│   Modified files:" + strings.Repeat(" ", termWidth-21) + "│\n")

	targetMsg := fmt.Sprintf("     • %s (%d items added)", state.TargetNoteName, len(selectedTodos))
	builder.WriteString("│" + targetMsg + strings.Repeat(" ", termWidth-2-runeCount(targetMsg)) + "│\n")

	sourceFiles := make(map[string]int)
	for _, todo := range selectedTodos {
		sourceFiles[todo.SourceFile]++
	}

	for fileName, count := range sourceFiles {
		itemText := "item"
		if count > 1 {
			itemText = "items"
		}
		sourceMsg := fmt.Sprintf("     • %s (%d %s completed)", fileName, count, itemText)
		builder.WriteString("│" + sourceMsg + strings.Repeat(" ", termWidth-2-runeCount(sourceMsg)) + "│\n")
	}

	currentLines := 9 + len(sourceFiles) + 1
	for i := currentLines; i < termHeight-6; i++ {
		builder.WriteString("│" + strings.Repeat(" ", termWidth-2) + "│\n")
	}

	builder.WriteString("├" + strings.Repeat("─", termWidth-2) + "┤\n")
	builder.WriteString("│ Enter=open note • u=undo • r=return to person selection • q=quit" +
		strings.Repeat(" ", termWidth-71) + "│\n")
	builder.WriteString("└" + strings.Repeat("─", termWidth-2) + "┘\n")

	return builder.String()
}

// Helper functions

// runeCount counts the number of runes in a string (for proper Unicode handling)
func runeCount(s string) int {
	return len([]rune(s))
}

// truncateString truncates a string to a maximum length and adds "..."
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
