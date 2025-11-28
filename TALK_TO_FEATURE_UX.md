# Talk-To Feature - UI/UX Design

## Overview
A new interactive view for managing "to-talk-X" tagged todos across notes. Users can collect todos tagged for specific people, select which ones to move, and add them to a target note while marking them complete in their original locations.

## Tag Format
Todos are tagged inline at the end of the todo line:
```
- [ ] Figure out how to do y to-talk-michael
- [ ] Discuss budget to-talk-alice
```

Only incomplete todos (`- [ ]`) are shown in the view. Completed todos (`- [x]`) are ignored.

## User Flow

### 1. Command Entry

**Command: `tt`** - Opens view with all to-talk-X todos grouped by person
**Command: `tt [name]`** - Opens view filtered to specific person (e.g., `tt michael`)

### 2. Person Selection View (when using `tt` without name)

```
TALK TO - Select Person

  > Michael (5 items)
    Alice (2 items)
    Bob (1 item)

[j/k=navigate, Enter=select person, q=quit]
```

**Behavior**:
- List shows all unique people with pending to-talk-X todos
- Shows count of incomplete todos for each person
- Navigation: j/k (vim-style) to move up/down
- Enter key selects the person and transitions to Todo Selection view
- q quits back to main CLI

### 3. Todo Selection View

Displayed when:
- User selects a person from Person Selection view (via Enter)
- User enters directly with `tt [name]`

```
TALK TO MICHAEL - Select Items to Move (5 found, 5 selected)

  [x] Figure out how to do y (from: Project Planning.md)
  [x] Discuss budget concerns (from: Q1 Review.md)
  [x] Ask about timeline (from: Sprint Notes.md)
  [x] Review architecture doc (from: Tech Notes.md)
  [x] Schedule follow-up (from: Weekly Notes.md)

[j/k=navigate, space=toggle selection, a=select all, n=none,
 Enter=continue to note selection, q=back/quit]
```

**Behavior**:
- All todos selected by default (`[x]` checkboxes)
- j/k navigation to move cursor
- Space toggles selection on current item
- `a` selects all items
- `n` deselects all items (none selected)
- Enter advances to Target Note Selection (only if at least 1 item selected)
- q goes back to Person Selection (or quits if entered via `tt [name]`)

### 4. Target Note Selection View

```
SELECT TARGET NOTE (3 items ready to move)

[f=find existing note, n=create new note, q=back]
```

**Pressing `f` - Find Existing Note (Modal Search)**:

Opens a side-panel search interface with vim-style modal editing:

```
SELECT TARGET NOTE (3 items ready to move)

┌─ FIND NOTE ─────────────────────┐
│ Search: mich_                   │
│                                  │
│ > one-to-one-michael.md         │
│   michael-project-notes.md      │
│   team-michael-feedback.md      │
│                                  │
│ [type to search]                │
│ [Esc=normal mode, i=insert]     │
│ [j/k=navigate, Enter=select]    │
└──────────────────────────────────┘
```

**Search Behavior**:
- Starts in INSERT mode (typing immediately filters)
- Real-time filtering: every keystroke updates matching notes list
- Navigate results:
  - While in INSERT mode: Arrow keys up/down to navigate results
  - Press Esc to enter NORMAL mode: use j/k to navigate
  - Press `i` to return to INSERT mode
- Enter selects the highlighted note and proceeds to Confirmation
- Esc (when in NORMAL mode) or `q` cancels and returns to Target Note Selection

**Pressing `n` - Create New Note**:

Prompts for note title:
```
Create new note: one-to-one with michael_

[Enter=create, Esc=cancel]
```

**Behavior**:
- User types title
- Backspace supported for editing
- Enter creates note using standard note creation flow
- Note created with today's date
- Proceeds to Confirmation view

### 5. Confirmation View

```
CONFIRM MOVE

Moving 3 items to: one-to-one-michael.md

  - Figure out how to do y (from: Project Planning.md)
  - Discuss budget concerns (from: Q1 Review.md)
  - Ask about timeline (from: Sprint Notes.md)

These items will be:
  • Added to top of one-to-one-michael.md (below title/frontmatter)
  • Marked as complete in their original notes
  • Tracked for undo

Continue? (y/n/c)
```

