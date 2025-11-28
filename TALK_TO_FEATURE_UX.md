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

### Multiple Tags on Same Todo
- A todo can have multiple to-talk tags: `- [ ] Discuss budget to-talk-michael to-talk-alice`
- The todo will appear in both person's lists (Michael and Alice)
- When moved from either person's view, the todo is marked complete in the original note
- This automatically removes it from all other person's views (since completed todos are filtered out)

### Todo Text Formatting
- When moving todos to target note, **remove the to-talk-X tag(s)** from the text
- Example: `- [ ] Figure out how to do y to-talk-michael` becomes `- [ ] Figure out how to do y`
- This keeps the target note clean since the context is now implicit (it's in a note for that person)

### Long Todo Titles
- Long todo titles should **wrap to multiple lines** in the list view
- If there are more todos than fit on screen, the list should be **scrollable**
- Navigation with j/k scrolls through all items, showing only what fits in the terminal viewport

### Nested/Indented Todos
- When a todo with to-talk-X tag has indented subtasks, **move the main task and all its subtasks together**
- Example source:
  ```
  - [ ] Main task to-talk-michael
    - [ ] Subtask 1
    - [ ] Subtask 2
  ```
- All three items are moved to the target note (with tag removed from main task)
- In the original note, only the main task is marked complete (subtasks remain as-is)

### Sorting Behavior
- **Person list**: Sort alphabetically by name (case-insensitive)
- **Todo list**: Sort by source file name, then by order within file
- **Note search results**: Sort by best match score (fuzzy search ranking)

### File Operation Error Handling
- Move operations are **atomic**: all-or-nothing
- If any file write fails during the move, **rollback all changes**
- Show error message explaining which file failed and that no changes were made
- User can retry the operation after resolving the issue

### Concurrent Modifications
- Work with snapshot of files taken when view is opened
- **Detect changes before move**: Check if source/target files have been modified externally
- If changes detected, show warning: "Files have been modified. Refresh view? (y/n)"
- User can refresh to reload current state or cancel the operation

### Same Source and Target Note
- **Allow moving todos within the same file**
- If target note is the same as the source note, the todo is moved to the top (after frontmatter)
- The original location is marked as complete
- This is useful for reorganizing todos within a file

### Additional Features
- **Keep it simple**: No additional shortcuts or quick actions beyond the core navigation and selection commands
- Focus on a clear, predictable workflow rather than adding complexity

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

## UI Mockups

### 1. Person Selection View

**Standard view (few items):**
```
┌──────────────────────────────────────────────────────────────────────────┐
│ TALK TO - Select Person                                                  │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│   > alice (12 items)                                                      │
│     bob (3 items)                                                         │
│     michael (7 items)                                                     │
│     sarah (1 item)                                                        │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
├──────────────────────────────────────────────────────────────────────────┤
│ j/k=navigate • Enter=select person • q=quit                               │
└──────────────────────────────────────────────────────────────────────────┘
```

**Scrollable view (many items):**
```
┌──────────────────────────────────────────────────────────────────────────┐
│ TALK TO - Select Person                                    (Showing 5-14) │
├──────────────────────────────────────────────────────────────────────────┤
│   ↑ (more above)                                                          │
│     david (2 items)                                                       │
│     emma (5 items)                                                        │
│     frank (1 item)                                                        │
│   > george (8 items)                                                      │
│     hannah (3 items)                                                      │
│     ian (12 items)                                                        │
│     jessica (4 items)                                                     │
│     kevin (6 items)                                                       │
│     laura (2 items)                                                       │
│   ↓ (more below)                                                          │
│                                                                           │
│                                                                           │
│                                                                           │
├──────────────────────────────────────────────────────────────────────────┤
│ j/k=navigate • Enter=select person • q=quit                               │
└──────────────────────────────────────────────────────────────────────────┘
```

**Behavior:**
- Shows only items that fit in viewport (10-12 items depending on terminal height)
- `↑` and `↓` indicators when there's more content above/below
- Header shows position when scrolling: "(Showing X-Y)"
- Navigation with j/k scrolls to keep selected item visible
- Selection wraps around (bottom → top, top → bottom)

### 2. Todo Selection View

**Standard view:**
```
┌──────────────────────────────────────────────────────────────────────────┐
│ TALK TO MICHAEL - Select Items to Move              (5 found, 5 selected) │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│ > [x] Figure out how to implement the new auth system                    │
│       (from: project-planning.md)                                        │
│                                                                           │
│   [x] Discuss budget concerns for Q2                                     │
│       (from: q1-review.md)                                               │
│                                                                           │
│   [x] Ask about timeline for the redesign project and whether we can    │
│       push it to next quarter                                            │
│       (from: sprint-notes.md)                                            │
│                                                                           │
│   [x] Review architecture doc                                            │
│       (from: tech-notes.md)                                              │
│                                                                           │
├──────────────────────────────────────────────────────────────────────────┤
│ j/k=nav • space=toggle • a=all • n=none • Enter=continue • q=back        │
└──────────────────────────────────────────────────────────────────────────┘
```

**Scrollable view (many items):**
```
┌──────────────────────────────────────────────────────────────────────────┐
│ TALK TO MICHAEL - Select Items to Move            (15 found, 12 selected) │
│                                                           (Showing 4-8/15) │
├──────────────────────────────────────────────────────────────────────────┤
│   ↑ (3 more above)                                                        │
│                                                                           │
│   [x] Ask about timeline for the redesign project and whether we can     │
│       push it to next quarter                                             │
│       (from: sprint-notes.md)                                             │
│                                                                           │
│ > [ ] Review architecture doc                                             │
│       (from: tech-notes.md)                                               │
│                                                                           │
│   [x] Schedule follow-up meeting next week                                │
│       (from: weekly-notes.md)                                             │
│                                                                           │
│   ↓ (7 more below)                                                        │
│                                                                           │
├──────────────────────────────────────────────────────────────────────────┤
│ j/k=nav • space=toggle • a=all • n=none • Enter=continue • q=back         │
└──────────────────────────────────────────────────────────────────────────┘
```

**Behavior:**
- Header shows person name and counts (X found, Y selected)
- `>` indicates cursor position on current todo
- `[x]` = selected for move, `[ ]` = not selected
- All todos selected by default when view opens
- Long todo titles wrap to multiple lines
- Source file shown in parentheses below each todo
- Scrollable when more todos than fit in viewport
- Shows "(Showing X-Y/Total)" and scroll indicators when scrolling

### 3. Target Note Selection View

```
┌──────────────────────────────────────────────────────────────────────────┐
│ SELECT TARGET NOTE                                                        │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│                                                                           │
│                                                                           │
│   3 items ready to move                                                  │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
├──────────────────────────────────────────────────────────────────────────┤
│ f=find existing note • n=create new note • q=back                        │
└──────────────────────────────────────────────────────────────────────────┘
```

**Behavior:**
- Simple prompt view showing count of items ready to move
- Waits for user to press f (find), n (new), or q (back)
- Pressing 'f' opens the Note Search Modal
- Pressing 'n' opens the Create Note prompt

### 4. Note Search Modal

```
┌──────────────────────────────────────────────────────────────────────────┐
│ SELECT TARGET NOTE                                                        │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│     ┌─ FIND NOTE ──────────────────────────────────────┐                │
│     │                                                    │                │
│     │ Search: mich_                        [INSERT MODE] │                │
│     │                                                    │                │
│     │ > one-to-one-michael.md                           │                │
│     │   michael-project-notes.md                        │                │
│     │   team-michael-feedback.md                        │                │
│     │   quarterly-review-michael.md                     │                │
│     │                                                    │                │
│     │                                                    │                │
│     │ Esc=normal mode • i=insert mode                   │                │
│     │ arrows/j/k=navigate • Enter=select • q=cancel     │                │
│     └────────────────────────────────────────────────────┘                │
│                                                                           │
│                                                                           │
│                                                                           │
├──────────────────────────────────────────────────────────────────────────┤
│ [Finding note...]                                                         │
└──────────────────────────────────────────────────────────────────────────┘
```

**Behavior:**
- Modal box overlaid on the Target Note Selection view
- Starts in INSERT mode - typing immediately filters results
- Shows current mode in header: [INSERT MODE] or [NORMAL MODE]
- Real-time fuzzy search filtering as user types
- `>` indicates currently selected note in results
- Navigation:
  - INSERT mode: arrow keys up/down to navigate results
  - NORMAL mode: j/k to navigate
  - Esc switches to NORMAL mode, i switches to INSERT mode
- Enter selects the highlighted note
- q or Esc (in NORMAL mode) cancels and returns to Target Note Selection
- Scrollable if many matching results

### 5. Confirmation View

```
┌──────────────────────────────────────────────────────────────────────────┐
│ CONFIRM MOVE                                                              │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│ Moving 3 items to: one-to-one-michael.md                                 │
│                                                                           │
│   • Figure out how to implement the new auth system                      │
│     (from: project-planning.md)                                          │
│                                                                           │
│   • Discuss budget concerns for Q2                                       │
│     (from: q1-review.md)                                                 │
│                                                                           │
│   • Ask about timeline for the redesign project                          │
│     (from: sprint-notes.md)                                              │
│                                                                           │
│ These items will be:                                                     │
│   • Added to top of one-to-one-michael.md (below frontmatter)            │
│   • Marked as complete in their original notes                           │
│   • Tracked for undo                                                     │
│                                                                           │
│ Continue?                                                                │
│                                                                           │
├──────────────────────────────────────────────────────────────────────────┤
│ y/Enter=execute • Esc/c=cancel, back to selection • q=quit to main       │
└──────────────────────────────────────────────────────────────────────────┘
```

**Behavior:**
- Shows summary of the move operation before executing
- Lists all selected todos with their source files
- Explains what will happen (added to target, marked complete in sources, tracked for undo)
- Three action options:
  - y or Enter: Execute the move
  - Esc or c: Cancel and return to Todo Selection to change selection
  - q: Quit entirely and return to main CLI

### 6. Success View

```
┌──────────────────────────────────────────────────────────────────────────┐
│ MOVE COMPLETED                                                            │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│                                                                           │
│   ✓ Successfully moved 3 items to one-to-one-michael.md                  │
│                                                                           │
│                                                                           │
│   Modified files:                                                         │
│     • one-to-one-michael.md (3 items added)                               │
│     • project-planning.md (1 item completed)                              │
│     • q1-review.md (1 item completed)                                     │
│     • sprint-notes.md (1 item completed)                                  │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
│                                                                           │
├──────────────────────────────────────────────────────────────────────────┤
│ Enter=open note • u=undo • r=return to person selection • q=quit         │
└──────────────────────────────────────────────────────────────────────────┘
```

**Behavior:**
- Displays after successful move operation
- Shows success confirmation with checkmark
- Summary of items moved and target note
- Lists all modified files with change counts
- Four action options:
  - Enter: Open the target note in editor
  - u: Undo the move operation (reverses all changes)
  - r: Return to person selection to move more todos
  - q: Quit to main CLI

## Next Steps
1. Create state transition diagrams showing flow between views
2. Plan technical implementation details
3. Define data structures and file formats