**Behavior**:
- y = Execute move, update all files, show success message
- n = Cancel and return to main CLI
- c = Cancel and return to Todo Selection view (restart selection process)

### 6. Post-Move State & Undo

**After successful move**:
```
✓ Successfully moved 3 items to one-to-one-michael.md

[u=undo, q=quit to main, Enter=return to person selection]
```

**Undo System** (similar to Weekly Planner):
- All moves tracked in session
- `u` key undoes the last move operation
- Undo reverses the entire batch:
  - Removes todos from target note
  - Marks them as incomplete in original notes
- Changes held in memory until user quits or saves
- On quit: prompt to save/discard changes if any pending

## UI State Management

### TalkToViewState Structure
```
- ViewMode: PersonSelection | TodoSelection | NoteSelection | Confirmation
- SelectedPerson: string (e.g., "michael")
- AllPeople: map[string]int (person -> count)
- PersonIndex: int (selection in person list)
- AvailableTodos: []File (all todos for selected person)
- SelectedTodos: []bool (parallel array for selections)
- TodoIndex: int (current cursor position)
- TargetNote: string (selected or created note)
- Changes: []MoveChange (track all moves for undo)
- UndoStack: []MoveChange
```

### Navigation Patterns
- **j/k**: Up/down navigation (vim-style)
- **Enter**: Confirm/select/advance
- **Space**: Toggle selection (in todo list)
- **q**: Back/quit
- **f**: Find note (in note selection)
- **n**: New note (in note selection) OR deselect all (in todo selection)
- **a**: Select all (in todo selection)
- **u**: Undo last move
- **Esc**: Mode switching in search (insert→normal)
- **i**: Mode switching in search (normal→insert)

## File Operations

### Reading To-Talk Todos
- Scan all markdown files in notes directory
- Parse todo lines matching pattern: `- [ ] .* to-talk-(\w+)`
- Extract person name from tag
- Filter out completed todos (`- [x]`)
- Group by person name

### Moving Todos
1. Read target note file
2. Insert todos at top (after YAML frontmatter and title)
3. Write updated target note
4. For each original note:
   - Read file
   - Find matching todo line
   - Change `- [ ]` to `- [x]`
   - Write updated file
5. Track change in undo stack

### Undo Move
1. Read target note
2. Remove inserted todos
3. Write target note
4. For each original note:
   - Read file
   - Change `- [x]` back to `- [ ]`
   - Write file

## Edge Cases & Behavior Rules

### No Todos Found
- **`tt` with no to-talk-X todos anywhere**: Show error message "No to-talk items found" and return immediately to command prompt (don't open view)
- **`tt michael` with no to-talk-michael todos**: Print message "No to-talk-michael items found" and return to command prompt (don't open view)

### Case Sensitivity
- Person names are **case-insensitive**
- `to-talk-Michael`, `to-talk-michael`, `to-talk-MICHAEL` are all treated as the same person
- Display using lowercase in the person selection view (normalized form)
- When filtering by command (`tt michael`), match case-insensitively

### New Note Creation
- When user presses `n` to create a new note:
  - Prompt for title
  - Create note using standard todo format (same as `ct` command)
  - Include standard YAML frontmatter: title, date-created, tags (empty or minimal)
  - After moving todos into the new note, **open it immediately in the editor**
  - This allows user to add context/meeting notes around the moved todos

### Post-Move Behavior
- After successfully moving todos to a **new** note: Open in editor immediately
- After moving todos to an **existing** note: Show success message, don't auto-open (user can manually open with `o` if needed)

## Architecture Alignment

This feature follows existing patterns:
- **State Management**: Similar to `ObjectivesViewState` and `WeekPlannerState`
- **Input Handling**: Similar to `objectives_input.go` and `week_planner_input.go` with action enums
- **UI Rendering**: Similar to `objectives_ui.go` with multiple view rendering functions
- **Multi-note Operations**: Similar to weekly planner's batch update system
- **Undo/Redo**: Follows weekly planner's undo stack pattern
- **Modal Search**: Inspired by objectives' search-and-link flow, enhanced with vim-style modes

## Next Steps
1. Create detailed UI mockups/wireframes
2. Design state transition diagrams
3. Plan technical implementation details
